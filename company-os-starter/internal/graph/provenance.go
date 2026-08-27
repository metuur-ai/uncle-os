package graph

import (
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// Trust tiers (R-4.3). Advisory labels, computed at read time and never stored.
//
// They answer one question a reader of an agent-written corpus keeps asking and
// this system could not previously answer: did a person look at this? The whole
// skills/ layer exists to direct agents producing PRDs and discovery briefs, and
// until now nothing recorded that fact.
const (
	// TierUnverified is the absence of any `verified:` entry. It is the default
	// and it is not an accusation — most documents are simply unreviewed.
	TierUnverified = "unverified"
	// TierMachineConfirmed means every confirmer was an agent or a process.
	TierMachineConfirmed = "machine-confirmed"
	// TierHumanReviewed means at least one confirmer was a person.
	TierHumanReviewed = "human-reviewed"
)

// humanActor is the `human:<id>` prefix from the actor convention (R-4.2).
// Agents are `<producer>/<version>`; automated processes are `process:<id>`.
const humanActor = "human:"

// TrustTier is R-4.3 and R-4.4: an advisory label derived from `verified:`,
// computed at read time and never written back.
//
// Nothing may gate on it. That constraint is the point rather than a caveat: a
// trust signal that can block turns into an approval field, and this system
// already has three of those which disagree with each other. This one is
// allowed to be wrong without stopping anyone's work.
//
// Derivation is deliberately generous about shape. `verified:` is authored by
// hand and by agents, so a scalar where a list belongs, or an entry missing
// `by`, degrades toward *unverified* rather than erroring — an unparseable
// provenance block must never be the reason a workspace fails to validate.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.3
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-4.4
func TrustTier(meta yamlio.PyMap) string {
	actors := VerifiedActors(meta)
	if len(actors) == 0 {
		return TierUnverified
	}
	for _, a := range actors {
		if strings.HasPrefix(a, humanActor) {
			return TierHumanReviewed
		}
	}
	return TierMachineConfirmed
}

// VerifiedActors returns the `by` of every entry in `verified:`, plus the
// entries R-5.1 derives from the legacy approval fields.
//
// The derivation is ONE-WAY and in-memory. `decisionOwner` and the two
// `approvedBy` fields are never rewritten, never migrated, and keep every gate
// that consumes them (R-5.2). Reconciling them for real changes approval
// semantics, which is why the parent change deferred it and why this one does
// not attempt it.
//
// Without the derivation the tier would read *unverified* for every document in
// every fixture on the day it shipped — a signal carrying no signal.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-5.1
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-5.2
func VerifiedActors(meta yamlio.PyMap) []string {
	var out []string
	for _, entry := range mappingsOf(meta.Get("verified")) {
		if by := entry.Get("by"); !yamlio.PyFalsy(by) {
			out = append(out, yamlio.PyString(by))
		}
	}
	// Legacy approval fields. A person's name in `decisionOwner` or `approvedBy`
	// is a human sign-off, so it derives a `human:` actor — that is the whole
	// reason the derivation exists.
	for _, key := range []string{"decisionOwner", "approvedBy"} {
		raw := meta.Get(key)
		// PyFalsy, not `== ""`. PyString renders an absent key as Python's
		// "None", so a string comparison silently derives an actor named None
		// for every document that lacks the field — which is every document.
		if yamlio.PyFalsy(raw) {
			continue
		}
		v := yamlio.PyString(raw)
		if model.IsTODO(v) {
			continue
		}
		if strings.HasPrefix(v, humanActor) || strings.HasPrefix(v, "process:") {
			out = append(out, v)
			continue
		}
		out = append(out, humanActor+v)
	}
	return out
}

// GeneratedBy is the `by` of a `generated:` mapping, or "" when absent (R-4.1).
func GeneratedBy(meta yamlio.PyMap) string {
	m, ok := meta.Get("generated").(yamlio.PyMap)
	if !ok {
		return ""
	}
	by := m.Get("by")
	if yamlio.PyFalsy(by) {
		return ""
	}
	return yamlio.PyString(by)
}

// mappingsOf normalizes `verified:` into a list of mappings. A single mapping
// written where a list belongs is accepted, because that is the shape a human
// reaches for first and rejecting it would fail a workspace over metadata that
// blocks nothing.
func mappingsOf(v yamlio.PyValue) []yamlio.PyMap {
	switch t := v.(type) {
	case yamlio.PyMap:
		return []yamlio.PyMap{t}
	case yamlio.PySeq:
		var out []yamlio.PyMap
		for _, e := range t {
			if m, ok := e.(yamlio.PyMap); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}
