---
type: tasks
id: tasks-tui-lifecycle-completion
title: TUI Lifecycle Completion — Tasks
status: complete
---

> **Complete 2026-08-29.** All five units landed (Amendments 5–8 to R-5.5, plus
> R-5.27). A change can be taken end to end from `company-os tui`. One thing
> found and deliberately not fixed is carried forward in unit 3: `reality new`
> misfiles a component under the wrong platform for *any* caller, and closing it
> in `RealityNew` is a CLI change that needs its own decision.

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
- [x] 3. `reality new` screen (est ~45m) — pickers for platform + component
  **LANDED 2026-08-29.** "new reality doc (writes)" added to `mutatingScreens`.
  Resolved at open time (R-5.26) like `add component`: the component list must
  both grow as components are added and **shrink** as docs are written, within
  one sitting. Only components without a reality doc are offered —
  `scaffold.RealityNew` refuses to overwrite, so the rest are certain errors.
  `make check` exit 0, five goldens byte-identical.

  **Required Amendment 7 to R-5.5** — the enumeration, plus one genuinely new
  clause. This is the first form whose two fields constrain each other, and a
  form cannot narrow one picker from another picker's value, so the catalog can
  hand `Build` a `(platform, component)` pair that exists nowhere. **That pair
  does not fail:** `RealityNew` never checks that the component lives under the
  platform it was given, so it would scaffold one platform's component
  underneath another and report success. `realityInvocation` refuses it in
  `Build` — `addInvocation`'s seam, for the same reason. The new clause is
  written generally, because the next two-field form will have the same shape.

  Tests: `TestRealityScreenRefusesAComponentFromAnotherPlatform` (both
  directions, plus a tree digest proving the refusal wrote nothing) and
  `TestRealityPickerOffersOnlyComponentsWithoutADoc` (reopening the screen after
  writing a doc, which is what proves R-5.26 holds here).

  **CLI hazard found, not fixed — needs a decision.** `company-os reality new
  --platform other svc` misfiles silently for *any* caller, not just the TUI:
  `internal/scaffold/commands.go:451` resolves the platform dir and writes,
  using `ws.FindComponent` only to look up a display name. The form guards the
  path it opens. Closing it properly means refusing in `RealityNew`, which is a
  CLI behavior change this unit's acceptance excludes — so it is deferred as its
  own decision, not folded in here.
- [x] 4. `prd complete` screen (est ~1h) — the done-gate refuses on unchecked
      checklist items or a stale reality doc, so the screen must render that
      refusal legibly rather than as a wall of text
  **LANDED 2026-08-29.** "complete PRD (writes)" added to `mutatingScreens`,
  resolved at open time (R-5.26) so the list shrinks as records are completed.
  Reuses `activePRDIDs` from unit 2. `make check` exit 0, five goldens
  byte-identical.

  **The wall of text was not there.** The refusal already names each unchecked
  count, each missing reality doc by component and path, and prints the exact
  `reality new` command that fixes it — six lines. Nothing was reformatted. The
  real defect was different and invisible until this screen existed: `prd
  complete`'s refusal is the only QUIET error in the system, `main.go` suppresses
  the `error:` line for it, and `runScreen` did not — so the TUI would have
  printed "done-check failed …" as the block header and again as an error line
  beneath it. Fixed in `runScreen`, and written up as **R-5.27** about quiet
  errors generally, since the duplication follows from the contract rather than
  from this command.

  **No `force` field, now or later** (recorded as a prohibition in R-5.5, not
  just as an omission). `--force` overrides the gate enforcing invariant 4; a
  gate waved through from a menu is not a gate, and the reader likeliest to
  reach for it there is the one who least knows what it protects. Not an R-5.10
  gap: R-5.10 requires every collected value to have a flag, not every flag to
  have a field.

  Tests: `TestCompleteScreenRendersTheDoneCheckRefusalOnce` counts the sentence
  rather than checking it appears (checking appearance passes on the bug) and
  asserts the reasons and the fix line survive the suppression;
  `TestCompleteScreenHasNoForceField` pins the prohibition at both the field and
  the preview.
- [x] 5. Docs: update the boundary disclosure in `company-os-starter/README.md`
      and both `TUTORIAL.md` copies — they currently say these four steps
      require the CLI, which is the honest statement *today* and becomes false
      as each unit lands
  **LANDED 2026-08-29.** All three "What the TUI cannot do" sections replaced
  with what it now does: the six lifecycle screens in order, and the
  cross-surface `bash` blocks deleted, since the loop no longer crosses surfaces.
  Both TUTORIAL copies were byte-identical in this section and were edited
  identically.

  The disclosure was **replaced, not deleted.** What a tool refuses is as
  load-bearing as what it offers, so the three genuine exclusions are now stated
  positively in all three docs: no `--force` field, no browsing screen writes,
  and no forms for `workspace sync` / `scratchpad init`.

  Also corrected `tuiform.go`'s header, which was stale from unit 1 onward and
  said `discover validate` "is deliberately absent from both this file and the
  read-only catalog" — false the moment unit 1 landed. It now records what the
  comment predicted, that the prediction was met, and which half of the
  prohibition still stands. `prd validate`'s weaker justification is stated
  there too rather than left for a reader to infer from a `(writes)` title that
  over-warns.

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
