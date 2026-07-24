---
type: prd
id: 2026-instant-refunds-servicing
title: Instant refunds — servicing view
status: proposed
team: cs-servicing
platform: customer-service
components: [case-desk]
governanceSnapshot: 2026-07-21
decisionOwner: noor.rahimi (Product Owner, CS Servicing)
created: 2026-07-21
initiative: initiative://instant-refunds
tags: [component/case-desk, kind/prd, platform/customer-service, status/proposed, team/cs-servicing]
---

# PRD: Instant refunds — servicing view

## Problem statement
Agents cannot see refund status; 62% of refund Cases are status inquiries the
customer could self-serve if status were visible.

## Success metrics
- Case desk shows live refund status on every refund-linked Case.
- Refund status inquiry Cases drop by half within 90 days.

## Proposed change
Case desk consumes the payments status API (consuming relationship — only
consumer-facing clauses apply) and renders refund state on the Case timeline.

## Affected components
- `case-desk`

## Applicable governance (snapshot 2026-07-21)

**case-desk**
- [ ] company: data-residency v1.0 (mandatory) — evidence:
- [ ] company: security-service-baseline v3.0 (mandatory) — evidence:
- [ ] identity: token-verification v1.0 (mandatory, via consuming) — evidence:

## Out of scope
Initiating refunds from the case desk (phase 2).

## Rollout and validation
Ships last of the three sister PRDs; see initiative://instant-refunds.
