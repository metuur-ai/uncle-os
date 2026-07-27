---
title: Use the agent skills
---

# Use the agent skills

A skill is a short, versioned Markdown file that tells a human or an agent *how
to author* one artifact correctly — the steps, and which of them are negotiable.
Skills carry no code and run nothing: `company-os` never executes a skill, it
only discovers, merges, and lints them. What the CLI enforces is the artifact
they produce.

Four layers exist, and every step is tagged with a tier, so a company can hand
down mandatory outcomes while a team — or one person — keeps its own method on
top.

## See what applies to you

```bash
$ company-os skills list
```

Real output, from a workspace with a platform skill, a team skill that extends
it, and one local personal rule:

```text
== agent skills (merged view across 4 layers) ==

layers (origin-labeled):
  [company] <none>
  [platform:ordering] creating-prd  id=skill://product/creating-prd authority=canonical
  [team:web] creating-prd-web  id=skill://product/creating-prd-web authority=team extends=platform-skill://ordering/creating-prd
  [personal] <none>

merged guidance (canonical steps first; personal rules last, non-overriding):

  creating-prd [platform:ordering, authority=canonical]
      1. (mandatory) Scaffold with `company-os prd new --from-discovery <id>`.
      2. (default) Follow the platform PRD structure.

  creating-prd-web [team:web, authority=team]
    layered on base platform-skill://ordering/creating-prd:
      [base] 1. (mandatory) Scaffold with `company-os prd new --from-discovery <id>`.
      [base] 2. (default) Follow the platform PRD structure.
      1. (default) Attach the Figma frame to every UI-visible change.

  personal rules (non-overriding — canonical mandatory steps always win):
    [personal:web] my-style

3 skill(s) across 3 populated layer(s).
```

This command never fails on absence. A fresh `company-os init` workspace ships
**no skills at all** — all four layers print `<none>`, `0 skill(s)`, exit 0.
A standalone team sees the company and platform layers empty for the same
reason. Nothing is broken; the layers simply aren't there yet.

## Where each layer lives

| Layer | Path | Notes |
|---|---|---|
| `company` | `company-os/skills/<name>.SKILL.md` | company baseline, applies to everything |
| `platform` | `platforms/<p>/skills/<name>.SKILL.md` | the platform's authoritative method |
| `team` | `teams/<t>/skills/<name>.SKILL.md` | the team's own skills and extensions |
| `personal` | `teams/<t>/scratchpad/personal-rules/*.md` | git-ignored, plain `.md`, never overriding |

Discovery globs **one level deep** and matches `*.SKILL.md` exactly, so
`skills/creating-prd/SKILL.md` is invisible — the file must be
`skills/creating-prd.SKILL.md`. (The four reference skills shipped in the
starter kit under `skills/<name>/SKILL.md` are examples to copy, not a workspace
layout; rename them on the way in.)

Personal rules are the one layer with a different shape: plain `*.md`, no
frontmatter required, and discovered only under a team's scratchpad. Create the
directory with `company-os scratchpad init --repo teams/<t>` — pointing it at the
workspace root instead puts the scratchpad where discovery will not look.

## The precedence rule

Every scaffolded `teams/<t>/team.yaml` commits the resolution order, so it is a
property of the workspace rather than a convention an agent has to be told:

```yaml
agentSkills:
  canonicalPath: skills/
  personalPath: scratchpad/personal-rules/
  precedence: canonical-mandatory > personal > canonical-default > canonical-guidance
  onConflict: prefer-canonical-and-inform-user
```

Read it as: a canonical **mandatory** step wins over anything personal; a
personal rule outranks canonical **default** and **guidance** steps; and when an
agent sets a personal rule aside, it says so rather than silently complying.

Each step declares its own tier on the numbered line:

- `(mandatory)` — not negotiable; the matching validator will block you.
- `(default)` — comply or explain. Diverging is legitimate but goes on the
  record as a [deviation](handle-a-deviation-or-exception.md).
- `(guidance)` — untracked. Ignore it freely.

Only that head line is parsed. Continuation lines are body text and never
appear in `skills list`, so put the action itself on the numbered line.

## Write a skill

Start from `templates/SKILL-template.md`:

```yaml
---
id: skill://governance/writing-adrs
type: skill
version: '1.0'
authority: canonical
appliesTo: ['company://all-teams']
inputs: [a decision with more than one defensible option]
outputs: [an ADR committed beside the artifact it constrains]
tags: []
---

# Writing an ADR

1. (mandatory) Record the decision before the code that assumes it merges.
2. (default) Name the options you rejected, and why.
3. (guidance) Length is yours.
```

Three things the gates care about:

