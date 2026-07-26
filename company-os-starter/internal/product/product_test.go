package product

// The product cluster's unit tests.
//
// The differential harness is the byte-level oracle for everything here, so this
// file deliberately does NOT re-assert output shapes it already covers. What it
// covers instead is the three things the harness cannot see:
//
//   - the two inherited selftest assertions, ST-034 and ST-035, which need a
//     TEAM OVERRIDE TEMPLATE and therefore a fixture no corpus workspace has;
//   - R-1.14, the OKF v0.2 Phase 0 done-gate date fix, which no fixture
//     exercises because every `updated:`/`created:` in the corpus is a
//     well-formed ISO date — measured, not assumed: see TestParseDate's header;
//   - the str.format subset, whose failure modes are unreachable without an
//     override template.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// ------------------------------------------------------------- fixtures

// fixture builds the smallest workspace the product commands need: one team with
// an ownership registry and one platform with a component descriptor, so
// governance resolves to something non-empty.
func fixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "company-os", "standards", "company-baseline.yaml"),
		"schemaVersion: '1.0'\ncontrols: []\n")
	write(t, filepath.Join(root, "platforms", "payments", "platform.yaml"),
		"schemaVersion: '1.0'\nid: platform://payments\n")
	write(t, filepath.Join(root, "platforms", "payments", "components", "checkout-service.yaml"),
		"schemaVersion: '1.0'\nid: component://checkout-service\n"+
			"metadata:\n  name: Checkout Service\n"+
			"ownership:\n  accountableTeam: core\n"+
			"relationships:\n  - platform: payments\n    type: belongs-to\n")
	write(t, filepath.Join(root, "platforms", "payments", "governance", "requirements.yaml"),
		"schemaVersion: '1.0'\nrequirements: []\n")
	write(t, filepath.Join(root, "teams", "core", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://core\n")
	write(t, filepath.Join(root, "teams", "core", "ownership", "components.yaml"),
		"schemaVersion: '1.0'\ncomponents:\n  - id: checkout-service\n    role: accountable\n")
	return workspace.New(root)
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

func messages(sections []model.GateResult, sev model.Severity) []string {
	var out []string
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Severity == sev {
				out = append(out, f.Message)
			}
		}
	}
	return out
}

// -------------------------------------------- ST-034 / ST-035 (GPF-R-4.4)

// customPRDTemplate is examples/selftest.py:200-209's override verbatim in
// intent: it omits `## {ps[1]}` (Success metrics) and keeps ps[0] and ps[2].
// Everything else, including the frontmatter the CLI owns, is present so that
// the ONLY reason validate can fail is the missing heading.
const customPRDTemplate = "---\ntype: prd\nid: {pid}\ntitle: {title}\nstatus: proposed\n" +
	"team: {team}\nplatform: {platform}\ncomponents: [{components}]\n" +
	"governanceSnapshot: {date}\ndecisionOwner: {title}-owner\ncreated: {date}\n" +
	"fromDiscovery: {discovery}\n" +
	"tags: [kind/prd, platform/{platform}, team/{team}, status/proposed]\n---\n\n" +
	"# PRD: {title}\n\n## {ps[0]}\nP\n\n## {ps[2]}\nC\n\n" +
	"## Affected components\n{component_list}\n\n" +
	"## Applicable governance (snapshot {date})\n{governance_checklist}\n"

