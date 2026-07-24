# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`company-os-starter/` is the reference implementation of a **Federated Company / Platform / Team Operating System**: a Git-based methodology for running governance, product, and engineering work as validated markdown + YAML artifacts. It is not an application — the only executable is one Python CLI (`bin/company-os`); everything else is templates, canonical skills, schemas, docs, and a fully worked example workspace.

The design intent, encoded throughout: **strict on artifacts, flexible on process.** Validators check outputs (schemas, links, ownership reconciliation, expiries) and never the method used to produce them.

## Commands

The CLI has no build step. Setup and use:

```bash
pip install pyyaml                              # only dependency; Python 3.9+
export PATH="$PWD/company-os-starter/bin:$PATH"
cd examples/workspace                           # the CLI operates on a workspace root
```

Workspace root resolution order: `--root` flag → `$COMPANY_OS_WORKSPACE_ROOT` → current directory.

Core lifecycle (run from a workspace root, e.g. `examples/workspace`):

```bash
company-os governance resolve --team <team>          # regenerate generated/effective-governance.yaml
company-os governance explain <component>            # why a rule applies to a component
company-os discover new --team <team> "<title>"      # scaffold a discovery brief
company-os discover validate --team <team> <brief-id>
company-os prd new --team <t> --platform <p> --components <id,...> --from-discovery <id>
company-os prd validate --platform <p> <prd-id>
company-os prd complete --platform <p> <prd-id>      # enforces reality-updated done-check; --force to override
company-os check ready|done --team <t> --components <id,...>
company-os deviation declare <rule> --team <t>       # comply-or-explain (default-tier rules only)
company-os exception request <rule> --team <t> --component <id> --expires <date>
company-os validate                                  # workspace CI gate (see below)
company-os graph build                               # derive tags from frontmatter for the whole workspace
company-os today --role developer|product-owner|...  # role-aware daily view
company-os scratchpad init --repo <path>             # create git-ignored local working area

# Growth + flavor (Phases 1–3): scaffold new units and inspect the ontology/skills
company-os init                                      # scaffold a fresh workspace (refuses to re-init)
company-os add platform|team|component ...           # grow the federation from the same templates as init
company-os reality new --platform <p> <component-id> # scaffold reality/components/<id>.md
company-os ids list                                  # list canonical IDs from ids/registry.yaml
company-os skills list                               # merged four-layer skill view

# Federation (Phase 4 — Option B; requires a workspace.yaml manifest + git ≥2.27)
company-os workspace sync [--frozen] [--only <repo>] # materialize read-only governance slices + write workspace.lock.yaml
company-os workspace status                          # per-repo pin/lock/slice drift status
```

There is no test suite. To exercise changes, run the CLI against `examples/workspace` and confirm `company-os validate` still exits 0. `docs/TUTORIAL.md` is an end-to-end walkthrough with real command output — use it as the acceptance path after changing CLI behavior.

## Architecture

### The federation (three OS layers + ontology)

A workspace is one directory containing four peer roots. Platforms are "teams" relative to Company OS; the same tier/deviation model applies at every level.

- `company-os/standards/` — company baseline controls applied to everything.
- `platforms/<p>/` — the authoritative catalog. Holds `components/<id>.yaml` (descriptors), `governance/requirements.yaml`, `reality/` (current-state truth), `change-records/active/` (live PRDs), `archive/prds/` (completed + outcomes), `skills/`.
- `teams/<t>/` — `ownership/components.yaml`, `governance/{deviations,exceptions}.yaml`, `product/discovery/`, `standards/definition-of-{ready,done}.md`, and `generated/`.
- `company-ontology/` — canonical IDs (`ids/registry.yaml`), concept notes, bounded contexts, context maps. Referenced by the other layers, never redefined by them.

### Invariants the CLI enforces (know these before editing)

1. **Rule tiers.** Every requirement/control is `mandatory` (only escapable via an expiring, approved *exception*), `default` (comply-or-explain via a *deviation*), or `guidance` (untracked). Mandatory rules must be written as verifiable **outcomes, not implementations** — this is what preserves team flexibility. `resolve_team_governance` rejects any deviation aimed at a mandatory rule.

