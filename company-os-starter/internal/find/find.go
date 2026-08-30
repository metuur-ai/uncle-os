// Package find implements `company-os find <query>` (ux-simplification 3.1):
// a read-only unified local search that fans out over five sources — canonical
// IDs, derived tags, frontmatter title/id, per-directory index entries, and
// feature-index component maps — and optionally appends a graphify hook.
//
// The design intent: a human without Obsidian and an agent without prior graph
// context have no single "everything about X". This package puts a front door
// in front of every local search mechanism the workspace already maintains,
// without removing or altering any of them.
//
// Search is not a gate: an empty result set prints "no matches" and exits 0.
// Non-zero exit is reserved for usage errors (missing query, unreadable
// workspace).
package find

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/ids"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// graphifyTimeout is the wall-clock limit for the graphify hook. The built-in
// sources already answered; graphify is supplementary, so a slow or hung
// process must not block the CLI.
const graphifyTimeout = 15 * time.Second

// Result is one search hit before it becomes a Finding. It carries the data
// the renderer needs: the workspace-relative path, the document title (when
// one exists), and the reason the document matched.
type Result struct {
	Path   string
	Title  string
	Reason string
	// Source is the section slug this result belongs to.
	Source string
	// Code is the finding code for the renderer.
	Code string
	// Rank orders results within a source: lower is first. Used to rank
	// exact-id matches before substring matches within the IDs section.
	Rank int
}

// Search runs the query across all five sources and returns one GateResult per
// non-empty source, plus a trailing graphify section when applicable. When no
// source matches, a single GateResult carries a CodeFindNoMatches finding.
//
// noGraphify skips the graphify hook and its hints entirely.
func Search(ws *workspace.Workspace, query string, noGraphify bool) ([]model.GateResult, error) {
	if query == "" {
		return nil, model.Usagef("find", "the following arguments are required: query")
	}

	var results []Result

	// Source 1: canonical IDs.
	idResults, err := searchIDs(ws, query)
	if err != nil {
		return nil, err
	}
	results = append(results, idResults...)

	// Sources 2+3: frontmatter tags, title, and id across all docs.
	fmResults, err := searchFrontmatter(ws, query)
	if err != nil {
		return nil, err
	}
	results = append(results, fmResults...)

	// Source 4: per-directory index.md content lines.
	idxResults, err := searchIndexes(ws, query)
	if err != nil {
		return nil, err
	}
	results = append(results, idxResults...)

	// Source 5: feature-index component maps.
	featResults, err := searchFeatureIndexes(ws, query)
	if err != nil {
		return nil, err
	}
	results = append(results, featResults...)

	// Build sections from results.
	sections := buildSections(results)

	// Graphify hook (unless --no-graphify).
	if !noGraphify {
		gSection := runGraphifyHook(ws, query)
		if gSection != nil {
			sections = append(sections, *gSection)
		}
	}

	// Empty result from built-in sources → "no matches", exit 0. The graphify
	// section (hint or output) is still appended when present, so the user
	// sees both the absence notice and any supplementary information.
	if len(results) == 0 {
		sections = append([]model.GateResult{{
			Ordinal: 1, Slug: "find", Title: "find",
			Findings: []model.Finding{{
				Severity: model.SevOK,
				Code:     model.CodeFindNoMatches,
				Message:  "no matches",
			}},
		}}, sections...)
	}

	return sections, nil
}

// searchIDs searches the canonical ID registry. Exact id matches rank first
// (CodeFindExactID, Rank 0), then substring matches (CodeFindSubstringID,
// Rank 1). Both are case-insensitive.
func searchIDs(ws *workspace.Workspace, query string) ([]Result, error) {
	entries, err := ids.Load(ws)
	if err != nil {
		return nil, err
	}
	lq := strings.ToLower(query)
	var exact, sub []Result
	for _, e := range entries {
		lid := strings.ToLower(e.ID)
		if lid == lq {
			exact = append(exact, Result{
				Path: e.DefinedIn, Title: e.ID, Reason: "exact id match",
				Source: model.SlugFindIDs, Code: model.CodeFindExactID, Rank: 0,
			})
		} else if strings.Contains(lid, lq) {
			sub = append(sub, Result{
				Path: e.DefinedIn, Title: e.ID, Reason: "id substring match",
				Source: model.SlugFindIDs, Code: model.CodeFindSubstringID, Rank: 1,
			})
		}
	}
	sort.Slice(exact, func(i, j int) bool { return exact[i].Path < exact[j].Path })
	sort.Slice(sub, func(i, j int) bool { return sub[i].Path < sub[j].Path })
	return append(exact, sub...), nil
}

