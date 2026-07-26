package main

// `skills list` — the merged view across the four skill layers.

import (
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdSkills is cmd_skills (bin/company-os:869-917). Like the other read-only
// views it formats nothing: internal/skills returns the record set and
// render.Skills turns it into bytes, so `out` goes unused here.
func cmdSkills(ws *workspace.Workspace, _ *Args, _ io.Writer) ([]model.GateResult, error) {
	return skills.List(ws)
}
