---
type: hld
id: hld-federation-enrichment
title: Federation Enrichment — porting Team OS mechanisms into Uncle OS
status: locked
tags: [kind/hld, status/locked]
---

# Federation Enrichment — High-Level Design

## Overview

Port mechanisms proven in the public Team OS example repo (Forge) into
`company-os-starter/`, each placed at the federation layer where it naturally
lives (company / platform / team), implemented at the **derived + validated**
altitude: generators emit structure and indexes, and `company-os validate`
drift-checks them the same way it already drift-checks tags. Where an item is
authored context rather than derivable (rosters, onboarding prose), the
"validated" part is *participation in the graph* — reserved frontmatter-core
fields — while the document body stays team-local guidance.

Eight candidate mechanisms were classified (2026-07-21 comparison research, §4).
Following PO/Lead review, **five are committed** and **three are deferred to
demand-driven** (designed here, built only when a real adopter needs them):

| # | Mechanism | Layer | Status | Shape |
|---|---|---|---|---|
| 2 | Feature→artifact index | platform | **Committed (MVP)** | Derived `generated/feature-index.yaml` |
| 8 | External-system delegation (`pointers`) | all layers (fractal) | **Committed (MVP)** | Frontmatter `pointers:` shape + core reservation |
| 1 | Nested CLAUDE.md routing chain | company + platform + team | **Committed** | Derived marker-block per root |
| 5 | Team identity block (roster/handles/channels) | team | **Committed** | `team.yaml` schema extension + core reservation |
| 6 | Onboarding guides + `today --role` pointer | company + team | **Committed (ride-along)** | New doc type + `today` wiring |
| 7 | data-catalog.yaml registry | platform (reality) | **Deferred — demand-driven** | Authored registry + schema/reference validation |
| 4 | Customer-account + call-summary structure | team | **Deferred — demand-driven** | Doc types + scaffold |
| 3 | Workflow orchestration pattern | company (skill contract) | **Deferred — demand-driven** | Template + reference-integrity validation |

Deferral rationale (PO, value grounds): #3/#4/#7 are the most Forge-specific,
carry the largest new contract surface, reach the narrowest set of adopters
(a platform with no warehouse has no tables; an infra team runs no customer
calls), and their "validated" content is thin. They stay in this HLD as designed
intent but are built only on real demand — never as speculative surface.

## Stakeholders & Impact

- **Adopting teams (standalone Team OS):** today a team layer carries only
  governance + discovery + standards. After this ships, a team can hold its
  identity (roster, channels) and onboarding, and gain its own CLAUDE.md node —
  becoming a usable "shared brain" without a company or platform above it. A
  platform-less, company-less team must validate green (new first-class
  acceptance test, see Success Criteria).
- **Platform / catalog owners:** gain a machine-readable, validated
  feature→artifact index — the MVP — replacing the hand-maintained
  `feature-index.yaml` that rotted in the Forge example.
- **AI agents working the workspace:** gain nested CLAUDE.md context nodes for
  progressive disclosure (Forge proved this at 33 nodes).
- **The `company-os` CLI + `validate` CI gate:** gain generators (`graph build`
  extensions) and new drift gates; the exit-code contract still holds.
- **Secondary consumers (Obsidian / OKF tools):** new reserved fields
  (`pointers`, `roster`, doc `type`s) extend the graph additively; the
  "preserve unknown fields" rule is preserved.

## Goals

- Every committed mechanism lands as **structure + derived files** (validatable),
  never prose-only convention — answering the failure the Forge example
  demonstrates (phantom paths, broken links, free-text status with nothing
  checking it).
- A standalone team workspace (no platform, no company) **remains valid and
  gains real utility**, and this is proven by a dedicated acceptance fixture.
- Generated artifacts are **derived, never hand-edited in their generated
  region**, and validate fails on drift — identical doctrine to today's `tags:`.
- **Drift-check-if-present, never require-presence:** every new gate passes when
  its artifact is absent. Enrichment is opt-in; the OS never forces scaffolding.
- Generators are **fail-safe**: they refuse to touch a hand-owned file with
  missing/unbalanced markers rather than corrupt it.
- The frontmatter-core contract is extended, not redefined: new fields are
  additive, unknown-field preservation holds.
