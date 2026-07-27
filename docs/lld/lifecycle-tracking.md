# Lifecycle Tracking — Low-Level Design

## Architecture

Four additions. Three extend the gate that already evaluates dated artifacts;
one adds a flag. No new subcommand, no new gate, no new package.

The three gate additions share one shape — read a date from frontmatter, compare
it to the run's single current-date seam, emit a warning with the delta — so they
share the seam, the tolerant parse, and the field convention rather than each
inventing its own.

### Part 1 — a warning band in the expiry gate

`internal/governance/gates.go` currently evaluates each dated commitment as a
two-way branch. Deviations:

```
past  := before(rd, now)      -> SevFail  CodeDeviationExpired
otherwise                     -> SevOK    CodeDeviationCurrent
```

Exceptions carry a third case already, for a missing `expires`:

```
falsy(ex)                     -> SevFail  CodeExceptionNoExpiry
past := before(ex, now)       -> SevFail  CodeExceptionExpired
otherwise                     -> SevOK    CodeExceptionValid
```

Each grows one branch between `past` and the OK case: the date is in the future
but within the window. Two new finding codes, both `SevWarn`, both carrying the
date and the remaining day count in `Fields` so a JSON consumer can sort by
urgency without parsing prose.

The existing `before()` helper already parses the date and reports a malformed
one as an artifact error, so the new branch reuses it and adds only a day-delta
computation. Neither existing code, message, nor severity changes — a past-due
commitment fails exactly as it does today.

**The window is a constant, not a flag.** A per-invocation `--window` would let
CI pass by narrowing it, which turns a governance signal into a preference.
Thirty days is the default and lives beside the codes it produces.

### Part 2 — the outcome-review check

`internal/roles/today.go:128-155` already walks `<platform>/archive/prds/*/outcome.md`,
loads its frontmatter, skips anything whose `status` is not `pending`, and emits
`CodeOutcomeReview` at `SevOK` with `due` rendered as a raw string. That walk is
the whole mechanism; what is missing is a comparison — and a tolerant one. The
value has never been parsed, so unlike a deviation's `reviewDate` it carries no
guarantee of being a date at all. It gets its own parse that degrades an
unparseable or absent value to a warning, rather than the shared `before()`
helper that raises an artifact error.

The check lives in the expiry gate rather than in `today` or in a gate of its
own. `today` is a view — it reports and never verdicts, and adding a severity
there would make role-filtered output the place governance failures hide. A
separate gate would renumber the sequence for one finding type. The expiry gate
is already "dated commitments that came due", and an outcome review is precisely
that — which is also why its title stops being accurate and has to change.

Severity is `SevWarn`, deliberately and permanently. `Problems()`
(`internal/model/model.go:148-160`) excludes warnings, so an overdue review
reports without failing the build. This is the difference between a check people
keep and a check people delete.

`today` keeps its own `SevOK` listing unchanged. The two are not redundant:
`today` answers "what reviews are outstanding", the gate answers "which of them
are late".

### Part 3 — `today --team`

`today` currently takes only `--role`. `internal/roles/today.go` iterates every
platform via `platformSection` and every team via `teamSection`. The flag filters
which sections are emitted:

- the named team's `teamSection`, and
- each `platformSection` for a platform owning at least one component the team
  claims in `teams/<t>/ownership/components.yaml`.

The platform filter matters. Filtering to the team alone would hide the active
PRDs and outcome reviews that are the team's actual work, since both live under
`platforms/<p>/`, not under the team.

**The mapping comes from the descriptors, not the generated file.** The
ownership registry lists `{id, relationship, repository}` and names no platform.
The component-to-platform edge lives in two places: `components.<id>.platforms[]`
inside `teams/<t>/generated/effective-governance.yaml`, and
`platforms/<p>/components/<id>.yaml`. The generated file is the tempting read —
it is one file and already loaded by `teamSection` — and it is the wrong one. It
is derived, it can be stale, and `teamSection` already has a branch for it being
missing entirely. A team whose governance has not been resolved would get a view
with no platform sections at all: not an error, just an empty answer that reads
as "nothing in flight". The descriptor is authoritative for this relationship by
standing invariant, so the filter scans descriptors and a team that owns
components but resolves to no platform says so.

An unknown team id is an argument error naming the unknown id and suggesting
close matches, matching how `prd new` already handles an unknown team.

### Part 4 — stale in-flight work

Two more walks over dated frontmatter: `platforms/<p>/change-records/active/*/prd.md`
at `status: proposed`, and `teams/<t>/product/discovery/*/brief.md` at
`status: draft`. Both carry `created:`. A brief at `validated` is deliberately
exempt — it is waiting on a PRD, not stalled.

The threshold is longer than Unit 1's window and likewise hard-coded, for the
same reason: a tunable staleness bar is a bar that gets tuned to silence.

`examples/workspace` has an empty `change-records/active/` and its one brief is
`validated`, so this fires on nothing in the acceptance corpus — the same
coverage caveat as the rest of this change, carried by synthesized workspaces.
The banking PRDs carry `created: 2026-07-21` and would age into the threshold,
which is why Unit 4's fixture-date invariant has to cover `created:` and not only
`due:`.

## Constraints

**All fixture expiry dates are far-future, so the warning window fires on
nothing.** Every live deviation and exception in `examples/workspace` and
`examples/federated` carries 2035 dates; `examples/failing-workspace` carries
deliberately stale 2020 ones. With a 30-day window, no fixture enters the
warning band and **no golden changes**. That is convenient and also a coverage
hole: the band ships with zero fixture coverage unless a test synthesizes a
workspace with a date inside it. It must.

