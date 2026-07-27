package main

// Unit 3 through the real dispatch path: `--json` on every subcommand.
//
// These drive run() rather than render.JSON directly, because what R-3.1..R-3.9
// actually promise is about the PROCESS — that the flag is accepted everywhere,
// that stdout carries one document and nothing else, that the code the payload
// reports is the code the process exits with, and that a failure still produces
// a document. A unit test on the encoder proves none of that.

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
)

// payload is the decoded document plus the process exit code and stderr, which
// are half of what the requirements are about.
type payload struct {
	doc      map[string]any
	exitCode int
	stderr   string
	stdout   string
}

// runJSON runs argv with --json, asserts stdout round-trips through
// encoding/json, and returns the decoded document.
func runJSON(t *testing.T, argv ...string) payload {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(append([]string{"--json"}, argv...), &out, &errb)
	p := payload{exitCode: code, stderr: errb.String(), stdout: out.String()}
	if err := json.Unmarshal(out.Bytes(), &p.doc); err != nil {
		t.Fatalf("%v: stdout is not one valid JSON document: %v\n--- stdout ---\n%s"+
			"\n--- stderr ---\n%s", argv, err, out.String(), errb.String())
	}
	// R-3.4 / R-3.5: every payload, no exceptions.
	if p.doc["schemaVersion"] != float64(render.SchemaVersion) {
		t.Errorf("%v: schemaVersion = %v, want %d",
			argv, p.doc["schemaVersion"], render.SchemaVersion)
	}
	build, ok := p.doc["build"].(map[string]any)
	if !ok {
		t.Fatalf("%v: payload carries no build object (R-3.5)", argv)
	}
	for _, k := range []string{"version", "commit", "goVersion", "platform"} {
		if s, _ := build[k].(string); s == "" {
			t.Errorf("%v: build.%s is empty (R-3.5)", argv, k)
		}
	}
	// R-3.8: the document says what the process is about to do.
	if p.doc["exitCode"] != float64(code) {
		t.Errorf("%v: payload exitCode = %v, process exited %d",
			argv, p.doc["exitCode"], code)
	}
	return p
}

func (p payload) sections(t *testing.T) []map[string]any {
	t.Helper()
	raw, ok := p.doc["sections"].([]any)
	if !ok {
		t.Fatalf("payload has no sections array (R-3.4a)")
	}
	out := make([]map[string]any, 0, len(raw))
	for _, s := range raw {
		out = append(out, s.(map[string]any))
	}
	return out
}

func (p payload) guidance(t *testing.T) []string {
	t.Helper()
	raw, ok := p.doc["guidance"].([]any)
	if !ok {
		t.Fatalf("payload has no guidance array (R-3.6)")
	}
	out := make([]string, 0, len(raw))
	for _, g := range raw {
		out = append(out, g.(string))
	}
	return out
}

// jsonWorkspace is a throwaway workspace with one of everything the product
// lifecycle needs, so the mutating commands have something to mutate.
func jsonWorkspace(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "ws")
	runOK(t, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "plat")
	runOK(t, "--root", root, "add", "component", "--platform", "plat", "svc")
	return root
}

// TestJSONOnEverySubcommand is R-3.1 and R-3.2: the flag is accepted by all
// sixteen, and every one of them puts exactly one JSON document on stdout with no
// prose around it.
//
// The corpus is one representative invocation per subcommand — the surface the
// flag has to cover is the sixteen COMMANDS, and a second action of the same
// command exercises the same dispatch seam. `workspace` runs against a monorepo
// workspace and therefore fails; that is R-3.8's case and is asserted below.
func TestJSONOnEverySubcommand(t *testing.T) {
	// `tui` is in the corpus below and must never actually start a UI here.
	// Under `go test` both streams are pipes and the gate refuses on its own,
	// but "on its own" is not a guarantee worth a hung test binary.
	withoutTTY(t)
	root := jsonWorkspace(t)
	repo := t.TempDir()
	cases := [][]string{
		{"--root", root, "validate"},
		{"--root", root, "ids", "list"},
		{"--root", root, "today"},
		{"--root", root, "skills", "list"},
		{"--root", root, "graph", "build"},
		{"--root", root, "governance", "resolve", "--team", "core"},
		{"--root", root, "check", "ready", "--team", "core", "--components", "svc"},
		{"--root", root, "reality", "new", "--platform", "plat", "svc"},
		{"--root", root, "discover", "new", "--team", "core", "A brief"},
		{"--root", root, "add", "team", "second"},
		{"--root", root, "deviation", "declare", "rule://x", "--team", "core",
			"--rationale", "because"},
		{"--root", root, "exception", "request", "rule://y", "--team", "core",
			"--component", "svc", "--expires", "2035-01-01"},
		{"--root", root, "scratchpad", "init", "--repo", repo},
		{"--root", root, "workspace", "status"},
		// R-3.8 on the exit-7 path: a refusal still puts one document on
		// stdout, so a consumer never has to tell "empty stdout" from "crashed".
		{"--root", root, "tui"},
		{"--root", filepath.Join(t.TempDir(), "fresh"), "init",
			"--company", "B", "--team", "t", "--platform", "p"},
	}
	seen := map[string]bool{}
	for _, argv := range cases {
		p := runJSON(t, argv...)
		if p.doc["command"] == nil || p.doc["command"] == "" {
			t.Errorf("%v: payload names no command", argv)
		}
		seen[p.doc["command"].(string)] = true
		// R-3.2: nothing but the document. Any stray Fprintf would leave text
		// before or after it, which json.Unmarshal in runJSON already rejects —
		// this catches the subtler case of a second document appended.
		if strings.Count(p.stdout, "\"schemaVersion\"") != 1 {
			t.Errorf("%v: stdout carries more than one document:\n%s", argv, p.stdout)
		}
	}
	// `prd` is driven separately because it needs a validated brief; assert here
	// that the other fifteen were all reached.
	seen["prd"] = true
	for name := range commands {
		if !seen[name] {
			t.Errorf("no --json case covers subcommand %q", name)
		}
	}
}

