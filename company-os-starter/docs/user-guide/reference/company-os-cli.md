---
title: company-os CLI reference
---

# `company-os` CLI reference

Extracted directly from the CLI's argument parser — every subcommand, flag,
and default listed here is what the CLI actually accepts, not an aspiration.
Examples use the Moonbeam Bakery workspace from
[tutorials/01](../tutorials/01-first-day-with-company-os.md) (team `web`,
platform `ordering`, component `online-ordering-app`).

## Global flag

```text
company-os --root <path> <command> ...
```

`--root` defaults to `$COMPANY_OS_WORKSPACE_ROOT`, then the current
directory. Every command except `init` and `scratchpad` requires running
inside (or pointed at, via `--root`) an existing workspace root — one where
at least one of `company-os/`, `platforms/`, `teams/`, `company-ontology/`,
`knowledge/` exists (or a `workspace.yaml` manifest, before the first sync).

---

## Exit codes

Every command exits with one of eight codes. This is a contract: the meaning
of a code will not change, and a reworded error message will not move an
invocation from one code to another. Scripts and agents should branch on the
code and never parse stdout.

| Code | Meaning | You get this when |
|---|---|---|
| `0` | success | the command did what you asked |
| `1` | validation failed | a `validate` gate reported `[FAIL]`, or `discover validate` / `prd validate` refused an artifact |
| `2` | usage error | unknown subcommand, bad flag, missing or invalid argument, or `company-os` with no subcommand at all |
| `3` | workspace error | you are not in a workspace root; or a platform, team, component, brief, PRD or manifest repo you named does not exist; or it exists somewhere other than where you said — `reality new --platform` for a component another platform's descriptor claims |
| `4` | artifact error | a YAML file or frontmatter block is malformed, or a `workspace.yaml` breaks its schema |
| `5` | precondition failed | a gate refused: `prd complete` before `reality/` was updated, or `prd new --from-discovery` against a brief that is not `validated` |
| `6` | external tool error | git is missing or older than 2.27, a clone or sparse-checkout failed, or `workspace sync --frozen` could not reconstruct a slice from the lock and the cache |
| `7` | interactive-mode error | a prompt was required and no terminal is attached — `company-os init` in CI without `--company/--team/--platform` |
| `8` | conflict | the destination already exists and the command refuses to overwrite it — `init` in a workspace root, a duplicate `add`, an existing `reality new` doc |

Two distinctions are worth learning, because they are the ones that change
what you do next:

- **`1` versus `5`.** `1` means an artifact is wrong. `5` means the artifact
  is fine and you have not done the work yet — most often, `prd complete`
  telling you to go update `reality/components/<id>.md`.
- **`3` versus `4`.** `3` means something is missing. `4` means something is
  there and unreadable. A `workspace.yaml` that does not exist is `3` (legal
  in monorepo mode); one that exists and breaks its schema is `4`.

Every failure code is non-zero, so a script that only tests success against
failure keeps working unchanged.

```bash
company-os validate
case $? in
  0) echo "clean" ;;
  1) echo "governance findings — read the [FAIL] lines" ;;
  5) echo "gate refused — finish the work, then retry" ;;
  *) echo "the run could not be completed; see stderr" ;;
esac
```

Diagnostics go to **stderr**; findings, listings and next-step guidance go to
**stdout**. A non-zero exit with empty stderr does not happen.

---

## `--json`

Every subcommand accepts `--json` (before the subcommand, like `--root`). It
replaces the human text on stdout with one envelope — the same records the
text renderer prints, encoded rather than reformatted, so the two can never
disagree about what was found.

```text
company-os --json <command> ...
```

```jsonc
{
  "schemaVersion": 1,                     // bumped only on a removal or repurpose
  "build": {"version": "...", "commit": "...", "goVersion": "...", "platform": "..."},
  "command": "discover",
  "action": "new",                        // omitted for commands with no action verb
  "root": "/abs/path/to/workspace",
  "exitCode": 0,                          // same code the process exits with
  "sections": [
    {
      "ordinal": 1,
      "slug": "discover-new",             // stable; the section's machine name
      "title": "2026-faster-checkout",
      "findings": [
        {
          "severity": "ok",               // "ok" | "warn" | "fail"
          "code": "discovery.created",     // stable; branch on this, not on message
          "subject": "...",               // present when the finding is about a named thing
          "path": "teams/web/product/discovery/2026-faster-checkout/brief.md",
          "message": "created teams/web/...",
          "fields": {"team": "web", "brief": "2026-faster-checkout"}
        }
      ]
    }
  ],
  "guidance": ["company-os discover validate --team web 2026-faster-checkout"],
  "error": "..."                          // present only when the command failed
}
```

