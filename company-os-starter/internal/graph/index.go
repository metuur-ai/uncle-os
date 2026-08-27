package graph

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// indexThreshold is D-2.3: a directory earns an index at two documents.
//
// One is deliberate, not arbitrary. At a threshold of 1 every leaf directory
// gets an index that restates the single file already named in its parent's
// listing, and the reader learns to skip indexes entirely — which costs the
// feature its whole point.
const indexThreshold = 2

// indexFrontmatter is R-2.3. `type: index` and nothing else.
//
// No `id:`, because nothing reads one: index.md is skipped by name during
// traversal (skipNames, tags.go), so it derives no tags and faces no core-field
// gate. No `title:` either — it would be the directory name, which is precisely
// what the description rubric in docs/FRONTMATTER-CORE.md calls a fact worth
// nothing. The `type:` exists for consumers outside this repository that filter
// on it.
const indexFrontmatter = "---\ntype: index\n---\n\n"

// BuildIndexes is R-2.1, R-2.2 and R-2.3: the content of every generated
// index.md, keyed by the absolute directory it belongs in.
//
// It is a pure function of docs and touches no filesystem. That is what lets the
// threshold be tested without a fixture tree, and it is why the wiring (R-2.6),
// the slice exclusion (R-2.7) and the hand-owned case (R-2.10) can each be
// layered on top without re-testing the rendering.
//
// Directories are counted by DIRECT children only. A recursive count would make
// every ancestor qualify and produce an index at every level of the tree.
//
// Ordering is total and derived entirely from the input — types sorted, then
// documents by path within a type — and nothing here reads a clock or a map
// iteration order. That is what R-2.11's idempotency rests on.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.1
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.2
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.3
func BuildIndexes(docs []Doc) map[string]string {
	byDir := map[string][]Doc{}
	for _, d := range docs {
		dir := filepath.Dir(d.Path)
		byDir[dir] = append(byDir[dir], d)
	}

	out := map[string]string{}
	for dir, group := range byDir {
		if len(group) < indexThreshold {
			continue
		}
		out[dir] = indexFrontmatter + RenderGeneratedRegion(buildIndexBlock(dir, group)) + "\n"
	}
	return out
}

// buildIndexBlock renders one directory's generated interior.
//
// The grouping idiom is BuildClaudeNode's (node.go), copied rather than
// reinvented: same "other" fallback for a missing type, same sorted type list,
// same stable sort by path within a type, and the same title fallback written as
// successive overrides — base name, then `id`, then `title` — which is Python's
// `title or id or Path(rel).name`. A future change to that chain is then one
// grep, not two divergent implementations.
func buildIndexBlock(dir string, docs []Doc) string {
	byType := map[string][]Doc{}
	var types []string
	for _, d := range docs {
		t := "other"
		if v := d.Meta.Get("type"); !yamlio.PyFalsy(v) {
			t = yamlio.PyString(v)
		}
		if _, seen := byType[t]; !seen {
			types = append(types, t)
		}
		byType[t] = append(byType[t], d)
	}
	sort.Strings(types)

	lines := []string{"## " + filepath.Base(dir) + " — index", ""}
	for _, t := range types {
		lines = append(lines, "- **"+t+"**")
		group := byType[t]
		sort.SliceStable(group, func(i, j int) bool { return group[i].Path < group[j].Path })
		for _, d := range group {
			// Entries sit directly in this directory by construction, so the
			// link is the base name.
			name := filepath.Base(d.Path)
			title := name
			if v := d.Meta.Get("id"); !yamlio.PyFalsy(v) {
				title = yamlio.PyString(v)
			}
			if v := d.Meta.Get("title"); !yamlio.PyFalsy(v) {
				title = yamlio.PyString(v)
			}
			line := "  - [" + title + "](" + name + ")"
			// An absent description renders nothing at all. A placeholder on
			// every line would train the reader to skip the column.
			if v := d.Meta.Get("description"); !yamlio.PyFalsy(v) {
				line += " — " + strings.TrimSpace(yamlio.PyString(v))
			}
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
