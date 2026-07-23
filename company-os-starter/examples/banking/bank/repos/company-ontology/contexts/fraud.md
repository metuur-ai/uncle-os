---
type: bounded-context
id: context://fraud
tags: [ontology/context]
aliases: [Fraud Context]
ownerPlatform: platform://fraud
ubiquitousLanguage: {Transaction: An observed account event evaluated by screening.,
  Alert: A screening hit awaiting triage., Case: An investigation grouping one or
    more Alerts., Disposition: 'The outcome of a Case (confirmed-fraud, false-positive).'}
forbiddenTerms: {Payment Order: Use Transaction. The payments-side instruction is
    out of scope once it enters screening.}
---

# Bounded Context: Fraud

Owns "Transaction". Cross-context translation with Payments is defined in
`context-maps/fraud-to-payments.md` — never redefine terms locally.
