package skills_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Shared fixture helpers for this package's tests.
//
// This file used to hold four differential tests — TestListMatchesReferenceCLI
// and its WithPersonalRules / OnResolvedExtends / OnPathlibStemEdgeCases
// variants — which ran company-os-starter/bin/company-os over the committed
// fixtures and compared `skills list` byte for byte at the library seam. R-9.3
// deleted that binary, so all four could only SKIP: 12 subtests reporting green
// while asserting nothing.
//
// They were removed rather than frozen because internal/difftest covers the same
// ground end to end — skills/list-<fixture> over all six committed workspaces
// plus the three failure-path ones, and skills/list-after-scratchpad for the
// populated personal-rules layer, each freezing stdout AND a hash of every file
// in the tree.
//
// One thing did NOT survive by that argument and was frozen instead:
// gate_oracle_test.go's 17 synthesized workspaces (skill shadowing, dangling
// extends, id-type collisions) exist in no committed fixture, so difftest cannot
// reach them. Its reference answers were recovered from tag python-cli-final and
// live in testdata/gate7_reference.json.
//
// The pathlib-stem edge case the fourth test covered — `Path(".md").stem` is
// ".md", not "" — is still pinned, by the "personal rule named exactly .md" case
// in gateCases.

// fixtures are every workspace root under examples/ that the difftest corpus
// exercises. Kept because skills_test.go walks them.
var fixtures = []string{
	"workspace",
	"standalone-team",
	"federated",
	"banking/small-company",
	"banking/bank/workspaces/team-payments-rails",
	"banking/bank/workspaces/team-fraud-detection",
	"failing-workspace",
	"failing-federated",
	"failing-federated-nolock",
}

func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return abs
}

func listText(t *testing.T, ws *workspace.Workspace) string {
	t.Helper()
	sections, err := skills.List(ws)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var b bytes.Buffer
	if err := render.Skills(&b, sections); err != nil {
		t.Fatalf("render.Skills: %v", err)
	}
	return b.String()
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
