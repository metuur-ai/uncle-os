package federation

// The git-guarded half of the inherited assertions (ST-077 .. ST-086), plus the
// lock-ordering proof that only a real sync can make.
//
// ST-076 is examples/selftest.py's `check(..., True)` skip sentinel — an
// assertion that asserts nothing. It is not ported; t.Skip says the same thing
// honestly.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// requireGitForTest skips when git is absent or older than the cone-mode floor,
// so a machine that only ever runs monorepo mode is never blocked (GPF-R-7.7).
func requireGitForTest(t *testing.T) {
	t.Helper()
	if !gitAvailable() {
		t.Skip("git not on PATH")
	}
	v, err := gitVersion()
	if err != nil || v[0] < MinGit[0] || (v[0] == MinGit[0] && v[1] < MinGit[1]) {
		t.Skipf("git %v is older than %v", v, MinGit)
	}
}

// sourceRepo is selftest.py:477-499: a repo carrying allowlisted governance
// content, a NESTED knowledge path, and two files the allowlist must not pull.
func sourceRepo(t *testing.T, dir string) (path, sha string) {
	t.Helper()
	src := filepath.Join(dir, "src")
	write(t, filepath.Join(src, "governance", "requirements.yaml"), "r: 1\n")
	write(t, filepath.Join(src, "components", "svc.yaml"), "id: svc\n")
	write(t, filepath.Join(src, "README.md"), "NOT governance\n")
	// A nested allowlist entry — the knowledge-catalog shape. Depth-1 entries
	// like governance/ never exercise the parent-permission path a re-sync needs.
	write(t, filepath.Join(src, "docs", "sdd", "spec.md"), "# Spec\n")
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", src}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "init", "-q", src).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	git("config", "user.email", "a@b.c")
	git("config", "user.name", "t")
	git("add", "-A")
	git("commit", "-q", "-m", "c1")
	return src, git("rev-parse", "HEAD")
}

// syncFixture builds one repo with TWO destinations — a governance slice under
// platforms/ and a knowledge slice under knowledge/ — and syncs it once.
func syncFixture(t *testing.T, paths ...string) (*workspace.Workspace, *Manifest, string, *SyncResult) {
	t.Helper()
	requireGitForTest(t)
	dir := tempRoot(t)
	src, sha := sourceRepo(t, dir)
	govPaths := "[governance/, components/]"
	if len(paths) == 1 {
		govPaths = paths[0]
	}
	wsDir := filepath.Join(dir, "ws")
	write(t, filepath.Join(wsDir, workspace.ManifestName), fmt.Sprintf(
		"version: 1\nrepos:\n  - name: p\n    url: file://%s\n"+
			"    pin: {commit: %s}\n    slices:\n"+
			"      - {paths: %s, localDirectory: platforms/p}\n"+
			"      - {paths: [docs/sdd], localDirectory: knowledge/p}\n",
		src, sha, govPaths))
	ws := wsAt(wsDir)
	m, err := LoadManifest(ws)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Sync(ws, m, SyncOptions{})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	t.Cleanup(func() { unfreeze(wsDir) })
	return ws, m, sha, res
}

// unfreeze restores write across a materialized workspace so t.TempDir's
// cleanup can remove it. Without it every test that syncs leaks a 0555 tree.
func unfreeze(root string) {
	rels := rglob(root)
	for i := len(rels) - 1; i >= 0; i-- {
		chmodWritable(filepath.Join(root, rels[i]))
	}
	chmodWritable(root)
}

func relFiles(t *testing.T, root, under string) []string {
	t.Helper()
	var out []string
	for _, rel := range rglob(filepath.Join(root, under)) {
		p := filepath.Join(root, under, rel)
		if fi, err := os.Lstat(p); err == nil && fi.Mode().IsRegular() {
			out = append(out, under+"/"+rel)
		}
	}
	return out
}

