package graph_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
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
