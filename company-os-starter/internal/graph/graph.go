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
	tagged, err := retag(docs, 1)
	if err != nil {
		return nil, err
	}
	changed := len(tagged.Findings)

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
// It returns the tag section as well as the aggregates (R-0.4). Python's
// rebuild_generated called rewrite_frontmatter_tags for its EFFECT and dropped
// the answer; that was fine while nothing downstream asked what had been
// rewritten, and stopped being fine when the repair path had to report exactly
// that. The list is returned, not re-derived (R-0.6).
//
// It still emits no summary line — that tally is cmd_graph's alone. The
// scaffolding commands print these lines BEFORE their own output, which is why
// the seam is ordered.
func Rebuild(ws *workspace.Workspace) ([]model.GateResult, error) {
	docs, err := IterGraphDocs(ws)
	if err != nil {
		return nil, err
	}
	tagged, err := retag(docs, 1)
	if err != nil {
		return nil, err
	}
	aggregates, err := rebuild(ws, docs, 2)
	if err != nil {
		return nil, err
	}
	return append([]model.GateResult{tagged}, aggregates...), nil
}

// retag is the single tag-rewriting traversal, shared by both entry points.
//
// It exists because of R-0.6. The repair path needs to report which documents
// it re-tagged, and the cheap way to get that list — walk the docs again
// afterwards and diff — would be a second traversal whose answer can disagree
// with what the first one actually wrote (a file changed underneath, a write
// that reported no-op). The set of rewritten files is therefore produced BY the
// writing loop, as its return value, and there is exactly one such loop.
//
// The findings are the section `graph build` has always emitted, unchanged, so
// sharing them with Rebuild costs `graph build` nothing (R-0.5).
//
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.4
// @spec req://uncle-os/derived-drift-repair@0.1#R-0.6
func retag(docs []Doc, ordinal int) (model.GateResult, error) {
	tagged := model.GateResult{Ordinal: ordinal, Slug: model.SectionTags, Title: "derived tags"}
	for _, d := range docs {
		wrote, err := RewriteFrontmatterTags(d.Path, d.Tags)
		if err != nil {
			return model.GateResult{}, err
		}
		if !wrote {
			continue
		}
		tagged.Findings = append(tagged.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeGraphTagged,
			Path:     d.Rel,
			Fields:   model.Fields{"path": d.Rel, "tags": d.Tags},
		})
	}
	return tagged, nil
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
