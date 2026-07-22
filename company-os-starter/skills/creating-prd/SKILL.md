---
id: skill://product/creating-prd
version: '1.3'
authority: canonical
appliesTo: ['company://all-platforms']
inputs:
- {a discovery brief with status: validated (or an explicit problem statement)}
outputs: [change-records/active/<id>/prd.md passing `company-os prd validate`]
tags: [authority/canonical, kind/skill, process/prd]
---

# Creating a PRD

1. (mandatory) Scaffold from the validated discovery — never copy an old PRD:
   `company-os prd new --team <team> --platform <platform> \
      --components <id,...> --from-discovery <discovery-id>`
   This injects the governance snapshot for the affected components,
   including any approved team deviations.
2. (mandatory) Set `decisionOwner` in the frontmatter. `prd validate`
   fails on TODO.
3. (mandatory) Write **Proposed change** as the future Representation of
   Reality: what will be true when this ships.
4. (default) Keep the PRD lean if your team has an approved deviation for
   `prd-structure`; otherwise follow the platform structure.
5. (guidance) Drafting method is yours — personal prompts, your own agent
   rules in `scratchpad/personal-rules/`, whatever works. The artifact
   contract is what is enforced, not your process.
6. (mandatory) Review the injected **Applicable governance** checklist with
   your tech lead during refinement, not at release time.
7. (mandatory) Validate before requesting review:
   `company-os prd validate <id> --platform <platform>`
8. (mandatory) If a mandatory requirement cannot be met, do NOT delete the
   checklist line. Use skill://governance/requesting-an-exception.
