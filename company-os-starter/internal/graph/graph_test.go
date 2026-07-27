package graph_test

import (
	"bytes"
	"crypto/sha256"
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
// They were removed rather than frozen because internal/difftest reaches the
// same ground end to end: graph/build-<fixture> for all six committed
// workspaces plus the three failure-path ones, graph/build-twice for
// idempotence, and graph/build-after-discover for a brand-new artifact — each
// freezing stdout AND a hash of every file in the tree. What is lost is the
// claim "matches Python"; what remains is "has not changed", which is all any
// test can assert once the reference is gone.
//
// The convergence property the first test pinned is NOT lost: three-pass
// convergence over the non-idempotent fixtures is exactly what a difftest
// golden records, and TestBuildIsIdempotentOnCommittedFixtures above still
// pins the two fixtures that must be fixed points from the start.

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

// TestRebuildEmitsNoTagsOrSummary pins the difference between the two entry
// points. rebuild_generated re-tags silently and prints no tally; a Rebuild
// that started emitting either would insert lines in front of every scaffolding
// command's output.
func TestRebuildEmitsNoTagsOrSummary(t *testing.T) {
	ws := copyFixture(t, "failing-workspace")
	sections, err := graph.Rebuild(ws)
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	for _, s := range sections {
		if s.Slug == model.SectionTags || s.Slug == model.SectionSummary {
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
