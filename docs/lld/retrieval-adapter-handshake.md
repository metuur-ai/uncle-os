---
type: lld
id: lld-retrieval-adapter-handshake
title: Retrieval Adapter Handshake — Low-Level Design
status: draft
tags: [kind/lld, status/draft]
---

# Retrieval Adapter Handshake — Low-Level Design

## Architecture

### New package: `internal/retrieval`

A leaf package. It imports `internal/model` (for exit codes and errors) and the
standard library, and nothing else from the module. Nothing below `cmd/` prints
or exits, per the existing module convention, so the adapter returns typed
errors and lets renderers speak.

```
internal/retrieval/
  engine.go     — Engine handle, LookPath, RequireEngine, version assert + cache
  exec.go       — run(ctx, args...) ([]byte, error); timeout, cancellation, stderr capture
  doctor.go     — Doctor() (DoctorReport, error)   — wraps `doctor --json`
  explain.go    — Explain(id) (ExplainResult, error) — wraps `graph explain --json`
  engine_test.go / exec_test.go — table tests over a fake engine binary
```

The public surface is deliberately four things:

```go
type Engine struct{ ... }                 // opaque; carries resolved path + cached version
func Discover(ctx context.Context) (*Engine, error)   // returns ErrEngineAbsent if not on PATH
func (e *Engine) Version() Version                    // cached, never re-shells
func (e *Engine) Doctor(ctx context.Context) (DoctorReport, error)
func (e *Engine) Explain(ctx context.Context, id string) (ExplainResult, error)
```

`Discover` is the only place `exec.LookPath` is called for the engine.
`ErrEngineAbsent` is a sentinel, so callers write `errors.Is(err,
retrieval.ErrEngineAbsent)` and choose a policy — they never inspect
`*exec.Error` or string-match "not found".

### The `RequireGit` precedent

`internal/federation/git.go` already solves this exact problem for `git` and is
the template, not merely an analogy:

| Concern | `git.go` (existing) | `retrieval` (new) |
|---|---|---|
| Availability | `gitAvailable()` → `exec.LookPath` | `Discover` → `exec.LookPath` |
| Version floor | `MinGit = [2]int{2, 27}` | `MinEngine` var, same shape |
| Version parse | regex over `git --version` **human text** | regex over `local-search -v` |
| Failure class | `model.ExitExternalTool` (6) | `model.ExitExternalTool` (6) |
| Guard placement | `RequireGit()` called at the *operation* (`sync.go:94`), not per command | `RequireEngine()` at the retrieval call site |
| Message shape | states what needs it, what does not, and the next command | identical |

Two consequences worth stating explicitly. First, `git.go:3-4` currently claims
git plumbing is "The only subprocess users in the whole module" — that comment
becomes false and must be amended in the same change. Second, `RequireGit` has
exactly one caller; the guard lives where the capability is *used*, so commands
that merely read a lock file are unaffected. `RequireEngine` inherits that
discipline: `init` and `check` must degrade, so neither calls it as a hard gate.

### Data flow

```
cmd/company-os/{scaffold,product}.go
        │  (policy: what does absence mean here?)
        ▼
internal/{scaffold,product}
        │  (typed call, context with deadline)
        ▼
internal/retrieval  ──exec──▶  local-search --json
        │
        ▼  DoctorReport / ExplainResult / ErrEngineAbsent / ErrEngineTooOld
```

Policy lives at the top, mechanism at the bottom. The adapter never decides
whether absence is fatal; it only reports absence.

### Unit 3 — registration on `init`

**`add` does not register, and the earlier draft saying it did was wrong.**
`cmdAdd` (`scaffold.go:113`) receives an already-resolved `workspace.Workspace`
and calls `scaffold.Add` (`scaffold.go:117`) to add a platform, team, or component
*inside* an existing workspace. It never creates a workspace root, so there is
never a new path for the engine to learn. Registration is scoped to `init` alone.

`cmdInit` (`cmd/company-os/scaffold.go:63`) delegates to `scaffold.Init` and then
builds a `GateResult` whose findings end with `CodeInitNext` — the guidance-chain
line. Registration rides along there:

