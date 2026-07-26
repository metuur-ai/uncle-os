// Package federation is `workspace sync` and `workspace status` — the Option B
// multi-repo governance federation (bin/company-os:2057-2658).
//
// Responsibility: manifest loading and validation, sparse-checkout, slice
// materialization at 0444/0555, and the lock. It depends on internal/workspace
// for path resolution and for MANIFEST_NAME/LOCK_NAME.
//
// It is the only package that shells out (git), the only one that writes files
// with modes other than the default, and the only one whose output order is
// derived from a walk rather than from a sort. FOUR orderings live in this
// cluster and no two of them are the same; they were measured against the
// committed fixtures, not assumed:
//
//   - workspace.lock.yaml's `files:` map (:2614) is INSERTION order — a nested
//     walk of manifest slice order, then that slice's `paths:` list order, then
//     sorted(rglob) within each subtree. examples/federated's manifest lists
//     governance/ before components/, so the committed lock is the REVERSE of
//     alphabetical. yamlio.OrderedMap carries it, and gate 8's [FAIL] line order
//     is that same order arriving back through the parser (:2521).
//   - aggregate_hash (:2436) runs `for rel in sorted(files)` — a plain STRING
//     sort over the very keys the lock emits in walk order. Same data, two
//     orders, one file. OrderedMap.SortedKeys is the second one.
//   - sorted(Path.rglob(...)) (:2422) is CPython PurePath COMPONENT-WISE order,
//     not string order: `sdd/adr/a.md` precedes `sdd/adr-x.md` there and follows
//     it under a byte sort. yamlio.PathLess/SortPaths carry it.
//   - _make_readonly (:2354) walks `sorted(rglob, reverse=True)` so children are
//     chmod'd before their parents. filepath.WalkDir is pre-order, so the walk is
//     collected and reversed explicitly rather than streamed.
//
// The manifest and the lock are handled as yamlio PyValue objects rather than
// Go structs. That is deliberate: `repo_pin` iterates a pin mapping in insertion
// order and reprs the leftovers into an error message, `status` interpolates a
// whole lock `pin:` dict through an f-string, and `--only` re-emits lock entries
// for repos it did not touch. Every one of those is a Python-object behaviour
// that a struct round trip loses.
//
// Nothing here exits or writes to stdout. Sync and Status return records; the
// gate-8 half returns findings with typed Fields and no composed prose (R-2.8,
// R-2.12), so internal/validate can call it without reaching into internals.
package federation