- **`type: skill` is required.** Without it the frontmatter gate `[1/N]` fails.
- **Never hand-write `tags:`.** Leave the list empty and run
  `company-os graph build` — it derives them from the frontmatter (`authority:
  canonical` becomes `authority/canonical`) and rewrites the file in place. A
  hand-written facet the deriver would not produce fails the derived-tag gate.
- **`id` is `skill://<scope>/<name>`** and must be unique across the layers that
  outrank you — see the next section.

Then:

```bash
$ company-os graph build
$ company-os validate
```

```text
[7/7] custom skills layering (shadowing + extends resolution)
  [ok] skills layered cleanly (1 canonical, 1 team; no shadowing or dangling extends)
```

The clean line counts canonical and team skills only. Personal rules are
deliberately excluded from the totals — they live in a git-ignored scratchpad, so
counting them would make `validate` output differ between your machine and a
fresh clone. They are still scanned for shadowing.

## Extend a canonical skill, never replace it

You cannot override a company or platform skill by redefining it. Reusing a
canonical skill's `id` **or** its file name from the team or personal layer fails
gate `[7/N]`:

```text
[FAIL] skill shadowing: teams/web/skills/writing-adrs.SKILL.md reuses the canonical id
'skill://governance/writing-adrs' of company-os/skills/writing-adrs.SKILL.md — extend it
with `extends: platform-skill://...` instead of replacing it
```

The sanctioned move is a distinct id and name plus an `extends:` pointer:

```yaml
---
id: skill://product/creating-prd-web
type: skill
version: '1.0'
authority: team
extends: platform-skill://ordering/creating-prd
appliesTo: ['team://web']
tags: []
---

# Creating a PRD (web additions)

1. (default) Attach the Figma frame to every UI-visible change.
```

`skills list` then renders the base skill's steps under your own, marked
`[base]` (see the sample output above), so the reader sees one merged procedure
instead of two files to reconcile. If the pointer resolves to nothing you get a
`warn` in `skills list` and a `[FAIL]` in `validate`:

```text
dangling extends: teams/web/skills/reviewing-prd.SKILL.md declares
extends: platform-skill://ordering/nope but no such base skill exists
```

## What ships, and what you inherit

The starter kit carries four reference skills, one per lifecycle step:

| Skill | Covers |
|---|---|
| `skill://product/running-discovery` | problem signal → validated brief |
| `skill://product/creating-prd` | validated brief → PRD that passes `prd validate` |
| `skill://product/completing-a-change` | shipped code → archived PRD + updated reality |
| `skill://governance/requesting-an-exception` | a mandatory rule a component cannot satisfy |

`examples/workspace` additionally carries `skill://governance/syncing-knowledge`
for the [knowledge catalog](sync-a-knowledge-catalog.md) workflow. Copy the ones
you want into `company-os/skills/` or `platforms/<p>/skills/`, renaming to
`<name>.SKILL.md`, and edit them — they are a starting point, not a dependency.

Skills also ride federation for free: `skills/` is in the default slice
allowlist, so a platform repo's skills land in every workspace that syncs it.

## Using skills from an agent

Skills are prose for the agent; the CLI is the contract. Every canonical skill
therefore opens with the same instruction — **run every command with `--json`
and branch on the exit code**:

```bash
$ company-os --json skills list
```

The envelope's `sections` are `layers`, `merged-guidance`, `personal-rules`
(omitted entirely when empty), and `summary`. Branch on `code`, never on prose:

| Code | Fields |
|---|---|
| `skills.layer-entry` | `layer`, `scope`, `name`, `id`, `authority`, `extends` |
| `skills.step` | `step` — the head line, tier tag included |
| `skills.personal-entry` | `team`, `name` (no id/authority: a personal rule overrides nothing) |
| `skills.summary` | `skills`, `layers` — as numbers |

