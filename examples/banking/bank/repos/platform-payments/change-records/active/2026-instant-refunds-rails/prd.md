---
type: prd
id: 2026-instant-refunds-rails
title: Instant refunds — payments rails
status: proposed
team: payments-rails
platform: payments
components: [payment-gateway, settlement-engine]
governanceSnapshot: 2026-07-21
decisionOwner: marta.silva (Group PM, Payments)
created: 2026-07-21
initiative: initiative://instant-refunds
tags: [component/payment-gateway, component/settlement-engine, kind/prd, platform/payments,
  status/proposed, team/payments-rails]
---

# PRD: Instant refunds — payments rails

## Problem statement
Refunds settle in 2-3 business days; 62% of refund-related CS Cases are pure
status inquiries. The instant-refunds initiative commits to seconds, not days.

## Success metrics
- Refund Payment Orders settle over instant rails for >=80% of eligible volume.
- Refund status inquiries to Customer Service drop by half within 90 days.

## Proposed change
A refund is a first-class Payment Order type on the instant rail. The gateway
accepts RefundOrder, the settlement engine settles it exactly once, status is
queryable end to end. ("Refund" is registered once in the ontology and mapped
across payments, cards, and customer-service contexts.)

## Affected components
- `payment-gateway`
- `settlement-engine`

## Applicable governance (snapshot 2026-07-21)

**payment-gateway, settlement-engine**
- [ ] company: data-residency v1.0 (mandatory) — evidence:
- [ ] company: security-service-baseline v3.0 (mandatory) — evidence:
- [ ] company: estimation-story-points v1.1 (default) — evidence:
- [ ] payments: settlement-finality v1.0 (mandatory) — evidence: duplicate-refund test planned
- [ ] payments: prd-structure v1.0 (default) — evidence:

## Out of scope
Chargebacks (cards PRD 2026-instant-refunds-disputes owns dispute flows).

## Rollout and validation
Sister PRDs: cards `2026-instant-refunds-disputes`, customer-service
`2026-instant-refunds-servicing`. NOTE: this cross-PRD link is by convention —
the CLI has no initiative artifact (gap G2); see ../../../../../initiatives/.
