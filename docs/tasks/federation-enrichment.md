---
type: tasks
id: tasks-federation-enrichment
title: Federation Enrichment — Tasks
status: draft
tags: [kind/tasks, status/draft]
---

# Federation Enrichment — Tasks

Source of truth: `docs/ears/federation-enrichment.md` (Units 0–7 committed).
Architecture constraints: `docs/lld/federation-enrichment.md`.
Target: single-file CLI `company-os-starter/bin/company-os`.

**Global acceptance (must hold after every phase):** `company-os validate`
exits 0 on `examples/workspace` **and** the new standalone-team fixture, **and**
`graph build; graph build` is a no-op diff.

**Deferred — NOT planned here:** EARS Units D1 (data-catalog #7), D2
(customer/call #4), D3 (workflow-skill #3). Build only on concrete demand;
D1 additionally blocked on defining the component→table reference surface.

**Ordering rule:** Phase 0 (Unit 0) lands before any generator. Phase 1 (Units
1–2) is the MVP. Phase 2 (Units 3–6) layers CLAUDE.md nodes + identity +
onboarding. Gate numbering churn (4→5→6 header labels) is expected and permitted
by the golden snapshot; tasks reference R-ids, not literal `[n/m]` strings.

---

## Phase 0 — Idempotency & characterization harness (Unit 0)

- [x] 0.1 Shared canonicalizer + semantic-compare helpers (est: ~40m)
  - why: The whole "derived + validated" altitude fails on day one if validate
    re-derives bytes that differ from what graph build wrote (pre-mortem R#1).
    A single compare path, parsed-structure for YAML and canonical-string for
    the block, with volatile fields excluded, is the foundation every later gate
    stands on.
  - acceptance: R-0.1 — compare by parsed-structure (YAML) / canonical-string
    (block), never raw bytes; R-0.2 — no volatile fields in derived artifacts
    (or masked from compare); R-0.3 — one function renders each artifact, called
    by both graph build and validate.
  - verify: unit-exercise `canonical_yaml()`/`blocks_equal()` on hand-crafted
    equal-but-reordered dicts and whitespace-varied blocks; both report equal.

- [x] 0.2 Deterministic ordering for aggregate consumers (deps: none, mutex: iter_graph_docs, est: ~20m)
  - why: `root.rglob("*.md")` is unsorted, so any aggregate (node, index) would
    drift across checkouts even with a canonicalizer. Sorting doc lists + dict
    keys is what makes idempotency real, not just careful serialization.
  - acceptance: R-0.4 — aggregate generators sort documents and dict keys so
    output is walk-order independent.
  - verify: build an aggregate over a fixture twice with shuffled filesystem
    order (or a stubbed unsorted walk); outputs are byte-identical.

- [x] 0.3 Hard `CLAUDE.md` name-skip in `iter_graph_docs` (deps: none, mutex: iter_graph_docs, est: ~10m)
  - why: Today a CLAUDE.md is spared graph ingestion only by the accident of
    having no frontmatter; the moment a node gains YAML it would be re-tagged and
    fight its own generator. Make the skip explicit (R-0.5) before nodes exist.
  - acceptance: R-0.5 — `CLAUDE.md` excluded from `iter_graph_docs` by name.
  - verify: add a CLAUDE.md with frontmatter to a fixture; `graph build` does not
    tag it and `validate` gate [4] does not iterate it.

- [x] 0.4 Characterization harness: golden snapshot + double-build no-op (deps: 0.2, 0.3, est: ~30m)
  - why: With no test suite, this script IS the safety net — it freezes gates
    [1–4] behavior and is the direct oracle for derive/validate asymmetry every
    later phase is written against.
  - acceptance: R-0.6 — `graph build; graph build; git diff --exit-code` clean;
    R-0.7 — committed script captures a golden `validate` stdout snapshot on
    `examples/workspace` and runs the double-build check.
  - verify: run the script on the untouched repo — snapshot matches, no-op diff
    is clean, exit 0.

- [x] 0.5 Standalone-team fixture validates green (deps: 0.3, 0.4, est: ~30m)
  - why: The central product promise is "a team adopts independently." It is
    currently untested — every fixture has a platform and company. Prove a
    platform-less, company-less team is valid before gates that could break it.
  - acceptance: R-0.8 — a workspace with only a team root exits 0 on `validate`;
    `graph build` emits only that team's node (Phase 2) and no feature-index.
  - verify: `examples/standalone-team/` validates green under current gates and
    stays green after each later phase (re-run in 2.3, 3.5, 7.1).

---

## Phase 1 — MVP: pointers + derived feature index (Units 1–2)

- [x] 1.1 `pointers:` shape + `pointer_errors()` (deps: 0.1, est: ~30m)
  - why: Pointers are the fractal delegation convention (#8) and the enabler for
    the index's external column and future evidence links — store references,
    never mirror external content.
  - acceptance: R-1.1/R-1.2 — one `pointers` shape (`label`+`system`+`url`|`id`)
    usable across doc kinds, each entry validated; R-1.3 — guidance-tier (warn)
    except where a gate consumes it; R-1.4 — references only, no fetching.
  - verify: a well-formed pointer validates; a missing-`system` pointer warns at
    exit 0; no network call exists in the path.

- [x] 1.2 Preserve-unknown-fields guard + FRONTMATTER-CORE additions (deps: 1.1, est: ~25m)
  - why: OKF interop depends on producers extending metadata and consumers never
    dropping it; new reserved fields must be documented without redefining tiers.
  - acceptance: R-1.5 — rewrites preserve unknown frontmatter keys; R-1.6 —
    `docs/FRONTMATTER-CORE.md` documents `pointers` (and reserves a doc type only
    in the phase its consumer ships).
  - verify: round-trip a doc with an unknown key through `graph build`; the key
    survives. FRONTMATTER-CORE renders with the new section.

- [x] 2.1 `build_feature_index()` generator + graph build wiring (deps: 0.1, 0.2, 1.1, est: ~90m)
  - why: The crown-jewel item (#2) — a derived, always-current map from component
    to all its artifacts, replacing the hand-maintained index that rotted in the
    Forge example. Largest new surface, so it lands first behind the harness.
  - acceptance: R-2.1 — writes `platforms/<p>/generated/feature-index.yaml`;
    R-2.2 — keyed by component id from `components`/`fromDiscovery`/`prd` edges,
    lists sorted; R-2.3 — collects `pointers` into `externalPointers`; R-2.4 —
    records discovery id edge only, never brief bodies; no `generatedAt`.
  - verify: run on `examples/workspace`; index matches the known component graph;
    `graph build; graph build` no-op (0.4 script passes).

- [x] 2.2 Gate for feature-index drift, absence-tolerant (deps: 2.1, 0.1, est: ~45m)
  - why: A derived file with no gate rots like Forge's did; a gate that fails on
    absence would force adoption and break the standalone promise. Drift-if-
    present is the exact embodiment of "opt-in enrichment."
  - acceptance: R-2.5 — rebuild + parsed-equality compare, FAIL on drift with a
    `run: company-os graph build` remedy; R-2.6 — absent index → PASS; R-2.7 —
    unresolved discovery/prd id → FAIL naming the id.
  - verify: hand-edit the committed index → validate FAILs with remedy; delete it
    → validate PASSes; standalone-team fixture (0.5) still green.

---

## Phase 2 — CLAUDE.md nodes, identity, onboarding (Units 3–6)

- [x] 3.1 Fail-safe `rewrite_generated_block()` (deps: 0.1, est: ~60m)
  - why: One greedy regex away from eating a user's hand-written CLAUDE.md
    (pre-mortem R#2). Must refuse rather than guess — a single data-loss incident
    and nobody runs the generator again.
  - acceptance: R-3.2 — literal-sentinel match, never the frontmatter `---`
    regex; R-3.3 — replace interior only, verbatim outside; R-3.4 — create when
    absent; R-3.5 — append (never overwrite) when prose exists without markers,
    and don't fail such files; R-3.6 — warn + mutate nothing on unbalanced/
    duplicated markers.
  - verify: fixtures for each marker state (none+absent, none+prose, one pair,
    unbalanced) behave per spec; rewriting an identical block yields an identical
    file.

- [x] 3.2 `build_claude_node()` pure render + graph build wiring (deps: 0.2, 3.1, est: ~75m)
  - why: Progressive-disclosure context for agents at every root (#1), Forge's
    33-node mechanism, but derived and pure so it never drifts.
  - acceptance: R-3.1 — write a block at each federation root; R-3.7 — Doc Index
    grouped by `type` + child links, a pure function of sorted docs + team.yaml
    (no time, no walk-order dependence).
  - verify: nodes render on `examples/workspace`; double-build no-op holds.

- [x] 3.3 Team identity block + `identity_errors()` (deps: 0.1, est: ~40m)
  - why: A team layer today has only id/name; a usable Team OS answers "who is on
    this team and where do we talk" in a queryable shape — and the node needs it.
  - acceptance: R-5.1 — optional `roster`/`channels`/`pointers` in `team.yaml`;
    R-5.2 — roster rows need `name`+`role`, channels rows `name`+`id`; R-5.3 —
    omitting them still validates; R-5.4 — fields reserved in FRONTMATTER-CORE.
  - verify: a team with a valid roster validates; a roster row missing `role`
    fails; a team.yaml without identity validates.

- [x] 3.4 Node renders identity summary (deps: 3.2, 3.3, est: ~20m)
  - why: Resolves the plan inversion the PO caught — the Phase-2 node must source
    its identity summary from the identity block that now lands in the same phase.
  - acceptance: R-3.8 / R-5.5 — a team node includes the roster/channels/pointers
    summary from `team.yaml`.
  - verify: a team with identity shows the summary in its generated block;
    changing the roster and re-building updates the summary.

- [x] 3.5 Gate for CLAUDE.md node drift, absence-tolerant (deps: 3.2, 0.1, est: ~40m)
  - why: Same drift-or-rot logic as the index gate, but on the higher-risk
    hand-owned file — so absence-tolerance matters even more.
  - acceptance: R-3.9 — canonical-string compare each block, FAIL on drift with
    remedy; R-3.10 — absent node → PASS.
  - verify: in-block hand-edit → FAIL; delete the node → PASS; standalone-team
    fixture (0.5) still green; gates [1–4] snapshot unchanged.

- [x] 4.1 `onboarding-guide` doc type + `today --role` pointer (deps: none, est: ~45m)
  - why: Company-level (governance/tiers) and team-level (this team's quirks)
    onboarding are different content for different readers; `today --role`
    already knows the role, so surface the right guide there.
  - acceptance: R-4.1 — `onboarding-guide` type derives `kind/onboarding`,
    requires `id`+`role`, no `status`; R-6.1 — guides at `company-os/onboarding/`
    and `teams/<t>/onboarding/`; R-6.2/R-6.3 — `today --role` prints the matching
    guide (team before company) or the role view unchanged with no error.
  - verify: with a matching guide, `today --role developer` prints the pointer;
    with none, it prints the role view and exits 0.

- [x] 4.2 Reserve-in-consuming-phase guard for deferred types (deps: 4.1, est: ~15m)
  - why: A reserved-but-unconsumed doc type is worse than absence — adopters
    author docs that "validate" but do nothing. Keep deferred types inert until
    their consumer ships.
  - acceptance: R-4.2 — `account-context`/`customer-call`/`data-catalog` are not
    given active required-field gates in committed scope.
  - verify: grep the CLI for those type strings — no required-field enforcement
    is wired; only their future phase would add it.

---

## Phase 7 — Acceptance & backward compatibility (Unit 7, throughout)

- [x] 7.1 Example fixtures per committed unit + commit regenerated workspace (deps: 2.1, 3.2, 3.3, 4.1, est: ~50m)
  - why: The only acceptance path is a green validate; the generated nodes+index
    must be committed or gate [5/6]/[6/6] fail on fresh checkout.
  - acceptance: R-7.1 — validate exits 0 on `examples/workspace` each phase;
    R-7.2 — fixtures exercise index, pointers, nodes, identity, onboarding, and
    the regenerated nodes + index are committed.
  - verify: fresh checkout → `validate` exits 0 with no `graph build` needed.

- [x] 7.2 Freeze gates [1–4] + preserve CLI conventions (deps: 2.2, 3.5, est: ~20m)
  - why: New gates must not perturb existing behavior or the single-file
    structure adopters depend on.
  - acceptance: R-7.3 — gates [1/4]–[4/4] identical output (modulo header) vs the
    golden snapshot; R-7.4 — `die/ok/warn/fail`, `frontmatter()`, and the
    next-command guidance chain intact.
  - verify: diff validate output against 0.4 snapshot — only gate-count header
    lines moved; CLI helpers unchanged.

- [x] 7.3 Correct stale gate-count docs (deps: none, est: ~15m)
  - why: The repo already suffers from docs that misdescribe the CLI; shipping
    new gates without fixing "3-step" text compounds it.
  - acceptance: R-7.5 — root `CLAUDE.md`, `TUTORIAL.md`, `docs/01` corrected to
    the actual gate count in this change.
  - verify: grep for "3-step"/"[n/3]" in those files returns nothing stale.
