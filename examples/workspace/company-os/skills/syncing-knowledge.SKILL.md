---
id: skill://governance/syncing-knowledge
type: skill
version: '1.0'
authority: canonical
appliesTo: ['company://all-teams']
inputs: [a repo whose documentation should join the workspace knowledge catalog]
outputs: [materialized slices + workspace.lock.yaml passing `company-os validate`]
tags: [authority/canonical]
---

# Syncing Knowledge Into the Catalog

The catalog is pull-only. Nothing here writes back to a source repo: you change
the source, bump the pin, and re-sync.

1. (mandatory) Pin the repo in `workspace.yaml` with `commit:` or `tag:`.
   A branch is rejected as non-reproducible. A tag is resolved to a SHA and
   recorded in the lock at sync time.
2. (mandatory) List ONLY documentation directories in `paths:`.
   `docs/sdd`, `specs/`, `architecture/`. The allowlist is the only thing
   keeping source code out, and there is no exclude form — never widen it to
   the repo root.
3. (mandatory) Set `localDirectory:` to `knowledge/<area>/<repo>`.
   The path below it mirrors the source, so `paths: [docs/sdd]` lands at
   `knowledge/<area>/<repo>/docs/sdd/`. There is no rename.
4. (default) Use a `slices:` list when one repo contributes several areas.
   One clone and one cache serve every destination. Targets must be disjoint —
   a nested pair is refused, because the outer slice would freeze the inner one.
5. (mandatory) Run `company-os workspace status` before every sync.
   It reports whether a pin or a slice target moved without a re-sync, which
   the file hashes alone cannot tell you.
6. (mandatory) Run `company-os workspace sync`, then commit three things at once.
   The manifest, the materialized slices, and `workspace.lock.yaml`. The lock
   diff is the audit trail of which content moved.
7. (mandatory) Never edit a synced file — change the source and re-sync.
   Slices are `0444` derived content and gate `[8/8]` fails on any hand-edit.
8. (default) Run `company-os graph build` after a sync.
   It refreshes the generated index in `knowledge/CLAUDE.md`, which is how an
   agent discovers what the catalog holds.
9. (guidance) Deciding WHICH docs are worth cataloging is yours.
   The contract is only that the allowlist stays narrow and the lock stays
   honest — not how you choose what goes in it.
