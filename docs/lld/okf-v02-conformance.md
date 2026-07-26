---
type: lld
id: lld-okf-v02-conformance
title: OKF v0.2 Conformance — Low-Level Design
status: locked
tags: [kind/lld, status/locked]
---

# OKF v0.2 Conformance — Low-Level Design

Implements `docs/hld/okf-v02-conformance.md`. Requirements are in
`docs/ears/okf-v02-conformance.md` (R-N.M).

Scope after review: **U0–U3**. Per-directory `index.md` generation is deferred to
its own change; the findings that cut it, and the decisions to carry forward, are
in the HLD's Deferred section.

## Architecture

This change is mostly documentation plus a frontmatter backfill. It adds no new
derived artifact, no new gate, and no new traversal. Three code touchpoints only.

### Component map

| Concern | Location | Change |
|---|---|---|
| Done-gate date compare | `cmd_prd` complete, `:679-682` | parse dates instead of comparing strings |
| Scaffolding — PRD / discovery / reality | `templates/*.md` + `*_TEMPLATE` strings | emit `title:`; reality emits `resource:` |
| Scaffolding — outcome review | `cmd_prd` complete, `:701-705` | emit `title:` (identity stays `prd:`) |
| Field documentation | `docs/FRONTMATTER-CORE.md` | document `title`/`resource`; reserve `generated`/`verified` |
| Conformance contract | new `company-os-starter/docs/CONFORMANCE.md` | new |
| Roadmap separation | `docs/ONTOLOGY-GUIDE.md` → new `docs/ONTOLOGY-ROADMAP.md` | split |
| Fixture claim corrections | two `contexts/communications.md` | reword `validate --ontology` as roadmap |
| Profile enum | `docs/01-flexibility-skills-and-role-views.md:121` | `minimal \| standard \| strict` |

### What `title:` touches downstream

`build_claude_node:1682` renders `title or id or filename`. Backfilling titles
therefore changes every committed `CLAUDE.md` generated block, which
`acceptance.sh` §4 catches via `s0 == s1 == s2`. It does **not** change validate's
stdout, which carries paths only — so the golden snapshots stay frozen (R-3.7,
R-3.8).

## Key Decisions

### 1. The done-gate parses dates and fails loudly on malformed input (R-0.1, R-0.2)

Today: `str(r_meta.get("updated","")) < str(meta.get("created",""))` (`:679-682`).
Correct for well-formed ISO dates by lexical accident; silently wrong for
`18/07/2026`, for an empty value, or for a YAML-parsed `datetime.date` whose
`str()` differs in shape.

Replace with `dt.date.fromisoformat` on both sides, matching what gate 2 already
does for `reviewDate` and `expires` (`:945-960`). A malformed or missing value
becomes an explicit done-check error naming the file and the offending value —
not a pass, and not a traceback. Behaviour on conformant input is unchanged.

### 2. `resource:` is authored, and scoped to documents describing an external asset (R-3.3, R-3.4)

**Authored, not derived.** `graph build` today rewrites exactly one frontmatter
field — `tags:`, via `rewrite_frontmatter_tags:1334-1346`. Two reasons not to
extend that authority here, replacing an earlier draft's I8 argument, which was
wrong:

- **The available derivation covers one document type only.** `iter_graph_docs`
  computes `cid = str(meta.get("id","")).replace("reality-","")` for
  `type: component-reality` (`:1377-1381`). A concept's resource is
  `capability://…`, a context's is `context://…`. Deriving for one type and
  authoring for the rest ships a field that means "machine-owned" in `reality/`
  and "human-owned" everywhere else, with nothing telling a reader which.
- **That derivation is a string mangle, not a lookup.** `str.replace` is unanchored
  and global — `id: reality-reality-sync` yields `sync` — and nothing checks the
  result against `ws.find_component()` (`:245-251`). Tolerable for a tag facet;
  not for a published URI that OKF §4.1 defines as identifying the asset.

