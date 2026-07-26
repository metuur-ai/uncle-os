---
id: skill://product/running-discovery
version: '1.1'
authority: canonical
appliesTo: ['company://all-teams']
inputs: ['a problem signal (tickets, metrics, stakeholder ask)']
outputs: [product/discovery/<id>/brief.md passing `company-os discover validate`]
tags: [authority/canonical, kind/skill, process/discovery]
---

# Running Discovery

**Agents: run every command with `--json` and branch on the exit code.** The
envelope carries `exitCode`, a per-finding `severity`/`code`, and a `guidance`
array holding the next command. Codes are a contract; the English in `message`
is not. Never parse prose, and never infer failure from the absence of a word.
Full envelope:
[reference/company-os-cli.md § `--json`](../../docs/user-guide/reference/company-os-cli.md#--json).

1. (mandatory) Scaffold — never start from a blank file or an old brief:

   ```bash
   company-os --json discover new "<title>" --team <team-id>
   ```

   - `0` — read the created path from the `discovery.created` finding's
     `fields.path`, and the next command from `guidance[0]`.
   - `3` — the team does not exist. Do not create it; ask.
   - `8` — a brief with this slug already exists. Open it rather than
     scaffolding a second one.

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
7. (mandatory) Validate. A brief that fails validation cannot become a PRD:

   ```bash
   out=$(company-os --json discover validate <id> --team <team-id>); rc=$?
   jq -r '.sections[].findings[] | select(.severity=="fail") | .code' <<<"$out"
   ```

   - `0` — the brief passes. It may still carry `warn` findings; see below.
   - `1` — rejected. Every blocking reason is a finding with
     `severity: "fail"`. Fix by `code`, never by message:

     | `code` | What to do |
     |---|---|
     | `product.section-heading-missing` | Restore the `## ` heading named in `fields.section`. Always blocking, whatever template produced the brief. |
     | `product.section-empty` (`fields.enforced: true`) | The team opted into blocking section enforcement. Write the section named in `fields.section`. |
     | `core.type-missing` / `core.identity-missing` / `core.status-missing` / `core.updated-missing` | Frontmatter is incomplete. Set the field the code names; do not invent an `id`. |

   - `3` — no brief by that id. Re-check the slug from step 1.

   A `product.section-empty` finding with `fields.enforced: false` is a
   `warn`, not a failure — the team has not opted into section enforcement and
   may use its own structure. Report it; do not treat it as blocking, and do
   not "fix" it by rewriting the brief into the default shape.

8. (mandatory) Record the **Decision** — `validated` or `invalidated`.
   Invalidated discoveries are kept: a killed idea with a reason is knowledge.
   Only a brief whose `discovery.validated` finding reports
   `fields.status: "validated"` can be passed to `prd new --from-discovery`.
