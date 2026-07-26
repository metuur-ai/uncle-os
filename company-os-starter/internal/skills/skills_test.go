package skills_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// The 11 assertions this file carries over from examples/selftest.py:253-306 are
// ST-038 through ST-048 in .devlocal/go-port/selftest-inventory.md. They keep
// the same fixtures the Python ones build, so a behaviour that drifts fails on
// the same shape it always did.

const skillFM = "---\nid: %s\ntype: skill\nversion: '1.0'\nauthority: %s\n" +
	"tags: [authority/%s]\n%s---\n\n# %s\n\n%s\n"

// mkskill is selftest.py's _mkskill (:229-232).
func mkskill(t *testing.T, path, id, auth, extends, steps string) {
	t.Helper()
	if auth == "" {
		auth = "canonical"
	}
	if steps == "" {
		steps = "1. (mandatory) do it"
	}
	ext := ""
	if extends != "" {
		ext = "extends: " + extends + "\n"
	}
	write(t, path, fmt.Sprintf(skillFM, id, auth, auth, ext, id, steps))
}

// fourLayerWorkspace is the fixture selftest.py builds at :235-248: one skill
// per shared layer plus one personal rule.
func fourLayerWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "company-os"), 0o755); err != nil {
		t.Fatal(err)
	}
	mkskill(t, filepath.Join(root, "company-os", "skills", "governing.SKILL.md"),
		"skill://company/governing", "", "", "")
	mkskill(t, filepath.Join(root, "platforms", "communications", "skills", "creating-prd.SKILL.md"),
		"skill://product/creating-prd", "", "", "")
	mkskill(t, filepath.Join(root, "teams", "core", "skills", "team-extra.SKILL.md"),
		"skill://team/team-extra", "team", "", "")
	write(t, filepath.Join(root, "teams", "core", "scratchpad", "personal-rules", "maria.md"),
		"---\ntype: personal-rule\ntags: [authority/personal]\n---\n\n# Maria\n- my rule\n")
	return workspace.New(root)
}

func discover(t *testing.T, ws *workspace.Workspace) []skills.Skill {
	t.Helper()
	got, err := skills.Discover(ws)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return got
}

func conflicts(t *testing.T, ws *workspace.Workspace) []skills.Conflict {
	t.Helper()
	got, err := skills.Conflicts(ws, discover(t, ws))
	if err != nil {
		t.Fatalf("Conflicts: %v", err)
	}
	return got
}

func only(t *testing.T, all []skills.Skill, layer skills.Layer) skills.Skill {
	t.Helper()
	var out []skills.Skill
	for _, s := range all {
		if s.Layer == layer {
			out = append(out, s)
		}
	}
	if len(out) != 1 {
		t.Fatalf("want exactly one %s skill, got %d", layer, len(out))
	}
	return out[0]
}

// ST-038 (selftest.py:253) — GPF-R-5.1.
func TestDiscoverSkills_AllFourLayers(t *testing.T) {
	got := discover(t, fourLayerWorkspace(t))
	if len(got) != 4 {
		t.Fatalf("discovered %d skills, want 4", len(got))
	}
	seen := map[skills.Layer]bool{}
	for _, s := range got {
		seen[s.Layer] = true
	}
	for _, want := range skills.Layers {
		if !seen[want] {
			t.Errorf("layer %s was not discovered", want)
		}
	}
	if len(seen) != 4 {
		t.Errorf("discovered %d distinct layers, want 4", len(seen))
	}
}

// ST-039 (selftest.py:256) — GPF-R-5.1.
func TestDiscoverSkills_PlatformOriginLabeled(t *testing.T) {
	p := only(t, discover(t, fourLayerWorkspace(t)), skills.LayerPlatform)
	if p.Platform != "communications" {
		t.Errorf("Platform = %q, want communications", p.Platform)
	}
	if p.Name != "creating-prd" {
		t.Errorf("Name = %q, want creating-prd (the file stem, .SKILL stripped)", p.Name)
	}
}

