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

// List builds the merged view (GPF-R-5.1, 5.3, 5.4).
func List(ws *workspace.Workspace) ([]model.GateResult, error) {
	found, err := Discover(ws)
	if err != nil {
		return nil, err
	}

	banner := model.GateResult{
		Slug:     model.SectionBanner,
		Findings: []model.Finding{{Code: model.CodeBanner}},
	}

	layers := model.GateResult{
		Slug:     model.SectionLayers,
		Findings: []model.Finding{{Code: model.CodeLayersHeader}},
	}
	for _, layer := range Layers {
		items := byLayer(found, layer)
		if len(items) == 0 {
			layers.Findings = append(layers.Findings, model.Finding{
				Code:   model.CodeLayerEmpty,
				Fields: model.Fields{"layer": string(layer)},
			})
			continue
		}
		for _, s := range items {
			f := model.Finding{
				Code: model.CodeLayerEntry,
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
		Slug:     model.SectionMerged,
		Findings: []model.Finding{{Code: model.CodeMergedHeader}},
	}
	for _, s := range found {
		if s.Layer == LayerPersonal {
			continue
		}
		merged.Findings = append(merged.Findings, model.Finding{
			Code: model.CodeSkillHeader,
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
					Code:   model.CodeBaseHeader,
					Path:   base.Rel,
					Fields: model.Fields{"extends": s.Extends.Text},
				})
				for _, head := range Steps(base.Body) {
					merged.Findings = append(merged.Findings, model.Finding{
						Code:   model.CodeBaseStep,
						Path:   base.Rel,
						Fields: model.Fields{"step": head},
					})
				}
			default:
				merged.Findings = append(merged.Findings, model.Finding{
					Severity: model.SevWarn,
					Code:     model.CodeDanglingExtendsWarning,
					Path:     s.Rel,
					Fields:   model.Fields{"extends": s.Extends.Text},
				})
			}
		}
		for _, head := range Steps(s.Body) {
			merged.Findings = append(merged.Findings, model.Finding{
				Code:   model.CodeStep,
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
			Slug:     model.SectionPersonal,
			Findings: []model.Finding{{Code: model.CodePersonalHeader}},
		}
		for _, p := range personal {
			s.Findings = append(s.Findings, model.Finding{
				Code:   model.CodePersonalEntry,
				Path:   p.Rel,
				Fields: model.Fields{"team": p.Team, "name": p.Name},
			})
		}
		sections = append(sections, s)
	}

	sections = append(sections, model.GateResult{
		Slug: model.SectionSummary,
		Findings: []model.Finding{{
			Code: model.CodeSummary,
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
