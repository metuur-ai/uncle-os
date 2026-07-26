package render_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/ids"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/roles"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// The expectations below are the Python CLI's stdout for examples/workspace,
// captured verbatim. examples/differential.py compares the two implementations
// on 47 invocations of these commands; this test is the fast guard that a
// records refactor cannot change the bytes without someone noticing before the
// harness runs.

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

func TestIDs_ReproducesPythonStdout(t *testing.T) {
	ws := exampleWorkspace(t)
	const listing = "canonical IDs (company-ontology/ids/registry.yaml):\n" +
		"  platform://communications  ->  platforms/communications/platform.yaml\n" +
		"  team://customer-engagement  ->  teams/customer-engagement/team.yaml\n" +
		"  component://customer-notification-service  ->  " +
		"platforms/communications/components/customer-notification-service.yaml\n" +
		"  capability://communications/message-delivery  ->  " +
		"company-ontology/concepts/capability--message-delivery.md\n" +
		"  context://communications  ->  company-ontology/contexts/communications.md\n" +
		"  req://communications/delivery-reliability  ->  " +
		"platforms/communications/governance/requirements.yaml\n" +
		"  req://communications/message-schema  ->  " +
		"platforms/communications/governance/requirements.yaml\n" +
		"7 id(s)\n"

	cases := []struct {
		name   string
		role   string
		filter ids.Filter
		want   string
	}{
		{"unfiltered", "", ids.Filter{}, listing},
		{"role-team-lead", "team-lead", ids.Filter{},
			"terms (plain-language, display-only — canonical terms are unchanged " +
				"in artifacts, IDs, tags, and validation):\n" +
				"  deviation — \"documented exception to a default rule\"\n" +
				"  exception — \"approved, expiring waiver of a mandatory rule\"\n" +
				"  governance — \"the requirements that apply to your components\"\n" +
				listing},
		// An unmapped role prints no legend and no error (GPF-R-3.3).
		{"role-unknown", "wizard", ids.Filter{}, listing},
		{"filter-team", "", ids.Filter{Team: "customer-engagement"},
			"canonical IDs (company-ontology/ids/registry.yaml):\n" +
				"  team://customer-engagement  ->  teams/customer-engagement/team.yaml\n" +
				"1 id(s) of 7\n"},
		// A filter that matches nothing still prints the header and the tally.
		{"filter-nomatch", "", ids.Filter{Prefix: "zzzz"},
			"canonical IDs (company-ontology/ids/registry.yaml):\n0 id(s) of 7\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sections, err := ids.List(ws, tc.role, tc.filter)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			var buf bytes.Buffer
			if err := render.IDs(&buf, sections); err != nil {
				t.Fatalf("IDs: %v", err)
			}
			if buf.String() != tc.want {
				t.Errorf("stdout mismatch\n got:\n%s\nwant:\n%s", buf.String(), tc.want)
			}
		})
	}
}

func TestToday_ReproducesPythonStdout(t *testing.T) {
	ws := exampleWorkspace(t)
	const devBlock = "\nteam customer-engagement " +
		"(governance generated 2026-07-18T14:39:06Z)\n" +
		"  - customer-notification-service: 3 platform requirement(s), " +
		"3 company control(s)\n"
	const poGlossary = "terms (plain-language, display-only — canonical terms are " +
		"unchanged in artifacts, IDs, tags, and validation):\n" +
		"  exception — \"promise with an expiry date\"\n" +
		"  deviation — \"documented exception to a default rule\"\n" +
		"  PRD — \"the plan for a change (product requirements)\"\n" +
		"  outcome review — \"the scheduled check on whether the change worked\"\n"
	const poBlock = "\nplatform communications: 0 active PRD(s)\n" +
		"  - outcome review due 2026-10-16: 2026-per-channel-quiet-hours\n"

	cases := []struct{ role, want string }{
		{"developer", "== today (developer) ==\n" + devBlock +
			"\nonboarding: teams/customer-engagement/onboarding/developer.md\n"},
		{"team-lead", "== today (team-lead) ==\n" +
			"terms (plain-language, display-only — canonical terms are unchanged " +
			"in artifacts, IDs, tags, and validation):\n" +
			"  deviation — \"documented exception to a default rule\"\n" +
			"  exception — \"approved, expiring waiver of a mandatory rule\"\n" +
			"  governance — \"the requirements that apply to your components\"\n" +
			devBlock},
		{"vp-engineering", "== today (vp-engineering) ==\n" + devBlock},
		{"product-owner", "== today (product-owner) ==\n" + poGlossary + poBlock},
		{"director-of-product", "== today (director-of-product) ==\n" + poGlossary + poBlock},
		// architect is in neither role set and has no onboarding guide here, so
		// the banner is the whole view.
		{"architect", "== today (architect) ==\n"},
	}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			sections, err := roles.Today(ws, tc.role)
			if err != nil {
				t.Fatalf("Today: %v", err)
			}
			var buf bytes.Buffer
			if err := render.Today(&buf, sections); err != nil {
				t.Fatalf("Today: %v", err)
			}
			if buf.String() != tc.want {
				t.Errorf("stdout mismatch\n got:\n%s\nwant:\n%s", buf.String(), tc.want)
			}
		})
	}
}

// TestToday_MissingGovernanceLine pins the warn at bin/company-os:1192, which
// warn() writes to STDOUT with a two-space indent and no leading blank line —
// the one shape in this view that is not preceded by one.
func TestToday_MissingGovernanceLine(t *testing.T) {
	ws := workspace.New(t.TempDir())
	if err := os.MkdirAll(filepath.Join(ws.Teams, "solo"), 0o755); err != nil {
		t.Fatal(err)
	}
	sections, err := roles.Today(ws, "developer")
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	var buf bytes.Buffer
	if err := render.Today(&buf, sections); err != nil {
		t.Fatalf("Today: %v", err)
	}
	want := "== today (developer) ==\n" +
		"  [warn] solo: no effective-governance.yaml — run governance resolve\n"
	if buf.String() != want {
		t.Errorf("stdout mismatch\n got:\n%s\nwant:\n%s", buf.String(), want)
	}
}
