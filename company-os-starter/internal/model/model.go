// Package model holds the record types every command returns and the exit-code
// contract main maps them onto. It is the bottom of the dependency graph: it
// imports nothing from the rest of the tree, and every other internal package
// may import it.
//
// The central rule of the port is "records before renderers" — commands compute
// GateResult/Finding values and return them; only cmd/company-os turns them into
// bytes and an exit status.
package model

import (
	"errors"
	"fmt"
)

// Severity is the three-way outcome of a single finding, rendered by the text
// renderer as "  [ok] ", "  [warn] " and "  [FAIL] " respectively.
type Severity int

const (
	SevOK Severity = iota
	SevWarn
	SevFail
)

// String is the JSON-facing name. The bracketed text markers are the text
// renderer's business, not the model's.
func (s Severity) String() string {
	switch s {
	case SevOK:
		return "ok"
	case SevWarn:
		return "warn"
	case SevFail:
		return "fail"
	}
	return "unknown"
}

// Fields is a finding's structured payload, consumed by BOTH renderers (R-2.3).
//
// The value type is `any` rather than `string` because measurement against the
// oracle rules `map[string]string` out on two independent counts:
//
//   - `bin/company-os:990` renders an ORDERED LIST inside the sentence —
//     "missing frontmatter ['team', 'components', 'governanceSnapshot']" — and
//     the element order is what the human reads. A string map cannot hold it.
//   - The counts in "[ok] communications: feature-index in sync (1 component(s))"
//     and "[ok] skills layered cleanly (2 canonical, 0 team; …)" must reach the
//     text output as numbers and the JSON output as `1`, not `"1"` (R-2.2 says
//     "typed fields").
//
// Key order carries no meaning: the text renderer addresses fields by name, and
// encoding/json sorts map keys, so JSON stays deterministic. Order *within* a
// slice value is load-bearing and preserved.
type Fields map[string]any

// Str returns the string at key, or "" when absent or of another type. The
// accessors never panic: a producer/renderer mismatch should render a wrong
// line, not crash a CLI mid-gate.
func (f Fields) Str(key string) string {
	s, _ := f[key].(string)
	return s
}

// Int returns the int at key, or 0 when absent or of another type.
func (f Fields) Int(key string) int {
	n, _ := f[key].(int)
	return n
}

// Strs returns the []string at key, or nil when absent or of another type.
// The slice order is preserved because it is rendered verbatim.
func (f Fields) Strs(key string) []string {
	s, _ := f[key].([]string)
	return s
}

// Finding is one rendered line's worth of structure, kept structured until the
// renderer runs. Message carries the sentence body only: no severity marker and
// no indentation.
//
// Subject is the line's prefix token and is render-ready text, including any
// quoting the oracle applies. `bin/company-os:946` prefixes the component id
// wrapped in single quotes ("[FAIL] 'svc-alpha': team 'ghost' claims …") while
// `:951` uses the bare id and `:941` uses the *team* id — three shapes inside
// one gate. Because Subject may therefore carry punctuation, the clean machine
// value always also lives in Fields, which is what JSON consumers read.
//
// The renderer applies one uniform rule — Subject, then ": ", then Message, and
// Message alone when Subject is empty. All seven "prefix shapes" in the LLD are
// seven distinct Subject *values*, not seven renderer branches.
type Finding struct {
	Severity Severity
	Code     string // stable machine identifier, e.g. CodeOwnershipAgrees
	Subject  string // render-ready prefix token; "" when the gate has no prefix
	Path     string // workspace-relative path, "" when not file-scoped
	Message  string // sentence body
	Fields   Fields
}

// GateResult groups the findings of one gate. A gate that ran and found nothing
// still returns a GateResult with zero findings — a flat []Finding cannot
// reproduce a golden in which a gate header is followed by no lines
// (examples/golden-validate.txt:11-12, R-2.1).
//
// Findings is one ordered slice and must never be bucketed by severity: the
// oracle interleaves the three severities in document order within a single gate
// (examples/failing-workspace-golden-validate.txt:26-31 is ok, four warns, then
// a fail for the next document).
//
// Two behaviours need no field of their own, and deliberately do not get one:
//
//   - The leading blank line is a property of the gate header, present on every
//     gate except the first (R-2.6) — that is exactly Ordinal > 1.
//   - Gate 4's conditional `[ok]` (`:1003-1008`) and gate 6's (`:1058`) are
//     expressed by the producer not appending an ok Finding. "This gate produced
//     failures and deliberately no ok line" is already the absence of a record.
type GateResult struct {
	Ordinal  int    // 1..N, rendered as [N/7] or [N/8]
	Slug     string // stable, JSON-facing, e.g. "ownership-reconciliation"
	Title    string // human header text, e.g. "ownership reconciliation"
	Findings []Finding
}

