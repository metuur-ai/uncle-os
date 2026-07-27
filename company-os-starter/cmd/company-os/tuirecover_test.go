package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
)

// TestNearbyRootsFindsTheReportedCase is the bug this feature came from, pinned
// against the real fixture: standing in examples/banking/bank/workspaces/ — a
// directory holding two workspace roots — `tui` used to exit 3 with the
// root-resolution order instead of offering either of them.
func TestNearbyRootsFindsTheReportedCase(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "..", "..",
		"examples", "banking", "bank", "workspaces"))
	if err != nil {
		t.Fatal(err)
	}
	got := nearbyRoots(dir)
	if len(got) != 2 {
		t.Fatalf("nearbyRoots(%s) = %v, want the two team workspaces", dir, got)
	}
	joined := strings.Join(got, " ")
	for _, want := range []string{"team-fraud-detection", "team-payments-rails"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in %v", want, got)
		}
	}
	if !sort.StringsAreSorted(got) {
		t.Errorf("results are not sorted: %v", got)
	}
}

// TestNearbyRootsIncludesItself covers `tui` run inside a root that the caller
// reached some other way — the scan must not skip the directory it starts in.
func TestNearbyRootsIncludesItself(t *testing.T) {
	dir, _ := filepath.Abs(filepath.Join("..", "..", "..", "examples", "workspace"))
	got := nearbyRoots(dir)
	if len(got) == 0 || got[0] != dir {
		t.Fatalf("nearbyRoots(%s) = %v, want it to include itself", dir, got)
	}
}

// TestNearbyRootsStopsAtOneLevel pins the depth decision. A root two levels down
// is deliberately NOT offered: a recursive scan of an arbitrary directory is
// slow where it matters and surprising everywhere.
func TestNearbyRootsStopsAtOneLevel(t *testing.T) {
	base := t.TempDir()
	deep := filepath.Join(base, "a", "b")
	if err := os.MkdirAll(filepath.Join(deep, "company-os"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := nearbyRoots(base); len(got) != 0 {
		t.Errorf("nearbyRoots found a root two levels down: %v", got)
	}
	if got := nearbyRoots(filepath.Join(base, "a")); len(got) != 1 {
		t.Errorf("nearbyRoots(one level up) = %v, want the child root", got)
	}
}

// TestRecoveryOmitsThePickerWhenNothingIsNearby: an empty picker is a dead end
// that looks like a feature.
func TestRecoveryOmitsThePickerWhenNothingIsNearby(t *testing.T) {
	var sel string
	screens := recoveryScreens(t.TempDir(), &sel)
	for _, s := range screens {
		if strings.Contains(s.Title, "found nearby") {
			t.Fatalf("empty dir offered a workspace picker: %q", s.Title)
		}
	}
	if len(screens) != 2 {
		t.Fatalf("got %d screens, want create + help", len(screens))
	}
}

// TestRecoveryPickerHandsOffTheChosenRoot is the contract cmdTUI's restart loop
// depends on: the picker records the absolute path and ends the run.
func TestRecoveryPickerHandsOffTheChosenRoot(t *testing.T) {
	base := t.TempDir()
	want := filepath.Join(base, "ws")
	if err := os.MkdirAll(filepath.Join(want, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	var selected string
	screens := recoveryScreens(base, &selected)
	if len(screens) == 0 || !strings.Contains(screens[0].Title, "found nearby") {
		t.Fatalf("first screen is not the picker: %+v", screens[0].Title)
	}
	body, err := screens[0].Run(screens[0].Choices[0])
	if !errors.Is(err, tui.ErrHandOff) {
		t.Fatalf("Run returned %v, want tui.ErrHandOff", err)
	}
	if body != "" {
		t.Errorf("hand-off rendered a body: %q", body)
	}
	if selected != want {
		t.Errorf("selected = %q, want %q", selected, want)
	}
}

// TestRecoveryScreensWriteNothing is R-5.8 for the new screens: building every
// form and reading the help must not touch the filesystem. Only Commit writes,
// and Commit is reached only past a confirmation.
func TestRecoveryScreensWriteNothing(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "ws", "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	before := hashTree(t, base)

	var selected string
	for _, s := range recoveryScreens(base, &selected) {
		if s.Form != nil {
			// Build is what runs when the reader asks to see the preview — it
			// happens before any confirmation exists.
			if _, err := s.Form.Build([]string{"Acme", "core", "web"}); err != nil {
				t.Fatalf("%s: Build: %v", s.Title, err)
			}
			continue
		}
		choices := s.Choices
		if len(choices) == 0 {
			choices = []string{""}
		}
		for _, c := range choices {
			if _, err := s.Run(c); err != nil && !errors.Is(err, tui.ErrHandOff) {
				t.Fatalf("%s: Run(%q): %v", s.Title, c, err)
			}
		}
	}
	if after := hashTree(t, base); after != before {
		t.Error("opening the recovery screens modified the filesystem")
	}
}

// TestRootHelpNamesEveryMarker keeps the help honest: it replaced an error that
// listed the markers, so dropping one would make the replacement worse than
// what it replaced.
func TestRootHelpNamesEveryMarker(t *testing.T) {
	got := rootHelp("/somewhere")
	for _, name := range []string{
		"company-os", "platforms", "teams", "company-ontology", "knowledge",
		"workspace.yaml", "COMPANY_OS_WORKSPACE_ROOT", "--root",
	} {
		if !strings.Contains(got, name) {
			t.Errorf("root help does not mention %q:\n%s", name, got)
		}
	}
}

func hashTree(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	var paths []string
	if err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, p)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	for _, p := range paths {
		rel, _ := filepath.Rel(root, p)
		h.Write([]byte(rel))
		if b, err := os.ReadFile(p); err == nil {
			h.Write(b)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