// ST-077 / ST-078: the sparse fetch plus the copy-time allowlist together
// materialize ONLY governance content. README.md and any non-allowlisted
// subtree stay out.
func TestWorkspaceSync_MaterializesOnlyAllowlistedPaths(t *testing.T) {
	ws, _, _, res := syncFixture(t, "[governance/]")
	got := relFiles(t, ws.Root, "platforms/p")
	if len(got) != 1 || got[0] != "platforms/p/governance/requirements.yaml" {
		t.Fatalf("materialized %v", got)
	}
	for _, f := range got {
		if strings.Contains(f, "README") {
			t.Fatalf("non-allowlisted file materialized: %s", f)
		}
	}
	if len(res.Repos) != 1 || res.Repos[0].Files != 2 {
		t.Fatalf("sync record = %+v; want 1 repo, 2 files across both slices", res.Repos)
	}
}

// ST-079 / ST-080: the second slice lands in its own target with the source
// subpath preserved, and the two destinations do not cross-contaminate.
func TestWorkspaceSync_SecondSliceLandsInOwnTargetIsolated(t *testing.T) {
	ws, _, _, _ := syncFixture(t)
	kfiles := relFiles(t, ws.Root, "knowledge/p")
	if len(kfiles) != 1 || kfiles[0] != "knowledge/p/docs/sdd/spec.md" {
		t.Fatalf("knowledge slice = %v", kfiles)
	}
	for _, f := range relFiles(t, ws.Root, "platforms/p") {
		if strings.Contains(f, "docs") {
			t.Fatalf("knowledge content leaked into the governance slice: %s", f)
		}
	}
	if exists(filepath.Join(ws.Root, "knowledge/p/governance")) {
		t.Fatal("governance content leaked into the knowledge slice")
	}
}

// ST-081: one repo means one clone and one cache directory, however many
// destinations it feeds.
func TestWorkspaceSync_OneCacheDirPerRepo(t *testing.T) {
	ws, _, _, _ := syncFixture(t)
	entries, err := os.ReadDir(filepath.Join(ws.Root, workspace.FederationCache))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "p" {
		t.Fatalf("cache dirs = %v", entries)
	}
}

// ST-082 / ST-083: ONE flat lock entry, recording the resolved SHA, both slices,
// and a files map that is the UNION across them.
func TestWorkspaceSync_LockRecordsSHAAndUnionFileHashes(t *testing.T) {
	ws, _, sha, _ := syncFixture(t)
	lock, err := LoadLock(ws)
	if err != nil || !lock.Usable {
		t.Fatalf("LoadLock = %v, %v", lock, err)
	}
	if len(lock.Repos) != 1 {
		t.Fatalf("lock repos = %d, want 1", len(lock.Repos))
	}
	lr := lock.Repos[0]
	if got := yamlio.PyString(lr.Get("resolvedCommit")); got != sha {
		t.Fatalf("resolvedCommit = %q, want %q", got, sha)
	}
	if sl, _ := lr.Get("slices").(yamlio.PySeq); len(sl) != 2 {
		t.Fatalf("slices recorded = %d, want 2", len(sl))
	}
	var keys []string
	for _, p := range LockFiles(lr) {
		keys = append(keys, p.K)
	}
	want := []string{
		"platforms/p/governance/requirements.yaml",
		"platforms/p/components/svc.yaml",
		"knowledge/p/docs/sdd/spec.md",
	}
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Fatalf("lock files = %v, want %v", keys, want)
	}
	// And the hashes are the real content hashes, keyed workspace-relative.
	for _, p := range LockFiles(lr) {
		body, err := os.ReadFile(filepath.Join(ws.Root, p.K))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(body)
		if hex.EncodeToString(sum[:]) != yamlio.PyString(p.V) {
			t.Fatalf("%s: recorded hash is not the file's", p.K)
		}
	}
}

