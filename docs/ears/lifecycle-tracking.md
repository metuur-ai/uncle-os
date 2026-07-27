# Lifecycle Tracking — EARS Specifications

## Unit 1: Expiry warning window

**Why:** A deviation's `reviewDate` and an exception's `expires` both encode
"we will revisit this by this day". The gate treats them as tripwires — silent
until past, then a hard failure — so a team's first signal that a governance
commitment came due is a red build. A warning band gives lead time without
weakening the failure.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL define a fixed warning window of 30 days for dated governance commitments. |
| R-1.2 | IF a deviation's `reviewDate` is in the future and within the warning window, THEN THE SYSTEM SHALL emit a warning naming the team, the rule, the date, and the number of days remaining. |
| R-1.3 | IF an exception's `expires` is in the future and within the warning window, THEN THE SYSTEM SHALL emit a warning naming the team, the rule, the date, and the number of days remaining. |
| R-1.4 | THE SYSTEM SHALL emit each warning under a finding code distinct from every code already present in `internal/model/codes.go`. |
| R-1.5 | THE SYSTEM SHALL carry the date and the remaining day count as separate typed fields, so a consumer can rank by urgency without parsing the message text. |
| R-1.6 | WHERE a commitment's date is in the past, THE SYSTEM SHALL emit the existing failing finding with its existing code, message, and severity unchanged. |
| R-1.7 | WHERE a commitment's date is beyond the warning window, THE SYSTEM SHALL emit the existing passing finding with its existing code, message, and severity unchanged. |
| R-1.8 | IF an exception has no `expires` value, THEN THE SYSTEM SHALL emit the existing failing finding and SHALL NOT evaluate the warning window. |
| R-1.8a | IF a deviation has no `reviewDate` value, THEN THE SYSTEM SHALL emit the existing passing finding unchanged and SHALL NOT evaluate the warning window. |
| R-1.9 | THE SYSTEM SHALL NOT expose the warning window as a command-line flag, an environment variable, or a workspace configuration value. |
| R-1.10 | WHEN a commitment's date is malformed, THE SYSTEM SHALL report the existing artifact error and SHALL NOT emit a warning. |
| R-1.11 | THE SYSTEM SHALL provide a seam, substitutable from within the package that owns the expiry gate, that lets a test replace the current date that gate reads, so window-boundary behaviour is assertable without depending on when the suite runs. |
| R-1.11a | THE SYSTEM SHALL route every date comparison in that gate through the one seam, so no two findings in a single run can disagree about what day it is. |
| R-1.12 | THE SYSTEM SHALL leave the production current-date source unchanged when no substitute is supplied. |

## Unit 2: Outcome-review overdue check

**Why:** `prd complete` archives a change and schedules an outcome review 90 days
out. Nothing compares that date to the calendar, so the loop the methodology
defines — change, archive, review the outcome — has no closing signal. A
workspace can carry reviews eighteen months overdue and pass every gate.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL evaluate every `outcome.md` beneath a platform's archived-PRD directory during validation. |
| R-2.1a | THE SYSTEM SHALL walk that directory once, through a single shared traversal used by both the validation check and the daily view, and SHALL NOT maintain two independent walks with two date policies. |
| R-2.2 | IF an outcome review's `status` is `pending` and its `due` date has passed, THEN THE SYSTEM SHALL emit a warning naming the PRD, the platform, and the due date. |
| R-2.3 | THE SYSTEM SHALL emit that finding at warning severity, and SHALL NOT cause a workspace to exit non-zero on account of an overdue outcome review. |
| R-2.4 | IF an outcome review's `status` is any value other than `pending`, THEN THE SYSTEM SHALL emit nothing for it. |
| R-2.5 | IF an outcome review has no `due` value, THEN THE SYSTEM SHALL emit a warning reporting the absent date rather than treating the review as current. |
| R-2.5a | IF an outcome review's `due` value is present but not a parseable date, THEN THE SYSTEM SHALL emit a warning naming the unparseable value, and SHALL NOT raise an artifact error. |
| R-2.5b | THE SYSTEM SHALL NOT cause a workspace holding a malformed or absent outcome-review `due` value to change its exit code. |
| R-2.6 | THE SYSTEM SHALL report this finding from the gate that already evaluates dated governance commitments, and SHALL NOT introduce a new gate for it. |
| R-2.6a | THE SYSTEM SHALL retitle that gate so its title names every dated artifact it now evaluates, rather than only deviations and exceptions. |
| R-2.6b | THE SYSTEM SHALL leave that gate's slug unchanged, so no JSON consumer keyed on it breaks. |
| R-2.6c | WHEN the gate title changes, THE SYSTEM SHALL update the golden snapshots and the rendered-gate test assertions that pin it. |
| R-2.7 | THE SYSTEM SHALL leave the existing outcome-review listing in the daily view unchanged in finding code and severity. |
| R-2.7a | WHERE the daily view lists an outcome review whose due date has passed, THE SYSTEM SHALL render it as overdue, so the surface a human reads does not present a long-overdue review identically to a current one. |
| R-2.7b | THE SYSTEM SHALL NOT introduce any failing severity into the daily view. |
| R-2.8 | THE SYSTEM SHALL carry the due date and the number of days overdue as separate typed fields. |

