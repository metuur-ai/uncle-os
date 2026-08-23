package graph_test

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// fixtures are every committed workspace `graph build` is expected to handle.
// They are not interchangeable: `failing-workspace` is the only one that
// exercises the WRITE path (a drifted `tags:`, a stale feature-index, a node
// with hand-edited prose inside the markers and one with no markers at all),
// and the banking fixtures are the only multi-root ones.
var fixtures = []string{
	"workspace", "standalone-team", "federated",
	filepath.Join("banking", "small-company"),
	"failing-workspace", "failing-federated", "failing-federated-nolock",
}

// idempotentFixtures are the two workspaces acceptance.sh §4 shasums. They are
// the ones whose committed state is asserted to be fully derived AND stable
// under a second build (R-0.6); the rest are not, and deliberately so — the
// cause is in rewrite_generated_block and is worth naming, because it looks
// like a port defect and is not. The CREATE and APPEND branches end their write
// with "\n", while the REPLACE branch splices in `text[ends[0].end():]` and
// END_RE's trailing `\s*$` has already eaten that newline. So a node that gets
// its markers from the append branch is rewritten once more on the next build,
// losing the final newline, and only then settles. `failing-workspace` and
// `failing-federated` each ship a CLAUDE.md with no markers, so both take that
// path; the fixtures acceptance.sh §4 covers ship theirs already marked and
// never do.
var idempotentFixtures = []string{"workspace", "standalone-team"}

// TestBuildIsIdempotentOnCommittedFixtures is R-0.6, the requirement the whole
// package is shaped around: the committed state is already fully derived, and a
// second build changes nothing.
func TestBuildIsIdempotentOnCommittedFixtures(t *testing.T) {
	for _, name := range idempotentFixtures {
		t.Run(name, func(t *testing.T) {
			ws := copyFixture(t, name)
			s0 := treeHash(t, ws.Root)
			if _, err := graph.Build(ws); err != nil {
				t.Fatalf("first build: %v", err)
			}
			s1 := treeHash(t, ws.Root)
			if _, err := graph.Build(ws); err != nil {
				t.Fatalf("second build: %v", err)
			}
			s2 := treeHash(t, ws.Root)
			if s0 != s1 {
				t.Fatalf("committed state was not fully derived:\n%s", diffTrees(s0, s1))
			}
			if s1 != s2 {
				t.Fatalf("graph build is not idempotent:\n%s", diffTrees(s1, s2))
			}
		})
	}
}

// TestBuildConvergesLikePython and TestBuildMatchesPythonBinary used to sit
// here. Both drove company-os-starter/bin/company-os over the same fixtures and
// compared stdout and the resulting file tree byte for byte. R-9.3 deleted that
// binary, so both could only skip — and pythonCLI() said so in its own skip
// message: "the Python reference is gone; this oracle retires with it".
//
// They were removed because their oracle is gone and cannot come back. See task
// 6.10 in docs/tasks/go-cli-tui-port.md for the measurement behind accepting the
// resulting gap: `graph build` has no end-to-end byte-level coverage, so a
// change to what it PRINTS is not caught here.
//
// What it WRITES is still pinned, and that is the half that matters most:
// TestBuildIsIdempotentOnCommittedFixtures above hashes the whole tree before
// and after two builds over the two fixtures that must be fixed points, and
// acceptance.sh §4 shasums the same property from the outside. The three-pass
// convergence the deleted test pinned over the non-idempotent fixtures is NOT
// covered by anything.

