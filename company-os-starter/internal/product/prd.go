package product

// cmd_prd (`bin/company-os:573-711`) — the platform-visible half of the
// lifecycle, and the enforcement point for invariant #4.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// processFields is the six-field process contract `prd validate` enforces
// (`:627-628`), in the order it reports them.
var processFields = []string{
	"title", "team", "platform", "components", "governanceSnapshot", "decisionOwner",
}

// outcomeDays is the outcome-review interval (`:702`).
const outcomeDays = 90

// Rebuild re-derives the generated artifacts, returning the records `graph
// build` would have produced for them.
//
// It is injected rather than imported for the reason internal/scaffold gives for
// its own Rebuild: keeping `product -> graph` out of the import graph is what
// keeps the seam a seam. internal/graph's Rebuild satisfies it as written.
type Rebuild func(*workspace.Workspace) ([]model.GateResult, error)

// PRDNew is `prd new` (`:574-623`), as records.
//
// The `missing` warnings are printed by the oracle BEFORE the change-record
// directory is created, so a run that then refuses at `:610` has already emitted
// them. That is why this returns its accumulated sections alongside the error
// rather than discarding them.
func PRDNew(ws *workspace.Workspace, team, platform, components, title, fromDiscovery string) (
	[]model.GateResult, error) {

	pdir, err := ws.PlatformDir(platform)
	if err != nil {
		return nil, err
	}
	ids := splitComponents(components)

	pid := ""
	if title != "" {
		pid = strconv.Itoa(today().Year()) + "-" + scaffold.Slugify(title)
	}
	problem, metrics, discovery := "<!-- Why now? -->", "<!-- Measurable. -->", "none"

	if fromDiscovery != "" {
		if team == "" {
			// `ws.team_dir(None)` is a TypeError traceback that writes nothing.
			// --team is not marked required on the `prd` sub-parser, so this is
			// argparse's own diagnostic for the requirement it could not
			// express (R-0.7a(l)).
			return nil, model.Usagef("prd",
				"the following arguments are required: --team")
		}
		tdir, err := ws.TeamDir(team)
		if err != nil {
			return nil, err
		}
		brief := filepath.Join(tdir, "product", "discovery", fromDiscovery, "brief.md")
		if _, err := os.Stat(brief); err != nil {
			return nil, model.Errorf(model.ExitWorkspace, "discovery brief not found: %s", brief)
		}
		meta, body, err := graph.ReadFrontmatter(brief)
		if err != nil {
			return nil, err
		}
		if status := strOf(meta, "status"); status != "validated" {
			// Exit 5: the brief exists and is well-formed, the WORKFLOW
			// precondition is unmet (.devlocal/go-port/exit-code-map.md:53).
			return nil, model.Errorf(model.ExitPrecondition,
				"discovery '%s' is '%s', not 'validated'. Run discover validate first.",
				fromDiscovery, status)
		}
		discovery = fromDiscovery
		if title == "" {
			// `meta["title"]` is a KeyError on a brief with no title.
			t := meta.Get("title")
			if t == nil {
				return nil, model.Errorf(model.ExitArtifact,
					"%s: missing required key 'title'", brief)
			}
			title = yamlio.PyString(t)
			pid = strconv.Itoa(today().Year()) + "-" + scaffold.Slugify(title)
		}
		if c, ok := sectionContentRaw(body, "Problem signal"); ok && c != "" {
			problem = c
		}
		if c, ok := sectionContentRaw(body, "Success criteria"); ok && c != "" {
			metrics = c
		}
	}

	if pid == "" {
		// Ruling G of the exit-code map: a hand-rolled conditional-requirement
		// check argparse cannot express. 1 today, 2 under the contract.
		return nil, model.Errorf(model.ExitUsage, "--title required (or --from-discovery)")
	}

	items, missing, err := Gather(ws, team, ids)
	if err != nil {
		return nil, err
	}
	// The warnings are printed by the oracle before the change record is
	// created, so a run that then refuses at `:610` has already emitted them.
	// `section` rebuilds the GateResult at every return rather than aliasing one
	// slice header, which is how a later append would silently fail to reach an
	// already-returned copy.
	var findings []model.Finding
	for _, cid := range missing {
		f := model.Fields{"component": cid, "team": team}
		findings = append(findings, warnFinding(model.CodePRDGovernanceUnresolved, cid, f))
	}
	section := func() []model.GateResult {
		return []model.GateResult{{
			Ordinal: 1, Slug: model.SlugPRDNew, Title: pid, Findings: findings,
		}}
	}

	dir := filepath.Join(pdir, "change-records", "active", pid)
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return section(), model.Errorf(model.ExitArtifact, "cannot create %s: %v", dir, err)
	}
	prd := filepath.Join(dir, "prd.md")
	if _, err := os.Stat(prd); err == nil {
		return section(), model.Errorf(model.ExitConflict, "%s already exists", prd)
	}
	tmpl, source, err := scaffold.ResolveTemplate(ws, scaffold.TemplatePRD, team, platform)
	if err != nil {
		return section(), err
	}
	checklist := ChecklistMarkdown(items)
	if checklist == "" {
		// `checklist or "- [ ] none resolved"` — an empty join is falsy.
		checklist = "- [ ] none resolved"
	}
	componentList := make([]string, 0, len(ids))
	for _, c := range ids {
		componentList = append(componentList, "- `"+c+"`")
	}
	text, err := formatTemplate(tmpl, source, map[string]any{
		"pid":                  pid,
		"title":                title,
		"team":                 team,
		"platform":             platform,
		"components":           strings.Join(ids, ", "),
		"date":                 today().Format(isoDate),
		"discovery":            discovery,
		"problem":              problem,
		"metrics":              metrics,
		"ps":                   PRDSections,
		"component_list":       strings.Join(componentList, "\n"),
		"governance_checklist": checklist,
	})
	if err != nil {
		return section(), err
	}
	if err := wrote(prd, os.WriteFile(prd, []byte(text), 0o666)); err != nil {
		return section(), err
	}

	rel := relTo(ws.Root, prd)
	findings = append(findings,
		okFinding(model.CodePRDCreated, "", rel, model.Fields{"path": rel, "prd": pid}),
		okFinding(model.CodeTemplateSource, "", "", model.Fields{"source": source}))
	// GPF-R-1.8: any component still missing its reality doc gets the exact
	// scaffold command it needs, so the chain stays unbroken through `complete`.
	for _, cid := range ids {
		if _, err := os.Stat(filepath.Join(pdir, "reality", "components", cid+".md")); err == nil {
			continue
		}
		f := model.Fields{"component": cid, "platform": platform}
		findings = append(findings, okFinding(model.CodePRDRealityNote, cid, "", f))
	}
	findings = append(findings, okFinding(model.CodePRDNext, "", "",
		model.Fields{"platform": platform, "prd": pid,
			model.FieldNext: "company-os prd validate --platform " + platform + " " + pid}))
	return section(), nil
}

