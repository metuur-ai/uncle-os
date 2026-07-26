package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// portedCommands are the subcommands whose implementation has landed, so the
// surface test must not invoke them. Delete an entry only when its command goes
// back to being a stub, which should never happen.
var portedCommands = map[string]bool{
	"init": true, "add": true, "reality": true, "scratchpad": true,
	"today": true, "ids": true, "skills": true,
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

func TestUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		argv []string
	}{
		{"bare invocation", nil},
		{"unknown subcommand", []string{"nope"}},
		{"bad positional choice", []string{"add", "widget", "x"}},
		{"missing required flag", []string{"check", "ready", "--team", "core"}},
		{"missing required positional", []string{"deviation", "declare", "--team", "core"}},
		{"unknown flag", []string{"validate", "--nope"}},
		{"bad flag choice", []string{"today", "--role", "wizard"}},
		{"flag without value", []string{"check", "ready", "--team"}},
		{"extra positional", []string{"skills", "list", "extra"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parse(tc.argv)
			if err == nil {
				t.Fatalf("parse(%q) succeeded, want a usage error", tc.argv)
			}
			if got := model.CodeOf(err); got != model.ExitUsage {
				t.Fatalf("exit code = %d, want %d", got, model.ExitUsage)
			}
		})
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
