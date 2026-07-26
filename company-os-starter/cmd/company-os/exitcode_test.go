package main

// The Unit 4 exit-code contract, one test per code (R-4.1..R-4.9, R-7.6).
//
// Every assertion here drives run() — the real dispatch path — rather than
// calling model.CodeOf on a hand-built error, because the thing that can rot is
// the WIRING: a producer reclassified in internal/ or a new branch in run() that
// forgets to consult CodeOf. A unit test on the error type would not notice
// either.
//
// Two codes have no producer in this build and say so explicitly rather than
// being skipped:
//
//   - 1 is `validate` reporting [FAIL], and internal/validate lands in Phase 3.
//   - 5 is `prd complete`'s done-gate refusal, and internal/product lands in
//     Phase 3.
//
// Both are exercised through a temporarily registered command, which proves the
// half that exists today — run()'s mapping — and will keep proving it when the
// real producers arrive.

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// wantExit runs argv through the real dispatch path and asserts the process exit
// code. stdout and stderr are captured so a failure reports what the binary
// actually said.
func wantExit(t *testing.T, want model.ExitCode, argv ...string) {
	t.Helper()
	var out, errb bytes.Buffer
	got := run(argv, &out, &errb)
	if got != int(want) {
		t.Errorf("%v exited %d, want %d\nstdout: %s\nstderr: %s",
			argv, got, want, out.String(), errb.String())
	}
}

// scratchWorkspace scaffolds a throwaway workspace and returns its root.
func scratchWorkspace(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "ws")
	runOK(t, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "plat")
	return root
}

// --------------------------------------------------------------- 0 (R-4.1)

func TestExit0Success(t *testing.T) {
	root := scratchWorkspace(t)
	for _, argv := range [][]string{
		{"--root", root, "ids", "list"},
		{"--root", root, "skills", "list"},
		{"--root", root, "today"},
		{"--root", root, "graph", "build"},
		{"--version"},
		{"--help"},
	} {
		wantExit(t, model.ExitOK, argv...)
	}
}

// --------------------------------------------------------------- 1 (R-4.2)

// TestExit1Validation covers the dispatch half of R-4.2: a command that returns
// any SevFail finding exits 1. The gate producers land in Phase 3; this is the
// seam they will hang off (main.go's model.HasFailure branch).
func TestExit1Validation(t *testing.T) {
	root := scratchWorkspace(t)
	restore := stubCommand(t, "validate", func(*workspace.Workspace, *Args, io.Writer) ([]model.GateResult, error) {
		return []model.GateResult{{
			Ordinal: 1, Slug: "ownership-reconciliation", Title: "ownership reconciliation",
			Findings: []model.Finding{{
				Severity: model.SevFail,
				Code:     model.CodeOwnershipDescriptorMissing,
				Message:  "no descriptor",
			}},
		}}, nil
	})
	defer restore()
	wantExit(t, model.ExitValidation, "--root", root, "validate")

	// The same command with no failing finding exits 0 — otherwise this test
	// would pass against a run() that returned 1 unconditionally.
	restore()
	restore = stubCommand(t, "validate", func(*workspace.Workspace, *Args, io.Writer) ([]model.GateResult, error) {
		return []model.GateResult{{Ordinal: 1, Slug: "s", Title: "t"}}, nil
	})
	defer restore()
	wantExit(t, model.ExitOK, "--root", root, "validate")
}

// --------------------------------------------------------------- 2 (R-4.3)

func TestExit2Usage(t *testing.T) {
	root := scratchWorkspace(t)
	for _, argv := range [][]string{
		{},                                     // bare invocation (R-1.4)
		{"frobnicate"},                         // unknown subcommand
		{"--root", root, "validate", "--nope"}, // unrecognized flag
		{"--root", root, "today", "--role", "nosuchrole"}, // invalid choice
		{"--root", root, "add", "component", "svc"},       // conditional --platform (:2021)
	} {
		wantExit(t, model.ExitUsage, argv...)
	}
}

// --------------------------------------------------------------- 3 (R-4.4)

