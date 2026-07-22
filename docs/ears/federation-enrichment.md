---
type: ears
id: ears-federation-enrichment
title: Federation Enrichment — EARS Specifications
status: locked
tags: [kind/ears, status/locked]
---

# Federation Enrichment — EARS Specifications

Keywords: `THE SYSTEM SHALL` (always-on), `WHEN` (event), `WHILE` (during a
state), `IF` (conditional/gate), `WHERE` (context-scoped). "The system" = the
`company-os` CLI and its validate gate. Units are grouped by build phase.
**Committed scope:** Units 0–7. **Deferred (demand-driven, designed not built):**
Units D1–D3.

---

## Unit 0: Idempotency & characterization harness (Phase 0)

**Why:** Two highest-severity risks (byte-symmetry, gate renumber) detonate in
the generator phases in a repo with no test suite. Land the oracle first so every
later phase is written against a red/green check.

| ID | EARS statement |
| --- | --- |
| R-0.1 | THE SYSTEM SHALL compare committed and freshly-derived artifacts by semantic equality — parsed-structure equality for YAML files, canonical-string (stripped, `\n`-normalized) for the CLAUDE.md block — never by raw byte comparison. |
| R-0.2 | THE SYSTEM SHALL NOT write volatile fields (timestamps such as `generatedAt`) into any derived artifact used by a drift gate; IF a volatile field is unavoidable, THE SYSTEM SHALL exclude it from the drift comparison. |
| R-0.3 | THE SYSTEM SHALL produce each derived artifact from a single function, called identically by `graph build` (write) and `validate` (render-and-compare). |
| R-0.4 | WHERE an aggregate generator lists documents, THE SYSTEM SHALL sort documents and dict keys deterministically so output does not depend on filesystem walk order. |
| R-0.5 | THE SYSTEM SHALL exclude every `CLAUDE.md` from `iter_graph_docs` by name, so generated nodes are never re-ingested as graph documents. |
| R-0.6 | WHEN `graph build` runs twice on unchanged input, THE SYSTEM SHALL produce no change on the second run (`git diff --exit-code` clean). |
| R-0.7 | THE SYSTEM SHALL ship a committed acceptance script capturing a golden `validate` stdout snapshot on `examples/workspace` and the double-build no-op check. |
| R-0.8 | WHERE a workspace contains only a team root (no `platforms/`, no `company-os/`), THE SYSTEM SHALL exit 0 on `validate`, and `graph build` SHALL emit only that team's CLAUDE.md node and no feature-index. |

---

## Unit 1: Pointers & frontmatter-core reservations (Phase 1, item #8)

**Why:** The `pointers` shape is the enabler for the feature index's external
column and is reused at every layer. Reserve it and the Phase-2 fields once,
additively, so producers emit them and consumers preserve them (the OKF rule).

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL define one `pointers:` shape (`label`, `system`, and at least one of `url` or `id`) usable in `team.yaml`, component descriptors, PRDs, and other docs. |
| R-1.2 | IF a doc declares a `pointers:` list, THE SYSTEM SHALL require each entry to carry `label`, `system`, and `url` or `id`. |
| R-1.3 | THE SYSTEM SHALL treat pointer well-formedness as guidance-tier (warn) except where a gate consumes a pointer, where it SHALL block. |
| R-1.4 | THE SYSTEM SHALL NOT fetch or mirror external content; it SHALL store references only. |
| R-1.5 | THE SYSTEM SHALL preserve unknown frontmatter fields on every doc it rewrites, never rejecting a doc for carrying extra keys. |
| R-1.6 | THE SYSTEM SHALL document each reserved field and doc type in `docs/FRONTMATTER-CORE.md` without redefining any existing tier, reserving a doc type only in the phase whose consumer ships. |

---

## Unit 2: Feature→artifact index (Phase 1 MVP, item #2)

**Why:** The strongest single idea in the research — replace Forge's
hand-maintained, rot-prone index with a derived, validated map from component to
all its artifacts, on the lowest-risk artifact (a file under `generated/`, never
hand-owned prose).

