---
type: lld
id: lld-ux-simplification-phase-2
title: UX Simplification Phase 2 — Low-Level Design
status: draft
---

# UX Simplification Phase 2 — Low-Level Design

> **HELD — not locked.** Revised 2026-08-28 to apply review. Every citation below
> was re-verified against source; the first draft carried six wrong ones. See
> `docs/hld/ux-simplification-phase-2.md` §Amendments.

**Design parent:** `docs/hld/ux-simplification-phase-2.md`
**Requirements:** `docs/ears/ux-simplification-phase-2.md`
**Research:** `.devlocal/research/2026-08-27-ux-simplification-second-pass.md`

## Architecture

Eight live units (U4 withdrawn, U9 added), grouped as the HLD groups its goals.
Nothing below `cmd/` prints or calls `os.Exit`; commands return
`[]model.GateResult`.

| Unit | Group | Kind | Primary files |
|---|---|---|---|
| U1 On-ramp docs + disclosure | A | docs + 2 output lines | `TUTORIAL.md` ×2, `README.md`, `cmd/company-os/args.go`, `internal/next/next.go`, `internal/scaffold/` |
| U2 Infer `prd new` context | B | Go, additive | `cmd/company-os/args.go`, new resolver in `internal/` |
| U3 Guidance chain real values | B | Go, one site | `internal/product/sections.go`, `internal/product/discover.go` |
| ~~U4~~ | — | **withdrawn** | premise falsified — HLD A1 |
| U5 Remove `graph` | B | Go, deletion | `cmd/company-os/args.go`, `commands.go`, `args_test.go` |
| U6 Document PRD contract split | C | comment + docs | `internal/product/check.go`, `TUTORIAL.md` |
| U7 Aggregate empty-section warnings | A | Go + transcripts | `internal/product/contract.go`, `TUTORIAL.md` ×2 |
| U8 `CONFORMANCE.md` | C | docs | `company-os-starter/docs/CONFORMANCE.md` |
| U9 Doc-truth fixes | A | docs | `TUTORIAL.md` ×2 |

### U1 — On-ramp docs and disclosure

`company-os tui` is registered late to break an init cycle
(`cmd/company-os/commands.go:53`), flagless by design (`args.go:349-353`), and
one of three commands exempt from the workspace-root fail-fast
(`main.go:61-65`) — safe as a blind first command, offering a recovery menu
outside a workspace (`tuirecover.go:121`).

Surface to document (verified by `grep 'Cmd: "' cmd/company-os/tui*.go`):

- read-only screens: `today` (`tui.go:147`), `validate` (`:153`),
  `governance explain` (`:173`), `skills list` (`:180`), `ids list` (`:186`),
  `workspace status` (`:192`)
- browsers with no CLI equivalent: overview, component, PRD, discovery
- mutating flows: `discover new` (`tuiform.go:102`), `prd new` (`:149`),
  `add team|platform|component` (`:247`)
- one-key repair: `derive` (`tuiadvise.go:79`, `:115`), `governance resolve`
  (`:149`), `add team --repair` (`:162`)

**The boundary this unit must disclose.** Absent from the interface:
`discover validate`, `prd validate`, `prd complete`, `reality new`, `deviation`,
`exception`, `next`, `find`, `check`. The consequence is not merely "finish in
the CLI" — `tuiform.go:272-289` (`validatedBriefIDs`) lists only
`status: validated` briefs in the PRD form's picker, so a TUI-created brief is
**absent from the next screen with no explanation**. Documentation must name this
before a user meets it.

Two output additions: the `next` empty-state (`internal/next/next.go:179-186`)
and `init`'s completion guidance. *Citation corrected:* the first draft cited
`internal/scaffold/scaffold.go:74-75`, which is inside `Slugify()`. The `init`
guidance emitter must be located before the task is written — `grep` for
`FieldNext` in `internal/scaffold/` returns nothing, so it is emitted elsewhere.

### U2 — Infer `prd new` context

**The Phase 1 resolver cannot be reused directly.** It lives at
`cmd/company-os/product.go:101` (`resolvePlatform`) and `:126` (`resolveTeam`) —
package `main`. *(The first draft cited `internal/product/product.go`, which does
not exist.)* Two consequences:

1. `resolvePlatform` keys off `ws.FindPRD(id)` (`product.go:108`), and `prd new`
   has no existing PRD to find. Platform inference must enumerate
   `ws.AllPlatforms()` (`internal/workspace/workspace.go:202`).
