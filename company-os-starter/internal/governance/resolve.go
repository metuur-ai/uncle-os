package governance

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

type (
	pyStr = yamlio.PyStr
	pySeq = yamlio.PySeq
	pyMap = yamlio.PyMap
	pyVal = yamlio.PyValue
)

// GeneratedName is the derived artifact resolve_team_governance owns.
const GeneratedName = "effective-governance.yaml"

// deviationRejected is the sentence recorded into the generated file for a
// deviation aimed at a mandatory rule (`bin/company-os:318`).
//
// It is recorded and the resolve CONTINUES — this is not an exit site. The
// refusal surfaces later as a `validate` gate failure, and `deviation declare`
// itself validates nothing and always exits 0. See
// .devlocal/go-port/exit-code-map.md § "Code 5's third example does not exist
// as an exit site": enforcing the tier at declare time would be a behaviour
// change outside the port, because the tier is only knowable after resolution.
const deviationRejected = "mandatory rules cannot be deviated; use an exception"

// Resolved is resolve_team_governance's `(out, result)` return
// (`bin/company-os:266-330`).
//
// Document is the whole generated mapping rather than a decoded struct because
// gather_prd_governance (`:556`) reads it back by key and `today` reads the
// committed file; one shape serves all three and no field is dropped in
// transit.
type Resolved struct {
	// Path is the absolute path of teams/<t>/generated/effective-governance.yaml.
	Path string
	// Rel is that path relative to the workspace root, in POSIX form.
	Rel string
	// Document is the emitted mapping.
	Document pyMap
	// Written is false when the guard (R-0.7c) found the derived content
	// unchanged and skipped the write.
	Written bool

	// components is the per-component tally the resolve block prints, computed
	// from what was emitted rather than re-read out of Document.
	components []component
}

// component is one entry of Document["components"], kept alongside so the
// renderer's tallies are computed from what was emitted rather than re-read.
type component struct {
	id         string
	platforms  []string
	company    int
	platformNs int
	warning    string
}

