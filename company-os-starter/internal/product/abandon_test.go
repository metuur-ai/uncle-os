package product

import (
	"os"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// draftFor scaffolds a draft for team `core` and returns its id and path.
func draftFor(t *testing.T, ws *workspace.Workspace, platform string) (string, string) {
	t.Helper()
	if _, err := DraftNew(ws, "core", platform, "Retire the widget", "", nil); err != nil {
		t.Fatal(err)
	}
	dir, err := DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	ids := DraftIDs(dir)
	if len(ids) != 1 {
		t.Fatalf("DraftIDs = %v, want exactly one", ids)
	}
	return ids[0], DraftPath(dir, ids[0])
}

// TestAbandonSetsStatus is R-8.1: the draft's status becomes `abandoned`, in
// place, with the rest of the frontmatter left alone.
func TestAbandonSetsStatus(t *testing.T) {
	ws := fixture(t)
	id, path := draftFor(t, ws, "payments")

	if _, err := PRDAbandon(ws, "core", id, nil); err != nil {
		t.Fatal(err)
	}

	meta, _, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strOf(meta, "status"); got != "abandoned" {
		t.Errorf("status = %q, want abandoned", got)
	}
	// The draft is kept, not removed or relocated (the whole point of a state
	// rather than an `rm`), and its routing survives.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("draft no longer at %s: %v", path, err)
	}
	if got := promoteToPlatform(meta); got != "payments" {
		t.Errorf("promoteTo.platform = %q, want payments (abandon must not scrub it)", got)
	}
}

// TestAbandonFrontmatterContract is R-10.9's parser contract: what abandon
// writes still opens `---\n`, closes `---\n`, and parses.
func TestAbandonFrontmatterContract(t *testing.T) {
	ws := fixture(t)
	id, path := draftFor(t, ws, "payments")
	if _, err := PRDAbandon(ws, "core", id, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "---\n") {
		t.Errorf("draft does not open with the fence:\n%s", first(string(data)))
	}
	if strings.Count(string(data), "\n---\n") < 1 {
		t.Errorf("draft has no closing fence:\n%s", first(string(data)))
	}
}

// TestAbandonRefusesNonDraft is R-8.2 and, for the `abandoned` case, R-8.4:
// there is no transition out, so abandoning an abandoned draft is the same
// conflict as abandoning a promoted one, and the status found is in the words.
func TestAbandonRefusesNonDraft(t *testing.T) {
	for _, status := range []string{"abandoned", "promoted", "proposed"} {
		t.Run(status, func(t *testing.T) {
			ws := fixture(t)
			id, path := draftFor(t, ws, "payments")
			setStatus(t, path, status)

			_, err := PRDAbandon(ws, "core", id, nil)
			if err == nil {
				t.Fatal("expected a conflict, got nil")
			}
			if code := model.CodeOf(err); code != model.ExitConflict {
				t.Errorf("exit = %d, want ExitConflict (%d)", code, model.ExitConflict)
			}
			if !strings.Contains(err.Error(), "'"+status+"'") {
				t.Errorf("error does not state the status found (%q): %v", status, err)
			}
		})
	}
}

// TestAbandonRefusalWritesNothing: a refused run leaves the draft exactly as it
// found it, the same guarantee promotion's pre-write refusals give (R-4.6).
func TestAbandonRefusalWritesNothing(t *testing.T) {
	ws := fixture(t)
	id, path := draftFor(t, ws, "payments")
	setStatus(t, path, "promoted")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PRDAbandon(ws, "core", id, nil); err == nil {
		t.Fatal("expected a conflict")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("a refused abandon rewrote the draft")
	}
}