// TestPRDValidateCustomTemplateMissingHeadingFails is ST-034: a PRD scaffolded
// from a team override that omits a required heading FAILS validate. Outputs are
// validated; templates are not (GPF-R-4.4). The heading check is deliberately
// independent of the team's format-enforcement opt-in, which this fixture does
// not set.
//
// TestPRDValidateNamesMissingHeading is ST-035, folded in: the same run must
// name the exact heading, because "your PRD is invalid" without the heading name
// is not actionable.
func TestPRDValidateCustomTemplateMissingHeadingFails(t *testing.T) {
	ws := fixture(t)
	write(t, filepath.Join(ws.Teams, "core", "templates", "prd.md"), customPRDTemplate)

	if _, err := PRDNew(ws, "core", "payments", "checkout-service", "Broken", ""); err != nil {
		t.Fatalf("prd new: %v", err)
	}
	id := time.Now().Format("2006") + "-broken"
	sections, err := PRDValidate(ws, "payments", id)
	if err != nil {
		t.Fatalf("prd validate: %v", err)
	}
	if !model.HasFailure(sections) {
		t.Fatalf("validate passed a PRD missing a required heading; findings=%v",
			messages(sections, model.SevOK))
	}
	fails := strings.Join(messages(sections, model.SevFail), "\n")
	if !strings.Contains(fails, "Success metrics") {
		t.Errorf("failure does not name the missing heading:\n%s", fails)
	}
}

// TestDiscoveryTemplateRendersEverySection is ST-016 for the discovery template
// and ST-017 for the PRD one: every name in the *_SECTIONS list has to reach the
// rendered document as a literal `## ` heading, because that list is also what
// `validate` greps. Losing this coupling is the failure mode where a renamed
// heading passes scaffolding and fails validation.
func TestDiscoveryTemplateRendersEverySection(t *testing.T) {
	discovery, err := formatTemplate(scaffold.DiscoveryTemplate, "test", map[string]any{
		"bid": "b", "title": "T", "team": "core", "date": "2026-01-01",
		"ds": DiscoverySections,
	})
	if err != nil {
		t.Fatalf("format discovery: %v", err)
	}
	for _, s := range DiscoverySections {
		if !strings.Contains(discovery, "\n## "+s+"\n") {
			t.Errorf("discovery template has no `## %s` heading", s)
		}
	}
	prd, err := formatTemplate(scaffold.PRDTemplate, "test", map[string]any{
		"pid": "p", "title": "T", "team": "core", "platform": "payments",
		"components": "c", "date": "2026-01-01", "discovery": "none",
		"problem": "P", "metrics": "M", "ps": PRDSections,
		"component_list": "- `c`", "governance_checklist": "- [ ] none resolved",
	})
	if err != nil {
		t.Fatalf("format prd: %v", err)
	}
	for _, s := range PRDSections {
		if !strings.Contains(prd, "\n## "+s+"\n") {
			t.Errorf("PRD template has no `## %s` heading", s)
		}
	}
}

// -------------------------------------------------- R-1.14 (OKF Phase 0)

// TestParseDate pins the done-gate's date comparison, which is the one place
// this port deliberately does not reproduce the oracle.
//
// It needs a unit test because the differential harness cannot reach it: every
// `updated:` and `created:` in every corpus fixture is a well-formed ISO date
// (measured — 6 reality docs, 6 PRDs), so every harness invocation exercises
// only the path where string comparison and date comparison agree. The bug lives
// entirely in the inputs no fixture supplies.
func TestParseDate(t *testing.T) {
	good := []string{"2026-07-18", "2026-07-18T10:30:00Z", "2026-07-18T10:30:00-05:00"}
	for _, s := range good {
		if _, ok := parseDate(s); !ok {
			t.Errorf("parseDate(%q) refused a well-formed date", s)
		}
	}
	// Each of these passed the oracle's string comparison silently: "18/07/2026"
	// sorts before every string starting with "2", so a reality doc dated in
	// dd/mm/yyyy was ALWAYS reported stale; "None" sorts after, so an absent
	// `updated:` read as up to date under some orderings.
	bad := []string{"18/07/2026", "", "None", "2026-13-45", "yesterday", "2026/07/18"}
	for _, s := range bad {
		if _, ok := parseDate(s); ok {
			t.Errorf("parseDate(%q) accepted a value that is not an ISO-8601 date", s)
		}
	}
}

