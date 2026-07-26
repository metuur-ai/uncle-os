package skills

import (
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// `skills list` (cmd_skills, bin/company-os:869-917) as records.
//
// The Python command interleaves discovery, `extends` resolution and printing
// in one pass, so its output shape and its traversal are the same code. Here
// List does the traversal and emits a code plus typed fields per line;
// internal/render.Skills is the only thing that knows the wording, the indents
// and the blank lines (R-2.8). Nothing below is a sentence.

// Section slugs. Each is one contiguous block of the merged view, which is what
// gives the TUI something to scroll by and --json something to key on.
const (
	SectionBanner   = "banner"
	SectionLayers   = "layers"
	SectionMerged   = "merged-guidance"
	SectionPersonal = "personal-rules"
	SectionSummary  = "summary"
)

// Line codes for `skills list`. They live here rather than in internal/model
// because model's codes are validate's render sites; a command's own codes
// belong to the package that produces them, as internal/ids does.
const (
	// CodeBanner is the "== agent skills ... ==" title (`:871`).
	CodeBanner = "skills.banner"
	// CodeLayersHeader opens the origin-labeled section (`:874`).
	CodeLayersHeader = "skills.layers-header"
	// CodeLayerEmpty is a layer with no skills (`:878`). Every layer is listed
	// whether or not it is populated, so the reader sees which are unused.
	CodeLayerEmpty = "skills.layer-empty"
	// CodeLayerEntry is one discovered skill in the origin-labeled section
	// (`:889`).
	CodeLayerEntry = "skills.layer-entry"
	// CodeMergedHeader opens the merged guidance section (`:894`).
	CodeMergedHeader = "skills.merged-header"
	// CodeSkillHeader is one skill's label in the merged view (`:899`).
	CodeSkillHeader = "skills.skill-header"
	// CodeBaseHeader announces a resolved `extends` base (`:903`).
	CodeBaseHeader = "skills.base-header"
	// CodeBaseStep is a step inherited from that base (`:905`).
	CodeBaseStep = "skills.base-step"
	// CodeDanglingExtendsWarning is the inline warning for an `extends` that
	// does not resolve (`:907`). It is the merged view's own notice, distinct
	// from gate 7's model.CodeSkillDanglingExtends finding: this one never
	// affects an exit code.
	CodeDanglingExtendsWarning = "skills.dangling-extends-warning"
	// CodeStep is one of the skill's own steps (`:910`).
	CodeStep = "skills.step"
	// CodePersonalHeader opens the personal-rules section (`:913`).
	CodePersonalHeader = "skills.personal-header"
	// CodePersonalEntry is one personal rule (`:915`).
	CodePersonalEntry = "skills.personal-entry"
	// CodeSummary is the trailing tally (`:916-917`).
	CodeSummary = "skills.summary"
)

// List builds the merged view (GPF-R-5.1, 5.3, 5.4).
func List(ws *workspace.Workspace) ([]model.GateResult, error) {
	found, err := Discover(ws)
	if err != nil {
		return nil, err
	}

	banner := model.GateResult{
		Slug:     SectionBanner,
		Findings: []model.Finding{{Code: CodeBanner}},
	}

	layers := model.GateResult{
		Slug:     SectionLayers,
		Findings: []model.Finding{{Code: CodeLayersHeader}},
	}
	for _, layer := range Layers {
		items := byLayer(found, layer)
		if len(items) == 0 {
			layers.Findings = append(layers.Findings, model.Finding{
				Code:   CodeLayerEmpty,
				Fields: model.Fields{"layer": string(layer)},
			})
			continue
		}
		for _, s := range items {
			f := model.Finding{
				Code: CodeLayerEntry,
				Path: s.Rel,
				Fields: model.Fields{
					"layer": string(layer),
					"scope": s.Scope(),
					"name":  s.Name,
				},
			}
			// A personal rule carries a fixed non-overriding notice instead of
			// its metadata, so its id and authority are deliberately absent
			// from the record too — the Python line does not read them.
			if layer != LayerPersonal {
				f.Fields["id"] = s.ID.Text
				f.Fields["authority"] = s.Authority.Text
				if s.Extends.Truthy {
					f.Fields["extends"] = s.Extends.Text
				}
			}
			layers.Findings = append(layers.Findings, f)
		}
	}

	merged := model.GateResult{
		Slug:     SectionMerged,
		Findings: []model.Finding{{Code: CodeMergedHeader}},
	}
	for _, s := range found {
		if s.Layer == LayerPersonal {
			continue
		}
		merged.Findings = append(merged.Findings, model.Finding{
			Code: CodeSkillHeader,
			Path: s.Rel,
			Fields: model.Fields{
				"name": s.Name, "layer": string(s.Layer),
				"scope": s.Scope(), "authority": s.Authority.Text,
			},
		})
		if s.Extends.Truthy {
			base, ok, err := ResolveExtends(ws, s.Extends)
			if err != nil {
				return nil, err
			}
			switch {
			case ok:
				merged.Findings = append(merged.Findings, model.Finding{
					Code:   CodeBaseHeader,
					Path:   base.Rel,
					Fields: model.Fields{"extends": s.Extends.Text},
				})
				for _, head := range Steps(base.Body) {
					merged.Findings = append(merged.Findings, model.Finding{
						Code:   CodeBaseStep,
						Path:   base.Rel,
						Fields: model.Fields{"step": head},
					})
				}
			default:
				merged.Findings = append(merged.Findings, model.Finding{
					Severity: model.SevWarn,
					Code:     CodeDanglingExtendsWarning,
					Path:     s.Rel,
					Fields:   model.Fields{"extends": s.Extends.Text},
				})
			}
		}
		for _, head := range Steps(s.Body) {
			merged.Findings = append(merged.Findings, model.Finding{
				Code:   CodeStep,
				Path:   s.Rel,
				Fields: model.Fields{"step": head},
			})
		}
	}

	sections := []model.GateResult{banner, layers, merged}

	// The personal section is omitted entirely when the layer is empty
	// (`:912`), which is not the same as an empty section: the header must not
	// print either.
	if personal := byLayer(found, LayerPersonal); len(personal) > 0 {
		s := model.GateResult{
			Slug:     SectionPersonal,
			Findings: []model.Finding{{Code: CodePersonalHeader}},
		}
		for _, p := range personal {
			s.Findings = append(s.Findings, model.Finding{
				Code:   CodePersonalEntry,
				Path:   p.Rel,
				Fields: model.Fields{"team": p.Team, "name": p.Name},
			})
		}
		sections = append(sections, s)
	}

	sections = append(sections, model.GateResult{
		Slug: SectionSummary,
		Findings: []model.Finding{{
			Code: CodeSummary,
			// Counts, so they reach the text output as numbers and JSON as
			// `4`, not `"4"` (R-2.3).
			Fields: model.Fields{
				"skills": len(found),
				"layers": populatedLayers(found),
			},
		}},
	})
	return sections, nil
}

func byLayer(all []Skill, layer Layer) []Skill {
	var out []Skill
	for _, s := range all {
		if s.Layer == layer {
			out = append(out, s)
		}
	}
	return out
}

// populatedLayers is `len({s["layer"] for s in skills})` (`:917`).
func populatedLayers(all []Skill) int {
	seen := map[Layer]bool{}
	for _, s := range all {
		seen[s.Layer] = true
	}
	return len(seen)
}
