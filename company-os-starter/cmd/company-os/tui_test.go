package main

// The `tui` subcommand's contract, tested through run() rather than through a
// pty. Everything R-5.1..R-5.4 and R-5.16 assert about it is observable at that
// seam: which argv reaches the UI, what a refusal writes and exits with, and
// whether the filesystem moved.
//
// The UI itself is tested in internal/tui, by calling Update and reading View.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// withoutTTY forces the gate closed for the duration of a test. Under `go test`
// both streams are already pipes, so the production probe refuses anyway — but a
// test that depends on that is one `go test 2>&1 >/dev/tty` away from launching
// an interactive program and hanging the suite.
func withoutTTY(t *testing.T) {
	t.Helper()
	prev := interactive
	interactive = func() bool { return false }
	t.Cleanup(func() { interactive = prev })
}

// tuiWorkspace scaffolds a real workspace to point the catalog at.
func tuiWorkspace(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "ws")
	var out, errOut bytes.Buffer
	if code := run([]string{"--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "plat"},
		&out, &errOut); code != 0 {
		t.Fatalf("init failed (%d): %s", code, errOut.String())
	}
	if code := run([]string{"--root", root, "add", "component",
		"--platform", "plat", "svc"}, &out, &errOut); code != 0 {
		t.Fatalf("add component failed (%d): %s", code, errOut.String())
	}
	return root
}

// TestTUIWithoutATTYExitsSevenAndChangesNothing is R-5.3, end to end: the exit
// code, the stream the message lands on, and the filesystem.
//
// The tree is hashed before and after rather than eyeballed, because "makes no
// filesystem change" is the half of the requirement a reader cannot see and the
// half a future edit is most likely to break — a screen that memoises to disk, a
// `governance resolve` slipped into the overview.
func TestTUIWithoutATTYExitsSevenAndChangesNothing(t *testing.T) {
	withoutTTY(t)
	root := tuiWorkspace(t)
	before := treeDigest(t, root)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "tui"}, &stdout, &stderr)

	if code != int(model.ExitInteractive) {
		t.Errorf("exit = %d, want %d (R-5.3)", code, model.ExitInteractive)
	}
	if stdout.Len() != 0 {
		t.Errorf("the refusal wrote to stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "terminal") {
		t.Errorf("stderr does not explain the refusal: %q", stderr.String())
	}
	// R-5.3 says "explanatory". A code and a noun are not an explanation for a
	// reader who does not know the binary has a non-interactive surface at all.
	if !strings.Contains(stderr.String(), "company-os validate") {
		t.Errorf("stderr names no way forward: %q", stderr.String())
	}
	if after := treeDigest(t, root); after != before {
		t.Error("the refusal changed the workspace (R-5.3)")
	}
}