## Unit 3: Team-scoped daily view

**Why:** `today` takes only `--role`, so on a workspace with several platforms it
prints all of them. The question a person actually asks is "what is in flight for
my team", and the answer currently requires reading past everyone else's.

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL accept an optional `--team` argument on the daily-view command. |
| R-3.2 | WHEN `--team` is supplied, THE SYSTEM SHALL emit that team's section among the sections the requested role already emits, and omit every other team's. |
| R-3.2a | WHERE the requested role emits no team sections at all, THE SYSTEM SHALL NOT begin emitting one because `--team` was supplied. |
| R-3.3 | WHEN `--team` is supplied, THE SYSTEM SHALL emit the section for each platform owning at least one component the named team appears against, and omit every other platform's. |
| R-3.3a | THE SYSTEM SHALL derive the component-to-platform mapping from each descriptor's declared platform-relationship entries, and SHALL NOT derive it from the directory a descriptor happens to sit in, nor from any generated file. |
| R-3.3c | WHERE a descriptor declares more than one platform relationship, THE SYSTEM SHALL treat every named platform as in scope, matching how effective governance already reads the same field. |
| R-3.3b | IF the named team owns at least one component but no platform section results, THEN THE SYSTEM SHALL emit a warning finding naming the team and the count of components that resolved to no platform, rather than emitting an empty view. |
| R-3.4 | WHEN `--team` is omitted, THE SYSTEM SHALL emit exactly the sections it emits today, in the same order. |
| R-3.5 | IF the named team does not exist, THEN THE SYSTEM SHALL fail with the same exit status and the same message shape that scaffolding a PRD already uses for an unknown team, and SHALL NOT introduce close-match suggestions on this path. |
| R-3.6 | THE SYSTEM SHALL NOT change what any emitted section computes or contains, only which sections are emitted. |
| R-3.7 | THE SYSTEM SHALL compose `--team` with `--role`, applying both filters together. |
| R-3.8 | THE SYSTEM SHALL NOT add a team filter to any command other than the daily view. |

## Unit 4: Fixture date hygiene

**Why:** One committed fixture holds a pending outcome review due 2026-10-16.
Under Unit 2 that fixture goes overdue on its own about 81 days from now and the
golden snapshot starts failing with no code change — a CI break that arrives by
calendar and reads as a regression. The corpus already has a convention for dates
that must stay valid, and this fixture does not follow it.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL move every committed pending outcome review's `due` date to a far-future date consistent with the convention the fixture corpus already uses for governance dates. |
| R-4.2 | THE SYSTEM SHALL apply R-4.1 to fixtures outside the acceptance path as well, so the corpus is uniform. |
| R-4.3 | THE SYSTEM SHALL NOT leave any committed fixture in a state where the passage of time alone changes a validation verdict. |
| R-4.3a | THE SYSTEM SHALL provide a test that sweeps the dated frontmatter of every committed fixture and fails on any date near enough to change a verdict within the foreseeable life of the corpus, so this invariant is enforced rather than merely stated. |
| R-4.4 | WHERE a fixture date is moved, THE SYSTEM SHALL regenerate every golden snapshot and derived index that renders it, including the feature index, which reads outcome-review frontmatter. |
| R-4.5 | THE SYSTEM SHALL accept that moving these dates leaves the overdue path with no committed-fixture coverage, and SHALL rely on synthesized workspaces for it per R-5.4. |

## Unit 5: Coverage

