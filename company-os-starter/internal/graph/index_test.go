package graph_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// doc builds a Doc the way IterGraphDocs would, without touching the disk.
// BuildIndexes is a pure function of []Doc, which is what lets the threshold be
// tested without a fixture tree.
func doc(dir, name string, kv ...string) graph.Doc {
	meta := yamlio.PyMap{}
	for i := 0; i+1 < len(kv); i += 2 {
		meta = meta.Set(kv[i], yamlio.PyStr(kv[i+1]))
	}
	return graph.Doc{
		Path: filepath.Join(dir, name),
		Rel:  filepath.ToSlash(filepath.Join(filepath.Base(dir), name)),
		Meta: meta,
	}
}

// TestBuildIndexesHonoursTheTwoDocumentThreshold is R-2.1.
//
// The boundary is asserted from both sides at once — one document must produce
// nothing, two must produce an index — because an off-by-one here is silent: a
// threshold of 1 yields an index in every leaf directory and reads as working.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.1
func TestBuildIndexesHonoursTheTwoDocumentThreshold(t *testing.T) {
	lonely := filepath.Join("ws", "lonely")
	pair := filepath.Join("ws", "pair")
	trio := filepath.Join("ws", "trio")

	got := graph.BuildIndexes([]graph.Doc{
		doc(lonely, "only.md", "type", "adr", "title", "Only"),
		doc(pair, "a.md", "type", "adr", "title", "A"),
		doc(pair, "b.md", "type", "adr", "title", "B"),
		doc(trio, "a.md", "type", "adr", "title", "A"),
		doc(trio, "b.md", "type", "adr", "title", "B"),
		doc(trio, "c.md", "type", "adr", "title", "C"),
	})

	if _, ok := got[lonely]; ok {
		t.Errorf("a directory holding one document got an index; threshold is 2")
	}
	if _, ok := got[pair]; !ok {
		t.Errorf("a directory holding two documents got no index; threshold is 2")
	}
	if _, ok := got[trio]; !ok {
		t.Errorf("a directory holding three documents got no index")
	}
	if len(got) != 2 {
		t.Errorf("index count = %d, want 2 (pair and trio)", len(got))
	}
}

// TestBuildIndexesCountsOnlyDirectChildren is R-2.1's other half: "directly
// holding". A recursive count would make every ancestor qualify and produce an
// index at every level of the tree — the noise the threshold exists to prevent.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.1
func TestBuildIndexesCountsOnlyDirectChildren(t *testing.T) {
	parent := filepath.Join("ws", "reality")
	child := filepath.Join(parent, "components")

	got := graph.BuildIndexes([]graph.Doc{
		doc(parent, "one.md", "type", "adr", "title", "One"),
		doc(child, "a.md", "type", "adr", "title", "A"),
		doc(child, "b.md", "type", "adr", "title", "B"),
	})

	if _, ok := got[parent]; ok {
		t.Errorf("parent qualified on its children's documents; the count must " +
			"be of direct children only")
	}
	if _, ok := got[child]; !ok {
		t.Errorf("child holding two documents got no index")
	}
}

// TestBuildIndexRendersTitleAndDescription is R-2.2 and R-2.3.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.2
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.3
func TestBuildIndexRendersTitleAndDescription(t *testing.T) {
	dir := filepath.Join("ws", "reality")
	got := graph.BuildIndexes([]graph.Doc{
		doc(dir, "svc-a.md", "type", "component-reality",
			"title", "Notification service",
			"description", "Fans out webhooks; retries are unbounded."),
		doc(dir, "svc-b.md", "type", "component-reality", "id", "reality-svc-b"),
		doc(dir, "note.md", "type", "adr", "title", "Zed decision"),
	})
	body := got[dir]
	if body == "" {
		t.Fatal("no index produced")
	}

	// R-2.3: frontmatter carrying type: index, then a marker block.
	if !strings.HasPrefix(body, "---\ntype: index\n---\n") {
		t.Errorf("index does not open with `type: index` frontmatter:\n%s", body)
	}
	if !strings.Contains(body, "company-os:generated:start") ||
		!strings.Contains(body, "company-os:generated:end") {
		t.Errorf("index carries no generated marker block:\n%s", body)
	}

	// R-2.2: title + description, grouped by type.
	want := []string{
		"- **adr**",
		"  - [Zed decision](note.md)",
		"- **component-reality**",
		"  - [Notification service](svc-a.md) — Fans out webhooks; retries are unbounded.",
		"  - [reality-svc-b](svc-b.md)",
	}
	for _, w := range want {
		if !strings.Contains(body, w) {
			t.Errorf("index is missing line %q\n--- got ---\n%s", w, body)
		}
	}

	// Grouping is by type and sorted, so `adr` precedes `component-reality`
	// even though its document sorts last by filename.
	if strings.Index(body, "- **adr**") > strings.Index(body, "- **component-reality**") {
		t.Errorf("types are not sorted:\n%s", body)
	}

	// A document with no description renders no trailing dash.
	if strings.Contains(body, "[reality-svc-b](svc-b.md) —") {
		t.Errorf("a doc without a description rendered an empty dash:\n%s", body)
	}
}

