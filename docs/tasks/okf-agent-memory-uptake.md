---
type: tasks
id: tasks-okf-agent-memory-uptake
title: OKF Agent-Memory Uptake — Tasks
status: draft
tags: [kind/tasks, status/draft]
---

# OKF Agent-Memory Uptake — Tasks

Source of truth: `docs/ears/okf-agent-memory-uptake.md` (Units 1–8).
Architecture constraints and measured anchors: `docs/lld/okf-agent-memory-uptake.md`.
Targets: `company-os-starter/internal/`, `company-os-starter/cmd/company-os/`,
`company-os-starter/docs/`, `company-os-starter/templates/`, and two repo-root docs
(`docs/01-flexibility-skills-and-role-views.md`, `docs/tasks/okf-v02-conformance.md`).

**Four things every task inherits:**

1. **The gate is `make check`** (gofmt + `go vet` + `go test ./...` +
   `acceptance.sh`), run from `company-os-starter/`.
2. **Golden freeze.** All five goldens under `examples/` stay byte-identical after
   every task. `acceptance.sh --update` is never run — a diff is a bug in an
   emission rule (LLD Decision 2), not a baseline to refresh.
3. **Renderer-only output.** Nothing below `cmd/` prints or exits. A new finding
   code means: constant in `internal/model/codes.go` under its host gate's block
   (R-8.1), one `case` per renderer that can emit it, `--json` fields covering
   every value the sentence shows.
4. **Templates move with emitters.** `templates/outcome-review.md` changes in the
   same commit as `outcomeDoc()` (task B.5).

**Slice order is the ordering rule** (LLD *Sequencing*, R-8.8): A → B → C → D.
Each slice is one PR, green on `make check` alone. C has the highest blast radius
and lands with its guard test (TC-S5) in the same PR. D is last because TC-S2
exercises everything from B and C.

**Global acceptance (must hold after every slice):** `make check` exits 0; the five
goldens are unchanged; `git status --porcelain examples/` is empty.

---

## Slice A — Definitions and docs (Units 7 + 6)

- [ ] A.1 Define version scheme, profile enum and log shape in `CONFORMANCE.md §6` (est: ~45m)
  - `company-os-starter/docs/CONFORMANCE.md` §6 currently says the version "is
    defined nowhere" (`:145-150`). Replace with: `YYYY.N`, numeric compare (year,
    then N), `2026.2` first and only value, forward-only bump (parent R-1.6);
    `profile` = `minimal | standard | strict` with the meanings in LLD Unit 7 and
    the sentence "gating by profile is not implemented (parent task 4.2)";
    compatibility rule (≤ V silent, > V or unparseable warns); one line: "`log.md`
    uses dated bullets; OKF §9 heading form is not used."
  - Verify: R-7.1, R-7.2, R-7.4, R-7.9 each have a sentence to point at.

- [ ] A.2 Reconcile the other two enum sites (est: ~10m)
  - `docs/01-flexibility-skills-and-role-views.md:121` comment → `# minimal |
    standard | strict`. `internal/scaffold/scaffold.go:196-197` already emits
    `2026.2` / `standard` — read, confirm, do not edit.
  - Verify: `grep -rn 'provisional' docs/ company-os-starter/docs/` returns nothing
    about `profile` (R-7.3).

- [ ] A.3 Gate-4 conformance check (est: ~1h)
  - New `internal/validate/versioncheck.go`, called at the tail of
    `frontmatterGate` (`validate.go:153-155`). For each `ws.AllPlatforms()`: absent
    `platform.yaml` or absent `conformance` block → nothing (R-7.7). Otherwise
    `frontmatter.conformance-version` (`path`, `declared`, `supported`) and
    `frontmatter.conformance-profile` (`path`, `declared`), both `SevWarn`.
  - Codes in the gate-4 block of `codes.go`; render cases in the validate renderer.
  - Tests: one platform each for `2026.2`/`standard` (silent), `2026.3` (warn),
    `banana` (warn), `profile: provisional` (warn), no `platform.yaml` (silent),
    `profile: minimal` (silent — the `examples/banking` case).
  - Verify: goldens unchanged — `examples/federated` has no `platform.yaml`; all
    nine fixture files declare `2026.2` with `standard`/`minimal`.

- [ ] A.4 `docs/SECURITY.md` (est: ~45m)
  - `company-os-starter/docs/SECURITY.md` with the five sections of R-6.1. Cite
    gate 9 and `internal/federation/lock.go:226` for integrity; cite the gate-4
    log-line check as "landing in slice C" (it does not exist yet — say so).
  - Add to `README.md`'s documentation list (R-6.2). The contract-rule-7 link
    lands in D.1.
  - Verify: the file exists; README links it; no other file changed.