// TestWriteFeatureIndexesGuardIsSemantic is task 2.4 stated as a test: an index
// whose BYTES differ from a fresh render but whose STRUCTURE does not must not
// be rewritten. A byte guard rewrites it, `graph build; graph build` stops
// being a no-op, and acceptance.sh §4 fails against Python-emitted bytes.
func TestWriteFeatureIndexesGuardIsSemantic(t *testing.T) {
	ws := copyFixture(t, "workspace")
	idx := filepath.Join(ws.Root, "platforms", "communications",
		"generated", "feature-index.yaml")
	original, err := os.ReadFile(idx)
	if err != nil {
		t.Fatal(err)
	}
	// Re-lay the very same document: keys reversed at the top level and the
	// block sequences turned into flow style. Nothing about the structure
	// changes; every byte does.
	reshaped, err := yamlio.PyDumpAutoFlow(reverseTop(t, original))
	if err != nil {
		t.Fatal(err)
	}
	if reshaped == string(original) {
		t.Fatal("reshaping produced identical bytes; the test proves nothing")
	}
	if err := os.WriteFile(idx, []byte(reshaped), 0o666); err != nil {
		t.Fatal(err)
	}
	written, err := graph.WriteFeatureIndexes(ws)
	if err != nil {
		t.Fatalf("WriteFeatureIndexes: %v", err)
	}
	if len(written) != 0 {
		t.Fatalf("rewrote %v; a semantically identical index must be left alone", written)
	}
	after, err := os.ReadFile(idx)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != reshaped {
		t.Fatal("the index was rewritten despite the guard reporting no write")
	}

	// The other half of the guard: a real structural change still writes.
	if err := os.WriteFile(idx, []byte("platform: communications\ncomponents: {}\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	written, err = graph.WriteFeatureIndexes(ws)
	if err != nil {
		t.Fatalf("WriteFeatureIndexes: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("wrote %v; a drifted index must be regenerated", written)
	}
	if after, _ := os.ReadFile(idx); string(after) != string(original) {
		t.Fatalf("regenerated index is not the committed one:\n%s", after)
	}
}

// TestDeriveTagsIsSortedAndDeduplicated pins the two properties every consumer
// depends on: the result is stable across runs, and a facet contributed twice
// appears once.
func TestDeriveTagsIsSortedAndDeduplicated(t *testing.T) {
	meta := yamlio.PyMap{
		{K: "type", V: yamlio.PyStr("prd")},
		{K: "platform", V: yamlio.PyStr("communications")},
		{K: "team", V: yamlio.PyStr("customer-engagement")},
		{K: "components", V: yamlio.PySeq{yamlio.PyStr("svc-a"), yamlio.PyStr("svc-a")}},
		{K: "boundedContext", V: yamlio.PyStr("context://communications")},
		{K: "status", V: yamlio.PyStr("proposed")},
		{K: "fromDiscovery", V: yamlio.PyStr("none")},
	}
	tags, err := graph.DeriveTags(meta, []string{"platform/communications", "capability/x"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"capability/x", "component/svc-a", "context/communications", "kind/prd",
		"platform/communications", "status/proposed", "team/customer-engagement",
	}
	if strings.Join(tags, ",") != strings.Join(want, ",") {
		t.Fatalf("tags = %v, want %v", tags, want)
	}
	if !sort.StringsAreSorted(tags) {
		t.Fatalf("tags are not sorted: %v", tags)
	}
}

// TestFeatureIndexUnresolvedOrderComesFromTheBuilder is R-0.11 at this site:
// the findings are ordered because BuildFeatureIndex inserted components in
// sorted order, not because anything sorts them afterwards. Iterating a Go map
// here would randomize a sequence a golden freezes.
func TestFeatureIndexUnresolvedOrderComesFromTheBuilder(t *testing.T) {
	ws := copyFixture(t, "workspace")
	pdir := filepath.Join(ws.Root, "platforms", "communications")
	idx, err := graph.BuildFeatureIndex(ws, pdir)
	if err != nil {
		t.Fatal(err)
	}
	components, ok := idx.Get("components").(yamlio.PyMap)
	if !ok {
		t.Fatal("no components mapping")
	}
	keys := make([]string, len(components))
	for i, p := range components {
		keys[i] = p.K
	}
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("component insertion order is not sorted: %v", keys)
	}
	// Same index, twenty times: a map-iteration order leak shows up as a
	// changing sequence rather than as a wrong one.
	first := fmt.Sprint(graph.FeatureIndexUnresolved(ws, idx))
	for i := 0; i < 20; i++ {
		if got := fmt.Sprint(graph.FeatureIndexUnresolved(ws, idx)); got != first {
			t.Fatalf("unresolved order is not deterministic: %s vs %s", first, got)
		}
	}
}

// TestRewriteGeneratedBlockIsFailSafe covers the four cases of R-3.3 to R-3.6
// through the public build, since an unbalanced node must leave the file byte
// for byte alone and still let every other root be written.
func TestRewriteGeneratedBlockIsFailSafe(t *testing.T) {
	ws := copyFixture(t, "workspace")
	node := filepath.Join(ws.Root, "company-os", "CLAUDE.md")
	original, err := os.ReadFile(node)
	if err != nil {
		t.Fatal(err)
	}
	// Two starts, one end: neither the balanced-pair branch nor the
	// no-markers branch applies.
	broken := append([]byte("<!-- company-os:generated:start -->\n"), original...)
	if err := os.WriteFile(node, broken, 0o666); err != nil {
		t.Fatal(err)
	}
	sections, err := graph.Build(ws)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	after, err := os.ReadFile(node)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, broken) {
		t.Fatal("an unbalanced node was mutated; the rewrite must be fail-safe")
	}
	var warned bool
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Code == model.CodeGraphNodeMarkersUnbalanced {
				warned = true
				if f.Severity != model.SevWarn {
					t.Errorf("marker imbalance reported at severity %v, want warn", f.Severity)
				}
				if f.Fields.Int("starts") != 2 || f.Fields.Int("ends") != 1 {
					t.Errorf("counts = %d start / %d end, want 2 / 1",
						f.Fields.Int("starts"), f.Fields.Int("ends"))
				}
			}
		}
	}
	if !warned {
		t.Fatal("no warning was reported for the unbalanced node")
	}
	// A warn must not make the run fail: the fail-safe answer is to leave the
	// file alone and keep going.
	if model.HasFailure(sections) {
		t.Error("a marker imbalance made graph build report a failure")
	}
}

