---
type: ears
slug: derived-drift-repair
status: locked
created: 2026-08-23
---

# Derived Drift Repair (`validate --fix`) — EARS Specifications

## Definitions

Used throughout; each is enumerated, not descriptive.

- **derived-condition code** — one of exactly five finding codes:
  `frontmatter.tags-drift`, `node.drift`, `feature-index.drift`, `node.absent`,
  `feature-index.absent`. These are every condition the rebuild writes for.
- **decision-class finding** — any finding whose code is not a derived-condition
  code.
- **hand-owned node** — a `CLAUDE.md` reported by gate 5 as `node.hand-owned`.
- **repair pass** — one invocation of the rebuild entry point against the workspace.
- **first run** / **second run** — the gate-set executions before and after a
  repair pass.

Mapping to design: Unit 0 is a prerequisite (LLD D5/P1, D9) with no user-visible
behaviour. Unit 1 realises HLD Goals 1–2, Unit 2 realises Goals 3 and 5, Unit 3
realises Goal 4, Unit 4 realises the reporting surface for Goals 1–2.

---

## Unit 0: Prerequisites (no user-visible behaviour)

**Why:** two clauses in Units 2 and 3 are unimplementable against the shipped code.
This unit makes them implementable and lands on its own.

| ID | EARS statement |
|---|---|
| R-0.1 | THE SYSTEM SHALL allow an error constructed by `model.Errorf` to carry an underlying error, and SHALL expose it via `Unwrap`. |
| R-0.2 | WHEN a write performed by a graph writer fails, THE SYSTEM SHALL propagate the underlying filesystem error such that `errors.Is(err, fs.ErrPermission)` reports truthfully. |
| R-0.3 | THE SYSTEM SHALL leave the rendered message text of every existing error unchanged by R-0.1 and R-0.2. |
| R-0.4 | THE SYSTEM SHALL report, from the rebuild entry point, one section naming each document whose derived tags it rewrote. |
| R-0.5 | THE SYSTEM SHALL produce, from `graph build`, output byte-identical to that produced before R-0.4. |
| R-0.6 | THE SYSTEM SHALL NOT derive the set of rewritten files by any traversal separate from the traversal that performed the writes. |

---

## Unit 1: Flag surface and partitioning

**Why:** the flag must be inert by default and must decide, from evidence rather
than from a guess, whether any repair is warranted at all.

| ID | EARS statement |
|---|---|
| R-1.1 | THE SYSTEM SHALL accept a boolean flag `--fix` on the `validate` command. |
| R-1.2 | WHERE `--fix` is absent, THE SYSTEM SHALL produce output, workspace mutations, and an exit code identical to those produced before this change. |
| R-1.3 | THE SYSTEM SHALL NOT accept `--fix` on any command other than `validate`. |
| R-1.4 | WHEN `--fix` is supplied, THE SYSTEM SHALL execute the first run to completion before performing any repair pass. |
| R-1.5 | WHEN the first run produces no finding carrying a derived-condition code, THE SYSTEM SHALL NOT perform a repair pass, and SHALL return the first run's results unchanged. |
| R-1.6 | WHEN the first run produces at least one finding carrying a derived-condition code, THE SYSTEM SHALL perform exactly one repair pass. |
| R-1.7 | THE SYSTEM SHALL NOT perform more than one repair pass per invocation. |
| R-1.8 | THE SYSTEM SHALL define the derived-condition code set as a single declared constant, and SHALL NOT determine membership by string pattern at runtime. |
| R-1.8a | THE SYSTEM SHALL partition the first run's findings by code alone, and SHALL NOT consider a finding's severity when deciding whether it carries a derived-condition code. |
| R-1.9 | THE SYSTEM SHALL fail its own test suite if any finding code declared in `model/codes.go` contains `drift`, case-insensitively, and appears in neither the derived-condition code set nor a declared decision-class allowlist. |
| R-1.10 | THE SYSTEM SHALL NOT alter, suppress, downgrade, reword, or re-order any decision-class finding as a consequence of `--fix` being supplied. |

---

## Unit 2: Repair execution

**Why:** repair must be the existing convergence path rather than a second
implementation of derivation, so that it cannot disagree with the gate that
detected the condition.

