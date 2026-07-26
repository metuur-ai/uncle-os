package main

// `company-os tui` — the TTY gate and the read-only screen catalog (Unit 5).
//
// Two things live here, and they are here rather than in internal/tui on
// purpose.
//
// The GATE is here because only package main may decide the process's fate: it
// probes the real terminal and returns an exit-7 error, which run() turns into a
// stderr line and a status the same way it turns every other coded error into
// one (R-5.3).
//
// The CATALOG is here because it is the only place that can reach both
// `commands` and `renderers`. Every screen executes the same Command function
// the flag CLI executes and formats it with the same Renderer — in-process, no
// shelling out, no parsing of its own output (R-5.12, R-5.13). internal/tui
// never sees a GateResult, so it cannot render one differently.
//
// This file must stay free of styling imports. ansi_test.go exempts
// internal/tui and nothing else, and a lipgloss import HERE would be exempt from
// nothing while sitting one call away from every golden the CLI produces.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// interactive reports whether both ends of the conversation are a real terminal.
//
// It is a variable so a test can assert the gate without owning a pty; the
// production value is the termios probe in tty.go, which is a real ioctl rather
// than an os.Stat/ModeCharDevice test — /dev/null is a character device too, and
// a CI job redirecting from it would otherwise be classified as interactive
// (see tty.go for the measurement).
//
// BOTH streams are required. Input alone is not enough: a TUI whose frames are
// being written into a file or a pipe has nobody to read them, and would fill
// that file with cursor-movement escapes until it was killed.
var interactive = func() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

// stdin is the reader the program is driven from, indirected for the same
// reason.
var stdin io.Reader = os.Stdin

// cmdTUI is `company-os tui` (R-5.1).
//
// The TTY check runs BEFORE RequireRoot, and `tui` is exempted from main's
// workspace assertion so that it can. R-5.3 is unconditional — "no TTY, exit 7,
// no filesystem change" — and if the root check ran first, the same
// non-interactive invocation would exit 7 inside a workspace and 3 outside one,
// which is two contracts wearing one name.
func cmdTUI(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error) {
	if !interactive() {
		// Named streams rather than "no TTY": the reader who hits this is
		// piping, redirecting, or running under CI, and "stdin and stdout must
		// both be a terminal" tells them which half of their command line to
		// change. The second line exists because R-5.3 says "explanatory", and
		// nothing is explained by a refusal that does not say what to run
		// instead.
		return nil, model.Errorf(model.ExitInteractive,
			"tui needs an interactive terminal: stdin and stdout must both be "+
				"one, and here at least one is a pipe, a file, or /dev/null.\n"+
				"  every other subcommand works without a terminal — "+
				"`company-os --help` lists them, `company-os validate` gives the "+
				"same findings as text, and --json gives them to a program.")
	}
	if err := ws.RequireRoot(); err != nil {
		return nil, err
	}
	// args.Root, not ws.Root: the previews must reproduce what the reader typed.
	// If they relied on the cwd or on $COMPANY_OS_WORKSPACE_ROOT, the previewed
	// command reproduces from the same place, and interpolating the resolved
	// absolute path instead would print a flag they did not use.
	return nil, tui.Run(screensFor(ws, args.Root), tui.Options{
		Input:   stdin,
		Output:  out,
		NoColor: noColorRequested(),
	})
}

// noColorRequested implements the NO_COLOR convention as published: the variable
// disables colour when it is PRESENT and NOT EMPTY. Testing only for presence
// would make `NO_COLOR=` — which shells produce readily — mean the opposite of
// what the user typed.
func noColorRequested() bool {
	v, ok := os.LookupEnv("NO_COLOR")
	return ok && v != ""
}

// ---------------------------------------------------------------- catalog

