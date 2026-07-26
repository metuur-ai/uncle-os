package main

import (
	"fmt"
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
	dest     func(*Args) *string
}

type flagSpec struct {
	name     string
	required bool
	def      string
	choices  []string
	str      func(*Args) *string // nil for booleans
	boolean  func(*Args) *bool   // non-nil means argparse action="store_true"
}

type cmdSpec struct {
	name  string
	help  string
	pos   []posSpec
	flags []flagSpec
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
			strFlag("company", func(a *Args) *string { return &a.Company }),
			strFlag("team", func(a *Args) *string { return &a.Team }),
			strFlag("platform", func(a *Args) *string { return &a.Platform }),
		},
	},
	{
		name: "add", help: "grow a workspace: add a platform, team, or component",
		pos: []posSpec{
			{name: "kind", choices: []string{"platform", "team", "component"},
				dest: func(a *Args) *string { return &a.Kind }},
			{name: "name", dest: func(a *Args) *string { return &a.Name }},
		},
		flags: []flagSpec{
			strFlag("platform", func(a *Args) *string { return &a.Platform }),
		},
	},
	{
		name: "reality", help: "scaffold a component reality doc",
		pos: []posSpec{
			{name: "action", choices: []string{"new"},
				dest: func(a *Args) *string { return &a.Action }},
			{name: "component", dest: func(a *Args) *string { return &a.ComponentArg }},
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
			strFlag("team", func(a *Args) *string { return &a.Team }),
			strFlag("platform", func(a *Args) *string { return &a.Platform }),
			strFlag("prefix", func(a *Args) *string { return &a.Prefix }),
			strFlag("role", func(a *Args) *string { return &a.Role }),
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
		name: "workspace", help: "federated multi-repo governance sync/status (Option B)",
		pos: []posSpec{
			{name: "action", choices: []string{"sync", "status"},
				dest: func(a *Args) *string { return &a.Action }},
		},
		flags: []flagSpec{
			boolFlag("frozen", func(a *Args) *bool { return &a.Frozen }),
			strFlag("only", func(a *Args) *string { return &a.Only }),
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

func usageErrf(format string, a ...any) error {
	return model.Errorf(model.ExitUsage, format, a...)
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
			return nil, errHelp
		case "version":
			return nil, errVersion
		case "json":
			if hasValue {
				return nil, usageErrf("--json does not take a value")
			}
			a.JSON = true
		case "root":
			if !hasValue {
				if i+1 >= len(argv) {
					return nil, usageErrf("argument --root: expected one argument")
				}
				i++
				value = argv[i]
			}
			a.Root = value
		default:
			return nil, usageErrf("unrecognized argument: %s", tok)
		}
	}

	if i >= len(argv) {
		return nil, usageErrf("the following arguments are required: cmd")
	}

	name := argv[i]
	spec, ok := lookupCommand(name)
	if !ok {
		return nil, usageErrf("argument cmd: invalid choice: %q (choose from %s)",
			name, quotedList(commandNames()))
	}
	a.Cmd = name

	if err := parseSubcommand(a, spec, argv[i+1:]); err != nil {
		return nil, err
	}

	// bin/company-os:2694 feeds the single `title` positional into both `title`
	// (discover new) and `id` (discover validate).
	if a.Cmd == "discover" {
		a.ID = a.TitleArg
	}
	return a, nil
}

func parseSubcommand(a *Args, spec cmdSpec, argv []string) error {
	// argparse defaults are applied before parsing so an absent flag still has
	// its documented value.
	for _, f := range spec.flags {
		if f.str != nil && f.def != "" {
			*f.str(a) = f.def
		}
	}

	seen := map[string]bool{}
	var positionals []string
	endOfFlags := false

	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		switch {
		case endOfFlags || !strings.HasPrefix(tok, "-") || tok == "-":
			positionals = append(positionals, tok)
			continue
		case tok == "--":
			endOfFlags = true
			continue
		}

		name, value, hasValue := splitFlag(tok)
		if name == "h" || name == "help" {
			return errHelp
		}
		f, ok := lookupFlag(spec, name)
		if !ok {
			return usageErrf("%s: unrecognized argument: %s", spec.name, tok)
		}
		if f.boolean != nil {
			if hasValue {
				return usageErrf("%s: argument --%s: ignored explicit argument %q",
					spec.name, name, value)
			}
			*f.boolean(a) = true
			seen[name] = true
			continue
		}
		if !hasValue {
			if i+1 >= len(argv) {
				return usageErrf("%s: argument --%s: expected one argument", spec.name, name)
			}
			i++
			value = argv[i]
		}
		if len(f.choices) > 0 && !contains(f.choices, value) {
			return usageErrf("%s: argument --%s: invalid choice: %q (choose from %s)",
				spec.name, name, value, quotedList(f.choices))
		}
		*f.str(a) = value
		seen[name] = true
	}

	for _, f := range spec.flags {
		if f.required && !seen[f.name] {
			return usageErrf("%s: the following arguments are required: --%s",
				spec.name, f.name)
		}
	}

	required := 0
	for _, p := range spec.pos {
		if !p.optional {
			required++
		}
	}
	if len(positionals) < required {
		var missing []string
		for _, p := range spec.pos[len(positionals):] {
			if !p.optional {
				missing = append(missing, p.name)
			}
		}
		return usageErrf("%s: the following arguments are required: %s",
			spec.name, strings.Join(missing, ", "))
	}
	if len(positionals) > len(spec.pos) {
		return usageErrf("%s: unrecognized arguments: %s",
			spec.name, strings.Join(positionals[len(spec.pos):], " "))
	}
	for idx, value := range positionals {
		p := spec.pos[idx]
		if len(p.choices) > 0 && !contains(p.choices, value) {
			return usageErrf("%s: argument %s: invalid choice: %q (choose from %s)",
				spec.name, p.name, value, quotedList(p.choices))
		}
		*p.dest(a) = value
	}
	return nil
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

func commandNames() []string {
	names := make([]string, 0, len(commandSpecs))
	for _, s := range commandSpecs {
		names = append(names, s.name)
	}
	sort.Strings(names)
	return names
}

func quotedList(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}