// TestRebuildEmitsNoSummary pins what still separates the two entry points.
//
// It used to also assert that Rebuild emitted no TAG section. R-0.4 inverted
// that half deliberately: the repair path has to report which documents it
// re-tagged, and Rebuild is the entry point it calls. What survives is the
// summary — the "scanned N, updated M" tally is cmd_graph's alone, and a
// Rebuild that emitted it would print a tally in front of every scaffolding
// command's output.
func TestRebuildEmitsNoSummary(t *testing.T) {
	ws := copyFixture(t, "failing-workspace")
	sections, err := graph.Rebuild(ws)
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	for _, s := range sections {
		if s.Slug == model.SectionSummary {
			t.Errorf("Rebuild emitted the %q section, which only `graph build` prints", s.Slug)
		}
	}
	// It still WRITES the tags — the fixture has a drifted brief — even though
	// it announces nothing.
	brief := filepath.Join(ws.Root, "teams", "ghost", "product", "discovery",
		"2035-drifted-tags", "brief.md")
	text, err := os.ReadFile(brief)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "tags: [kind/discovery, status/draft, team/ghost]") {
		t.Fatalf("Rebuild did not re-derive tags:\n%s", text)
	}
}

// ------------------------------------------------------------------ helpers

func copyFixture(t *testing.T, name string) *workspace.Workspace {
	t.Helper()
	src, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", name))
	if err != nil {
		t.Fatalf("resolving fixture: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "ws")
	if err := copyTree(src, dst); err != nil {
		t.Fatalf("copying fixture: %v", err)
	}
	return workspace.New(dst)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o777)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		// Federated fixtures ship 0444 slices; the copy has to be writable or
		// a later t.TempDir cleanup fails on some platforms.
		return os.WriteFile(target, data, 0o666)
	})
}

