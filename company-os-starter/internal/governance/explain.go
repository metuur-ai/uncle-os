package governance

import (
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/ids"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// UnresolvedError reports `governance explain` finding no generated governance
// naming the component (`bin/company-os:364-369`).
//
// It is exit 3 per .devlocal/go-port/exit-code-map.md § B, and the map is
// explicit that this is the WEAKEST 3 in that group: the one message covers
// both "the id does not exist anywhere" and "the id exists but
// `governance resolve` was never run for its team", and the Python code cannot
// tell them apart. The map's recommendation to split it into 3 and 5 is
// deliberately NOT taken — splitting changes an observable exit code, which is
// a behaviour change and not a port. The two causes are kept together, exactly
// as the oracle keeps them.
type UnresolvedError struct {
	// Component is the id as the message renders it.
	Component string
	// Suggestions are the up-to-three closest registered component ids.
	Suggestions []string

	coded *model.Error
}

func (e *UnresolvedError) Error() string { return e.coded.Error() }
func (e *UnresolvedError) Unwrap() error { return e.coded }

// Explain is cmd_governance's `explain` action (`bin/company-os:348-370`), as
// records.
//
// It emits one section per TEAM whose generated governance names the component:
// the oracle's loop sets `found` and keeps going rather than returning, so a
// component owned by two teams prints two blocks.
//
// component arrives already rendered as Python would render it, so an omitted
// positional is the literal "None" the f-string produces — see the caller.
func Explain(ws *workspace.Workspace, component string) ([]model.GateResult, error) {
	var sections []model.GateResult
	for _, tdir := range ws.AllTeams() {
		effPath := filepath.Join(tdir, "generated", GeneratedName)
		eff, err := loadOr(effPath, pyMap{})
		if err != nil {
			return nil, err
		}
		effMap, ok := eff.(pyMap)
		if !ok {
			// `eff.get("components", {})` raises AttributeError (R-0.7a(j)).
			return nil, model.Errorf(model.ExitArtifact,
				"%s: expected a mapping at the document root", relTo(ws.Root, effPath))
		}
		comps, ok := getDefault(effMap, "components", pyMap{}).(pyMap)
		if !ok {
			return nil, model.Errorf(model.ExitArtifact,
				"%s: 'components:' must be a mapping", relTo(ws.Root, effPath))
		}
		comp, ok := comps.Get(component).(pyMap)
		// `if not comp: continue` is truthiness — an EMPTY component entry is
		// skipped just as an absent one is, and `found` stays false.
		if !ok || yamlio.PyFalsy(comp) {
			continue
		}
		team, err := index(effMap, "team", relTo(ws.Root, effPath))
		if err != nil {
			return nil, err
		}
		s, err := explainSection(comp, component, yamlio.PyString(team),
			relTo(ws.Root, effPath), len(sections)+1)
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	if len(sections) > 0 {
		return sections, nil
	}

	// GPF-R-2.3: name the unknown id and suggest up to 3 closest registered
	// component ids. internal/ids owns the difflib transliteration; this is its
	// first live caller.
	suggestions, err := ids.Suggest(ws, component, "component")
	if err != nil {
		return nil, err
	}
	hint := ""
	if len(suggestions) > 0 {
		hint = "\n  did you mean: " + joinComma(suggestions)
	}
	coded, _ := model.Errorf(model.ExitWorkspace,
		"component '%s' not in any generated governance file; "+
			"run: company-os governance resolve --team <team>%s",
		component, hint).(*model.Error)
	return nil, &UnresolvedError{Component: component, Suggestions: suggestions, coded: coded}
}

// explainSection is the `:357-363` body for one team's generated governance.
func explainSection(comp pyMap, component, team, path string, ordinal int) (model.GateResult, error) {
	s := model.GateResult{Ordinal: ordinal, Slug: model.SlugExplain, Title: team}
	header := model.Fields{"component": component, "team": team}
	s.Findings = append(s.Findings, model.Finding{
		Severity: model.SevOK,
		Code:     model.CodeExplainComponent,
		Subject:  component,
		Path:     path,
		Message:  Message(model.CodeExplainComponent, header),
		Fields:   header,
	})

	// `comp["requirements"]["platform"]` and `comp["platforms"]` are INDEXED,
	// not .get(): a missing key is a KeyError and stops the command.
	reqs, err := index(comp, "requirements", path)
	if err != nil {
		return model.GateResult{}, err
	}
	reqsMap, ok := reqs.(pyMap)
	if !ok {
		return model.GateResult{}, model.Errorf(model.ExitArtifact,
			"%s: components.%s.requirements must be a mapping", path, component)
	}
	platform, err := index(reqsMap, "platform", path)
	if err != nil {
		return model.GateResult{}, err
	}
	platformMap, ok := platform.(pyMap)
	if !ok {
		return model.GateResult{}, model.Errorf(model.ExitArtifact,
			"%s: components.%s.requirements.platform must be a mapping", path, component)
	}
	owners, err := index(comp, "platforms", path)
	if err != nil {
		return model.GateResult{}, err
	}
	ownerSeq, ok := owners.(pySeq)
	if !ok {
		return model.GateResult{}, model.Errorf(model.ExitArtifact,
			"%s: components.%s.platforms must be a sequence", path, component)
	}

	for _, pair := range platformMap {
		pid := pair.K
		rel, err := relationshipFor(ownerSeq, pid, path, component)
		if err != nil {
			return model.GateResult{}, err
		}
		list, ok := pair.V.(pySeq)
		if !ok {
			return model.GateResult{}, model.Errorf(model.ExitArtifact,
				"%s: components.%s.requirements.platform.%s must be a sequence",
				path, component, pid)
		}
		for _, item := range list {
			r, ok := item.(pyMap)
			if !ok {
				return model.GateResult{}, model.Errorf(model.ExitArtifact,
					"%s: components.%s.requirements.platform.%s entries must be mappings",
					path, component, pid)
			}
			rid, err := index(r, "id", path)
			if err != nil {
				return model.GateResult{}, err
			}
			version, err := index(r, "version", path)
			if err != nil {
				return model.GateResult{}, err
			}
			level, err := index(r, "level", path)
			if err != nil {
				return model.GateResult{}, err
			}
			fields := model.Fields{
				"component":    component,
				"team":         team,
				"platform":     pid,
				"relationship": rel,
				"requirement":  yamlio.PyString(rid),
				"version":      yamlio.PyString(version),
				"level":        yamlio.PyString(level),
				"deviated":     r.Get("deviation") != nil,
			}
			s.Findings = append(s.Findings, model.Finding{
				Severity: model.SevOK,
				Code:     model.CodeExplainRequirement,
				Subject:  yamlio.PyString(rid),
				Path:     path,
				Message:  Message(model.CodeExplainRequirement, fields),
				Fields:   fields,
			})
		}
	}
	return s, nil
}

// relationshipFor is `next((p["relationship"] for p in comp["platforms"]
// if p["id"] == pid), "?")` (`:359`).
//
// The generator swallows nothing: `p["id"]` is evaluated for every element up
// to the match, so a malformed element BEFORE the matching one raises even
// though the answer would have been found. Only exhausting the sequence
// produces the "?" default, which is what a component carrying requirements for
// a platform it no longer belongs to renders.
func relationshipFor(owners pySeq, pid, path, component string) (string, error) {
	for _, o := range owners {
		m, ok := o.(pyMap)
		if !ok {
			return "", model.Errorf(model.ExitArtifact,
				"%s: components.%s.platforms entries must be mappings", path, component)
		}
		id, err := index(m, "id", path)
		if err != nil {
			return "", err
		}
		if yamlio.PyEqual(id, pyStr(pid)) {
			rel, err := index(m, "relationship", path)
			if err != nil {
				return "", err
			}
			return yamlio.PyString(rel), nil
		}
	}
	return "?", nil
}

func joinComma(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
