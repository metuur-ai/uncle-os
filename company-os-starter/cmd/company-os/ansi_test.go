package main

// R-3.10: no subcommand other than `tui` emits an ANSI escape sequence.
//
// This is the only invariant in the port with no natural alarm. bin/company-os
// emits zero escapes, every golden was captured against that, and normalize()
// does not strip them — so the day a styled string reaches a non-TUI path, the
// goldens go red somewhere far from the cause and the diff is invisible in a
// terminal, because the escape is what makes it invisible. Phase 7 puts Lipgloss
// in the same binary; an import that leaks one package sideways is a plausible
// accident, not a hypothetical.
//
// Two assertions, deliberately different in kind:
//
//   - TestNoANSIEscapesInSource is structural. It reads the tree with go/ast and
//     fails on an escape byte in any string literal, or on an import of a
//     styling library, anywhere outside internal/tui. It catches the defect at
//     the line that introduced it, before the code has to run.
//   - TestNoANSIEscapesAtRuntime is empirical. It builds the binary and sweeps
//     real invocations, because the structural check can only see escapes that
//     are spelled out — one assembled at runtime from an integer, or arriving
//     from a dependency, is only visible in the bytes.
//
// Test files are excluded from the structural walk, following
// architecture_test.go: the invariant constrains what the binary can print, and
// this file has to name the escape byte to test for it.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// esc is the byte that opens every ANSI sequence. Colour, cursor movement,
// screen clears and terminal queries all start here, so one byte is the whole
// check — matching CSI ("\x1b[") alone would miss OSC, and miss a bare escape.
const esc = 0x1b

// styleLibs are import paths that exist to write escape sequences. A non-TUI
// package importing one has already lost the guarantee even if it has not yet
// called anything: the intent is unambiguous, and catching it at the import is
// what makes the rule enforceable before the styled string is written.
//
// Matched as substrings so a version suffix or a subpackage still trips.
var styleLibs = []string{
	"lipgloss",
	"termenv",
	"fatih/color",
	"gookit/color",
	"logrusorgru/aurora",
	"go-colorable",
	"bubbletea",
	"bubbles",
}

// tuiPrefix is the one package allowed to emit escapes (R-3.10). It does not
// exist until Phase 7 — the walk simply never reaches it, and starts exempting
// it the moment it appears, with no edit here.
var tuiPrefix = filepath.Join("internal", "tui")

// goRoots are the module's Go trees, relative to this package. `vendor/` is
// excluded on purpose: it is the *Python* PyYAML tree (see go.work), not a Go
// vendor directory, and holds nothing this rule governs.
var goRoots = []string{
	filepath.Join("..", "..", "internal"),
	filepath.Join("..", "..", "cmd"),
}

func TestNoANSIEscapesInSource(t *testing.T) {
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving module root: %v", err)
	}
	fset := token.NewFileSet()
	var violations []string
	scanned := 0

	for _, root := range goRoots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") ||
				strings.HasSuffix(path, "_test.go") {
				return nil
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(moduleRoot, abs)
			if err != nil {
				return err
			}
			if strings.HasPrefix(rel, tuiPrefix+string(filepath.Separator)) {
				return nil // the one package R-3.10 exempts
			}
			scanned++

			file, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				return perr
			}

			for _, imp := range file.Imports {
				p, uerr := strconv.Unquote(imp.Path.Value)
				if uerr != nil {
					continue
				}
				for _, lib := range styleLibs {
					if strings.Contains(p, lib) {
						violations = append(violations, fmt.Sprintf(
							"%s: imports %q — a styling library outside %s "+
								"(R-3.10: only `tui` may emit escapes)",
							fset.Position(imp.Pos()), p, tuiPrefix))
					}
				}
			}

			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || (lit.Kind != token.STRING && lit.Kind != token.CHAR) {
					return true
				}
				if !literalHasESC(lit) {
					return true
				}
				violations = append(violations, fmt.Sprintf(
					"%s: string literal %s contains an ANSI escape (0x1b) — "+
						"R-3.10 forbids escape sequences outside %s; the goldens "+
						"were captured against a binary that emits none",
					fset.Position(lit.Pos()), lit.Value, tuiPrefix))
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", root, err)
		}
	}

	// A walk that silently matched nothing would pass forever. Guard the guard.
	if scanned == 0 {
		t.Fatalf("scanned no Go files under %v — the walk is broken, not clean", goRoots)
	}
	for _, v := range violations {
		t.Error(v)
	}
}

