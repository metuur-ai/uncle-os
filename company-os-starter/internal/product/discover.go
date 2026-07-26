package product

// cmd_discover (`bin/company-os:409-464`) — the team-private half of the
// lifecycle.
//
// Both actions resolve the team directory FIRST (`:410`), before either branch
// runs, so an unknown team is exit 3 whichever action was asked for.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// DiscoverNew is `discover new` (`:411-424`), as records.
//
// The refusal at `:417` is exit 8 and names the ABSOLUTE path, because that is
// what the oracle's f-string interpolates; the success line names the
// workspace-relative one. The difference is not cosmetic drift — it is two
// different expressions in the source, and R-0.8 freezes both.
func DiscoverNew(ws *workspace.Workspace, team, title string) ([]model.GateResult, error) {
	tdir, err := ws.TeamDir(team)
	if err != nil {
		return nil, err
	}
	if title == "" {
		// argparse leaves `title` as None for `discover new --team t` with no
		// positional, and slugify(None) is an AttributeError traceback — exit 1,
		// nothing written. R-0.7a(l) replaces the traceback with a diagnostic;
		// the code is 2 because a positional that must be supplied for the
		// action to mean anything is a usage error, which is what the parser
		// would have said had `nargs="?"` been expressible per action.
		//
		// The check runs AFTER TeamDir because cmd_discover resolves the team
		// first (`:410`), so `discover new --team nope` is an unknown team on
		// both sides whether or not a title was supplied.
		return nil, model.Usagef("discover",
			"the following arguments are required: title")
	}
	bid := strconv.Itoa(today().Year()) + "-" + scaffold.Slugify(title)
	dir := filepath.Join(tdir, "product", "discovery", bid)
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", dir, err)
	}
	brief := filepath.Join(dir, "brief.md")
	if _, err := os.Stat(brief); err == nil {
		return nil, model.Errorf(model.ExitConflict, "%s already exists", brief)
	}
	tmpl, source, err := scaffold.ResolveTemplate(ws, scaffold.TemplateDiscoveryBrief, team, "")
	if err != nil {
		return nil, err
	}
	text, err := formatTemplate(tmpl, source, map[string]any{
		"bid":   bid,
		"title": title,
		"team":  team,
		"date":  today().Format(isoDate),
		"ds":    DiscoverySections,
	})
	if err != nil {
		return nil, err
	}
	if err := wrote(brief, os.WriteFile(brief, []byte(text), 0o666)); err != nil {
		return nil, err
	}

	rel := relTo(ws.Root, brief)
	created := model.Fields{"path": rel, "brief": bid, "team": team}
	tmplFields := model.Fields{"source": source}
	next := model.Fields{"team": team, "brief": bid}
	// The sentence wraps this command in prose ("fill …, then run: …"); R-3.6
	// wants the command alone, so it is spelled out rather than sliced out.
	next[model.FieldNext] = "company-os discover validate --team " + team + " " + bid
	return []model.GateResult{{
		Ordinal: 1, Slug: model.SlugDiscoverNew, Title: bid,
		Findings: []model.Finding{
			okFinding(model.CodeDiscoveryCreated, "", rel, created),
			okFinding(model.CodeTemplateSource, "", "", tmplFields),
			okFinding(model.CodeDiscoveryNext, "", "", next),
		},
	}}, nil
}

// DiscoverValidate is `discover validate` (`:425-463`), as records.
//
// The status flip is the LAST thing it does, so a brief that fails validation is
// left exactly as it was. The refusal exits 1 through the shared fail() helper
// (.devlocal/go-port/exit-code-map.md § H), which here means SevFail findings and
// the dispatcher's HasFailure mapping rather than an error return.
func DiscoverValidate(ws *workspace.Workspace, team, id string) ([]model.GateResult, error) {
	tdir, err := ws.TeamDir(team)
	if err != nil {
		return nil, err
	}
	if id == "" {
		// Same `nargs="?"` positional as DiscoverNew's title — `:2694` feeds it
		// into both dests — and the same oracle outcome: `tdir / None` is a
		// TypeError traceback at `:428`, exit 1, nothing written. Without this
		// guard filepath.Join silently drops the empty segment and the port
		// names `.../discovery/brief.md`, a path the user never asked about.
		// R-0.7a(l); argparse's own name for the positional is `title`.
		return nil, model.Usagef("discover",
			"the following arguments are required: title")
	}
	brief := filepath.Join(tdir, "product", "discovery", id, "brief.md")
	if _, err := os.Stat(brief); err != nil {
		return nil, model.Errorf(model.ExitWorkspace, "no brief at %s", brief)
	}
	meta, body, err := graph.ReadFrontmatter(brief)
	if err != nil {
		return nil, err
	}
	errs := CoreFieldErrors(meta)
	blocking, format := sectionIssues(body, DiscoverySections)
	errs = append(errs, blocking...)

	enforced, err := FormatChecksEnforced(ws, team)
	if err != nil {
		return nil, err
	}
	errs, warnings := applyFormatPolicy(errs, format, enforced)

	s := model.GateResult{Ordinal: 1, Slug: model.SlugDiscoverValidate, Title: id}
	// Python prints every warn() as it is computed, which is before the first
	// fail(); the record order is that print order.
	for _, w := range warnings {
		s.Findings = append(s.Findings, w.finding(model.SevWarn))
	}
	if len(errs) > 0 {
		for _, e := range errs {
			s.Findings = append(s.Findings, e.finding(model.SevFail))
		}
		return []model.GateResult{s}, nil
	}

	// `text.replace("status: draft", "status: validated", 1)` — a plain string
	// replace over the WHOLE file with a count of 1, not a frontmatter edit. A
	// brief already `status: validated` therefore validates again and is
	// rewritten byte-identically.
	raw, err := os.ReadFile(brief)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", brief, err)
	}
	updated := strings.Replace(string(raw), "status: draft", "status: validated", 1)
	if err := wrote(brief, os.WriteFile(brief, []byte(updated), 0o666)); err != nil {
		return nil, err
	}

	validated := model.Fields{"brief": id, "status": "validated", "team": team}
	next := model.Fields{"team": team, "brief": id}
	next[model.FieldNext] = Message(model.CodeDiscoveryValidateNext, next)
	s.Findings = append(s.Findings,
		okFinding(model.CodeDiscoveryValidated, "", relTo(ws.Root, brief), validated),
		okFinding(model.CodeDiscoveryValidateNext, "", "", next))
	return []model.GateResult{s}, nil
}
