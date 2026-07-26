package governance

// `deviation declare` and `exception request` — the only two commands in the
// CLI that READ-MODIFY-WRITE a file a human authored.
//
// Everything unusual about this file follows from that. The load side is
// yamlio.PyLoadFile and the write side is yamlio.PyWriteFile, which is
// `dump_yaml` (`bin/company-os:65-69`) —
// `safe_dump(sort_keys=False, default_flow_style=False)` — transliterated. That
// pairing is deliberate and it is what makes the differential harness compare
// equal on file_tree:
//
//   - PyYAML's own safe_dump already fails to reproduce the committed bytes of
//     every deviations.yaml and exceptions.yaml in the corpus. The oracle
//     REFLOWS these files today: the committed flow-style `- {rule: …, tier: …}`
//     entries come back out as block mappings and `tags: [a, b]` as an
//     indentless block sequence. Matching Python means reproducing that reflow,
//     not preserving the authored layout.
//   - so the R-0.7a(g) carve-out — an authored artifact re-emitted under
//     yaml.v3's layout — is NOT exercised here. It exists for the case where no
//     PyYAML-compatible emitter is available; one is, and using it makes the
//     bytes identical rather than merely sanctioned.
//
// The consequence for R-0.7a(b) is the same trade the oracle makes: safe_load
// discards comments, so a comment in a hand-authored deviations.yaml does not
// survive the append here either. Preserving them would take the yaml.Node
// round trip, and that round trip is precisely what would reintroduce the
// layout divergence this file avoids. Byte parity with the oracle wins.
//
// Neither command validates. `deviation declare` accepts a rule aimed at a
// mandatory requirement and exits 0; the refusal is recorded by Resolve and
// surfaces as a validate gate failure. See deviationRejected in resolve.go.

import (
	"path/filepath"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// reviewDays is the comply-or-explain review interval (`:1119`).
const reviewDays = 180

// Declared is `deviation declare`'s outcome.
type Declared struct {
	// Rule is the rule id as given.
	Rule string
	// Rel is teams/<t>/governance/deviations.yaml relative to the root.
	Rel string
	// ReviewDate is TODAY + 180 days, rendered as date.isoformat().
	ReviewDate string
	// Team is the declaring team.
	Team string
}

// Declare is cmd_deviation (`bin/company-os:1112-1125`).
func Declare(ws *workspace.Workspace, teamID, rule, rationale string) (Declared, error) {
	tdir, err := ws.TeamDir(teamID)
	if err != nil {
		return Declared{}, err
	}
	path := filepath.Join(tdir, "governance", "deviations.yaml")
	review := today().AddDate(0, 0, reviewDays).Format("2006-01-02")

	data, err := loadOr(path, pyMap{
		{K: "schemaVersion", V: pyStr("1.0")},
		{K: "team", V: pyStr(teamID)},
		{K: "deviations", V: pySeq{}},
	})
	if err != nil {
		return Declared{}, err
	}
	next, err := appendEntry(data, path, "deviations", pyMap{
		{K: "rule", V: pyStr(rule)},
		{K: "tier", V: pyStr("default")},
		{K: "status", V: pyStr("declared")},
		{K: "rationale", V: pyStr(orDefault(rationale, "TODO: why does the team deviate?"))},
		{K: "reviewDate", V: pyStr(review)},
	})
	if err != nil {
		return Declared{}, err
	}
	if err := yamlio.PyWriteFile(path, next); err != nil {
		return Declared{}, err
	}
	return Declared{Rule: rule, Rel: relTo(ws.Root, path), ReviewDate: review, Team: teamID}, nil
}

// Requested is `exception request`'s outcome.
type Requested struct {
	// Rel is teams/<t>/governance/exceptions.yaml relative to the root.
	Rel string
	// Expires is the --expires value verbatim. The oracle neither parses nor
	// validates it; `validate` gate 2 is where a garbage or past date surfaces.
	Expires string
}

// Request is cmd_exception (`bin/company-os:1128-1138`).
func Request(ws *workspace.Workspace, teamID, rule, component, reason, expires string) (Requested, error) {
	tdir, err := ws.TeamDir(teamID)
	if err != nil {
		return Requested{}, err
	}
	path := filepath.Join(tdir, "governance", "exceptions.yaml")

	data, err := loadOr(path, pyMap{
		{K: "schemaVersion", V: pyStr("1.0")},
		{K: "team", V: pyStr(teamID)},
		{K: "exceptions", V: pySeq{}},
	})
	if err != nil {
		return Requested{}, err
	}
	next, err := appendEntry(data, path, "exceptions", pyMap{
		{K: "rule", V: pyStr(rule)},
		{K: "component", V: pyStr(component)},
		{K: "reason", V: pyStr(orDefault(reason, "TODO"))},
		{K: "compensatingControls", V: pySeq{pyStr("TODO")}},
		{K: "approvedBy", V: pyStr("TODO: rule owner")},
		{K: "expires", V: pyStr(expires)},
	})
	if err != nil {
		return Requested{}, err
	}
	if err := yamlio.PyWriteFile(path, next); err != nil {
		return Requested{}, err
	}
	return Requested{Rel: relTo(ws.Root, path), Expires: expires}, nil
}

// appendEntry is `data.setdefault(key, []).append(entry)`.
//
// setdefault means an ABSENT key is created at the END of the mapping — after
// `tags:`, if the file has one — and a present key keeps its authored position.
// PyMap.Set already has both halves of that behaviour, which is why the list is
// written back through it rather than mutated in place.
func appendEntry(data pyVal, path, key string, entry pyMap) (pyMap, error) {
	m, ok := data.(pyMap)
	if !ok {
		// `data.setdefault` on a non-mapping raises AttributeError and the file
		// is never rewritten. R-0.7a(j): same outcome, exit 4, nothing written.
		return nil, model.Errorf(model.ExitArtifact,
			"%s: expected a mapping at the document root", path)
	}
	existing := m.Get(key)
	if existing == nil {
		existing = pySeq{}
	}
	list, ok := existing.(pySeq)
	if !ok {
		// `.append` on a non-list raises AttributeError.
		return nil, model.Errorf(model.ExitArtifact, "%s: '%s' must be a sequence", path, key)
	}
	// A fresh backing array: the loaded slice may be shared with the document
	// the caller still holds.
	next := make(pySeq, 0, len(list)+1)
	next = append(next, list...)
	next = append(next, entry)

	out := make(pyMap, len(m))
	copy(out, m)
	return out.Set(key, next), nil
}

// orDefault is Python's `value or "default"`: an EMPTY string takes the
// default, which is also what an omitted --rationale/--reason produces.
func orDefault(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// today is TODAY (`bin/company-os:31`) — `datetime.date.today()`, the local
// calendar date.
//
// It is truncated to midnight so that AddDate below is pure CALENDAR
// arithmetic, as `date + timedelta(days=180)` is. Adding 180*24h to a wall
// clock instead lands on the previous day whenever the interval crosses a DST
// spring-forward, which is a one-day-wrong reviewDate written into an authored
// file twice a year.
func today() time.Time {
	n := time.Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}
