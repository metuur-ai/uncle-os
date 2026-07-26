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
	"discover":   notImplemented("discover"),
	"prd":        notImplemented("prd"),
	"governance": notImplemented("governance"),
	"check":      notImplemented("check"),
	"validate":   notImplemented("validate"),
	"deviation":  notImplemented("deviation"),
	"exception":  notImplemented("exception"),
	"scratchpad": cmdScratchpad,
	"today":      cmdToday,
	"graph":      notImplemented("graph"),
	"ids":        cmdIDs,
	"skills":     cmdSkills,
	"workspace":  notImplemented("workspace"),
}

func notImplemented(name string) Command {
	return func(*workspace.Workspace, *Args, io.Writer) ([]model.GateResult, error) {
		return nil, model.Errorf(model.ExitValidation, "%s: not implemented", name)
	}
}