// literalHasESC reports whether a string or rune literal carries an escape byte,
// however it is spelled. Unquoting is what makes that "however": "\x1b[",
// "\033[", "\u001b[", '\x1b' and a raw backquoted literal holding a real 0x1b
// all decode to the same byte, so one check covers every encoding instead of a
// list of spellings that a new one can walk around.
func literalHasESC(lit *ast.BasicLit) bool {
	if lit.Kind == token.CHAR {
		r, _, _, err := strconv.UnquoteChar(strings.Trim(lit.Value, "'"), '\'')
		return err == nil && r == esc
	}
	if v, err := strconv.Unquote(lit.Value); err == nil {
		return strings.IndexByte(v, esc) >= 0
	}
	// Unquote failed, so fall back to the source text rather than skipping the
	// literal: a check that goes quiet on input it cannot parse is not a check.
	raw := lit.Value
	for _, spelling := range []string{`\x1b`, `\x1B`, `\033`, `\u001b`, `\u001B`, `\e`} {
		if strings.Contains(raw, spelling) {
			return true
		}
	}
	return strings.IndexByte(raw, esc) >= 0
}

// TestNoANSIEscapesAtRuntime asserts the property the goldens actually depend
// on: not one byte of real output matches an escape. It builds and execs the
// binary rather than calling run() in-process, because R-3.10 is a statement
// about what the shipped artifact writes to a pipe, and the tty-detection seam
// (tty.go) means the in-process path is not the same path.
func TestNoANSIEscapesAtRuntime(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain on PATH; cannot build the binary to sweep")
	}
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "company-os")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the binary: %v\n%s", err, out)
	}

	ws := filepath.Join(tmp, "ws")
	// The corpus spans the axes that could each acquire styling independently:
	// global flags, a mutating command, list output, the record renderers, the
	// federation gate, and all three failure shapes (usage, refusal, crash-ish).
	corpus := [][]string{
		{"--version"},
		{"--help"},
		{},
		{"--not-a-flag"},
		{"nonesuch"},
		{"--root", ws, "init", "--company", "Acme", "--team", "core", "--platform", "plat"},
		{"--root", ws, "ids", "list"},
		{"--root", ws, "skills", "list"},
		{"--root", ws, "today"},
		{"--root", ws, "today", "--role", "developer"},
		{"--root", ws, "graph", "build"},
		{"--root", ws, "governance", "resolve", "--team", "core"},
		{"--root", ws, "workspace", "status"},
		{"--root", ws, "add", "component", "--platform", "plat", "--id", "svc-a",
			"--name", "Service A", "--team", "core"},
		{"--root", ws, "reality", "new", "--platform", "plat", "svc-a"},
		{"--root", ws, "validate"},
		{"--root", ws, "prd", "--help"},
		{"--root", filepath.Join(tmp, "nowhere"), "today"},
	}
	// The worked example is richer than anything init can scaffold — every gate
	// renderer, every severity. Skipped rather than failed if the layout moves,
	// since this test is not the one that should police the fixture tree.
	if example, err := filepath.Abs(
		filepath.Join("..", "..", "..", "examples", "workspace")); err == nil {
		if _, serr := os.Stat(example); serr == nil {
			corpus = append(corpus,
				[]string{"--root", example, "validate"},
				[]string{"--root", example, "today"},
				[]string{"--root", example, "graph", "build"},
				[]string{"--root", example, "governance", "resolve",
					"--team", "customer-engagement"},
				[]string{"--root", example, "skills", "list"},
				[]string{"--root", example, "ids", "list"},
			)
		}
	}

	for _, argv := range corpus {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(bin, argv...)
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		cmd.Stdin = nil // never interactive; the init wizard must not engage
		_ = cmd.Run()   // a non-zero exit is expected for several of these
		for _, stream := range []struct {
			name string
			buf  *bytes.Buffer
		}{{"stdout", &stdout}, {"stderr", &stderr}} {
			if i := bytes.IndexByte(stream.buf.Bytes(), esc); i >= 0 {
				t.Errorf("company-os %s wrote an ANSI escape to %s at byte %d: %q",
					strings.Join(argv, " "), stream.name, i,
					excerpt(stream.buf.Bytes(), i))
			}
		}
	}
}

// excerpt returns the bytes around an offset, quoted, so a failure shows the
// sequence instead of printing it and vanishing into the terminal.
func excerpt(b []byte, at int) string {
	lo, hi := at-20, at+20
	if lo < 0 {
		lo = 0
	}
	if hi > len(b) {
		hi = len(b)
	}
	return string(b[lo:hi])
}