// TestAbandonRebuilds is R-8.5: derived artifacts are rebuilt BEFORE the
// command returns, and the rebuild's own sections come back with the result.
func TestAbandonRebuilds(t *testing.T) {
	ws := fixture(t)
	id, _ := draftFor(t, ws, "payments")

	called := 0
	var rebuild Rebuild = func(*workspace.Workspace) ([]model.GateResult, error) {
		called++
		return []model.GateResult{{Ordinal: 99, Slug: "derived"}}, nil
	}
	out, err := PRDAbandon(ws, "core", id, rebuild)
	if err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("rebuild called %d times, want 1", called)
	}
	if len(out) != 2 || out[1].Slug != "derived" {
		t.Errorf("derived sections not spliced into the result: %+v", out)
	}
	if out[0].Findings[0].Code != model.CodePRDAbandoned {
		t.Errorf("first finding = %q, want %q",
			out[0].Findings[0].Code, model.CodePRDAbandoned)
	}
}

// TestAbandonRequiresTeam: a draft is team-private, so there is no run of this
// command without --team.
func TestAbandonRequiresTeam(t *testing.T) {
	ws := fixture(t)
	if _, err := PRDAbandon(ws, "", "anything", nil); err == nil {
		t.Fatal("expected a usage error")
	}
}

// TestAbandonUnknownDraft: an id with no artifact behind it is a workspace
// error naming the path, not a silent success.
func TestAbandonUnknownDraft(t *testing.T) {
	ws := fixture(t)
	_, err := PRDAbandon(ws, "core", "2026-nothing-here", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if code := model.CodeOf(err); code != model.ExitWorkspace {
		t.Errorf("exit = %d, want ExitWorkspace (%d)", code, model.ExitWorkspace)
	}
}

// TestAbandonedDraftIsNotPromotable is half of R-8.3: promotion refuses an
// abandoned draft, and says so in the status it found. (The other halves — the
// gates and the role view — are exclusions by status elsewhere; see
// internal/validate/promotion.go and DraftsSection.)
func TestAbandonedDraftIsNotPromotable(t *testing.T) {
	ws := fixture(t)
	id, _ := draftFor(t, ws, "payments")
	if _, err := PRDAbandon(ws, "core", id, nil); err != nil {
		t.Fatal(err)
	}
	_, err := PRDPromote(ws, "core", id, nil)
	if err == nil {
		t.Fatal("expected promotion to refuse an abandoned draft")
	}
	if code := model.CodeOf(err); code != model.ExitConflict {
		t.Errorf("exit = %d, want ExitConflict (%d)", code, model.ExitConflict)
	}
	if !strings.Contains(err.Error(), "'abandoned'") {
		t.Errorf("refusal does not name the status found: %v", err)
	}
}

// TestAbandonedDraftIsNotListed is the other half of R-8.3 that this package
// owns: `today --team` stops speaking about the draft (R-9.2).
func TestAbandonedDraftIsNotListed(t *testing.T) {
	ws := fixture(t)
	id, _ := draftFor(t, ws, "payments")
	if _, _, err := DraftsSection(ws, "core", 1); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := DraftsSection(ws, "core", 1); !ok {
		t.Fatal("an open draft is not listed")
	}
	if _, err := PRDAbandon(ws, "core", id, nil); err != nil {
		t.Fatal(err)
	}
	_, ok, err := DraftsSection(ws, "core", 1)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("an abandoned draft is still listed by today --team")
	}
}

// setStatus rewrites just the `status:` line, leaving the rest of the artifact
// alone — the cheapest way to fabricate a draft in a state no command produces.
func setStatus(t *testing.T, path, status string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := strings.Replace(string(data), "status: draft\n", "status: "+status+"\n", 1)
	if out == string(data) {
		t.Fatalf("no `status: draft` line to replace in %s", path)
	}
	if err := os.WriteFile(path, []byte(out), 0o666); err != nil {
		t.Fatal(err)
	}
}

func first(s string) string {
	if i := strings.Index(s, "\n"); i >= 0 && i < 120 {
		return s[:i]
	}
	if len(s) > 120 {
		return s[:120]
	}
	return s
}
