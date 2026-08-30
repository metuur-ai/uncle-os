---
id: skill://product/creating-prd
type: skill
version: '1.5'
authority: canonical
appliesTo: ['company://all-platforms']
inputs:
- {a discovery brief with status: validated (or an explicit problem statement)}
outputs: [change-records/active/<id>/prd.md passing `company-os prd validate`]
tags: [authority/canonical]
---

# Creating a PRD

**Agents: run every command with `--json` and branch on the exit code.** The
envelope carries `exitCode`, a per-finding `severity`/`code`, and a `guidance`
array holding the next command. Codes are a contract; the English in `message`
is not. Never parse prose. Full envelope:
[reference/company-os-cli.md § `--json`](../../docs/user-guide/reference/company-os-cli.md#--json).

**Two entry points, one artifact.** Steps 1–8 are the direct path: you already
know the platform, the components and the shape of the change, so the PRD
starts platform-visible. When the change is still being worked up, start
instead from [the draft branch](#the-draft-branch) at step 9 — a draft lives
in the team tree until you promote it. Both paths end at the same change
record under `platforms/<platform>/change-records/active/`.

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

## The draft branch

Use this when the change is not platform-ready yet. A draft is a PRD the team
owns and `validate` tolerates: it lives at
`teams/<team>/product/change-records/draft/<draft-id>/prd.md` with
`status: draft`, and nothing appears under `platforms/` until you promote it.
Promotion — not editing — is what makes a change platform-visible.

9. (default) Scaffold a draft:

    ```bash
    company-os --json prd new --team <team> --draft "<title>" \
       [--platform <platform>] [--from-discovery <discovery-id>]
    ```

    `--platform` is optional here and only presets the draft's
    `promoteTo.platform` — the promotion target. Leave it off and set
    `promoteTo.platform` in the draft's frontmatter before you promote.
    `--from-discovery` requires a `status: validated` brief and copies its
    Problem and Success sections forward, exactly as on the platform path.

    - `0` — the draft's path is in the created finding's `fields`. The next
      thing to do is edit that file; do not promote yet.
    - `5` — the discovery brief is not `status: validated`. Go back to
      skill://product/running-discovery.
    - `3` — the team, or the platform you named, does not exist.
    - `8` — a draft with this id already exists in the team's draft
      directory. Open it rather than scaffolding a second one.

10. (mandatory) Validate the draft as you work it up:

    ```bash
    company-os --json prd validate --team <team> <draft-id>
    ```

    A draft is held to the core frontmatter contract only — `type`, an `id`
    identity, and `status`. `title`, `platform`, `components`,
    `governanceSnapshot` and `decisionOwner` are **not** required at draft
    time, and the body is not checked for the PRD section headings. Those are
    requirements of the **promotion** contract, not of draft validation. Do
    not fill them with placeholder text to satisfy a check that is not being
    made.

    - `0` — the draft is well-formed for its stage.
    - `1` — the core frontmatter is incomplete. Branch on `code` exactly as
      in step 7's table (`core.type-missing`, `core.identity-missing`,
      `core.status-missing`).
    - `3` — no draft by that id in that team's draft directory.

11. (mandatory) Promote when the change is real:

    ```bash
    company-os --json prd promote --team <team> <draft-id>
    ```

    The target platform is not a flag on `promote`: it is read from the
    draft's `promoteTo.platform`, which must agree with the frontmatter
    `platform` field. Promotion copies the draft forward to
    `platforms/<platform>/change-records/active/<id>/prd.md` with
    `status: proposed`. The draft stays where it is, flipped to
    `status: promoted` and frozen — editing it afterwards fails the
    promotion-integrity gate in `company-os validate`.

    Promotion is where the full PRD contract binds, and it is a precondition
    of `prd complete`: a draft can never be completed. Promote it, then
    complete the resulting change record under
    skill://product/completing-a-change.

    - `0` — promoted. Pick up at step 7 against the new record:
      `company-os --json prd validate <id> --platform <platform>`.
    - `5` — not ready. The refusal names, in one pass, every process field
      still missing or `TODO` (`title`, `team`, `platform`, `components`,
      `governanceSnapshot`, `decisionOwner`) and every missing section
      (`Problem statement`, `Success metrics`, `Proposed change`). Nothing
      was written. Supply them and re-run.
    - `4` — the draft declares no target platform. Set
      `promoteTo.platform`.
    - `3` — `promoteTo.platform` names a platform that does not exist.
    - `8` — the draft is not `status: draft`, or a change record with that id
      already exists on the target platform.

12. (default) Retire a draft that will not be promoted:

    ```bash
    company-os --json prd abandon --team <team> <draft-id>
    ```

    Sets `status: abandoned`, after which every gate, `today`, and promotion
    ignore it. There is no way back — create a new draft instead. Never `rm`
    a draft; that is hand-repair of tooling state.

    - `8` — the draft is not `status: draft`.
