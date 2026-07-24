---
type: bounded-context
id: context://communications
tags: [ontology/context]
aliases: [Communications Context]
ownerPlatform: platform://communications
ubiquitousLanguage: {Message: A unit of content addressed to a Recipient over a Channel.,
  Recipient: The addressable target of a Message. NOT called "Customer" here., Template: A
    parameterized definition from which Messages are rendered., Delivery: 'The observable
    outcome of sending a Message (delivered, failed, queued).', Channel: 'A transport
    for Messages (email, push).'}
forbiddenTerms: {Customer: 'Use Recipient. "Customer" belongs to context://crm and
    context://billing.'}
---

# Bounded Context: Communications

One model, one vocabulary. Documents whose frontmatter declares
`boundedContext: context://communications` must use this ubiquitous language.
`company-os validate --ontology` flags forbidden terms in canonical docs.