// readOnlyScreens is R-5.4's enumerated list, in its stated order.
//
// Six of the ten are a subcommand: they build the same *Args the parser would
// have built and hand it to runScreen, which dispatches through `commands` and
// formats through `renderers`. The other four — the overview and the three
// browsers — have no single-command twin, so they list what is on disk. Every
// one of the ten is READ-ONLY, which is a property of this list and not of the
// UI: `discover validate` is deliberately absent from the discovery browser
// because it rewrites `status: draft` to `status: validated` in the brief
// (internal/product/discover.go), and a browser that edits what you browse is
// exactly the defect shipping read-only screens first exists to avoid.
// root is the raw --root the reader typed, threaded through so every derived
// invocation reproduces from where they stood (see cmdTUI).
func readOnlyScreens(ws *workspace.Workspace, root string) []tui.Screen {
	components := componentCatalog(ws)
	return []tui.Screen{
		{
			Title: "workspace overview",
			Run:   func(string) (string, error) { return overviewText(ws), nil },
		},
		{
			Title:   "today (role view)",
			Prompt:  "role",
			Choices: roleChoices(),
			Run: func(role string) (string, error) {
				return runScreen(ws, &Args{Root: root, Cmd: "today", Role: role})
			},
		},
		{
			Title: "validate results",
			Run: func(string) (string, error) {
				return runScreen(ws, &Args{Root: root, Cmd: "validate"})
			},
		},
		{
			Title: "component browser",
			Run:   func(string) (string, error) { return componentText(components), nil },
		},
		{
			Title: "PRD browser",
			Run:   func(string) (string, error) { return prdText(ws), nil },
		},
		{
			Title: "discovery browser",
			Run:   func(string) (string, error) { return discoveryText(ws), nil },
		},
		{
			Title:   "governance explain",
			Prompt:  "component",
			Choices: componentIDList(components),
			Run: func(cid string) (string, error) {
				return runScreen(ws, &Args{Root: root,
					Cmd: "governance", Action: "explain", ComponentArg: cid})
			},
		},
		{
			Title: "skills list",
			Run: func(string) (string, error) {
				return runScreen(ws, &Args{Root: root, Cmd: "skills", Action: "list"})
			},
		},
		{
			Title: "ids list",
			Run: func(string) (string, error) {
				return runScreen(ws, &Args{Root: root, Cmd: "ids", Action: "list"})
			},
		},
		{
			Title: "workspace status",
			Run: func(string) (string, error) {
				return runScreen(ws, &Args{Root: root, Cmd: "workspace", Action: "status"})
			},
		},
	}
}

// screensFor is the whole catalog the UI is driven with: R-5.4's ten read-only
// screens, then R-5.5's two mutating forms.
//
// The order is the requirement's and is load-bearing for more than tidiness —
// the read-only screens are what a reader lands on, and the two that write are
// below them, marked, and reachable only through a preview and a confirmation.
func screensFor(ws *workspace.Workspace, root string) []tui.Screen {
	return append(readOnlyScreens(ws, root), mutatingScreens(ws, root)...)
}

// runScreen executes one subcommand exactly as run() does and returns its text.
//
// The order matters and is run()'s: the command writes its own prose to the
// buffer first, the renderer appends the records after. A command may return
// records AND an error — `workspace sync` and `prd complete` both do — so the
// buffer is returned either way and the caller shows both.
func runScreen(ws *workspace.Workspace, args *Args) (string, error) {
	cmd, ok := commands[args.Cmd]
	if !ok {
		return "", model.Errorf(model.ExitUsage,
			"no handler registered for %q", args.Cmd)
	}
	var buf bytes.Buffer
	results, err := cmd(ws, args, &buf)
	if r, ok := renderers[args.Cmd]; ok && len(results) > 0 {
		if rerr := r(&buf, results); rerr != nil && err == nil {
			err = rerr
		}
	}
	header := screenCommand(args)
	if header == "" {
		return buf.String(), err
	}
	return "$ " + header + "\n\n" + buf.String(), err
}