- [ ] A.5 Close parent task 1.3 (est: ~10m)
  - `docs/tasks/okf-v02-conformance.md` task 1.3: append "**Update 2026-09-08.**
    Completed by `okf-agent-memory-uptake` slice A (tasks A.1–A.3): enum
    `minimal | standard | strict` per R-1.7, three sync points reconciled, gate-4
    warn check added." Match the existing 2026-08-26 note style. Tick the box only
    if every clause of parent R-1.7 is met — including "every shipped fixture
    conformant", which A.3's tests prove.
  - Verify: R-7.10.

## Slice B — Provenance and freshness (Units 1 + 2, TC-S6)

- [ ] B.1 Export `graph.VerifiedEntries` (est: ~15m)
  - Rename `mappingsOf` (`internal/graph/provenance.go:120`) to exported
    `VerifiedEntries`; update its callers in `graph`. No behaviour change.
  - Verify: `go test ./internal/graph/...`; `governance` can import `graph`
    (cycle-free — LLD Measured facts).

- [ ] B.2 `provenance_gate.go` — shape checks (est: ~2h)
  - New `internal/governance/provenance_gate.go`: `provenanceFindings(ws)
    []model.Finding` walking `graph.IterGraphDocs(ws)`; **swallow** the iterator
    error (R-1.11). Checks R-1.2…R-1.6 with the codes in LLD Unit 1 table; actor
    regex verbatim; new `parseProvenanceDate` (YYYY-MM-DD or RFC3339) — `before()`
    untouched (R-1.7). No `SevOK` lines (R-1.9). Each message begins `provenance:`
    (R-1.13).
  - `ExpiryGate` (`gates.go:142`) appends the result after the team loop.
  - Codes in the gate-2 block of `codes.go`; render cases.
  - Tests: one fixture per check; a doc with non-mapping frontmatter proves gate 2
    still completes and gate 4 still reports it; a doc with `generated.at: not-a-date`
    produces no finding (R-1.8).
  - Verify: five goldens byte-identical (R-1.12) — fixtures carry zero
    provenance frontmatter, so gate 2 emits nothing new.

- [ ] B.3 `stale_after` check + `--stale` (est: ~1.5h)
  - Same file: `stale_after <= today` → `expiry.stale` (`path`, `staleAfter`,
    `today`), `SevWarn`, promoted to `SevFail` when `opts.Stale`;
    unparseable → `expiry.stale-after-unparseable` (`path`, `value`), warn always.
  - `internal/validate`: `type RunOptions struct{ Stale bool }`;
    `RunWith(ws, RunOptions)`; `Run(ws)` = `RunWith(ws, RunOptions{})` (R-2.7 —
    the four existing `Run` callers do not change). Thread `Stale` to
    `ExpiryGate`.
  - `cmd/company-os/args.go`: `Stale bool` beside `Fix` (`:52`); flag entry beside
    `fix` (`:225`). `cmd/company-os/validate.go:31,48` → `RunWith(ws,
    RunOptions{Stale: args.Stale})`; `--fix` still runs `FixDerived` first (R-2.6).
  - Tests use `2000-01-01` / `2099-01-01` (R-2.12, no clock seam).
  - Verify: `validate --stale --fix` parses; goldens unchanged.

- [ ] B.4 Document `stale_after` in `FRONTMATTER-CORE.md` (est: ~20m)
  - Tier 2, beside `updated`, marked *optional*; heading amended to "required per
    doc family unless marked optional"; definition text from R-2.1 including "clear
    the signal by advancing or removing the field". Also add the one-line
    provenance-consistency note (what gate 2 now checks) and the sentence
    "derivation is not logged" (R-3.13 — written here so slice C is doc-complete
    when it lands).
  - Verify: R-2.1.

- [ ] B.5 `outcome.md` producer + template sync (est: ~30m)
  - `outcomeDoc()` (`internal/product/prd.go:639`): add `stale_after: <due>`
    directly under `due:` (R-2.8).
  - `templates/outcome-review.md`: add `stale_after: <YYYY-MM-DD>` **and** the
    `generated:` placeholder block it already lacks (R-2.9; pre-existing drift).
  - Update `internal/product` tests that assert `outcome.md` bytes.
  - Verify: no fixture runs `prd complete`, so goldens unchanged; template and
    emitter field sets match line-for-line.

