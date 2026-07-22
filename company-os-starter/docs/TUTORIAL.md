---
type: doc
tags: [doc/company-os-starter, kind/tutorial]
---

# Tutorial: From Idea to Archived PRD with `company-os`

This walkthrough uses the populated example in `examples/workspace/` and the
reference CLI in `bin/company-os`. Every command and output below was executed
against this kit — you can reproduce the whole session.

**The principle behind everything:** strict on artifacts, flexible on process.
The CLI and skills guide you; the validators enforce only the contract. How you
think, draft, or prompt your agent is your business.

New here? Start with [GOLDEN-PATH.md](GOLDEN-PATH.md) — it takes an empty
directory to a completed change on a fresh workspace you scaffold yourself.

## 0. Setup

```bash
# requirements: python3 + pyyaml  (pip install pyyaml)
export PATH="$PWD/bin:$PATH"
cd examples/workspace        # or: export COMPANY_OS_WORKSPACE_ROOT=$PWD/examples/workspace
```

The workspace layout (each directory is its own Git repo in real life):

```text
examples/workspace/
├── company-os/standards/company-baseline.yaml      # 3 company controls
├── platforms/communications/
│   ├── governance/requirements.yaml                # tiered, versioned rules
│   ├── components/customer-notification-service.yaml   # SINGLE SOURCE for
│   │                                               #   platform links + owner team
│   ├── reality/components/...                      # current behavior docs
│   ├── change-records/active/                      # live PRDs
│   ├── archive/prds/                               # completed PRDs + outcomes
│   └── skills/                                     # canonical platform skills
└── teams/customer-engagement/
    ├── ownership/components.yaml                   # what the team owns
    ├── governance/deviations.yaml                  # comply-or-explain records
    ├── standards/definition-of-{ready,done}.md     # team baseline
    ├── generated/effective-governance.yaml         # DERIVED — never hand-edit
    └── scratchpad/                                 # local-only, git-ignored
```

## 0.5 Configuring paths on your machine

Your absolute paths are yours; the committed YAML must stay portable. The rule
is simple: **no `/Users/yourname/...` string is ever committed.** Absolute paths
live only in environment variables or git-ignored local files. That way you,
your teammates, and CI can clone the same repos to completely different
locations and everything still resolves.

### Precedence (highest wins)

```text
1. CLI flag              company-os --root /abs/path ...
2. Environment variable  $COMPANY_OS_WORKSPACE_ROOT
3. Repo-local override   .company-os.local.yaml        (git-ignored)
4. User-level config     ~/.company-os/config.yaml     (outside every repo)
5. Committed shared      config/repositories.yaml      (relative dirs only)
6. Built-in default      current working directory
```

### The normal case: one env var per machine

Set the workspace root once. The CLI's `--root` already defaults to
`$COMPANY_OS_WORKSPACE_ROOT`, so after this you can run commands from anywhere:

```bash
# you
export COMPANY_OS_WORKSPACE_ROOT=/Users/javier/work/company-knowledge
# a teammate, different path, same commands
export COMPANY_OS_WORKSPACE_ROOT=/home/alice/projects/company
# CI
export COMPANY_OS_WORKSPACE_ROOT=/workspace/company

company-os governance resolve --team customer-engagement   # resolves under your root
```

Or pass it inline without exporting anything:

```bash
company-os --root /Users/javier/work/company-knowledge validate
```

Effective path = `root` + the repo's **relative** `directory`. The committed
`config/repositories.yaml` holds only those relative directories and the *name*
of the env var — never a value:

```yaml
# config/repositories.yaml   (committed, identical for everyone)
workspace:
  rootVariable: COMPANY_OS_WORKSPACE_ROOT   # the NAME of the env var, not a path
repositories:
  - id: platform-os-communications
    directory: platforms/communications                       # relative
  - id: customer-notification-service
    directory: components/customer-notification-service       # relative
```

### The odd case: one repo lives somewhere off the tree

If you cloned a repo into an unusual spot, override just that repo in a
**git-ignored** local file — absolute paths are allowed here because this file
is never committed:

```yaml
# .company-os.local.yaml   (NOT committed)
workspace:
  root: /Users/javier/work/company-os
repositories:
  customer-notification-service:
    localPath: /Users/javier/work/notification-experiments    # absolute override
```

