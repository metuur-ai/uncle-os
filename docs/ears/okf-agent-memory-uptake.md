---
type: ears
id: ears-okf-agent-memory-uptake
title: OKF Agent-Memory Uptake — EARS Specifications
status: draft
---

# OKF Agent-Memory Uptake — EARS Specifications

Keywords: `THE SYSTEM SHALL` (always-on), `WHEN` (event), `WHILE` (during a
state), `IF` (conditional/gate), `WHERE` (context-scoped). "The system" = the
`company-os` CLI, its validate gate, and the shipped documentation and fixtures.

Requirement IDs are `R-<unit>.<n>`. Code binds with
`@spec req://uncle-os/okf-agent-memory-uptake@0.1#R-x.y`.

**Committed scope:** Units 1–7 + cross-cutting Unit 8. **Declined (recorded, not
built):** T8 MCP shim.

---

## Unit 1: Provenance consistency (gate 2, warn)

**Why:** `generated:` and `verified:` shipped as parsed-but-unchecked fields. A
document can claim a verification dated before its own generation and pass. The
trust tier stays advisory; these checks read shapes and dates, never the tier.

| ID | EARS statement |
| --- | --- |
| R-1.1 | WHEN `validate` runs gate 2, THE SYSTEM SHALL walk every graph document (`IterGraphDocs`) after the deviation/exception findings and emit provenance findings, without changing gate 2's ordinal, slug or title. |
| R-1.2 | IF `generated:` is a mapping whose `by` is absent or empty, THE SYSTEM SHALL emit one `SevWarn` finding `expiry.provenance-generated-no-by` with field `path`. |
| R-1.3 | IF `generated:` is present and not a mapping, THE SYSTEM SHALL emit one `SevWarn` finding `expiry.provenance-generated-not-mapping` with field `path`. |
| R-1.4 | IF `generated.at` and `verified[i].at` both parse as `YYYY-MM-DD` or RFC3339 and `verified[i].at` is earlier than `generated.at`, THE SYSTEM SHALL emit one `SevWarn` finding `expiry.provenance-superseded-verification` per such entry with fields `path`, `index`, `verifiedAt`, `generatedAt`. |
| R-1.5 | IF an actor string in `generated.by` or any `verified[i].by` does not match `^(?:[a-zA-Z][\w.-]*:\S+\|[^\s/]+/[^\s/]+)$`, THE SYSTEM SHALL emit one `SevWarn` finding `expiry.provenance-actor-format` with fields `path`, `actor`. |
| R-1.6 | IF an actor string has the form `<prefix>:<rest>` and `<prefix>` is neither `human` nor `process`, THE SYSTEM SHALL emit one `SevWarn` finding `expiry.provenance-actor-prefix` with fields `path`, `actor`, `prefix`. |
| R-1.7 | THE SYSTEM SHALL parse provenance dates with a dedicated parser and SHALL NOT modify the `before()` parser used for `reviewDate`/`expires`. |
| R-1.8 | THE SYSTEM SHALL NOT emit any finding for a `generated.at` or `verified[i].at` value that fails to parse; an unparseable date affects only whether R-1.4 can be evaluated. |
| R-1.9 | THE SYSTEM SHALL NOT emit per-document `SevOK` findings for Unit 1 checks. |
| R-1.10 | THE SYSTEM SHALL NOT read `TrustTier()` in any Unit 1 check. |
| R-1.11 | IF the document walk returns an error, THE SYSTEM SHALL NOT abort gate 2; it SHALL emit no finding for the error and leave gate 4 to report it as today. |
| R-1.12 | WHEN the shipped fixtures are validated after this unit, THE SYSTEM SHALL produce output byte-identical to all five goldens: `golden-validate.txt`, `federated-golden-validate.txt`, `failing-workspace-golden-validate.txt`, `failing-federated-golden-validate.txt`, `failing-federated-nolock-golden-validate.txt`. |
| R-1.13 | THE SYSTEM SHALL begin each Unit 1 finding's rendered message with `provenance:` so the text output is self-labelling under the gate 2 header. |

## Unit 2: `stale_after` freshness signal

