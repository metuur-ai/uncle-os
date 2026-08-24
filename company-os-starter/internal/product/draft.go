package product

// Draft residency (R-1.1 … R-1.3). One place derives the draft location so
// promotion, abandonment, `today` and the integrity gate all address the same
// path instead of each re-joining it.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// DraftDir is `teams/<team>/product/change-records/draft` (R-1.1). The team
// must exist — an unknown team is the same NotFoundError every other product
// command raises. The draft directory itself need not exist; see DraftIDs.
func DraftDir(ws *workspace.Workspace, team string) (string, error) {
	tdir, err := ws.TeamDir(team)
	if err != nil {
		return "", err
	}
	return filepath.Join(tdir, "product", "change-records", "draft"), nil
}

// DraftPath is the artifact for one draft: `<draft-dir>/<draft-id>/prd.md`
// (R-1.2).
func DraftPath(draftDir, draftID string) string {
	return filepath.Join(draftDir, draftID, "prd.md")
}

// DraftIDs lists the draft ids under draftDir, sorted by name.
//
// R-1.3: an absent draft directory means the team has no drafts. That is a
// non-event, not a finding — every draft-less workspace would otherwise grow
// one — so the missing-directory case is indistinguishable here from an empty
// one. An entry counts as a draft only if it carries a prd.md (R-1.2); stray
// files and empty directories are skipped rather than reported, for the same
// reason.
func DraftIDs(draftDir string) []string {
	entries, err := os.ReadDir(draftDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if _, err := os.Stat(DraftPath(draftDir, e.Name())); err == nil {
			out = append(out, e.Name())
		}
	}
	return out
}

// draftPlaceholder is the value an unfilled process field carries. It is the
// token `prd validate` already reads as "not yet supplied" (`:627-628`), so a
// draft's unfilled fields fail the same test a change record's do rather than
// needing a second vocabulary.
const draftPlaceholder = "TODO"

// draftPlaceholderFields are the process fields a new draft leaves unfilled
// (R-2.4), in the order the frontmatter writes them.
//
// `team` is absent from the list because a draft always knows its team — it
// lives in that team's directory — which is the one place the draft's process
// contract is more determined than a change record's, not less.
var draftPlaceholderFields = []string{
	"title", "platform", "components", "governanceSnapshot", "decisionOwner",
}

