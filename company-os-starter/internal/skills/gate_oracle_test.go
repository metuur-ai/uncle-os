package skills_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Gate 7 has an oracle for exactly three shapes: the two clean lines in the
// passing goldens and the two-failure block in
// examples/failing-workspace-golden-validate.txt. Everything else about
// skill_conflicts and the two counts — id reuse across differing file names, a
// canonical TEAM skill, a personal rule colliding with one, an id-less pair,
// multiple canonical skills matched by a single offender — is unmeasured by any
// committed fixture.
//
// These tests build those workspaces and read gate 7 back out of the reference
// CLI's own `validate` output, so the comparison is against Python rather than
// against a re-reading of Python.

// pythonGate7 runs the reference `validate` and returns its gate-7 block:
// the header line plus every line up to the blank one that opens gate 8 or the
// trailer. Exit status is ignored — a synthesized workspace fails other gates
// by construction, and a fixture that fails identically under both
// implementations is still passing evidence.
func pythonGate7(t *testing.T, root, wsPath string) string {
	t.Helper()
	cli := filepath.Join(root, "company-os-starter", "bin", "company-os")
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("reference CLI not present: %v", err)
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; the reference oracle cannot run")
	}
	cmd := exec.Command("python3", cli, "--root", wsPath, "validate")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	_ = cmd.Run()

	var block []string
	in := false
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(line, "[7/") {
			in = true
		} else if in && line == "" {
			break
		}
		if in {
			block = append(block, line)
		}
	}
	if len(block) == 0 {
		t.Fatalf("reference validate produced no gate 7 block:\n%s", out.String())
	}
	return strings.Join(block, "\n") + "\n"
}

// gateMatchesReference is the assertion every case below shares.
func gateMatchesReference(t *testing.T, ws *workspace.Workspace) {
	t.Helper()
	want := pythonGate7(t, repoRoot(t), ws.Root)
	g, err := skills.Gate(ws, 7)
	if err != nil {
		t.Fatalf("Gate: %v", err)
	}
	if got := renderGate(g, 7); got != want {
		t.Errorf("gate 7 diverges from the reference CLI\n--- got ---\n%s--- want ---\n%s",
			got, want)
	}
}

