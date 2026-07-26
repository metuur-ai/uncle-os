package federation

// The git-free half of the 38 federation assertions inherited from
// examples/selftest.py (ST-049 .. ST-075), plus the ordering proofs the port
// needed that Python never had to state.
//
// Every assertion that Python writes as `_dies(lambda: ...)` becomes "returns a
// non-nil error carrying the right exit code" here — the LLD's first structural
// ruling, since nothing below cmd/ may exit.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// ---------------------------------------------------------------- helpers

// pv lifts a Go literal into the PyValue the loader would have produced.
func pv(v any) yamlio.PyValue {
	switch t := v.(type) {
	case nil:
		return yamlio.PyNull{}
	case string:
		return yamlio.PyStr(t)
	case []string:
		out := yamlio.PySeq{}
		for _, s := range t {
			out = append(out, yamlio.PyStr(s))
		}
		return out
	case yamlio.PyValue:
		return t
	}
	panic("pv: unsupported literal")
}

// pmap builds a mapping from alternating key/value literals, preserving order.
func pmap(kv ...any) yamlio.PyMap {
	out := yamlio.PyMap{}
	for i := 0; i+1 < len(kv); i += 2 {
		out = append(out, yamlio.PyPair{K: kv[i].(string), V: pv(kv[i+1])})
	}
	return out
}

func pseq(items ...any) yamlio.PySeq {
	out := yamlio.PySeq{}
	for _, it := range items {
		out = append(out, pv(it))
	}
	return out
}

// wsAt builds a Workspace rooted at dir.
func wsAt(dir string) *workspace.Workspace { return workspace.New(dir) }

// tempRoot is tempRoot(t) put through the same symlink resolution
// workspace.New applies. On darwin tempRoot(t) hands back a /var/... path whose
// realpath is /private/var/..., so a fixture built at the raw path and read
// through ws.Root would compute workspace-relative keys full of "../.." — which
// is a test artifact, not a behaviour, and every assertion here would have to
// repeat it.
func tempRoot(t *testing.T) string {
	t.Helper()
	return workspace.New(t.TempDir()).Root
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

// wantCode asserts an error is present and carries the expected exit code, which
// is the half of "it dies" that Python's SystemExit could not express.
func wantCode(t *testing.T, err error, code model.ExitCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("want an error with code %d, got nil", code)
	}
	if got := model.CodeOf(err); got != code {
		t.Fatalf("exit code = %d, want %d (err: %v)", got, code, err)
	}
}

// ------------------------------------------- pin validation (GPF-R-6.3), 6

// ST-049 / ST-050: exactly one of commit:/tag: is accepted and echoed back.
func TestRepoPin_AcceptsCommit(t *testing.T) {
	got, err := RepoPin(pmap("name", "r", "pin", pmap("commit", "abc123")))
	if err != nil {
		t.Fatal(err)
	}
	if got != (Pin{Kind: "commit", Ref: "abc123"}) {
		t.Fatalf("pin = %+v", got)
	}
}

func TestRepoPin_AcceptsTag(t *testing.T) {
	got, err := RepoPin(pmap("name", "r", "pin", pmap("tag", "v1.0.0")))
	if err != nil {
		t.Fatal(err)
	}
	if got != (Pin{Kind: "tag", Ref: "v1.0.0"}) {
		t.Fatalf("pin = %+v", got)
	}
}

// ST-051 / ST-052: a floating ref is not reproducible and is refused. The
// message reprs the offending keys as a Python list, which is what the
// differential harness compares byte-for-byte.
func TestRepoPin_RejectsBranch(t *testing.T) {
	_, err := RepoPin(pmap("name", "r", "pin", pmap("branch", "main")))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "pin key(s) ['branch'] are floating") {
		t.Fatalf("message = %q", err)
	}
}

func TestRepoPin_RejectsBareRef(t *testing.T) {
	_, err := RepoPin(pmap("name", "r", "pin", pmap("ref", "HEAD")))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "['ref']") {
		t.Fatalf("message = %q", err)
	}
}

// ST-053: both set is ambiguous, not permissive.
func TestRepoPin_RejectsBothCommitAndTag(t *testing.T) {
	_, err := RepoPin(pmap("name", "r", "pin", pmap("commit", "a", "tag", "v1")))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "(got ['commit', 'tag'])") {
		t.Fatalf("message = %q", err)
	}
}

