package main

// The governance cluster's dispatch handlers: governance resolve/explain,
// deviation declare and exception request.
//
// None of them formats anything — internal/governance returns the record set
// and render.Governance turns it into bytes — so `out` goes unused here, as it
// does for the other record-returning commands.

import (
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdGovernance is cmd_governance (bin/company-os:331-370).
func cmdGovernance(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	if args.Action == "resolve" {
		if args.Team == "" {
			// `--team` is not marked required in the oracle's parser, so
			// `governance resolve` with no team reaches `ws.team_dir(None)` and
			// raises `TypeError: unsupported operand type(s) for /: 'PosixPath'
			// and 'NoneType'` — a traceback, exit 1, nothing written.
			//
			// Reproducing the traceback is neither possible nor desirable, but
			// the FILESYSTEM effect is not carved out (R-0.7a(j)): without this
			// guard Go resolves teams/ itself as the team directory and writes
			// teams/generated/effective-governance.yaml, a file the oracle never
			// creates. The diagnostic is argparse's own wording for a required
			// argument, which is what the flag should have been. Exit 2 is the
			// contract's own category for it: "a missing required argument".
			return nil, model.Usagef("governance",
				"the following arguments are required: --team")
		}
		return governance.ResolveSections(ws, args.Team)
	}
	return governance.Explain(ws, explainTarget(args.ComponentArg))
}

// explainTarget renders the optional `component` positional the way the oracle's
// f-strings do.
//
// argparse defaults an unsupplied `nargs="?"` positional to None, and every
// place `explain` uses it — the die() message and `suggest_ids`, which calls
// `str(unknown)` — renders that as the four characters "None". The Go parser
// has no None, so the empty string is mapped here, at the argparse seam, rather
// than teaching internal/governance about a sentinel it has no other use for.
func explainTarget(component string) string {
	if component == "" {
		return "None"
	}
	return component
}

// cmdDeviation is cmd_deviation (bin/company-os:1112-1125). It validates
// nothing: a deviation aimed at a mandatory rule is accepted here and exits 0,
// and the refusal surfaces later through `governance resolve` and `validate`.
func cmdDeviation(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	return governance.DeclareSections(ws, args.Team, args.Rule, args.Rationale)
}

// cmdException is cmd_exception (bin/company-os:1128-1138).
func cmdException(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	return governance.RequestSections(ws, args.Team, args.Rule, args.Component,
		args.Reason, args.Expires)
}
