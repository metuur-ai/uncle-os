package product

import (
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// draftFields returns the CodeDraft rows of a section, keyed by draft id.
func draftFields(t *testing.T, s model.GateResult) map[string]model.Fields {
	t.Helper()
	out := map[string]model.Fields{}
	for _, f := range s.Findings {
		if f.Code == model.CodeDraft {
			out[f.Fields.Str("draft")] = f.Fields
		}
	}
	return out
}

// TestDraftsSectionListsOpenDrafts is R-9.1: id, title and target platform, one
// row per open draft, under a banner.
func TestDraftsSectionListsOpenDrafts(t *testing.T) {
	ws := fixture(t)
	if _, err := DraftNew(ws, "core", "payments", "Faster checkout", "", nil); err != nil {
		t.Fatal(err)
	}
	// A title is a placeholder on a brand-new draft (R-2.4), so fill one in to
	// prove the filled case renders the real value.
	dir, _ := DraftDir(ws, "core")
	id := DraftIDs(dir)[0]
	setField(t, DraftPath(dir, id), "title", "Faster checkout")

	s, ok, err := DraftsSection(ws, "core", 3)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("an open draft produced no section")
	}
	if s.Slug != model.SlugDrafts || s.Ordinal != 3 || s.Title != "core" {
		t.Errorf("section = %+v, want slug=%s ordinal=3 title=core", s, model.SlugDrafts)
	}
	if s.Findings[0].Code != model.CodeDraftsHeader {
		t.Fatalf("first finding = %q, want the banner", s.Findings[0].Code)
	}
	if got := s.Findings[0].Fields.Int("drafts"); got != 1 {
		t.Errorf("banner count = %d, want 1", got)
	}
	rows := draftFields(t, s)
	row, seen := rows[id]
	if !seen {
		t.Fatalf("draft %q is not listed; got %v", id, rows)
	}
	if row.Str("title") != "Faster checkout" {
		t.Errorf("title = %q, want %q", row.Str("title"), "Faster checkout")
	}
	if row.Str("platform") != "payments" {
		t.Errorf("platform = %q, want payments", row.Str("platform"))
	}
}

// TestDraftsSectionExcludesClosed is R-9.2: `promoted` and `abandoned` drafts
// are not open drafts. A status nobody has ever heard of is excluded too, which
// is what matching by exclusion buys.
func TestDraftsSectionExcludesClosed(t *testing.T) {
	for _, status := range []string{"promoted", "abandoned", "invented"} {
		t.Run(status, func(t *testing.T) {
			ws := fixture(t)
			id, path := draftFor(t, ws, "payments")
			setStatus(t, path, status)

			s, ok, err := DraftsSection(ws, "core", 1)
			if err != nil {
				t.Fatal(err)
			}
			if ok {
				t.Errorf("a %s draft is still listed: %v", status, draftFields(t, s))
			}
			_ = id
		})
	}
}

// TestDraftsSectionPlaceholdersRenderUnset is R-9.3, on the artifact the
// scaffold ACTUALLY writes: `prd new --draft` with no --platform leaves both
// title and promoteTo.platform as TODO, and neither may reach the reader.
func TestDraftsSectionPlaceholdersRenderUnset(t *testing.T) {
	ws := fixture(t)
	if _, err := DraftNew(ws, "core", "", "Something vague", "", nil); err != nil {
		t.Fatal(err)
	}
	s, ok, err := DraftsSection(ws, "core", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("a placeholder-filled draft is still an open draft")
	}
	for id, row := range draftFields(t, s) {
		if row.Str("title") != draftsUnset {
			t.Errorf("%s: title = %q, want %q", id, row.Str("title"), draftsUnset)
		}
		if row.Str("platform") != draftsUnset {
			t.Errorf("%s: platform = %q, want %q", id, row.Str("platform"), draftsUnset)
		}
		if row.Str("title") == placeholder || row.Str("platform") == placeholder {
			t.Errorf("%s: the literal placeholder reached the reader", id)
		}
	}
}

// TestDraftsSectionOmittedWhenEmpty is R-9.4: no open drafts means no section,
// not an empty one. Both the never-drafted team and the all-closed team.
func TestDraftsSectionOmittedWhenEmpty(t *testing.T) {
	ws := fixture(t)
	if _, ok, err := DraftsSection(ws, "core", 1); err != nil || ok {
		t.Errorf("a team with no draft dir got a section (ok=%v, err=%v)", ok, err)
	}

	_, path := draftFor(t, ws, "payments")
	setStatus(t, path, "promoted")
	s, ok, err := DraftsSection(ws, "core", 1)
	if err != nil {
		t.Fatal(err)
	}
	if ok || len(s.Findings) != 0 {
		t.Errorf("a team whose only draft is closed got a section: %+v", s)
	}
}

// TestDraftsSectionScopesToTeam is R-9.5: one team in, that team's drafts out.
func TestDraftsSectionScopesToTeam(t *testing.T) {
	ws := fixture(t)
	write(t, ws.Root+"/teams/platform/team.yaml", "schemaVersion: '1.0'\nid: team://platform\n")
	if _, err := DraftNew(ws, "core", "payments", "Core work", "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := DraftNew(ws, "platform", "payments", "Platform work", "", nil); err != nil {
		t.Fatal(err)
	}

	s, ok, err := DraftsSection(ws, "core", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("core has an open draft")
	}
	rows := draftFields(t, s)
	if len(rows) != 1 {
		t.Fatalf("core's section lists %d drafts, want 1: %v", len(rows), rows)
	}
	for id := range rows {
		if id != derivePRDID("Core work") {
			t.Errorf("core's section lists %q, which is not core's draft", id)
		}
	}
}

// TestDraftsSectionUnknownTeam: an unknown team is the same not-found every
// other team-scoped command raises, not an empty listing that reads as "no
// drafts" about a team that does not exist.
func TestDraftsSectionUnknownTeam(t *testing.T) {
	ws := fixture(t)
	if _, _, err := DraftsSection(ws, "nope", 1); err == nil {
		t.Fatal("expected an error for an unknown team")
	}
}

// setField rewrites one top-level scalar frontmatter key in place, through the
// same writer the commands use so the fence stays the parser's.
func setField(t *testing.T, path, key, value string) {
	t.Helper()
	meta, body, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeArtifact(path, meta.Set(key, yamlio.PyStr(value)), body); err != nil {
		t.Fatal(err)
	}
}
