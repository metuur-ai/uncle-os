---
type: prd
id: 2026-instant-refunds-disputes
title: Instant refunds — card disputes
status: proposed
team: cards-issuing
platform: cards
components: [card-issuing]
governanceSnapshot: 2026-07-21
decisionOwner: lena.kovacs (Product Owner, Cards Issuing)
created: 2026-07-21
initiative: initiative://instant-refunds
tags: [component/card-issuing, kind/prd, platform/cards, status/proposed, team/cards-issuing]
---

# PRD: Instant refunds — card disputes

## Problem statement
A won dispute today credits the card in 2-3 days via the legacy batch. The
initiative commits an instant refund the moment a dispute is resolved.

## Success metrics
- Resolved disputes trigger an instant RefundOrder within 60 seconds for >=90%.
- Dispute-resolution NPS +10 within a quarter.

## Proposed change
On dispute resolution, card-issuing emits a RefundOrder to the payments
gateway (tokenized card reference only, per pan-handling).

## Affected components
- `card-issuing`

## Applicable governance (snapshot 2026-07-21)

**card-issuing**
- [ ] company: data-residency v1.0 (mandatory) — evidence:
- [ ] company: security-service-baseline v3.0 (mandatory) — evidence:
- [ ] cards: pan-handling v2.0 (mandatory) — evidence: token-only interface, no PAN leaves the boundary

## Out of scope
Refund settlement itself (payments PRD 2026-instant-refunds-rails).

## Rollout and validation
Coordinated behind the initiative flag; see initiative://instant-refunds.
