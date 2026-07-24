---
id: reality-transaction-screening
type: component-reality
authority: canonical
updated: 2026-07-21
boundedContext: context://fraud
tags: [authority/canonical, component/transaction-screening, kind/reality, platform/fraud]
---

# Transaction Screening — Current Behavior

Consumes PaymentOrderAccepted events (translated to Transactions at the
boundary per `map://payments--fraud`), scores them, raises Alerts into Cases.

## Current limitations
- Legacy scoring engine cannot yet evidence req://payments/settlement-finality
  consumer clauses — covered by an expiring exception (see team-fraud-detection).
