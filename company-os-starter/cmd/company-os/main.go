// Command company-os is the Company OS federation CLI.
//
// This package is the only one in the module allowed to call os.Exit or write to
// stdout. Everything below cmd/ returns records and errors; the mapping from
// those to output and to one of the eight exit codes happens here and nowhere
// else. internal/architecture_test.go enforces that.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main's testable body. It returns the process exit code and never exits.
func run(argv []string, stdout, stderr io.Writer) int {
	args, err := parse(argv)
	var hr *helpRequest
	switch {
	case errors.As(err, &hr):
		fmt.Fprint(stdout, help(hr.scope))
		return int(model.ExitOK)
	case errors.Is(err, errVersion):
		fmt.Fprint(stdout, versionLine())
		return int(model.ExitOK)
	case err != nil:
		// A bare invocation lands here too: argparse prints usage and exits 2.
		// Both lines are scoped to the parser that caught the mistake, so a
		// sub-parser error names the subcommand and shows *its* usage rather
		// than the top-level one (R-1.4a).
		var ue *usageError
		if errors.As(err, &ue) {
			writeUsageError(stderr, ue.scope, ue.Error())
			return int(model.CodeOf(err))
		}
		fmt.Fprint(stderr, usage())
		fmt.Fprintf(stderr, "company-os: error: %v\n", err)
		return int(model.CodeOf(err))
	}

	ws := workspace.New(workspace.Resolve(args.Root))

	// `scratchpad` operates on --repo (any repo) and `init` creates a workspace
	// where none exists yet, so both are exempt; every other command scans the
	// workspace and must fail-fast outside a workspace root (bin/company-os:2774).
	//
	// `tui` is exempt here and asserts the root itself, one line after its TTY
	// check. R-5.3 is unconditional — no TTY means exit 7 — and checking the
	// root first would make the same non-interactive invocation exit 7 inside a
	// workspace and 3 outside one.
	if args.Cmd != "scratchpad" && args.Cmd != "init" && args.Cmd != "tui" {
		if err := ws.RequireRoot(); err != nil {
			return fail(args, stdout, stderr, err)
		}
	}

	cmd, ok := commands[args.Cmd]
	if !ok {
		return fail(args, stdout, stderr,
			model.Errorf(model.ExitUsage, "no handler registered for %q", args.Cmd))
	}

	results, err := cmd(ws, args, stdout)

	// Rendering precedes the error branch because a command may return records
	// AND an error: those records are output the oracle had already printed
	// before it refused — `prd complete`'s done-check block (bin/company-os:700),
	// `prd new`'s unresolved-component warnings (`:605`), `workspace sync`'s
	// completed repos (`:2589`) and now `validate`'s completed gates (R-2.6a),
	// all on stdout, all followed by a non-zero exit. Dropping them on the error
	// path would lose a whole stdout block.
	//
	// `--json` replaces the text renderer rather than joining it (R-3.2): one
	// document on stdout and no prose, error path included (R-3.8).
	if args.JSON {
		payload := render.Result{
			Command: args.Cmd, Action: args.Action, Root: ws.Root,
			Sections: results, Err: err, ExitCode: exitCode(results, err),
		}
		if jerr := render.JSON(stdout, payload); jerr != nil {
			fmt.Fprintf(stderr, "error: %v\n", jerr)
			return int(model.CodeOf(jerr))
		}
	} else if r, ok := renderers[args.Cmd]; ok && len(results) > 0 {
		if rerr := r(stdout, results); rerr != nil {
			fmt.Fprintf(stderr, "error: %v\n", rerr)
			return int(model.CodeOf(rerr))
		}
	}
	if err != nil {
		// A conditional-requirement check argparse could not express still
		// reports as an argument error: the sub-parser's usage line, then the
		// `company-os <sub>: error: …` diagnostic (R-1.4a). Without this the
		// three such checks in command code emitted a bare `error: …` line that
		// the greppable selector task 1.1a pinned does not match.
		var ue *model.UsageError
		if errors.As(err, &ue) {
			writeUsageError(stderr, ue.Scope, ue.Error())
			return int(model.CodeOf(err))
		}
		// A quiet error has already said what it had to say, on the command's
		// own output stream. `prd complete`'s refusal is the only one: it prints
		// its block to stdout and writes nothing to stderr.
		if !model.IsQuiet(err) {
			fmt.Fprintf(stderr, "error: %v\n", err)
		}
		return int(model.CodeOf(err))
	}
	return int(exitCode(results, err))
}

// exitCode is the process status for a completed dispatch: the error's code when
// there is one, ExitValidation when any gate reported [FAIL], otherwise 0.
//
// It exists so that `--json` can PUBLISH the code in the same document it
// publishes the findings in (R-3.8) — the text path decides the same thing on
// its way out, and two independent decisions would eventually disagree.
func exitCode(results []model.GateResult, err error) model.ExitCode {
	switch {
	case err != nil:
		return model.CodeOf(err)
	case model.HasFailure(results):
		return model.ExitValidation
	}
	return model.ExitOK
}