// treeHash renders "path  sha256" per file, sorted — the same shape
// acceptance.sh §4's snapshot() produces, so a failure here reads like one
// from there.
func treeHash(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%x  %s", sha256.Sum256(data), filepath.ToSlash(rel)))
		return nil
	})
	if err != nil {
		t.Fatalf("hashing %s: %v", root, err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// diffTrees names only the paths whose hash line differs, which is the useful
// part of a two-thousand-line comparison.
func diffTrees(a, b string) string {
	seen := map[string]bool{}
	for _, l := range strings.Split(a, "\n") {
		seen[l] = true
	}
	var out []string
	for _, l := range strings.Split(b, "\n") {
		if !seen[l] {
			out = append(out, "  candidate-only: "+l)
		}
	}
	seen = map[string]bool{}
	for _, l := range strings.Split(b, "\n") {
		seen[l] = true
	}
	for _, l := range strings.Split(a, "\n") {
		if !seen[l] {
			out = append(out, "  reference-only: "+l)
		}
	}
	return strings.Join(out, "\n")
}

// reverseTop reverses a document's top-level key order without touching its
// content, so the reshaped file parses to the same structure.
func reverseTop(t *testing.T, raw []byte) yamlio.PyValue {
	t.Helper()
	v, err := yamlio.PyLoadBytes(raw, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := v.(yamlio.PyMap)
	if !ok {
		t.Fatal("fixture is not a mapping")
	}
	out := make(yamlio.PyMap, 0, len(m))
	for i := len(m) - 1; i >= 0; i-- {
		out = append(out, m[i])
	}
	return out
}

// TestGraphWritersPropagatePermission is R-0.2 against a real EACCES, not a
// fake: it chmods the containing directory and lets the kernel produce the
// error.
//
// The failure it prevents is D5. A repair pass that touches a file inside a
// read-only synced slice must be able to say "this is a slice, re-sync it
// instead" rather than "cannot write …: permission denied", and the only way to
// tell those apart is structurally. Before R-0.1/R-0.2, errors.Is against
// fs.ErrPermission returned false for every one of these writers, because
// model.Errorf flattened the cause through fmt.Sprintf.
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.2
func TestGraphWritersPropagatePermission(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores the mode bits this test depends on")
	}

	t.Run("RewriteFrontmatterTags", func(t *testing.T) {
		dir := t.TempDir()
		doc := filepath.Join(dir, "brief.md")
		if err := os.WriteFile(doc, []byte("---\ntype: discovery\n---\n\nbody\n"), 0o666); err != nil {
			t.Fatal(err)
		}
		defer chmodRO(t, doc)()

		_, err := graph.RewriteFrontmatterTags(doc, []string{"kind/discovery"})
		assertPermission(t, err)
	})

	t.Run("PyWriteCanonical", func(t *testing.T) {
		dir := t.TempDir()
		defer chmodRO(t, dir)()

		err := yamlio.PyWriteCanonical(filepath.Join(dir, "feature-index.yaml"),
			yamlio.PyMap{})
		assertPermission(t, err)
	})

	// The node writer has no exported entry point, so this drives it the way a
	// repair pass would: through Rebuild, against a workspace root whose
	// CLAUDE.md cannot be replaced.
	t.Run("rewriteGeneratedBlock via Rebuild", func(t *testing.T) {
		ws := copyFixture(t, "workspace")
		root := filepath.Join(ws.Root, "company-os")
		if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); err != nil {
			t.Skipf("fixture has no node to write: %v", err)
		}
		// Force a rewrite by leaving a MARKED node with a stale interior.
		//
		// This deliberately does not use a marker-less file. Since hand-owned
		// nodes stopped being adopted, a file with no markers is never written
		// at all — sealing one would prove nothing, because no write is
		// attempted and no permission error can surface. The interior must be
		// wrong so the replace branch actually reaches the filesystem.
		if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"),
			[]byte("# company-os\n\n<!-- company-os:generated:start -->\nstale\n<!-- company-os:generated:end -->\n"), 0o666); err != nil {
			t.Fatal(err)
		}
		defer chmodRO(t, filepath.Join(root, "CLAUDE.md"))()

		_, err := graph.Rebuild(ws)
		assertPermission(t, err)
	})
}

