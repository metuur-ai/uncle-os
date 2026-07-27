package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Args is the parsed command line. It is shaped after the argparse Namespace
// that bin/company-os builds in main() (:2661-2781). Field names diverge from
// the Python attribute names only where a positional and a flag would collide:
// TitleArg is discover's `title` positional while Title is prd's --title, and
// ComponentArg is the reality/governance positional while Component is
// exception's --component.
type Args struct {
	Cmd  string
	Root string // raw --root value, "" when not supplied
	JSON bool   // new in Go: structured output

	// positionals
	Action       string // discover|prd|governance|deviation|exception|scratchpad|graph|ids|skills|workspace|reality
	Kind         string // add.kind, check.kind
	Name         string // add.name
	ComponentArg string // reality.component, governance.component
	TitleArg     string // discover.title
	ID           string // prd.id; discover mirrors TitleArg here, as :2694 does
	Rule         string // deviation.rule, exception.rule

	// flags
	Company       string
	Team          string
	Platform      string
	Components    string
	Title         string
	FromDiscovery string
	Force         bool
	Repair        bool
	Rationale     string
	Component     string
	Expires       string
	Reason        string
	Repo          string
	Role          string
	Prefix        string
	Frozen        bool
	Only          string
}

type posSpec struct {
	name     string
	optional bool // argparse nargs="?"
	choices  []string
	// help is the argument's `help=` string, verbatim from the oracle's parser.
	// Empty means the oracle passes no help= for it, and --help prints the
	// invocation bare — R-0.8 freezes the absence as much as the presence.
	help string
	dest func(*Args) *string
}

type flagSpec struct {
	name     string
	required bool
	def      string
	choices  []string
	help     string
	str      func(*Args) *string // nil for booleans
	boolean  func(*Args) *bool   // non-nil means argparse action="store_true"
}

type cmdSpec struct {
	name  string
	help  string
	pos   []posSpec
	flags []flagSpec
	// goOnly marks a subcommand the Python oracle does not have. It parses,
	// dispatches, and appears in `--help` like any other; it is only kept out of
	// commandNames(), which feeds two argparse-mirroring strings that the
	// differential harness compares byte-for-byte against a parser that has
	// never heard of it. See commandNames.
	goOnly bool
}

func strFlag(name string, dest func(*Args) *string) flagSpec {
	return flagSpec{name: name, str: dest}
}

func reqStrFlag(name string, dest func(*Args) *string) flagSpec {
	return flagSpec{name: name, required: true, str: dest}
}

func boolFlag(name string, dest func(*Args) *bool) flagSpec {
	return flagSpec{name: name, boolean: dest}
}

