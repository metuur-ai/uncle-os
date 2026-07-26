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

// TestIsRootStandaloneTeamFixture is selftest.py:91-92 (ST-013): any ONE
// canonical root suffices. examples/standalone-team has teams/ and nothing else,
// which is the shape a team gets before it has a platform to belong to — if this
// regressed, `company-os today` would refuse to run for exactly the users least
// able to diagnose it.
func TestIsRootStandaloneTeamFixture(t *testing.T) {
	dir := fixtureDir(t, "standalone-team")
	// Guard the guard: the assertion below is only meaningful while the fixture
	// really is teams/-only.
	for _, other := range []string{"company-os", "platforms", "company-ontology", KnowledgeRoot} {
		if _, err := os.Stat(filepath.Join(dir, other)); err == nil {
			t.Fatalf("%s grew a %s/ root; it no longer proves that one root suffices",
				dir, other)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "teams")); err != nil {
		t.Fatalf("%s has no teams/ dir: %v", dir, err)
	}
	if !New(dir).IsRoot() {
		t.Fatalf("%s is not recognised as a workspace root", dir)
	}
}
