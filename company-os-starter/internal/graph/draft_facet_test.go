package graph_test

// R-3.6: a draft gets a `team/<t>` location facet.
//
// No derivation code was written for this. IterGraphDocs already infers the
// location facet from the parent directory of each graph ROOT — `teams/<t>` sits
// under `teams`, so every document anywhere beneath it carries `team/<t>`
// (tags.go:201-209). What was unproven is that the inference survives the DEPTH
// a draft introduces: `teams/<t>/product/change-records/draft/<id>/prd.md` is
// four directories below the root, deeper than any document the derivation was
// written against, and a facet keyed off the immediate parent instead of the
// root would silently produce `team/draft` or nothing at all.
//
// So the test is about depth, not about drafts as such — which is why it asserts
// the facet against a hand-placed file rather than one DraftNew wrote. DraftNew
// stamps `team:` into the frontmatter, and `team:` also derives a `team/` tag; a
// fixture carrying both cannot tell the two sources apart, and would go on
// passing after the directory inference was removed.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// TestDraftDerivesTeamLocationFacet is R-3.6 at draft depth.
//
// The document deliberately omits `team:` in frontmatter. The only way
// `team/core` can appear is the parent-directory inference.
func TestDraftDerivesTeamLocationFacet(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, filepath.Join(root, "teams", "core", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://core\n")
	rel := filepath.Join("teams", "core", "product", "change-records", "draft",
		"2026-deep", "prd.md")
	writeDoc(t, filepath.Join(root, rel),
		"---\ntype: prd\nid: 2026-deep\nstatus: draft\n---\n\n# Draft\n")

	docs, err := graph.IterGraphDocs(workspace.New(root))
	if err != nil {
		t.Fatalf("IterGraphDocs: %v", err)
	}

	var tags []string
	found := false
	for _, d := range docs {
		if filepath.ToSlash(d.Rel) == filepath.ToSlash(rel) {
			found = true
			tags = d.Tags
		}
	}
	if !found {
		t.Fatalf("the graph traversal did not reach the draft at %s; docs=%d", rel, len(docs))
	}
	if !hasTag(tags, "team/core") {
		t.Errorf("draft did not derive its team location facet; tags=%v", tags)
	}
	// The other two facets come from `type:` and `status:` and are asserted here
	// only so a derivation that collapsed to a single tag would not pass on the
	// one tag this test happens to name.
	for _, want := range []string{"kind/prd", "status/draft"} {
		if !hasTag(tags, want) {
			t.Errorf("draft is missing %q; tags=%v", want, tags)
		}
	}
	// The draft directory must not leak into the vocabulary as a location.
	for _, tag := range tags {
		if strings.HasPrefix(tag, "team/") && tag != "team/core" {
			t.Errorf("a directory below the team root produced a location facet: %q", tag)
		}
	}
}

// TestDraftTagsAreIdempotent is R-3.4's derived-artifact half at the unit level:
// building twice over a workspace holding a draft leaves the draft's bytes
// unchanged.
//
// A second build that rewrote the file would mean `validate` reports tag drift
// on a workspace nobody edited, which is the failure mode R-3.4 rules out. The
// fixture includes a `platform: TODO` placeholder because that is what an
// unrouted draft actually carries, and the resulting `platform/TODO` facet has
// to round-trip like any other derived tag rather than being special-cased away.
func TestDraftTagsAreIdempotent(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, filepath.Join(root, "teams", "core", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://core\n")
	path := filepath.Join(root, "teams", "core", "product", "change-records",
		"draft", "2026-idem", "prd.md")
	writeDoc(t, path, "---\ntype: prd\nid: 2026-idem\nstatus: draft\n"+
		"team: core\nplatform: TODO\ncomponents: []\n---\n\n# Draft\n")

	ws := workspace.New(root)
	if _, err := graph.Build(ws); err != nil {
		t.Fatalf("first build: %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := graph.Build(ws); err != nil {
		t.Fatalf("second build: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("a second graph build rewrote the draft:\n--- first ---\n%s\n--- second ---\n%s",
			first, second)
	}
	if !strings.Contains(string(first), "team/core") {
		t.Errorf("the written draft carries no team facet:\n%s", first)
	}
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func writeDoc(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}
