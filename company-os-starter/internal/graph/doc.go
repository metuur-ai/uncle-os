// Package graph derives the semantic layer from frontmatter.
//
// Responsibility: tag derivation, the feature index, CLAUDE.md nodes, and
// rebuildGenerated — the mandatory bridge between the write path and the derive
// path. rebuildGenerated lives here, not in scaffold, so the dependency stays
// one-way (scaffold -> graph) instead of cycling.
//
// Not implemented — Phase 5.
package graph
