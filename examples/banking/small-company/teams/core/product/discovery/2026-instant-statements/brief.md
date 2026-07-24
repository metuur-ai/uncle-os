---
type: discovery-brief
id: 2026-instant-statements
title: Instant statement downloads
status: validated
team: core
created: 2026-07-14
tags: [kind/discovery, status/validated, team/core]
---

# Discovery: Instant statement downloads

## Problem signal
31% of support email in June was "where is my statement PDF?" (87 of 280
messages). Statements are generated nightly; users opening the app during the
day see nothing for the current period.

## Hypothesis
We believe that generating statements on demand will cut statement-related
support email by 70% and remove the nightly batch as a single point of failure.

## Success criteria
- Statement-related support email drops below 30/month within 60 days.
- On-demand generation p95 under 4 seconds.
- Nightly batch retired.

## Affected components (initial guess)
- `banking-app`

## Risks and open questions
- PDF rendering load during month-end peaks.

## Decision
validated — 2026-07-16, signal is quantified and criteria are numeric.