**One fixture is a dated time bomb.**
`examples/workspace/platforms/communications/archive/prds/2026-per-channel-quiet-hours/outcome.md`
is `status: pending, due: 2026-10-16`. Today is 2026-07-27. Under the new check
that fixture goes overdue on its own, roughly 81 days from now, and
`examples/golden-validate.txt` starts failing with no code change — a CI break
that arrives by calendar and looks like a regression. The repository already has
a convention for dates that must stay valid: every deviation and exception uses
2035. This fixture's `due` must be moved into 2035 to match. The banking fixture
carries the same shape (`due: 2026-10-18`) but is not validated by
`acceptance.sh`; it is moved anyway so the corpus is consistent.

**`today` is not golden-pinned.** `acceptance.sh` invokes `validate`,
`graph build`, `workspace sync`, `workspace status`, and `sync --frozen` — never
`today`. The `--team` flag therefore has no golden impact and needs Go coverage
to be tested at all.

**The clock is not injectable, and the spec demands boundary tests.**
`today()` (`internal/governance/declare.go:184-185`) is a package-private
`time.Now()` wrapper with no seam, and `ExpiryGate` calls it internally. A test
asserting "29 days out warns, 31 days out does not" would have to compute its
fixture dates from its own `time.Now()` and hope both reads land on the same
calendar day — which they will not, reliably, at a window boundary or across
midnight. A substitutable current-date seam is therefore a precondition of the
coverage this change requires, not an optional nicety.

**Adding outcome reviews makes the gate's title false.**
`ExpiryGateTitle = "deviation and exception expiry"`
(`internal/governance/gates.go:29`) is a constant, rendered into all five
goldens as `[2/N] deviation and exception expiry` and asserted three times in
`internal/model/model_test.go`. Reporting outcome reviews from that gate under
that title mislabels them. The title changes; the slug `governance-expiry` does
not, because a slug is a JSON contract and a title is prose. That splits the
churn into golden and test updates without breaking a consumer.

**A malformed date currently cannot reach the gate, and must not start to.**
`before()` returns an artifact error on an unparseable date, and callers
propagate it — that is exit 4, not a finding. Deviations and exceptions are
already validated against it. Outcome reviews are not: `today` renders an absent
`due` as the literal string `"None"` for Python compatibility, and nothing has
ever parsed the value. Routing `due` through `before()` unchanged would make a
workspace with one typo'd outcome date start exiting 4 where it exits 0 today —
a regression against this change's own non-regression goal. The outcome path
needs its own tolerant parse that degrades to a warning.

**Warnings must not change exit codes.** `main.go:128-137` maps
`model.HasFailure(results)` to `ExitValidation`. Warnings are outside
`Problems()`, so exit behavior is unchanged by construction — but any new finding
added at `SevFail` by mistake silently converts a reporting change into a
breaking one.

**The subcommand surface is pinned byte-for-byte.** Adding `--team` to `today`
changes its per-subcommand usage string, which `args_test.go:313-339` asserts by
exact match, and the parser count guard at `args_test.go:168-180` must still
agree. This is mechanical but not optional.

**Nothing under `internal/` may print or exit.**
`cmd/company-os/architecture_test.go:24-42` enforces it. All three additions
return `model.GateResult` or `Finding` values and compose their own `Message:`,
as the surrounding code already does.

## Key Decisions

**Overdue outcome reviews warn, never fail.** The strongest argument for failing
is that the methodology calls the review part of "done". The stronger
counter-argument is operational: an unreviewed outcome blocks a build that has
nothing to do with the outcome, on a repository the reviewer may not own. Checks
that block unrelated work get disabled, and a disabled check reports nothing.
A warning that stays enabled beats a failure that gets removed.

**The window is 30 days and hard-coded.** Configurable urgency is a way to make
a signal disappear. If 30 proves wrong the constant moves in one place, which is
a smaller change than deprecating a flag people have set to 1.

**The outcome check joins the expiry gate rather than becoming gate 9.** One
finding type does not justify renumbering the sequence, and the expiry gate's
subject is already "dated commitments". This also keeps the change independent of
the `cos-` namespace work, which does add a gate — the two can land in either
order without colliding.

**`today` reports; `validate` verdicts.** The overdue comparison could have gone
into `today`, which already does the directory walk. Keeping severity out of
`today` preserves a clean split: a view that never fails anything, and a gate
that is the single place a workspace's health is decided. It also avoids
governance findings being hidden behind `--role`.

**`--team` widens to the team's platforms rather than filtering to the team
directory.** A team's work product lives under `platforms/<p>/change-records/`
and `platforms/<p>/archive/`, not under `teams/<t>/`. A literal directory filter
would produce an empty, technically-correct view.

**Rejected: a `--window` flag on `validate`.** See above — it converts a
governance signal into a per-run preference and gives CI a way to pass by
narrowing it.

**Rejected: a `prd list` / `discover list` command.** `today` already enumerates
active PRDs per platform; with `--team` it answers the same question the new
command would, without adding a subcommand to a surface pinned by exact-match
tests.

## Out of Scope

- Any change to how a deviation, exception, or outcome review is created.
- Notification, scheduling, or anything that runs without the user invoking it.
- A history or changelog command over `log.md`.
- Making an overdue outcome review fail the build, now or behind a flag.
- Renaming, renumbering, or altering any existing finding code, message, or
  severity.
- `--team` on any command other than `today`.
