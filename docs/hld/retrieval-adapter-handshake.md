---
type: hld
id: hld-retrieval-adapter-handshake
title: Retrieval Adapter Handshake — consuming the local-search engine contract
status: draft
tags: [kind/hld, status/draft]
---

# Retrieval Adapter Handshake — High-Level Design

## Overview

Give Company OS a single, typed, subprocess-isolated way to talk to the
`local-search` engine, and wire the two cheapest consumers of that channel:
registering a freshly scaffolded workspace with the engine (`init`), and
surfacing engine drift once, loudly, in `check`. Nothing above
`internal/retrieval` learns that a subprocess exists.

The organising constraint is not retrieval — it is **absence**. `local-search` is
a separate binary on a separate release cadence, installed on some machines and
not others. Company OS must stay *fully usable* without it. Every design decision
below falls out of that: the adapter is a leaf package with one entry point, the
engine handshake is performed at the call site rather than at process start, and
the only new failure mode a missing engine can introduce is a warning — and, after
review, not even that: with the engine absent, `check` is byte-identical to today
and only `init` speaks up.

**What this slice is honestly for.** It delivers no new user-facing verb. Units 1
and 2 are plumbing and carry most of the risk; Unit 3's payoff is collected
outside `company-os`; Unit 4 speaks only to users who already installed the
engine. The case for doing it now is *de-risking*: it pins the subprocess
contract, the version floor, and the absence invariants before `#1b`/`#2` arrive
with real result types and a user-visible verb to argue about. If the goal were
user value soonest, `#5 check --fix` is the stronger next slice by the backlog's
own assessment — that is a sequencing decision taken with open eyes, not a claim
this slice out-values it.

