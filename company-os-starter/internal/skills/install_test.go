package skills_test

// `skills install`: the five dispositions and the version comparison that
// picks between them.
//
// The dispositions are the whole contract — an agent's next move is decided by
// which code came back — so each gets its own case against a real temp
// workspace rather than a mocked filesystem.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	skillfiles "github.com/metuur-ai/uncle-os/company-os-starter/skills"
)

// installFixture is an empty workspace root: `skills install` creates the
// destination itself, so nothing needs to exist first.
func installFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	return workspace.New(t.TempDir())
}

func outcomeFor(t *testing.T, res *skills.InstallResult, name string) skills.InstallOutcome {
	t.Helper()
	for _, o := range res.Outcomes {
		if o.Name == name {
			return o
		}
	}
	t.Fatalf("no outcome for %q", name)
	return skills.InstallOutcome{}
}

// anEmbeddedSkill returns one embedded skill to build fixtures around, so the
// tests never hard-code a skill name that a later rename would falsify.
func anEmbeddedSkill(t *testing.T) skillfiles.File {
	t.Helper()
	all, err := skillfiles.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no embedded skills")
	}
	return all[0]
}

// TestInstallWritesEverySkillIntoTheCompanyLayer is the fresh-install case: an
// empty workspace ends up holding every skill the binary carries, flat, where
// discovery looks.
func TestInstallWritesEverySkillIntoTheCompanyLayer(t *testing.T) {
	ws := installFixture(t)

	res, err := skills.Install(ws, nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	embedded, err := skillfiles.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Outcomes) != len(embedded) {
		t.Fatalf("installed %d skills, want %d", len(res.Outcomes), len(embedded))
	}
	for _, o := range res.Outcomes {
		if o.Code != model.CodeSkillsInstalled {
			t.Errorf("%s: code %q, want %q", o.Name, o.Code, model.CodeSkillsInstalled)
		}
	}

	// Discovery is the point of the placement, so assert it rather than the path.
	found, err := skills.Discover(ws)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(found) != len(embedded) {
		t.Fatalf("discovery found %d installed skills, want %d", len(found), len(embedded))
	}
}

// TestInstallIsIdempotent is the unchanged case: running twice changes nothing
// the second time. A command that reported work on every run would make "did
// anything change" unanswerable.
func TestInstallIsIdempotent(t *testing.T) {
	ws := installFixture(t)
	if _, err := skills.Install(ws, nil); err != nil {
		t.Fatalf("first Install: %v", err)
	}
	res, err := skills.Install(ws, nil)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	for _, o := range res.Outcomes {
		if o.Code != model.CodeSkillsUnchanged {
			t.Errorf("%s: second run reported %q, want %q", o.Name, o.Code, model.CodeSkillsUnchanged)
		}
	}
}

