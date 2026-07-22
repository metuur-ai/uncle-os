---
type: doc
tags: [doc/company-os-starter, kind/guide, ontology/guide]
---

# Ontology Guide: Defining Meaning, Correlating with Tags, Tracing with EARS `@spec`

**Companion to:** the Federated OS proposal and the `company-os-starter` kit.
A worked example lives in `examples/workspace/company-ontology/`.

Three questions, three mechanisms, one graph:

| Question | Mechanism | Source of truth |
|---|---|---|
| What does this thing *mean*? | Ontology (canonical IDs + concept notes) | `company-ontology/` |
| What is this doc *about*? | Tags + wikilinks (derived from frontmatter) | frontmatter → generated |
| Does the code *satisfy* the requirement? | EARS clauses + `@spec` annotations | `requirements.yaml` → code/tests |

The invariant that keeps all three honest: **IDs are canonical; tags and links
are derived.** Humans and agents write stable IDs in frontmatter; tooling
generates the Obsidian-friendly tags and wikilinks from them. Never the reverse.

---

# Part 1 — How to Define the Ontology: a 7-Step Guide

Keep it **thin**. The enterprise ontology exists for *identity and
integration*, not to be the one true model of everything (that would rebuild
the monolith DDD exists to avoid). Rich models live inside bounded contexts.

## Step 1 — Harvest the nouns
Sweep existing docs (PRDs, architecture docs, org charts) and list every
recurring noun: Customer, Order, Component, Capability, Team, Requirement…
Don't define anything yet. Output: a flat list with where each noun appears.

## Step 2 — Split shared concepts from contextual ones
For each noun ask: *does it mean exactly the same thing everywhere?*
- Yes → candidate for the shared upper ontology. Governance/structure concepts
  almost always qualify: `Component, Capability, Platform, Team, Requirement,
  Policy, Risk, Evidence, Repository`.
- No (Customer, Account, Product…) → it belongs *inside* bounded contexts,
  possibly under different names. The shared layer records only its identity
  and the mappings between contexts.

Rule of thumb: if the upper ontology exceeds ~30 concepts, you are modeling
someone's domain instead of the enterprise skeleton. Push it down.

## Step 3 — Assign canonical IDs
Every concept, context, and instance gets a stable, URL-style ID:

```text
concept://component                         # a concept (a "type")
component://customer-notification-service   # an instance
capability://communications/message-delivery
context://communications
req://communications/delivery-reliability   # a requirement (see Part 3)
map://crm--communications                   # a context map
skill://product/creating-prd                # a canonical skill (scope/name)
platform-skill://communications/creating-prd # extend-reference to a platform-layer skill
```

Two schemes address custom skills (Unit 5, `docs/ears/golden-path-flavor-federation.md`):

- `skill://<scope>/<name>` is a skill's own canonical `id:`. The `<scope>` is a
  semantic namespace (`product`, `governance`, `company`…), **not** the platform.
- `platform-skill://<platform>/<name>` is the reference form used in a skill's
  `extends:` frontmatter to layer on a platform-layer base. It resolves to the
  base skill **file** `platforms/<platform>/skills/<name>.SKILL.md` — whose own
  `id:` is a `skill://<scope>/<name>` — where `<name>` is that file's stem name
  and `<platform>` is the platform directory. A team/personal skill reusing a
  canonical skill's `skill://` id or name is *shadowing* and fails `validate`;
  `extends: platform-skill://…` is the sanctioned, additive alternative.

Register every ID in `company-ontology/ids/registry.yaml` with a `definedIn`
pointer. One ID, one definition location — validation fails on unregistered
or doubly-defined IDs. IDs never change; names and aliases can.

## Step 4 — Define concepts as notes with relationships
One markdown file per concept in `concepts/`, frontmatter carrying the machine
part, body carrying the human definition (see `concepts/component.md`):

```yaml
---
type: concept
id: concept://component
aliases: [Component]
relationships:
  - owned-by: concept://team
  - implements: concept://capability
taxonomy:
  subtypes: [api, service, worker, library, database, ui]
---
```

Note: **taxonomy is part of the ontology**, not a separate layer — `subtypes`
is just the is-a relationship living beside `owned-by` and `implements`.

## Step 5 — Declare bounded contexts and their ubiquitous language
One file per context in `contexts/`. This is where DDD plugs in: the context
declares its vocabulary — and, just as importantly, its *forbidden* terms:

```yaml
ubiquitousLanguage:
  Recipient: The addressable target of a Message. NOT called "Customer" here.
forbiddenTerms:
  Customer: Use Recipient. "Customer" belongs to context://crm.
```

A platform is usually one bounded context (or a small cluster). Every canonical
doc in that platform declares `boundedContext: context://communications` and is
lint-checked against the vocabulary (Part 4).

