---
type: ears
id: ears-go-cli-tui-port
title: Go CLI + TUI Port — EARS Specifications
status: draft
tags: [kind/ears, status/draft]
---

# Go CLI + TUI Port — EARS Specifications

Keywords: `THE SYSTEM SHALL` (always-on), `WHEN` (event), `WHILE` (during a
state), `IF` (conditional/gate), `WHERE` (context-scoped). "The system" = the
`company-os` Go binary.

Intent arrow: `docs/hld/go-cli-tui-port.md` → `docs/lld/go-cli-tui-port.md` →
this file → code and tests.

Units 0–4 and 6–9 are **Phase 1**. Unit 5 is **Phase 2** and does not start until
the Phase 1 parity gate (R-9.1) is green.

---

## Unit 0: Behavioral parity (the acceptance oracle)

**Why:** This is a port, not a redesign. But the oracle is weaker than it looks:
the golden snapshots cover one command out of sixteen, only on the passing path,
and they stop working on 2027-01-01. Every requirement below exists to make
"parity is proven" a statement that is actually true when R-9.1 reads it.

| ID | EARS statement |
|---|---|
| R-0.0 | BEFORE any Go implementation work begins, THE SYSTEM SHALL have its fixture expiry dates pushed out and both golden files re-baselined **from the Python CLI**, because `TODAY` (`bin/company-os:31`) compared against `examples/workspace/teams/customer-engagement/governance/exceptions.yaml:9` (`expires: '2026-12-31'`) breaks both goldens on 2027-01-01 and both deviations on 2027-01-14, independent of this change. |
| R-0.1 | WHEN `validate` is run against `examples/workspace`, THE SYSTEM SHALL emit stdout byte-identical to `examples/golden-validate.txt` after `normalize()` (`examples/acceptance.sh:20`). |
| R-0.2 | WHEN `validate` is run against `examples/federated`, THE SYSTEM SHALL emit stdout byte-identical to `examples/federated-golden-validate.txt` after normalization. |
| R-0.3 | THE SYSTEM SHALL pass `examples/acceptance.sh` with no edits to that script. |
| R-0.4 | WHERE a workspace has no `workspace.yaml` manifest, THE SYSTEM SHALL print gate headers numbered `[N/7]`; WHERE a manifest is present, THE SYSTEM SHALL print `[N/8]` and run the federated slice-integrity gate. |
| R-0.5 | THE SYSTEM SHALL compute the gate-count denominator at run time and SHALL NOT hardcode it. |
| R-0.6 | WHEN `graph build` is run twice in succession, THE SYSTEM SHALL leave the workspace byte-identical after the second run. |
| R-0.7 | IF any observable output, exit status, or filesystem effect differs from the Python CLI for the same inputs, and that difference is not listed in R-0.7a, THE SYSTEM SHALL be treated as defective and SHALL be corrected rather than the golden files re-baselined. |
| R-0.7a | THE SYSTEM SHALL permit exactly these sanctioned differences from the Python CLI, and no others: (a) `write_feature_indexes`' idempotency guard (`bin/company-os:1530-1537`) becomes a semantic compare rather than a byte compare; (b) YAML comments survive a read-modify-write of `deviations.yaml` and `exceptions.yaml`, which PyYAML's `safe_load` destroys today; (c) the "reality template not found" failure at `:2041` ceases to exist because R-1.11 embeds the template; (d) the OKF Phase 0 done-gate date-parsing fix; (e) malformed YAML and malformed frontmatter exit 4 with a diagnostic instead of raising an uncaught `YAMLError` and exiting 1 through a Python traceback — `load_yaml` (`:58`) and `frontmatter` (`:76`) have no `try` today, so code 4's headline case has zero exit sites and the Go port adds them; (f) the hand-rolled conditional-requirement checks at `:601` and `:2021` exit 2 rather than 1; (g) a read-modify-write of an **authored** YAML artifact re-emits it under `yaml.v3`'s layout rather than PyYAML's — measured, 66 of 112 committed YAML documents differ on re-emit, from three PyYAML *emitter* policies (80-column plain-scalar folding on 41, indentless block sequences on 24, single-quoting of a plain `://` scalar inside a flow mapping on 1) that `yaml.v3` neither implements nor exposes a knob for. PyYAML's own `safe_dump` already fails to reproduce the committed bytes of every `deviations.yaml`, `exceptions.yaml`, and `ids/registry.yaml` in the corpus, so `deviation declare` and `exception request` reflow their target file under the Python CLI **today**; the Go port changes only the shape of that reflow, once, at the transition. The node tree survives structurally identical on 112 of 112 and emit reaches a fixed point after one pass, which is what R-0.6 needs. This carve-out SHALL NOT extend to derived artifacts — see R-0.7c; (h) a tab character used as intra-line whitespace (after a `:`, trailing, or inside a flow collection) or inside a plain scalar's content is accepted rather than raising PyYAML's `ScannerError`. Measured against vendored PyYAML 6.0.2, `yaml.v3` rejects a tab in every *structural* position — as indentation, before a key, after a block-sequence `-` — exactly as PyYAML does, and accepts a tab inside a quoted or block-literal scalar exactly as PyYAML does, so the residual divergence cannot change a document's shape; it can only convert an uncaught `ScannerError` into the value the author intended. THE SYSTEM SHALL NOT add a pre-scan to close it: PyYAML accepts tabs inside quoted and block-literal scalars, so any pre-scan cheap enough to be worth writing would reject documents the Python CLI loads today — a regression in the opposite and worse direction; (i) usage and argument-error text — on stderr for errors and on **stdout** for `--help` — is not byte-identical to argparse's, and the top-level usage line legitimately differs in content because Go carries `--json` and `--version` flags Python does not have. Measured: argparse wraps to `COLUMNS`, so `COLUMNS=200 company-os nosuch` emits usage on one line and the non-TTY default of 80 wraps it across three — byte-parity here is parity against an environment variable, and any golden captured from it is not reproducible across CI runners. The *content* is not carved out; see R-1.4a; (j) a **well-formed** YAML document of the wrong *shape* — a non-list `ids:`, a non-mapping registry root, a null `components:` — exits 4 with a diagnostic rather than raising an uncaught `AttributeError`/`TypeError` and exiting 1 through a Python traceback. This is the sibling of (e), which covers only *malformed* YAML; these documents parse cleanly and fail at attribute access. THE FILESYSTEM EFFECT IS NOT CARVED OUT: Python writes nothing on these paths, so the Go port SHALL write nothing; (l) an **argument** of the wrong shape reaching command code — an optional positional argparse allows through as `None`, such as `discover new` with no title (`bin/company-os:2688` declares `title` with `nargs="?"`, so `slugify(None)` raises `AttributeError` at `:73`) — exits 2 with a diagnostic rather than an uncaught traceback exiting 1. This is the CLI-argument sibling of (j), whose stated subject is a YAML document; without it the case has no clause naming it literally; (k) argparse's unique-prefix flag abbreviation (`--plat` → `--platform`) is not reproduced. Zero callers in the repository, zero harness coverage, and argparse's ambiguity resolution is a nontrivial algorithm — closed as a sanctioned omission rather than left open as a permanent deferral. |
| R-0.7c | THE SYSTEM SHALL guard every write of a **derived** YAML artifact on a semantic compare of the parsed structure and SHALL skip the write when it is unchanged — covering `write_feature_indexes` (`:1530-1537`), `resolve_team_governance`'s `generated/effective-governance.yaml` (`:329-330`), and `register_id`'s `ids/registry.yaml` (`:1823`) — because `generated/effective-governance.yaml` is byte-stable under PyYAML's own `safe_dump` today, so re-emitting it under `yaml.v3`'s layout would dirty the tree and break R-0.10. |
| R-0.7d | THE SYSTEM SHALL collapse duplicate mapping keys on load, keeping the first key's position and the last value, so that a read-modify-write emits one pair as PyYAML's `safe_dump` does rather than preserving both. |
| R-0.7e | THE SYSTEM SHALL expand YAML aliases to their anchor's value as PyYAML does, and SHALL reject a merge-key or value-key token (`<<`, `=`) used in **any position except a mapping key** — value, sequence item, or bare document — with a syntax error rather than accepting it and materializing an author-never-typed `!!merge` tag into their file. Measured: PyYAML raises `ConstructorError` in all three non-key positions and accepts `<<: {…}`, `<<: *anchor`, `!!merge <<:`, and `=: 5` as legitimate keys. |
| R-0.8 | THE SYSTEM SHALL NOT add, remove, reword, or reorder any human-facing output line that exists in the Python CLI today. |
| R-0.9 | THE SYSTEM SHALL capture, from the Python CLI and before its deletion, golden snapshots exercising at least one `[FAIL]` finding per gate and at least one `[warn]` line, because neither committed golden contains a single `[FAIL]` or `[warn]` line against 15 failure sites in `cmd_validate`. |
| R-0.10 | WHEN `graph build` and `governance resolve` are run against `examples/workspace` from a clean git tree, THE SYSTEM SHALL leave `git status` clean. |
| R-0.11 | THE SYSTEM SHALL **deterministically order** — not necessarily sort — every map-driven emission and every map-driven finding loop, because Go map iteration is randomized by design and two of these reach byte-frozen stdout. Measured, the required order differs per site: `workspace.lock.yaml`'s `files:` map (`:2614`) is **insertion order** — a nested walk of manifest slice order, then that slice's `paths:` list order, then `sorted(rglob)` within each subtree — and `examples/federated`'s manifest lists `governance/` before `components/`, so the committed lock is the reverse of alphabetical; gate 8's finding loop (`:2521`) inherits that same order back through the parser, and `examples/failing-federated-golden-validate.txt` freezes it; `feature_index_unresolved` (`:1519`) is sorted, but only incidentally, because `build_feature_index` (`:1440`) iterates `sorted(cids)`. Sorting the first two breaks a committed golden. |
| R-0.11a | THE SYSTEM SHALL reproduce CPython's `PurePath` component-wise ordering wherever Python uses `sorted(Path.rglob(...))`, which is not string ordering: `sdd/adr/a.md` precedes `sdd/adr-x.md` under `PurePath` and follows it under a byte sort, because `/` (0x2F) sorts after `-` and `.`. Note `aggregate_hash` (`:2436`) uses a plain **string** sort over the very keys `workspace.lock.yaml` emits in walk order — the same data under two different orders. |

