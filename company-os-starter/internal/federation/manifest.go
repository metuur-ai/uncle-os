package federation

// The hand-owned manifest: load, validate, normalize (bin/company-os:2088-2245).
//
// Every check is eager, so a malformed workspace.yaml fails identically for
// `sync`, `status` and `validate` — which is the property that keeps gate 8 from
// having to re-validate anything.

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// DefaultSlicePaths is the governance-only allowlist a repo gets when it names
// no `paths:` (bin/company-os:203-205). Directory entries pull the whole
// subtree; a file entry pulls just that file. `governance/requirements.yaml`
// duplicates content already covered by `governance/`, which is load-bearing
// for one reason only: it proves the lock's files map keeps a path at its FIRST
// insertion position rather than moving it on re-set.
var DefaultSlicePaths = []string{
	"governance/", "components/", "governance/requirements.yaml",
	"reality/", "skills/", "templates/",
}

// repoNameRe is the plain-label constraint on `name:` (bin/company-os:2190).
// The name keys the federation cache directory, so `../evil` must not reach it.
var repoNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// remedy is repo_pin's shared tail (bin/company-os:2201).
const remedy = "use commit:/tag: only, then: company-os workspace sync"

// Manifest is a parsed, fully validated workspace.yaml. Repos keeps the raw
// mappings rather than a struct: repo_pin's error message reprs whatever keys
// the pin actually carried, and `status` interpolates the lock's pin mapping
// verbatim.
type Manifest struct {
	Data  yamlio.PyMap
	Repos []yamlio.PyMap
}

// Slice is one normalized destination: where the content lands and which source
// paths are allowed into it.
type Slice struct {
	LocalDirectory string
	Paths          []string
}

// Pin is the resolved reproducible reference. Exactly one of commit:/tag: —
// a branch or a bare ref is floating and rejected (GPF-R-6.3).
type Pin struct {
	Kind string // "commit" or "tag"
	Ref  string
}

// ManifestPath is the manifest's absolute path.
func ManifestPath(ws *workspace.Workspace) string {
	return path.Join(ws.Root, workspace.ManifestName)
}

// LockPath is the lock's absolute path.
func LockPath(ws *workspace.Workspace) string {
	return path.Join(ws.Root, workspace.LockName)
}

// LoadManifest returns the parsed manifest, or (nil, nil) when absent —
// monorepo mode, in which none of this package runs (GPF-R-6.1).
//
// It is load_manifest (bin/company-os:2092-2135). Validation is eager and in
// the oracle's order: per-entry shape, then duplicate names, then the global
// cross-repo target-disjointness check that no per-entry pass could make.
func LoadManifest(ws *workspace.Workspace) (*Manifest, error) {
	// The existence test comes first and is its own answer: absence is monorepo
	// mode, while an empty file is a manifest that says nothing and dies below.
	// PyLoadFile cannot tell those two apart — both give nil.
	if _, err := os.Stat(ManifestPath(ws)); err != nil {
		return nil, nil
	}
	raw, err := yamlio.PyLoadFile(ManifestPath(ws))
	if err != nil {
		return nil, err
	}
	// `load_yaml(p, {}) or {}` (:2098) is a TRUTHINESS fallback, so an empty
	// document, an empty list and a literal `0` all become {} here.
	data, isMap := raw.(yamlio.PyMap)
	if yamlio.PyFalsy(raw) {
		data, isMap = yamlio.PyMap{}, true
	}
	repos, reposIsList := data.Get("repos").(yamlio.PySeq)
	if !isMap || !reposIsList {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: must be a mapping with a 'repos:' list. "+
				"then: company-os workspace sync", workspace.ManifestName)
	}
	if len(repos) == 0 {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: 'repos:' is empty — remove the file for monorepo "+
				"mode, or add at least one repo. then: company-os workspace sync",
			workspace.ManifestName)
	}

	m := &Manifest{Data: data}
	seen := map[string]bool{}
	// targets is an ordered association list, not a map: the overlap message
	// names the FIRST target already registered, which is dict insertion order.
	type target struct{ rel, owner string }
	var targets []target

	for i, entry := range repos {
		if err := validateRepoEntry(entry, i); err != nil {
			return nil, err
		}
		repo := entry.(yamlio.PyMap)
		nameVal := repo.Get("name")
		// Dedup on repr, not on str(): `name: 1` and `name: '1'` are distinct
		// dict members in Python and both pass the plain-label regex.
		if key := yamlio.PyRepr(nameVal); seen[key] {
			return nil, model.Errorf(model.ExitArtifact,
				"%s: duplicate repo name '%s'", workspace.ManifestName,
				yamlio.PyString(nameVal))
		} else {
			seen[key] = true
		}
		name := yamlio.PyString(nameVal)
		slices, err := RepoSlices(repo)
		if err != nil {
			return nil, err
		}
		for _, s := range slices {
			rel, err := SliceRel(name, s.LocalDirectory)
			if err != nil {
				return nil, err
			}
			parts := pathParts(rel)
			for _, other := range targets {
				op := pathParts(other.rel)
				// Compare on parts, not str.HasPrefix, so 'platforms/comms' and
				// 'platforms/comms-v2' are siblings rather than nested.
				if partsPrefix(parts, op) || partsPrefix(op, parts) {
					return nil, model.Errorf(model.ExitArtifact,
						"%s: slice target '%s' (repo '%s') overlaps '%s' "+
							"(repo '%s') — targets must be disjoint. "+
							"then: company-os workspace sync",
						workspace.ManifestName, rel, name, other.rel, other.owner)
				}
			}
			targets = append(targets, target{rel: rel, owner: name})
		}
		m.Repos = append(m.Repos, repo)
	}
	return m, nil
}

