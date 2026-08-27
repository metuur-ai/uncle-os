package product

// Scan helpers exported for `company-os next` (ux-simplification 1.1).
//
// These wrap the existing contract and done-check logic so the next-action
// scanner can reuse it without re-implementing any validation. Each function
// reads the same files the corresponding `prd validate` / `prd complete`
// command reads, and returns structured records the next package turns into
// priority-ranked actions.

import (
	"os"
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// ActivePRD is one active PRD with its contract and done-check status.
type ActivePRD struct {
	ID             string
	Platform       string
	Path           string // workspace-relative path to the PRD directory
	ContractIssues []Issue
	DoneProblems   []Issue
	MissingReality []string
}

// ActivePRDScan returns every active PRD across all platforms with its
// contract issues and done-check problems.
//
// It reuses CoreFieldErrors for the core-field contract, sectionIssues for
// the required headings, and doneCheck for the invariant-#4 gate. None of
// those are re-implemented here — the call chain is identical to what
// PRDValidate and PRDComplete already exercise.
func ActivePRDScan(ws *workspace.Workspace) ([]ActivePRD, error) {
	var prds []ActivePRD
	for _, pdir := range ws.AllPlatforms() {
		platform := filepath.Base(pdir)
		active := filepath.Join(pdir, "change-records", "active")
		entries, err := os.ReadDir(active)
		if err != nil {
			continue // absent active/ is not an error
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			prdPath := filepath.Join(active, e.Name(), "prd.md")
			if _, err := os.Stat(prdPath); err != nil {
				continue
			}
			meta, body, err := graph.ReadFrontmatter(prdPath)
			if err != nil {
				return nil, err
			}

			// Contract issues: core fields + required sections.
			issues := CoreFieldErrors(meta)
			for _, field := range processFields {
				if truthy(meta, field) && !yamlio.PyEqual(meta.Get(field), yamlio.PyStr("TODO")) {
					continue
				}
				issues = append(issues, Issue{
					Code:   model.CodePRDProcessField,
					Fields: model.Fields{"field": field},
				})
			}
			blocking, _ := sectionIssues(body, PRDSections)
			issues = append(issues, blocking...)

			// Done-check: unchecked items + stale reality.
			doneProblems, missingReality, err := doneCheck(ws, pdir, prdPath, meta, body)
			if err != nil {
				return nil, err
			}

			prds = append(prds, ActivePRD{
				ID:             e.Name(),
				Platform:       platform,
				Path:           relTo(ws.Root, prdPath),
				ContractIssues: issues,
				DoneProblems:   doneProblems,
				MissingReality: missingReality,
			})
		}
	}
	return prds, nil
}

// PendingOutcome is one archived PRD whose outcome review is still pending.
type PendingOutcome struct {
	PRD      string
	Platform string
	Due      string
	Path     string // workspace-relative path to outcome.md
}

// PendingOutcomeScan returns every archived PRD with a pending outcome review.
//
// It reads the same outcome.md frontmatter that `today` does, but returns
// structured records rather than findings. The caller decides whether a due
// date is "now" or "later" — this function reports all pending ones.
func PendingOutcomeScan(ws *workspace.Workspace) ([]PendingOutcome, error) {
	var outcomes []PendingOutcome
	for _, pdir := range ws.AllPlatforms() {
		platform := filepath.Base(pdir)
		arch := filepath.Join(pdir, "archive", "prds")
		entries, err := os.ReadDir(arch)
		if err != nil {
			continue // absent archive/ is not an error
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			outcomePath := filepath.Join(arch, e.Name(), "outcome.md")
			if _, err := os.Stat(outcomePath); err != nil {
				continue
			}
			meta, _, err := graph.ReadFrontmatter(outcomePath)
			if err != nil {
				return nil, err
			}
			if strOf(meta, "status") != "pending" {
				continue
			}
			outcomes = append(outcomes, PendingOutcome{
				PRD:      e.Name(),
				Platform: platform,
				Due:      strOf(meta, "due"),
				Path:     relTo(ws.Root, outcomePath),
			})
		}
	}
	return outcomes, nil
}
