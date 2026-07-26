package main

import (
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/validate"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdValidate is `company-os validate` (bin/company-os:922-1107).
//
// It takes no flags and returns every gate. The exit status is not decided here:
// run() maps a record set containing any [FAIL] onto ExitValidation (1), which
// is `sys.exit(0 if problems == 0 else 1)` at `:1107`.
func cmdValidate(ws *workspace.Workspace, _ *Args, _ io.Writer) ([]model.GateResult, error) {
	return validate.Run(ws)
}
