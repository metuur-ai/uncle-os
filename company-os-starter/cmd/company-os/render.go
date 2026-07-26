package main

import (
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

// renderers is the text renderer per subcommand. A command with no entry emits
// nothing yet, which is where the unported commands sit.
var renderers = map[string]Renderer{
	"ids":    render.IDs,
	"today":  render.Today,
	"skills": render.Skills,
}
