---
type: context-map
id: map://payments--fraud
tags: [ontology/context-map]
upstream: context://payments
downstream: context://fraud
---

# Context Map: Payments -> Fraud

| Payments term | Fraud term | Translation rule |
|---|---|---|
| Payment Order | Transaction | A Payment Order becomes a Transaction when emitted to the screening stream |
| Settlement | (not modeled) | Fraud screens pre-settlement only |

Integration style: published-language (Payments publishes PaymentOrderAccepted
events; Fraud translates at its boundary, an Anti-Corruption Layer).
