package scaffold

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// treeHash fingerprints every file under root by path AND content, so a test
// can assert "nothing else moved" rather than only checking the file it cares
// about. A repair that fixed the target and quietly rewrote a neighbour would
// pass a narrower assertion.
func treeHash(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	var paths []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		paths = append(paths, p)
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	sort.Strings(paths)
	for _, p := range paths {
		rel, _ := filepath.Rel(root, p)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		h.Write([]byte(rel))
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TestRepairRestoresOnlyWhatIsMissing is the invariant that makes `--repair`
// safe to offer from the TUI: it fills gaps and touches nothing else.
//
// The assertion is the whole-tree hash taken BEFORE the file was removed. A
// repair that restored the file but perturbed any other byte — a re-registered
// id written differently, a regenerated block with a new timestamp — fails here,
// which a per-file comparison would not catch.
func TestRepairRestoresOnlyWhatIsMissing(t *testing.T) {
	root := initWorkspace(t)
	ws := workspace.New(root)

	target := filepath.Join(root, "teams", "core", "standards", "definition-of-ready.md")
	original := read(t, target)
	before := treeHash(t, root)

	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	res, err := RepairTeam(ws, "core", nil)
	if err != nil {
		t.Fatalf("RepairTeam: %v", err)
	}

	if got := read(t, target); got != original {
		t.Errorf("repaired file is not what creation wrote\n got: %q\nwant: %q", got, original)
	}
	if after := treeHash(t, root); after != before {
		t.Errorf("repair changed something other than the missing file\n"+
			"before=%s\n after=%s", before, after)
	}
	if len(res.Written) != 1 ||
		!strings.HasSuffix(res.Written[0], "definition-of-ready.md") {
		t.Errorf("Written = %v, want exactly the one missing file", res.Written)
	}
	if len(res.Skipped) != 3 {
		t.Errorf("Skipped = %v, want the three files that were present", res.Skipped)
	}
}

// TestRepairIsANoopWhenNothingIsMissing pins that a speculative repair is safe
// and says so. Silence would read as failure, and a rebuild on a no-op would
// make `--repair` something you have to think before running.
func TestRepairIsANoopWhenNothingIsMissing(t *testing.T) {
	root := initWorkspace(t)
	before := treeHash(t, root)

	res, err := RepairTeam(workspace.New(root), "core", nil)
	if err != nil {
		t.Fatalf("RepairTeam: %v", err)
	}
	if len(res.Written) != 0 {
		t.Errorf("Written = %v, want none", res.Written)
	}
	if len(res.Skipped) != 4 {
		t.Errorf("Skipped = %v, want all four scaffolded files", res.Skipped)
	}
	if len(res.Generated) != 0 {
		t.Errorf("Generated = %v; a no-op repair must not rebuild", res.Generated)
	}
	if after := treeHash(t, root); after != before {
		t.Error("a no-op repair modified the workspace")
	}
}

// TestRepairNeverOverwritesAModifiedFile is the property a user's edits depend
// on: repair is not "restore to pristine". A file they changed is a file that
// exists, and existing files are skipped.
func TestRepairNeverOverwritesAModifiedFile(t *testing.T) {
	root := initWorkspace(t)
	target := filepath.Join(root, "teams", "core", "standards", "definition-of-done.md")
	const mine = "---\ntype: team-standard\n---\n\n# my own definition of done\n"
	if err := os.WriteFile(target, []byte(mine), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := RepairTeam(workspace.New(root), "core", nil); err != nil {
		t.Fatalf("RepairTeam: %v", err)
	}
	if got := read(t, target); got != mine {
		t.Errorf("repair overwrote an edited file:\n got: %q\nwant: %q", got, mine)
	}
}

// TestRepairRefusesATeamThatDoesNotExist keeps `--repair` and `add team`
// distinct. Creating the team here would make the two interchangeable and hide
// which one happened.
func TestRepairRefusesATeamThatDoesNotExist(t *testing.T) {
	root := initWorkspace(t)
	before := treeHash(t, root)

	_, err := RepairTeam(workspace.New(root), "ghost", nil)
	if err == nil {
		t.Fatal("RepairTeam(ghost) = nil; want an error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error does not say the team is absent: %v", err)
	}
	if after := treeHash(t, root); after != before {
		t.Error("a refused repair wrote something")
	}
}

// TestRepairWritesWhatCreationWouldHave closes the loop the whole design rests
// on: repair and creation read one definition, so their output cannot differ.
// Asserted by rebuilding a team from scratch in a second workspace and diffing
// every scaffolded file.
func TestRepairWritesWhatCreationWouldHave(t *testing.T) {
	repaired := initWorkspace(t)
	files, err := teamFiles(repaired, "core")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := os.Remove(f.Path); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := RepairTeam(workspace.New(repaired), "core", nil); err != nil {
		t.Fatalf("RepairTeam: %v", err)
	}

	created := initWorkspace(t)
	for _, f := range files {
		rel := f.Rel
		a := read(t, filepath.Join(repaired, filepath.FromSlash(rel)))
		b := read(t, filepath.Join(created, filepath.FromSlash(rel)))
		if a != b {
			t.Errorf("%s differs between repair and creation\n repair: %q\ncreate: %q", rel, a, b)
		}
	}
}
