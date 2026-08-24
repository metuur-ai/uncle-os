package product

import (
	"path/filepath"
	"reflect"
	"testing"
)

// TestDraftDir is R-1.1: one fixed location, derived from the team directory.
func TestDraftDir(t *testing.T) {
	ws := fixture(t)
	got, err := DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(ws.Root, "teams", "core", "product", "change-records", "draft")
	if got != want {
		t.Errorf("DraftDir = %q, want %q", got, want)
	}
}

// TestDraftDirUnknownTeam: an unknown team is the same not-found every other
// product command raises, not a silently-joined path.
func TestDraftDirUnknownTeam(t *testing.T) {
	ws := fixture(t)
	if _, err := DraftDir(ws, "nope"); err == nil {
		t.Fatal("expected an error for an unknown team")
	}
}

// TestDraftPath is R-1.2: `<draft-dir>/<draft-id>/prd.md`.
func TestDraftPath(t *testing.T) {
	got := DraftPath("/d", "2026-thing")
	want := filepath.Join("/d", "2026-thing", "prd.md")
	if got != want {
		t.Errorf("DraftPath = %q, want %q", got, want)
	}
}

// TestDraftIDsAbsentDir is R-1.3: no draft directory means no drafts and no
// error to report — the empty and the missing case are indistinguishable.
func TestDraftIDsAbsentDir(t *testing.T) {
	ws := fixture(t)
	dir, err := DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	if got := DraftIDs(dir); got != nil {
		t.Errorf("DraftIDs on an absent dir = %v, want nil", got)
	}
}

// TestDraftIDsEnumerates: sorted ids, and only entries that actually carry a
// prd.md (R-1.2).
func TestDraftIDsEnumerates(t *testing.T) {
	ws := fixture(t)
	dir, err := DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	write(t, DraftPath(dir, "2026-beta"), "---\ntype: prd\n---\n")
	write(t, DraftPath(dir, "2026-alpha"), "---\ntype: prd\n---\n")
	write(t, filepath.Join(dir, "2026-empty", ".keep"), "")
	write(t, filepath.Join(dir, "stray.md"), "not a draft")

	want := []string{"2026-alpha", "2026-beta"}
	if got := DraftIDs(dir); !reflect.DeepEqual(got, want) {
		t.Errorf("DraftIDs = %v, want %v", got, want)
	}
}
