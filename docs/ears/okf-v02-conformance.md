---
type: ears
id: ears-okf-v02-conformance
title: OKF v0.2 Conformance — EARS Specifications
status: locked
tags: [kind/ears, status/locked]
---

# OKF v0.2 Conformance — EARS Specifications

Keywords: `THE SYSTEM SHALL` (always-on), `WHEN` (event), `WHILE` (during a
state), `IF` (conditional/gate), `WHERE` (context-scoped). "The system" = the
`company-os` CLI, its validate gate, and the shipped documentation and fixtures.

**Committed scope:** Units 0–3, plus cross-cutting Units 5–6.
**Deferred (decided, not built):** D1 provenance and trust, D2 per-directory
index generation.

---

## Unit 0: Done-gate date correctness (U0)

**Why:** Invariant #4 of the methodology is "a change is done only when reality is
updated," and the gate enforcing it compares dates as raw strings
(`bin/company-os:679-682`). Writing a normative spec that documents this gate
while the gate misbehaves on malformed input is worse than writing no spec. One
line, no dependents, no snapshot delta — land it first.

| ID | EARS statement |
| --- | --- |
| R-0.1 | WHEN `prd complete` compares a reality document's `updated:` against a PRD's `created:`, THE SYSTEM SHALL parse both values as ISO-8601 calendar dates and compare the parsed dates, never their string forms. |
| R-0.2 | IF either value is absent, empty, or not parseable as an ISO-8601 date, THE SYSTEM SHALL record a done-check error naming the offending file and value, and SHALL NOT raise an unhandled exception. |
| R-0.3 | WHERE both dates are well-formed, THE SYSTEM SHALL produce the same accept/refuse outcome as before this change. |
| R-0.4 | **Amended 2026-07-27 — see [Amendment 1](#amendment-1--r-04-partial-retirement-2026-07-27).** THE SYSTEM SHALL cover R-0.2 with an automated test, and `bash examples/acceptance.sh` SHALL exit 0. The clause naming `examples/selftest.py` as that test's location is retired. |

### Amendment 1 — R-0.4 partial retirement (2026-07-27)

**Authority:** R-9.3 of `docs/ears/go-cli-tui-port.md`, which required the
deletion of `examples/selftest.py`, executed by task 6.5. This amendment records
the consequence for R-0.4 and unblocks task 0.1.

**Statement as originally locked:**

> R-0.4 | THE SYSTEM SHALL cover R-0.2 with a check in `examples/selftest.py`,
> and `bash examples/acceptance.sh` SHALL exit 0.

| Clause | Status | Disposition |
| --- | --- | --- |
| C1 — the R-0.2 check lives in `examples/selftest.py` | **RETIRED** 2026-07-27 | The file does not exist; R-9.3 deleted it along with the Python reference implementation. A requirement naming a deleted file cannot be met, and cannot be failed either — which is worse, because it holds a task open without testing anything. |
| C2 — R-0.2 is covered by an automated test | **IN FORCE — met by NEW work, not by continuity** | Covered in Go by `TestDoneGateNamesAMalformedDate` and `TestParseDate` in `internal/product/product_test.go`. **Corrected 2026-07-27:** an earlier version of this row said the coverage "survives the port." It does not — nothing survived. `python-cli-final:examples/selftest.py` has 86 `check()` sites and **none** of them tests the done-gate date, consistent with task 0.1's own record that the Python side was never fixed. The Go tests are new work delivered under port task 2.6. The outcome is genuinely met; the continuity was not. |
| C3 — `bash examples/acceptance.sh` exits 0 | **IN FORCE — unchanged** | `acceptance.sh` was not deleted and is still the acceptance half of `make check`. |

**Verified on the binary, not inferred from the test names (2026-07-27).** A
reality document carrying `updated: 18/07/2026` — the exact malformed case R-0.2
names — makes `prd complete` refuse with
`cannot compare dates: platforms/web/reality/components/billing.md has
'updated: 18/07/2026', which is not an ISO-8601 date (YYYY-MM-DD)`. File named,
offending value named, no unhandled exception, no silent pass.

**Why this is a retirement and not a quiet reinterpretation.** Task 0.1 recorded
its own exit condition — it stays open "until either the Python side is fixed or
R-9.3 deletes it." R-9.3 deleted it. Closing the task without amending R-0.4
would leave a locked requirement pointing at a file nobody can restore, and the
next reader would have no way to tell whether it had been met, waived, or
forgotten.

---

## Unit 1: Conformance and versioning contract (U1)

**Why:** Company OS has no written definition of "conformant," no methodology
version, and eight fixtures committing `companyOsVersion: '2026.2'` against a
version defined nowhere — two of them with a `profile:` value outside the only
documented enum.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL ship a normative document at `company-os-starter/docs/CONFORMANCE.md` containing, as distinct sections: Goals, Non-Goals, Terminology, a conformance clause, a MUST-NOT-reject list, versioning, and "considered and deferred". |
| R-1.2 | THE SYSTEM SHALL state its conformance clause using RFC-2119 keywords, declaring `type` and an identity field (`id`, or `prd` for outcome reviews, per `bin/company-os:133-135`) as the required floor; `title` and `description` as recommended on all documents; and `resource` as recommended ONLY on documents describing an asset external to the knowledge base, and absent elsewhere. |
| R-1.3 | THE SYSTEM SHALL enumerate what tooling MUST NOT reject a workspace or document for, covering at minimum: unknown `type` values, unknown frontmatter keys, broken cross-links, missing optional documents, and missing federation roots. |
| R-1.4 | **Amended 2026-07-27 — see [Amendment 2](#amendment-2--re-plan-against-the-go-binary-2026-07-27).** WHERE the conformance document lists a MUST-NOT-reject item, THE SYSTEM SHALL cite the implementation site that already honours it. ~~`bin/company-os` file:line~~ — that file was deleted by R-9.3; citations resolve to the Go package, symbol and line instead. |
| R-1.5 | THE SYSTEM SHALL define `2026.2` as the first methodology version carrying a written conformance clause, SHALL designate everything prior as `unversioned`, and SHALL NOT reconstruct a retroactive `2026.1`. |
| R-1.6 | THE SYSTEM SHALL define a forward-only bump rule — the version increments when the conformance clause changes what tooling MUST accept or reject; documentation, fixtures, and new non-blocking fields do not bump — and THIS CHANGE SHALL NOT bump it, so the eight fixtures already committing `2026.2` remain correct unedited. |
| R-1.7 | THE SYSTEM SHALL define the `profile:` enum as `minimal \| standard \| strict`, SHALL correct `docs/01-flexibility-skills-and-role-views.md:121` which currently documents `standard \| strict \| provisional`, and SHALL leave every shipped fixture conformant — including `profile: minimal` at `examples/banking/bank/repos/platform-lending/platform.yaml:4` and `examples/banking/small-company/platforms/product/platform.yaml:4`. |
| R-1.8 | THE SYSTEM SHALL declare which OKF version it targets and SHALL state, per OKF-recommended field, whether Company OS adopts, supersedes, or declines it. |
| R-1.9 | THE SYSTEM SHALL either update or annotate as historical the stale `OKF v0.1` claims at `docs/00-original-proposal.md:3` and `docs/01-flexibility-skills-and-role-views.md:8`. |
| R-1.10 | THE SYSTEM SHALL define, in Terminology, at minimum: *reality*, *deviation*, *exception*, *component*, *canonical*, *authority*, *tier*, *federation root*, *generated artifact*, and *absence tolerance*. |
| R-1.11 | THE SYSTEM SHALL record N1–N8 from the HLD in "considered and deferred", each with one sentence of reasoning. |
| R-1.12 | THE SYSTEM SHALL reserve `generated:` and `verified:` as inert frontmatter field names in `docs/FRONTMATTER-CORE.md`, following the pattern used for reserved doc types (`:113-118`), stating their intended meaning and pointing at the deferred change. |
| R-1.13 | THE SYSTEM SHALL verify every SHOULD in the conformance clause against the CLI's blocking field checks (`core_field_errors:128-145`, gate 3 `:975`, `cmd_prd` validate `:627-630`), and SHALL document any field that blocks anywhere as conditionally required — specifically `title`, which is required today for `type: prd`. |
| R-1.14 | THE SYSTEM SHALL amend `docs/FRONTMATTER-CORE.md:8-12` so that only one document claims to be the interop contract, and SHALL link it to `CONFORMANCE.md`. |
| R-1.15 | **Amended 2026-07-27 — see [Amendment 2](#amendment-2--re-plan-against-the-go-binary-2026-07-27).** IF a document carries `generated:` or `verified:`, THE SYSTEM SHALL pass validation and preserve both fields unchanged through `graph build`, covered by an automated test. ~~a check in `examples/selftest.py`~~ — retired with the file, exactly as R-0.4's C1 was. |
| R-1.16 | THE SYSTEM SHALL produce no CLI diff in this unit; `bash examples/acceptance.sh` SHALL exit 0 with golden snapshots unchanged. |

---

## Unit 2: Roadmap and shipped capability separated (U2)

**Why:** `ONTOLOGY-GUIDE.md` reads as reference documentation for six capabilities
that do not exist in the CLI, and two shipped fixtures assert one of them as fact.
Fixtures are what adopters copy.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL move unshipped ontology material — `ears:` requirement blocks, `@spec` markers, `spec trace`, `validate --ontology`, vocabulary linting, `## Graph` wikilink blocks, and per-clause PRD checklists — out of `docs/ONTOLOGY-GUIDE.md` into `docs/ONTOLOGY-ROADMAP.md`. |
| R-2.2 | THE SYSTEM SHALL open the roadmap document with a not-yet-available banner in the style of `docs/user-guide/explanation/observer-roadmap.md:5-9`, and SHALL link forward to it from `ONTOLOGY-GUIDE.md`. |
| R-2.3 | THE SYSTEM SHALL correct `examples/workspace/company-ontology/contexts/communications.md:20`, which asserts "`company-os validate --ontology` flags forbidden terms in canonical docs" as shipped fact, to the roadmap wording already used at `examples/banking/bank/repos/company-ontology/contexts/payments.md:20`. |
| R-2.4 | WHERE `grep -rn "validate --ontology\|spec trace\|@spec" company-os-starter/ examples/` returns a hit, THE SYSTEM SHALL have that hit inside a document or section carrying the banner, or worded as roadmap. Hits inside `docs/{hld,lld,ears}/` are exempt, as those legitimately discuss the deferred work. |
| R-2.5 | WHERE the same false claim appears inside a read-only federated slice (`examples/federated/company-ontology/contexts/communications.md:20`), THE SYSTEM SHALL correct it at its source and re-sync rather than editing slice bytes, so `workspace.lock.yaml` hashes stay valid (I9). |
| R-2.6 | THE SYSTEM SHALL produce no CLI diff in this unit; `bash examples/acceptance.sh` SHALL exit 0 with golden snapshots unchanged. |

---

## Unit 3: `title` and `resource` (U3)

**Why:** `title` has a consumer today — `build_claude_node` renders
`title or id or filename` (`:1682`), and 12 of 17 fixture documents lack it, so
generated context nodes list filenames. `resource` is OKF-recommended and Company
OS dropped it despite its own founding proposal using it
(`00-original-proposal.md:170`). `description:` is deliberately NOT in this unit —
its only consumer is the deferred index.

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL document `title` and `resource` in `docs/FRONTMATTER-CORE.md` with their tier and their consumer, stating that `title` is non-blocking except where the process contract already requires it, and that `resource` applies only to documents describing an external asset. |
| R-3.2 | THE SYSTEM SHALL backfill `title:` on every document in `examples/workspace/` and `examples/standalone-team/` that lacks one. |
| R-3.3 | THE SYSTEM SHALL treat `resource:` as producer-authored; `graph build` SHALL NOT write or rewrite it, and no validate gate SHALL check it. |
| R-3.4 | THE SYSTEM SHALL add `resource: component://<component-id>` to every `type: component-reality` document in `examples/workspace/` and `examples/standalone-team/` ONLY. Federated slices SHALL NOT be edited (I9). |
| R-3.5 | WHEN a user runs `prd new`, `discover new`, or `reality new`, THE SYSTEM SHALL emit a document carrying `title:`, with `reality new` additionally emitting `resource:`, and with `templates/*.md` and the CLI's `*_TEMPLATE` strings in sync. |
| R-3.6 | WHEN `prd complete` writes `outcome.md` (`bin/company-os:701-705`), THE SYSTEM SHALL emit `title:` on it, SHALL keep its identity field as `prd:` per `core_field_errors:132-135`, and SHALL NOT assume `id:` exists in any title-fallback path. |
| R-3.7 | WHEN `title:` is added to fixture documents, THE SYSTEM SHALL regenerate and commit the affected `CLAUDE.md` generated blocks in the same change, so the acceptance harness's double-build check (`s0 == s1 == s2`) stays green. |
| R-3.8 | THE SYSTEM SHALL leave `examples/golden-validate.txt` and `examples/federated-golden-validate.txt` unchanged by this unit, because validate's stdout carries document paths and per-root status but never a `title`. IF this unit produces a golden diff, THE SYSTEM SHALL treat it as an unintended regression and SHALL NOT re-baseline. |
| R-3.9 | IF a document omits `title` or `resource`, THE SYSTEM SHALL emit neither an error nor a warning beyond the blocking checks that exist today. |

---

## Unit 5: Backward compatibility (cross-cutting)

**Why:** The adopter running `validate` in CI has no pain today. This unit exists
so that stays true. Every requirement here is a veto over the others.

| ID | EARS statement |
| --- | --- |
| R-5.1 | IF a workspace passed `company-os validate` before this change, THE SYSTEM SHALL exit 0 on that workspace after this change, with no manual migration other than the documented `graph build` re-run of R-3.7. |
| R-5.2 | THE SYSTEM SHALL introduce no new blocking check on any field this change adds or documents; `title`, `resource`, `generated`, and `verified` SHALL introduce no new blocking behaviour. |
| R-5.3 | THE SYSTEM SHALL introduce no new warn line into `validate` output. |
| R-5.4 | THE SYSTEM SHALL preserve unknown frontmatter fields and tolerate unknown `type` values, unchanged. |
| R-5.5 | THE SYSTEM SHALL keep the mandatory / default / guidance tier model untouched; nothing introduced here SHALL be mandatory-tier. |
| R-5.6 | THE SYSTEM SHALL leave both golden snapshots byte-identical across this entire change; `examples/acceptance.sh --update` SHALL NOT be run. |
| R-5.7 | THE SYSTEM SHALL NOT write into, or invalidate the lock hashes of, any materialized federated slice (I9). |

---

## Unit 6: Fixture coverage (cross-cutting)

**Why:** `examples/banking/` is the largest fixture (38 markdown files) and the
acceptance harness does not exercise it. Editing it without an oracle is editing
untested content.

| ID | EARS statement |
| --- | --- |
| R-6.1 | THE SYSTEM SHALL exit 0 on `bash examples/acceptance.sh` at the end of every unit, not only at the end of the change. |
| R-6.2 | IF `examples/banking/` is edited beyond the `profile:` enum reconciliation of R-1.7, THE SYSTEM SHALL first add it to the acceptance harness's validate-exit-code fixture loop. |
| R-6.3 | IF adding `examples/banking/` to the harness proves larger than this change can absorb, THE SYSTEM SHALL leave that fixture untouched and SHALL record in `CONFORMANCE.md` that it is a prior-version fixture not yet backfilled. |

---

## Amendment 2 — Re-plan against the Go binary (2026-07-27)

**Authority:** R-9.8 of `docs/ears/go-cli-tui-port.md` — "THE SYSTEM SHALL leave
OKF v0.2 Phases 1–3 unimplemented, to be re-planned against the Go binary as a
separate change." This is that re-plan. It changes no committed outcome; it
re-points the evidence those outcomes are written against.

This spec was locked while `company-os-starter/bin/company-os` was the
implementation. R-9.3 deleted that file. Every `bin/company-os:NNN` citation in
this document therefore names a line nobody can open, and two requirements
(R-1.4, R-1.15) named it *normatively* — they could not be satisfied as written.

**How the anchors were resolved — corrected 2026-07-27 after review.** The first
version of this amendment claimed each anchor was resolved "by following the
comment the port left behind." That is only true of some rows. The Go comments
carry Python line references, but those references and this spec's disagree by up
to 57 lines, and in at least two cases both are stale — `:1682` for
`build_claude_node` lands inside `iter_knowledge_docs` in the tagged Python
source, and the Go citation `:1346-1392` for `iter_graph_docs` starts inside
`rewrite_frontmatter_tags`. Those rows were resolved by matching the quoted
expression or the symbol, not the line.

**The authoritative source is available and was used for the correction.** R-9.2
required tagging the final Python commit; the tag `python-cli-final` exists and
`git show python-cli-final:company-os-starter/bin/company-os` returns all 2781
lines. An earlier version of this amendment stated that "the file that would
settle it is deleted." That was **wrong** — the spec family made this settleable
on purpose, and the amendment forgot its own recovery path. Every line number
below marked *(settled)* was read from that tag.

### Anchor map — every `bin/company-os` citation in this spec

| Cited in | Python anchor | Go site today |
| --- | --- | --- |
| R-0.1, R-0.3 (Unit 0 Why) | `:679-682` done-gate date compare | `internal/product/prd.go:466` and `parseDate` — the comment there quotes the original string comparison |
| R-1.2 | `:133-135` identity field (`id`, or `prd`) | `internal/product/contract.go:70` — the `!truthy(meta,"id") && !truthy(meta,"prd")` branch of `CoreFieldErrors` |
| R-1.13 | `core_field_errors:128-145` | `internal/product/contract.go:65` `CoreFieldErrors` |
| R-1.13 | gate 3 `:975` | `internal/product/check.go:36` `contractFields` — the four-field list the gate fails on. (`:89` `Gate` is the enclosing function; the check itself is 53 lines above it.) |
| R-1.13 | `cmd_prd` validate `:627-630` — **the Go source cites `:627-628`**; a third stale-citation discrepancy, flagged here rather than left for task 1.1 to rediscover | `internal/product/prd.go:22-25` `processFields`. **Deliberately asymmetric with gate 3**: six fields vs four, because a workspace sweep does not demand `decisionOwner` or the `platform` of a document whose location states it (`check.go:33-35`). |
| R-1.13 (format checks) | `:171-175` `format_checks_enforced` | `internal/product/contract.go:99` `FormatChecksEnforced` |
| R-3.6 | `:701-705` `outcome.md` writer | `internal/product/prd.go:361` (call) and `:569-574` `outcomeDoc` (content) |
| Unit 3 Why | `:1682` `build_claude_node` title fallback — **stale; the function begins at `:1709` (settled)** and `:1682` is inside `iter_knowledge_docs`, a different title-derivation site | `internal/graph/node.go:319-324` — resolved by matching the quoted expression `meta.get("title") or meta.get("id") or Path(rel).name`, not by line |
| D-2.2 | `cmd_graph:1713` → **`:1770` (settled)**; `rebuild_generated:1746` → **`:1803` (settled)** | **Two distinct symbols, not one.** `cmd_graph` is `internal/graph/graph.go:16` `Build`; `rebuild_generated` is `internal/graph/graph.go:59` `Rebuild`. D-2.2 is specifically about generating from *both* paths, so an earlier version of this row — which mapped both names onto `Rebuild` — erased the requirement's subject. Both of this spec's line numbers are stale by 57; the Go comment's `:1803` is correct. |
| D-2.5 | `iter_graph_docs` skip-list — **`:1361` (settled)**; the Go source's `:1346-1392` range starts inside `rewrite_frontmatter_tags` | `internal/graph/tags.go:186` `IterGraphDocs` |

### Clause dispositions

| Clause | Status | Disposition |
| --- | --- | --- |
| R-1.4 — cite a `bin/company-os` file:line | **AMENDED** | The outcome — every MUST-NOT-reject item is backed by a citable implementation site rather than an assertion — is unchanged and is the whole point of the requirement. Only the file it names changes. |
| R-1.15 — the preserve check lives in `examples/selftest.py` | **RETIRED, and the restated obligation is UNMET** | The file was deleted by R-9.3, as with R-0.4's C1. **But the outcomes differ and an earlier version of this row wrongly called them identical.** R-0.4's obligation is met by a real Go test; R-1.15's is not. `python-cli-final:examples/selftest.py` never carried a `generated:`/`verified:` check, and `verified` appears **nowhere** in Go source or tests today. What exists is generic — `TestRewriteFrontmatterTagsPreservesUnknownKeys` — and it uses a plain string scalar, whereas D-1.1 defines `generated:` as a mapping `{by, at}` and `verified:` as a sequence of mappings `[{by, at}]`. Nested collections round-trip through `PyDumpAutoFlow` (`internal/graph/tags.go:302`), which is exactly where R-0.7a(g)'s flow/block re-layout divergence lives — so the untested case is the case R-1.15 is about. **This is new work in Phase 1, not a preserved guarantee.** Task 1.6 carries it. |
| Every other `bin/company-os:NNN` citation | **RE-POINTED, not amended** | These appear in *Why* prose and acceptance evidence, not in normative statements. The anchor map above is the resolution; no requirement text changes. |

### R-1.15 vs R-1.16 — a contradiction this amendment creates

R-1.15 as amended requires an automated test. R-1.16 requires that this unit
"produce no CLI diff," which the re-plan defines as an empty diff of
`company-os-starter/internal` and `cmd`. A Go test for graph-build field
preservation lands in `internal/`. **Both cannot hold as literally written.**

**Ruling:** R-1.16's subject is *shipped CLI behaviour*, not the repository tree —
it exists so this unit cannot change what the tool does, which is why its own
statement pairs it with "`acceptance.sh` SHALL exit 0 with golden snapshots
unchanged." A test adds no behaviour and moves no golden. R-1.16 is therefore not
engaged by test-only additions, and `_test.go` files are outside its subject.

Recorded here rather than in a task's verify line, where it first lived: a
carve-out to a locked requirement is a requirement change, and this project's
arrow of intent puts those in the spec.

### Re-measured facts (2026-07-27)

The spec's *Why* sections quote counts taken while the Python CLI shipped. Three
moved, and the tasks planned against them are sized wrong until restated:

| Claim as locked | Measured today | Effect |
| --- | --- | --- |
| Unit 3: "12 of 17 fixture documents lack `title`" | **14 of 16** documents carrying frontmatter `type:` in `examples/workspace/` + `examples/standalone-team/` lack it | R-3.2's backfill is larger, not smaller — the port added `syncing-knowledge.SKILL.md` and other typed documents |
| R-1.7: `profile:` enum | **Unchanged and still wrong.** 6 fixtures `standard`, 2 `minimal`; `docs/01-flexibility-skills-and-role-views.md:121` still documents `standard \| strict \| provisional` | R-1.7 stands exactly as written — `minimal` used but undocumented, `provisional` documented but unused, `strict` neither |
| R-1.6: "eight fixtures commit `2026.2`" | **8.** Unchanged | R-1.6 stands |
| R-6.2: `examples/banking/` is 38 markdown files | **38.** Unchanged | R-6.2/R-6.3 stand |

### What the port added that this spec predates

R-1.1's conformance document was scoped before the CLI had a machine-readable
surface or an exit-code contract. Both now exist and belong in it:

- **Exit codes.** `docs/ears/go-cli-tui-port.md` Unit 4 defines a full contract
  (exit 2 usage, 3 workspace, 7 environment, 8 refuse-to-overwrite). A conformance
  document that tells an adopter what tooling MUST NOT reject, while saying
  nothing about what the tool *returns* when it does reject, is half a contract.
- **`--json`.** Unit 3 made every gate result machine-readable. The MUST-NOT-reject
  list (R-1.3) should state that it binds the JSON surface identically, or it will
  be read as applying to human output only.
- **`knowledge/`.** The synced catalog root did not exist when Unit 1 was written.
  It is a node root but not a graph-docs root, and R-1.10's Terminology owes it a
  definition alongside *federation root*.
- **A code-side producer of `conformance:` now exists, and it lands inside Unit 1's
  own subject.** `scaffoldPlatform` (`internal/scaffold/scaffold.go:196-197`)
  writes `companyOsVersion: "2026.2"` and `profile: standard` into every new
  platform. R-1.5, R-1.6 and R-1.7 were written when those values existed only in
  committed fixtures; the `profile:` enum now has a **third** sync point —
  `docs/01-flexibility-skills-and-role-views.md:121`, the fixtures, and this
  emitter — and task 1.3's targets name only the first two. Harmless today, since
  `standard` is in the corrected enum, which is exactly why it would be missed.

These are additions to R-1.1's section list, not new units. They are recorded here
rather than legislated, because adding requirements to a locked spec is a change
this amendment has no authority to make — R-9.8 authorized a re-plan, not a
re-scope.

---

## Deferred — decided, not built

### Unit D1: Provenance and trust

**Why deferred:** the largest remaining gap and the intended next change. Not a
metadata addition — Company OS encodes sign-off today in three incompatible shapes
that are load-bearing in blocking gates: `decisionOwner` on PRDs (literal `TODO`
hard-fails, `:627-630`), `approvedBy` on deviations, and `approvedBy` on exceptions
where `'TODO: rule owner'` passes validation
(`examples/workspace/teams/customer-engagement/governance/exceptions.yaml:8`).
Unifying them changes approval semantics. U1 reserves the names so it lands cleanly.

| ID | EARS statement (designed, not implemented) |
| --- | --- |
| D-1.1 | THE SYSTEM WOULD record content origin as `generated: {by, at}` and confirmation as `verified: [{by, at}]`, kept distinct because a document's author need not be its confirmer. |
| D-1.2 | THE SYSTEM WOULD identify actors as `<producer>/<version>` for agents, `human:<id>` for people, and `process:<id>` for automated processes. |
| D-1.3 | THE SYSTEM WOULD derive a trust tier from `verified` — absent ⇒ unverified, non-human actors only ⇒ machine-confirmed, any `human:` actor ⇒ human-reviewed — as an advisory signal, never access control. |
| D-1.4 | THE SYSTEM WOULD reconcile `decisionOwner` and the two `approvedBy` fields against this model, including the exception-vs-PRD `TODO` asymmetry. |

### Unit D2: `description` and per-directory `index.md`

**Why deferred:** cut from this change after technical and product review. Two
blockers and one dependency — the federation collision (D-2.1), the wiring gap
(D-2.2), and the fact that `description`'s only consumer is the index, so shipping
it alone would create a field with no consumer. Its quality also cannot be reviewed
without a rendered index: the test for a bad description is whether it still reads
true pasted onto a sibling, and that only works side by side.

| ID | EARS statement (designed, not implemented) |
| --- | --- |
| D-2.1 | THE SYSTEM WOULD skip any path under a manifest-declared slice root when generating or drift-checking indexes, since `examples/federated/` materializes two graph documents inside a read-only slice and any write there breaks gate 8's lock hashes. |
| D-2.2 | THE SYSTEM WOULD generate indexes from every derived-artifact path — `cmd_graph:1713` **and** `rebuild_generated:1746` (call sites `:713`, `:1925`, `:1952`, `:1959`, `:1970`, `:1993`) — so `company-os init` does not produce a workspace failing its own validate, which `examples/selftest.py` already asserts against. |
| D-2.3 | THE SYSTEM WOULD generate an index only in a directory directly holding **2 or more** graph documents, yielding three indexes in `examples/workspace/` and zero in `examples/standalone-team/`. |
| D-2.4 | THE SYSTEM WOULD NOT exclude `archive/prds/<id>/`; at exactly two documents it qualifies, and with R-3.6's titled `outcome.md` its index surfaces the 90-day outcome obligation that is otherwise invisible. |
| D-2.5 | THE SYSTEM WOULD add `index.md` to the `iter_graph_docs` skip-list in the same atomic commit as the generator, never after, so generated indexes are never re-ingested as graph documents. |
| D-2.6 | THE SYSTEM WOULD perform the drift check inside gate 5 with its printed header byte-identical to `examples/golden-validate.txt:28`, reporting one aggregate line per federation root, adding no gate and renumbering none. |
| D-2.7 | THE SYSTEM WOULD require a `description:` rubric: a description must carry at least one fact not derivable from the filename, `title:`, `type:`, or directory path, and must NOT read as true when pasted onto a sibling document. Its template placeholder must be quoted, since a useful description usually contains a colon and would otherwise be a YAML parse error. |
| D-2.8 | Open for that spec: whether `index.md` carries frontmatter at all and if so its `type`; the semantics of a qualifying directory with no index; how a pre-existing hand-written `index.md` is treated; and index removal when a directory drops below the threshold. |
