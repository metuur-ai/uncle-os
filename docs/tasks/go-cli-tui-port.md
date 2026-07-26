---
type: tasks
id: tasks-go-cli-tui-port
title: Go CLI + TUI Port — Tasks
status: draft
tags: [kind/tasks, status/draft]
---

# Go CLI + TUI Port — Tasks

Source of truth: `docs/ears/go-cli-tui-port.md` (Units 0–9).
Architecture constraints: `docs/lld/go-cli-tui-port.md`.
Intent: `docs/hld/go-cli-tui-port.md`.
Targets: new Go module under `company-os-starter/`, replacing
`company-os-starter/bin/company-os`.

**Global acceptance (must hold after every phase from P3 onward):** the Go
binary reproduces `examples/golden-validate.txt` and
`examples/federated-golden-validate.txt` byte-for-byte after `normalize()`, and
`graph build; graph build` is a no-op diff.

**Ordering rule.** Phase 0 is Python-only oracle repair and blocks every Go
task — the current oracle covers one command, only on the passing path, and
expires 2027-01-01. Phases 1–2 build the substrate bottom-up. Phase 3 is where
parity becomes measurable. Phase 6 is the irreversible one. Phase 7 (TUI) does
not start until 6.1 is green.

**Deferred — NOT planned here:** OKF v0.2 Phases 1–3 (re-plan against the Go
binary as a separate change). Phase 0 of OKF is in scope, at task 3.9.

---

## Phase 0 — Oracle repair, in Python, before any Go (Unit 0, Unit 7)

- [x] 0.1 Push fixture expiry dates out and re-baseline both goldens (est: ~30m)
  - DONE 2026-07-26. Only two `TODAY` comparison sites exist (`:959` deviations
    `reviewDate`, `:970` exceptions `expires`); the other 11 `TODAY` uses generate
    dates. 7 fixture files moved to 2035 preserving each value's original quoting
    — the unquoted `reviewDate: 2035-01-15` is deliberately left unquoted because
    task 1.2 needs it to exercise YAML 1.1 scalar resolution. Both goldens
    re-baselined; `acceptance.sh` PASS. Two follow-ups surfaced: (a)
    `examples/banking/bank/workspaces/team-fraud-detection` validates FAIL with 5
    pre-existing non-date problems (missing descriptors + no `workspace.lock.yaml`)
    — `acceptance.sh` never runs it, but task 0.3's harness will trip on it;
    (b) `examples/banking/bank/repos/platform-fraud/archive/prds/2026-alert-triage-queues/prd.md:41`
    narrates the old 2026-12-31 expiry in prose and is now inconsistent with the
    fixture it describes.
  - why: `TODAY` (`bin/company-os:31`) is compared against `expires: '2026-12-31'`
    in `examples/workspace/teams/customer-engagement/governance/exceptions.yaml:9`.
    On 2027-01-01 golden line 9 flips to `[FAIL]`, `PASS` becomes
    `FAIL — 1 problem(s)`, and both goldens break on their own — 158 days from
    now. Re-baselining mid-port would create exactly the "was it the port or the
    rule?" ambiguity that justified deferring OKF. Do it now, from Python, while
    a reference implementation still exists.
  - acceptance: R-0.0
  - verify: `examples/acceptance.sh` passes; `git diff` shows only fixture dates
    and the two golden files; the new expiry is at least 2 years out.

- [x] 0.2 Capture failure-path goldens from the Python CLI (deps: 0.1, est: ~90m)
  - DONE 2026-07-26. **All 15 render sites in `cmd_validate` now have an oracle**
    (14 `fail()` + the single `warn()` at `:1013`). Three new fixtures, three new
    goldens, wired into `acceptance.sh` as section 2c:
    `examples/failing-workspace/` → 15 findings covering all 13 `fail()` sites in
    gates 1–7 plus 4 `warn` lines; `examples/failing-federated/` → 4 of the 5
    `federated_slice_problems()` shapes at `:1100`;
    `examples/failing-federated-nolock/` → the 5th, which cannot co-occur because
    the absent-lock branch returns early. **Gate 4's conditional `[ok]`
    (`:1003-1008`) is frozen**: the reality doc omits `updated:` only, a field
    that feeds neither `derive_tags` nor the CLAUDE.md node, so it emits its
    `[FAIL]` with no `[ok]` and nothing else moves. All three gate-5 `[ok]`
    shapes (`:1030` absent, `:1036` hand-owned, `:1040` in sync) render too, and
    every gate emits at least one `[ok]` beside its failures so the snapshot
    freezes the interleaving, not just the failures. Two structural facts forced
    the fixture shape: gate 6 `continue`s after index drift (so drift and
    unresolved-reference findings need two platforms — `alpha` and `beta`), and
    the `kind='prd'` half of `:1062` is only reachable when an archived PRD's
    `id` differs from its directory name. **Known gap, deliberate:** `:1002`
    renders one of `core_field_errors`' five strings and `:1025` one of
    `identity_errors`' shapes — both are pure functions of a dict and are
    directly unit-testable in Go without an oracle, unlike render order, prefix
    policy, and the conditional, which are not.
  - why: neither committed golden contains a single `[FAIL]` or `[warn]` line,
    against 15 failure sites in `cmd_validate`. "Parity-gated cutover" currently
    rests on a happy-path snapshot. Once `bin/company-os` is deleted there is no
    runnable reference from which to generate one.
  - acceptance: R-0.9
  - verify: new golden fixtures exist exercising ≥1 `[FAIL]` per gate and ≥1
    `[warn]`; `acceptance.sh` diffs them; all pass against Python.