// ST-040 (selftest.py:259) — GPF-R-5.1.
func TestDiscoverSkills_PersonalRuleNotCanonical(t *testing.T) {
	p := only(t, discover(t, fourLayerWorkspace(t)), skills.LayerPersonal)
	if p.Team != "core" {
		t.Errorf("Team = %q, want core", p.Team)
	}
	if p.IsCanonical() {
		t.Error("a personal rule must not carry canonical authority")
	}
	// A personal rule keeps its plain stem: no .SKILL marker to strip.
	if p.Name != "maria" {
		t.Errorf("Name = %q, want maria", p.Name)
	}
}

// ST-041 (selftest.py:265) — GPF-R-5.4. The merged view ranks canonical
// (company/platform/team) skills above personal rules, which render last.
func TestSkills_CanonicalPrecedePersonal(t *testing.T) {
	ws := fourLayerWorkspace(t)
	all := discover(t, ws)
	var shared, personal int
	for _, s := range all {
		if s.Layer == skills.LayerPersonal {
			personal++
			continue
		}
		shared++
	}
	if shared != 3 || personal != 1 {
		t.Fatalf("shared=%d personal=%d, want 3 and 1", shared, personal)
	}
	// The ordering claim is about the rendered view, so assert it there: every
	// shared skill's block must appear before the personal-rules section.
	out := listText(t, ws)
	section := strings.Index(out, "\n  personal rules (non-overriding")
	if section < 0 {
		t.Fatal("merged view has no personal-rules section")
	}
	for _, name := range []string{"governing [company", "creating-prd [platform", "team-extra [team"} {
		at := strings.Index(out, name)
		if at < 0 {
			t.Fatalf("merged view omits %q", name)
		}
		if at > section {
			t.Errorf("%q renders after the personal rules; canonical steps must rank above them", name)
		}
	}
}

// ST-042 (selftest.py:269) — GPF-R-5.2, absence/clean tolerant.
func TestSkillConflicts_CleanLayeringIsEmpty(t *testing.T) {
	if got := conflicts(t, fourLayerWorkspace(t)); len(got) != 0 {
		t.Errorf("clean layering reported %d conflict(s): %+v", len(got), got)
	}
	// And a workspace with no skills at all is equally clean.
	empty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(empty, "company-os"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := conflicts(t, workspace.New(empty)); len(got) != 0 {
		t.Errorf("a workspace with no skills reported %d conflict(s)", len(got))
	}
}

// ST-043 (selftest.py:275) and ST-044 (:276) — GPF-R-5.2. A team skill reusing
// a canonical FILE NAME shadows it even though its own id differs, and the
// record names both sides.
func TestSkillConflicts_DetectsShadowingAndNamesBothFiles(t *testing.T) {
	ws := fourLayerWorkspace(t)
	mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "creating-prd.SKILL.md"),
		"skill://team/creating-prd-copy", "team", "", "")

	got := conflicts(t, ws)
	var shadowing []skills.Conflict
	for _, c := range got {
		if c.Kind == skills.ConflictShadowing {
			shadowing = append(shadowing, c)
		}
	}
	if len(shadowing) != 1 {
		t.Fatalf("got %d shadowing conflicts, want 1: %+v", len(shadowing), got)
	}
	c := shadowing[0]
	if c.Skill != "teams/core/skills/creating-prd.SKILL.md" {
		t.Errorf("Skill = %q", c.Skill)
	}
	if c.Shadows != "platforms/communications/skills/creating-prd.SKILL.md" {
		t.Errorf("Shadows = %q", c.Shadows)
	}
	// The ids differ, so the name is what collided.
	if c.Reason != skills.ReasonName || c.ReasonValue != "creating-prd" {
		t.Errorf("Reason = %q/%q, want name/creating-prd", c.Reason, c.ReasonValue)
	}
	msg := c.Finding().Message
	for _, want := range []string{c.Skill, c.Shadows} {
		if !strings.Contains(msg, want) {
			t.Errorf("message omits %q: %s", want, msg)
		}
	}
}

