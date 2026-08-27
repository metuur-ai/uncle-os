package workspace

// Inherited from examples/selftest.py (task 6.1). ST-011..ST-013, `:87-93`.
//
// TestIsRoot already covers the predicate against synthetic directories,
// including the empty-dir case (ST-011) and a `teams/`-only dir (ST-013's
// shape). What it does not cover — and what selftest.py did — is the predicate
// against the two committed fixtures. That matters because the fixtures are what
// every other harness (acceptance.sh, the goldens, the differential corpus)
// resolves through: if IsRoot ever stopped recognising examples/workspace, every
// one of those would start skipping rather than failing.

import (
	"os"
	"path/filepath"
	"testing"
)

// fixtureDir resolves examples/<name>, skipping when the test binary is running
// outside a checkout (R-6.7 — the binary must not need files beside it).
func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", name))
	if err != nil {
		t.Fatalf("resolving examples/%s: %v", name, err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	return dir
}

// TestIsRootFullWorkspaceFixture is selftest.py:89-90 (ST-012).
func TestIsRootFullWorkspaceFixture(t *testing.T) {
	dir := fixtureDir(t, "workspace")
	if !New(dir).IsRoot() {
		t.Fatalf("%s is not recognised as a workspace root", dir)
	}
}

// TestIsRootStandaloneTeamFixture was selftest.py:91-92 (ST-013): any ONE
// canonical root suffices, demonstrated by a teams/-only fixture. Task 4.1
// (docs/tasks/ux-simplification.md) rebuilt examples/standalone-team into a
// real, fully-validating minimal workspace — one company, one platform, one
// team — so it is no longer teams/-only; the single-root shape ST-013 wanted
// is still covered directly by TestIsRoot's synthetic "canonical dir" case.
// What remains worth guarding here is the same thing TestIsRootFullWorkspaceFixture
// guards for examples/workspace: this committed fixture, whatever its current
// shape, must keep resolving as a workspace root, or `company-os today`/`next`
// would refuse to run against the project's own one-team onboarding example.
func TestIsRootStandaloneTeamFixture(t *testing.T) {
	dir := fixtureDir(t, "standalone-team")
	if !New(dir).IsRoot() {
		t.Fatalf("%s is not recognised as a workspace root", dir)
	}
}
