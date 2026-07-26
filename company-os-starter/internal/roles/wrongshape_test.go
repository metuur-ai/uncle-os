package roles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmd_today reads effective-governance.yaml and every outcome.md with no schema
// in between: `eff.get("components", {}).items()`, `e["requirements"]["platform"]`
// and `len(v)` are all applied to whatever the file happens to hold. The port
// answered each of them with a Go type switch that narrowed the accepted shapes,
// which produced two different failures at once — an exit 0 where Python raises,
// and a silently WRONG count where both exit 0.
//
// Measured with the vendored PyYAML 6.0.2 (`today --role developer`):
//
//	components: null            → AttributeError: 'NoneType' has no 'items', exit 1
//	components: [a]             → AttributeError: 'list' has no 'items',     exit 1
//	c1: {}                      → KeyError: 'requirements',                  exit 1
//	platform: null              → AttributeError: 'NoneType' has no 'values', exit 1
//	platform: {p1: 7}           → TypeError: object of type 'int' has no len, exit 1
//	platform: {p1: abcd}        → exit 0, "4 platform requirement(s)"
//	company: abc                → exit 0, "3 company control(s)"
//
// R-0.7a(j) carves out the traceback and the exit code, not the outcome: a
// non-zero exit for the first five, the right number for the last two.
func writeGov(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, "teams", "core", "generated")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "effective-governance.yaml"),
		[]byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

func govWorkspace(t *testing.T) (*workspace.Workspace, string) {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{
		filepath.Join("company-os"),
		filepath.Join("platforms", "p1"),
		filepath.Join("teams", "core"),
		filepath.Join("company-ontology"),
	} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o777); err != nil {
			t.Fatal(err)
		}
	}
	return workspace.New(root), root
}

func TestTodayRefusesAWrongShapedGovernanceFile(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"components-null", "generatedAt: 2026-01-01\ncomponents: null\n"},
		{"components-sequence", "generatedAt: 2026-01-01\ncomponents:\n- a\n"},
		{"components-scalar", "generatedAt: 2026-01-01\ncomponents: nope\n"},
		{"component-not-a-mapping", "generatedAt: 2026-01-01\ncomponents:\n  c1: 7\n"},
		{"requirements-missing", "generatedAt: 2026-01-01\ncomponents:\n  c1: {}\n"},
		{"requirements-null", "generatedAt: 2026-01-01\ncomponents:\n  c1:\n    requirements: null\n"},
		{"platform-missing", "generatedAt: 2026-01-01\ncomponents:\n  c1:\n    requirements:\n      company: []\n"},
		{"platform-null", "generatedAt: 2026-01-01\ncomponents:\n  c1:\n    requirements:\n      platform: null\n      company: []\n"},
		{"platform-value-has-no-length", "generatedAt: 2026-01-01\ncomponents:\n  c1:\n    requirements:\n      platform:\n        p1: 7\n      company: []\n"},
		{"company-missing", "generatedAt: 2026-01-01\ncomponents:\n  c1:\n    requirements:\n      platform: {}\n"},
		{"company-has-no-length", "generatedAt: 2026-01-01\ncomponents:\n  c1:\n    requirements:\n      platform: {}\n      company: 7\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws, root := govWorkspace(t)
			writeGov(t, root, c.body)

			_, err := Today(ws, "developer")
			if err == nil {
				t.Fatal("expected a refusal; Python raises here and exits 1")
			}
			if code := model.CodeOf(err); code != model.ExitArtifact {
				t.Errorf("exit code = %d, want %d (R-0.7a(j))", code, model.ExitArtifact)
			}
		})
	}
}

