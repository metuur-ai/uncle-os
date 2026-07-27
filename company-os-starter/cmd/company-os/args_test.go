package main

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// portedCommands are the subcommands whose implementation has landed, so the
// surface test must not invoke them. Delete an entry only when its command goes
// back to being a stub, which should never happen.
var portedCommands = map[string]bool{
	"init": true, "add": true, "reality": true, "scratchpad": true,
	"today": true, "ids": true, "skills": true, "workspace": true,
	"graph": true, "governance": true, "deviation": true, "exception": true,
	"discover": true, "prd": true, "check": true, "validate": true,
	// `tui` is listed for the same reason as the rest — it is implemented — but
	// it is the one entry whose exclusion is load-bearing rather than tidy:
	// dispatching it from here would consult the real terminal, and on a
	// developer's machine that means the surface test opens a full-screen UI.
	// Its own contract is in tui_test.go, with the TTY probe stubbed.
	"tui": true,
}

// TestEverySubcommandParses walks the whole surface ported from
// bin/company-os:2661-2781: one valid invocation per subcommand, asserting the
// positionals and flags land in the fields the commands will read.
func TestEverySubcommandParses(t *testing.T) {
	cases := []struct {
		argv  []string
		check func(*testing.T, *Args)
	}{
		{
			argv: []string{"init", "--company", "Acme", "--team", "core", "--platform", "plat"},
			check: func(t *testing.T, a *Args) {
				want(t, "company", a.Company, "Acme")
				want(t, "team", a.Team, "core")
				want(t, "platform", a.Platform, "plat")
			},
		},
		{
			argv: []string{"add", "component", "billing-api", "--platform", "plat"},
			check: func(t *testing.T, a *Args) {
				want(t, "kind", a.Kind, "component")
				want(t, "name", a.Name, "billing-api")
				want(t, "platform", a.Platform, "plat")
			},
		},
		{
			argv: []string{"reality", "new", "billing-api", "--platform", "plat"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "new")
				want(t, "component", a.ComponentArg, "billing-api")
				want(t, "platform", a.Platform, "plat")
			},
		},
		{
			argv: []string{"discover", "new", "--team", "core", "Faster onboarding"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "new")
				want(t, "team", a.Team, "core")
				want(t, "title", a.TitleArg, "Faster onboarding")
				// :2694 mirrors the single positional into both title and id.
				want(t, "id", a.ID, "Faster onboarding")
			},
		},
		{
			argv: []string{"prd", "new", "--team", "core", "--platform", "plat",
				"--components", "a,b", "--from-discovery", "2026-x", "--force"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "new")
				want(t, "components", a.Components, "a,b")
				want(t, "from-discovery", a.FromDiscovery, "2026-x")
				if !a.Force {
					t.Error("--force did not set Force")
				}
			},
		},
		{
			argv: []string{"governance", "explain", "billing-api", "--team", "core"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "explain")
				want(t, "component", a.ComponentArg, "billing-api")
			},
		},
		{
			argv: []string{"check", "ready", "--team", "core", "--components", "a"},
			check: func(t *testing.T, a *Args) {
				want(t, "kind", a.Kind, "ready")
				want(t, "components", a.Components, "a")
			},
		},
		{
			argv:  []string{"validate"},
			check: func(t *testing.T, a *Args) { want(t, "cmd", a.Cmd, "validate") },
		},
		{
			argv: []string{"deviation", "declare", "req://x", "--team", "core",
				"--rationale", "why"},
			check: func(t *testing.T, a *Args) {
				want(t, "rule", a.Rule, "req://x")
				want(t, "rationale", a.Rationale, "why")
			},
		},
		{
			argv: []string{"exception", "request", "req://x", "--team", "core",
				"--component", "billing-api", "--expires", "2035-01-01", "--reason", "r"},
			check: func(t *testing.T, a *Args) {
				want(t, "rule", a.Rule, "req://x")
				want(t, "component", a.Component, "billing-api")
				want(t, "expires", a.Expires, "2035-01-01")
			},
		},
		{
			argv: []string{"scratchpad", "init", "--repo", "/tmp/repo"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "init")
				want(t, "repo", a.Repo, "/tmp/repo")
			},
		},
		{
			argv:  []string{"today", "--role", "architect"},
			check: func(t *testing.T, a *Args) { want(t, "role", a.Role, "architect") },
		},
		{
			argv:  []string{"graph", "build"},
			check: func(t *testing.T, a *Args) { want(t, "action", a.Action, "build") },
		},
		{
			argv: []string{"ids", "list", "--team", "core", "--platform", "plat",
				"--prefix", "component://", "--role", "developer"},
			check: func(t *testing.T, a *Args) {
				want(t, "prefix", a.Prefix, "component://")
				want(t, "role", a.Role, "developer")
			},
		},
		{
			argv:  []string{"skills", "list"},
			check: func(t *testing.T, a *Args) { want(t, "action", a.Action, "list") },
		},
		{
			// R-5.1/R-5.2: `tui` is a bare subcommand with no flags and no
			// positionals, so the ONLY way to reach the interactive UI is to
			// type its name. Any flag added here would be a second launcher.
			argv: []string{"tui"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "")
				want(t, "role", a.Role, "")
			},
		},
		{
			argv: []string{"workspace", "sync", "--frozen", "--only", "platform-x"},
			check: func(t *testing.T, a *Args) {
				want(t, "action", a.Action, "sync")
				want(t, "only", a.Only, "platform-x")
				if !a.Frozen {
					t.Error("--frozen did not set Frozen")
				}
			},
		},
	}

	if len(cases) != len(commandSpecs) {
		t.Fatalf("covered %d subcommands, the surface has %d", len(cases), len(commandSpecs))
	}

	for _, tc := range cases {
		t.Run(tc.argv[0], func(t *testing.T) {
			a, err := parse(tc.argv)
			if err != nil {
				t.Fatalf("parse(%q): %v", tc.argv, err)
			}
			if a.Cmd != tc.argv[0] {
				t.Fatalf("cmd = %q, want %q", a.Cmd, tc.argv[0])
			}
			tc.check(t, a)

			// Surface only: every not-yet-ported command must reach its stub
			// rather than a parse error. A command whose phase has landed is
			// exercised by its own package's tests and by the differential
			// harness, not from here — calling it would write to disk.
			if _, ported := portedCommands[a.Cmd]; ported {
				return
			}
			if _, err := commands[a.Cmd](nil, a, io.Discard); err == nil ||
				!strings.Contains(err.Error(), "not implemented") {
				t.Fatalf("dispatch returned %v, want a not-implemented error", err)
			}
		})
	}
}

