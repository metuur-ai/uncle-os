// Package next implements `company-os next` (ux-simplification 1.1): a
// read-only scan that names the single highest-priority pending action and
// the exact command to perform it.
//
// Priority order (per docs/tasks/ux-simplification.md):
//  1. expired or soon-due deviation reviewDate / exception expires
//  2. active PRD failing its artifact contract
//  3. active PRD with unchecked checklist items and/or stale reality
//  4. completed PRD with an outcome review due
//  5. nothing pending → suggest `company-os discover new`
//
// The package reuses existing scan functions — governance.ExpiryScan for
// escape-hatch expiry, product.ActivePRDScan for active-PRD contract and
// done-check, product.PendingOutcomeScan for archived outcome reviews —
// rather than re-implementing any validation logic.
package next

import (
	"fmt"
	"sort"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/product"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Scan returns the next-action records. When all is false, only the single
// highest-priority action is returned; when true, every pending item is
// grouped by kind.
func Scan(ws *workspace.Workspace, all bool) ([]model.GateResult, error) {
	// Priority 1: expired or soon-due escape hatches.
	expiryItems, err := governance.ExpiryScan(ws, governance.SoonDueWindow)
	if err != nil {
		return nil, err
	}
	var expiryPending []governance.ExpiryItem
	for _, it := range expiryItems {
		if it.Status == "expired" || it.Status == "soon-due" {
			expiryPending = append(expiryPending, it)
		}
	}

	// Priority 2+3: active PRDs — contract issues and done-check problems.
	activePRDs, err := product.ActivePRDScan(ws)
	if err != nil {
		return nil, err
	}
	var contractPRDs, doneCheckPRDs []product.ActivePRD
	for _, prd := range activePRDs {
		if len(prd.ContractIssues) > 0 {
			contractPRDs = append(contractPRDs, prd)
		}
		if len(prd.DoneProblems) > 0 || len(prd.MissingReality) > 0 {
			doneCheckPRDs = append(doneCheckPRDs, prd)
		}
	}

	// Priority 4: pending outcome reviews.
	outcomes, err := product.PendingOutcomeScan(ws)
	if err != nil {
		return nil, err
	}

	// Single-action mode: return the first non-empty priority.
	if !all {
		return singleAction(ws, expiryPending, contractPRDs, doneCheckPRDs, outcomes)
	}
	return allActions(ws, expiryPending, contractPRDs, doneCheckPRDs, outcomes)
}

// singleAction returns one GateResult carrying the highest-priority pending
// action. Priority 5 (nothing pending) suggests `discover new`.
func singleAction(ws *workspace.Workspace,
	expiry []governance.ExpiryItem,
	contract, doneCheck []product.ActivePRD,
	outcomes []product.PendingOutcome,
) ([]model.GateResult, error) {

	s := model.GateResult{Ordinal: 1, Slug: model.SlugNext, Title: "next"}

	// Priority 1: expiry.
	if len(expiry) > 0 {
		it := expiry[0]
		// Sort expired before soon-due, then by date ascending so the most
		// urgent is first.
		sort.Slice(expiry, func(i, j int) bool {
			if expiry[i].Status != expiry[j].Status {
				return expiry[i].Status == "expired"
			}
			return expiry[i].Date < expiry[j].Date
		})
		it = expiry[0]
		cmd := fmt.Sprintf("company-os %s declare %s --team %s", it.Kind, it.Rule, it.Team)
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevFail,
			Code:     model.CodeNextExpiry,
			Subject:  it.Rule,
			Message:  expiryMessage(it),
			Fields: model.Fields{
				"kind": it.Kind, "team": it.Team, "rule": it.Rule,
				"date": it.Date, "status": it.Status,
				model.FieldNext: cmd,
			},
		})
		return []model.GateResult{s}, nil
	}

	// Priority 2: contract issues.
	if len(contract) > 0 {
		prd := contract[0]
		cmd := fmt.Sprintf("company-os prd validate --platform %s %s", prd.Platform, prd.ID)
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevFail,
			Code:     model.CodeNextContract,
			Subject:  prd.ID,
			Path:     prd.Path,
			Message: fmt.Sprintf("PRD %s has %d contract issue(s)",
				prd.ID, len(prd.ContractIssues)),
			Fields: model.Fields{
				"prd": prd.ID, "platform": prd.Platform,
				"issues":        len(prd.ContractIssues),
				model.FieldNext: cmd,
			},
		})
		return []model.GateResult{s}, nil
	}

	// Priority 3: done-check problems.
	if len(doneCheck) > 0 {
		prd := doneCheck[0]
		cmd := fmt.Sprintf("company-os prd complete --platform %s %s", prd.Platform, prd.ID)
		unchecked := 0
		for _, p := range prd.DoneProblems {
			if p.Code == model.CodeDoneChecklistUnchecked {
				unchecked += p.Fields.Int("count")
			}
		}
		stale := len(prd.MissingReality)
		msg := fmt.Sprintf("PRD %s has %d unchecked item(s) and %d stale/missing reality doc(s)",
			prd.ID, unchecked, stale)
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevWarn,
			Code:     model.CodeNextDoneCheck,
			Subject:  prd.ID,
			Path:     prd.Path,
			Message:  msg,
			Fields: model.Fields{
				"prd": prd.ID, "platform": prd.Platform,
				"unchecked": unchecked, "stale": stale,
				model.FieldNext: cmd,
			},
		})
		return []model.GateResult{s}, nil
	}

	// Priority 4: pending outcome reviews.
	if len(outcomes) > 0 {
		o := outcomes[0]
		// Sort by due date ascending so the most urgent is first.
		sort.Slice(outcomes, func(i, j int) bool {
			return outcomes[i].Due < outcomes[j].Due
		})
		o = outcomes[0]
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevWarn,
			Code:     model.CodeNextOutcome,
			Subject:  o.PRD,
			Path:     o.Path,
			Message:  fmt.Sprintf("outcome review for %s due %s", o.PRD, o.Due),
			Fields: model.Fields{
				"prd": o.PRD, "platform": o.Platform, "due": o.Due,
			},
		})
		return []model.GateResult{s}, nil
	}

	// Priority 5: nothing pending.
	s.Findings = append(s.Findings, model.Finding{
		Severity: model.SevOK,
		Code:     model.CodeNextEmpty,
		Message:  "no pending actions; start with company-os discover new",
		Fields: model.Fields{
			model.FieldNext: "company-os discover new --team <team> \"<title>\"",
		},
	})
	return []model.GateResult{s}, nil
}