| ID | EARS statement |
| --- | --- |
| R-2.1 | WHEN `company-os graph build` runs, THE SYSTEM SHALL write `platforms/<p>/generated/feature-index.yaml` for every platform. |
| R-2.2 | THE SYSTEM SHALL key the index by component id and, per component, record its reality doc, active PRDs, archived PRDs, discovery ids, and pending outcomes, resolved from frontmatter reference edges (`components`, `fromDiscovery`, `prd`), with all lists sorted. |
| R-2.3 | THE SYSTEM SHALL collect any `pointers:` from those artifacts into an `externalPointers` list per component. |
| R-2.4 | THE SYSTEM SHALL NOT copy discovery-brief bodies into the index; it SHALL record only the discovery id edge exposed via a PRD's `fromDiscovery`. |
| R-2.5 | WHEN `company-os validate` runs, THE SYSTEM SHALL, as gate [6/6], rebuild each index and FAIL on parsed-structure drift from the committed file with a `run: company-os graph build` remedy. |
| R-2.6 | IF the index is absent for a platform, gate [6/6] SHALL PASS (drift-check-if-present). |
| R-2.7 | IF an index entry references a discovery or PRD id that resolves to no document, THE SYSTEM SHALL FAIL gate [6/6] with the unresolved id named. |

---

## Unit 3: Generated CLAUDE.md context nodes (Phase 2, item #1)

**Why:** Progressive disclosure for agents (Forge proved it at 33 nodes),
derived and drift-checked so it cannot rot, and fail-safe so it never corrupts
hand-owned prose.

| ID | EARS statement |
| --- | --- |
| R-3.1 | WHEN `graph build` runs, THE SYSTEM SHALL write a generated context block into a `CLAUDE.md` at each federation root: `company-os/`, every `platforms/<p>/`, every `teams/<t>/`, and `company-ontology/`. |
| R-3.2 | THE SYSTEM SHALL delimit the block with literal `company-os:generated:start`/`:end` HTML-comment markers and replace only content between them, never using the frontmatter `---` regex. |
| R-3.3 | WHERE a `CLAUDE.md` contains exactly one balanced marker pair, THE SYSTEM SHALL replace only the interior and copy all bytes before start and after end verbatim. |
| R-3.4 | IF a federation root has no `CLAUDE.md`, THE SYSTEM SHALL create one with a minimal hand-owned header and the block. |
| R-3.5 | WHERE a `CLAUDE.md` exists without markers, THE SYSTEM SHALL append the block without altering existing prose, and SHALL NOT fail validation for that marker-less file. |
| R-3.6 | IF markers are missing-as-a-pair, unbalanced, or duplicated, THE SYSTEM SHALL warn and mutate nothing (fail-safe; never partial-write). |
| R-3.7 | THE SYSTEM SHALL render into the block a Doc Index grouped by doc `type` and links to child federation roots, as a pure function of sorted docs and `team.yaml` (no time, no walk-order dependence). |
| R-3.8 | WHERE the root is a team, THE SYSTEM SHALL include a summary of that team's `roster`, `channels`, and `pointers` (Unit 6) in the block. |
| R-3.9 | WHEN `validate` runs, THE SYSTEM SHALL, as gate [5/6], canonical-string-compare each committed block to a fresh render and FAIL on drift with a `run: company-os graph build` remedy. |
| R-3.10 | IF no `CLAUDE.md` node exists at a root, gate [5/6] SHALL PASS (drift-check-if-present; enrichment is opt-in). |

---

## Unit 4: New doc-type kind tags (Phase 1–2)

**Why:** Committed docs (onboarding) must derive `kind/*` tags to participate in
the graph; reserve each type's tag additively, its required-field gate only when
its consumer ships.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL recognise `onboarding-guide` as a valid `type` that derives a `kind/onboarding` tag, requiring `id` and `role` and NOT requiring `status`. |
| R-4.2 | THE SYSTEM SHALL reserve additional doc types (`account-context`, `customer-call`, `data-catalog`) only in the deferred phase that ships their consumer, never activating their required-field gates while inert. |

---

## Unit 5: Team identity block (Phase 2, item #5)

**Why:** A team layer today has only id/name; a usable Team OS answers "who is on
this team and where do we talk" in a queryable shape, and the Phase-2 node must
render it (resolves the R-3.8 dependency).

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL accept optional `roster`, `channels`, and `pointers` blocks in `teams/<t>/team.yaml`. |
| R-5.2 | IF `roster` is present, THE SYSTEM SHALL require each row to carry `name` and `role`; IF `channels` is present, each row SHALL carry `name` and `id`. |
| R-5.3 | THE SYSTEM SHALL keep identity blocks optional — a `team.yaml` without them SHALL still validate. |
| R-5.4 | THE SYSTEM SHALL reserve `roster`/`channels`/`pointers` field names in `docs/FRONTMATTER-CORE.md` so they are uniformly queryable across teams. |
| R-5.5 | WHEN the team CLAUDE.md node is rendered (R-3.8), THE SYSTEM SHALL source its identity summary from these blocks. |

