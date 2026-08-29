---
type: doc
tags: [doc/company-os-starter, kind/tutorial]
---

# Tutorial: From Idea to Archived PRD with `company-os`

This walkthrough uses the populated example in `../examples/workspace/` and the
`company-os` CLI. Every command and output below was executed
against this kit — you can reproduce the whole session.

**The principle behind everything:** strict on artifacts, flexible on process.
The CLI and skills guide you; the validators enforce only the contract. How you
think, draft, or prompt your agent is your business.

New here? Start with [GOLDEN-PATH.md](GOLDEN-PATH.md) — it takes an empty
directory to a completed change on a fresh workspace you scaffold yourself.

## 0. Setup

```bash
# requirements: the `company-os` binary on your PATH — nothing else.
# Download a release artifact and `chmod +x` it, or build from source:
#   make install        (from company-os-starter/; puts it in ~/.local/bin)
cd ../examples/workspace     # or: export COMPANY_OS_WORKSPACE_ROOT=$PWD/../examples/workspace
```

The workspace layout (each directory is its own Git repo in real life):

```text
../examples/workspace/
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

Not sure what to do first — in this workspace, or any workspace? Ask the tool
instead of reading the layout above:

```bash
$ company-os next
next: outcome review for 2026-per-channel-quiet-hours due 2026-10-16
  platform: communications
```

`next` is read-only: it scans for the single highest-priority pending action
(an expiring deviation/exception, a PRD failing its contract, an unchecked
governance item, a stale reality doc, a due outcome review) and prints the
exact command to run. `--all` lists every pending item instead of just the
top one. Run it any time you land in an unfamiliar workspace — before reading
docs, before running anything else.

### Prefer a menu? `company-os tui`

If you would rather browse than memorize commands, `company-os tui` opens a
menu-driven UI over the same commands. It is safe as a first command — it runs
outside a workspace too, and offers to scaffold one there. Four views exist
*only* in the TUI: a workspace overview, and browsers for components, PRDs and
discovery briefs. It also gives one-key fixes for regenerating derived state and
re-resolving governance.

**The whole loop is in the menu.** Every step below — write a brief, validate
it, write the PRD from it, check the PRD, describe current state, complete the
change — has a guided form. Each previews the exact command it is about to run
and writes nothing until you confirm, so the menu teaches the CLI rather than
replacing it.

```bash
company-os tui        # and that is the loop, end to end
```

Two deliberate omissions. `prd complete --force` has no field: it overrides the
check that a change is not done until reality is updated, and that stays a
typed, deliberate act. And no browsing screen writes — `discover validate`
rewrites the brief it is given, so it lives with the forms, never in the
discovery browser. `workspace sync` and `scratchpad init` have no forms either;
both need a value the workspace does not contain.

The rest of this tutorial uses the CLI throughout, because a tutorial has to
show you what ran. If you run this loop often the CLI is the faster surface
anyway — and as of the flag inference shown below, the commands are complete as
written, not abbreviations.

## 0.5 Configuring paths on your machine

Your absolute paths are yours; the committed YAML must stay portable. The rule
is simple: **no `/Users/yourname/...` string is ever committed.** Absolute paths
live only in environment variables — never in a tracked file. That way you,
your teammates, and CI can clone the same repos to completely different
locations and everything still resolves.

### Precedence (highest wins)

```text
1. CLI flag              company-os --root /abs/path ...
2. Environment variable  $COMPANY_OS_WORKSPACE_ROOT
3. Built-in default      current working directory
```

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

A per-repo `.company-os.local.yaml` override, a `~/.company-os/config.yaml`
multi-workspace switcher, and a repo-cloning step driven by a committed
`config/repositories.yaml` are proposed but not implemented — see
`00-original-proposal.md` for that design. (`company-os workspace sync` does
exist today, but for the unrelated federated multi-repo manifest — see §8
below — not for this proposal.)

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

Try validating the empty brief. The brief validates — and the CLI tells you what
you have not filled in yet, without blocking you:

```bash
$ company-os discover validate 2026-per-channel-quiet-hours --team customer-engagement
  [warn] 3 sections are empty: Problem signal, Hypothesis, Success criteria — format guidance only; the team may use its own structure (opt in via standards/doc-formats.yaml)
  [ok] brief '2026-per-channel-quiet-hours' validated (status: validated)
```

That is *strict on artifacts, flexible on process* in one line: the **headings**
are the contract and a missing one always fails, but whether a section has prose
in it yet is your team's business. A team that wants empty sections to block
opts in with `enforce: true` in `teams/<t>/standards/doc-formats.yaml`, and then
gets one blocking `[FAIL]` per empty section instead of this single warning.

Fill the three mandatory sections (how you research them — interviews, data,
prototypes — is `guidance`-tier, i.e. your choice), then:

```bash
$ company-os discover validate 2026-per-channel-quiet-hours --team customer-engagement
  [ok] brief '2026-per-channel-quiet-hours' validated (status: validated)
