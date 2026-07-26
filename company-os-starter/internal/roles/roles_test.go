package roles_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/roles"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

func exampleWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", "workspace"))
	if err != nil {
		t.Fatalf("resolving fixture: %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	return workspace.New(root)
}

// TestRoleGlossaryLines_MappedRoleShowsCanonicalTerm ports selftest.py ST-036
// (:217): a mapped role yields at least two terms and one of them is the
// canonical word "exception", shown alongside its plain-language label.
func TestRoleGlossaryLines_MappedRoleShowsCanonicalTerm(t *testing.T) {
	terms := roles.Terms("product-owner")
	if len(terms) < 2 {
		t.Fatalf("got %d term(s), want at least 2", len(terms))
	}
	found := false
	for _, term := range terms {
		if term.Canonical == "exception" {
			found = true
			if term.Plain == "" || term.Plain == term.Canonical {
				t.Errorf("canonical term %q carries no distinct plain label", term.Canonical)
			}
		}
	}
	if !found {
		t.Errorf("glossary for product-owner does not mention the canonical term 'exception': %v", terms)
	}
}

// TestRoleGlossaryLines_UnmappedRoleIsEmpty ports ST-037 (:219): an unmapped
// role yields nothing, with no error (GPF-R-3.3).
func TestRoleGlossaryLines_UnmappedRoleIsEmpty(t *testing.T) {
	for _, role := range []string{"nobody-role", "developer", "architect", "vp-engineering", ""} {
		if got := roles.Terms(role); len(got) != 0 {
			t.Errorf("Terms(%q) = %v, want empty", role, got)
		}
		if _, ok := roles.GlossarySection(role, 1); ok {
			t.Errorf("GlossarySection(%q) produced a section, want none", role)
		}
	}
}

// TestGlossaryIsPureRead pins GPF-R-3.2: rendering a role view never mutates an
// artifact. The fixture's mtimes are unchanged after every mapped role is asked
// for its legend.
func TestGlossaryIsPureRead(t *testing.T) {
	ws := exampleWorkspace(t)
	before := snapshot(t, ws.Root)
	for _, role := range []string{"product-owner", "director-of-product", "team-lead"} {
		roles.Terms(role)
		roles.GlossarySection(role, 1)
	}
	if after := snapshot(t, ws.Root); after != before {
		t.Error("the glossary touched the workspace")
	}
}

// TestToday_ArchitectRendersHeaderOnly pins the role sets at `:1173` and `:1188`:
// architect is in neither, so its view is the banner and whatever onboarding
// guide exists — never a platform or team block.
func TestToday_ArchitectRendersHeaderOnly(t *testing.T) {
	sections, err := roles.Today(exampleWorkspace(t), "architect")
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	for _, s := range sections {
		if s.Slug == model.SlugPlatform || s.Slug == model.SlugTeam {
			t.Errorf("architect view contains a %q section", s.Slug)
		}
	}
	if len(sections) == 0 || sections[0].Slug != model.SlugHeader {
		t.Fatalf("first section is not the header: %v", sections)
	}
}

// TestToday_DeveloperCarriesGovernanceCounts pins the two counts at `:1196-1198`
// as ints, so the text renderer reads numbers rather than re-parsing a sentence
// (R-2.3), plus the per-platform split the sentence discards.
func TestToday_DeveloperCarriesGovernanceCounts(t *testing.T) {
	sections, err := roles.Today(exampleWorkspace(t), "developer")
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	f := findingByCode(t, sections, model.CodeComponent)
	if got := f.Fields.Str("component"); got != "customer-notification-service" {
		t.Fatalf("component = %q", got)
	}
	if got := f.Fields.Int("platformRequirements"); got != 3 {
		t.Errorf("platformRequirements = %d, want 3", got)
	}
	if got := f.Fields.Int("companyControls"); got != 3 {
		t.Errorf("companyControls = %d, want 3", got)
	}
	plats := f.Fields.Strs("platforms")
	if len(plats) != 1 || plats[0] != "communications" {
		t.Errorf("platforms = %v, want [communications]", plats)
	}
}

// TestToday_MissingGovernanceIsAWarning pins `:1191-1193`: an absent
// effective-governance.yaml warns and continues, and the warning does not make
// the command fail.
func TestToday_MissingGovernanceIsAWarning(t *testing.T) {
	ws := workspace.New(t.TempDir())
	if err := os.MkdirAll(filepath.Join(ws.Teams, "solo"), 0o755); err != nil {
		t.Fatal(err)
	}
	sections, err := roles.Today(ws, "developer")
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	f := findingByCode(t, sections, model.CodeGovernanceMissing)
	if f.Severity != model.SevWarn {
		t.Errorf("severity = %v, want warn", f.Severity)
	}
	if got := f.Fields.Str("team"); got != "solo" {
		t.Errorf("team = %q, want solo", got)
	}
	if model.HasFailure(sections) {
		t.Error("a missing governance file must not fail the command")
	}
}

// TestToday_ProductOwnerReadsPRDsAndOutcomes pins `:1174-1187`: the active count
// comes from the directory listing, and an archived outcome surfaces only while
// its status is pending.
func TestToday_ProductOwnerReadsPRDsAndOutcomes(t *testing.T) {
	sections, err := roles.Today(exampleWorkspace(t), "product-owner")
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	p := findingByCode(t, sections, model.CodePlatform)
	if got := p.Fields.Str("platform"); got != "communications" {
		t.Errorf("platform = %q", got)
	}
	if got := p.Fields.Int("activePRDs"); got != 0 {
		t.Errorf("activePRDs = %d, want 0", got)
	}
	o := findingByCode(t, sections, model.CodeOutcomeReview)
	if got := o.Fields.Str("prd"); got != "2026-per-channel-quiet-hours" {
		t.Errorf("prd = %q", got)
	}
	// An unquoted `due: 2026-10-16` is a PyYAML date; it must render as Python
	// renders it, not as a Go time.Time (R-1.6).
	if got := o.Fields.Str("due"); got != "2026-10-16" {
		t.Errorf("due = %q, want 2026-10-16", got)
	}
}

// TestOnboardingGuide_TeamBeforeCompany pins R-6.2's precedence and the absent
// case at `:1165`, which must be "no line" rather than an error.
func TestOnboardingGuide_TeamBeforeCompany(t *testing.T) {
	ws := workspace.New(t.TempDir())
	company := filepath.Join(ws.Company, "onboarding")
	team := filepath.Join(ws.Teams, "solo", "onboarding")
	for _, d := range []string{company, team} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(company, "developer.md"), []byte("c"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := roles.OnboardingGuide(ws, "architect"); ok {
		t.Error("architect has no guide, want no result")
	}
	got, ok := roles.OnboardingGuide(ws, "developer")
	if !ok || got != "company-os/onboarding/developer.md" {
		t.Fatalf("guide = %q (%v), want the company one", got, ok)
	}
	if err := os.WriteFile(filepath.Join(team, "developer.md"), []byte("t"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok = roles.OnboardingGuide(ws, "developer")
	if !ok || got != "teams/solo/onboarding/developer.md" {
		t.Fatalf("guide = %q (%v), want the team one to win", got, ok)
	}
	if strings.Contains(got, `\`) {
		t.Errorf("guide path is not POSIX-separated: %q", got)
	}
}

func findingByCode(t *testing.T, sections []model.GateResult, code string) model.Finding {
	t.Helper()
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Code == code {
				return f
			}
		}
	}
	t.Fatalf("no finding with code %q", code)
	return model.Finding{}
}

// snapshot is a cheap "did anything change" fingerprint: every path under root
// with its size and modification time.
func snapshot(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		b.WriteString(path)
		b.WriteString(info.ModTime().String())
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return b.String()
}