2. **The new resolver must live in `internal/`**, not beside the Phase 1 pair in
   `cmd/`. U3 substitutes inside `internal/product`, and Go forbids importing
   `main`. Placing it in `cmd/` makes U3 fail to compile.

Resolution sources:

| Flag | Source | Confidence |
|---|---|---|
| `--team` | the brief's directory via `ws.FindDiscovery` (`workspace.go:285`) | lookup, not enumeration — strongest |
| `--platform` | enumerate `ws.AllPlatforms()`; unique → use | strong |
| `--components` | the resolved team's `ownership/components.yaml`; unique → use | weakest; drives the governance checklist |

**Parser mechanism — exactly one safe implementation.** The suspension at
`cmd/company-os/args.go:677-681` is a guarded `continue` keyed on `a.Action`
(the comment at `:663-664` confirms positionals are bound before this loop).
Add `a.Action == "new"` to that list — a one-line change.

**Do not implement by setting `required: false` on `prd --platform`.** That flips
`flagUsage()` to bracket the flag, changing every `prd` usage line from
`--platform PLATFORM` to `[--platform PLATFORM]`, pinned at
`cmd/company-os/args_test.go:636`.

### U3 — Guidance chain real values

**One site, not three.** `sections.go:109-116` (`CodePRDRealityNote`,
`CodePRDNext`) already interpolates resolved values via `f.Str("platform")`. The
only placeholder in the chain is `CodeDiscoveryValidateNext` at
`sections.go:98-101`:

```
"company-os prd new --team %s --from-discovery %s "+
    "--platform <platform-id> --components <comp-id,...>"
```

Substitute where the U2 resolver returns a unique candidate; otherwise emit
today's placeholder. Also update the `FieldNext` value (`discover.go:157`).

Depends on U2, and on U2 having placed its resolver in `internal/`.

### U5 — Remove `graph`

`derive` and `graph build` both dispatch to `cmdGraph` (`commands.go:39-40`),
which ignores `Args` (`views.go:42`). The harness that required `graph` was
deleted (`company-os-starter/Makefile:5-6`), so the alias has no remaining
consumer.

Remove the `graph` entry from `commandSpecs` (`args.go:318-324`) and from the
dispatch table (`commands.go:40`). `commandNames()` (`args.go:759-768`) then
drops it, which changes two user-facing strings — the `invalid choice:`
diagnostic (`args.go:521`) and the `usage:` line (`args.go:777`) — pinned at
`args_test.go:266`, `:312`, `:319`. `usage()` already diverges from
`commandNames()` via the `goOnly` filter (`main.go:275-288`), so no new mechanism
is needed.

> **⚠ The hazard that makes this the riskiest unit in the change.**
> `examples/acceptance.sh:123,125` invoke `graph build` with `>/dev/null 2>&1`
> and **no exit-code assertion**. They are the two writes inside the double-build
> idempotency check (`:119-132`). Remove `graph` without touching them and both
> exit 2, write nothing, and `s0 == s1 == s2` holds trivially **forever** — the
> suite stays green while the idempotency gate is dead. Nothing announces it.
> Fix both lines to `derive` in the same commit, add exit-code assertions, and
> prove the check still fails on a deliberately non-idempotent derivation before
> trusting it again.

**Full blast radius** — every `graph build` call site, all of which must land in
one commit:

| File:line | Kind | Failure mode if missed |
|---|---|---|
| `examples/acceptance.sh:123,125` | harness | **silent** — check dies green |
| `cmd/company-os/ansi_test.go:216,235` | test | loud |
| `cmd/company-os/exitcode_test.go:67` | test | loud |
| `cmd/company-os/json_test.go:125` | test | loud |
| `cmd/company-os/args_test.go:20,159,163` | test | loud |
| `cmd/company-os/tuiadvise_test.go:95` | test comment | loud |

### U6 — Document the PRD contract split

Gate 3 checks four fields (`internal/product/check.go:36`, `contractFields`);
`prd validate` checks six (`internal/product/prd.go:22-24`, `processFields`).
`check.go:32-35` already carries a partial rationale — **extend it** rather than
writing a competing one. Add the tutorial line. No code change.

### U7 — Aggregate empty-section warnings

`internal/product/contract.go:138-156` emits one `SevWarn` finding per empty
section, each carrying the full opt-in sentence. When more than one section in a
document is empty, emit a single finding listing them. `applyFormatPolicy`
(`contract.go:164`) already returns `[]Issue` that callers fold into findings, so
collapsing N into one is contained within `internal/product`. Teams that opted in
via `doc-formats.yaml` (`contract.go:111-128`) keep per-section findings.

