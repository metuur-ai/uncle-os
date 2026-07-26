// Package validate runs the workspace CI gates.
//
// Responsibility: each gate returns one model.GateResult, including the gates
// that ran and found nothing — a flat finding list cannot reproduce a golden in
// which a gate header is followed by zero lines.
//
// Not implemented — Phase 3.
package validate