## Step 6 — Write context maps for every integration
Where two contexts exchange data, a `context-maps/` note records the
translation table (Customer→Recipient), the direction (upstream/downstream),
and the integration style (ACL, published language…). Context maps are the
*only* place cross-context vocabulary may legally appear side by side.

## Step 7 — Govern the ontology like code
- New concept / renamed term / new context = pull request against
  `company-ontology`, reviewed by the semantics owner (usually architecture),
  separate from governance changes.
- Team OS **references** ontology, never defines it. A team needing a new
  concept opens a PR upstream.
- Deprecate, don't delete: mark concepts `status: deprecated` with a
  `replacedBy:` so old links keep resolving.

---

# Part 2 — Tags: Correlating Content Across Directories, Platforms, and Teams

Directories give you *one* hierarchy. Tags give you every other slice: "all
mandatory requirements touching message-delivery, across every platform and
team," regardless of which repo the files sit in. In an assembled Obsidian
vault (every OS repo mounted as a folder), tags + wikilinks turn the
federation into one navigable graph.

## 2.1 The tag namespace (nested, faceted)

```text
#kind/prd  #kind/adr  #kind/reality  #kind/discovery  #kind/skill  #kind/outcome
#platform/communications
#team/customer-engagement
#context/communications
#component/customer-notification-service
#capability/message-delivery
#req/communications/delivery-reliability
#tier/mandatory   #tier/default   #tier/guidance
#status/active    #status/completed   #status/deprecated
#spec/communications/delivery-reliability      (see Part 3)
```

Each axis is a facet; Obsidian's nested-tag pane and search compose them:

```text
tag:#capability/message-delivery tag:#kind/prd            → every PRD ever touching the capability
tag:#team/customer-engagement tag:#tier/mandatory         → the team's mandatory surface
tag:#component/customer-notification-service -tag:#status/completed → live work on the component
```

## 2.2 Derived, never hand-written

Tags duplicate information that already exists in frontmatter IDs — so hand
writing them would create drift. The rule:

```text
frontmatter IDs  ──(company-os graph build)──►  tags: [...] block + wikilinks
```

`graph build` reads `platform:`, `team:`, `components:`, `boundedContext:`,
requirement references, and `status:` from each doc's frontmatter and rewrites
a generated `tags:` array (Obsidian reads frontmatter tags natively). Editing
a tag by hand is futile — the next build overwrites it. Change the frontmatter
instead. `validate --ontology` fails if committed tags differ from a fresh
derivation, same pattern as `effective-governance.yaml`.

## 2.3 Wikilinks and hub notes: making the graph mean something

