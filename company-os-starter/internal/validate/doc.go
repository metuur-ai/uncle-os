// Package validate runs the workspace CI gates.
//
// Responsibility: compose the seven or eight gates in the oracle's order and
// return one model.GateResult each, including the gates that ran and found
// nothing — a flat finding list cannot reproduce a golden in which a gate header
// is followed by zero lines.
//
// Six of the eight gates are produced by the cluster that owns their subject
// matter; only gate 4 is built here, because it is the only one that spans two
// clusters. Nothing in this package composes prose.
package validate
