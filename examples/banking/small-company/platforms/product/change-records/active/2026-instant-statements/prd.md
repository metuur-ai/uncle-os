---
type: prd
id: 2026-instant-statements
title: Instant statement downloads
status: proposed
team: core
platform: product
components: [banking-app]
governanceSnapshot: 2026-07-17
decisionOwner: ana.ruiz (Team Lead, acting Product Owner)
created: 2026-07-17
fromDiscovery: 2026-instant-statements
tags: [component/banking-app, discovery/2026-instant-statements, kind/prd, platform/product,
  status/proposed, team/core]
---

# PRD: Instant statement downloads

## Problem statement
31% of June support email was "where is my statement PDF?"; statements only
exist after the nightly batch runs.

## Success metrics
- Statement-related support email drops below 30/month within 60 days.
- On-demand generation p95 under 4 seconds.
- Nightly batch retired.

## Proposed change
Statements render on demand for any period, cached after first request. The
nightly batch is removed. Reality doc gains an "on-demand statements" section.

## Affected components
- `banking-app`

## Applicable governance (snapshot 2026-07-17)

**banking-app**
- [ ] company: security-service-baseline v1.0 (mandatory) — evidence:
- [ ] company: change-log v1.0 (default) *(team deviation applies — PR description is the change log)*
- [ ] product: release-safety v1.0 (mandatory) — evidence:

## Out of scope
Historical statement backfill before 2024.

## Rollout and validation
Feature-flagged to the team first, then 10% of users, then all. `prd complete`
will refuse until the reality doc's `updated:` postdates this PRD.
