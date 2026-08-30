package next

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// write creates a file with the given body, creating parent directories as needed.
func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

// emptyFixture builds the smallest workspace with no pending actions.
func emptyFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "company-os", "standards", "company-baseline.yaml"),
		"schemaVersion: '1.0'\ncontrols: []\n")
	write(t, filepath.Join(root, "platforms", "p1", "platform.yaml"),
		"schemaVersion: '1.0'\nid: platform://p1\n")
	write(t, filepath.Join(root, "teams", "t1", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://t1\n")
	write(t, filepath.Join(root, "teams", "t1", "governance", "deviations.yaml"),
		"schemaVersion: '1.0'\nteam: t1\ndeviations: []\n")
	write(t, filepath.Join(root, "teams", "t1", "governance", "exceptions.yaml"),
		"schemaVersion: '1.0'\nteam: t1\nexceptions: []\n")
	return workspace.New(root)
}

// expiryFixture builds a workspace with one expired deviation.
func expiryFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	ws := emptyFixture(t)
	write(t, filepath.Join(ws.Root, "teams", "t1", "governance", "deviations.yaml"),
		"schemaVersion: '1.0'\nteam: t1\ndeviations:\n"+
			"  - rule: company-standard://estimation/story-points\n"+
			"    tier: default\n    status: approved\n"+
			"    rationale: team uses cycle time\n"+
			"    reviewDate: 2020-01-01\n")
	return ws
}

// outcomeFixture builds a workspace with one pending outcome review.
func outcomeFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	ws := emptyFixture(t)
	prdDir := filepath.Join(ws.Root, "platforms", "p1", "archive", "prds", "2025-test-prd")
	write(t, filepath.Join(prdDir, "prd.md"),
		"---\ntype: prd\nid: 2025-test-prd\nstatus: completed\n"+
			"title: Test PRD\ncreated: 2025-01-01\n---\n\n# Test\n")
	write(t, filepath.Join(prdDir, "outcome.md"),
		"---\ntype: outcome-review\nprd: 2025-test-prd\n"+
			"due: 2025-06-01\nstatus: pending\n---\n\n# Outcome\n")
	return ws
}

// contractFixture builds a workspace with an active PRD that has contract issues.
func contractFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	ws := emptyFixture(t)
	prdDir := filepath.Join(ws.Root, "platforms", "p1", "change-records", "active", "2026-broken-prd")
	// Missing required sections and process fields.
	write(t, filepath.Join(prdDir, "prd.md"),
		"---\ntype: prd\nid: 2026-broken-prd\nstatus: proposed\n---\n\n# Broken\n")
	return ws
}