// TRAP 2, write side. The lock's files: map is INSERTION order — manifest slice
// order, then that slice's paths: order, then the walk. It is the reverse of
// alphabetical for the committed fixtures, so a sort here would look tidy and
// break a committed golden. Swapping the two entries in `paths:` must swap the
// two blocks in the lock.
func TestWorkspaceSync_LockFilesOrderFollowsPathsListNotSort(t *testing.T) {
	keys := func(pathsList string) []string {
		ws, _, _, _ := syncFixture(t, pathsList)
		lock, err := LoadLock(ws)
		if err != nil || !lock.Usable {
			t.Fatalf("LoadLock = %v", err)
		}
		var out []string
		for _, p := range LockFiles(lock.Repos[0]) {
			if strings.HasPrefix(p.K, "platforms/") {
				out = append(out, p.K)
			}
		}
		return out
	}
	const gov = "platforms/p/governance/requirements.yaml"
	const comp = "platforms/p/components/svc.yaml"

	got := keys("[governance/, components/]")
	if len(got) != 2 || got[0] != gov || got[1] != comp {
		t.Fatalf("lock order = %v, want %v then %v — this is the committed "+
			"fixture's order and it is NOT sorted", got, gov, comp)
	}
	got = keys("[components/, governance/]")
	if len(got) != 2 || got[0] != comp || got[1] != gov {
		t.Fatalf("lock order = %v, want %v then %v — reordering paths: did not "+
			"reorder the lock, so something sorted", got, comp, gov)
	}
}

// Every materialized path is frozen: files 0444, directories 0555. The
// differential harness compares mode bits, so this is the local proof of what
// the harness proves end to end.
func TestWorkspaceSync_MaterializedTreeIsReadOnly(t *testing.T) {
	ws, _, _, _ := syncFixture(t)
	for _, base := range []string{"platforms/p", "knowledge/p"} {
		root := filepath.Join(ws.Root, base)
		for _, rel := range rglob(root) {
			fi, err := os.Lstat(filepath.Join(root, rel))
			if err != nil {
				t.Fatal(err)
			}
			want := os.FileMode(modeFileReadonly)
			if fi.IsDir() {
				want = modeDirReadonly
			}
			if got := fi.Mode().Perm(); got != want {
				t.Fatalf("%s/%s mode = %04o, want %04o", base, rel, got, want)
			}
		}
	}
}

// ST-084: re-syncing over an already-frozen tree succeeds. This is the
// regression guard for the nested allowlist entry (docs/sdd) whose parent the
// first run left 0555 — removal needs write on the PARENT, which chmod-ing only
// the entry does not supply.
func TestWorkspaceSync_ResyncOverReadOnlyNestedSlice(t *testing.T) {
	ws, m, _, _ := syncFixture(t)
	if _, err := Sync(ws, m, SyncOptions{}); err != nil {
		t.Fatalf("re-sync over a frozen tree failed: %v", err)
	}
}

