---
id: skill://product/running-discovery
version: '1.0'
authority: canonical
appliesTo: ['company://all-teams']
inputs: ['a problem signal (tickets, metrics, stakeholder ask)']
outputs: [product/discovery/<id>/brief.md passing `company-os discover validate`]
tags: [authority/canonical, kind/skill, process/discovery]
---

# Running Discovery

1. (mandatory) Scaffold — never start from a blank file or an old brief:
   `company-os discover new "<title>" --team <team-id>`
2. (mandatory) Fill **Problem signal** with evidence, not opinion. Quantify it.
3. (mandatory) Write the **Hypothesis** as: we believe <change> for <who>
   will achieve <outcome>.
4. (mandatory) Define **Success criteria** as numbers with a time window.
   If you cannot measure it, the discovery is not done.
5. (guidance) Method is yours: interviews, data analysis, prototypes,
   RICE/WSJF prioritization — the team chooses. Keep raw notes in
   `scratchpad/`, never in the brief.
6. (default) List an initial guess of affected components — it drives which
   platform governance the future PRD will inherit.
7. (mandatory) Validate: `company-os discover validate <id> --team <team-id>`.
   A brief that fails validation cannot become a PRD.
8. (mandatory) Record the **Decision** — `validated` or `invalidated`.
   Invalidated discoveries are kept: a killed idea with a reason is knowledge.
