# Company OS / Team OS and BMAD artifacts: platform and component levels

**Date:** 2026-08-03
**Status:** Research — documentation of current state
**Question:** How can Company OS and/or Team OS be used to work with BMAD artifacts for platform and components?

Paths are relative to the repo root unless noted; Go source lives under `company-os-starter/`.
BMAD evidence is from `/Users/javierbenavides/others/ai-agents/production-grade-agent-skills/`
(`llms-full.txt`, 4,229 lines) and its `docs/reference/2026-08-03-bmad-platform-component-phase-compatibility.md`.

---

## 0. Two findings that frame everything else

**0.1 There is no existing integration.** `grep -ril "bmad"` across this entire
repository returns zero hits. No template, skill, schema, gate, or doc references BMAD.
Nothing has been built, and nothing is half-built.

**0.2 The platform/component framing in the question is Company OS vocabulary, not
BMAD vocabulary.** The BMAD compatibility doc dated today states it directly
(`docs/reference/2026-08-03-bmad-platform-component-phase-compatibility.md:12-17`):

> The user's framing — "the balance between **Platform-level** and **Component-level**" — does
> **not** appear in the BMAD documentation that is present in this repo. […] No match for
> `platform-level` / `component-level` as a vocabulary pair. No match for `scale level`, no
> `Level 0`–`Level 4` numbered project-level system.

A grep for `multi-repo|monorepo|cross-project|multi-service|microservice` across those
4,229 lines yields no feature. **BMAD is single-project, single-root.** The platform/component
axis is precisely the axis Company OS models natively and BMAD does not.

---

## 1. What each system models

### 1.0 Shape at a glance

```text
COMPANY OS                                  BMAD
(federated, N repos, governance-first)      (single root, 1 project, execution-first)

  company-os/        ── baseline              _bmad/          ── config (TOML)
  company-ontology/  ── canonical IDs         docs/           ── kernel + knowledge
  platforms/<p>/     ── catalog + reality     _bmad-output/   ── generated
  teams/<t>/         ── ownership + product        planning-artifacts/
  knowledge/         ── synced foreign docs        implementation-artifacts/
                        (0444, read-only)          specs/spec-{slug}/

  ▲ models platform vs component natively    ▲ no platform/component vocabulary
  ▲ multi-repo via workspace.yaml            ▲ no multi-repo / monorepo feature
  ▲ stops at the PRD                         ▲ starts at the PRD, goes down to stories
```

### 1.1 Company OS artifact inventory

Authored peer roots: `company-os/`, `platforms/<p>/`, `teams/<t>/`, `company-ontology/`,
plus optional synced `knowledge/`.

Real layout — `examples/workspace/` verbatim, trimmed to one platform and one team:

```text
examples/workspace/
├── company-os/
│   ├── CLAUDE.md
│   ├── onboarding/developer.md
│   ├── skills/syncing-knowledge.SKILL.md         # note: <name>.SKILL.md, not <name>/SKILL.md
│   └── standards/company-baseline.yaml
├── company-ontology/
│   ├── CLAUDE.md                                 # generated index
│   ├── ids/registry.yaml                         # canonical IDs
│   ├── concepts/{capability--message-delivery,component}.md
│   ├── contexts/communications.md                # ubiquitousLanguage + forbiddenTerms
│   ├── context-maps/crm-to-communications.md
│   └── taxonomies/requirement-types.yaml
├── platforms/
│   └── communications/
│       ├── platform.yaml
│       ├── CLAUDE.md
│       ├── components/customer-notification-service.yaml   # ← single source of truth
│       ├── governance/requirements.yaml
│       ├── reality/components/                             # current-state truth
│       ├── change-records/active/                          # live PRDs
│       ├── archive/prds/                                   # completed + outcome.md
│       ├── generated/feature-index.yaml                    # derived — never hand-edit
│       ├── skills/creating-prd.SKILL.md
│       └── log.md
└── teams/
    └── customer-engagement/
        ├── team.yaml
        ├── CLAUDE.md
        ├── ownership/components.yaml
        ├── governance/{deviations,exceptions}.yaml
        ├── product/discovery/                     # team-private, draft→validated
        ├── standards/definition-of-{ready,done}.md
        ├── generated/effective-governance.yaml    # derived — never hand-edit
        ├── onboarding/developer.md
        └── scratchpad/personal-rules/             # git-ignored; loses to canonical-mandatory
```

Team OS standalone is a strict subset — `examples/standalone-team/` is the whole thing:

```text
examples/standalone-team/
└── teams/solo/
    ├── team.yaml
    ├── CLAUDE.md
    └── onboarding/developer.md
```

