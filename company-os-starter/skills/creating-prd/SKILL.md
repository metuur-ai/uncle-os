---
id: skill://product/creating-prd
version: '1.4'
authority: canonical
appliesTo: ['company://all-platforms']
inputs:
- {a discovery brief with status: validated (or an explicit problem statement)}
outputs: [change-records/active/<id>/prd.md passing `company-os prd validate`]
tags: [authority/canonical, kind/skill, process/prd]
---

# Creating a PRD

**Agents: run every command with `--json` and branch on the exit code.** The
envelope carries `exitCode`, a per-finding `severity`/`code`, and a `guidance`
array holding the next command. Codes are a contract; the English in `message`
is not. Never parse prose. Full envelope:
[reference/company-os-cli.md § `--json`](../../docs/user-guide/reference/company-os-cli.md#--json).

1. (mandatory) Scaffold from the validated discovery — never copy an old PRD:

   ```bash
   company-os --json prd new --team <team> --platform <platform> \
      --components <id,...> --from-discovery <discovery-id>
   ```

   This injects the governance snapshot for the affected components,
   including any approved team deviations.

   - `0` — the PRD path is in the `prd.created` finding's `fields`.
   - `5` — the discovery brief is not `status: validated`. Go back to
     skill://product/running-discovery. Do **not** hand-edit the brief's
     status to get past this.
   - `3` — the team, platform, or a named component does not exist. The error
     names the unknown id and suggests close matches; use one or ask.
   - `8` — a PRD with this id already exists. Open it.

   Two findings from a *successful* run are load-bearing and easy to miss
   because they are not failures:

   - `prd.governance-unresolved` (`warn`) — a component the team's effective
     governance does not name. Its rules were **not** injected into the
     checklist. Run `company-os --json governance resolve --team <team>` and
     re-scaffold, or raise it. Do not proceed assuming the checklist is
     complete.
   - `prd.reality-note` — a listed component has no reality doc yet. Scaffold
     it now using the command in `guidance`; step 1 of
     skill://product/completing-a-change will block on it otherwise.

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

   ```bash
   out=$(company-os --json prd validate <id> --platform <platform>); rc=$?
   jq -r '.sections[].findings[] | select(.severity=="fail")
          | [.code, .subject] | @tsv' <<<"$out"
   ```

   - `0` — passes. `warn` findings may remain; report them, do not silence them.
   - `1` — rejected. Branch on `code`:

     | `code` | What to do |
     |---|---|
     | `prd.process-field` | The field named in `fields` is missing or still `TODO` — most often `decisionOwner`. |
     | `product.section-heading-missing` | Restore the `## ` heading named in `fields.section`. Always blocking, whatever template produced the PRD. |
     | `product.section-empty` (`fields.enforced: true`) | The team opted into blocking section enforcement. Write the section. |
     | `core.type-missing` / `core.identity-missing` / `core.status-missing` / `core.updated-missing` | Frontmatter is incomplete. Set the field the code names. |

   - `3` — no PRD by that id on that platform. Check both arguments.

   A `product.section-empty` finding with `fields.enforced: false` is a
   `warn`: format guidance only, and the team may use its own structure.

8. (mandatory) If a mandatory requirement cannot be met, do NOT delete the
   checklist line. Use skill://governance/requesting-an-exception.