// TestTUIRefusesOutsideAWorkspaceWithSevenNotThree pins the ordering the
// exemption in run() exists for: R-5.3 is unconditional, so the same
// non-interactive invocation must exit 7 whether or not it is standing in a
// workspace. Without the exemption this returns 3 and the contract is two
// contracts.
func TestTUIRefusesOutsideAWorkspaceWithSevenNotThree(t *testing.T) {
	withoutTTY(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", t.TempDir(), "tui"}, &stdout, &stderr)
	if code != int(model.ExitInteractive) {
		t.Errorf("exit = %d outside a workspace, want %d", code, model.ExitInteractive)
	}
}

// TestNoOtherSubcommandNeedsATTY is R-5.16. Every other command runs with the
// gate forced CLOSED and must be unaffected — none of them may consult it, and
// none may acquire a TTY requirement later without failing here.
func TestNoOtherSubcommandNeedsATTY(t *testing.T) {
	withoutTTY(t)
	root := tuiWorkspace(t)
	repo := t.TempDir()
	for _, argv := range [][]string{
		{"--root", root, "validate"},
		{"--root", root, "today"},
		{"--root", root, "ids", "list"},
		{"--root", root, "skills", "list"},
		{"--root", root, "graph", "build"},
		{"--root", root, "governance", "resolve", "--team", "core"},
		{"--root", root, "check", "ready", "--team", "core", "--components", "svc"},
		{"--root", root, "reality", "new", "--platform", "plat", "svc"},
		{"--root", root, "discover", "new", "--team", "core", "A brief"},
		{"--root", root, "add", "team", "second"},
		{"--root", root, "scratchpad", "init", "--repo", repo},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(argv, &stdout, &stderr); code == int(model.ExitInteractive) {
			t.Errorf("%v exited 7 with no TTY — R-5.16 forbids any subcommand "+
				"but `tui` requiring one: %s", argv, stderr.String())
		}
	}
}

// TestQuittingTheRecoveryMenuExitsZero is R-5.23, and it is the ONLY externally
// observable surface of R-5.17's exemption from R-4.4.
//
// Before R-5.17, `tui` outside a workspace root exited 3. It now opens a recovery
// menu, and a reader who quits without choosing gets 0 — because quitting a menu
// is not a failure, so R-4.10 is not engaged (Amendment 1's ruling). R-4.10
// explicitly contemplates callers branching on exit status, so this number is
// part of the contract and not an implementation detail.
//
// It had no test until 2026-07-27; task 7.7's verify line claimed it while
// nothing checked it. Found in review by @uncle-dev:uncle-lead.
func TestQuittingTheRecoveryMenuExitsZero(t *testing.T) {
	prevTTY := interactive
	interactive = func() bool { return true }
	t.Cleanup(func() { interactive = prevTTY })

	// "q" quits from any screen (R-5.14). Driving the real UI rather than
	// calling recoveryScreens directly is the point: the exit code is produced
	// by run(), and that is what a caller sees.
	prevIn := stdin
	stdin = strings.NewReader("q")
	t.Cleanup(func() { stdin = prevIn })

	dir := t.TempDir() // not a workspace root, and holds none
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--root", dir, "tui"}, &stdout, &stderr); code != 0 {
		t.Errorf("quitting the recovery menu exited %d, want 0\nstderr: %s",
			code, stderr.String())
	}
}

// TestBareInvocationDoesNotLaunchTheTUI is R-5.2. A bare `company-os` prints
// usage and exits 2, exactly as it did before this subcommand existed — the
// property that keeps an agent or a CI job from landing in an interactive app.
//
// The gate is left OPEN here on purpose: if a bare invocation could reach the
// UI, closing the gate would hide it behind the same exit 7 the real launcher
// produces, and the test would pass for the wrong reason.
func TestBareInvocationDoesNotLaunchTheTUI(t *testing.T) {
	prev := interactive
	interactive = func() bool { return true }
	t.Cleanup(func() { interactive = prev })

	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != int(model.ExitUsage) {
		t.Errorf("bare invocation exit = %d, want %d", code, model.ExitUsage)
	}
	if !strings.Contains(stderr.String(), "required: cmd") {
		t.Errorf("bare invocation did not report a missing subcommand: %q", stderr.String())
	}
}

// TestNoEnvironmentVariableLaunchesTheTUI is the other half of R-5.2. The
// variables a launcher would plausibly be hung off are set to every value that
// usually means "yes", and a bare invocation still prints usage.
func TestNoEnvironmentVariableLaunchesTheTUI(t *testing.T) {
	prev := interactive
	interactive = func() bool { return true }
	t.Cleanup(func() { interactive = prev })

	for _, name := range []string{
		"COMPANY_OS_TUI", "COMPANY_OS_INTERACTIVE", "TUI", "INTERACTIVE", "TERM",
	} {
		for _, value := range []string{"1", "true", "yes", "xterm-256color"} {
			t.Setenv(name, value)
			var stdout, stderr bytes.Buffer
			if code := run(nil, &stdout, &stderr); code != int(model.ExitUsage) {
				t.Errorf("%s=%s changed a bare invocation to exit %d (R-5.2)",
					name, value, code)
			}
		}
	}
}