2. **Single source of truth.** The component descriptor (`platforms/<p>/components/<id>.yaml`) is authoritative for both component↔platform relationships *and* the accountable team. `company-os validate` step [1/7] fails if a team's ownership registry claims `accountable` but the descriptor's `ownership.accountableTeam` disagrees. Humans edit one file; tooling reconciles the rest.

3. **Generated files are derived, never hand-edited.** `teams/<t>/generated/effective-governance.yaml` is produced by `governance resolve` (merges company baseline + platform requirements filtered by relationship/componentType + team deviations). CI regenerates and diffs. The same rule applies to frontmatter `tags:` — `graph build` derives them from IDs (`derive_tags`); editing tags by hand is overwritten on the next build.

4. **A change is done only when reality is updated.** `prd complete` refuses to archive while any `- [ ]` governance checklist item is unchecked, or while a component's `reality/components/<id>.md` has an `updated:` date older than the PRD's `created:` date. On success it moves the PRD to `archive/prds/`, writes an `outcome.md` due in 90 days, and appends `log.md`.

5. **Deviations and exceptions expire.** `validate` step [2/7] fails on any past `reviewDate` (deviation) or missing/past `expires` (exception).

### Product lifecycle (where artifacts live and move)

Discovery is team-private (`teams/<t>/product/discovery/`, status draft→validated). Once a PRD proposes changing platform reality it becomes a platform-visible change record (`platforms/<p>/change-records/active/`, status proposed→completed). Completion archives it and schedules an outcome review. `prd new --from-discovery` requires the brief to be `status: validated` and copies its Problem/Success sections forward.

### Ontology, tags, and `@spec` (the graph layer)

`docs/ONTOLOGY-GUIDE.md` is the spec for the semantic layer. Three mechanisms, one rule — **IDs are canonical; tags and wikilinks are derived**:

- **Meaning:** canonical URL-style IDs (`component://`, `capability://`, `req://`, `context://`) registered once in `ids/registry.yaml`.
- **Aboutness:** faceted nested tags (`#kind/prd`, `#platform/...`, `#tier/mandatory`) derived from frontmatter by `graph build`.
- **Satisfaction:** platform requirements carry numbered EARS clauses (`R1…Rn`); code and tests bind to them with grep-able `@spec req://.../<id>@<version>#<clause>` markers. A mandatory clause with zero test-side `@spec` sites is meant to block completion.

Bounded contexts declare a `ubiquitousLanguage` and `forbiddenTerms`; canonical docs are vocabulary-linted against their context. Note: `docs/ONTOLOGY-GUIDE.md` describes several validation/trace subcommands (`validate --ontology`, `spec trace`) as the target design — the current `bin/company-os` implements `graph build` and the core `validate` gate but not those; treat the guide as the roadmap, the CLI source as ground truth for what exists today.

## Conventions when editing

- **Editing the CLI (`bin/company-os`):** it is a single self-contained file with no framework. Preserve the `die`/`ok`/`warn`/`fail` output helpers and the `frontmatter()` parser (expects `^---\n...\n---\n` exactly). Every mutating command prints the next command in the workflow — keep that guidance chain intact.
- **Artifacts carry YAML frontmatter** with a `type:` and `tags:`. Never hand-write `tags:` — set the source fields and run `graph build`. Never hand-edit anything under `generated/`.
- **Templates** in `templates/` are the contract surface; the `*_TEMPLATE` strings in `bin/company-os` must stay in sync with the section names the corresponding `validate` command greps for (e.g. changing a PRD `## Success metrics` heading requires updating both the template and `cmd_prd`).
- **Skills** (`skills/*/SKILL.md`) tag each step `(mandatory)` / `(default)` / `(guidance)`. On conflict, canonical mandatory steps win over personal rules in `scratchpad/personal-rules/` — and the agent should say so.