**Golden risk is zero,** determined not guessed: `CodeSectionEmpty` reaches only
`discover.go:120` and `prd.go:301`; gate 3 never calls `sectionIssues`;
`nextscan.go:71` and `promote.go:512` discard the format half.
`grep 'format guidance only' examples/*.txt` returns nothing.

**The real coupling is the tutorials.** `TUTORIAL.md:164-166` and `:224-226`, and
`company-os-starter/docs/TUTORIAL.md:167-169` and `:227-229`, contain the exact
three-warning transcripts this unit changes. They must be updated in the same
commit.

### U8 — `CONFORMANCE.md`

Specified as R-1.1 (`docs/ears/okf-v02-conformance.md:78`), seven named sections,
task 1.2 `[ ]`. File confirmed absent.

**OKF-required floor:** `type:` present and non-empty
(`internal/product/contract.go:79-81`); preserve unknown keys and tolerate
unknown types (`docs/FRONTMATTER-CORE.md:18-20`,
`TestRewriteFrontmatterTagsPreservesUnknownKeys`); markdown+YAML substrate;
tolerate broken cross-links. `title`, `description`, `resource` are recommended;
`index.md` may be generated.

**Company-OS additions:** `id:` as a second required field (`contract.go:83-85`);
the 8/9 gates and exit-1 posture; lifecycle-typed vocabulary driving blocking
gates; derived-and-overwritten `tags:`; the canonical URI registry; tiering,
deviations and expiring exceptions; federation and lock hashing; mandatory
generation with drift-as-failure.

**Record without fixing** (HLD N7): `docs/00-original-proposal.md:3` claims OKF
v0.1 against an EARS targeting v0.2; `companyOsVersion: "2026.2"` in nine
fixtures against a version defined nowhere; `resource:` absent from all
fixtures; no single interop-contract claimant (R-1.14).

### U9 — Doc-truth fixes

Two verified false statements, both cheap:

1. `TUTORIAL.md:469-472` lists `CLAUDE.md` context nodes among the sources `find`
   fronts. No path in `find.go` reads one — `searchIndexes` matches base name
   `index.md` only (`find.go:338`). Both tutorial copies.
2. Stale `[1/7]`…`[7/7]` gate transcripts against the CLI's real `[1/8]`…`[8/8]`.

**Scope correction (2026-08-28).** Phase 1 recorded this as living in
`TUTORIAL.md` §8 (`docs/tasks/ux-simplification.md:230`). It does not — both
`TUTORIAL.md` copies return zero matches for `[1/7]`/`[7/7]`. The stale
transcripts are in the user guide instead:

| File | Lines |
|---|---|
| `company-os-starter/docs/user-guide/how-to/run-the-validation-gate.md` | 38, 45 |
| `company-os-starter/docs/user-guide/how-to/use-the-agent-skills.md` | 146 |
| `company-os-starter/docs/user-guide/how-to/grow-a-workspace.md` | 50, 86 |
| `company-os-starter/docs/user-guide/tutorials/01-first-day-with-company-os.md` | 241, 248 |

Four files rather than one, and outside the files U1 opens — so U9 is a slightly
larger unit than first scoped, and independent of U1.

## Constraints

**C1 — Goldens stay byte-identical.** All five; no `acceptance.sh --update`.
Determined, not assumed: no golden pins output any live unit touches. Goldens
capture `validate` stdout only (`acceptance.sh:35,53,77`); `init` output is
discarded (`:153`); the double-build compares checksums, not text (`:120-131`).

**C2 — House conventions.** `[]model.GateResult`; only `cmd/` and
`internal/render/` write output; nothing below `cmd/` calls `os.Exit` or prints.
Every mutating command prints the next command. The `frontmatter()` contract is
untouched. **U3 must not resolve this by calling from `internal/` into `cmd/`** —
that is both a compile error and a layering violation.

**C3 — Frozen gate ordinals.** Gates 1–7 never renumbered
(`internal/governance/gates.go:22-23`); gate 8 before the federation gate, which
stays last (`internal/validate/validate.go:91-97`).

**C4 — Explicit flags always win, and inference is unique-match-only.**
`prd new` writes into a platform's change-records, so a wrong inference misplaces
an artifact. No best-guess path.

**C5 — Additive only, with one deletion (U5).** Recorded in HLD A5 as a known
limitation: a design that forbids removal cannot reduce concepts. This is why the
change is held.

**C6 — Templates stay reference material.** Only
`templates/reality-component.md` is embedded (`templates/embed.go:34-35`).

