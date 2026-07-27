package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// freshWorkspace is what `company-os init` leaves behind, produced by running
// the real command through run().
//
// Calling scaffold.Init directly does NOT reproduce it: without a Rebuild no
// CLAUDE.md is written at all, and the node gate reads that as "absent -> pass",
// so the fixture would be silently weaker than a real workspace. Driving run()
// is the only version whose fresh state is worth asserting about.
func freshWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := filepath.Join(t.TempDir(), "ws")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := run([]string{"--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "web"}, &out, &errb); code != 0 {
		t.Fatalf("init exited %d\n%s%s", code, out.String(), errb.String())
	}
	return workspace.New(root)
}

func titles(ds []diagnosis) string {
	var b strings.Builder
	for _, d := range ds {
		b.WriteString(d.Title)
		b.WriteString("\n")
	}
	return b.String()
}

// TestAdviseOffersGraphBuildForADriftedBlock is the motivating scenario: a
// generated block that no longer matches what the workspace would produce, which
// previously surfaced only as a [FAIL] line at the end of `validate`.
//
// The drift is induced here rather than taken from a fresh `init`, because a
// fresh `init` does not drift. An earlier version of this comment asserted the
// opposite — that `company-os init && company-os validate` exits 1 against the
// real binary while the same init through run() comes out in sync, and that the
// difference was not understood. Re-measured 2026-07-26 across three builds
// (working tree, the committed binary, the installed one) and both invocation
// forms (cwd and --root): every one exits 0. There is no divergence; the
// original claim was wrong. See task 6.11.
//
// So this test induces the condition directly, which is what it should have
// done regardless: the drifted-block path is reached by editing a generated
// block, not by scaffolding a workspace.
func TestAdviseOffersGraphBuildForADriftedBlock(t *testing.T) {
	ws := freshWorkspace(t)
	node := filepath.Join(ws.Root, "teams", "core", "CLAUDE.md")
	body, err := os.ReadFile(node)
	if err != nil {
		t.Fatal(err)
	}
	// Insert INSIDE the generated markers: outside them is hand-owned prose the
	// gate deliberately tolerates, so editing there would prove nothing. Matched
	// on the marker token rather than a whole line — the start marker carries a
	// "do not edit" note that would make a literal comparison brittle.
	lines := strings.Split(string(body), "\n")
	inserted := false
	for i, l := range lines {
		if strings.Contains(l, "company-os:generated:start") {
			lines = append(lines[:i+1], append([]string{"drifted line"}, lines[i+1:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		t.Fatalf("no generated start marker in %s; the fixture changed shape", node)
	}
	if err := os.WriteFile(node, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}

	got := advise(ws, ws.Root)
	if len(got) == 0 {
		t.Fatal("advise found nothing; it should offer graph build")
	}
	first := got[0]
	if !strings.Contains(first.Title, "drifted") {
		t.Fatalf("first diagnosis is %q, want the drifted-block one\nall:\n%s",
			first.Title, titles(got))
	}
	if first.Fix == nil || first.Fix.Cmd != "graph" || first.Fix.Action != "build" {
		t.Fatalf("fix = %+v, want `graph build`", first.Fix)
	}
	if !strings.Contains(first.Detail, "teams/core/CLAUDE.md") {
		t.Errorf("detail does not name the drifted file:\n%s", first.Detail)
	}
}

// TestAdviseIsSilentOnAHealthyWorkspace: an advisor that always has something to
// say is noise, and the menu it heads would stop being read.
func TestAdviseIsSilentOnAHealthyWorkspace(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", "workspace"))
	if err != nil {
		t.Fatal(err)
	}
	ws := workspace.New(root)
	if got := advise(ws, root); len(got) != 0 {
		t.Errorf("advise on the committed example workspace found %d problem(s):\n%s",
			len(got), titles(got))
	}
}

// TestAdviceHeadsTheCatalog is R-5.18's ordering clause. It also pins what a
// freshly scaffolded workspace actually looks like: `init` writes no
// teams/<t>/generated/, so governance is unresolved and the advisor is NOT
// empty there — while `validate` still exits 0 (GPF-R-1.7). Those two facts
// together are the ones an earlier round of comments got backwards.
func TestAdviceHeadsTheCatalog(t *testing.T) {
	ws := freshWorkspace(t)
	ds := advise(ws, ws.Root)
	if len(ds) == 0 {
		t.Fatal("a fresh workspace has unresolved governance; advise found nothing")
	}
	screens := screensFor(ws, ws.Root)
	if len(screens) < len(ds) {
		t.Fatalf("catalog has %d screens, fewer than the %d diagnoses", len(screens), len(ds))
	}
	for i, d := range ds {
		if screens[i].Title != d.Title {
			t.Errorf("screen %d is %q, want the diagnosis %q", i, screens[i].Title, d.Title)
		}
	}
}

// TestAdviseOffersRepairForAMissingTeamFile ties the detector to the fixer: the
// offer must be the invocation that actually resolves what was detected.
func TestAdviseOffersRepairForAMissingTeamFile(t *testing.T) {
	ws := freshWorkspace(t)
	target := filepath.Join(ws.Root, "teams", "core", "standards", "definition-of-ready.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	var found *diagnosis
	for _, d := range advise(ws, ws.Root) {
		if strings.Contains(d.Title, "restore") {
			d := d
			found = &d
		}
	}
	if found == nil {
		t.Fatal("no repair offered for a missing team file")
	}
	if found.Fix == nil || found.Fix.Cmd != "add" || found.Fix.Kind != "team" ||
		found.Fix.Name != "core" || !found.Fix.Repair {
		t.Fatalf("fix = %+v, want `add team core --repair`", found.Fix)
	}
	if !strings.Contains(found.Detail, "definition-of-ready.md") {
		t.Errorf("detail does not name the missing file:\n%s", found.Detail)
	}
}

// TestAdviseNeverOffersToWriteAManifest. A manifest needs repo URLs and commit
// pins nothing can infer; a form producing a plausible-but-wrong one would be
// worse than the missing file.
func TestAdviseNeverOffersToWriteAManifest(t *testing.T) {
	ws := freshWorkspace(t)
	if err := os.WriteFile(filepath.Join(ws.Root, "workspace.lock.yaml"),
		[]byte("version: 1\nrepos: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var seen bool
	for _, d := range advise(ws, ws.Root) {
		if strings.Contains(d.Title, workspace.ManifestName) {
			seen = true
			if d.Fix != nil {
				t.Errorf("a fix was offered for the manifest: %+v", d.Fix)
			}
		}
	}
	if !seen {
		t.Error("a lock without a manifest was not reported at all")
	}
}

// TestAdviseWritesNothing is R-5.8 for the advisor: detection runs on every TUI
// start, so it must be incapable of changing the workspace it inspects.
func TestAdviseWritesNothing(t *testing.T) {
	ws := freshWorkspace(t)
	before := hashTree(t, ws.Root)
	_ = advise(ws, ws.Root)

	// Building every offer is what happens when the reader opens one to see the
	// preview, before any confirmation exists.
	for _, s := range adviceScreens(ws, ws.Root) {
		if s.Form != nil {
			if _, err := s.Form.Build(nil); err != nil {
				t.Fatalf("%s: Build: %v", s.Title, err)
			}
			continue
		}
		if _, err := s.Run(""); err != nil {
			t.Fatalf("%s: Run: %v", s.Title, err)
		}
	}
	if after := hashTree(t, ws.Root); after != before {
		t.Error("the advisor modified the workspace it inspected")
	}
}

// TestAdviceOffersRoundTripThroughTheParser is R-5.6/R-5.10 for the new offers:
// every previewed line must be one the real parser reads back into the same
// *Args. An offer the CLI cannot reproduce is not an offer.
func TestAdviceOffersRoundTripThroughTheParser(t *testing.T) {
	ws := freshWorkspace(t)
	if err := os.Remove(filepath.Join(ws.Root, "teams", "core", "standards",
		"definition-of-done.md")); err != nil {
		t.Fatal(err)
	}
	ds := advise(ws, ws.Root)
	if len(ds) < 2 {
		t.Fatalf("expected several diagnoses, got %d", len(ds))
	}
	for _, d := range ds {
		if d.Fix == nil {
			continue
		}
		line := screenCommand(d.Fix)
		if line == "" {
			t.Errorf("%s: empty preview", d.Title)
			continue
		}
		tokens := shellSplit(line)
		if len(tokens) == 0 || tokens[0] != "company-os" {
			t.Errorf("%s: preview %q does not start with the program name", d.Title, line)
			continue
		}
		back, err := parse(tokens[1:])
		if err != nil {
			t.Errorf("%s: preview %q does not parse: %v", d.Title, line, err)
			continue
		}
		if got, want := screenCommand(back), line; got != want {
			t.Errorf("%s: round trip changed the invocation\n got: %s\nwant: %s",
				d.Title, got, want)
		}
	}
}