// sectionContentRaw is `grab()` (`:594-596`): the section body stripped, but
// with its HTML comments LEFT IN — unlike the validate sweep, which removes them
// before testing for emptiness. A brief whose Problem signal is only its
// scaffolded hint therefore carries that hint forward into the PRD.
func sectionContentRaw(body []byte, section string) (string, bool) {
	raw, ok := sectionBody(body, section)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(raw), true
}

// requirePRDID rejects `prd validate` / `prd complete` with no `id`.
//
// `id` is declared `nargs="?"` (`bin/company-os:2707`) because `prd new` derives
// it from --title, so argparse lets the other two actions through with None and
// `pdir / "change-records" / "active" / None` is a TypeError traceback at `:633`
// and `:673` — exit 1, nothing written. R-0.7a(l) replaces it with a diagnostic
// at exit 2. Without the guard filepath.Join drops the empty segment and the
// port reports `no active PRD at …/active/prd.md`, naming a path that does not
// correspond to anything the user asked for.
//
// Both callers run this AFTER PlatformDir, because the oracle evaluates
// `ws.platform_dir(args.platform)` first on the same expression: an unknown
// platform is exit 3 on both sides whether or not an id was supplied.
func requirePRDID(id string) error {
	if id != "" {
		return nil
	}
	return model.Usagef("prd", "the following arguments are required: id")
}