**C7 — TUTORIAL.md copies stay in sync**, and their captured transcripts stay
accurate (U7, U9).

**C8 — Ordering.** U2 → U3 (hard: U3 needs U2's resolver in `internal/`).
U5 before U2 — both mutate the `commandSpecs` literal in `cmd/company-os/args.go`,
and U5 is the smaller diff. U1's docs half is independent of its two output
lines and can land first. U6, U8, U9 are fully independent.

**Recommended landing order:** `U8 → U6 → U9 → U1(docs) → U5 → U2 → U3 →
U1(output lines) → U7`.

## Key Decisions

**D1 — Optimize for the practiced non-technical operator.** *(Rewritten
2026-08-28. The original read "Optimize for non-technical users. U1 is therefore
the centerpiece and lands first; CLI ergonomics follow." Senior review flagged
that as contradicted-but-unretracted once the workload evidence arrived, and it
was right to.)*

The owner gave two facts, not two competing answers: **non-technical** is an
identity fact, **~40 changes/month** is a frequency fact. They compose into one
person — a practiced non-technical operator — and that person is served by
neither of the obvious readings:

- A **menu** fails them at repetition 400. U1 is correctly demoted.
- **Engineer-shaped CLI ergonomics** would be the wrong correction. U2/U3/U7 are
  not engineer features: U2 removes three IDs this person must hold, U3 makes the
  printed line pasteable, U7 removes noise from a screen they read ~480 times a
  year. Nothing there assumes an engineer.

So the audience never changed; its *cadence* did. The distinction matters because
"we demoted the non-technical unit" and "we retargeted at the practiced version
of the same person" are one commit apart and a year of drift apart.

**Open and gating (recorded, not resolved): headcount.** 40 changes/month ÷ N
producers decides this. N=1 is ~2 changes/business-day and deep muscle memory, so
the demotion is strongly right. N=8 is ~5/month each — roughly weekly, not muscle
memory — and a menu plausibly wins, which would make the demotion wrong. The same
number decides whether an observation session has a valid participant to recruit
at all. It was drafted as one of five owner questions and is the one that did not
come back.

U1 and U7 remain the Group A success condition; SC1–SC4 and SC10 reflect it.

**D2 — Document the PRD contract split rather than widen gate 3.** *Rejected:*
widening gate 3 — better enforcement, but changes goldens and can fail active
PRDs in the field. *Also rejected:* narrowing `prd validate`, which loses a useful
author-time check.

**D3 — Remove `graph` entirely.** The harness justifying it is gone; only our own
tests pin the strings. *Rejected:* hiding it in help while keeping the alias
forever, which was the first draft's plan and rested on a dead dependency.

**D4 — `--components` inference may be dropped.** It drives the governance
checklist, so a wrong list costs more than a wrong platform. If fixture testing
shows loose resolution, ship U2 with `--team` and `--platform` only.

**D5 — Suspend `--platform` for `prd new` specifically, not unconditionally.**
Add `a.Action == "new"` to the existing per-action list. *Rejected:*
unconditional suspension (changes error messages for four other actions) and
`required: false` (breaks `args_test.go:636`).

**D6 — Disclose the interface boundary rather than close it or hide it.** The
exclusion of `discover validate` is principled (`tuiform.go:31-36`), so closing it
is a design change (HLD A4). *Rejected:* promoting the interface without
disclosure — a user meeting the empty picker unwarned is worse off than one who
never opened it.

**D7 — Aggregate warnings only where they do not block.** Where
`doc-formats.yaml` sets `enforce: true`, per-section detail is earned.

**D8 — `CONFORMANCE.md` records inaccuracies without fixing them.** They belong
to the parent conformance change.

**D9 — Deferral clauses removed.** The first draft pre-authorized four units to
defer against a golden risk that does not exist, with no stated expectation — so a
skip would have read as compliance. Verification remains, with the expected
outcome stated: it passes.

## Out of Scope

- Closing the interactive interface's lifecycle gap (HLD A4 — named follow-up)
- `find` staleness (HLD N9, A6 — named follow-up)
- `add component`'s `team://TODO` placeholder (HLD N10)
- Widening gate 3; any gate semantics or exit-code change (N1, N3)
- The `profile:` enum and conceptual reduction generally (N8, A5) — the subject of
  the validation protocol, not of this change
- The unbuilt OKF conformance phases; fixing what `CONFORMANCE.md` records (N7, N8)
- A style-preserving YAML emitter — a separate yamlio project
- An `ingest` workflow — the one OKF mechanism with no coverage
- Any migration for existing workspaces