*(The earlier draft justified this on I8, claiming a derived `resource:` would be
a second machine-generated assertion of component identity. That premise is false:
`:1377-1381` already derives `component/<cid>` into frontmatter, `graph build`
already writes it, and gate 4 already blocks on it drifting. The conclusion was
right; the reasoning was not, and a maintainer checking it would have reopened the
decision.)*

**Scoped, not universal.** OKF §4.1 defines `resource` as the URI of *the asset the
document describes*, distinct from the document itself. A reality document
describes something outside the knowledge base — a running service. A discovery
brief describes nothing external; it *is* the artifact, and its `id:` already
identifies it. `definition-of-ready.md:2` already carries
`id: team-standard://customer-engagement/definition-of-ready`; adding `resource:`
there would duplicate it.

Rule: **`resource:` is present if and only if the document describes an asset
external to the knowledge base.** Today that is `type: component-reality` and
nothing else. This is why R-1.2 splits the recommendation rather than listing
`title` and `resource` in one breath — undifferentiated, the field reads as
inconsistently applied when it is in fact correctly scoped.

### 3. The conformance clause is checked against the CLI's blocking gates (R-1.4, R-1.13)

`title` cannot be documented as a plain SHOULD. It is blocking today in two places:

```python
# gate 3, :975
missing = [f for f in ["title", "team", "components", "governanceSnapshot"] if not meta.get(f)]
```
```python
# prd validate, :627-630
for field in ["title", "team", "platform", "components", "governanceSnapshot", "decisionOwner"]:
    if not meta.get(field) or meta.get(field) == "TODO":
```

For `type: prd` it is a MUST, enforced, exit 1. So every SHOULD in the clause is
verified against `core_field_errors:128-145`, gate 3 `:975`, and `cmd_prd`
validate `:627-630`, and anything that blocks anywhere is documented as
conditionally required. Symmetrically, every MUST-NOT-reject item cites a
`bin/company-os` file:line — an uncited item is a bug in the claim, not a
formatting gap.

### 4. `2026.2` is the floor; the bump rule is forward-only (R-1.5, R-1.6)

Defining `2026.2` retroactively means reconstructing what `2026.1` contained,
which means deriving a past release's contents from current code — not evidence,
and circular.

Instead: **`2026.2` is the first methodology version with a written conformance
clause.** Everything before it is `unversioned`. Nothing becomes retroactively
non-conformant.

**Bump rule:** `N` increments when the conformance clause changes what tooling
MUST accept or reject. Documentation, fixtures, and new non-blocking fields do not
bump. Chosen deliberately so **this change does not bump** — which makes the eight
fixtures already committing `2026.2` correct as they stand, rather than eight
files to reconcile for no user outcome.

**`profile:` is resolved in the same breath**, because it is already
self-contradictory in shipped artifacts.
`docs/01-flexibility-skills-and-role-views.md:121` documents
`standard | strict | provisional`, while
`examples/banking/bank/repos/platform-lending/platform.yaml:4` and
`examples/banking/small-company/platforms/product/platform.yaml:4` both ship
`profile: minimal`. Writing CONFORMANCE.md against the documented enum makes two
fixtures non-conformant on day one — the exact failure this change exists to
prevent. The enum becomes `minimal | standard | strict`; `provisional` is dropped
as used nowhere.

### 5. The conformance contract is a new file, not a promotion of FRONTMATTER-CORE (R-1.1)

`docs/FRONTMATTER-CORE.md:8-12` already calls itself "the whole interop contract",
and its "What validates what" table (`:139-147`) is the closest existing thing to a
conformance boundary. They stay separate because they answer different questions:
`FRONTMATTER-CORE.md` answers "which fields, on which document?" — a lookup, read
repeatedly. `CONFORMANCE.md` answers "what does conformant mean, what version is
this, what may tooling reject?" — a contract, read once and cited. Merging produces
one long document serving neither reader. `FRONTMATTER-CORE.md:8-12` is amended so
only one document claims to be the contract, and gains a link.

### 6. The ontology roadmap is split into its own file, and the fixtures are corrected too (R-2.1, R-2.4)