// TestReadOnlyScreensAreTheEnumeratedTen is R-5.4. The list is asserted by
// title, in order, so that adding a screen or dropping one is a decision someone
// has to make here rather than a silent drift away from the requirement.
func TestReadOnlyScreensAreTheEnumeratedTen(t *testing.T) {
	ws := workspace.New(tuiWorkspace(t))
	got := readOnlyScreens(ws, "")
	want := []string{
		"workspace overview",
		"today (role view)",
		"validate results",
		"component browser",
		"PRD browser",
		"discovery browser",
		"governance explain",
		"skills list",
		"ids list",
		"workspace status",
	}
	if len(got) != len(want) {
		t.Fatalf("%d screens, R-5.4 enumerates %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Title != want[i] {
			t.Errorf("screen %d is %q, want %q", i, got[i].Title, want[i])
		}
	}
}

// TestEveryScreenRunsAndWritesNothing is the property that makes "read-only
// screens ship first" (R-5.4, R-5.5) a fact rather than an intention.
//
// Every screen is executed — including every choice of every picker, since a
// screen is only read-only for the arguments it was actually run with — and the
// workspace is hashed around the whole sweep.
func TestEveryScreenRunsAndWritesNothing(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)
	before := treeDigest(t, root)

	for _, s := range readOnlyScreens(ws, "") {
		args := s.Choices
		if len(args) == 0 {
			args = []string{""}
		}
		for _, a := range args {
			body, err := s.Run(a)
			// An error is legitimate — `workspace status` on a monorepo
			// workspace has no manifest — but silence is not: a screen that
			// returns neither text nor a reason renders an empty frame.
			if err == nil && strings.TrimSpace(body) == "" {
				t.Errorf("screen %q (%q) produced no output and no error", s.Title, a)
			}
		}
	}
	if after := treeDigest(t, root); after != before {
		t.Errorf("a read-only screen changed the workspace (R-5.4):\n%s", diffTrees(t, before, after))
	}
}

// TestScreensRenderThroughTheRealRenderers is R-5.13: a screen's body is what
// the flag CLI would have printed, byte for byte, because it went through the
// same Command and the same Renderer.
func TestScreensRenderThroughTheRealRenderers(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)

	for _, tc := range []struct {
		screen string
		choice string
		argv   []string
	}{
		{"validate results", "", []string{"--root", root, "validate"}},
		{"skills list", "", []string{"--root", root, "skills", "list"}},
		{"ids list", "", []string{"--root", root, "ids", "list"}},
		{"today (role view)", "product-owner",
			[]string{"--root", root, "today", "--role", "product-owner"}},
	} {
		s := screenByTitle(t, readOnlyScreens(ws, ""), tc.screen)
		body, _ := s.Run(tc.choice)

		var stdout, stderr bytes.Buffer
		run(tc.argv, &stdout, &stderr)

		// The screen prefixes the derived invocation; the records follow it
		// unedited. Comparing the remainder is what proves no re-derivation.
		_, records, ok := strings.Cut(body, "\n\n")
		if !ok {
			t.Fatalf("%s: body carries no command header:\n%s", tc.screen, body)
		}
		if records != stdout.String() {
			t.Errorf("%s: the TUI body is not the CLI's output (R-5.13)\n"+
				"--- tui ---\n%s\n--- cli ---\n%s", tc.screen, records, stdout.String())
		}
	}
}