// TestPriorityOrdering verifies that an expiry outranks a PRD issue which
// outranks an outcome review.
func TestPriorityOrdering(t *testing.T) {
	// Build a workspace with ALL three priority levels present.
	root := t.TempDir()
	write(t, filepath.Join(root, "company-os", "standards", "company-baseline.yaml"),
		"schemaVersion: '1.0'\ncontrols: []\n")
	write(t, filepath.Join(root, "platforms", "p1", "platform.yaml"),
		"schemaVersion: '1.0'\nid: platform://p1\n")
	write(t, filepath.Join(root, "teams", "t1", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://t1\n")

	// Priority 1: expired deviation.
	write(t, filepath.Join(root, "teams", "t1", "governance", "deviations.yaml"),
		"schemaVersion: '1.0'\nteam: t1\ndeviations:\n"+
			"  - rule: company-standard://estimation/story-points\n"+
			"    tier: default\n    status: approved\n"+
			"    rationale: team uses cycle time\n"+
			"    reviewDate: 2020-01-01\n")
	write(t, filepath.Join(root, "teams", "t1", "governance", "exceptions.yaml"),
		"schemaVersion: '1.0'\nteam: t1\nexceptions: []\n")

	// Priority 2: active PRD with contract issues.
	prdDir := filepath.Join(root, "platforms", "p1", "change-records", "active", "2026-broken-prd")
	write(t, filepath.Join(prdDir, "prd.md"),
		"---\ntype: prd\nid: 2026-broken-prd\nstatus: proposed\n---\n\n# Broken\n")

	// Priority 4: pending outcome review.
	outDir := filepath.Join(root, "platforms", "p1", "archive", "prds", "2025-done-prd")
	write(t, filepath.Join(outDir, "prd.md"),
		"---\ntype: prd\nid: 2025-done-prd\nstatus: completed\n"+
			"title: Done PRD\ncreated: 2025-01-01\n---\n\n# Done\n")
	write(t, filepath.Join(outDir, "outcome.md"),
		"---\ntype: outcome-review\nprd: 2025-done-prd\n"+
			"due: 2025-06-01\nstatus: pending\n---\n\n# Outcome\n")

	ws := workspace.New(root)
	sections, err := Scan(ws, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) == 0 || len(sections[0].Findings) == 0 {
		t.Fatal("expected at least one finding")
	}

	// The single-action mode should return the expiry (priority 1).
	f := sections[0].Findings[0]
	if f.Code != model.CodeNextExpiry {
		t.Errorf("expected priority-1 expiry, got %s: %s", f.Code, f.Message)
	}
}

// TestEmptyWorkspace verifies that an empty workspace returns the
// "no pending actions" finding.
func TestEmptyWorkspace(t *testing.T) {
	ws := emptyFixture(t)
	sections, err := Scan(ws, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) == 0 || len(sections[0].Findings) == 0 {
		t.Fatal("expected at least one finding")
	}
	f := sections[0].Findings[0]
	if f.Code != model.CodeNextEmpty {
		t.Errorf("expected CodeNextEmpty, got %s: %s", f.Code, f.Message)
	}
}

// TestExpiryOutranksOutcome verifies that an expired deviation outranks
// a pending outcome review.
func TestExpiryOutranksOutcome(t *testing.T) {
	ws := expiryFixture(t)
	// Add an outcome review too.
	outDir := filepath.Join(ws.Root, "platforms", "p1", "archive", "prds", "2025-done-prd")
	write(t, filepath.Join(outDir, "prd.md"),
		"---\ntype: prd\nid: 2025-done-prd\nstatus: completed\n"+
			"title: Done PRD\ncreated: 2025-01-01\n---\n\n# Done\n")
	write(t, filepath.Join(outDir, "outcome.md"),
		"---\ntype: outcome-review\nprd: 2025-done-prd\n"+
			"due: 2025-06-01\nstatus: pending\n---\n\n# Outcome\n")

	sections, err := Scan(ws, false)
	if err != nil {
		t.Fatal(err)
	}
	f := sections[0].Findings[0]
	if f.Code != model.CodeNextExpiry {
		t.Errorf("expected expiry to outrank outcome, got %s", f.Code)
	}
}

// TestContractOutranksOutcome verifies that an active PRD with contract
// issues outranks a pending outcome review.
func TestContractOutranksOutcome(t *testing.T) {
	ws := contractFixture(t)
	// Add an outcome review too.
	outDir := filepath.Join(ws.Root, "platforms", "p1", "archive", "prds", "2025-done-prd")
	write(t, filepath.Join(outDir, "prd.md"),
		"---\ntype: prd\nid: 2025-done-prd\nstatus: completed\n"+
			"title: Done PRD\ncreated: 2025-01-01\n---\n\n# Done\n")
	write(t, filepath.Join(outDir, "outcome.md"),
		"---\ntype: outcome-review\nprd: 2025-done-prd\n"+
			"due: 2025-06-01\nstatus: pending\n---\n\n# Outcome\n")

	sections, err := Scan(ws, false)
	if err != nil {
		t.Fatal(err)
	}
	f := sections[0].Findings[0]
	if f.Code != model.CodeNextContract {
		t.Errorf("expected contract to outrank outcome, got %s", f.Code)
	}
}

// TestOutcomeWhenNothingElse verifies that a pending outcome review is
// surfaced when no higher-priority items exist.
func TestOutcomeWhenNothingElse(t *testing.T) {
	ws := outcomeFixture(t)
	sections, err := Scan(ws, false)
	if err != nil {
		t.Fatal(err)
	}
	f := sections[0].Findings[0]
	if f.Code != model.CodeNextOutcome {
		t.Errorf("expected CodeNextOutcome, got %s: %s", f.Code, f.Message)
	}
}

// TestAllGrouping verifies that --all returns group headers for each
// priority kind that has pending items.
func TestAllGrouping(t *testing.T) {
	// Build a workspace with expiry + outcome.
	root := t.TempDir()
	write(t, filepath.Join(root, "company-os", "standards", "company-baseline.yaml"),
		"schemaVersion: '1.0'\ncontrols: []\n")
	write(t, filepath.Join(root, "platforms", "p1", "platform.yaml"),
		"schemaVersion: '1.0'\nid: platform://p1\n")
	write(t, filepath.Join(root, "teams", "t1", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://t1\n")
	write(t, filepath.Join(root, "teams", "t1", "governance", "deviations.yaml"),
		"schemaVersion: '1.0'\nteam: t1\ndeviations:\n"+
			"  - rule: company-standard://estimation/story-points\n"+
			"    tier: default\n    status: approved\n"+
			"    rationale: team uses cycle time\n"+
			"    reviewDate: 2020-01-01\n")
	write(t, filepath.Join(root, "teams", "t1", "governance", "exceptions.yaml"),
		"schemaVersion: '1.0'\nteam: t1\nexceptions: []\n")
	outDir := filepath.Join(root, "platforms", "p1", "archive", "prds", "2025-done-prd")
	write(t, filepath.Join(outDir, "prd.md"),
		"---\ntype: prd\nid: 2025-done-prd\nstatus: completed\n"+
			"title: Done PRD\ncreated: 2025-01-01\n---\n\n# Done\n")
	write(t, filepath.Join(outDir, "outcome.md"),
		"---\ntype: outcome-review\nprd: 2025-done-prd\n"+
			"due: 2025-06-01\nstatus: pending\n---\n\n# Outcome\n")

	ws := workspace.New(root)
	sections, err := Scan(ws, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) == 0 {
		t.Fatal("expected sections")
	}

	// Count group headers and item codes.
	var groups []string
	var codes []string
	for _, f := range sections[0].Findings {
		if f.Code == model.CodeNextGroup {
			groups = append(groups, f.Fields.Str("group"))
		} else {
			codes = append(codes, f.Code)
		}
	}

	// Should have two groups: expiry and outcome.
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d: %v", len(groups), groups)
	}
	if groups[0] != "expiry" {
		t.Errorf("expected first group 'expiry', got %q", groups[0])
	}
	if groups[1] != "outcome" {
		t.Errorf("expected second group 'outcome', got %q", groups[1])
	}

	// Should have one expiry item and one outcome item.
	if len(codes) != 2 {
		t.Errorf("expected 2 items, got %d: %v", len(codes), codes)
	}
	if codes[0] != model.CodeNextExpiry {
		t.Errorf("expected first item CodeNextExpiry, got %s", codes[0])
	}
	if codes[1] != model.CodeNextOutcome {
		t.Errorf("expected second item CodeNextOutcome, got %s", codes[1])
	}
}

// TestAllEmpty verifies that --all on an empty workspace returns the
// "no pending actions" finding.
func TestAllEmpty(t *testing.T) {
	ws := emptyFixture(t)
	sections, err := Scan(ws, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) == 0 || len(sections[0].Findings) == 0 {
		t.Fatal("expected at least one finding")
	}
	f := sections[0].Findings[0]
	if f.Code != model.CodeNextEmpty {
		t.Errorf("expected CodeNextEmpty, got %s", f.Code)
	}
}
