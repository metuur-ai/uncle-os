---
type: initiative            # PROPOSED artifact kind — gap G2 in the research doc.
id: initiative://instant-refunds
status: active
sponsor: director-of-product
tags: [kind/initiative]
memberPrds:
  - {platform: platform://payments, prd: change-records/active/2026-instant-refunds-rails/prd.md}
  - {platform: platform://cards, prd: change-records/active/2026-instant-refunds-disputes/prd.md}
  - {platform: platform://customer-service, prd: change-records/active/2026-instant-refunds-servicing/prd.md}
---

# Initiative: Instant Refunds

Today the CLI has NO cross-platform artifact: this work decomposes into one PRD
per platform, each validated/completed independently against its own platform
(`prd new --platform <p>`). This file models the proposed lightweight initiative
doc: it references member PRDs by platform + path (IDs canonical, links
derived), so a future `validate` step could flag dangling members. Shared
vocabulary comes from the ontology, not from this file: "Refund" would be
registered once and mapped across context://payments, context://cards, and
customer-service.
