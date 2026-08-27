package graph

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/federation"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
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

// indexHeader is the new-file preamble for a generated index.md: R-2.3's
// frontmatter, and nothing else. rewriteGeneratedBlockWithHeader appends the
// marker region after it.
const indexHeader = indexFrontmatter

// sliceRoots is R-2.7: the absolute directory of every manifest-declared slice.
//
// Nothing may be written into one. A materialized slice ships 0444/0555 and its
// bytes are hashed into workspace.lock.yaml, so a write there fails gate 8 —
// invariant I9. The drift check must skip the same paths, or a slice directory
// that qualifies but cannot be written would report drift forever.
//
// An absent manifest is not an error: LoadManifest returns (nil, nil) for a
// monorepo, which is the common case and yields no exclusions.
//
// Today no slice directory in examples/federated holds two graph documents, so
// the threshold happens to shield them anyway. That is a coincidence of the
// current fixture, not a guarantee — one more document in a source repo turns it
// into a gate-8 failure. Hence the explicit exclusion.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.7
func sliceRoots(ws *workspace.Workspace) ([]string, error) {
	m, err := federation.LoadManifest(ws)
	if err != nil || m == nil {
		return nil, err
	}
	var out []string
	for _, repo := range m.Repos {
		slices, err := federation.RepoSlices(repo)
		if err != nil {
			// A malformed repo entry is the manifest gate's business, not the
			// index writer's. Excluding nothing here would let a write reach a
			// slice, so the safe reading of an unparseable entry is to skip
			// generation entirely rather than guess at its destination.
			return nil, err
		}
		for _, s := range slices {
			if s.LocalDirectory == "" {
				continue
			}
			out = append(out, filepath.Join(ws.Root, filepath.FromSlash(s.LocalDirectory)))
		}
	}
	return out, nil
}

// excluded reports whether dir is at or beneath any of roots.
func excluded(dir string, roots []string) bool {
	for _, r := range roots {
		if dir == r || strings.HasPrefix(dir, r+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// WriteIndexes materializes the generated index.md files (R-2.6), removes the
// ones whose directory no longer qualifies (R-2.9), and writes nothing into a
// manifest-declared slice (R-2.7).
//
// It goes through rewriteGeneratedBlockWithHeader rather than os.WriteFile so a
// hand-written index.md — one carrying no markers — is left alone and reported
// rather than clobbered (R-2.10), exactly as a hand-owned CLAUDE.md is.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.6
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.9
func WriteIndexes(ws *workspace.Workspace, docs []Doc) ([]model.Finding, error) {
	skip, err := sliceRoots(ws)
	if err != nil {
		return nil, err
	}
	want := BuildIndexes(docs)

	dirs := make([]string, 0, len(want))
	for dir := range want {
		if !excluded(dir, skip) {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs) // findings are rendered in order; map order is not one

	var out []model.Finding
	for _, dir := range dirs {
		path := filepath.Join(dir, "index.md")
		outcome, _, err := rewriteGeneratedBlockWithHeader(
			path, buildIndexBlock(dir, groupOf(want, dir, docs)), indexHeader)
		if err != nil {
			return nil, err
		}
		rel := relTo(ws.Root, path)
		switch outcome {
		case nodeWritten:
			out = append(out, model.Finding{Severity: model.SevOK,
				Code: model.CodeGraphDirIndexWritten, Path: rel,
				Fields: model.Fields{"path": rel}})
		case nodeHandOwned:
			out = append(out, model.Finding{Severity: model.SevOK,
				Code: model.CodeGraphDirIndexHandOwned, Path: rel,
				Fields: model.Fields{"path": rel}})
		}
	}

	stale, err := staleIndexes(ws, want, skip)
	if err != nil {
		return nil, err
	}
	for _, path := range stale {
		if err := os.Remove(path); err != nil {
			return nil, model.Wrapf(model.ExitArtifact, err,
				"cannot remove %s: %v", path, err)
		}
		rel := relTo(ws.Root, path)
		out = append(out, model.Finding{Severity: model.SevOK,
			Code: model.CodeGraphDirIndexRemoved, Path: rel,
			Fields: model.Fields{"path": rel}})
	}
	return out, nil
}

// groupOf recovers the documents that belong to dir. BuildIndexes already
// decided which directories qualify; this re-selects the members so the block
// renderer sees the same set.
func groupOf(want map[string]string, dir string, docs []Doc) []Doc {
	var out []Doc
	for _, d := range docs {
		if filepath.Dir(d.Path) == dir {
			out = append(out, d)
		}
	}
	return out
}

// staleIndexes is R-2.9: generated index.md files whose directory no longer
// holds two graph documents.
//
// Only files carrying a balanced marker pair are returned. A hand-written
// index.md is never deleted — deleting authored content because a threshold
// moved would be the worst possible failure mode here, and it is the same
// judgement rewriteGeneratedBlock already makes for a marker-less CLAUDE.md.
func staleIndexes(ws *workspace.Workspace, want map[string]string, skip []string) ([]string, error) {
	var out []string
	for _, root := range graphRoots(ws) {
		if !exists(root) {
			continue
		}
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // an unreadable subtree is skipped, as IterGraphDocs does
			}
			if d.IsDir() || d.Name() != "index.md" {
				return nil
			}
			dir := filepath.Dir(p)
			if _, keep := want[dir]; keep || excluded(dir, skip) {
				return nil
			}
			raw, readErr := os.ReadFile(p)
			if readErr != nil {
				return nil
			}
			if _, ok := ExtractGeneratedBlock(readText(raw)); ok {
				out = append(out, p)
			}
			return nil
		})
		if err != nil {
			return nil, model.Wrapf(model.ExitArtifact, err,
				"cannot scan %s for stale indexes: %v", root, err)
		}
	}
	sort.Strings(out)
	return out, nil
}