**Why:** After a PRD completes, its outcome and reality documents carry no signal
for when they should be re-read. Deferred as N4 by the parent; the user chose
gate 2 as its home.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL document `stale_after: YYYY-MM-DD` in `docs/FRONTMATTER-CORE.md` under Tier 2 as an optional field on any typed document, defined as "on or after this date the document should be re-read; nothing is deleted or archived because of it; clear the signal by advancing or removing the field", and SHALL amend the Tier 2 heading to "required per doc family unless marked optional". |
| R-2.2 | WHEN gate 2 walks a document whose `stale_after` parses and is less than or equal to today, THE SYSTEM SHALL emit one finding `expiry.stale` with fields `path`, `staleAfter`, `today`. |
| R-2.3 | WHILE `validate` runs without `--stale`, THE SYSTEM SHALL emit `expiry.stale` at `SevWarn`. |
| R-2.4 | WHILE `validate` runs with `--stale`, THE SYSTEM SHALL emit `expiry.stale` at `SevFail` and the exit code SHALL reflect it. |
| R-2.5 | IF `stale_after` is present and does not parse, THE SYSTEM SHALL emit one `SevWarn` finding `expiry.stale-after-unparseable` with fields `path`, `value`, regardless of `--stale`. |
| R-2.6 | THE SYSTEM SHALL accept `--stale` together with `--fix`; `FixDerived` runs first, then gates run with R-2.4 in force. |
| R-2.7 | THE SYSTEM SHALL expose `validate.RunWith(ws, RunOptions{Stale bool})` and keep `validate.Run(ws)` as `RunWith(ws, RunOptions{})`, so existing callers compile unchanged. |
| R-2.8 | WHEN `prd complete` writes `archive/prds/<id>/outcome.md`, THE SYSTEM SHALL include `stale_after: <due>` directly under the existing `due:` line, with the same value. |
| R-2.9 | THE SYSTEM SHALL update `templates/outcome-review.md` to carry `stale_after: <YYYY-MM-DD>` and a `generated:` placeholder block, so the template and `outcomeDoc()` emit the same field set. |
| R-2.10 | THE SYSTEM SHALL NOT delete, move, archive or rewrite any document because its `stale_after` has passed. |
| R-2.11 | THE SYSTEM SHALL NOT let `stale_after` or `--stale` influence `check ready`, `check done` or `prd complete`. |
| R-2.12 | THE SYSTEM SHALL NOT introduce an injectable clock; tests SHALL use far-dated fixture values. |

## Unit 3: Log discipline

**Why:** Two of eleven mutating commands write `log.md`. The original proposal
asked for an append-only knowledge log; OKF appends on every mutation. The
`- date:` bullet shape is kept (§9 Q2); status transitions log too (user,
2026-09-08).

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL provide `internal/logbook` importing only `model`, with `Append(path, date, text)` that opens `path` with `O_APPEND\|O_CREATE\|O_WRONLY` and writes exactly `- <date>: <text>\n`, and `Line(fields)` that renders the `CodeLogAppended` sentence. |
| R-3.2 | THE SYSTEM SHALL route the existing `prd complete` and `prd promote` log writes through `logbook.Append` with their line text unchanged, so `examples/workspace/platforms/communications/log.md` and existing tests keep their bytes. |
| R-3.3 | WHEN `deviation declare` succeeds, THE SYSTEM SHALL append to `teams/<t>/log.md`: `deviation declared for rule \`<rule>\`; review <reviewDate>`. |
| R-3.4 | WHEN `exception request` succeeds, THE SYSTEM SHALL append to `teams/<t>/log.md`: `exception requested for rule \`<rule>\` on \`<component>\`; expires <expires>`. |
| R-3.5 | WHEN `discover new` succeeds, THE SYSTEM SHALL append to `teams/<t>/log.md`: `discovery brief \`<id>\` created`. |
| R-3.6 | WHEN `discover validate` succeeds, THE SYSTEM SHALL append to `teams/<t>/log.md`: `discovery brief \`<id>\` validated`. |
| R-3.7 | WHEN `prd new` succeeds, THE SYSTEM SHALL append to `platforms/<p>/log.md`: `PRD \`<id>\` proposed`, suffixed ` (from discovery <brief>)` when `--from-discovery` was given. |
| R-3.8 | WHEN `prd new --draft` succeeds, THE SYSTEM SHALL append to `teams/<t>/log.md`: `PRD draft \`<id>\` created`, with the R-3.7 suffix rule. |
| R-3.9 | WHEN `prd abandon` succeeds, THE SYSTEM SHALL append to `teams/<t>/log.md`: `PRD draft \`<id>\` abandoned`. |
| R-3.10 | WHEN `add platform`, `add team` or `add component` succeeds, THE SYSTEM SHALL append to the created or owning root's `log.md`: `<kind> \`<name>\` added`. |
| R-3.11 | WHEN `reality new` succeeds, THE SYSTEM SHALL append to `platforms/<p>/log.md`: `reality doc for \`<component>\` scaffolded`. |
| R-3.12 | WHEN `workspace sync` succeeds, THE SYSTEM SHALL append to `<root>/log.md`: `workspace synced: <n> repo(s), lock updated`. |
| R-3.13 | THE SYSTEM SHALL NOT append to any `log.md` from `init`, `scratchpad init`, `governance resolve`, `graph build`/`derive`, `validate --fix` or `skills install`; `docs/FRONTMATTER-CORE.md` SHALL state that derivation is not logged. |
| R-3.14 | WHEN a `GateResult`-returning command appends a log line, THE SYSTEM SHALL emit one `CodeLogAppended` finding with field `path` beside its next-step guidance, without dropping or reordering that guidance. |
| R-3.15 | WHEN `deviation declare` or `exception request` appends a log line, THE SYSTEM SHALL expose the path as `Log` on `Declared`/`Requested`, render it as one text line, and emit it as a `"log"` key in `--json`. |
| R-3.16 | WHEN gate 4 runs, THE SYSTEM SHALL read `log.md` under `<root>` and every `NodeRoots` root as text and, for each non-blank line not matching `^- \d{4}-\d{2}-\d{2}: \S`, emit one `SevWarn` finding `frontmatter.log-line-format` with fields `path`, `line`, `text` (≤ 80 characters). |
| R-3.17 | THE SYSTEM SHALL keep `log.md` in `skipNames`. |
| R-3.18 | THE SYSTEM SHALL NOT rewrite, reorder, deduplicate or delete lines in any `log.md`. |

