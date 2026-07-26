package governance

// validate's first two gates: ownership reconciliation (bin/company-os:936-951)
// and deviation/exception expiry (`:954-974`).
//
// They live here rather than in internal/validate because both are governance
// questions — "does the team registry agree with the platform catalog" and "is
// this escape hatch still in date" — and internal/validate's job is composition
// and ordering, not derivation. Each returns a finished model.GateResult, so
// nothing about either gate requires reaching into this package.

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// The two gates' identities. The header text is frozen by the golden snapshots
// and gates 1-7 are never renumbered, so the ordinal is the caller's.
const (
	OwnershipGateSlug  = "ownership-reconciliation"
	OwnershipGateTitle = "ownership reconciliation"

	ExpiryGateSlug  = "governance-expiry"
	ExpiryGateTitle = "deviation and exception expiry"
)

// isoDate is the date layout `datetime.date.fromisoformat` accepts.
const isoDate = "2006-01-02"

// OwnershipGate is validate's gate 1 (`:936-951`): every component a team's
// ownership registry claims, checked against the platform catalog descriptor
// that is the single source of truth for the accountable team.
//
// The three findings carry three DIFFERENT subjects — the team id, the
// single-quoted component id, the bare component id — which is the shape the
// LLD calls "gate 1 has three prefix shapes". They are three Subject values, not
// three renderer branches (R-2.5).
func OwnershipGate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: OwnershipGateSlug, Title: OwnershipGateTitle}
	for _, tdir := range ws.AllTeams() {
		teamID := filepath.Base(tdir)
		path := filepath.Join(tdir, "ownership", "components.yaml")
		ownership, err := loadOr(path, pyMap{})
		if err != nil {
			return g, err
		}
		owned, err := seqAt(ownership, "components", relTo(ws.Root, path))
		if err != nil {
			return g, err
		}
		for _, raw := range owned {
			comp, ok := raw.(pyMap)
			if !ok {
				// `comp["id"]` on a non-mapping raises TypeError (R-0.7a(j)).
				return g, model.Errorf(model.ExitArtifact,
					"%s: ownership components entries must be mappings", relTo(ws.Root, path))
			}
			cidVal, err := index(comp, "id", "ownership component")
			if err != nil {
				return g, err
			}
			cid := yamlio.PyString(cidVal)

			platform, descPath, found := ws.FindComponent(cid)
			var desc pyMap
			haveDesc := false
			if found {
				// find_component returns load_yaml(f) with the DEFAULT None, so an
				// empty or otherwise falsy descriptor is `desc is None` exactly as
				// an absent file is (R-1.7a).
				loaded, err := yamlio.PyLoadFile(descPath)
				if err != nil {
					return g, err
				}
				if !yamlio.PyFalsy(loaded) {
					m, ok := loaded.(pyMap)
					if !ok {
						// `desc.get("ownership", {})` raises AttributeError.
						return g, model.Errorf(model.ExitArtifact,
							"%s: expected a mapping at the document root", relTo(ws.Root, descPath))
					}
					desc, haveDesc = m, true
				}
			}
			if !haveDesc {
				fields := model.Fields{"team": teamID, "component": cid}
				g.Findings = append(g.Findings, finding(model.SevFail,
					model.CodeOwnershipDescriptorMissing, teamID, fields))
				continue
			}

			// `(desc.get("ownership", {}) or {}).get("accountableTeam", "")` — the
			// `or {}` is what makes an explicit `ownership: null` behave as an
			// absent one.
			ownershipBlock, ok := getDefault(desc, "ownership", pyMap{}).(pyMap)
			if !ok && !yamlio.PyFalsy(getDefault(desc, "ownership", pyMap{})) {
				// A truthy non-mapping reaches `.get` and raises AttributeError.
				return g, model.Errorf(model.ExitArtifact,
					"%s: 'ownership' must be a mapping", relTo(ws.Root, descPath))
			}
			acc := getDefault(ownershipBlock, "accountableTeam", pyStr(""))

			// Python compares a loaded value against a str with `!=`, so a
			// non-string accountableTeam is never equal and always mismatches.
			claimsAccountable := yamlio.PyEqual(comp.Get("relationship"), pyStr("accountable"))
			if claimsAccountable && !yamlio.PyEqual(acc, pyStr("team://"+teamID)) {
				fields := model.Fields{
					"component":       cid,
					"team":            teamID,
					"accountableTeam": yamlio.PyString(acc),
				}
				g.Findings = append(g.Findings, finding(model.SevFail,
					model.CodeOwnershipAccountableMismatch, "'"+cid+"'", fields))
				continue
			}
			fields := model.Fields{"component": cid, "platform": platform}
			g.Findings = append(g.Findings, finding(model.SevOK,
				model.CodeOwnershipAgrees, cid, fields))
		}
	}
	return g, nil
}

