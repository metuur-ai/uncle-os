package governance

// The record sets the four actions return, and the one function in the package
// that produces human prose.
//
// Message is a pure function of (code, Fields) and nothing else (R-2.8, R-2.12):
// every number in a sentence is read back out of Fields here, so a field that
// stops being populated changes the rendered line rather than going quietly
// missing. internal/render calls it; nothing below composes a sentence.

import (
	"fmt"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// ResolveSections is cmd_governance's `resolve` action (`:335-346`), as records.
func ResolveSections(ws *workspace.Workspace, teamID string) ([]model.GateResult, error) {
	res, err := Resolve(ws, teamID)
	if err != nil {
		return nil, err
	}
	s := model.GateResult{Ordinal: 1, Slug: model.SlugResolve, Title: teamID}

	headline := model.Fields{"team": teamID, "components": len(res.components)}
	s.Findings = append(s.Findings,
		model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeGovernanceResolved,
			Subject:  teamID,
			Message:  Message(model.CodeGovernanceResolved, headline),
			Fields:   headline,
		},
		model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeGovernanceWrote,
			Path:     res.Rel,
			// `written` is not printed. It is the R-0.7c guard's answer — false
			// means the derived content was unchanged and the file was left
			// alone — and a UI that shows "wrote X" wants to know.
			Fields:  model.Fields{"path": res.Rel, "written": res.Written},
			Message: Message(model.CodeGovernanceWrote, model.Fields{"path": res.Rel}),
		})

	for _, c := range res.components {
		platforms := c.platforms
		if platforms == nil {
			platforms = []string{}
		}
		fields := model.Fields{
			"component":            c.id,
			"team":                 teamID,
			"platforms":            platforms,
			"companyControls":      c.company,
			"platformRequirements": c.platformNs,
		}
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeGovernanceComponent,
			Subject:  c.id,
			Message:  Message(model.CodeGovernanceComponent, fields),
			Fields:   fields,
		})
		if c.warning == "" {
			continue
		}
		warnFields := model.Fields{"component": c.id, "team": teamID, "warning": c.warning}
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevWarn,
			Code:     model.CodeGovernanceNoDescriptor,
			Subject:  c.id,
			Message:  Message(model.CodeGovernanceNoDescriptor, warnFields),
			Fields:   warnFields,
		})
	}
	return []model.GateResult{s}, nil
}

// DeclareSections is cmd_deviation (`:1112-1125`), as records.
func DeclareSections(ws *workspace.Workspace, teamID, rule, rationale string) ([]model.GateResult, error) {
	d, err := Declare(ws, teamID, rule, rationale)
	if err != nil {
		return nil, err
	}
	declared := model.Fields{"rule": d.Rule, "path": d.Rel, "team": d.Team}
	review := model.Fields{"reviewDate": d.ReviewDate, "team": d.Team}
	// The sentence reads "review due <date>; re-run: <command>"; R-3.6 wants the
	// command on its own.
	review[model.FieldNext] = "company-os governance resolve --team " + d.Team
	return []model.GateResult{{
		Ordinal: 1, Slug: model.SlugDeviation, Title: teamID,
		Findings: []model.Finding{
			{
				Severity: model.SevOK,
				Code:     model.CodeDeviationDeclared,
				Subject:  d.Rule,
				Path:     d.Rel,
				Message:  Message(model.CodeDeviationDeclared, declared),
				Fields:   declared,
			},
			{
				Severity: model.SevOK,
				Code:     model.CodeDeviationReviewDue,
				Message:  Message(model.CodeDeviationReviewDue, review),
				Fields:   review,
			},
		},
	}}, nil
}

// RequestSections is cmd_exception (`:1128-1138`), as records.
func RequestSections(ws *workspace.Workspace, teamID, rule, component, reason, expires string) ([]model.GateResult, error) {
	r, err := Request(ws, teamID, rule, component, reason, expires)
	if err != nil {
		return nil, err
	}
	drafted := model.Fields{"path": r.Rel, "expires": r.Expires}
	return []model.GateResult{{
		Ordinal: 1, Slug: model.SlugException, Title: teamID,
		Findings: []model.Finding{
			{
				Severity: model.SevOK,
				Code:     model.CodeExceptionDrafted,
				Path:     r.Rel,
				Message:  Message(model.CodeExceptionDrafted, drafted),
				Fields:   drafted,
			},
			{
				Severity: model.SevOK,
				Code:     model.CodeExceptionApproval,
				Message:  Message(model.CodeExceptionApproval, nil),
				Fields:   model.Fields{},
			},
		},
	}}, nil
}

// Message composes one governance line's sentence from its code and its typed
// fields, and from nothing else.
//
// CodeExplainRequirement is the one code whose Message is only half a line: the
// second line it renders ("applies because …") is the renderer's, because it is
// indented differently and the pair is one record. Both halves read the same
// Fields.
func Message(code string, f model.Fields) string {
	// validate's gates 1 and 2 are the same cluster asked a different question;
	// their eight sentences live in gates.go next to the loops that detect them.
	if s, ok := gateMessage(code, f); ok {
		return s
	}
	switch code {
	case model.CodeGovernanceResolved:
		return fmt.Sprintf("resolved governance for team '%s' (%d component(s))",
			f.Str("team"), f.Int("components"))
	case model.CodeGovernanceWrote:
		return "wrote " + f.Str("path")
	case model.CodeGovernanceComponent:
		return fmt.Sprintf("%s: platforms [%s], %d company + %d platform requirement(s)",
			f.Str("component"), platformList(f.Strs("platforms")),
			f.Int("companyControls"), f.Int("platformRequirements"))
	case model.CodeGovernanceNoDescriptor:
		return f.Str("component") + ": " + f.Str("warning")
	case model.CodeExplainComponent:
		return fmt.Sprintf("component '%s' (team %s):", f.Str("component"), f.Str("team"))
	case model.CodeExplainRequirement:
		deviated := ""
		if d, _ := f["deviated"].(bool); d {
			deviated = " [deviation applied]"
		}
		return fmt.Sprintf("%s v%s (%s)%s",
			f.Str("requirement"), f.Str("version"), f.Str("level"), deviated)
	case model.CodeDeviationDeclared:
		return fmt.Sprintf("declared deviation from %s in %s", f.Str("rule"), f.Str("path"))
	case model.CodeDeviationReviewDue:
		return fmt.Sprintf("review due %s; re-run: company-os governance resolve --team %s",
			f.Str("reviewDate"), f.Str("team"))
	case model.CodeExceptionDrafted:
		return fmt.Sprintf("exception drafted in %s (expires %s)",
			f.Str("path"), f.Str("expires"))
	case model.CodeExceptionApproval:
		return "note: mandatory rules require approval by the rule owner before this is valid."
	}
	return ""
}

// ExplainReason is the second line CodeExplainRequirement renders (`:363`).
func ExplainReason(f model.Fields) string {
	return fmt.Sprintf("applies because the component '%s' platform '%s'",
		f.Str("relationship"), f.Str("platform"))
}

// platformList is `", ".join(...) or "-"` (`:341`): an empty join is falsy and
// becomes a dash, so a component with no platform relationship still renders a
// complete sentence.
func platformList(ids []string) string {
	joined := strings.Join(ids, ", ")
	if joined == "" {
		return "-"
	}
	return joined
}