- After a successful scaffold, if `Discover` succeeds, run
  `local-search repo add <absolute workspace root>`.
- The explicit path matters. `confirmRepoAdd` (`local-search cli/main.go:269-275`)
  only prompts when it must *infer* the cwd; passing the path takes the
  non-interactive branch. No TTY is required, so `ExitInteractive` is not in
  play and engine-side `repo add --yes` is not a prerequisite.
- **`repo add` is not a cheap registration call.** Verified: it persists the
  entry (`local-search cli/main.go:242`) and then runs a *synchronous* index build
  (`cli/main.go:261`, body at `cli/main.go:988-1010`). On a large workspace this is
  minutes, not milliseconds. Registration therefore gets its own indexing-sized
  deadline, announces itself *before* it blocks, and treats deadline expiry as
  "registered, indexing incomplete" — the entry is already persisted at that
  point, so calling it a failed registration would be a lie.
- Registration emits its own finding before `CodeInitNext`, so the user is told
  it happened. Silent side effects on someone's machine-wide federation are not
  acceptable.
- **The engine names repos by directory basename** (`cli/main.go:210`,
  `filepath.Base`) and rejects a name-or-path collision (`cli/main.go:220`,
  `cli/main.go:2980-2987`). Two workspaces called `platform` would collide, so we
  pass an explicit identity derived from the full root path.
- Absence, or a non-zero `repo add`, downgrades to an advisory finding. `init`
  still succeeds; a workspace that exists but is unregistered is a recoverable
  state with a printed remedy.
- **"Already registered" is detected by a pre-check, not by interpreting the
  failure.** This is forced, not chosen. `repo add` has no `--json` mode, and
  `die` (`cli/main.go:3007-3010`) exits 1 for *every* failure — already-registered,
  bad path, mkdir failure. Exit status cannot separate them and the failure prose
  is the only distinguishing signal. So before invoking `repo add`, the adapter
  reads the engine's registry file and checks whether this absolute root is
  already present; if it is, registration is reported as an ok outcome and
  `repo add` is not invoked at all.

  This is a **declared, versioned coupling to an engine data file** — a real cost,
  recorded here rather than buried. It is the lesser of the two evils: the
  alternative is scraping a human-readable string that the engine is free to
  reword in any release. When engine-side `repo add --json` lands, the pre-check
  is deleted and the coupling goes with it.

  The read is a pre-check, so it races: two `company-os init` runs on the same
  root, or an init racing a user's `local-search scan`, can both observe "absent"
  and both call `repo add`. The engine's own name/path collision check
  (`cli/main.go:220`) is the backstop; a `repo add` that fails *after* a
  pre-check said absent is reported as an advisory finding with the remedy, never
  as an init failure. Company OS does not lock a file it does not own.
- `--no-engine` opts out of registration entirely, for CI and for users who do
  not want their workspace in the machine-wide index. `examples/acceptance.sh`
  passes it on every `company-os init`, so the acceptance run cannot touch the
  developer's machine-wide registry. That flag plumbing is part of this change —
  today `acceptance.sh` runs `init` with no such flag, and the earlier draft
  described the guard as though it already existed.

**Announcing before blocking, without breaking the print convention.** R-3.10
requires the user to be told indexing is starting *before* it starts, but nothing
below `cmd/` may print, and a `GateResult` is only returned after the command
completes. A finding cannot warn ahead of a call that has not returned. The seam:
`scaffold.Init` accepts an optional progress sink — the same shape as the existing
`stdinPrompt(promptWriter(args, out))` precedent at `scaffold.go:68`, which already
threads an `io.Writer` down for interaction. `cmd/` supplies a writer; the adapter
writes one line to it immediately before the blocking `repo add` and nothing after.
Under `--json` the sink is nil and the line is suppressed, so machine-readable
output stays clean. No package below `cmd/` acquires a print statement of its own;
it writes to a writer its caller handed it.

### Unit 4 — engine drift in `check`

`cmdCheck` (`cmd/company-os/product.go:52`) is a three-line delegation to
`product.Check(ws, team, components, kind)`, which returns `[]model.GateResult`.
The engine report is appended as one additional `GateResult` carrying up to three
findings: version, floor satisfaction, index staleness. It is sourced from
`Doctor()` — one call, one deadline.