No `platforms/`, so `prd new` (which requires `--platform`, `cmd/company-os/args.go:165`) is
not exercisable here at all. Standalone Team OS is discovery + governance resolution only.

| Level | Artifact | Path |
|---|---|---|
| Company | baseline controls | `company-os/standards/company-baseline.yaml` |
| Platform | component descriptor | `platforms/<p>/components/<id>.yaml` |
| Platform | requirements | `platforms/<p>/governance/requirements.yaml` |
| Platform | current-state truth | `platforms/<p>/reality/components/<id>.md` |
| Platform | live PRD | `platforms/<p>/change-records/active/` |
| Platform | completed PRD + outcome | `platforms/<p>/archive/prds/`, `outcome.md`, `log.md` |
| Team | ownership | `teams/<t>/ownership/components.yaml` |
| Team | deviations / exceptions | `teams/<t>/governance/{deviations,exceptions}.yaml` |
| Team | discovery brief | `teams/<t>/product/discovery/<id>/brief.md` |
| Team | DoR / DoD | `teams/<t>/standards/definition-of-{ready,done}.md` |
| Team | derived governance | `teams/<t>/generated/effective-governance.yaml` |
| Ontology | canonical IDs | `company-ontology/ids/registry.yaml` |

The lifecycle type set is **closed** (`internal/product/contract.go:28-33`):

```go
var LifecycleTypes = map[string]bool{
	"discovery-brief": true, "prd": true, "adr": true, "outcome-review": true,
}
```

### 1.2 BMAD artifact inventory

| BMAD artifact | Path pattern | Scope | Schema-validated |
|---|---|---|---|
| `kernel.md` | `<project_knowledge>/kernel.md` (default `docs/kernel.md`) | Platform-wide | No |
| `project_knowledge/` | `<project_knowledge>/index.md` + `*.md` | Platform-wide | No (indexed by `context.py`) |
| `prd.md` | `_bmad-output/planning-artifacts/` (via v4→v6 migration) | Platform-wide | No (checklist critique only) |
| `ARCHITECTURE-SPINE.md` | root never stated | Platform-wide | No |
| `stories.yaml` | `<spec-folder>/stories.yaml` | Per-component | No |
| `stories/<id>-*.md` | `<spec-folder>/stories/<story-id>-<slug>.md` | Per-story | No (status enum only) |
| `spec-<slug>.md` | `{implementation_artifacts}/spec-<slug>.md` | Per-story/run | No (status enum only) |
| `DESIGN.md` / `EXPERIENCE.md` | root never stated | Platform-wide | No |

Layout as documented (roots are per-key configurable, so this is the default resolution):

```text
<project-root>/
├── _bmad/
│   └── config.toml                        # [modules.bmm] — each root key overridable
├── docs/                                  # {project_knowledge} default
│   ├── kernel.md                          # always-loaded global ruleset
│   ├── index.md                           # built by context.py
│   ├── ARCHITECTURE-SPINE.md              # root never stated in docs
│   ├── DESIGN.md  /  EXPERIENCE.md        # root never stated in docs
│   └── *.md
└── _bmad-output/
    ├── planning-artifacts/
    │   └── prd.md                         # platform-wide standing requirements
    ├── implementation-artifacts/
    │   └── spec-<slug>.md                 # per story/run execution trail
    └── specs/
        └── spec-<slug>/                   # {spec-folder}
            ├── SPEC.md                    # REFERENCED whole by a story run
            ├── stories.yaml               # the work breakdown
            └── stories/
                ├── <story-id>-<slug>.md   # COPIED into a story run
                └── ...
```

BMAD writes under **three** roots, not one: `_bmad/` (config), `docs/` (durable knowledge),
`_bmad-output/` (generated), the latter with `planning-artifacts/`,
`implementation-artifacts/`, `specs/spec-{slug}/`. Roots are configurable per-key in TOML
under `[modules.bmm]` (`llms-full.txt:1112-1121`).

**BMAD ships no JSON schemas.** Validation is script-level only: `sprint_plan.py validate`,
`context.py`, and a configurable PRD checklist.

---

## 2. The granularity gap

Company OS's validated chain is **discovery brief → PRD → component reality updated →
outcome review**, with ADR as a fourth standalone lifecycle type.

