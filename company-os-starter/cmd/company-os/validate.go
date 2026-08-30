package main

import (
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/validate"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdValidate is `company-os validate` (bin/company-os:922-1107).
//
// It takes one optional flag, --fix (ux-simplification 2.1). Without it the
// behaviour is byte-identical to the pre-flag CLI: the gates run against
// whatever the tree carries and the renderer produces the same bytes. With
// --fix, FixDerived regenerates every derived artifact through the SAME code
// paths `governance resolve` and `graph build` use, and the normal gates then
// run against the refreshed tree. The gate report is unchanged in shape; one
// summary line and one JSON field carry the count of files whose bytes
// actually changed.
//
// The exit status is not decided here: run() maps a record set containing any
// [FAIL] onto ExitValidation (1), which is `sys.exit(0 if problems == 0 else
// 1)` at `:1107`.
func cmdValidate(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	if args.Fix {
		n, err := validate.FixDerived(ws)
		if err != nil {
			return nil, err
		}
		results, err := validate.Run(ws)
		if err != nil {
			return results, err
		}
		// Append the fix-summary section. It is not a gate — Ordinal 0 and
		// the dedicated slug keep the renderer from numbering it, and the
		// JSON encoder from treating it as one.
		results = append(results, model.GateResult{
			Slug: model.SlugFixSummary,
			Findings: []model.Finding{{
				Severity: model.SevOK,
				Code:     model.CodeFixRegenerated,
				Fields:   model.Fields{"regenerated": n},
			}},
		})
		return results, nil
	}
	return validate.Run(ws)
}
