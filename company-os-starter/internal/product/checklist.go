package product

// gather_prd_governance (`bin/company-os:551-570`), decomposed.
//
// The oracle built a markdown STRING and handed the same string to two unrelated
// consumers: `prd new` interpolated it into an artifact, `check ready|done`
// printed it. R-2.12 names it one of the four worst offenders, so it is split:
// Gather answers "which requirements apply", ChecklistMarkdown answers "what
// does that look like as markdown". Nothing else composes a checklist line.
//
// Gather has a side effect that is easy to miss and load-bearing for file-tree
// parity: when teams/<t>/generated/effective-governance.yaml is absent it calls
// governance.Resolve, which WRITES that file. So `check ready` against a team
// that never ran `governance resolve` creates the generated artifact. That is
// the oracle's behaviour (`:557-558`) and the differential harness compares the
// whole resulting tree.

import (
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// ChecklistItem is one line of the applicable-governance checklist, or the bold
// component header that opens a component's block.
type ChecklistItem struct {
	// Code is CodeChecklistComponent or CodeChecklistItem.
	Code string
	// Component is the component id every item belongs to.
	Component string
	// Scope is "company" for a baseline control, else the platform id. Empty on
	// a component header.
	Scope string
	// ID, Version and Level name the requirement.
	ID      string
	Version string
	Level   string
	// Deviated is `"deviation" in r` — the team has a recorded deviation from
	// this requirement.
	Deviated bool
}

// Fields is the item's --json payload. A component header carries only its
// component, which is what makes the two codes distinguishable without a type
// switch.
func (c ChecklistItem) Fields() model.Fields {
	if c.Code == model.CodeChecklistComponent {
		return model.Fields{"component": c.Component}
	}
	return model.Fields{
		"component": c.Component,
		"scope":     c.Scope,
		"id":        c.ID,
		"version":   c.Version,
		"level":     c.Level,
		"deviated":  c.Deviated,
	}
}

// Gather is gather_prd_governance (`:551-570`). It returns the items for the
// named components in the order the oracle emits them, plus the ids the team's
// effective governance does not name.
func Gather(ws *workspace.Workspace, team string, components []string) ([]ChecklistItem, []string, error) {
	eff, err := effectiveGovernance(ws, team)
	if err != nil {
		return nil, nil, err
	}
	all, ok := eff.Get("components").(pyMap)
	if !ok {
		// `eff["components"]` is a KeyError on an absent key and `.get` on a
		// non-mapping is an AttributeError; both are tracebacks that write
		// nothing (R-0.7a(j)).
		return nil, nil, model.Errorf(model.ExitArtifact,
			"teams/%s/generated/%s: no 'components' mapping", team, governance.GeneratedName)
	}

	var items []ChecklistItem
	var missing []string
	for _, cid := range components {
		comp, ok := all.Get(cid).(pyMap)
		if !ok {
			// `.get(cid)` returning None is the documented miss. A present
			// entry that is not a mapping raises one line later, so both land
			// here; only the miss is reported as missing.
			if all.Get(cid) == nil || yamlio.PyIsNone(all.Get(cid)) {
				missing = append(missing, cid)
				continue
			}
			return nil, nil, model.Errorf(model.ExitArtifact,
				"effective governance: component '%s' is not a mapping", cid)
		}
		items = append(items, ChecklistItem{Code: model.CodeChecklistComponent, Component: cid})

		reqs, ok := comp.Get("requirements").(pyMap)
		if !ok {
			return nil, nil, model.Errorf(model.ExitArtifact,
				"effective governance: component '%s' has no 'requirements' mapping", cid)
		}
		company, err := requirementList(reqs, "company", cid)
		if err != nil {
			return nil, nil, err
		}
		for _, c := range company {
			it, err := item(c, cid, "company")
			if err != nil {
				return nil, nil, err
			}
			items = append(items, it)
		}
		// `.items()` over an insertion-ordered dict: the platform blocks come
		// out in the order resolve wrote them, never sorted.
		platforms, ok := reqs.Get("platform").(pyMap)
		if !ok {
			return nil, nil, model.Errorf(model.ExitArtifact,
				"effective governance: component '%s' has no 'platform' mapping", cid)
		}
		for _, p := range platforms {
			list, ok := p.V.(pySeq)
			if !ok {
				return nil, nil, model.Errorf(model.ExitArtifact,
					"effective governance: %s/%s requirements must be a sequence", cid, p.K)
			}
			for _, r := range list {
				it, err := item(r, cid, p.K)
				if err != nil {
					return nil, nil, err
				}
				items = append(items, it)
			}
		}
	}
	return items, missing, nil
}

func requirementList(reqs pyMap, key, cid string) (pySeq, error) {
	v := reqs.Get(key)
	if v == nil {
		return nil, model.Errorf(model.ExitArtifact,
			"effective governance: component '%s' has no '%s' requirements", cid, key)
	}
	s, ok := v.(pySeq)
	if !ok {
		return nil, model.Errorf(model.ExitArtifact,
			"effective governance: %s '%s' requirements must be a sequence", cid, key)
	}
	return s, nil
}

// item is the shared body of `:561` and `:565`. The two Python lines differ only
// in the scope label and in the deviation marker, which is why one record type
// covers both.
func item(v pyVal, cid, scope string) (ChecklistItem, error) {
	m, ok := v.(pyMap)
	if !ok {
		return ChecklistItem{}, model.Errorf(model.ExitArtifact,
			"effective governance: %s/%s entries must be mappings", cid, scope)
	}
	return ChecklistItem{
		Code:      model.CodeChecklistItem,
		Component: cid,
		Scope:     scope,
		ID:        yamlio.PyString(m.Get("id")),
		Version:   yamlio.PyString(m.Get("version")),
		Level:     yamlio.PyString(m.Get("level")),
		// `"deviation" in r` is a KEY test, so a recorded `deviation: null`
		// still counts.
		Deviated: m.Get("deviation") != nil,
	}, nil
}

// effectiveGovernance is `:553-558`: read the generated file, and only re-derive
// when it is absent or empty. `load_yaml(path)` defaults to None, so an empty or
// falsy document also triggers the re-derivation — and the re-derivation writes.
func effectiveGovernance(ws *workspace.Workspace, team string) (pyMap, error) {
	tdir, err := ws.TeamDir(team)
	if err != nil {
		return nil, err
	}
	v, err := yamlio.PyLoadFile(filepath.Join(tdir, "generated", governance.GeneratedName))
	if err != nil {
		return nil, err
	}
	if !yamlio.PyFalsy(v) {
		m, ok := v.(pyMap)
		if !ok {
			return nil, model.Errorf(model.ExitArtifact,
				"teams/%s/generated/%s: expected a mapping at the document root",
				team, governance.GeneratedName)
		}
		return m, nil
	}
	res, err := governance.Resolve(ws, team)
	if err != nil {
		return nil, err
	}
	return res.Document, nil
}

// ChecklistMarkdown is the second half of gather_prd_governance: `"\n".join(lines)`
// over the same lines the oracle built inline (`:559-566`).
//
// The leading newline on a component header is the oracle's `f"\n**{cid}**"`, so
// the whole fragment starts with a blank line whenever anything resolved. Both
// consumers strip or guard that themselves, exactly as Python does.
func ChecklistMarkdown(items []ChecklistItem) string {
	lines := make([]string, 0, len(items))
	for _, it := range items {
		if it.Code == model.CodeChecklistComponent {
			lines = append(lines, "\n**"+it.Component+"**")
			continue
		}
		dev := ""
		if it.Deviated {
			dev = " *(team deviation applies)*"
		}
		lines = append(lines,
			"- [ ] "+it.Scope+": "+it.ID+" v"+it.Version+" ("+it.Level+")"+dev+" — evidence: ")
	}
	return strings.Join(lines, "\n")
}