// ST-045 (selftest.py:289) and ST-046 (:291) — GPF-R-5.3.
func TestResolveExtends(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "company-os"), 0o755); err != nil {
		t.Fatal(err)
	}
	mkskill(t, filepath.Join(root, "platforms", "communications", "skills", "creating-prd.SKILL.md"),
		"skill://product/creating-prd", "", "", "")
	ws := workspace.New(root)

	base, found, err := skills.ResolveExtends(ws, uri("platform-skill://communications/creating-prd"))
	if err != nil {
		t.Fatalf("ResolveExtends: %v", err)
	}
	if !found {
		t.Fatal("a platform-skill:// URI must resolve to the platform-layer base skill file")
	}
	if base.Name != "creating-prd" {
		t.Errorf("base Name = %q, want creating-prd", base.Name)
	}
	// The URI addresses the FILE, so the base keeps its own semantic id.
	if base.ID.String() != "skill://product/creating-prd" {
		t.Errorf("base ID = %q", base.ID)
	}

	if _, found, err := skills.ResolveExtends(ws, uri("platform-skill://communications/ghost")); err != nil || found {
		t.Errorf("unknown target: found=%v err=%v, want false/nil", found, err)
	}
	// A malformed URI is the same answer as a missing file, never an error.
	for _, bad := range []string{"", "skill://product/creating-prd", "platform-skill://onlyone"} {
		if _, found, err := skills.ResolveExtends(ws, uri(bad)); err != nil || found {
			t.Errorf("ResolveExtends(%q): found=%v err=%v, want false/nil", bad, found, err)
		}
	}
}

// ST-047 (selftest.py:298) — GPF-R-5.3: extending is the sanctioned
// alternative to shadowing, so a distinctly-named extender is clean.
func TestSkillConflicts_ValidExtendsIsNotConflict(t *testing.T) {
	ws := extendsWorkspace(t)
	mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "creating-prd-mobile.SKILL.md"),
		"skill://team/creating-prd-mobile", "team",
		"platform-skill://communications/creating-prd", "1. (default) add mobile mock")
	if got := conflicts(t, ws); len(got) != 0 {
		t.Errorf("a valid extends reported %d conflict(s): %+v", len(got), got)
	}
}

// ST-048 (selftest.py:304) — GPF-R-5.3.
func TestSkillConflicts_DanglingExtendsNamesURI(t *testing.T) {
	ws := extendsWorkspace(t)
	mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "broken.SKILL.md"),
		"skill://team/broken", "team", "platform-skill://communications/nonexistent", "")

	got := conflicts(t, ws)
	if len(got) != 1 || got[0].Kind != skills.ConflictDanglingExtends {
		t.Fatalf("got %+v, want one dangling-extends conflict", got)
	}
	if got[0].Extends != "platform-skill://communications/nonexistent" {
		t.Errorf("Extends = %q", got[0].Extends)
	}
	if !strings.Contains(got[0].Finding().Message, got[0].Extends) {
		t.Errorf("message must name the unresolved URI: %s", got[0].Finding().Message)
	}
}

// extendsWorkspace is selftest.py's second fixture (:281-286): a platform base
// skill and nothing else.
func extendsWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "company-os"), 0o755); err != nil {
		t.Fatal(err)
	}
	mkskill(t, filepath.Join(root, "platforms", "communications", "skills", "creating-prd.SKILL.md"),
		"skill://product/creating-prd", "", "", "")
	return workspace.New(root)
}

// uri builds the Value a caller would have read from frontmatter. Going through
// a real file would work too but obscures which input is under test.
func uri(s string) skills.Value {
	return skills.NewValue(s)
}

// ---------------------------------------------------------------- gate 7