// writeInstalled puts a stand-in skill at the destination with a chosen version.
func writeInstalled(t *testing.T, ws *workspace.Workspace, name, version, marker string) string {
	t.Helper()
	dir := filepath.Join(ws.Root, "company-os", "skills")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+skills.Suffix)
	body := "---\nid: skill://product/" + name + "\ntype: skill\nversion: '" + version +
		"'\nauthority: canonical\ntags: [authority/canonical]\n---\n\n# " + marker + "\n\n" +
		"1. (mandatory) " + marker + "\n"
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestInstallUpdatesAnOlderSkill is the update case: a stale file is replaced
// and both versions are reported, so the change is auditable from the envelope.
func TestInstallUpdatesAnOlderSkill(t *testing.T) {
	ws := installFixture(t)
	target := anEmbeddedSkill(t)
	path := writeInstalled(t, ws, target.Name, "0.1", "STALE")

	res, err := skills.Install(ws, nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	got := outcomeFor(t, res, target.Name)
	if got.Code != model.CodeSkillsUpdated {
		t.Fatalf("code %q, want %q", got.Code, model.CodeSkillsUpdated)
	}
	if got.From != "0.1" || got.To != target.Version {
		t.Errorf("versions from=%q to=%q, want from=0.1 to=%s", got.From, got.To, target.Version)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "STALE") {
		t.Error("stale body survived the update")
	}
}

// TestInstallKeepsALocallyNewerSkill is the downgrade guard: a workspace ahead
// of the binary means a stale CLI, and the file is left exactly as it was.
func TestInstallKeepsALocallyNewerSkill(t *testing.T) {
	ws := installFixture(t)
	target := anEmbeddedSkill(t)
	path := writeInstalled(t, ws, target.Name, "99.0", "AHEAD")

	res, err := skills.Install(ws, nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if got := outcomeFor(t, res, target.Name); got.Code != model.CodeSkillsLocallyNewer {
		t.Fatalf("code %q, want %q", got.Code, model.CodeSkillsLocallyNewer)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "AHEAD") {
		t.Error("a newer local skill was overwritten — that is a downgrade")
	}
}

// TestInstallKeepsAnUnreadableSkill is the never-guess case: frontmatter that
// yields no version is a gate-4 problem for the user to see, not a file to
// silently replace.
func TestInstallKeepsAnUnreadableSkill(t *testing.T) {
	ws := installFixture(t)
	target := anEmbeddedSkill(t)
	dir := filepath.Join(ws.Root, "company-os", "skills")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, target.Name+skills.Suffix)
	if err := os.WriteFile(path, []byte("no frontmatter here\n"), 0o666); err != nil {
		t.Fatal(err)
	}

	res, err := skills.Install(ws, nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if got := outcomeFor(t, res, target.Name); got.Code != model.CodeSkillsUnreadable {
		t.Fatalf("code %q, want %q", got.Code, model.CodeSkillsUnreadable)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "no frontmatter here\n" {
		t.Error("an unreadable file was overwritten")
	}
}

// TestInstalledSkillsPassValidate is the guarantee that makes the command
// usable: what it writes leaves the workspace green, not in tag drift or
// missing a core field.
func TestInstalledSkillsPassValidate(t *testing.T) {
	ws := installFixture(t)
	if _, err := skills.Install(ws, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	assertSkillsGate4Clean(t, ws)
}

// TestInstallRebuildsDerivedArtifacts is the regression for the defect
// acceptance caught and the unit tests above did not: writing into
// company-os/skills/ changes a graph-docs root, so the company CLAUDE.md
// context node and the directory's index go stale and gate 5 fails. A
// previously-green workspace must not be reddened by a command that only added
// the files the CLI itself ships.
//
// The rebuild is asserted through its OUTPUT — the derived lines it returns —
// rather than by inspecting files, so any future derivation this command must
// trigger is covered without being named here.
func TestInstallRebuildsDerivedArtifacts(t *testing.T) {
	ws := installFixture(t)
	called := false
	rebuild := skills.Rebuild(func(got *workspace.Workspace) ([]string, error) {
		called = true
		if got.Root != ws.Root {
			t.Errorf("rebuild ran against %q, want %q", got.Root, ws.Root)
		}
		return []string{"derived: one line"}, nil
	})

	res, err := skills.Install(ws, rebuild)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !called {
		t.Fatal("install did not rebuild derived artifacts — gate 5 will fail on the " +
			"files it just wrote")
	}
	if len(res.Generated) != 1 || res.Generated[0] != "derived: one line" {
		t.Errorf("Generated = %v, want the rebuild's lines passed through", res.Generated)
	}

	// Idempotent runs rebuild too: a workspace whose derived artifacts were
	// already stale should not stay that way because nothing needed writing.
	called = false
	if _, err := skills.Install(ws, rebuild); err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !called {
		t.Error("a no-change install skipped the rebuild")
	}
}

// TestVersionComparisonIsNumericPerSegment fixes the ordering rule.
//
// The cases are derived from whatever version the embedded skill actually
// carries, so bumping a skill's version does not falsify the test. The third
// case is the one that matters: with an embedded 1.m where m is a single digit,
// an installed 1.10 is NEWER numerically (10 > m) but OLDER as a string
// ("1.10" < "1.3" — '1' sorts before '3' at the third character). A string
// comparison would call it stale and overwrite it, which is the downgrade this
// ordering exists to prevent. It arrives the first time a skill reaches its
// tenth revision.
func TestVersionComparisonIsNumericPerSegment(t *testing.T) {
	target := anEmbeddedSkill(t)
	maj, min, ok := twoSegments(target.Version)
	if !ok {
		t.Skipf("embedded version %q is not major.minor; ordering cases are not derivable",
			target.Version)
	}

	cases := []struct {
		installed string
		want      string
		why       string
	}{
		{fmt.Sprintf("%d.%d", maj, min), model.CodeSkillsUnchanged, "same version"},
		{fmt.Sprintf("%d.%d", maj+1, min), model.CodeSkillsLocallyNewer, "higher major"},
		{fmt.Sprintf("%d.%d", maj-1, min), model.CodeSkillsUpdated, "lower major"},
	}
	if maj >= 1 && min >= 2 && min <= 9 {
		cases = append(cases, struct {
			installed string
			want      string
			why       string
		}{
			fmt.Sprintf("%d.10", maj), model.CodeSkillsLocallyNewer,
			fmt.Sprintf("10 > %d numerically, though %d.10 < %d.%d as a string",
				min, maj, maj, min),
		})
	}

	for _, tc := range cases {
		// A fresh root per case: the disposition must depend on the installed
		// version alone, not on what a previous case left behind.
		ws := installFixture(t)
		writeInstalled(t, ws, target.Name, tc.installed, "MARK")
		res, err := skills.Install(ws, nil)
		if err != nil {
			t.Fatalf("Install: %v", err)
		}
		if got := outcomeFor(t, res, target.Name); got.Code != tc.want {
			t.Errorf("embedded v%s, installed v%s: code %q, want %q (%s)",
				target.Version, tc.installed, got.Code, tc.want, tc.why)
		}
	}
}

// twoSegments parses "major.minor", reporting false for any other shape.
func twoSegments(v string) (maj, min int, ok bool) {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return 0, 0, false
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	min, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}
	return maj, min, true
}