---

## Unit 1: Command surface and parsing fidelity

**Why:** Engineers, CI pipelines, and all four shipped agent skills invoke
specific commands with specific flags today. Drift in the surface silently breaks
callers this change never intended to touch. Two of these requirements are about
*parsing*, not output — they produce wrong behavior rather than wrong bytes, which
is why they outrank everything cosmetic.

| ID | EARS statement |
|---|---|
| R-1.1 | THE SYSTEM SHALL accept all 16 subcommands with the same names, actions, positional arguments, flags, defaults, and choice sets as `main()` in `bin/company-os:2661-2781`. |
| R-1.2 | THE SYSTEM SHALL resolve the workspace root in the order `--root` flag, then `$COMPANY_OS_WORKSPACE_ROOT`, then the current directory. |
| R-1.3 | WHEN any subcommand other than `init` or `scratchpad` is run outside a workspace root, THE SYSTEM SHALL fail fast without performing any work. |
| R-1.4 | WHEN `company-os` is invoked with no subcommand, THE SYSTEM SHALL print usage help, exit 2, and SHALL NOT start an interactive session. |
| R-1.4a | WHEN an argument is invalid, THE SYSTEM SHALL emit a diagnostic naming the subcommand, the offending argument, the offending value, and the valid choices — semantically equivalent to argparse's `company-os <sub>: error: argument <name>: invalid choice: '<v>' (choose from a, b, c)` — on a single line, plus a sub-parser-scoped usage line. R-0.7a(i) waives argparse's *wrapping*, not its *content*: dropping the diagnostic removes a human-facing line that R-0.8 forbids removing, on 18 of 193 harness invocations, on a stream four shipped agent skills read. |
| R-1.5 | THE SYSTEM SHALL preserve the frontmatter parser's exact semantics, accepting only documents opening with `---\n` and closing with `\n---\n`, with a differential test against the Python regex because Go requires `(?s)` and Go's `$` anchor semantics differ from Python's. |
| R-1.6 | THE SYSTEM SHALL resolve YAML scalars to the same values PyYAML's `safe_load` resolves them to, including YAML 1.1 timestamp, boolean (`yes`/`no`/`on`/`off`), sexagesimal, and leading-zero octal forms — such that `reviewDate: 2027-01-15` unquoted (`examples/workspace/teams/customer-engagement/governance/deviations.yaml:8`) renders as `2027-01-15` and not `2027-01-15 00:00:00 +0000 UTC`, while `reviewDate: '2027-01-14'` quoted on line 10 of the same file resolves through the same code path. |
| R-1.7 | THE SYSTEM SHALL preserve unknown keys, authored key order, quote styles, and comments through every read-modify-write of a YAML artifact, using node-level access throughout and never struct unmarshal — subject to the layout carve-out at R-0.7a(g). |
| R-1.7a | THE SYSTEM SHALL reproduce Python's `yaml.safe_load(...) or {}` as **truthiness**, not a nil check: `0`, `0.0`, `false`, `no`, `off`, `''`, `""`, `[]`, `{}`, `null`, and `~` all collapse to an empty mapping, while a truthy non-mapping is returned as authored. A nil-only implementation is wrong on ten of the sixteen measured cases. |
| R-1.7b | THE SYSTEM SHALL reject a multi-document YAML stream and an unconstructible calendar date (e.g. `2035-02-30`), matching PyYAML's `ComposerError` and `ValueError`, where `yaml.v3` silently accepts the first document or falls back to a string. |
| R-1.8 | THE SYSTEM SHALL preserve the next-command guidance chain: every mutating command SHALL print the next command in the workflow. |
| R-1.9 | WHERE a mutating command prints no next step today — `governance resolve` (`:334-346`), `exception request` (`:1127-1138`), `scratchpad init` (`:1141-1155`), `graph build` (`:1770-1780`) — THE SYSTEM SHALL continue to print exactly what it prints today, because R-0.8 outranks R-1.8. Closing those gaps is a separate change. |
| R-1.10 | THE SYSTEM SHALL resolve user template overrides from workspace-relative paths only — `teams/<t>/templates/`, `platforms/<p>/templates/`, `company-os/templates/` — and SHALL NOT read any template from a path relative to the binary. |
| R-1.11 | THE SYSTEM SHALL embed all built-in templates in the binary, including `templates/reality-component.md` (`bin/company-os:526-529`). |
| R-1.12 | THE SYSTEM SHALL normalize path separators to POSIX form when writing derived content, preserving the accommodation at `bin/company-os:1747-1748`. |
| R-1.13 | THE SYSTEM SHALL reproduce `slugify`'s case-folding behavior (`:72-73`), noting that Python's `.lower()` and Go's `strings.ToLower` diverge on a small set of code points and the result feeds filesystem paths. |
| R-1.14 | THE SYSTEM SHALL apply the OKF v0.2 Phase 0 done-gate date-parsing fix (`docs/tasks/okf-v02-conformance.md:44-57`), which is carved out of the OKF deferral because R-0.7 would otherwise compel faithful reproduction of a known bug. |