// partsPrefix reports whether a starts with b, component-wise.
func partsPrefix(a, b []string) bool {
	if len(b) > len(a) {
		return false
	}
	for i := range b {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// validateRepoEntry is _validate_repo_entry (bin/company-os:2177-2196). Order
// matters: the shape check has to run before any key is read, and the pin and
// slice checks run last so a missing `name:` is reported as a missing name
// rather than as a nameless pin.
func validateRepoEntry(entry yamlio.PyValue, i int) error {
	repo, ok := entry.(yamlio.PyMap)
	if !ok {
		return model.Errorf(model.ExitArtifact, "%s: repos[%d] must be a mapping",
			workspace.ManifestName, i)
	}
	if repo.Get("root") != nil {
		return model.Errorf(model.ExitArtifact,
			"%s: repos[%d] uses 'root:' — renamed to "+
				"'localDirectory:' (the local directory the slice lands in). "+
				"rename it, then: company-os workspace sync",
			workspace.ManifestName, i)
	}
	for _, f := range []string{"name", "url"} {
		if yamlio.PyFalsy(repo.Get(f)) {
			return model.Errorf(model.ExitArtifact,
				"%s: repos[%d] missing required '%s:'", workspace.ManifestName, i, f)
		}
	}
	name := yamlio.PyString(repo.Get("name"))
	if !repoNameRe.MatchString(name) {
		return model.Errorf(model.ExitArtifact,
			"%s: repos[%d] name '%s' must be a plain label "+
				"([A-Za-z0-9._-], not starting with '.') — it keys the git cache "+
				"directory", workspace.ManifestName, i, name)
	}
	if _, err := RepoPin(repo); err != nil { // raises on floating / ambiguous pins
		return err
	}
	if _, err := RepoSlices(repo); err != nil { // raises on a malformed shape
		return err
	}
	return nil
}

// RepoPin is repo_pin (bin/company-os:2197-2216): exactly one of commit:/tag:.
// A branch name or a bare `ref:` is non-reproducible and rejected (GPF-R-6.3).
func RepoPin(repo yamlio.PyMap) (Pin, error) {
	name := "?"
	if v := repo.Get("name"); v != nil {
		name = yamlio.PyString(v)
	}
	pin, ok := repo.Get("pin").(yamlio.PyMap)
	if !ok {
		return Pin{}, model.Errorf(model.ExitArtifact,
			"repo '%s': 'pin:' must be a mapping with exactly one of "+
				"commit:/tag: — floating refs (branch/ref) are rejected (GPF-R-6.3). "+
				"%s", name, remedy)
	}
	var floating []string
	for _, p := range pin {
		if p.K != "commit" && p.K != "tag" {
			floating = append(floating, p.K)
		}
	}
	if len(floating) > 0 {
		return Pin{}, model.Errorf(model.ExitArtifact,
			"repo '%s': pin key(s) %s are floating and "+
				"non-reproducible; %s (GPF-R-6.3)",
			name, yamlio.PyStrings(floating), remedy)
	}
	var present []string
	for _, k := range []string{"commit", "tag"} {
		if !yamlio.PyFalsy(pin.Get(k)) {
			present = append(present, k)
		}
	}
	if len(present) != 1 {
		got := yamlio.PyStrings(present)
		if len(present) == 0 { // `present or 'neither'` — a str, so no quotes
			got = "neither"
		}
		return Pin{}, model.Errorf(model.ExitArtifact,
			"repo '%s': pin must set EXACTLY ONE of commit:/tag: "+
				"(got %s). %s", name, got, remedy)
	}
	return Pin{Kind: present[0], Ref: yamlio.PyString(pin.Get(present[0]))}, nil
}

// RepoSlices is _repo_slices (bin/company-os:2142-2175): the normalized
// destination list for one repo entry.
//
// A repo declares EITHER a top-level localDirectory:/paths: pair (one slice) OR
// a slices: list (N slices sharing one clone and one cache). Mixing them is
// rejected rather than silently ignoring the top-level keys.
func RepoSlices(repo yamlio.PyMap) ([]Slice, error) {
	name := "?"
	if v := repo.Get("name"); v != nil {
		name = yamlio.PyString(v)
	}
	raw := repo.Get("slices")
	if raw == nil || isNull(raw) {
		if yamlio.PyFalsy(repo.Get("localDirectory")) {
			return nil, model.Errorf(model.ExitArtifact,
				"repo '%s': needs either 'localDirectory:' (one slice) or "+
					"'slices:' (several). then: company-os workspace sync", name)
		}
		paths, err := slicePaths(name, repo.Get("paths"), "paths")
		if err != nil {
			return nil, err
		}
		return []Slice{{
			LocalDirectory: yamlio.PyString(repo.Get("localDirectory")),
			Paths:          paths,
		}}, nil
	}
	// Key PRESENCE, not truthiness: `paths: null` alongside `slices:` is still
	// the ambiguous shape this rejects.
	for _, k := range []string{"localDirectory", "paths"} {
		if repo.Get(k) != nil {
			return nil, model.Errorf(model.ExitArtifact,
				"repo '%s': cannot set both 'slices:' and top-level '%s:' "+
					"— move it into a slices entry, then: company-os workspace sync",
				name, k)
		}
	}
	list, ok := raw.(yamlio.PySeq)
	if !ok || len(list) == 0 {
		return nil, model.Errorf(model.ExitArtifact,
			"repo '%s': 'slices:' must be a non-empty list of "+
				"{paths, localDirectory} entries", name)
	}
	out := make([]Slice, 0, len(list))
	for j, entry := range list {
		s, ok := entry.(yamlio.PyMap)
		if !ok {
			return nil, model.Errorf(model.ExitArtifact,
				"repo '%s': slices[%d] must be a mapping", name, j)
		}
		if s.Get("root") != nil {
			return nil, model.Errorf(model.ExitArtifact,
				"repo '%s': slices[%d] uses 'root:' — renamed to "+
					"'localDirectory:'. rename it, then: company-os workspace sync",
				name, j)
		}
		if yamlio.PyFalsy(s.Get("localDirectory")) {
			return nil, model.Errorf(model.ExitArtifact,
				"repo '%s': slices[%d] missing required 'localDirectory:'", name, j)
		}
		paths, err := slicePaths(name, s.Get("paths"), fmt.Sprintf("slices[%d].paths", j))
		if err != nil {
			return nil, err
		}
		out = append(out, Slice{
			LocalDirectory: yamlio.PyString(s.Get("localDirectory")),
			Paths:          paths,
		})
	}
	return out, nil
}

// slicePaths is _slice_paths (bin/company-os:2137-2140). An absent or null
// `paths:` falls back to the governance-only default; anything else must be a
// non-empty list.
func slicePaths(name string, v yamlio.PyValue, where string) ([]string, error) {
	if v == nil || isNull(v) {
		return append([]string(nil), DefaultSlicePaths...), nil
	}
	list, ok := v.(yamlio.PySeq)
	if !ok || len(list) == 0 {
		return nil, model.Errorf(model.ExitArtifact,
			"repo '%s': '%s' must be a non-empty allowlist", name, where)
	}
	out := make([]string, 0, len(list))
	for _, p := range list {
		out = append(out, yamlio.PyString(p))
	}
	return out, nil
}

// SliceRel is slice_rel (bin/company-os:2219-2241): the workspace-relative
// POSIX path a slice target lands at, or an error. It refuses paths that escape
// the workspace or land outside a canonical root.
func SliceRel(name, localDirectory string) (string, error) {
	raw := localDirectory
	rel := strings.Trim(raw, "/")
	if strings.HasPrefix(raw, "/") || hasDotDot(rel) || rel == "" {
		return "", model.Errorf(model.ExitArtifact,
			"repo '%s': localDirectory '%s' must be a relative "+
				"path inside the workspace (no absolute paths, no '..').", name, raw)
	}
	parts := pathParts(rel)
	if !isCanonicalRoot(parts[0]) {
		return "", model.Errorf(model.ExitArtifact,
			"repo '%s': localDirectory '%s' must land under one "+
				"of %s/", name, raw, strings.Join(workspace.CanonicalRoots, "/, "))
	}
	// `knowledge/` itself is CLI-owned: it holds the generated CLAUDE.md node.
	// Targeting it bare would chmod that node 0444 via makeReadonly, after which
	// graph build cannot rewrite it and gate 5 reports unfixable drift. Depth >= 2
	// is NOT a blanket rule — `company-os` and `company-ontology` are legitimate
	// depth-1 targets whose node ships inside the slice.
	if parts[0] == workspace.KnowledgeRoot && len(parts) < 2 {
		return "", model.Errorf(model.ExitArtifact,
			"repo '%s': localDirectory '%s' must name an area under "+
				"%s/ (e.g. %s/components/%s), not the catalog root itself",
			name, raw, workspace.KnowledgeRoot, workspace.KnowledgeRoot, name)
	}
	return rel, nil
}

// SliceTargetRoot is the absolute path a slice materializes into.
func SliceTargetRoot(ws *workspace.Workspace, name, localDirectory string) (string, error) {
	rel, err := SliceRel(name, localDirectory)
	if err != nil {
		return "", err
	}
	return path.Join(ws.Root, rel), nil
}

// SliceKey is slice_key (bin/company-os:2439-2444): an order-independent
// identity for a slice set, used for manifest-vs-lock drift.
//
// Python returns a sorted list of (localDirectory, tuple(sorted(paths))) and
// compares those lists. Only equality is ever asked, so any injective canonical
// encoding answers the same question; \x00 and \x01 cannot occur in a path.
func SliceKey(slices []Slice) string {
	keys := make([]string, 0, len(slices))
	for _, s := range slices {
		paths := make([]string, 0, len(s.Paths))
		for _, p := range s.Paths {
			paths = append(paths, strings.Trim(p, "/"))
		}
		sort.Strings(paths)
		keys = append(keys, strings.Trim(s.LocalDirectory, "/")+"\x01"+
			strings.Join(paths, "\x00"))
	}
	sort.Strings(keys)
	return strings.Join(keys, "\x02")
}

// SliceKeyOf is SliceKey over the raw `slices:` value of a LOCK entry, where the
// entries are whatever was recorded rather than a validated Slice. Python reads
// them with `.get("localDirectory", "")` and `.get("paths") or []`, so a lock
// written before slice recording compares as the empty set rather than failing.
func SliceKeyOf(v yamlio.PyValue) string {
	list, _ := v.(yamlio.PySeq)
	slices := make([]Slice, 0, len(list))
	for _, e := range list {
		m, ok := e.(yamlio.PyMap)
		if !ok {
			continue
		}
		s := Slice{}
		if ld := m.Get("localDirectory"); ld != nil {
			s.LocalDirectory = yamlio.PyString(ld)
		}
		if ps, ok := m.Get("paths").(yamlio.PySeq); ok {
			for _, p := range ps {
				s.Paths = append(s.Paths, yamlio.PyString(p))
			}
		}
		slices = append(slices, s)
	}
	return SliceKey(slices)
}

func isCanonicalRoot(s string) bool {
	for _, r := range workspace.CanonicalRoots {
		if r == s {
			return true
		}
	}
	return false
}

func isNull(v yamlio.PyValue) bool {
	_, ok := v.(yamlio.PyNull)
	return ok
}

// pathParts splits a POSIX path the way PurePosixPath.parts does for the
// relative paths this package handles: empty components are dropped.
func pathParts(p string) []string {
	var out []string
	for _, c := range strings.Split(p, "/") {
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}

// hasDotDot is `".." in Path(rel).parts` — a component-wise test, so a file
// named `..hidden` is not mistaken for a traversal.
func hasDotDot(rel string) bool {
	for _, c := range pathParts(rel) {
		if c == ".." {
			return true
		}
	}
	return false
}
