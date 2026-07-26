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
// Everything else stays derived rather than stored. The [N/M] denominator is
// len(Gates) — gate 8 exists only in federated mode (`:930`), so the gate list
// already carries it. The trailer count is Problems().
type Report struct {
	Root  string
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

// Finding codes — one per render site in cmd_validate (bin/company-os:922-1107),
// stable across message rewordings (R-2.4). The `fail(c)` and `fail(pr)` sites
// at `:1078` and `:1100` pass through pre-composed prose from skill_conflicts
// and federated_slice_problems; R-2.12 requires decomposing those, so each
// distinct sentence they produce gets its own code here rather than one code per
// call site.
const (
	// Gate 1 — ownership reconciliation (:941, :946, :951).
	CodeOwnershipDescriptorMissing   = "ownership.descriptor-missing"
	CodeOwnershipAccountableMismatch = "ownership.accountable-mismatch"
	CodeOwnershipAgrees              = "ownership.agrees"

	// Gate 2 — deviation and exception expiry (:960, :963, :968, :971, :974).
	CodeDeviationExpired  = "expiry.deviation-expired"
	CodeDeviationCurrent  = "expiry.deviation-current"
	CodeExceptionNoExpiry = "expiry.exception-no-expiry"
	CodeExceptionExpired  = "expiry.exception-expired"
	CodeExceptionValid    = "expiry.exception-valid"

	// Gate 3 — active PRD contracts (:990, :993).
	CodePRDFrontmatterMissing = "prd.frontmatter-missing"
	CodePRDContractPresent    = "prd.contract-present"

	// Gate 4 — frontmatter core and tag derivation (:1002, :1006, :1010, :1013).
	CodeFrontmatterCoreField = "frontmatter.core-field"
	CodeTagsDrift            = "frontmatter.tags-drift"
	CodeFrontmatterInSync    = "frontmatter.in-sync"
	CodePointerGuidance      = "frontmatter.pointer-guidance"

	// Gate 5 — CLAUDE.md context node drift (:1025, :1029, :1033, :1037, :1041).
	CodeNodeIdentity  = "node.identity"
	CodeNodeAbsent    = "node.absent"
	CodeNodeHandOwned = "node.hand-owned"
	CodeNodeDrift     = "node.drift"
	CodeNodeInSync    = "node.in-sync"

	// Gate 6 — feature-index drift (:1050, :1055, :1062, :1066).
	CodeFeatureIndexAbsent     = "feature-index.absent"
	CodeFeatureIndexDrift      = "feature-index.drift"
	CodeFeatureIndexUnresolved = "feature-index.unresolved-reference"
	CodeFeatureIndexInSync     = "feature-index.in-sync"

	// Gate 7 — custom skills layering (:1078 decomposed, :1087).
	CodeSkillShadowing       = "skills.shadowing"
	CodeSkillDanglingExtends = "skills.dangling-extends"
	CodeSkillsClean          = "skills.clean"

	// Gate 8 — federated slice integrity (:1100 decomposed, :1103).
	CodeSliceLockMissing      = "federation.lock-missing"
	CodeRepoNotLocked         = "federation.repo-not-locked"
	CodeSliceSetDrift         = "federation.slice-set-drift"
	CodeSliceFileMissing      = "federation.slice-file-missing"
	CodeSliceHandEdited       = "federation.slice-hand-edited"
	CodeFederationSlicesMatch = "federation.slices-match"
)

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

// Errorf builds an *Error with the given code.
func Errorf(code ExitCode, format string, a ...any) error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, a...)}
}

// CodeOf reports the exit code an error should produce. An error that carries no
// code resolves to ExitValidation (1), matching today's Python behavior, where an
// uncaught exception exits 1 through a traceback.
func CodeOf(err error) ExitCode {
	if err == nil {
		return ExitOK
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
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