// renderGate is the minimal text renderer from internal/model/model_test.go,
// narrowed to one gate. It exists to prove the same thing task 1.7 proved for
// the whole report: that gate 7's golden lines come out of the records, with no
// out-of-band state and no per-code branch in the renderer.
func renderGate(g model.GateResult, denominator int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%d/%d] %s\n", g.Ordinal, denominator, g.Title)
	for _, f := range g.Findings {
		marker := map[model.Severity]string{
			model.SevOK: "ok", model.SevWarn: "warn", model.SevFail: "FAIL",
		}[f.Severity]
		line := f.Message
		if f.Subject != "" {
			line = f.Subject + ": " + f.Message
		}
		fmt.Fprintf(&b, "  [%s] %s\n", marker, line)
	}
	return b.String()
}

// goldenGate7 slices gate 7's header and findings out of a committed golden:
// the header line plus every line until the blank one that opens the next gate.
func goldenGate7(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "examples", name))
	if err != nil {
		t.Fatalf("reading golden: %v", err)
	}
	var out []string
	in := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "[7/") {
			in = true
		} else if in && line == "" {
			break
		}
		if in {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		t.Fatalf("%s has no gate 7 block", name)
	}
	return strings.Join(out, "\n") + "\n"
}

// TestGateReproducesGoldens is the acceptance check for gate 7: the same
// producer, run against three fixtures, reproduces all three committed shapes —
// two clean lines carrying DIFFERENT counts and the two-failure shape.
func TestGateReproducesGoldens(t *testing.T) {
	cases := []struct {
		fixture, golden string
		denominator     int
	}{
		{"workspace", "golden-validate.txt", 7},
		{"federated", "federated-golden-validate.txt", 8},
		{"failing-workspace", "failing-workspace-golden-validate.txt", 7},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			path := filepath.Join(repoRoot(t), "examples", tc.fixture)
			if _, err := os.Stat(path); err != nil {
				t.Skipf("fixture absent: %v", err)
			}
			g, err := skills.Gate(workspace.New(path), 7)
			if err != nil {
				t.Fatalf("Gate: %v", err)
			}
			got, want := renderGate(g, tc.denominator), goldenGate7(t, tc.golden)
			if got != want {
				t.Errorf("gate 7 does not reproduce %s\n--- got ---\n%s--- want ---\n%s",
					tc.golden, got, want)
			}
			if g.Slug != "skills-layering" {
				t.Errorf("Slug = %q", g.Slug)
			}
		})
	}
}

// TestGateCountsReachTheTextRendererThroughFields is the R-2.3 proof named in
// the task's acceptance: the two numbers in the clean line are ints in Fields,
// and the sentence is composed FROM them. Changing the field changes the text,
// which a pre-composed string could not do.
func TestGateCountsReachTheTextRendererThroughFields(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "workspace")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	g, err := skills.Gate(workspace.New(path), 7)
	if err != nil {
		t.Fatalf("Gate: %v", err)
	}
	f := g.Findings[0]
	if f.Code != model.CodeSkillsClean {
		t.Fatalf("Code = %q", f.Code)
	}
	// Ints, not strings: JSON must emit 2, not "2".
	if got, ok := f.Fields["canonical"].(int); !ok || got != 2 {
		t.Errorf("Fields[canonical] = %#v, want int 2", f.Fields["canonical"])
	}
	if got, ok := f.Fields["team"].(int); !ok || got != 0 {
		t.Errorf("Fields[team] = %#v, want int 0", f.Fields["team"])
	}
	// The counts are read back out of Fields to build the sentence.
	mutated := model.Fields{"canonical": 41, "team": 9}
	want := "skills layered cleanly (41 canonical, 9 team; no shadowing or dangling extends)"
	if got := skills.Message(model.CodeSkillsClean, mutated); got != want {
		t.Errorf("Message() = %q, want %q", got, want)
	}
}