// ST-054: an empty pin names 'neither' — a bare str, so no quotes.
func TestRepoPin_RejectsEmptyPin(t *testing.T) {
	_, err := RepoPin(pmap("name", "r", "pin", yamlio.PyMap{}))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "(got neither)") {
		t.Fatalf("message = %q", err)
	}
}

// ------------------------------------------ multi-slice normalization, 7

// ST-055: a bare localDirectory:/paths: pair still normalizes to one slice —
// backward compatibility with every single-slice manifest ever written.
func TestRepoSlices_BarePairNormalizesToOne(t *testing.T) {
	got, err := RepoSlices(pmap("name", "r", "url", "file:///x",
		"localDirectory", "knowledge/a", "paths", []string{"docs/sdd"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].LocalDirectory != "knowledge/a" ||
		len(got[0].Paths) != 1 || got[0].Paths[0] != "docs/sdd" {
		t.Fatalf("slices = %+v", got)
	}
}

// ST-056: an omitted paths: falls back to the governance-only default.
func TestRepoSlices_OmittedPathsUseDefault(t *testing.T) {
	got, err := RepoSlices(pmap("name", "r", "localDirectory", "knowledge/a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || strings.Join(got[0].Paths, ",") != strings.Join(DefaultSlicePaths, ",") {
		t.Fatalf("paths = %v, want %v", got[0].Paths, DefaultSlicePaths)
	}
}

// ST-057: one entry per slice, order preserved.
func TestRepoSlices_ListYieldsOnePerSlice(t *testing.T) {
	got, err := RepoSlices(pmap("name", "r", "slices", pseq(
		pmap("paths", []string{"docs/sdd"}, "localDirectory", "knowledge/a"),
		pmap("paths", []string{"arch"}, "localDirectory", "knowledge/b"))))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].LocalDirectory != "knowledge/a" ||
		got[1].LocalDirectory != "knowledge/b" {
		t.Fatalf("slices = %+v", got)
	}
}

// ST-058 / ST-059: mixing the two shapes is rejected rather than silently
// ignoring the top-level keys.
func TestRepoSlices_RejectsSlicesPlusTopLevelLocalDir(t *testing.T) {
	_, err := RepoSlices(pmap("name", "r", "localDirectory", "knowledge/a",
		"slices", pseq(pmap("localDirectory", "knowledge/b"))))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "cannot set both 'slices:' and top-level 'localDirectory:'") {
		t.Fatalf("message = %q", err)
	}
}

func TestRepoSlices_RejectsSlicesPlusTopLevelPaths(t *testing.T) {
	_, err := RepoSlices(pmap("name", "r", "paths", []string{"d"},
		"slices", pseq(pmap("localDirectory", "knowledge/b"))))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "top-level 'paths:'") {
		t.Fatalf("message = %q", err)
	}
}

// The presence test is on the KEY, not on truthiness: `paths: null` alongside
// `slices:` is the same ambiguous shape and must be refused too. Python's
// `if k in repo` says so; a PyFalsy test here would have let it through.
func TestRepoSlices_RejectsSlicesPlusNullPaths(t *testing.T) {
	_, err := RepoSlices(pmap("name", "r", "paths", nil,
		"slices", pseq(pmap("localDirectory", "knowledge/b"))))
	wantCode(t, err, model.ExitArtifact)
}

// ST-060: a repo must name a destination.
func TestRepoSlices_RejectsNoDestination(t *testing.T) {
	_, err := RepoSlices(pmap("name", "r"))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "needs either 'localDirectory:'") {
		t.Fatalf("message = %q", err)
	}
}

// ST-061: the renamed key is named in the message, so the fix is mechanical.
func TestRepoSlices_RejectsRootKeyNamingRename(t *testing.T) {
	_, err := RepoSlices(pmap("name", "r", "slices", pseq(pmap("root", "knowledge/a"))))
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "renamed to 'localDirectory:'") {
		t.Fatalf("message = %q", err)
	}
}

// ------------------------------ hashing and target-path safety, 5 (+ traps)

