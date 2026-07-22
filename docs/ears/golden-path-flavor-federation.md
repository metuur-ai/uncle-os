# Golden Path, Flavor & Federation — EARS Specifications

Arrow of intent: HLD → LLD → EARS → code/tests. To change a behaviour, update these tables first.

> Requirement IDs carry the `GPF-` prefix to avoid collision with `docs/ears/federation-enrichment.md`.
> Units 6–7 define the **federation base** that the federation-enrichment spec builds on.

## Unit 1: Init wizard, golden path & growth

**Why:** The terminal-only, multi-step setup is the first adoption blocker; a new user must reach a working workspace from one command and one document — and the guidance chain must stay unbroken through `prd complete` (reality doc included) and through day-2 growth (adding teams/platforms/components without hand-copying YAML).

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-1.1 | WHEN `company-os init` is run in an empty directory, THE SYSTEM SHALL interactively scaffold the four peer roots (`company-os/`, `platforms/<p>/`, `teams/<t>/`, `company-ontology/`), seed `ids/registry.yaml`, and print the first golden-path command. |
| GPF-R-1.2 | IF `company-os init` is run inside an existing workspace, THE SYSTEM SHALL refuse and exit non-zero without modifying any file. |
| GPF-R-1.3 | WHEN `init` is given `--company`, `--team`, and `--platform`, THE SYSTEM SHALL scaffold non-interactively with those values. |
| GPF-R-1.4 | IF the wizard is aborted before completion, THE SYSTEM SHALL leave no partially scaffolded workspace behind. |
| GPF-R-1.5 | WHEN any mutating command (existing or new) completes, THE SYSTEM SHALL print the next command in the workflow (guidance chain preserved and extended). |
| GPF-R-1.6 | THE SYSTEM SHALL ship one golden-path document whose steps match the CLI's printed guidance chain from environment prerequisites (Python, pyyaml, PATH) through `prd complete`. |
| GPF-R-1.7 | WHEN `init` completes on a fresh directory, THE SYSTEM SHALL produce a workspace on which `company-os validate` exits 0. |
| GPF-R-1.8 | WHEN `prd new` completes and any target component lacks `reality/components/<id>.md`, THE SYSTEM SHALL print a scaffold command (`company-os reality new --platform <p> <component-id>`) that generates the reality doc from `templates/reality-component.md`, so the guidance chain remains unbroken through `prd complete`. |
| GPF-R-1.9 | WHEN `company-os add platform\|team\|component` is run, THE SYSTEM SHALL scaffold the new unit from the same templates `init` uses, and SHALL refuse to overwrite any existing file. |

## Unit 2: ID discovery & error guidance

**Why:** Users must currently already know canonical `component://`-style IDs; closing this gap removes the highest-frequency friction without new surface.

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-2.1 | WHEN `company-os ids list` is run, THE SYSTEM SHALL list canonical IDs read from `company-ontology/ids/registry.yaml` (never a parallel index). |
| GPF-R-2.2 | WHERE `--team`, `--platform`, or `--prefix` is given to `ids list`, THE SYSTEM SHALL filter the listing accordingly. |
| GPF-R-2.3 | IF a command fails because a supplied component or rule ID is unknown, THE SYSTEM SHALL name the unknown ID and suggest up to 3 closest registered IDs. |

## Unit 3: Role views & terminology translation

**Why:** Terminology load (deviation, exception, EARS, ontology, bounded context) blocks non-technical users; translation must aid reading without forking the ubiquitous language.

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-3.1 | WHERE `--role` is given to a command that supports it — which SHALL include at least `today` and `ids list` — THE SYSTEM SHALL display plain-language labels for canonical terms alongside (never instead of) the canonical term. |
| GPF-R-3.2 | THE SYSTEM SHALL keep all artifacts, IDs, tags, and validator messages in canonical vocabulary; translation is display-only. |
| GPF-R-3.3 | IF a canonical term has no translation entry for the active role, THE SYSTEM SHALL display the canonical term unchanged. |

## Unit 4: Template overrides

**Why:** Teams need flavor (their own wording, examples, language) in scaffolded artifacts without gaining the power to weaken the artifact contract. Required headings are matched in English only; no alias mechanism exists (localization deferred — see HLD Non-Goals).

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-4.1 | WHEN a scaffolding command needs a template, THE SYSTEM SHALL resolve it first-found-wins through `teams/<t>/templates/` → `platforms/<p>/templates/` → `company-os/templates/` → the built-in `*_TEMPLATE` string. |
| GPF-R-4.2 | WHERE no override file exists, THE SYSTEM SHALL scaffold byte-identically to today's built-in templates. |
| GPF-R-4.3 | THE SYSTEM SHALL inject frontmatter (`type:`, IDs, tag-source fields), status fields, and `deviations.yaml`/`exceptions.yaml` row seeds itself, regardless of which template was resolved. |
| GPF-R-4.4 | IF an artifact produced from any template lacks a required section heading, THE SYSTEM SHALL fail `validate` naming the missing heading (outputs are validated, templates are not). |