// DraftNew is `prd new --team <t> --draft "<title>"` (R-2.1 … R-2.10).
//
// It is a separate entry point from PRDNew rather than a branch inside it
// because the two write different artifacts to different roots under different
// contracts; what they share — the id derivation, the template chain and the
// discovery copy-forward — they share as functions.
//
// platform is optional. Supplied, it fills both the routing field
// `promoteTo.platform` and the process field `platform` (R-2.5); omitted, both
// stay placeholders and the draft says so.
//
// rebuild is the same injection PRDComplete takes, and runs for the same reason
// (R-2.8): tags are derived, never authored (R-2.10), so a draft that returned
// before the derivation would leave `validate` red until someone remembered
// `graph build`.
func DraftNew(ws *workspace.Workspace, team, platform, title, fromDiscovery string,
	rebuild Rebuild) ([]model.GateResult, error) {

	if team == "" {
		// A draft is team-private, so unlike `prd new` there is no run of this
		// command without a team. --team is not argparse-required on the `prd`
		// sub-parser, so the diagnostic is raised here (R-0.7a(l)).
		return nil, model.Usagef("prd", "the following arguments are required: --team")
	}
	// PlatformDir is the same up-front existence check `prd new` performs, kept
	// here so an unknown platform is exit 3 before anything is written.
	if platform != "" {
		if _, err := ws.PlatformDir(platform); err != nil {
			return nil, err
		}
	}
	draftDir, err := DraftDir(ws, team)
	if err != nil {
		return nil, err
	}

	title, problem, metrics, discovery, err := carryDiscovery(ws, team, fromDiscovery, title)
	if err != nil {
		return nil, err
	}
	did := derivePRDID(title)
	if did == "" {
		return nil, model.Errorf(model.ExitUsage, "--title required (or --from-discovery)")
	}

	// R-2.6: the conflict is decided before the directory is made, so a refused
	// run leaves the workspace exactly as it found it.
	prd := DraftPath(draftDir, did)
	if _, err := os.Stat(prd); err == nil {
		return nil, model.Errorf(model.ExitConflict, "%s already exists", prd)
	}

	// R-2.3: the existing team -> platform -> company -> built-in probe order,
	// unchanged. A team that overrode templates/prd.md drafts with its own
	// template, which is the point of reusing the chain rather than adding one.
	tmpl, source, err := scaffold.ResolveTemplate(ws, scaffold.TemplatePRD, team, platform)
	if err != nil {
		return nil, err
	}
	target := platform
	if target == "" {
		target = draftPlaceholder
	}
	text, err := formatTemplate(tmpl, source, map[string]any{
		"pid":       did,
		"title":     title,
		"team":      team,
		"platform":  target,
		"date":      today().Format(isoDate),
		"discovery": discovery,
		"problem":   problem,
		"metrics":   metrics,
		"ps":        PRDSections,
		// A draft names no components yet (R-2.4), so the three
		// component-derived substitutions are empty and the governance
		// checklist resolves to nothing — the same text `prd new` writes when
		// --components is empty.
		"components":           "",
		"component_list":       "",
		"governance_checklist": "- [ ] none resolved",
	})
	if err != nil {
		return nil, err
	}
	body, err := draftBody(text, source)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(prd), 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact,
			"cannot create %s: %v", filepath.Dir(prd), err)
	}
	doc := draftFrontmatter(did, team, target) + body
	if err := wrote(prd, os.WriteFile(prd, []byte(doc), 0o666)); err != nil {
		return nil, err
	}

	rel := relTo(ws.Root, prd)
	out := []model.GateResult{{
		Ordinal: 1, Slug: model.SlugPRDNew, Title: did,
		Findings: []model.Finding{
			okFinding(model.CodePRDCreated, "", rel, model.Fields{"path": rel, "prd": did}),
			okFinding(model.CodeTemplateSource, "", "", model.Fields{"source": source}),
		},
	}}
	if rebuild != nil {
		derived, err := rebuild(ws)
		if err != nil {
			return nil, err
		}
		out = append(out, derived...)
	}
	// R-2.9: the closing line names the file, and names no command — see
	// model.CodePRDDraftNext.
	out = append(out, model.GateResult{
		Ordinal: len(out) + 1, Slug: model.SlugPRDNew, Title: did,
		Findings: []model.Finding{
			okFinding(model.CodePRDDraftNext, "", rel,
				model.Fields{"path": rel, "team": team, "draft": did}),
		},
	})
	return out, nil
}

// draftFrontmatter is the block every new draft opens with.
//
// It is written here rather than taken from the template because a draft's
// frontmatter is the system's contract, not the team's: `status: draft` is what
// keeps the draft out of the platform PRD gate, and the absent `tags:` is
// R-2.10 — tag facets are derived by `graph build`, never authored.
//
// There is deliberately no `level:` key (R-2.2). `level` already means rule tier
// across this workspace; the promotion target is `promoteTo.platform` and
// nothing else.
func draftFrontmatter(id, team, platform string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("type: prd\n")
	b.WriteString("id: " + id + "\n")
	b.WriteString("status: draft\n")
	b.WriteString("created: " + today().Format(isoDate) + "\n")
	b.WriteString("promoteTo:\n")
	b.WriteString("  platform: " + platform + "\n")
	b.WriteString("title: " + draftPlaceholder + "\n")
	b.WriteString("team: " + team + "\n")
	b.WriteString("platform: " + platform + "\n")
	b.WriteString("components: []\n")
	b.WriteString("governanceSnapshot: " + draftPlaceholder + "\n")
	b.WriteString("decisionOwner: " + draftPlaceholder + "\n")
	b.WriteString("---\n")
	return b.String()
}

// draftBody strips the resolved template's own frontmatter, keeping the prose
// (R-2.3). The template's block describes a `status: proposed` change record; a
// draft carries draftFrontmatter's instead, and two fences in one file parse as
// one fence plus a body that starts with a stray `---`.
//
// A template with no frontmatter contributes all of itself, which is what a team
// that overrode templates/prd.md with a bare outline would expect.
func draftBody(text, label string) (string, error) {
	doc, err := frontmatter.Parse([]byte(text))
	if err != nil {
		return "", model.Errorf(model.ExitArtifact, "%s: %v", label, err)
	}
	return string(doc.Body), nil
}
