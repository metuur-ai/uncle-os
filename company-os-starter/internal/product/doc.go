// Package product owns the discovery -> PRD -> completion lifecycle.
//
// Responsibility: `discover new|validate`, `prd new|validate|complete`, the
// composable `check ready|done` gates, and validate gate 3 (active PRD
// contracts), whose subject is a change record and therefore belongs here
// rather than in internal/validate.
//
// Four things about this cluster are worth knowing before editing it.
//
// `prd complete` is the enforcement point for invariant #4 of the whole
// methodology — a change is done only when reality is updated. Its refusal is
// exit 5 and not 1 (.devlocal/go-port/exit-code-map.md § I): 5 means "go update
// reality/", 1 would mean "your artifact is malformed", and that distinction is
// the most useful one the exit contract makes for an agent driving the command.
// On success it MOVES the change record into archive/prds/, writes an
// outcome.md due in 90 days, appends log.md, and only then re-derives the
// generated artifacts — the one command whose own output brackets the derived
// output rather than following it.
//
// The refusal is also the one place a command prints a whole block to STDOUT and
// still exits non-zero without a stderr diagnostic. That is why Complete returns
// its records AND an error: the caller renders the records and model.IsQuiet
// keeps the dispatcher from adding an `error: …` line the oracle never prints.
//
// gather_prd_governance (`bin/company-os:551-570`) built a markdown checklist
// STRING and handed it to two unrelated consumers: `prd new` interpolated it
// into an artifact, `check` printed it. R-2.12 splits that in two — Gather
// returns []ChecklistItem, and ChecklistMarkdown is the one function that turns
// items into markdown. Both consumers call the renderer; neither composes a
// sentence.
//
// The done-gate's date comparison is the ONE place in this port that
// deliberately does not reproduce Python (R-1.14, OKF v0.2 Phase 0, sanctioned
// as R-0.7a(d)). See realityDateFinding in prd.go.
package product
