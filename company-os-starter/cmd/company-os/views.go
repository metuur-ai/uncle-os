package main

// The two read-only role/ontology views: `ids list` and `today`.
//
// Neither handler formats anything. Both return the records their internal
// package computed and let `renderers` turn them into bytes, which is what keeps
// the same record set available to --json and to the TUI without a second
// traversal of the workspace (R-2.9). `out` is unused here for that reason.

import (
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/ids"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/roles"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdIDs is cmd_ids (bin/company-os:1275-1302).
func cmdIDs(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	return ids.List(ws, args.Role, ids.Filter{
		Prefix:   args.Prefix,
		Team:     args.Team,
		Platform: args.Platform,
	})
}

// cmdToday is cmd_today (bin/company-os:1168-1203).
func cmdToday(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	return roles.Today(ws, args.Role)
}