```text
                COMPANY OS                      │            BMAD
                                                │
  company baseline (standards/)                 │
        ↓                                       │
  platform requirements.yaml                    │   kernel.md
        ↓                                       │   project_knowledge/
  team deviations + exceptions                  │   prd.md  (platform-wide)
        ↓  `governance resolve`                 │   ARCHITECTURE-SPINE.md
  generated/effective-governance.yaml           │
        ↓                                       │
  discovery brief   (draft → validated)         │
        ↓  `prd new --from-discovery`           │
  PRD  (proposed → completed)                   │
        ↓  Gather()                             │
  ┌ - [ ] checklist items ─────────────┐        │
  │ derived · no id · no file          │        │
  │ no frontmatter · no status         │        │
  │ no schema · no validator           │        │
  │ consumed only by uncheckedItems()  │        │
  └────────────────────────────────────┘        │
                                                │
  ══════ VALIDATED MODELLING STOPS ══════       │   ══ BMAD'S WORK BEGINS ══
                                                │
        (nothing)                               │   stories.yaml
        (nothing)                               │   stories/<id>-<slug>.md
        (nothing)                               │   spec-<slug>.md
                                                │
        ↓  `prd complete` (done-check)          │
  reality/components/<id>.md updated            │        (no equivalent)
        ↓                                       │
  archive/prds/ + outcome.md (+90d) + log.md    │        (no equivalent)
```

The dividing line is the finding: the two artifact sets are **stacked, not overlapping**.

**Below the PRD, Company OS models nothing as a first-class artifact.** Evidence is
negative-space and consistent across four surfaces:

1. `LifecycleTypes` is closed — no task, story, ticket, epic, or work-item type
   (`internal/product/contract.go:28-33`).
2. `grep "type: " templates/*.md` returns exactly six: `skill`, `adr`, `discovery-brief`,
   `outcome-review`, `prd`, `component-reality`.
3. No scaffolding command creates one. There is no `task new` / `story new`.
4. The one sub-PRD unit that exists is **derived and unvalidated**: the checklist item.
   `ChecklistItem` (`internal/product/checklist.go:50`) is generated from effective
   governance by `Gather` (`checklist.go:67`), rendered as `- [ ]` by `ChecklistMarkdown`
   (`checklist.go:209`), and consumed only by counting — `uncheckedItems`
   (`internal/product/prd.go:491`) returns an integer and `prd complete` blocks if non-zero.
   It has no id, no file, no frontmatter, no status, no owner, no schema entry, no validator.

BMAD's `stories.yaml`, `stories/<id>-*.md`, and `spec-<slug>.md` all live **below** the
level at which Company OS stops modelling. The two systems' artifact sets are therefore
largely complementary rather than overlapping: Company OS's finest validated grain (the PRD)
sits directly above BMAD's coarsest execution grain (the story set).

---

## 3. The join points that actually exist

### 3.1 `knowledge/` — read-only catalog sync (the built-for-this mechanism)

`knowledge/` is described in the project CLAUDE.md as *the catalog: read-only documentation
slices synced from repos that are not Company OS workspaces (a component repo's `docs/sdd`,
`specs/`). Never authored by hand, never scaffolded by `init`.* BMAD's `docs/` and
`_bmad-output/` are exactly this shape.

A real manifest, `examples/banking/bank/workspaces/team-fraud-detection/workspace.yaml`
(abridged) — note that the team repo *is* the workspace root and everything else is a slice:

```yaml
version: 1
repos:
  - name: company-os
    url: https://git.example.com/bank/company-os.git
    localDirectory: company-os
    pin: {tag: v2026.2}
    paths: [standards/, onboarding/, skills/, templates/]
  - name: platform-fraud
    url: https://git.example.com/bank/platform-fraud.git
    localDirectory: platforms/fraud
    pin: {commit: 1111111111111111111111111111111111111111}
    paths: [governance/, components/, reality/, skills/, templates/]
  - name: platform-payments
    url: https://git.example.com/bank/platform-payments.git
    localDirectory: platforms/payments
    pin: {commit: 2222222222222222222222222222222222222222}
    paths: [governance/, components/, reality/]     # ← narrower slice: read-only consumer
```

The `knowledge/` destination in practice, from `examples/failing-federated/workspace.yaml`:

```yaml
  - name: never-synced
    url: https://git.example.com/acme/never-synced.git
    localDirectory: knowledge/never-synced          # ← foreign docs land here
    pin: {commit: 2222222222222222222222222222222222222222}
    paths: [docs/]
```

Authored vs. materialized, in the same tree:

```text
team-fraud-detection/                 (the team's own git repo — WRITABLE)
├── workspace.yaml                    ✎ authored
├── SYNC-NOTE.md                      ✎ authored
├── teams/fraud-detection/            ✎ authored — the only writable content
│   ├── team.yaml
│   ├── ownership/components.yaml
│   ├── governance/{deviations,exceptions}.yaml
│   ├── product/discovery/2026-alert-triage-queues/brief.md
│   ├── standards/definition-of-{ready,done}.md
│   └── scratchpad/personal-rules/priya-review-checklist.md
│
│   ······ everything below appears only after `workspace sync` ······
├── company-os/                       🔒 0555 / files 0444
├── company-ontology/                 🔒
├── platforms/fraud/                  🔒
├── platforms/payments/               🔒
├── knowledge/<bmad-repo>/            🔒 ← where BMAD docs/ would land
└── workspace.lock.yaml               ⚙ generated: per-file SHA-256 + resolved slice set
```