## Unit 4: The agent contract

**Why:** A fresh workspace has no file telling an agent how to behave; existing
workspaces have no way to get one. OKF `bootstrap` writes `AGENTS.md`; Claude
Code reads `CLAUDE.md`; `skills install` is the precedent for writing embedded
canonical text into a workspace.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL provide `scaffold.WriteContract(root)` that, per file (`AGENTS.md`, `CLAUDE.md`): IF absent, writes it; IF present and containing `<!-- company-os:agent-contract:start -->`, leaves it byte-for-byte unchanged; IF present without the marker, appends a newline and the delimited block; and SHALL NEVER truncate. |
| R-4.2 | `AGENTS.md`'s block SHALL contain the Minimal Agent Contract embedded via `go:embed` from `internal/scaffold/contract_agents.md`, with nine rules covering: search before writing (`company-os find`); never edit `generated/`, generated markers or `knowledge/`; never hand-write `tags:` (`company-os derive`); run `validate` before claiming done and the reality-updated rule of `prd complete`; follow the printed next command; record decisions in artifacts; never persist secrets/PII (link `docs/SECURITY.md`); agents never write `human:` verifications; where skills live. |
| R-4.3 | `CLAUDE.md`'s block SHALL contain exactly the import line `@AGENTS.md` and one comment stating that Claude Code resolves the import, other runtimes read `AGENTS.md`, and content below the block is hand-owned. |
| R-4.4 | WHEN `init` has moved the four roots into the target, THE SYSTEM SHALL call `WriteContract(target)` as a subsequent step in `Init`. |
| R-4.5 | IF `WriteContract` fails during `init`, THE SYSTEM SHALL NOT fail `init`; it SHALL emit one `SevWarn` finding `scaffold.contract-not-written` with fields `path`, `error` naming `company-os skills install` as the repair. |
| R-4.6 | WHEN `skills install` runs, THE SYSTEM SHALL call `WriteContract(ws.Root)` after writing skills and before the rebuild, and SHALL report per file `written`, `appended` or `unchanged`. |
| R-4.7 | THE SYSTEM SHALL NOT version or compare the contract text; a present start marker means the file is left alone. |
| R-4.8 | THE SYSTEM SHALL add `AGENTS.md` to `skipNames` so an `AGENTS.md` placed inside a node root is never evaluated as a typed document (the workspace root is already outside the walk). |
| R-4.9 | THE SYSTEM SHALL NOT add `<root>` to `NodeRoots`; `graph build` and gate 5 SHALL ignore both root files. |
| R-4.10 | THE SYSTEM SHALL NOT write or modify `AGENTS.md`/`CLAUDE.md` from any command other than `init` and `skills install`. |
| R-4.11 | WHEN `init` is followed by two `graph build` runs, THE SYSTEM SHALL produce an identical tree hash with both root files present. |

