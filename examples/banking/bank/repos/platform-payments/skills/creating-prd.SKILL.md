---
id: skill://product/creating-prd
type: skill
version: '1.3'
authority: canonical
appliesTo: ['company://all-platforms']
inputs:
- {a discovery brief with status: validated (or an explicit problem statement)}
outputs: [change-records/active/<id>/prd.md passing `company-os prd validate`]
tags: [authority/canonical, platform/payments]
---

# Creating a PRD (payments platform copy)

1. (mandatory) Scaffold from the validated discovery — never copy an old PRD:
   `company-os prd new --team <team> --platform payments \
      --components <id,...> --from-discovery <discovery-id>`
2. (mandatory) Set `decisionOwner` in the frontmatter.
3. (mandatory) Write **Proposed change** as the future Representation of Reality.
4. (default) State the affected Rail(s) explicitly in the Proposed change.
5. (guidance) Drafting method is yours; the artifact contract is what is enforced.
6. (mandatory) Review the injected **Applicable governance** checklist during
   refinement, not at release time.
7. (mandatory) `company-os prd validate <id> --platform payments` before review.
8. (mandatory) An unmet mandatory requirement is never deleted — use
   skill://governance/requesting-an-exception.