// Resolve is resolve_team_governance (`bin/company-os:266-330`): ownership ->
// components -> platform relationships -> requirements (company baseline plus
// platform mandatory/default) merged with the team's deviations.
//
// Two things about it are load-bearing and easy to lose in a port:
//
//   - `ws.platform_dir(pid)` at `:302` is inside the relationship loop, so a
//     descriptor naming a platform that does not exist fails the whole command
//     with exit 3 and writes nothing.
//   - a deviation aimed at a mandatory rule is RECORDED, not refused; see
//     deviationRejected above.
func Resolve(ws *workspace.Workspace, teamID string) (Resolved, error) {
	tdir, err := ws.TeamDir(teamID)
	if err != nil {
		return Resolved{}, err
	}
	ownership, err := loadOr(filepath.Join(tdir, "ownership", "components.yaml"), pyMap{})
	if err != nil {
		return Resolved{}, err
	}
	deviations, err := loadOr(filepath.Join(tdir, "governance", "deviations.yaml"), pyMap{})
	if err != nil {
		return Resolved{}, err
	}
	devRules, err := deviationIndex(deviations)
	if err != nil {
		return Resolved{}, err
	}
	baseline, err := loadOr(filepath.Join(ws.Company, "standards", "company-baseline.yaml"), pyMap{})
	if err != nil {
		return Resolved{}, err
	}

	components := pyMap{}
	applied := pySeq{}
	var view []component

	owned, err := seqAt(ownership, "components", "ownership/components.yaml")
	if err != nil {
		return Resolved{}, err
	}
	for _, raw := range owned {
		comp, ok := raw.(pyMap)
		if !ok {
			// `comp["id"]` on a non-mapping raises TypeError before anything is
			// written (R-0.7a(j)).
			return Resolved{}, model.Errorf(model.ExitArtifact,
				"%s: ownership components entries must be mappings",
				relTo(ws.Root, filepath.Join(tdir, "ownership", "components.yaml")))
		}
		cidVal, err := index(comp, "id", "ownership component")
		if err != nil {
			return Resolved{}, err
		}
		cid := yamlio.PyString(cidVal)

		entry := pyMap{
			{K: "platforms", V: pySeq{}},
			{K: "requirements", V: pyMap{
				{K: "company", V: pySeq{}},
				{K: "platform", V: pyMap{}},
			}},
		}
		seen := component{id: cid}

		// The company baseline applies to everything the team is accountable
		// for, descriptor or not — it is appended before the descriptor lookup.
		controls, err := seqAt(baseline, "controls", "company-os/standards/company-baseline.yaml")
		if err != nil {
			return Resolved{}, err
		}
		company := pySeq{}
		for _, c := range controls {
			ctrl, ok := c.(pyMap)
			if !ok {
				return Resolved{}, model.Errorf(model.ExitArtifact,
					"company-os/standards/company-baseline.yaml: controls entries must be mappings")
			}
			id, err := index(ctrl, "id", "company baseline control")
			if err != nil {
				return Resolved{}, err
			}
			company = append(company, pyMap{
				{K: "id", V: id},
				{K: "level", V: getDefault(ctrl, "level", pyStr("mandatory"))},
				{K: "version", V: pyStr(yamlio.PyString(getDefault(ctrl, "version", pyStr("1.0"))))},
			})
		}
		entry = setNested(entry, "requirements", "company", company)
		seen.company = len(company)

		_, descPath, found := ws.FindComponent(cid)
		var desc pyMap
		haveDesc := false
		if found {
			loaded, err := yamlio.PyLoadFile(descPath)
			if err != nil {
				return Resolved{}, err
			}
			// find_component returns load_yaml(f), whose default is None, and
			// `or default` makes that TRUTHINESS: an empty or falsy descriptor
			// reaches the `desc is None` branch exactly as an absent one does
			// (R-1.7a).
			if !yamlio.PyFalsy(loaded) {
				m, ok := loaded.(pyMap)
				if !ok {
					// `desc.get(...)` on a non-mapping raises AttributeError.
					return Resolved{}, model.Errorf(model.ExitArtifact,
						"%s: expected a mapping at the document root", relTo(ws.Root, descPath))
				}
				desc, haveDesc = m, true
			}
		}
		if !haveDesc {
			seen.warning = "no component descriptor found in any platform catalog"
			entry = append(entry, yamlio.PyPair{K: "warning", V: pyStr(seen.warning)})
			components = components.Set(cid, entry)
			view = append(view, seen)
			continue
		}

		rels, err := seqAt(desc, "platformRelationships", relTo(ws.Root, descPath))
		if err != nil {
			return Resolved{}, err
		}
		for _, r := range rels {
			rel, ok := r.(pyMap)
			if !ok {
				return Resolved{}, model.Errorf(model.ExitArtifact,
					"%s: platformRelationships entries must be mappings", relTo(ws.Root, descPath))
			}
			platformVal, err := index(rel, "platform", "platformRelationships entry")
			if err != nil {
				return Resolved{}, err
			}
			platformStr, ok := platformVal.(pyStr)
			if !ok {
				// `rel["platform"].split("://")` raises AttributeError.
				return Resolved{}, model.Errorf(model.ExitArtifact,
					"%s: platformRelationships[].platform must be a string", relTo(ws.Root, descPath))
			}
			pid := localName(string(platformStr))
			relationship, err := index(rel, "relationship", "platformRelationships entry")
			if err != nil {
				return Resolved{}, err
			}

			platforms, _ := entry.Get("platforms").(pySeq)
			entry = entry.Set("platforms", append(platforms, pyMap{
				{K: "id", V: pyStr(pid)},
				{K: "relationship", V: relationship},
			}))

			pdir, err := ws.PlatformDir(pid)
			if err != nil {
				return Resolved{}, err
			}
			reqs, err := loadOr(filepath.Join(pdir, "governance", "requirements.yaml"), pyMap{})
			if err != nil {
				return Resolved{}, err
			}
			reqPath := relTo(ws.Root, filepath.Join(pdir, "governance", "requirements.yaml"))
			list, err := seqAt(reqs, "requirements", reqPath)
			if err != nil {
				return Resolved{}, err
			}

			applicable := pySeq{}
			for _, item := range list {
				r, ok := item.(pyMap)
				if !ok {
					return Resolved{}, model.Errorf(model.ExitArtifact,
						"%s: requirements entries must be mappings", reqPath)
				}
				keep, err := applies(r, relationship, desc, reqPath)
				if err != nil {
					return Resolved{}, err
				}
				if !keep {
					continue
				}
				rid, err := index(r, "id", "platform requirement")
				if err != nil {
					return Resolved{}, err
				}
				level := getDefault(r, "level", pyStr("default"))
				built := pyMap{
					{K: "id", V: rid},
					{K: "level", V: level},
					{K: "version", V: pyStr(yamlio.PyString(getDefault(r, "version", pyStr("1.0"))))},
					{K: "title", V: getDefault(r, "title", rid)},
				}
				ruleID := "platform-standard://" + pid + "/" + yamlio.PyString(rid)
				if d, ok := devRules[ruleID]; ok {
					if yamlio.PyEqual(level, pyStr("mandatory")) {
						built = append(built,
							yamlio.PyPair{K: "deviationRejected", V: pyStr(deviationRejected)})
					} else {
						rationale, ok := getDefault(d, "rationale", pyStr("")).(pyStr)
						if !ok {
							// `.strip()` on a non-str raises AttributeError.
							return Resolved{}, model.Errorf(model.ExitArtifact,
								"deviations.yaml: rationale for %s must be a string", ruleID)
						}
						built = append(built, yamlio.PyPair{K: "deviation", V: pyMap{
							{K: "status", V: getDefault(d, "status", yamlio.PyNull{})},
							{K: "rationale", V: pyStr(strings.TrimSpace(string(rationale)))},
						}})
						applied = append(applied, pyStr(ruleID))
					}
				}
				applicable = append(applicable, built)
			}
			if len(applicable) > 0 {
				entry = addPlatformRequirements(entry, pid, applicable)
			}
		}
		tally(&seen, entry)
		components = components.Set(cid, entry)
		view = append(view, seen)
	}

	result := pyMap{
		{K: "generatedAt", V: pyStr(now())},
		{K: "team", V: pyStr(teamID)},
		{K: "components", V: components},
		{K: "deviationsApplied", V: applied},
	}

	out := filepath.Join(tdir, "generated", GeneratedName)
	written, err := writeGuarded(out, result)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{
		Path: out, Rel: relTo(ws.Root, out), Document: result, Written: written,
		components: view,
	}, nil
}