// fail reports an error raised BEFORE any command ran — not a workspace root, no
// handler. The diagnostic goes to stderr either way (R-3.9); under `--json`
// stdout still gets a valid document rather than nothing at all (R-3.8), because
// a consumer that has to distinguish "empty stdout" from "crashed" has no
// contract.
func fail(args *Args, stdout, stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "error: %v\n", err)
	if args.JSON {
		_ = render.JSON(stdout, render.Result{
			Command: args.Cmd, Action: args.Action, Root: workspace.Resolve(args.Root),
			Err: err, ExitCode: model.CodeOf(err),
		})
	}
	return int(model.CodeOf(err))
}

// versionLine is what `company-os --version` prints (R-6.8). One line, four
// facts, in the order a bug report needs them.
//
// `version` alone was not enough: the Makefile derives it from
// `git describe --tags --always --dirty`, so on an untagged tree it degrades to
// a bare abbreviated sha and the reader cannot tell a release name from a commit
// — and once a tag exists, the tag alone no longer identifies the tree it was
// built from. Version and commit together are the "version and build
// identifier" R-6.8 asks for; the Go version and platform are there because the
// only three artifacts R-6.2 ships differ by exactly those, and both are
// compile-time constants that cost no reproducibility.
//
// Wording lives here, not in internal/model, per R-2.8.
func versionLine() string {
	b := model.BuildInfo()
	return fmt.Sprintf("company-os %s (commit %s, %s, %s)\n",
		b.Version, b.Commit, b.GoVersion, b.Platform)
}

// help renders `--help` for one parser: the top-level one when scope is empty,
// otherwise the named subcommand's.
//
// Sub-parser help lists that subcommand's own usage, positionals, and flags and
// nothing else, which is what argparse does and what the reader asked for.
// R-0.7a(i) waives byte-identity with argparse's LAYOUT — specifically the usage
// line's wrap to $COLUMNS — not the content, so every help= string in the
// oracle's parser is reproduced verbatim and the gutter is argparse's own.
func help(scope string) string {
	spec, ok := lookupCommand(scope)
	if scope == "" || !ok {
		return usage()
	}
	// No description line: none of the sub-parsers sets `description=`, so
	// argparse prints none, and adding one would add a human-facing line R-0.8
	// freezes. Section bodies list metavars, not usage fragments — argparse
	// drops the `[…]` optionality brackets there.
	type action struct{ invocation, help string }
	positionals := make([]action, 0, len(spec.pos))
	for _, p := range spec.pos {
		positionals = append(positionals, action{posMetavar(p), p.help})
	}
	options := []action{{"-h, --help", "show this help message and exit"}}
	for _, f := range spec.flags {
		options = append(options, action{strings.Trim(flagUsage(f), "[]"), f.help})
	}

	// argparse's gutter: `help_position = min(longest_invocation + 2 + 2, 24)`,
	// computed across EVERY action in the parser. That is why `validate --help`
	// puts its help text at column 14 and `prd --help` at 24 — one long flag
	// name moves the whole column for the parser it belongs to.
	width := 0
	for _, a := range append(append([]action{}, positionals...), options...) {
		if len(a.invocation) > width {
			width = len(a.invocation)
		}
	}
	if width += 4; width > 24 {
		width = 24
	}
	width -= 4

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", usageLine(scope))
	section := func(title string, actions []action) {
		fmt.Fprintf(&b, "\n%s:\n", title)
		for _, a := range actions {
			switch {
			case a.help == "":
				fmt.Fprintf(&b, "  %s\n", a.invocation)
			case len(a.invocation) > width:
				// argparse breaks to the next line rather than pushing the
				// gutter out for one long invocation.
				fmt.Fprintf(&b, "  %s\n%*s%s\n", a.invocation, width+4, "", a.help)
			default:
				fmt.Fprintf(&b, "  %-*s  %s\n", width, a.invocation, a.help)
			}
		}
	}
	if len(positionals) > 0 {
		section("positional arguments", positionals)
	}
	section("options", options)
	return b.String()
}

func usage() string {
	var b strings.Builder
	b.WriteString("usage: company-os [--root ROOT] [--json] [--version] <command> ...\n\n")
	b.WriteString("  --root ROOT   workspace root " +
		"(default: $COMPANY_OS_WORKSPACE_ROOT or cwd)\n")
	b.WriteString("  --json        emit structured output instead of text\n")
	b.WriteString("  --version     print the version and exit\n\n")
	b.WriteString("commands:\n")
	width := 0
	for _, s := range commandSpecs {
		if len(s.name) > width {
			width = len(s.name)
		}
	}
	for _, s := range commandSpecs {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, s.name, s.help)
	}
	return b.String()
}
