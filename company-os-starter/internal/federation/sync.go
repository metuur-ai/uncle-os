package federation

// `workspace sync` and `workspace status` (bin/company-os:2542-2658).
//
// Both return records. Sync returns its PARTIAL record set alongside an error
// when it fails mid-run, because the oracle prints each repo's line as that repo
// completes and only then dies — a caller that dropped the records on error
// would lose stdout the Python CLI had already written.

import (
	"math/big"
	"os"
	"path"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// SyncOptions are the two flags `workspace sync` accepts.
type SyncOptions struct {
	// Frozen materializes strictly from the lock, with no network.
	Frozen bool
	// Only restricts the run to one repo by name; other repos keep their
	// existing lock entries.
	Only string
}

// RepoSync is one repo's outcome. Restored distinguishes the two render shapes:
// the online path names the pin it fetched, the frozen path does not because it
// used the lock's resolved commit instead.
type RepoSync struct {
	Name     string
	SHA      string
	Pin      Pin
	Targets  []string
	Files    int
	Restored bool
}

// SyncResult is what one `workspace sync` did. It is returned even when Sync
// errors, carrying the repos that had already completed.
type SyncResult struct {
	Frozen bool
	// RepoCount is the number of repos this run covers, after --only.
	RepoCount int
	Repos     []RepoSync
	// LockRepos is the entry count written to the lock; meaningless when Frozen,
	// where nothing is written.
	LockRepos int
	// Wrote reports whether the lock was rewritten.
	Wrote bool
	// Complete is false when Sync returned an error partway. The oracle prints
	// its trailer only after the loop finishes, so a caller rendering a partial
	// result must not print one — and "Wrote" cannot stand in for this, because
	// the --frozen path completes successfully without ever writing the lock.
	Complete bool
}

// RepoStatus is one repo's line in `workspace status`.
type RepoStatus struct {
	Name    string
	Pin     Pin
	Targets []string
	// Locked is false when the repo has no lock entry at all — never synced.
	Locked bool
	// SHA is the lock's resolvedCommit, or "?" when the entry omits it.
	SHA string
	// State is the per-file hash verdict. Only meaningful when neither drift
	// flag is set, because the oracle reports drift in preference to it.
	State State
	// PinDrift means the manifest pin no longer matches the lock's.
	PinDrift bool
	// LockPin is repr() of the lock's whole `pin:` mapping, which the drift line
	// interpolates verbatim.
	LockPin string
	// SliceDrift means the slice SET moved without a re-sync — the case the
	// per-file hashes cannot see, because the old files are still there and
	// still hash clean.
	SliceDrift bool
}

// StatusResult is what one `workspace status` found.
type StatusResult struct {
	RepoCount int
	Repos     []RepoStatus
	// ActionNeeded selects the trailing next-step line.
	ActionNeeded bool
}

// Sync is cmd_workspace_sync (bin/company-os:2556-2632).
func Sync(ws *workspace.Workspace, m *Manifest, opts SyncOptions) (*SyncResult, error) {
	if err := RequireGit(); err != nil {
		return nil, err
	}
	repos := m.Repos
	if opts.Only != "" {
		want := yamlio.PyRepr(yamlio.PyStr(opts.Only))
		var filtered []yamlio.PyMap
		for _, r := range repos {
			if yamlio.PyRepr(r.Get("name")) == want {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			return nil, model.Errorf(model.ExitWorkspace,
				"--only '%s' matches no repo in %s", opts.Only, workspace.ManifestName)
		}
		repos = filtered
	}
	baseLock, err := LoadLock(ws)
	if err != nil {
		return nil, err
	}
	if opts.Frozen && !baseLock.Usable {
		return nil, model.Errorf(model.ExitExternalTool,
			"--frozen requires %s, but it is missing or malformed at "+
				"%s. run online first: company-os workspace sync",
			workspace.LockName, ws.Root)
	}
	// `locked = dict(lock_by_name)` — entries for repos this --only run does not
	// touch are preserved verbatim, so a scoped sync cannot silently drop them.
	locked := newEntryList()
	for _, r := range baseLock.Repos {
		locked.set(yamlio.PyRepr(r.Get("name")), r)
	}

	res := &SyncResult{Frozen: opts.Frozen, RepoCount: len(repos)}
	for _, repo := range repos {
		nameVal := repo.Get("name")
		name := yamlio.PyString(nameVal)
		pin, err := RepoPin(repo)
		if err != nil {
			return res, err
		}
		cache := path.Join(ws.Root, workspace.FederationCache, name)
		if opts.Frozen {
			entry, err := syncFrozen(ws, baseLock, nameVal, name, cache)
			if err != nil {
				return res, err
			}
			res.Repos = append(res.Repos, *entry)
			continue
		}
		slices, err := RepoSlices(repo)
		if err != nil {
			return res, err
		}
		sha, err := fetchPinned(cache, yamlio.PyString(repo.Get("url")), pin)
		if err != nil {
			return res, err
		}
		files, err := materializeAll(ws, cache, name, slices, sha)
		if err != nil {
			return res, err
		}
		locked.set(yamlio.PyRepr(nameVal), lockEntry(nameVal, repo.Get("url"),
			slices, pin, sha, files))
		res.Repos = append(res.Repos, RepoSync{
			Name: name, SHA: sha, Pin: pin,
			Targets: targetsLabel(slices), Files: files.Len(),
		})
	}

	if err := ensureCacheGitignored(ws); err != nil {
		return res, err
	}
	if opts.Frozen {
		res.Complete = true
		return res, nil
	}
	ordered := yamlio.PySeq{}
	for _, repo := range m.Repos {
		if e, ok := locked.get(yamlio.PyRepr(repo.Get("name"))); ok {
			ordered = append(ordered, e)
		}
	}
	doc := yamlio.PyMap{
		{K: "version", V: yamlio.PyInt{N: big.NewInt(1)}},
		{K: "generatedFrom", V: yamlio.PyStr(workspace.ManifestName)},
		{K: "repos", V: ordered},
	}
	if err := yamlio.PyWriteFile(LockPath(ws), doc); err != nil {
		return res, err
	}
	res.Wrote = true
	res.LockRepos = len(ordered)
	res.Complete = true
	return res, nil
}

// syncFrozen is the offline branch (bin/company-os:2574-2600). It materializes
// from the LOCK, not from the manifest — otherwise a manifest edit silently
// redirects the "strictly from the lock" path, which is the one guarantee
// --frozen makes.
func syncFrozen(ws *workspace.Workspace, lock *Lock, nameVal yamlio.PyValue, name, cache string) (*RepoSync, error) {
	lr, ok := lock.ByName(nameVal)
	if !ok {
		return nil, model.Errorf(model.ExitExternalTool,
			"--frozen: %s has no entry for repo '%s' (lock "+
				"does not cover the manifest). re-run online: "+
				"company-os workspace sync", workspace.LockName, name)
	}
	rawSlices := lr.Get("slices")
	if yamlio.PyFalsy(rawSlices) {
		return nil, model.Errorf(model.ExitExternalTool,
			"--frozen: %s entry for '%s' predates slice "+
				"recording. re-run online once: company-os workspace sync",
			workspace.LockName, name)
	}
	slices, err := lockSlices(rawSlices, name)
	if err != nil {
		return nil, err
	}
	sha := yamlio.PyString(lr.Get("resolvedCommit"))
	if !isDir(path.Join(cache, ".git")) {
		rel, relErr := relTo(ws.Root, cache)
		if relErr != nil {
			rel = cache
		}
		return nil, model.Errorf(model.ExitExternalTool,
			"--frozen: no local cache for '%s' at "+
				"%s and network is disabled. run "+
				"online first: company-os workspace sync", name, rel)
	}
	files, err := materializeAll(ws, cache, name, slices, sha)
	if err != nil {
		return nil, err
	}
	if AggregateHash(files) != yamlio.PyString(lr.Get("sliceHash")) {
		return nil, model.Errorf(model.ExitExternalTool,
			"--frozen: materialized slice for '%s' does not match the "+
				"hash in %s (cache/objects inconsistent with lock). "+
				"run online: company-os workspace sync", name, workspace.LockName)
	}
	return &RepoSync{
		Name: name, SHA: sha, Restored: true,
		Targets: targetsLabel(slices), Files: files.Len(),
	}, nil
}

// Status is cmd_workspace_status (bin/company-os:2635-2658).
func Status(ws *workspace.Workspace, m *Manifest) (*StatusResult, error) {
	lock, err := LoadLock(ws)
	if err != nil {
		return nil, err
	}
	res := &StatusResult{RepoCount: len(m.Repos)}
	for _, repo := range m.Repos {
		nameVal := repo.Get("name")
		pin, err := RepoPin(repo)
		if err != nil {
			return nil, err
		}
		st := RepoStatus{Name: yamlio.PyString(nameVal), Pin: pin}
		lr, ok := lock.ByName(nameVal)
		if !ok {
			res.Repos = append(res.Repos, st)
			res.ActionNeeded = true
			continue
		}
		st.Locked = true
		st.SHA = "?"
		if v := lr.Get("resolvedCommit"); v != nil {
			st.SHA = yamlio.PyString(v)
		}
		pinLock, _ := lr.Get("pin").(yamlio.PyMap)
		st.LockPin = yamlio.PyRepr(pinLock)
		// Two independent tests, as the oracle writes them: the pinned value must
		// still match, AND the lock's truthy pin keys must be exactly [kind] — so
		// a lock carrying both commit: and tag: is drift even if one of them
		// agrees.
		var truthy []string
		for _, p := range pinLock {
			if !yamlio.PyFalsy(p.V) {
				truthy = append(truthy, p.K)
			}
		}
		st.PinDrift = yamlio.PyRepr(pinLock.Get(pin.Kind)) != yamlio.PyRepr(yamlio.PyStr(pin.Ref)) ||
			len(truthy) != 1 || truthy[0] != pin.Kind
		slices, err := RepoSlices(repo)
		if err != nil {
			return nil, err
		}
		st.Targets = targetsLabel(slices)
		st.SliceDrift = SliceKey(slices) != SliceKeyOf(lr.Get("slices"))
		if st.State, err = SliceState(ws, lr); err != nil {
			return nil, err
		}
		if st.PinDrift || st.SliceDrift || st.State != StateClean {
			res.ActionNeeded = true
		}
		res.Repos = append(res.Repos, st)
	}
	return res, nil
}

// lockEntry builds one flat lock record (bin/company-os:2612-2616). Key order is
// the emitted order and is frozen by examples/federated/workspace.lock.yaml.
func lockEntry(name, url yamlio.PyValue, slices []Slice, pin Pin, sha string, files *yamlio.OrderedMap) yamlio.PyMap {
	sl := yamlio.PySeq{}
	for _, s := range slices {
		paths := yamlio.PySeq{}
		for _, p := range s.Paths {
			paths = append(paths, yamlio.PyStr(p))
		}
		sl = append(sl, yamlio.PyMap{
			{K: "localDirectory", V: yamlio.PyStr(s.LocalDirectory)},
			{K: "paths", V: paths},
		})
	}
	return yamlio.PyMap{
		{K: "name", V: name},
		{K: "url", V: url},
		{K: "slices", V: sl},
		{K: "pin", V: yamlio.PyMap{{K: pin.Kind, V: yamlio.PyStr(pin.Ref)}}},
		{K: "resolvedCommit", V: yamlio.PyStr(sha)},
		{K: "sliceHash", V: yamlio.PyStr(AggregateHash(files))},
		{K: "files", V: filesPyMap(files)},
	}
}

// filesPyMap converts the hash map to an emittable mapping in INSERTION order.
// Keys(), never SortedKeys() — see doc.go's first ordering.
func filesPyMap(files *yamlio.OrderedMap) yamlio.PyMap {
	out := make(yamlio.PyMap, 0, files.Len())
	for _, k := range files.Keys() {
		v, _ := files.Get(k)
		out = append(out, yamlio.PyPair{K: k, V: yamlio.PyStr(v.Value)})
	}
	return out
}

// lockSlices normalizes a lock entry's recorded `slices:` for re-materialization.
// Python indexes both keys directly, so a lock entry missing either is a crash
// there; here it is a code-6 refusal naming the lock, which is the same class of
// answer --frozen's other four failures give.
func lockSlices(v yamlio.PyValue, name string) ([]Slice, error) {
	list, ok := v.(yamlio.PySeq)
	if !ok {
		return nil, model.Errorf(model.ExitExternalTool,
			"--frozen: %s entry for '%s' has a malformed slices: list. "+
				"re-run online once: company-os workspace sync", workspace.LockName, name)
	}
	out := make([]Slice, 0, len(list))
	for _, e := range list {
		m, ok := e.(yamlio.PyMap)
		if !ok || m.Get("localDirectory") == nil || m.Get("paths") == nil {
			return nil, model.Errorf(model.ExitExternalTool,
				"--frozen: %s entry for '%s' has a malformed slices: list. "+
					"re-run online once: company-os workspace sync",
				workspace.LockName, name)
		}
		s := Slice{LocalDirectory: yamlio.PyString(m.Get("localDirectory"))}
		paths, _ := m.Get("paths").(yamlio.PySeq)
		for _, p := range paths {
			s.Paths = append(s.Paths, yamlio.PyString(p))
		}
		out = append(out, s)
	}
	return out, nil
}

// targetsLabel is _targets_label (bin/company-os:2447-2451): a repo's
// destinations in manifest order. A single-slice repo renders exactly as the
// old single-localDirectory output did.
func targetsLabel(slices []Slice) []string {
	out := make([]string, 0, len(slices))
	for _, s := range slices {
		out = append(out, s.LocalDirectory)
	}
	return out
}

// ensureCacheGitignored is ensure_cache_gitignored (bin/company-os:2529-2539).
// The git cache is machine-owned; the lock and the slices ARE tracked.
func ensureCacheGitignored(ws *workspace.Workspace) error {
	gi := path.Join(ws.Root, ".gitignore")
	entry := strings.SplitN(workspace.FederationCache, "/", 2)[0] + "/" // ".company-os/"
	var lines []string
	if raw, err := os.ReadFile(gi); err == nil {
		lines = splitLines(string(raw))
	}
	for _, l := range lines {
		if l == entry || l == workspace.FederationCache+"/" {
			return nil
		}
	}
	f, err := os.OpenFile(gi, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return model.Errorf(model.ExitExternalTool, "cannot write %s: %v", gi, err)
	}
	defer f.Close()
	var b strings.Builder
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		b.WriteString("\n")
	}
	b.WriteString("# company-os federation cache (machine-owned; do not commit)\n")
	b.WriteString(entry + "\n")
	if _, err := f.WriteString(b.String()); err != nil {
		return model.Errorf(model.ExitExternalTool, "cannot write %s: %v", gi, err)
	}
	return nil
}

// splitLines is str.splitlines(): no trailing empty element for a file that ends
// in a newline, which is what makes the blank-separator test above correct.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	out := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

// relTo is Path.relative_to for the one message that renders the cache path
// workspace-relative.
func relTo(base, target string) (string, error) {
	if !strings.HasPrefix(target, base+"/") {
		return "", model.Errorf(model.ExitExternalTool, "%s is not under %s", target, base)
	}
	return strings.TrimPrefix(target, base+"/"), nil
}

// entryList is an insertion-ordered name->entry association list. A Go map would
// be adequate here (the emission order is rebuilt from the manifest), but the
// package rule is that no map whose order could ever become observable is
// ranged over.
type entryList struct {
	keys []string
	vals []yamlio.PyMap
}

func newEntryList() *entryList { return &entryList{} }

func (e *entryList) set(key string, v yamlio.PyMap) {
	for i, k := range e.keys {
		if k == key {
			e.vals[i] = v
			return
		}
	}
	e.keys = append(e.keys, key)
	e.vals = append(e.vals, v)
}

func (e *entryList) get(key string) (yamlio.PyMap, bool) {
	for i, k := range e.keys {
		if k == key {
			return e.vals[i], true
		}
	}
	return nil, false
}
