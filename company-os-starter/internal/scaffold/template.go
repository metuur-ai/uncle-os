package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/templates"
)

// The three built-in templates, one per name resolve_template accepts
// (bin/company-os:522-529). Two of them are constants rather than //go:embed
// because no file on disk holds their text: templates/discovery-brief.md and
// templates/prd.md are human reference copies with different placeholders, so
// embedding those files would change what `discover new` and `prd new` write.
// The third, reality-component, is the embedded file — see templates/embed.go.
//
// Both constants are byte-for-byte ports of the Python module strings; the
// literal backticks force the concatenation, since a Go raw string cannot
// contain one. TestBuiltinsMatchPythonModuleStrings pins them to the oracle.

// DiscoveryTemplate is DISCOVERY_TEMPLATE (bin/company-os:378-405) verbatim.
const DiscoveryTemplate = `---
type: discovery-brief
id: {bid}
title: {title}
status: draft
team: {team}
created: {date}
tags: [kind/discovery, team/{team}, status/draft]
---

# Discovery: {title}

## {ds[0]}
<!-- What evidence triggered this? Support tickets, metrics, stakeholder ask. -->

## {ds[1]}
<!-- We believe that <change> for <who> will achieve <outcome>. -->

## {ds[2]}
<!-- Measurable. How will we know the hypothesis is true/false? -->

## Affected components (initial guess)
<!-- component ids, if known -->

## Risks and open questions

## Decision
<!-- validated | invalidated, and why. Set by ` + "`" + `discover validate` + "`" + `. -->
`

// PRDTemplate is PRD_TEMPLATE (bin/company-os:470-506) verbatim.
const PRDTemplate = `---
type: prd
id: {pid}
title: {title}
status: proposed
team: {team}
platform: {platform}
components: [{components}]
governanceSnapshot: {date}
decisionOwner: TODO
created: {date}
fromDiscovery: {discovery}
tags: [kind/prd, platform/{platform}, team/{team}, status/proposed]
---

# PRD: {title}

## {ps[0]}
{problem}

## {ps[1]}
{metrics}

## {ps[2]}
<!-- What will be different in the Representation of Reality when done? -->

## Affected components
{component_list}

## Applicable governance (snapshot {date})
<!-- Injected by ` + "`" + `prd new` + "`" + `. Check items off as evidence is linked. -->
{governance_checklist}

## Out of scope

## Rollout and validation
`

// Template names accepted by ResolveTemplate. They are also the override
// filenames probed under each templates/ directory, as <name>.md.
const (
	TemplateDiscoveryBrief   = "discovery-brief"
	TemplatePRD              = "prd"
	TemplateRealityComponent = "reality-component"
)

// The provenance labels for the three built-ins, byte-identical to
// bin/company-os:523-529.
//
// These strings are printed to the user — `  template: built-in PRD_TEMPLATE`
// under `prd new` (:612), `discover new` (:419) and `reality new` (:2053) — so
// R-0.8 forbids rewording them, and the differential oracle ST-030 asserts the
// `prd new` line contains "PRD_TEMPLATE". They therefore keep naming Python
// identifiers that no longer exist in this codebase: the label is frozen output,
// not a description of the implementation. The Go constants holding the text
// stay idiomatically named (DiscoveryTemplate, PRDTemplate); only the printed
// label is a quotation of the oracle.
const (
	SourceBuiltinDiscovery = "built-in DISCOVERY_TEMPLATE"
	SourceBuiltinPRD       = "built-in PRD_TEMPLATE"
	SourceBuiltinReality   = "built-in templates/reality-component.md"
)

// ErrUnknownTemplate reports a template name with no override chain and no
// built-in, the (None, None) return of _builtin_template (bin/company-os:530).
// Every call site passes a literal from the TemplateXxx set, so this is a
// programming error rather than a user-facing one.
var ErrUnknownTemplate = errors.New("unknown template name")

// builtinTemplate returns the built-in body and provenance label for a template
// name — the last link in the resolution chain (bin/company-os:518-530).
func builtinTemplate(name string) (text, source string, ok bool) {
	switch name {
	case TemplateDiscoveryBrief:
		return DiscoveryTemplate, SourceBuiltinDiscovery, true
	case TemplatePRD:
		return PRDTemplate, SourceBuiltinPRD, true
	case TemplateRealityComponent:
		return templates.RealityComponent, SourceBuiltinReality, true
	}
	return "", "", false
}

// ResolveTemplate returns the body and provenance label for a scaffolding
// template, first-found-wins through the override chain then the built-in
// (bin/company-os:533-548, R-1.10).
//
// The probe order is teams/<t>/templates/<name>.md, then
// platforms/<p>/templates/<name>.md, then company-os/templates/<name>.md, each
// skipped when its scope was not supplied. Every candidate is workspace-relative
// and none is relative to the binary, which is what R-6.7 requires and what the
// embedded built-in makes possible.
//
// Pass "" for team or platform to omit that link, matching Python's `if team:` /
// `if platform:` guards.
func ResolveTemplate(ws *workspace.Workspace, name, team, platform string) (text, source string, err error) {
	for _, c := range templateCandidates(ws, name, team, platform) {
		b, err := os.ReadFile(c.path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			// Python's path.exists() guard is followed by an unguarded
			// read_text(); an override that exists but cannot be read is a
			// traceback there. Here it is an artifact fault with a diagnostic.
			return "", "", model.Errorf(model.ExitArtifact,
				"cannot read template override %s: %v", c.label, err)
		}
		return string(b), c.label, nil
	}
	text, source, ok := builtinTemplate(name)
	if !ok {
		return "", "", fmt.Errorf("%w: %q", ErrUnknownTemplate, name)
	}
	return text, source, nil
}

// templateCandidate is one override location: where to look, and the label to
// print when it wins. The label is always the POSIX workspace-relative form,
// never the absolute path that was probed.
type templateCandidate struct {
	path  string
	label string
}

func templateCandidates(ws *workspace.Workspace, name, team, platform string) []templateCandidate {
	file := name + ".md"
	var out []templateCandidate
	if team != "" {
		out = append(out, templateCandidate{
			path:  filepath.Join(ws.Teams, team, "templates", file),
			label: "teams/" + team + "/templates/" + file,
		})
	}
	if platform != "" {
		out = append(out, templateCandidate{
			path:  filepath.Join(ws.Platforms, platform, "templates", file),
			label: "platforms/" + platform + "/templates/" + file,
		})
	}
	return append(out, templateCandidate{
		path:  filepath.Join(ws.Company, "templates", file),
		label: "company-os/templates/" + file,
	})
}
