package graph

// Tag derivation: bin/company-os:1307-1392.
//
// The rule the Ontology Guide states and this file enforces: IDs are canonical,
// tags are DERIVED. Nothing hand-writes a `tags:` list; `graph build` recomputes
// it from the frontmatter's own id fields and overwrites whatever was there,
// except for the four manually curated facets iterGraphDocs carries forward.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// kindTag is KIND_TAG (bin/company-os:1307-1311).
var kindTag = map[string]string{
	"prd": "kind/prd", "discovery-brief": "kind/discovery",
	"component-reality": "kind/reality", "outcome-review": "kind/outcome",
	"adr": "kind/adr", "concept": "ontology/concept",
	"bounded-context": "ontology/context", "context-map": "ontology/context-map",
	"onboarding-guide": "kind/onboarding",
}

// curatedFacets are the tag namespaces a human owns. Everything else in a
// `tags:` list is regenerated from scratch on every build, so a hand-added
// `kind/…` disappears and a hand-added `capability/…` survives (`:1391`).
var curatedFacets = map[string]bool{
	"ontology": true, "capability": true, "req": true, "spec": true,
}

// skipNames are the markdown files iterGraphDocs never ingests (`:1379`).
// CLAUDE.md is skipped BY NAME so a generated context node is never re-ingested
// as a graph document once it carries frontmatter of its own.
var skipNames = map[string]bool{
	"log.md": true, "README.md": true, "CLAUDE.md": true,
}

// Doc is one frontmatter markdown document and the tags derived from it — the
// tuple iter_graph_docs yields (`:1361`). It is shared by `graph build`, which
// writes Tags back, and validate gate 4, which compares them against what is
// committed.
type Doc struct {
	// Path is absolute, as Python's Path objects are here.
	Path string
	// Rel is Path relative to the workspace root, in posix form — the shape
	// every rendered line uses.
	Rel string
	// Meta is the loaded frontmatter mapping. Never empty: a document whose
	// frontmatter is absent or falsy is skipped, matching `if not meta`.
	Meta yamlio.PyMap
	// Tags is derive_tags' output: sorted, deduplicated, derived.
	Tags []string
}

// DeriveTags is derive_tags (`:1317-1343`).
//
// extra carries the facets the caller inferred from a document's LOCATION
// rather than its content — `platform/x` for anything under platforms/x — plus
// the curated tags being preserved. The result is sorted because Python returns
// `sorted(tags)` over a set; Python's sort of str is codepoint order, which is
// what sort.Strings gives for UTF-8.
func DeriveTags(meta yamlio.PyMap, extra []string) ([]string, error) {
	set := make(map[string]bool, len(extra)+8)
	for _, t := range extra {
		set[t] = true
	}
	if t, ok := meta.Get("type").(yamlio.PyStr); ok {
		if tag, ok := kindTag[string(t)]; ok {
			set[tag] = true
		}
	}
	addIfTruthy(set, "platform/", meta.Get("platform"))
	addIfTruthy(set, "team/", meta.Get("team"))

	components, err := pyIter(meta.Get("components"), "components")
	if err != nil {
		return nil, err
	}
	for _, c := range components {
		set["component/"+yamlio.PyString(c)] = true
	}

	// `str(bc).split("://")[-1]` — the LAST segment, so a bare id and a
	// context:// URL land on the same tag.
	if bc := meta.Get("boundedContext"); !yamlio.PyFalsy(bc) {
		text := yamlio.PyString(bc)
		parts := strings.Split(text, "://")
		set["context/"+parts[len(parts)-1]] = true
	}

	addIfTruthy(set, "status/", meta.Get("status"))
	addIfTruthy(set, "authority/", meta.Get("authority"))
	// `meta.get("fromDiscovery") and meta["fromDiscovery"] != "none"` — the
	// sentinel is compared as a string, so a PRD that declares no discovery
	// origin gets no discovery facet rather than a `discovery/none` one.
	if fd := meta.Get("fromDiscovery"); !yamlio.PyFalsy(fd) && yamlio.PyString(fd) != "none" {
		set["discovery/"+yamlio.PyString(fd)] = true
	}
	addIfTruthy(set, "prd/", meta.Get("prd"))
	if t, ok := meta.Get("type").(yamlio.PyStr); ok && string(t) == "onboarding-guide" {
		addIfTruthy(set, "role/", meta.Get("role"))
	}

	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out, nil
}

