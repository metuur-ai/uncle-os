// Package roles owns the role-scoped read views: the plain-language glossary
// (GPF-R-3.1/3.2/3.3) and the `today` daily view.
//
// The package is named for the role concept rather than for the `today` command
// because role_glossary_lines (`bin/company-os:1260`) has two callers —
// cmd_today (`:1171`) and cmd_ids (`:1277`). A package called `today` could not
// hold it without internal/ids depending on internal/today, which inverts what
// the two things are. The dependency runs ids -> roles, one way.
//
// Everything here is read-only. The glossary in particular is display-only and
// touches no file, which is what GPF-R-3.2 requires: rendering a role view can
// never mutate an artifact.
//
// No sentence is composed here. Producers emit codes and typed Fields;
// internal/render turns them into the lines the Python CLI prints (R-2.8).
package roles

import "github.com/metuur-ai/uncle-os/company-os-starter/internal/model"

// SlugGlossary names the legend section, shared by `today` and `ids list`.
const SlugGlossary = "glossary"

// CodeGlossaryTerm is one canonical term paired with its plain-language label.
const CodeGlossaryTerm = "role.glossary-term"

// Term is one row of the legend. Both halves stay separate all the way to the
// renderer: the canonical term is the thing that must never change in artifacts,
// and the plain label is the thing that must never leak into one.
type Term struct {
	Canonical string
	Plain     string
}

// terms is ROLE_TERMS (`bin/company-os:1239-1257`), verbatim and in order. A
// role with no entry displays unchanged with no error (GPF-R-3.3), which is the
// nil return below rather than a sentinel.
var terms = map[string][]Term{
	"product-owner": {
		{"exception", "promise with an expiry date"},
		{"deviation", "documented exception to a default rule"},
		{"PRD", "the plan for a change (product requirements)"},
		{"outcome review", "the scheduled check on whether the change worked"},
	},
	"director-of-product": {
		{"exception", "promise with an expiry date"},
		{"deviation", "documented exception to a default rule"},
		{"PRD", "the plan for a change (product requirements)"},
		{"outcome review", "the scheduled check on whether the change worked"},
	},
	"team-lead": {
		{"deviation", "documented exception to a default rule"},
		{"exception", "approved, expiring waiver of a mandatory rule"},
		{"governance", "the requirements that apply to your components"},
	},
}

// Terms returns the legend for a role, or nil when the role has no entry. It is
// role_glossary_lines minus the formatting: pure, reads nothing, writes nothing.
func Terms(role string) []Term {
	return terms[role]
}

// GlossarySection wraps Terms as a section. ok is false for an unmapped role,
// which is how "no glossary, no error" reaches the renderer — an absent section
// rather than an empty one, so the renderer never has to decide whether to print
// the legend's intro line.
func GlossarySection(role string, ordinal int) (model.GateResult, bool) {
	t := Terms(role)
	if len(t) == 0 {
		return model.GateResult{}, false
	}
	g := model.GateResult{Ordinal: ordinal, Slug: SlugGlossary}
	for _, term := range t {
		g.Findings = append(g.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     CodeGlossaryTerm,
			Subject:  term.Canonical,
			Fields: model.Fields{
				"canonical": term.Canonical,
				"plain":     term.Plain,
				"role":      role,
			},
		})
	}
	return g, true
}
