---
type: bounded-context
id: context://payments
tags: [ontology/context]
aliases: [Payments Context]
ownerPlatform: platform://payments
ubiquitousLanguage: {Payment Order: 'An instruction to move funds from a debtor to
    a creditor.', Settlement: The irrevocable transfer of funds between institutions.,
  Clearing: Exchange and reconciliation of payment messages prior to Settlement.,
  Rail: 'A scheme/network a Payment Order travels on (SEPA, SWIFT, instant).'}
forbiddenTerms: {Transaction: 'Use Payment Order or Settlement. "Transaction" belongs
    to context://fraud, where it means an observed account event under screening.'}
---

# Bounded Context: Payments

Same word, different meaning: what Fraud calls a "Transaction" is an observed
event; Payments reasons about Payment Orders and Settlements. Canonical docs
declaring `boundedContext: context://payments` are vocabulary-linted against
this list (`validate --ontology`, roadmap).