Two properties make this usable as an agent's plan: the tier is inside the
`step` string, so an agent can tell what it may not skip; and `guidance[0]` on
every mutating command is the next command in the workflow, so the agent follows
the chain rather than guessing. Full envelope:
[reference/company-os-cli.md § `--json`](../reference/company-os-cli.md#--json).

The boundary worth stating plainly: a skill tells an agent what to produce and
which steps are mandatory. It does not grant the agent authority to relax a
mandatory step, hand-edit anything under `generated/`, or edit a synced slice.
Those are gate failures no matter which skill or personal rule asked for them.

## Sample prompts

Nothing loads a skill automatically. An agent finds one of two ways: you name it,
or it runs `company-os skills list` and reads the merged view. Both work; naming
it is faster and leaves less room for the agent to invent a procedure.

The examples below use the ids from `examples/workspace` — team
`customer-engagement`, platform `communications`, component
`customer-notification-service`, discovery brief
`2026-per-channel-quiet-hours`. Substitute your own.

**Orient before acting.** Worth making the first prompt of any session in an
unfamiliar workspace:

```text
Run `company-os --json skills list` from the workspace root and tell me which
skills apply to team customer-engagement, which layer each comes from, and
which steps are mandatory. Don't do any work yet.
```

**Run discovery.** The skill's own steps are the plan; you supply the problem:

```text
Follow skill://product/running-discovery to open a discovery brief for team
customer-engagement on this problem: notification volume spikes overnight and
users can't mute per channel. Run each command with --json, follow guidance[0],
and stop at the first non-zero exit code and show me the failing finding codes.
```

**Create a PRD from a validated brief.** State the inputs the skill requires so
the agent doesn't improvise them:

```text
Follow skill://product/creating-prd to scaffold a PRD for platform
communications from discovery 2026-per-channel-quiet-hours, components
customer-notification-service. Treat the two non-failure findings as
load-bearing: if you see prd.governance-unresolved or prd.reality-note, stop
and tell me rather than proceeding. Do not hand-edit the brief's status to get
past a step 1 failure.
```

**Complete a change.** This is the prompt most worth being explicit in, because
the skill's whole point is that merging is not done:

```text
Follow skill://product/completing-a-change for PRD <id> on platform
communications. Reality first: update reality/components/customer-notification-service.md
and its `updated:` date to reflect what actually shipped, then check off the
governance checklist items that are genuinely satisfied, with evidence links,
then run prd complete. Never pass --force. If it exits 5, report the done.*
codes that fired and stop — don't tick a box to get past one.
```

**Request an exception.** The prompt that matters here is the one that resists
the shortcut:

```text
customer-notification-service can't move to the current message envelope this
quarter. Follow skill://governance/requesting-an-exception: confirm the tier with
governance explain first, and if message-schema really is mandatory, draft the
exception with a real --expires and --reason plus compensatingControls. Leave
approvedBy empty — that's the rule owner's signature, not yours. Show me the
entry before you write it.
```

**Sync the knowledge catalog.** The skill maps the repo and directory through
`workspace.yaml`; the agent should not fetch anything itself:

```text
Follow skill://governance/syncing-knowledge to add
github.com/acme/component-library's docs/sdd to the catalog at
knowledge/components/component-library, pinned to tag v1.2.0. Write the
workspace.yaml entry, run workspace status, then sync, then graph build, then
validate. Do not edit anything under knowledge/ directly.
```

**Author or extend a skill.** Useful when a team keeps re-explaining the same
local step:

```text
Our team always attaches the Figma frame to UI-visible PRDs. Write that as a
team skill under teams/customer-engagement/skills/ that extends the platform's
creating-prd rather than redefining it: distinct id and file name, `extends:
platform-skill://communications/creating-prd`, `type: skill`, empty tags. Then
run graph build and validate and show me gate [7/N].
```

**When a mandatory step blocks you.** The failure mode to prompt against is an
agent that "resolves" a gate by editing the thing being checked:

```text
validate is failing. Diagnose it and fix the cause, not the symptom. You may not
hand-edit anything under generated/, any file in a synced slice, or a
frontmatter status field to make a gate pass. If the only fix is a governance
decision, stop and tell me which one.
```

What makes these work, and what to carry into your own:

- **Name the skill by id.** `skill://product/creating-prd` is unambiguous;
  "write a PRD" invites the agent to invent a structure.
- **Name the inputs the skill declares.** Every skill's frontmatter lists
  `inputs:` and `outputs:`. Supplying the inputs is most of the prompt.
- **Ask for `--json` and exit-code branching.** Then the agent reports finding
  codes instead of paraphrasing errors.
- **Say where to stop.** "Stop at the first non-zero exit" turns a silent
  workaround into a question you get to answer.
- **Forbid the shortcuts explicitly.** `--force`, editing `generated/`, editing a
  slice, flipping a `status:` field. The gates catch all four, but catching them
  after a commit is more expensive than not doing them.
- **Let personal rules stay personal.** You don't need to restate them in the
  prompt; an agent reading `skills list` already sees them, and already knows a
  canonical mandatory step outranks them.

## Related

- [reference/company-os-cli.md § `skills`](../reference/company-os-cli.md#skills) — the command reference
- [how-to/run-the-validation-gate.md](run-the-validation-gate.md) — what gate `[7/N]` checks
- [how-to/handle-a-deviation-or-exception.md](handle-a-deviation-or-exception.md) — the record a `(default)` divergence needs
- [explanation/how-it-fits-together.md](../explanation/how-it-fits-together.md) — how skills relate to Local Search and the CLI
- [explanation/github-mcp-and-automation.md](../explanation/github-mcp-and-automation.md) — what a skill may *not* delegate to an external tool
- `templates/SKILL-template.md` — the authoring template
