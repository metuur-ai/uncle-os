---
id: reality-payment-gateway
type: component-reality
authority: canonical
updated: 2026-07-21
boundedContext: context://payments
tags: [authority/canonical, component/payment-gateway, kind/reality, platform/payments]
---

# Payment Gateway — Current Behavior

Accepts Payment Orders over API, routes to SEPA and instant rails, emits
PaymentOrderAccepted events consumed by Fraud screening. Settlement status
is queryable per order.
