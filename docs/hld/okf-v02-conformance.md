---
type: hld
id: hld-okf-v02-conformance
title: OKF v0.2 Conformance — High-Level Design
status: locked
tags: [kind/hld, status/locked]
---

# OKF v0.2 Conformance — High-Level Design

## Overview

Company OS declares the Google Open Knowledge Format as its documentation
standard — `company-os-starter/docs/00-original-proposal.md:3` ("Documentation
standard: Google Open Knowledge Format (OKF) v0.1") and `:29` ("Every Company
OS, Platform OS, and Team OS repository should be an **OKF Knowledge Bundle**").
`bin/company-os:983` labels validate gate 4 "the OKF/Obsidian interop contract".

The implementation has drifted from that claim in two directions at once. It
**subtracted**: the shipped example workspaces contain zero `index.md`, zero
`description:`, and zero `resource:` — despite the proposal's own minimum-concept
example carrying `resource: component://customer-notification-service`
(`00-original-proposal.md:170`). And it **froze**: OKF has since published v0.2,
which supersedes v0.1, retires `timestamp:`, and adds five frontmatter families.

Separately, Company OS has never written down what "conformant" means. There is
no document using MUST/SHOULD/MAY, no conformance clause, and no methodology
version — while eight fixtures commit
`conformance: {companyOsVersion: '2026.2', profile: …}`, naming a version defined
nowhere and, in two cases, a profile value (`minimal`) outside the only documented
enum (`docs/01-flexibility-skills-and-role-views.md:121`).

This change closes the definitional gap. It writes the normative conformance and
versioning document, separates shipped capability from roadmap in the ontology
guide, adopts the OKF-recommended fields Company OS dropped, and fixes a
date-comparison defect in the `prd complete` done-gate.

**Scope was cut after technical and product review.** Per-directory `index.md`
generation — originally U4 — is deferred to its own change, because it collides
with federation mode in a way that requires design work this spec does not
contain (see Deferred). `description:` ships with it rather than here, since its
only consumer is the index and its quality cannot be reviewed without one.

Source analysis: `.devlocal/research/2026-07-25-okf-v02-vs-company-os-comparison.md`.

## Stakeholders & Impact

**The solo-team adopter** (`examples/standalone-team/` — one team, three markdown
files, three of four federation roots absent). Their pain is not navigation;
three files need no index. It is that "am I doing this right?" has no answer,
because no document defines a conformant workspace. The on-ramp promise in
`examples/standalone-team/EXAMPLE_README.md` — start solo, later join a federation
without restructuring — is asserted, not specified. After this ships they can read
one conformance clause that says the floor is `type` + `id`. They must see **no
new failures and no new warnings**, and no new generated files whatsoever.

**The federation adopter** (`examples/workspace/`, `examples/banking/`). Their
`CLAUDE.md` context nodes currently list documents by `id` or filename, because 12
of 17 fixture documents carry no `title:` (`build_claude_node:1682` falls back
`title or id or filename`). After this ships those nodes read as titles. Their
migration cost is one `graph build` re-run, documented.

**AI agents, as consumers.** Better `CLAUDE.md` doc indexes and a
machine-followable `resource:` URI on reality documents, replacing
filename-to-component inference. The larger navigation win — one index per
directory — arrives with the deferred change.

**AI agents, as producers.** The `skills/` layer exists to direct agents producing
PRDs, briefs, and reality docs. Nothing records that a PRD was drafted by an agent
rather than a person. **This change does not fix that** — it reserves the field
names and defers the semantics. Stated plainly so no reader infers otherwise.

**CI pipelines running `company-os validate`.** A constraint, not an opportunity.
A workspace that exits 0 at commit N-1 must exit 0 at commit N. The kit's own
oracle — `examples/acceptance.sh`, with `examples/selftest.py` and two frozen
golden snapshots — is the verification surface for every criterion here.

**Maintainers of the kit.** `docs/ONTOLOGY-GUIDE.md` (368 lines) documents
`spec trace`, `validate --ontology`, vocabulary linting, `ears:` requirement
blocks, `## Graph` wikilink blocks, and per-clause PRD checklists. Grep of
`bin/company-os` confirms zero occurrences of `ears`, `@spec`, or `spec trace`.
Worse, two shipped fixtures assert the unshipped command as fact —
`examples/workspace/company-ontology/contexts/communications.md:20` and the
federated twin both say "`company-os validate --ontology` flags forbidden terms in
canonical docs", while the banking fixture correctly writes "(`validate
--ontology`, roadmap)". Fixtures are what adopters copy.

## Goals

1. **"Conformant" has a written definition.** A reader can point at one document
   and one section and state what a Company OS document MUST carry, what it
   SHOULD carry, and what tooling MUST NOT reject on.
2. **The methodology has a version, and it is the version already committed in
   fixtures.** `2026.2` and every `profile:` value in the repo resolve to
   definitions, and no shipped fixture is non-conformant against the new document
   on the day it ships.
3. **Every documented CLI capability is either runnable today or visibly labelled
   as not-yet-shipped** — in the docs *and* in the fixtures adopters copy.
4. **Every reality document points at the asset it describes with a
   machine-followable URI**, drawn from the vocabulary registered in
   `company-ontology/ids/registry.yaml`.
5. **Generated `CLAUDE.md` context nodes name documents by title** rather than by
   id or filename.
6. **The `prd complete` done-gate compares dates as dates.** A malformed
   `updated:` or `created:` produces a clear failure naming the file, not a silent
   wrong answer.
7. **No workspace that passed `validate` before this change fails after it**,
   beyond the one documented `graph build` re-run.
8. **The kit's own CI gate is green**, and the golden snapshots are unchanged —
   this change re-baselines nothing.

## Non-Goals

Decided "no"s, recorded so they do not read as oversights. They belong in the
conformance document's "considered and deferred" section.

- **N1 — Provenance and trust tiers** (`generated`, `verified`, the actor
  convention, OKF §5.2/§5.3/§7). The largest remaining gap and the intended next
  change. Deferred because it is not a metadata addition: Company OS encodes
  sign-off in three incompatible shapes that are load-bearing in blocking gates —
  `decisionOwner` on PRDs (a literal `TODO` hard-fails,
  `bin/company-os:627-630`), `approvedBy` on deviations, and `approvedBy` on
  exceptions where `'TODO: rule owner'` passes validation today
  (`examples/workspace/teams/customer-engagement/governance/exceptions.yaml:8`).
  Unifying them changes approval semantics and needs its own proposal. This
  change reserves the names only.
- **N2 — OKF `sources[]` with credibility signals.** `usage_count` is a catalog
  concept with no analogue here, and `docs/FRONTMATTER-CORE.md:76-77` forbids
  fetching external content, so the machinery would have nothing to feed on.
- **N3 — Attested Computation** (OKF §10). No sanctioned-computation problem in
  this domain.
- **N4 — `stale_after:` general freshness.** Deferred until a consumer surface
  exists; today it would add permanent CI noise for zero user outcome.
- **N5 — Making `title`, `description`, or `resource` required** beyond where the
  process contract already requires them.
- **N6 — Implementing anything from `ONTOLOGY-GUIDE.md` Part 3.** This change
  *labels* that material; it does not build it.
- **N7 — Changing the four-root federation layout** or the artifact directory
  structure.
- **N8 — Repository directory synchronization via GitHub MCP / Actions.** A
  separate initiative; the source→destination mapping it asks for already ships as
  `paths:` + `root:`. See
  `.devlocal/research/2026-07-25-knowledge-catalog-directory-sync.md`.

**Invariants — a change violating any of these is rejected regardless of merit:**

- **I1. Backward compatibility.** No workspace conformant at commit N-1 fails at
  commit N, beyond the documented `graph build` re-run.
- **I2. Absence tolerance.** `examples/standalone-team/` keeps passing with three
  of four federation roots missing.
- **I3. The tier model.** mandatory / default / guidance untouched. Nothing here
  becomes mandatory-tier.
- **I4. Generated files are derived, never hand-edited.**
- **I5. Gate 1–7 numbering and printed strings are frozen**
  (`bin/company-os:915-918`: "Never renumber gates 1-7"). Gate 8 belongs to
  federation.
- **I6. Idempotency.** Two consecutive `graph build` runs leave the workspace
  byte-identical.
- **I7. Unknown fields and unknown types are preserved, never rejected.**
- **I8. Single source of truth for ownership** (validate gate 1).
- **I9. Federated slices are read-only derived content.** Nothing may write into a
  materialized slice or invalidate `workspace.lock.yaml` hashes.

## Success Criteria

1. `company-os-starter/docs/CONFORMANCE.md` exists with: Goals, Non-Goals,
   Terminology, a numbered RFC-2119 conformance clause, an explicit
   MUST-NOT-reject list with a `bin/company-os` file:line beside each item, a
   version scheme, a declared OKF target version, and "considered and deferred"
   covering N1–N8.
2. Every SHOULD in the conformance clause has been checked against the CLI's
   blocking field checks; `title` is documented as conditionally required, because
   it blocks today for `type: prd` at `bin/company-os:975` and `:627-630`.
3. `2026.2` and the profile enum `minimal | standard | strict` are defined, and
   all eight fixtures committing a `conformance:` block are conformant against the
   new document without edits.
4. `grep -rn "validate --ontology\|spec trace\|@spec" company-os-starter/ examples/`
   returns hits only inside files or sections carrying a not-yet-available banner,
   or worded as roadmap.
5. Every `type: component-reality` document in `examples/workspace/` and
   `examples/standalone-team/` carries `resource: component://<id>`; documents that
   describe no external asset carry no `resource:`.
6. Every document in those two fixtures carries `title:`, and their committed
   `CLAUDE.md` blocks have been regenerated to match.
7. `company-os prd complete` emits an `outcome.md` carrying `title:`, and fails
   with a message naming the file and the bad value when a reality doc carries a
   malformed `updated:`.
8. `bash examples/acceptance.sh` exits 0 **with `examples/golden-validate.txt` and
   `examples/federated-golden-validate.txt` unchanged.** A golden diff in this
   change is a regression, not a re-baseline.

## Build Phases

| Unit | Content | CLI diff | `golden-validate.txt` | Committed generated files |
|---|---|---|---|---|
| U0 | Done-gate date parsing + selftest check | 1 line | unchanged | unchanged |
| U1 | Conformance & versioning document; reserve `generated`/`verified` | none | unchanged | unchanged |
| U2 | Ontology roadmap separation + fixture claim corrections | none | unchanged | unchanged |
| U3 | `title` + `resource`; templates incl. the `outcome.md` writer; fixtures | templates only | **unchanged** | `CLAUDE.md` blocks change |

Two properties of the frozen artifacts drive this table, both verified against the
harness rather than assumed:

- **`examples/golden-validate.txt` records only validate's stdout** — gate headers,
  document *paths* (gate 4), and per-root status (gate 5). It never contains a
  title. U3 therefore re-baselines **nothing**.
- **`examples/acceptance.sh` §4 requires the committed workspace to be already
  fully derived** (`s0 == s1 == s2`). U3 changes what `build_claude_node` renders,
  because it falls back `title or id or filename` (`bin/company-os:1682`), so the
  committed `CLAUDE.md` blocks must be regenerated and committed in the same
  change or §4 goes red.

## Deferred — per-directory `index.md` generation

Cut from this change after review; recorded here so the next spec starts from the
findings rather than rediscovering them.

**Why it was cut.** Two blockers and one dependency:

1. **It collides with federation mode.** The read-only slice in
   `examples/federated/` contains two graph documents —
   `platforms/communications/reality/components/customer-notification-service.md`
   and `platforms/communications/skills/creating-prd.SKILL.md`. Generating an index
   beside either mutates slice bytes, so the content hash no longer matches
   `workspace.lock.yaml` and **gate 8 fails**, breaking `acceptance.sh` §2b and §3.
   `acceptance.sh:72-75` already excludes `federated` from the double-build loop
   for exactly this reason. Any index design must skip manifest-declared slice
   roots (I9).
2. **Generation must be wired into every derived-artifact path, not just
   `graph build`.** `rebuild_generated` (`:1746`) has seven call sites — `:713`
   `prd complete`, `:1925` `init`, `:1952`/`:1959`/`:1970` `add`, `:1993`
   `reality new`. Wiring only `cmd_graph` means `company-os init` produces a
   workspace that fails its own validate, which `examples/selftest.py` already
   asserts against.
3. **`description:` belongs with it.** Its only consumer is the index, so shipping
   it here would create a field with no consumer — the exact drift this change
   exists to correct. Its quality also cannot be reviewed without a rendered
   index: the test for a bad description is whether it still reads true when
   pasted onto a sibling, and that comparison only works side by side.

**Decisions already made, to carry forward:**

- **Threshold: a directory gets an index only at ≥2 documents.** At ≥1, seven of
  ten generated files in `examples/workspace/` would be single-entry — a file
  whose whole content links the file beside it. ≥2 yields three indexes there and
  **zero** in `examples/standalone-team/`, honouring I2.
- **`archive/prds/<id>/` is not excluded.** It sits exactly at the threshold, and
  once `prd complete` emits a titled `outcome.md` (U3 here), its index surfaces
  the 90-day outcome obligation that is otherwise invisible until someone opens
  the file.
- **Gate 5 absorbs the drift check with its printed header byte-identical.** The
  header is one statement at `bin/company-os:1006` matching
  `examples/golden-validate.txt:28`; index status appends beneath it as one
  aggregate line per federation root, matching gate 6's shape. No new gate, no
  renumbering (I5).
- **`index.md` joins the `iter_graph_docs` skip-list in the same atomic commit as
  the generator**, never after — otherwise generated indexes are re-ingested as
  graph documents and idempotency breaks (I6).
- **Open questions for that spec:** whether `index.md` carries frontmatter at all
  (and if so, its `type`); the semantics of a qualifying directory with no index
  (absence-tolerant like gates 5 and 6, or a hard fail); how a pre-existing
  hand-written `index.md` is treated; and index removal when a directory drops
  below the threshold.
