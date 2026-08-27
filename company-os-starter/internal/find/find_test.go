package find

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// setupFixture creates a minimal workspace under a temp directory with just
// enough structure for find to exercise each source.
func setupFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Create the canonical roots so workspace.IsRoot returns true.
	for _, d := range []string{"company-os", "platforms", "teams", "company-ontology", "knowledge"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Registry with two IDs: one that matches exactly and one by substring.
	regDir := filepath.Join(root, "company-ontology", "ids")
	if err := os.MkdirAll(regDir, 0755); err != nil {
		t.Fatal(err)
	}
	reg := `schemaVersion: '1.0'
kind: IdRegistry
ids:
- {id: 'component://alpha-svc', definedIn: platforms/test/components/alpha-svc.yaml}
- {id: 'component://beta-worker', definedIn: platforms/test/components/beta-worker.yaml}
tags: [ontology/registry]
`
	if err := os.WriteFile(filepath.Join(regDir, "registry.yaml"), []byte(reg), 0644); err != nil {
		t.Fatal(err)
	}

	// Platform with a component and feature-index.
	pdir := filepath.Join(root, "platforms", "test")
	for _, d := range []string{
		"components",
		"generated",
		"reality/components",
		"change-records/active",
	} {
		if err := os.MkdirAll(filepath.Join(pdir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Component descriptor.
	desc := `---
type: component
id: alpha-svc
title: Alpha Service
tags: [kind/component, platform/test]
---
`
	if err := os.WriteFile(filepath.Join(pdir, "components", "alpha-svc.yaml"), []byte(desc), 0644); err != nil {
		t.Fatal(err)
	}

	// Reality doc with a title containing the query term.
	reality := `---
type: component-reality
id: reality-alpha-svc
title: Alpha Service Reality
tags: [kind/reality, platform/test]
---
Current state of alpha-svc.
`
	if err := os.WriteFile(filepath.Join(pdir, "reality", "components", "alpha-svc.md"), []byte(reality), 0644); err != nil {
		t.Fatal(err)
	}

	// Feature-index.
	fi := `components:
  alpha-svc:
    reality: reality/components/alpha-svc.md
    archivedPrds:
    - 2026-alpha-improvement
platform: test
`
	if err := os.WriteFile(filepath.Join(pdir, "generated", "feature-index.yaml"), []byte(fi), 0644); err != nil {
		t.Fatal(err)
	}

	// Per-directory index.md (generated format with markers).
	idxDir := filepath.Join(pdir, "reality", "components")
	idx := `---
type: index
---

## components — index

- **component-reality**
  - [Alpha Service Reality](alpha-svc.md)
`
	if err := os.WriteFile(filepath.Join(idxDir, "index.md"), []byte(idx), 0644); err != nil {
		t.Fatal(err)
	}
	// Create a second doc in the same directory so the index would qualify.
	second := `---
type: component-reality
id: reality-beta-worker
title: Beta Worker Reality
tags: [kind/reality, platform/test]
---
`
	if err := os.WriteFile(filepath.Join(idxDir, "beta-worker.md"), []byte(second), 0644); err != nil {
		t.Fatal(err)
	}

	// Team with a doc carrying tags.
	tdir := filepath.Join(root, "teams", "team-a")
	for _, d := range []string{"product/discovery/2026-alpha-discovery"} {
		if err := os.MkdirAll(filepath.Join(tdir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}
	teamDoc := `---
type: discovery-brief
id: 2026-alpha-discovery
title: Alpha Discovery Brief
tags: [kind/discovery, team/team-a, component/alpha-svc]
---
`
	if err := os.WriteFile(filepath.Join(tdir, "product", "discovery", "2026-alpha-discovery", "brief.md"), []byte(teamDoc), 0644); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestExactIDFirst(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	results, err := Search(ws, "component://alpha-svc", true)
	if err != nil {
		t.Fatal(err)
	}

	// The IDs section should appear and the first finding should be the exact
	// match.
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	first := results[0]
	if first.Slug != model.SlugFindIDs {
		t.Fatalf("first section: want %s, got %s", model.SlugFindIDs, first.Slug)
	}
	if len(first.Findings) == 0 {
		t.Fatal("IDs section has no findings")
	}
	if first.Findings[0].Code != model.CodeFindExactID {
		t.Fatalf("first finding: want %s, got %s", model.CodeFindExactID, first.Findings[0].Code)
	}
}

func TestSubstringOverTagsTitleID(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// "alpha" should match:
	// - IDs: component://alpha-svc (substring)
	// - Tags: component/alpha-svc tag on the discovery brief
	// - Title: "Alpha Service Reality" on the reality doc
	// - Feature-index: alpha-svc component
	results, err := Search(ws, "alpha", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}

	// Collect all slugs present.
	slugs := map[string]bool{}
	for _, s := range results {
		slugs[s.Slug] = true
	}
	for _, want := range []string{model.SlugFindIDs, model.SlugFindTags, model.SlugFindFrontmatter, model.SlugFindFeature} {
		if !slugs[want] {
			t.Errorf("missing section %s in results", want)
		}
	}
}

func TestIndexContentHit(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// "Alpha Service Reality" appears in the index.md content.
	results, err := Search(ws, "Alpha Service Reality", true)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, s := range results {
		if s.Slug == model.SlugFindIndex {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected an index section for 'Alpha Service Reality'")
	}
}

func TestFeatureIndexHit(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// "2026-alpha-improvement" is an archived PRD id in the feature-index.
	results, err := Search(ws, "2026-alpha-improvement", true)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, s := range results {
		if s.Slug == model.SlugFindFeature {
			found = true
			for _, f := range s.Findings {
				if f.Code != model.CodeFindFeature {
					t.Errorf("want code %s, got %s", model.CodeFindFeature, f.Code)
				}
			}
			break
		}
	}
	if !found {
		t.Error("expected a feature-index section for '2026-alpha-improvement'")
	}
}

func TestNoMatchesExits0(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	results, err := Search(ws, "nonsense-xyz-123", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 section for no-matches, got %d", len(results))
	}
	if len(results[0].Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(results[0].Findings))
	}
	if results[0].Findings[0].Code != model.CodeFindNoMatches {
		t.Fatalf("want %s, got %s", model.CodeFindNoMatches, results[0].Findings[0].Code)
	}
}

func TestDeterminism(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// Run search twice and compare the record sets byte-for-byte by
	// serializing to a deterministic string representation.
	r1, err := Search(ws, "alpha", true)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := Search(ws, "alpha", true)
	if err != nil {
		t.Fatal(err)
	}

	s1 := serializeResults(r1)
	s2 := serializeResults(r2)
	if s1 != s2 {
		t.Errorf("non-deterministic output:\nrun 1:\n%s\nrun 2:\n%s", s1, s2)
	}
}

// serializeResults produces a deterministic string for comparison.
func serializeResults(results []model.GateResult) string {
	var b strings.Builder
	for _, s := range results {
		b.WriteString(s.Slug + "\n")
		for _, f := range s.Findings {
			b.WriteString(f.Code + "|" + f.Path + "|" + f.Subject + "|" + f.Message + "\n")
		}
	}
	return b.String()
}

func TestGraphifyHookSkippedWhenAbsent(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// Clear PATH so graphify is not found, and ensure no graphify-out/ exists.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	// No graphify binary, no graphify-out/ → no hint, no section.
	results, err := Search(ws, "alpha", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range results {
		if s.Slug == model.SlugFindGraphify {
			t.Error("expected no graphify section when neither binary nor graph.json exists")
		}
	}
}

func TestGraphifyHookHintWhenGraphJSONOnly(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// Create graphify-out/graph.json but no binary on PATH.
	gDir := filepath.Join(root, "graphify-out")
	if err := os.MkdirAll(gDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gDir, "graph.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// With --no-graphify, no hint section should appear even though graph.json
	// exists.
	results, err := Search(ws, "alpha", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range results {
		if s.Slug == model.SlugFindGraphify {
			t.Error("--no-graphify should suppress the graphify section entirely")
		}
	}

	// Without --no-graphify but with an empty PATH (no graphify binary),
	// the hook should produce a hint.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	results, err = Search(ws, "alpha", false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range results {
		if s.Slug == model.SlugFindGraphify {
			found = true
			if len(s.Findings) != 1 {
				t.Fatalf("expected 1 hint finding, got %d", len(s.Findings))
			}
			if s.Findings[0].Code != model.CodeFindGraphifyHint {
				t.Fatalf("want %s, got %s", model.CodeFindGraphifyHint, s.Findings[0].Code)
			}
			break
		}
	}
	if !found {
		t.Error("expected a graphify hint section when graph.json exists but binary is absent")
	}
}

func TestEmptyQueryIsUsageError(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	_, err := Search(ws, "", true)
	if err == nil {
		t.Fatal("expected usage error for empty query")
	}
}

func TestCaseInsensitive(t *testing.T) {
	root := setupFixture(t)
	ws := workspace.New(root)

	// "ALPHA" should match just like "alpha".
	r1, err := Search(ws, "ALPHA", true)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := Search(ws, "alpha", true)
	if err != nil {
		t.Fatal(err)
	}

	if serializeResults(r1) != serializeResults(r2) {
		t.Error("case-insensitive search should produce identical results")
	}
}