Commit a `.company-os.local.example.yaml` so teammates know the shape, and make
sure the real file is ignored. `company-os scratchpad init` writes these ignore
rules for you:

```gitignore
.company-os.local.yaml
.env
.env.local
scratchpad/
```

### Multiple workspaces: user-level config

Juggling a primary and an experimental checkout? Keep them in a user-level file
outside every repo and switch by name:

```yaml
# ~/.company-os/config.yaml
activeWorkspace: primary
workspaces:
  primary:
    root: /Users/javier/work/company-knowledge
  experimental:
    root: /Users/javier/work/company-experimental
```

### What goes where

| Committed YAML (portable) | Env vars / git-ignored local files (machine-specific) |
|---|---|
| Relative `directory` per repo | Absolute `root` path |
| Git remote URLs, stable IDs | Per-repo `localPath` overrides |
| The *name* of the env var | Obsidian vault location, workspace selection |

### Kit status

The reference `bin/company-os` in this kit implements layers **1, 2, and 6**
today: `--root`, `$COMPANY_OS_WORKSPACE_ROOT`, and cwd fallback — enough to
point it at any path on your machine. Layers 3–5 (`.company-os.local.yaml`
merging, `~/.company-os/config.yaml`, and a `workspace sync` that clones repos
into the relative directories) are specified here and in the proposal but not
yet wired into the CLI.

## 1. Resolve the team's effective governance

Ownership → component descriptors → platform relationships → requirements,
merged with the team's approved deviations:

```bash
$ company-os governance resolve --team customer-engagement
resolved governance for team 'customer-engagement' (1 component(s))
wrote teams/customer-engagement/generated/effective-governance.yaml
  customer-notification-service: platforms [communications], 3 company + 3 platform requirement(s)
```

Ask *why* a rule applies to you:

```bash
$ company-os governance explain customer-notification-service
component 'customer-notification-service' (team customer-engagement):
  - delivery-reliability v2.1 (mandatory)
      applies because the component 'belongs-to' platform 'communications'
  - message-schema v1.3 (mandatory)
      applies because the component 'belongs-to' platform 'communications'
  - prd-structure v1.0 (default) [deviation applied]
      applies because the component 'belongs-to' platform 'communications'
```

Note the last line: the team's lean-PRD deviation (a `default`-tier rule) is
already merged in. Mandatory rules can never be deviated — only excepted (§7).

## 2. Discovery

```bash
$ company-os discover new "Per-channel quiet hours" --team customer-engagement
created teams/customer-engagement/product/discovery/2026-per-channel-quiet-hours/brief.md
next: fill Problem signal, Hypothesis, Success criteria, then run: ...
```

Try validating the empty brief — the contract pushes back:

```bash
$ company-os discover validate 2026-per-channel-quiet-hours --team customer-engagement
  [FAIL] section 'Problem signal' is empty
  [FAIL] section 'Hypothesis' is empty
  [FAIL] section 'Success criteria' is empty
```

Fill the three mandatory sections (how you research them — interviews, data,
prototypes — is `guidance`-tier, i.e. your choice), then:

```bash
$ company-os discover validate 2026-per-channel-quiet-hours --team customer-engagement
  [ok] brief '2026-per-channel-quiet-hours' validated (status: validated)
```

## 3. Create the PRD from the validated discovery

```bash
$ company-os prd new --team customer-engagement --platform communications \
    --components customer-notification-service \
    --from-discovery 2026-per-channel-quiet-hours
created platforms/communications/change-records/active/2026-per-channel-quiet-hours/prd.md
```

Three things happened automatically:

1. The **Problem statement** and **Success metrics** were copied from the
   discovery brief (no re-typing, no drift).
2. A **governance snapshot** was stamped (`governanceSnapshot: 2026-07-18`) —
   if the platform tightens a rule next month, this PRD is still evaluated
   against the version it started with.
3. The **applicable governance checklist** was injected, deviation included:

```markdown
## Applicable governance (snapshot 2026-07-18)

**customer-notification-service**
- [ ] company: security-service-baseline v3.0 (mandatory) — evidence:
- [ ] company: customer-data-privacy v2.2 (mandatory) — evidence:
- [ ] company: tier-1-observability v1.4 (default) — evidence:
- [ ] communications: delivery-reliability v2.1 (mandatory) — evidence:
- [ ] communications: message-schema v1.3 (mandatory) — evidence:
- [ ] communications: prd-structure v1.0 (default) *(team deviation applies)* — evidence:
```