```

`--team` is optional when the brief id is unique across the workspace — the
CLI searches `teams/*/product/discovery/` for it. On a second brief in this
same workspace:

```bash
$ company-os discover validate 2026-test-flag-free
  [warn] 3 sections are empty: Problem signal, Hypothesis, Success criteria — format guidance only; the team may use its own structure (opt in via standards/doc-formats.yaml)
  [ok] brief '2026-test-flag-free' validated (status: validated)
```

Several teams with a brief of the same id → the command lists every candidate
and asks you to disambiguate with `--team`; an explicit `--team` always wins
over inference.

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

Same inference as `discover validate`: `--platform` is optional when the PRD
id is unique across `platforms/*/change-records/active/` and
`platforms/*/archive/prds/`. `prd complete` infers it the same way. Dropping
the flag on a fresh PRD:

```bash
$ company-os prd validate 2026-test-flag-free
  [warn] 3 sections are empty: Problem statement, Success metrics, Proposed change — format guidance only; the team may use its own structure (opt in via standards/doc-formats.yaml)
  [FAIL] process contract field 'decisionOwner' missing or TODO
```

> **Why `prd validate` is stricter than `company-os validate`.** They check
> different things on purpose, so a PRD can pass the workspace gate and still
> fail here. Gate `[3/8] active PRD contracts` is the **active-record floor** —
> a workspace-wide sweep asserting four fields (`title`, `team`, `components`,
> `governanceSnapshot`) on every PRD already live under
> `change-records/active/`. `prd validate` is the **author-time pre-flight** —
> it adds `platform` and `decisionOwner`, which the author is expected to settle
> before delivering, and which the gate does not demand of a document whose
> location already states its platform. If the two ever agree exactly, one of
> them is redundant.

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
validating workspace /path/to/examples/workspace

[1/8] ownership reconciliation
  [ok] customer-notification-service: registry and descriptor agree (communications)

[2/8] deviation and exception expiry
  [ok] customer-engagement: deviation platform-standard://communications/prd-structure current (review 2035-01-15)
  [ok] customer-engagement: deviation company-standard://estimation/story-points current (review 2035-01-14)
  [warn] customer-engagement: exception for platform-standard://communications/message-schema is approved by a placeholder (TODO: rule owner) — replace it with the rule owner
  [ok] customer-engagement: exception platform-standard://communications/message-schema valid until 2035-12-31

[3/8] active PRD contracts

[4/8] frontmatter core and tag derivation (interop contract)
  [ok] company-os/onboarding/developer.md: core fields + tags in sync
  ...                                       # one line per frontmatter document
  [ok] company-ontology/contexts/communications.md: core fields + tags in sync

[5/8] CLAUDE.md context node drift (fail-safe, absence-tolerant)
  [ok] company-os/CLAUDE.md: context node in sync
  [ok] platforms/communications/CLAUDE.md: context node in sync
  [ok] teams/customer-engagement/CLAUDE.md: context node in sync
  [ok] company-ontology/CLAUDE.md: context node in sync
  [ok] platforms/communications: directory indexes in sync (1 index(es))
  [ok] teams/customer-engagement: directory indexes in sync (1 index(es))
  [ok] company-ontology: directory indexes in sync (1 index(es))

[6/8] feature-index drift (derived component->artifact map)
  [ok] communications: feature-index in sync (1 component(s))

[7/8] custom skills layering (shadowing + extends resolution)
  [ok] skills layered cleanly (2 canonical, 0 team; no shadowing or dangling extends)

[8/8] promotion integrity (promoted drafts match their change records)

PASS
```

Note the `[warn]` in gate 2: warnings name something worth fixing (here, an
exception still carrying the placeholder approver `company-os exception request`
wrote) but they do not fail the gate. Only `[FAIL]` lines count as problems.

The gate count is dynamic: the eight gates above run in monorepo mode. In a
**federated** workspace (a `workspace.yaml` manifest is present) validate adds a
ninth gate — `[9/9] federated slice integrity` — which fails if a materialized
governance slice was hand-edited (its content hash no longer matches
`workspace.lock.yaml`). With no manifest the ninth gate does not exist and the
output is byte-for-byte the eight-gate form above.

It fails (exit 1, blocking merge) when: a team claims ownership the component
descriptor doesn't confirm (single-source rule), a deviation passes its
`reviewDate`, an exception is missing or past its `expires`, an active PRD
is missing contract fields, a doc's frontmatter core or derived tags drift, a
generated `CLAUDE.md` context node or per-directory `index.md` is stale, a
platform's derived `feature-index.yaml` is out of date, or a promoted draft no
longer matches its change record. The absence-tolerant gates pass when the
artifact is absent. Wire it as CI:

```yaml
# .github/workflows/os-validate.yml (any OS repo)
- name: install company-os        # one static binary, no runtime dependency
  env: {COMPANY_OS_VERSION: v1.0.0}
  run: |
    curl -fsSLo /usr/local/bin/company-os \
      <release-url>/$COMPANY_OS_VERSION/company-os_${COMPANY_OS_VERSION}_linux_amd64
    chmod +x /usr/local/bin/company-os
- run: company-os --root . validate
- run: company-os governance resolve --team <team> && git diff --exit-code teams/*/generated/
```

The second check ensures `effective-governance.yaml` is truly derived — if
someone hand-edited it, the regenerated file differs and CI fails.

Locally, `--fix` regenerates that same derived state (effective-governance,
tags, indexes, CLAUDE.md context nodes) before running the gates, so you don't
have to remember `governance resolve` and `derive` yourself:

```bash
$ company-os validate --fix
...
PASS
validate --fix: 0 file(s) regenerated
```

`--fix` is a **local convenience only**. CI keeps the strict two-step check
above (`validate` + `git diff --exit-code` against the freshly regenerated
files) — a workspace with drift still fails CI even if `--fix` would have
silently repaired it locally.

## 9. Where personal flexibility lives

```bash
$ company-os scratchpad init --repo teams/customer-engagement
```

See `scratchpad/personal-rules/maria-prd-style.md` in the example: Maria's
agent applies her drafting style *on top of* `skill://product/creating-prd`.
Per `team.yaml` precedence, `canonical-mandatory > personal > canonical-default
> canonical-guidance` — her rules can reshape how the PRD gets written, never
whether it passes `prd validate`.

