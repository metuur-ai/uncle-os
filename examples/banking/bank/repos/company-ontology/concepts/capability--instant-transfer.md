---
type: concept
id: capability://payments/instant-transfer
tags: [context/payments, ontology/capability, ontology/concept]
aliases: [Instant Transfer, instant-transfer]
boundedContext: context://payments
implementedBy: ['component://payment-gateway']
---

# Instant Transfer (capability)

The ability to execute a Payment Order over an instant rail with Settlement
confirmed to the debtor within seconds, and status observable end to end.
Vocabulary is bound to `context://payments` — this doc says "Payment Order",
never "Transaction" (see `contexts/payments.md` forbiddenTerms).
