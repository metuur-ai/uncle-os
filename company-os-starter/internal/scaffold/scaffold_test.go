package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// --------------------------------------------------------------- slugify

// TestSlugify covers R-1.13. The cases are the ones the 0.3 corpus drives
// (`init/slugify`, `add/platform-slugify`) plus the boundary shapes re.sub +
// strip produce: a run of specials collapses to ONE dash, and leading/trailing
// dashes are stripped after the substitution, not before.
func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Acme Inc.", "acme-inc"},
		{"Core Team!", "core-team"},
		{"My Platform!!", "my-platform"},
		{"Weird Name!!", "weird-name"},
		{"newplat", "newplat"},
		{"---leading---", "leading"},
		{"  ", ""},
		{"", ""},
		{"a  b", "a-b"},
		{"MiXeD123", "mixed123"},
		{"Ünïcødé", "n-c-d"},
	}
	for _, c := range cases {
		if got := Slugify(c.in); got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestTitleCase pins Python's str.title(), which is what turns an id back into
// a display name. The digit case is the non-obvious one: a digit is not cased,
// so the letter after it is uppercased.
func TestTitleCase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"core team", "Core Team"},
		{"my platform", "My Platform"},
		{"newcomp", "Newcomp"},
		{"v2 rocket", "V2 Rocket"},
		{"v2rocket", "V2Rocket"},
		{"", ""},
	}
	for _, c := range cases {
		if got := titleCase(c.in); got != c.want {
			t.Errorf("titleCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ------------------------------------------------------------ exit codes

func codeOf(t *testing.T, err error) model.ExitCode {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	return model.CodeOf(err)
}

// TestWriteNewRefusesToClobber is GPF-R-1.9 and the exit-code map's ruling A:
// every "already exists" refusal is code 8, and it mutates nothing.
func TestWriteNewRefusesToClobber(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b.yaml")
	if err := writeNew(path, "first\n"); err != nil {
		t.Fatalf("writeNew: %v", err)
	}
	err := writeNew(path, "second\n")
	if got := codeOf(t, err); got != model.ExitConflict {
		t.Fatalf("exit code = %d, want %d", got, model.ExitConflict)
	}
	var conflictErr *ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("error %v is not a *ConflictError", err)
	}
	if conflictErr.Path != path {
		t.Errorf("Path = %q, want %q", conflictErr.Path, path)
	}
	if got, _ := os.ReadFile(path); string(got) != "first\n" {
		t.Errorf("the refused write mutated the file: %q", got)
	}
	if !strings.Contains(err.Error(), "refusing to overwrite existing file: "+path) {
		t.Errorf("message = %q", err.Error())
	}
}

// ------------------------------------------------------------- _prompt

// TestPromptFlagWins, TestPromptAsksWhenInteractive and
// TestPromptRefusesWithoutTerminal are the three arms of GPF-R-1.3.
func TestPromptFlagWins(t *testing.T) {
	got, err := resolvePrompt(nil, "given", "company", "Company name", "My Company")
	if err != nil {
		t.Fatalf("resolvePrompt: %v", err)
	}
	if got != "given" {
		t.Errorf("= %q, want %q", got, "given")
	}
}

func TestPromptAsksWhenInteractive(t *testing.T) {
	var asked []string
	ask := func(label, def string) (string, error) {
		asked = append(asked, label+"|"+def)
		return "  typed  \n", nil
	}
	got, err := resolvePrompt(ask, "", "team", "First team id", "core")
	if err != nil {
		t.Fatalf("resolvePrompt: %v", err)
	}
	if got != "typed" {
		t.Errorf("= %q, want %q — the answer is stripped", got, "typed")
	}
	if len(asked) != 1 || asked[0] != "First team id|core" {
		t.Errorf("asked = %v", asked)
	}

	blank := func(string, string) (string, error) { return "\n", nil }
	got, err = resolvePrompt(blank, "", "team", "First team id", "core")
	if err != nil {
		t.Fatalf("resolvePrompt: %v", err)
	}
	if got != "core" {
		t.Errorf("empty answer = %q, want the default %q", got, "core")
	}
}

// TestPromptRefusesWithoutTerminal pins the message and the code. It is 7, not
// 2: the three flags are optional to argparse and the requirement materializes
// only from the absence of a TTY (.devlocal/go-port/exit-code-map.md § D).
func TestPromptRefusesWithoutTerminal(t *testing.T) {
	_, err := resolvePrompt(nil, "", "company", "Company name", "My Company")
	if got := codeOf(t, err); got != model.ExitInteractive {
		t.Fatalf("exit code = %d, want %d", got, model.ExitInteractive)
	}
	want := "non-interactive run: --company is required when no terminal is " +
		"attached (pass --company, --team, and --platform)"
	if err.Error() != want {
		t.Fatalf("message\n got: %q\nwant: %q", err.Error(), want)
	}
}

// ----------------------------------------------------------------- init

func TestInitScaffoldsEveryRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	ws := workspace.New(root)
	res, err := Init(ws, InitOptions{Company: "Acme Inc.", Team: "Core Team!", Platform: "My Platform!!"})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if res.Team != "core-team" || res.Platform != "my-platform" || res.Company != "Acme Inc." {
		t.Fatalf("result = %+v — the ids are slugified, the company name is not", res)
	}
	for _, rel := range []string{
		"company-os/company.yaml",
		"company-os/standards/company-baseline.yaml",
		"platforms/my-platform/platform.yaml",
		"platforms/my-platform/governance/requirements.yaml",
		"teams/core-team/team.yaml",
		"teams/core-team/standards/definition-of-ready.md",
		"teams/core-team/standards/definition-of-done.md",
		"teams/core-team/onboarding/developer.md",
		"company-ontology/ids/registry.yaml",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	// GPF-R-1.7's precondition: the seeded registry names both new units.
	registry := read(t, filepath.Join(root, "company-ontology", "ids", "registry.yaml"))
	for _, id := range []string{"platform://my-platform", "team://core-team"} {
		if !strings.Contains(registry, id) {
			t.Errorf("registry does not name %s:\n%s", id, registry)
		}
	}
}

// TestInitRefusesInsideAWorkspace is GPF-R-1.2 — refuse, exit 8, mutate
// nothing. Code 8 and not 3: 3 means the workspace object is ABSENT, and here
// it is present (exit-code-map § A).
func TestInitRefusesInsideAWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "teams"), 0o777); err != nil {
		t.Fatal(err)
	}
	before := tree(t, root)

	_, err := Init(workspace.New(root), InitOptions{Company: "A", Team: "t", Platform: "p"})
	if got := codeOf(t, err); got != model.ExitConflict {
		t.Fatalf("exit code = %d, want %d", got, model.ExitConflict)
	}
	if !strings.Contains(err.Error(), "is already a workspace root") ||
		!strings.Contains(err.Error(), "refusing to re-init") {
		t.Errorf("message = %q", err.Error())
	}
	if got := tree(t, root); got != before {
		t.Errorf("the refusal mutated the workspace:\nbefore %v\nafter  %v", before, got)
	}
}