// TestGateMatchesReferenceOnSynthesizedWorkspaces walks the conflict and count
// shapes the committed corpus never reaches.
func TestGateMatchesReferenceOnSynthesizedWorkspaces(t *testing.T) {
	cases := []struct {
		name  string
		build func(t *testing.T, ws *workspace.Workspace)
	}{
		{"clean four layers", func(*testing.T, *workspace.Workspace) {}},
		{"team skill reuses a canonical name", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "creating-prd.SKILL.md"),
				"skill://team/creating-prd-copy", "team", "", "")
		}},
		{"team skill reuses a canonical id under another name", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "elsewhere.SKILL.md"),
				"skill://product/creating-prd", "team", "", "")
		}},
		// One offender colliding with TWO canonical skills emits two findings,
		// ordered by the canonical list, not by the offender.
		{"one offender shadows two canonical skills", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "company-os", "skills", "shared.SKILL.md"),
				"skill://company/shared", "", "", "")
			mkskill(t, filepath.Join(ws.Root, "platforms", "communications", "skills", "shared.SKILL.md"),
				"skill://product/shared", "", "", "")
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "shared.SKILL.md"),
				"skill://team/shared", "team", "", "")
		}},
		// A personal rule is scanned for shadowing even though it never counts.
		{"personal rule shadows a canonical name", func(t *testing.T, ws *workspace.Workspace) {
			write(t, filepath.Join(ws.Root, "teams", "core", "scratchpad", "personal-rules", "creating-prd.md"),
				"---\ntype: personal-rule\n---\n\n# mine\n")
		}},
		// authority decides the canonical count, layer decides the team count.
		{"canonical team skill counts as canonical", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "team-canon.SKILL.md"),
				"skill://team/team-canon", "canonical", "", "")
		}},
		// ...and a canonical TEAM skill is still not shadowable.
		{"canonical team skill cannot be shadowed", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "team-canon.SKILL.md"),
				"skill://team/team-canon", "canonical", "", "")
			write(t, filepath.Join(ws.Root, "teams", "core", "scratchpad", "personal-rules", "team-canon.md"),
				"---\ntype: personal-rule\n---\n\n# mine\n")
		}},
		{"id-less skills do not shadow each other", func(t *testing.T, ws *workspace.Workspace) {
			write(t, filepath.Join(ws.Root, "company-os", "skills", "alpha.SKILL.md"),
				"---\ntype: skill\nauthority: canonical\n---\n\n# alpha\n")
			write(t, filepath.Join(ws.Root, "teams", "core", "skills", "beta.SKILL.md"),
				"---\ntype: skill\nauthority: team\n---\n\n# beta\n")
		}},
		{"dangling extends", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "broken.SKILL.md"),
				"skill://team/broken", "team", "platform-skill://communications/nonexistent", "")
		}},
		{"malformed extends URI is dangling", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "broken.SKILL.md"),
				"skill://team/broken", "team", "not-a-uri", "")
		}},
		{"valid extends is clean", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "creating-prd-mobile.SKILL.md"),
				"skill://team/creating-prd-mobile", "team",
				"platform-skill://communications/creating-prd", "1. (default) add mobile mock")
		}},
		// Shadowing and dangling extends together, to freeze the emission order
		// across the two loops rather than only within one.
		{"shadowing and dangling extends together", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "broken.SKILL.md"),
				"skill://team/broken", "team", "platform-skill://communications/nonexistent", "")
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "creating-prd.SKILL.md"),
				"skill://team/creating-prd-copy", "team", "", "")
		}},
		// `s["id"] == k["id"]` is Python's `==`, which spans the numeric tower:
		// 5 == 5.0 and True == 1 are conflicts even though str() renders them
		// differently. No committed golden carries this shape, and Value.Equal
		// compared kind-and-text, so gate 7 stayed CLEAN where Python reports
		// [FAIL].
		{"int id shadowed by the same value as a float", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "company-os", "skills", "numeric.SKILL.md"),
				"5", "", "", "")
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "numeric-copy.SKILL.md"),
				"5.0", "team", "", "")
		}},
		{"int id shadowed by the same value as a bool", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "company-os", "skills", "one.SKILL.md"),
				"1", "", "", "")
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "one-copy.SKILL.md"),
				"true", "team", "", "")
		}},
		// ...and the negative: a str "5" is not the int 5, which is what the
		// kind test was there for and must survive the fix.
		{"string id does not shadow the same digits as an int", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "company-os", "skills", "digits.SKILL.md"),
				"5", "", "", "")
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "digits-copy.SKILL.md"),
				"'5'", "team", "", "")
		}},
		// A personal rule named exactly `.md` has the stem ".md" under pathlib,
		// not "" — and "" would collide with every other empty-named skill.
		{"personal rule named exactly .md", func(t *testing.T, ws *workspace.Workspace) {
			write(t, filepath.Join(ws.Root, "teams", "core", "scratchpad", "personal-rules", ".md"),
				"---\ntype: personal-rule\n---\n\n# dotfile\n")
			write(t, filepath.Join(ws.Root, "teams", "core", "scratchpad", "personal-rules", "other.md"),
				"---\ntype: personal-rule\n---\n\n# other\n")
		}},
		// A company-layer skill that is NOT canonical protects nothing.
		{"non-canonical company skill is not shadowable", func(t *testing.T, ws *workspace.Workspace) {
			mkskill(t, filepath.Join(ws.Root, "company-os", "skills", "draft.SKILL.md"),
				"skill://company/draft", "guidance", "", "")
			mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "draft.SKILL.md"),
				"skill://team/draft", "team", "", "")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := fourLayerWorkspace(t)
			tc.build(t, ws)
			gateMatchesReference(t, ws)
		})
	}
}