// TestScreenCommandIsDerivedFromTheExecutedArgs is R-5.7's read-only ancestor:
// the header a screen shows is computed from the same *Args the screen runs, so
// the two cannot drift. A hand-written label would pass nothing here.
func TestScreenCommandIsDerivedFromTheExecutedArgs(t *testing.T) {
	for _, tc := range []struct {
		args *Args
		want string
	}{
		{&Args{Cmd: "validate"}, "company-os validate"},
		{&Args{Cmd: "skills", Action: "list"}, "company-os skills list"},
		{&Args{Cmd: "workspace", Action: "status"}, "company-os workspace status"},
		{&Args{Cmd: "today", Role: "architect"}, "company-os today --role architect"},
		{&Args{Cmd: "governance", Action: "explain", ComponentArg: "svc"},
			"company-os governance explain svc"},
	} {
		if got := screenCommand(tc.args); got != tc.want {
			t.Errorf("screenCommand(%+v) = %q, want %q", tc.args, got, tc.want)
		}
	}
	// The default is elided rather than printed: `--role developer` is what
	// argparse fills in, so showing it would teach a flag the reader never has
	// to type.
	if got := screenCommand(&Args{Cmd: "today", Role: "developer"}); got != "company-os today" {
		t.Errorf("the role default leaked into the derived command: %q", got)
	}
}

// TestRoleChoicesComeFromTheParser guards the one place the UI could offer a
// value the CLI rejects.
func TestRoleChoicesComeFromTheParser(t *testing.T) {
	got := roleChoices()
	if len(got) == 0 {
		t.Fatal("no roles")
	}
	for _, role := range got {
		if _, err := parse([]string{"today", "--role", role}); err != nil {
			t.Errorf("the TUI offers role %q, which the parser rejects: %v", role, err)
		}
	}
}

// TestGovernanceExplainPickerMatchesTheCatalog: the picker must offer exactly
// the components that exist, or it offers a lookup that dies.
func TestGovernanceExplainPickerMatchesTheCatalog(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)
	s := screenByTitle(t, readOnlyScreens(ws, ""), "governance explain")
	if len(s.Choices) == 0 {
		t.Fatal("no components offered for a workspace that has one")
	}
	for _, cid := range s.Choices {
		if _, _, found := ws.FindComponent(cid); !found {
			t.Errorf("the picker offers %q, which is not a component", cid)
		}
	}
}

// TestNoColorRequested implements the published convention: present AND
// non-empty. `NO_COLOR=` is what a shell produces from an unset variable
// expansion and must not mean the opposite of what was typed.
func TestNoColorRequested(t *testing.T) {
	for _, tc := range []struct {
		set   bool
		value string
		want  bool
	}{
		{false, "", false},
		{true, "", false},
		{true, "1", true},
		{true, "0", true}, // the convention is presence, not truthiness
		{true, "anything", true},
	} {
		os.Unsetenv("NO_COLOR")
		if tc.set {
			t.Setenv("NO_COLOR", tc.value)
		}
		if got := noColorRequested(); got != tc.want {
			t.Errorf("NO_COLOR set=%v value=%q -> %v, want %v",
				tc.set, tc.value, got, tc.want)
		}
	}
}

func screenByTitle(t *testing.T, screens []tui.Screen, title string) tui.Screen {
	t.Helper()
	for _, s := range screens {
		if s.Title == title {
			return s
		}
	}
	t.Fatalf("no screen titled %q", title)
	return tui.Screen{}
}

// treeDigest hashes every path and every file's contents under root. Comparing
// two of these is how "changed nothing" is asserted without listing what a
// command might have touched.
func treeDigest(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		if d.IsDir() {
			lines = append(lines, "d "+rel)
			return nil
		}
		body, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		sum := sha256.Sum256(body)
		lines = append(lines, "f "+rel+" "+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// diffTrees reports the lines that differ, so a failure names the file rather
// than two hashes.
func diffTrees(t *testing.T, before, after string) string {
	t.Helper()
	was := map[string]bool{}
	for _, l := range strings.Split(before, "\n") {
		was[l] = true
	}
	var out []string
	for _, l := range strings.Split(after, "\n") {
		if !was[l] {
			out = append(out, "  + "+l)
		}
	}
	return strings.Join(out, "\n")
}
