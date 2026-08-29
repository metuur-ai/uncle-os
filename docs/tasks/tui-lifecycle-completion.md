---
type: tasks
id: tasks-tui-lifecycle-completion
title: TUI Lifecycle Completion — Tasks
status: draft
---

# TUI Lifecycle Completion — Tasks

**Goal, in the owner's words:** *"I want to make it simple for users, non
developer, to follow."*

**The problem, stated once:** a non-developer cannot finish a change. Today
`company-os tui` lets them create a discovery brief and open the PRD form — and
their own brief is not in that form's picker, because it lists only
`status: validated` briefs and **nothing in the UI validates one**
(`cmd/company-os/tuiform.go:272-289`, `validatedBriefIDs`). To get past that
screen they must leave for a terminal. Same for `prd validate`, `reality new`
and `prd complete`.

Everything in ux-simplification Phase 2 made the *CLI* faster. None of it
changed the above. This change does.

## Scope

Add four mutating screens to the TUI so a whole change can be completed without
typing a command:

| Screen | Command it wraps | Status |
|---|---|---|
| validate discovery brief | `discover validate <id>` | **1st — unblocks the create→PRD path** |
| validate PRD | `prd validate <id>` | 2nd |
| scaffold reality doc | `reality new --platform <p> <component>` | 3rd |
| complete PRD | `prd complete <id>` | 4th |

## Why this is tractable

1. The TUI dispatches through the **same command table** as the CLI
   (`cmd/company-os/commands.go:47-53`). No new logic — a screen builds an
   `*Args` and `runScreen` dispatches it.
2. The preview-and-confirm pattern already exists: `invocation.Preview()` shows
   the exact command, `Commit()` runs it (`tuiform.go:60-73`).
3. The design is already prescribed by the source. `tuiform.go:31-36` says
   `discover validate` "is a mutation wearing a read-only name… If it is ever
   offered, it belongs HERE, behind a preview and a confirmation, not in a
   browser."

## The one real design constraint

These are **mutations**, and the exclusion was deliberate, not an oversight. Each
screen must live in `mutatingScreens`, be titled `(writes)` like its neighbours,
and show what will change before it changes it. A browsing screen must never
quietly edit what is being browsed — that is the rule the current gap exists to
protect, and it survives this change.

## Units

- [x] 1. `discover validate` screen (est ~1h)
  **LANDED 2026-08-28.** Screen added to `mutatingScreens`, titled
  "validate discovery brief (writes)", sited immediately after the screen that
  creates what it validates. Picker lists **draft** briefs via a new
  `draftBriefIDs`; `validatedBriefIDs` and it now share `briefIDsWithStatus`.
  The form collects the brief id only — the team resolves from the brief's own
  directory (ux-simplification 1.2), which keeps the previewed line short enough
  for a non-developer to read. `make check` exit 0, five goldens byte-identical.

  **Required Amendment 5 to R-5.5** of `docs/ears/go-cli-tui-port.md`, not a test
  edit. R-5.5 forbade a form for `discover validate` *anywhere*, and a test
  enforced it across the whole catalog. The amendment follows the standard
  Amendment 4 set — a mutating form ships on an observed request — and the
  request is the owner's restated goal. **The load-bearing half is unchanged and
  still asserted:** no *browsing* screen may reach `discover validate`. The test
  now separates the two prohibitions rather than collapsing them into "nowhere",
  which had also forbidden the one safe home for it.
  - picker lists **draft** briefs (the inverse of `validatedBriefIDs`), because
    offering an already-validated brief is a no-op that looks like a choice
  - preview shows `company-os discover validate --team <t> <id>`
  - acceptance: create a brief in the TUI, validate it in the TUI, and see it
    appear in the "new PRD" form's picker — the exact path that is broken today
- [x] 2. `prd validate` screen (est ~45m) — picker lists active PRDs
  **LANDED 2026-08-28.** "validate PRD (writes)" added to `mutatingScreens`,
  sited immediately after the screen that creates what it checks. Picker lists
  active records via a new `activePRDIDs` — archived ones under `archive/prds/`
  are excluded, because a check on finished work is a choice that does nothing.
  The form collects the PRD id only; `--platform` is resolved from the id
  (ux-simplification 1.2, `resolvePlatform`). `make check` exit 0, five goldens
  byte-identical.

  **Required Amendment 6 to R-5.5** — the enumeration only. Amendment 5 had
  already named this as one of the three remaining gaps and set the standard.
  The one real judgement: `prd validate` mutates nothing, so `(writes)` and a
  place in `mutatingScreens` over-warn. It goes there anyway, because the
  read-only catalog dispatches no commands at all and that is structural rather
  than per-command; the first browsing screen to dispatch would be the precedent,
  and the next one added there would not be so harmless.

  Added `TestValidatePRDPickerOffersActiveRecordsOnly` — the generic catalog
  tests run on a fixture with no change records, where every picker is
  legitimately empty, so nothing until now proved a picker ever fills.

  **Known cosmetic wart, not introduced here:** the preview renders
  `company-os prd validate <id> --platform ''`, because `screenCommand` always
  prints a flag marked `required` and the parser's suspension of `--platform`
  for `validate` lives in `parseSubcommand`, which the renderer cannot see.
  Unit 1 shipped the same shape (`discover validate <id> --team ''`). Teaching
  the renderer the suspension rules would duplicate them; the fix belongs in one
  place or neither, and it is not this unit's scope.
- [ ] 3. `reality new` screen (est ~45m) — pickers for platform + component
- [ ] 4. `prd complete` screen (est ~1h) — the done-gate refuses on unchecked
      checklist items or a stale reality doc, so the screen must render that
      refusal legibly rather than as a wall of text
- [ ] 5. Docs: update the boundary disclosure in `company-os-starter/README.md`
      and both `TUTORIAL.md` copies — they currently say these four steps
      require the CLI, which is the honest statement *today* and becomes false
      as each unit lands

## Acceptance (every unit)

- `make check` exits 0; all five goldens byte-identical
- No CLI behavior change — the TUI is a caller of the command table, not a
  second implementation
- Nothing below `cmd/` prints or calls `os.Exit`

## Context carried forward

Prior work: `docs/{hld,lld,ears}/ux-simplification-phase-2.md` (Phase 2, shipped
— Units 1,2,3,6,7,8,9; U4 withdrawn, U5 held behind the `acceptance.sh`
landmine). Amendment A4 of the HLD names this change as the named follow-up.

Deliberately **not** doing, and why:

- `profile: solo` — killed (HLD A7): optimizes quarterly setup
- U5 (`remove graph`) — held: zero per-change value, real landmine
- observation sessions / headcount question — superseded. The owner restated the
  goal directly; no study is needed to establish that a non-developer cannot use
  a menu that dead-ends.

Open risks recorded in HLD A8/A9 and unchanged by this work: throughput rose
while artifact-fidelity feedback fell; U2's inference expires at the first
`add platform`.
