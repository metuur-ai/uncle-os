---
type: lld
id: lld-federation-enrichment
title: Federation Enrichment — Low-Level Design
status: locked
tags: [kind/lld, status/locked]
---

# Federation Enrichment — Low-Level Design

All changes are confined to `company-os-starter/`. The CLI stays a single
self-contained file (`bin/company-os`) with the `die/ok/warn/fail` helpers and
the `^---\n...\n---\n` `frontmatter()` parser preserved. Every mutating command
keeps printing the next command in the workflow.

Committed scope: #2 feature-index, #8 pointers, #1 CLAUDE.md nodes, #5 identity,
#6 onboarding. Deferred (demand-driven): #7 data-catalog, #4 customer/call,
#3 workflow-skill — designed in §Deferred, not implemented in this initiative.

## Architecture

### Reuse the existing spine — one traversal, three consumers

`iter_graph_docs(ws)` (`bin/company-os:759-791`) already walks every frontmatter
doc under company / platforms / teams / company-ontology and yields
`(path, meta, derived_tags)`. It is the shared spine for `graph build` (writes
tags) and gate [4/4] (drift-checks tags). We keep it as the **doc source** and
add two aggregate consumers — but they **collect then sort** before rendering
rather than emitting from inside the per-doc loop:

- `tags:` — per **doc**, independent, order-immune (already `sorted()`).
- CLAUDE.md node — per **root** (aggregate over a root's docs).
- feature-index — per **platform** (aggregate over reference edges).

Because two outputs aggregate, iteration order matters and
`root.rglob("*.md")` (`:775`) is **unsorted**. Aggregate builders MUST sort doc
lists and dict keys before rendering, or generated files drift across checkouts.

### Component map

| Piece | Where | New / changed | Phase |
|---|---|---|---|
| Shared canonicalizer + parsed-compare | new `canonical_yaml()`, `blocks_equal()` | new | 0 |
| Deterministic doc ordering | sort `:775` rglob for aggregate consumers | changed | 0 |
| Hard `CLAUDE.md` name-skip | `iter_graph_docs` skip set (`:776`) | changed | 0 |
| Characterization harness | `examples/` golden snapshot + double-build script | new | 0 |
| Standalone-team fixture | new `examples/standalone-team/` | new | 0 |
| `pointers` shape check | new `pointer_errors()` | new | 1 |
| Feature-index builder | new `build_feature_index(ws, pdir)` | new | 1 |
| Validate gate [6/6] | `cmd_validate` (`:543-628`) | changed | 1 |
| `graph build` writes index | `cmd_graph` (`:794-802`) | changed | 1 |
| New doc types (kind tags) | `KIND_TAG` (`:713-716`) | changed | 1–2 |
| CLAUDE.md node generator | new `build_claude_node(root, docs)` | new | 2 |
| Fail-safe block rewrite | new `rewrite_generated_block(path, block)` | new | 2 |
| Validate gate [5/6] | `cmd_validate` | changed | 2 |
| `roster`/`channels` shape check | new `identity_errors()` | new | 2 |
| `today --role` onboarding pointer | `cmd_today` (`:679-706`) | changed | 2 |

### Data flow

```
frontmatter docs ──iter_graph_docs (sorted)──┬─▶ derive_tags ──▶ tags:            (existing)
                                             ├─▶ build_claude_node(root, docs)  (Phase 2, #1)
                                             └─▶ build_feature_index(ws, pdir)  (Phase 1, #2)

validate: gate4 tags │ gate5 CLAUDE.md block (canonical-string) │ gate6 index (parsed-equality)
          — every new gate PASSES when its artifact is absent —
```

## Key Decisions

### 0. Idempotency: compare parsed structure, exclude volatile fields, sort everything

The house precedent is gate [4/4]: it compares `sorted(meta["tags"])` to a fresh
`sorted()` derivation (`:619-620`) — **structure, not bytes**. We follow it:

- **feature-index:** compare **parsed-YAML dict equality** (committed vs fresh),
  with volatile keys excluded — sidesteps every `yaml.safe_dump` ordering/quoting
  divergence (note `:60-61` vs `:754` already dump with different
  `default_flow_style`).
- **CLAUDE.md block:** no parse target → **canonical-string compare**: extract
  the between-markers substring, `.strip()`, normalize to `\n`.
- **No volatile fields in derived artifacts.** Drop `generatedAt`/timestamps from
  the feature-index and nodes entirely (git already timestamps the commit). This
  kills pre-mortem R#1 at the source — a fresh render is byte-for-byte the intent
  of the committed one. If a timestamp is ever wanted, it MUST be masked from the
  compare, stated explicitly.
- **Single-source rendering:** each artifact is produced by exactly one function,
  called identically by `cmd_graph` (writes) and `cmd_validate` (renders +
  compares). Build and validate must never compute an artifact via divergent code
  paths, or they will drift.

### 1. Generated CLAUDE.md nodes use a fail-safe marker block (#1, Phase 2)

A federation root's `CLAUDE.md` is **hand-owned prose** with one generated region:

```
<!-- company-os:generated:start — do not edit; run `company-os graph build` -->
... generated Doc Index + child links + (teams) identity summary ...
<!-- company-os:generated:end -->
```

`rewrite_generated_block(path, block)` contract (fail-safe, never guess):

- Match **only** the literal HTML-comment sentinels, non-greedy, `re.DOTALL`.
  **Never** reuse the frontmatter regex `^---\n(.*?)\n---\n` (`:71`,`:747`) —
  CLAUDE.md prose legitimately contains `---` (rules, table separators); that
  regex would truncate at the first stray fence and clobber prose.
- **0 markers + file absent** → create with a minimal hand-owned header + block.
- **0 markers + file exists** (hand-written CLAUDE.md) → **append** the block,
  altering no existing prose; validate does **not** fail such a marker-less file.
- **exactly 1 start + 1 end, in order** → replace only the interior.
- **any other combination** (start without end, end before start, >1 pair) →
  `warn` and mutate nothing. Silent partial rewrite is the failure mode designed
  out.
- **Byte-preservation invariant:** everything before start and after end copied
  verbatim; rewriting an identical block yields an identical file.

`build_claude_node(root, docs)` is a **pure** function of (sorted docs,
`team.yaml`) — no time, no rglob-order dependence. It renders a Doc Index grouped
by `type` + child-root links; for a team root it appends the roster/channels/
pointers summary from `team.yaml` (#5). Emitted at each root in
`iter_graph_docs`' root list: `company-os/`, each `platforms/<p>/`, each
`teams/<t>/`, `company-ontology/`.

**`CLAUDE.md` hard name-skip** (Phase 0): add `"CLAUDE.md"` to the
`iter_graph_docs` skip set at `:776`. Today it is saved only accidentally by
`if not meta: continue` (`:779`) — a CLAUDE.md that gains frontmatter would be
ingested and its own block re-tagged. The `.yaml` derived files are never matched
by the `*.md` rglob, so no feedback loop there.

### 2. Feature index is a platform-scoped derived file (#2, Phase 1 MVP)

`build_feature_index(ws, pdir)` writes `platforms/<p>/generated/feature-index.yaml`
by following the reference edges already in frontmatter (`components`,
`fromDiscovery`, `prd`). Keyed by component id (the catalog's stable unit):

```yaml
platform: <pid>            # no generatedAt — volatile fields excluded (Decision 0)
components:
  customer-notification-service:
    reality: reality/components/customer-notification-service.md
    activePrds: [2026-...]           # sorted
    archivedPrds: [2026-per-channel-quiet-hours]
    discovery: [2026-per-channel-quiet-hours]   # via fromDiscovery edges only
    outcomes: [{prd: ..., due: 2026-10-16, status: pending}]
    externalPointers: [...]          # collected from pointers: on those docs (#8)
```

- Under `generated/` → same "derived, never hand-edited, CI regenerates and
  diffs" contract as `effective-governance.yaml`.
- Records only the discovery **id** edge (already public via a PRD's
  `fromDiscovery`); never copies team-private brief bodies.
- All lists sorted; multi-source join reuses patterns from `cmd_today`
  (`:679-706`) and the PRD commands — assembly, not new infrastructure, but the
  largest new surface and highest drift risk, hence MVP-first behind the harness.

### 3. New gates appended; [1–4] frozen; absence-tolerant (Phase 1–2)

`cmd_validate` grows 4→6:

- **[5/6] CLAUDE.md node drift** (Phase 2) — for each root, render fresh and
  canonical-string-compare to the committed block. **Absent node → PASS.**
- **[6/6] feature-index drift** (Phase 1) — for each platform, rebuild and
  parsed-equality-compare. **Absent index → PASS.** Unresolved discovery/prd id
  in an entry → FAIL, id named.

`core_field_errors` and gates [1/4]–[4/4] are untouched (R-10.3). The renumber is
cosmetic (`"[n/4]"` strings at `:548,569,592,612`). Safety net without a test
suite: (a) golden `validate` stdout snapshot pre-change, diff after allowing only
header lines to move; (b) `graph build; graph build; git diff --exit-code`; both
committed as a ~10-line acceptance script (Phase 0).

### 4. Identity + pointers are validated shapes (#5, #8)

- `pointer_errors(meta)`: if `pointers:` present, each entry needs `label` +
  `system` + one of `url`/`id`. Guidance-tier (warn) except where a gate consumes
  it (index reference integrity), where it blocks.
- `identity_errors(team_yaml)`: if `roster:` present, each row needs `name` +
  `role`; `channels:` rows need `name` + `id`. Optional — a `team.yaml` without
  them still validates. Feeds the team node summary.
- Both reserved in `docs/FRONTMATTER-CORE.md`, additively.

### 5. Onboarding split + today wiring (#6, Phase 2)

- Company/platform onboarding: `company-os/onboarding/<role>.md`
  (`type: onboarding-guide`, `role: developer|product-owner|...`).
- Team onboarding: `teams/<t>/onboarding/<role>.md`.
- `cmd_today` gains: after the role view, if a guide matching `args.role` exists
  (team scope first, then company), print `onboarding: <path>`; if none, print
  the role view unchanged, no error. Read-only, additive.

## Constraints

- Single-file CLI, no framework; preserve `die/ok/warn/fail` and the exact
  `frontmatter()` regex (used only for real frontmatter, never for CLAUDE.md
  blocks).
- Python 3.9+, PyYAML only; no new dependencies.
- Generated files land under `generated/` or inside marker blocks; `graph build`
  is the only writer; regenerate-only-when-changed to avoid PR churn.
- Templates are examples, not contracts; adding to them must not turn body
  sections into blocking checks by default.
- Commit the regenerated `examples/workspace` (nodes + index) as part of the
  change, or gate [5/6]/[6/6] fail on fresh checkout.
- Acceptance = `validate` exits 0 on `examples/workspace` AND the standalone-team
  fixture AND `graph build; graph build` is a no-op diff.

## Deferred — demand-driven (designed, not built)

Built only when a concrete adopter needs them; kept here so intent is not lost.

- **#7 data-catalog** — authored `platforms/<p>/reality/data-catalog.yaml`,
  entries `{table, database, owner, description}`, schema-checked, single-source
  (not derived). **Blocked on an undefined surface:** components have no way today
  to declare which warehouse tables they use, so reference integrity (old R-4.3)
  is unimplementable until a `tables:`/pointer convention is defined. Define that
  first when demand arrives.
- **#4 customer-account/call** — scaffold `teams/<t>/product/customers/accounts/
  <account>/{account-context.md, calls/{summaries,transcripts}/}`; new doc types
  participate via frontmatter core; team-private; bodies guidance.
- **#3 workflow-skill** — extend `templates/SKILL-template.md` with
  `kind: workflow`, ordered `steps:`, `reference:` golden output, `cadence`
  split; `workflow_skill_errors` verifies declared step files + reference exist
  (reference integrity, not body format).

## Out of Scope

- Ontology roadmap commands (`validate --ontology`, `spec trace`).
- Cross-team aggregated customer views.
- Any external-system live integration (MCP pulls, Slack posting, Figma reads).
- Migrating Forge example content; we add minimal fixtures under `examples/`.
- Auto-generating onboarding prose or roster values (authored, not derived).