// TestBuildIndexesIsDeterministic is what R-2.11's idempotency rests on. Go
// randomizes map iteration order, so an accidental dependence on it surfaces
// here rather than in the acceptance harness's double-build check much later.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.11
func TestBuildIndexesIsDeterministic(t *testing.T) {
	dir := filepath.Join("ws", "d")
	in := []graph.Doc{
		doc(dir, "a.md", "type", "adr", "title", "A"),
		doc(dir, "b.md", "type", "prd", "title", "B"),
		doc(dir, "c.md", "type", "adr", "title", "C"),
		doc(dir, "d.md", "type", "concept", "title", "D"),
	}
	first := graph.BuildIndexes(in)[dir]
	for i := 0; i < 20; i++ {
		if again := graph.BuildIndexes(in)[dir]; again != first {
			t.Fatalf("BuildIndexes is not deterministic\n--- run 1 ---\n%s\n--- run %d ---\n%s",
				first, i+2, again)
		}
	}
}

// TestWriteIndexesLeavesAHandWrittenIndexAlone is R-2.10.
//
// The first spec draft had this FAIL gate 5. That contradicted the decision the
// repo already settled for CLAUDE.md — a marker-less file is hand-owned, left
// alone, reported, and PASSES (internal/graph/node.go, internal/graph/gates.go)
// — and it would have added a blocking check, violating R-5.5 and invariant I1.
//
// Deleting authored content because a threshold moved is the worst failure this
// feature could have, so it is asserted from both sides: the write path must not
// clobber it, and the stale sweep must not delete it.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.10
func TestWriteIndexesLeavesAHandWrittenIndexAlone(t *testing.T) {
	ws := workspace.New(t.TempDir())
	dir := filepath.Join(ws.Company, "standards")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	const authored = "# My own index\n\nI wrote this by hand and it is not generated.\n"
	idx := filepath.Join(dir, "index.md")
	write(t, idx, authored)

	// Two documents, so the directory qualifies and the writer WANTS this path.
	write(t, filepath.Join(dir, "a.md"), "---\ntype: adr\nid: a\nstatus: draft\n---\n\n# A\n")
	write(t, filepath.Join(dir, "b.md"), "---\ntype: adr\nid: b\nstatus: draft\n---\n\n# B\n")

	docs, err := IterGraphDocsFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	findings, err := graph.WriteIndexes(ws, docs)
	if err != nil {
		t.Fatalf("WriteIndexes: %v", err)
	}
	if got := readFile(t, idx); got != authored {
		t.Errorf("a hand-written index.md was rewritten.\nwant: %q\ngot:  %q", authored, got)
	}
	if !hasCode(findings, model.CodeGraphDirIndexHandOwned) {
		t.Errorf("hand-owned index was not reported; silence is what makes a user "+
			"think --fix is broken. findings: %v", findings)
	}

	// Now drop below the threshold and re-run: the stale sweep must still not
	// touch it, because it carries no generated markers.
	if err := os.Remove(filepath.Join(dir, "b.md")); err != nil {
		t.Fatal(err)
	}
	docs, err = IterGraphDocsFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.WriteIndexes(ws, docs); err != nil {
		t.Fatalf("WriteIndexes after drop: %v", err)
	}
	if got := readFile(t, idx); got != authored {
		t.Errorf("the stale sweep deleted or rewrote a hand-written index.md; got %q", got)
	}
}

// TestWriteIndexesRemovesAStaleGeneratedIndex is R-2.9, the counterpart: a
// GENERATED index whose directory dropped below the threshold is deleted.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.9
func TestWriteIndexesRemovesAStaleGeneratedIndex(t *testing.T) {
	ws := workspace.New(t.TempDir())
	dir := filepath.Join(ws.Company, "standards")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "a.md"), "---\ntype: adr\nid: a\nstatus: draft\n---\n\n# A\n")
	write(t, filepath.Join(dir, "b.md"), "---\ntype: adr\nid: b\nstatus: draft\n---\n\n# B\n")

	docs, err := IterGraphDocsFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.WriteIndexes(ws, docs); err != nil {
		t.Fatal(err)
	}
	idx := filepath.Join(dir, "index.md")
	if _, err := os.Stat(idx); err != nil {
		t.Fatalf("index was not generated in the first place: %v", err)
	}

	if err := os.Remove(filepath.Join(dir, "b.md")); err != nil {
		t.Fatal(err)
	}
	docs, err = IterGraphDocsFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	findings, err := graph.WriteIndexes(ws, docs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(idx); !os.IsNotExist(err) {
		t.Errorf("a generated index survived its directory dropping to one document")
	}
	if !hasCode(findings, model.CodeGraphDirIndexRemoved) {
		t.Errorf("removal was not reported: %v", findings)
	}
}

// IterGraphDocsFor is a thin alias so the tests above read as the pipeline does.
func IterGraphDocsFor(ws *workspace.Workspace) ([]graph.Doc, error) {
	return graph.IterGraphDocs(ws)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func hasCode(findings []model.Finding, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}
