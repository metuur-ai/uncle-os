package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Governance writes the four governance-cluster commands as the Python CLI
// writes them: `governance resolve` (bin/company-os:335-346), `governance
// explain` (`:348-363`), `deviation declare` (`:1122-1124`) and `exception
// request` (`:1137-1138`).
//
// One renderer serves all four because they share one line grammar — a flat
// block with no gate headers and no blank lines anywhere. What differs is the
// indent per code, which is exactly what the switch below carries.
//
// Two of the four print no next-step line. R-1.9 keeps it that way: `governance
// resolve` and `exception request` end where the oracle ends them, and R-0.8
// outranks R-1.8. Do not complete the chain here.
func Governance(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		for _, f := range s.Findings {
			if err := governanceLine(w, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func governanceLine(w io.Writer, f model.Finding) error {
	switch f.Code {
	case model.CodeGovernanceResolved,
		model.CodeGovernanceWrote,
		model.CodeExplainComponent,
		model.CodeDeviationDeclared,
		model.CodeDeviationReviewDue,
		model.CodeExceptionDrafted,
		model.CodeExceptionApproval:
		return writeLines(w, f.Message)

	case model.CodeGovernanceComponent:
		return writeLines(w, "  "+f.Message)

	case model.CodeGovernanceNoDescriptor:
		// warn() writes to stdout with a two-space indent
		// (bin/company-os:50-51), so this line is part of the view.
		return writeLines(w, "  [warn] "+f.Message)

	case model.CodeExplainRequirement:
		// One record, two lines: the requirement and the sentence explaining
		// why it applies. The second is six-space indented and never appears
		// without the first.
		return writeLines(w, "  - "+f.Message, "      "+governance.ExplainReason(f.Fields))
	}
	return fmt.Errorf("render: governance: no rule for finding code %q", f.Code)
}