- [ ] B.6 TC-S6 StaleAndProvenanceSignals (est: ~45m)
  - `cmd/company-os/scenarios_test.go` (new file, on `runArgs`,
    `selftest_test.go:30`): `init` a temp workspace, drop one doc with
    `stale_after: 2000-01-01`, `generated: {by: process:x, at: 2026-01-02}`,
    `verified: [{by: human:a, at: 2026-01-01}]`. Assert via `--json`: exit 0 with
    exactly one `expiry.stale` warn and one
    `expiry.provenance-superseded-verification` warn; `--stale` → exit ≠ 0,
    `expiry.stale` at `fail`, otherwise identical finding set (R-5.7).
  - Verify: R-5.7; `make check`.

## Slice C — Log discipline (Unit 3, TC-S5)

- [ ] C.1 `internal/logbook` (est: ~45m)
  - New package importing only `model`: `Append(path, date, text) error`
    (`O_APPEND|O_CREATE|O_WRONLY`, `0o666`, writes `- <date>: <text>\n`) and
    `Line(f model.Fields) string` for `CodeLogAppended` (R-3.1).
  - Move the sentence from `internal/product/sections.go:200-201` into
    `logbook.Line`; `product`'s case delegates.
  - Verify: `go vet` shows no cycle; unit test for exact bytes.

- [ ] C.2 Re-point the two existing writers (est: ~30m)
  - `appendLog` (`prd.go:652`, call `:414`) and `appendPromotionLog`
    (`promote.go:587`) call `logbook.Append` with their line text unchanged
    (R-3.2).
  - Verify: `examples/workspace/platforms/communications/log.md` unchanged;
    promote/complete tests pass without edits to their expected strings.

- [ ] C.3 Governance writers + struct fields (est: ~1h)
  - `Declare` (`declare.go:59`) and `Request` (`:101`): append per R-3.3/R-3.4 to
    `teams/<t>/log.md`; `Declared`/`Requested` gain `Log string`;
    `internal/render/governance.go` prints `log -> <path>` and emits `"log"` in
    JSON (R-3.15).
  - Verify: `--json` for both commands includes `log`; text shows one line.

- [ ] C.4 Product writers (est: ~1.5h)
  - `DiscoverNew` (`discover.go:27`), `DiscoverValidate` (`:96`), `PRDNew`
    (`prd.go:43`), `DraftNew` (`draft.go:88`), `PRDAbandon` (`abandon.go:40`) —
    lines per R-3.5…R-3.9; each adds one `CodeLogAppended` finding with `path`
    beside its `next:` finding without reordering it (R-3.14).
  - Verify: `next:` still last in each command's text output; goldens unchanged
    (no fixture runs these).

- [ ] C.5 Scaffold + federation writers (est: ~1h)
  - `scaffold.Add` (`commands.go:367`) → R-3.10; `RealityNew` (`:474`) → R-3.11;
    `federation.Sync` (`sync.go:93`) → `<root>/log.md` per R-3.12. Render cases in
    the scaffold and federation renderers call `logbook.Line`.
  - Verify: `init`, `graph build`, `validate --fix`, `skills install` write no
    `log.md` (R-3.13) — assert in a test.

- [ ] C.6 Gate-4 log-line shape check (est: ~45m)
  - New `internal/validate/logformat.go`, called at the tail of `frontmatterGate`
    after A.3's check: `<root>/log.md` and each `NodeRoots(ws)` root; non-blank
    lines must match `^- \d{4}-\d{2}-\d{2}: \S`; else
    `frontmatter.log-line-format` (`path`, `line`, `text` ≤ 80) `SevWarn` (R-3.16).
    `log.md` stays in `skipNames` (R-3.17).
  - Verify: all four fixture `log.md` files conform (measured), so goldens
    unchanged; a test with a bad line produces exactly one warn.

- [ ] C.7 TC-S5 LogTrail (est: ~45m)
  - In `scenarios_test.go`: run the TC-S2 chain (build it here; D.4 reuses it)
    and assert one new `- YYYY-MM-DD: ` line per mutating step in the `log.md`
    named in LLD Unit 3, none for the failing `prd complete`, and zero
    `frontmatter.log-line-format` findings in the final `validate` (R-5.6).
  - Verify: R-5.6; `make check`.

## Slice D — Contract and scenarios (Units 4 + 5, TC-S1…S4)

