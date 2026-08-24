---
id: skill://product/creating-prd
type: skill
version: '1.3'
authority: canonical
appliesTo: ['company://all-platforms']
inputs:
- {a discovery brief with status: validated (or an explicit problem statement)}
outputs: [change-records/active/<id>/prd.md passing `company-os prd validate`]
tags: [authority/canonical, platform/communications]
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

Steps 1–8 are the direct path, for a change whose platform and components are
already known. When the change still needs working up, start at step 9
instead: a draft lives in the team tree, `validate` tolerates it, and nothing
appears under `platforms/` until it is promoted.

9. (default) Draft it in the team tree first:
   `company-os prd new --team <team> --draft "<title>" [--platform <platform>]`
   The draft lands in `teams/<team>/product/change-records/draft/<id>/prd.md`
   with `status: draft`.
10. (mandatory) Validate the draft as you work:
    `company-os prd validate <draft-id> --team <team>`
    Only the core frontmatter contract applies here. `platform`,
    `components`, `governanceSnapshot` and `decisionOwner` are required by
    promotion, not by draft validation — leave them `TODO` until you promote.
11. (mandatory) Promote when the change is real:
    `company-os prd promote <draft-id> --team <team>`
    The target comes from the draft's `promoteTo.platform`, not from a flag.
    This is where the full PRD contract binds, and it is a precondition of
    `prd complete` — a draft cannot be completed. Then pick up at step 7
    against the promoted record.
12. (default) Retire a draft you will not promote:
    `company-os prd abandon <draft-id> --team <team>`. Never `rm` it.