## 9.5 Tags everywhere: `derive`

Every artifact in the kit now carries a `tags:` block — skills, templates,
docs, standards, YAML configs, and workspace documents. Two kinds exist:

- **Static facets** on canonical content (skills: `kind/skill authority/canonical
  process/prd`; configs: `kind/requirements platform/communications`).
- **Derived facets** on workspace docs, generated from frontmatter IDs by:

```bash
$ company-os derive
  tagged platforms/communications/archive/prds/2026-per-channel-quiet-hours/prd.md
  ...
derive: 11 doc(s) scanned, 9 updated
$ company-os derive
derive: 11 doc(s) scanned, 0 updated     # idempotent
```

The archived PRD, for instance, ends up with
`[component/customer-notification-service, discovery/2026-..., kind/prd,
platform/communications, status/completed, team/customer-engagement]` — all
derived from fields it already had. Scaffolds (`discover new`, `prd new`,
`prd complete`) emit starter tags at creation; `derive` keeps them true
as `status:` and other fields change. Hand-edited tags in derived facets are
overwritten on the next build — change the frontmatter, not the tag. Manually
curated `ontology/*`, `capability/*`, `req/*`, and `spec/*` facets are
preserved.

Open the workspace as an Obsidian vault and the tag pane gives you the
cross-repo slices from the Ontology Guide: `#kind/prd` for every PRD across
platforms, `#component/customer-notification-service` for everything touching
that component, `#team/customer-engagement #status/completed` for the team's
shipped work. Wire `derive && git diff --exit-code` into CI (same pattern
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

## 11. Finding things: `company-os find`

Local search is otherwise fragmented across derived tags, the ids registry,
per-directory `index.md` files, and feature-indexes.
`find` is one front door over all of them — case-insensitive substring match,
exact-id hits ranked first, grouped output by match kind:

```bash
$ company-os find customer-notification-service
== canonical IDs ==
  platforms/communications/components/customer-notification-service.yaml  [component://customer-notification-service]  id substring match

== derived tags ==
  platforms/communications/archive/prds/2026-per-channel-quiet-hours/prd.md  [Per-channel quiet hours]  tag: component/customer-notification-service
  platforms/communications/reality/components/customer-notification-service.md  tag: component/customer-notification-service

== frontmatter ==
  platforms/communications/reality/components/customer-notification-service.md  id field match

== feature-index ==
  platforms/communications/generated/feature-index.yaml  [customer-notification-service]  feature-index component match

graphify installed but no graphify-out/graph.json here — run graphify to build the graph
```

A query that doesn't hit anything is not a failure — search is not a gate:

```bash
$ company-os find zzznotarealthing
no matches

graphify installed but no graphify-out/graph.json here — run graphify to build the graph
```

**The graphify hook:** if a `graphify` binary is on `PATH` and
`graphify-out/graph.json` exists under the workspace root, `find` appends a
`graphify query "<query>"` section using the built graph's EXTRACTED/INFERRED
edges. Either missing → the quiet hint line above, exit 0 regardless. Skip the
hook explicitly with `--no-graphify`:

```bash
$ company-os find customer-notification-service --no-graphify
== canonical IDs ==
  platforms/communications/components/customer-notification-service.yaml  [component://customer-notification-service]  id substring match
...
```

`--json` emits the same records structured, for scripting or an agent to
consume directly.