// ST-062: sliceHash is computed ONCE over the union, so merging the same files
// in either order gives the same digest — reordering `slices:` with no content
// change cannot flip it and make --frozen die blaming the object cache.
func TestAggregateHash_MergeOrderIndependent(t *testing.T) {
	f1 := yamlio.NewOrderedMap()
	f1.SetString("b/2.md", "bb")
	f1.SetString("a/1.md", "aa")
	f2 := yamlio.NewOrderedMap()
	f2.SetString("c/3.md", "cc")

	ab := yamlio.NewOrderedMap()
	ab.Update(f1)
	ab.Update(f2)
	ba := yamlio.NewOrderedMap()
	ba.Update(f2)
	ba.Update(f1)

	if strings.Join(ab.Keys(), ",") == strings.Join(ba.Keys(), ",") {
		t.Fatal("the two merges must differ in INSERTION order or this proves nothing")
	}
	if AggregateHash(ab) != AggregateHash(ba) {
		t.Fatalf("aggregate hash depends on merge order: %s != %s",
			AggregateHash(ab), AggregateHash(ba))
	}
}

// TRAP 3. aggregate_hash iterates `sorted(files)` — a plain STRING sort — over
// the very keys the lock emits in walk order. Feeding it the digest of a
// hand-built sorted stream pins that, and would fail the moment someone
// "unified" the two orderings onto Keys().
func TestAggregateHash_IteratesSortedKeysNotInsertionOrder(t *testing.T) {
	files := yamlio.NewOrderedMap()
	files.SetString("z.md", "1")
	files.SetString("a.md", "2")
	if got, want := strings.Join(files.Keys(), ","), "z.md,a.md"; got != want {
		t.Fatalf("insertion order = %q, want %q", got, want)
	}
	if got := AggregateHash(files); got != sha256Of("a.md\x002\nz.md\x001\n") {
		t.Fatalf("aggregate hash = %s; it is not the sorted-key stream", got)
	}
}

// ST-063: knowledge/ is the catalog root and holds a generated CLAUDE.md node.
// Targeting it bare would freeze that node 0444 and make gate 5 unfixable.
func TestSliceRel_RejectsBareKnowledgeRoot(t *testing.T) {
	_, err := SliceRel("r", "knowledge")
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "must name an area under knowledge/") {
		t.Fatalf("message = %q", err)
	}
}

// ST-064: depth >= 2 under knowledge/ is fine and comes back unchanged.
func TestSliceRel_AcceptsKnowledgeArea(t *testing.T) {
	got, err := SliceRel("r", "knowledge/components/x")
	if err != nil || got != "knowledge/components/x" {
		t.Fatalf("SliceRel = %q, %v", got, err)
	}
}

// Depth >= 2 is NOT a blanket rule: company-os/ and company-ontology/ are
// legitimate depth-1 targets whose node ships inside the slice.
func TestSliceRel_AcceptsDepthOneNonKnowledgeRoot(t *testing.T) {
	for _, root := range []string{"company-os", "company-ontology"} {
		if got, err := SliceRel("r", root); err != nil || got != root {
			t.Fatalf("SliceRel(%q) = %q, %v", root, got, err)
		}
	}
}

// ST-065 plus the two escape shapes the same site rejects.
func TestSliceRel_RejectsNonCanonicalRoot(t *testing.T) {
	_, err := SliceRel("r", "elsewhere/x")
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "must land under one of") {
		t.Fatalf("message = %q", err)
	}
}

func TestSliceRel_RejectsEscapes(t *testing.T) {
	for _, bad := range []string{"/etc/passwd", "../outside", "", "platforms/../../x"} {
		if _, err := SliceRel("r", bad); err == nil {
			t.Fatalf("SliceRel(%q) accepted", bad)
		}
	}
	// A component that merely STARTS with dots is not a traversal.
	if _, err := SliceRel("r", "platforms/..hidden"); err != nil {
		t.Fatalf("platforms/..hidden rejected: %v", err)
	}
}

// ST-066: the repo name keys the federation cache directory, so traversal must
// not reach it.
func TestValidateRepoEntry_RejectsUnsafeName(t *testing.T) {
	err := validateRepoEntry(pmap("name", "../evil", "url", "file:///x",
		"localDirectory", "knowledge/a", "pin", pmap("commit", "abc")), 0)
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "must be a plain label") {
		t.Fatalf("message = %q", err)
	}
}

