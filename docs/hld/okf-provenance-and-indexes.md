---
type: hld
id: hld-okf-provenance-and-indexes
title: OKF Provenance and Indexes — High-Level Design
status: draft
---

# OKF Provenance and Indexes — High-Level Design

## Overview

Promotes the two units deferred from `docs/ears/okf-v02-conformance.md` into a
buildable change: **D2** (`description:` frontmatter plus generated per-directory
`index.md`) and **D1** (`generated:` / `verified:` provenance and a derived trust
tier).

These are the two items from section 5 of
`.devlocal/research/2026-07-25-okf-v02-vs-company-os-comparison.md` that were
designed but never built. The other six items of that section are already covered
by the parent EARS (Units 1–3) or recorded there as decided non-goals N3 and N4.
This document does not restate them.

The parent change writes the conformance contract. This one adds the two fields
that contract recommends but no shipped document carries, and gives the first of
them a consumer.

**Relationship to the parent:** `docs/ears/okf-v02-conformance.md` §"Deferred —
decided, not built" is the source. D-1.1–D-1.4 and D-2.1–D-2.8 are lifted here and
made normative. Task 1.6 of the parent reserves `generated:` and `verified:` as
inert names so this change lands without a rename; that reservation is a
precondition, not a dependency to duplicate.

## Stakeholders & Impact

**Agents working inside a workspace** are the primary consumer. An agent opening
`platforms/communications/reality/` today must `ls` the directory and open files
to learn what is there — there is no `index.md` anywhere in any fixture. After
this ships it reads one file listing each document with a one-sentence
description. This is the progressive-disclosure argument, and it applies harder to
Company OS than to a flat OKF bundle because the trees are deep.

**Reviewers and approvers** gain an answer to a question the system cannot
currently answer: was this PRD drafted by an agent or written by a person? The
entire `skills/` layer exists to direct agents producing PRDs and discovery
briefs, and nothing records that fact. Trust today is inferred from three
incompatible fields.

**Adopters running `company-os init`** must not receive a workspace that fails its
own `validate`. This is the failure mode D-2.2 exists to prevent and it constrains
where index generation is wired.

**Federated consumers** are affected only negatively if this is done wrong.
`examples/federated/` materializes graph documents inside a read-only `0444` slice
whose bytes are hashed into `workspace.lock.yaml`. Writing an index there breaks
gate `[8/8]`.

## Goals

1. Every directory holding two or more graph documents carries a generated
   `index.md` listing each document with its title and description.
2. `description:` is documented, emitted by scaffolding, and backfilled across the
   two monorepo fixtures — with a rubric strong enough that descriptions are not
   filename restatements.
3. `generated:` and `verified:` are live fields: validated, preserved through
   `derive`, and emitted where the producer is known.
4. A trust tier is derivable from `verified:` — absent means unverified, non-human
   actors only means machine-confirmed, any `human:` actor means human-reviewed —
   and is advisory only, never access control.
5. Existing approval fields keep their exact current meaning, and a `verified:`
   entry is derived from them one-way so the trust tier is meaningful on day one
   without editing a single fixture's approval data.
6. `company-os init` still produces a workspace that passes `validate`.
7. Two consecutive `derive` runs leave the workspace byte-identical.

## Non-Goals

- **N1 — Reconciling `decisionOwner` and `approvedBy` into `verified:`.** Derivation
  is one-way and read-only. The legacy fields stay authoritative and unchanged.
  Full reconciliation is D-1.4 of the parent and remains deferred, because
  unifying them changes approval semantics.
- **N2 — Closing the exception/PRD `TODO` asymmetry as a blocking check.**
  `internal/scaffold/template.go:67` emits `decisionOwner: TODO` which hard-fails
  `prd validate`; `internal/governance/declare.go:121` emits
  `approvedBy: 'TODO: rule owner'` which passes. This change surfaces the second at
  **warn tier only**. Making it block would fail workspaces that pass today,
  including the shipped `examples/workspace` fixture.
- **N3 — `stale_after:`.** Unchanged from the parent's N4. Still no consumer
  surface.
- **N4 — OKF `sources[]` and Attested Computation.** Unchanged from the parent's
  N2 and N3.
- **N5 — Making `description:`, `title:`, or `resource:` required.** They stay
  recommended. Unchanged from the parent's N5.
- **N6 — Generating indexes inside `knowledge/`.** It is a node root but not a
  graph-docs root (`internal/graph/tags.go:191`). Its slices are foreign,
  `0444`, and carry no `type:` frontmatter.
- **N7 — Changing gate numbering.** Index drift reports inside gate 5. No gate is
  added and none is renumbered.

## Success Criteria

1. `examples/workspace/` contains exactly three generated `index.md` files;
   `examples/standalone-team/` contains zero. Both still pass `validate`.
2. `git diff --stat examples/federated` is empty, and
   `company-os --root examples/federated validate` exits 0 at `[8/8]`.
3. Every document in the two monorepo fixtures that carries frontmatter `type:`
   also carries a `description:` satisfying the rubric.
4. A document carrying `generated: {by, at}` and `verified: [{by, at}]` validates
   and survives `derive` byte-identical, covered by a Go test written against
   those nested shapes rather than a string scalar.
5. `company-os init` followed immediately by `company-os validate` exits 0.
6. `make check` exits 0.
7. Both golden snapshots are regenerated **deliberately and reviewed line by
   line** — this change moves them, unlike the parent change which froze them.

## Invariants

Inherited from the parent HLD. A change violating any is rejected regardless of
merit.

- **I1. Backward compatibility.** No workspace conformant before this fails after,
  beyond a documented `derive` re-run. A new warn line is permitted; a new
  blocking check is not.
- **I2. Absence tolerance.** `examples/standalone-team/` keeps passing with three
  of four federation roots missing.
- **I3. The tier model.** mandatory / default / guidance untouched. Nothing here
  becomes mandatory-tier.
- **I4. Generated files are derived, never hand-edited.** `index.md` joins that
  set.
- **I5. Gate 1–7 numbering and printed strings are frozen.** Gate 5's header stays
  byte-identical.
- **I6. Idempotency.** Two consecutive `derive` runs leave the workspace
  byte-identical.
- **I7. Unknown fields and unknown types are preserved, never rejected.**
- **I9. Federated slices are read-only derived content.** Nothing writes into a
  materialized slice or invalidates `workspace.lock.yaml` hashes.

**Note on R-5.6.** The parent change required both goldens byte-identical and
forbade `acceptance.sh --update`. That constraint belonged to the parent, whose
units were documentation-only. This change adds output lines to gate 5 and new
files to two fixtures, so the goldens necessarily move. They are re-baselined once,
deliberately, with the diff reviewed — not refreshed reflexively on a red run.