// TestConflictRecordsCarryNoProse is R-2.8/R-2.12 stated as a check rather than
// as a comment: the detection layer emits facts, and every sentence in the gate
// is a pure function of (code, Fields).
func TestConflictRecordsCarryNoProse(t *testing.T) {
	path := filepath.Join(repoRoot(t), "examples", "failing-workspace")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	ws := workspace.New(path)
	got := conflicts(t, ws)
	if len(got) != 2 {
		t.Fatalf("got %d conflicts, want 2", len(got))
	}
	// Every string field is a path, an identity or an enum — never a fragment
	// of the rendered sentence.
	for _, c := range got {
		for name, v := range map[string]string{
			"Skill": c.Skill, "Shadows": c.Shadows,
			"ReasonValue": c.ReasonValue, "Extends": c.Extends,
		} {
			if strings.ContainsAny(v, " —`") {
				t.Errorf("%s field %s = %q looks like prose", c.Kind, name, v)
			}
		}
	}
	// Shadowing before dangling extends: the emission order is the oracle's.
	if got[0].Kind != skills.ConflictShadowing || got[1].Kind != skills.ConflictDanglingExtends {
		t.Errorf("order = %s, %s; want shadowing then dangling-extends", got[0].Kind, got[1].Kind)
	}
	// Rebuilding each finding's message from its own fields must be stable.
	for _, c := range got {
		f := c.Finding()
		if got := skills.Message(f.Code, f.Fields); got != f.Message {
			t.Errorf("Message(%s) is not a pure function of Fields:\n got %q\nwant %q",
				f.Code, got, f.Message)
		}
		if f.Subject != "" {
			t.Errorf("gate 7 uses no line prefix; Subject = %q", f.Subject)
		}
		if f.Severity != model.SevFail {
			t.Errorf("conflict severity = %v, want fail", f.Severity)
		}
	}
}

// TestCountIgnoresLayerForCanonicalAndPersonalForTeam freezes the two easily
// mis-ported halves of `:1085-1086`: the canonical count is NOT filtered by
// layer, and the personal layer contributes to neither number.
func TestCountIgnoresLayerForCanonicalAndPersonalForTeam(t *testing.T) {
	ws := fourLayerWorkspace(t)
	// A team skill declaring canonical authority. It counts as canonical even
	// though it is not eligible to be shadowed.
	mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "team-canon.SKILL.md"),
		"skill://team/team-canon", "canonical", "", "")

	c := skills.Count(discover(t, ws))
	// company + platform + the new team skill.
	if c.Canonical != 3 {
		t.Errorf("Canonical = %d, want 3 (authority, not layer, decides)", c.Canonical)
	}
	// team-extra + team-canon; the personal rule is not counted.
	if c.Team != 2 {
		t.Errorf("Team = %d, want 2", c.Team)
	}
	// A canonical TEAM skill cannot be shadowed, so the identically-named
	// personal-layer file below must not trip the gate.
	write(t, filepath.Join(ws.Root, "teams", "core", "scratchpad", "personal-rules", "team-canon.md"),
		"---\ntype: personal-rule\n---\n\n# mine\n")
	if got := conflicts(t, ws); len(got) != 0 {
		t.Errorf("shadowing a canonical TEAM skill must not be a conflict: %+v", got)
	}
}

// TestShadowingDetectsIDReuseAcrossDifferentNames covers the `id` half of
// `:852`, which the failing-workspace fixture exercises but the selftest
// fixtures do not: same id, different file name.
func TestShadowingDetectsIDReuseAcrossDifferentNames(t *testing.T) {
	ws := fourLayerWorkspace(t)
	mkskill(t, filepath.Join(ws.Root, "teams", "core", "skills", "differently-named.SKILL.md"),
		"skill://product/creating-prd", "team", "", "")
	got := conflicts(t, ws)
	if len(got) != 1 {
		t.Fatalf("got %+v, want one conflict", got)
	}
	if got[0].Reason != skills.ReasonID || got[0].ReasonValue != "skill://product/creating-prd" {
		t.Errorf("Reason = %q/%q, want id/skill://product/creating-prd",
			got[0].Reason, got[0].ReasonValue)
	}
}