What is safe to rely on:

- **`exitCode` and `severity`/`code`** are the contract. Codes are stable
  across message rewordings; messages are not. Never grep `message`.
- **`guidance`** is the next-command chain, always an array — empty, never
  absent. `guidance[0]` is the command to run next.
- **`fields`** carries typed values (counts are numbers, lists keep their
  authored order), so a consumer never has to re-parse a sentence.
- **`sections` and `findings` are always arrays.** A gate that ran and found
  nothing is `"findings": []` — not the same fact as a gate that did not run.
- Adding a field is **not** a schema break. Only a removal or a repurpose
  bumps `schemaVersion`.

The envelope is written even on failure, with `exitCode` and `error` set; the
same diagnostic still goes to stderr. So a failing run is machine-readable
without a second invocation:

```bash
out=$(company-os --json prd complete "$id" --platform "$p"); rc=$?
case $rc in
  0) jq -r '.guidance[]' <<<"$out" ;;                       # what to do next
  5) jq -r '.sections[].findings[]
            | select(.severity=="fail") | .code' <<<"$out" ;;  # what is blocking
  *) jq -r '.error' <<<"$out" >&2; exit $rc ;;
esac
```

---

## `init`

Scaffold a brand-new workspace (the four source roots: `company-os/`,
`platforms/<p>/`, `teams/<t>/`, `company-ontology/`). Refuses to run inside
an existing workspace root. The fifth root, `knowledge/`, is not scaffolded —
it appears only when a knowledge slice is synced into it.

```text
company-os init [--company NAME] [--team ID] [--platform ID]
```

All three flags are optional — omit any of them and the CLI prompts
interactively, defaulting to `My Company` / `core` / `platform-1`. Passing
all three skips every prompt.

```bash
$ company-os init --company "Moonbeam Bakery" --team web --platform ordering
initialized workspace at /Users/you/moonbeam-os
  company: Moonbeam Bakery | first team: web | first platform: ordering
next: cd /Users/you/moonbeam-os && company-os discover new --team web "<discovery title>"
```

## `add`

Grow an existing workspace: another platform, team, or component.

```text
company-os add {platform,team,component} <name> [--platform ID]
```

`--platform` is required (and only meaningful) when `kind` is `component`.

```bash
$ company-os add component online-ordering-app --platform ordering
added component 'online-ordering-app' to platform 'ordering'
next: company-os reality new --platform ordering online-ordering-app
```

## `reality`

Scaffold a component's reality doc — the file describing current, true
behavior.

```text
company-os reality new <component> --platform ID
```

`--platform` is required. Refuses to overwrite an existing reality doc (exit
`8`), and refuses a component whose descriptor names a different platform (exit
`3`).

The second refusal is newer than the command. Until 2026-08-29 a mismatched pair
was written rather than refused: the doc landed under the platform you named,
and nothing read it — the component browser looks under the *owning* platform,
and so does `prd complete`'s done-check. You saw `created` and the gate you ran
it to satisfy went on refusing, with nothing connecting the two.

A component the workspace does not know at all is still scaffolded, because
writing a reality doc before its descriptor exists is a legitimate order of work.

```bash
$ company-os reality new online-ordering-app --platform ordering
```

## `discover`

The discovery-brief workflow (team-private).

```text
company-os discover new --team ID "<title>"
company-os discover validate --team ID <brief-id>
```

`--team` is required for both actions. The positional argument is the
brief's title for `new` and its id for `validate`.

```bash
$ company-os discover new "Same-day pickup slots" --team web
$ company-os discover validate 2026-same-day-pickup-slots --team web
```

## `prd`

The PRD (platform-visible change record) workflow.

```text
company-os prd new --platform ID --team ID --components ID[,ID...] \
    [--from-discovery BRIEF-ID] [--title TEXT] [id]
company-os prd validate --platform ID <prd-id>
company-os prd complete --platform ID <prd-id> [--force]
```

`--platform` is required for every platform-addressed action.
`--components` defaults to an empty string — pass a comma-separated list.
`--force` overrides the done-check on `complete` (use sparingly; it exists
for genuinely exceptional cases, not to skip the reality-update rule
routinely).

```bash
$ company-os prd new --team web --platform ordering \
    --components online-ordering-app \
    --from-discovery 2026-same-day-pickup-slots

$ company-os prd validate 2026-same-day-pickup-slots --platform ordering

$ company-os prd complete 2026-same-day-pickup-slots --platform ordering
```

`prd complete` refuses to archive while any governance checklist item in the
PRD is unchecked, or while the reality doc for any listed component has an
`updated:` date older than the PRD's `created:` date.