---

## Unit 2: Findings as records

**Why:** The Python CLI destroys structure at the point of computation —
`validate` counts problems in a bare integer (`:923`) and prints an English
sentence at each of 15 failure sites. Records are the precondition for JSON, for
the TUI, and for exit codes that mean something. A flat finding list is not
enough: the golden has a gate header with zero findings under it.

| ID | EARS statement |
|---|---|
| R-2.1 | THE SYSTEM SHALL model validation output as an ordered list of gate results, each carrying its ordinal, stable slug, human title, and its findings — such that a gate that produced no findings still renders its header, as `examples/golden-validate.txt:11-12` requires. |
| R-2.2 | THE SYSTEM SHALL represent every gate result, conflict, and validation error as a structured record carrying at minimum severity, stable code, subject, path, message, and typed fields. |
| R-2.3 | THE SYSTEM SHALL make the typed fields available to **both** the text and JSON renderers, because human lines such as `[ok] communications: feature-index in sync (1 component(s))` carry counts that must reach the text output. |
| R-2.4 | THE SYSTEM SHALL assign each record a stable machine-readable code that does not change when its human message is reworded, and SHALL declare every such code in one reviewable location so the set can be checked for collisions and naming consistency — the codes are the `--json` contract under R-3.4, and a per-producer-package const block gives them no place to be reviewed and forces `internal/render` to import every cluster to reach them. |
| R-2.5 | THE SYSTEM SHALL carry an explicit per-gate line-prefix policy covering all seven observed shapes, and SHALL NOT assume a uniform prefix rule; gate 5 alone uses three distinct shapes (`:1030`, `:1036`, `:1040`) and gates 7 and 8 use none. |
| R-2.6 | THE SYSTEM SHALL model the leading blank line as a property of the gate header, present on every gate except the first. |
| R-2.6a | THE SYSTEM SHALL carry the total gate count explicitly rather than deriving it from the length of the gate list, so that a run aborting mid-gate still renders completed gates under the true denominator. RULING: this resolves an R-0.8/R-2.6 collision found at the parity checkpoint. `exception/garbage-expires` crashes the oracle at `:970` after it has already printed the banner, all of gate 1, and three of gate 2's four lines; a derived denominator would render those as `[1/2]`/`[2/2]` where the oracle wrote `[1/7]`/`[2/7]`, so the port dropped them instead — removing six human-facing lines that R-0.8 forbids removing. The count is knowable before any gate runs (`:930` decides 7 vs 8 from manifest presence), so this was never a real conflict: carrying it costs nothing and satisfies both. |
| R-2.7 | THE SYSTEM SHALL reproduce gate 4's conditional `[ok]` (`:1003-1008`): a document with core-field errors emits its failures and no `[ok]` line. |
| R-2.8 | THE SYSTEM SHALL NOT compose human-readable sentences anywhere outside a renderer. |
| R-2.9 | THE SYSTEM SHALL render human text and JSON from the same record set, such that neither renderer can report a finding the other does not have. |
| R-2.10 | THE SYSTEM SHALL NOT write to stdout or stderr, and SHALL NOT terminate the process, from any package below the command-dispatch layer — which requires `Workspace.require_root` (`:230`), `platform_dir` (`:238`), and `team_dir` (`:244`) to become error-returning rather than calling `die()`. |
| R-2.11 | THE SYSTEM SHALL return results and errors from every command entry point rather than exiting from within it. |
| R-2.12 | THE SYSTEM SHALL apply record-based computation to every check returning pre-composed prose in the Python CLI, including skill conflicts (`:837-866`), federated slice problems (`:2490-2526`), core-field errors, pointer errors, identity errors, and PRD governance checklists (`:551-570`). |

