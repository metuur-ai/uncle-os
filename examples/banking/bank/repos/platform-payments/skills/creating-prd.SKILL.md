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

Steps 1–8 are the direct path, for a change whose platform and Rails are
already known. When the change still needs working up, start at step 9: a
draft lives in the team tree and nothing lands on payments until promotion.

9. (default) Draft in the team tree first:
   `company-os prd new --team <team> --draft "<title>" [--platform payments]`
   It lands in `teams/<team>/product/change-records/draft/<id>/prd.md` with
   `status: draft`.
10. (mandatory) Validate the draft while you work:
    `company-os prd validate <draft-id> --team <team>`
    Only the core frontmatter contract applies. `platform`, `components`,
    `governanceSnapshot` and `decisionOwner` are required by promotion, not by
    draft validation — the affected Rail(s) can stay `TODO` until then.
11. (mandatory) Promote when the change is real:
    `company-os prd promote <draft-id> --team <team>`
    The target comes from the draft's `promoteTo.platform`, not a flag. This
    is where the full PRD contract and the payments governance checklist bind,
    and it is a precondition of `prd complete` — a draft cannot be completed.
    Then pick up at step 6 against the promoted record.
12. (default) Retire a draft you will not promote:
    `company-os prd abandon <draft-id> --team <team>`. Never `rm` it.