// TestTodayCountsLenTheWayPythonDoes is the both-exit-0 half: `len(v)` is not
// "length of a sequence". Narrowing it reported 0 where Python reports 4, on a
// line a human reads as a fact about their governance.
func TestTodayCountsLenTheWayPythonDoes(t *testing.T) {
	cases := []struct {
		name             string
		body             string
		platform, compan int
	}{
		{"sequences", "components:\n  c1:\n    requirements:\n      platform:\n        p1: [r1, r2]\n      company: [x]\n", 2, 1},
		{"mapping-valued-platform", "components:\n  c1:\n    requirements:\n      platform:\n        p1: {a: 1, b: 2, c: 3}\n      company: []\n", 3, 0},
		{"string-valued-platform", "components:\n  c1:\n    requirements:\n      platform:\n        p1: abcd\n      company: []\n", 4, 0},
		{"string-valued-company", "components:\n  c1:\n    requirements:\n      platform: {}\n      company: abc\n", 0, 3},
		// len() counts CHARACTERS, so a two-byte rune is one.
		{"non-ascii-string", "components:\n  c1:\n    requirements:\n      platform:\n        p1: éé\n      company: []\n", 2, 0},
		{"two-platforms", "components:\n  c1:\n    requirements:\n      platform:\n        p1: [a]\n        p2: bcd\n      company: []\n", 4, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws, root := govWorkspace(t)
			writeGov(t, root, "generatedAt: 2026-01-01\n"+c.body)

			sections, err := Today(ws, "developer")
			if err != nil {
				t.Fatalf("Today: %v", err)
			}
			f := findFinding(t, sections, model.CodeComponent)
			if got := f.Fields["platformRequirements"]; got != c.platform {
				t.Errorf("platformRequirements = %v, want %d", got, c.platform)
			}
			if got := f.Fields["companyControls"]; got != c.compan {
				t.Errorf("companyControls = %v, want %d", got, c.compan)
			}
		})
	}
}

// TestTodayOutcomeDueRendersLikePython is P6. `{m.get('due')}` has no default,
// so an absent key f-strings as "None" and a container f-strings as its repr.
func TestTodayOutcomeDueRendersLikePython(t *testing.T) {
	cases := []struct {
		name, frontmatter, want string
	}{
		{"absent", "status: pending\n", "None"},
		{"explicit-null", "status: pending\ndue: null\n", "None"},
		{"date", "status: pending\ndue: 2026-10-16\n", "2026-10-16"},
		{"mapping", "status: pending\ndue: {a: 1, b: [2, 3]}\n", "{'a': 1, 'b': [2, 3]}"},
		{"sequence", "status: pending\ndue: [x, y]\n", "['x', 'y']"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws, root := govWorkspace(t)
			dir := filepath.Join(root, "platforms", "p1", "archive", "prds", "old-1")
			if err := os.MkdirAll(dir, 0o777); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "outcome.md"),
				[]byte("---\n"+c.frontmatter+"---\nbody\n"), 0o666); err != nil {
				t.Fatal(err)
			}

			sections, err := Today(ws, "product-owner")
			if err != nil {
				t.Fatalf("Today: %v", err)
			}
			f := findFinding(t, sections, model.CodeOutcomeReview)
			if got := f.Fields["due"]; got != c.want {
				t.Errorf("due = %q, want %q", got, c.want)
			}
		})
	}
}

// TestTodayPRDFieldsRenderContainers covers the same fallback at the active-PRD
// row, where a container previously collapsed to the "?" default.
func TestTodayPRDFieldsRenderContainers(t *testing.T) {
	ws, root := govWorkspace(t)
	dir := filepath.Join(root, "platforms", "p1", "change-records", "active", "act-1")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prd.md"),
		[]byte("---\nstatus: {a: 1}\nteam: [x, y]\n---\nbody\n"), 0o666); err != nil {
		t.Fatal(err)
	}

	sections, err := Today(ws, "product-owner")
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	f := findFinding(t, sections, model.CodeActivePRD)
	if got := f.Fields["status"]; got != "{'a': 1}" {
		t.Errorf("status = %q, want %q", got, "{'a': 1}")
	}
	if got := f.Fields["team"]; got != "['x', 'y']" {
		t.Errorf("team = %q, want %q", got, "['x', 'y']")
	}
}

func findFinding(t *testing.T, sections []model.GateResult, code string) model.Finding {
	t.Helper()
	var seen []string
	for _, s := range sections {
		for _, f := range s.Findings {
			seen = append(seen, f.Code)
			if f.Code == code {
				return f
			}
		}
	}
	t.Fatalf("no %q finding; saw %s", code, strings.Join(seen, ", "))
	return model.Finding{}
}