// TestInitRefusesOnAManifestAlone covers GPF-R-6.2: workspace.yaml marks a root
// before any canonical directory exists, so a federated workspace that has
// never been synced still refuses re-init.
func TestInitRefusesOnAManifestAlone(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, workspace.ManifestName), []byte("repos: []\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	_, err := Init(workspace.New(root), InitOptions{Company: "A", Team: "t", Platform: "p"})
	if got := codeOf(t, err); got != model.ExitConflict {
		t.Fatalf("exit code = %d, want %d", got, model.ExitConflict)
	}
}

// TestInitIsAtomic is GPF-R-1.4, and it is the reason `init` scaffolds into a
// staging directory at all (bin/company-os:1982). It aborts the run at the last
// possible moment before the move — inside rebuild_generated, which is the only
// caller-supplied step that runs against the fully scaffolded staging tree — and
// asserts the target is untouched.
func TestInitIsAtomic(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	boom := errors.New("aborted mid-scaffold")
	var staged string

	_, err := Init(workspace.New(root), InitOptions{
		Company: "Acme", Team: "core", Platform: "platform-1",
		Rebuild: func(ws *workspace.Workspace) ([]string, error) {
			// Everything is on disk at this point, just not where the user
			// would see it — which is exactly the property under test.
			staged = ws.Root
			if _, err := os.Stat(filepath.Join(staged, "teams", "core", "team.yaml")); err != nil {
				t.Errorf("staging tree incomplete when rebuild ran: %v", err)
			}
			return nil, boom
		},
	})
	if !errors.Is(err, boom) {
		t.Fatalf("Init = %v, want the aborting error", err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("an aborted init left %s behind (%v)", root, err)
	}
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Errorf("the staging directory %s was not cleaned up", staged)
	}
}

