---
type: ears
id: ears-retrieval-adapter-handshake
title: Retrieval Adapter Handshake — EARS Specifications
status: draft
tags: [kind/ears, status/draft]
---

# Retrieval Adapter Handshake — EARS Specifications

Keywords: `THE SYSTEM SHALL` (always-on), `WHEN` (event), `WHILE` (during a
state), `IF` (conditional/gate), `WHERE` (context-scoped). "The system" = the
`company-os` CLI, including `internal/retrieval`. "The engine" = the
`local-search` binary.

**Committed scope:** Units 1–4. Units 1 and 2 are the adapter (#1a); Unit 3 is
registration (#3); Unit 4 is engine drift in `check` (#4).

---

## Unit 1: Engine discovery, version floor, and absence

**Why:** Company OS must remain fully usable on a machine with no `local-search`
installed. Absence is the design's central case, not its error path — so it is
specified before any successful call is.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL confine all engine subprocess execution to `internal/retrieval`; no package outside it SHALL name the engine binary. (Stated as "name the binary" rather than "import `os/exec` for that purpose" so that it is mechanically greppable: `os/exec` is legitimately imported today by `internal/federation/git.go` and several test files, and no guard test can infer intent.) |
| R-1.2 | THE SYSTEM SHALL resolve the engine exactly once per process via `PATH` lookup and SHALL cache the resolved path. |
| R-1.3 | IF the engine is not found on `PATH`, THE SYSTEM SHALL return a distinguishable sentinel absence value that callers can test without inspecting an operating-system error type or matching error text. |
| R-1.4 | WHILE the engine is absent, THE SYSTEM SHALL leave every command's output and exit code unchanged from its behaviour prior to this change, with the sole exceptions of the advisory findings required by R-3.5 and R-4.3. (Stated against the pre-change baseline rather than against "behaviour when the engine is present", because `init` and `check` deliberately differ between present and absent — that difference is the feature.) |
| R-1.5 | THE SYSTEM SHALL declare a minimum supported engine version as a single named constant. |
| R-1.6 | WHEN the engine version is first needed, THE SYSTEM SHALL determine it by invoking the engine's version flag, and SHALL cache the result for the remainder of the process. |
| R-1.7 | THE SYSTEM SHALL NOT invoke the engine while serving `--help`, `--version`, or an argument-parsing failure. |
| R-1.8 | IF the installed engine version is below the declared minimum, THE SYSTEM SHALL report the installed version, the required version, and the upgrade command. |
| R-1.9 | WHERE the system produces a non-zero exit code because the engine is absent, is below the minimum version, exceeded its timeout, or failed to execute, THE SYSTEM SHALL use the external-tool exit code (6), and SHALL NOT use the usage (2), workspace (3), or artifact (4) codes. **Not reachable in this slice** — no command in scope (#3, #4) exits non-zero on any engine condition, so this clause carries no test-side `@spec` obligation here. It governs #1b/#2, where a command may genuinely require the engine, and is stated now so the exit-code decision (D1) is recorded once rather than relitigated. |
| R-1.9a | THE SYSTEM SHALL amend the documented meaning of exit code 6 to cover external tools generally rather than git specifically, before any code attributes that code to the engine. |
| R-1.10 | THE SYSTEM SHALL NOT fail or alter any gate result solely because the engine is absent, and SHALL NOT warn about engine absence except as permitted by R-3.5 and R-4.3. |
| R-1.11 | WHEN the system reports engine absence or a version-floor breach to a user, THE SYSTEM SHALL state that the condition affects retrieval only and that no other command is impaired, and SHALL print the remedying command. (Reworded from "name the command that requires the engine": in this slice no command requires it, so the original phrasing had no true instance.) |
| R-1.12 | IF the engine's version output cannot be parsed into a comparable version, THE SYSTEM SHALL treat the engine as unusable for version-gated work, report the raw output it could not parse, and SHALL NOT panic or assume the engine satisfies the floor. |

---

## Unit 2: Invocation contract — JSON, timeout, and failure isolation

**Why:** A subprocess is an unbounded hazard: it can hang, emit garbage, or
change its output format between releases. These requirements make each of those
a typed, bounded, non-contagious failure.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL request machine-readable JSON output for every engine invocation that offers a JSON mode, excepting the version query. WHERE a wrapped subcommand offers no JSON mode, THE SYSTEM SHALL record that fact explicitly alongside the invocation and SHALL derive its outcome by the means specified for it, never by parsing its prose. (`repo add` is the one such subcommand in scope; see R-3.9a.) |
| R-2.2 | THE SYSTEM SHALL NOT parse human-readable engine output, with the single exception of the version line permitted by R-2.3. |
| R-2.3 | WHERE the version line is parsed, THE SYSTEM SHALL treat it as the sole permitted scrape, on the grounds that it establishes whether the JSON contract is available. |
| R-2.4 | IF a required JSON output mode is unavailable on the installed engine, THE SYSTEM SHALL fail the version assertion and SHALL NOT fall back to parsing human-readable output. |
| R-2.5 | THE SYSTEM SHALL bound every engine invocation with a deadline and SHALL propagate caller context cancellation to the subprocess. |
| R-2.6 | WHEN an engine invocation exceeds its deadline, THE SYSTEM SHALL terminate the subprocess, return a timeout failure attributed to the engine, and allow the calling command to complete. |
| R-2.7 | IF the engine exits non-zero, THE SYSTEM SHALL capture its diagnostic output and include it in the returned error rather than discarding it. |
| R-2.7a | WHERE an engine subcommand uses its exit code to convey report severity rather than invocation failure, THE SYSTEM SHALL treat a well-formed JSON document on standard output as a successful invocation and SHALL derive severity from the parsed report, not from the exit code. |
| R-2.7b | THE SYSTEM SHALL record, per wrapped engine subcommand, whether its exit code means invocation failure or report severity, and SHALL NOT apply one interpretation to both. |
| R-2.8 | IF engine output is not valid JSON or does not match the expected shape, THE SYSTEM SHALL return a parse failure attributed to the engine and SHALL NOT panic, and SHALL NOT attribute the failure to the workspace or its artifacts. |
| R-2.9 | THE SYSTEM SHALL return typed results to callers and SHALL NOT expose subprocess handles, raw output buffers, or operating-system process errors across the package boundary. |
| R-2.10 | THE SYSTEM SHALL pass its full test suite on a machine where the engine is absent from `PATH`, and SHALL NOT require a real engine installation for any test. |
| R-2.11 | THE SYSTEM SHALL exercise, under test, each of: engine absent, engine below minimum version, engine exiting non-zero, engine emitting malformed JSON, and engine exceeding its deadline. |
| R-2.12 | THE SYSTEM SHALL produce identical test and acceptance results whether or not a real engine is installed on the developing machine. |
| R-2.13 | WHILE running its test or acceptance suites, THE SYSTEM SHALL NOT create, modify, or delete any state outside the test's temporary directory, and in particular SHALL NOT mutate the engine's machine-wide repository registry or trigger a real index build. |
| R-2.13a | THE SYSTEM SHALL enforce R-2.13 structurally by redirecting the home directory that locates the engine's registry to a per-test temporary directory, and SHALL NOT rely on convention alone. |
| R-2.13b | THE SYSTEM SHALL run its acceptance suite with a sanitised executable search path, so that acceptance results do not depend on whether a real engine is installed on the machine running it. |
| R-2.14 | THE SYSTEM SHALL bound each engine invocation by a deadline appropriate to that subcommand, and SHALL NOT apply a single global deadline to invocations whose expected duration differs by orders of magnitude. |
| R-2.15 | THE SYSTEM SHALL bound the volume of engine output it retains in memory, and IF that bound is exceeded THE SYSTEM SHALL terminate the invocation and return a failure attributed to the engine rather than accumulating unbounded output. |
| R-2.16 | THE SYSTEM SHALL make per-subcommand deadlines injectable under test, so that deadline behaviour is verifiable without waiting the production deadline in real time. |
| R-2.17 | WHILE exercising the engine-absent case under test, THE SYSTEM SHALL replace the executable search path with a directory known to contain no engine, and SHALL NOT rely on prepending to the inherited path. (Prepending shadows a real engine only for cases that install a fixture; the absent case installs nothing and would otherwise find a developer's real engine, defeating R-2.12 on its own headline case.) |

---

## Unit 3: Workspace registration on `init`

**Why:** Registration is what makes a workspace reachable by the machine-wide
federation. Doing it at scaffold time removes an undocumented manual step — but
it mutates state outside the workspace, so it must be announced and refusable.

**Scoped to `init` only.** An earlier draft included `add`. That was wrong on the
facts: `cmdAdd` receives an already-resolved workspace and adds a platform, team,
or component *inside* it — it never creates a workspace root. Registering on
`add` would therefore re-register an already-registered root on every invocation,
which is not a new capability, only a new way to hit R-3.9 on every call.

| ID | EARS statement |
| --- | --- |
| R-3.1 | WHEN `init` successfully scaffolds a workspace AND the engine is present and satisfies the minimum version, THE SYSTEM SHALL register the workspace with the engine. |
| R-3.2 | THE SYSTEM SHALL pass the absolute workspace root explicitly when registering, so that registration requires no terminal and no interactive confirmation. |
| R-3.3 | WHEN registration succeeds, THE SYSTEM SHALL emit a finding stating that the workspace was registered and naming the registered path. |
| R-3.4 | THE SYSTEM SHALL NOT register a workspace silently. |
| R-3.5 | IF the engine is absent, below the minimum version, or registration fails, THE SYSTEM SHALL complete the scaffold successfully, report the unregistered state as advisory, and print the command that completes registration later. |
| R-3.6 | THE SYSTEM SHALL provide on `init` an explicit opt-out flag that suppresses registration entirely, and WHEN it is supplied THE SYSTEM SHALL NOT invoke the engine at all. |
| R-3.7 | THE SYSTEM SHALL continue to print the next command in the workflow as the final finding of `init`, after any registration finding. |
| R-3.8 | THE SYSTEM SHALL NOT change the exit code of `init` on the basis of registration outcome. |
| R-3.9 | WHEN the workspace is already registered with the engine, THE SYSTEM SHALL treat that as a successful outcome and SHALL NOT present it as a failure or as an advisory defect. |
| R-3.9a | THE SYSTEM SHALL determine prior registration by reading the engine's repository registry file before invoking registration, and SHALL treat that file as a declared, version-pinned coupling to engine internals rather than as a scrape. |
| R-3.9b | THE SYSTEM SHALL verify, as part of the version assertion, that the engine's registry file is in the format this coupling expects, and IF it is not, THE SYSTEM SHALL skip registration and report the workspace as unregistered under R-3.5 rather than misreporting its state. |
| R-3.10 | WHERE registration causes the engine to build an index synchronously, THE SYSTEM SHALL apply a deadline sized for indexing rather than the default invocation deadline, SHALL tell the user that indexing is in progress before it begins — emitted through the writer the command layer already supplies, never printed from within the adapter, and suppressed when output is machine-readable — and IF the deadline is reached THE SYSTEM SHALL report the workspace as registered with indexing incomplete and name the command that resumes it — never as a failed registration. |
| R-3.11 | THE SYSTEM SHALL supply the engine with a registration identity derived from the workspace root path such that two workspaces with the same directory basename do not collide. |
| R-3.12 | THE SYSTEM SHALL determine registration outcome from the engine's exit status, its machine-readable output, and the registry-file read declared in R-3.9a — and SHALL NOT parse the engine's human-readable diagnostic text to distinguish "already registered" from other failures. (The engine exits 1 identically for "already registered", an unresolvable path, and a filesystem failure, and `repo add` has no JSON mode, so exit status alone cannot separate them. The registry read is what makes this clause satisfiable; without it R-3.9 and R-3.12 are jointly unsatisfiable.) |
| R-3.13 | WHERE the registry is read before registration and written by the engine during it, THE SYSTEM SHALL treat the intervening window as racy: IF registration fails after the pre-check reported the workspace absent, THE SYSTEM SHALL re-read the registry before reporting, and IF the workspace is present on re-read THE SYSTEM SHALL report success under R-3.9. |
| R-3.14 | THE SYSTEM SHALL NOT present the registry-file pre-check as authoritative evidence of registration; only a completed registration or a post-failure re-read (R-3.13) SHALL satisfy R-3.3. |

**R-3.9a, R-3.9b and R-3.13 are provisional.** They exist solely because `repo add` has no
`--json` and routes every failure through one `die` helper. The upstream fix is tracked as
**#1a** in `.devlocal/research/2026-08-23-local-search-improvements-backlog.md` (effort S).
If it lands before this unit is implemented, delete all three clauses and the pre-check
with them, and satisfy R-3.12 from the engine's machine-readable outcome alone. Do not
harden the registry coupling further while that item is open.

---

## Unit 4: Engine drift surfaced by `check`

**Why:** Engine drift should be discovered while the user is already reading
output, not at the moment they urgently need a retrieval answer. `check` is that
moment — but its exit code gates completion, so the report must be advisory.

| ID | EARS statement |
| --- | --- |
| R-4.1 | WHEN `check` runs AND the engine is present, THE SYSTEM SHALL report the installed engine version, whether it satisfies the declared minimum, and whether the engine's index is stale. |
| R-4.2 | THE SYSTEM SHALL make at most one invocation per wrapped engine subcommand per `check` run, each bounded by its own deadline, and SHALL reuse the process-cached version probe rather than repeating it. (`check` makes two invocations, not one: the version probe cannot be replaced by the report's embedded version field, because the report is the JSON contract whose availability the version probe exists to establish — see D2.) |
| R-4.3 | WHERE the engine is present, THE SYSTEM SHALL render every engine condition — below minimum, index stale, timed out, malformed output, non-zero exit — as advisory output. Engine absence is reported by `init` under R-3.5, not by `check` (R-4.5a). |
| R-4.4 | THE SYSTEM SHALL NOT allow any engine condition to produce a failing gate result or to change the exit code of `check`; the exit code SHALL be determined solely by the product readiness and completion gates. |
| R-4.5 | WHEN `check` reports a version-floor breach or a stale index on an engine that is present, THE SYSTEM SHALL print the upgrade or re-scan command. |
| R-4.5a | WHILE the engine is absent, THE SYSTEM SHALL emit no engine finding and no install suggestion from `check`. (Until a command exists that consumes the index, an install prompt in the readiness/completion path advertises a capability the system cannot yet use, to the majority population, on a command they run constantly. Revisit when `uncle ask` lands.) |
| R-4.6 | THE SYSTEM SHALL add the engine report additively and SHALL NOT renumber, re-slug, or reorder any existing gate. |
| R-4.7 | WHILE the engine is absent, THE SYSTEM SHALL leave `check` output byte-identical to its output prior to this change. |
| R-4.8 | THE SYSTEM SHALL expose the engine report to `check` through an entry point that returns findings only and cannot return an error, so that no engine condition has a path by which it can propagate as a command-level failure. |
| R-4.9 | THE SYSTEM SHALL verify by test that `check` returns the same exit code with the engine absent, healthy, reporting warnings, reporting failures, below the minimum version, and hanging until its deadline. |

---

## Traceability

Requirements in this document are bound to implementation and tests by grep-able
markers of the form:

```
@spec req://uncle-os/retrieval-adapter-handshake@0.1#R-2.5
```

Per the ontology rule, a mandatory clause with no test-side `@spec` site is not
considered satisfied. R-1.4, R-1.10, R-2.10, R-4.4, and R-4.7 are the
load-bearing absence invariants and each requires a test-side marker. R-2.7a,
R-3.10, and R-4.8 encode pre-mortem findings against verified engine behaviour
and carry the same obligation. R-2.17, R-2.13a, and R-2.13b are the hermeticity
clauses — without a marker on each, the suite silently becomes machine-dependent,
which is the failure mode least likely to be noticed by the person who caused it.

**R-1.9 is the one clause exempt from the marker obligation in this slice**, and
only because no code path in scope can reach it (see its note). If a later slice
makes a command require the engine, R-1.9 acquires the obligation with it.