// TestDefaultsMatchArgparse pins the two flags with non-empty argparse defaults.
func TestDefaultsMatchArgparse(t *testing.T) {
	a, err := parse([]string{"prd", "validate", "--platform", "p"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want(t, "components", a.Components, "")

	a, err = parse([]string{"today"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want(t, "role", a.Role, "developer")
}

// TestRootResolutionOrder covers R-1.2. workspace.Resolve owns the env and cwd
// fallbacks; parse only has to carry the flag through unset when absent.
func TestRootResolutionOrder(t *testing.T) {
	a, err := parse([]string{"--root", "/ws", "validate"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want(t, "root", a.Root, "/ws")

	a, err = parse([]string{"validate"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want(t, "root", a.Root, "")
}

// topUsage is the usage line the top-level parser prints above its errors. It
// carries --json and --version, which are Go-only, on top of argparse's line.
const topUsage = "usage: company-os [-h] [--root ROOT] [--json] [--version] " +
	"{init,add,reality,discover,prd,governance,check,validate,deviation," +
	"exception,scratchpad,today,graph,ids,skills,workspace} ..."

// TestArgumentErrorDiagnostics covers R-1.4a. Every `wantErr` below is a verbatim
// copy of what the Python CLI emits for the same argv — captured from
// bin/company-os with the vendored PyYAML and transcribed into
// .devlocal/go-port/argparse-truth.md, which records the same shapes under both
// COLUMNS=200 and COLUMNS=80 to show that this line does not wrap.
//
// R-0.7a(i) waives argparse's COLUMNS-dependent wrapping of the `usage:` line,
// so `wantUsage` is asserted against the Go renderer's own unwrapped line rather
// than against argparse's bytes. The error line is asserted exactly.
//
// One `usage:` line has since DIVERGED from Python on purpose: `add` carries
// `[--repair]`, a Go-only flag with no Python counterpart (GPF-R-1.9a). The
// error lines below are still verbatim Python. A usage line growing a flag the
// reference never had is the expected consequence of adding surface, not drift —
// but it does mean these four strings are no longer transcriptions, and a future
// reader should not treat them as evidence of what Python printed.
func TestArgumentErrorDiagnostics(t *testing.T) {
	cases := []struct {
		name      string
		argv      []string
		wantUsage string
		wantErr   string
	}{
		// --- bare invocation ---
		{
			"bare invocation", nil, topUsage,
			"company-os: error: the following arguments are required: cmd",
		},
		{
			"global flag only, no subcommand", []string{"--root", "/tmp"}, topUsage,
			"company-os: error: the following arguments are required: cmd",
		},
		{
			// The missing subcommand outranks the unknown flag, as it does in
			// argparse, where the required-subparser check runs before extras.
			"unknown global flag, no subcommand", []string{"--bogus"}, topUsage,
			"company-os: error: the following arguments are required: cmd",
		},

		// --- unknown subcommand: choices in DECLARATION order, never sorted ---
		{
			"unknown subcommand", []string{"frobnicate"}, topUsage,
			"company-os: error: argument cmd: invalid choice: 'frobnicate' " +
				"(choose from init, add, reality, discover, prd, governance, check, " +
				"validate, deviation, exception, scratchpad, today, graph, ids, " +
				"skills, workspace)",
		},
		{
			"unknown subcommand after --root", []string{"--root=/tmp", "nosuch"}, topUsage,
			"company-os: error: argument cmd: invalid choice: 'nosuch' " +
				"(choose from init, add, reality, discover, prd, governance, check, " +
				"validate, deviation, exception, scratchpad, today, graph, ids, " +
				"skills, workspace)",
		},

		// --- unknown flag: always reported by the ROOT parser ---
		{
			"unknown flag after a subcommand", []string{"validate", "--nope"}, topUsage,
			"company-os: error: unrecognized arguments: --nope",
		},
		{
			"unknown flag with an inline value", []string{"skills", "list", "--bogus=v"},
			topUsage, "company-os: error: unrecognized arguments: --bogus=v",
		},
		{
			"unknown global flag before a subcommand", []string{"--bogus", "validate"},
			topUsage, "company-os: error: unrecognized arguments: --bogus",
		},
		{
			// --json is Go-only and global-only; after a subcommand argparse and
			// Go agree that it is unrecognized.
			"--json after a subcommand", []string{"validate", "--json"}, topUsage,
			"company-os: error: unrecognized arguments: --json",
		},

		// --- invalid choice on a positional ---
		{
			"skills action", []string{"skills", "show"},
			"usage: company-os skills [-h] {list}",
			"company-os skills: error: argument action: invalid choice: 'show' " +
				"(choose from list)",
		},
		{
			"ids action", []string{"ids", "show"},
			"usage: company-os ids [-h] [--team TEAM] [--platform PLATFORM] " +
				"[--prefix PREFIX] [--role ROLE] {list}",
			"company-os ids: error: argument action: invalid choice: 'show' " +
				"(choose from list)",
		},
		{
			"graph action", []string{"graph", "rebuild"},
			"usage: company-os graph [-h] {build}",
			"company-os graph: error: argument action: invalid choice: 'rebuild' " +
				"(choose from build)",
		},
		{
			"scratchpad action", []string{"scratchpad", "reset"},
			"usage: company-os scratchpad [-h] [--repo REPO] {init}",
			"company-os scratchpad: error: argument action: invalid choice: 'reset' " +
				"(choose from init)",
		},
		{
			"workspace action", []string{"workspace", "pull"},
			"usage: company-os workspace [-h] [--frozen] [--only ONLY] {sync,status}",
			"company-os workspace: error: argument action: invalid choice: 'pull' " +
				"(choose from sync, status)",
		},
		{
			"reality action", []string{"reality", "delete", "x", "--platform", "communications"},
			"usage: company-os reality [-h] --platform PLATFORM {new} component",
			"company-os reality: error: argument action: invalid choice: 'delete' " +
				"(choose from new)",
		},
		{
			"discover action",
			[]string{"discover", "frobnicate", "--team", "customer-engagement", "x"},
			"usage: company-os discover [-h] --team TEAM {new,validate} [title]",
			"company-os discover: error: argument action: invalid choice: 'frobnicate' " +
				"(choose from new, validate)",
		},
		{
			"deviation action",
			[]string{"deviation", "revoke", "x", "--team", "customer-engagement"},
			"usage: company-os deviation [-h] --team TEAM [--rationale RATIONALE] " +
				"{declare} rule",
			"company-os deviation: error: argument action: invalid choice: 'revoke' " +
				"(choose from declare)",
		},
		{
			"add kind", []string{"add", "widget", "x"},
			"usage: company-os add [-h] [--platform PLATFORM] [--repair] " +
				"{platform,team,component} name",
			"company-os add: error: argument kind: invalid choice: 'widget' " +
				"(choose from platform, team, component)",
		},
		{
			"check kind",
			[]string{"check", "sideways", "--team", "customer-engagement", "--components", "x"},
			"usage: company-os check [-h] --team TEAM --components COMPONENTS {ready,done}",
			"company-os check: error: argument kind: invalid choice: 'sideways' " +
				"(choose from ready, done)",
		},

		// --- the check-order proofs: a bad choice fires during consumption ---
		{
			"bad choice outranks the missing required flag", []string{"reality", "bogus"},
			"usage: company-os reality [-h] --platform PLATFORM {new} component",
			"company-os reality: error: argument action: invalid choice: 'bogus' " +
				"(choose from new)",
		},
		{
			"bad choice outranks the missing required positional", []string{"add", "bad-kind"},
			"usage: company-os add [-h] [--platform PLATFORM] [--repair] " +
				"{platform,team,component} name",
			"company-os add: error: argument kind: invalid choice: 'bad-kind' " +
				"(choose from platform, team, component)",
		},
		{
			"bad choice outranks the surplus positional", []string{"skills", "bogus", "extra"},
			"usage: company-os skills [-h] {list}",
			"company-os skills: error: argument action: invalid choice: 'bogus' " +
				"(choose from list)",
		},

		// --- invalid choice on a flag ---
		{
			"today --role", []string{"today", "--role", "wizard"}, todayUsage,
			"company-os today: error: argument --role: invalid choice: 'wizard' " +
				"(choose from developer, team-lead, product-owner, architect, " +
				"vp-engineering, director-of-product)",
		},
		{
			"today --role, inline form", []string{"today", "--role=bogus"}, todayUsage,
			"company-os today: error: argument --role: invalid choice: 'bogus' " +
				"(choose from developer, team-lead, product-owner, architect, " +
				"vp-engineering, director-of-product)",
		},
		{
			"flag choice outranks the surplus positional",
			[]string{"today", "--role", "wizard", "extra"}, todayUsage,
			"company-os today: error: argument --role: invalid choice: 'wizard' " +
				"(choose from developer, team-lead, product-owner, architect, " +
				"vp-engineering, director-of-product)",
		},

		// --- missing required flags: ONE line, ALL of them, declaration order ---
		{
			"missing --team", []string{"discover", "new", "Title"},
			"usage: company-os discover [-h] --team TEAM {new,validate} [title]",
			"company-os discover: error: the following arguments are required: --team",
		},
		{
			"missing --platform",
			[]string{"prd", "new", "--team", "customer-engagement", "--title", "T"},
			prdUsage,
			"company-os prd: error: the following arguments are required: --platform",
		},
		{
			"missing --components", []string{"check", "ready", "--team", "customer-engagement"},
			"usage: company-os check [-h] --team TEAM --components COMPONENTS {ready,done}",
			"company-os check: error: the following arguments are required: --components",
		},
		{
			"missing --expires",
			[]string{"exception", "request", "x", "--team", "ce", "--component", "c"},
			exceptionUsage,
			"company-os exception: error: the following arguments are required: --expires",
		},
		{
			"missing --component",
			[]string{"exception", "request", "x", "--team", "ce", "--expires", "2035-01-01"},
			exceptionUsage,
			"company-os exception: error: the following arguments are required: --component",
		},
		{
			"two missing flags on one line",
			[]string{"exception", "request", "some-rule", "--team", "t"}, exceptionUsage,
			"company-os exception: error: the following arguments are required: " +
				"--component, --expires",
		},
		{
			"two missing flags on one line, check", []string{"check", "ready"},
			"usage: company-os check [-h] --team TEAM --components COMPONENTS {ready,done}",
			"company-os check: error: the following arguments are required: --team, --components",
		},

		// --- missing required positionals ---
		{
			"missing add name", []string{"add", "platform"},
			"usage: company-os add [-h] [--platform PLATFORM] [--repair] " +
				"{platform,team,component} name",
			"company-os add: error: the following arguments are required: name",
		},
		{
			"missing deviation rule", []string{"deviation", "declare", "--team", "t"},
			"usage: company-os deviation [-h] --team TEAM [--rationale RATIONALE] " +
				"{declare} rule",
			"company-os deviation: error: the following arguments are required: rule",
		},
		{
			"missing skills action", []string{"skills"},
			"usage: company-os skills [-h] {list}",
			"company-os skills: error: the following arguments are required: action",
		},
		{
			"missing governance action", []string{"governance"},
			"usage: company-os governance [-h] [--team TEAM] {resolve,explain} [component]",
			"company-os governance: error: the following arguments are required: action",
		},
		{
			"missing reality component", []string{"reality", "new", "--platform", "p"},
			"usage: company-os reality [-h] --platform PLATFORM {new} component",
			"company-os reality: error: the following arguments are required: component",
		},
		{
			"two positionals missing", []string{"add"},
			"usage: company-os add [-h] [--platform PLATFORM] [--repair] " +
				"{platform,team,component} name",
			"company-os add: error: the following arguments are required: kind, name",
		},

		// --- positionals AND flags missing on the same line, in _actions order ---
		{
			"positional then flags", []string{"check"},
			"usage: company-os check [-h] --team TEAM --components COMPONENTS {ready,done}",
			"company-os check: error: the following arguments are required: " +
				"kind, --team, --components",
		},
		{
			"positional then flag", []string{"discover"},
			"usage: company-os discover [-h] --team TEAM {new,validate} [title]",
			"company-os discover: error: the following arguments are required: action, --team",
		},
		{
			"two positionals then a flag", []string{"deviation"},
			"usage: company-os deviation [-h] --team TEAM [--rationale RATIONALE] " +
				"{declare} rule",
			"company-os deviation: error: the following arguments are required: " +
				"action, rule, --team",
		},

		// --- too many positionals: reported by the ROOT parser ---
		{
			"surplus positional after skills", []string{"skills", "list", "extra"}, topUsage,
			"company-os: error: unrecognized arguments: extra",
		},
		{
			"surplus positional after validate", []string{"validate", "extra"}, topUsage,
			"company-os: error: unrecognized arguments: extra",
		},
		{
			"surplus positional after init", []string{"init", "extra"}, topUsage,
			"company-os: error: unrecognized arguments: extra",
		},
		{
			"surplus positional loses to the required check",
			[]string{"discover", "new", "t1", "t2"},
			"usage: company-os discover [-h] --team TEAM {new,validate} [title]",
			"company-os discover: error: the following arguments are required: --team",
		},
		{
			"surplus positional loses to the required check, prd",
			[]string{"prd", "new", "id1", "extra"}, prdUsage,
			"company-os prd: error: the following arguments are required: --platform",
		},

		// --- a flag missing its value ---
		{
			"global --root", []string{"--root"}, topUsage,
			"company-os: error: argument --root: expected one argument",
		},
		{
			"sub-parser --role", []string{"today", "--role"}, todayUsage,
			"company-os today: error: argument --role: expected one argument",
		},
		{
			"sub-parser --company", []string{"init", "--company"},
			"usage: company-os init [-h] [--company COMPANY] [--team TEAM] " +
				"[--platform PLATFORM]",
			"company-os init: error: argument --company: expected one argument",
		},

		// --- a value given to a store_true flag ---
		{
			"--force=v", []string{"prd", "new", "--platform", "p", "--force=v"}, prdUsage,
			"company-os prd: error: argument --force: ignored explicit argument 'v'",
		},
		{
			"--frozen=v", []string{"workspace", "sync", "--frozen=v"},
			"usage: company-os workspace [-h] [--frozen] [--only ONLY] {sync,status}",
			"company-os workspace: error: argument --frozen: ignored explicit argument 'v'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parse(tc.argv); model.CodeOf(err) != model.ExitUsage {
				t.Fatalf("parse(%q) exit code = %d, want %d",
					tc.argv, model.CodeOf(err), model.ExitUsage)
			}

			var stdout, stderr bytes.Buffer
			if code := run(tc.argv, &stdout, &stderr); code != 2 {
				t.Fatalf("run(%q) = %d, want 2", tc.argv, code)
			}
			if stdout.Len() != 0 {
				t.Errorf("run(%q) wrote to stdout:\n%s", tc.argv, stdout.String())
			}

			lines := strings.Split(strings.TrimRight(stderr.String(), "\n"), "\n")
			if len(lines) != 2 {
				t.Fatalf("stderr is %d lines, want exactly 2 (usage + error):\n%s",
					len(lines), stderr.String())
			}
			if lines[0] != tc.wantUsage {
				t.Errorf("usage line\n got: %s\nwant: %s", lines[0], tc.wantUsage)
			}
			if lines[1] != tc.wantErr {
				t.Errorf("error line\n got: %s\nwant: %s", lines[1], tc.wantErr)
			}
		})
	}
}

// Usage lines long enough to be worth naming once.
const (
	todayUsage = "usage: company-os today [-h] [--role {developer,team-lead," +
		"product-owner,architect,vp-engineering,director-of-product}]"
	prdUsage = "usage: company-os prd [-h] [--team TEAM] --platform PLATFORM " +
		"[--components COMPONENTS] [--title TITLE] " +
		"[--from-discovery FROM_DISCOVERY] [--force] {new,validate,complete} [id]"
	exceptionUsage = "usage: company-os exception [-h] --team TEAM " +
		"--component COMPONENT --expires EXPIRES [--reason REASON] {request} rule"
)

// TestErrorLineIsGreppable pins the selector the differential harness keys on:
// the diagnostic is exactly one line, it is the LAST line of stderr, and it
// matches `^company-os( <sub>)?: error: `. The usage block above it is waived
// under R-0.7a(i); this line is not.
func TestErrorLineIsGreppable(t *testing.T) {
	selector := regexp.MustCompile(`^company-os(?: [a-z]+)?: error: .+$`)
	for _, argv := range [][]string{
		nil, {"frobnicate"}, {"skills", "show"}, {"today", "--role", "wizard"},
		{"check", "ready"}, {"validate", "--nope"}, {"add", "platform"},
	} {
		var stdout, stderr bytes.Buffer
		run(argv, &stdout, &stderr)
		lines := strings.Split(strings.TrimRight(stderr.String(), "\n"), "\n")
		var matched []string
		for _, l := range lines {
			if selector.MatchString(l) {
				matched = append(matched, l)
			}
		}
		if len(matched) != 1 {
			t.Fatalf("run(%q): %d lines match the selector, want exactly 1:\n%s",
				argv, len(matched), stderr.String())
		}
		if matched[0] != lines[len(lines)-1] {
			t.Errorf("run(%q): the selected line is not the last line of stderr", argv)
		}
	}
}

// TestConditionalRequirementsAreArgparseShaped covers the six requirements
// argparse CANNOT express — a positional declared `nargs="?"` because it is
// required for one action and meaningless for another, and a flag required only
// alongside another flag. They are checked in command code, below cmd/, and
// before *model.UsageError existed their diagnostic came out as a bare
// `error: …` line: no `company-os ` prefix, no sub-parser usage line, and
// therefore invisible to the selector TestErrorLineIsGreppable pins.
//
// R-1.4a requires the diagnostic AND a sub-parser-scoped usage line; R-0.7a(l)
// is what sanctions replacing the oracle's traceback on these paths at all.
func TestConditionalRequirementsAreArgparseShaped(t *testing.T) {
	root := scratchWorkspace(t)
	selector := regexp.MustCompile(`^company-os(?: [a-z]+)?: error: .+$`)
	for _, tc := range []struct {
		scope string
		argv  []string
	}{
		{"discover", []string{"discover", "new", "--team", "core"}},
		{"discover", []string{"discover", "validate", "--team", "core"}},
		{"governance", []string{"governance", "resolve"}},
		{"prd", []string{"prd", "new", "--platform", "plat",
			"--from-discovery", "2026-x"}},
		{"prd", []string{"prd", "validate", "--platform", "plat"}},
		{"prd", []string{"prd", "complete", "--platform", "plat"}},
	} {
		var stdout, stderr bytes.Buffer
		argv := append([]string{"--root", root}, tc.argv...)
		if code := run(argv, &stdout, &stderr); code != 2 {
			t.Errorf("run(%q) = %d, want 2\n%s", tc.argv, code, stderr.String())
			continue
		}
		if stdout.Len() != 0 {
			t.Errorf("run(%q) wrote to stdout: %q", tc.argv, stdout.String())
		}
		lines := strings.Split(strings.TrimRight(stderr.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("run(%q): stderr is %d lines, want usage + diagnostic:\n%s",
				tc.argv, len(lines), stderr.String())
			continue
		}
		if want := usageLine(tc.scope); lines[0] != want {
			t.Errorf("run(%q) usage line = %q, want %q", tc.argv, lines[0], want)
		}
		if !selector.MatchString(lines[1]) {
			t.Errorf("run(%q): %q does not match the greppable selector",
				tc.argv, lines[1])
		}
		if !strings.HasPrefix(lines[1], "company-os "+tc.scope+": error: ") {
			t.Errorf("run(%q): diagnostic is not scoped to %q: %q",
				tc.argv, tc.scope, lines[1])
		}
	}
}

// TestSubcommandHelpIsScoped pins that `company-os <sub> --help` answers from
// the SUB-parser. Printing the root help there drops every one of that
// subcommand's own flags, which R-0.7a(i) does not waive — (i) waives argparse's
// layout, not answering a different question.
func TestSubcommandHelpIsScoped(t *testing.T) {
	for _, tc := range []struct {
		argv []string
		want []string
	}{
		{[]string{"validate", "--help"}, []string{"usage: company-os validate"}},
		{[]string{"prd", "-h"}, []string{"usage: company-os prd",
			"--from-discovery FROM_DISCOVERY", "{new,validate,complete}"}},
		{[]string{"--help"}, []string{"usage: company-os [--root ROOT]"}},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.argv, &stdout, &stderr); code != 0 {
			t.Fatalf("run(%q) = %d, want 0", tc.argv, code)
		}
		for _, w := range tc.want {
			if !strings.Contains(stdout.String(), w) {
				t.Errorf("run(%q) stdout is missing %q:\n%s",
					tc.argv, w, stdout.String())
			}
		}
	}
	// The root command list belongs to the root parser only.
	var stdout, stderr bytes.Buffer
	run([]string{"validate", "--help"}, &stdout, &stderr)
	if strings.Contains(stdout.String(), "scaffold a new workspace") {
		t.Errorf("`validate --help` printed the top-level command list:\n%s",
			stdout.String())
	}
}

// TestBareInvocationExitsTwo covers R-1.4 end to end, including the usage text
// argparse prints alongside it.
func TestBareInvocationExitsTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: company-os") {
		t.Fatalf("stderr did not carry usage:\n%s", stderr.String())
	}
}

func TestHelpAndVersionExitZero(t *testing.T) {
	for _, argv := range [][]string{{"--help"}, {"-h"}, {"--version"}} {
		var stdout, stderr bytes.Buffer
		if code := run(argv, &stdout, &stderr); code != 0 {
			t.Fatalf("run(%q) = %d, want 0", argv, code)
		}
		if stdout.Len() == 0 {
			t.Fatalf("run(%q) wrote nothing to stdout", argv)
		}
	}
}

// TestFlagsAndPositionalsInterleave is the reason the parser is hand-rolled: Go's
// flag package stops at the first non-flag argument, argparse does not.
func TestFlagsAndPositionalsInterleave(t *testing.T) {
	a, err := parse([]string{"discover", "--team", "core", "validate", "2026-brief"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want(t, "action", a.Action, "validate")
	want(t, "id", a.ID, "2026-brief")
}

func TestEqualsFormFlags(t *testing.T) {
	a, err := parse([]string{"--root=/ws", "prd", "new", "--platform=p", "--title=T"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want(t, "root", a.Root, "/ws")
	want(t, "platform", a.Platform, "p")
	want(t, "title", a.Title, "T")
}

func TestHelpSentinelIsNotAUsageError(t *testing.T) {
	_, err := parse([]string{"--help"})
	if !errors.Is(err, errHelp) {
		t.Fatalf("err = %v, want errHelp", err)
	}
}

func want(t *testing.T, field, got, expected string) {
	t.Helper()
	if got != expected {
		t.Errorf("%s = %q, want %q", field, got, expected)
	}
}