// ST-085 / ST-086: --frozen materializes from the lock with the source repo
// gone, and reproduces the online tree bit for bit.
func TestWorkspaceSync_FrozenReproducesOnlineTreeOffline(t *testing.T) {
	ws, m, _, _ := syncFixture(t)
	before := treeHash(t, ws.Root)

	// Wipe both slices and make the source unreachable.
	for _, base := range []string{"platforms", "knowledge"} {
		p := filepath.Join(ws.Root, base)
		unfreeze(p)
		if err := os.RemoveAll(p); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(filepath.Dir(ws.Root), "src")
	if err := os.Rename(src, src+".gone"); err != nil {
		t.Fatal(err)
	}

	res, err := Sync(ws, m, SyncOptions{Frozen: true})
	if err != nil {
		t.Fatalf("--frozen sync failed: %v", err)
	}
	if len(res.Repos) != 1 || !res.Repos[0].Restored || res.Wrote {
		t.Fatalf("frozen result = %+v; must restore and must not rewrite the lock", res)
	}
	if got := treeHash(t, ws.Root); got != before {
		t.Fatalf("frozen tree %s != online tree %s", got, before)
	}
}

// --frozen with no local cache is an external-tool refusal, not a silent
// partial materialization: the guarantee is "from the lock, offline", and
// without the object cache neither half is available.
func TestWorkspaceSync_FrozenWithoutCacheRefuses(t *testing.T) {
	ws, m, _, _ := syncFixture(t)
	cache := filepath.Join(ws.Root, workspace.FederationCache)
	unfreeze(cache)
	if err := os.RemoveAll(cache); err != nil {
		t.Fatal(err)
	}
	res, err := Sync(ws, m, SyncOptions{Frozen: true})
	wantCode(t, err, model.ExitExternalTool)
	if !strings.Contains(err.Error(), "no local cache for 'p' at "+workspace.FederationCache+"/p") {
		t.Fatalf("message = %q", err)
	}
	// The header had already printed by the time it failed, so the record set
	// must survive the error — and it must NOT claim completion.
	if res == nil || res.Complete || len(res.Repos) != 0 {
		t.Fatalf("partial result = %+v", res)
	}
}

// --only scopes the run but must not drop the other repos' lock entries.
func TestWorkspaceSync_OnlyUnknownRepoIsAWorkspaceError(t *testing.T) {
	ws, m, _, _ := syncFixture(t)
	res, err := Sync(ws, m, SyncOptions{Only: "nope"})
	wantCode(t, err, model.ExitWorkspace)
	if res != nil {
		t.Fatalf("result = %+v; --only fails before the header line", res)
	}
}

// The cache is machine-owned and must never land in the workspace repo.
func TestWorkspaceSync_GitignoresTheCache(t *testing.T) {
	ws, m, _, _ := syncFixture(t)
	raw, err := os.ReadFile(filepath.Join(ws.Root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), ".company-os/\n") {
		t.Fatalf(".gitignore = %q", raw)
	}
	// Idempotent: a second sync must not append a second stanza.
	if _, err := Sync(ws, m, SyncOptions{}); err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(ws.Root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(raw) {
		t.Fatalf(".gitignore grew on re-sync:\n%q\n%q", raw, again)
	}
}

// Status over a freshly synced workspace reports every target and calls it
// clean, which is the only thing that exercises the multi-target label.
func TestWorkspaceStatus_ReportsEveryTargetClean(t *testing.T) {
	ws, m, sha, _ := syncFixture(t)
	res, err := Status(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Repos) != 1 {
		t.Fatalf("status repos = %d", len(res.Repos))
	}
	r := res.Repos[0]
	if !r.Locked || r.PinDrift || r.SliceDrift || r.State != StateClean {
		t.Fatalf("status = %+v", r)
	}
	if strings.Join(r.Targets, ", ") != "platforms/p, knowledge/p" {
		t.Fatalf("targets = %v", r.Targets)
	}
	if r.SHA != sha || res.ActionNeeded {
		t.Fatalf("status = %+v (sha %s)", res, sha)
	}
}

// A hand-edited slice flips status to drifted and gate 8 to [FAIL] — the same
// hash comparison reached two ways.
func TestWorkspaceStatus_DriftedAfterHandEdit(t *testing.T) {
	ws, m, _, _ := syncFixture(t)
	victim := filepath.Join(ws.Root, "platforms/p/governance/requirements.yaml")
	chmodWritable(victim)
	write(t, victim, "r: 2\n")

	res, err := Status(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	if res.Repos[0].State != StateDrifted || !res.ActionNeeded {
		t.Fatalf("status = %+v", res.Repos[0])
	}
	g, err := Gate(ws, m, 8)
	if err != nil {
		t.Fatal(err)
	}
	if g.Ordinal != 8 || g.Slug != GateSlug || len(g.Findings) != 1 ||
		g.Findings[0].Code != model.CodeSliceHandEdited {
		t.Fatalf("gate = %+v", g)
	}
}

// treeHash digests the materialized slices' paths and contents, which is what
// "bit-reproducible" means for --frozen.
func treeHash(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	for _, base := range []string{"platforms", "knowledge"} {
		dir := filepath.Join(root, base)
		for _, rel := range rglob(dir) {
			p := filepath.Join(dir, rel)
			fi, err := os.Lstat(p)
			if err != nil || !fi.Mode().IsRegular() {
				continue
			}
			body, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(body)
			fmt.Fprintf(h, "%s/%s%s\n", base, rel, hex.EncodeToString(sum[:]))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
