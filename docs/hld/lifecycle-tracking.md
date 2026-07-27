# Lifecycle Tracking — High-Level Design

## Overview

Company OS can tell you what is broken. It cannot tell you what is about to
break, and in one case it cannot tell you that something broke at all.

Three commitments in the methodology carry a date. A deviation has a
`reviewDate`. An exception has an `expires`. A completed PRD schedules an
outcome review `due` 90 days out. Every one of those dates encodes the same
promise: *we accepted a temporary state, and we will revisit it by this day.*

The tooling handles two of them as tripwires and the third not at all. The
expiry gate emits `SevFail` and nothing else — a deviation whose `reviewDate` is
tomorrow is completely silent, and the day after it is a hard CI failure. There
is no state between "fine" and "already broken", so a team's first signal that a
governance commitment came due is a red build. Outcome reviews are worse:
`prd complete` writes one with `status: pending` and a `due` date, `today` lists
it, and nothing anywhere compares that date to the calendar. A workspace can
carry outcome reviews eighteen months overdue and pass every gate.

The middle of the lifecycle is open too. A PRD at `status: proposed` since
January and a discovery brief at `status: draft` for fourteen months are
invisible to every gate, though both carry `created:`. Nothing asks "what is
stuck?"

This change adds the missing time horizon. Dated governance commitments learn a
warning window, an overdue outcome review stops being invisible, stalled
in-flight work gets an age, and `today` gains the filter that makes it usable on
a workspace with more than one team.

None of it is judgment. Each addition is a date compared to today and a count —
exactly the kind of deterministic verdict the CLI exists to render, and exactly
the kind of thing a skill could never enforce.

## Stakeholders & Impact

**Teams carrying deviations and exceptions.** Today the only notification that a
commitment expired is a failing gate, after the fact. After this change a
commitment inside the warning window is visible while there is still time to
renew it, extend it, or let it lapse deliberately.

**Platform owners who completed a change.** The methodology says a change is not
finished when it ships, it is finished when its outcome is reviewed. That loop
currently has no closing enforcement. After this change an overdue outcome
review is visible in the validation output rather than only as a line in `today`
that nobody compares against a calendar.

**Anyone running `today` on a real workspace.** Its only flag is `--role`, so on
a workspace with six platforms it prints all six. Adding `--team` makes it answer
"what is in flight for *my* team" instead of "what is in flight anywhere".

**CI owners.** Two gates gain a severity they did not emit before. Warnings do
not fail a build, so no pipeline that passes today starts failing — but output
that was previously either silent or fatal now has a middle band, and anything
parsing that output should expect it.

## Goals

1. A deviation or exception coming due inside a fixed 30-day window is reported
   as a warning, naming the date and the days remaining, without failing the
   build.
2. An outcome review whose `due` date has passed while `status: pending` is
   reported, so the change-completion loop has a closing signal, and the daily
   view shows it as overdue rather than printing it like any other row.
3. A PRD or discovery brief that has sat in its opening status past a threshold
   is reported with its age.
4. `company-os today` accepts `--team` and restricts its output to that team and
   the platforms that team's components belong to.
5. No workspace that passes validation today begins failing, or aborting, because
   of this change.
6. The gate reporting these findings is titled for what it now checks, while its
   machine-readable slug stays as it is.
7. `make check` passes, and the five golden snapshots agree with the binary.

## Non-Goals

- **Overdue outcome reviews do not fail the build.** An unreviewed outcome is a
  management problem, not a broken artifact. A gate that failed CI for it would
  be disabled within a month, and a disabled check reports nothing at all.
- **No new subcommand.** Everything lands on `validate` and `today`, both of
  which already exist and are already the places people look.
- **No notification, scheduling, or calendar integration.** The CLI reports
  state when run. Deciding when to run it is outside the tool.
- **No history command.** "What happened last quarter" is a narrative question,
  and `git log` over `archive/prds/` already answers it better than a bespoke
  command would.
- **The warning window is not a grace period.** A past-due deviation stays a
  hard failure. The window only adds lead time before that point.
- **`--team` does not change what `today` computes**, only which sections it
  prints.

## Success Criteria

Observable when this ships:

- Setting a fixture deviation's `reviewDate` to a date inside the window makes
  `validate` report a warning naming the days remaining, and exit 0.
- Moving that same date into the past restores the existing hard failure, with
  its message unchanged.
- Setting a fixture outcome review's `due` into the past makes `validate` report
  it, naming the PRD and the date, and exit 0.
- Marking that review any status other than `pending` silences the report.
- A synthesized PRD left at `proposed` past the threshold is reported with its
  age; the same PRD marked completed is not.
- `company-os today --team <t>` prints that team's section and the platforms
  carrying its components, and omits every other team.
- A workspace holding an outcome review with a typo'd or missing `due` date
  still exits 0, reporting the bad value rather than aborting on it.
- `company-os validate` still exits 0 on `examples/workspace` and
  `examples/federated`; the goldens change only in the retitled gate line and
  where a fixture date was deliberately moved.
