---
id: skill://product/reality-from-prds
type: skill
version: '1.0'
authority: canonical
appliesTo: ['company://all-platforms']
inputs: [a component descriptor, the platform's completed PRDs, a workspace root]
outputs: [a reality doc that passes `company-os validate`]
tags: [authority/canonical]
---

# Writing a Reality Doc from PRDs

Reality is a snapshot, PRDs are the ledger. This skill writes the snapshot
from the ledger — and nothing else.

**Precondition (mandatory): this skill works from completed PRDs only.** A
reality doc describes what a component does TODAY. Completed (archived) PRDs
under `platforms/<platform>/archive/prds/` are shipped change and belong in
reality; records under `platforms/<platform>/change-records/active/` are
proposals and do not — note them, exclude them. History lives in the PRDs,
never in the reality doc: a removed feature simply stops being described.

**Agents: run every command with `--json` and branch on the exit code.** The
envelope carries `exitCode`, a per-finding `severity`/`code`, and a `guidance`
array holding the next command. Codes are a contract; the English in `message`
is not. Never parse prose. Full envelope:
[reference/company-os-cli.md § `--json`](../../docs/user-guide/reference/company-os-cli.md#--json).

1. (mandatory) Resolve the template the doc must follow: read
   `platforms/<platform>/templates/reality-component.md` if it exists,
   otherwise `company-os/templates/reality-component.md`, otherwise use the
   canonical shape (frontmatter contract plus present-tense body sections —
   built-in: `## Business rules`, `## Current limitations`). Team templates are
   not consulted: `reality new` skips them too. Substitute the template's
   placeholders: `reality-<component-id>`, `<YYYY-MM-DD>`, `<Component Name>`.
2. (mandatory) Collect the ledger: every
   `platforms/<platform>/archive/prds/<id>/prd.md` whose `components:` names
   this component. List active records from
   `platforms/<platform>/change-records/active/` as in flight, excluded from
   the doc.
3. (mandatory) Determine the delta. No reality doc exists — draft it from all
   completed PRDs. It exists — compare its `updated:` against each completed
   PRD's `created:`; a PRD newer than `updated:` is not yet included, so read
   it and classify its impact on the doc: behavior added, behavior changed,
   behavior removed.
4. (mandatory) Present the plan and wait for confirmation. Show the proposed
   doc (new) or the delta — every item citing its PRD. A removal must be
   stated explicitly: "rule X drops out because PRD Y removed it." Do not
   write anything until the user confirms.
5. (mandatory) Write `platforms/<platform>/reality/components/<component>.md`
   following the resolved template: present tense, current behavior only — no
   history, no plans; cite the PRD next to every rule it produced and list
   every source PRD under `## Sources`; frontmatter contract:
   `type: component-reality`, `id: reality-<component>`,
   `authority: canonical`, `updated:` set to today and never predating any
   completed PRD's `created:`. Write no `tags:` — `graph build` derives them.
6. (mandatory) Derive tags and verify:

   ```bash
   company-os --json graph build
   out=$(company-os --json validate); rc=$?
   ```

   - `0` — gates pass. If a PRD prompted this update and is still active, the
     next step in the workflow is
     `company-os --json prd complete <id> --platform <platform>` — branch on
     its exit code (skill://product/completing-a-change).
   - `1` — a gate failed. Branch on the findings' `code`:
     `frontmatter.tags-drift` — tags were hand-written; delete the `tags:`
     key and rerun `graph build`. `core.type-missing` — the `type:` key
     broke; restore `type: component-reality`. Report any other code and stop.
   - `4` — the doc's frontmatter is unreadable; fix the YAML and retry.
   - `3` — you are not in a workspace root.
