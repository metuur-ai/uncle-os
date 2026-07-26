---
id: skill://product/completing-a-change
version: '1.2'
authority: canonical
appliesTo: ['company://all-platforms']
inputs: [an active PRD whose implementation has shipped]
outputs: [archived PRD + updated reality docs + scheduled outcome review]
tags: [authority/canonical, kind/skill, process/change-completion]
---

# Completing a Change

A change is NOT done when the code merges. It is done when reality is updated.

**Agents: run every command with `--json` and branch on the exit code.** The
envelope carries `exitCode`, a per-finding `severity`/`code`, and a `guidance`
array holding the next command. Codes are a contract; the English in `message`
is not. Never parse prose — in particular, never decide the done-gate passed
because you did not see a refusal sentence. Full envelope:
[reference/company-os-cli.md § `--json`](../../docs/user-guide/reference/company-os-cli.md#--json).

1. (mandatory) Update the Representation of Reality doc for every affected
   component (`reality/components/<id>.md`): present tense, current behavior,
   bump the `updated` date. Missing doc? Scaffold it —
   `company-os --json reality new <component-id> --platform <platform>`.
2. (mandatory) Check off every line in the PRD's governance checklist with a
   link to evidence. Unchecked mandatory items block completion. To see what
   is still open before you attempt completion:

   ```bash
   company-os --json check done --team <team> --components <id,...> \
     | jq -r '.sections[].findings[] | select(.code=="check.checklist") | .message'
   ```

3. (default) Update the component catalog entry if capabilities changed.
4. (mandatory) Complete it:

   ```bash
   out=$(company-os --json prd complete <id> --platform <platform>); rc=$?
   ```

   - `0` — archived. See step 5.
   - `5` — **the done-gate refused.** This is the enforcement point of the
     methodology's fourth invariant: a change is done only when reality is
     updated. It is not an error to route around. Every blocker is a finding;
     branch on `code`:

     | `code` | What it means | Fix |
     |---|---|---|
     | `done.checklist-unchecked` | A governance checklist line is still `- [ ]` | Do the work, then check it off with evidence. Never just tick the box. |
     | `done.reality-missing` | A component named by the PRD has no reality doc | Scaffold it; the exact command is in the accompanying `done.fix` finding. |
     | `done.reality-stale` | A reality doc's `updated:` predates the PRD's `created:` | Go back to step 1 for the component in `fields`. |
     | `done.reality-date-invalid` | A reality doc's `updated:` will not parse as a date | Rewrite it as `YYYY-MM-DD`. |

     ```bash
     jq -r '.sections[].findings[]
            | select(.severity=="fail") | [.code, .subject] | @tsv' <<<"$out"
     ```

     `--force` exists for incidents only. An agent must never pass it on its
     own initiative — if the gate refuses, report which codes fired and stop.

   - `3` — no PRD by that id on that platform, or the platform is unknown.
   - `1` — the PRD itself fails its contract. Fix it under
     skill://product/creating-prd step 7, then retry.

5. (mandatory) On exit `0` the command archives the PRD, appends `log.md`, and
   schedules the outcome review (+90 days) — findings `prd.archived`,
   `prd.log-appended`, and `prd.outcome-scheduled`. Confirm all three are
   present before reporting the change complete; the outcome review's due date
   is in the `prd.outcome-scheduled` finding's `fields`. The outcome review
   compares actuals to the PRD's success metrics — the PO owns it.
