package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// The corpus is deliberately weighted to the shapes the scaffolding commands
// actually emit, plus the four decisions that are easy to get wrong and
// invisible until a file-tree diff:
//
//   - '1.0' and '2026.2' are STRINGS that the implicit resolver would reclaim as
//     floats, so safe_dump single-quotes them. Emitting them plain would make
//     every scaffolded platform.yaml and company.yaml read back wrong.
//   - a block sequence that is a mapping VALUE is indentless (`tags:\n- x`),
//     which is the single largest visual difference from yaml.v3.
//   - the 85-column `precedence:` line in a scaffolded team.yaml is NOT folded,
//     because no space on it sits past column 80.
//   - empty collections go to flow style: `requirements: []`, `components: {}`.
var emitCorpus = []struct{ name, yaml string }{
	{"company", `
schemaVersion: '1.0'
kind: Company
metadata:
  id: acme-inc
  name: Acme Inc.
tags:
- kind/company
`},
	{"baseline-78-column-plain", `
controls:
- id: security-baseline
  version: '1.0'
  level: default
  requirement: Services authenticate inbound calls and encrypt data in transit.
`},
	{"team-85-column-unfolded", `
agentSkills:
  canonicalPath: skills/
  precedence: canonical-mandatory > personal > canonical-default > canonical-guidance
  onConflict: prefer-canonical-and-inform-user
`},
	{"empty-collections", `
requirements: []
components: {}
platform:
  id: my-platform
`},
	{"registry-flow-in-block-out", `
schemaVersion: '1.0'
kind: IdRegistry
ids:
- {id: 'platform://communications', definedIn: platforms/communications/platform.yaml}
- {id: 'component://customer-notification-service', definedIn: platforms/communications/components/customer-notification-service.yaml}
tags: [ontology/registry]
`},
	{"scalars-needing-quotes", `
a: '1.0'
b: '2026.2'
c: 'yes'
d: 'null'
e: ''
f: '2035-01-15'
g: '0755'
h: '12:30'
`},
	{"scalars-staying-plain", `
a: platform://communications
b: repo://svc
c: skills/
d: scratchpad/personal-rules/
e: canonical-mandatory > personal
f: Acme Inc.
g: team://TODO
`},
	{"non-string-scalars", `
a: true
b: false
c: 42
d: -7
e: 3.5
f: null
g: 2035-01-15
h: 1e30
`},
	{"indicator-scalars", `
a: '#hash'
b: '- dash'
c: 'key: value'
d: 'trailing '
e: ' leading'
f: '---fence'
g: "quote'inside"
h: '@at'
`},
	{"nested-and-long", `
platformRelationships:
- platform: platform://communications
  relationship: belongs-to
deep:
  one:
    two:
      three: a much longer sentence that runs past the eightieth column so the fold is exercised here
`},
	{"unicode-forces-double-quotes", `
name: Café Ñandú
`},
	{"long-single-quoted", `
value: '1.0 aaaaaaaa bbbbbbbb cccccccc dddddddd eeeeeeee ffffffff gggggggg hhhhhhhh iiiiiiii'
`},
}

// TestEmitterMatchesPyYAML re-runs yaml.safe_dump over the same documents and
// asserts byte equality. It is the only thing standing between this port and a
// silently different file tree: the differential harness compares file BYTES,
// and every YAML artifact the scaffolding commands write is produced here.
//
// It follows the oracle pattern of internal/frontmatter and internal/workspace —
// skip, never pass, when python3 or the vendored PyYAML is unavailable.
func TestEmitterMatchesPyYAML(t *testing.T) {
	env := oracleEnv(t)
	for _, c := range emitCorpus {
		t.Run(c.name, func(t *testing.T) {
			src := strings.TrimLeft(c.yaml, "\n")
			path := filepath.Join(t.TempDir(), "doc.yaml")
			if err := os.WriteFile(path, []byte(src), 0o666); err != nil {
				t.Fatal(err)
			}
			loaded, err := loadPy(path)
			if err != nil {
				t.Fatalf("loadPy: %v", err)
			}
			got, err := pyDump(loaded)
			if err != nil {
				t.Fatalf("pyDump: %v", err)
			}
			want := safeDump(t, env, src)
			if got != want {
				t.Fatalf("emitter diverged from safe_dump\n--- python\n%s--- go\n%s", want, got)
			}
		})
	}
}