---

## Unit 6: Onboarding split & role wiring (Phase 2, item #6)

**Why:** Company-level onboarding (governance/tiers/lifecycle) and team-level
onboarding (this team's quirks) are different content for different readers;
`today --role` already knows the reader's role.

| ID | EARS statement |
| --- | --- |
| R-6.1 | THE SYSTEM SHALL support `onboarding-guide` docs at `company-os/onboarding/<role>.md` and `teams/<t>/onboarding/<role>.md`. |
| R-6.2 | WHEN `company-os today --role <r>` runs, THE SYSTEM SHALL, after the role view, print a pointer to the onboarding guide matching `<r>` if one exists, checking team scope before company scope. |
| R-6.3 | IF no onboarding guide matches the role, THE SYSTEM SHALL print the role view unchanged with no error. |
| R-6.4 | THE SYSTEM SHALL derive an onboarding guide's tags from its `role` and scope, and its body SHALL remain guidance. |

---

## Unit 7: Acceptance & backward compatibility (all committed phases)

**Why:** The project's only acceptance path is a green validate; new mechanisms
must not break existing adopters, and the standalone-team promise must be tested.

| ID | EARS statement |
| --- | --- |
| R-7.1 | THE SYSTEM SHALL keep `company-os validate` exiting 0 on `examples/workspace` after each phase. |
| R-7.2 | THE SYSTEM SHALL extend `examples/workspace` with fixtures exercising each committed unit (index, pointers, nodes, identity, onboarding) and SHALL commit the regenerated nodes + index. |
| R-7.3 | THE SYSTEM SHALL preserve the semantics of existing gates [1/4]–[4/4], renumbered under a 6-gate run, verified against a golden stdout snapshot, with no behavior change. |
| R-7.4 | THE SYSTEM SHALL keep the single-file CLI, the `die/ok/warn/fail` helpers, the `frontmatter()` parser, and the next-command guidance chain intact. |
| R-7.5 | WHERE any doc (root `CLAUDE.md`, `TUTORIAL.md`, `docs/01`) states validate has 3 steps, THE SYSTEM SHALL correct it to the actual gate count in the same change. |

---

# Deferred — demand-driven (designed, NOT built in this initiative)

These units are retained so intent is not lost. They are implemented only when a
concrete adopter needs them, each behind its own reserve-in-consuming-phase rule.

## Unit D1: Data-catalog registry (item #7)

**Why:** Give warehouse-bearing platforms a validated table registry. **Blocked
on an undefined surface** — components have no way today to declare which tables
they use, so reference integrity is unimplementable until that convention exists.

| ID | EARS statement |
| --- | --- |
| R-D1.1 | THE SYSTEM SHALL accept an authored `platforms/<p>/reality/data-catalog.yaml` whose entries each declare `table`, `database`, `owner`, and `description`, treated as single-source (not derived, never overwritten by `graph build`). |
| R-D1.2 | IF an entry is missing a required field, THE SYSTEM SHALL FAIL validation naming the entry and field. |
| R-D1.3 | Before implementation, THE SYSTEM SHALL define how a component/artifact references a warehouse table; only then MAY reference integrity (absent-table → FAIL) be built. |

## Unit D2: Customer-account & call structure (item #4)

**Why:** Give team-private discovery a structured evidence layer to cite.

| ID | EARS statement |
| --- | --- |
| R-D2.1 | THE SYSTEM SHALL scaffold, under `teams/<t>/product/customers/accounts/<account>/`, an `account-context.md` and `calls/{summaries,transcripts}/`. |
| R-D2.2 | THE SYSTEM SHALL treat account and call docs as graph participants via frontmatter core, keeping them team-private (never surfaced into platform-level generated files) and their bodies guidance. |

## Unit D3: Workflow-skill orchestration pattern (item #3)

**Why:** Upgrade canonical skills to Forge's repeatable-process shape with the
linkage validated, not the prose.

| ID | EARS statement |
| --- | --- |
| R-D3.1 | THE SYSTEM SHALL extend `templates/SKILL-template.md` to express `kind: workflow` with ordered `steps:`, a `reference:` golden-output path, and a `cadence` split of `changesEveryCycle` vs `stable`. |
| R-D3.2 | WHERE a skill declares `kind: workflow`, THE SYSTEM SHALL verify every `steps[].file` and the `reference` path exist relative to the skill directory, FAIL naming any missing path, and leave bodies as guidance. |