// PRDValidate is `prd validate` (`:625-669`), as records. Its refusal exits 1
// through the shared fail() helper (exit-code map § H).
func PRDValidate(ws *workspace.Workspace, platform, id string) ([]model.GateResult, error) {
	pdir, err := ws.PlatformDir(platform)
	if err != nil {
		return nil, err
	}
	if err := requirePRDID(id); err != nil {
		return nil, err
	}
	prd := filepath.Join(pdir, "change-records", "active", id, "prd.md")
	if _, err := os.Stat(prd); err != nil {
		return nil, model.Errorf(model.ExitWorkspace, "no active PRD at %s", prd)
	}
	meta, body, err := graph.ReadFrontmatter(prd)
	if err != nil {
		return nil, err
	}
	errs := CoreFieldErrors(meta)
	for _, field := range processFields {
		if truthy(meta, field) && !yamlio.PyEqual(meta.Get(field), yamlio.PyStr("TODO")) {
			continue
		}
		errs = append(errs, Issue{
			Code:   model.CodePRDProcessField,
			Fields: model.Fields{"field": field},
		})
	}
	blocking, format := sectionIssues(body, PRDSections)
	errs = append(errs, blocking...)

	// `if meta.get("team") and format_checks_enforced(ws, meta["team"])`: a PRD
	// with no team never enforces, whatever its team's config says.
	enforced := false
	if truthy(meta, "team") {
		enforced, err = FormatChecksEnforced(ws, yamlio.PyString(meta.Get("team")))
		if err != nil {
			return nil, err
		}
	}
	errs, warnings := applyFormatPolicy(errs, format, enforced)

	s := model.GateResult{Ordinal: 1, Slug: model.SlugPRDValidate, Title: id}
	for _, w := range warnings {
		s.Findings = append(s.Findings, w.finding(model.SevWarn))
	}
	if len(errs) > 0 {
		for _, e := range errs {
			s.Findings = append(s.Findings, e.finding(model.SevFail))
		}
		return []model.GateResult{s}, nil
	}
	s.Findings = append(s.Findings,
		okFinding(model.CodePRDContractOK, "", relTo(ws.Root, prd), model.Fields{"prd": id}),
		okFinding(model.CodePRDValidateNext, "", "",
			model.Fields{"platform": platform, "prd": id,
				model.FieldNext: "company-os prd complete --platform " + platform + " " + id}))
	return []model.GateResult{s}, nil
}

// PRDComplete is `prd complete` (`:671-711`) — invariant #4.
//
// On refusal it returns BOTH the records (the done-check block, which the oracle
// prints to stdout) and ErrDoneCheck, a quiet exit-5 error: the caller renders
// the records and the dispatcher adds no `error: …` line, because the oracle
// writes nothing to stderr here.
//
// On success the filesystem effects are, in order: move the change record into
// archive/prds/, rewrite its status, write outcome.md, append log.md. Only then
// does the derived output run, and only after THAT does the next-step line
// print — the one command whose own output brackets the rebuild rather than
// following it.
func PRDComplete(ws *workspace.Workspace, platform, id string, force bool,
	rebuild Rebuild) ([]model.GateResult, error) {

	pdir, err := ws.PlatformDir(platform)
	if err != nil {
		return nil, err
	}
	if err := requirePRDID(id); err != nil {
		return nil, err
	}
	src := filepath.Join(pdir, "change-records", "active", id)
	prd := filepath.Join(src, "prd.md")
	if _, err := os.Stat(prd); err != nil {
		return nil, model.Errorf(model.ExitWorkspace, "no active PRD at %s", prd)
	}
	meta, body, err := graph.ReadFrontmatter(prd)
	if err != nil {
		return nil, err
	}

	problems, missingReality, err := doneCheck(ws, pdir, prd, meta, body)
	if err != nil {
		return nil, err
	}
	if len(problems) > 0 && !force {
		s := model.GateResult{Ordinal: 1, Slug: model.SlugPRDComplete, Title: id}
		// The banner is a plain print(), not a fail(): the [FAIL] lines below
		// carry the failure, and counting the banner as one would inflate every
		// --json problem tally by one.
		s.Findings = append(s.Findings,
			okFinding(model.CodeDoneCheckHeader, "", "", model.Fields{"prd": id}))
		for _, p := range problems {
			s.Findings = append(s.Findings, p.finding(model.SevFail))
		}
		for _, cid := range missingReality {
			f := model.Fields{"component": cid, "platform": platform}
			s.Findings = append(s.Findings, okFinding(model.CodeDoneFix, cid, "", f))
		}
		return []model.GateResult{s}, ErrDoneCheck
	}

	dst := filepath.Join(pdir, "archive", "prds", id)
	if err := os.MkdirAll(filepath.Dir(dst), 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", filepath.Dir(dst), err)
	}
	if err := shutilMove(src, dst); err != nil {
		return nil, err
	}

	archived := filepath.Join(dst, "prd.md")
	raw, err := os.ReadFile(archived)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", archived, err)
	}
	// Two plain replaces with a count of 1, over the whole file — not a
	// frontmatter edit. `in-delivery` is tried second and unconditionally, so a
	// PRD that was `proposed` has already lost its match by then.
	text := strings.Replace(string(raw), "status: proposed", "status: completed", 1)
	text = strings.Replace(text, "status: in-delivery", "status: completed", 1)
	if err := wrote(archived, os.WriteFile(archived, []byte(text), 0o666)); err != nil {
		return nil, err
	}

	due := today().AddDate(0, 0, outcomeDays).Format(isoDate)
	outcome := filepath.Join(dst, "outcome.md")
	if err := wrote(outcome, os.WriteFile(outcome, []byte(outcomeDoc(id, due, strOf(meta, "title"))), 0o666)); err != nil {
		return nil, err
	}

	log := filepath.Join(pdir, "log.md")
	if err := appendLog(log, today().Format(isoDate), id, due); err != nil {
		return nil, err
	}

	relDst, relLog := relTo(ws.Root, dst), relTo(ws.Root, log)
	s := model.GateResult{Ordinal: 1, Slug: model.SlugPRDComplete, Title: id}
	s.Findings = append(s.Findings,
		okFinding(model.CodePRDArchived, "", relDst, model.Fields{"path": relDst, "prd": id}),
		okFinding(model.CodeOutcomeScheduled, "", "", model.Fields{"due": due, "prd": id}),
		okFinding(model.CodeLogAppended, "", relLog, model.Fields{"path": relLog}))
	out := []model.GateResult{s}

	if rebuild != nil {
		derived, err := rebuild(ws)
		if err != nil {
			return nil, err
		}
		out = append(out, derived...)
	}
	out = append(out, model.GateResult{
		Ordinal: len(out) + 1, Slug: model.SlugPRDComplete, Title: id,
		Findings: []model.Finding{
			okFinding(model.CodePRDCompleteNext, "", "",
				model.Fields{model.FieldNext: "company-os validate"}),
		},
	})
	return out, nil
}