## Unit 5: Custom skills layering

**Why:** Teams and individuals need to add process guidance additively while canonical skills remain the single authority — conflicts must be impossible by construction, not adjudicated. (The rule that *agents* must announce canonical-over-personal overrides lives in the skill/agent guidance docs; below is only what the CLI itself can verify.)

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-5.1 | WHEN skills are discovered, THE SYSTEM SHALL merge four layers — `company-os/skills/`, `platforms/<p>/skills/`, `teams/<t>/skills/`, and personal `scratchpad/personal-rules/` — labeling each skill with its origin layer. |
| GPF-R-5.2 | IF a team or personal skill reuses the ID/name of a canonical skill (shadowing), THE SYSTEM SHALL fail `validate` identifying both files. |
| GPF-R-5.3 | WHERE a skill declares `extends: platform-skill://…` frontmatter, THE SYSTEM SHALL present it layered on the base skill (base steps plus extension steps); the `platform-skill://` scheme SHALL be documented as a canonical ID scheme in the ontology conventions. |
| GPF-R-5.4 | WHEN the merged skill view is rendered, THE SYSTEM SHALL order canonical `(mandatory)` steps above personal rules and label personal rules as non-overriding. |

## Unit 6: Federation activation & manifest

**Why:** Multi-repo organizations need one governance view without adopting submodules or copying files; monorepo users must be untouched. This unit and Unit 7 define the federation base that `docs/ears/federation-enrichment.md` builds on.

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-6.1 | IF no `workspace.yaml` exists at the workspace root, THE SYSTEM SHALL behave exactly as monorepo mode today (zero behavior change). |
| GPF-R-6.2 | WHEN `workspace.yaml` is present at the workspace root, THE SYSTEM SHALL operate in federated mode, treating each listed repo as a source of read-only slices. |
| GPF-R-6.3 | THE SYSTEM SHALL accept only explicit pins per repo — `commit: <sha>` or `tag: <tag>` — and SHALL reject floating refs (branch names, bare `ref:`). |
| GPF-R-6.4 | WHERE a pin is a tag, THE SYSTEM SHALL resolve it to a commit SHA at sync time and record the SHA in `workspace.lock.yaml`. |

## Unit 7: Workspace sync, lock & status

**Why:** Federation must be reproducible, minimal (governance slices only — Option B), CI-safe offline, and honest about its day-2 costs (pin bumps, commit-back, ownership transfer).

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-7.1 | WHEN `company-os workspace sync` runs, THE SYSTEM SHALL fetch only governance-relevant paths (`governance/`, `components/`, `governance/requirements.yaml`, `reality/`, `skills/`, `templates/`) via sparse/filtered shallow git — never full repo content. |
| GPF-R-7.2 | WHEN sync completes, THE SYSTEM SHALL write every resolved repo SHA and per-slice content hashes to `workspace.lock.yaml`. |
| GPF-R-7.3 | WHILE `--frozen` is active, THE SYSTEM SHALL perform no network access and materialize strictly from `workspace.lock.yaml`; IF the lock is missing or does not cover the manifest, THE SYSTEM SHALL fail with a clear message. |
| GPF-R-7.4 | WHEN `company-os workspace status` runs, THE SYSTEM SHALL report drift between manifest pins, the lock file, and the materialized slices. |
| GPF-R-7.5 | IF a materialized slice's content hash differs from the hash recorded in `workspace.lock.yaml` (hand-edit), THE SYSTEM SHALL fail `validate` naming the edited path (slices are derived content, like `generated/`). |
| GPF-R-7.6 | WHERE federated slices are materialized, THE SYSTEM SHALL resolve templates (Unit 4) and skills (Unit 5) across them identically to monorepo layout. |
| GPF-R-7.7 | IF git is unavailable, THE SYSTEM SHALL fail only federation commands with a clear message; monorepo commands SHALL not require git. |
| GPF-R-7.8 | THE SYSTEM SHALL ship a federation runbook document covering pin-bump → sync → resolve → commit-back, the two-PR ownership-transfer flow, and a reusable per-repo CI recipe using `sync --frozen`. |

## Unit 8: Fail-fast & parser consistency

**Why:** Silent success in the wrong directory (F8) and frontmatter parsing differences between artifact types (F11) erode trust in every other guarantee in this spec.

| ID    | EARS statement |
| ----- | ------------- |
| GPF-R-8.1 | IF a workspace-scoped command is run outside a workspace root, THE SYSTEM SHALL fail fast, naming the root-resolution order (`--root` → `$COMPANY_OS_WORKSPACE_ROOT` → cwd) instead of succeeding silently. |
| GPF-R-8.2 | THE SYSTEM SHALL parse YAML frontmatter through the single `frontmatter()` contract (`^---\n...\n---\n`) identically across all artifact types. |