// Report is the whole of one `validate` run. A bare []GateResult is not enough:
// the oracle's first line is "validating workspace <root>" (`:924`), and Root is
// the one piece of the golden that cannot be derived from the gates.
//
// Total is the [N/M] denominator and is CARRIED, not derived from len(Gates)
// (R-2.6a). The oracle decides 7-versus-8 from manifest presence at `:930`,
// before gate 1 runs, and then prints every header against that number — so on a
// run that aborts mid-gate the gates that already printed still belong under the
// original denominator. Deriving it would render `[1/2]`/`[2/2]` where the
// oracle wrote `[1/7]`/`[2/7]`, which is a false claim about how much of the
// workspace was checked; the alternative the port took before this field existed
// was to drop the completed gates entirely, which removes human-facing lines
// R-0.8 forbids removing. The count is knowable before any gate runs, so neither
// lie is necessary.
//
// The trailer count stays derived: it is Problems().
type Report struct {
	Root  string
	Total int
	Gates []GateResult
}

// Problems counts the findings that make the run fail, reproducing the Python
// `problems` integer (`:923`). Warnings are excluded: the failing-workspace
// oracle has 15 fails and 4 warns and its trailer reads "FAIL — 15 problem(s)".
func (r Report) Problems() int {
	n := 0
	for _, g := range r.Gates {
		for _, f := range g.Findings {
			if f.Severity == SevFail {
				n++
			}
		}
	}
	return n
}

// ExitCode maps the report onto the process status: 0 when clean, 1 when any
// gate reported [FAIL] (`:1107`).
func (r Report) ExitCode() ExitCode {
	if r.Problems() > 0 {
		return ExitValidation
	}
	return ExitOK
}

// ExitCode is the process exit status. The eight codes are the contract agents
// branch on; see docs/lld/go-cli-tui-port.md § "Exit codes" and the full
// per-site classification at .devlocal/go-port/exit-code-map.md.
type ExitCode int

const (
	// ExitOK is success.
	ExitOK ExitCode = 0
	// ExitValidation means one or more gates reported [FAIL].
	ExitValidation ExitCode = 1
	// ExitUsage means bad flags, a missing required argument, an unknown
	// subcommand, or a bare invocation. This is already argparse's behavior.
	ExitUsage ExitCode = 2
	// ExitWorkspace means not a workspace root, or a required workspace object
	// does not exist.
	ExitWorkspace ExitCode = 3
	// ExitArtifact means malformed YAML/frontmatter or a schema violation.
	ExitArtifact ExitCode = 4
	// ExitPrecondition means a done-gate refusal or an unvalidated discovery brief.
	ExitPrecondition ExitCode = 5
	// ExitExternalTool means git is missing or too old, a clone/sparse-checkout
	// failed, or --frozen lock reconciliation failed.
	ExitExternalTool ExitCode = 6
	// ExitInteractive means a TTY was required but absent.
	ExitInteractive ExitCode = 7
	// ExitConflict means refusing to overwrite an existing artifact.
	ExitConflict ExitCode = 8
)

// Error is an error carrying an exit code. Packages below cmd/ return these
// instead of calling os.Exit, which is what keeps the whole tree usable from the
// TUI and from --json without a second code path.
type Error struct {
	Code ExitCode
	Msg  string
}

func (e *Error) Error() string { return e.Msg }

// ExitCode makes *Error an ExitCoder. It is the reason CodeOf needs exactly one
// lookup rather than one per error type.
func (e *Error) ExitCode() ExitCode { return e.Code }

// ExitCoder is an error that knows which of the eight codes it maps to.
//
// Most producers return *Error and get this for free. A package that already has
// a typed error its own callers match on — internal/yamlio's SyntaxError is the
// case that forced this — implements the method instead of wrapping, so the type
// keeps working with errors.As on both sides of the seam.
//
// This interface is what lets cmd/company-os classify without string matching,
// which R-4.1..R-4.9 require: a message reworded upstream must never move an
// exit code.
type ExitCoder interface {
	error
	ExitCode() ExitCode
}

