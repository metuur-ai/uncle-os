package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// TestScaffoldGuidanceChain pins the exact stdout of the four scaffolding
// commands. Two requirements meet here and they pull in opposite directions:
//
//   - R-1.8 — every mutating command prints the next command in the workflow.
//   - R-1.9 — `scratchpad init` prints NO next step today and must keep printing
//     exactly what it prints today, because R-0.8 (byte-frozen output) outranks
//     R-1.8. Closing that gap is a separate change; a well-meaning "next:" line
//     added here is a regression.
//
// The differential harness proves this against Python, but only where Python is
// installed. This is the CI-side guard.
func TestScaffoldGuidanceChain(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	// The printed root is the RESOLVED one — Path.resolve() follows the /var ->
	// /private/var symlink macOS puts on every temp dir (task 1.6).
	resolved := workspace.New(root).Root

	out := runOK(t, "--root", root, "init", "--company", "Acme Inc.",
		"--team", "Core Team!", "--platform", "My Platform!!")
	wantLines(t, out, []string{
		"initialized workspace at " + resolved,
		"  company: Acme Inc. | first team: core-team | first platform: my-platform",
		`next: cd ` + resolved + ` && company-os discover new --team core-team "<discovery title>"`,
	})

	out = runOK(t, "--root", root, "add", "platform", "second")
	wantLines(t, out, []string{
		"added platform 'second'",
		"next: company-os add component --platform second <component-id>",
	})

	out = runOK(t, "--root", root, "add", "team", "second")
	wantLines(t, out, []string{
		"added team 'second'",
		`next: company-os discover new --team second "<discovery title>"`,
	})

	out = runOK(t, "--root", root, "add", "component", "billing-api", "--platform", "second")
	wantLines(t, out, []string{
		"added component 'billing-api' to platform 'second'",
		"next: company-os reality new --platform second billing-api",
	})

	out = runOK(t, "--root", root, "reality", "new", "--platform", "second", "billing-api")
	wantLines(t, out, []string{
		"created platforms/second/reality/components/billing-api.md",
		"  template: built-in templates/reality-component.md",
		"next: fill in Business rules / Current limitations, then continue: " +
			"company-os prd complete --platform second <prd-id>",
	})

	// R-1.9: one line, and no "next:".
	out = runOK(t, "scratchpad", "init", "--repo", root)
	wantLines(t, out, []string{
		"initialized " + filepath.Join(root, "scratchpad") + " and updated .gitignore",
	})
	if strings.Contains(out, "next:") {
		t.Errorf("scratchpad init grew a next-step line, which R-1.9 forbids:\n%s", out)
	}
}

// TestInitOutsideAWorkspaceIsExempt covers the require_root carve-out at
// bin/company-os:2774 from the dispatch side: `init` and `scratchpad` are the
// only two commands that must run outside a workspace root, and the tests above
// would pass even if the exemption were broken for only one of them.
func TestInitOutsideAWorkspaceIsExempt(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := run([]string{"--root", filepath.Join(dir, "nope"), "validate"}, &out, &errb); code == 0 {
		t.Fatal("validate outside a workspace root exited 0")
	}
	for _, argv := range [][]string{
		{"--root", filepath.Join(dir, "a"), "init", "--company", "A", "--team", "t", "--platform", "p"},
		{"scratchpad", "init", "--repo", filepath.Join(dir, "b")},
	} {
		out.Reset()
		errb.Reset()
		if code := run(argv, &out, &errb); code != 0 {
			t.Errorf("%v exited %d: %s", argv, code, errb.String())
		}
	}
}

func runOK(t *testing.T, argv ...string) string {
	t.Helper()
	var out, errb bytes.Buffer
	if code := run(argv, &out, &errb); code != 0 {
		t.Fatalf("%v exited %d\nstderr: %s", argv, code, errb.String())
	}
	return out.String()
}

// wantLines asserts the command's stdout is exactly these lines. It is an
// equality check rather than a Contains check because rebuild_generated's output
// (once internal/graph lands) must appear BEFORE them, and a Contains check
// would not notice it landing after.
func wantLines(t *testing.T, got string, lines []string) {
	t.Helper()
	want := strings.Join(lines, "\n") + "\n"
	if got != want {
		t.Errorf("stdout\n got: %q\nwant: %q", got, want)
	}
}