// TestJSONPRDLifecycle covers the sixteenth command and, with it, R-3.6 on the
// three-step chain the guidance exists to keep unbroken.
func TestJSONPRDLifecycle(t *testing.T) {
	root := jsonWorkspace(t)

	brief := runJSON(t, "--root", root, "discover", "new", "--team", "core", "A brief")
	next := brief.guidance(t)
	if len(next) != 1 || !strings.HasPrefix(next[0], "company-os discover validate ") {
		t.Fatalf("discover new guidance = %v (R-3.6)", next)
	}
	// The guidance is a runnable command, which is the whole point of R-3.6:
	// replay it and the chain continues.
	argv := append([]string{"--root", root}, strings.Fields(next[0])[1:]...)
	validated := runJSON(t, argv...)
	if g := validated.guidance(t); len(g) != 1 ||
		!strings.HasPrefix(g[0], "company-os prd new ") {
		t.Fatalf("discover validate guidance = %v", g)
	}

	briefID := ""
	for _, s := range brief.sections(t) {
		for _, f := range s["findings"].([]any) {
			fields, _ := f.(map[string]any)["fields"].(map[string]any)
			if id, ok := fields["brief"].(string); ok {
				briefID = id
			}
		}
	}
	if briefID == "" {
		t.Fatal("discover new payload never names the brief it created (R-3.7)")
	}

	created := runJSON(t, "--root", root, "prd", "new", "--team", "core",
		"--platform", "plat", "--components", "svc", "--from-discovery", briefID)
	if g := created.guidance(t); len(g) != 1 ||
		!strings.HasPrefix(g[0], "company-os prd validate ") {
		t.Fatalf("prd new guidance = %v (R-3.6)", g)
	}
	// R-3.7: `prd new` is one of the commands named as producing no findings, and
	// its envelope has to say what it made rather than come back empty.
	if len(created.sections(t)) == 0 {
		t.Fatal("prd new emitted an empty document (R-3.7)")
	}
	if !strings.Contains(created.stdout, "\"prd.created\"") {
		t.Errorf("prd new payload does not name what it created:\n%s", created.stdout)
	}
}

// TestJSONEnvelopeForFindinglessCommands is R-3.7 for the four scaffolding
// commands. Each has to name the thing it made, in fields, not only in prose.
func TestJSONEnvelopeForFindinglessCommands(t *testing.T) {
	root := jsonWorkspace(t)
	cases := []struct {
		name  string
		argv  []string
		code  string
		field string
	}{
		{"init", []string{"--root", filepath.Join(t.TempDir(), "w"), "init",
			"--company", "C", "--team", "t", "--platform", "p"},
			model.CodeInitCreated, "root"},
		{"add", []string{"--root", root, "add", "platform", "other"},
			model.CodeAddCreated, "id"},
		{"reality", []string{"--root", root, "reality", "new", "--platform", "plat", "svc"},
			model.CodeRealityCreated, "path"},
		{"scratchpad", []string{"--root", root, "scratchpad", "init", "--repo", t.TempDir()},
			model.CodeScratchpadCreated, "path"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := runJSON(t, tc.argv...)
			found := ""
			for _, s := range p.sections(t) {
				for _, raw := range s["findings"].([]any) {
					f := raw.(map[string]any)
					if f["code"] != tc.code {
						continue
					}
					fields, _ := f["fields"].(map[string]any)
					found, _ = fields[tc.field].(string)
				}
			}
			if found == "" {
				t.Errorf("%s payload has no %s.%s (R-3.7):\n%s",
					tc.name, tc.code, tc.field, p.stdout)
			}
		})
	}
}