// doneCheck is `:678-698`: the unchecked-checklist tally plus one verdict per
// component named in the PRD's frontmatter.
func doneCheck(ws *workspace.Workspace, pdir, prd string, meta pyMap, body []byte) (
	problems []Issue, missingReality []string, err error) {

	if n := uncheckedItems(body); n > 0 {
		problems = append(problems, Issue{
			Code:   model.CodeDoneChecklistUnchecked,
			Fields: model.Fields{"count": n},
		})
	}

	created, createdOK := parseDate(strOf(meta, "created"))
	if !createdOK {
		// R-1.14: the oracle would silently order this string against another
		// string. Naming it is the whole point of the fix.
		problems = append(problems, Issue{
			Code: model.CodeDoneRealityDateInvalid,
			Fields: model.Fields{
				"path":  relTo(ws.Root, prd),
				"field": "created",
				"value": strOf(meta, "created"),
			},
		})
	}

	components, err := componentIDs(meta)
	if err != nil {
		return nil, nil, err
	}
	for _, cid := range components {
		reality := filepath.Join(pdir, "reality", "components", cid+".md")
		if _, err := os.Stat(reality); err != nil {
			rel := relTo(ws.Root, reality)
			problems = append(problems, Issue{
				Code:   model.CodeDoneRealityMissing,
				Fields: model.Fields{"component": cid, "path": rel},
			})
			missingReality = append(missingReality, cid)
			continue
		}
		rMeta, _, err := graph.ReadFrontmatter(reality)
		if err != nil {
			return nil, nil, err
		}
		updatedText := strOf(rMeta, "updated")
		updated, updatedOK := parseDate(updatedText)
		switch {
		case !updatedOK:
			problems = append(problems, Issue{
				Code: model.CodeDoneRealityDateInvalid,
				Fields: model.Fields{
					"path":      relTo(ws.Root, reality),
					"field":     "updated",
					"value":     updatedText,
					"component": cid,
				},
			})
		case createdOK && updated.Before(created):
			problems = append(problems, Issue{
				Code:   model.CodeDoneRealityStale,
				Fields: model.Fields{"component": cid, "updated": updatedText},
			})
		}
	}
	return problems, missingReality, nil
}

// parseDate is the R-1.14 / OKF v0.2 Phase 0 fix, and the ONE place in this port
// that deliberately does not reproduce the oracle (sanctioned as R-0.7a(d)).
//
// `bin/company-os:679-682` compares the two dates as raw STRINGS:
//
//	if str(r_meta.get("updated", "")) < str(meta.get("created", "")):
//
// That is right for well-formed ISO dates by lexical accident, and silently
// wrong for everything else — `18/07/2026` sorts before every `2` and passes the
// gate, an absent value becomes "" and fails it with a sentence claiming the doc
// is stale, and a value PyYAML parsed into a datetime.date compares as its
// repr. The gate that enforces invariant #4 cannot be allowed to answer by
// accident.
//
// Two spellings parse: a bare calendar date, and a full RFC 3339 timestamp for
// the case where the author wrote a time. Anything else is refused BY NAME, so a
// malformed date is a visible done-check problem rather than an invisible pass.
func parseDate(s string) (time.Time, bool) {
	if t, err := time.Parse(isoDate, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
	}
	return time.Time{}, false
}