- [ ] D.1 Contract text + `WriteContract` (est: ~1.5h)
  - `internal/scaffold/contract_agents.md` (nine rules per R-4.2; rule 1 names
    `company-os find`, rule 3 `company-os derive`, rule 7 links
    `docs/SECURITY.md`), embedded via `go:embed`. `internal/scaffold/contract.go`:
    `WriteContract(root) (ContractResult, error)` with the absent/marker/append
    rule per file (R-4.1); `CLAUDE.md` block = `@AGENTS.md` + one comment (R-4.3).
    Marker `company-os:agent-contract:*` — confirm no match against `startRE`/
    `endRE` (`node.go:38-39`).
  - Tests: absent → written; marker present → bytes identical; present without
    marker → original bytes + `\n` + block; never truncates.
  - Verify: R-4.1, R-4.2, R-4.3, R-4.7.

- [ ] D.2 Hook into `init` and `skills install` (est: ~1h)
  - `scaffold.Init` (`commands.go:91`): after `moveRoots` (`:128`), call
    `WriteContract(target)`; on error emit `scaffold.contract-not-written`
    (`path`, `error`) as `SevWarn` naming `company-os skills install`; do **not**
    fail `init` (R-4.4, R-4.5). Code in a `scaffold.*` block of `codes.go`.
  - `skills.Install` (`install.go:81`): call `WriteContract(ws.Root)` after the
    skills loop, before `rebuild`; `InstallResult.Contract`; renderer prints one
    line per file (R-4.6).
  - Add `AGENTS.md` to `skipNames` (`tags.go:52`) (R-4.8). Do not add root to
    `NodeRoots` (R-4.9).
  - Verify: `init` in a dir with a pre-existing `CLAUDE.md` appends; `skills
    install` on `examples/workspace` copy writes both files and `validate` still
    passes; two `graph build` runs → identical tree hash (R-4.11); goldens
    unchanged (fixtures are not re-scaffolded).

- [ ] D.3 TUTORIAL paragraph (est: ~15m)
  - `docs/TUTORIAL.md`: one paragraph after `init` output showing `AGENTS.md` /
    `CLAUDE.md`, and one sentence at `skills install` noting it also writes the
    contract into an existing workspace. Use real command output.
  - Verify: output pasted from a real run, not typed.

- [ ] D.4 TC-S1…TC-S4 (est: ~2h)
  - TC-S1 FreshWorkspaceContract (R-5.2): anchor phrase per rule — keep the nine
    anchors in a table in the test so rule edits fail loudly.
  - TC-S2 SkillChainFromInit (R-5.3): reuse C.7's chain; drive `next:` from
    `--json` `FieldNext`; on the first `prd complete` assert non-zero + a
    `CodeDoneFix` finding and execute its `fix:` verbatim; second `prd complete`
    exit 0; final `validate` PASS; `outcome.md` `stale_after == due`.
  - TC-S3 AntiDuplication (R-5.4): `ExitConflict`, error names existing path,
    first `brief.md` byte-identical, directory listing otherwise unchanged.
  - TC-S4 HumanCorrectionPreserved (R-5.5).
  - Verify: `go test ./cmd/... -run TestScenario` green; no LLM anywhere (R-5.9).

- [ ] D.5 `docs/AGENT_TESTING.md` (est: ~30m)
  - Six rows (TC-S1…S6) with pass criteria; the OKF definition of done
    (`ROADMAP.md:516-538`); which TC-S2 steps are skill-driven (`add component`,
    the retry after `fix:`) vs `next:`-driven; "new scenario ⇒ new row" (R-5.8).
  - Verify: every scenario in `scenarios_test.go` has a row and vice versa.

---

## Dependencies

```
A.1 ─┬─ A.2 ─ A.3 ─ A.5
     └─ A.4
B.1 ─ B.2 ─ B.3 ─ B.6
              B.4 (docs, parallel)   B.5 (parallel; needs B.3's field name only)
C.1 ─ C.2 ─ C.3 ─ C.4 ─ C.5 ─ C.6 ─ C.7     (C.7 needs B.5 for stale_after in outcome.md)
D.1 ─ D.2 ─ D.3
        └── D.4 ─ D.5                        (D.4 needs C.7's chain and B.6's file)
```

## Estimates

| Slice | Tasks | Est. |
|---|---|---|
| A | 5 | ~3h |
| B | 6 | ~5.5h |
| C | 7 | ~6h |
| D | 5 | ~5.5h |
| **Total** | **23** | **~20h** |

## Not planned here

- T8 MCP shim (declined). `profile`-based gating (parent 4.2). Fixture backfill of
  `stale_after`/`generated:`. Skill file edits. LLM-driven scenarios. Clock seam.