func TestExit3Workspace(t *testing.T) {
	empty := t.TempDir()
	root := scratchWorkspace(t)
	for _, argv := range [][]string{
		{"--root", empty, "today"},                                         // require_root (:230)
		{"--root", empty, "ids", "list"},                                   // same site, another command
		{"--root", "/nonexistent/xyz", "ids", "list"},                      // --root at a path that is not there
		{"--root", root, "add", "component", "svc", "--platform", "ghost"}, // platform_dir (:238)
		{"--root", root, "reality", "new", "--platform", "ghost", "svc"},   // same
		{"--root", root, "workspace", "status"},                            // no manifest (:2547)
	} {
		wantExit(t, model.ExitWorkspace, argv...)
	}
}

// --------------------------------------------------------------- 4 (R-4.5)

// TestExit4Artifact covers all three shapes code 4 is defined over: a manifest
// schema violation (the 21 sites at bin/company-os:2102-2238), malformed YAML
// (R-0.7a(e) — Python has NO exit site for this and raises a traceback), and a
// well-formed document of the wrong shape (R-0.7a(j)).
func TestExit4Artifact(t *testing.T) {
	t.Run("manifest schema violation", func(t *testing.T) {
		root := scratchWorkspace(t)
		writeFile(t, filepath.Join(root, workspace.ManifestName),
			"version: 1\nrepos: []\n")
		wantExit(t, model.ExitArtifact, "--root", root, "workspace", "status")
	})

	t.Run("malformed manifest YAML", func(t *testing.T) {
		root := scratchWorkspace(t)
		writeFile(t, filepath.Join(root, workspace.ManifestName),
			"version: 1\nrepos: [unclosed\n")
		wantExit(t, model.ExitArtifact, "--root", root, "workspace", "status")
	})

	t.Run("malformed skill frontmatter", func(t *testing.T) {
		root := scratchWorkspace(t)
		writeFile(t, filepath.Join(root, "company-os", "skills", "bad.SKILL.md"),
			"---\nid: [unclosed\n---\nbody\n")
		wantExit(t, model.ExitArtifact, "--root", root, "skills", "list")
	})

	t.Run("skill frontmatter of the wrong shape", func(t *testing.T) {
		root := scratchWorkspace(t)
		writeFile(t, filepath.Join(root, "company-os", "skills", "shape.SKILL.md"),
			"---\n- a\n- list\n---\nbody\n")
		wantExit(t, model.ExitArtifact, "--root", root, "skills", "list")
	})

	t.Run("generated governance of the wrong shape", func(t *testing.T) {
		root := scratchWorkspace(t)
		writeFile(t,
			filepath.Join(root, "teams", "core", "generated", "effective-governance.yaml"),
			"generatedAt: 2026-01-01\ncomponents: null\n")
		wantExit(t, model.ExitArtifact, "--root", root, "today")
	})
}

// TestArtifactErrorsCarryCode4 is the type-level half: the two errors that reach
// dispatch without a *model.Error wrapper must still resolve to 4, or every
// caller that forwards them verbatim silently reports a validation failure.
func TestArtifactErrorsCarryCode4(t *testing.T) {
	_, err := yamlio.Load([]byte("a: [unclosed\n"))
	var se *yamlio.SyntaxError
	if !errors.As(err, &se) {
		t.Fatalf("malformed YAML returned %T, want *yamlio.SyntaxError", err)
	}
	if got := model.CodeOf(err); got != model.ExitArtifact {
		t.Errorf("CodeOf(yamlio.SyntaxError) = %d, want %d", got, model.ExitArtifact)
	}
	// Wrapping must not lose it — internal/skills forwards through fmt.Errorf.
	if got := model.CodeOf(wrap(err)); got != model.ExitArtifact {
		t.Errorf("CodeOf(wrapped SyntaxError) = %d, want %d", got, model.ExitArtifact)
	}

	_, err = frontmatter.Parse([]byte("---\n\xff\xfe\n---\n"))
	if !errors.Is(err, frontmatter.ErrInvalidUTF8) {
		t.Fatalf("invalid UTF-8 returned %v, want ErrInvalidUTF8", err)
	}
	if got := model.CodeOf(err); got != model.ExitArtifact {
		t.Errorf("CodeOf(ErrInvalidUTF8) = %d, want %d", got, model.ExitArtifact)
	}
}

