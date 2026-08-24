---
title: Take a change from discovery to done
---

# Take a change from discovery to done

The full first-time walkthrough of this loop, with narration and expected
output, lives in
[tutorials/01-first-day-with-company-os.md](../tutorials/01-first-day-with-company-os.md).
This page is the same loop as a repeatable recipe once you already know the
shape.

## 1. Capture the problem (team-private)

```bash
company-os discover new "<title>" --team <team>
# ...fill Problem signal, Hypothesis, Success criteria in brief.md...
company-os discover validate <brief-id> --team <team>
```

Discovery is team-private exploration — it stays under `teams/<t>/product/discovery/`
until it's validated.

## 2. Turn it into a PRD (platform-visible)

```bash
company-os prd new --team <team> --platform <platform> \
  --components <comp-id[,comp-id...]> \
  --from-discovery <brief-id>
```

A PRD is a change record the platform can see — it lands in
`platforms/<platform>/change-records/active/<id>/`. `--from-discovery`
copies the Problem statement and Success metrics forward and stamps a
`governanceSnapshot` so the PRD is judged against the rules that existed
when work started, not whatever they've drifted to since.

```bash
company-os prd validate <prd-id> --platform <platform>
```

Fixes frontmatter and required-section gaps until it passes.

### 2b. The draft branch — when the change isn't platform-ready yet

Step 2 assumes you already know the platform and the components. When you
don't, draft it in the team tree first and promote it later:

```bash
company-os prd new --team <team> --draft "<title>" \
  [--platform <platform>] [--from-discovery <brief-id>]
# ...work the draft up in teams/<team>/product/change-records/draft/<id>/prd.md...
company-os prd validate <draft-id> --team <team>
company-os prd promote <draft-id> --team <team>
```

A draft carries `status: draft` and stays team-private. `validate` tolerates
it: only the core frontmatter contract applies, so `title`, `platform`,
`components`, `governanceSnapshot` and `decisionOwner` can stay `TODO` while
you think. `prd promote` is where the full PRD contract binds — it copies the
draft forward into exactly the change record step 2 would have produced,
under `platforms/<platform>/change-records/active/<id>/`, and refuses while
any of those fields or the required sections are still missing. The target
platform comes from the draft's `promoteTo.platform`, not from a flag.

From there, rejoin the recipe at step 3 — a promoted record is a normal
change record. A draft cannot skip promotion: `prd complete` addresses
records by platform only. To retire a draft you decide not to pursue, run
`company-os prd abandon <draft-id> --team <team>` rather than deleting it.

## 3. Check readiness before pulling into a sprint

```bash
company-os check ready --team <team> --components <comp-id[,comp-id...]>
```

Composed on demand from the team's `standards/definition-of-ready.md` plus
the resolved governance checklist — never a static, stale document.

## 4. Update reality, then complete

```bash
# edit platforms/<platform>/reality/components/<comp-id>.md, bump `updated:`
# check off the PRD's governance checklist with evidence links
company-os prd complete <prd-id> --platform <platform>
```

`prd complete` refuses to archive while any governance checklist item is
unchecked, or while the component's reality doc's `updated:` date is older
than the PRD's `created:` date. This is not a formality — the rule is
literally that a change isn't done until the Representation of Reality
reflects it.

On success it archives the PRD to `archive/prds/`, schedules an
`outcome.md` (due in 90 days), and appends `log.md`.

## 5. Confirm the workspace

```bash
company-os validate
```

## When things go sideways

| Symptom | Cause | Fix |
|---|---|---|
| `discover validate` fails on all three sections | brief created but never filled in | edit `brief.md`'s Problem signal / Hypothesis / Success criteria |
| `prd new` errors about missing `--platform` | `--platform` is required even when `--from-discovery` is set | pass the platform the components belong to |
| `prd complete` fails with "reality doc ... not updated" | you completed before editing the reality doc | edit it and bump `updated:` to today or later, then retry |
| `check ready`/`check done` output looks empty for governance | components not yet passed through `governance resolve` | run `company-os governance resolve --team <team>` first |