// Errorf builds an *Error with the given code.
func Errorf(code ExitCode, format string, a ...any) error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, a...)}
}

// CodeOf reports the exit code an error should produce.
//
// The lookup walks the %w chain and stops at the OUTERMOST ExitCoder, so a
// caller that re-classifies an inner error by wrapping it wins — which is what
// makes `fmt.Errorf("...: %w", err)` safe everywhere else in the tree.
//
// An error that carries no code resolves to ExitValidation (1), matching today's
// Python behavior, where an uncaught exception exits 1 through a traceback.
func CodeOf(err error) ExitCode {
	if err == nil {
		return ExitOK
	}
	var c ExitCoder
	if errors.As(err, &c) {
		return c.ExitCode()
	}
	return ExitValidation
}

// HasFailure reports whether any gate produced a SevFail finding, which is what
// maps a successful command run onto ExitValidation.
func HasFailure(results []GateResult) bool {
	for _, r := range results {
		for _, f := range r.Findings {
			if f.Severity == SevFail {
				return true
			}
		}
	}
	return false
}

// QuietError is an error whose diagnostic has ALREADY been written to the
// command's own output stream, so the dispatcher must not print it again.
//
// Exactly one site needs it: `prd complete`'s done-gate refusal
// (bin/company-os:699-707) prints its whole block — a banner, one `[FAIL]` line
// per problem, one `fix:` pointer per missing reality doc — to STDOUT and then
// exits 5 with nothing on stderr. Without this the port would either lose the
// block (records are dropped on the error path) or gain an `error: …` line the
// oracle never writes.
//
// It is a wrapper rather than a flag on *Error so that CodeOf keeps working
// unchanged: the wrapped error is still the outermost ExitCoder.
type QuietError struct{ Err error }

func (q *QuietError) Error() string { return q.Err.Error() }

func (q *QuietError) Unwrap() error { return q.Err }

// ExitCode forwards the wrapped error's code, keeping *QuietError an ExitCoder
// so CodeOf stops at it rather than falling through to the default.
func (q *QuietError) ExitCode() ExitCode { return CodeOf(q.Err) }

// Quiet marks an error as already-reported. Quiet(nil) is nil.
func Quiet(err error) error {
	if err == nil {
		return nil
	}
	return &QuietError{Err: err}
}

// IsQuiet reports whether the error's diagnostic has already been emitted.
func IsQuiet(err error) bool {
	var q *QuietError
	return errors.As(err, &q)
}

// UsageError marks an error as an argparse-shaped ARGUMENT error scoped to one
// sub-parser, so the dispatcher renders it the way argparse renders the errors
// it catches itself: the sub-parser's `usage:` line, then
// `company-os <scope>: error: <msg>` (R-1.4a).
//
// The parser in cmd/company-os produces these for everything argparse can
// express. This type exists for the requirements argparse CANNOT express — a
// positional declared `nargs="?"` because it is required for one action and
// meaningless for another, or a flag required only in the presence of another
// flag. Those checks can only run in command code, below cmd/, where the
// parser's own usageError type is out of reach; without a marker their
// diagnostic came out as a bare `error: …` line with no usage block and no
// `company-os ` prefix, which is the one stderr shape four shipped agent skills
// cannot grep for. R-0.7a(l) is the clause that sanctions replacing the oracle's
// traceback on these paths.
//
// Scope is the SUBCOMMAND, never the action: argparse interpolates its own
// `prog` before `: error:`, and a sub-parser's prog is `company-os <sub>`.
//
// Like QuietError this is a wrapper rather than a field on *Error, so CodeOf
// still stops at the outermost ExitCoder.
type UsageError struct {
	Scope string
	Err   error
}

func (u *UsageError) Error() string { return u.Err.Error() }

func (u *UsageError) Unwrap() error { return u.Err }

// ExitCode forwards the wrapped error's code, keeping *UsageError an ExitCoder.
func (u *UsageError) ExitCode() ExitCode { return CodeOf(u.Err) }

// Usagef builds an ExitUsage error scoped to the named subcommand.
func Usagef(scope, format string, a ...any) error {
	return &UsageError{Scope: scope, Err: Errorf(ExitUsage, format, a...)}
}