// TestRebuildReportsHandOwnedNodes is the caller half of R-2.12: refusing to
// write a marker-less CLAUDE.md is correct, but refusing SILENTLY is the bug.
// The unit test on rewriteGeneratedBlock proves the refusal; only this proves
// the user is told which file was skipped.
//
// Worth stating why this needs its own test: no golden fixture contains a
// marker-less node, so the acceptance suite passes whether or not this finding
// is emitted. Without this test the reporting path has zero coverage.
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-2.12
func TestRebuildReportsHandOwnedNodes(t *testing.T) {
	ws := copyFixture(t, "workspace")
	root := filepath.Join(ws.Root, "company-os")
	node := filepath.Join(root, "CLAUDE.md")
	if _, err := os.Stat(node); err != nil {
		t.Skipf("fixture has no node to write: %v", err)
	}
	const prose = "# company-os\n\nhand-written, never marked\n"
	if err := os.WriteFile(node, []byte(prose), 0o666); err != nil {
		t.Fatal(err)
	}

	gates, err := graph.Rebuild(ws)
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	var found *model.Finding
	for _, gate := range gates {
		for i, f := range gate.Findings {
			if f.Code == model.CodeNodeHandOwned {
				found = &gate.Findings[i]
			}
		}
	}
	if found == nil {
		t.Fatal("Rebuild skipped a marker-less CLAUDE.md and reported nothing; " +
			"a user cannot tell a silent skip from a successful write")
	}
	// The finding has to name the file, or it cannot be acted on.
	if !strings.Contains(found.Path, "CLAUDE.md") {
		t.Errorf("hand-owned finding does not name the file: Path = %q", found.Path)
	}
	if found.Severity != model.SevOK {
		t.Errorf("severity = %v, want SevOK — a hand-owned node is a legitimate "+
			"terminal state, not a failure", found.Severity)
	}
	// And the refusal itself still holds: reporting must not become writing.
	raw, err := os.ReadFile(node)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != prose {
		t.Errorf("hand-owned node was modified:\n got: %q\nwant: %q", string(raw), prose)
	}
}

