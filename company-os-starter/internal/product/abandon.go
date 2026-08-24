package product

// Retiring a draft (R-8.1 … R-8.5).
//
// Abandonment is the one lifecycle move that produces nothing: no record, no
// log line, no provenance. That is the point — the draft never became real, so
// there is nothing to trace it to. What it does produce is a STATE, which is
// what separates `prd abandon` from `rm`: the draft stays where it was, still
// readable, and every consumer that filters on status stops speaking about it.

import (
	"os"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// statusAbandoned is the terminal draft status (R-8.1).
//
// R-8.4 is enforced by omission and cannot be enforced any other way: no
// command in this package writes any status onto a draft carrying this one,
// because both mutators (PRDPromote, PRDAbandon) refuse anything that is not
// `draft` before they write. Adding a transition out would mean adding a
// branch here; there is none, and this comment is the reason there is none.
const statusAbandoned = "abandoned"

// PRDAbandon is `prd abandon --team <t> <draft-id>` (R-8.1 … R-8.5).
//
// It mirrors PRDPromote's refusal order deliberately — team, id, residency,
// status — so the two mutators disagree about nothing, and a draft that
// `promote` calls a conflict `abandon` calls a conflict too, in the same words.
//
// R-8.3 needs no code here. `abandoned` is not `draft`, so promotion's step 1
// already refuses it; it is not `promoted`, so the promotion-integrity gate
// already passes over it (R-7.7); and `today` lists `draft` and nothing else
// (R-9.2). Excluding it a fourth time, by name, would be the second spelling of
// a state the rest of the system matches by exclusion.
func PRDAbandon(ws *workspace.Workspace, team, draftID string, rebuild Rebuild) (
	[]model.GateResult, error) {

	if team == "" {
		// Same reason as DraftNew and PRDPromote: a draft is team-private, and
		// --team is not argparse-required on the `prd` sub-parser.
		return nil, model.Usagef("prd", "the following arguments are required: --team")
	}
	if err := requirePRDID(draftID); err != nil {
		return nil, err
	}
	draftDir, err := DraftDir(ws, team)
	if err != nil {
		return nil, err
	}
	path := DraftPath(draftDir, draftID)
	if _, err := os.Stat(path); err != nil {
		return nil, model.Errorf(model.ExitWorkspace, "no draft at %s", path)
	}
	meta, body, err := graph.ReadFrontmatter(path)
	if err != nil {
		return nil, err
	}
	rel := relTo(ws.Root, path)

	// R-8.2. The status found is IN the sentence, because "not a draft" alone
	// leaves the reader unable to tell an already-abandoned draft from a
	// promoted one — and only one of those two is a mistake worth undoing.
	if status := strOf(meta, "status"); status != statusDraft {
		return nil, model.Errorf(model.ExitConflict,
			"%s: status is '%s', not 'draft'", rel, status)
	}

	// R-8.1. Rewritten in place through the same writer promotion uses, so the
	// `^---\n…\n---\n` fence and the key order are one implementation and not
	// two. Nothing else in the frontmatter is touched: a `promoteTo` that was
	// filled in stays filled in, because the draft is kept as a record of what
	// was proposed, not scrubbed.
	if err := writeArtifact(path, meta.Set("status", yamlio.PyStr(statusAbandoned)), body); err != nil {
		return nil, err
	}

	// R-8.5. `graph build` derives a `status/` facet, so the draft's own tags
	// are stale the instant the line above lands. Returning before the rebuild
	// would leave `validate` red on a successful command — the same trap
	// DraftNew and PRDPromote avoid, avoided the same way.
	out := []model.GateResult{{
		Ordinal: 1, Slug: model.SlugPRDAbandon, Title: draftID,
		Findings: []model.Finding{
			okFinding(model.CodePRDAbandoned, "", rel,
				model.Fields{"path": rel, "draft": draftID, "team": team}),
		},
	}}
	if rebuild != nil {
		derived, err := rebuild(ws)
		if err != nil {
			return nil, err
		}
		out = append(out, derived...)
	}
	return out, nil
}