This is scope **#1a, #3, #4** of the improvement backlog
(`.devlocal/research/2026-08-23-os-improvements-backlog.md`). The result-typed
half of the adapter (#1b, `find` / `search`) and the `ask` verb (#2) are
explicitly out of scope and blocked on engine-side work.

Units below are numbered as in the EARS spec; the backlog scope item each one
serves is in the last column.

| Unit | Name | Engine dependency | Shape | Backlog |
|---|---|---|---|---|
| 1 | Engine discovery, version floor, absence | none — exercised against shipped `-v` | New leaf package | #1a |
| 2 | Invocation contract — JSON, timeout, failure isolation | none — exercised against `graph explain --json` / `doctor --json` | Same package | #1a |
| 3 | Workspace registration on `init` | none — `local-search repo add <path>` is non-interactive when given an explicit path | Scaffold ride-along + guidance-chain line | #3 |
| 4 | Engine drift reported by `check` | none | New advisory gate finding | #4 |

## Stakeholders & Impact

**Company OS users without `local-search` installed** — the majority today. Their
pain: none yet, because nothing calls the engine. What changes after this ships:
**nothing at all.** `check` output stays byte-identical for them (R-4.7). An
earlier draft gave them a `[WARN]` line plus an install command on every `check`
run; that was an unsolicited prompt in the readiness/completion path, advertising
retrieval this slice does not yet deliver, aimed at the majority population. It
is withdrawn until `uncle ask` exists to make the suggestion worth acting on.
Every existing command keeps its current exit code. This is the population most
at risk from a careless implementation, and the one the EARS invariants are
written to protect.

**Company OS users with `local-search` installed** — their workspace is
registered with the machine-wide federation at `init` time instead of never, so
retrieval works the first time they reach for it rather than after an
undocumented manual `repo add`. Stated honestly: that payoff is *collected*
through the `local-search` CLI and the agent skill, not through `company-os` —
this slice adds no verb that queries the index. What `company-os` contributes is
that the workspace is discoverable at all. `check` tells them when their engine is too old
or their index is stale, at a moment when they are already reading output, rather
than at the moment they urgently need an answer.

**CLI maintainers** — gain one bounded subprocess seam with a documented contract
and a version floor, instead of ad-hoc `exec.Command` calls accumulating at call
sites. They inherit a second external-tool dependency to keep working; the
mitigation is that it is optional in a way `git` is not.

**Downstream agents and skills** — unchanged surface. They see typed results or a
typed absence, never a parse error and never a hang.

## Goals

- `internal/retrieval` is the sole subprocess user for the engine, exposing typed
  results and a typed *absent* state; no caller branches on `exec.ErrNotFound`.
- A minimum engine version is asserted once per process, cached, and enforced
  before any JSON contract is relied upon.
- With `local-search` absent from `PATH`, `go test ./...` passes and every
  non-retrieval command produces byte-identical output and exit codes to today.
- A hung or wedged engine cannot hang a Company OS command — every invocation is
  bounded by a timeout and honours context cancellation.
- `company-os init` registers the new workspace with the engine when one is
  present, announces the indexing work *before* it blocks, and prints the next
  command. `add` does not register: it operates inside an already-registered
  workspace root, so there is nothing new for the engine to learn.
- Re-registering an already-known root is a **success**, distinguished from a
  wrong path by reading the engine's registry file before invoking `repo add` —
  not by scraping the failure prose, which the exit code cannot separate from any
  other failure.
- `make check` produces identical results whether or not a real engine is
  installed, and no test run mutates the developer's machine-wide engine state.
- `company-os check` reports engine presence, version-floor satisfaction, and
  index staleness as advisory findings.

## Non-Goals

- **`uncle ask` and the `find` / `search` result types (#1b, #2).** Blocked on
  engine-side work; adding a second already-shaped JSON call to a working adapter
  is small and deliberately deferred.
- **Any fallback to scraping human-readable engine output**, with exactly one
  bounded exception (the `-v` version line) justified in the LLD. The engine's
  on-disk registry file is *read* as a pre-check (a versioned coupling declared
  in the LLD and EARS) — reading a data file is not scraping output, but it is a
  real coupling and is named as one rather than hidden.
- **Making retrieval mandatory anywhere.** No gate may move from pass to fail
  because `local-search` is missing. This is an invariant, not a preference.
- **Vendoring, embedding, or auto-installing `local-search`.** Company OS
  references an external tool; it never ships or fetches one.
- **Changing `local-search` itself.** Engine-side conveniences (`repo add --yes`,
  `version --json`) are assumed to land later; this design works without them.
- **Re-numbering or re-slugging existing `validate` gates.** The `check` finding
  is additive.

One new public surface is introduced and is in scope: **`company-os init
--no-engine`**, which skips registration outright, for CI and for users who do
not want their workspace in a machine-wide index. It is the only new flag, it is
opt-out rather than opt-in because the registered case is the one users want by
default, and `examples/acceptance.sh` must exercise it so the acceptance run
never touches the machine-wide registry.

## Success Criteria

Each criterion below is reachable from the test fixture described in D8 — none
requires a real engine on the machine running the suite.

1. **Absence is invisible.** With no `local-search` on `PATH`: `make check` is
   green, `company-os validate` on `examples/workspace` still exits 0, and
   `company-os check` output is **byte-identical to the pre-change binary**.
   Falsified by any diff at all.
2. **Registration is idempotent.** With a conforming engine: `company-os init`
   leaves the workspace registered and says so; a second `init` on the same root
   reports registration as an ok outcome, not a defect — and does so *without*
   the second run invoking `repo add` at all. Falsified by a fixture that records
   its invocations showing a second `repo add`, or by an `[ERROR]`/non-zero exit
   on the second run.
3. **A below-floor engine warns and nothing fails.** `check` reports the floor
   breach as an advisory finding naming the installed version, the required
   version, and the upgrade command; `check`'s exit code is unchanged from the
   engine-absent case. Falsified by an exit-code change or a missing field.
4. **A stale index is reported, and staleness alone never fails a gate.** With a
   fixture reporting `index_stale`, `check` emits exactly one advisory finding
   naming the staleness and the rescan command, and exits as it would with no
   engine. Falsified by a gate transitioning to fail, or by staleness passing
   silently.
5. **A hung engine cannot hang Company OS.** A fixture that sleeps past the
   deadline is abandoned at the timeout; the calling command completes within the
   deadline plus a small margin and reports the timeout as an engine fault, not a
   workspace fault. Falsified by a command that outlives its deadline.
6. **Test runs do not touch the developer's machine.** After `make check` on a
   machine with a real engine installed, the machine-wide registry file's
   modification time and contents are unchanged. Falsified by any mutation.

`No package outside internal/retrieval names the engine binary` is not a success
criterion — it is a structural constraint, stated under Constraints in the LLD and
enforced by a grep test. It describes the design rather than an observable
outcome, so it does not belong here.