// screenCommand renders the invocation an *Args is equivalent to, DERIVED from
// the same value that gets executed rather than written out per screen (R-5.6,
// R-5.7).
//
// It is the ONLY producer of a previewed command line in this binary. A
// read-only screen shows it as a header; a form shows it as the thing it is
// about to run, and then runs THAT *Args. There is no second spelling of a
// command anywhere, so there is nothing for the preview to drift away from.
//
// The guarantee is stated as a law rather than as a habit, and tuiform_test.go
// asserts it over every command in the spec table:
//
//	parse(shellSplit(screenCommand(a))) == a   for every a the parser can produce
//
// Three things follow from having to make that law TOTAL, and each is the reason
// for a piece of code below that would otherwise look like polish:
//
//   - Every token is shell-quoted. A discovery title is free text, so an
//     unquoted preview is not a command anyone can paste, and the law would fail
//     on the first title with a space in it.
//   - The pre-subcommand globals are rendered. --root is a value the parser can
//     hold; a preview that silently dropped it would print a command that runs
//     somewhere else.
//   - A positional beginning with `-` moves behind an argparse `--` guard, which
//     forces the flags in front of it. That layout is used ONLY when a
//     positional needs it, so the ordinary case still reads the way a person
//     types it.
func screenCommand(a *Args) string {
	spec, ok := lookupCommand(a.Cmd)
	if !ok {
		return ""
	}
	parts := []string{"company-os"}
	// Global options belong to the top-level parser, so they precede the
	// subcommand — argparse does not accept them after it.
	if a.Root != "" {
		parts = append(parts, "--root", shellQuote(a.Root))
	}
	if a.JSON {
		parts = append(parts, "--json")
	}
	parts = append(parts, a.Cmd)

	var positionals, flags []string
	guard := false
	for _, p := range spec.pos {
		v := *p.dest(a)
		// An empty OPTIONAL positional is simply absent, which is what argparse
		// would have seen. An empty REQUIRED one is a different fact — the value
		// really is the empty string — and eliding it would print a command line
		// the parser rejects.
		if v == "" && p.optional {
			continue
		}
		if strings.HasPrefix(v, "-") {
			guard = true
		}
		positionals = append(positionals, shellQuote(v))
	}
	for _, f := range spec.flags {
		if f.boolean != nil {
			if *f.boolean(a) {
				flags = append(flags, "--"+f.name)
			}
			continue
		}
		// A flag left at its argparse default is elided: printing `--role
		// developer` would teach a flag the reader never has to type, and
		// re-parsing the shorter line yields the same value. A REQUIRED flag is
		// always printed, default or not, for the same reason as above — the
		// parser refuses the line without it.
		if v := *f.str(a); f.required || v != f.def {
			flags = append(flags, "--"+f.name, shellQuote(v))
		}
	}

	if guard {
		parts = append(parts, flags...)
		parts = append(parts, "--")
		parts = append(parts, positionals...)
	} else {
		parts = append(parts, positionals...)
		parts = append(parts, flags...)
	}
	return strings.Join(parts, " ")
}

// shellQuote makes one token safe to paste into a POSIX shell, and safe to split
// back out again.
//
// Single quotes are used rather than double, because inside single quotes the
// shell expands nothing at all — no $, no backtick, no backslash — so a
// discovery title containing `$(…)` previews as text and pastes as text. The
// only sequence that needs handling is a single quote itself, spelled the
// standard way: close, escape one, reopen.
func shellQuote(s string) string {
	if s != "" && strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(
			"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"+
				"0123456789_@%+=:,./-", r)
	}) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// roleChoices is `today --role`'s choice set, read out of the parser spec rather
// than repeated here — the two would otherwise drift the first time a role is
// added, and the TUI would offer a value argparse rejects.
func roleChoices() []string {
	spec, ok := lookupCommand("today")
	if !ok {
		return nil
	}
	for _, f := range spec.flags {
		if f.name == "role" {
			return append([]string(nil), f.choices...)
		}
	}
	return nil
}

// ------------------------------------------------------- the four listings

// componentRow is one row of the component browser.
type componentRow struct {
	id       string
	platform string
	reality  bool
}

// componentCatalog walks every platform's component directory once. Both the
// browser and the `governance explain` picker read it, so a workspace with a
// hundred components is scanned once per session rather than once per keystroke.
func componentCatalog(ws *workspace.Workspace) []componentRow {
	var rows []componentRow
	for _, pdir := range ws.AllPlatforms() {
		platform := filepath.Base(pdir)
		matches, err := filepath.Glob(filepath.Join(pdir, "components", "*.yaml"))
		if err != nil {
			continue
		}
		sort.Strings(matches)
		for _, f := range matches {
			id := strings.TrimSuffix(filepath.Base(f), ".yaml")
			reality := filepath.Join(pdir, "reality", "components", id+".md")
			_, statErr := os.Stat(reality)
			rows = append(rows, componentRow{
				id: id, platform: platform, reality: statErr == nil})
		}
	}
	return rows
}

func componentIDList(rows []componentRow) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.id)
	}
	return out
}

