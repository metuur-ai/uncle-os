---
type: context-map
id: map://crm--communications
tags: [ontology/context-map]
upstream: context://crm
downstream: context://communications
---

# Context Map: CRM -> Communications

| CRM term | Communications term | Translation rule |
|---|---|---|
| Customer | Recipient | A Customer becomes a Recipient when subscribed to >=1 Channel |
| Contact preference | Channel subscription | Preferences project onto per-channel subscriptions |

Integration style: published-language (CRM publishes CustomerSubscribed events;
Communications translates at its boundary, an Anti-Corruption Layer).
