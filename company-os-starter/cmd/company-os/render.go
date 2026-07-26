package main

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
)

// Renderer turns one command's records into bytes. It is per-command rather than
// global because the commands do not share a line grammar: `validate` prints
// gate headers and severity markers, `today` prints a banner and indented rows,
// `ids list` prints neither. What they do share is that no package below cmd/
// composes any of those sentences (R-2.8) — the record set is the contract, and
// the renderer is the only thing that knows the wording.
type Renderer func(io.Writer, []model.GateResult) error

// renderers is the text renderer per subcommand. Every subcommand has one.
//
// This map is the TEXT side only. `--json` has no per-command entry and never
// will: one encoder covers the whole surface because every command returns the
// same record type (R-3.4b, internal/render/json.go).
var renderers = map[string]Renderer{
	"graph":      render.Graph,
	"ids":        render.IDs,
	"today":      render.Today,
	"skills":     render.Skills,
	"governance": render.Governance,
	"validate":   render.Validate,
	"discover":   render.Product,
	"prd":        render.Product,
	"check":      render.Product,
	"deviation":  render.Governance,
	"exception":  render.Governance,
	"init":       renderPlain,
	"add":        renderPlain,
	"reality":    renderPlain,
	"scratchpad": renderPlain,
	"workspace":  renderPlain,
}

// renderPlain writes each finding's Message on its own line and applies no
// grammar of its own.
//
// The five commands it serves — the four scaffolding ones and `workspace` —
// have no line grammar worth factoring: no severity markers, no gate headers, no
// per-code indent, and the two blank lines `workspace` emits are literal `\n`s
// inside the oracle's own f-strings. Their Message is the finished line, so this
// is the whole renderer.
func renderPlain(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		for _, f := range s.Findings {
			if _, err := fmt.Fprintln(w, f.Message); err != nil {
				return err
			}
		}
	}
	return nil
}
