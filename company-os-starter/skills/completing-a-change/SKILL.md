---
id: skill://product/completing-a-change
version: '1.1'
authority: canonical
appliesTo: ['company://all-platforms']
inputs: [an active PRD whose implementation has shipped]
outputs: [archived PRD + updated reality docs + scheduled outcome review]
tags: [authority/canonical, kind/skill, process/change-completion]
---

# Completing a Change

A change is NOT done when the code merges. It is done when reality is updated.

1. (mandatory) Update the Representation of Reality doc for every affected
   component (`reality/components/<id>.md`): present tense, current behavior,
   bump the `updated` date.
2. (mandatory) Check off every line in the PRD's governance checklist with a
   link to evidence. Unchecked mandatory items block completion.
3. (default) Update the component catalog entry if capabilities changed.
4. (mandatory) Run: `company-os prd complete <id> --platform <platform>`.
   The command refuses to archive if reality docs are older than the PRD
   or checklist items are open. Do not use --force outside incidents.
5. (mandatory) The command archives the PRD, appends `log.md`, and schedules
   the outcome review (+90 days). The outcome review compares actuals to the
   PRD's success metrics — the PO owns it.