func addIfTruthy(set map[string]bool, prefix string, v yamlio.PyValue) {
	if yamlio.PyFalsy(v) {
		return
	}
	set[prefix+yamlio.PyString(v)] = true
}

// pyIter is `for x in (value or [])` over a loaded object. Python iterates a
// str by CHARACTER and a dict by KEY, and raises TypeError on a number — so a
// `components: svc-a` typo would silently produce one tag per letter rather
// than one tag. Reproducing the str and dict cases keeps the observable
// behaviour; the TypeError case becomes exit 4 with a diagnostic per
// R-0.7a(j), which writes nothing exactly as the traceback did.
func pyIter(v yamlio.PyValue, field string) ([]yamlio.PyValue, error) {
	if yamlio.PyFalsy(v) {
		return nil, nil
	}
	switch t := v.(type) {
	case yamlio.PySeq:
		return t, nil
	case yamlio.PyStr:
		out := make([]yamlio.PyValue, 0, len(t))
		for _, r := range string(t) {
			out = append(out, yamlio.PyStr(string(r)))
		}
		return out, nil
	case yamlio.PyMap:
		out := make([]yamlio.PyValue, 0, len(t))
		for _, p := range t {
			out = append(out, yamlio.PyStr(p.K))
		}
		return out, nil
	}
	return nil, model.Errorf(model.ExitArtifact,
		"'%s:' must be a list, a mapping or a string", field)
}

// ReadFrontmatter is frontmatter() (`:76-82`) with safe_load applied: the
// mapping and the body. A document with no fence, or with falsy frontmatter,
// yields a nil mapping — Python's `{}` — and the whole text as the body.
//
// A frontmatter block that parses to something other than a mapping exits 4:
// Python reaches `meta.get(...)` on a list and raises AttributeError, writing
// nothing (R-0.7a(j)).
func ReadFrontmatter(path string) (yamlio.PyMap, []byte, error) {
	doc, err := frontmatter.ParseFile(path)
	if err != nil {
		return nil, nil, model.Errorf(model.ExitArtifact, "%v", err)
	}
	if !doc.HasFrontmatter {
		return nil, doc.Body, nil
	}
	v, err := yamlio.PyLoadBytes(doc.YAML, path)
	if err != nil {
		return nil, nil, err
	}
	if yamlio.PyFalsy(v) {
		return nil, doc.Body, nil
	}
	m, ok := v.(yamlio.PyMap)
	if !ok {
		return nil, nil, model.Errorf(model.ExitArtifact,
			"%s: frontmatter must be a mapping", path)
	}
	return m, doc.Body, nil
}

// IterGraphDocs is iter_graph_docs (`:1346-1392`), shared by `graph build`
// (which writes tags) and gate 4 (which compares committed tags against a fresh
// derivation) — one traversal, one derivation, so the writer and the drift gate
// cannot disagree.
//
// knowledge/ is deliberately NOT a root here. It is a node root but not a
// graph-docs root: the slices under it are foreign, read-only, and carry no
// `type:` frontmatter, so deriving tags for them would mean writing to a 0444
// tree.
func IterGraphDocs(ws *workspace.Workspace) ([]Doc, error) {
	var out []Doc
	for _, root := range graphRoots(ws) {
		if !exists(root) {
			continue
		}
		// The location facet is inferred from the parent directory's name, so a
		// doc that omits `platform:`/`team:` still gets one.
		locKey := ""
		switch filepath.Base(filepath.Dir(root)) {
		case "platforms":
			locKey = "platform"
		case "teams":
			locKey = "team"
		}
		mds, err := rglobMarkdown(root)
		if err != nil {
			return nil, err
		}
		for _, md := range mds {
			// `"scratchpad" in md.parts` tests the ABSOLUTE path's components,
			// not the workspace-relative ones. That is a real trap — a checkout
			// living under a directory named scratchpad skips every document —
			// and iter_knowledge_docs (`:1673`) carries a comment saying so.
			// Reproduced rather than fixed: R-0.7 makes the Python behaviour the
			// contract, and quietly ingesting more documents than the oracle
			// would is a silent divergence.
			if hasPathComponent(md, "scratchpad") || skipNames[filepath.Base(md)] {
				continue
			}
			meta, _, err := ReadFrontmatter(md)
			if err != nil {
				return nil, err
			}
			if len(meta) == 0 {
				continue
			}
			var extra []string
			if locKey != "" {
				extra = append(extra, locKey+"/"+filepath.Base(root))
			}
			if t, ok := meta.Get("type").(yamlio.PyStr); ok && string(t) == "component-reality" {
				// `str(meta.get("id", "")).replace("reality-", "")` — str.replace
				// removes EVERY occurrence, not just the id's prefix.
				id := ""
				if v := meta.Get("id"); v != nil {
					id = yamlio.PyString(v)
				}
				if cid := strings.ReplaceAll(id, "reality-", ""); cid != "" {
					extra = append(extra, "component/"+cid)
				}
			}
			existing, err := pyIter(meta.Get("tags"), "tags")
			if err != nil {
				return nil, err
			}
			for _, t := range existing {
				text := yamlio.PyString(t)
				if curatedFacets[strings.SplitN(text, "/", 2)[0]] {
					extra = append(extra, text)
				}
			}
			tags, err := DeriveTags(meta, extra)
			if err != nil {
				return nil, err
			}
			out = append(out, Doc{Path: md, Rel: relTo(ws.Root, md), Meta: meta, Tags: tags})
		}
	}
	return out, nil
}

