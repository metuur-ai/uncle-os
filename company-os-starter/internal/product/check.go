package product

// compose_checklist / cmd_check (`bin/company-os:714-733`) and validate gate 3
// (`:975-993`).
//
// `check` is the read-only composition of the two halves of a definition: the
// team's own baseline document, verbatim, and the governance checklist `prd new`
// would inject. It is a view, not a gate — a missing baseline and an unresolved
// component are both warnings and neither changes the exit code.
//
// Gate 3 lives here rather than in internal/validate because its subject is a
// change record. internal/validate composes it as one of its seven or eight
// GateResults; the shape (Gate + Message) is internal/governance's, and
// internal/skills' and internal/federation's, so the composition is uniform.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// GateSlug and GateTitle name validate's third gate.
const (
	GateSlug  = "active-prd-contracts"
	GateTitle = "active PRD contracts"
)

// contractFields is gate 3's four-field check (`:988-989`). It is deliberately
// SHORTER than `prd validate`'s six: the gate is a workspace sweep and does not
// demand a decisionOwner or a platform of a document whose location already
// states it.
var contractFields = []string{"title", "team", "components", "governanceSnapshot"}

// Check is cmd_check (`:731-733`) through compose_checklist (`:714-729`).
//
// kind is "ready" or "done" and reaches the output only as part of the baseline
// filename, which is why the two subcommands share one implementation.
func Check(ws *workspace.Workspace, team, components, kind string) ([]model.GateResult, error) {
	tdir, err := ws.TeamDir(team)
	if err != nil {
		return nil, err
	}
	ids := splitComponents(components)

	name := "definition-of-" + kind + ".md"
	baseline := filepath.Join(tdir, "standards", name)
	head := model.GateResult{Ordinal: 1, Slug: model.SlugCheckBaseline, Title: kind}
	head.Findings = append(head.Findings,
		okFinding(model.CodeCheckBaselineHeader, "", "", model.Fields{"file": name, "kind": kind}))
	if raw, err := os.ReadFile(baseline); err == nil {
		text := strings.TrimSpace(string(raw))
		head.Findings = append(head.Findings, okFinding(model.CodeCheckBaselineText, "",
			relTo(ws.Root, baseline), model.Fields{"text": text, "path": relTo(ws.Root, baseline)}))
	} else if os.IsNotExist(err) {
		head.Findings = append(head.Findings,
			warnFinding(model.CodeCheckBaselineMissing, "", model.Fields{"path": relTo(ws.Root, baseline)}))
	} else {
		// Python's read_text() on an existing but unreadable file is a
		// traceback; R-0.7a(e) makes it a diagnostic.
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", baseline, err)
	}

	items, missing, err := Gather(ws, team, ids)
	if err != nil {
		return nil, err
	}
	gov := model.GateResult{Ordinal: 2, Slug: model.SlugCheckGovernance, Title: kind}
	gov.Findings = append(gov.Findings, okFinding(model.CodeCheckGovernanceHeader, "", "",
		model.Fields{"components": ids}))
	// `print(checklist.strip() if checklist.strip() else "(none resolved)")` —
	// the fragment's leading blank line is stripped here and kept in `prd new`.
	markdown := strings.TrimSpace(ChecklistMarkdown(items))
	if markdown == "" {
		markdown = "(none resolved)"
	}
	gov.Findings = append(gov.Findings, okFinding(model.CodeCheckChecklist, "", "",
		model.Fields{"markdown": markdown, "items": items}))
	for _, cid := range missing {
		gov.Findings = append(gov.Findings,
			warnFinding(model.CodeCheckUnresolved, cid, model.Fields{"component": cid}))
	}
	return []model.GateResult{head, gov}, nil
}

// Gate is validate's gate 3 (`:975-993`): every active change record in every
// platform, checked against the four-field process contract.
//
// A platform with no change-records/active/ is skipped, as is a directory in it
// holding no prd.md — both are `continue`s in the oracle and neither produces a
// line, which is why a workspace with no active PRDs renders a gate header
// followed by nothing (golden-validate.txt:11-12).
func Gate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: GateSlug, Title: GateTitle}
	for _, pdir := range ws.AllPlatforms() {
		active := filepath.Join(pdir, "change-records", "active")
		entries, err := os.ReadDir(active)
		if err != nil {
			continue
		}
		for _, e := range entries {
			prd := filepath.Join(active, e.Name(), "prd.md")
			if _, err := os.Stat(prd); err != nil {
				continue
			}
			meta, _, err := graph.ReadFrontmatter(prd)
			if err != nil {
				return g, err
			}
			var missing []string
			for _, f := range contractFields {
				if !truthy(meta, f) {
					missing = append(missing, f)
				}
			}
			subject := filepath.Base(pdir) + "/" + e.Name()
			fields := model.Fields{
				"platform": filepath.Base(pdir),
				"prd":      e.Name(),
				"path":     relTo(ws.Root, prd),
			}
			if len(missing) > 0 {
				fields["missing"] = missing
				g.Findings = append(g.Findings, model.Finding{
					Severity: model.SevFail,
					Code:     model.CodePRDFrontmatterMissing,
					Subject:  subject,
					Path:     relTo(ws.Root, prd),
					Message:  Message(model.CodePRDFrontmatterMissing, fields),
					Fields:   fields,
				})
				continue
			}
			fields["missing"] = []string{}
			g.Findings = append(g.Findings, model.Finding{
				Severity: model.SevOK,
				Code:     model.CodePRDContractPresent,
				Subject:  subject,
				Path:     relTo(ws.Root, prd),
				Message:  Message(model.CodePRDContractPresent, fields),
				Fields:   fields,
			})
		}
	}
	return g, nil
}
