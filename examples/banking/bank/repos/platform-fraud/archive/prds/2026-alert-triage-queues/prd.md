---
type: prd
id: 2026-alert-triage-queues
title: Priority queues for alert triage
status: completed
team: fraud-detection
platform: fraud
components: [transaction-screening]
governanceSnapshot: 2026-07-10
decisionOwner: ines.beck (Product Owner)
created: 2026-07-10
fromDiscovery: 2026-alert-triage-queues
tags: [component/transaction-screening, discovery/2026-alert-triage-queues, kind/prd,
  platform/fraud, status/completed, team/fraud-detection]
---

# PRD: Priority queues for alert triage

## Problem statement
High-risk Alerts dispositioned in 6.4h median against a 2h SLA; single FIFO
queue is 71% low-risk noise.

## Success metrics
- High-risk median time-to-disposition < 2h within 30 days.
- False-negative rate unchanged (±0.1pp).

## Proposed change
Alerts are ranked into priority queues at creation; analysts pull from the
highest non-empty queue. Reality doc gains a "triage queues" section.

## Affected components
- `transaction-screening`

## Applicable governance (snapshot 2026-07-10)

**transaction-screening**
- [x] company: data-residency v1.0 (mandatory) — evidence: EU-only deployment attestation 2026
- [x] company: security-service-baseline v3.0 (mandatory) — evidence: annual review 2026-03
- [x] company: estimation-story-points v1.1 (default) *(team deviation applies — t-shirt sizes)*
- [x] fraud: alert-latency v1.2 (mandatory) — evidence: queue-latency dashboard + SLA test suite
- [x] payments: settlement-finality v1.0 (mandatory, via consuming) *(team exception applies — legacy scoring engine, expires 2026-12-31)*

## Out of scope
Auto-disposition of low-risk Alerts.

## Rollout and validation
Shadow-ranked for one week, then cut over per analyst pod.
