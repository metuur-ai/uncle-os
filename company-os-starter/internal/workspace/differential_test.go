package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDifferentialAgainstPythonOracle re-runs the Python CLI over the same three
// failure paths this package ports and asserts the stderr it produces is
// byte-identical to the errors returned here. It follows the oracle pattern
// established by internal/frontmatter: when python3 or the CLI is unavailable it
// skips rather than passes, so a missing oracle can never look like agreement.
//
// The frozen literals in workspace_test.go are what CI asserts on every machine;
// this test is what stops them drifting away from the reference implementation
// while one still exists.
func TestDifferentialAgainstPythonOracle(t *testing.T) {
	cli, env := oracle(t)

	// Python's die() writes "error: <msg>\n"; the Go error carries <msg>.
	strip := func(s string) string {
		return strings.TrimSuffix(strings.TrimPrefix(s, "error: "), "\n")
	}

	t.Run("require_root", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "not-a-workspace")
		ws := New(dir)
		got := ws.RequireRoot()
		if got == nil {
			t.Fatal("RequireRoot() = nil")
		}
		want := strip(runOracle(t, cli, env, "--root", dir, "validate"))
		if got.Error() != want {
			t.Fatalf("diverged from Python\n go: %q\npy: %q", got.Error(), want)
		}
	})

	root := makeRoot(t)
	ws := New(root)

	t.Run("platform_dir", func(t *testing.T) {
		_, got := ws.PlatformDir("ghost")
		if got == nil {
			t.Fatal("PlatformDir(ghost) = nil")
		}
		want := strip(runOracle(t, cli, env,
			"--root", root, "prd", "validate", "--platform", "ghost", "some-prd"))
		if got.Error() != want {
			t.Fatalf("diverged from Python\n go: %q\npy: %q", got.Error(), want)
		}
	})

	t.Run("team_dir", func(t *testing.T) {
		_, got := ws.TeamDir("ghost")
		if got == nil {
			t.Fatal("TeamDir(ghost) = nil")
		}
		want := strip(runOracle(t, cli, env,
			"--root", root, "governance", "resolve", "--team", "ghost"))
		if got.Error() != want {
			t.Fatalf("diverged from Python\n go: %q\npy: %q", got.Error(), want)
		}
	})
}

// oracle locates bin/company-os and the vendored PyYAML relative to this
// package, and returns the environment the CLI needs. It skips the test when
// anything is missing.
func oracle(t *testing.T) (cli string, env []string) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; cannot re-run the oracle")
	}
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Skipf("cannot locate the module root: %v", err)
	}
	cli = filepath.Join(moduleRoot, "bin", "company-os")
	vendor := filepath.Join(moduleRoot, "vendor")
	for _, p := range []string{cli, vendor} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("oracle unavailable: %v", err)
		}
	}
	return cli, append(os.Environ(), "PYTHONPATH="+vendor)
}

// runOracle runs the Python CLI and returns its stderr. The CLI is expected to
// fail here; a run that succeeds means the fixture no longer reaches the branch
// under test, which is a test bug rather than a pass.
func runOracle(t *testing.T, cli string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("python3", append([]string{cli}, args...)...)
	cmd.Env = env
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		t.Fatalf("oracle %v succeeded; it no longer reaches the failure path", args)
	}
	if stderr.Len() == 0 {
		t.Fatalf("oracle %v produced no stderr", args)
	}
	return stderr.String()
}
