---
type: ears
id: ears-okf-provenance-and-indexes
title: OKF Provenance and Indexes — EARS Specifications
status: draft
---

# OKF Provenance and Indexes — EARS Specifications

**Source:** Units D1 and D2 of `docs/ears/okf-v02-conformance.md`
§"Deferred — decided, not built", promoted to normative.
**Design:** `docs/hld/okf-provenance-and-indexes.md`,
`docs/lld/okf-provenance-and-indexes.md`.
**Precondition:** task 1.6 of the parent reserves `generated:` and `verified:` as
inert names. This change makes them live.

Build order: Unit 1 → Unit 2 → Unit 3 → Unit 4 → Unit 5.

---

## Unit 1: `description:` as a documented, emitted field

**Why:** the index is worthless without it. An index listing titles alone is
`ls` with extra steps, and `description:` measures zero occurrences across every
fixture today. The field ships first so the generator in Unit 2 has something to
render.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL document `description:` in `company-os-starter/docs/FRONTMATTER-CORE.md` as recommended on every document, non-blocking, with the index named as its consumer. |
| R-1.2 | THE SYSTEM SHALL define a description rubric requiring at least one fact not derivable from the document's filename, `title:`, `type:`, or directory path. |
| R-1.3 | THE SYSTEM SHALL state that a description MUST NOT read as true when pasted onto a sibling document in the same directory. |
| R-1.4 | WHERE a template emits a `description:` placeholder, THE SYSTEM SHALL quote the placeholder value, since a useful description commonly contains a colon and would otherwise be a YAML parse error. |
| R-1.5 | WHEN `prd new`, `discover new`, or `reality new` scaffolds a document, THE SYSTEM SHALL emit a `description:` field, updating both the built-in string in `internal/scaffold/template.go` and its peer file under `company-os-starter/templates/`. |
| R-1.6 | WHEN `prd complete` writes `outcome.md`, THE SYSTEM SHALL emit a `description:` field. |
| R-1.7 | IF a document omits `description:`, THE SYSTEM SHALL validate it unchanged, producing neither an error nor a warning. |

---

## Unit 2: Per-directory `index.md` generation

**Why:** the progressive-disclosure payoff. An agent opening
`platforms/communications/reality/` today must `ls` and open files to learn what
is there. Deep trees make this matter more here than in a flat OKF bundle.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL generate an `index.md` in any directory directly holding two or more graph documents, and SHALL NOT generate one below that threshold. |
| R-2.2 | THE SYSTEM SHALL render each entry as a link carrying the document's title and its `description:`, grouped by `type:`, reusing the title fallback chain at `internal/graph/node.go:319-324` (`title` → `id` → filename). |
| R-2.3 | THE SYSTEM SHALL write `index.md` with YAML frontmatter carrying `type: index` and a `company-os:generated` marker block delimiting the generated interior. |
| R-2.4 | THE SYSTEM SHALL register `index` in the type→tag vocabulary at `internal/graph/tags.go:24-28` as an inert type, against which no required-field gate runs. |
| R-2.5 | THE SYSTEM SHALL add `"index.md"` to `skipNames` (`internal/graph/tags.go:41`) in the same commit as the generator, so a generated index is never re-ingested as a graph document, counted toward its own directory's threshold, or listed inside itself. |
| R-2.6 | THE SYSTEM SHALL generate indexes from every derived-artifact path — `rebuildGenerated` (`cmd/company-os/scaffold.go:46`) and its records-returning twin (`cmd/company-os/product.go:26`) — so that `company-os init` does not produce a workspace failing its own `validate`. |
| R-2.7 | WHERE a path lies beneath a manifest-declared slice root, THE SYSTEM SHALL NOT generate an index there and SHALL NOT drift-check one there. |
| R-2.8 | THE SYSTEM SHALL NOT generate an index anywhere under `knowledge/`, which is a node root but not a graph-docs root (`internal/graph/tags.go:191`). |
| R-2.9 | WHEN a directory that previously qualified drops below two graph documents, THE SYSTEM SHALL delete its generated `index.md`. |
| R-2.10 | IF an `index.md` exists without a `company-os:generated` marker block, THE SYSTEM SHALL leave it untouched, SHALL report it as hand-owned at severity OK, and SHALL pass — matching the settled `CLAUDE.md` precedent at `internal/graph/node.go:126-145` and `internal/graph/gates.go:291` ("hand-owned, no generated markers (-> pass)"). |
| R-2.11 | THE SYSTEM SHALL leave the workspace byte-identical after two consecutive `derive` runs (I6). |
| R-2.12 | THE SYSTEM SHALL generate exactly three `index.md` files in `examples/workspace/` and zero in `examples/standalone-team/`. |

---

## Unit 3: Index drift detection

**Why:** an index is a generated artifact, and invariant I4 makes derived-vs-
authored a hard boundary. An index that silently goes stale is worse than none,
because a reader trusts it.

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL perform index drift detection inside gate 5, adding no gate and renumbering none (I5). |
| R-3.2 | THE SYSTEM SHALL leave gate 5's printed header byte-identical to its committed form, `[5/N] CLAUDE.md context node drift (fail-safe, absence-tolerant)`. |
| R-3.3 | WHEN a committed `index.md` differs from a fresh derivation, THE SYSTEM SHALL fail gate 5 and name the file. |
| R-3.4 | WHERE a directory qualifies for an index but none exists, THE SYSTEM SHALL fail gate 5 and name the directory. |
| R-3.5 | THE SYSTEM SHALL report index drift as one aggregate line per federation root. |
| R-3.6 | WHILE three of four federation roots are absent, THE SYSTEM SHALL keep passing `validate` on `examples/standalone-team/` (I2). |

