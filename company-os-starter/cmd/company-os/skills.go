package main

// `skills list` — the merged view across the four skill layers.
// `skills install` — write the canonical skills into company-os/skills/.

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdSkills is cmd_skills (bin/company-os:869-917) plus the `install` action.
// Like the other read-only views it formats nothing: internal/skills returns
// the record set and render.Skills turns it into bytes, so `out` goes unused.
func cmdSkills(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	if args.Action == "install" {
		return skillsInstall(ws)
	}
	return skills.List(ws)
}

// skillsInstall turns one InstallResult into the command's block.
//
// The sentences live here, as they do for every other mutating command: the
// next-step line wraps its command in prose, so the bare command rides in
// Fields[FieldNext] and the rendered line is spelled separately (R-3.6).
func skillsInstall(ws *workspace.Workspace) ([]model.GateResult, error) {
	res, err := skills.Install(ws, skills.Rebuild(rebuildGenerated))
	if err != nil {
		return nil, err
	}

	var findings []model.Finding
	var changed int
	for _, o := range res.Outcomes {
		fields := model.Fields{"skill": o.Name, "path": o.Rel, "version": o.To}
		var msg string
		sev := model.SevOK
		switch o.Code {
		case model.CodeSkillsInstalled:
			changed++
			msg = fmt.Sprintf("installed %s (v%s)", o.Rel, o.To)
		case model.CodeSkillsUpdated:
			changed++
			fields["from"] = o.From
			msg = fmt.Sprintf("updated %s (v%s -> v%s)", o.Rel, o.From, o.To)
		case model.CodeSkillsUnchanged:
			msg = fmt.Sprintf("unchanged %s (v%s)", o.Rel, o.To)
		case model.CodeSkillsLocallyNewer:
			fields["from"] = o.From
			sev = model.SevWarn
			msg = fmt.Sprintf("kept %s: installed v%s is newer than this binary's v%s — "+
				"upgrade the CLI rather than downgrading the skill", o.Rel, o.From, o.To)
		case model.CodeSkillsUnreadable:
			sev = model.SevWarn
			msg = fmt.Sprintf("kept %s: its frontmatter version could not be read", o.Rel)
		}
		findings = append(findings, model.Finding{
			Severity: sev, Code: o.Code, Subject: o.Name, Path: o.Rel,
			Message: msg, Fields: fields,
		})
	}

	next := "company-os skills list"
	findings = append(findings,
		line(model.CodeSkillsInstallSummary,
			fmt.Sprintf("%d skill(s) in %s, %d changed", len(res.Outcomes), res.Dir, changed),
			model.Fields{"count": len(res.Outcomes), "changed": changed, "path": res.Dir}),
		line(model.CodeSkillsInstallNext, "next: review what a session now sees: "+next,
			model.Fields{model.FieldNext: next}),
	)

	// The rebuild's lines precede the command's own, as they do for every other
	// scaffolding command (see generatedSection).
	return append(generatedSection(res.Generated), model.GateResult{
		Ordinal: 1, Slug: model.SlugSkillsInstall, Title: res.Dir, Findings: findings,
	}), nil
}