Validation enforces the artifact contract, nothing more:

```bash
$ company-os prd validate 2026-per-channel-quiet-hours --platform communications
  [FAIL] frontmatter field 'decisionOwner' missing or TODO
  [FAIL] section 'Proposed change' is empty

# ...fill decisionOwner and Proposed change...

$ company-os prd validate 2026-per-channel-quiet-hours --platform communications
  [ok] PRD '2026-per-channel-quiet-hours' passes the artifact contract
```

## 4. Composable Definition of Ready during refinement

Team baseline + resolved governance, composed on demand — no giant static
checklist:

```bash
$ company-os check ready --team customer-engagement \
    --components customer-notification-service
== Team baseline (definition-of-ready.md) ==
## Team Definition of Ready
- The problem and expected outcome are clear.
...
== Applicable governance (customer-notification-service) ==
- [ ] communications: delivery-reliability v2.1 (mandatory) — evidence:
...
```

`check done` works the same way against `definition-of-done.md`.

## 5. Complete the change — reality first, archive second

Try to complete before updating the reality doc:

```bash
$ company-os prd complete 2026-per-channel-quiet-hours --platform communications
done-check failed — a change is not done until reality is updated:
  [FAIL] reality doc for 'customer-notification-service' not updated since PRD created (reality updated: 2026-01-10)
```

This is the "balance vs. transactions" rule with teeth. Update
`reality/components/customer-notification-service.md` (describe the new quiet
hours behavior, bump `updated:`), check off the governance items with evidence
links, then:

```bash
$ company-os prd complete 2026-per-channel-quiet-hours --platform communications
archived -> platforms/communications/archive/prds/2026-per-channel-quiet-hours
outcome review scheduled (due 2026-10-16)
appended platforms/communications/log.md
```

The PRD is now history; reality is current; an `outcome.md` with the success
metrics is waiting for actuals in 90 days.

## 6. Role views

```bash
$ company-os today --role product-owner
== today (product-owner) ==
platform communications: 0 active PRD(s)
  - outcome review due 2026-10-16: 2026-per-channel-quiet-hours

$ company-os today --role developer
team customer-engagement (governance generated 2026-07-18T...)
  - customer-notification-service: 3 platform requirement(s), 3 company control(s)
```

## 7. Flexibility with an audit trail: deviations and exceptions

Deviate from a `default` rule (comply-or-explain, auto-scheduled re-review):

```bash
$ company-os deviation declare "company-standard://estimation/story-points" \
    --team customer-engagement \
    --rationale "Team forecasts with cycle time instead of points."
declared deviation from company-standard://estimation/story-points in teams/customer-engagement/governance/deviations.yaml
review due 2027-01-14; re-run: company-os governance resolve --team customer-engagement
```