### Team drafts

A draft is a PRD a team works up before it becomes platform-visible. It lives
at `teams/<team>/product/change-records/draft/<draft-id>/prd.md`, carries
`status: draft`, and `validate` tolerates it. `promote` copies it forward into
a normal change record; nothing appears under `platforms/` before that.

```text
company-os prd new --team ID --draft ["<title>"] [--platform ID] \
    [--from-discovery BRIEF-ID] [id]
company-os prd validate --team ID <draft-id>
company-os prd promote --team ID <draft-id>
company-os prd abandon --team ID <draft-id>
```

`--team` is required for all four. `--draft` is what selects the draft branch
of `prd new`; without it `prd new` behaves exactly as documented above.
`--platform` is **optional** on `prd new --draft`: it presets the draft's
`promoteTo.platform`, and when omitted you set that key in the draft's
frontmatter before promoting. The remaining process fields (`title`,
`components`, `governanceSnapshot`, `decisionOwner`) are scaffolded as
placeholders — they are a promotion-time requirement, not a creation-time
one.

`prd promote` takes no `--platform`: the target is read from the draft's
`promoteTo.platform`, which must name a platform that exists and must agree
with the draft's frontmatter `platform` field. Promotion writes
`platforms/<platform>/change-records/active/<id>/prd.md` with
`status: proposed`, flips the draft to `status: promoted`, and records a
digest of the draft on the new record — after which editing the draft fails
the promotion-integrity gate in `company-os validate`.

`prd abandon` sets `status: abandoned`, a terminal state ignored by every
gate, by `today`, and by promotion. There is no transition back.

```bash
$ company-os prd new --team web --draft "Same-day pickup slots" \
    --platform ordering --from-discovery 2026-same-day-pickup-slots

$ company-os prd validate 2026-same-day-pickup-slots --team web

$ company-os prd promote 2026-same-day-pickup-slots --team web

$ company-os prd abandon 2026-same-day-pickup-slots --team web
```

Exit codes on the draft branch, all drawn from the table above:

| Code | `prd new --draft` | `prd validate --team` | `prd promote` | `prd abandon` |
|---|---|---|---|---|
| `0` | draft created | draft is well-formed for its stage | promoted | abandoned |
| `1` | — | core frontmatter incomplete (`core.type-missing`, `core.identity-missing`, `core.status-missing`) | — | — |
| `3` | team or named platform does not exist | no draft by that id for that team | `promoteTo.platform` names a platform that does not exist | no draft by that id for that team |
| `4` | — | — | the draft declares no target platform (`promoteTo.platform` absent or still a placeholder) | — |
| `5` | `--from-discovery` brief is not `status: validated` | — | not ready: a process field (`title`, `team`, `platform`, `components`, `governanceSnapshot`, `decisionOwner`) is missing or `TODO`, or a required section (`Problem statement`, `Success metrics`, `Proposed change`) is absent — every omission is reported in one pass and no file is written | — |
| `8` | a draft with that id already exists | — | the draft is not `status: draft`, or a record with that id already exists on the target platform | the draft is not `status: draft` |

Draft validation is deliberately weaker than PRD validation:
`prd validate --team` checks the core frontmatter contract only (`type`, an
`id` identity, `status`) and does not check the body for PRD section
headings. `title`, `platform`, `components`, `governanceSnapshot` and
`decisionOwner` are required by **promotion**, not by draft validation.

## `governance`

Resolve or explain effective governance.

```text
company-os governance resolve [--team ID]
company-os governance explain <component> [--team ID]
```

Neither `action` requires `--team` at the argument-parser level, but
`resolve` needs a team id to do anything meaningful — always pass one in
practice. `explain` looks the component up across every team's already-
generated `effective-governance.yaml`, so `--team` there is informational,
not a filter.

```bash
$ company-os governance resolve --team web
$ company-os governance explain online-ordering-app
```

## `check`

Composable Definition of Ready / Definition of Done.

```text
company-os check {ready,done} --team ID --components ID[,ID...]
```

`--team` and `--components` are both required.

```bash
$ company-os check ready --team web --components online-ordering-app
$ company-os check done --team web --components online-ordering-app
```

## `validate`

Run every workspace validation gate. No arguments.

```bash
$ company-os validate
```

7 gates in a monorepo workspace, 8 when a `workspace.yaml` federation
manifest is present. Full gate-by-gate breakdown:
[how-to/run-the-validation-gate.md](../how-to/run-the-validation-gate.md).

## `deviation`

Declare a comply-or-explain deviation from a `default`-tier rule.

