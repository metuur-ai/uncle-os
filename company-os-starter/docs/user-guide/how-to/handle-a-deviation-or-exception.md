---
title: Handle a deviation or exception
---

# Handle a deviation or exception

Company OS rules come in three tiers: `mandatory`, `default`, and
`guidance`. You can't opt out of `mandatory` rules and you don't need to
track `guidance` ones — but `default` rules are comply-or-explain, and
`mandatory` rules can be excepted through an owned, expiring process. Both
are explicit, both expire, and both get checked in `company-os validate`.

## Deviating from a default-tier rule

Say Moonbeam's `web` team forecasts with cycle time instead of story points,
against the company's default estimation standard:

```bash
$ company-os deviation declare "company-standard://estimation/story-points" \
    --team web \
    --rationale "Team forecasts with cycle time instead of points."
declared deviation from company-standard://estimation/story-points in teams/web/governance/deviations.yaml
review due 2027-01-19; re-run: company-os governance resolve --team web
```

That review date isn't cosmetic — `declare` sets it 180 days out
automatically. `company-os validate` gate `[2/7]` fails once a deviation's
`reviewDate` is in the past, so a deviation is a standing decision you'll be
asked to re-affirm, not a permanent escape hatch.

`deviation declare` only accepts `default`-tier rules —
`resolve_team_governance` rejects any deviation aimed at a `mandatory` rule
outright. That's the whole point of the tier: mandatory rules are outcomes
the company needs guaranteed everywhere, not implementation details a team
can choose out of.

## Excepting a mandatory rule

Mandatory rules can only be escaped through an **exception**: expiring,
tied to one component, and requiring the rule owner's sign-off.

```bash
$ company-os exception request "platform-standard://ordering/order-confirmation-sla" \
    --team web --component legacy-pos-bridge \
    --expires 2026-12-31 \
    --reason "Legacy POS integration can't emit confirmations synchronously yet."
exception drafted in teams/web/governance/exceptions.yaml (expires 2026-12-31)
note: mandatory rules require approval by the rule owner before this is valid.
```

`--expires` is required — there's no such thing as a permanent exception.
`company-os validate` gate `[2/7]` fails on any exception missing an
`expires` date, or one that's already past it.

## Which one do I need?

| Question | Answer |
|---|---|
| Is the rule `mandatory` or `default`? | Check with `company-os governance explain <component>` — it prints each rule's tier. |
| The rule is `default` and we just work differently | `deviation declare` |
| The rule is `mandatory` and we genuinely can't comply yet | `exception request` — and get the rule owner's sign-off |
| The rule is `mandatory` and we think it shouldn't apply at all | Neither — that's a governance conversation with whoever owns the rule, not a CLI command |

## After declaring either one

Re-resolve governance so the deviation/exception is reflected in the merged
view teams and PRDs actually see:

```bash
$ company-os governance resolve --team web
```

## When things go sideways

| Symptom | Cause | Fix |
|---|---|---|
| `validate` gate `[2/7]` fails: "deviation ... expired" | `reviewDate` passed | re-run `deviation declare` for the same rule, or drop it if it's no longer needed |
| `validate` gate `[2/7]` fails: "exception ... has NO expiry" | hand-edited `exceptions.yaml` without an `expires` field | always create exceptions via `exception request`, never by hand |
| Deviation rejected / rule still shows as required | you tried to deviate a `mandatory` rule | mandatory rules can only be excepted, never deviated — use `exception request` |
| Exception "drafted" but still enforced | mandatory-rule exceptions need the rule owner's approval before they're valid, per the CLI's own note | follow up with the rule owner; the CLI records the request, it doesn't grant it |