**`doctor` reports severity through its exit code.** Verified:
`local-search cli/doctor.go:82-87` exits 2 on failures and 1 on warnings, and
index staleness is emitted as a *warning* (`cli/doctor.go:271`). A naive "non-zero
means the invocation failed" rule would convert the single fact Unit 4 exists to
report into "engine failed". The adapter therefore parses well-formed JSON on
stdout regardless of exit status for this subcommand, and records per-subcommand
which interpretation applies.

The critical detail: `check` is the `ready` / `done` gate. Its exit code is
load-bearing for the completion workflow. **No engine condition may contribute a
`[FAIL]`.** Absent, too-old, timed-out, and malformed-JSON all render as `[WARN]`
and leave the exit code determined solely by the product gates.

## Constraints

- **Go module, static binary, no runtime dependency.** The engine is discovered
  at runtime or not at all; it is never a build- or link-time dependency.
- **Nothing below `cmd/` may call `os.Exit` or print** (module convention). The
  adapter returns `model.Error` values carrying an exit code.
- **`frontmatter()` parser contract and the guidance chain (R-1.8) are
  non-negotiable.** Every mutating command prints its next command — registration
  in `init` is a mutation of machine state and is therefore announced.
- **Exit-code map is a published contract** (`internal/model/model.go:177-199`,
  characterised by `cmd/company-os/exitcode_test.go` and
  `internal/model/model_test.go:418`). Codes may be *used* in new places; their
  meanings may not be redefined.

  This change **does** edit that contract, and schedules the edit explicitly:
  `model.go:192-193` documents `ExitExternalTool` as *git-scoped* — "git is
  missing or too old, a clone/sparse-checkout failed, or `--frozen` lock
  reconciliation failed". Reusing code 6 for the engine widens that meaning, so
  the doc comment is amended in the same change, alongside the `git.go:3-4`
  amendment above. `TestExitCodeContract` asserts numeric values only and will not
  catch the drift; since the comment is declared a contract, the amendment is
  reviewed as one rather than slipped in.
- **`make check` is the gate**: gofmt + vet + `go test ./...` +
  `examples/acceptance.sh`, plus `company-os validate` exiting 0 on
  `examples/workspace`.
- **Tests must pass on a machine with no `local-search`.** This forbids tests
  that shell out to a real engine; the fake-binary fixture is the only route.

## Key Decisions

**D1 — Missing/too-old engine is `ExitExternalTool` (6), not `ExitUsage` (2).**
This overrides the backlog's recommendation to "reuse the `notImplemented`
reasoning at `commands.go:57-67` — not actionable is exit 2". Verified against
the source: `ExitExternalTool` is documented as "git is missing or too old, a
clone/sparse-checkout failed" — a missing or too-old `local-search` is the same
category, and `RequireGit` already returns 6 for precisely "not found on PATH"
and "too old". Exit 2 means *bad flags or unknown subcommand*; a correct
invocation on a machine lacking an optional tool is not a usage error, and
conflating them would make the code useless for its stated purpose. Rejected
alternative: a new `ExitEngine` code — the map is a published contract and 6
already covers it. *Note this decision is scoped to future retrieval commands;
within this change's scope (#3, #4) nothing exits non-zero on absence at all.*

**D2 — Parse the `-v` line; one bounded exception, with precedent.** Scraping
human output is otherwise forbidden. The exception holds because the version line
is a documented literal and because it is what *establishes* whether the JSON
contract exists — a chicken-and-egg the JSON path cannot resolve. This is not a
novel concession: `gitVersion()` (`git.go:67-84`) regex-scrapes `git --version`
today for the same reason. Migration to `version --json` when engine #2 lands is
a one-function change behind `Version()`. Everywhere else, if `--json` is
unavailable, the version assert fails — no scraping fallback.

**D3 — Version cached per process, asserted on first use, not at start-up.**
Asserting at start-up would shell out on every `company-os` invocation including
`--help`, taxing the majority who have no engine. Lazy + cached costs the users
who use retrieval and nobody else.

