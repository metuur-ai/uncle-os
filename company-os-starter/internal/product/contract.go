package product

// The process-level artifact contract: core_field_errors (`:128-145`), the
// team's format-enforcement opt-in (`:171-175`), and the section sweep both
// `validate` actions run.
//
// core_field_errors is not product-specific — validate gate 4 (`:1000-1002`)
// calls it over every graph document. It lives here because this is where its
// first consumer landed, exactly as pointer_errors lives in internal/graph next
// to the feature index that first needed it. Gate 4 imports CoreFieldErrors
// rather than growing a second copy; a second copy is how the doc-level gate and
// the two `validate` subcommands would start disagreeing about what a valid
// artifact is.
//
// It returns RECORDS, not sentences (R-2.12). Message() below is the only place
// the five sentences exist.

import (
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// LifecycleTypes is LIFECYCLE_TYPES (`bin/company-os:125`) — the doc types whose
// `status:` is part of the contract because something moves them through one.
var LifecycleTypes = map[string]bool{
	"discovery-brief": true,
	"prd":             true,
	"adr":             true,
	"outcome-review":  true,
}

// DiscoverySections is DISCOVERY_SECTIONS (`:376`): the SINGLE source shared by
// the discovery template (interpolated as `{ds[i]}`) and the `discover validate`
// heading check (GPF-R-4.3). Changing a name here is a two-file change — the
// built-in template in internal/scaffold has to change with it, which is what
// TestDiscoveryTemplateRendersEverySection exists to catch.
var DiscoverySections = []string{"Problem signal", "Hypothesis", "Success criteria"}

// PRDSections is PRD_SECTIONS (`:468`), the same contract for PRDs.
var PRDSections = []string{"Problem statement", "Success metrics", "Proposed change"}

// Issue is one contract violation, kept structured until Message runs.
type Issue struct {
	Code   string
	Fields model.Fields
}

func (i Issue) finding(sev model.Severity) model.Finding {
	return model.Finding{
		Severity: sev,
		Code:     i.Code,
		Message:  Message(i.Code, i.Fields),
		Fields:   i.Fields,
	}
}

// CoreFieldErrors is core_field_errors (`bin/company-os:128-145`): the
// process-level contract only, never the body format.
//
// Order is the oracle's — type, identity, status, updated, role — because the
// caller renders them in list order and the golden fixes that order.
func CoreFieldErrors(meta yamlio.PyMap) []Issue {
	var out []Issue
	if !truthy(meta, "type") {
		out = append(out, Issue{Code: model.CodeCoreTypeMissing, Fields: model.Fields{}})
	}
	if !truthy(meta, "id") && !truthy(meta, "prd") {
		out = append(out, Issue{Code: model.CodeCoreIdentityMissing, Fields: model.Fields{}})
	}
	// `t = meta.get("type")` is compared against a set of str, so only a str can
	// match; a null or a list is simply not in LIFECYCLE_TYPES.
	t, _ := meta.Get("type").(yamlio.PyStr)
	if LifecycleTypes[string(t)] && !truthy(meta, "status") {
		out = append(out, Issue{
			Code:   model.CodeCoreStatusMissing,
			Fields: model.Fields{"type": string(t)},
		})
	}
	if t == "component-reality" && !truthy(meta, "updated") {
		out = append(out, Issue{Code: model.CodeCoreUpdatedMissing, Fields: model.Fields{}})
	}
	if t == "onboarding-guide" && !truthy(meta, "role") {
		out = append(out, Issue{Code: model.CodeCoreRoleMissing, Fields: model.Fields{}})
	}
	return out
}

// FormatChecksEnforced is format_checks_enforced (`:171-175`): section-structure
// checks are team-local GUIDANCE by default, and a team opts into blocking
// enforcement through standards/doc-formats.yaml.
//
// The team id is interpolated into a path with `str(team)`, so an empty id
// probes teams/standards/doc-formats.yaml — a file that does not exist — and the
// answer is "not enforced". That is Python's answer too and no call site
// depends on it, so it is left alone rather than guarded.
func FormatChecksEnforced(ws *workspace.Workspace, team string) (bool, error) {
	path := filepath.Join(ws.Teams, team, "standards", "doc-formats.yaml")
	v, err := yamlio.PyLoadFile(path)
	if err != nil {
		return false, err
	}
	// `load_yaml(path, {}) or {}` then `.get("enforce")`: a non-mapping document
	// reaches `.get` and raises AttributeError, writing nothing (R-0.7a(j)).
	if yamlio.PyFalsy(v) {
		return false, nil
	}
	cfg, ok := v.(pyMap)
	if !ok {
		return false, model.Errorf(model.ExitArtifact,
			"%s: expected a mapping at the document root", path)
	}
	return !yamlio.PyFalsy(cfg.Get("enforce")), nil
}

// sectionIssues sweeps the required headings of one document
// (`:437-447` for a brief, `:637-646` for a PRD — the same loop twice).
//
// It returns two lists, and the split is the whole point of GPF-R-4.4: a MISSING
// heading is an artifact-contract failure and always blocks, whatever template
// produced the document, while an EMPTY section is format guidance the team may
// opt into enforcing. Outputs are validated even when produced from a custom
// override template.
func sectionIssues(body []byte, sections []string) (blocking, format []Issue) {
	for _, s := range sections {
		content, found := sectionContent(body, s)
		if !found {
			blocking = append(blocking, Issue{
				Code:   model.CodeSectionHeadingMissing,
				Fields: model.Fields{"section": s},
			})
			continue
		}
		if content == "" {
			format = append(format, Issue{
				Code:   model.CodeSectionEmpty,
				Fields: model.Fields{"section": s},
			})
		}
	}
	return blocking, format
}

// applyFormatPolicy folds the format issues into the blocking list or turns them
// into warnings, per the team's opt-in (`:448-453`, `:647-652`).
//
// The `enforced` field on each format issue is what makes Message able to render
// both sentences off one code: the blocking form is the bare "section 'X' is
// empty", the guidance form appends the opt-in pointer.
func applyFormatPolicy(errs, format []Issue, enforced bool) (blocking, warnings []Issue) {
	for _, f := range format {
		f.Fields = model.Fields{"section": f.Fields.Str("section"), "enforced": enforced}
		if enforced {
			errs = append(errs, f)
		} else {
			warnings = append(warnings, f)
		}
	}
	return errs, warnings
}