```text
company-os deviation declare <rule> --team ID [--rationale TEXT]
```

`--team` is required. Sets `reviewDate` 180 days out automatically; rejects
any rule that resolves to `mandatory` tier.

```bash
$ company-os deviation declare "company-standard://estimation/story-points" \
    --team web --rationale "Team forecasts with cycle time instead of points."
```

## `exception`

Request an exception to a `mandatory`-tier rule.

```text
company-os exception request <rule> --team ID --component ID --expires DATE [--reason TEXT]
```

`--team`, `--component`, and `--expires` are all required — there is no
exception without an expiry date.

```bash
$ company-os exception request "platform-standard://ordering/order-confirmation-sla" \
    --team web --component legacy-pos-bridge \
    --expires 2026-12-31 --reason "Legacy POS can't emit confirmations synchronously yet."
```

## `scratchpad`

Initialize the local-only, git-ignored scratchpad in any repo (exempt from
the workspace-root requirement, unlike every other command).

```text
company-os scratchpad init [--repo PATH]
```

`--repo` defaults to the current directory. Creates
`scratchpad/{drafts,brainstorms,personal-rules,experiments,inbox}/` and
appends ignore rules (`scratchpad/`, `.company-os.local.yaml`, `.env`,
`.env.local`) to `.gitignore`.

```bash
$ company-os scratchpad init --repo teams/web
```

## `today`

Role-aware daily view.

```text
company-os today [--role ROLE]
```

`--role` defaults to `developer`; choices are `developer`, `team-lead`,
`product-owner`, `architect`, `vp-engineering`, `director-of-product`.

```bash
$ company-os today --role product-owner
```

## `derive`

Re-derive tags and generated aggregates (feature-index, `CLAUDE.md` context
nodes) from frontmatter across the whole workspace. This is what makes the
workspace legible to the two read-side tools — Local Search indexes the
derived `tags:`, and Obsidian reads the same frontmatter tags natively. See
[explanation/obsidian-and-local-search.md](../explanation/obsidian-and-local-search.md).

```text
company-os derive
```

```bash
$ company-os derive
derive: 12 doc(s) scanned, 3 updated
```

Run this after any change that would drift derived content — `validate`
gates `[4/N]`–`[6/N]` will otherwise fail.

> `company-os graph build` is the pre-`derive` spelling. It still works and
> dispatches to exactly the same code, so existing scripts and muscle memory
> keep working, but `derive` is the name to use and the one the CLI's own
> remediation hints print.

## `ids`

List canonical IDs from `company-ontology/ids/registry.yaml`.

```text
company-os ids list [--team ID] [--platform ID] [--prefix TEXT] [--role ROLE]
```

All filters are optional and combinable. `--role`, if given, also prints a
plain-language glossary for that role above the ID listing.

```bash
$ company-os ids list --prefix component://
$ company-os ids list --platform ordering
```

## `skills`

List merged agent skills across all four layers (company, platform, team,
personal), or install the canonical ones the binary carries.

```text
company-os skills list
company-os skills install
```

```bash
$ company-os skills list
```

Layers that don't exist yet (e.g. no company or platform root in a
standalone-team workspace) simply show as empty — this command never errors
on absence. A freshly `init`ed workspace has no skills at all and reports
`0 skill(s) across 0 populated layer(s)` — until you install them.

### `skills install`

Writes the canonical skills the binary ships with into `company-os/skills/`,
flat and named the way discovery matches. This is what turns the `agentSkills`
pointer in a team's `team.yaml` into files that actually exist.

```bash
$ company-os skills install
  index company-os/skills/index.md
  node company-os/CLAUDE.md
  installed company-os/skills/completing-a-change.SKILL.md (v1.3)
  installed company-os/skills/creating-prd.SKILL.md (v1.5)
  installed company-os/skills/reality-from-prds.SKILL.md (v1.0)
  installed company-os/skills/requesting-an-exception.SKILL.md (v1.1)
  installed company-os/skills/running-discovery.SKILL.md (v1.2)

5 skill(s) in company-os/skills, 5 changed
next: review what a session now sees: company-os skills list
```

It compares the installed `version:` against the binary's and reports one code
per skill, so an agent branches on codes rather than prose:

| `code` | What happened |
|---|---|
| `skills.installed` | No file was there; it was written. |
| `skills.updated` | The installed version was older; it was replaced. Both versions are in `fields`. |
| `skills.unchanged` | Already at this version. Left alone — your edits at an unchanged version survive. |
| `skills.locally-newer` | The installed version is **newer** than this binary's. Left alone: upgrade the CLI rather than downgrade the skill. Warn. |
| `skills.unreadable` | The installed file's version could not be read. Left alone. Warn. |