## Unit 5: Agent-behaviour scenario tests

**Why:** `acceptance.sh` proves determinism, not that an agent following the
skills reaches `prd complete` from a fresh clone. OKF's definition of done is
adopted as the pass criterion.

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL ship `cmd/company-os/scenarios_test.go` with `TestScenario<Name>` functions built on `runArgs`, reading `next:`/`fix:` commands from `--json` fields, run by `go test ./...`. |
| R-5.2 | TC-S1 FreshWorkspaceContract: WHEN `init` runs in an empty directory, `AGENTS.md` SHALL contain both delimiters and an anchor phrase from each of the nine R-4.2 rules, `CLAUDE.md` SHALL contain `@AGENTS.md`, and `validate` SHALL exit 0. |
| R-5.3 | TC-S2 SkillChainFromInit: WHEN the chain `init → add component → discover new → discover validate → prd new --from-discovery --components → prd validate → prd complete → reality new → prd complete → validate` runs, every step joined by a printed `next:` SHALL be driven by that value; the first `prd complete` SHALL exit non-zero with the done-check and print `fix: company-os reality new …`, which the test executes verbatim; the second `prd complete` SHALL exit 0; the final `validate` SHALL be `PASS`; `outcome.md` SHALL carry `stale_after` equal to `due`. |
| R-5.4 | TC-S3 AntiDuplication: WHEN `discover new` runs twice with the same title, the second call SHALL exit `ExitConflict`, its error SHALL name the existing path, the first `brief.md` SHALL be byte-identical, and no other file SHALL be created. |
| R-5.5 | TC-S4 HumanCorrectionPreserved: WHEN a PRD carrying `decisionOwner: <name>` is completed, the archived PRD SHALL retain that field and `generated:` SHALL appear only in `outcome.md`. |
| R-5.6 | TC-S5 LogTrail: WHEN the TC-S2 chain runs, each mutating step SHALL have appended exactly one `- YYYY-MM-DD: ` line to the `log.md` named in Unit 3 (the failing `prd complete` appends none), and gate 4 SHALL emit zero `frontmatter.log-line-format` findings. |
| R-5.7 | TC-S6 StaleAndProvenanceSignals: WHEN a fixture carries `stale_after: 2000-01-01` and `verified[0].at < generated.at`, `validate` SHALL exit 0 with exactly one `expiry.stale` warn and one `expiry.provenance-superseded-verification` warn; `validate --stale` SHALL exit non-zero with `expiry.stale` at `fail` and no other difference. |
| R-5.8 | THE SYSTEM SHALL ship `docs/AGENT_TESTING.md` listing TC-S1…TC-S6 with pass criteria, the definition of done "a completely new agent can continue the workspace using the repository alone", which TC-S2 steps are skill-driven rather than `next:`-driven, and the rule that a new scenario requires a row. |
| R-5.9 | THE SYSTEM SHALL NOT involve any LLM in the scenario suite. |

## Unit 6: `docs/SECURITY.md`

| ID | EARS statement |
| --- | --- |
| R-6.1 | THE SYSTEM SHALL ship `company-os-starter/docs/SECURITY.md` with sections: never-persist table (secrets/tokens, credentials, PII, production payloads, session transcripts, chain-of-thought; covering `reality/`, `product/discovery/`, `change-records/`, `scratchpad/`); knowledge integrity (`0444`/`0555`, hash-locked, gate 9, fix = upstream + `workspace sync`); poisoning (untraceable content stays in `scratchpad/`; no `sources[]`); audit trail (`log.md` append-only + `git log`; gate 4 line-format warn); deletion (note in `log.md`; history rewrite only for leaked secrets). |
| R-6.2 | THE SYSTEM SHALL link `docs/SECURITY.md` from contract rule 7 and from `README.md`'s documentation list. |

## Unit 7: Version handshake and enum reconciliation

**Why:** Nine fixtures commit `companyOsVersion: "2026.2"` against a version
defined nowhere; `profile` has three contradictory definitions. Completes parent
task 1.3.

