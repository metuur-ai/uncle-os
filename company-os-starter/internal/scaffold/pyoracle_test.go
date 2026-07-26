package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// TestScaffoldedArtifactsMatchPyYAML closes the loop on the documents nobody
// authored: it re-dumps what `init` and `add` actually wrote and compares to
// safe_dump. A regression in a scaffold dict — a reordered key, a value that
// stopped being a string — fails here rather than in the harness.
//
// The emitter's own corpus lives with the emitter, in internal/yamlio; this test
// stays here because only this package knows what the scaffolds contain.
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