// TestInitLeavesNothingBehindOnAPromptRefusal is the other abort point: the
// wizard fails before any directory is created at all.
func TestInitLeavesNothingBehindOnAPromptRefusal(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	_, err := Init(workspace.New(root), InitOptions{Company: "Acme"})
	if got := codeOf(t, err); got != model.ExitInteractive {
		t.Fatalf("exit code = %d, want %d", got, model.ExitInteractive)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("a refused init created %s", root)
	}
}

// TestInitRunsRebuildAgainstStagingBeforeTheMove pins the ordering the seam
// depends on: internal/graph derives the generated artifacts inside the staging
// tree, so they land in the target atomically with the sources.
func TestInitRunsRebuildAgainstStagingBeforeTheMove(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	res, err := Init(workspace.New(root), InitOptions{
		Company: "Acme", Team: "core", Platform: "platform-1",
		Rebuild: func(ws *workspace.Workspace) ([]string, error) {
			if ws.Root == root {
				t.Error("rebuild ran against the target, not the staging directory")
			}
			marker := filepath.Join(ws.Root, "company-os", "CLAUDE.md")
			if err := os.WriteFile(marker, []byte("derived\n"), 0o666); err != nil {
				return nil, err
			}
			return []string{"  node company-os/CLAUDE.md"}, nil
		},
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if len(res.Generated) != 1 || res.Generated[0] != "  node company-os/CLAUDE.md" {
		t.Errorf("Generated = %q — rebuild's lines must reach the renderer", res.Generated)
	}
	if got := read(t, filepath.Join(root, "company-os", "CLAUDE.md")); got != "derived\n" {
		t.Errorf("the derived artifact did not move into place: %q", got)
	}
}

// ------------------------------------------------------------------ add

func TestAddPlatformTeamComponent(t *testing.T) {
	root := initWorkspace(t)
	ws := workspace.New(root)

	if _, err := Add(ws, AddPlatform, "Weird Name!!", "", nil); err != nil {
		t.Fatalf("add platform: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "platforms", "weird-name", "platform.yaml")); err != nil {
		t.Errorf("platform not scaffolded: %v", err)
	}

	if _, err := Add(ws, AddTeam, "Second Team", "", nil); err != nil {
		t.Fatalf("add team: %v", err)
	}
	for _, rel := range []string{
		"teams/second-team/team.yaml",
		"teams/second-team/standards/definition-of-ready.md",
		"teams/second-team/standards/definition-of-done.md",
		"teams/second-team/onboarding/developer.md",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}

	res, err := Add(ws, AddComponent, "New Comp", "weird-name", nil)
	if err != nil {
		t.Fatalf("add component: %v", err)
	}
	if res.ID != "new-comp" || res.Platform != "weird-name" {
		t.Errorf("result = %+v", res)
	}
	desc := read(t, filepath.Join(root, "platforms", "weird-name", "components", "new-comp.yaml"))
	if !strings.Contains(desc, "name: New Comp") {
		t.Errorf("descriptor does not carry the title-cased name:\n%s", desc)
	}

	registry := read(t, filepath.Join(root, "company-ontology", "ids", "registry.yaml"))
	for _, id := range []string{"platform://weird-name", "team://second-team", "component://new-comp"} {
		if !strings.Contains(registry, id) {
			t.Errorf("registry does not name %s:\n%s", id, registry)
		}
	}
}

// TestAddComponentWithoutPlatform is exit 2, not 3: --platform is optional to
// argparse, so this is the missing-required-argument error argparse could not
// express (exit-code-map row 16).
func TestAddComponentWithoutPlatform(t *testing.T) {
	ws := workspace.New(initWorkspace(t))
	_, err := Add(ws, AddComponent, "x", "", nil)
	if got := codeOf(t, err); got != model.ExitUsage {
		t.Fatalf("exit code = %d, want %d", got, model.ExitUsage)
	}
	if err.Error() != "add component requires --platform <platform-id>" {
		t.Errorf("message = %q", err.Error())
	}
}

// TestAddComponentUnknownPlatform routes through workspace.PlatformDir, so the
// object-absent code 3 reaches the caller unchanged.
func TestAddComponentUnknownPlatform(t *testing.T) {
	ws := workspace.New(initWorkspace(t))
	_, err := Add(ws, AddComponent, "x", "nope", nil)
	if got := codeOf(t, err); got != model.ExitWorkspace {
		t.Fatalf("exit code = %d, want %d", got, model.ExitWorkspace)
	}
}

// TestAddDuplicatePlatformRefuses is the _write_new refusal reached from a
// command: code 8, and the pre-existing descriptor is untouched.
func TestAddDuplicatePlatformRefuses(t *testing.T) {
	root := initWorkspace(t)
	ws := workspace.New(root)
	before := read(t, filepath.Join(root, "platforms", "platform-1", "platform.yaml"))

	_, err := Add(ws, AddPlatform, "platform-1", "", nil)
	if got := codeOf(t, err); got != model.ExitConflict {
		t.Fatalf("exit code = %d, want %d", got, model.ExitConflict)
	}
	if got := read(t, filepath.Join(root, "platforms", "platform-1", "platform.yaml")); got != before {
		t.Error("the refused add rewrote the existing descriptor")
	}
}

// TestRegisterIDIsIdempotent covers the `if not any(...)` guard at :1823. Two
// registrations of one id leave one entry AND leave the file byte-identical,
// which is what keeps `git status` clean across repeated scaffolds.
func TestRegisterIDIsIdempotent(t *testing.T) {
	root := initWorkspace(t)
	path := filepath.Join(root, "company-ontology", "ids", "registry.yaml")

	if err := registerID(root, "platform://dup", "platforms/dup/platform.yaml"); err != nil {
		t.Fatalf("registerID: %v", err)
	}
	once := read(t, path)
	if err := registerID(root, "platform://dup", "platforms/dup/platform.yaml"); err != nil {
		t.Fatalf("registerID (again): %v", err)
	}
	if twice := read(t, path); twice != once {
		t.Errorf("second registration changed the file:\n%s\n---\n%s", once, twice)
	}
	if n := strings.Count(once, "platform://dup"); n != 1 {
		t.Errorf("id appears %d times, want 1", n)
	}
}

// --------------------------------------------------------------- reality

func TestRealityNewSubstitutesAndNamesItsTemplate(t *testing.T) {
	root := initWorkspace(t)
	ws := workspace.New(root)
	if _, err := Add(ws, AddComponent, "billing-api", "platform-1", nil); err != nil {
		t.Fatalf("add component: %v", err)
	}

	res, err := RealityNew(ws, "platform-1", "billing-api", nil)
	if err != nil {
		t.Fatalf("RealityNew: %v", err)
	}
	if res.Path != "platforms/platform-1/reality/components/billing-api.md" {
		t.Errorf("Path = %q", res.Path)
	}
	if res.Source != SourceBuiltinReality {
		t.Errorf("Source = %q, want %q", res.Source, SourceBuiltinReality)
	}
	body := read(t, filepath.Join(root, filepath.FromSlash(res.Path)))
	for _, want := range []string{"id: reality-billing-api", "# Billing Api — Current Behavior"} {
		if !strings.Contains(body, want) {
			t.Errorf("body is missing %q:\n%s", want, body)
		}
	}
	for _, unwanted := range []string{"<YYYY-MM-DD>", "<Component Name>", "reality-<component-id>"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("placeholder %q survived:\n%s", unwanted, body)
		}
	}
}

// TestRealityNewFallsBackToTheComponentID covers `(desc or {}).get(...) or cid`
// for a component with no descriptor at all.
func TestRealityNewFallsBackToTheComponentID(t *testing.T) {
	root := initWorkspace(t)
	res, err := RealityNew(workspace.New(root), "platform-1", "never-declared", nil)
	if err != nil {
		t.Fatalf("RealityNew: %v", err)
	}
	body := read(t, filepath.Join(root, filepath.FromSlash(res.Path)))
	if !strings.Contains(body, "# never-declared — Current Behavior") {
		t.Errorf("did not fall back to the component id:\n%s", body)
	}
}

// TestRealityNewRefusesToOverwrite pins code 8 and the workspace-RELATIVE path
// in the message — `reality new` renders the path differently from _write_new,
// which uses the absolute one.
func TestRealityNewRefusesToOverwrite(t *testing.T) {
	root := initWorkspace(t)
	ws := workspace.New(root)
	if _, err := RealityNew(ws, "platform-1", "svc", nil); err != nil {
		t.Fatalf("RealityNew: %v", err)
	}
	_, err := RealityNew(ws, "platform-1", "svc", nil)
	if got := codeOf(t, err); got != model.ExitConflict {
		t.Fatalf("exit code = %d, want %d", got, model.ExitConflict)
	}
	want := "platforms/platform-1/reality/components/svc.md already exists — refusing to overwrite"
	if err.Error() != want {
		t.Fatalf("message\n got: %q\nwant: %q", err.Error(), want)
	}
}

// TestRealityNewPrefersAWorkspaceOverride is GPF-R-4.1 reaching this command:
// the platform-scoped override wins over the embedded built-in, and its label
// is what gets printed.
func TestRealityNewPrefersAWorkspaceOverride(t *testing.T) {
	root := initWorkspace(t)
	override := filepath.Join(root, "platforms", "platform-1", "templates", "reality-component.md")
	if err := os.MkdirAll(filepath.Dir(override), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(override, []byte("flavored <Component Name>\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	res, err := RealityNew(workspace.New(root), "platform-1", "svc", nil)
	if err != nil {
		t.Fatalf("RealityNew: %v", err)
	}
	if res.Source != "platforms/platform-1/templates/reality-component.md" {
		t.Errorf("Source = %q", res.Source)
	}
	if got := read(t, filepath.Join(root, filepath.FromSlash(res.Path))); got != "flavored svc\n" {
		t.Errorf("body = %q", got)
	}
}

// ------------------------------------------------------------ scratchpad

func TestScratchpadInit(t *testing.T) {
	repo := t.TempDir()
	res, err := ScratchpadInit(repo)
	if err != nil {
		t.Fatalf("ScratchpadInit: %v", err)
	}
	if res.Base != filepath.Join(repo, "scratchpad") {
		t.Errorf("Base = %q", res.Base)
	}
	for _, sub := range scratchpadDirs {
		if fi, err := os.Stat(filepath.Join(repo, "scratchpad", sub)); err != nil || !fi.IsDir() {
			t.Errorf("missing %s: %v", sub, err)
		}
	}
	if got := read(t, filepath.Join(repo, "scratchpad", "README.md")); got != scratchpadReadme {
		t.Errorf("README = %q", got)
	}
	if got := read(t, filepath.Join(repo, ".gitignore")); got != gitignoreBlock {
		t.Errorf(".gitignore = %q", got)
	}
}

// TestScratchpadInitDoesNotAppendTwice is the `if "scratchpad/" not in existing`
// guard, and the reason the 0.3 corpus runs `scratchpad init` twice in one tree.
func TestScratchpadInitDoesNotAppendTwice(t *testing.T) {
	repo := t.TempDir()
	if _, err := ScratchpadInit(repo); err != nil {
		t.Fatalf("first: %v", err)
	}
	once := read(t, filepath.Join(repo, ".gitignore"))
	if _, err := ScratchpadInit(repo); err != nil {
		t.Fatalf("second: %v", err)
	}
	if twice := read(t, filepath.Join(repo, ".gitignore")); twice != once {
		t.Errorf(".gitignore grew on the second run:\n%q", twice)
	}
}

// TestScratchpadInitPreservesAnExistingGitignore checks the append, not a
// rewrite: an unrelated .gitignore keeps its content and gains the block.
func TestScratchpadInitPreservesAnExistingGitignore(t *testing.T) {
	repo := t.TempDir()
	gi := filepath.Join(repo, ".gitignore")
	if err := os.WriteFile(gi, []byte("node_modules/\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := ScratchpadInit(repo); err != nil {
		t.Fatalf("ScratchpadInit: %v", err)
	}
	if got := read(t, gi); got != "node_modules/\n"+gitignoreBlock {
		t.Errorf(".gitignore = %q", got)
	}
}

// TestScratchpadBaseRendersLikePathlib pins the printed string: `Path(".") /
// "scratchpad"` is `scratchpad`, not `./scratchpad`, and that string reaches
// stdout verbatim.
func TestScratchpadBaseRendersLikePathlib(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	for _, repo := range []string{"", "."} {
		res, err := ScratchpadInit(repo)
		if err != nil {
			t.Fatalf("ScratchpadInit(%q): %v", repo, err)
		}
		if res.Base != "scratchpad" {
			t.Errorf("ScratchpadInit(%q).Base = %q, want %q", repo, res.Base, "scratchpad")
		}
	}
	res, err := ScratchpadInit("sub/dir")
	if err != nil {
		t.Fatalf("ScratchpadInit: %v", err)
	}
	if res.Base != filepath.Join("sub", "dir", "scratchpad") {
		t.Errorf("Base = %q", res.Base)
	}
}

// --------------------------------------------------------------- helpers

// initWorkspace scaffolds a workspace the way `init` does and returns its root,
// so the `add` and `reality` tests start from the same tree the corpus does.
func initWorkspace(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "ws")
	if _, err := Init(workspace.New(root), InitOptions{
		Company: "Acme", Team: "core", Platform: "platform-1",
	}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return root
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// tree is a stable snapshot of every path under root with its mode, so a test
// can assert a refusal mutated nothing.
func tree(t *testing.T, root string) string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		out = append(out, rel+" "+fi.Mode().String())
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return strings.Join(out, "\n")
}