// TestShadowingIgnoresAbsentIDs is the `s["id"] and` guard at `:852`: two
// skills that both omit `id` share the value None, and None == None. Without
// the truthiness test every id-less team skill would shadow every id-less
// canonical one.
func TestShadowingIgnoresAbsentIDs(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "company-os", "skills", "alpha.SKILL.md"),
		"---\ntype: skill\nauthority: canonical\n---\n\n# alpha\n")
	write(t, filepath.Join(root, "teams", "core", "skills", "beta.SKILL.md"),
		"---\ntype: skill\nauthority: team\n---\n\n# beta\n")
	if got := conflicts(t, workspace.New(root)); len(got) != 0 {
		t.Errorf("two id-less skills must not shadow each other: %+v", got)
	}
}

// ---------------------------------------------------------------- steps

// TestSteps freezes parse_skill_steps (`:820-823`), which decides what the
// merged view shows under each skill.
func TestSteps(t *testing.T) {
	cases := []struct {
		name, body string
		want       []string
	}{
		{"tier marker required", "1. do it\n", nil},
		{"marker must follow the number", "1. Write it. (mandatory)\n", nil},
		{"all three tiers", "1. (mandatory) a\n2. (default) b\n3. (guidance) c\n",
			[]string{"1. (mandatory) a", "2. (default) b", "3. (guidance) c"}},
		{"leading and trailing space is stripped", "   4. (default) b   \n",
			[]string{"4. (default) b"}},
		{"space between number and paren is optional", "5.(guidance) c\n",
			[]string{"5.(guidance) c"}},
		{"multi-digit ordinals", "10. (mandatory) x\n", []string{"10. (mandatory) x"}},
		{"unknown tier is not a step", "1. (optional) x\n", nil},
		{"only the head line is kept", "1. (mandatory) head\n   continuation\n",
			[]string{"1. (mandatory) head"}},
		{"empty body", "", nil},
		// Python's \s and \d are Unicode-wide and its splitlines() breaks on
		// more than \n; Go's regexp, strings.Split and strings.TrimSpace are all
		// narrower. Every expectation below was measured against CPython rather
		// than inferred from the pattern.
		{"form feed separates lines", "x\f2. (default) b\n", []string{"2. (default) b"}},
		{"vertical tab separates lines", "x\v2. (default) b\n", []string{"2. (default) b"}},
		{"file separator separates lines", "x\x1c3. (guidance) c\n", []string{"3. (guidance) c"}},
		{"interior no-break space matches but is not stripped",
			"1.\u00a0(mandatory) a\n", []string{"1.\u00a0(mandatory) a"}},
		{"leading no-break space is whitespace", " 1. (mandatory) a\n",
			[]string{"1. (mandatory) a"}},
		{"unicode digits count", "١. (mandatory) a\n", []string{"١. (mandatory) a"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := skills.Steps(tc.body)
			if len(got) != len(tc.want) {
				t.Fatalf("Steps() = %q, want %q", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("Steps()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestName freezes _skill_name (`:775-782`).
func TestName(t *testing.T) {
	cases := map[string]string{
		"/a/creating-prd.SKILL.md": "creating-prd",
		"/a/maria.md":              "maria",
		"/a/a.b.md":                "a.b",
		"/a/noext":                 "noext",
		"/a/.SKILL.md":             "",
	}
	for path, want := range cases {
		if got := skills.Name(path); got != want {
			t.Errorf("Name(%q) = %q, want %q", path, got, want)
		}
	}
}

// TestDiscoverIsAbsenceTolerant is the standalone-team case: no company-os/,
// no platforms/, no skills anywhere.
func TestDiscoverIsAbsenceTolerant(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "teams", "solo"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := discover(t, workspace.New(root))
	if len(got) != 0 {
		t.Fatalf("got %d skills, want 0", len(got))
	}
	g, err := skills.Gate(workspace.New(root), 7)
	if err != nil {
		t.Fatalf("Gate: %v", err)
	}
	if len(g.Findings) != 1 || g.Findings[0].Severity != model.SevOK {
		t.Fatalf("an empty workspace must pass gate 7: %+v", g.Findings)
	}
}
