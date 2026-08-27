package graph_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// TestProvenanceFieldsSurviveDerive is R-4.5 and R-4.6.
//
// This is the one place this change can corrupt user data silently.
// TestRewriteFrontmatterTagsPreservesUnknownKeys already covers unknown-key
// preservation, but it uses a plain string SCALAR. `generated:` is a mapping and
// `verified:` is a sequence of mappings, and nested collections re-layout
// through PyDumpAutoFlow — the site of the flow/block divergence recorded as
// R-0.7a(g) in the port. The untested shape is precisely the one this
// requirement is about, which is why R-4.6 names it explicitly.
//
// Byte equality of the whole frontmatter block is asserted, not key presence: a
// mapping that survives as a mapping but comes back inline where it was block
// has still changed the file, and would show up as gate-4 drift on someone
// else's machine.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.5
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.6
func TestProvenanceFieldsSurviveDerive(t *testing.T) {
	ws := workspace.New(t.TempDir())
	dir := filepath.Join(ws.Company, "standards")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	const body = `---
type: adr
id: adr-provenance
status: draft
generated:
  by: claude-opus-5/1
  at: 2026-08-26
verified:
  - by: human:ada
    at: 2026-08-26
  - by: process:nightly-audit
    at: 2026-08-25
tags: [kind/adr, status/draft]
---

# Provenance
`
	path := filepath.Join(dir, "adr-provenance.md")
	write(t, path, body)

	for i := 0; i < 2; i++ {
		if _, err := graph.Rebuild(ws); err != nil {
			t.Fatalf("derive run %d: %v", i+1, err)
		}
		got := readFile(t, path)
		if got != body {
			t.Fatalf("derive run %d changed the document.\n--- want ---\n%s\n--- got ---\n%s",
				i+1, body, got)
		}
	}
}

// TestTrustTier is R-4.3, R-4.8 and R-5.3.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.3
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.8
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-5.3
func TestTrustTier(t *testing.T) {
	for _, tc := range []struct{ name, fm, want string }{
		{"absent", "type: adr\nid: a\n", graph.TierUnverified},
		{"empty list", "type: adr\nid: a\nverified: []\n", graph.TierUnverified},
		{"agent only", "type: adr\nid: a\nverified:\n  - by: claude-opus-5/1\n",
			graph.TierMachineConfirmed},
		{"process only", "type: adr\nid: a\nverified:\n  - by: process:nightly\n",
			graph.TierMachineConfirmed},
		{"one human among agents", "type: adr\nid: a\nverified:\n  - by: claude-opus-5/1\n  - by: human:ada\n",
			graph.TierHumanReviewed},
		{"single mapping not a list", "type: adr\nid: a\nverified:\n  by: human:ada\n",
			graph.TierHumanReviewed},
		{"entry without by degrades", "type: adr\nid: a\nverified:\n  - at: 2026-08-26\n",
			graph.TierUnverified},

		// R-5.1: the legacy fields derive a human actor, so the tier means
		// something on day one instead of reading `unverified` everywhere.
		{"decisionOwner derives human", "type: prd\nid: a\ndecisionOwner: Ada (PM)\n",
			graph.TierHumanReviewed},
		{"approvedBy derives human", "type: adr\nid: a\napprovedBy: Ada\n",
			graph.TierHumanReviewed},

		// R-5.3: a scaffolded placeholder derives nothing. Both spellings, since
		// the two scaffolders disagree and that asymmetry is the point.
		{"decisionOwner TODO derives nothing", "type: prd\nid: a\ndecisionOwner: TODO\n",
			graph.TierUnverified},
		{"approvedBy TODO derives nothing", "type: adr\nid: a\napprovedBy: 'TODO: rule owner'\n",
			graph.TierUnverified},
	} {
		t.Run(tc.name, func(t *testing.T) {
			meta := parseFM(t, tc.fm)
			if got := graph.TrustTier(meta); got != tc.want {
				t.Errorf("TrustTier = %q, want %q\nfrontmatter:\n%s", got, tc.want, tc.fm)
			}
		})
	}
}

// TestTrustTierNeverGates is R-4.4. The tier is advisory; no gate may consume
// it. Asserted by construction: nothing outside this package and its tests
// references TrustTier, so a future gate that starts depending on it has to
// change this list first.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.4
func TestTrustTierNeverGates(t *testing.T) {
	roots := []string{"../validate", "../product", "../governance", "../federation"}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
				return nil
			}
			raw, readErr := os.ReadFile(p)
			if readErr != nil {
				return nil
			}
			if strings.Contains(string(raw), "TrustTier(") {
				t.Errorf("%s references TrustTier; R-4.4 forbids a gate depending "+
					"on an advisory signal", p)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func parseFM(t *testing.T, fm string) yamlio.PyMap {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "d.md")
	write(t, path, "---\n"+fm+"---\n\n# X\n")
	meta, _, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	return meta
}