// graphRoots is iter_graph_docs' `roots` (`:1352-1353`).
func graphRoots(ws *workspace.Workspace) []string {
	roots := []string{ws.Company}
	roots = append(roots, ws.AllPlatforms()...)
	roots = append(roots, ws.AllTeams()...)
	return append(roots, filepath.Join(ws.Root, "company-ontology"))
}

// RewriteFrontmatterTags is rewrite_frontmatter_tags (`:1346-1358`): a
// read-modify-write that re-serializes the frontmatter but never re-implements
// the fence parse. It reports whether the file changed.
//
// The emitter is PyDumpAutoFlow, not PyDump: safe_dump runs here with
// default_flow_style=None, which is what makes committed frontmatter read
// `tags: [a, b]` inline. Dumping in block style would rewrite every document on
// the first build.
func RewriteFrontmatterTags(path string, tags []string) (bool, error) {
	meta, body, err := ReadFrontmatter(path)
	if err != nil {
		return false, err
	}
	if len(meta) == 0 {
		return false, nil
	}
	want := make(yamlio.PySeq, len(tags))
	for i, t := range tags {
		want[i] = yamlio.PyStr(t)
	}
	if yamlio.PyEqual(meta.Get("tags"), want) {
		return false, nil
	}
	// dict assignment: an existing `tags:` keeps its position, a new one is
	// appended. That is why a re-tagged document does not also get its keys
	// reordered.
	meta = meta.Set("tags", want)
	fm, err := yamlio.PyDumpAutoFlow(meta)
	if err != nil {
		return false, model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", path, err)
	}
	text := "---\n" + strings.TrimSpace(fm) + "\n---\n" + string(body)
	if err := os.WriteFile(path, []byte(text), 0o666); err != nil {
		return false, model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	return true, nil
}

// ------------------------------------------------------------- filesystem

// rglobMarkdown is `sorted(root.rglob("*.md"))`. The sort is PurePath order via
// yamlio.SortPaths, NOT sort.Strings: CPython compares paths component-wise, so
// `sdd/adr/a.md` precedes `sdd/adr-x.md` where a byte sort puts it after
// (R-0.11a). Directories are excluded — pathlib would yield one named `*.md`
// and Python would then raise on read; nothing in the corpus has one, and a
// crash is not behaviour worth reproducing.
func rglobMarkdown(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// An unreadable subtree is skipped rather than fatal: rglob yields
			// what it can reach and Python surfaces nothing here.
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot walk %s: %v", root, err)
	}
	yamlio.SortPaths(out)
	return out, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// hasPathComponent is `name in path.parts`.
func hasPathComponent(path, name string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == name {
			return true
		}
	}
	return false
}

// relTo is Path.relative_to(base).as_posix(). It assumes base is a lexical
// prefix of path, which every caller guarantees by construction; a path that is
// not under base comes back unchanged rather than as a chain of "..", because
// Python raises there and no caller is prepared for a "../" answer.
func relTo(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// under reports whether path is base or lives beneath it — Python's
// `path.relative_to(base)` succeeding, which is the test group_docs_by_root
// uses to assign each document to exactly one federation root.
func under(base, path string) (string, bool) {
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}
