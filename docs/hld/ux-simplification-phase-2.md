---
type: hld
id: hld-ux-simplification-phase-2
title: UX Simplification Phase 2 — High-Level Design
status: draft
---

# UX Simplification Phase 2 — High-Level Design

> **HELD — not locked.** Revised 2026-08-28 to apply product-owner, tech-lead and
> senior-engineer review. The documents are now correct, but the change remains
> held pending diagnosis validation:
> `.devlocal/research/2026-08-28-ux-diagnosis-validation-protocol.md`.
> Units 6 and 8 survived review unchanged and are cleared to land independently.
> Amendments are recorded in §Amendments rather than silently absorbed.

**Source:** `.devlocal/research/2026-08-27-ux-simplification-second-pass.md`
**Predecessor:** `docs/tasks/ux-simplification.md` (Phase 1, all six units landed
2026-08-26/27; the complaint persisted).
**Design:** `docs/lld/ux-simplification-phase-2.md`,
`docs/ears/ux-simplification-phase-2.md`.

## Overview

Phase 1 attacked the command chain and landed in full. The complaint that
motivated it — "following the process requires holding too much in one's head" —
survived it. That outcome is the most important input to this document, and it is
the reason the change is held rather than locked: it is not yet established
whether the remaining problem is ergonomic or conceptual.

What review did establish is that several defects assumed by the first draft of
this design do not exist. `add team` does not produce a workspace its own
validator rejects. The committed goldens do not constrain the two most
user-visible units. The Python differential harness that justified keeping a
duplicate command was deleted months ago. Those corrections are applied
throughout; the units that remain are the ones whose premises survived checking.

The interactive interface is the centre of this change and also its largest open
risk. `company-os tui` ships with browsers that exist nowhere else, guided
creation flows and one-key drift repair — and it cannot complete a unit of work.
A user who creates a discovery brief in the TUI will not find it in the next
screen's picker. This design therefore documents the interface *and its boundary*
together; it does not promote it as a complete path.

This change is **additive**: it adds capability, documentation and inference,
with one deletion (a duplicate command name). It adds no enforcement, no gate,
and no check on how an artifact was produced. That constraint is inherited from
Phase 1 and is recorded in §Amendments as a known limitation of this approach,
not as a virtue.

## Stakeholders & Impact

**Primary: non-technical practitioners — product owners, product managers,
business analysts, designers.** Named in the 2026-07-22 journey review and
confirmed as the target for this change.

- *Today:* the documented path to a first artifact runs through 30 invocable
  command forms and 25 flags. A menu-driven interface exists and no tutorial or
  README mentions it — but it cannot finish a job.
- *After:* the interface is documented as the entry point for browsing and
  creation, with an explicit, honest statement of which steps require the CLI and
  the exact commands that close the loop.

**Secondary: engineers.** `prd new` is the only lifecycle step that still
hard-fails without `--platform` in workspaces where one value is legal. After
this change it infers what the workspace already knows.

**Secondary: LLM agents.** The guidance chain prints `<platform-id>` placeholders
rather than resolved values, forcing a discovery detour. After this change it
emits runnable commands wherever inference resolves uniquely.

**Secondary: adopters evaluating the method.** "What must I do because of the
standard, and what is this project's opinion?" currently has no answer to point
at. `CONFORMANCE.md` answers it, and is the prerequisite for arguing any later
removal.

**Not affected:** existing workspaces. Every current invocation continues to
behave identically.

## Goals

Grouped by who they serve, because review found the first draft mislabelled an
engineer-facing change as a non-technical one.