Flow and the two distinct gate-8 failures:

```text
  source repo (e.g. a BMAD component repo)
        │  paths: [docs/, _bmad-output/planning-artifacts/]
        │  pin: {tag|commit}
        ▼
  git sparse checkout (≥2.27) → cache
        │
        ▼  `company-os workspace sync`   [--frozen] [--only <repo>]
  knowledge/<repo>/…   chmod 0444 files / 0555 dirs
        │
        ├──► workspace.lock.yaml   { per-file SHA-256, resolved slice set }
        │
        ▼  `company-os validate` → gate [8/8]
   ┌────────────────────────────────┬──────────────────────────────────────┐
   │ file hash ≠ lock               │ manifest slice set ≠ lock slice set  │
   │ CodeSliceHandEdited            │ CodeSliceSetDrift                    │
   │ "someone edited a slice"       │ "paths: changed, no re-sync"         │
   │                                │ ← exists because old files still     │
   │                                │   hash clean; nothing else catches it│
   └────────────────────────────────┴──────────────────────────────────────┘
```

Gate coverage for anything living under `knowledge/`:

```text
  gate [1/8] ownership reconciliation ─┐
  gate [2/8] deviation/exception expiry │
  gate [3/8] …                          ├─ SKIPPED for knowledge/
  gate [4/8] …                          │  (foreign docs, no `type:` frontmatter, 0444)
  gate [5/8] …                          │
  gate [6/8] …                          │
  gate [7/8] skill conflicts ───────────┘
  gate [8/8] slice hash integrity ────── APPLIES

  `graph build`  ──────────────────────── SKIPPED (node root, not a graph-docs root)
  CLAUDE.md index ─────────────────────── APPLIES

                        ⇒ "Indexed, not governed."
```

Mechanics, from `internal/federation/`:

- A `workspace.yaml` manifest names each repo with a `localDirectory:` destination and an
  include-only `paths:` allowlist; a repo contributing several areas uses a `slices:` list of
  `{paths, localDirectory}` — one clone, one cache, N destinations. Targets must be disjoint.
- `workspace sync` materializes each slice at `0444` (files) / `0555` (dirs) via git sparse
  checkout, requires git ≥ 2.27, and writes `workspace.lock.yaml` with a per-file SHA-256 hash
  map plus the resolved slice set. Flags: `--frozen`, `--only`.
- Gate 8 (`internal/federation/lock.go:23`) fails on a hand-edit (`CodeSliceHandEdited`) **and**
  on a slice-set change made without a re-sync (`CodeSliceSetDrift`) — the second check exists
  because the old files still hash clean, so nothing else would catch it.

**Critically, `knowledge/` is a node root but not a graph-docs root.** `internal/graph/node.go:135-137`
states it in source: *"knowledge/ is included and is the reason this list is not graphRoots().
Indexed, not governed."* So synced BMAD documents:

- **are** indexed into the generated `CLAUDE.md` context node (`node.go:144`),
- **are** covered by gate 8 hash integrity,
- **are not** tag-derived, not frontmatter-validated, and never written to
  (`tags.go:191-194`: *"the slices under it are foreign, read-only, and carry no `type:`
  frontmatter, so deriving tags for them would mean writing to a 0444 tree"*).

This answers the "foreign markdown with no Company OS frontmatter" case directly: it is
carried, hashed, and indexed — and exempt from gates 1–7.

### 3.2 `@spec` markers — the only designed binding to code

The intent is stated at
`examples/banking/bank/repos/code-transaction-screening/tests/test_settlement_finality.py:3-4`:
*"Code repos are NOT modeled by the Company OS CLI. The only binding back to governance is
grep-able @spec markers tying tests to EARS clauses."*

Syntax, from the only two real instances (`:8`, `:14`):

```python
# @spec req://payments/settlement-finality@1.0#R2
```

**No Go code parses `@spec`.** `grep -rn '@spec' --include='*.go' .` → zero hits across all
138 files. `spec trace` and `validate --ontology` do not exist in the 17-command table
(`cmd/company-os/args.go:103-287`). Further, the traceability chain is **broken at the data
layer too**: `grep -rln "clauses:" examples/` → zero files; across all 9 `requirements.yaml`,
`requirement:` is a single prose string, never enumerated into R1..Rn. A mandatory clause
with zero `@spec` sites blocks nothing; its only force is a human checklist line in a
Definition of Done (`.../fraud-detection/standards/definition-of-done.md:9`).

Two incompatible numbering conventions coexist and are not interconvertible: dotted `R-1.1`
in `docs/ears/`, flat `#R1` in `@spec`.

### 3.3 Template override chain — three names only

`internal/scaffold/template.go:153-202`, first found wins:

```text
  resolve("prd")
        │
        ├─1─► teams/<team>/templates/prd.md              ✔ found → stop
        ├─2─► platforms/<platform>/templates/prd.md
        ├─3─► company-os/templates/prd.md
        └─4─► built-in Go constant / embedded templates/

  overridable names — exactly three (template.go:98-102):
      discovery-brief   ✔
      prd               ✔
      reality-component ✔
      outcome-review    ✘ hardcoded at internal/product/prd.go:568-571
      team scaffolds    ✘ internal/scaffold/teamtemplates.go:3-9
      any YAML scaffold ✘

  regardless of which layer won:
      required headings missing → BLOCKS     (contract.go:118-125)
      empty sections            → opt-in per team via standards/doc-formats.yaml
                                  with `enforce: true` (contract.go:99-116)
```

On disk, an override that shadows the built-in PRD template for one team only:

```text
workspace/
├── company-os/templates/prd.md             # layer 3 — company default
├── platforms/communications/templates/
│   └── prd.md                              # layer 2 — platform variant
└── teams/customer-engagement/
    ├── templates/prd.md                    # layer 1 — WINS for this team
    └── standards/doc-formats.yaml          # `enforce: true` → empty-section checks on
```

The overridable name set is exactly three (`template.go:98-102`):
`discovery-brief`, `prd`, `reality-component`. `outcome-review` is **not** overridable
(hardcoded at `internal/product/prd.go:568-571`); nor are the team markdown scaffolds
(`internal/scaffold/teamtemplates.go:3-9`) or any YAML scaffold.

Missing required headings always block regardless of which template produced the document
(`internal/product/contract.go:118-125`, explicitly noting custom override templates);
empty-section checks are opt-in per team via `teams/<t>/standards/doc-formats.yaml` with
`enforce: true` (`contract.go:99-116`).

### 3.4 Four-layer skills — where BMAD process guidance would live

`internal/skills/skills.go:29-38` and `:197-225`:

| Layer | Directory | Glob |
|---|---|---|
| company | `company-os/skills/` | `*.SKILL.md` |
| platform | `platforms/<p>/skills/` | `*.SKILL.md` |
| team | `teams/<t>/skills/` | `*.SKILL.md` |
| personal | `teams/<t>/scratchpad/personal-rules/` | `*.md` |

Discovery is **one level deep and non-recursive**, so `skills/<name>/SKILL.md` is invisible —
the file must be `skills/<name>.SKILL.md` (`Suffix`, `skills.go:22`).

```text
workspace/
├── company-os/skills/
│   └── syncing-knowledge.SKILL.md          ✔ discovered   (authority: canonical)
│   └── syncing-knowledge/SKILL.md          ✘ INVISIBLE — non-recursive, wrong suffix
├── platforms/communications/skills/
│   └── creating-prd.SKILL.md               ✔ target of `extends:`
└── teams/customer-engagement/
    ├── skills/
    │   └── creating-prd.SKILL.md           ⚠ gate 7 fails if it shadows a canonical skill
    │                                          extends: platform-skill://communications/creating-prd
    └── scratchpad/personal-rules/
        └── priya-review-checklist.md       ✔ personal layer (plain *.md, git-ignored)
```

`extends:` grammar is `^platform-skill://<platform>/<name>$` and resolves to the **platform
layer only**. The declared precedence —

```text
canonical-mandatory  >  personal  >  canonical-default  >  canonical-guidance
└──────────────────────────┬──────────────────────────────────────────────┘
        written into team.yaml as DATA ONLY (scaffold.go:234-240);
        grep finds no Go code that reads it. Enforced by prose, for the agent.
```

Gate 7 (`internal/skills/conflicts.go:182-207`) enforces two rules: a team/personal skill may
not shadow a company/platform skill marked `authority: canonical`, and `extends:` must resolve
(`skills.go:394-433`).

Step tier tags `(mandatory)`/`(default)`/`(guidance)` are **parsed but not enforced**
(`pysem.go:79-81`; sole consumers are display sites `list.go:93`, `:109`).

### 3.5 The CLI as a machine interface

`company-os-starter/docs/user-guide/reference/company-os-cli.md` publishes the contract in
normative terms:

> Every command exits with one of eight codes. This is a contract […] **Scripts and agents
> should branch on the code and never parse stdout.**

The `--json` payload carries `schemaVersion: 1` ("bumped only on a removal or repurpose"),
a stable `slug`, and `exitCode`, and is documented as *"the same information the text renderer
prints, encoded rather than reformatted, so the two can never disagree."* Exit `1` means an
artifact is wrong; exit `5` means the artifact is fine and the work is not done yet.

Correspondingly, `docs/user-guide/explanation/github-mcp-and-automation.md:5`:
**"Company OS ships no MCP server and no MCP client."** The CLI subprocess plus `--json` is
the intended and only integration path.

---

## 4. What does not exist

**There is no plugin or hook mechanism for adding a validated artifact type.** Definitively —
no exec-out to user scripts, no Go-plugin loading, no config-declared validators anywhere
under `internal/` or `cmd/`. Supporting evidence:

- Gates are a hardcoded Go slice — `internal/validate/validate.go:82-92`, a literal
  `[]func(int) (model.GateResult, error)` naming eight functions from six packages.
  Ordinals are positional (`i+1` at `:100`); gate 8 runs only if a federation manifest exists.
- `type:` → tag mapping is a hardcoded 9-entry map (`internal/graph/tags.go:23-29`).
- Lifecycle rules are literal conditionals (`internal/product/contract.go:65-89`).
- `add` is a closed switch over `platform|team|component` (`internal/scaffold/commands.go:368-424`).
- `company-os-starter/schemas/` contains only `SCHEMAS.md` — no machine-readable schemas.

A genuinely new artifact type requires editing at minimum `internal/graph/tags.go:23`,
`internal/product/contract.go:65`, `internal/validate/validate.go:82`,
`internal/scaffold/commands.go:368`, and `cmd/company-os/args.go:116`.

Also absent: `company-os skill new`, `template init`, template lint, `spec trace`,
`validate --ontology`, vocabulary/`forbiddenTerms` linting, wikilink generation, registry
validation, and any `graph.json`.

---

## 5. Level-by-level correspondence

| Concern | BMAD | Company OS | Relationship |
|---|---|---|---|
| Global rules always loaded | `kernel.md` | `company-os/standards/` + `teams/<t>/generated/effective-governance.yaml` | Both exist. Company OS's is **derived and tiered**; BMAD's is authored prose. |
| Durable project facts | `project_knowledge/` | `company-ontology/` + `reality/` | Both exist. Company OS splits meaning (IDs) from current-state truth. |
| Global requirements | `prd.md` (platform-wide) | `platforms/<p>/governance/requirements.yaml` | Naming collision: BMAD's `prd.md` is platform-scoped; a Company OS PRD is a **change record**, not a standing requirement. |
| Architecture decisions | `ARCHITECTURE-SPINE.md` | `templates/adr.md` (`type: adr`) | ADR is a lifecycle type but has **no scaffold subcommand** — `grep '"adr"' cmd/ internal/scaffold/` returns nothing outside the template. |
| Change proposal | — | `change-records/active/` → `archive/prds/` + `outcome.md` | Company OS only. BMAD has no archive/outcome-review concept. |
| Work breakdown | `stories.yaml`, `stories/*.md` | derived `- [ ]` checklist items only | BMAD only, as first-class files. |
| Execution trail | `spec-<slug>.md` | — | BMAD only. |
| UX spec | `DESIGN.md` / `EXPERIENCE.md` | — | BMAD only. |
| Readiness gate | `bmad-sprint-planning` (PASS/CONCERNS/FAIL) | `check ready` | Both exist; different inputs. Company OS derives from governance and requires `governance resolve` first (`CodeCheckUnresolved`). |
| Done gate | — | `prd complete` done-check | Company OS only, and it is the strictest thing in either system. |
| Ownership / accountability | — | descriptor `ownership.accountableTeam`, gate 1 | Company OS only. |
| Multi-repo | none | `workspace sync` + lock + gate 8 | Company OS only. |

### 5.1 The two gates worth naming precisely

**Company OS `prd complete`** (`internal/product/prd.go:296`, gated by `doneCheck` at `:397`)
blocks on: any unchecked governance checklist item (`CodeDoneChecklistUnchecked`), or a
component whose `reality/components/<id>.md` is missing (`CodeDoneRealityMissing`), undated
(`CodeDoneRealityDateInvalid`), or whose `updated:` predates the PRD's `created:`
(`CodeDoneRealityStale`). On success it archives the PRD, writes `outcome.md` due
`outcomeDays = 90` days out (`prd.go:27`), and appends `log.md`. `--force` overrides.

**BMAD `bmad-sprint-planning`** (`llms-full.txt:3117-3123`) inventories whatever planning
artifacts exist, *"identifying documents by reading them, not by filename patterns"*, and asks
one question: *"could a developer implement these epics without inventing decisions nothing
records?"* A missing document type is only a finding if stories depend on it.

The first is a **closing** gate keyed on reality drift; the second is an **opening** gate keyed
on specification completeness. They do not overlap:

```text
   BMAD `bmad-sprint-planning`            COMPANY OS `prd complete`
   ── OPENING gate ──                     ── CLOSING gate ──

   asks:                                  asks:
   "could a developer implement           "has reality been updated to
    these epics without inventing          match what we said we'd do?"
    decisions nothing records?"

   input: whatever planning artifacts     input: effective-governance checklist
   exist, identified by READING them,            + reality/components/<id>.md dates
   not by filename patterns

   verdict: PASS / CONCERNS / FAIL        blocks on:
                                            CodeDoneChecklistUnchecked  (any `- [ ]`)
   a missing doc type is a finding          CodeDoneRealityMissing
   ONLY if stories depend on it             CodeDoneRealityDateInvalid
                                            CodeDoneRealityStale (updated: < created:)

                                          on success:
                                            → archive/prds/
                                            → outcome.md due +90d (prd.go:27)
                                            → append log.md
                                          `--force` overrides
```

### 5.2 Story sharding vs. governance derivation

BMAD's sharding COPIES into a story run: the `stories.yaml` entry's `title` and `description`,
plus each sibling `stories/*.md` file's Code Map, Design Notes, Spec Change Log, Tasks &
Acceptance checklist state, and Auto Run Result. It REFERENCES (loads whole from disk):
`<spec-folder>/SPEC.md` and companions (`llms-full.txt:3386`, `:3392`, `:3398`).

Company OS's analogue is `Gather` (`internal/product/checklist.go:67`), which reads
`teams/<t>/generated/effective-governance.yaml` and renders per-component `- [ ]` lines. Both
are context-narrowing derivations from a broader source; Company OS's output is a checklist
inside one document, BMAD's is a set of standalone files.

---

## 6. Observed constraints relevant to any integration

1. **Component identity is flat and global.** `component://customer-notification-service`
   carries no platform segment, whereas `capability://` and `req://` use
   `<platform>/<local>` (`registry.yaml:4-10`).
2. **Nothing validates the ID registry.** `internal/ids/ids.go:121` `List` emits only `SevOK`
   findings — *structurally incapable* of reporting a problem. The shipped reference workspace
   already carries three unregistered live IDs (`req://communications/prd-structure`,
   `concept://component`, `map://crm--communications`), and the generated
   `company-ontology/CLAUDE.md` index is a strict superset of the registry beside it.
3. **Tags are plain strings, not `#`-prefixed, and are not slugified.** The `#kind/prd` form in
   the project CLAUDE.md is Obsidian display convention; stored data is `kind/prd`
   (`tags.go:123`, `:87`, `:95`, `:104`). `status: In Review` yields the tag
   `status/In Review`, space included. There is **no `tier/` rule** — grep of `internal/graph`
   for `tier` returns zero non-test hits.
4. **Four facets survive `graph build`** — `ontology`, `capability`, `req`, `spec`
   (`tags.go:34-36`). This four-key map is the entire footprint of the `@spec`/`req` concept
   in compiled Go: it preserves a human-typed `spec/…` prefix; it does not create, parse,
   resolve, or verify one.
5. **`prd` requires `--platform`** (`cmd/company-os/args.go:165`), so on
   `examples/standalone-team` — which ships only `team.yaml`, `CLAUDE.md`, and
   `onboarding/developer.md` — the PRD lifecycle is not exercisable at all. Team OS standalone
   is discovery plus governance resolution only.
6. **Latent trap:** the `scratchpad` skip test is on the **absolute** path in `IterGraphDocs`
   (`tags.go:215-221`) but the **workspace-relative** path in `iterKnowledgeDocs`
   (`node.go:161-163`). A checkout living under a directory named `scratchpad` silently skips
   every graph document.
7. **Documentation asserts behaviour the Go tree does not implement** in at least three places:
   `docs/ONTOLOGY-GUIDE.md:316` (`spec trace`), `contexts/communications.md:20`
   (`validate --ontology`), and `docs/user-guide/explanation/observer-roadmap.md:48`
   (*"**Started.** `@spec` marker and wikilink extraction…"*).

---

## 7. Summary of the answer

- Company OS models the **platform/component axis natively** (descriptors, requirements,
  ownership, reality, multi-repo federation). BMAD does not model it at all — the framing
  in the question is absent from BMAD's own documentation.
- Company OS's validated artifact chain **stops at the PRD**. BMAD's `stories.yaml`,
  `stories/*.md`, and `spec-*.md` all live below that line. The overlap is narrow; the
  collision is nominal (`prd.md`) rather than structural.
- The mechanism purpose-built for foreign artifacts is `knowledge/` + `workspace sync`:
  BMAD output lands read-only at `0444`, hash-gated by gate 8, indexed into `CLAUDE.md`,
  and exempt from gates 1–7 — *indexed, not governed*.
- The mechanism designed to bind governance to code is the `@spec` marker, which **no Go code
  parses**, and for which **no clause data exists** in any shipped `requirements.yaml`.
- Extension without editing Go is limited to: three overridable template names, four skill
  layers, a per-team format-enforcement opt-in, and four preserved tag facets. There is **no
  plugin or hook surface**, and no MCP layer is planned.
- The supported programmatic surface is the CLI subprocess: eight documented exit codes and a
  `schemaVersion: 1` JSON payload, with an explicit directive never to parse stdout.

---

## Appendix A. Mechanical result of pointing the existing sync at a BMAD repo

**This is not a recommendation.** It is what the documented `workspace sync` mechanism would
mechanically produce, given a manifest entry whose `paths:` names BMAD output directories.
No such manifest exists in this repository today.

Source side — a BMAD-run component repo:

```text
code-transaction-screening/            (a normal code repo, NOT a Company OS workspace)
├── src/
├── tests/
│   └── test_settlement_finality.py    # carries: # @spec req://payments/settlement-finality@1.0#R2
├── _bmad/config.toml
├── docs/
│   ├── kernel.md
│   └── ARCHITECTURE-SPINE.md
└── _bmad-output/
    ├── planning-artifacts/prd.md
    └── specs/spec-screening-v2/
        ├── SPEC.md
        ├── stories.yaml
        └── stories/002-throttle-rules.md
```

Manifest entry (allowlist is include-only; targets must be disjoint across the manifest):

```yaml
  - name: code-transaction-screening
    url: https://git.example.com/bank/code-transaction-screening.git
    localDirectory: knowledge/transaction-screening
    pin: {commit: <sha>}
    paths:
      - docs/
      - _bmad-output/planning-artifacts/
      - _bmad-output/specs/
```

Destination side, after `company-os workspace sync`:

```text
team-payments-rails/
├── workspace.yaml                                    ✎
├── workspace.lock.yaml                               ⚙ hashes every file below
├── teams/payments-rails/                             ✎ writable
├── platforms/payments/                               🔒 governed slice — gates 1–8 apply
│   ├── components/transaction-screening.yaml         ← source of truth for ownership
│   ├── governance/requirements.yaml
│   └── reality/components/transaction-screening.md   ← `prd complete` reads this date
└── knowledge/transaction-screening/                  🔒 indexed, NOT governed
    ├── CLAUDE.md                                     ⚙ generated index (node.go:144)
    ├── docs/kernel.md                                0444 · no type: · no tags derived
    ├── docs/ARCHITECTURE-SPINE.md                    0444
    ├── _bmad-output/planning-artifacts/prd.md        0444 · NOT a Company OS `type: prd`
    └── _bmad-output/specs/spec-screening-v2/
        ├── SPEC.md                                   0444
        ├── stories.yaml                              0444
        └── stories/002-throttle-rules.md             0444
```

Consequences that follow from the source, not from opinion:

| Behaviour | Result | Why |
|---|---|---|
| Hand-edit `knowledge/.../stories.yaml` | gate 8 fails, `CodeSliceHandEdited` | files are `0444`, hash-pinned in the lock |
| Change `paths:` without re-syncing | gate 8 fails, `CodeSliceSetDrift` | old files still hash clean — nothing else catches it |
| `graph build` over that tree | skipped | `knowledge/` is a node root, not a graph-docs root (`node.go:135-137`) |
| BMAD `prd.md` present | ignored by `prd validate` | it is not `type: prd` frontmatter and not under `change-records/active/` |
| `#R2` marker in the synced test | resolves to nothing | no Go parses `@spec`; no `clauses:` exist in any `requirements.yaml` |
| Update BMAD docs upstream | requires pin bump + re-sync | *"Never edit a slice — change the source repo, bump the pin, re-sync."* |

The `prd complete` done-check remains keyed on
`platforms/<p>/reality/components/<id>.md` having an `updated:` date newer than the PRD's
`created:` date. Nothing under `knowledge/` can satisfy it, because the reality file lives in
the governed platform slice and is a Company OS artifact with `type: component-reality`.
