package main

import (
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Command is the dispatch seam. Every subcommand returns records and an error;
// main is the only place that turns either into bytes and an exit code. This is
// what lets the TUI and --json reuse one code path, and it is the structural
// difference from bin/company-os, where die() (:41-55) exits from anywhere in
// the call tree.
//
// out is the stdout writer. It is a parameter rather than a package global
// because the commands that print prose rather than findings — every mutating
// command, which R-1.8 requires to print its next step — still may not reach
// for os.Stdout themselves: they format a record the internal package returned.
type Command func(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error)

// commands maps each subcommand name to its implementation. Every entry is a
// stub until its phase lands; the names, flags, and choice sets are already
// final (task 1.1 is surface only).
var commands = map[string]Command{
	"init":       cmdInit,
	"add":        cmdAdd,
	"reality":    cmdReality,
	"discover":   cmdDiscover,
	"prd":        cmdPRD,
	"governance": cmdGovernance,
	"check":      cmdCheck,
	"validate":   cmdValidate,
	"deviation":  cmdDeviation,
	"exception":  cmdException,
	"scratchpad": cmdScratchpad,
	"today":      cmdToday,
	"graph":      cmdGraph,
	"ids":        cmdIDs,
	"skills":     cmdSkills,
	"workspace":  cmdWorkspace,
}

// `tui` is registered here rather than in the literal above because R-5.12
// makes it self-referential: the TUI executes screens by dispatching back
// through this very map, so cmdTUI reads `commands` and `commands` would name
// cmdTUI — an initialization cycle the compiler rejects. Assigning after the
// literal is initialized breaks it without weakening the property. The cycle is
// the point: there is exactly one dispatch table, and the UI is a caller of it.
func init() { commands["tui"] = cmdTUI }

// notImplemented is the stub for a subcommand whose phase has not landed.
//
// The code is 2, not 1. Code 1 means "a validate subcommand reported [FAIL]"
// (.devlocal/go-port/exit-code-map.md § H), so a stubbed `company-os validate`
// returning 1 was byte-indistinguishable from a real gate failure: anyone
// pointing CI at this binary mid-port got a red that looked exactly like a
// governance red, and — worse — a green the day the stub was replaced would look
// like a fix. 2 says what is true, that this invocation is not something the
// binary can act on, and no CI branch reads it as a verdict.
func notImplemented(name string) Command {
	return func(*workspace.Workspace, *Args, io.Writer) ([]model.GateResult, error) {
		return nil, model.Errorf(model.ExitUsage,
			"%s: not implemented in this build", name)
	}
}