| ID | EARS statement |
| --- | --- |
| R-7.1 | THE SYSTEM SHALL define in `CONFORMANCE.md §6` the `companyOsVersion` scheme `YYYY.N` compared numerically (year, then N), declare `2026.2` as the first and only value, and restate the forward-only bump rule (parent R-1.6). |
| R-7.2 | THE SYSTEM SHALL define in `CONFORMANCE.md §6` the `profile` enum as exactly `minimal \| standard \| strict` (parent R-1.7), with `standard` = all gates, `minimal` = gates 1–4, `strict` = reserved, and SHALL state that gating by profile is not implemented (parent task 4.2). |
| R-7.3 | THE SYSTEM SHALL correct `docs/01-flexibility-skills-and-role-views.md:121` to `# minimal \| standard \| strict`, dropping `provisional`. |
| R-7.4 | THE SYSTEM SHALL define the compatibility rule: a validator built for `V` accepts declared versions `≤ V` silently and warns on `> V` or unparseable; legacy-switch behaviour applies only once a second version exists. |
| R-7.5 | WHEN gate 4 runs, for each platform whose `platform.yaml` declares `conformance.companyOsVersion` that is unparseable or greater than the built-in `V`, THE SYSTEM SHALL emit one `SevWarn` finding `frontmatter.conformance-version` with fields `path`, `declared`, `supported`. |
| R-7.6 | WHEN gate 4 runs, for each platform whose `platform.yaml` declares `conformance.profile` outside the enum, THE SYSTEM SHALL emit one `SevWarn` finding `frontmatter.conformance-profile` with fields `path`, `declared`. |
| R-7.7 | IF `platform.yaml` is absent for a platform, or present without a `conformance` block, THE SYSTEM SHALL emit no finding and no error. |
| R-7.8 | THE SYSTEM SHALL NOT change which gates run based on `profile` in this change. |
| R-7.9 | THE SYSTEM SHALL add a note in `CONFORMANCE.md §6` that `log.md` uses dated bullets and the OKF §9 heading form is not used. |
| R-7.10 | THE SYSTEM SHALL add an "Update 2026-09-08 — completed by `okf-agent-memory-uptake` Unit 7" note under task 1.3 of `docs/tasks/okf-v02-conformance.md`. |

## Unit 8: Cross-cutting

| ID | EARS statement |
| --- | --- |
| R-8.1 | THE SYSTEM SHALL register every new finding code (`expiry.provenance-*`, `expiry.stale*`, `frontmatter.log-line-format`, `frontmatter.conformance-*`, `scaffold.contract-not-written`) in `internal/model/codes.go` under its host gate's block, and render each through exactly one `case` per renderer that can emit it, with `--json` carrying every field the sentence shows. |
| R-8.2 | THE SYSTEM SHALL keep gates `[1/8]…[8/8]` and the conditional `[9]` at their current ordinals, slugs and titles. |
| R-8.3 | THE SYSTEM SHALL introduce no new default-`SevFail` finding; `--stale` is the only promotion path. |
| R-8.4 | THE SYSTEM SHALL add no runtime dependency beyond the Go standard library. |
| R-8.5 | WHEN `examples/acceptance.sh` runs after each slice, THE SYSTEM SHALL pass with every golden unchanged; `--update` SHALL NOT be used. |
| R-8.6 | THE SYSTEM SHALL NOT call `os.Exit` or print below `cmd/`. |
| R-8.7 | THE SYSTEM SHALL NOT modify any file under `skills/` or bump any skill `version:`. |
| R-8.8 | THE SYSTEM SHALL land in four slices — A (Units 7+6), B (1+2, TC-S6), C (3, TC-S5), D (4+5) — each green on `make check` alone. |

---

## Declined

- **T8 stdio MCP shim.** Standing decision
  (`docs/user-guide/explanation/github-mcp-and-automation.md:5`) holds; the only
  MCP in the user's workflow is GitHub MCP for remote repository operations.

## Resolved during review (2026-09-08)

- `profile` enum: parent R-1.7 adopted; `01-flexibility…md:121` corrected (user).
- `CLAUDE.md`: `@AGENTS.md` import only (user).
- Pre-existing workspaces: `skills install` writes the contract (user).
- `discover validate` and `prd abandon` log (user).
- `discover new` duplicate title already refuses with `ExitConflict` (measured).
- No clock seam; far-dated fixtures (lead recommendation, adopted).
- Code prefixes follow the host gate (`expiry.*`, `frontmatter.*`) per the
  `codes.go` contract.
- `logbook` owns the log sentence to avoid `product` import cycles.
- TC-S2 rewritten from the measured `next:`/`fix:` chain; `add component` added;
  first `prd complete` deliberately fails.