**Why:** Every fixture date is far-future, so the warning band and the overdue
check fire on nothing in the committed corpus. Both features would ship with zero
coverage from the fixture-driven suite, and R-4.1 deliberately widens that gap.

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL provide test coverage for a deviation dated inside the warning window, asserting warning severity and the reported day count. |
| R-5.2 | THE SYSTEM SHALL provide test coverage for an exception dated inside the warning window, asserting warning severity and the reported day count. |
| R-5.3 | THE SYSTEM SHALL provide test coverage at each window boundary, asserting that the day after the window is passing and the day the date falls into the past is failing. |
| R-5.4 | THE SYSTEM SHALL provide test coverage for an overdue pending outcome review, a future pending one, a non-pending overdue one, and one with no due date. |
| R-5.5 | THE SYSTEM SHALL provide test coverage for the team-scoped daily view, asserting both which sections appear and which are omitted. |
| R-5.6 | THE SYSTEM SHALL construct every date-sensitive test case relative to the clock the gate reads, and SHALL NOT hard-code a calendar date that will change meaning as time passes. |

## Unit 7: Stale in-flight work

**Why:** The lifecycle runs discovery (draft to validated), PRD (proposed to
completed), archive, outcome review. Units 1 and 2 close the governance side and
the tail. The middle is open: a PRD sitting at `status: proposed` since January
and a discovery brief at `status: draft` for fourteen months are invisible to
every gate, though both already carry `created:`. "What is stuck?" is asked more
often than "what expires soon?", and answering it is what makes the retitled gate
honest about the whole lifecycle rather than only its end.

| ID | EARS statement |
| --- | --- |
| R-7.1 | THE SYSTEM SHALL define a fixed staleness threshold for in-flight work, longer than the warning window of Unit 1. |
| R-7.2 | IF an active PRD's `status` is `proposed` and its `created` date is older than the threshold, THEN THE SYSTEM SHALL emit a warning naming the PRD, the platform, the date, and the age. |
| R-7.3 | IF a discovery brief's `status` is `draft` and its `created` date is older than the threshold, THEN THE SYSTEM SHALL emit a warning naming the brief, the team, the date, and the age. |
| R-7.4 | THE SYSTEM SHALL emit these findings at warning severity, and SHALL NOT cause a workspace to exit non-zero on account of stale in-flight work. |
| R-7.5 | WHERE a discovery brief's `status` is `validated`, THE SYSTEM SHALL emit nothing for it, since a validated brief is waiting on a PRD rather than stalled. |
| R-7.6 | IF an in-flight artifact has no `created` value or an unparseable one, THEN THE SYSTEM SHALL emit a warning naming the artifact and the missing or bad value, and SHALL NOT raise an artifact error. |
| R-7.7 | THE SYSTEM SHALL report these findings from the same gate as Units 1 and 2, and SHALL NOT introduce a new gate. |
| R-7.8 | THE SYSTEM SHALL read the same current-date seam Unit 1 defines, so every age in one run is measured against one day. |
| R-7.9 | THE SYSTEM SHALL carry the created date and the age as separate typed fields. |
| R-7.10 | THE SYSTEM SHALL treat an in-flight artifact's `created` date as subject to the fixture-date invariant of Unit 4, so no committed fixture becomes stale by the passage of time. |

## Unit 8: Non-regression

**Why:** These additions insert new severities into two of the most-parsed
outputs in the system. Stating the boundary makes an accidental behavior change a
spec violation rather than a judgment call.

| ID | EARS statement |
| --- | --- |
| R-8.1 | THE SYSTEM SHALL NOT cause any workspace that exits 0 today to exit non-zero after this change. |
| R-8.2 | THE SYSTEM SHALL NOT alter any existing finding code, message text, severity, or emission order. |
| R-8.3 | THE SYSTEM SHALL NOT add, remove, or renumber any validation gate. |
| R-8.4 | THE SYSTEM SHALL NOT add a subcommand, and SHALL change exactly one subcommand's argument surface. |
| R-8.5 | WHEN the daily-view usage string changes, THE SYSTEM SHALL update the exact-match parser assertions and the subcommand count guard that pin it. |
| R-8.6 | THE SYSTEM SHALL keep every addition inside packages that neither print nor exit, composing finding text the way the surrounding code already does. |
| R-8.7 | WHEN `make check` runs, THE SYSTEM SHALL complete gofmt, vet, `go test ./...`, and `examples/acceptance.sh` without failure. |
| R-8.7a | THE SYSTEM SHALL leave the interactive terminal UI unchanged, which it consumes already-rendered screens and reasons about no finding severity. |
| R-8.8 | THE SYSTEM SHALL remain behaviorally independent of the `cos-` skill namespace change, so that either may land first without altering the other's specified behavior. |
| R-8.8a | THE SYSTEM SHALL acknowledge that both changes rewrite the same golden line — one its counter, one its title — and SHALL require whichever lands second to regenerate the goldens by running the binary. |
| R-8.8b | THE SYSTEM SHALL NOT produce any golden snapshot by hand-editing or hand-merging its text. |