// applies is the `appliesTo` filter at `:296-301`.
func applies(r pyMap, relationship pyVal, desc pyMap, reqPath string) (bool, error) {
	appliesTo := getDefault(r, "appliesTo", pyMap{})
	scope, ok := appliesTo.(pyMap)
	if !ok {
		// `applies.get(...)` on a non-mapping raises AttributeError.
		return false, model.Errorf(model.ExitArtifact, "%s: appliesTo must be a mapping", reqPath)
	}
	rels := scope.Get("relationships")
	if !yamlio.PyFalsy(rels) {
		in, err := contains(rels, relationship)
		if err != nil {
			return false, model.Errorf(model.ExitArtifact,
				"%s: appliesTo.relationships is not a container", reqPath)
		}
		if !in {
			return false, nil
		}
	}
	ctypes := scope.Get("componentTypes")
	if !yamlio.PyFalsy(ctypes) {
		metadata, ok := getDefault(desc, "metadata", pyMap{}).(pyMap)
		if !ok {
			return false, model.Errorf(model.ExitArtifact,
				"%s: component metadata must be a mapping", reqPath)
		}
		in, err := contains(ctypes, metadata.Get("type"))
		if err != nil {
			return false, model.Errorf(model.ExitArtifact,
				"%s: appliesTo.componentTypes is not a container", reqPath)
		}
		if !in {
			return false, nil
		}
	}
	return true, nil
}

// deviationIndex is `{d["rule"]: d for d in deviations.get("deviations", [])}`
// (`:270`). A repeated rule keeps the LAST entry, which is what makes
// `deviation declare` twice behave as one declaration.
func deviationIndex(deviations pyVal) (map[string]pyMap, error) {
	list, err := seqAt(deviations, "deviations", "governance/deviations.yaml")
	if err != nil {
		return nil, err
	}
	out := map[string]pyMap{}
	for _, item := range list {
		d, ok := item.(pyMap)
		if !ok {
			return nil, model.Errorf(model.ExitArtifact,
				"governance/deviations.yaml: deviations entries must be mappings")
		}
		rule, err := index(d, "rule", "deviation")
		if err != nil {
			return nil, err
		}
		switch rule.(type) {
		case pySeq, pyMap:
			// An unhashable dict key raises TypeError before anything is read.
			return nil, model.Errorf(model.ExitArtifact,
				"governance/deviations.yaml: a deviation rule must be a scalar")
		}
		// Only a str key can ever equal the `platform-standard://…` lookup, so a
		// non-str rule is indexed under its Python str() and simply never hits.
		out[yamlio.PyString(rule)] = d
	}
	return out, nil
}