Tags classify; **wikilinks draw the edges** Obsidian's graph renders. The trick
that makes cross-repo correlation work: every ontology entry (concept,
capability, context, component descriptor's companion note) is a **hub note**,
and every doc that references its ID also gets a generated wikilink to it:

```markdown
<!-- generated block at the bottom of a PRD -->
## Graph
[[capability--message-delivery]] · [[component--customer-notification-service]] ·
[[context--communications]] · [[team--customer-engagement]]
```

Because hub notes carry `aliases:` ("Message Delivery"), prose mentions can be
linked naturally too. In the graph view, the hub becomes the star center:
click `capability--message-delivery` and you see the PRDs, reality docs, ADRs,
requirements, and teams orbiting it — across every repo in the vault. That is
the cross-directory correlation you asked for, and it doubles as agent
context-selection: "load everything one hop from this hub."

## 2.4 Naming conventions that keep Obsidian happy

- Hub note filenames mirror IDs with `--`: `capability--message-delivery.md`
  (slashes are illegal in filenames; `--` round-trips cleanly).
- One hub per ID; everything else links to it, nothing redefines it.
- Tags use `/` nesting exactly parallel to ID paths, so translating between
  `req://communications/delivery-reliability` and
  `#req/communications/delivery-reliability` is mechanical.

---

# Part 3 — EARS + `@spec`: Requirements You Can Trace to Code

## 3.1 EARS in one minute

EARS (Easy Approach to Requirements Syntax) gives every requirement one of six
shapes, which kills ambiguity and — because the shapes are regular — makes
requirements parseable:

| Pattern | Template |
|---|---|
| Ubiquitous | The `<system>` shall `<response>` |
| Event-driven | **When** `<trigger>`, the `<system>` shall `<response>` |
| State-driven | **While** `<state>`, the `<system>` shall `<response>` |
| Unwanted behavior | **If** `<condition>`, **then** the `<system>` shall `<response>` |
| Optional feature | **Where** `<feature is present>`, the `<system>` shall `<response>` |
| Complex | Combinations of the above |

## 3.2 EARS inside `requirements.yaml`

Platform requirements stay outcome-shaped, but the verification checklist
becomes numbered EARS clauses — each clause gets a stable anchor (`R1…Rn`):

```yaml
- id: delivery-reliability          # full id: req://communications/delivery-reliability
  level: mandatory
  version: "2.1"
  ears:
    - id: R1
      pattern: event
      clause: >
        When a message delivery attempt fails, the notification service shall
        queue the message for retry with an idempotency key.
    - id: R2
      pattern: unwanted
      clause: >
        If a message exhausts all retries, then the notification service shall
        move it to a recoverable dead-letter state and emit a delivery-failed event.
    - id: R3
      pattern: ubiquitous
      clause: >
        The notification service shall expose delivery status for every
        accepted message.
    - id: R4
      pattern: state
      clause: >
        While a recipient's quiet hours are active for a channel, the service
        shall hold non-urgent messages and deliver them at the window end.
```

Note R4 uses the Communications ubiquitous language: *recipient*, *channel*,
*message* — EARS clauses are written **in the vocabulary of their bounded
context**, which is how ontology and spec-writing reinforce each other.

## 3.3 The `@spec` annotation

A grep-able marker that binds an implementation or test to an EARS clause:

```text
@spec req://communications/delivery-reliability@2.1#R2
```

`ID @ version # clause`. It lives everywhere the requirement is realized:

```python
# notification_service/retry.py
# @spec req://communications/delivery-reliability@2.1#R1
# @spec req://communications/delivery-reliability@2.1#R2
def handle_failed_delivery(msg): ...
```

```python
# tests/test_delivery_reliability.py
def test_exhausted_retries_move_to_recoverable_dlq():
    """@spec req://communications/delivery-reliability@2.1#R2"""
```

```gherkin
# features/quiet_hours.feature
@spec-req-communications-delivery-reliability-2.1-R4
Scenario: non-urgent message held during quiet hours
```

And in OKF docs, `graph build` converts `@spec` mentions into the tag
`#spec/communications/delivery-reliability` plus a wikilink to the
requirement's hub note — so specs, code pointers, PRDs, and tests all orbit
the same node in the Obsidian graph.

## 3.4 The spec-driven development loop

```text
EARS clauses (requirements.yaml, versioned)
      │  prd new injects them into the PRD checklist per clause
      ▼
PRD: - [ ] delivery-reliability@2.1#R4 — evidence:
      │  developer implements; tests carry @spec
      ▼
company-os spec trace   →  coverage matrix:
      R1  code ✓  test ✓
      R2  code ✓  test ✓
      R3  code ✓  test ✗   ← FAIL: mandatory clause without a test
      R4  code ✓  test ✓
      │
      ▼
prd complete: evidence links = the traced tests; done only when
mandatory clauses are covered and reality docs are updated
```

`spec trace` is a simple scanner: grep `@spec` across the component repos
listed in the team's ownership, parse `id@version#clause`, join against
`requirements.yaml`, and report per-clause coverage. Because the PRD pinned
`governanceSnapshot`, the trace evaluates against the clause versions the work
started with. When a requirement bumps to 2.2, stale `@spec ...@2.1` markers
surface automatically as "annotated against a superseded version."

## 3.5 Rules that keep `@spec` honest

1. Only `mandatory` clauses *require* trace coverage; `default` clauses report
   but don't block (deviations may apply); `guidance` is never traced.
2. An `@spec` pointing at an unregistered ID or nonexistent clause fails
   `validate --ontology` — annotations can't rot silently.
3. One clause may have many `@spec` sites; a mandatory clause with **zero**
   test-side sites blocks `prd complete`.
4. Never paraphrase the clause at the annotation site — the ID is the link;
   duplicated prose would drift.

---

# Part 4 — Validation: What `validate --ontology` Adds

Extending the kit's CI gate with semantic checks:

```text
[4/6] id registry            every component://, capability://, req://, context://
                             reference in any frontmatter resolves in ids/registry.yaml;
                             every registry entry's definedIn file exists
[5/6] vocabulary lint        canonical docs declaring boundedContext X must not use
                             X's forbiddenTerms (e.g. "Customer" in a Communications
                             reality doc → FAIL with "use Recipient")
[6/6] derivation freshness   committed tags/graph blocks match a fresh `graph build`;
                             @spec annotations parse and resolve to live clauses;
                             mandatory-clause test coverage per `spec trace`
```

Scratchpads and archives are exempt from vocabulary lint (history keeps its
original words); only `authority: canonical` docs are held to it.

---

# Quick-Start Checklist

1. Create `company-ontology/` (example scaffold: `examples/workspace/company-ontology/`).
2. Steps 1–3: harvest nouns, split shared vs contextual, register IDs.
3. Write hub notes for your top ~10 shared concepts and your first bounded
   context with its ubiquitous language + forbidden terms.
4. Add `boundedContext:` and ontology IDs to frontmatter of canonical docs.
5. Rewrite one mandatory requirement's checklist as numbered EARS clauses.
6. Annotate its implementation and tests with `@spec id@version#clause`.
7. Wire `graph build` (tags + hub wikilinks) and `validate --ontology` into CI.
8. Open the assembled vault in Obsidian: filter `#capability/...`, click a hub,
   watch the federation become one graph.
