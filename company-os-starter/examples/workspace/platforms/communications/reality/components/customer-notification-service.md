---
id: reality-customer-notification-service
type: component-reality
authority: canonical
updated: 2026-07-18
pointers:
  - {label: Delivery runbook, system: confluence, url: 'https://example.invalid/cns-runbook'}
tags: [authority/canonical, component/customer-notification-service, kind/reality,
  platform/communications]
---

# Customer Notification Service — Current Behavior

Delivers templated outbound customer notifications over email and push.
Supports delivery tracking, retries with idempotency keys, and enforcement
of customer channel preferences.

## Quiet hours
Recipients can define quiet hours per channel. Non-urgent messages queued
during quiet hours deliver at the window end; urgent-class messages bypass.

## Current limitations
- Quiet hours not yet configurable via admin API (UI only).