// --------------------------------------------------------------- 5 (R-4.6)

// TestExit5Precondition covers the dispatch half of R-4.6. The producer is
// `prd complete`'s done-gate (bin/company-os:703), which lands with
// internal/product in Phase 3.
func TestExit5Precondition(t *testing.T) {
	root := scratchWorkspace(t)
	restore := stubCommand(t, "prd", func(*workspace.Workspace, *Args, io.Writer) ([]model.GateResult, error) {
		return nil, model.Errorf(model.ExitPrecondition,
			"reality/components/svc.md was last updated before this PRD was created")
	})
	defer restore()
	wantExit(t, model.ExitPrecondition, "--root", root, "prd", "complete",
		"--platform", "plat", "2026-thing")
}

// --------------------------------------------------------------- 6 (R-4.7)

func TestExit6ExternalTool(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH; the --frozen path cannot be reached")
	}
	root := scratchWorkspace(t)
	writeFile(t, filepath.Join(root, workspace.ManifestName),
		"version: 1\n"+
			"repos:\n"+
			"  - name: a\n"+
			"    url: file:///nonexistent\n"+
			"    localDirectory: platforms/a\n"+
			"    paths: [docs/]\n"+
			"    pin: {commit: '0000000000000000000000000000000000000000'}\n")
	// --frozen with no workspace.lock.yaml is bin/company-os:2564, held at 6
	// with its four siblings so an agent reads one code for "the frozen restore
	// failed" (.devlocal/go-port/exit-code-map.md § C).
	wantExit(t, model.ExitExternalTool, "--root", root, "workspace", "sync", "--frozen")
}

// --------------------------------------------------------------- 7 (R-4.8)

// TestExit7Interactive is bin/company-os:1961. `go test` runs with stdin
// detached, which is the condition itself — the flags are OPTIONAL to the parser
// and become required only when no terminal is attached, which is why this is 7
// and not 2 (.devlocal/go-port/exit-code-map.md § D).
func TestExit7Interactive(t *testing.T) {
	if isTerminal(os.Stdin) {
		t.Skip("stdin is a terminal; the wizard would prompt instead of refusing")
	}
	root := filepath.Join(t.TempDir(), "ws")
	wantExit(t, model.ExitInteractive, "--root", root, "init")
	wantExit(t, model.ExitInteractive, "--root", root, "init", "--company", "Acme")
}

// --------------------------------------------------------------- 8 (R-4.9)

func TestExit8Conflict(t *testing.T) {
	root := scratchWorkspace(t)
	// :1971 — the target is already a workspace root.
	wantExit(t, model.ExitConflict, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "plat")
	// :1797 — _write_new refuses to overwrite platforms/plat/platform.yaml.
	wantExit(t, model.ExitConflict, "--root", root, "add", "platform", "plat")
	// :2037 — `reality new` refuses to overwrite an existing reality doc.
	runOK(t, "--root", root, "add", "component", "svc", "--platform", "plat")
	runOK(t, "--root", root, "reality", "new", "--platform", "plat", "svc")
	wantExit(t, model.ExitConflict, "--root", root, "reality", "new", "--platform", "plat", "svc")
}

// ------------------------------------------------------------------ helpers

// stubCommand temporarily replaces one entry in the dispatch table and returns
// the restore func. Calling restore twice is safe; the second call is a no-op on
// an already-restored entry.
func stubCommand(t *testing.T, name string, fn Command) func() {
	t.Helper()
	prev, existed := commands[name]
	commands[name] = fn
	restored := false
	return func() {
		if restored {
			return
		}
		restored = true
		if existed {
			commands[name] = prev
		} else {
			delete(commands, name)
		}
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// wrap is the %w forwarding internal/skills does, isolated so the test states
// what it is exercising.
func wrap(err error) error { return errors.Join(err) }
