package product

// The open-draft listing `today --team <t>` renders (R-9.1 … R-9.5).
//
// It lives here rather than in internal/roles because everything it has to know
// is here: where drafts live (R-1.1 … R-1.3), which status is open (R-9.2), and
// what counts as still-a-placeholder (R-9.3). internal/roles receives it as an
// injected roles.DraftsSection and decides only where the section goes and
// whether it appears — see the comment on that type for why the dependency runs
// this way round.
//
// No sentence is composed here; the codes and Fields are turned into lines by
// internal/render.

import (
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// draftsUnset is what a placeholder renders as (R-9.3).
//
// The rule earns its keep because the scaffold WRITES placeholders — DraftNew
// leaves `title: TODO` and, without --platform, `promoteTo.platform: TODO` — so
// without it the common case is a listing of literal TODOs, which reads as data
// rather than as an absence and tells the author nothing they did not know.
const draftsUnset = "unset"

// DraftsSection is one team's open drafts as a section, plus R-9.4's "is there
// anything to say" (false = omit the section, do not render it empty).
//
// R-9.2 is by exclusion, not enumeration: only `draft` is listed, so `promoted`
// and `abandoned` fall out, and so does a status this unit has never heard of.
// Enumerating the two exclusions would make a typo'd status render as open.
//
// R-9.5 is the signature: one team in, that team's drafts out. There is no walk
// over ws.AllTeams() to accidentally widen.
func DraftsSection(ws *workspace.Workspace, team string, ordinal int) (
	model.GateResult, bool, error) {

	draftDir, err := DraftDir(ws, team)
	if err != nil {
		return model.GateResult{}, false, err
	}

	s := model.GateResult{Ordinal: ordinal, Slug: model.SlugDrafts, Title: team}
	// R-1.3 via DraftIDs: an absent draft directory is a team with no drafts,
	// not a finding.
	for _, id := range DraftIDs(draftDir) {
		path := DraftPath(draftDir, id)
		meta, _, err := graph.ReadFrontmatter(path)
		if err != nil {
			return model.GateResult{}, false, err
		}
		if strOf(meta, "status") != statusDraft {
			continue
		}
		rel := relTo(ws.Root, path)
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeDraft,
			Subject:  id,
			Path:     rel,
			Fields: model.Fields{
				"draft":    id,
				"team":     team,
				"path":     rel,
				"title":    orUnset(strOf(meta, "title")),
				"platform": orUnset(promoteToPlatform(meta)),
			},
		})
	}
	if len(s.Findings) == 0 {
		return model.GateResult{}, false, nil
	}

	// The banner is prepended rather than emitted first, because whether it is
	// emitted at all depends on what the loop found (R-9.4).
	s.Findings = append([]model.Finding{{
		Severity: model.SevOK,
		Code:     model.CodeDraftsHeader,
		Subject:  team,
		Path:     relTo(ws.Root, draftDir),
		Fields:   model.Fields{"team": team, "drafts": len(s.Findings)},
	}}, s.Findings...)
	return s, true, nil
}

// orUnset applies R-9.3 through promotion's own predicate rather than a second
// one: the fields this listing calls unset are exactly the fields `prd promote`
// will ask the author to supply, and two answers to "is this filled in yet?"
// would let a draft read as ready in one view and unfinished in the other.
//
// The target platform is `promoteTo.platform` and not the process field
// `platform` (R-9.1 via R-4.1) — on a draft where the two disagree, this line
// shows where the draft is actually going rather than what its prose claims.
func orUnset(v string) string {
	if IsUnset(v) {
		return draftsUnset
	}
	return v
}