// commandSpecs mirrors the sub-parsers of bin/company-os:2661-2781, in the same
// order, with the same positional names, choice sets, required markers, and
// defaults.
var commandSpecs = []cmdSpec{
	{
		name: "init", help: "scaffold a new workspace (four peer roots)",
		flags: []flagSpec{
			{name: "company", help: "company name (skips the prompt)",
				str: func(a *Args) *string { return &a.Company }},
			{name: "team", help: "first team id (skips the prompt)",
				str: func(a *Args) *string { return &a.Team }},
			{name: "platform", help: "first platform id (skips the prompt)",
				str: func(a *Args) *string { return &a.Platform }},
		},
	},
	{
		name: "add", help: "grow a workspace: add a platform, team, or component",
		pos: []posSpec{
			{name: "kind", choices: []string{"platform", "team", "component"},
				dest: func(a *Args) *string { return &a.Kind }},
			{name: "name", help: "id of the new platform/team/component",
				dest: func(a *Args) *string { return &a.Name }},
		},
		flags: []flagSpec{
			{name: "platform", help: "target platform (required for add component)",
				str: func(a *Args) *string { return &a.Platform }},
			// `add team <id> --repair` fills in scaffolded files an existing team
			// is missing. Without it there is no path back: `add team` refuses
			// once the directory exists, so one deleted standards file could only
			// be recovered by hand.
			boolFlag("repair", func(a *Args) *bool { return &a.Repair }),
		},
	},
	{
		name: "reality", help: "scaffold a component reality doc",
		pos: []posSpec{
			{name: "action", choices: []string{"new"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "component", help: "component id",
				dest: func(a *Args) *string { return &a.ComponentArg }},
		},
		flags: []flagSpec{
			reqStrFlag("platform", func(a *Args) *string { return &a.Platform }),
		},
	},
	{
		name: "discover", help: "discovery brief workflow",
		pos: []posSpec{
			{name: "action", choices: []string{"new", "validate"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "title", optional: true,
				help: "title (new) or brief id (validate)",
				dest: func(a *Args) *string { return &a.TitleArg }},
		},
		flags: []flagSpec{
			reqStrFlag("team", func(a *Args) *string { return &a.Team }),
		},
	},
	{
		name: "prd", help: "PRD workflow",
		pos: []posSpec{
			{name: "action", choices: []string{"new", "validate", "complete"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "id", optional: true, dest: func(a *Args) *string { return &a.ID }},
		},
		flags: []flagSpec{
			strFlag("team", func(a *Args) *string { return &a.Team }),
			reqStrFlag("platform", func(a *Args) *string { return &a.Platform }),
			{name: "components", def: "",
				str: func(a *Args) *string { return &a.Components }},
			strFlag("title", func(a *Args) *string { return &a.Title }),
			strFlag("from-discovery", func(a *Args) *string { return &a.FromDiscovery }),
			boolFlag("force", func(a *Args) *bool { return &a.Force }),
		},
	},
	{
		name: "governance", help: "resolve / explain effective governance",
		pos: []posSpec{
			{name: "action", choices: []string{"resolve", "explain"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "component", optional: true,
				dest: func(a *Args) *string { return &a.ComponentArg }},
		},
		flags: []flagSpec{
			strFlag("team", func(a *Args) *string { return &a.Team }),
		},
	},
	{
		name: "check", help: "composable DoR / DoD",
		pos: []posSpec{
			{name: "kind", choices: []string{"ready", "done"},
				dest: func(a *Args) *string { return &a.Kind }},
		},
		flags: []flagSpec{
			reqStrFlag("team", func(a *Args) *string { return &a.Team }),
			reqStrFlag("components", func(a *Args) *string { return &a.Components }),
		},
	},
	{
		name: "validate", help: "workspace validation gates",
	},
	{
		name: "deviation", help: "declare a comply-or-explain deviation",
		pos: []posSpec{
			{name: "action", choices: []string{"declare"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "rule", dest: func(a *Args) *string { return &a.Rule }},
		},
		flags: []flagSpec{
			reqStrFlag("team", func(a *Args) *string { return &a.Team }),
			strFlag("rationale", func(a *Args) *string { return &a.Rationale }),
		},
	},
	{
		name: "exception", help: "request an exception to a mandatory rule",
		pos: []posSpec{
			{name: "action", choices: []string{"request"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "rule", dest: func(a *Args) *string { return &a.Rule }},
		},
		flags: []flagSpec{
			reqStrFlag("team", func(a *Args) *string { return &a.Team }),
			reqStrFlag("component", func(a *Args) *string { return &a.Component }),
			reqStrFlag("expires", func(a *Args) *string { return &a.Expires }),
			strFlag("reason", func(a *Args) *string { return &a.Reason }),
		},
	},
	{
		name: "scratchpad", help: "init the local-only scratchpad",
		pos: []posSpec{
			{name: "action", choices: []string{"init"},
				dest: func(a *Args) *string { return &a.Action }},
		},
		flags: []flagSpec{
			strFlag("repo", func(a *Args) *string { return &a.Repo }),
		},
	},
	{
		name: "today", help: "role-aware daily view",
		flags: []flagSpec{
			{
				name: "role", def: "developer",
				choices: []string{"developer", "team-lead", "product-owner",
					"architect", "vp-engineering", "director-of-product"},
				str: func(a *Args) *string { return &a.Role },
			},
		},
	},
	{
		name: "graph", help: "derive tags/graph metadata from frontmatter",
		pos: []posSpec{
			{name: "action", choices: []string{"build"},
				dest: func(a *Args) *string { return &a.Action }},
		},
	},
	{
		name: "ids", help: "list canonical IDs from the ontology registry",
		pos: []posSpec{
			{name: "action", choices: []string{"list"},
				dest: func(a *Args) *string { return &a.Action }},
		},
		flags: []flagSpec{
			{name: "team", help: "filter to IDs defined under this team",
				str: func(a *Args) *string { return &a.Team }},
			{name: "platform", help: "filter to IDs defined under this platform",
				str: func(a *Args) *string { return &a.Platform }},
			{name: "prefix", help: "filter to IDs starting with this prefix",
				str: func(a *Args) *string { return &a.Prefix }},
			{name: "role", help: "also show a plain-language glossary for this role",
				str: func(a *Args) *string { return &a.Role }},
		},
	},
	{
		name: "skills", help: "list merged agent skills across the four layers",
		pos: []posSpec{
			{name: "action", choices: []string{"list"},
				dest: func(a *Args) *string { return &a.Action }},
		},
	},
	{
		// R-5.1: the ONE launcher. It takes no flags and no positionals, so
		// there is no way to reach the UI except by typing its name — no bare
		// invocation, no other subcommand, no environment variable (R-5.2).
		// A bare `company-os` still prints help and exits 2, as it always has.
		name: "tui", help: "interactive terminal UI (read-only views)", goOnly: true,
	},
	{
		name: "workspace", help: "federated multi-repo governance sync/status (Option B)",
		pos: []posSpec{
			{name: "action", choices: []string{"sync", "status"},
				dest: func(a *Args) *string { return &a.Action }},
		},
		flags: []flagSpec{
			{name: "frozen",
				help:    "CI: materialize strictly from workspace.lock.yaml, no network",
				boolean: func(a *Args) *bool { return &a.Frozen }},
			{name: "only", help: "limit sync to a single repo by name",
				str: func(a *Args) *string { return &a.Only }},
		},
	},
}

func lookupCommand(name string) (cmdSpec, bool) {
	for _, s := range commandSpecs {
		if s.name == name {
			return s, true
		}
	}
	return cmdSpec{}, false
}

// sentinel results of parse that are not errors in the usual sense.
type sentinel string

func (s sentinel) Error() string { return string(s) }

const (
	errHelp    = sentinel("help requested")
	errVersion = sentinel("version requested")
)

// helpRequest is `-h`/`--help`, carrying the scope of the parser that saw it.
//
// argparse answers `company-os prd --help` from the `prd` SUB-parser, so its
// output is prd's own usage, positionals, and flags — not the top-level command
// list. Returning a bare sentinel lost that scope and made every sub-help print
// the root help, which drops six of prd's flags and answers a question the user
// did not ask. R-0.7a(i) waives argparse's *layout* for help text; it does not
// waive printing the wrong parser's.
//
// Is keeps `errors.Is(err, errHelp)` working for callers that only need to know
// that help was requested.
type helpRequest struct{ scope string }

func (h *helpRequest) Error() string { return string(errHelp) }

func (h *helpRequest) Is(target error) bool { return target == errHelp }

// usageError is an argparse-shaped argument error. argparse reports one of these
// as two pieces of stderr — a `usage:` line scoped to the parser that caught the
// mistake, then one line reading `<prog>: error: <message>` — and R-1.4a requires
// both. Scope is what distinguishes `company-os: error: ...` from
// `company-os skills: error: ...`; it also selects which usage line is printed,
// because printing the top-level usage for a sub-parser error loses the
// subcommand's own flags and choice sets.
//
// R-0.7a(i) waives argparse's COLUMNS-dependent *wrapping* of the usage line. It
// does not waive the error line, which argparse emits unwrapped and which is
// therefore reproducible byte-for-byte across CI runners.
type usageError struct {
	scope string // subcommand name; "" means the top-level parser caught it
	coded *model.Error
}

func (e *usageError) Error() string { return e.coded.Msg }

// Unwrap exposes the coded error so model.CodeOf still resolves ExitUsage (2),
// which is what every one of these paths exits with under Python.
func (e *usageError) Unwrap() error { return e.coded }

// prog renders argparse's `prog` for a scope: the value argparse interpolates
// before `: error:`.
func prog(scope string) string {
	if scope == "" {
		return "company-os"
	}
	return "company-os " + scope
}

// writeUsageError emits the two lines argparse writes for an argument error, in
// argparse's order. Both the parser's own usageError and the *model.UsageError
// that command code returns for a conditional requirement land here, so the two
// cannot drift into two different stderr shapes.
func writeUsageError(w io.Writer, scope, msg string) {
	fmt.Fprintln(w, usageLine(scope))
	fmt.Fprintf(w, "%s: error: %s\n", prog(scope), msg)
}

// rootErrf builds an error the top-level parser caught. argparse dispatches
// sub-parsers through parse_known_args, so unrecognized tokens bubble up here
// even when they appeared after a subcommand.
func rootErrf(format string, a ...any) error {
	return &usageError{coded: &model.Error{
		Code: model.ExitUsage, Msg: fmt.Sprintf(format, a...)}}
}

// subErrf builds an error the named sub-parser caught.
func subErrf(scope, format string, a ...any) error {
	return &usageError{scope: scope, coded: &model.Error{
		Code: model.ExitUsage, Msg: fmt.Sprintf(format, a...)}}
}

// parse splits argv into the global options, the subcommand, and that
// subcommand's flags and positionals. Like argparse, flags and positionals may
// be interleaved after the subcommand; unlike Go's flag package, a positional
// does not terminate flag parsing.
//
// The global --root/--json/--version options are pre-subcommand only, matching
// argparse, where an option on the parent parser is not accepted after the
// sub-parser takes over.
func parse(argv []string) (*Args, error) {
	a := &Args{}
	var extras []string

	i := 0
	for ; i < len(argv); i++ {
		tok := argv[i]
		if tok == "--" {
			i++
			break
		}
		if !strings.HasPrefix(tok, "-") || tok == "-" {
			break
		}
		name, value, hasValue := splitFlag(tok)
		switch name {
		case "h", "help":
			return nil, &helpRequest{}
		case "version":
			return nil, errVersion
		case "json":
			if hasValue {
				return nil, rootErrf("argument --json: ignored explicit argument %s",
					reprValue(value))
			}
			a.JSON = true
		case "root":
			if !hasValue {
				if i+1 >= len(argv) {
					return nil, rootErrf("argument --root: expected one argument")
				}
				i++
				value = argv[i]
			}
			a.Root = value
		default:
			// argparse leaves an unknown option for the root parser to report as
			// `unrecognized arguments`, and only after the required-subcommand
			// check: `company-os --bogus` reports the missing cmd, not --bogus.
			extras = append(extras, tok)
		}
	}

	if i >= len(argv) {
		return nil, rootErrf("the following arguments are required: cmd")
	}

	name := argv[i]
	spec, ok := lookupCommand(name)
	if !ok {
		return nil, rootErrf("argument cmd: invalid choice: %s (choose from %s)",
			reprValue(name), strings.Join(commandNames(), ", "))
	}
	a.Cmd = name

	subExtras, err := parseSubcommand(a, spec, argv[i+1:])
	if err != nil {
		return nil, err
	}
	// Every unrecognized token — an unknown flag or a surplus positional, before
	// or after the subcommand — is reported by the *root* parser, on one line,
	// in argv order, and only once the sub-parser's own checks have passed.
	if extras = append(extras, subExtras...); len(extras) > 0 {
		return nil, rootErrf("unrecognized arguments: %s", strings.Join(extras, " "))
	}

	normalizeArgs(a)
	return a, nil
}

// normalizeArgs applies the post-parse fixups that make an *Args well-formed,
// for every producer of one.
//
// There are two producers now: this parser, and the TUI's forms, which build an
// *Args directly from collected answers rather than from argv (R-5.12). Any rule
// applied here and not there would make the same command mean two things
// depending on how it was launched, so the rule lives in one function both call.
func normalizeArgs(a *Args) {
	// bin/company-os:2694 feeds the single `title` positional into both `title`
	// (discover new) and `id` (discover validate).
	if a.Cmd == "discover" {
		a.ID = a.TitleArg
	}
}

// parseSubcommand consumes one sub-parser's argv and returns the tokens argparse
// would have left in parse_known_args' extras list for the root parser to report.
//
// The check order is argparse's and is load-bearing: value-level checks fire
// while a token is being consumed, the required check fires after the whole argv
// is consumed, and unrecognized tokens are reported last, by the caller. That is
// why `reality bogus` reports the bad choice rather than the missing --platform,
// `add bad-kind` reports the bad choice rather than the missing `name`, and
// `discover new t1 t2` reports the missing --team rather than the surplus `t2`.
func parseSubcommand(a *Args, spec cmdSpec, argv []string) ([]string, error) {
	// argparse defaults are applied before parsing so an absent flag still has
	// its documented value.
	for _, f := range spec.flags {
		if f.str != nil && f.def != "" {
			*f.str(a) = f.def
		}
	}

	seen := map[string]bool{}
	// Indices into argv, not values, so surplus positionals and unknown flags can
	// be merged back into one argv-ordered `unrecognized arguments` list.
	var posIdx, extraAt []int
	endOfFlags := false

	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		switch {
		case endOfFlags || !strings.HasPrefix(tok, "-") || tok == "-":
			posIdx = append(posIdx, i)
			continue
		case tok == "--":
			endOfFlags = true
			continue
		}

		name, value, hasValue := splitFlag(tok)
		if name == "h" || name == "help" {
			return nil, &helpRequest{scope: spec.name}
		}
		f, ok := lookupFlag(spec, name)
		if !ok {
			// parse_known_args consumes an unknown option string on its own and
			// does not swallow the token after it: `--bogus=v` is reported
			// verbatim, and `--bogus v` leaves `v` to the positional matcher.
			extraAt = append(extraAt, i)
			continue
		}
		if f.boolean != nil {
			if hasValue {
				return nil, subErrf(spec.name, "argument --%s: ignored explicit argument %s",
					name, reprValue(value))
			}
			*f.boolean(a) = true
			seen[name] = true
			continue
		}
		if !hasValue {
			if i+1 >= len(argv) {
				return nil, subErrf(spec.name, "argument --%s: expected one argument", name)
			}
			i++
			value = argv[i]
		}
		if len(f.choices) > 0 && !contains(f.choices, value) {
			return nil, subErrf(spec.name, "argument --%s: invalid choice: %s (choose from %s)",
				name, reprValue(value), strings.Join(f.choices, ", "))
		}
		*f.str(a) = value
		seen[name] = true
	}

	matched := len(posIdx)
	if matched > len(spec.pos) {
		matched = len(spec.pos)
	}
	for slot := 0; slot < matched; slot++ {
		p, value := spec.pos[slot], argv[posIdx[slot]]
		if len(p.choices) > 0 && !contains(p.choices, value) {
			return nil, subErrf(spec.name, "argument %s: invalid choice: %s (choose from %s)",
				p.name, reprValue(value), strings.Join(p.choices, ", "))
		}
		*p.dest(a) = value
	}
	surplus := posIdx[matched:]

	// argparse emits ONE line listing EVERY missing argument, comma-separated, in
	// parser._actions order — declaration order, positionals bare and flags with
	// their dashes. Measured, every sub-parser declares all its required
	// positionals before all its required flags, so this ordering is exact.
	var missing []string
	for slot := matched; slot < len(spec.pos); slot++ {
		if !spec.pos[slot].optional {
			missing = append(missing, spec.pos[slot].name)
		}
	}
	for _, f := range spec.flags {
		if f.required && !seen[f.name] {
			missing = append(missing, "--"+f.name)
		}
	}
	if len(missing) > 0 {
		return nil, subErrf(spec.name, "the following arguments are required: %s",
			strings.Join(missing, ", "))
	}

	// Unknown flags and surplus positionals are one list, reported in argv order.
	extraAt = append(extraAt, surplus...)
	sort.Ints(extraAt)
	extras := make([]string, len(extraAt))
	for i, at := range extraAt {
		extras[i] = argv[at]
	}
	return extras, nil
}

// reprValue renders a value the way argparse does, via Python's repr: single
// quotes, switching to double quotes only when the value itself contains a
// single quote and no double quote.
func reprValue(v string) string {
	if strings.Contains(v, "'") && !strings.Contains(v, `"`) {
		return `"` + v + `"`
	}
	return "'" + v + "'"
}

// splitFlag strips the leading dashes and splits an optional `=value`.
func splitFlag(tok string) (name, value string, hasValue bool) {
	name = strings.TrimLeft(tok, "-")
	if idx := strings.Index(name, "="); idx >= 0 {
		return name[:idx], name[idx+1:], true
	}
	return name, "", false
}

func lookupFlag(spec cmdSpec, name string) (flagSpec, bool) {
	for _, f := range spec.flags {
		if f.name == name {
			return f, true
		}
	}
	return flagSpec{}, false
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// commandNames lists the subcommands in DECLARATION order. argparse renders a
// choice set in the order the choices were registered and never sorts it, so
// sorting here would reword a human-facing line (R-0.8).
// commandNames is the subcommand list argparse INTERPOLATES INTO A DIAGNOSTIC:
// the `{init,add,…}` group of the top-level usage line, and the `(choose from
// …)` tail of `argument cmd: invalid choice:`. The second of those is compared
// byte-for-byte by the differential harness — `usage/unknown-subcommand` waives
// the usage block above it and deliberately keeps the diagnostic line (R-1.4a)
// — so a name the oracle's parser does not have cannot appear here without
// putting DIVERGE off zero for a cosmetic gain.
//
// goOnly subcommands are therefore skipped. They are NOT hidden: `--help` walks
// commandSpecs directly and lists them, `company-os tui --help` answers from the
// sub-parser, and dispatch is unaffected. The only thing this costs is that a
// typo near `tui` is not offered `tui` as a suggestion — which is the smaller
// loss, and it reverses on its own at cutover, when R-9.3 deletes the oracle and
// there is no longer a second parser to match.
func commandNames() []string {
	names := make([]string, 0, len(commandSpecs))
	for _, s := range commandSpecs {
		if s.goOnly {
			continue
		}
		names = append(names, s.name)
	}
	return names
}

// usageLine renders the `usage:` line argparse prints above an argument error,
// scoped to the parser that caught it. It is deliberately never wrapped:
// argparse wraps to `$COLUMNS`, R-0.7a(i) waives that, and a golden captured
// from a wrapped line is not reproducible across CI runners.
func usageLine(scope string) string {
	if scope == "" {
		return "usage: company-os [-h] [--root ROOT] [--json] [--version] {" +
			strings.Join(commandNames(), ",") + "} ..."
	}
	spec, ok := lookupCommand(scope)
	if !ok {
		return usageLine("")
	}
	parts := []string{"usage: company-os", scope, "[-h]"}
	for _, f := range spec.flags {
		parts = append(parts, flagUsage(f))
	}
	for _, p := range spec.pos {
		parts = append(parts, posUsage(p))
	}
	return strings.Join(parts, " ")
}

func flagUsage(f flagSpec) string {
	var core string
	switch {
	case f.boolean != nil:
		core = "--" + f.name
	case len(f.choices) > 0:
		core = "--" + f.name + " {" + strings.Join(f.choices, ",") + "}"
	default:
		// argparse's default metavar is the dest, uppercased.
		core = "--" + f.name + " " + strings.ToUpper(strings.ReplaceAll(f.name, "-", "_"))
	}
	if f.required {
		return core
	}
	return "[" + core + "]"
}

// posMetavar is argparse's metavar for a positional: the choice set when it has
// one, otherwise the dest. It carries no optionality brackets — those belong to
// the usage line, not to the `positional arguments:` section.
func posMetavar(p posSpec) string {
	if len(p.choices) > 0 {
		return "{" + strings.Join(p.choices, ",") + "}"
	}
	return p.name
}

func posUsage(p posSpec) string {
	if p.optional {
		return "[" + posMetavar(p) + "]"
	}
	return posMetavar(p)
}