// allActions returns one GateResult per priority kind that has pending items,
// plus a trailing empty-action finding when nothing is pending.
func allActions(ws *workspace.Workspace,
	expiry []governance.ExpiryItem,
	contract, doneCheck []product.ActivePRD,
	outcomes []product.PendingOutcome,
) ([]model.GateResult, error) {

	s := model.GateResult{Ordinal: 1, Slug: model.SlugNext, Title: "next"}
	any := false

	// Sort expired before soon-due, then by date ascending.
	sort.Slice(expiry, func(i, j int) bool {
		if expiry[i].Status != expiry[j].Status {
			return expiry[i].Status == "expired"
		}
		return expiry[i].Date < expiry[j].Date
	})

	if len(expiry) > 0 {
		any = true
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeNextGroup,
			Message:  "expired/soon-due escape hatches:",
			Fields:   model.Fields{"group": "expiry"},
		})
		for _, it := range expiry {
			cmd := fmt.Sprintf("company-os %s declare %s --team %s", it.Kind, it.Rule, it.Team)
			s.Findings = append(s.Findings, model.Finding{
				Severity: model.SevFail,
				Code:     model.CodeNextExpiry,
				Subject:  it.Rule,
				Message:  expiryMessage(it),
				Fields: model.Fields{
					"kind": it.Kind, "team": it.Team, "rule": it.Rule,
					"date": it.Date, "status": it.Status,
					"grouped":       true,
					model.FieldNext: cmd,
				},
			})
		}
	}

	if len(contract) > 0 {
		any = true
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeNextGroup,
			Message:  "active PRDs with contract issues:",
			Fields:   model.Fields{"group": "contract"},
		})
		for _, prd := range contract {
			cmd := fmt.Sprintf("company-os prd validate --platform %s %s", prd.Platform, prd.ID)
			s.Findings = append(s.Findings, model.Finding{
				Severity: model.SevFail,
				Code:     model.CodeNextContract,
				Subject:  prd.ID,
				Path:     prd.Path,
				Message: fmt.Sprintf("%s (platform %s): %d issue(s)",
					prd.ID, prd.Platform, len(prd.ContractIssues)),
				Fields: model.Fields{
					"prd": prd.ID, "platform": prd.Platform,
					"issues":        len(prd.ContractIssues),
					"grouped":       true,
					model.FieldNext: cmd,
				},
			})
		}
	}

	if len(doneCheck) > 0 {
		any = true
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeNextGroup,
			Message:  "active PRDs with incomplete done-check:",
			Fields:   model.Fields{"group": "done-check"},
		})
		for _, prd := range doneCheck {
			cmd := fmt.Sprintf("company-os prd complete --platform %s %s", prd.Platform, prd.ID)
			unchecked := 0
			for _, p := range prd.DoneProblems {
				if p.Code == model.CodeDoneChecklistUnchecked {
					unchecked += p.Fields.Int("count")
				}
			}
			stale := len(prd.MissingReality)
			s.Findings = append(s.Findings, model.Finding{
				Severity: model.SevWarn,
				Code:     model.CodeNextDoneCheck,
				Subject:  prd.ID,
				Path:     prd.Path,
				Message: fmt.Sprintf("%s (platform %s): %d unchecked item(s), %d stale/missing reality doc(s)",
					prd.ID, prd.Platform, unchecked, stale),
				Fields: model.Fields{
					"prd": prd.ID, "platform": prd.Platform,
					"unchecked": unchecked, "stale": stale,
					"grouped":       true,
					model.FieldNext: cmd,
				},
			})
		}
	}

	if len(outcomes) > 0 {
		any = true
		sort.Slice(outcomes, func(i, j int) bool {
			return outcomes[i].Due < outcomes[j].Due
		})
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeNextGroup,
			Message:  "pending outcome reviews:",
			Fields:   model.Fields{"group": "outcome"},
		})
		for _, o := range outcomes {
			s.Findings = append(s.Findings, model.Finding{
				Severity: model.SevWarn,
				Code:     model.CodeNextOutcome,
				Subject:  o.PRD,
				Path:     o.Path,
				Message:  fmt.Sprintf("%s (platform %s): due %s", o.PRD, o.Platform, o.Due),
				Fields: model.Fields{
					"prd": o.PRD, "platform": o.Platform, "due": o.Due,
					"grouped": true,
				},
			})
		}
	}

	if !any {
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeNextEmpty,
			Message:  "no pending actions; start with company-os discover new",
			Fields: model.Fields{
				model.FieldNext: "company-os discover new --team <team> \"<title>\"",
			},
		})
	}

	return []model.GateResult{s}, nil
}

// expiryMessage composes the human-readable sentence for one expiry item.
func expiryMessage(it governance.ExpiryItem) string {
	switch it.Status {
	case "expired":
		return fmt.Sprintf("%s for %s expired %s — re-review or remove",
			it.Kind, it.Rule, it.Date)
	case "soon-due":
		return fmt.Sprintf("%s for %s due %s — re-review or remove",
			it.Kind, it.Rule, it.Date)
	}
	return ""
}