A banner on a 150-line section inside a 368-line document is easy to scroll past; a
separate file cannot be read by accident. It matches the existing pattern —
`docs/user-guide/explanation/observer-roadmap.md` is a file, not a banner, and
opens with the language to copy: "**Not yet available.** Everything on this page is
a design vision, not a shipped tool. No commands below can be run today."

`ONTOLOGY-GUIDE.md` keeps Parts 1–2 (canonical IDs, the registry, tag derivation,
`graph build`) — all runnable. The unshipped material moves out.

**The check extends to fixtures, not just docs.** Two shipped context documents
assert the unshipped command as fact:

- `examples/workspace/company-ontology/contexts/communications.md:20`
- `examples/federated/company-ontology/contexts/communications.md:20`

both reading "`company-os validate --ontology` flags forbidden terms in canonical
docs", while `examples/banking/bank/repos/company-ontology/contexts/payments.md:20`
correctly writes "(`validate --ontology`, roadmap)". Fixtures are what adopters
copy, so a false capability claim there is worse than one in the guide. The
banking wording is the model.

The federated copy sits inside a read-only slice whose lock hashes would break if
edited in place (I9) — it is corrected at its source and re-synced, or carved out
explicitly. R-2.5 covers this.

### 7. Scaffolding covers all four document-emitting paths (R-3.5, R-3.6)

Three are obvious: `prd new`, `discover new`, `reality new`. The fourth is easy to
miss — `prd complete` writes `outcome.md` inline at `:701-705`:

```python
outcome.write_text(
    f"---\ntype: outcome-review\nprd: {args.id}\ndue: {due}\nstatus: pending\n"
    f"tags: [kind/outcome, prd/{args.id}, status/pending]\n---\n\n"
```

No `title:`, no `id:`. Without fixing it, R-3.2 backfills the fixture's committed
`outcome.md` and the very next `prd complete` emits a non-conformant one. Its
identity field is `prd:`, not `id:` — accepted by `core_field_errors:132-135` — so
any title-fallback logic must not assume `id` exists.

This also matters for the deferred index change: a titled `outcome.md` is what
makes an `archive/prds/<id>/` index worth generating.

## Constraints

- **Python 3.9+, stdlib + PyYAML only.** `bin/company-os` is one self-contained
  file. Preserve the `die`/`ok`/`warn`/`fail` helpers and `frontmatter()`'s exact
  `^---\n...\n---\n` contract.
- **Gate 1–7 numbering and printed strings are frozen** (I5, `:915-918`). This
  change adds no gate and changes no gate output.
- **The golden snapshots must not move.** `examples/acceptance.sh` is the oracle;
  a diff in this change is a regression, not a re-baseline (R-5.6).
- **`description:` values will need quoting.** A useful description usually
  contains a colon, and `description: Validated: 412 tickets` is a YAML parse
  error under PyYAML. When `description:` lands with the deferred index change,
  its template placeholder must be quoted and FRONTMATTER-CORE must say so.
- **`examples/banking/` has no acceptance.sh coverage** (38 markdown files,
  exercised by nothing). It is added to the fixture loop before it is edited, or
  left untouched — R-6.2/R-6.3.
- **Templates and the CLI's `*_TEMPLATE` strings must stay in sync**, per the repo
  `CLAUDE.md`.

## Out of Scope

Deferred with intent; each recorded in `CONFORMANCE.md`'s "considered and
deferred" section.

- Per-directory `index.md` generation and `description:` — deferred together to
  their own change; see the HLD's Deferred section for the blockers and the
  decisions already made.
- Provenance and trust semantics — `generated`, `verified`, the actor convention,
  trust-tier derivation. **Names reserved here, semantics deferred.** (N1)
- OKF `sources[]` and per-claim footnote attribution. (N2)
- Attested Computation. (N3)
- `stale_after:` and any freshness view. (N4)
- Building anything from the ontology roadmap. (N6)
- Repository directory sync via GitHub MCP / Actions. (N8) The source→destination
  mapping already ships as `paths:` + `root:`; the unmet parts are direction
  (push-back, forbidden by the read-only slice model) and CI automation. See
  `.devlocal/research/2026-07-25-knowledge-catalog-directory-sync.md`.
- Resolving governance-checklist `evidence:` URIs.