**Group A — the non-technical on-ramp (this change's success condition):**

1. A person who has never used the tool reaches a working interactive menu
   without reading the command list.
2. The four capabilities reachable only through that interface — workspace
   overview, component, PRD and discovery browsers — are discoverable from the
   documentation.
3. A reader following only the documented path can complete one full discovery →
   PRD → complete lifecycle, crossing between interface and CLI at points the
   documentation names in advance.
4. First-run output does not present a freshly scaffolded artifact as a wall of
   warnings.

**Group B — CLI and agent ergonomics:**

5. `company-os prd new` completes without `--platform`, `--team` or
   `--components` whenever the workspace admits exactly one candidate for each.
6. Every command the guidance chain prints is directly runnable wherever
   inference resolves uniquely.
7. `company-os --help` lists 19 commands, with no duplicate name for tag
   derivation.

**Group C — contract documentation:**

8. A reader can determine from a single document which requirements come from OKF
   and which are this project's own.
9. The difference between gate 3's four-field PRD contract and `prd validate`'s
   six-field contract is documented as intentional.

*(The former goal "explicit flags win over inference" was an invariant, not a
goal; it now lives in EARS Unit 0 as R-0.13.)*

## Non-Goals

**N1 — No gate semantics change.** Gates `[1/8]`–`[8/8]` and `[1/9]`–`[9/9]` keep
their exact semantics and output. Exit codes unchanged.

**N2 — No golden updates.** All five committed goldens stay byte-identical, and
no unit runs `acceptance.sh --update`. *Review correction:* this is far less
constraining than the first draft assumed — no golden pins the output any unit
here touches. It is a genuine invariant, not a scoping veto.

**N3 — Gate 3 is not widened.** The two-contract mismatch is resolved by
documenting the split. Widening gate 3 would change goldens and could fail active
PRDs in the field; that is a governance change and needs its own EARS unit.

**N4 — No new TUI screens in this change.** The interface's inability to complete
a lifecycle is real, understood, and **recorded as a named follow-up change**
(§Amendments A4). It is not left as an open question, and it is disclosed in the
documentation this change writes.

**N5 — No process enforcement.** Strict on artifacts, flexible on process.

**N6 — The CLI remains fully usable without the interactive interface.** Nothing
becomes TUI-only; no CLI capability is deprecated.

**N7 — `CONFORMANCE.md` documents, it does not change conformance.** Any
requirement it reveals as unmet is recorded, not fixed here.

**N8 — The unbuilt OKF conformance phases stay unbuilt.** `title:`/`resource:`
emission, the `profile:` enum and `ONTOLOGY-ROADMAP.md` remain the parent
change's work. *Note:* the `profile:` enum is also the primitive a
conceptual-simplification approach would need — see §Amendments A5.

**N9 — `find`'s staleness is not fixed here.** `find` consumes four derived
surfaces with no freshness check and can silently return stale results
(`internal/find/find.go:124-428`). This is a real trust defect affecting the
primary audience most. It is out of scope for this change and is assigned as a
named follow-up (§Amendments A6) rather than left ownerless.

**N10 — `add component` keeps writing `accountableTeam: team://TODO`.** The
component half of the scaffolder-placeholder problem is not addressed here.
Unlike the team case (see A2), this one is real — gate 1 does fail on an
unreplaced `team://TODO`. It is deliberately deferred because it is reachable
from the TUI's `add component` form and should be designed together with the
interface work in A4.

## Success Criteria

Every unit has at least one observable criterion. Review found the first draft
gave none to the two user-visible units.

| # | Criterion | Unit |
|---|---|---|
| SC1 | `TUTORIAL.md` (both copies) and `README.md` name `company-os tui` before the first CLI invocation is taught | U1 |
| SC2 | The documentation names which lifecycle steps the interface performs and which require the CLI, listing `discover validate`, `prd validate` and `prd complete` explicitly | U1 |
| SC3 | A reader following only the documented path completes one full discovery → PRD → complete lifecycle | U1 |
| SC4 | `company-os next` (empty) and `company-os init` each offer the interface alongside existing guidance | U1 |
| SC5 | `prd new --from-discovery <id> "<title>"` succeeds on `examples/standalone-team` with no other flags, producing a PRD byte-identical to the fully-flagged invocation | U2 |
| SC6 | Ambiguous inference names every candidate; absent inference gives the existing not-found error | U2 |
| SC7 | `company-os discover validate <id>` prints a `prd new` command containing no `<placeholder>` tokens on a single-platform workspace | U3 |
| SC8 | `company-os --help` lists 19 commands; `graph` is absent from help, dispatch and `commandNames()` | U5 |
| SC9 | `check.go` and `TUTORIAL.md` each state why `prd validate` enforces two fields gate 3 does not | U6 |
| SC10 | A freshly scaffolded brief produces one empty-section warning line, not three; both `TUTORIAL.md` transcript blocks match | U7 |
| SC11 | `company-os-starter/docs/CONFORMANCE.md` exists and separates the OKF floor from Company-OS additions, with a source citation per claim | U8 |
| SC12 | No document claims `find` fronts `CLAUDE.md` context nodes; `TUTORIAL.md` §8 shows the real `[1/8]`…`[8/8]` gate output | U9 |

**Gates on the work itself:**

- SC13 — `make check` exits 0.
- SC14 — All five committed goldens byte-identical.
- SC15 — `company-os validate` exits 0 on `examples/workspace` and
  `examples/standalone-team`.
- SC16 — Every inference has an overriding flag; every existing flag-carrying
  invocation behaves byte-identically.

## Amendments

Recorded per house convention: when review contradicts a specified premise, the
contradiction is written down rather than silently absorbed.

**A1 — Unit 4 withdrawn (premise falsified).** The first draft claimed `add team`
produces a workspace its validator rejects. It does not. Verified live:
`add team newteam` → `validate` → gates 1 and 2 clean, exit 0. Gates are
absence-tolerant by construction (`internal/governance/gates.go:47`,
`loadOr(path, pyMap{})`), so a missing file and an empty list are byte-identical;
`examples/acceptance.sh:151-162` has asserted this since it was written. The
former SC "reduce hand-authored files from 12 to 9" measured files that gate
nothing. **Unit 4 is removed.** Its number is retained and marked withdrawn
rather than renumbering the rest.

**A2 — The goldens never constrained this change.** The first draft made
byte-identical goldens a design veto and pre-authorized four units to defer
against it. All five goldens snapshot `validate` output only; empty-section
warnings are emitted by `discover validate`/`prd validate`, and `init` stdout is
discarded (`examples/acceptance.sh:153`). `grep "is empty" examples/*.txt`
returns nothing. The deferral clauses are removed; verification remains as a
cheap check with a stated expected outcome.

**A3 — The Python differential harness does not exist.** `company-os-starter/Makefile:5-6`
records that R-9.3 deleted the Python reference and retired the harness with it.
The first draft's N4/D3 kept `graph build` alive for it. `commandNames()` still
feeds two user-facing strings (`cmd/company-os/args.go:521`, `:777`) pinned by
our own tests (`args_test.go:266,312,319`), which are updatable. **Decision:
`graph` is removed entirely**, superseding the hide-in-help approach.

**A4 — The interactive interface cannot complete a lifecycle (named follow-up).**
`cmd/company-os/tuiform.go:272-289` populates the PRD form's `from-discovery`
picker only with `status: validated` briefs, and the command that validates a
brief is, per that function's own comment, "not called from anywhere in the UI."
`internal/product/prd.go:192-196` then refuses a non-validated brief. A user who
creates a brief in the interface will not see it in the next screen. The
exclusion is deliberate and well-argued (`tuiform.go:31-36`: `discover validate`
is "a mutation wearing a read-only name"), so closing it is a UI design change,
not a bug fix. **Follow-up change: interactive lifecycle completion** — offer
`discover validate`, `prd validate`, `prd complete` and `reality new` behind
preview-and-confirm, as that comment prescribes. Until it lands, U1 discloses the
boundary.

**A5 — This design cannot reduce concepts, and that may be the wrong axis.**
Senior review's central objection: Phase 1 was a complete ergonomic intervention
that failed to resolve the complaint, and this change is aimed at the same axis
while its additive-only constraint makes removal structurally impossible. The
alternative — a `profile: solo` where concepts do not exist rather than being
easier to invoke (fewer roots, fewer gates, no deviations/exceptions, discovery
and PRD collapsed) — is not attempted here and is excluded by N8. **This is why
the change is held rather than locked.** The validation protocol exists to decide
between the two axes before more is specified.

**A6 — `find` staleness assigned, not dropped — and PROMOTED by the workload
evidence.** See N9. `internal/find/find.go:124-428` consumes four derived
surfaces with no freshness check and returns stale hits silently. At ~40
changes/month that is hundreds of consultations a year with an invisible failure
mode, hitting hardest the user least equipped to detect it. Cheap: reuse the
drift detection already in `cmd/company-os/tuiadvise.go:69-115`. **This is the
highest-value code item not currently scoped.**

**A7 — `profile: solo` is killed, in writing, not parked.** The two-root finding
is real (a `platforms/` + `teams/` workspace validates clean and idempotent with
zero code changes) and the work is S–M. It is still **not worth doing**: it
optimizes workspace *setup*, which the owner reports happens **quarterly**,
against an explicit instruction to design for today's reality. Left on the board
as "the big conceptual option" it will keep attracting attention it no longer
deserves. Recorded as closed so the next reader does not re-derive it. Reopen
only if adoption — not per-change cost — becomes the goal.

**A8 — the throughput/fidelity risk, named because nobody had named it.** Read
U2, U3, U7 and U6 together rather than separately. U2 removes the step where an
author states which platform and components a change touches; U3 turns the
guidance chain into a paste-through script, so a full lifecycle can run with no
command composed by a human; U7 collapses the warnings that were — clumsily — the
only signal that an artifact was still empty; U6 blesses gate 3 accepting four
fields where `prd validate` wants six. **Individually each removes friction;
together they remove the moments at which an operator is obliged to look at the
artifact.** At 40 changes/month with agents in the loop, that composition is a
well-formed pipeline for producing artifacts that pass every gate and say
nothing. Gates measure structure; nothing measures whether a reality doc
describes reality. Throughput rose, the only quality feedback fell, and no
instrument watches the second number. Not addressed here; recorded so the next
change cannot claim it was unforeseen.

**A9 — U2's inference has a quarterly expiry.** Inference is unique-match-only,
which is correct. It therefore works today and **stops working at the first
`add platform`** — an event the owner reports happens quarterly. By then the
docs, the operators' habits and every agent prompt will have standardized on the
flag-free form, and it breaks for all of them simultaneously with a usage error.
Survivable if anticipated: the help text and tutorial should say inference holds
*while* the workspace admits one candidate, and someone should decide before it
happens whether ambiguity ought to prompt rather than fail.