// TestJSONInventsNoGuidance is R-1.9 on the JSON side. Four mutating commands
// print no next step in the oracle and R-0.8 outranks R-1.8 for all four, so
// `--json` must publish nothing rather than "complete" the chain — a consumer
// replaying `.guidance` would otherwise run a command a human never sees offered.
func TestJSONInventsNoGuidance(t *testing.T) {
	root := jsonWorkspace(t)
	for _, argv := range [][]string{
		{"--root", root, "governance", "resolve", "--team", "core"},
		{"--root", root, "exception", "request", "rule://z", "--team", "core",
			"--component", "svc", "--expires", "2035-01-01"},
		{"--root", root, "scratchpad", "init", "--repo", t.TempDir()},
		{"--root", root, "graph", "build"},
	} {
		if g := runJSON(t, argv...).guidance(t); len(g) != 0 {
			t.Errorf("%v published guidance the oracle does not print: %v", argv, g)
		}
	}
}

// TestJSONCountsAreNumbers is R-2.3 through the wire: a count that arrives as
// "1" instead of 1 forces every consumer to re-parse it.
func TestJSONCountsAreNumbers(t *testing.T) {
	root := jsonWorkspace(t)
	p := runJSON(t, "--root", root, "validate")
	for _, s := range p.sections(t) {
		if s["slug"] != model.SlugWorkspace {
			continue
		}
		fields := s["findings"].([]any)[0].(map[string]any)["fields"].(map[string]any)
		if _, ok := fields["gates"].(float64); !ok {
			t.Errorf("banner gates field is %T, want a number", fields["gates"])
		}
		if _, ok := fields["complete"].(bool); !ok {
			t.Errorf("banner complete field is %T, want a bool", fields["complete"])
		}
		return
	}
	t.Fatal("validate payload has no banner section")
}

// TestJSONFailureStillEmitsJSON is R-3.8: a failing command writes a document on
// stdout and its diagnostic on stderr (R-3.9), and the two agree about the code.
func TestJSONFailureStillEmitsJSON(t *testing.T) {
	root := jsonWorkspace(t)
	cases := []struct {
		name string
		argv []string
		want model.ExitCode
	}{
		{"not-a-workspace", []string{"--root", t.TempDir(), "validate"}, model.ExitWorkspace},
		{"monorepo-has-no-manifest", []string{"--root", root, "workspace", "status"},
			model.ExitWorkspace},
		{"unknown-team", []string{"--root", root, "governance", "resolve", "--team", "ghost"},
			model.ExitWorkspace},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := runJSON(t, tc.argv...)
			if p.exitCode != int(tc.want) {
				t.Errorf("exited %d, want %d", p.exitCode, tc.want)
			}
			if s, _ := p.doc["error"].(string); s == "" {
				t.Error("failing payload carries no error (R-3.8)")
			}
			if !strings.Contains(p.stderr, "error: ") {
				t.Errorf("diagnostic did not reach stderr (R-3.9): %q", p.stderr)
			}
		})
	}
}

// TestJSONFlagLeavesTextOutputAlone is R-3.3. The five golden snapshots are the
// real gate; this one guards the seam the goldens cannot see — that adding the
// flag did not change what happens WITHOUT it — by running each command both
// ways and asserting the text side is still text.
func TestJSONFlagLeavesTextOutputAlone(t *testing.T) {
	root := jsonWorkspace(t)
	for _, argv := range [][]string{
		{"--root", root, "ids", "list"},
		{"--root", root, "today"},
		{"--root", root, "skills", "list"},
		{"--root", root, "graph", "build"},
		{"--root", root, "validate"},
	} {
		var out, errb bytes.Buffer
		run(argv, &out, &errb)
		if strings.HasPrefix(strings.TrimSpace(out.String()), "{") {
			t.Errorf("%v emitted JSON without --json:\n%s", argv, out.String())
		}
	}
}

// TestJSONIsNotInTheDifferentialCorpus used to live here. It read
// examples/differential.py as a FILE and asserted the string "--json" did not
// appear in it, because `--json` was a Go-only flag with no Python counterpart:
// an invocation carrying it could not have been compared against the oracle.
//
// Both halves of that premise are gone. The Python CLI was deleted by R-9.3, and
// the corpus now lives in internal/difftest as a golden characterization suite
// with no second implementation to agree with — so `--json` invocations are not
// merely permitted there, they are worth ADDING, which is the exact opposite of
// what this test enforced. It also read the file with a t.Skipf on error, so
// after the delete it would have skipped in silence rather than failing.
//
// Not replaced by an equivalent: there is nothing left to constrain.