- [x] 0.3 Build the cross-implementation differential harness (deps: 0.1, est: ~3h)
  - DONE 2026-07-26 → `examples/differential.py`. **288 invocations, 331 command
    steps, all 16 subcommands, 38 distinct fixture trees** (9 committed, 29
    synthesized: bad manifests, empty dirs, the git source repo); Python-vs-Python reports zero
    divergence in ~51s. Each invocation runs against a pristine temp copy of its
    fixture (committed fixtures are never mutated) and compares stdout, stderr,
    exit code, and the whole resulting file tree (path + mode bits + content).
    Multi-step invocations run in one copy so lifecycles are reachable
    (`prd new → validate → complete`, `deviation declare → governance resolve`,
    `workspace sync → status → sync --frozen`). Coverage is weighted to the ten
    commands the 0.4 NOTE named: ids 26, today 21, governance 18, discover 16,
    check 14, graph 14, skills 12, deviation 10, exception 10, scratchpad 8.
    127 of 265 base invocations exercise ≥1 non-zero step, so the failure paths
    that will carry the 8-code contract are already frozen. Normalization is
    exactly two substitutions, applied to streams *and* to file bytes before
    hashing: the per-run temp workspace path → `<WS>`, and the generated UTC
    timestamp `NOW` (`:32`, surfaces as `generatedAt`) → `<TS>`. `TODAY`-derived
    dates are deliberately NOT normalized. Detection was proven, not assumed:
    four throwaway single-line mutations of the Python CLI were each caught —
    a stdout string (`today` header), a written-file body (`outcome.md`
    heading), an exit code (`die` 1→3), and a mode bit (slice `0444`→`0644`,
    which is what puts invariant #6 under differential coverage). Rulings:
    (a) the corpus DOES sweep `examples/banking/` and the 0.2 `failing-*`
    fixtures — `team-fraud-detection` fails identically on both sides and that
    is the evidence we want, not a defect; (b) the tree snapshot skips `.git`
    internals inside `.company-os/federation-cache` (reflog timestamps, pack
    nondeterminism — no CLI produced them); the materialized slices themselves
    are fully compared; (c) `workspace sync` is covered for real via a
    synthesized `file://` source repo pinned at a fixed-date commit, and SKIPs
    loudly per-invocation when git is absent or < 2.27. **One declared hole:**
    `workspace/sync-missing-ref` is PARTIAL — git relays the local `upload-pack`
    process's stderr on a separate pipe, so its two `not our ref` lines
    interleave in a non-deterministic ORDER between runs of ONE binary. Its
    stderr is excluded (exit code and file tree still compared) and the harness
    prints the exclusion with its reason in every report rather than folding it
    into PASS; the wrapping `` error: `git ...` failed (exit 128) `` line stays
    deterministically covered by `workspace/sync-bad-pin`. That is the only
    stream excluded anywhere in the corpus.
  - why: `acceptance.sh` byte-freezes one command out of sixteen, and `go test`
    is written against the Go implementation so it cannot detect a Go/Python
    divergence by construction. Without this harness, R-9.1's "parity is proven"
    is not a true statement, and `prd complete` — the command enforcing invariant
    #4 of the methodology — ships unverified.
  - acceptance: R-7.1, R-7.2, R-7.3a
  - verify: harness runs Python against Python over all 16 commands and all three
    fixtures, reporting zero divergence in stdout, stderr, exit code, and
    resulting file tree; it is wired to accept a second binary path.
  - NOTE from 0.4: this harness is the **only** oracle for `discover`,
    `deviation`, `exception`, `check`, `governance`, `today`, `graph`, `ids`,
    `skills`, and `scratchpad` — selftest never covered them. Weight coverage
    accordingly rather than spreading evenly.
  - NOTE from 0.1: `examples/banking/bank/workspaces/team-fraud-detection` is not
    exit-0 clean (5 pre-existing non-date problems). Decide whether the harness
    sweeps `examples/banking/` at all, or restricts to the three fixtures
    `acceptance.sh` uses. Divergence on a fixture that fails identically under
    both binaries is still parity — do not "fix" the fixture to make the harness
    green.

- [x] 0.4 Inventory the `selftest.py` assertions as a port checklist (deps: none, est: ~45m)
  - DONE 2026-07-26 → `.devlocal/go-port/selftest-inventory.md`. **Count corrected:
    86 `check()` call sites, 85 real assertions** (the 87th grep hit was the
    helper's own definition at `:21`; ST-076 is a skip sentinel). 10 are
    git≥2.27-conditional, so unconditional coverage is 75; ~34 are compound and
    fan out. Distribution: federation 38 (44%), scaffold 14, skills 11, graph 7 —
    `governance`, `validate`, `render`, and `model` inherit **zero**. Flagged: 14
    assertions test `die()`/`SystemExit` and must invert under R-2.10; 8 test
    private Python functions; 3 have no Go analogue and need an explicit ruling.
    **Brief correction:** selftest drives only **7 of 16** subcommands via
    subprocess. `discover`, `deviation`, `exception`, `check`, `governance`,
    `today`, `graph`, `ids`, `skills`, `scratchpad` were never covered — that gap
    belongs to 0.3, not to this port.
  - why: `R-9.3` deletes `examples/selftest.py`, which carries 87 `check()`
    assertions across 10 commands via real subprocess invocations. `R-7.4`'s
    "coverage for every internal package" is a promise, not a port. An explicit
    list is what makes the deletion safe.
  - acceptance: R-7.3 (checklist half)
  - verify: `.devlocal/go-port/selftest-inventory.md` lists all 87 assertions with
    command, behavior asserted, and target Go test name.

- [x] 0.5 Classify every failure path against the exit-code contract (deps: none, est: ~60m)
  - DONE 2026-07-26 → `.devlocal/go-port/exit-code-map.md`. **Count corrected: 52
    `die()` call sites + 5 `sys.exit` = 56 failure paths** (53 counted the `def`).
    Zero unclassified. Distribution: code 4 → 22, codes 3 and 6 → 10 each, code 8
    → 5, code 1 → 3, code 2 → 2, code 5 → 2, code 7 → 1, carve-out → 1. Four
    rulings that refine the LLD: `:2318` is 4 not 6 (git succeeded; the fault is
    an abbreviated SHA in `workspace.yaml`); `:2547` is 3 not 4 (absent, not
    malformed, and absence is legal in monorepo mode); `:2564` stays 6 as a
    deliberate asymmetry (malformed manifest = 4, malformed lock = 6, justified
    by hand-authored vs machine-generated); `:367` is the weakest 3 and collapses
    "component missing" with "governance resolve never run" — cheap to split in
    Go. Two contract defects found and fixed in the specs: code 4's headline case
    has **zero** sites today (bad YAML raises an uncaught `YAMLError` → traceback
    → exit 1), and code 5's "deviation aimed at a mandatory rule" is not an exit
    site at all.
  - why: the exit-code contract has eight categories and the file has 53 exit
    sites. Several do not map cleanly — "already exists" (`:417`, `:610`, `:1797`,
    `:1971`, `:2037`), not-found lookups, `--frozen` lock failures. Inferring the
    mapping during implementation guarantees inconsistency; agents will branch on
    these codes.
  - acceptance: R-4.11
  - verify: `.devlocal/go-port/exit-code-map.md` maps every `die()` line number to
    exactly one code from Unit 4; zero unclassified.

---

## Phase 1 — Substrate: module, YAML, frontmatter, workspace (Units 1–2)

- [x] 1.1 Scaffold the Go module and dispatch skeleton (deps: 0.1, est: ~60m)
  - DONE 2026-07-26. Module `github.com/metuur-ai/uncle-os/company-os-starter`
    (derived from `origin`, with the `company-os-starter/` subdirectory in the
    path so `go install …/cmd/company-os@latest` resolves — the LLD's
    `github.com/<org>/company-os` was a placeholder), `go 1.22`, toolchain
    go1.25.7, zero dependencies. 12 internal packages created per the LLD layout;
    10 are doc-only, `model` and `workspace` carry real types because the
    dispatch signature and the root-requirement exemption need them.
    `internal/tui/` deliberately absent (Phase 7). **The stdlib `flag` package is
    unusable here and that is the one non-obvious finding**: `flag.Parse` stops
    at the first non-flag token, so `prd new --platform p` would never see
    `--platform`, while argparse permutes. The parser is hand-rolled off a
    declarative spec table mirroring `:2661-2781` one sub-parser at a time —
    same order, positional names, `choices`, `required`, `nargs="?"`, and
    defaults (`--components ""`, `--role developer`). `--root/--json/--version`
    are pre-subcommand only, matching argparse's parent-parser scoping.
    `require_root`'s message is byte-identical to Python's, `/private` symlink
    resolution included (`filepath.EvalSymlinks` reproduces `Path.resolve()`;
    plain `Abs` would have diverged on every macOS temp dir the 0.3 harness
    creates). Exit codes live in `main` alone: usage → 2, `RequireRoot` → 3,
    `model.CodeOf` for everything else, unclassified errors → 1 to match
    Python's uncaught-traceback path. R-2.10 is enforced by an AST test
    (`cmd/company-os/architecture_test.go`) that bans `os.Exit`, `os.Stdout`,
    `fmt.Print*` and `log.Fatal*|Panic*|Print*` below `cmd/`, resolved through
    the file's import names so a local `log` variable cannot trip it; proven by
    injecting a violation into `internal/model` and watching it fail.
    **Known deferrals, both surface-only:** argparse's unique-prefix flag
    abbreviation (`--plat` → `--platform`) is not reproduced, and usage/error
    stderr text is clear but not argparse-byte-identical — neither is covered by
    R-1.1 ("surface only") or by any committed golden, but if the 0.3 harness
    starts asserting usage-error stderr, this is where to look.
  - why: everything below hangs off a uniform `(ws, args) → ([]GateResult, error)`
    entry point. The Python file already has this shape at `:2777`; establishing
    it first is what keeps `os.Exit` out of the tree.
  - acceptance: R-1.1 (surface only), R-2.11
  - verify: `go build ./...` succeeds; all 16 subcommands parse their flags and
    return "not implemented"; nothing below `cmd/` imports `os.Exit` or `fmt.Print`.

- [x] 1.2 `internal/yamlio` — YAML 1.1 scalar resolution (deps: 1.1, est: ~4h)
  - why: PyYAML is YAML 1.1, `yaml.v3` is YAML 1.2. Verified: `reviewDate:
    2027-01-15` unquoted resolves to `time.Time` in Go and renders as
    `2027-01-15 00:00:00 +0000 UTC`, breaking golden line 7. This affects
    *parsing*, so it produces wrong behavior in every gate, not just wrong bytes.
    It is the single most dangerous divergence in the port and it is invisible
    until a golden diff.
  - acceptance: R-1.6
  - verify: table test covering unquoted date, quoted date, `yes`/`no`/`on`/`off`,
    sexagesimal, and leading-zero octal; each resolves to the same rendered string
    PyYAML produces.

- [x] 1.3 `internal/yamlio` — `yaml.Node` round-trip with preserve-unknown (deps: 1.2, est: ~3h)
  - DONE 2026-07-26 → `internal/yamlio/document.go`. API: `Load` /
    `LoadFrontmatter` → `*Document`; `Root`, `Bytes`, `IsFalsy`; free functions
    `MapGet`/`MapSet`/`MapKeys`/`SeqAppend` plus `NewString`/`NewMapping`/
    `NewSequence`. No struct unmarshal anywhere. **Round-trip measured over all
    112 git-tracked YAML documents under `examples/` (63 `.yaml` + 49 markdown
    frontmatter blocks): 46 byte-identical, 66 not.** All 66 are frozen with
    their cause in `testdata/roundtrip-divergent.txt`; zero are
    uncharacterized. Three causes, all PyYAML-emitter policies yaml.v3 neither
    implements nor exposes a knob for: 41 × 80-column line folding, 24 ×
    indentless block sequence (`key:\n- item`), 1 × flow-context requoting (a
    plain `://` scalar inside `{...}` comes back single-quoted — the only style
    drift in the corpus, frozen in `roundtrip-restyled.txt`). Two more layout
    losses are pinned by unit test: a multi-space gap before a line comment
    collapses to one space, and a multi-line folded scalar is re-folded onto one
    line. **Every one is layout-only** — the node tree survives re-emit
    STRUCTURALLY identical (kind, tag, value, anchor, all three comment slots,
    shape) on 112/112, and emit reaches a fixed point after one pass, which is
    what `graph build; graph build` needs. This is the divergence task 2.4
    already anticipated ("makes a custom PyYAML-compatible emitter
    unnecessary"); the exposure is bounded because `rewrite_frontmatter_tags`
    (`:1353`) early-returns when tags already match, so R-0.10 does not depend
    on byte-identity. Modify fidelity is corpus-driven, not sampled: rewriting
    one top-level value across the 45 eligible byte-identical documents changes
    exactly one line each, with key order intact. Seam findings: `or {}` is
    Python **truthiness**, not a `None` check — `0`, `false`, `''`, `[]`, `{}`
    all collapse to an empty mapping (measured, 16 rows); `Load` is deliberately
    stricter than `yaml.Unmarshal` on two points PyYAML rejects and yaml.v3
    accepts silently — a multi-document stream (`ComposerError`) and an
    unconstructible date like `2035-02-30` (surfaced by walking every scalar
    through 1.2's `Resolve`). **One measured gap left open:** `type:\tprd` raises
    `ScannerError` under PyYAML and parses cleanly under yaml.v3 — closing it
    means reimplementing PyYAML's scanner, so the divergence is asserted in
    `TestSyntaxErrorPropagation` rather than hidden. `<<` merge keys are also
    not flattened (no committed artifact uses one). R-0.7a(b) verified:
    comments in all three slots survive a read-modify-write.
  - why: three independent reasons, all mandatory — unknown frontmatter keys must
    survive a tag rewrite (`selftest.py:44-50`); Go map iteration is randomized
    where Python dicts are insertion-ordered; and hand-authored files are
    read-modify-written by `deviation declare` (`:1121`) and `exception request`
    (`:1136`). Struct unmarshal breaks all three.
  - acceptance: R-1.7
  - verify: round-trip test on a frontmatter block with unknown keys, non-alpha
    key order, and mixed quote styles returns byte-identical output.

- [ ] 1.4 `internal/yamlio` — deterministic map ordering (deps: 1.3, est: ~90m)
  - why: two map-driven loops reach byte-frozen stdout. Gate 8 renders `[FAIL]`
    line order from `(lr.get("files") or {}).items()` (`:2521`), and
    `workspace.lock.yaml` (`:2614`) embeds a `files:` map that would re-emit in
    random order every sync, breaking the lock's own reproducibility claim.
  - acceptance: R-0.11
  - verify: emit the same lock 50 times, assert byte-identical every time; gate 8
    finding order stable across 50 runs.

- [x] 1.5 `internal/frontmatter` with a differential regex test (deps: 1.1, est: ~90m)
  - DONE 2026-07-26. 40-document corpus, every expectation **measured** against
    the real `frontmatter()` (loaded with `SourceFileLoader` per
    `examples/selftest.py:11-15`, vendored PyYAML on `sys.path`), transcript in
    `.devlocal/go-port/frontmatter-truth.md`. Two layers assert it: a frozen
    truth table and `TestDifferentialAgainstPythonOracle`, which re-runs Python
    live over the same corpus (`internal/frontmatter/testdata/oracle.py`) and
    skips rather than passes when the oracle is unavailable. 83 assertions, 0
    skips. **The dominant finding was not the regex but `Path.read_text()`**: it
    decodes UTF-8 (raising `UnicodeDecodeError`, so the function never returns)
    and applies universal-newline translation *before* the regex, which is the
    only reason a CRLF artifact parses at all — the pattern contains no `\r`.
    Decoding therefore lives inside `Parse`, not in the caller. Two measured
    surprises worth knowing downstream: `---\n---\n` is **rejected** (a blank
    line between fences is required), and `---\n---\n---\n` is accepted with
    group 1 = `---`. The `$` trap did not bite because the trailing `(.*)` is
    greedy and reaches end-of-text first — verified by mutant: Python's
    `(.*?)$` yields `'body'` where `(.*)$` yields `'body\n'`. The Go pattern is
    written `\A…\z` so a future edit cannot reintroduce it. **Seam:** this
    package stops at the fence and returns the block as raw bytes;
    `yaml.safe_load(...) or {}` and its `ScannerError` propagation are 1.3/1.4's.
    Test teeth proven by mutation — dropping `(?s)`, dropping the newline
    translation, or dropping the UTF-8 check each fails the suite.
  - why: R-7.4's frontmatter clause is one of the two clauses that survive
    retirement, so its semantics are inviolable. Go needs `(?s)` and Go's `$`
    anchor differs from Python's, which is the kind of difference that passes
    review and fails at runtime.
  - acceptance: R-1.5
  - verify: differential test against the Python regex over malformed, empty,
    trailing-newline, and CRLF documents; identical accept/reject decisions.

- [x] 1.6 `internal/workspace` with error-returning lookups (deps: 1.1, est: ~2h)
  - DONE 2026-07-26. All 8 methods of `Workspace` (`:211-263`) ported; the three
    `die()` sites now return `*workspace.NotFoundError`, which wraps a
    `*model.Error{Code: ExitWorkspace}` and carries `Kind`/`ID`/`Dir`, so `main`
    classifies by `errors.As` and never by substring — all three are code 3 per
    `.devlocal/go-port/exit-code-map.md`. All three messages were **measured**,
    not read: the Python CLI was run (vendored PyYAML on `PYTHONPATH`) for each
    path and its stderr frozen, then `differential_test.go` re-runs that oracle
    live on every `go test` and skips rather than passes when it is unavailable.
    **The dominant finding was that 1.1's `EvalSymlinks`-with-`Abs`-fallback is
    only half of `Path.resolve()`**: `EvalSymlinks` fails outright on a missing
    path, so a root that does not exist kept its symlinked prefix unresolved —
    measured, `Path('/tmp/nope').resolve()` is `/private/tmp/nope` while the
    fallback yielded `/tmp/nope`. Since `require_root`'s message names the root,
    every "root does not exist" error diverged on every macOS temp dir. `New` now
    walks up to the longest existing ancestor, resolves that, and re-appends the
    tail; `filepath.Abs` already collapses `..` lexically, which is what
    `Path('/tmp/nope/../other').resolve() == '/private/tmp/other'` shows Python
    does too. Three smaller measured points: `platform_dir`/`team_dir` test
    `Path.exists()`, not `is_dir()`, so a *file* of the right name is "found";
    `all_platforms`/`all_teams` need `os.Stat` per entry rather than
    `DirEntry.IsDir()` because `Path.is_dir()` follows symlinks; and
    `find_component` returns `(None, None)` on a miss, so it is a `found bool`,
    not an error. **Seam:** `FindComponent` returns the descriptor *path*, not
    its contents — `load_yaml` is 1.3/1.4's. **Two deliberate divergences:** exit
    3 replaces Python's hardcoded 1 (the contract), and where `platforms/` exists
    but is not a directory Python raises `NotADirectoryError` and exits 1 through
    a traceback while `subdirs` returns empty, because R-2.10 forbids exiting and
    neither function has an error return in Python. Test teeth proven by three
    mutants (single→double quotes in the message, the naive `Abs` fallback,
    `DirEntry.IsDir()`), each caught. **Gap, deliberate:** the `init`/`scratchpad`
    exemption lives in `cmd/company-os/main.go:49` and is asserted here only from
    this side (`RequireRoot` never exempts anything); an end-to-end exemption test
    belongs in `cmd/company-os` and does not exist yet.
  - why: this is not a transliteration. `require_root` (`:230`), `platform_dir`
    (`:238`), and `team_dir` (`:244`) all call `die()`, which R-2.10 forbids below
    the dispatch layer. All three change signature and the change ripples into
    every caller. `MANIFEST_NAME`/`LOCK_NAME` live here, not in
    `internal/federation`, or `workspace ↔ federation` is an import cycle that
    will not compile.
  - acceptance: R-1.2, R-1.3, R-2.10
  - verify: `go build ./...` with no cycle; root resolution honours `--root` >
    `$COMPANY_OS_WORKSPACE_ROOT` > cwd; all three lookups return errors.

- [x] 1.7 `internal/model` — `GateResult`, `Finding`, severities, exit codes (deps: 1.1, est: ~90m)
  - DONE 2026-07-26. Verified by reproducing `golden-validate.txt`,
    `failing-workspace-golden-validate.txt` and
    `failing-federated-golden-validate.txt` **byte-for-byte from records alone**,
    via a test-local renderer in `internal/model/model_test.go` (the real one is
    3.2). Four measured corrections to the LLD sketch, all now folded into
    `docs/lld/go-cli-tui-port.md`: (a) **`[]GateResult` is not the top of the
    model** — line 1 `validating workspace <root>` (`:924`) is underivable, so a
    `Report{Root, Gates}` wraps them; `[N/M]` stays `len(Gates)` and the trailer
    stays a `SevFail` count (warns do not count — the failing golden has 15 fails,
    4 warns, trailer 15). (b) **`Fields map[string]string` is wrong** — `:990`
    renders an *ordered list* inside the sentence (`missing frontmatter ['team',
    'components', 'governanceSnapshot']`) which a string map cannot hold, and
    R-2.2's "typed" requires JSON `1` not `"1"`; it is `map[string]any` with
    non-panicking `Str`/`Int`/`Strs`. Key order is meaningless, order *within* a
    slice value is load-bearing. (c) **The prefix table was wrong about gate 1**,
    which has the same three-shape property it attributed uniquely to gate 5:
    `:941` prefixes the *team* id and `:946` the component id **in single
    quotes**. (d) **No per-code prefix discriminator is needed** — all seven
    shapes are seven `Subject` *values* under one uniform rule, which the
    test-local renderer implements as a single branch to prove it. Consequence:
    `Subject` is render-ready text, so the clean id must also live in `Fields`.
    Three things needed no field at all: the leading blank line is `Ordinal > 1`;
    gate 4's and gate 6's conditional `[ok]` is the *absence* of a record; and the
    `:1013` warn is a loop producing N one-line findings, not one multi-line
    finding — its real constraint is that `Findings` stays one document-ordered
    slice, never bucketed by severity (failing golden `:26-31`).
  - why: a flat `[]Finding` cannot reproduce the golden —
    `examples/golden-validate.txt:11-12` is gate 3's header with zero findings
    under it, and a renderer driven only by findings cannot know gate 3 ran.
  - acceptance: R-2.1, R-2.2, R-2.4
  - verify: a `GateResult` with empty `Findings` renders its header; `Fields` is
    reachable from both renderers.

- [x] 1.8 Embed templates with `//go:embed` (deps: 1.1, est: ~45m)
  - DONE 2026-07-26. **Only one template is embedded**, because only one is read
    from disk: `templates/reality-component.md`. The disk files
    `templates/discovery-brief.md` and `templates/prd.md` are human reference
    copies whose placeholder text differs from `DISCOVERY_TEMPLATE` (`:378`) and
    `PRD_TEMPLATE` (`:470`) — embedding them would have changed what `discover
    new` and `prd new` write. Those two are Go constants
    (`internal/scaffold/template.go`), pinned to the Python module strings by
    `TestBuiltinsMatchPythonModuleStrings`, which regex-extracts both literals
    out of `bin/company-os` and compares. **The embed directive lives in
    `templates/embed.go`** (`package templates`) because `//go:embed` cannot name
    a parent directory, so nothing under `internal/` can reach
    `company-os-starter/templates/`; that package holds the embedded bytes and
    nothing else, so all resolution logic stays inside `internal/`, under
    `architecture_test.go`'s AST-enforced no-exit/no-stdout rule.
    `ResolveTemplate` is in `internal/scaffold`, matching the LLD's
    "`internal/scaffold/ … (+ embedded templates)`", and `internal/product` will
    import it for `discover new` / `prd new`. **All six provenance labels are
    frozen verbatim**, including `built-in DISCOVERY_TEMPLATE` and `built-in
    PRD_TEMPLATE`, which name Python identifiers that no longer exist: they are
    byte-frozen human-facing output under R-0.8, and ST-030 asserts the `prd new`
    line contains `PRD_TEMPLATE`. The Go constants keep idiomatic names; only the
    printed label quotes the oracle. R-6.7 verified empirically, not by
    inspection: the `company-os` binary and a compiled `internal/scaffold` test
    binary were each copied alone into an empty directory and run there — the CLI
    starts (`--version`, bare-invocation exit 2) and all three built-ins resolve
    with correct labels with nothing on disk beside the binary. Note `install.sh`
    copies `templates/` wholesale into the Python kit root, so `embed.go` rides
    along there inertly until `bin/` is deleted in Phase 6.
  - why: `_builtin_template` (`:526-529`) is the one place the CLI reads a file
    from beside the binary, and R-6.7 forbids it. Discovery and PRD templates are
    already module strings (`:522-525`).
  - acceptance: R-1.11, R-6.7
  - verify: binary runs correctly from a directory containing nothing else;
    workspace-relative overrides still resolve.

---

## Phase 2 — Command clusters, bottom-up by coupling (Unit 1)

Ordered by the coupling map in research §4c: most self-contained first.

- [ ] 2.1 `internal/scaffold` — init, add, reality, scratchpad (deps: 1.6, 1.8, est: ~5h)
  - why: pure leaf from the callee side, and it establishes the template and
    write paths every later cluster reuses. `init`'s atomic staging-directory
    behavior (`:1982`) is the only transactional write in the system and must
    survive.
  - acceptance: R-1.1 (these commands), R-1.10, R-1.13
  - verify: differential harness reports zero divergence for `init`, `add`,
    `reality new`, `scratchpad init`; aborted `init` leaves nothing behind.

- [ ] 2.2 `internal/skills` — four-layer merge, shadowing, extends (deps: 1.6, est: ~3h)
  - why: 169 lines, exactly 2 external call sites (both in gate 7). Cleanest cut
    after federation and a good early proof that the record model works — its
    `[ok]` line carries counts (`2 canonical, 0 team`) that must reach the text
    renderer through `Fields`.
  - acceptance: R-1.1 (`skills list`), R-2.3, R-2.12
  - verify: `skills list` matches Python byte-for-byte; gate 7's line renders its
    counts from `Fields`, not from a pre-composed string.

- [ ] 2.3 `internal/graph` — tags, feature-index, CLAUDE.md nodes, `rebuildGenerated` (deps: 1.3, 1.4, 1.6, est: ~6h)
  - why: highest fan-in of any non-validate cluster, reached from gates 4/5/6 and
    from `rebuild_generated` (`:1807`). `rebuild_generated` (6 call sites) is the
    mandatory bridge between the write path and the derive path and belongs here,
    with a one-way `scaffold → graph` dependency — placing it in `scaffold` is the
    natural wrong guess and creates a cycle.
  - acceptance: R-0.6, R-1.1 (`graph build`)
  - verify: `graph build; graph build` is a no-op diff; differential harness clean.

- [ ] 2.4 Change `write_feature_indexes`' idempotency guard to a semantic compare (deps: 2.3, est: ~30m)
  - why: the guard at `:1530-1537` is a **byte** compare against a fresh render,
    so Go's emitter signature rewrites every `feature-index.yaml` on first build
    and `acceptance.sh:76-89`'s `s0 == s1` fails. This four-line change matches
    what gate 6 already does at `:1053`, removes the exposure permanently rather
    than matching a signature that could drift again, and makes a custom
    PyYAML-compatible emitter unnecessary.
  - acceptance: R-0.7a(a), R-0.10
  - verify: `acceptance.sh` §4 double-build passes with Python-emitted committed
    files untouched; `git status` clean after `graph build` and `governance
    resolve`.

- [ ] 2.5 `internal/governance` — resolve, explain, deviations, exceptions (deps: 1.3, 2.3, est: ~5h)
  - why: `deviation declare` and `exception request` are the two read-modify-write
    paths on hand-authored YAML, where `yaml.Node` fidelity is load-bearing. This
    is also where comment preservation inverts behavior — PyYAML destroys comments
    today, Go keeps them, which is sanctioned by R-0.7a(b) rather than silently
    shipped.
  - acceptance: R-1.1 (these commands), R-1.7, R-0.7a(b)
  - verify: differential harness clean modulo the sanctioned comment difference;
    `effective-governance.yaml` regenerates with `git status` clean.

- [ ] 2.6 `internal/product` — discover, prd, check (deps: 2.3, 2.5, est: ~6h)
  - why: `prd complete` enforces invariant #4 of the whole methodology — a change
    is done only when reality is updated. It has no byte-level oracle today, which
    is precisely why 0.3's harness had to exist before this task.
  - acceptance: R-1.1 (these commands), R-2.12
  - verify: differential harness clean for all three `prd` actions, both `discover`
    actions, and both `check` kinds, on passing and refusing paths.

- [ ] 2.7 `internal/federation` — manifest, sparse-checkout, slices, lock (deps: 1.4, 1.6, est: ~7h)
  - why: ~510 lines and the most self-contained cluster (2 external callers), but
    it carries the fiddly filesystem work: `_make_readonly` (`:2354-2360`) uses
    `sorted(rglob, reverse=True)` so children are chmod'd before parents, and
    `filepath.WalkDir` is pre-order — Go must collect and reverse-sort explicitly
    or it will chmod a directory read-only before its contents.
  - acceptance: R-1.1 (`workspace sync|status`), R-0.11
  - verify: `--frozen` sync from lock with no network reproduces the committed
    slice tree with `0444`/`0555` modes; lock emission byte-stable across runs.

- [ ] 2.8 `internal/ids` and `today` (deps: 1.6, est: ~2h)
  - why: both flatten rich structures to prose today and are the two commands
    where `--json` has the most obvious immediate value.
  - acceptance: R-1.1 (`ids list`, `today`)
  - verify: differential harness clean across all six `--role` values and all four
    `ids list` filters.

---

## Phase 3 — Validate, renderers, and the parity gate (Units 0, 2)

- [ ] 3.1 `internal/validate` — the 7/8 gates returning `GateResult` (deps: 2.2–2.7, est: ~8h)
  - why: `cmd_validate` reaches into six clusters and is the largest function in
    the file (186 lines). It sits above everything and nothing depends on it,
    which is what makes it the last thing built and the first thing measured.
  - acceptance: R-0.4, R-0.5, R-2.1, R-2.7
  - verify: gate denominator computed at run time (7 vs 8 by fixture); gate 3
    renders its header with zero findings; gate 4 emits no `[ok]` for a document
    carrying core-field errors.

- [ ] 3.2 `internal/render/text` — the per-gate prefix policy (deps: 1.7, 3.1, est: ~5h)
  - why: there is no uniform prefix rule and pretending there is one is how a
    records refactor breaks the golden silently. Seven distinct shapes: gate 1
    prefixes component, 2 team, 3 a compound `platform/prd-id`, 4 path, 6
    platform, 7 and 8 nothing — and gate 5 alone uses three shapes (`:1030`,
    `:1036`, `:1040`). The leading blank line is a property of the gate header,
    present on every gate except the first.
  - acceptance: R-2.5, R-2.6, R-2.8, R-0.1, R-0.2
  - verify: both committed goldens reproduce byte-for-byte after `normalize()`;
    both failure-path goldens from 0.2 reproduce.

- [ ] 3.3 **Parity checkpoint** — differential harness against the Go binary (deps: 3.2, est: ~4h)
  - why: this is the moment the port becomes a measurable claim rather than an
    assertion. Everything downstream is gated on it.
  - acceptance: R-0.3, R-0.7, R-0.8, R-7.9 (first pass)
  - verify: harness from 0.3 reports zero divergence across all 16 commands, three
    fixtures, passing and failing paths; `acceptance.sh` passes unmodified.

---

## Phase 4 — Agent surfaces: JSON, exit codes, version (Units 3, 4, 6)

- [ ] 4.1 `internal/render/json` with `schemaVersion` and build id (deps: 3.2, est: ~4h)
  - why: agents are already first-class consumers driving the CLI from four
    shipped skills. An unversioned schema for a machine-facing contract is a
    breaking change with a fuse on it.
  - acceptance: R-3.1, R-3.2, R-3.3, R-3.4, R-3.5, R-3.9
  - verify: `--json` on every subcommand emits valid JSON with `schemaVersion` and
    build id; default output unchanged (3.3 still green).

- [ ] 4.2 JSON envelopes for guidance and for finding-less commands (deps: 4.1, est: ~2h)
  - why: R-3.2 forbids prose on stdout and R-1.8 requires every mutating command
    print its next step — left unresolved, `--json` silently deletes the system's
    best existing affordance for exactly the consumer it was written for. And
    `prd new` produces no findings at all, so its envelope has to be defined
    rather than defaulted to an empty document.
  - acceptance: R-3.6, R-3.7
  - verify: `prd new --json` emits a populated envelope naming what it created and
    a `guidance` field carrying the next command.

- [ ] 4.3 Wire the exit-code contract (deps: 0.5, 3.3, est: ~3h)
  - why: every non-zero path is indistinguishable today, so CI and agents parse
    stdout to tell drift from an expired exception. Note code 2 already exists —
    argparse exits 2 for a bad flag — so this documents as much as it introduces.
  - acceptance: R-4.1 through R-4.10
  - verify: a test per code; every `die()` site from 0.5's map reaches its assigned
    code; `acceptance.sh:62`'s zero/non-zero assertions unchanged.

- [ ] 4.4 `--version` and ANSI-free guarantee (deps: 1.1, est: ~90m)
  - why: for a binary distributed by copy with no package manager, a user cannot
    tell what they are running and no bug report is actionable. The Python CLI
    emits zero ANSI codes; nothing currently states that must stay true, and the
    goldens depend on it.
  - acceptance: R-6.8, R-3.10
  - verify: `--version` reports version and build id; no subcommand emits an
    escape sequence.

- [ ] 4.5 Document the exit-code contract (deps: 4.3, est: ~60m)
  - why: no such contract has ever existed, and adopters' CI branches on exit
    status today.
  - acceptance: R-4.12
  - verify: contract published in user-facing docs with all eight codes.

---

## Phase 5 — Distribution (Unit 6)

- [ ] 5.1 Release build matrix and checksums (deps: 3.3, est: ~3h)
  - why: `CGO_ENABLED=0` static linking is what makes "no runtime dependency"
    true rather than aspirational.
  - acceptance: R-6.1, R-6.2, R-6.10
  - verify: three artifacts build reproducibly; checksums published; `ldd`/`otool`
    confirm no dynamic dependencies.

- [ ] 5.2 macOS signing and notarization (deps: 5.1, est: ~4h)
  - why: a downloaded, unsigned binary carries `com.apple.quarantine` and is
    killed on first exec with "cannot be opened because the developer cannot be
    verified." The workaround is strictly harder than `pip install pyyaml` for the
    audience this change is justified on — the headline value inverts.
  - acceptance: R-6.3
  - verify: a **downloaded** darwin artifact runs on a clean Mac with no
    right-click-open and no `xattr` step. If notarization is not pursued, the
    workaround is documented and the HLD records it as an accepted cost.

- [ ] 5.3 Clean-machine install verification (deps: 5.2, est: ~2h)
  - why: SC4 is the only success criterion that tests the actual user journey, and
    verifying it against a locally built binary would pass while the real journey
    fails.
  - acceptance: R-6.4, R-6.5
  - verify: downloaded artifact, clean macOS arm64 and clean Linux amd64, neither
    with Python; every subcommand runs following only published docs.

- [ ] 5.4 Upgrade and version-skew position (deps: 4.4, est: ~90m)
  - why: R-6.4 covers copying the binary once. The federation model assumes shared
    workspaces, so two people on different builds is a real user-facing condition
    with no stated answer.
  - acceptance: R-6.9
  - verify: a newer binary operates on a workspace last written by an older one;
    the skew position is documented.

---

## Phase 6 — Cutover, irreversible (Units 7, 8, 9)

- [ ] 6.1 Port all 85 `selftest.py` assertions to named Go tests (deps: 0.4, 3.3, est: ~8h)
  - why: `selftest.py` is the only behavioral oracle covering `prd`, `discover`,
    `init`, `deviation`, and `exception`. Deleting it against an unquantified
    promise of package coverage is how a port loses the checks nobody remembered.
  - acceptance: R-7.3, R-7.4, R-7.5, R-7.6, R-7.7, R-7.8
  - verify: every line of 0.4's inventory maps to a passing named Go test; zero
    unported.

- [ ] 6.2 Retire R-7.4's single-file and helper clauses (deps: 3.3, est: ~2h)
  - why: a project whose product *is* the methodology cannot silently override its
    own locked EARS requirement. Two of the four clauses — frontmatter parser and
    guidance chain — are load-bearing and stay in force; retiring them by accident
    would be worse than not retiring anything.
  - acceptance: R-8.1, R-8.2, R-8.3, R-8.4, R-8.5, R-8.6, R-8.7
  - verify: `docs/ears/federation-enrichment.md:149` amended; all five repeating
    documents updated; `company-os validate` still exits 0.

- [ ] 6.3 Tag the final Python commit and document rollback (deps: 6.1, est: ~30m)
  - why: after deletion no runnable reference implementation exists from which to
    generate a golden for any defect found later. The tag is the only recovery
    path and it must exist before the delete, not after.
  - acceptance: R-9.2
  - verify: tag exists; `git checkout <tag> -- company-os-starter/bin/company-os`
    restores a working CLI.

- [ ] 6.4 **Parity gate** — final differential run (deps: 6.1, 6.3, mutex: cutover, est: ~2h)
  - why: R-9.1 permits deletion if and only if parity holds. This is where that
    condition is actually evaluated.
  - acceptance: R-7.9, R-9.1
  - verify: harness zero divergence, `go test ./...` green, `acceptance.sh` green,
    all four goldens reproduce.

- [ ] 6.5 Delete the Python implementation (deps: 6.4, mutex: cutover, est: ~60m)
  - why: two implementations in parity is a maintenance tax with no expiry date.
    The port is worth doing only if it ends with one binary.
  - acceptance: R-9.3, R-9.5
  - verify: `bin/company-os`, `install.sh`, `vendor/`, `selftest.py` gone;
    `acceptance.sh` still green against the Go binary.

- [ ] 6.6 Handle stranded launchers on existing installs (deps: 6.5, est: ~60m)
  - why: `install.sh` generated a bash launcher on every existing install that
    `exec`s a path which no longer exists. R-9.6 covers documentation, not
    installs.
  - acceptance: R-9.4
  - verify: documented migration step for an existing install; a stranded launcher
    fails with an actionable message rather than a bare `no such file`.

- [ ] 6.7 Update all documentation referencing the Python path (deps: 6.5, est: ~3h)
  - why: five distinct invocation patterns are documented and in use, all assuming
    a runnable Python script.
  - acceptance: R-9.6
  - verify: no doc references `pip install pyyaml`, `PYTHONPATH`, the generated
    launcher, or `bin/company-os` as an executable; `README.md:29-35`,
    `docs/GOLDEN-PATH.md:24`, `docs/FEDERATION-RUNBOOK.md:448` updated.

- [ ] 6.8 Update the four shipped agent skills for `--json` and exit codes (deps: 4.3, 6.5, est: ~2h)
  - why: the agent-facing capability delivered in Units 3 and 4 is invisible to
    the agents already shipping in this repository, which still instruct plain
    command invocation and prose parsing.
  - acceptance: R-9.7
  - verify: all four `company-os-starter/skills/*/SKILL.md` use `--json` and branch
    on documented exit codes.

- [ ] 6.9 Record the named Go codebase owner (deps: none, est: ~15m)
  - why: there is no Go precedent in this repository, and `local-search` being Go
    is precedent for the ecosystem rather than evidence of this team's capacity to
    maintain 4500–6000 lines of it.
  - acceptance: R-9.9
  - verify: owner named in the HLD Stakeholders section.

---

## Phase 7 — TUI, gated on Phase 6 (Unit 5)

Does not start until 6.4 is green. Read-only screens ship complete before any
mutating form is written.

- [ ] 7.1 Bubble Tea shell, `tui` subcommand, TTY gate (deps: 6.4, est: ~4h)
  - why: an agent or CI job must never be able to land in an interactive app, and
    a TUI that cannot be exited is worse than no TUI for an audience defined as
    blocked on terminal fluency.
  - acceptance: R-5.1, R-5.2, R-5.3, R-5.14, R-5.16
  - verify: `tui` with no TTY exits 7 and changes nothing; `q`, `Esc`, `Ctrl-C`
    exit from every screen; bare `company-os` still prints help.

- [ ] 7.2 Read-only screens, enumerated and tested (deps: 7.1, est: ~12h)
  - why: this subset serves what POs actually do — they read status far more often
    than they scaffold — and it needs none of the mutation machinery, which
    eliminates the entire class of "the TUI wrote the wrong thing" defect.
  - acceptance: R-5.4, R-5.13
  - verify: all ten named screens exist and are asserted by test; validate results
    render from the same records the text and JSON renderers consume.

- [ ] 7.3 Terminal robustness — `NO_COLOR`, narrow, resize (deps: 7.2, est: ~3h)
  - why: for an audience defined by terminal unfamiliarity, a UI that garbles
    below 80 columns fails on the machines it exists to serve.
  - acceptance: R-5.15
  - verify: legible at 60 columns; honours `NO_COLOR`; survives resize.

- [ ] 7.4 Derived command preview (deps: 7.2, est: ~4h)
  - why: this is the property that justifies interactive mutation at all — every
    TUI action reduces to a flag-complete invocation reproducible in CI. Derived
    from the args structure, never hand-written per screen, because a hand-written
    preview drifts from what actually runs and quietly destroys the guarantee.
  - acceptance: R-5.6, R-5.7, R-5.10
  - verify: preview string is generated from the same struct the command executes;
    a test asserts preview and execution cannot diverge.

- [ ] 7.5 First mutating forms — `discover new`, `prd new` (deps: 7.4, est: ~8h)
  - why: the two commands a product owner actually authors. No forms are built for
    `workspace sync` or `scratchpad init` — nobody will scaffold infrastructure
    through a form instead of typing the command.
  - acceptance: R-5.5, R-5.8, R-5.9, R-5.11, R-5.12
  - verify: no write occurs before confirmation; cancellation before confirmation
    leaves the workspace untouched; scaffolding delegates to the same code path
    `init`'s `_prompt` wizard uses, so GPF-R-1.3/1.4 hold for both.

- [ ] 7.6 Product-owner journey verification (deps: 7.3, 7.5, est: ~2h)
  - why: not one Phase 1 success criterion involves a member of the audience the
    TUI is justified on. If nobody will run this test, the PO justification is
    struck and the TUI is re-justified on engineer value alone — which is
    defensible, but it has to be said rather than assumed.
  - acceptance: HLD Success Criterion 9
  - verify: a product owner, given only a download link and no prior use, reads
    workspace status, opens a component's governance, and inspects a validate
    failure — unassisted, without prior shell configuration.