// TestDoneGateNamesAMalformedDate is OKF v0.2 R-0.2: a malformed `updated:`
// produces a done-check error naming the file and the offending value, never an
// unhandled exception and never a silent pass.
//
// The oracle's answer to this input is `"18/07/2026" < "2026-01-01"` -> True ->
// "not updated since PRD created", a sentence that is accidentally right for the
// wrong reason and would be accidentally WRONG the moment the PRD's own created
// date sorted the other way.
func TestDoneGateNamesAMalformedDate(t *testing.T) {
	ws := fixture(t)
	seedActivePRD(t, ws, "2026-thing", "created: 2026-01-01", "updated: 18/07/2026")

	sections, err := PRDComplete(ws, "payments", "2026-thing", false, nil)
	if err != ErrDoneCheck {
		t.Fatalf("done-gate error = %v, want ErrDoneCheck", err)
	}
	if got := model.CodeOf(err); got != model.ExitPrecondition {
		t.Fatalf("exit code = %d, want %d", got, model.ExitPrecondition)
	}
	fails := strings.Join(messages(sections, model.SevFail), "\n")
	if !strings.Contains(fails, "18/07/2026") {
		t.Errorf("refusal does not name the offending value:\n%s", fails)
	}
	if !strings.Contains(fails, "reality/components/checkout-service.md") {
		t.Errorf("refusal does not name the file:\n%s", fails)
	}
}