// docInfo is one frontmatter document's searchable fields.
type docInfo struct {
	Rel   string   // workspace-relative path
	Title string   // frontmatter title:, or ""
	ID    string   // frontmatter id:, or ""
	Tags  []string // frontmatter tags: entries
}

// walkDocs walks every workspace root (including knowledge/) for .md files,
// reads frontmatter, and yields searchable docInfo tuples. It tolerates
// read-only files (0444) and skips scratchpad paths and binary-unreadable
// files silently.
func walkDocs(ws *workspace.Workspace) ([]docInfo, error) {
	roots := searchRoots(ws)
	var out []docInfo
	for _, root := range roots {
		if !dirExists(root) {
			continue
		}
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // unreadable subtree → skip
			}
			if d.IsDir() {
				if d.Name() == "scratchpad" {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(d.Name()) != ".md" {
				return nil
			}
			info, readErr := readDocInfo(ws, p)
			if readErr != nil {
				return nil // tolerate per-file read errors
			}
			if info != nil {
				out = append(out, *info)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// readDocInfo extracts searchable fields from one markdown file's
// frontmatter. Returns nil when the file has no frontmatter or no searchable
// fields.
func readDocInfo(ws *workspace.Workspace, path string) (*docInfo, error) {
	doc, err := frontmatter.ParseFile(path)
	if err != nil {
		return nil, err
	}
	if !doc.HasFrontmatter {
		return nil, nil
	}
	v, err := yamlio.PyLoadBytes(doc.YAML, path)
	if err != nil {
		return nil, err
	}
	if yamlio.PyFalsy(v) {
		return nil, nil
	}
	m, ok := v.(yamlio.PyMap)
	if !ok {
		return nil, nil
	}
	rel := relTo(ws.Root, path)
	info := &docInfo{Rel: rel}
	if t := m.Get("title"); !yamlio.PyFalsy(t) {
		info.Title = yamlio.PyString(t)
	}
	if id := m.Get("id"); !yamlio.PyFalsy(id) {
		info.ID = yamlio.PyString(id)
	}
	if tags := m.Get("tags"); !yamlio.PyFalsy(tags) {
		if seq, ok := tags.(yamlio.PySeq); ok {
			for _, t := range seq {
				info.Tags = append(info.Tags, yamlio.PyString(t))
			}
		}
	}
	// Only return docs that have at least one searchable field.
	if info.Title == "" && info.ID == "" && len(info.Tags) == 0 {
		return nil, nil
	}
	return info, nil
}

// searchFrontmatter searches derived tags, title, and id fields across all
// frontmatter documents. A document can match on multiple fields; each match
// becomes a separate finding so the renderer can report the match reason.
func searchFrontmatter(ws *workspace.Workspace, query string) ([]Result, error) {
	docs, err := walkDocs(ws)
	if err != nil {
		return nil, err
	}
	lq := strings.ToLower(query)
	var tagResults, titleResults, idResults []Result

	// Track seen paths to avoid duplicate entries when a doc matches on
	// multiple tags or fields. Each (path, code) pair appears at most once.
	type pathCode struct {
		path string
		code string
	}
	seen := map[pathCode]bool{}

	for _, d := range docs {
		// Source 2: tags.
		for _, tag := range d.Tags {
			if strings.Contains(strings.ToLower(tag), lq) {
				key := pathCode{d.Rel, model.CodeFindTag}
				if !seen[key] {
					seen[key] = true
					tagResults = append(tagResults, Result{
						Path: d.Rel, Title: d.Title,
						Reason: "tag: " + tag,
						Source: model.SlugFindTags, Code: model.CodeFindTag,
					})
				}
				break // one tag hit per doc per source is enough
			}
		}
		// Source 3a: title.
		if d.Title != "" && strings.Contains(strings.ToLower(d.Title), lq) {
			key := pathCode{d.Rel, model.CodeFindTitle}
			if !seen[key] {
				seen[key] = true
				titleResults = append(titleResults, Result{
					Path: d.Rel, Title: d.Title,
					Reason: "title match",
					Source: model.SlugFindFrontmatter, Code: model.CodeFindTitle,
				})
			}
		}
		// Source 3b: id field.
		if d.ID != "" && strings.Contains(strings.ToLower(d.ID), lq) {
			key := pathCode{d.Rel, model.CodeFindFieldID}
			if !seen[key] {
				seen[key] = true
				idResults = append(idResults, Result{
					Path: d.Rel, Title: d.Title,
					Reason: "id field match",
					Source: model.SlugFindFrontmatter, Code: model.CodeFindFieldID,
				})
			}
		}
	}

	sort.Slice(tagResults, func(i, j int) bool { return tagResults[i].Path < tagResults[j].Path })
	sort.Slice(titleResults, func(i, j int) bool { return titleResults[i].Path < titleResults[j].Path })
	sort.Slice(idResults, func(i, j int) bool { return idResults[i].Path < idResults[j].Path })

	var out []Result
	out = append(out, tagResults...)
	out = append(out, titleResults...)
	out = append(out, idResults...)
	return out, nil
}

// searchIndexes walks the workspace for per-directory index.md files and
// searches their content lines for the query. These are generated files —
// their content lists the documents in the directory, so a hit means the
// query appears in one of those listings.
func searchIndexes(ws *workspace.Workspace, query string) ([]Result, error) {
	lq := strings.ToLower(query)
	var results []Result
	seen := map[string]bool{}

	roots := searchRoots(ws)
	for _, root := range roots {
		if !dirExists(root) {
			continue
		}
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if d.Name() == "scratchpad" {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name() != "index.md" {
				return nil
			}
			rel := relTo(ws.Root, p)
			if seen[rel] {
				return nil
			}
			data, readErr := os.ReadFile(p)
			if readErr != nil {
				return nil
			}
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), lq) {
					seen[rel] = true
					results = append(results, Result{
						Path: rel, Title: filepath.Base(filepath.Dir(p)) + "/index.md",
						Reason: "index content match",
						Source: model.SlugFindIndex, Code: model.CodeFindIndexEntry,
					})
					break
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Path < results[j].Path })
	return results, nil
}

// searchFeatureIndexes searches each platform's generated feature-index.yaml
// for component ids and artifact ids that match the query.
func searchFeatureIndexes(ws *workspace.Workspace, query string) ([]Result, error) {
	lq := strings.ToLower(query)
	var results []Result

	for _, pdir := range ws.AllPlatforms() {
		fiPath := filepath.Join(pdir, "generated", "feature-index.yaml")
		v, err := yamlio.PyLoadFile(fiPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if yamlio.PyFalsy(v) {
			continue
		}
		idx, ok := v.(yamlio.PyMap)
		if !ok {
			continue
		}
		components, ok := idx.Get("components").(yamlio.PyMap)
		if !ok {
			continue
		}
		rel := relTo(ws.Root, fiPath)
		for _, pair := range components {
			cid := pair.K
			entry, _ := pair.V.(yamlio.PyMap)

			// Check component id.
			if strings.Contains(strings.ToLower(cid), lq) {
				results = append(results, Result{
					Path: rel, Title: cid,
					Reason: "feature-index component match",
					Source: model.SlugFindFeature, Code: model.CodeFindFeature,
				})
				continue
			}
			// Check artifact ids within the component entry.
			if matchArtifactID(entry, lq) {
				results = append(results, Result{
					Path: rel, Title: cid,
					Reason: "feature-index artifact match",
					Source: model.SlugFindFeature, Code: model.CodeFindFeature,
				})
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].Title < results[j].Title
	})
	return results, nil
}

// matchArtifactID checks whether any artifact id within a feature-index
// component entry matches the query.
func matchArtifactID(entry yamlio.PyMap, lq string) bool {
	for _, key := range []string{"activePrds", "archivedPrds", "discovery"} {
		if seq, ok := entry.Get(key).(yamlio.PySeq); ok {
			for _, v := range seq {
				if strings.Contains(strings.ToLower(yamlio.PyString(v)), lq) {
					return true
				}
			}
		}
	}
	if seq, ok := entry.Get("outcomes").(yamlio.PySeq); ok {
		for _, v := range seq {
			if m, ok := v.(yamlio.PyMap); ok {
				if prd := m.Get("prd"); !yamlio.PyFalsy(prd) {
					if strings.Contains(strings.ToLower(yamlio.PyString(prd)), lq) {
						return true
					}
				}
			}
		}
	}
	return false
}

// buildSections converts flat results into one GateResult per source, in the
// canonical source order: ids → tags → frontmatter → index → feature.
func buildSections(results []Result) []model.GateResult {
	// Canonical source order.
	order := []struct {
		slug  string
		title string
	}{
		{model.SlugFindIDs, "canonical IDs"},
		{model.SlugFindTags, "derived tags"},
		{model.SlugFindFrontmatter, "frontmatter"},
		{model.SlugFindIndex, "index entries"},
		{model.SlugFindFeature, "feature-index"},
	}

	bySource := map[string][]Result{}
	for _, r := range results {
		bySource[r.Source] = append(bySource[r.Source], r)
	}

	var sections []model.GateResult
	ordinal := 1
	for _, o := range order {
		rs, ok := bySource[o.slug]
		if !ok || len(rs) == 0 {
			continue
		}
		// Sort by rank first (exact before substring), then by path.
		sort.SliceStable(rs, func(i, j int) bool {
			if rs[i].Rank != rs[j].Rank {
				return rs[i].Rank < rs[j].Rank
			}
			return rs[i].Path < rs[j].Path
		})
		s := model.GateResult{Ordinal: ordinal, Slug: o.slug, Title: o.title}
		for _, r := range rs {
			s.Findings = append(s.Findings, model.Finding{
				Severity: model.SevOK,
				Code:     r.Code,
				Subject:  r.Title,
				Path:     r.Path,
				Message:  r.Reason,
				Fields: model.Fields{
					"path": r.Path, "title": r.Title, "reason": r.Reason,
				},
			})
		}
		sections = append(sections, s)
		ordinal++
	}
	return sections
}

// runGraphifyHook executes the graphify hook when conditions are met and
// returns a GateResult for the graphify section, or nil when no hook ran and
// no hint applies.
//
// The hook matrix:
//   - binary on PATH AND graph.json exists → run graphify, append output
//   - binary present but no graph.json → hint
//   - graph.json present but no binary → hint
//   - neither present → nil (no hint, no section)
//   - timeout or non-zero exit → one-line error reason, still exit 0
func runGraphifyHook(ws *workspace.Workspace, query string) *model.GateResult {
	graphJSON := filepath.Join(ws.Root, "graphify-out", "graph.json")
	binaryPath, binErr := exec.LookPath("graphify")
	graphExists := fileExists(graphJSON)

	// Neither present → no hint, no section.
	if binErr != nil && !graphExists {
		return nil
	}

	s := model.GateResult{
		Ordinal: 99, // appended last; renderer does not print ordinals
		Slug:    model.SlugFindGraphify,
		Title:   "graphify",
	}

	// One present but not the other → hint.
	if binErr != nil && graphExists {
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeFindGraphifyHint,
			Message:  "graphify not detected — install it for graph search",
		})
		return &s
	}
	if binErr == nil && !graphExists {
		_ = binaryPath // used only to confirm presence
		// The binary is installed, so "install it" would be a lie: what is
		// missing is the graph itself. Name the fix the user can run.
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeFindGraphifyHint,
			Message:  "graphify installed but no graphify-out/graph.json here — run graphify to build the graph",
		})
		return &s
	}

	// Both present → run graphify query.
	ctx, cancel := context.WithTimeout(context.Background(), graphifyTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "query", query)
	cmd.Dir = ws.Root
	out, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeFindGraphifyError,
			Message:  "graphify: timed out after 15s",
		})
		return &s
	}
	if err != nil {
		reason := err.Error()
		if ee, ok := err.(*exec.ExitError); ok {
			reason = strings.TrimSpace(string(ee.Stderr))
			if reason == "" {
				reason = ee.Error()
			}
		}
		// Truncate long error messages.
		if len(reason) > 120 {
			reason = reason[:117] + "..."
		}
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeFindGraphifyError,
			Message:  "graphify: " + reason,
		})
		return &s
	}

	// Append graphify output as findings.
	text := strings.TrimSpace(string(out))
	if text == "" {
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeFindGraphify,
			Message:  "graphify: no results",
		})
		return &s
	}
	for _, line := range strings.Split(text, "\n") {
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeFindGraphify,
			Message:  line,
		})
	}
	return &s
}

// searchRoots returns every directory find walks. It includes knowledge/
// because synced slices are local-searchable even though they are
// indexed-not-governed — find writes nothing, so the 0444 mode is irrelevant.
func searchRoots(ws *workspace.Workspace) []string {
	return []string{
		ws.Company,
		ws.Platforms,
		ws.Teams,
		filepath.Join(ws.Root, "company-ontology"),
		filepath.Join(ws.Root, "knowledge"),
	}
}

// relTo returns the workspace-relative path in forward-slash form.
func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