func componentText(rows []componentRow) string {
	if len(rows) == 0 {
		return "no component descriptors under platforms/*/components/."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d component(s)\n\n", len(rows))
	for _, r := range rows {
		state := "reality doc missing"
		if r.reality {
			state = "reality doc present"
		}
		fmt.Fprintf(&b, "  %s\n      platform: %s\n      %s\n", r.id, r.platform, state)
	}
	b.WriteString("\nopen `governance explain` for the rules that apply to one of these.\n")
	return b.String()
}

// prdText lists change records: the live ones first, then the archive. Both come
// from the directory layout invariant #4 describes — active PRDs under
// change-records/active/, completed ones under archive/prds/.
func prdText(ws *workspace.Workspace) string {
	var b strings.Builder
	total := 0
	for _, pdir := range ws.AllPlatforms() {
		platform := filepath.Base(pdir)
		for _, group := range []struct{ label, rel string }{
			{"active", filepath.Join("change-records", "active")},
			{"archived", filepath.Join("archive", "prds")},
		} {
			ids := subdirNames(filepath.Join(pdir, group.rel))
			for _, id := range ids {
				total++
				fmt.Fprintf(&b, "  [%s] %s/%s\n", group.label, platform, id)
				if s := frontmatterField(
					filepath.Join(pdir, group.rel, id, "prd.md"), "status"); s != "" {
					fmt.Fprintf(&b, "      status: %s\n", s)
				}
			}
		}
	}
	if total == 0 {
		return "no change records under platforms/*/change-records/ or */archive/prds/."
	}
	return fmt.Sprintf("%d change record(s)\n\n", total) + b.String()
}

// discoveryText lists team-private discovery briefs. It reads `status` and
// nothing else — running `discover validate` here would REWRITE the brief, which
// is why the browser is a listing rather than a command (R-5.4 ships read-only
// screens first precisely to make that class of mistake impossible).
func discoveryText(ws *workspace.Workspace) string {
	var b strings.Builder
	total := 0
	for _, tdir := range ws.AllTeams() {
		team := filepath.Base(tdir)
		dir := filepath.Join(tdir, "product", "discovery")
		for _, id := range subdirNames(dir) {
			total++
			fmt.Fprintf(&b, "  %s/%s\n", team, id)
			if s := frontmatterField(filepath.Join(dir, id, "brief.md"), "status"); s != "" {
				fmt.Fprintf(&b, "      status: %s\n", s)
			}
		}
	}
	if total == 0 {
		return "no discovery briefs under teams/*/product/discovery/."
	}
	return fmt.Sprintf("%d discovery brief(s)\n\n", total) + b.String()
}

// overviewText is the shape of the federation in one screen: which peer roots
// exist, what is in them, and whether this workspace is federated.
func overviewText(ws *workspace.Workspace) string {
	var b strings.Builder
	fmt.Fprintf(&b, "workspace %s\n\n", ws.Root)

	platforms := baseNames(ws.AllPlatforms())
	teams := baseNames(ws.AllTeams())
	fmt.Fprintf(&b, "  platforms  %d\n", len(platforms))
	for _, p := range platforms {
		fmt.Fprintf(&b, "      %s\n", p)
	}
	fmt.Fprintf(&b, "  teams      %d\n", len(teams))
	for _, t := range teams {
		fmt.Fprintf(&b, "      %s\n", t)
	}
	fmt.Fprintf(&b, "  components %d\n", len(componentCatalog(ws)))

	b.WriteString("\n  peer roots\n")
	for _, name := range workspace.CanonicalRoots {
		fmt.Fprintf(&b, "      %-19s %s\n", name, presence(filepath.Join(ws.Root, name)))
	}
	fmt.Fprintf(&b, "      %-19s %s\n", workspace.ManifestName,
		presence(filepath.Join(ws.Root, workspace.ManifestName)))
	fmt.Fprintf(&b, "      %-19s %s\n", workspace.LockName,
		presence(filepath.Join(ws.Root, workspace.LockName)))

	b.WriteString("\nopen `validate results` for the gate report.\n")
	return b.String()
}

func presence(path string) string {
	if _, err := os.Stat(path); err == nil {
		return "present"
	}
	return "absent"
}

func baseNames(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, filepath.Base(p))
	}
	return out
}

// subdirNames lists the immediate subdirectory names of dir, sorted. An absent
// dir yields nothing: a platform with no archive is normal, not an error.
func subdirNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if fi, err := os.Stat(filepath.Join(dir, e.Name())); err == nil && fi.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// frontmatterField reads one scalar out of a document's frontmatter, or "" when
// the file is missing, has no frontmatter, or does not carry the key. A browser
// that refused to list a malformed document would hide the very document its
// reader most needs to find.
func frontmatterField(path, key string) string {
	meta, _, err := graph.ReadFrontmatter(path)
	if err != nil || meta == nil {
		return ""
	}
	v := meta.Get(key)
	if v == nil {
		return ""
	}
	return yamlio.PyString(v)
}