---

## Unit 3: Machine-readable output

**Why:** Agents already run non-interactively and already drive the CLI from four
shipped skills. What they lack is not a mode but structured output. Today an
agent wanting to know which gate failed must parse English.

| ID | EARS statement |
|---|---|
| R-3.1 | THE SYSTEM SHALL accept a global `--json` flag on every subcommand. |
| R-3.2 | WHEN `--json` is passed, THE SYSTEM SHALL emit the command's results as JSON on stdout and SHALL emit no human-formatted prose on stdout. |
| R-3.3 | WHEN `--json` is not passed, THE SYSTEM SHALL emit exactly the output it emits today, satisfying Unit 0. |
| R-3.4 | THE SYSTEM SHALL include a `schemaVersion` field in every `--json` payload, and SHALL NOT remove or repurpose a documented field without incrementing it. |
| R-3.4a | THE SYSTEM SHALL name the top-level record array `sections`, not `gates`. `GateResult` became the section carrier for non-validate commands during implementation — `ids list`, `skills list`, and `today` are all modeled as `[]GateResult` — so `gates` reads wrong for three of sixteen commands, and R-3.4 freezes the name on first publish. |
| R-3.4b | THE SYSTEM SHALL implement JSON as one encoder over the record types, not one renderer per command, so that R-7.7's tuple-equality between the text and JSON renderers holds by construction rather than by sixteen separate agreements. |
| R-3.5 | THE SYSTEM SHALL include the binary's version and build identifier in every `--json` payload. |
| R-3.6 | WHEN `--json` is passed to a mutating command, THE SYSTEM SHALL emit the next-command guidance as a structured field rather than omitting it, so that R-3.2 does not silently delete the affordance R-1.8 protects. |
| R-3.7 | WHERE a command produces no findings — `prd new`, `add`, `reality new`, `scratchpad init` — THE SYSTEM SHALL emit a defined JSON envelope describing what it created, and SHALL NOT emit an empty document. |
| R-3.8 | WHEN `--json` is passed and the command fails, THE SYSTEM SHALL still emit valid JSON on stdout and SHALL set the exit code per Unit 4. |
| R-3.9 | THE SYSTEM SHALL emit diagnostics and progress messages on stderr, never interleaved into `--json` stdout. |
| R-3.10 | THE SYSTEM SHALL emit no ANSI escape sequences from any subcommand other than `tui`, because the golden files depend on it and nothing currently states it. |