// ExpiryGate is validate's gate 2 (`:954-974`): a deviation whose review date
// has passed, and an exception with no expiry or a past one.
//
// Every finding in the gate is prefixed with the TEAM id, including the ones
// about a specific rule — the rule is inside the sentence, not in the prefix.
//
// On error this returns the findings accumulated SO FAR, not a zero GateResult,
// and that is a contract rather than an accident. The oracle prints as it goes,
// so an unparseable date at `:970` leaves the lines above it already on stdout;
// internal/validate keeps the partial gate and renders it under the carried
// denominator (R-2.6a). Returning an empty result here would silently delete
// those lines again. `before` is the site that reaches it: an `expires:` or
// `reviewDate:` value Python hands to `date.fromisoformat` without a try.
func ExpiryGate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: ExpiryGateSlug, Title: ExpiryGateTitle}
	now := today()
	for _, tdir := range ws.AllTeams() {
		teamID := filepath.Base(tdir)

		devPath := filepath.Join(tdir, "governance", "deviations.yaml")
		dev, err := loadOr(devPath, pyMap{})
		if err != nil {
			return g, err
		}
		list, err := seqAt(dev, "deviations", relTo(ws.Root, devPath))
		if err != nil {
			return g, err
		}
		for _, raw := range list {
			d, ok := raw.(pyMap)
			if !ok {
				return g, model.Errorf(model.ExitArtifact,
					"%s: deviations entries must be mappings", relTo(ws.Root, devPath))
			}
			ruleVal, err := index(d, "rule", "deviation")
			if err != nil {
				return g, err
			}
			rd := d.Get("reviewDate")
			fields := model.Fields{
				"team": teamID, "rule": yamlio.PyString(ruleVal),
				"reviewDate": yamlio.PyString(rd),
			}
			// `if rd and date.fromisoformat(str(rd)) < TODAY` — a falsy reviewDate
			// short-circuits into the ok branch and renders as its own str().
			past, err := before(rd, now, relTo(ws.Root, devPath), "reviewDate")
			if err != nil {
				return g, err
			}
			if past {
				g.Findings = append(g.Findings, finding(model.SevFail,
					model.CodeDeviationExpired, teamID, fields))
				continue
			}
			g.Findings = append(g.Findings, finding(model.SevOK,
				model.CodeDeviationCurrent, teamID, fields))
		}

		excPath := filepath.Join(tdir, "governance", "exceptions.yaml")
		exc, err := loadOr(excPath, pyMap{})
		if err != nil {
			return g, err
		}
		exceptions, err := seqAt(exc, "exceptions", relTo(ws.Root, excPath))
		if err != nil {
			return g, err
		}
		for _, raw := range exceptions {
			e, ok := raw.(pyMap)
			if !ok {
				return g, model.Errorf(model.ExitArtifact,
					"%s: exceptions entries must be mappings", relTo(ws.Root, excPath))
			}
			// `e.get("rule")` — absent renders as None rather than raising, unlike
			// the deviation loop's `d["rule"]` one block above.
			rule := yamlio.PyString(e.Get("rule"))
			ex := e.Get("expires")
			fields := model.Fields{"team": teamID, "rule": rule, "expires": yamlio.PyString(ex)}
			if yamlio.PyFalsy(ex) {
				delete(fields, "expires")
				g.Findings = append(g.Findings, finding(model.SevFail,
					model.CodeExceptionNoExpiry, teamID, fields))
				continue
			}
			past, err := before(ex, now, relTo(ws.Root, excPath), "expires")
			if err != nil {
				return g, err
			}
			if past {
				g.Findings = append(g.Findings, finding(model.SevFail,
					model.CodeExceptionExpired, teamID, fields))
				continue
			}
			g.Findings = append(g.Findings, finding(model.SevOK,
				model.CodeExceptionValid, teamID, fields))
		}
	}
	return g, nil
}

// before is `value and dt.date.fromisoformat(str(value)) < TODAY`.
//
// A falsy value never reaches the parse, matching the `and`. A value that is not
// an ISO date raises ValueError in Python and exits 1 through a traceback; here
// it is a coded error carrying the same status, so the exit code agrees and the
// diagnostic names the file (R-0.7a).
func before(v pyVal, now time.Time, path, field string) (bool, error) {
	if yamlio.PyFalsy(v) {
		return false, nil
	}
	text := yamlio.PyString(v)
	d, err := time.Parse(isoDate, text)
	if err != nil {
		return false, model.Errorf(model.ExitValidation,
			"%s: '%s: %s' is not an ISO-8601 date (YYYY-MM-DD)", path, field, text)
	}
	return d.Before(now), nil
}

// finding builds one gate record, composing its sentence through Message and
// nowhere else.
func finding(sev model.Severity, code, subject string, f model.Fields) model.Finding {
	return model.Finding{
		Severity: sev,
		Code:     code,
		Subject:  subject,
		Message:  Message(code, f),
		Fields:   f,
	}
}

// gateMessage composes the eight gate-1 and gate-2 sentences. It is folded into
// the package's single Message below rather than exported separately, so this
// cluster still has exactly one function that produces prose.
func gateMessage(code string, f model.Fields) (string, bool) {
	switch code {
	case model.CodeOwnershipDescriptorMissing:
		return fmt.Sprintf("owns '%s' but no descriptor in any platform catalog",
			f.Str("component")), true
	case model.CodeOwnershipAccountableMismatch:
		return fmt.Sprintf(
			"team '%s' claims accountable but descriptor says '%s' (single-source rule: "+
				"descriptor ownership must match team registry)",
			f.Str("team"), f.Str("accountableTeam")), true
	case model.CodeOwnershipAgrees:
		return fmt.Sprintf("registry and descriptor agree (%s)", f.Str("platform")), true
	case model.CodeDeviationExpired:
		return fmt.Sprintf("deviation for %s expired %s — re-review or remove",
			f.Str("rule"), f.Str("reviewDate")), true
	case model.CodeDeviationCurrent:
		return fmt.Sprintf("deviation %s current (review %s)",
			f.Str("rule"), f.Str("reviewDate")), true
	case model.CodeExceptionNoExpiry:
		return fmt.Sprintf("exception for %s has NO expiry — invalid", f.Str("rule")), true
	case model.CodeExceptionExpired:
		return fmt.Sprintf("exception for %s expired %s", f.Str("rule"), f.Str("expires")), true
	case model.CodeExceptionValid:
		return fmt.Sprintf("exception %s valid until %s", f.Str("rule"), f.Str("expires")), true
	}
	return "", false
}
