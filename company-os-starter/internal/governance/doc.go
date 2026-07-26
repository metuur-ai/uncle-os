// Package governance resolves effective team governance and explains why a rule
// applies to a component.
//
// Responsibility: `governance resolve`, `governance explain`, `deviation
// declare`, `exception request`, and the tier model (mandatory / default /
// guidance).
//
// Two things about this cluster are worth knowing before editing it.
//
// It owns the ONLY two commands in the CLI that read-modify-write a file a
// human authored — `deviation declare` and `exception request`. declare.go's
// header records why the load/emit pair is PyLoadFile/PyWriteFile and not a
// yaml.Node round trip: the oracle already reflows both files, so matching its
// bytes means reproducing that reflow, and R-0.7a(g) goes unexercised here
// rather than being relied on.
//
// It also owns the third and last R-0.7c site,
// `teams/<t>/generated/effective-governance.yaml`. That one is DERIVED, so its
// write is guarded on a semantic compare and skipped when the derivation has
// not changed — see writeGuarded in resolve.go, which also explains why the
// compare has to hold `generatedAt` equal. The authored files get no such
// guard; the derived/authored split is the whole of the difference between
// R-0.7c and R-0.7a(g).
//
// The tier model is enforced by RECORDING, not refusing: a deviation aimed at a
// mandatory rule is written into the generated file as `deviationRejected` and
// the resolve continues, because the tier is only knowable after resolution.
package governance