// TestDoneGateAcceptsAFreshRealityDoc is OKF v0.2 R-0.3 from the accepting side:
// well-formed input yields the same outcome as the string comparison did. The
// reality doc is newer than the PRD, so the only remaining problem is the
// unchecked checklist item, and removing that lets the archive run.
func TestDoneGateAcceptsAFreshRealityDoc(t *testing.T) {
	ws := fixture(t)
	seedActivePRD(t, ws, "2026-fresh", "created: 2026-01-01", "updated: 2026-06-01")

	sections, err := PRDComplete(ws, "payments", "2026-fresh", false, nil)
	if err != nil {
		t.Fatalf("done-gate refused a fresh reality doc: %v (%v)",
			err, messages(sections, model.SevFail))
	}
	archived := filepath.Join(ws.Platforms, "payments", "archive", "prds", "2026-fresh")
	for _, name := range []string{"prd.md", "outcome.md"} {
		if _, err := os.Stat(filepath.Join(archived, name)); err != nil {
			t.Errorf("archive is missing %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(ws.Platforms, "payments", "log.md")); err != nil {
		t.Errorf("log.md was not appended: %v", err)
	}
}

// TestDoneGateStaleRealityRefuses is R-0.3 from the refusing side, with the
// oracle's exact sentence.
func TestDoneGateStaleRealityRefuses(t *testing.T) {
	ws := fixture(t)
	seedActivePRD(t, ws, "2026-stale", "created: 2026-06-01", "updated: 2026-01-01")

	sections, err := PRDComplete(ws, "payments", "2026-stale", false, nil)
	if err != ErrDoneCheck {
		t.Fatalf("done-gate error = %v, want ErrDoneCheck", err)
	}
	want := "reality doc for 'checkout-service' not updated since PRD created " +
		"(reality updated: 2026-01-01)"
	if fails := messages(sections, model.SevFail); !contains(fails, want) {
		t.Errorf("refusal = %v, want one line %q", fails, want)
	}
}

// TestDoneGateForceArchivesAnyway pins --force: the refusal is skipped, nothing
// is printed about it, and the archive happens.
func TestDoneGateForceArchivesAnyway(t *testing.T) {
	ws := fixture(t)
	seedActivePRD(t, ws, "2026-forced", "created: 2026-06-01", "updated: 2026-01-01")

	sections, err := PRDComplete(ws, "payments", "2026-forced", true, nil)
	if err != nil {
		t.Fatalf("--force still refused: %v", err)
	}
	if fails := messages(sections, model.SevFail); len(fails) > 0 {
		t.Errorf("--force printed refusals: %v", fails)
	}
	body, err := os.ReadFile(filepath.Join(ws.Platforms, "payments",
		"archive", "prds", "2026-forced", "prd.md"))
	if err != nil {
		t.Fatalf("archived prd.md: %v", err)
	}
	if !strings.Contains(string(body), "status: completed") {
		t.Errorf("archived PRD status was not rewritten:\n%s", body)
	}
}

// TestShutilMoveNestsIntoAnExistingDirectory is the destination rule that a
// plain os.Rename gets wrong, and that `prd/full-lifecycle-force` caught: a
// second completion of the same id does not fail and does not overwrite — it
// lands NESTED, exactly as shutil.move does.
func TestShutilMoveNestsIntoAnExistingDirectory(t *testing.T) {
	ws := fixture(t)
	seedActivePRD(t, ws, "2026-twice", "created: 2026-06-01", "updated: 2026-07-01")
	if _, err := PRDComplete(ws, "payments", "2026-twice", true, nil); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	seedActivePRD(t, ws, "2026-twice", "created: 2026-06-01", "updated: 2026-07-01")
	if _, err := PRDComplete(ws, "payments", "2026-twice", true, nil); err != nil {
		t.Fatalf("second complete: %v", err)
	}
	nested := filepath.Join(ws.Platforms, "payments", "archive", "prds",
		"2026-twice", "2026-twice", "prd.md")
	if _, err := os.Stat(nested); err != nil {
		t.Errorf("second archive did not nest into the existing directory: %v", err)
	}
}

// ------------------------------------------------------- str.format subset

// TestFormatTemplateRefusesWhatItCannotRender covers the subset boundary. A
// partial implementation that silently dropped `{title!r}` or `{date:%Y}` would
// write a MALFORMED artifact from a valid override; refusing by name is what
// keeps a template mistake visible.
func TestFormatTemplateRefusesWhatItCannotRender(t *testing.T) {
	args := map[string]any{"title": "T", "ps": []string{"A", "B"}}
	for _, tmpl := range []string{
		"{title!r}", "{date:%Y}", "{missing}", "{ps}", "{ps[9]}", "{unclosed", "}",
	} {
		if _, err := formatTemplate(tmpl, "test", args); err == nil {
			t.Errorf("formatTemplate(%q) accepted an unsupported field", tmpl)
		}
	}
	got, err := formatTemplate("{{{title}}} {ps[1]}", "test", args)
	if err != nil {
		t.Fatalf("formatTemplate: %v", err)
	}
	if got != "{T} B" {
		t.Errorf("formatTemplate = %q, want %q", got, "{T} B")
	}
}

// ------------------------------------------------------------- helpers

// seedActivePRD writes an active change record and its component's reality doc,
// with the two dates under test. The PRD carries exactly one unchecked checklist
// item so the caller can also observe the checklist half of the done-gate.
func seedActivePRD(t *testing.T, ws *workspace.Workspace, id, created, updated string) {
	t.Helper()
	write(t, filepath.Join(ws.Platforms, "payments", "change-records", "active", id, "prd.md"),
		"---\ntype: prd\nid: "+id+"\ntitle: Thing\nstatus: proposed\nteam: core\n"+
			"platform: payments\ncomponents: [checkout-service]\n"+
			"governanceSnapshot: 2026-01-01\ndecisionOwner: someone\n"+created+"\n---\n\n"+
			"# PRD: Thing\n\n## Applicable governance\n- [x] company: c v1 (mandatory)\n")
	write(t, filepath.Join(ws.Platforms, "payments", "reality", "components", "checkout-service.md"),
		"---\ntype: component-reality\nid: reality-checkout-service\n"+updated+"\n---\n\n"+
			"# Checkout Service\n")
}

func contains(items []string, want string) bool {
	for _, s := range items {
		if s == want {
			return true
		}
	}
	return false
}