---

## Unit 4: `generated:` and `verified:` provenance

**Why:** the largest conceptual gap. The whole `skills/` layer exists to direct
agents producing PRDs and discovery briefs, and nothing in the system records
that a document was agent-drafted rather than written by a person.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL record content origin as `generated: {by, at}` and confirmation as `verified: [{by, at}]`, kept distinct because a document's author need not be its confirmer. |
| R-4.2 | THE SYSTEM SHALL identify actors as `<producer>/<version>` for agents, `human:<id>` for people, and `process:<id>` for automated processes. |
| R-4.3 | THE SYSTEM SHALL derive a trust tier from `verified:` — absent yields *unverified*, entries with no `human:` actor yield *machine-confirmed*, any `human:` actor yields *human-reviewed*. |
| R-4.4 | THE SYSTEM SHALL treat the trust tier as advisory, computed at read time, never stored, and SHALL NOT allow any gate or access decision to depend on it. |
| R-4.5 | IF a document carries `generated:` or `verified:`, THE SYSTEM SHALL validate it and SHALL preserve both fields byte-identical through `derive`. |
| R-4.6 | THE SYSTEM SHALL cover R-4.5 with an automated Go test written against the mapping and sequence-of-mappings shapes specifically, because `TestRewriteFrontmatterTagsPreservesUnknownKeys` uses a plain string scalar and nested collections re-layout through `PyDumpAutoFlow` (`internal/graph/tags.go:302`). |
| R-4.7 | WHEN a producing command writes a document and the producer is known, THE SYSTEM SHALL emit `generated: {by, at}`. |
| R-4.8 | IF a document omits both fields, THE SYSTEM SHALL validate it unchanged and resolve its tier to *unverified*. |

---

## Unit 5: One-way derivation from existing approval fields

**Why:** without it the trust tier reads *unverified* for every document in every
fixture on the day it ships — a signal with no signal. Derivation makes it
meaningful without touching approval semantics, which is what made full
reconciliation too large to carry here.

| ID | EARS statement |
| --- | --- |
| R-5.1 | WHERE a document carries `decisionOwner:` or `approvedBy:`, THE SYSTEM SHALL derive a `verified:` entry in memory for trust-tier computation. |
| R-5.2 | THE SYSTEM SHALL NOT rewrite, migrate, or remove `decisionOwner:` or `approvedBy:`, and SHALL leave every gate consuming them unchanged. |
| R-5.3 | IF `decisionOwner:` or `approvedBy:` holds a literal `TODO` value, THE SYSTEM SHALL derive no `verified:` entry from it. |
| R-5.4 | WHEN `approvedBy:` holds a literal `TODO` value, THE SYSTEM SHALL emit a warning naming the file, and SHALL NOT fail validation. |
| R-5.5 | THE SYSTEM SHALL NOT introduce any new blocking check (I1). |

---

## Unit 6: Fixtures, compatibility, and verification

**Why:** every requirement here is a veto over the rest. Verified once at the end
against the whole diff rather than assumed per unit.

| ID | EARS statement |
| --- | --- |
| R-6.1 | THE SYSTEM SHALL backfill `description:` onto every document in `examples/workspace/` and `examples/standalone-team/` carrying frontmatter `type:`, satisfying the R-1.2 and R-1.3 rubric. |
| R-6.2 | THE SYSTEM SHALL leave `examples/federated/` untouched, verified by an empty `git diff --stat examples/federated`. |
| R-6.3 | THE SYSTEM SHALL keep `company-os --root examples/federated validate` exiting 0 at `[8/8]` with `workspace.lock.yaml` hashes valid (I9). |
| R-6.4 | THE SYSTEM SHALL leave the `examples/banking/` fixture untouched, as the acceptance harness does not exercise it. |
| R-6.5 | THE SYSTEM SHALL keep any workspace that passed `validate` before this change passing after it, beyond one documented `derive` re-run (I1). |
| R-6.6 | THE SYSTEM SHALL preserve unknown fields and unknown types (I7) and SHALL leave the mandatory/default/guidance tier model untouched (I3). |
| R-6.7 | WHEN the golden snapshots move, THE SYSTEM SHALL re-baseline them in the same commit as the change causing the move, with the diff reviewed line by line, and SHALL NOT run `acceptance.sh --update` to clear a failing build. |
| R-6.8 | THE SYSTEM SHALL commit regenerated `CLAUDE.md` blocks and generated `index.md` files in the same commit as the `description:` backfill, so the harness's double-build check (`s0 == s1 == s2`) stays green. |
| R-6.9 | THE SYSTEM SHALL exit 0 from `make check` — gofmt, `go vet`, `go test ./...`, and `examples/acceptance.sh`. |
| R-6.10 | THE SYSTEM SHALL add an acceptance case running `company-os init` into a scratch directory followed by `company-os validate`, asserting exit 0, because the harness does not cover that path today and it is the exact D-2.2 failure mode R-2.6 exists to prevent. |

---

## Resolved during pre-mortem

**R-2.10 was drafted wrong and is corrected above.** The first draft said a
hand-written `index.md` should cause gate 5 to fail with a named error. That
contradicts a decision this repo already settled for `CLAUDE.md`: a marker-less
file is hand-owned, is left alone, is reported, and **passes**
(`internal/graph/node.go:126-145`, `internal/graph/gates.go:291`). The comment at
`:131-138` records why the alternative was rejected — appending markers silently
converted hand-owned files into managed ones and put `graph build` in direct
conflict with the repair pass. Failing instead would also have introduced a new
blocking check, violating R-5.5 and invariant I1. R-2.10 now follows the
precedent.
