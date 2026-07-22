---
type: concept
id: concept://component
tags: [ontology/concept]
aliases: [Component]
relationships:
  - owned-by: concept://team
  - implements: concept://capability
  - relates-to: concept://platform
taxonomy:
  subtypes: [api, service, worker, library, database, ui]
---

# Component

A deployable or maintainable unit of software with a single accountable team.
A component may relate to several platforms (`belongs-to`, `supports`,
`publishes-to`) but has exactly one canonical descriptor, which is the single
source of truth for its relationships and ownership.