---

## Unit 4: Exit-code contract

**Why:** Callers distinguishing drift from an expired exception from a missing
file must parse stdout today. `examples/acceptance.sh:62` tests only zero against
non-zero, so distinct non-zero codes are safe to introduce and every existing CI
gate keeps reading identically. Note that code 2 is **already today's behavior** —
argparse exits 2 for a bad flag or unknown subcommand; the claim that "every
non-zero path returns 1" is false and is corrected here.

| ID | EARS statement |
|---|---|
| R-4.1 | WHEN a command succeeds, THE SYSTEM SHALL exit 0. |
| R-4.2 | WHEN `validate` reports one or more `[FAIL]` findings, THE SYSTEM SHALL exit 1. |
| R-4.3 | WHEN invocation is malformed — unknown subcommand, bad flag, missing required argument, bare invocation — THE SYSTEM SHALL exit 2, preserving argparse's existing behavior. |
| R-4.4 | **Amended 2026-07-26 — `tui` is exempt when a TTY is attached; see [Amendment 1](#amendment-1--r-44-exemption-for-tui-2026-07-26).** WHEN the workspace root is absent or a required workspace object does not exist — platform (`:238`), team (`:244`), component (`:367`), and the lookups at `:430`, `:584`, `:636`, `:676`, `:2561` — THE SYSTEM SHALL exit 3, because the invocation was well-formed and the workspace did not contain the object. |
| R-4.5 | WHEN a YAML document, frontmatter block, or schema-governed artifact is malformed, THE SYSTEM SHALL exit 4. |
| R-4.6 | WHEN a precondition gate refuses — a done-check with unchecked items or stale reality (`:703`), or a PRD sourced from an unvalidated discovery brief — THE SYSTEM SHALL exit 5. NOTE: a deviation aimed at a mandatory rule is **not** a code-5 site; `resolve_team_governance` (`:317-319`) records `deviationRejected` and continues, surfacing the refusal as a `validate` gate failure (exit 1), and `deviation declare` validates nothing. Enforcing at declare time would break R-9.1 parity. |
| R-4.7 | WHEN an external tool is unavailable or fails — git absent, git older than 2.27, clone or sparse-checkout failure, `--frozen` lock reconciliation failure (`:2564`-`:2594`) — THE SYSTEM SHALL exit 6. |
| R-4.8 | WHEN an interactive prompt is required and no TTY is attached — `tui`, and `_prompt` at `:1961` — THE SYSTEM SHALL exit 7. |
| R-4.9 | WHEN a command refuses to overwrite an existing artifact — `:417`, `:610`, `:1797`, `:1971`, `:2037` — THE SYSTEM SHALL exit 8. |
| R-4.10 | THE SYSTEM SHALL use only non-zero codes for every failure, such that any existing caller branching on zero against non-zero behaves identically to today. |
| R-4.11 | THE SYSTEM SHALL classify all 56 failure paths in `bin/company-os` — 52 `die()` call sites and 5 `sys.exit` statements — to a code in this contract, as a tracked deliverable rather than an inference made during implementation. Delivered at `.devlocal/go-port/exit-code-map.md`. |
| R-4.12 | THE SYSTEM SHALL document the exit-code contract in user-facing documentation, because no such contract has ever existed. |

---

## Unit 5: Interactive terminal UI (Phase 2)

**Why:** The recorded friction is an onboarding wall:
`.devlocal/research/2026-07-22-simplicity-user-journey-review.md:96`. A product
owner should be able to run one binary and navigate the workspace without
composing flag-complete commands from memory. This unit is Phase 2 because it is
strictly downstream of parity, has no oracle, and is the only unbounded element
of the scope. Note that it addresses the *navigation* half of the wall only —
friction F5 (35-40 domain terms) and F6 (EARS and `@spec` syntax) are untouched.

| ID | EARS statement |
|---|---|
| R-5.1 | THE SYSTEM SHALL provide an interactive terminal UI launched only by the explicit subcommand `company-os tui`. |
| R-5.2 | THE SYSTEM SHALL NOT launch the TUI from a bare invocation, from any other subcommand, or from any environment-variable trigger. |
| R-5.3 | IF `tui` is invoked with no TTY attached, THE SYSTEM SHALL print an explanatory message to stderr, exit 7, and make no filesystem change. |
| R-5.4 | THE SYSTEM SHALL ship read-only screens first, enumerated and asserted by test: workspace overview, `today --role`, validate results, component browser, PRD browser, discovery browser, `governance explain`, `skills list`, `ids list`, and `workspace status`. |
| R-5.5 | THE SYSTEM SHALL ship mutating forms only after R-5.4 is complete, one at a time, beginning with `discover new` and `prd new`, and SHALL NOT build forms for `workspace sync` or `scratchpad init`. |
| R-5.6 | WHEN the TUI is about to perform a mutating action, THE SYSTEM SHALL display the exact flag-complete `company-os` invocation equivalent to that action before executing it. |
| R-5.7 | THE SYSTEM SHALL derive the previewed command from the same argument structure it executes, and SHALL NOT hand-write a preview string per screen, because a hand-written preview drifts from what runs and destroys the property justifying interactive mutation. |
| R-5.8 | WHILE a mutating action is previewed, THE SYSTEM SHALL require explicit confirmation and SHALL make no filesystem change until confirmation is given. |
| R-5.9 | WHEN a mutating action is cancelled before confirmation, THE SYSTEM SHALL leave the workspace exactly as it was; THE SYSTEM SHALL NOT claim transactional rollback for failures occurring after execution begins, because only `init` is atomic today (`:1982`). |
| R-5.10 | THE SYSTEM SHALL ensure every value the TUI collects has a command-line flag equivalent, so that any TUI action is reproducible in CI and by an agent. |
| R-5.11 | THE SYSTEM SHALL delegate scaffolding to the same code path `init`'s existing `_prompt` wizard (`:1954-1965`) uses, so that GPF-R-1.3 and GPF-R-1.4 hold for both interactive paths without a second implementation. |
| R-5.12 | THE SYSTEM SHALL execute TUI actions by calling the same in-process command functions the flag CLI calls, and SHALL NOT shell out to itself or parse its own rendered output. |
| R-5.13 | THE SYSTEM SHALL render validate results in the TUI from the same records the text and JSON renderers consume. |
| R-5.14 | THE SYSTEM SHALL exit the TUI on `q`, `Esc`, or `Ctrl-C` from any screen, leaving no partial write. |
| R-5.15 | THE SYSTEM SHALL honour `NO_COLOR` within the TUI and SHALL degrade legibly on a terminal narrower than 80 columns and on resize. |
| R-5.16 | THE SYSTEM SHALL NOT require a TTY, a terminal size, or any interactive capability for any subcommand other than `tui`. |
| R-5.17 | WHEN `tui` is invoked with a TTY attached from a directory that is not a workspace root, THE SYSTEM SHALL open a recovery menu rather than exiting — the workspace roots found at or one level below that directory, a previewed `init` for the current directory, and an explanation of what a root is — and WHEN a root is selected SHALL re-enter the normal catalog with that path as `--root`. The scan SHALL NOT descend more than one level, because a root surfacing from four levels down is not predictable from where the reader is standing. Added 2026-07-26; exempts `tui` from R-4.4 — see [Amendment 1](#amendment-1--r-44-exemption-for-tui-2026-07-26). |
| R-5.18 | WHEN the TUI opens on a workspace, THE SYSTEM SHALL present, ahead of every other screen, each derived artifact that is missing or stale together with the command that produces it, because each is otherwise reported only where the reader must already be looking: an unresolved `effective-governance.yaml` as a `[warn]` line inside `today --role`, a drifted generated block as a `[FAIL]` line at the end of `validate`. |
| R-5.19 | THE SYSTEM SHALL derive that list from the same checks `validate` runs, so the advisor and the gate cannot disagree about whether something is wrong; WHERE nothing is wrong the list SHALL be empty, so a healthy workspace presents the catalog it presented before this requirement existed. |
| R-5.20 | WHILE the advisor is detecting problems, and WHILE a reader is opening an offer to read its preview, THE SYSTEM SHALL make no filesystem change — detection runs on every TUI start, so it must be incapable of altering what it inspects. |
| R-5.21 | THE SYSTEM SHALL express every offered fix as a single flag-complete `company-os` invocation executed through the preview-and-confirm path of R-5.6–R-5.8, such that the previewed line parses back into the invocation that runs (R-5.10). |
| R-5.22 | IF a detected problem cannot be resolved by an invocation the system can construct — a missing federation manifest needs repo URLs and commit pins nothing can infer — THE SYSTEM SHALL report and explain it and SHALL NOT offer a fix, because a form that writes a plausible-but-wrong manifest is worse than the missing file. |

### Amendment 1 — R-4.4 exemption for `tui` (2026-07-26)

**Authority:** R-5.17, specified and added 2026-07-26. Executed by task 7.7 of
`docs/tasks/go-cli-tui-port.md`.

R-4.4 requires exit 3 when the workspace root is absent, and until this date
`tui` obeyed it: `cmdTUI` called `RequireRoot` and returned exit 3 with the
root-resolution order printed. R-5.17 replaces that path for `tui` alone.

| Aspect | Disposition |
| --- | --- |
| Every subcommand other than `tui` | **Unchanged.** R-4.4 holds as written. |
| `tui` with no TTY | **Unchanged.** R-5.3's exit 7 is decided before the root question is asked, so the non-interactive contract is untouched and no filesystem change occurs. |
| `tui` with a TTY, outside a root | **Exempted.** Opens the recovery menu instead of exiting 3. |
| Exit code when the reader quits that menu without choosing | **0, not 3.** This is the observable surface of the exemption and the part a caller could branch on under R-4.10. |

**Why the exemption rather than better error text.** The behaviour was measured
standing in `examples/banking/bank/workspaces/`, a directory holding two
workspace roots: `tui` refused and printed the resolution order. For an audience
defined as blocked on terminal fluency, printing a resolution order is the wrong
answer to "I am one directory too high" when both correct answers are in the
directory being read. The content of the old error was not deleted — it is the
menu's third entry (`rootHelp`), shown rather than fatal.

---

## Unit 6: Distribution

**Why:** The goal that selected Go over Python is one sentence: a user copies a
binary into `~/.local/bin` and starts working. If any prerequisite survives, the
weeks spent porting bought nothing — and an unsigned macOS binary has a
prerequisite that is harder than `pip install pyyaml`.

| ID | EARS statement |
|---|---|
| R-6.1 | THE SYSTEM SHALL be distributed as a single statically linked executable requiring no interpreter, runtime, or installed library on the target machine. |
| R-6.2 | THE SYSTEM SHALL build with `CGO_ENABLED=0` for `darwin/arm64`, `darwin/amd64`, and `linux/amd64`. |
| R-6.3 | THE SYSTEM SHALL ship darwin release artifacts that are Developer-ID-signed and notarized; IF notarization is not performed, THE SYSTEM SHALL document the `xattr -d com.apple.quarantine` first-run workaround in the install documentation and record it in the HLD as an accepted cost. |
| R-6.4 | WHEN a user follows the published install documentation using a **downloaded release artifact**, THE SYSTEM SHALL run every subcommand with no further setup; verification against a locally built binary SHALL NOT satisfy this requirement. |
| R-6.5 | THE SYSTEM SHALL function correctly on a machine with no Python installation. |
| R-6.6 | THE SYSTEM SHALL NOT require, generate, or depend on a shell launcher script. |
| R-6.7 | THE SYSTEM SHALL NOT read any file from a path relative to the binary's own location. |
| R-6.8 | THE SYSTEM SHALL report its version and build identifier via `company-os --version`. |
| R-6.9 | THE SYSTEM SHALL operate correctly on a workspace last written by a different build, and THE SYSTEM SHALL state in documentation whether version skew across a federated workspace is supported. |
| R-6.10 | THE SYSTEM SHALL publish checksums alongside each released artifact. |
| R-6.11 | THE SYSTEM SHALL NOT claim, test, or release a Windows build under this change. |

---

## Unit 7: Testing and the parity oracle

**Why:** The first draft of this spec asserted that `go test` passing plus
`acceptance.sh` passing means parity. That is false twice: `acceptance.sh`
byte-freezes one command, and `go test` is written against the Go implementation
and cannot detect a Go/Python divergence by construction. The cutover decision
rests on this unit, so it has to be true.

| ID | EARS statement |
|---|---|
| R-7.1 | THE SYSTEM SHALL provide a cross-implementation differential harness that runs the Python and Go binaries over identical fixtures across all 16 commands and diffs stdout, stderr, exit code, and the resulting filesystem tree. |
| R-7.1a | THE SYSTEM SHALL provide the harness with a declared-divergence registry keyed by invocation and stream, each entry naming the R-0.7a clause or `.devlocal/go-port/exit-code-map.md` line that sanctions it, and SHALL fail both on an undeclared divergence **and on a stale declaration** whose invocation now passes. Without it every sanctioned carve-out lives only in prose while the tool gating the cutover exits 1 on any difference — R-0.7a(g) alone sanctions a re-emit divergence on 66 of 112 committed documents, so R-7.9 cannot be evaluated as written. |
| R-7.2 | THE SYSTEM SHALL exercise the differential harness on both the passing and the failing path of every command, and over the `workspace`, `standalone-team`, and `federated` fixtures. |
| R-7.3 | THE SYSTEM SHALL port each of the 85 real assertions in `examples/selftest.py` (86 `check()` call sites, one a skip sentinel) to a named Go test, tracked against the inventory at `.devlocal/go-port/selftest-inventory.md`, before `selftest.py` is deleted. |
| R-7.3a | THE SYSTEM SHALL treat R-7.3 as necessary but not sufficient for R-7.4, because selftest drives only 7 of 16 subcommands through subprocess — `discover`, `deviation`, `exception`, `check`, `governance`, `today`, `graph`, `ids`, `skills`, and `scratchpad` have zero inherited behavioral coverage and depend entirely on the differential harness. |
| R-7.4 | THE SYSTEM SHALL provide native `go test` coverage for every internal package. |
| R-7.5 | THE SYSTEM SHALL include tests asserting byte-identical reproduction of both committed golden snapshots and both failure-path snapshots captured under R-0.9. |
| R-7.6 | THE SYSTEM SHALL include tests covering each exit code in the Unit 4 contract. |
| R-7.7 | THE SYSTEM SHALL include tests asserting that each `--json` renderer reports the same `{gate, code, severity, subject}` tuple set as its text counterpart. |
| R-7.8 | THE SYSTEM SHALL include a differential test for the frontmatter regex and for YAML 1.1 scalar resolution, covering the unquoted-date, quoted-date, and boolean-word cases named in R-1.6. |
| R-7.9 | WHEN the differential harness reports zero divergence, the `go test` suite passes, and `examples/acceptance.sh` passes, THE SYSTEM SHALL be considered at parity. |

---

## Unit 8: Retiring the single-file requirement

**Why:** `docs/ears/federation-enrichment.md:149` (R-7.4) mandates keeping the
single-file CLI and its helpers, repeated in five other documents. No document
states why; the constraint post-dates the founding proposal — which never mentions
"single file," "self-contained," or "monolith" — and calcified into a locked
requirement. A project whose product *is* the methodology must retire its own
locked requirement with the ceremony it demands of adopters. Two of the four
clauses are load-bearing and survive.

| ID | EARS statement |
|---|---|
| R-8.1 | THE SYSTEM SHALL formally retire the single-file clause of `docs/ears/federation-enrichment.md:149` (R-7.4) as part of this change. |
| R-8.2 | THE SYSTEM SHALL formally retire the `die`/`ok`/`warn`/`fail` helper clause of R-7.4, superseded by the renderer architecture in Unit 2. |
| R-8.3 | THE SYSTEM SHALL keep the frontmatter-parser clause of R-7.4 in force, per R-1.5. |
| R-8.4 | THE SYSTEM SHALL keep the next-command guidance-chain clause of R-7.4 in force, per R-1.8. |
| R-8.5 | THE SYSTEM SHALL update every document repeating the single-file requirement: `CLAUDE.md:102`, `docs/lld/golden-path-flavor-federation.md:49`, `docs/lld/federation-enrichment.md:12`, `docs/lld/okf-v02-conformance.md:204-206`, and the two task files that restate it. |
| R-8.6 | THE SYSTEM SHALL record that a terminal UI shipped inside the binary, with no runtime dependency, is not covered by the "Web/GUI surfaces" non-goal at `docs/hld/golden-path-flavor-federation.md:47-52`. |
| R-8.7 | THE SYSTEM SHALL record that the four-times-stated dependency policy governs runtime prerequisites on the user's machine, and that build-time Go modules linked into a static binary do not violate it. |

---

## Unit 9: Retiring the Python implementation

**Why:** Two implementations kept in parity is a maintenance tax with no expiry
date. The port is worth doing only if it ends with one binary — but deleting the
reference implementation destroys the ability to generate an oracle for any defect
found afterwards, so the gate and the rollback both have to be real.

| ID | EARS statement |
|---|---|
| R-9.1 | IF and only if R-7.9 parity holds, THE SYSTEM SHALL permit deletion of `company-os-starter/bin/company-os`. |
| R-9.2 | BEFORE deletion, THE SYSTEM SHALL tag the final Python commit and document `git checkout <tag> -- company-os-starter/bin/company-os` as the recovery path, because after deletion no runnable reference implementation exists from which to generate a golden. |
| R-9.3 | WHEN parity is proven, THE SYSTEM SHALL delete `company-os-starter/bin/company-os`, `company-os-starter/install.sh`, `company-os-starter/vendor/`, and `examples/selftest.py`. |
| R-9.4 | THE SYSTEM SHALL handle the launcher `install.sh` generated on existing installs, which after deletion points at a file that no longer exists. |
| R-9.5 | THE SYSTEM SHALL NOT ship both implementations simultaneously in a released state. |
| R-9.6 | THE SYSTEM SHALL update every document referencing the Python invocation path, the `pip install pyyaml` prerequisite, the generated launcher, or `PYTHONPATH`, including `README.md:29-35`, `docs/GOLDEN-PATH.md:24`, and `docs/FEDERATION-RUNBOOK.md:448`. |
| R-9.7 | THE SYSTEM SHALL update the four shipped agent skills at `company-os-starter/skills/*/SKILL.md` to use `--json` and the exit-code contract, so the agent-facing capability delivered in Units 3 and 4 reaches the agents already in the repository. |
| R-9.8 | THE SYSTEM SHALL leave OKF v0.2 Phases 1–3 unimplemented, to be re-planned against the Go binary as a separate change; Phase 0 is implemented under R-1.14. |
| R-9.9 | THE SYSTEM SHALL record a named owner for the Go codebase before implementation begins, because there is no Go precedent in this repository and `local-search` being Go is precedent for the ecosystem rather than evidence of this team's capacity. |
