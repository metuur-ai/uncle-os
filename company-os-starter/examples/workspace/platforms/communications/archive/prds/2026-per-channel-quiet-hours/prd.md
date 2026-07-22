---
type: prd
id: 2026-per-channel-quiet-hours
title: Per-channel quiet hours
status: completed
team: customer-engagement
platform: communications
components: [customer-notification-service]
governanceSnapshot: 2026-07-18
decisionOwner: maria.lopez (Product Owner)
created: 2026-07-18
fromDiscovery: 2026-per-channel-quiet-hours
tags: [component/customer-notification-service, discovery/2026-per-channel-quiet-hours,
  kind/prd, platform/communications, status/completed, team/customer-engagement]
---

# PRD: Per-channel quiet hours

## Problem statement
412 support tickets in Q2 about notifications arriving at night.
Opt-out rate for push is 3x email; exit surveys cite "too many pings after hours".

## Success metrics
- Push opt-out rate drops from 4.2% to below 3.0% within 60 days of launch.
- Delivery success rate stays above 99.5%.
- Support tickets about night-time notifications drop by half.

## Proposed change
Recipients can define quiet hours per channel (push, email). Messages queued
during quiet hours are delivered at the window end; urgent-class messages bypass
quiet hours. Reality docs will gain a Quiet Hours business rule.

## Affected components
- `customer-notification-service`

## Applicable governance (snapshot 2026-07-18)
<!-- Injected by `prd new`. Check items off as evidence is linked. -->

**customer-notification-service**
- [x] company: security-service-baseline v3.0 (mandatory) — evidence: repo://customer-notification-service/docs/security-review-2026.md
- [x] company: customer-data-privacy v2.2 (mandatory) — evidence: privacy DPIA #418
- [x] company: tier-1-observability v1.4 (default) — evidence: grafana dashboard notif-quiet-hours
- [x] communications: delivery-reliability v2.1 (mandatory) — evidence: integration test suite qh-delivery
- [x] communications: message-schema v1.3 (mandatory) — evidence: envelope schema unchanged (n/a note)
- [x] communications: prd-structure v1.0 (default) *(team deviation applies)* — evidence: lean PRD per approved deviation

## Out of scope

## Rollout and validation