// -------------------------- manifest-level target disjointness, 4

func manifestWith(t *testing.T, body string) (*workspace.Workspace, error) {
	t.Helper()
	dir := tempRoot(t)
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o777); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, workspace.ManifestName), "version: 1\nrepos:\n"+body)
	ws := wsAt(dir)
	_, err := LoadManifest(ws)
	return ws, err
}

// ST-067: two slices of ONE repo nesting one target inside another is rejected —
// the outer slice's read-only freeze captures the inner one's parent, and the
// next re-sync cannot remove it.
func TestLoadManifest_RejectsNestedTargetsInOneRepo(t *testing.T) {
	_, err := manifestWith(t,
		"  - name: a\n    url: file:///x\n    pin: {commit: dead}\n    slices:\n"+
			"      - {paths: [d], localDirectory: knowledge/a}\n"+
			"      - {paths: [e], localDirectory: knowledge/a/b}\n")
	wantCode(t, err, model.ExitArtifact)
	if !strings.Contains(err.Error(), "targets must be disjoint") {
		t.Fatalf("message = %q", err)
	}
}

// ST-068: the cross-repo case needs the global view no per-entry pass has.
func TestLoadManifest_RejectsTwoReposSharingTarget(t *testing.T) {
	_, err := manifestWith(t,
		"  - name: a\n    url: file:///x\n    pin: {commit: dead}\n"+
			"    localDirectory: knowledge/x\n"+
			"  - name: b\n    url: file:///y\n    pin: {commit: beef}\n"+
			"    localDirectory: knowledge/x\n")
	wantCode(t, err, model.ExitArtifact)
}

// ST-069: sibling targets sharing a string prefix are NOT nested. This is the
// assertion that forbids a strings.HasPrefix implementation.
func TestLoadManifest_SiblingPrefixTargetsAccepted(t *testing.T) {
	_, err := manifestWith(t,
		"  - name: a\n    url: file:///x\n    pin: {commit: dead}\n    slices:\n"+
			"      - {paths: [d], localDirectory: platforms/comms}\n"+
			"      - {paths: [e], localDirectory: platforms/comms-v2}\n")
	if err != nil {
		t.Fatalf("sibling prefixes rejected: %v", err)
	}
}

// ST-070: no workspace.yaml => monorepo mode, no behavior change (GPF-R-6.1).
func TestLoadManifest_AbsentManifestIsMonorepoMode(t *testing.T) {
	m, err := LoadManifest(wsAt(tempRoot(t)))
	if err != nil || m != nil {
		t.Fatalf("LoadManifest = %v, %v; want nil, nil", m, err)
	}
}

// An EMPTY workspace.yaml is not absence: it is a manifest that says nothing,
// and the repos: check refuses it. Only the file's existence switches modes.
func TestLoadManifest_EmptyManifestIsNotMonorepoMode(t *testing.T) {
	dir := tempRoot(t)
	write(t, filepath.Join(dir, workspace.ManifestName), "")
	_, err := LoadManifest(wsAt(dir))
	wantCode(t, err, model.ExitArtifact)
}

// -------------------------------- drift detection (GPF-R-7.5), 5

// driftFixture is selftest.py:418-461 — a workspace whose slice file, lock and
// manifest agree, ready to be perturbed one axis at a time.
func driftFixture(t *testing.T) (*workspace.Workspace, string, string) {
	t.Helper()
	dir := tempRoot(t)
	slice := filepath.Join(dir, "platforms", "p", "governance", "requirements.yaml")
	write(t, slice, "schemaVersion: '1.0'\n")
	good := sha256Of("schemaVersion: '1.0'\n")
	write(t, filepath.Join(dir, workspace.ManifestName),
		"version: 1\nrepos:\n  - name: p\n    url: file:///x\n"+
			"    localDirectory: platforms/p\n    pin: {commit: deadbeef}\n"+
			"    paths: [governance/]\n")
	return wsAt(dir), slice, good
}

func writeLock(t *testing.T, ws *workspace.Workspace, body string) {
	t.Helper()
	write(t, LockPath(ws), body)
}

