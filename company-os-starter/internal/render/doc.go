// Package render turns []model.GateResult into bytes.
//
// Responsibility: text.go reproduces the Python output byte-for-byte, including
// the per-gate prefix policy and the rule that a blank line belongs to the gate
// header rather than to a finding; json.go emits the --json form. Neither writes
// to stdout — both take an io.Writer supplied by cmd/.
//
// Not implemented — Phase 3.
package render
