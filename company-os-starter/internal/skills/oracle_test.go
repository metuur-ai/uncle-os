package skills_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// This file makes examples/differential.py's `skills` comparison at the library
// seam as well as at the process seam.
//
// The harness diffs two BINARIES over a corpus of invocations; the same
// comparison is made here directly against bin/company-os, over the same
// committed fixtures, so a divergence fails `go test` instead of waiting for
// someone to run the harness. It also reaches two shapes the corpus reaches
// only through another cluster's command: a populated personal-rules layer
// (differential.py runs `scratchpad init` first) and a RESOLVED `extends`,
// which no committed fixture has at all.

// fixtures are every workspace root under examples/ that the differential
// corpus lists (examples/differential.py:73-85).
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

// pythonSkillsList runs the reference CLI, skipping the whole comparison when
// python3 is unavailable rather than failing a build that has no oracle to
// measure against.
func pythonSkillsList(t *testing.T, root, wsPath string) string {
	t.Helper()
	cli := filepath.Join(root, "company-os-starter", "bin", "company-os")
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("reference CLI not present: %v", err)
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; the reference oracle cannot run")
	}
	cmd := exec.Command("python3", cli, "--root", wsPath, "skills", "list")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("reference CLI failed on %s: %v\n%s", wsPath, err, errb.String())
	}
	return out.String()
}

// listText is the command end to end: the records internal/skills produces, put
// through the renderer cmd/company-os hands them to. Testing anything less than
// the pair would leave the seam between them unmeasured.
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

// TestListMatchesReferenceCLI is the byte-for-byte parity check for
// `skills list` (task 2.2 acceptance, R-1.1).
func TestListMatchesReferenceCLI(t *testing.T) {
	root := repoRoot(t)
	for _, fx := range fixtures {
		t.Run(fx, func(t *testing.T) {
			wsPath := filepath.Join(root, "examples", filepath.FromSlash(fx))
			if _, err := os.Stat(wsPath); err != nil {
				t.Skipf("fixture absent: %v", err)
			}
			want := pythonSkillsList(t, root, wsPath)

			if got := listText(t, workspace.New(wsPath)); got != want {
				t.Errorf("skills list diverges from the reference CLI\n%s",
					firstDiff(got, want))
			}
		})
	}
}

// TestListMatchesReferenceCLIWithPersonalRules covers the one corpus entry the
// committed fixtures cannot: differential.py's `skills/list-after-scratchpad`
// runs `scratchpad init` first so the git-ignored personal-rules layer becomes
// discoverable. The directory is created here directly, since `scratchpad init`
// belongs to another cluster.
func TestListMatchesReferenceCLIWithPersonalRules(t *testing.T) {
	root := repoRoot(t)
	src := filepath.Join(root, "examples", "failing-workspace")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	ws := copyTree(t, src)

	// Two personal rules under one team, one of them with no frontmatter at
	// all, which is the ({}, whole-text) branch of the parser.
	dir := filepath.Join(ws, "teams", "ghost", "scratchpad", "personal-rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "b-second.md"),
		"---\ntype: personal-rule\n---\n\n1. (default) my own step\n")
	write(t, filepath.Join(dir, "a-first.md"), "no frontmatter here\n")

	want := pythonSkillsList(t, root, ws)
	if got := listText(t, workspace.New(ws)); got != want {
		t.Errorf("skills list diverges with a populated personal layer\n%s",
			firstDiff(got, want))
	}
}

// TestListMatchesReferenceCLIOnResolvedExtends covers the merged view's
// base-then-extension branch (`:903-905`), which no committed fixture reaches:
// examples/failing-workspace has a DANGLING extends only, so without this the
// `layered on base` path would ship unmeasured.
func TestListMatchesReferenceCLIOnResolvedExtends(t *testing.T) {
	root := repoRoot(t)
	src := filepath.Join(root, "examples", "workspace")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	ws := copyTree(t, src)

	write(t, filepath.Join(ws, "teams", "customer-engagement", "skills", "creating-prd-mobile.SKILL.md"),
		"---\nid: skill://team/creating-prd-mobile\ntype: skill\nauthority: team\n"+
			"extends: platform-skill://communications/creating-prd\n---\n\n"+
			"# Mobile\n\n1. (default) add a mobile mock\n")

	want := pythonSkillsList(t, root, ws)
	if got := listText(t, workspace.New(ws)); got != want {
		t.Errorf("skills list diverges on a resolved extends\n%s",
			firstDiff(got, want))
	}
	if !strings.Contains(want, "layered on base") {
		t.Fatal("oracle did not exercise the base-layering branch")
	}
}

// firstDiff reports the first differing line with its neighbours, which is far
// easier to act on than two full merged views.
func firstDiff(got, want string) string {
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		gl, wl := at(g, i), at(w, i)
		if gl == wl {
			continue
		}
		var b strings.Builder
		b.WriteString("first difference at line ")
		b.WriteString(itoa(i + 1))
		b.WriteString("\n")
		for j := i - 2; j <= i+2; j++ {
			if j < 0 {
				continue
			}
			if j == i {
				b.WriteString("  got  | " + at(g, j) + "\n")
				b.WriteString("  want | " + at(w, j) + "\n")
				continue
			}
			b.WriteString("       | " + at(w, j) + "\n")
		}
		return b.String()
	}
	return "outputs are equal"
}

func at(lines []string, i int) string {
	if i < 0 || i >= len(lines) {
		return "<eof>"
	}
	return lines[i]
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

// copyTree copies a fixture into a temp dir so a test never mutates a committed
// workspace, matching differential.py's copy-per-invocation rule.
func copyTree(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copying fixture: %v", err)
	}
	return dst
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

// TestListMatchesReferenceCLIOnPathlibStemEdgeCases covers the file names where
// PurePath.stem and filepath.Ext disagree.
//
// `Path(".md").stem` is ".md" — pathlib treats a leading dot as part of the name
// — while filepath.Ext(".md") is ".md", so trimming it left an EMPTY skill name.
// The name is the shadowing identity AND the label `skills list` prints, so the
// old behavior both printed a blank entry and made every empty-named skill
// collide. Reachable at teams/<t>/scratchpad/personal-rules/.md.
func TestListMatchesReferenceCLIOnPathlibStemEdgeCases(t *testing.T) {
	root := repoRoot(t)
	src := filepath.Join(root, "examples", "failing-workspace")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	ws := copyTree(t, src)

	dir := filepath.Join(ws, "teams", "ghost", "scratchpad", "personal-rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".md", "trailing-dot.", "a.b.md", "plain.md"} {
		write(t, filepath.Join(dir, name),
			"---\ntype: personal-rule\n---\n\n1. (default) step\n")
	}

	want := pythonSkillsList(t, root, ws)
	if got := listText(t, workspace.New(ws)); got != want {
		t.Errorf("skills list diverges on a pathlib stem edge case\n%s",
			firstDiff(got, want))
	}
}