Except a `mandatory` rule (expiring, owned, needs the rule owner's approval):

```bash
$ company-os exception request "platform-standard://communications/message-schema" \
    --team customer-engagement --component legacy-fax-gateway \
    --expires 2026-12-31 \
    --reason "Legacy protocol cannot carry the standard envelope."
exception drafted in teams/customer-engagement/governance/exceptions.yaml (expires 2026-12-31)
note: mandatory rules require approval by the rule owner before this is valid.
```

## 8. The CI gate

```bash
$ company-os validate
[1/7] ownership reconciliation
  [ok] customer-notification-service: registry and descriptor agree (communications)
[2/7] deviation and exception expiry
  [ok] customer-engagement: deviation platform-standard://communications/prd-structure current (review 2027-01-15)
  [ok] customer-engagement: deviation company-standard://estimation/story-points current (review 2027-01-14)
  [ok] customer-engagement: exception platform-standard://communications/message-schema valid until 2026-12-31
[3/7] active PRD contracts
[4/7] frontmatter core and tag derivation (interop contract)
  [ok] platforms/communications/reality/components/customer-notification-service.md: frontmatter core and tags in sync
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
  [ok] company-os/CLAUDE.md: context node in sync
  [ok] platforms/communications/CLAUDE.md: context node in sync
  [ok] teams/customer-engagement/CLAUDE.md: context node in sync
  [ok] company-ontology/CLAUDE.md: context node in sync
[6/7] feature-index drift (derived component->artifact map)
  [ok] communications: feature-index in sync (1 component(s))
[7/7] custom skills layering (shadowing + extends resolution)
  [ok] skills layered cleanly (1 canonical, 0 team, 1 personal; no shadowing or dangling extends)
PASS
```

The gate count is dynamic: the seven gates above run in monorepo mode. In a
**federated** workspace (a `workspace.yaml` manifest is present) validate adds an
eighth gate — `[8/8] federated slice integrity` — which fails if a materialized
governance slice was hand-edited (its content hash no longer matches
`workspace.lock.yaml`). With no manifest the eighth gate does not exist and the
output is byte-for-byte the seven-gate form above.

It fails (exit 1, blocking merge) when: a team claims ownership the component
descriptor doesn't confirm (single-source rule), a deviation passes its
`reviewDate`, an exception is missing or past its `expires`, an active PRD
is missing contract fields, a doc's frontmatter core or derived tags drift, a
generated `CLAUDE.md` context node is stale, or a platform's derived
`feature-index.yaml` is out of date. The last three gates are absence-tolerant
— they pass when the artifact is absent. Wire it as CI:

```yaml
# .github/workflows/os-validate.yml (any OS repo)
- run: pip install pyyaml
- run: bin/company-os --root . validate
- run: bin/company-os governance resolve --team <team> && git diff --exit-code teams/*/generated/
```

The second check ensures `effective-governance.yaml` is truly derived — if
someone hand-edited it, the regenerated file differs and CI fails.

## 9. Where personal flexibility lives

```bash
$ company-os scratchpad init --repo teams/customer-engagement
```

See `scratchpad/personal-rules/maria-prd-style.md` in the example: Maria's
agent applies her drafting style *on top of* `skill://product/creating-prd`.
Per `team.yaml` precedence, `canonical-mandatory > personal > canonical-default
> canonical-guidance` — her rules can reshape how the PRD gets written, never
whether it passes `prd validate`.

## 9.5 Tags everywhere: `graph build`

Every artifact in the kit now carries a `tags:` block — skills, templates,
docs, standards, YAML configs, and workspace documents. Two kinds exist:

- **Static facets** on canonical content (skills: `kind/skill authority/canonical
  process/prd`; configs: `kind/requirements platform/communications`).
- **Derived facets** on workspace docs, generated from frontmatter IDs by:

```bash
$ company-os graph build
  tagged platforms/communications/archive/prds/2026-per-channel-quiet-hours/prd.md
  ...
graph build: 11 doc(s) scanned, 9 updated
$ company-os graph build
graph build: 11 doc(s) scanned, 0 updated     # idempotent
```

The archived PRD, for instance, ends up with
`[component/customer-notification-service, discovery/2026-..., kind/prd,
platform/communications, status/completed, team/customer-engagement]` — all
derived from fields it already had. Scaffolds (`discover new`, `prd new`,
`prd complete`) emit starter tags at creation; `graph build` keeps them true
as `status:` and other fields change. Hand-edited tags in derived facets are
overwritten on the next build — change the frontmatter, not the tag. Manually
curated `ontology/*`, `capability/*`, `req/*`, and `spec/*` facets are
preserved.

Open the workspace as an Obsidian vault and the tag pane gives you the
cross-repo slices from the Ontology Guide: `#kind/prd` for every PRD across
platforms, `#component/customer-notification-service` for everything touching
that component, `#team/customer-engagement #status/completed` for the team's
shipped work. Wire `graph build && git diff --exit-code` into CI (same pattern
as `governance resolve`) so committed tags can never drift from the
frontmatter they derive from.

## 10. The full loop, in one picture

```text
signal ──> discover new ──> discover validate ──> prd new (snapshot+checklist)
                                                        │
                                                  prd validate
                                                        │
                                            build (check ready / check done)
                                                        │
                                      update reality docs + link evidence
                                                        │
                                                  prd complete
                                                 /            \
                                    archive + log.md      outcome review (+90d)
                                                                │
                                                        learnings ──> next signal
```