// TestEmitterIsAFixedPoint is what `graph build; graph build` and repeated
// `add` runs depend on: re-dumping an emitter's own output must not change it.
func TestEmitterIsAFixedPoint(t *testing.T) {
	for _, c := range emitCorpus {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			first := filepath.Join(dir, "a.yaml")
			if err := os.WriteFile(first, []byte(strings.TrimLeft(c.yaml, "\n")), 0o666); err != nil {
				t.Fatal(err)
			}
			loaded, err := loadPy(first)
			if err != nil {
				t.Fatalf("loadPy: %v", err)
			}
			once, err := pyDump(loaded)
			if err != nil {
				t.Fatalf("pyDump: %v", err)
			}
			second := filepath.Join(dir, "b.yaml")
			if err := os.WriteFile(second, []byte(once), 0o666); err != nil {
				t.Fatal(err)
			}
			reloaded, err := loadPy(second)
			if err != nil {
				t.Fatalf("loadPy (round 2): %v", err)
			}
			twice, err := pyDump(reloaded)
			if err != nil {
				t.Fatalf("pyDump (round 2): %v", err)
			}
			if once != twice {
				t.Fatalf("not a fixed point\n--- first\n%s--- second\n%s", once, twice)
			}
		})
	}
}

// TestScaffoldedArtifactsMatchPyYAML closes the loop on the documents nobody
// authored: it re-dumps what `init` and `add` actually wrote and compares to
// safe_dump. A regression in a scaffold dict — a reordered key, a value that
// stopped being a string — fails here rather than in the harness.
func TestScaffoldedArtifactsMatchPyYAML(t *testing.T) {
	env := oracleEnv(t)
	root := initWorkspace(t)
	if _, err := Add(workspace.New(root), AddComponent, "billing-api", "platform-1", nil); err != nil {
		t.Fatalf("add component: %v", err)
	}
	for _, rel := range []string{
		"company-os/company.yaml",
		"company-os/standards/company-baseline.yaml",
		"platforms/platform-1/platform.yaml",
		"platforms/platform-1/governance/requirements.yaml",
		"platforms/platform-1/components/billing-api.yaml",
		"teams/core/team.yaml",
		"company-ontology/ids/registry.yaml",
	} {
		t.Run(rel, func(t *testing.T) {
			src := read(t, filepath.Join(root, filepath.FromSlash(rel)))
			if want := safeDump(t, env, src); src != want {
				t.Fatalf("not what safe_dump would write\n--- python\n%s--- go\n%s", want, src)
			}
		})
	}
}

// oracleEnv locates the vendored PyYAML and skips when it or python3 is absent.
func oracleEnv(t *testing.T) []string {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; cannot re-run the oracle")
	}
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Skipf("cannot locate the module root: %v", err)
	}
	vendor := filepath.Join(moduleRoot, "vendor")
	if _, err := os.Stat(filepath.Join(vendor, "yaml")); err != nil {
		t.Skipf("vendored PyYAML unavailable: %v", err)
	}
	return append(os.Environ(), "PYTHONPATH="+vendor)
}

const safeDumpScript = `
import sys, yaml
sys.stdout.write(yaml.safe_dump(yaml.safe_load(sys.stdin.read()),
                                sort_keys=False, default_flow_style=False))
`

func safeDump(t *testing.T, env []string, src string) string {
	t.Helper()
	cmd := exec.Command("python3", "-c", safeDumpScript)
	cmd.Env = env
	cmd.Stdin = strings.NewReader(src)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("safe_dump oracle failed: %v\n%s", err, errb.String())
	}
	return out.String()
}
