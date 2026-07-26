package graph

// The two entry points: `graph build` (cmd_graph, bin/company-os:1787-1797) and
// RebuildGenerated (rebuild_generated, `:1803-1810`).
//
// They are the same derivation with a different mouth. `graph build` also
// re-tags every document and prints a summary; RebuildGenerated re-tags
// silently and reports only the derived aggregates, because it runs INSIDE
// another command whose own output has to follow.

import (
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Build is cmd_graph: derive tags for every frontmatter doc, then rebuild the
// derived aggregates.
//
// The four sections come back in emission order and each is a section, not a
// gate — `graph build` prints no headers. Ordinal exists for --json and the TUI.
func Build(ws *workspace.Workspace) ([]model.GateResult, error) {
	docs, err := IterGraphDocs(ws)
	if err != nil {
		return nil, err
	}
	tagged := model.GateResult{Ordinal: 1, Slug: model.SectionTags, Title: "derived tags"}
	changed := 0
	for _, d := range docs {
		wrote, err := RewriteFrontmatterTags(d.Path, d.Tags)
		if err != nil {
			return nil, err
		}
		if !wrote {
			continue
		}
		changed++
		tagged.Findings = append(tagged.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeGraphTagged,
			Path:     d.Rel,
			Fields:   model.Fields{"path": d.Rel, "tags": d.Tags},
		})
	}

	aggregates, err := rebuild(ws, docs, 2)
	if err != nil {
		return nil, err
	}

	summary := model.GateResult{Ordinal: 4, Slug: model.SectionSummary, Title: "summary",
		Findings: []model.Finding{{
			Severity: model.SevOK,
			Code:     model.CodeGraphSummary,
			Fields:   model.Fields{"scanned": len(docs), "updated": changed},
		}}}
	return append(append([]model.GateResult{tagged}, aggregates...), summary), nil
}

// Rebuild is rebuild_generated (`:1803-1810`): re-derive tags and the generated
// aggregates through the same code path as `graph build`, so a freshly
// scaffolded workspace validates green without a separate build step.
//
// It returns only the aggregate sections. That is not an omission — Python's
// rebuild_generated calls rewrite_frontmatter_tags directly, without cmd_graph's
// print, and emits no summary line either. The scaffolding commands print these
// lines BEFORE their own output, which is why the seam is ordered.
func Rebuild(ws *workspace.Workspace) ([]model.GateResult, error) {
	docs, err := IterGraphDocs(ws)
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		if _, err := RewriteFrontmatterTags(d.Path, d.Tags); err != nil {
			return nil, err
		}
	}
	return rebuild(ws, docs, 1)
}

// rebuild is write_feature_indexes followed by write_claude_nodes, in that
// order. The order is observable: every "wrote index" line precedes every
// "node" line in the oracle's output.
func rebuild(ws *workspace.Workspace, docs []Doc, ordinal int) ([]model.GateResult, error) {
	indexes := model.GateResult{Ordinal: ordinal, Slug: model.SectionFeatureIndexes,
		Title: "feature indexes"}
	written, err := WriteFeatureIndexes(ws)
	if err != nil {
		return nil, err
	}
	for _, rel := range written {
		indexes.Findings = append(indexes.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeGraphIndexWritten,
			Path:     rel,
			Fields:   model.Fields{"path": rel},
		})
	}

	nodeFindings, err := writeClaudeNodes(ws, docs)
	if err != nil {
		return nil, err
	}
	nodes := model.GateResult{Ordinal: ordinal + 1, Slug: model.SectionClaudeNodes,
		Title: "context nodes", Findings: nodeFindings}
	return []model.GateResult{indexes, nodes}, nil
}
