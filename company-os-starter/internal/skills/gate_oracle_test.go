package skills_test

import (
	"encoding/json"
	"os"
	"path/filepath"
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
// These tests build those workspaces and compare gate 7 against the reference
// CLI's own `validate` output, so the comparison is against Python rather than
// against a re-reading of Python.
//
// That output used to be produced live, by running company-os-starter/bin/
// company-os on each synthesized workspace. R-9.3 deleted it, which left this
// file able only to SKIP — 17 subtests reporting green while asserting nothing.
// The reference was recovered from tag python-cli-final and its answers frozen
// into testdata/gate7_reference.json rather than let this coverage go: these 17
// workspaces are synthesized here and exist in no committed fixture, so no
// fixture-driven suite could reach them however broad it was.
//
// The tradeoff, stated rather than glossed: the answers can no longer be
// re-derived in-tree. What they still catch is gate 7 drifting away from the
// behaviour Python defined, which is what the file is for. Regenerating means
// checking out the tag; the capture procedure is recorded in
// docs/tasks/go-cli-tui-port.md task 6.10.

// frozenGate7 returns the reference gate-7 block for one case, by name.
func frozenGate7(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "gate7_reference.json"))
	if err != nil {
		t.Fatalf("reading the frozen gate-7 answers: %v", err)
	}
	var frozen struct {
		Provenance string            `json:"provenance"`
		Gate7      map[string]string `json:"gate7"`
	}
	if err := json.Unmarshal(raw, &frozen); err != nil {
		t.Fatalf("decoding the frozen gate-7 answers: %v", err)
	}
	want, ok := frozen.Gate7[name]
	if !ok {
		t.Fatalf("no frozen reference answer for case %q; a case was added without "+
			"capturing its answer, which would otherwise assert nothing", name)
	}
	return want
}

// gateMatchesReference is the assertion every case below shares.
func gateMatchesReference(t *testing.T, name string, ws *workspace.Workspace) {
	t.Helper()
	want := frozenGate7(t, name)
	g, err := skills.Gate(ws, 7)
	if err != nil {
		t.Fatalf("Gate: %v", err)
	}
	if got := renderGate(g, 7); got != want {
		t.Errorf("gate 7 diverges from the reference CLI\n--- got ---\n%s--- want ---\n%s",
			got, want)
	}
}

// gateCases are the conflict and count shapes the committed corpus never
// reaches. Package-level rather than inline so the capture harness that froze
// the reference answers walked exactly the same list the test asserts on.
var gateCases = []struct {
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

// TestGateMatchesReferenceOnSynthesizedWorkspaces walks the conflict and count
// shapes the committed corpus never reaches.
func TestGateMatchesReferenceOnSynthesizedWorkspaces(t *testing.T) {
	for _, tc := range gateCases {
		t.Run(tc.name, func(t *testing.T) {
			ws := fourLayerWorkspace(t)
			tc.build(t, ws)
			gateMatchesReference(t, tc.name, ws)
		})
	}
}

// TestEveryGateCaseHasAFrozenAnswer fails loudly if a case is added without
// capturing the reference answer for it. Without this the new case would call
// frozenGate7, miss, and the failure would read as a bug in gate 7 rather than
// as missing testdata.
func TestEveryGateCaseHasAFrozenAnswer(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "gate7_reference.json"))
	if err != nil {
		t.Fatalf("reading the frozen gate-7 answers: %v", err)
	}
	var frozen struct {
		Gate7 map[string]string `json:"gate7"`
	}
	if err := json.Unmarshal(raw, &frozen); err != nil {
		t.Fatal(err)
	}
	if len(frozen.Gate7) != len(gateCases) {
		t.Errorf("%d frozen answers for %d cases", len(frozen.Gate7), len(gateCases))
	}
	for _, tc := range gateCases {
		if _, ok := frozen.Gate7[tc.name]; !ok {
			t.Errorf("case %q has no frozen reference answer", tc.name)
		}
	}
}