Two things it deliberately does not do: it never deletes a file, and it never
merges. A team that edits a skill and bumps its `version:` keeps that file
forever; a team that edits without bumping loses those edits the next time the
skill's shipped version moves.

The command rebuilds derived artifacts before returning — the two lines above
the install output — so `company-os validate` exits 0 straight afterwards with
no `graph build` in between.

Discovery is one level deep and matches `*.SKILL.md`, except for the personal
layer (`teams/<t>/scratchpad/personal-rules/*.md`, git-ignored). Precedence,
authoring, and the `extends:` mechanism:
[how-to/use-the-agent-skills.md](../how-to/use-the-agent-skills.md).

## `workspace`

Federated multi-repo governance sync/status. Requires a `workspace.yaml`
manifest at the workspace root — without one, every action dies immediately
telling you this is a monorepo workspace and needs no federation. Requires
git ≥ 2.27. Full walkthrough:
[docs/FEDERATION-RUNBOOK.md](../../../docs/FEDERATION-RUNBOOK.md).

```text
company-os workspace sync [--frozen] [--only NAME]
company-os workspace status
```

`--frozen` materializes strictly from `workspace.lock.yaml` with no network
access (CI-safe); it dies if the lock is missing or doesn't cover a repo.
`--only` limits `sync` to a single repo by name.

Each repo declares a destination one of two ways. A single slice uses the
top-level pair:

```yaml
  - name: platform-communications
    url: https://github.com/acme/platform-communications.git
    localDirectory: platforms/communications     # where it lands
    pin: {tag: v2.1.0}                           # commit: or tag:, never a branch
    paths: [governance/, components/, reality/]  # what to pull
```

A repo contributing several areas uses `slices:` instead — one clone, one
cache, one checkout, N destinations:

```yaml
  - name: component-library
    url: https://github.com/acme/component-library.git
    pin: {tag: v1.2.0}
    slices:
      - {paths: [docs/sdd],       localDirectory: knowledge/components/component-library}
      - {paths: [.claude/skills], localDirectory: knowledge/skills/component-library}
```

Setting `slices:` alongside a top-level `localDirectory:` or `paths:` is
rejected rather than silently ignoring the top-level key, and slice targets
must be disjoint across the whole manifest — equal or nested targets are
refused.

```bash
$ company-os workspace sync
$ company-os workspace status
```

`sync` writes `workspace.lock.yaml`; `status` prints one line per repo listing
every target, and reports pin drift, slice-set drift (a target or allowlist
changed without a re-sync), and materialized-slice cleanliness, then suggests
either `workspace sync` or `validate` as the next step.

---

## `tui`

Interactive terminal UI over the current workspace: ten read-only screens and
nine forms that write.

```text
company-os tui
```

Every read-only screen runs the same code as its equivalent command. Every form
previews the exact flag-complete `company-os` invocation it is about to run and
writes nothing until you confirm; the preview is derived from the arguments that
get executed, so it cannot drift from what actually happens.

`Esc` goes back one level and quits only at the top level. `Ctrl-C` always
quits. `q` quits except while typing into a field.

The forms cover a whole change — `discover new`, `discover validate`, `prd new`,
`prd validate`, `reality new`, `prd complete` — plus `add team`,
`add platform`, and `add component`. Every picker is built when its screen is
opened, so each step offers what the step before it created.

Three deliberate omissions:

- No form for `workspace sync` or `scratchpad init`: both need values that
  cannot be derived from the workspace, and a form that supplies a
  plausible-but-wrong repo URL, commit pin, or external path is worse than no
  form.
- No field for `prd complete --force`, which overrides the check that a change
  is not done until reality is updated. Use it from a terminal or not at all.
- No command runs from a read-only screen. `discover validate` rewrites the
  brief it is given, so it is a form and never part of the discovery browser.

This subcommand has no `--json` and is not part of the agent contract — it is a
human surface. Scripts and agents use the underlying commands, where the
structured envelope and the differentiated exit codes live.

Full walkthrough, including the keys and the create-a-platform-then-a-component
sequence: [how-to/browse-and-scaffold-in-the-tui.md](../how-to/browse-and-scaffold-in-the-tui.md).

---

## See also

- [reference/configuration.md](configuration.md) — every file/env var these
  commands actually read.
- [reference/troubleshooting.md](troubleshooting.md) — symptom → cause →
  fix.
- [how-to/browse-and-scaffold-in-the-tui.md](../how-to/browse-and-scaffold-in-the-tui.md)
  — the interactive surface over these commands.