// TestGraphWriterMessagesUnchanged is R-0.3 at the call sites 0.2 touched. The
// wrap must be invisible: same prefix, same rendered cause, same exit code.
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.3
func TestGraphWriterMessagesUnchanged(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores the mode bits this test depends on")
	}
	dir := t.TempDir()
	doc := filepath.Join(dir, "brief.md")
	if err := os.WriteFile(doc, []byte("---\ntype: discovery\n---\n\nbody\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	defer chmodRO(t, doc)()

	_, err := graph.RewriteFrontmatterTags(doc, []string{"kind/discovery"})
	if err == nil {
		t.Fatal("write into a read-only directory succeeded")
	}
	if !strings.HasPrefix(err.Error(), "cannot write "+doc+": ") {
		t.Errorf("message shape changed by wrapping:\n%s", err.Error())
	}
	if !strings.HasSuffix(err.Error(), "permission denied") {
		t.Errorf("cause no longer renders into the message:\n%s", err.Error())
	}
	if got := model.CodeOf(err); got != model.ExitArtifact {
		t.Errorf("CodeOf = %v, want %v — wrapping must not move an exit code", got, model.ExitArtifact)
	}
}

// chmodRO seals a path against writes and returns the restore func, so
// t.TempDir's own cleanup can still remove the tree.
//
// It seals FILES at 0444 and DIRECTORIES at 0555, which is what `workspace sync`
// materializes. The distinction is load-bearing: on POSIX a read-only directory
// still permits overwriting an existing file — only entry creation and removal
// are governed by the directory bit — so sealing the parent would not exercise
// the writers at all. Two of the three write paths overwrite in place.
func chmodRO(t *testing.T, path string) func() {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	mode := os.FileMode(0o444)
	if info.IsDir() {
		mode = 0o555
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	return func() { _ = os.Chmod(path, info.Mode().Perm()) }
}

func assertPermission(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("write into a read-only directory succeeded")
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("errors.Is(err, fs.ErrPermission) = false for %T: %v", err, err)
	}
}

// TestRebuildReportsRetaggedDocs is R-0.4: the repair path's whole reporting
// surface depends on Rebuild answering "what did you re-tag?".
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.4
func TestRebuildReportsRetaggedDocs(t *testing.T) {
	ws := copyFixture(t, "failing-workspace")
	sections, err := graph.Rebuild(ws)
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	var tags *model.GateResult
	for i := range sections {
		if sections[i].Slug == model.SectionTags {
			tags = &sections[i]
		}
	}
	if tags == nil {
		t.Fatal("Rebuild reported no tag section")
	}
	// The fixture's drifted brief is the document that must be named.
	const want = "teams/ghost/product/discovery/2035-drifted-tags/brief.md"
	var found bool
	for _, f := range tags.Findings {
		if f.Path == want {
			found = true
			if f.Code != model.CodeGraphTagged {
				t.Errorf("code = %v, want %v", f.Code, model.CodeGraphTagged)
			}
		}
	}
	if !found {
		t.Fatalf("the re-tagged document was not named in the section:\n%+v", tags.Findings)
	}
}

// TestRetagIsTheOnlyTraversal is R-0.6, enforced structurally rather than by
// inspection: every path the section names must actually have been rewritten by
// the run that produced it.
//
// A re-derived list would pass a naive "is it non-empty" check while being
// wrong. So this runs Rebuild twice. The second run rewrites nothing — the
// first one already converged the workspace — and a section built from a
// separate traversal (say, "docs whose derived tags differ from a fresh
// derivation") would still be free to report entries on that second run.
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.6
func TestRetagIsTheOnlyTraversal(t *testing.T) {
	ws := copyFixture(t, "failing-workspace")
	if _, err := graph.Rebuild(ws); err != nil {
		t.Fatalf("first Rebuild: %v", err)
	}
	sections, err := graph.Rebuild(ws)
	if err != nil {
		t.Fatalf("second Rebuild: %v", err)
	}
	for _, s := range sections {
		if s.Slug != model.SectionTags {
			continue
		}
		if len(s.Findings) != 0 {
			t.Errorf("the converged second run still named %d re-tagged document(s); "+
				"the list is not coming from the writing loop:\n%+v",
				len(s.Findings), s.Findings)
		}
	}
}

// TestBuildOutputUnchangedByR04 is R-0.5. `graph build` emitted a tag section
// long before Rebuild did; refactoring the two entry points onto one shared
// traversal must not have moved anything in it.
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.5
func TestBuildOutputUnchangedByR04(t *testing.T) {
	ws := copyFixture(t, "failing-workspace")
	sections, err := graph.Build(ws)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(sections) != 4 {
		t.Fatalf("got %d sections, want the 4 `graph build` has always emitted", len(sections))
	}
	want := []struct {
		ordinal int
		slug    string
	}{
		{1, model.SectionTags},
		{2, model.SectionFeatureIndexes},
		{3, model.SectionClaudeNodes},
		{4, model.SectionSummary},
	}
	for i, w := range want {
		if sections[i].Slug != w.slug || sections[i].Ordinal != w.ordinal {
			t.Errorf("section %d = %q ordinal %d, want %q ordinal %d",
				i, sections[i].Slug, sections[i].Ordinal, w.slug, w.ordinal)
		}
	}
	// The summary tally must still count the tag rewrites, which now arrive
	// via the shared traversal rather than a local counter.
	var scanned, updated int
	for _, f := range sections[3].Findings {
		scanned, updated = f.Fields.Int("scanned"), f.Fields.Int("updated")
	}
	if scanned == 0 {
		t.Error("summary scanned 0 documents")
	}
	if updated != len(sections[0].Findings) {
		t.Errorf("summary updated = %d but the tag section names %d documents",
			updated, len(sections[0].Findings))
	}
}