// ST-071: a slice whose bytes hash to the recorded value is clean — and
// slice_state reads ONLY the files: key, which is what keeps --only per-repo and
// gate 8 cheap. The entry below deliberately carries no slices:, url: or pin:.
func TestSliceState_CleanWhenHashesMatch(t *testing.T) {
	ws, _, good := driftFixture(t)
	lr := pmap("name", "p", "resolvedCommit", "deadbeef", "sliceHash", "x",
		"files", pmap("platforms/p/governance/requirements.yaml", good))
	got, err := SliceState(ws, lr)
	if err != nil || got != StateClean {
		t.Fatalf("SliceState = %q, %v", got, err)
	}
}

// ST-074: a hand-edited slice file is drifted, not clean.
func TestSliceState_DriftedOnHandEdit(t *testing.T) {
	ws, slice, good := driftFixture(t)
	lr := pmap("name", "p", "files",
		pmap("platforms/p/governance/requirements.yaml", good))
	write(t, slice, "schemaVersion: '1.0'\n# sneaky\n")
	got, err := SliceState(ws, lr)
	if err != nil || got != StateDrifted {
		t.Fatalf("SliceState = %q, %v", got, err)
	}
}

// A lock entry that records no files at all is 'missing', not vacuously clean.
func TestSliceState_MissingWhenLockRecordsNoFiles(t *testing.T) {
	ws, _, _ := driftFixture(t)
	got, err := SliceState(ws, pmap("name", "p"))
	if err != nil || got != StateMissing {
		t.Fatalf("SliceState = %q, %v", got, err)
	}
}

const inSyncLock = "version: 1\nrepos:\n- name: p\n  resolvedCommit: deadbeef\n" +
	"  sliceHash: x\n  slices:\n  - localDirectory: platforms/p\n    paths:\n    - governance/\n" +
	"  files:\n    platforms/p/governance/requirements.yaml: %s\n"