- `company-os validate` exits 0 on `examples/workspace` after every phase, and
  `graph build; graph build` is a no-op diff.

## Non-Goals

- No change to the existing four validate gates' semantics; new gates are
  appended (count moves 4→6). Root `CLAUDE.md`'s stale "3-step" text is corrected.
- No implementation of the roadmap ontology commands (`validate --ontology`,
  `spec trace`, EARS `ears:` blocks).
- No mirroring of external system *content* — tickets, Figma, Slack stay
  external; we store **pointers** only (#8).
- No new runtime dependencies (Python 3.9+ / PyYAML only).
- No forced adoption: identity, onboarding, and all deferred items are opt-in;
  a team omitting them still validates.
- Document body/section format stays guidance (opt-in blocking via the existing
  `standards/doc-formats.yaml`).
- No building of #3/#4/#7 in this initiative unless a concrete adopter asks; the
  R-4.3 table-reference surface (#7) is explicitly undefined until then.

## Success Criteria

Split per mechanism so each is atomically acceptance-testable:

1. **Idempotency (semantic).** After `graph build`, a second `graph build`
   produces no diff, and `validate` re-derives every artifact and matches the
   committed copy under a *semantic* compare (parsed-structure equality for
   YAML, canonical-string for the CLAUDE.md block, volatile fields excluded).
   No timestamp or walk-order difference can cause drift.
2. **Standalone team validates.** A workspace containing only a team root (no
   `platforms/`, no `company-os/`) exits 0 on `validate`; `graph build` emits
   only that team's CLAUDE.md node and no feature-index.
3. **Feature index (#2, MVP).** `graph build` writes
   `platforms/<p>/generated/feature-index.yaml` keyed by component id from
   frontmatter edges; gate [6/6] FAILs on a hand-edit with a
   `run: company-os graph build` remedy, and PASSES when the file is absent.
4. **Pointers (#8).** A well-formed `pointers:` entry validates; a malformed one
   warns (guidance) except where a gate consumes it, where it blocks.
5. **CLAUDE.md nodes (#1).** `graph build` writes a marker-delimited block at
   each federation root, preserving all prose outside the markers; gate [5/6]
   FAILs on an in-block hand-edit and PASSES when no node exists; the generator
   refuses to write a file whose markers are missing/unbalanced.
6. **Team identity (#5).** A team declaring `roster`/`channels` validates and
   its node renders the identity summary; a team omitting them still validates.
7. **Onboarding (#6).** `today --role <r>` prints a pointer to a matching guide
   when one exists (team scope before company), and prints the role view
   unchanged with no error when none exists.
8. **Backward compatibility.** Gates [1/4]–[4/4] produce identical output
   (modulo the renumbered header) on `examples/workspace`, verified by a golden
   snapshot; the regenerated `examples/workspace` (nodes + index) is committed.

## Build Phases

- **Phase 0 — Characterization harness (de-risk first).** Before any generator:
  capture a golden `validate` stdout snapshot on `examples/workspace`; add the
  `graph build; graph build; git diff --exit-code` no-op check; add a
  standalone-team (no platform/company) fixture that must validate green; land
  the shared canonicalizer + parsed-compare decision, deterministic sorted
  iteration, and the hard `CLAUDE.md` name-skip. This is the oracle every later
  phase is written against.
- **Phase 1 — MVP: derived index (lowest risk, highest value).** `pointers:`
  shape (#8) + only the frontmatter-core fields Phase 2 will consume; the
  feature-index generator (#2) as a file under `generated/` (never touches
  hand-owned prose); one drift gate [6/6] with absence-tolerance. Proves the
  entire derived+validated altitude on the safest artifact.
- **Phase 2 — CLAUDE.md nodes + identity.** Generated marker-block nodes (#1)
  with fail-safe rewrite and gate [5/6]; team identity (#5) moved here so the
  node can render its roster/channels summary (resolves the R-2.6 inversion);
  onboarding split + `today --role` pointer (#6) as a ride-along.
- **Deferred — demand-driven (not built unless requested):** data-catalog
  registry (#7, requires defining the table-reference surface first),
  customer-account/call structure (#4), workflow-skill pattern (#3).
