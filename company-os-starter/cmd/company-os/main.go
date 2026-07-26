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
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main's testable body. It returns the process exit code and never exits.
func run(argv []string, stdout, stderr io.Writer) int {
	args, err := parse(argv)
	switch {
	case errors.Is(err, errHelp):
		fmt.Fprint(stdout, usage())
		return int(model.ExitOK)
	case errors.Is(err, errVersion):
		fmt.Fprintf(stdout, "company-os %s\n", version)
		return int(model.ExitOK)
	case err != nil:
		// A bare invocation lands here too: argparse prints usage and exits 2.
		fmt.Fprint(stderr, usage())
		fmt.Fprintf(stderr, "company-os: error: %v\n", err)
		return int(model.CodeOf(err))
	}

	ws := workspace.New(workspace.Resolve(args.Root))

	// `scratchpad` operates on --repo (any repo) and `init` creates a workspace
	// where none exists yet, so both are exempt; every other command scans the
	// workspace and must fail-fast outside a workspace root (bin/company-os:2774).
	if args.Cmd != "scratchpad" && args.Cmd != "init" {
		if err := ws.RequireRoot(); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return int(model.CodeOf(err))
		}
	}

	cmd, ok := commands[args.Cmd]
	if !ok {
		fmt.Fprintf(stderr, "error: no handler registered for %q\n", args.Cmd)
		return int(model.ExitUsage)
	}

	results, err := cmd(ws, args, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return int(model.CodeOf(err))
	}

	// The JSON renderer lands in Phase 4; the gate renderer in Phase 3. Until
	// then a command without an entry in `renderers` returns no records.
	if r, ok := renderers[args.Cmd]; ok {
		if err := r(stdout, results); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return int(model.CodeOf(err))
		}
	}
	if model.HasFailure(results) {
		return int(model.ExitValidation)
	}
	return int(model.ExitOK)
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