// ST-072: an in-sync workspace reports zero problems and one file checked.
func TestFederatedSliceProblems_InSyncIsClean(t *testing.T) {
	ws, _, good := driftFixture(t)
	writeLock(t, ws, fmt.Sprintf(inSyncLock, good))
	m, err := LoadManifest(ws)
	if err != nil {
		t.Fatal(err)
	}
	res, err := SliceFindings(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	if res.Files != 1 {
		t.Fatalf("files checked = %d, want 1", res.Files)
	}
	if len(res.Findings) != 1 || res.Findings[0].Severity != model.SevOK ||
		res.Findings[0].Code != model.CodeFederationSlicesMatch {
		t.Fatalf("findings = %+v; want one ok line", res.Findings)
	}
}

// ST-073: the manifest's target moved but the on-disk bytes still hash clean.
// This is the case nothing else catches — the per-file loop reports nothing.
func TestFederatedSliceProblems_SliceSetDriftDetected(t *testing.T) {
	ws, _, good := driftFixture(t)
	writeLock(t, ws, fmt.Sprintf(strings.Replace(inSyncLock,
		"localDirectory: platforms/p", "localDirectory: platforms/other", 1), good))
	m, err := LoadManifest(ws)
	if err != nil {
		t.Fatal(err)
	}
	res, err := SliceFindings(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, f := range res.Findings {
		if f.Code == model.CodeSliceSetDrift {
			found = true
			if f.Fields.Str("repo") != "p" {
				t.Fatalf("repo field = %q", f.Fields.Str("repo"))
			}
			if !strings.Contains(f.Message, "slice set") ||
				!strings.Contains(f.Message, "workspace sync") {
				t.Fatalf("message = %q", f.Message)
			}
		}
	}
	if !found {
		t.Fatalf("slice-set drift not reported: %+v", res.Findings)
	}
}

// ST-075: the hand-edit problem names the offending file and the remedy — and,
// per R-2.12, the path is a typed field rather than only a substring of prose.
func TestFederatedSliceProblems_HandEditNamesPathAndRemedy(t *testing.T) {
	ws, slice, good := driftFixture(t)
	writeLock(t, ws, fmt.Sprintf(inSyncLock, good))
	m, err := LoadManifest(ws)
	if err != nil {
		t.Fatal(err)
	}
	write(t, slice, "schemaVersion: '1.0'\n# sneaky\n")
	res, err := SliceFindings(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 {
		t.Fatalf("findings = %+v, want exactly one", res.Findings)
	}
	f := res.Findings[0]
	if f.Code != model.CodeSliceHandEdited || f.Severity != model.SevFail {
		t.Fatalf("finding = %+v", f)
	}
	const rel = "platforms/p/governance/requirements.yaml"
	if f.Path != rel || f.Fields.Str("path") != rel {
		t.Fatalf("path = %q / field %q", f.Path, f.Fields.Str("path"))
	}
	if !strings.Contains(f.Message, rel) || !strings.Contains(f.Message, "workspace sync") {
		t.Fatalf("message = %q", f.Message)
	}
}

// A missing lock is one problem naming both files, not a crash and not silence.
func TestFederatedSliceProblems_MissingLock(t *testing.T) {
	ws, _, _ := driftFixture(t)
	m, err := LoadManifest(ws)
	if err != nil {
		t.Fatal(err)
	}
	res, err := SliceFindings(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 || res.Findings[0].Code != model.CodeSliceLockMissing {
		t.Fatalf("findings = %+v", res.Findings)
	}
	if res.Files != 0 {
		t.Fatalf("files = %d, want 0", res.Files)
	}
}

// A repo the lock does not cover is reported per repo and does not abort the
// scan of the others.
func TestFederatedSliceProblems_RepoNotLocked(t *testing.T) {
	ws, _, good := driftFixture(t)
	writeLock(t, ws, fmt.Sprintf(strings.Replace(inSyncLock, "- name: p", "- name: other", 1), good))
	m, err := LoadManifest(ws)
	if err != nil {
		t.Fatal(err)
	}
	res, err := SliceFindings(ws, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 || res.Findings[0].Code != model.CodeRepoNotLocked ||
		res.Findings[0].Fields.Str("repo") != "p" {
		t.Fatalf("findings = %+v", res.Findings)
	}
}

// TRAP 2, read side. Gate 8 renders its [FAIL] lines in the LOCK's document
// order, which is the emission order arriving back through the parser and the
// reverse of alphabetical for every committed fixture. Swapping the two files in
// the lock must swap the two findings — if it does not, something sorted.
func TestFederatedSliceProblems_FindingOrderFollowsLockDocumentOrder(t *testing.T) {
	order := func(first, second string) []string {
		dir := tempRoot(t)
		write(t, filepath.Join(dir, workspace.ManifestName),
			"version: 1\nrepos:\n  - name: p\n    url: file:///x\n"+
				"    localDirectory: platforms/p\n    pin: {commit: deadbeef}\n"+
				"    paths: [governance/, components/]\n")
		writeLock(t, wsAt(dir), "version: 1\nrepos:\n- name: p\n"+
			"  slices:\n  - localDirectory: platforms/p\n    paths:\n    - governance/\n    - components/\n"+
			"  files:\n    "+first+": deadbeef\n    "+second+": deadbeef\n")
		ws := wsAt(dir)
		m, err := LoadManifest(ws)
		if err != nil {
			t.Fatal(err)
		}
		res, err := SliceFindings(ws, m)
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, f := range res.Findings {
			paths = append(paths, f.Fields.Str("path"))
		}
		return paths
	}
	const gov = "platforms/p/governance/requirements.yaml"
	const comp = "platforms/p/components/svc.yaml"

	// The committed shape: governance before components, i.e. NOT sorted.
	if got := order(gov, comp); len(got) != 2 || got[0] != gov || got[1] != comp {
		t.Fatalf("findings = %v, want %v then %v", got, gov, comp)
	}
	// Swap the two lines in the lock and the two findings must swap with them.
	if got := order(comp, gov); len(got) != 2 || got[0] != comp || got[1] != gov {
		t.Fatalf("findings = %v, want %v then %v — the order was not read "+
			"from the lock", got, comp, gov)
	}
}

// -------------------------------------------------- the filesystem traps

// TRAP 1. makeReadonly must leave files 0444 and dirs 0555 everywhere below the
// slice root, and must leave the root itself alone. The mode bits are compared
// by the differential harness on every path, so this is a byte-level contract,
// not hygiene.
func TestMakeReadonly_FreezesEveryDescendantAndSparesRoot(t *testing.T) {
	root := filepath.Join(tempRoot(t), "slice")
	write(t, filepath.Join(root, "governance", "requirements.yaml"), "r: 1\n")
	write(t, filepath.Join(root, "docs", "sdd", "spec.md"), "# Spec\n")
	write(t, filepath.Join(root, "top.md"), "x\n")
	rootBefore := statMode(t, root)

	makeReadonly(root)

	for _, rel := range rglob(root) {
		p := filepath.Join(root, rel)
		fi, err := os.Lstat(p)
		if err != nil {
			t.Fatal(err)
		}
		want := os.FileMode(modeFileReadonly)
		if fi.IsDir() {
			want = modeDirReadonly
		}
		if got := fi.Mode().Perm(); got != want {
			t.Fatalf("%s mode = %04o, want %04o", rel, got, want)
		}
	}
	if got := statMode(t, root); got != rootBefore {
		t.Fatalf("root mode changed to %04o (was %04o) — rglob(\"*\") yields "+
			"descendants only", got, rootBefore)
	}
	// Leave the tree removable for t.Cleanup.
	forceRemove(root)
}

// The reverse walk is what makes the freeze survivable: a materializeSlice over
// an already-frozen tree has to restore write across the WHOLE subtree before
// removing anything, because unlink needs write on the PARENT. Depth-1 allowlist
// entries never exercised it; a nested one (docs/sdd) does.
func TestForceRemove_RemovesAFrozenNestedTree(t *testing.T) {
	root := filepath.Join(tempRoot(t), "slice")
	write(t, filepath.Join(root, "docs", "sdd", "spec.md"), "# Spec\n")
	makeReadonly(root)
	forceRemove(root)
	if exists(root) {
		t.Fatal("frozen tree survived forceRemove")
	}
}

// TRAP 4. `sorted(Path.rglob("*"))` is CPython PurePath COMPONENT-WISE order,
// not a byte sort: '/' (0x2F) sorts above '-' and '.', so a directory sorts
// before a sibling file whose name extends its own. Measured under CPython 3.12
// and reproduced here.
func TestRglob_UsesPurePathOrderNotStringOrder(t *testing.T) {
	root := tempRoot(t)
	write(t, filepath.Join(root, "sdd", "adr", "a.md"), "a")
	write(t, filepath.Join(root, "sdd", "adr", "z.md"), "z")
	write(t, filepath.Join(root, "sdd", "adr-x.md"), "x")
	write(t, filepath.Join(root, "sdd", "adr.md"), "m")

	want := []string{"sdd", "sdd/adr", "sdd/adr/a.md", "sdd/adr/z.md",
		"sdd/adr-x.md", "sdd/adr.md"}
	if got := strings.Join(rglob(root), " "); got != strings.Join(want, " ") {
		t.Fatalf("rglob order:\n got %s\nwant %s", got, strings.Join(want, " "))
	}
}

// hashTree keys the map by workspace-relative path, in the order the allowlist
// and the walk produce — and a path reached twice keeps its FIRST position,
// which is what DefaultSlicePaths' `governance/` + `governance/requirements.yaml`
// pair depends on.
func TestHashTree_InsertionOrderFollowsAllowlistThenWalk(t *testing.T) {
	dir := tempRoot(t)
	target := filepath.Join(dir, "platforms", "p")
	write(t, filepath.Join(target, "governance", "requirements.yaml"), "r: 1\n")
	write(t, filepath.Join(target, "components", "svc.yaml"), "id: svc\n")

	files, err := hashTree(wsAt(dir), target,
		[]string{"governance/", "components/", "governance/requirements.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"platforms/p/governance/requirements.yaml",
		"platforms/p/components/svc.yaml",
	}
	if got := files.Keys(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("insertion order = %v, want %v (allowlist order, and a "+
			"re-hashed path must not move to the end)", got, want)
	}
	// And the same key set sorts the other way, which is the second ordering.
	if got := files.SortedKeys(); got[0] != want[1] {
		t.Fatalf("SortedKeys = %v; it must not agree with Keys here", got)
	}
}

// sha256Of is the hex digest of a literal, for pinning both the lock's recorded
// hashes and AggregateHash's byte stream without a fixture file.
func sha256Of(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func statMode(t *testing.T, p string) os.FileMode {
	t.Helper()
	fi, err := os.Lstat(p)
	if err != nil {
		t.Fatal(err)
	}
	return fi.Mode().Perm()
}
