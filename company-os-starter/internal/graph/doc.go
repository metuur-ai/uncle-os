// Package graph derives the semantic layer from frontmatter.
//
// Responsibility: tag derivation, the feature index, CLAUDE.md context nodes,
// and RebuildGenerated — the mandatory bridge between the write path and the
// derive path. RebuildGenerated lives here, not in scaffold, so the dependency
// stays one-way (scaffold -> graph) instead of cycling; scaffold declares
// `type Rebuild func(*workspace.Workspace) ([]string, error)` and cmd/company-os
// injects the implementation. This package therefore SATISFIES that type and
// never imports scaffold.
//
// Everything here is derived. Two invariants follow and are worth stating
// before reading any of it:
//
//   - Derived artifacts are compared SEMANTICALLY, never by bytes (R-0.7c).
//     The Python original compares feature-index bytes (`bin/company-os:1535`)
//     and gets away with it because the same PyYAML emitter wrote them; this
//     port's emitter signature differs on some documents, so a byte compare
//     would rewrite every index on the first build and `graph build; graph
//     build` would never settle. Every write below is guarded on
//     yamlio.PyDumpCanonical of the parsed structure, which is exactly what
//     gate 6 already does (`:1053`).
//
//   - Order is reproduced by MECHANISM, not by result (R-0.11). Python reaches
//     a sorted feature index because `build_feature_index` iterates
//     `sorted(cids)`, not because `feature_index_unresolved` sorts anything;
//     and it reaches PurePath order — not string order — wherever it walks with
//     `sorted(Path.rglob(...))`. The two disagree (`sdd/adr/a.md` before
//     `sdd/adr-x.md`), so the walks below use yamlio.SortPaths while the places
//     Python sorts plain strings use sort.Strings. Each site names which.
package graph