**D4 — Absence is a sentinel error, not a nil `*Engine`.** A nil handle invites
nil-deref at each call site and makes "absent" indistinguishable from "not looked
yet". `errors.Is(err, ErrEngineAbsent)` forces each caller to state a policy.

**D5 — Every invocation is context-bounded with a default deadline.** A wedged
engine must not hang `check`. Timeout renders as an engine fault (6 where an exit
code is produced at all), never as a workspace or artifact fault, because nothing
about the workspace is wrong.

**D6 — Registration is announced and opt-out-able, never silent.** `repo add`
mutates machine-wide state outside the workspace. The guidance chain already
requires mutating commands to speak; this extends the same courtesy to a side
effect the user did not directly request.

**D7 — Engine findings in `check` are advisory-only.** Directly enforces the HLD
invariant. If a missing optional tool could fail the done-gate, Company OS would
have acquired a hard dependency by the back door.

**D8 — Fake-binary fixture for tests.** Tests build a tiny Go binary into
`t.TempDir()` and have it emit canned JSON, canned version lines, non-zero exits,
malformed JSON, and a deliberate hang. This exercises the real `exec` path with no
real engine and no network.

Three details are load-bearing, and the earlier draft got the first one wrong:

- **`PATH` is *replaced* with the fixture directory, not prepended.** Prepending
  shadows a real engine for the present case but does nothing for the
  engine-*absent* case: a test that installs nothing still finds the developer's
  real engine and silently tests the wrong thing. Since the absence invariants
  (R-1.4, R-1.10, R-4.4, R-4.7) are the most important clauses in this spec, the
  case that must be hermetic is exactly the one prepending leaves exposed.
  `t.Setenv("PATH", <fixture dir or empty>)` gives identical results on a machine
  with and without a real engine. Note `t.Setenv` forbids `t.Parallel()` in these
  packages — accepted, the suite is small.
- **`HOME` is redirected to `t.TempDir()` too.** The engine resolves its registry
  at `~/.local-search/repos` (`cli/main.go:33-34`). R-2.13's "SHALL NOT mutate
  machine-wide state" is otherwise enforced only by convention, and the blast
  radius is writing junk into a developer's real registry and triggering real
  scans. A guard test asserts `HOME` is a temp dir.
- **Deadlines are injectable.** R-2.11's deadline test would otherwise cost real
  wall-clock and be flaky under load. Timeouts are fields on the `Engine` handle,
  defaulted at construction, overridden in tests.

`examples/acceptance.sh` runs `company-os init --no-engine` for the same
hermeticity reason (see Unit 3). It exercises the real binary with a real `PATH`,
so the opt-out flag — not the fixture — is what protects it.

**D9 — The advisory-only rule gets a structural guard, not discipline.** The
`check` entry point into the adapter returns findings and *no error*, so there is
no code path by which an engine condition can reach `main.go`'s error branch —
which is evaluated before `HasFailure` and would otherwise exit 6. Stating the
rule in prose is insufficient: `git.go:44-49` teaches the opposite reflex
(return `model.Errorf` on subprocess trouble), and one idiomatic `if err != nil
{ return nil, err }` in `product.Check` would block `prd complete` on every
machine without the engine. An exit-code-identity test across all engine states
locks it.

**D10 — Per-subcommand deadlines, not one global timeout.** `doctor` is
sub-second; `repo add` triggers a full index build. A single deadline either
hangs `check` or aborts registration mid-index. Each wrapped subcommand declares
its own.


## Out of Scope

- `internal/retrieval` `find` / `search` result types (#1b) — blocked on engine.
- `uncle ask` (#2) and the 17-verb → 6-verb surface cut (#10).
- `check --fix` (#5), Obsidian vault shape (#6–#9), vocabulary renames (#12–#13).
- EARS clause numbering (#14) and `@spec` resolution (#15–#16). Unit 4 here
  reports *engine* drift only; it does not compute clause coverage.
- Engine-side changes: `repo add --yes`, `version --json`, unified `find`.
- Any change to existing gate ordinals, slugs, or exit codes.