// uncheckedItems is `len(re.findall(r"- \[ \] .*", body))` (`:678`).
func uncheckedItems(body []byte) int {
	return len(uncheckedRe.FindAll(body, -1))
}

// componentIDs is `meta.get("components", [])` as Python iterates it.
//
// A list yields its elements. A STRING yields its characters, which is not
// defensive padding: `components: svc-alpha` is valid YAML and the oracle would
// look for `s.md`, `v.md`, `c.md`… Reproducing that keeps a mis-authored PRD
// failing the same way on both sides instead of passing on one.
func componentIDs(meta pyMap) ([]string, error) {
	v := meta.Get("components")
	switch t := v.(type) {
	case nil:
		return nil, nil
	case yamlio.PyNull:
		// `.get(key, [])` returns the explicit None, and `for cid in None`
		// raises TypeError (R-0.7a(j)).
		return nil, model.Errorf(model.ExitArtifact, "prd: 'components' is null and not iterable")
	case pySeq:
		out := make([]string, 0, len(t))
		for _, e := range t {
			out = append(out, yamlio.PyString(e))
		}
		return out, nil
	case yamlio.PyStr:
		out := make([]string, 0, len(t))
		for _, r := range string(t) {
			out = append(out, string(r))
		}
		return out, nil
	case pyMap:
		out := make([]string, 0, len(t))
		for _, p := range t {
			out = append(out, p.K)
		}
		return out, nil
	}
	return nil, model.Errorf(model.ExitArtifact, "prd: 'components' is not iterable")
}

// shutilMove is shutil.move (`bin/company-os:698`), destination rule included.
//
// The destination rule is not a detail, and reproducing it is the difference
// between matching the oracle's file tree and not. When dst is an EXISTING
// DIRECTORY, shutil.move does not merge and does not overwrite: it moves src
// INSIDE dst, as `dst/<basename(src)>`. Completing a PRD whose id already sits
// in archive/prds/ therefore produces
//
//	archive/prds/<id>/<id>/prd.md
//
// and leaves the previously archived record in place at archive/prds/<id>/ —
// which is also the record `prd complete` then rewrites the status of and hangs
// the new outcome.md off, because both of those paths are built from dst rather
// than from where the move actually landed. That is exactly what the differential
// fixture `prd/full-lifecycle-force` exercises: examples/workspace already holds
// an archived 2026-per-channel-quiet-hours. A plain os.Rename fails with EEXIST
// there and writes nothing.
//
// A pre-existing dst/<basename> is shutil.Error, a traceback and exit 1 in the
// oracle. Here it is exit 8: refusing to overwrite an existing artifact is what
// the code means, and R-0.7a(e) sanctions naming it instead of raising.
func shutilMove(src, dst string) error {
	real := dst
	if fi, err := os.Stat(dst); err == nil && fi.IsDir() {
		real = filepath.Join(dst, filepath.Base(src))
		if _, err := os.Lstat(real); err == nil {
			return model.Errorf(model.ExitConflict,
				"destination path '%s' already exists", real)
		}
	}
	if err := scaffold.Move(src, real); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot archive %s: %v", src, err)
	}
	return nil
}

// outcomeDoc is the outcome review written at `:701-706`, byte for byte.
func outcomeDoc(id, due, title string) string {
	return "---\ntype: outcome-review\nprd: " + id + "\ndue: " + due + "\nstatus: pending\n" +
		"tags: [kind/outcome, prd/" + id + ", status/pending]\n---\n\n" +
		"# Outcome review: " + title + "\n\n" +
		"## Success metrics vs. actuals\n\n## Verdict\n\n## Learnings\n"
}

// appendLog is `open(log, "a")` (`:707-709`). The file is created when absent,
// which is what `"a"` does.
func appendLog(path, date, id, due string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return model.Errorf(model.ExitArtifact, "cannot open %s: %v", path, err)
	}
	line := "- " + date + ": PRD `" + id + "` completed and archived; outcome review due " + due + "\n"
	if _, err := f.WriteString(line); err != nil {
		f.Close()
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	return nil
}