| ID | EARS statement |
|---|---|
| R-2.1 | THE SYSTEM SHALL implement the repair pass by invoking `graph.Rebuild`. |
| R-2.2 | THE SYSTEM SHALL NOT contain any derivation of tags, context-node blocks, or feature indexes that is reachable only from the repair pass. |
| R-2.3 | WHEN a repair pass completes without error, THE SYSTEM SHALL execute the second run against the same workspace. |
| R-2.4 | THE SYSTEM SHALL derive its reported gate results, for an invocation in which a repair pass completed, from the second run only. |
| R-2.5 | THE SYSTEM SHALL NOT perform network access during a repair pass. |
| R-2.6 | WHERE a derived artifact's current content already equals its freshly derived content, THE SYSTEM SHALL NOT write that file during a repair pass. |
| R-2.7 | WHEN `--fix` is supplied against a workspace in which the first run produces no derived-condition finding, THE SYSTEM SHALL leave every file in the workspace byte-identical. |
| R-2.8 | WHEN two `--fix` invocations run consecutively against the same workspace with no intervening change, THE SYSTEM SHALL write no bytes during the second. |
| R-2.9 | THE SYSTEM SHALL leave a repaired workspace's tree byte-identical to the tree produced by `graph build` followed by `validate` from the same starting state. |
| R-2.10 | THE SYSTEM SHALL NOT modify `workspace.lock.yaml` during a repair pass. |
| R-2.11 | THE SYSTEM SHALL NOT write a `CLAUDE.md` that carries no generated markers. |
| R-2.12 | WHEN any command declines to write a `CLAUDE.md` under R-2.11, THE SYSTEM SHALL emit a finding naming that file, whether the command is a repair pass or a build. |
| R-2.12a | THE SYSTEM SHALL emit the R-2.12 finding from the code path that performs the write, and SHALL NOT require its caller to re-derive which files were declined. |

---

## Unit 3: Unrepairable writes

**Why:** a derived file can live inside a read-only synced slice. The person who
hits this cannot act on a permission error; they can act on "re-sync the slice."

| ID | EARS statement |
|---|---|
| R-3.1 | IF a write attempted during a repair pass fails with a permission error, THEN THE SYSTEM SHALL report a finding that names the workspace-relative path of that file. |
| R-3.2 | IF a write attempted during a repair pass fails with a permission error, THEN THE SYSTEM SHALL report, in that same finding, the remedy of changing the source repository, updating its pin, and re-running `workspace sync`. |
| R-3.3 | IF a write attempted during a repair pass fails with a permission error, THEN THE SYSTEM SHALL leave that file's content and mode unchanged. |
| R-3.4 | IF a write attempted during a repair pass fails with a permission error, THEN THE SYSTEM SHALL NOT attempt to change that file's mode in order to complete the write. |
| R-3.5 | IF a write attempted during a repair pass fails with a permission error, THEN THE SYSTEM SHALL abort the repair pass, and SHALL NOT execute the second run. |
| R-3.6 | IF a repair pass aborts for any reason, THEN THE SYSTEM SHALL report the first run's gate results together with the abort finding. |
| R-3.7 | WHEN a repair pass has aborted, THE SYSTEM SHALL exit non-zero. |
| R-3.8 | THE SYSTEM SHALL classify a write failure as a permission error by testing the returned error against `fs.ErrPermission`, and SHALL NOT classify it by parsing a rendered message. |
| R-3.9 | IF a repair pass aborts for a reason other than a permission error, THEN THE SYSTEM SHALL exit with the code that failure produces when `--fix` is absent. |

---

## Unit 4: Reporting and exit codes

**Why:** the run must state what it changed, and its exit code must remain a
truthful statement about the workspace as it now exists on disk.

| ID | EARS statement |
|---|---|
| R-4.1 | WHEN a repair pass has written at least one file or skipped at least one hand-owned node, THE SYSTEM SHALL report one section, distinct from the gate list, enumerating each such file. |
| R-4.2 | THE SYSTEM SHALL populate that section from the findings returned by the repair pass, and SHALL NOT re-derive the list of written files independently. |
| R-4.3 | THE SYSTEM SHALL render the repair section without a `[n/total]` gate prefix, and SHALL NOT include it in the gate count. |
| R-4.4 | THE SYSTEM SHALL report exactly one workspace banner record per invocation. |
| R-4.5 | WHEN a repair pass completed, THE SYSTEM SHALL report the second run's banner as that record. |
| R-4.6 | WHEN the second run reports no failing gate, THE SYSTEM SHALL exit zero. |
| R-4.7 | WHEN the second run reports at least one failing gate, THE SYSTEM SHALL exit with the same code a non-`--fix` run reporting those same gate results would produce. |
| R-4.8 | THE SYSTEM SHALL NOT exit with the usage exit code for any invocation in which `--fix` was parsed successfully. |
| R-4.9 | WHEN a repair pass completed, THE SYSTEM SHALL NOT emit the first run's gate results in any output stream. |
| R-4.10 | WHERE `--json` is supplied together with `--fix`, THE SYSTEM SHALL emit a single JSON document containing both the repair section and the reported gate results. |
| R-4.11 | THE SYSTEM SHALL emit no output and perform no exit from any package below `cmd/`. |
| R-4.12 | WHEN a repair pass wrote no files and skipped no hand-owned node, THE SYSTEM SHALL omit the repair section rather than emitting it empty. |

---

## Traceability

Requirements in this document bind to implementation and tests by grep-able
`@spec` markers of the form `@spec req://uncle-os/derived-drift-repair@0.1#R-n.m`.
Per the ontology rule, every mandatory clause SHALL carry at least one test-side
marker.