// writeGuarded is dump_yaml(result, out) with R-0.7c's semantic guard.
//
// generatedAt is the reason the guard cannot be a plain structural compare: it
// carries NOW, so a fresh result NEVER equals the committed one and an
// unguarded writer rewrites the file on every invocation. That single changed
// line is enough to break R-0.10 — measured, `bin/company-os governance resolve`
// on a clean examples/workspace leaves exactly that one-line diff behind today.
//
// So the compare neutralizes the timestamp and asks the question the guard is
// actually for: did the DERIVED CONTENT change? When it did not, the file is
// left alone, which keeps the tree clean and — because the differential
// harness normalizes every `YYYY-MM-DDTHH:MM:SSZ` to <TS> on both sides — is
// indistinguishable from Python's rewrite.
//
// The compare form is PyDumpCanonical, the same key-sorted canonical dump gate 6
// and write_feature_indexes use, so two spellings of one document agree.
func writeGuarded(out string, result pyMap) (bool, error) {
	committed, err := yamlio.PyLoadFile(out)
	if err == nil {
		if prior, ok := committed.(pyMap); ok {
			same, cmpErr := sameDerivedContent(prior, result)
			if cmpErr == nil && same {
				return false, nil
			}
		}
	}
	// A committed file that is absent, malformed, or of another shape is simply
	// overwritten — Python never reads it at all.
	if err := yamlio.PyWriteFile(out, result); err != nil {
		return false, err
	}
	return true, nil
}

// sameDerivedContent compares two generated documents with generatedAt held
// equal, so the answer is about the derivation and not about the clock.
func sameDerivedContent(a, b pyMap) (bool, error) {
	left, err := yamlio.PyDumpCanonical(withoutTimestamp(a))
	if err != nil {
		return false, err
	}
	right, err := yamlio.PyDumpCanonical(withoutTimestamp(b))
	if err != nil {
		return false, err
	}
	return left == right, nil
}

// withoutTimestamp copies m with generatedAt blanked. PyMap.Set mutates in
// place for a key that exists, so the copy is not optional.
func withoutTimestamp(m pyMap) pyMap {
	out := make(pyMap, len(m))
	copy(out, m)
	if out.Get("generatedAt") != nil {
		out = out.Set("generatedAt", pyStr(""))
	}
	return out
}

// tally reads the resolve block's two numbers back out of the ENTRY rather than
// accumulating them as the loop runs, because the oracle computes them from the
// finished dict (`:340-341`) and the two answers differ.
//
// A descriptor that names the same platform twice appends TWICE to `platforms:`
// — a list — but its requirements land in `requirements.platform[pid]` once,
// because a dict assignment overwrites. So the platform list carries the
// duplicate and the requirement count does not. A running sum gets the second
// number wrong in exactly that case.
func tally(seen *component, entry pyMap) {
	if platforms, ok := entry.Get("platforms").(pySeq); ok {
		for _, p := range platforms {
			if m, ok := p.(pyMap); ok {
				seen.platforms = append(seen.platforms, yamlio.PyString(m.Get("id")))
			}
		}
	}
	reqs, _ := entry.Get("requirements").(pyMap)
	platform, _ := reqs.Get("platform").(pyMap)
	for _, pair := range platform {
		if list, ok := pair.V.(pySeq); ok {
			seen.platformNs += len(list)
		}
	}
}

// addPlatformRequirements writes entry["requirements"]["platform"][pid].
func addPlatformRequirements(entry pyMap, pid string, list pySeq) pyMap {
	reqs, _ := entry.Get("requirements").(pyMap)
	platform, _ := reqs.Get("platform").(pyMap)
	return setNested(entry, "requirements", "platform", platform.Set(pid, list))
}

// setNested writes entry[outer][inner] = v without aliasing the nested maps.
func setNested(entry pyMap, outer, inner string, v pyVal) pyMap {
	nested, _ := entry.Get(outer).(pyMap)
	copied := make(pyMap, len(nested))
	copy(copied, nested)
	return entry.Set(outer, copied.Set(inner, v))
}

// now is NOW (`bin/company-os:32`). Python computes it once at import; here it
// is computed per call, which is the same value for any single invocation.
func now() string { return time.Now().UTC().Format("2006-01-02T15:04:05Z") }
