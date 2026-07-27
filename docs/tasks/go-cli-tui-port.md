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

**Phase 2 running order, revised after the mid-implementation review.** 2.4a
first and alone — it carries a reachable panic and a type-corrupting float bug in
code that 2.5 and 2.7 are about to be written against. Then **2.7 in parallel
with 2.3**: they share zero code, 2.7 is the longest single pole (~510 lines) and
carries 38 of the 85 inherited selftest assertions (44% of all inherited
coverage), so holding it until last puts the best-instrumented cluster between
the port and 3.1. Then 2.4 → 2.5 → 2.6. Do not defer 2.4; `acceptance.sh` §4
breaks the moment 2.3 lands without it.

**The riskiest thing still unbuilt is 2.5's `deviation declare` / `exception
request`** — the only two commands that read-modify-write a hand-authored file,
and therefore the only two that will trip R-0.7a(g) by design. They are blocked
on 0.6 having a waiver mechanism, not on their own logic.

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

- [x] 0.6 Declared-divergence registry in the differential harness (deps: 0.3, blocks 3.3, est: ~90m)
  - DONE 2026-07-26. Registry is data, in `examples/declared-divergences.txt`
    (blank-line-separated `key: value` records, no new dependency), loaded and
    enforced by `examples/differential.py`. 36 entries: 16 `exit_code`, 20
    `stderr`. **Invocation ids are exact — globs are deliberately unsupported**,
    because a glob firing on 7 of its 8 matches looks healthy while hiding the
    eighth; exact ids also make staleness decidable per entry. `authority:` is
    required *and shape-validated* (`R-0.7a(<a-k>)` | `R-<n>.<n>[a]` |
    `.devlocal/go-port/exit-code-map.md:<n>`), so an uncited waiver cannot load.
    `exit_code` entries carry `expect: <ref> -> <cand>` and waive only that exact
    transition — a wrong code is still a hard DIVERGE, proven by forcing 5 where
    3 was declared. stderr entries use `waive: usage-block`, which drops only the
    lines *before* the first `company-os…: error:` line on both sides and then
    compares the remainder: R-0.7a(i) waives argparse's COLUMNS wrapping, R-1.4a
    does not waive the diagnostic. Today that collapses a 24-line block diff to
    the one-line diagnostic diff task 1.1a must close — 19 of the 20 stderr
    entries are therefore still DIVERGE (correctly), and `usage/no-args` is
    already DECLARED. Report distinguishes PASS / DECLARED / DIVERGE / STALE /
    SKIP; every DECLARED prints its citation, and `--list-waivers` dumps the
    registry. STALE = the entry's invocation ran and no longer diverges on that
    stream; **not evaluated in self-check mode** (reference *is* candidate, so no
    entry can fire and all 36 would false-positive) — suppressed loudly in the
    header and the summary, never silently. Nothing is declared for the
    unimplemented commands.
    **Two unrecorded behavior changes found, both need an R-0.7a amendment and
    are left undeclared so they stay visible:** (1) `--help` output diverges on
    **stdout** (`usage/help`, `usage/validate-help`, `usage/prd-help`) but
    R-0.7a(i) is scoped to *stderr*, so nothing sanctions it; (2) the top-level
    `usage:` line, the flag summary, and the per-command list are a Go-authored
    rewrite, not a re-wrap of argparse's — a content change, which (i)'s
    "*content* is not carved out" arguably forbids.
  - why: `examples/differential.py` exits 1 on any DIVERGE and has no waiver
    mechanism. Every sanctioned carve-out currently lives in prose in a task-file
    bullet, while 3.3's acceptance is literally "zero divergence" — and R-0.7a(g)
    alone sanctions a re-emit divergence on 66 of 112 committed documents, which
    `deviation declare` and `exception request` will trigger by design. Without a
    registry, 3.3 is not a checkpoint, it is a negotiation.
  - acceptance: R-7.1a
  - verify: registry keyed by `(invocation id, stream)`, each entry citing an
    R-0.7a clause or an `exit-code-map.md` line; harness fails on an undeclared
    divergence AND on a stale entry whose invocation now passes; prove the stale
    check by declaring a divergence that does not exist.

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

- [x] 1.4 `internal/yamlio` — deterministic map ordering (deps: 1.3, est: ~90m)
  - DONE 2026-07-26. **The headline finding is that the required order is NOT
    sorted** — R-0.11's verb and the LLD's "explicitly sorted" prescription were
    both wrong for two of the three sites, and implementing them as written
    breaks a committed golden. Measured, not reasoned: rebuilding
    `examples/federated/workspace.lock.yaml`'s `files:` map through the REAL
    `hash_tree`/`_materialize_all` (loaded with `SourceFileLoader`, vendored
    PyYAML on `sys.path`) reproduces the committed key order exactly, and
    `sorted()` over the same keys does not — the lock records `governance/…`
    before `components/…` because that is the manifest's `paths:` list order.
    Gate 8 then iterates `safe_load` of that lock, so its `[FAIL]` order is the
    lock's DOCUMENT order; confirmed against the live CLI by swapping the two
    `files:` lines in a scratch copy of `examples/failing-federated`, which
    swapped the two `[FAIL]` lines. Only gate 6 is sorted, and incidentally —
    `build_feature_index` (`:1440`) iterates `cids = sorted(...)`, so the loop
    inherits sorted order rather than imposing it. **Two orders live in one key
    set:** emission uses walk order while `aggregate_hash` (`:2436`) uses
    `sorted(files)`, a plain string sort — hence `Keys` and `SortedKeys` are
    separate methods, not an argument. A third order sits underneath: CPython's
    `sorted(Path)` compares **component-wise**, not as raw strings, so
    `sdd/adr/a.md` precedes `sdd/adr-x.md` where `sort.Strings` reverses them;
    `PathLess` reproduces it and agrees with CPython 3.12.11 on 208/208 groups
    (8 frozen + 200 randomized) via `testdata/pathorder_oracle.py`, which skips
    rather than passes when Python is unavailable. Primitives: `MapPairs`
    (ordered node iteration — the ergonomic path, so no Go map is involved at
    all), `OrderedMap` (insertion-ordered builder with `dict` assignment and
    `dict.update` semantics; its internal index is a Go map that is never ranged
    over), and `PathLess`/`SortPaths`. Test teeth proven by mutation: degrading
    `PathLess` to `sort.Strings`, or sorting `OrderedMap.Keys`, or sorting
    `MapPairs` — i.e. implementing the LLD as written — each fails the suite
    against the real fixture or the real golden.
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

- [x] 2.1 `internal/scaffold` — init, add, reality, scratchpad (deps: 1.6, 1.8, est: ~5h)
  - DONE 2026-07-26. Differential over the 35 invocations these four commands
    own: **18 PASS / 18 DIVERGE**, and **every divergence is accounted for, none
    is in this port**. Measured twice — once against `bin/company-os`, once
    against a reference that loads it unmodified and replaces exactly one module
    attribute, `rebuild_generated = lambda ws: None`. The second run has **zero
    FILE DIFF and zero FILE TREE divergence across the whole corpus**, and the
    only diverging steps are failure paths and `validate`. Residual 18: 9 exit
    codes (all exactly as `.devlocal/go-port/exit-code-map.md` assigns — 7×2 for
    `_prompt` non-TTY, 8×4 for the conflict sites, 3×2 via `PlatformDir`, 2×1
    for `add component` without `--platform`), 4 argparse usage stderr (task
    1.1's declared deferral), 5 whose diverging step is `validate` (task 3.1).
    The other 10 are `rebuild_generated` alone. **The dominant finding was that
    a real PyYAML-compatible emitter is not optional here**: `register_id`
    (`:1815`) re-dumps the WHOLE `ids/registry.yaml` through `safe_dump`, so one
    `add` rewrites seven flow-style entries into block style on
    `examples/workspace` — arbitrary authored content flowing out through an
    emitter whose bytes the harness compares. `internal/scaffold/pyemit.go`
    transliterates `analyze_scalar` and the five writer primitives from
    `vendor/yaml/emitter.py`; a "wrap at 80" approximation is wrong on
    `team.yaml`, whose 85-column `precedence:` line PyYAML does NOT fold because
    no space on it sits past column 80. Proven by a live oracle over 12
    documents plus the 7 artifacts `init`/`add` actually write, with teeth shown
    by three mutants (width 80→60, indentless off, implicit-resolution quoting
    off) — each caught. **The one real defect found and fixed was TTY
    detection**: `os.Stdin.Stat()` + `os.ModeCharDevice` reports TRUE for
    `/dev/null`, so CI would have been treated as interactive; `cmd/company-os/
    tty.go` does the termios ioctl instead. Second measured surprise: the
    printed root is `Path.resolve()`d, so `init` prints `/private/var/...` on
    macOS. **Seam:** `scaffold.Rebuild func(*workspace.Workspace) ([]string,
    error)` — a parameter, not an import, because the LLD's one-way
    `scaffold → graph` edge would otherwise block this task on 2.3. It returns
    LINES because the order is load-bearing (rebuild's `  wrote index …` /
    `  node …` precede `added platform 'x'`) and only `cmd/` may print. **The
    dispatch seam changed**: `Command` now takes an `io.Writer`, since every
    mutating command prints prose rather than findings. `scratchpad init` still
    prints one line and no next step (R-1.9 over R-1.8), pinned by test.
  - why: pure leaf from the callee side, and it establishes the template and
    write paths every later cluster reuses. `init`'s atomic staging-directory
    behavior (`:1982`) is the only transactional write in the system and must
    survive.
  - acceptance: R-1.1 (these commands), R-1.10, R-1.13
  - verify: differential harness reports zero divergence for `init`, `add`,
    `reality new`, `scratchpad init`; aborted `init` leaves nothing behind.

- [x] 2.2 `internal/skills` — four-layer merge, shadowing, extends (deps: 1.6, est: ~3h)
  - DONE 2026-07-26. `skills list` is **10/12 PASS** on the differential; the two
    DIVERGE entries are `skills/bad-action` (argparse error text) and
    `skills/not-a-root` (exit 1 -> 3, the sanctioned `require_root` change in the
    exit-code map) — both are dispatch-layer shapes that diverge identically for
    the already-landed `ids` and `today`, and neither is reachable from this
    cluster. Because those two are the only harness-visible holes, the
    comparison was ALSO made at the library seam: `oracle_test.go` diffs the Go
    output against `bin/company-os` over all 9 committed fixtures plus two
    shapes the corpus reaches only through another cluster's command (a
    populated personal-rules layer, and a RESOLVED `extends` — no committed
    fixture has one, only a dangling one). `gate_oracle_test.go` does the same
    for gate 7 through Python's own `validate` over 13 synthesized workspaces,
    which is what covers `skill_conflicts` shapes no fixture reaches: id reuse
    under a different file name, one offender shadowing two canonical skills, a
    canonical TEAM skill (counts as canonical, cannot be shadowed), an id-less
    pair, and a non-canonical company skill. Three subtleties the merge hid:
    (a) `n_can` (`:1085`) filters by **authority, not layer**, while the
    shadowing target list (`:842`) additionally requires company/platform — a
    canonical team skill is counted but not protected; (b) `s["id"] and ...`
    (`:852`) is load-bearing, since without it every id-less team skill would
    shadow every id-less canonical one via `None == None`; (c) `parse_skill_steps`
    uses Python's Unicode `\s`/`\d` and `str.splitlines()`, so Go's `regexp`,
    `strings.Split` and `TrimSpace` all under-match — measured against CPython
    and reproduced in `pysem.go`.
  - why: 169 lines, exactly 2 external call sites (both in gate 7). Cleanest cut
    after federation and a good early proof that the record model works — its
    `[ok]` line carries counts (`2 canonical, 0 team`) that must reach the text
    renderer through `Fields`.
  - acceptance: R-1.1 (`skills list`), R-2.3, R-2.12
  - verify: `skills list` matches Python byte-for-byte; gate 7's line renders its
    counts from `Fields`, not from a pre-composed string.

- [x] 2.3 `internal/graph` — tags, feature-index, CLAUDE.md nodes, `rebuildGenerated` (deps: 1.3, 1.4, 1.6, est: ~6h)
  - DONE 2026-07-26, together with 2.4. Four files — `tags.go`, `featureindex.go`,
    `node.go`, `graph.go` — plus `internal/render/graph.go`. `cmd/company-os`
    wires the existing `scaffold.Rebuild` seam to `graph.Rebuild` rendered
    through `render.Graph` into a buffer, so the five sentences `graph build`
    prints are composed in exactly one place and the scaffolding commands emit
    them ahead of their own output. Byte parity against the Python binary is
    pinned by `TestBuildMatchesPythonBinary` over all 7 monorepo fixtures
    (stdout + file tree).
    **`rewrite_frontmatter_tags` forced a real emitter addition.** safe_dump
    runs there with `default_flow_style=None`, not `False` — which is why
    committed frontmatter reads `tags: [a, b]` inline and `pointers:` stays
    block with flow-style items. `internal/yamlio` had only the `False` form,
    so writing block style would have rewritten every document on the first
    build. Added `PyDumpAutoFlow` (flow writers, `flow_level`,
    `allow_flow_plain`, the `best_style` rule) and `PyDumpCanonical`
    (`sort_keys=True` + `allow_unicode=True`), both pinned to the vendored
    PyYAML in `pyflow_test.go`.
    **Two order traps confirmed by measurement, not assumption:** `sorted(rglob)`
    is PurePath order (`yamlio.SortPaths`) while `group_docs_by_root`'s
    `sorted(v, key=x[0])` sorts the relative path as a plain STRING — the walk
    order and the rendered order are different orders over the same data.
    **`iter_graph_docs` tests `"scratchpad" in md.parts` on the ABSOLUTE path**
    while `iter_knowledge_docs` tests the workspace-relative one; the asymmetry
    is reproduced, not fixed, because R-0.7 makes the oracle the contract.
    Exports what gates 4/5/6 will need (`IterGraphDocs`, `DeriveTags`,
    `BuildFeatureIndex`, `FeatureIndexUnresolved`, `ExtractGeneratedBlock`,
    `BuildClaudeNode`, `GroupDocsByRoot`, `NodeRoots`, `RootTeamMeta`,
    `IdentitySummary`, `PointerErrors`); `identity_errors` (`:1560`) is
    deliberately left for task 3.x, since gate 5 is its only caller.
    **Finding for whoever writes gate 5:** `rewrite_generated_block` is NOT a
    fixed point after one pass on a node it had to CREATE or APPEND to. Those
    two branches end their write with `"\n"`; the REPLACE branch splices in
    `text[ends[0].end():]`, and `END_RE`'s trailing `\s*$` has already eaten
    that newline. So such a node is rewritten once more on the next build and
    only then settles. `examples/failing-workspace` and
    `examples/failing-federated` each ship a marker-less CLAUDE.md and take
    that path; Python does the identical thing, on identical bytes — measured,
    and pinned by `TestBuildConvergesLikePython`. This is why acceptance.sh §4
    covers only `workspace` and `standalone-team`, whose nodes are already
    marked.
  - why: highest fan-in of any non-validate cluster, reached from gates 4/5/6 and
    from `rebuild_generated` (`:1807`). `rebuild_generated` (6 call sites) is the
    mandatory bridge between the write path and the derive path and belongs here,
    with a one-way `scaffold → graph` dependency — placing it in `scaffold` is the
    natural wrong guess and creates a cycle.
  - acceptance: R-0.6, R-1.1 (`graph build`)
  - verify: `graph build; graph build` is a no-op diff; differential harness clean.

- [x] 2.4 Change `write_feature_indexes`' idempotency guard to a semantic compare (deps: 2.3, est: ~30m)
  - DONE 2026-07-26, in the same session as 2.3 — the guard landed with the
    writer rather than after it, so no build ever ran with the byte compare.
    `WriteFeatureIndexes` compares `yamlio.PyDumpCanonical` of the parsed
    committed document against the same of a fresh derivation, which is
    literally what gate 6 does at `:1053`. `TestWriteFeatureIndexesGuardIsSemantic`
    proves both halves: an index re-laid with reversed top-level keys and flow
    style (every byte different, structure identical) is left alone, and a
    structurally drifted one is still regenerated.
    **R-0.7c's other two sites, since this task only names one.**
    `register_id`'s `ids/registry.yaml` was still unguarded — 2.4a landed the
    emitter relocation but not the guard — so it is guarded here:
    `internal/scaffold/scaffold.go` now returns without writing when the id is
    already registered, which is the one branch where `data` is provably
    `loaded` itself. Verified: `add component <existing-id>` leaves the file
    byte-unchanged where Python reflows it.
    `resolve_team_governance`'s `generated/effective-governance.yaml`
    (`:329-330`) is **NOT reachable** — `internal/governance` is still a stub
    and belongs to task 2.5. R-0.7c is therefore two-thirds met; 2.5 owns the
    last third, and its own verify line ("`effective-governance.yaml`
    regenerates with `git status` clean") is exactly that requirement.
    Consequently R-0.10 is verified for `graph build` on `examples/workspace`
    only; the `governance resolve` half waits on 2.5.
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

- [x] 1.1a Restore the argument-error diagnostic (deps: 1.1, blocks 3.3, est: ~2h)
  - why: task 1.1 deferred argparse's usage stderr as "not byte-identical." Half
    right. Measured, argparse wraps to `COLUMNS` — `COLUMNS=200` emits usage on
    one line, the non-TTY default of 80 wraps it across three — so byte-parity is
    parity against an environment variable and was correctly deferred. But Go does
    not merely differ in bytes: it **drops the diagnostic entirely**. Python names
    the subcommand, the offending argument, its value, and the valid choices; Go
    prints a generic top-level usage block. That removes a human-facing line on
    18 of 193 harness invocations, on a stream four shipped agent skills read.
  - acceptance: R-1.4a, R-0.7a(i)
  - verify: `company-os skills bogus` names the subcommand, argument, value, and
    choices on one line; the usage *block* is waived in 0.6's registry while the
    error *line* is compared exactly.

---

- [x] 2.4a Relocate the PyYAML emitter to `internal/yamlio`; fix P0–P6; consolidate codes (deps: 2.1, 2.8, est: ~4h)
  - PRIORITY: run this FIRST in Phase 2. Two of the defects below are in code
    that 2.5 and 2.7 are about to be written against — fixing them now is one
    edit session, fixing them after is three.
  - P0 (panic): `internal/scaffold/pyemit.go:590` slices `text[end+1:end]` because
    the escape branch at `:584-586` sets `start = end + 1` and the fold test at
    `:588` then passes. Python (`vendor/yaml/emitter.py:959-963`) tolerates
    `start > end` and yields `""`. 151 of 700 fuzz documents panicked; one repro
    leaves a half-applied workspace (platform written, registry not).
  - P1 (type corruption): `pyemit.go:112` uses `strconv.FormatFloat` where PyYAML
    uses `repr(float).lower()` (`vendor/yaml/representer.py:171`). Integral floats
    lose `.0`, which **changes the YAML type to `!!int` on reload**; the exponent
    threshold is 1e6 in Go against 1e16 in Python. 104 of 114 float samples
    mismatched. This is round-trip corruption of an authored file.
  - P2: `pyemit.go:265-274` never emits the explicit-key `? key` form.
    `check_simple_key` (`emitter.py:437-455`) falls back to `? key\n: value` at
    key length ≥ 123, empty key, or multiline key. 201 of 1200 structural fuzz
    documents diverged.
  - P3–P6 (R-0.7a(j)): four wrong-shape-YAML paths where Python raises and writes
    nothing but Go proceeds — `scaffold.go:294-300` rewrites a registry whose
    `ids:` holds a bare string; `scaffold.go:265-272` misses `IsFalsy` on a `[]`
    registry (R-1.7a, and `internal/yamlio` already has the helper); `roles/today.go:231,236-257`
    prints a line where Python raises, and narrows `len(v)` to sequences so a
    mapping-valued `platform:` silently reports 0 instead of 4; `roles/today.go:185`
    defaults `due` to `""` where Python interpolates `None`. Match the observable
    outcome — non-zero exit and **write nothing** — not the traceback.
  - Also: move every `Code*`/`Slug*` const into `internal/model/codes.go` (R-2.4)
    before 2.3 and 2.7 add two more clusters' worth; fix
    `cmd/company-os/commands.go:45`, where `notImplemented` returns
    `model.ExitValidation` so `validate` exits 1 mid-port and is
    indistinguishable from a real gate failure; correct the false "autojunk is
    inert" comment at `internal/ids/suggest.go:146-151` (`b` is the user-supplied
    id, and it diverges at ≥200 runes — live once 2.5 lands); fix
    `internal/skills/skills.go:138` (a file named exactly `.md` yields `""`) and
    `:73` (`Value.Equal` refuses cross-type equality where Python `==` gives
    `5 == 5.0`, so gate 7 would MISS a shadowing conflict Python reports).
  - RULING NEEDED, one line, do not debate at 3.3: `internal/scaffold/commands.go:458-462`
    `pathJoin` uses `filepath.Join`, which cleans `..`; pathlib does not.
    `scratchpad init --repo "a/.."` prints `initialized a/../scratchpad` under
    Python and `initialized scratchpad` under Go. Files identical, one line
    differs. Either match Python or declare it in 0.6's registry.
  - acceptance: R-0.7c (registry half), R-0.7a(j), R-1.7a, R-2.4
  - why: task 2.1 found that a PyYAML-compatible emitter is **mandatory**, reversing
    this plan's earlier rejection of one — `register_id` (`:1815`) re-dumps the
    whole `ids/registry.yaml` through `safe_dump`, so a single `add` on
    `examples/workspace` rewrites seven flow-style entries into block style, and
    the differential harness compares those bytes. An approximation fails:
    `team.yaml`'s 85-column `precedence:` line is one PyYAML does *not* fold,
    because there is no space past column 80. It landed in
    `internal/scaffold/pyemit.go` only because `internal/yamlio` was under
    concurrent edit; `internal/governance` (2.5) and `internal/federation` (2.7)
    both need it and must not import `internal/scaffold` to get it. Task 2.8 also
    left `register_id`'s unconditional rewrite unguarded to avoid colliding with
    2.1 — R-0.7c requires it.
  - acceptance: R-0.7c (registry half)
  - verify: `internal/governance` and `internal/federation` reach the emitter
    without importing `internal/scaffold`; `add` on an already-registered id
    leaves `ids/registry.yaml` byte-unchanged; differential harness clean for
    `add` and `init`.

- [x] 2.5 `internal/governance` — resolve, explain, deviations, exceptions (deps: 1.3, 2.3, 2.4a, est: ~5h)
  - DONE 2026-07-26. Five files (`resolve.go`, `explain.go`, `declare.go`,
    `sections.go`, `pysem.go`) plus `internal/render/governance.go` and
    `cmd/company-os/governance.go`.
    **The read-modify-write matched Python BYTE FOR BYTE and R-0.7a(g) was never
    needed.** The premise in this task's `why:` is now wrong in the direction
    that helps: the carve-out exists for the case where no PyYAML-compatible
    emitter is available, and 2.4a made one available in `internal/yamlio`. So
    `deviation declare` and `exception request` load with `PyLoadFile` and write
    with `PyWriteFile` — object level, not `yaml.Node` — and reproduce the
    oracle's own reflow exactly: `examples/workspace`'s `deviations.yaml` is
    committed in FLOW style with a folded multi-line rationale, an UNQUOTED
    YAML-1.1 timestamp beside a quoted one, and a `://` scalar inside a flow
    mapping, and all four come back out through `safe_dump`'s block layout with
    identical bytes on both sides. `file_tree` is clean on every one of the 20
    deviation/exception invocations. Consequence: **R-0.7a(b) is also not
    exercised** — `safe_load` discards comments and so does this path, because
    the `yaml.Node` round trip that would preserve them is exactly what would
    reintroduce the layout divergence. Byte parity beat comment preservation;
    that inverts what this task predicted and is the better outcome.
    **R-0.7c's third and last site is closed.** `writeGuarded` compares
    `PyDumpCanonical` of the committed document against a fresh derivation with
    **`generatedAt` held equal on both sides** — without that neutralization the
    guard is inert, because `NOW` changes every second and a fresh result never
    equals the committed one. Measured: the oracle itself leaves a one-line
    `generatedAt` diff behind on a clean `examples/workspace` today, which is
    precisely the R-0.10 breakage. With the guard, `governance resolve` and
    `graph build` on `examples/workspace` and `examples/federated` leave
    `git status` clean, and the harness still passes because it normalizes every
    `YYYY-MM-DDTHH:MM:SSZ` to `<TS>` — including the committed one. Both halves
    are mutation-tested: removing the guard and removing the neutralization each
    fail the suite. R-0.10 is now verified for both commands.
    **Three things deliberately NOT done**, each of which looked like an
    improvement: `deviation declare` still validates nothing and exits 0 on a
    mandatory rule (the rejection is *recorded* by `Resolve` as
    `deviationRejected` and surfaces later as a validate `[FAIL]`, per
    exit-code-map § "Code 5's third example"); `:367` keeps collapsing "unknown
    component" with "resolve was never run" as one exit 3 rather than splitting
    into 3 and 5 as the map suggests is cheap; and neither `governance resolve`
    nor `exception request` gained a `next:` line (R-1.9 over R-1.8).
    **One new refusal, forced.** `--team` is not marked required on the
    `governance` sub-parser, so `governance resolve` with no team reaches
    `ws.team_dir(None)` and raises `TypeError` — a traceback, exit 1, nothing
    written. Go would otherwise resolve `teams/` itself as the team directory
    and create `teams/generated/effective-governance.yaml`, a file the oracle
    never creates; R-0.7a(j) does not carve out the filesystem effect. Guarded
    in `cmdGovernance` as exit 2. The omitted `explain` positional is mapped to
    the four characters `"None"` at the same seam, because argparse's `None`
    reaches both the die() message and `suggest_ids`.
    `governance explain` is `internal/ids.Suggest`'s first live caller and needed
    no change to it.
    **Differential, measured against the same waiver registry** (`--only` on the
    three groups): PASS 0 → 19, DIVERGE 35 → 16 across 38 invocations; global
    PASS 98 → 118, DIVERGE 101 → 81, DECLARED 89 unchanged, STALE 0. Every one
    of the 16 residual divergences is one of three known classes and none is a
    `file_tree` or `stdout` difference in these commands: 11 are the exit
    1 → 3 not-found reclassification awaiting a declared-divergence entry, 1 is
    the `resolve` no-team traceback above, and 5 are invocations whose LAST step
    is `validate`, which is still a stub (task 3.x).
  - why: `deviation declare` and `exception request` are the two read-modify-write
    paths on hand-authored YAML, where `yaml.Node` fidelity is load-bearing. This
    is also where comment preservation inverts behavior — PyYAML destroys comments
    today, Go keeps them, which is sanctioned by R-0.7a(b) rather than silently
    shipped.
  - acceptance: R-1.1 (these commands), R-1.7, R-0.7a(b)
  - verify: differential harness clean modulo the sanctioned comment difference;
    `effective-governance.yaml` regenerates with `git status` clean.

- [x] 2.6 `internal/product` — discover, prd, check (deps: 2.3, 2.5, est: ~6h)
  - DONE 2026-07-26. Seven files (`contract.go`, `checklist.go`, `discover.go`,
    `prd.go`, `check.go`, `sections.go`, `pysem.go`) plus
    `internal/render/product.go` and `cmd/company-os/product.go`. Differential
    **PASS 118 → 146, DIVERGE 81 → 53, DECLARED 89, STALE 0**; the 47 product
    invocations are now 31 PASS / 1 DECLARED / 15 DIVERGE, and **every one of
    the 15 is exit-code-only** — no stdout and no file_tree difference anywhere
    in the cluster. Those 15 are the sanctioned 1→3/5/8/2 re-classifications and
    need entries in `examples/declared-divergences.txt`, whose header
    deliberately pre-declared nothing for `prd`/`discover`/`check`; the sole
    non-exit-code case is `discover/new-no-title`, where `slugify(None)` is an
    AttributeError traceback in the oracle and a diagnostic here (R-0.7a(e)).
    **One real defect the harness caught, in `prd complete`:** `shutil.move`
    does not fail when the destination directory exists — it moves the source
    INSIDE it, as `archive/prds/<id>/<id>/`, and the status rewrite and
    `outcome.md` then target the record already sitting at `archive/prds/<id>/`
    because both paths are built from `dst` rather than from where the move
    landed. `os.Rename` returns EEXIST there and writes nothing;
    `prd/full-lifecycle-force` is exactly that fixture. `internal/scaffold`'s
    `move` is now exported as `Move` and `shutilMove` adds the destination rule
    on top, so the two callers of one Python function share one implementation.
    **R-2.12 discharged:** `gather_prd_governance` returns `[]ChecklistItem` and
    `ChecklistMarkdown` is the only place a checklist line exists; `prd new`
    interpolates it into the artifact, `check` prints it stripped.
    **R-1.14 / OKF v0.2 Phase 0 landed here** — see task 0.1 of
    `docs/tasks/okf-v02-conformance.md`, still open there because this port fixed
    the Go side only. **No corpus fixture reaches it** (measured: all 6 reality
    docs and all 6 PRDs carry well-formed ISO dates), so the fix is covered by
    unit tests — `TestParseDate`, `TestDoneGateNamesAMalformedDate`,
    `TestDoneGateAcceptsAFreshRealityDoc`, `TestDoneGateStaleRealityRefuses` —
    and produced zero harness divergence. Gate 3 is exposed as
    `product.Gate(ws, ordinal)` + `product.Message`, matching
    governance/skills/federation, ready for 3.1; `CoreFieldErrors` is exported
    from here for gate 4 rather than copied. ST-034/ST-035 ported and marked
    off, plus ST-016/ST-017 (the `*_SECTIONS` ↔ template coupling), which needed
    the same `str.format` subset.
    **One change outside the cluster:** `run()` in `cmd/company-os/main.go` now
    renders a command's records BEFORE the error branch and skips the stderr
    diagnostic for a `model.QuietError`. `prd complete`'s refusal is the only
    command in the CLI that prints a whole block to stdout and still exits
    non-zero with an empty stderr; without both halves the port either loses the
    block or invents an `error:` line. No-op for every other command, which
    returns nil records with its error.
  - why: `prd complete` enforces invariant #4 of the whole methodology — a change
    is done only when reality is updated. It has no byte-level oracle today, which
    is precisely why 0.3's harness had to exist before this task.
  - acceptance: R-1.1 (these commands), R-2.12, R-1.14
  - verify: differential harness clean for all three `prd` actions, both `discover`
    actions, and both `check` kinds, on passing and refusing paths.

- [x] 2.7 `internal/federation` — manifest, sparse-checkout, slices, lock (deps: 1.4, 1.6, est: ~7h)
  - DONE 2026-07-26. Five files (`manifest.go`, `git.go`, `materialize.go`,
    `lock.go`, `sync.go`) plus `cmd/company-os/federation.go`. The manifest and
    the lock are carried as `yamlio` **PyValue objects, not Go structs** — three
    sites need Python-object behaviour a struct round trip loses: `repo_pin`
    iterates the pin mapping in insertion order and reprs the leftovers
    (`['branch']`), `status` interpolates a whole lock `pin:` dict through an
    f-string (`{'commit': 'abc'}`), and `--only` re-emits untouched lock entries
    verbatim. That forced one addition to `internal/yamlio`:
    **`PyRepr`/`PyString`/`PyStrings` in a new `pyrepr.go`** — the object-level
    twins of `pyobject.go`'s node-level `pyRepr`/`PyText`, pinned to them by
    `TestPyReprAgreesWithNodeRepr` on the same 13 documents (the precedent is
    `PyFalsy`, which already exists on both sides for the same reason).
    **All 38 inherited selftest assertions discharged** — 37 ported, ST-076 is
    `check(..., True)` and became a `t.Skip`; marked off in
    `.devlocal/go-port/selftest-inventory.md`.
    **The four measured traps, each mutation-tested:**
    (1) `_make_readonly`'s `sorted(rglob, reverse=True)` is ported as an explicit
    collect-and-reverse — but inverting it to a forward walk changes **nothing**:
    POSIX `chmod` needs ownership, not parent write, and `0555` keeps the search
    bit, so a pre-order walk produces byte-identical modes. What IS load-bearing
    is the *deep* chmod in `_force_remove` — dropping it fails ST-084 and leaves
    `t.TempDir` unable to clean up, because `unlink` needs write on the PARENT.
    The reverse sort is kept for fidelity, not for correctness; it is the
    `_force_remove`/`materialize_slice` restore passes that matter.
    (2) the lock's `files:` INSERTION order is `yamlio.OrderedMap` → `PyMap`,
    proven by syncing the same repo with `paths: [governance/, components/]` and
    then `[components/, governance/]` and asserting the lock's two blocks swap;
    emitting `SortedKeys()` there fails both that test and the union-hash test.
    (3) `aggregate_hash`'s plain string sort over the same keys is pinned against
    a hand-built sorted digest stream, so "unifying the two orderings" fails.
    (4) `sorted(rglob)`'s PurePath order goes through `yamlio.SortPaths`;
    `sort.Strings` flips `sdd/adr/a.md` past `sdd/adr-x.md`.
    Gate 8 is exposed as `federation.Gate(ws, manifest, ordinal)` returning a
    `model.GateResult`, over `SliceFindings` returning
    `Integrity{Findings, Files, Repos}` — `federated_slice_problems`' five
    English sentences decomposed into five codes with typed `Fields`
    (`repo`/`path`/`manifest`/`lock`), plus the SevOK line, and a pure
    `Message(code, Fields)` that is the package's only prose (the
    `internal/skills` gate-7 shape). Task 3.1 needs nothing else from here.
    Exit codes follow `.devlocal/go-port/exit-code-map.md` including both of its
    refinements: `:2318` (abbreviated SHA) is **4**, `:2547` (no `workspace.yaml`)
    is **3**. **Differential, `workspace`+`workspace-git` (77 invocations):
    PASS 1 -> 10, DIVERGE 75 -> 66.** Of the 66 remaining, **60 diverge on
    `exit_code` alone** and are blocked on task 4.3 adding their waivers to
    `examples/declared-divergences.txt` (which task 0.6 forbids this task to
    touch); 3 are `validate`-dependent (3.1); 2 are the malformed-`workspace.yaml`
    fixture, where Go **adds** an exit site the map already sanctions; and 1 is a
    **defect in the Python oracle** (below). Every other stream — stdout, stderr,
    and the file tree *including mode bits* — matches byte-for-byte.
  - FOUND, needs a waiver at 4.3: `workspace/status-failing-federated` **crashes
    the Python CLI**. `examples/failing-federated/workspace.lock.yaml` writes
    `resolvedCommit: 1111111111111111111111111111111111111111` unquoted, PyYAML
    resolves 40 digits to an **int**, and `:2651`'s `sha[:12]` raises
    `TypeError: 'int' object is not subscriptable` — exit 1, no stdout past the
    header. Go renders it through `str()` and completes. Do NOT "fix" the fixture:
    it is a golden input for gate 8. Declare the divergence.
  - why: ~510 lines and the most self-contained cluster (2 external callers), but
    it carries the fiddly filesystem work: `_make_readonly` (`:2354-2360`) uses
    `sorted(rglob, reverse=True)` so children are chmod'd before parents, and
    `filepath.WalkDir` is pre-order — Go must collect and reverse-sort explicitly
    or it will chmod a directory read-only before its contents.
  - acceptance: R-1.1 (`workspace sync|status`), R-0.11
  - verify: `--frozen` sync from lock with no network reproduces the committed
    slice tree with `0444`/`0555` modes; lock emission byte-stable across runs.

- [x] 2.8 `internal/ids` and `today` (deps: 1.6, est: ~2h)
  - DONE 2026-07-26. Landed as **two** packages, `internal/ids` and
    `internal/roles` — not `internal/ids` + `internal/today` as this entry and
    the task-0.4 inventory assumed. `role_glossary_lines` (`:1260`) has two
    callers, `cmd_today` (`:1171`) and `cmd_ids` (`:1277`), so a package named
    for the `today` command cannot hold it without `internal/ids` importing
    `internal/today`; naming the second package for the *role* concept gives a
    one-way `ids -> roles` edge. The LLD's "Package layout" section, which
    listed neither package, is amended with both and with the ruling.
    Differential: **42 of 47 PASS** (`ids` 24/26, `today` 18/21). The five
    remaining divergences are all out of slice and identical for every
    subcommand: `ids/not-a-root` and `today/not-a-root` (Python `die()` exits 1,
    Go exits 3 — the R-4.x contract that `.devlocal/go-port/exit-code-map.md:45`
    already assigns to `:230`, landing at task 4.3); `ids/bad-action` and
    `today/bad-role` (argparse's per-subparser usage block on stderr vs the Go
    `usage()` shipped by task 1.1); and `today/after-resolve`, whose step 1 is
    `governance resolve` (task 2.5, still a stub) so `today` correctly warns
    about the `effective-governance.yaml` that step never wrote. Selftest
    ST-026–ST-029 and ST-036–ST-037 ported and marked off in the inventory;
    `suggest_ids`' difflib dependency is transliterated and pinned against
    CPython on 200 random ratio pairs plus 15 `get_close_matches` queries
    (`internal/ids/testdata/difflib-oracle.json`). **Follow-up, deliberately not
    taken:** `register_id` (`:1813`) writes `ids/registry.yaml` unguarded and
    R-0.7c now requires a semantic-compare guard; its only callers are `init`
    and `add`, which task 2.1 was landing concurrently, so it is left for
    whoever lands it — it belongs in `internal/ids` on the same reasoning that
    puts `rebuild_generated` in `internal/graph`.
  - why: both flatten rich structures to prose today and are the two commands
    where `--json` has the most obvious immediate value.
  - acceptance: R-1.1 (`ids list`, `today`)
  - verify: differential harness clean across all six `--role` values and all four
    `ids list` filters.

---

## Phase 3 — Validate, renderers, and the parity gate (Units 0, 2)

- [x] 3.1 `internal/validate` — the 7/8 gates returning `GateResult` (deps: 2.2–2.7, est: ~8h)
  - why: `cmd_validate` reaches into six clusters and is the largest function in
    the file (186 lines). It sits above everything and nothing depends on it,
    which is what makes it the last thing built and the first thing measured.
  - DONE 2026-07-26, with 3.2. **`internal/validate` is 150 lines and composes
    almost nothing of its own.** Three of the eight gates already existed
    (`product.Gate`, `skills.Gate`, `federation.Gate`); four more were added to
    the cluster that owns their subject matter rather than here —
    `governance.OwnershipGate`/`ExpiryGate` (1, 2) and
    `graph.NodeGate`/`FeatureIndexGate` (5, 6) — because "which team is
    accountable" and "has this derived artifact drifted" are those packages'
    questions, and a gate living above them would have to re-derive the answer.
    Four helpers had to be ported to support them, all into `internal/graph`:
    `IdentityErrors` (`identity_errors`, `:1543`), `BlocksEqual`/`canonicalBlock`
    (`:109-117`), and `TagsInSync` (gate 4's `sorted(tags) == derived`).
  - **Gate 4 is the one gate built here, and it is not a placement preference —
    it is a cycle.** It needs `graph.IterGraphDocs` *and*
    `product.CoreFieldErrors`, and `internal/product` already imports
    `internal/graph`, so gate 4 in either package needs the import the other way.
    `internal/validate` sits above both. It composes no prose: each core-field
    finding's sentence comes from `product.Message`, each tags/pointer finding's
    from the new `graph.Message`.
  - **The banner is a section, not a gate** (`Slug: model.SlugWorkspace`,
    `Ordinal: 0`). The dispatch seam is `[]model.GateResult`, not `model.Report`,
    so the workspace root — the one line of the report no gate can derive — had
    to travel in-band. Keeping it out of the gate list is what keeps `[N/M]`'s
    denominator equal to the number of real gates.
  - **A malformed artifact aborts the run, and the banner still ships.**
    `:924` prints it before `:929` loads the manifest, so the oracle has already
    written that line by the time it dies — measured on a `workspace.yaml` with a
    non-list `repos:`. `Run` returns `[]GateResult{banner}` *with* the error, and
    the banner carries `complete: false` so the renderer does not print `PASS`
    over a run that reached no verdict. Deliberate residual: on a MID-gate abort
    the completed gates are dropped rather than rendered, because the denominator
    is derived from the gate list and a truncated list renders `[3/3]` where the
    oracle wrote `[3/7]`. One differential case reaches it
    (`exception/garbage-expires`, `expires: not-a-date`); it is a declared
    divergence either way, because Python's stderr there is a ValueError
    traceback.
  - Gate 2's date compare returns **exit 1** for an unparseable date, not 4 — the
    oracle's `dt.date.fromisoformat` raises ValueError and exits 1 through a
    traceback, and this is the one place in the cluster where the artifact-family
    default would have been wrong.
  - acceptance: R-0.4, R-0.5, R-2.1, R-2.7
  - verify: gate denominator computed at run time (7 vs 8 by fixture); gate 3
    renders its header with zero findings; gate 4 emits no `[ok]` for a document
    carrying core-field errors.

- [x] 3.2 `internal/render/text` — the per-gate prefix policy (deps: 1.7, 3.1, est: ~5h)
  - why: there is no uniform prefix rule and pretending there is one is how a
    records refactor breaks the golden silently. Seven distinct shapes: gate 1
    prefixes component, 2 team, 3 a compound `platform/prd-id`, 4 path, 6
    platform, 7 and 8 nothing — and gate 5 alone uses three shapes (`:1030`,
    `:1036`, `:1040`). The leading blank line is a property of the gate header,
    present on every gate except the first.
  - DONE 2026-07-26. **The premise above is wrong and task 1.7 already proved it:
    there IS a uniform rule.** `render.Validate` is 60 lines with no per-gate
    branch at all — emit `Subject`, then `": "`, then `Message`; `Message` alone
    when `Subject` is empty. The "seven distinct prefix shapes" are seven distinct
    `Subject` VALUES chosen by the producers, which is why `Subject` is documented
    as render-ready text that may carry punctuation: gate 1 emits `ghost`,
    `'svc-alpha'` and `svc-beta` from three sites inside one loop, and gate 5
    emits `<root>/team.yaml`, a bare `<root>` and `<root>/CLAUDE.md`. A renderer
    that branched per gate would need re-deriving every time a producer changed
    its mind about a prefix; this one cannot go out of date.
  - Three things stay derived rather than stored, exactly as 1.7's test-local
    renderer had them: the blank line is `Ordinal > 1` (it survives a gate with
    zero findings — `failing-federated-golden-validate.txt:3-4`), the `[N/M]`
    denominator is the gate count, and the trailer counts `SevFail` only
    (`failing-workspace` is 15 fails + 4 warns and reads 15). The one thing that
    is NOT derivable is whether the run finished; see 3.1's `complete` note.
  - `severityMarker` returns an error for an unknown severity rather than a
    fallback string. `[FAIL]` is what CI and four shipped skills grep for, so a
    producer/renderer mismatch has to be loud.
  - **ALL FIVE goldens reproduce byte-for-byte after `normalize()`**, asserted in
    Go (`internal/validate/golden_test.go`) rather than only by `acceptance.sh`,
    so the parity claim survives R-9.3 deleting the reference. Exit status is
    asserted separately from the diff, following `acceptance.sh:83-90`: a fixture
    that silently started passing would still diff clean against a re-baselined
    golden.
  - acceptance: R-2.5, R-2.6, R-2.8, R-0.1, R-0.2
  - verify: both committed goldens reproduce byte-for-byte after `normalize()`;
    both failure-path goldens from 0.2 reproduce.

- [x] 3.3 **Parity checkpoint** — differential harness against the Go binary (deps: 3.2, est: ~4h)
  - why: this is the moment the port becomes a measurable claim rather than an
    assertion. Everything downstream is gated on it.
  - acceptance: R-0.3, R-0.7, R-0.8, R-7.9 (first pass)
  - verify: harness from 0.3 reports zero divergence across all 16 commands, three
    fixtures, passing and failing paths; `acceptance.sh` passes unmodified.
  - DONE 2026-07-26. **`make differential`: 288 invocations, PASS 169,
    DECLARED 119, PARTIAL 0, DIVERGE 0, STALE 0, SKIP 0 — "ZERO UNDECLARED
    DIVERGENCE across the corpus".** `make check` green (gofmt clean, `go vet`
    clean, 974 tests across 16 packages, `examples/acceptance.sh` PASS
    unmodified — R-0.3). All five goldens reproduce byte-for-byte **from the Go
    binary** — `workspace`, `federated`, `failing-workspace`,
    `failing-federated`, `failing-federated-nolock` — R-0.1/R-0.2/R-0.9.
    `graph build` + `governance resolve` on `examples/workspace` leave
    `git status --porcelain -- examples/workspace` empty (R-0.10).
  - REGISTRY, AUDITED AS A WHOLE — 126 records over 119 invocations. Every one
    cites an authority (the harness refuses to load one that does not), none is
    stale (STALE=0 proves every record FIRED; a typo'd invocation id or a
    divergence that got fixed would both surface here), and **zero are
    `file_tree`** — across 288 invocations the two implementations' filesystem
    effects are byte-identical, which is the strongest single result in this
    checkpoint. By stream: 106 `exit_code`, 16 `stderr`, 4 `stdout`.
      - The 106 exit-code records are the Unit 4 contract and lose no teeth: each
        pins one exact `<ref> -> <cand>` transition, so any OTHER pair is still a
        hard DIVERGE. Distribution `1 -> 4` ×51 (manifest schema), `1 -> 3` ×33,
        `1 -> 8` ×6, `1 -> 6` ×6, `1 -> 2` ×4, `1 -> 5` ×3, `1 -> 7` ×2,
        `1 -> 0` ×1. Seven are pinned to a `step:` so a step that is supposed to
        SUCCEED cannot be absorbed by its own invocation's waiver.
      - 10 of the 16 stderr records use `waive: usage-block`, which drops only
        argparse's COLUMNS-wrapped block and still compares the diagnostic line
        byte-for-byte (R-1.4a).
      - Only 10 records are `whole-stream`, and 6 of those exist because the
        reference side is a Python TRACEBACK — no shared structure to compare.
  - THREE DEFECTS FOUND AND FIXED RATHER THAN DECLARED:
    1. **R-1.4a, six sites.** Three conditional requirements argparse cannot
       express (`discover new` with no title, `governance resolve` with no
       `--team`, `prd new --from-discovery` with no `--team`) emitted a bare
       `error: <sub> <action>: …` line — no `company-os ` prefix, no sub-parser
       usage line, invisible to the selector task 1.1a pinned. Three MORE had the
       same shape and were worse: `discover validate`, `prd validate`, and
       `prd complete` with their `nargs="?"` positional omitted reached command
       code with an empty id, `filepath.Join` silently dropped the empty segment,
       and the port reported `no active PRD at …/active/prd.md` at exit 3 —
       naming a path the user never asked about, and a code R-0.7a(l) does not
       sanction. All six now emit the argparse pair and exit 2, via
       `model.UsageError` (a `QuietError`-shaped wrapper, since the producers are
       below `cmd/` and cannot reach the parser's own type). Pinned by
       `TestConditionalRequirementsAreArgparseShaped`.
    2. **Sub-command `--help` answered from the wrong parser.** Every
       `company-os <sub> --help` printed the ROOT help, dropping that
       subcommand's own usage line, positionals, and flags. R-0.7a(i) waives
       argparse's LAYOUT, not answering a different question, so this was a
       defect. `help(scope)` now renders the named sub-parser with argparse's
       section order, metavars, `help=` strings verbatim, and its
       `min(longest + 4, 24)` gutter. Measured across all 16 sub-parsers:
       **8 are now byte-identical to the oracle** (validate, discover,
       governance, check, reality, scratchpad, graph, skills) and the other 8
       differ in the usage line's WRAP and nothing else. `usage/validate-help`
       went from DIVERGE to PASS and needs no registry entry at all — the reason
       PASS is 169 rather than 168.
    3. Task 0.6 flagged `--help` on stdout as unsanctioned; R-0.7a(i) has since
       been widened to cover stdout, so `usage/help` and `usage/prd-help` are now
       declarable, and are declared.
  - TWO THINGS THE AUDIT WOULD NOT SIGN OFF SILENTLY:
    1. **`exception/garbage-expires` stdout is the thinnest-justified waiver in
       the file.** The oracle crashes inside gate 2 having printed the banner,
       all of gate 1, and three of gate 2's four lines; the port prints the
       banner alone. That drops six human-facing lines R-0.8 forbids removing.
       It is not an oversight — `internal/validate.go:87-105` chooses it
       deliberately, because R-2.6 derives the `[N/M]` denominator from the
       length of the gate LIST, so returning the completed gates would render
       `[1/2]` and `[2/2]` where the oracle wrote `[1/7]` and `[2/7]`, a false
       claim about how much of the workspace was checked. It is a genuine
       **R-0.8-versus-R-2.6 collision** and it wants a ruling, not another
       waiver: either R-0.7a gains a clause naming a mid-gate abort, or the total
       gate count is carried on the banner record so partial output can render
       the true denominator. Left declared, with the collision written into the
       registry as the audit trail.
       **RULED AND CLOSED 2026-07-26 — R-2.6a, the second option.** The count is
       decided at `:930` from manifest presence, before gate 1 runs, so it was
       never a real conflict. `model.Report` gained `Total`, the banner record
       carries it as `fields.gates`, `render.Validate` reads it instead of
       counting the gate list, and `validate.Run` now returns the completed gates
       and the aborting gate's partial findings alongside the error (every gate
       producer already returned its partial `GateResult`; `internal/governance`'s
       `ExpiryGate` is the one the corpus reaches and now says so in its contract).
       The two stdouts are **byte-identical** — measured, same sha1 after `<WS>`
       normalization — so the stdout waiver was DELETED. Registry 126 -> 125
       records; only the stderr entry remains, Python's being a ValueError
       traceback. Differential unchanged at PASS 169 / DECLARED 119 / DIVERGE 0 /
       STALE 0 (the invocation is still DECLARED on its stderr alone).
    2. **`workspace/status-failing-federated` is the only `1 -> 0` in the
       registry** — the single waiver that turns a reference FAILURE into a
       candidate SUCCESS, which is the transition an agent branching on exit
       status is most exposed to. It is correctly sanctioned (the oracle crashes
       on `sha[:12]` where PyYAML resolved an unquoted 40-digit `resolvedCommit`
       to an int; the port's PyRepr layer renders Python's own `str()` and
       completes, and the fixture is a golden input for gate 8 that must NOT be
       "fixed"), but it deserves to be named rather than buried among 105
       same-shaped siblings.
  - ONE-LINE FOLLOW-UP FOR WHOEVER OWNS THE HARNESS: `differential.py`'s
    `AUTHORITY_RE` accepts `R-0.7a([a-k])` and stops one letter short of the new
    (l) clause, so the three entries (l) governs cite `R-0.7a(j)` (a truthful
    citation for the same SHAPE) plus `R-4.3`, and name (l) in prose. Widening
    the character class to `[a-l]` lets them cite the clause that actually
    governs them. Not changed here — the harness logic is out of this task's
    scope.

---

## Phase 4 — Agent surfaces: JSON, exit codes, version (Units 3, 4, 6)

- [x] 4.1 `internal/render/json` with `schemaVersion` and build id (deps: 3.2, est: ~4h)
  - why: agents are already first-class consumers driving the CLI from four
    shipped skills. An unversioned schema for a machine-facing contract is a
    breaking change with a fuse on it.
  - acceptance: R-3.1, R-3.2, R-3.3, R-3.4, R-3.5, R-3.9
  - verify: `--json` on every subcommand emits valid JSON with `schemaVersion` and
    build id; default output unchanged (3.3 still green).
  - DONE 2026-07-26. **ONE encoder, 150 lines, `internal/render/json.go`** —
    R-3.4b taken literally. It is a writer over `[]model.GateResult` and knows
    three things: the shape of a `Finding`, that a `Fields` entry named
    `model.FieldNext` is guidance, and that every payload carries
    `schemaVersion` (1) and `model.BuildInfo()` embedded as `build` (R-3.5 —
    task 4.4's struct, not re-derived). It knows no command name, no gate, no
    code. `cmd/company-os/render.go`'s per-command map is the TEXT side only and
    has no `--json` counterpart; adding a seventeenth command adds zero lines
    here.
  - The top-level array is `sections`, not `gates` (R-3.4a), frozen by
    `schemaVersion` on this publish.
  - **The precondition R-3.4b needed, and the only structural change in this
    task:** five commands — `init`, `add`, `reality new`, `scratchpad init`,
    `workspace sync|status` — wrote prose straight to `out` and returned no
    records, so a single encoder over the record types could not reach them.
    They now return findings whose `Message` is the finished line, and a
    seven-line `renderPlain` writes those lines back out. The format strings
    MOVED; they did not change — proven by the corpus, not by inspection.
  - Verified: `--json` on all 16 subcommands round-trips through
    `encoding/json`, carries `schemaVersion` + a fully populated `build`, and
    reports an `exitCode` equal to the code the process actually exits with
    (`cmd/company-os/json_test.go`). Counts serialize as numbers and ordered
    list values keep their order (`Fields` is passed through untouched).
  - **R-7.7 tuple equality holds, on all five fixtures, not two**
    (`internal/validate/json_test.go`). The assertion parses `render.Validate`'s
    rendered TEXT back into `{gate, code, severity, subject}` tuples and compares
    them to the tuples decoded from the `--json` document — comparing the two
    renderers' *inputs* would have passed no matter what either dropped.
    Mutation-checked: making the text renderer skip warns fails it.
  - R-3.3: all five goldens still reproduce byte-for-byte, `make check` green,
    `make differential` unchanged at PASS 169 / DECLARED 119 / DIVERGE 0 /
    STALE 0. `--json` is Go-only and confirmed absent from the corpus by a test
    that greps `examples/differential.py`, so the harness cannot silently start
    comparing a flag the oracle does not have.

- [x] 4.2 JSON envelopes for guidance and for finding-less commands (deps: 4.1, est: ~2h)
  - why: R-3.2 forbids prose on stdout and R-1.8 requires every mutating command
    print its next step — left unresolved, `--json` silently deletes the system's
    best existing affordance for exactly the consumer it was written for. And
    `prd new` produces no findings at all, so its envelope has to be defined
    rather than defaulted to an empty document.
  - acceptance: R-3.6, R-3.7
  - verify: `prd new --json` emits a populated envelope naming what it created and
    a `guidance` field carrying the next command.
  - DONE 2026-07-26. **R-3.6 is a Fields KEY, not a code table.** Producers that
    compose a next-step sentence also set `model.FieldNext` to the bare command;
    the encoder lifts every finding that has it into a top-level `guidance`
    array and needs to know nothing about which codes those are. That matters
    because the command is not always the whole sentence — `discover new` reads
    "fill …, then run: <cmd>", `deviation declare` reads "review due <date>;
    re-run: <cmd>", `workspace sync` puts a second command behind a `#` comment
    — so parsing it back out of the prose would have been guesswork.
  - `guidance` is `[]` and never `null`, so `.guidance | length` needs no
    special case. It is empty exactly where the oracle prints no next step:
    `governance resolve`, `exception request`, `scratchpad init`, `graph build`
    (R-1.9 outranks R-1.8 there, and the JSON does not invent what the text does
    not print — asserted).
  - R-3.7: the five finding-less commands each name what they created in
    `fields` as well as in the sentence — `init.created.root`, `add.created.id`
    (+ `kind`, `platform`), `reality.created.path`, `scratchpad.created.path`,
    `prd.created` — so the envelope is a description, never an empty document.
    `rebuild_generated`'s derived lines ride in their own `generated` section
    ahead of the command's own, preserving the oracle's ordering, and that
    section is omitted entirely rather than emitted empty.
  - R-3.8/R-3.9 wired at the same seam: a failure still writes one document on
    stdout carrying `error` and `exitCode`, the diagnostic still goes to stderr
    only, and the pre-dispatch failures (not a workspace root, no handler) go
    through the same path. `init`'s interactive prompt is redirected to stderr
    under `--json` — it is progress, not results.

- [x] 4.3 Wire the exit-code contract (deps: 0.5, 2.7, est: ~3h)
  - ORDERING CORRECTED: was `deps: 3.3`, which was circular. 3.3's gate is "zero
    divergence", but nine-plus of the known divergences ARE the exit codes this
    task wires — 3.3 could not go green until 4.3 landed, and 4.3 could not start
    until 3.3 was green. All 56 paths are already classified in
    `.devlocal/go-port/exit-code-map.md`, so nothing blocks this on the renderer.
    Deferring the *work* to Phase 4 was right; ordering the *gate* after it was
    not.
  - DONE 2026-07-26. The wiring was mostly already in place from Phases 1–2; the
    audit found **three real gaps**, all of the same shape — an error type that
    reached dispatch without a code and therefore resolved to **1**, the code
    reserved for "a `validate` subcommand reported `[FAIL]`":
    (a) `yamlio.SyntaxError`, which is EVERY malformed-YAML path outside
    `PyLoadFile` (skills frontmatter, graph tags, node parsing) — R-0.7a(e)
    requires 4 and it was returning 1;
    (b) `frontmatter.ErrInvalidUTF8`, same, R-4.5;
    (c) `internal/skills`' two wrong-shape refusals (`frontmatter is not a
    mapping`, `must be a scalar`) at 1, where `internal/roles` and
    `internal/scaffold` already return 4 for the identical condition under
    R-0.7a(j). The skills sites carried a comment *arguing* for 1; that reading
    predates (j) and is now corrected.
    Mechanism: `model.CodeOf` resolves through a new `model.ExitCoder` interface
    (`errors.As`, outermost wins) instead of matching `*model.Error` only, so a
    package with its own typed error implements one method rather than wrapping
    and losing the type. No string matching anywhere in `main`.
    Tests: `cmd/company-os/exitcode_test.go`, one per code (R-4.1..R-4.9,
    R-7.6), all driven through `run()` rather than through `CodeOf` on a
    hand-built error, because the thing that rots is the wiring. Codes 1 and 5
    have no producer in this build (`internal/validate`, `internal/product` are
    Phase 3) and are exercised through a temporarily registered command, which
    proves `run()`'s half and keeps proving it when the producers land.
  - FINDING, and it corrects the corpus rather than the map:
    `examples/differential.py` annotates `workspace/sync-bad-pin` as
    "short commit pin -> exit 4", i.e. as `:2318`. **It is not.**
    `git fetch origin <abbrev-sha>` fails outright ("couldn't find remote ref"),
    so both implementations die at `:2263` and `:2318` is never reached. The
    map's classification of `:2318` as 4 is right; **no fixture in the corpus
    reaches it**, so that ruling is untested by the harness. Left alone — the
    corpus is task 0.3's file and the note is a comment, not an assertion.
  - Differential, attributable to this task alone: **PASS 98 / DECLARED 26 /
    DIVERGE 164 -> PASS 98 / DECLARED 89 / DIVERGE 101, STALE 0.** 67 registry
    entries added over 63 invocations (60 exit-code-only + the two crash cases
    below). Of the 101 remaining, 87 are the seven still-unimplemented commands,
    11 are multi-step chains whose later step is one of those, and 3 are `--help`
    stdout (R-0.7a(i), task 1.1a). A later run with task 2.5's
    `internal/governance` also landed reads **PASS 118 / DECLARED 89 / DIVERGE
    81 / STALE 0** — DECLARED and STALE are unchanged, which is the check that
    matters here: none of the 67 new waivers went stale when a neighbouring
    cluster started passing.
  - acceptance: R-4.1 through R-4.10
  - verify: a test per code; every `die()` site from 0.5's map reaches its assigned
    code; `acceptance.sh:62`'s zero/non-zero assertions unchanged.

- [x] 4.3a Declare the `prd` / `discover` / `check` exit-code divergences (deps: 4.3, 2.6, est: ~45m)
  - DONE 2026-07-26. Task 0.6's registry header deliberately declared nothing for
    these three commands so that real defects could not hide behind a waiver the
    day they landed. 2.6 landed them; this closes the gap. Registry only — no Go
    source, no fixture, no golden touched.
  - **The question this task existed to answer: are the residual divergences
    exit-code-only? Measured yes.** Across all three clusters the harness
    reported **15 diverging invocations, 16 blocks, and exactly one of those
    blocks was not an exit code** — `discover/new-no-title`'s stderr, which is a
    Python `AttributeError` traceback. **Zero stdout blocks. Zero file_tree
    blocks.** The diagnostics and the filesystem effects were already
    byte-identical; only the status differed. `compare()` diffs every step and
    snapshots the whole tree after the last one, so this is a positive result,
    not an untested surface.
  - 16 entries over 15 invocations, by transition:
    **1 → 8** ×2 (`:417` `discover new` already-exists, `:610` `prd new`
    already-exists);
    **1 → 3** ×8 — `:244` team_dir ×3 (`discover new`, `discover validate`,
    `check ready`), `:430` no-brief, `:584` discovery-not-found, `:238`
    platform_dir, `:636` and `:676` no-active-PRD;
    **1 → 5** ×3 — `:587` unvalidated brief, `:703` done-gate ×2;
    **1 → 2** ×2 — `:601` (R-0.7a(f)) and `discover/new-no-title`;
    plus **1 stderr** `whole-stream` waiver.
    The reported shape count was **1 → 3 ×6**; the measurement found **×8**. The
    two the summary missed are `discover/validate-unknown-team` and
    `prd/validate-missing`, both genuine `team_dir`/`no active PRD` sites.
  - Four multi-step entries carry an explicit `step:` rather than the match-any
    default (`discover/new-twice-conflict` 2, `prd/new-twice-conflict` 2,
    `prd/new-draft-discovery` 2, `prd/full-lifecycle` 3), because step 1 of each
    is supposed to SUCCEED and an unstepped waiver would absorb it if it stopped.
  - **`discover/new-no-title` is the one entry whose authority is a judgment
    call, and it is flagged in the registry as such.** `d.add_argument("title",
    nargs="?")` (`bin/company-os:2688`) makes the title optional to argparse, so
    `cmd_discover` runs with `None`, `:412` calls `slugify(None)`, and `:73`
    raises `AttributeError` — an interpreter traceback, not a classified `die()`
    site, so no exit-code-map line applies. R-4.3 mandates 2 for a missing
    required argument and that is what the port emits. R-0.7a(j) is cited for the
    traceback-versus-diagnostic *shape* (same exception class, same
    unhandled-exception exit path) but its stated subject is a well-formed YAML
    document of the wrong shape, not a CLI argument. **R-0.7a should gain a
    clause naming this case literally.** Not `:601`/R-0.7a(f) — that is `cmd_prd`
    and a real `die()`.
  - **Observation for whoever owns 2.6's usage text, not fixed here:** the Go
    diagnostic is `error: discover new: the following arguments are required:
    title` — no `company-os ` prefix and **no sub-parser usage line**, where
    every other Go usage error emits `usage: company-os <sub> …` plus
    `company-os <sub>: error: …`. R-1.4a asks for "a diagnostic … plus a
    sub-parser-scoped usage line". The practical cost is registry-visible: with
    no `company-os…: error:` line, `waive: usage-block` strips nothing and would
    hard-DIVERGE, so this stream can only be waived **whole** — strictly weaker
    teeth than the ten `usage-block` entries above it.
  - Nothing declared for `validate` or for any chain ending in it — 2.7's owner
    is mid-implementation and a pre-declared waiver would hide the defect it is
    meant to catch. `prd validate` and `discover validate` are cluster commands,
    not that gate.
  - Differential, attributable to this task alone: **PASS 146 / DECLARED 89 /
    DIVERGE 53 / STALE 0 -> PASS 146 / DECLARED 104 / DIVERGE 38 / STALE 0.**
    PASS unmoved (no behavior changed), 15 invocations moved DIVERGE -> DECLARED,
    PARTIAL 0. The 38 remaining are `validate`/`governance`/`deviation`/
    `exception`, the chains ending in them, and the three `--help` stdout cases.
  - TEETH PROOF: `prd/complete-banking-active`'s `expect:` temporarily set to
    `1 -> 7`. Hard **DIVERGE**, annotated `declared 1 -> 7
    (.devlocal/go-port/exit-code-map.md:59, …:277), observed 1 -> 5`, and the run
    exited non-zero. Reverted; totals returned to 104/38/0. A waiver still waives
    one transition and not a stream.
  - acceptance: R-7.1a, R-4.3, R-4.4, R-4.6, R-4.9
  - verify: every added entry reports DECLARED, not DIVERGE; STALE stays 0; no
    `prd`/`discover`/`check` invocation diverges on stdout or file_tree.

- [x] 4.4 `--version` and ANSI-free guarantee (deps: 1.1, est: ~90m)
  - DONE 2026-07-26. Two halves, both landed.
    **`--version` (R-6.8).** The flag existed but printed one token
    (`company-os 206f90a`) and the Makefile's `-X main.commit=` was stamping a
    symbol `main` never declared — silently dropped by the linker for the whole
    of Phases 1–3. New form:
    `company-os <version> (commit <sha>, go1.25.7, darwin/arm64)`. `version`
    alone could not satisfy R-6.8 because `git describe --tags --always --dirty`
    degrades to a bare abbreviated sha on an untagged tree (indistinguishable
    from a release name) and, once a tag exists, no longer identifies the tree.
    Go version and platform are in because the only three artifacts R-6.2 ships
    differ by exactly those, and both are compile-time constants. **No build
    date** — it would make two builds of one source differ and cost the
    reproducibility R-6.10's checksums exist to be worth having; the commit says
    it precisely already.
    **The vars moved from `main` to `internal/model`** (`buildinfo.go`,
    `model.BuildInfo() model.Build`), and the Makefile now stamps
    `-X $(MODULE)/internal/model.{version,commit}`. R-3.5 puts the same
    identifier in every `--json` payload and the JSON encoder is an internal
    package — left in `main` it would have had to be handed down, and two call
    sites would be free to disagree about what "the build" is. **Task 4.1 calls
    `model.BuildInfo()` and embeds the returned `model.Build`**; its four json
    tags (`version`, `commit`, `goVersion`, `platform`) are the payload shape and
    are frozen by R-3.4 on first publish. The struct carries no wording (R-2.8):
    `cmd/company-os.versionLine()` composes the human line. `BuildInfo`
    normalises `""` to `dev`/`unknown`, which covers both a plain `go build` and
    the sharper `make build VERSION=` stamping an empty string over a good
    default.
    **ANSI-free (R-3.10)** is now a test, not a convention —
    `cmd/company-os/ansi_test.go`, mirroring `architecture_test.go`'s AST walk.
    Two assertions of deliberately different kind: `TestNoANSIEscapesInSource`
    walks `internal/` + `cmd/` with `go/ast`, `strconv.Unquote`s every
    STRING/CHAR literal and fails on byte `0x1b` in the *decoded* value — so
    `\x1b`, `\033`, `\u001b` and a raw ESC in a backquoted literal are one check
    rather than a list of spellings a new one can walk around — plus an import
    check against eight styling libraries (lipgloss, termenv, fatih/color,
    gookit/color, aurora, go-colorable, bubbletea, bubbles). `internal/tui` is
    exempted by path prefix; it does not exist yet, the walk simply never reaches
    it, and it starts being exempted the moment Phase 7 creates it with no edit
    to the test. `_test.go` files are excluded, following the R-2.10 precedent. A
    `scanned == 0` guard fails the test if the walk ever matches nothing, so a
    broken walk cannot read as a clean tree.
    `TestNoANSIEscapesAtRuntime` builds the binary and execs 24 invocations
    (global flags, all three failure shapes, every landed renderer, the scratch
    workspace and `examples/workspace`), asserting not one byte of stdout or
    stderr is `0x1b`. It execs rather than driving `run()` in-process because
    R-3.10 is a claim about what the shipped artifact writes to a pipe and
    `tty.go`'s isatty seam means the in-process path is not the same path.
  - MUTATION PROOF, four injections, each reverted:
    (1) `const boldOn = "\x1b[1m"` in `internal/render/ids.go` -> FAIL naming
    `../../internal/render/ids.go:16:16`;
    (2) the same as octal `"\033[2m"` -> FAIL at `:16:13`, proving the decode
    step and not a substring match is what is doing the work;
    (3) `import _ "github.com/charmbracelet/lipgloss"` in a throwaway
    `internal/mutproof` -> FAIL naming the import line;
    (4) `fmt.Fprint(w, string(rune(27))+"[1m")` in `render.IDs` — assembled at
    runtime, no literal holds the byte — **static check PASSES, runtime check
    FAILS** on both workspaces with the offending bytes quoted. That fourth one
    is the argument for keeping both: the structural check can only see escapes
    that are spelled out.
  - The runtime sweep found **zero** escape sequences in the current build. The
    tree was already clean; the test is what keeps it that way.
  - Not a differential surface: `--version` appears in no invocation in
    `examples/differential.py`, and `usage()` is byte-unchanged, so the three
    `--help` stdout divergences are exactly as 1.1a left them. Differential
    **PASS 118 / DECLARED 89 / DIVERGE 81 / STALE 0 before and after** — no
    movement in any bucket.
  - why: for a binary distributed by copy with no package manager, a user cannot
    tell what they are running and no bug report is actionable. The Python CLI
    emits zero ANSI codes; nothing currently states that must stay true, and the
    goldens depend on it.
  - acceptance: R-6.8, R-3.10
  - verify: `--version` reports version and build id; no subcommand emits an
    escape sequence.

- [x] 4.5 Document the exit-code contract (deps: 4.3, est: ~60m)
  - DONE 2026-07-26, alongside 4.3 — splitting them would have published a
    contract nobody could yet run. Primary home is
    `company-os-starter/docs/user-guide/reference/company-os-cli.md`, a new
    top-level `## Exit codes` section directly after the global-flag section
    (it is global, not per-subcommand): all eight codes with a "you get this
    when" column, the two discriminations that change what the reader does next
    (1-vs-5 and 3-vs-4), the R-4.10 guarantee that zero-vs-non-zero callers are
    unaffected, and a `case $?` example. Cross-linked from
    `how-to/run-the-validation-gate.md` where CI wiring is discussed — that page
    said only "exits non-zero" and is where an adopter writing a pipeline
    actually lands — and from the user-guide index.
  - why: no such contract has ever existed, and adopters' CI branches on exit
    status today.
  - acceptance: R-4.12
  - verify: contract published in user-facing docs with all eight codes.

---

## Phase 5 — Distribution (Unit 6)

- [x] 5.1 Release build matrix and checksums (deps: 3.3, est: ~3h)
  - DONE 2026-07-26. Three findings, each of which broke a claim the target was
    already making.
  - **Reproducibility was path-dependent.** A cold-cache rebuild at the same
    path was already byte-identical, so the naive "build twice" check passed —
    but the same source at a different path produced different bytes, because
    without `-trimpath` the compiler writes the checkout's absolute path into
    the pclntab and `-s -w` does not strip it. Isolating it needed
    `-buildvcs=false` on both sides: Go stamps `vcs.revision`/`vcs.time`/
    `vcs.modified` into every artifact, and a copied tree with no `.git` differs
    for that reason alone. With VCS held constant the two builds differed by
    path; with `-trimpath` added they are identical. Added `-trimpath` to
    `BUILDFLAGS` (both `build` and `release`), and proved it end to end: **two
    independent `git clone`s of the same commit at different-length paths now
    produce byte-identical artifacts and an identical `SHA256SUMS`.** New
    `make repro` target builds the matrix twice — second pass against a
    throwaway `GOCACHE` — and fails on any checksum movement.
  - Side effect worth knowing: `cmd/go` omits the `-ldflags` build setting from
    embedded metadata whenever `-trimpath` is set, so `go version -m` no longer
    shows the version stamp. The stamp is verified against the artifact's own
    bytes instead, plus an actual `--version` exec of the host-native artifact.
  - **`deps-check` proved nothing about a release artifact.** It inspected
    `$(BIN_OUT)` — the local host build — and never looked in `dist/` at all.
    Rewritten to depend on `release` and inspect each artifact: `go version -m`
    for `CGO_ENABLED=0` and `-trimpath` (arch-independent, the only check that
    works on all three from one host); `otool -L` for both darwin artifacts —
    which *does* read a foreign-arch Mach-O, so cross-compiled darwin is
    genuinely covered — and `file`'s `statically linked` for the ELF, because
    `otool` answers `is not an object file` on ELF and `ldd` does not exist on
    macOS. That last gap is real and is what 5.3's Linux box closes.
    Mutation-proved: `make deps-check LDFLAGS='-s -w'` and
    `make deps-check BUILDFLAGS=` each fail on all three artifacts.
  - **"Statically linked" is literally true only on Linux.** Every darwin
    artifact links `/usr/lib/libSystem.B.dylib` and `/usr/lib/libresolv.9.dylib`
    regardless of `CGO_ENABLED` — Apple ships no static `libSystem`, so a fully
    static Mach-O does not exist. `net` is not in the dependency graph; the Go
    darwin runtime emits both unconditionally. Both are part of macOS, so
    R-6.1's user-visible promise holds; `deps-check` fails on any *third* dylib
    and the wording no longer overclaims.
  - Version stamping verified in a **released** artifact, not a local build:
    `dist/company-os_<v>_darwin_arm64 --version` →
    `company-os 206f90a-dirty (commit 206f90a, go1.25.7, darwin/arm64)`. The
    4.4-class failure (a symbol the linker silently drops) cannot recur
    unnoticed — `deps-check` now fails on it.
  - `make check` exit 0; differential PASS 169 / DECLARED 119 / DIVERGE 0 /
    STALE 0, unchanged.
  - acceptance: R-6.1, R-6.2, R-6.10
  - verify: three artifacts build reproducibly; checksums published; `ldd`/`otool`
    confirm no dynamic dependencies.

- [x] 5.2 macOS Gatekeeper: satisfied by the install path, not by signing (deps: 5.1, est: ~4h)
  - **Notarization is BLOCKED on an Apple Developer account and was NOT
    performed. It is also no longer required** — see the closing note at the end
    of this entry, which supersedes the framing below. The investigation record
    is kept because it is why signing was not shipped as a half-measure.
  - R-6.3's fallback clause was in force at the time and had been
    satisfied in full.
  - Quarantine was measured, not assumed, and the assumption in the LLD was
    wrong in a way that matters. A fresh, never-executed binary carrying
    `com.apple.quarantine` does **not** get "killed on first exec" with a
    message — from a terminal it **hangs indefinitely and prints nothing**. The
    process never appears in `ps`; the exec blocks in the kernel awaiting a
    Gatekeeper verdict that arrives via a GUI dialog. No error, no exit code,
    nothing to grep for. `spctl -a -t exec` → `rejected`.
    `xattr -d com.apple.quarantine` recovers it completely (exit 0, correct
    output). Both directions reproduced on a freshly built, unique cdhash —
    the first attempt was confounded by Gatekeeper caching a positive verdict
    for a cdhash that had already run.
  - **Signing alone buys nothing.** A Developer ID Application certificate *was*
    available on the build host, so this was tested rather than reasoned about:
    signed with `--options runtime --timestamp`, verified `valid on disk` and
    `satisfies its Designated Requirement`, `TeamIdentifier` set, full Apple
    chain — and `spctl` still says `rejected`,
    `source=Unnotarized Developer ID`, and the quarantined binary still hangs.
    So `make sign` is not a partial mitigation and is not presented as one.
    (Also learned the hard way: `--sign "<name>"` fails with `ambiguous` the
    moment a renewed and an expiring cert share a name — the 40-char SHA-1 is
    the only reliable argument. Documented.)
  - `make sign` / `make notarize` added. Both refuse to run without
    `CODESIGN_IDENTITY` / `NOTARY_PROFILE` and print the exact commands to
    obtain them; neither can silently emit an unsigned artifact. Verified both
    guard paths exit 1.
  - Also established: **a bare Mach-O executable cannot be stapled** — there is
    nowhere in a flat file to store the ticket, and `xcrun stapler staple`
    fails. `notarytool` will not accept a bare executable either, hence the
    `ditto -c -k` zip. Consequence: Gatekeeper resolves the ticket online on
    first run; offline first-run requires shipping a `.dmg`/`.pkg` and stapling
    that. Documented so the maintainer who has the account does not discover it
    mid-release.
  - Fallback delivered: quarantine workaround documented at the point of
    failure in
    `company-os-starter/docs/user-guide/tutorials/01-first-day-with-company-os.md`
    (§ "macOS: the first run will be blocked"), linked from
    `company-os-starter/docs/GOLDEN-PATH.md`, maintainer procedure and CI secret
    list in the new
    `company-os-starter/docs/user-guide/how-to/release-and-upgrade.md`, and
    recorded as an accepted cost in a new `## Accepted Costs` section of
    `docs/hld/go-cli-tui-port.md` — including the honest statement that the
    port's headline justification inverts on macOS for downloaded artifacts.
  - **CLOSED 2026-07-26 without signing, by fixing the install path instead.**
    The blocker was mis-scoped: every measurement above is about a **browser**
    download. `com.apple.quarantine` is applied by the downloading application
    (Safari, Chrome, Mail), never by `curl`, `wget` or `tar`, and Gatekeeper
    only adjudicates files carrying it. `company-os-starter/install.sh` — the
    same approach the `local-search` CLI ships with — fetches by `curl`, so the
    installed binary has no quarantine attribute and runs unsigned with no
    prompt. Verified end to end on macOS 15 / arm64: the installed file carries
    only `com.apple.provenance` and `validate` exits 0, an attribute profile
    byte-identical to the already-in-use `local-search` binary. `spctl` still
    says `rejected`; that is the wrong question for an unquarantined file and is
    now documented as such rather than chased.
  - Release artifacts renamed `company-os-<os>-<arch>` (no version in the
    filename) because `/releases/latest/download/<name>` requires a fixed name —
    the version lives in the tag and in `--version`. `install.sh` is copied into
    `dist/` and checksummed with the binaries.
  - Docs realigned so the one-liner is the install instruction everywhere
    (GOLDEN-PATH, first-day tutorial, `make release` output), with the browser
    path kept as a documented fallback carrying its `xattr` fix. The HLD's
    accepted cost was rewritten rather than deleted: the old entry concluded the
    port's justification "inverts on macOS", which was true only of the path
    nobody needs to take.
  - `make sign` / `make notarize` are retained and still refuse to run without
    credentials, but are off the critical path.
  - why: a downloaded, unsigned binary carries `com.apple.quarantine` and is
    killed on first exec with "cannot be opened because the developer cannot be
    verified." The workaround is strictly harder than `pip install pyyaml` for the
    audience this change is justified on — the headline value inverts.
  - acceptance: R-6.3
  - verify: a **downloaded** darwin artifact runs on a clean Mac with no
    right-click-open and no `xattr` step. If notarization is not pursued, the
    workaround is documented and the HLD records it as an accepted cost.

- [ ] 5.3 Clean-machine install verification (deps: 5.2, est: ~2h)
  - **BLOCKED 2026-07-26 — on hardware, not on code.** R-6.4 is satisfied only
    by a *downloaded release artifact* on a *clean machine*, and says in terms
    that verification against a locally built binary does not satisfy it. This
    needs (a) a published release with a real download URL, (b) a clean macOS
    arm64 machine or fresh VM with no Python and no Go, and (c) a clean Linux
    amd64 machine likewise. None of the three exists here.
  - **The published install line does not work yet, and currently does something
    worse than failing. Measured 2026-07-26, not assumed:**

    | URL | Status |
    |---|---|
    | `raw.githubusercontent.com/…/main/company-os-starter/install.sh` | **200 — serves the PYTHON-era installer** |
    | `github.com/…/releases/latest/download/company-os-darwin-arm64` | 404 — no releases exist |

    All of this work is on branch `go-cli`, 14 commits ahead of `main`. `main`
    still carries the pre-port tree, so the one-liner documented in GOLDEN-PATH,
    the first-day tutorial and the procedure below currently installs the
    deleted Python CLI and its vendored PyYAML. That is a silent wrong answer,
    not an error.

    **Two things clear it, both outside this task:** merge `go-cli` to `main`,
    and publish one release. `.github/workflows/release.yml` was added so the
    second is `git tag v0.1.0 && git push origin v0.1.0` — it gates on
    `make check`, builds via `make release`, runs `make deps-check` before
    publishing, and marks the release `latest` so the unversioned artifact URL
    resolves. Until both happen, `install.sh` fails with an explicit "no release
    has been published yet" message pointing at `make install` from source,
    rather than a bare download error.
  - **No longer gated on 5.2.** The dependency was "the macOS half is only
    meaningful once notarization is done"; `install.sh` removed that — a
    `curl`-fetched binary is never quarantined, so the unsigned artifact is the
    shipping artifact. The procedure in
    `docs/user-guide/how-to/release-and-upgrade.md` was rewritten accordingly:
    step 2 is now the published one-liner plus an `xattr` assertion that the
    installed binary carries no quarantine attribute, and the browser download
    moved to step 3 as a secondary check, since some users will take that path
    whatever the docs say.
  - **Do not mark this done from a workstation run.** The whole point of the
    requirement is that a local pass and a real failure are indistinguishable.
  - The executable acceptance procedure a human must follow is written out step
    by step in `company-os-starter/docs/user-guide/how-to/release-and-upgrade.md`
    § "Clean-machine acceptance procedure": machine prep (assert no Python, no
    Go), install via the published one-liner, assert `xattr` shows no quarantine
    attribute, then a second pass through the browser-download path with
    checksum verification to confirm the documented `xattr -d` fallback is
    sufficient, then the 20-command surface and the pass condition. Do not `scp`
    the binary from a workstation — installing something other than what a user
    would get is how this check passes while the real path is broken.
  - **What was verified locally, which is real evidence but the wrong kind.**
    Extending 1.8's empty-directory work to the full surface: every invocation
    in the differential corpus with a real on-disk fixture — **227**, spanning
    all 16 subcommands, including the 13 that run in an empty directory — was
    executed twice over identical fixture copies, once normally and once under
    `env -i` with `PATH` containing nothing but the binary and `git`, `HOME`
    pointed at an empty directory, and no Python of any version reachable
    (asserted, not assumed). **Exit code, stdout and stderr identical in all
    227.** The binary was then relocated to an unrelated directory and re-run,
    re-confirming R-6.7.
  - **That local evidence can no longer be regenerated.** The probe lived in a
    scratchpad and imported `differential.py` to reuse its corpus; both that
    file and the Go corpus that replaced it are gone (task 6.10). The result
    above stands as a record of a measurement taken, not as something a future
    session can re-run. Re-establishing it would mean rebuilding an invocation
    list first — which is a reason to run the real clean-machine check rather
    than reach for the local substitute again.
  - The one gap this cannot close from macOS: the linux artifact's runtime
    linkage. `otool` cannot read ELF and `ldd` does not exist here, so
    `make deps-check` falls back to `file`'s structural `statically linked`.
    Running `ldd ./company-os` on the clean Linux box (expected:
    `not a dynamic executable`) is a required step of the procedure above.
  - acceptance: R-6.4, R-6.5
  - verify: downloaded artifact, clean macOS arm64 and clean Linux amd64, neither
    with Python; every subcommand runs following only published docs.

- [x] 5.4 Upgrade and version-skew position (deps: 4.4, est: ~90m)
  - DONE 2026-07-26. Position taken and documented in
    `company-os-starter/docs/user-guide/how-to/release-and-upgrade.md`
    § "Version skew across a shared workspace":
    **skew is supported within a workspace-format major version, and is
    invisible.**
  - The position rests on a structural fact rather than a promise:
    `model.BuildInfo()` has exactly two consumers — the `--version` line and the
    `build` object in `--json` — and is written into **no** workspace artifact.
    Compatibility is carried by the artifacts' own `schemaVersion` (`"1.0"` in
    descriptors and governance files, `version: 1` in `workspace.lock.yaml`),
    which moves when the format moves and not when the binary does.
  - Measured with two binaries stamped `0.9.0-old` and `9.9.9-new`, run over
    `examples/workspace`: (a) old-writes/new-reads and new-writes/old-reads both
    clean across `validate`, `today`, `governance explain`, `ids list`,
    `skills list`, `check ready`; (b) trees written by the two builds are
    **byte-identical** (same SHA-256 over the whole tree); (c) re-running
    `graph build` and `governance resolve` under the *other* build is a **no-op
    diff**, so a shared workspace does not churn and CI's regenerate-and-diff
    gate does not fail because a teammate is a version behind; (d) a full-text
    sweep of the resulting workspace finds **neither version string anywhere**.
  - The cost is stated rather than hidden: because nothing is recorded, nothing
    can warn. A future format change will not announce itself to an older
    binary. The mitigation is a release-process rule, written down — a
    workspace-format change is a `schemaVersion` bump plus a validator gate, and
    must never ship as a silent behavior change in a generator.
  - Upgrade-in-place covered. `make install` over an existing install works and
    reports the new version. Hardened it while here: the copy now lands on a
    sibling path and is `mv`'d over the target. rename(2) is atomic (no window
    where a half-written binary is on `PATH`) and unlinks rather than truncates
    the old inode — which is what an in-place `cp` cannot do on Linux while the
    old binary is executing (`ETXTBSY`). Verified Darwin permits the in-place
    write, so this is protection for the Linux release target, not for the
    build host.
  - Documented for users too: upgrading is a file copy, no migration, no state
    outside the workspace — and a *downloaded* replacement re-acquires
    `com.apple.quarantine` even though the binary it replaces ran fine.
  - acceptance: R-6.9
  - verify: a newer binary operates on a workspace last written by an older one;
    the skew position is documented.

---

## Phase 6 — Cutover, irreversible (Units 7, 8, 9)

- [x] 6.1 Port all 85 `selftest.py` assertions to named Go tests (deps: 0.4, 3.3, est: ~8h)
  - why: `selftest.py` is the only behavioral oracle covering `prd`, `discover`,
    `init`, `deviation`, and `exception`. Deleting it against an unquantified
    promise of package coverage is how a port loses the checks nobody remembered.
  - acceptance: R-7.3, R-7.4, R-7.5, R-7.6, R-7.7, R-7.8
  - verify: every line of 0.4's inventory maps to a passing named Go test; zero
    unported.

- [x] 6.2 Retire R-7.4's single-file and helper clauses (deps: 3.3, est: ~2h)
  - DONE 2026-07-26. Retirement is recorded, not performed silently:
    `docs/ears/federation-enrichment.md` gains **Amendment 1**, a dated record
    naming its authority (Unit 8, R-8.1–R-8.7), quoting the original locked
    statement verbatim, and splitting it into four separately-dispositioned
    clauses — C1 single-file **RETIRED**, C2 `die`/`ok`/`warn`/`fail` **RETIRED**,
    C3 `frontmatter()` parser **IN FORCE**, C4 next-command guidance chain **IN
    FORCE**. The R-7.4 row itself is rewritten to state only the two surviving
    clauses and links to the amendment, so a reader who never scrolls cannot
    conclude the whole requirement went away. **The clause split is the point**:
    the original bundled four independent clauses under one ID, which is exactly
    the shape that loses a load-bearing clause inside a retirement aimed at
    another. C3 and C4 are now individually citable and carry their downstream
    enforcement (R-1.5's differential regex test, R-1.8 plus R-3.6's `guidance`
    array) in the disposition column.
    **Eight documents updated, not the six R-8.5 enumerates** — two were not on
    the list: `docs/lld/golden-path-flavor-federation.md:11` ("All changes live
    in the single self-contained CLI") and `docs/lld/federation-enrichment.md:195`
    (a second Constraints restatement below the `:12` one already listed). Full
    set: `CLAUDE.md:102`, `docs/lld/golden-path-flavor-federation.md:11,49`,
    `docs/lld/federation-enrichment.md:12,195`,
    `docs/lld/okf-v02-conformance.md:204-206`,
    `docs/tasks/federation-enrichment.md:13,208`,
    `docs/tasks/golden-path-flavor-federation.md:13,276`. The two task files get
    an `AMENDED` bullet rather than a rewrite — their completed acceptance was
    measured against R-7.4 as locked at the time and is not retroactively wrong.
    R-8.6 and R-8.7 are recorded as a **scoping note beneath** the non-goal at
    `docs/hld/golden-path-flavor-federation.md:53`, not as a rewrite of it, so
    the original decision stays legible: "Web/GUI surfaces" rejects surfaces with
    a *runtime of their own*, which a terminal UI compiled into the same static
    binary is not; and the dependency policy governs **runtime** prerequisites,
    which build-time Go modules are not. R-8.7 is additionally recorded at its
    other live sites — `docs/GOLDEN-PATH.md` §0, `docs/lld/okf-v02-conformance.md`,
    `docs/lld/federation-enrichment.md`, and `docs/hld/federation-enrichment.md:89`
    (a fifth statement of the policy the task's four-site list omits).
    `bin/company-os:13` and `install.sh:5-8` are left alone — task 6.5 deletes
    both. Note the port satisfies that policy's *intent* more completely than the
    status quo: the Python CLI needed an interpreter plus a vendored PyYAML on
    the user's machine; the Go binary needs nothing.
  - why: a project whose product *is* the methodology cannot silently override its
    own locked EARS requirement. Two of the four clauses — frontmatter parser and
    guidance chain — are load-bearing and stay in force; retiring them by accident
    would be worse than not retiring anything.
  - acceptance: R-8.1, R-8.2, R-8.3, R-8.4, R-8.5, R-8.6, R-8.7
  - verify: `docs/ears/federation-enrichment.md:149` amended; all five repeating
    documents updated; `company-os validate` still exits 0.

- [x] 6.3 Tag the final Python commit and document rollback (deps: 6.1, est: ~30m)
  - **DONE 2026-07-26.** The blocker below was resolved by committing the port:
    commit `210de50` on `go-cli` contains both the Python CLI *and* the parity
    apparatus (`differential.py`, `declared-divergences.txt`, all five goldens),
    which is exactly what no prior commit had. Tagged `python-cli-final`
    (annotated, not pushed). **Rollback proven, not assumed:** extracting
    `bin/company-os` + `vendor/` + `templates/` from the tag into a clean temp
    directory and running `validate` against `examples/workspace` returns `PASS`.
    All three paths are required — `bin/company-os:36` resolves `TEMPLATES_DIR`
    at import, so a bin-only checkout imports fine and passes `validate` while
    every scaffolding command fails on a missing template.
  - Superseded analysis, kept because it explains why the tag waited:
    Rollback procedure written to `release-and-upgrade.md` §"Break glass:
    recovering the Python reference implementation": tag name, checkout command,
    capability boundary, and the exact command sequence a human runs at cutover.
    **Why no tag now.** The whole Go port is uncommitted working tree; `HEAD`
    (`206f90a`) predates it. `bin/company-os` *is* tracked and byte-identical to
    the working-tree parity oracle, so `HEAD` would recover a runnable CLI — but
    `examples/differential.py` is modified-uncommitted and
    `examples/declared-divergences.txt` is untracked, so **no commit in the
    repository contains the harness the oracle is useful with.** Tagging today
    yields a Python CLI with a pre-port harness, no divergence ledger, and no Go
    implementation to compare against — it recovers the oracle and discards the
    apparatus that makes it an oracle. R-9.2's constraint is *tag before delete*,
    an ordering, not "tag immediately".
    **Ordering that must be preserved at cutover** (written out in the doc):
    commit the port → tag that commit → push the tag → delete in a *separate*
    commit. Collapsing port and deletion into one commit destroys the recovery
    point entirely: no commit would then hold both a working oracle and the Go
    build it was proven against.
    **R-9.2's literal command is insufficient and the doc says so.**
    `git checkout <tag> -- company-os-starter/bin/company-os` alone does not
    restore a working CLI. `bin/company-os:36` resolves
    `TEMPLATES_DIR = SCRIPT_DIR.parent / "templates"` at import, so a bin-only
    recovery imports fine and passes `validate` while every scaffolding command
    fails on a missing template — the slowest possible way to discover the
    mistake. Recovery must take `bin/company-os`, `vendor/` (the pure-Python
    PyYAML, tracked, 19 files — so no `pip install` is required) and
    `templates/`. Verified by extracting exactly those three paths from `HEAD`
    via `git archive` into a throwaway dir and running `validate` against
    `examples/workspace`: PASS.
    Tag name `python-cli-final`, annotated, deliberately not `v`-prefixed so it
    never globs in with binary release tags. Repo currently has **zero** tags.
  - why: after deletion no runnable reference implementation exists from which to
    generate a golden for any defect found later. The tag is the only recovery
    path and it must exist before the delete, not after.
  - acceptance: R-9.2
  - verify: tag exists; `git checkout <tag> -- company-os-starter/bin/company-os`
    restores a working CLI.
  - **remaining**: a human runs the tag sequence in `release-and-upgrade.md`
    §"Creating the tag" at the cutover commit. Blocks 6.4/6.5.

- [x] 6.4 **Parity gate** — final differential run (deps: 6.1, 6.3, mutex: cutover, est: ~2h)
  - **GREEN 2026-07-26.** R-7.9 evaluated on all three counts:
    `examples/differential.py` exits 0 over 288 invocations (PASS 169,
    DECLARED 119, DIVERGE 0, STALE 0); `go test ./...` 1065 tests across 17
    packages, zero failures; `examples/acceptance.sh` PASS. All **five** goldens
    reproduce byte-for-byte **from the Go binary**, verified individually rather
    than through the harness — `golden-validate`, `federated-golden-validate`,
    `failing-workspace`, `failing-federated`, `failing-federated-nolock`. (The
    verify line below said "four"; there are five, since task 0.2 added three
    failure-path snapshots and this list predates them.)
  - NOTE on R-9.9: the Go codebase owner is still `TBD` in the HLD. R-9.9 requires
    it recorded "before implementation begins", which has long passed, and the
    "before 6.4 is declared green" wording added during task 6.9 was an editorial
    tightening rather than the requirement. It gates ongoing maintenance
    accountability, not this mechanical verification — but it remains genuinely
    open and someone must accept it.
  - why: R-9.1 permits deletion if and only if parity holds. This is where that
    condition is actually evaluated.
  - acceptance: R-7.9, R-9.1

- [x] 6.5 Delete the Python implementation (deps: 6.4, mutex: cutover, est: ~60m)
  - why: two implementations in parity is a maintenance tax with no expiry date.
    The port is worth doing only if it ends with one binary.
  - acceptance: R-9.3, R-9.5
  - verify: `bin/company-os`, `install.sh`, `vendor/`, `selftest.py` gone;
    `acceptance.sh` still green against the Go binary.

- [x] 6.6 Handle stranded launchers on existing installs (deps: 6.5, est: ~60m)
  - DONE 2026-07-26. **Executed ahead of its stated 6.5 dependency, deliberately
    — the dependency is backwards.** The only code lever for this task lives in
    `install.sh`, which 6.5 deletes. Doing 6.6 after 6.5 leaves nothing to edit.
    **The task's premise is empirically wrong, and the real hazard is worse.**
    Ran `install.sh` into a scratch prefix and inspected the result: it does not
    install a file, it installs a *tree* — `bin/ templates/ skills/ schemas/
    docs/ vendor/ README.md`, **79 files, 1.2 MB**, copied to
    `$COMPANY_OS_PREFIX/share/company-os/`, plus a 6-line bash launcher at
    `$COMPANY_OS_PREFIX/bin/company-os` that pins an absolute interpreter path,
    exports `PYTHONPATH=$HOME/vendor` and `exec`s the kit's own copy of the
    entrypoint. Because the kit is a **self-contained copy**, deleting the
    Python implementation from the repository does not strand it at all —
    verified: with the repo files conceptually gone, the installed launcher ran
    `validate` against `examples/workspace` to `PASS`, exit 0.
    So the failure mode is not ENOENT, it is **silent shadowing**. Verified with
    a fake Go binary later on `PATH`: the stale launcher wins, every real
    subcommand runs the old implementation and succeeds, and the one command
    that would expose the substitution — `--version` — hits a flag the Python
    CLI never had (`grep -c '\-\-version'` finds only `git --version` at
    `bin/company-os:2269`) and answers with an argparse usage banner, exit 2. A
    user can be a year behind with no symptom. Bare ENOENT (exit 2) occurs only
    in the half-migrated state where the kit root was deleted and the launcher
    left behind.
    **What was actually delivered.** (1) Migration procedure in
    `release-and-upgrade.md` §"Migrating off the Python kit" — ordered `rm`s
    (launcher before kit root, so the machine never sits in the ENOENT state),
    `type -a company-os` to expose shadowing, and a verification step keyed on
    `--version` returning a version line rather than a usage banner. Corrected
    the adjacent "there is no migration step" sentence, which was true only for
    binary-to-binary upgrades. (2) Three `troubleshooting.md` rows keyed on the
    symptoms a user actually sees — usage-banner-instead-of-version, `--json`
    rejected / behavior unchanged after installing, and the ENOENT. (3)
    `configuration.md`'s existing 6.7 block now warns that the old command
    *still working* is the problem, and cross-references rather than duplicates.
    (4) `install.sh` generates a self-diagnosing launcher: guards
    `[ -f $COMPANY_OS_HOME/bin/company-os ]` and on failure prints what the file
    is, the two exact `rm` commands with `$0` and the kit root resolved, and the
    doc pointer — exit 127. Verified before/after in scratch prefixes: bare
    `python3: can't open file ... [Errno 2]` exit 2 → actionable message exit
    127; healthy path and `--uninstall` unaffected (`validate` PASS, 0 residue).
    **Honest limit — the existing-install population is unreachable.** The
    launcher is a generated file on someone else's disk. Nothing shipped later
    can modify or remove it; there is no update channel, no phone-home, and the
    kit carries no version. The `install.sh` change therefore does **not** fix
    the installs this task names — it only benefits installs created between now
    and cutover, and installs made from `python-cli-final` during an R-9.2
    break-glass recovery (where the shadowing hazard genuinely recurs). For
    everyone else the mechanism is documentation plus one accident of layout:
    `make install` writes to `$PREFIX/bin/company-os`, the exact path the
    launcher occupies, so the default upgrade overwrites it. The orphaned 1.2 MB
    kit root survives that and always needs the manual `rm -rf`.
    **Two decisions left to the human at cutover, not taken here:**
    (a) R-9.3 mandates deleting `install.sh`; consider instead replacing it with
    a ~10-line migration notice, since `./install.sh --uninstall` is the
    documented uninstaller and deleting it removes the uninstaller along with
    the thing it uninstalls — after which `./install.sh` is itself a bare
    ENOENT. (b) The one lever that *would* reach existing installs is a check in
    the **Go binary** — on startup, if `$PREFIX/share/company-os/bin/company-os`
    exists, warn that a stale kit is present and may be shadowing. Not
    implemented here: Go source was out of scope for this task and concurrently
    owned. Worth a task if adopter installs are known to exist.
  - why: `install.sh` generated a bash launcher on every existing install that
    `exec`s a path which no longer exists. R-9.6 covers documentation, not
    installs.
  - acceptance: R-9.4
  - verify: documented migration step for an existing install; a stranded launcher
    fails with an actionable message rather than a bare `no such file`.

- [x] 6.7 Update all documentation referencing the Python path (deps: 6.5, est: ~3h)
  - DONE 2026-07-26. **13 documents**, against the 3 the acceptance line names.
    New install story stated once per audience and cross-referenced, never
    duplicated: adopter (`docs/user-guide/tutorials/01-first-day-with-company-os.md`
    §1, `docs/GOLDEN-PATH.md` §0, `company-os-starter/README.md`), contributor
    (`CLAUDE.md`, both `TUTORIAL.md` copies), CI
    (`docs/user-guide/how-to/run-the-validation-gate.md`,
    `docs/user-guide/how-to/sync-a-knowledge-catalog.md`,
    `docs/FEDERATION-RUNBOOK.md` ×2 patterns), reference
    (`docs/user-guide/reference/configuration.md`,
    `docs/user-guide/reference/company-os-cli.md`).
    **`TUTORIAL.md` exists twice** — repo root and `company-os-starter/docs/` —
    diverging only in relative paths and one extra intro paragraph. Both were
    updated; whoever consolidates them later should know they are near-duplicates,
    not one file with a symlink.
    **Two references the starting list did not name and that a grep for
    `pip install` alone would miss:**
    (1) `docs/user-guide/reference/configuration.md:66-80` documented
    `COMPANY_OS_PREFIX`, the kit root at `$PREFIX/share/company-os/`, and
    `./install.sh --uninstall` — an entire configuration surface that ceases to
    exist. Replaced with "nothing to configure" plus an explicit *migration*
    block naming both paths to delete, which is the doc half of task 6.6's
    stranded launcher.
    (2) `FEDERATION-RUNBOOK.md` Pattern B runs `python - <<'PY'` with `yaml` for
    the workflow's **own** manifest surgery — unrelated to the CLI, and it would
    have broken silently once the `pip install pyyaml` step above it was removed.
    Kept, with `pip install pyyaml` moved inside that step and a comment saying
    it is the workflow's dependency, not the CLI's, plus a `yq` alternative. It
    is the only `pip install` left anywhere in the docs and it is correct.
    CI snippets now install by pinned download (`COMPANY_OS_VERSION`, `curl` +
    `chmod +x`) — pinned deliberately, because two builds validating one
    workspace in the same week is a confusing failure. Release URLs are the
    placeholder `<release-url>` since task 5.1 owns the artifact host.
    **Task 5.2's macOS quarantine caveat has a reserved place, not a stub**:
    an `INSTALL-CAVEATS` HTML comment in `tutorials/01` marks the single home
    for it, and `GOLDEN-PATH.md` carries a matching comment pointing there
    rather than inviting a second copy.
    Not touched, deliberately: `examples/acceptance.sh` and `examples/selftest.py`
    (other tasks own them), `company-os-starter/vendor/README.md` and
    `install.sh` (deleted by 6.5), `.devlocal/**` (gitignored research), and the
    root `README.md`, whose "single-file" describes the graph-explorer tool and
    has nothing to do with the CLI.
  - why: five distinct invocation patterns are documented and in use, all assuming
    a runnable Python script.
  - acceptance: R-9.6
  - verify: no doc references `pip install pyyaml`, `PYTHONPATH`, the generated
    launcher, or `bin/company-os` as an executable; `README.md:29-35`,
    `docs/GOLDEN-PATH.md:24`, `docs/FEDERATION-RUNBOOK.md:448` updated.

- [x] 6.8 Update the four shipped agent skills for `--json` and exit codes (deps: 4.3, 6.5, est: ~2h)
  - DONE 2026-07-26. All four rewritten; **every step's `(mandatory)`/`(default)`/
    `(guidance)` tag is byte-identical to before**, verified by diffing the tag
    sequence against `git show HEAD:` (5/8/5/8 steps, same tiers in the same
    order). Versions bumped 1.1→1.2, 1.3→1.4, 1.0→1.1, 1.0→1.1.
    **Task 4.5 documented the exit codes but nothing documented `--json`** —
    the envelope had zero user-facing description, so skills citing it would
    have pointed at nothing. Added `## --json` to
    `docs/user-guide/reference/company-os-cli.md` (annotated envelope, the four
    guarantees consumers may rely on, the additive-fields rule, and the
    write-on-failure property), and every skill links to it rather than
    restating the shape.
    **Every code, field name, and exit code in the four skills was verified
    against the built binary, not read off the source.** Two claims were wrong
    on the first pass and are the reason this was worth doing empirically:
    `governance explain` publishes the tier as `fields.level`, not `fields.tier`;
    and `check done` emits `check.checklist`, while `checklist.item` is the
    *injection*-side code `prd new` writes into an artifact and is never a
    `check` finding. Branch tables were then confirmed live — `discover new`
    3/8, `prd new --from-discovery` on a draft brief → 5, `prd validate` missing
    id → 3, `exception request` without `--expires` → 2, and a real
    `prd complete` refusal returning exit 5 with `severity: "fail"` findings
    `done.checklist-unchecked` + `done.reality-missing`.
    Reshaping per skill: **running-discovery** — branch table for `discover
    validate` exit 1 keyed on `code`, plus the trap that
    `product.section-empty` with `fields.enforced: false` is a `warn` an agent
    must report and must not "fix" by rewriting someone's brief into the default
    shape. **creating-prd** — same for `prd validate`, plus the two findings a
    *successful* `prd new` emits that are easy to miss because they are not
    failures: `prd.governance-unresolved` (rules silently absent from the
    checklist) and `prd.reality-note`. **completing-a-change** — exit 5 is framed
    as the enforcement point of invariant #4 rather than an error to route
    around, with a four-code table, and an explicit prohibition on an agent
    passing `--force` on its own initiative; step 5 now requires confirming all
    three of `prd.archived`/`prd.log-appended`/`prd.outcome-scheduled` before
    reporting a change complete. **requesting-an-exception** — gains a
    tier-lookup preamble, because choosing exception-vs-deviation was previously
    left to the agent's judgement and a deviation aimed at a `mandatory` rule is
    rejected outright; also an explicit "never fill `approvedBy` on someone's
    behalf".
  - why: the agent-facing capability delivered in Units 3 and 4 is invisible to
    the agents already shipping in this repository, which still instruct plain
    command invocation and prose parsing.
  - acceptance: R-9.7
  - verify: all four `company-os-starter/skills/*/SKILL.md` use `--json` and branch
    on documented exit codes.

- [x] 6.10 Retire the Python parity scaffolding left behind by 6.5 (deps: 6.5, est: ~4h)
  - DONE 2026-07-26. 6.5 deleted the Python CLI but not the scaffolding built to
    compare against it, which left **18 test functions — 55 counting subtests —
    green because they could only SKIP**, against 1236 real passes. This is
    exactly the failure mode their own comments were written to prevent ("skip,
    never pass, so a missing oracle can never look like agreement"); it had
    already happened, permanently, with nothing reporting it. `make differential`
    was dead too: `PY_CLI` pointed at a path that no longer existed.
  - The full inventory, since a partial one is how this was missed the first time
    (an initial `grep` was truncated by a broken pipe and its partial output
    taken as complete — the count was reported as 4 before being re-measured):

    | Location | Tests | Subtest skips | Needed |
    |---|---|---|---|
    | `governance/oracle_test.go` | 4 | 22 | `bin/company-os` |
    | `skills/gate_oracle_test.go` | 1 | 17 | `bin/company-os` |
    | `skills/oracle_test.go` | 4 | 12 | `bin/company-os` |
    | `yamlio/pyemit_test.go`, `pyflow_test.go` | 4 | 4 | vendored PyYAML |
    | `graph/graph_test.go` | 2 | 2 | both |
    | `frontmatter`, `workspace` differential tests | 2 | 2 | `bin/company-os` |
    | `scaffold/pyoracle_test.go` | 1 | 1 | vendored PyYAML |
    | `governance/governance_test.go` | 1 | 1 | `bin/company-os` |

    Two of them (`graph_test.go`) already skipped with the message "the Python
    reference is gone; this oracle retires with it" — seen previously and left.
  - **The corpus was ported to Go, evaluated, and then REMOVED. That reversal is
    the substance of this task and is recorded rather than tidied away.**

    `examples/differential.py` (1361 lines) was first rebuilt as
    `company-os-starter/internal/difftest`: the same **288** invocations across
    all 16 subcommands, each against a fresh fixture copy, with per-step
    exit/stdout/stderr and the whole resulting file tree frozen into `testdata/`.
    It worked. Corpus parity was proved rather than asserted — the id list was
    captured from `differential.py --list` *before* that file was deleted, and a
    test failed if any of the 288 went missing. (The real count is **288**, not
    the 227 quoted in task 5.3; that figure was the subset with a real on-disk
    fixture.) The goldens were proved able to fail, twice.

    It was then removed, on a measurement. Three mutations, each checked against
    the suite WITHOUT the corpus:

    | Mutation | Rest of suite | corpus | acceptance.sh |
    |---|---|---|---|
    | `[warn]` → `[WARN]` in `render/governance.go` | **missed** | caught (3) | — |
    | `##` heading in `scaffold/template.go` | caught | caught (8) | — |
    | renamed a key in the generated governance file | caught | caught (**135**) | **missed** |

    One unique catch in three, and its class is specific: **rendered bytes**.
    Unit tests exercise the model, not the text that reaches a terminal —
    `internal/render` has tests and they passed while the label was wrong.

    Against that, the third row is the cost. A one-line change reddened 135 of
    288 snapshots. Nobody reviews 135 golden diffs honestly; they run `-update`
    and skim. A suite with that blast radius trains a reviewer to rubber-stamp,
    which is the exact failure it exists to prevent. 2.1M of `testdata` was the
    smaller objection.

    The deciding frame: the ten oracle tests deleted below had been **skipping
    since cutover**, asserting nothing. So removing the corpus does not regress
    from a working state — it returns the repository to where it was, minus 55
    misleading green skips. The corpus was new coverage that had never existed in
    working form, and it was judged not worth its maintenance shape.
  - **What is consequently NOT covered, stated plainly rather than left to be
    discovered.** No end-to-end byte-level coverage exists for `discover`,
    `deviation`, `exception`, `check`, `governance`, `today`, `graph`, `ids`,
    `skills` or `scratchpad`. `acceptance.sh` freezes `validate` alone (5
    fixtures, passing and failing paths). A change to what any other command
    PRINTS will not be caught by any test. What each command WRITES is still
    covered at the library seam by its own package's tests.

    The proportionate fix, if this ever bites: add a handful of golden
    assertions to `acceptance.sh` for `today` / `skills list` / `ids list` —
    roughly 1% of the corpus's size for most of its unique value. Deliberately
    not done pre-emptively.
  - The **entire declared-divergence subsystem was never ported** — the waiver
    registry, its citation grammar, and all 631 records in
    `examples/declared-divergences.txt`. Every one sanctioned a Python/Go
    difference; against frozen output there is no second implementation, so any
    difference is a regression and waiving it would defeat the point.
  - **One finding from the port worth keeping even though the code is gone.**
    `workspace sync` records the source `file://` URL in `workspace.lock.yaml`,
    and a synthetic git repo lives in a per-run temp directory. The Python
    harness never saw this because it ran both binaries inside ONE temp dir in
    ONE process, so the path cancelled out of the comparison. Anything that
    freezes `workspace sync` output across runs must normalize that path or it
    will fail at random.
  - **Each skip-only test was decided on its own merits, not swept.** Tag
    `python-cli-final` — the recovery path 6.3 exists to provide — was checked
    out to recover both the Python CLI and the vendored PyYAML 6.0.2, so
    "freeze" was a real option rather than a wish:

    - **Frozen (unique coverage, could not be recovered any other way).**
      `skills/gate_oracle_test.go`'s 17 synthesized workspaces — skill shadowing,
      dangling `extends`, and the id-type collisions where Python's `==` spans
      the numeric tower (`5 == 5.0`, `True == 1`) — exist in NO committed
      fixture, so no fixture-driven suite could reach them however broad.
      Reference gate-7 blocks captured into `testdata/gate7_reference.json`.
      `internal/yamlio`'s four emitter tests frozen into
      `testdata/pyyaml_safedump.json` (61 cases, 4 dump modes). These are the
      reason the `python-cli-final` checkout was worth doing, and they survive
      the corpus's removal.
    - **Converted to self-contained property tests (no oracle needed).**
      `governance_test.go`'s duplicate-relationship tally now asserts the
      property directly — the platform appears twice in the list while the
      requirement counts are unchanged — which was the whole of what the
      differential checked there. `scaffold/pyoracle_test.go` became a
      fixed-point check under the Go emitter; not a weaker claim, because
      `TestEmitterMatchesPyYAML` pins that emitter to `safe_dump` byte for byte.
      `internal/yamlio/testdata/pathorder_oracle.py`, the last thing in the suite
      needing `python3` on `PATH`, became `testdata/pathorder_cpython.json`
      (208 groups, CPython 3.12.11; the group set is seeded, so reproducible).
    - **Deleted, and NOT replaced.** `governance/oracle_test.go`'s and
      `skills/oracle_test.go`'s eight tests and `graph_test.go`'s two. At the
      time they were removed the corpus covered the same ground; the corpus was
      then removed too, so nothing covers it now. That is the gap enumerated
      above, and it is the honest outcome: these ten had asserted nothing since
      cutover, so no working coverage was lost — but none was gained either.
      Their shared helpers were kept; only the oracle-specific ones went.

    Every frozen file was mutation-checked: corrupting two entries in
    `pyyaml_safedump.json` produces four failures, and one in
    `gate7_reference.json` produces six. They assert.
  - **What was NOT ported, and why.**
    `examples/banking/.../test_settlement_finality.py` stays Python: it is the
    *governed* artifact, not the governor. Company OS is language-agnostic and
    the example exists to show a polyglot federation; it also sits inside a
    federation slice whose bytes are hashed into `workspace.lock.yaml`, so
    rewriting it would force a re-sync and a gate `[8/8]` re-baseline for nothing.
    It is the only `.py` left tracked in the repository.
  - **Regeneration procedure**, since nothing frozen here can be re-derived from
    the working tree. Check out the tag into a scratch directory:

    ```
    git archive python-cli-final company-os-starter/bin \
        company-os-starter/vendor company-os-starter/templates | tar -x -C <dir>
    ```

    All three paths are required — `bin/company-os:36` resolves `TEMPLATES_DIR`
    at import, so a bin-only checkout passes `validate` while every scaffolding
    command fails. `gate7_reference.json` is produced by walking `gateCases`,
    running `python3 bin/company-os --root <ws> validate` on each synthesized
    workspace and slicing out the `[7/` block. `pyyaml_safedump.json` needs the
    four `safe_dump` calls recorded beside the tests that consume them.
  - **Honest limitations, stated rather than glossed.** A differential proves two
    implementations agree; a golden only proves behaviour has not changed since
    someone last read the diff — so a golden accepted without reading it protects
    nothing. The frozen answers cannot notice if they were wrong when captured,
    only if behaviour drifts away from them. And `scaffold/pyoracle_test.go` now
    rests on the same emitter it used to be independent of.
  - why: a test that can only skip is worse than no test, because a skip reads as
    a pass in a green summary — and `make check` was reporting green over 55 of
    them. The end state keeps the coverage that nothing else could provide (the
    17 skill-conflict workspaces, the PyYAML emitter answers) and accepts the
    loss of the rest, having measured what that loss actually is rather than
    assuming it either way.
  - acceptance: R-9.3 (completion). **R-7.1 is NOT met and is not claimed** —
    the differential corpus was preserved through the port and then deliberately
    retired; the requirement predates the deletion of the reference
    implementation it was written for.
  - verify: no tracked `.py` outside the banking fixture; **zero** Python-related
    SKIP in `go test ./... -v` (was 55); `make check` green; the frozen answers
    in `gate7_reference.json` and `pyyaml_safedump.json` go red when corrupted.

- [x] 6.9 Record the named Go codebase owner (deps: none, est: ~15m)
  - **CLOSED 2026-07-26. Owner: javierbenavides**, recorded in
    `docs/hld/go-cli-tui-port.md` Stakeholders. Scope accepted: review of
    `internal/**` and `cmd/**`, the `go.mod` toolchain pin, and being the person
    a defect lands on.
  - The field was first added as a deliberate `TBD` placeholder and left
    blocking, because naming an owner requires a person to accept the
    responsibility and cannot be inferred from a repository. No name was
    invented in the interim; the placeholder held for four commits.
  - **Recorded out of order, which is worth keeping in the record rather than
    smoothing over.** R-9.9 requires the owner "before implementation begins."
    This was filled after the port shipped, after `151f159` deleted the Python
    reference, and after 6.10 retired the differential corpus. The guarantee the
    requirement was written to provide — that somebody had accepted the
    maintenance burden *before* the only oracle was destroyed — was not
    delivered. The acceptance is genuine; the sequencing was not honoured.
  - why: there is no Go precedent in this repository, and `local-search` being Go
    is precedent for the ecosystem rather than evidence of this team's capacity to
    maintain 4500–6000 lines of it.
  - acceptance: R-9.9 (met late — see above)
  - verify: owner named in the HLD Stakeholders section.

- [x] 6.11 Re-measure the claimed `init` → `validate` divergence (deps: 6.5, est: ~1h)
  - **CLOSED 2026-07-26. The claim was false. There is no divergence.**
  - During task 7.7 a comment was written into `cmd/company-os/tuiadvise.go` and
    `tuiadvise_test.go` asserting that `company-os init && company-os validate`
    exits 1 with `teams/<t>/CLAUDE.md: generated block drifted` against the real
    binary — "measured twice, with and without `--root`" — while the same `init`
    driven in-process through `run()` came out in sync, and that the reason was
    not understood. It was carried as a live loose end.
  - **Re-measured across three builds and both invocation forms. All six
    combinations exit 0:** a binary built from the working tree, the committed
    `company-os-starter/company-os` (`de085be`), and the installed
    `~/.local/bin/company-os` (`117ff23`), each run from inside the new workspace
    and again from its parent with `--root`. Gate `[5/7]` reports every context
    node in sync in every case. **GPF-R-1.7 holds.**
  - **What the original observation most likely was.** A fresh `init` *does*
    leave one thing undone — it writes no `teams/<t>/generated/`, so
    `effective-governance.yaml` is absent and `today --role` warns about it.
    That is a real fresh-workspace finding and it is the one the advisor's second
    diagnosis fires on. It is not a drifted block, and `validate` does not fail
    on it. The two were conflated.
  - **What was corrected rather than left standing:** the file comment in
    `tuiadvise.go`, the comment on `TestAdviseOffersGraphBuildForADriftedBlock`,
    and the `freshWorkspace` helper's account of why it drives `run()`. A false
    claim in a comment outlives the session that wrote it and is worse than no
    comment, because the next reader has no reason to doubt it.
  - `cmd/company-os/zz_probe_test.go` is the untracked probe written to chase
    this. It is a no-op without `PROBE_ROOT` and is left in place pending a
    decision to keep or delete it — it is not part of the suite's meaning.
  - why: an unexplained divergence between the binary and its own test harness
    would undermine every in-process test in the package. It turned out to be a
    misreading, which is a better outcome than the one that was recorded — but
    only if the record is corrected rather than inherited.
  - acceptance: GPF-R-1.7 (verified, not merely assumed)
  - verify: `init` then `validate` exits 0 from the cwd and via `--root`, on a
    working-tree build and on the last two released ones.

---

## Phase 7 — TUI, gated on Phase 6 (Unit 5)

Does not start until 6.4 is green. Read-only screens ship complete before any
mutating form is written.

- [x] 7.1 Bubble Tea shell, `tui` subcommand, TTY gate (deps: 6.4, est: ~4h)
  - DONE 2026-07-26. bubbletea v0.25.0 + bubbles v0.18.0 + lipgloss v0.9.1 — the
    first Go runtime deps, sanctioned by R-8.7. `internal/tui` is a Company-OS-
    ignorant Bubble Tea shell over a `[]Screen` of closures; `cmd/company-os/tui.go`
    holds the gate and the catalog. **The one non-obvious finding is a parity
    trap**: `commandNames()` feeds `argument cmd: invalid choice: … (choose from
    …)`, which `usage/unknown-subcommand` compares byte-for-byte (its
    `waive: usage-block` strips the usage lines and deliberately keeps the
    diagnostic). Listing `tui` there would have moved DIVERGE off 0 for a
    cosmetic gain, so `cmdSpec.goOnly` keeps it out of that string only — it
    still parses, dispatches, and shows in `--help`. The gate probes BOTH stdin
    and stdout through tty.go's termios ioctl and runs BEFORE `RequireRoot`
    (`tui` is exempted in main), so R-5.3 is unconditional: no TTY is exit 7
    whether or not you are standing in a workspace. `q`/`Esc`/`Ctrl-C` quit
    ahead of every other key in `Update`, from all three modes; going back is
    `backspace`/`left`, because a key that sometimes exits is the trap R-5.14
    exists to prevent. **ANSI exemption verified four ways**, not assumed:
    baseline passes with internal/tui importing all three styling libs; a
    lipgloss import in `internal/render` FAILS; a `"\x1b[31m"` literal there
    FAILS; and retargeting `tuiPrefix` makes internal/tui's own imports FAIL,
    proving the walk reaches those files rather than missing them. The 24-
    invocation runtime sweep still finds zero escape bytes with lipgloss linked
    into the binary.
  - why: an agent or CI job must never be able to land in an interactive app, and
    a TUI that cannot be exited is worse than no TUI for an audience defined as
    blocked on terminal fluency.
  - acceptance: R-5.1, R-5.2, R-5.3, R-5.14, R-5.16
  - verify: `tui` with no TTY exits 7 and changes nothing; `q`, `Esc`, `Ctrl-C`
    exit from every screen; bare `company-os` still prints help.

- [x] 7.2 Read-only screens, enumerated and tested (deps: 7.1, est: ~12h)
  - DONE 2026-07-26. Ten catalog entries in ONE table, not ten views. Six are a
    subcommand: they build the same `*Args` the parser would have built and go
    through `commands[…]` then `renderers[…]`, in run()'s own order, so R-5.13
    holds by construction — `TestScreensRenderThroughTheRealRenderers` asserts
    the body is byte-identical to `run()`'s stdout for validate/skills/ids/today.
    The other four (overview + component/PRD/discovery browsers) have no
    single-command twin and list what is on disk. **`discover validate` is
    deliberately NOT the discovery browser's detail view**: it rewrites
    `status: draft` → `status: validated` in the brief
    (`internal/product/discover.go:150`), so a browser built on it would edit
    what you browse — exactly the defect read-only-first exists to prevent.
    `prd validate` was checked and IS read-only, but the browser stayed a listing
    for symmetry. `TestEveryScreenRunsAndWritesNothing` executes every screen at
    every choice of every picker and hashes the whole tree around the sweep. The
    header's `$ company-os …` line is DERIVED from the executed `*Args` via
    `screenCommand` — the read-only ancestor of 7.4's preview, ~15 lines over the
    existing spec table; a `Screen.Command` field existed briefly and was deleted
    because a second, statically written copy is precisely the drift R-5.7
    forbids.
  - why: this subset serves what POs actually do — they read status far more often
    than they scaffold — and it needs none of the mutation machinery, which
    eliminates the entire class of "the TUI wrote the wrong thing" defect.
  - acceptance: R-5.4, R-5.13
  - verify: all ten named screens exist and are asserted by test; validate results
    render from the same records the text and JSON renderers consume.

- [x] 7.3 Terminal robustness — `NO_COLOR`, narrow, resize (deps: 7.2, est: ~3h)
  - DONE 2026-07-26. `NO_COLOR` follows the published convention (present AND
    non-empty — `NO_COLOR=` must not mean the opposite of what was typed) and
    drops every attribute, not only colour, which turns "honours NO_COLOR" into
    a byte assertion: no frame in any mode contains 0x1b. Nothing is lost,
    because the selection was always carried by a `> ` marker and the styling was
    decoration. Bodies are hard-wrapped in `relayout` — the viewport does not
    wrap, so a long finding would be folded by the terminal outside the
    viewport's line accounting and scrolling would stop meaning anything —
    carrying the leading indent onto continuations, since a continuation flush
    left reads as a new record. Below 80 columns the footer shortens; headers
    truncate with an ellipsis so a long title cannot push the layout down a row.
    Verified twice: headlessly at 40/60/100 columns and across a 200→50 resize
    with a body already open, and over a real pty at `PTYCOLS=60 NO_COLOR=1` —
    zero lines over 60 columns, zero SGR sequences emitted.
  - why: for an audience defined by terminal unfamiliarity, a UI that garbles
    below 80 columns fails on the machines it exists to serve.
  - acceptance: R-5.15
  - verify: legible at 60 columns; honours `NO_COLOR`; survives resize.

- [x] 7.4 Derived command preview (deps: 7.2, est: ~4h)
  - DONE 2026-07-26. `screenCommand(*Args)` is now the ONLY producer of a
    command line in the binary, and preview/execution are made
    structurally incapable of diverging by a **single-value** design plus a
    **law**, not by a matching test. The value: `invocation{ws, args}`
    (`cmd/company-os/tuiform.go`) satisfies the new `tui.Action` interface with
    `Preview() = screenCommand(i.args)` and `Commit() = runScreen(i.ws, i.args)`
    — two readers of ONE field, so they can no more disagree than two callers of
    the same getter. `tui.Form` exposes exactly one hook, `Build(values)
    (Action, error)`; a separate Preview closure beside a separate Commit
    closure was rejected precisely because it would be two independently written
    functions behind one screen. The law:
    `parse(shellSplit(screenCommand(a))) == a` for every `*Args` the spec table
    can describe — preview is a right inverse of the real parser, so the only
    way to satisfy it is to render exactly what the parser reads back. The
    corpus is GENERATED from `commandSpecs` (19 hostile free-text values × 4
    global combinations × every command), so a flag added tomorrow is covered
    tomorrow. Three things exist only because the law had to be TOTAL: POSIX
    single-quoting (a title with a space is not otherwise a command), rendering
    the pre-subcommand `--root`/`--json` globals, and moving a leading-dash
    positional behind an argparse `--` guard, which forces flags in front of it
    for that case only. Two empty-value rules fell out of the same totality
    requirement: an empty REQUIRED positional prints as `''` and a required flag
    always prints, because eliding either yields a line the parser rejects.
    Mutation-proved, not assumed: removing the quoting, dropping one flag from
    the render, and hand-writing a preview literal inside `internal/tui` each
    fail. The last is caught by a go/ast guard in `internal/tui/form_test.go`
    (same shape as `ansi_test.go`) that rejects any string literal in that
    package containing `company-os` followed by anything — the bare program name
    is the menu title and stays legal. `--root` is threaded from `args.Root`,
    not `ws.Root`: interpolating the resolved absolute path would print a flag
    the reader never typed.
  - why: this is the property that justifies interactive mutation at all — every
    TUI action reduces to a flag-complete invocation reproducible in CI. Derived
    from the args structure, never hand-written per screen, because a hand-written
    preview drifts from what actually runs and quietly destroys the guarantee.
  - acceptance: R-5.6, R-5.7, R-5.10
  - verify: preview string is generated from the same struct the command executes;
    a test asserts preview and execution cannot diverge.

- [x] 7.5 First mutating forms — `discover new`, `prd new` (deps: 7.4, est: ~8h)
  - DONE 2026-07-26. Two screens, twelve in the catalog. **new discovery brief
    (writes)**: `team` (picker over `ws.AllTeams()`) → `--team`, `title` (free
    text) → the `title` positional. **new PRD (writes)**: `platform` (picker) →
    `--platform`, `title` → `--title`, `components` → `--components`, `team`
    (optional picker) → `--team`, `from-discovery` (optional picker, offering
    ONLY `status: validated` briefs, since `internal/product` refuses anything
    else with exit 5) → `--from-discovery`. `--force` is not collected: it means
    nothing to `new`. R-5.10 is checked, not asserted — every field is filled
    with a value unique to it and the previewed line is parsed back, so a field
    that does not survive the trip is a value the TUI can collect and the CLI
    cannot reproduce. `title` is marked optional in the PRD form so the real
    "--title required (or --from-discovery)" rule stays in `internal/product`
    rather than being re-implemented in the UI (R-5.12). **R-5.8/R-5.9 proved on
    a real tree**: `TestCancelledFormLeavesTheWorkspaceExactlyAsItWas` fills the
    form completely, reaches the preview, and leaves by each of q / esc / ctrl-c
    / n, hashing the workspace around every step — and separately after EVERY
    keystroke of every form. R-5.9's second half is honoured by NOT promising
    it: once `Commit` begins there is no rollback, because only `init` stages
    into a temp dir, and the code comment says so. **R-5.11/R-5.12**:
    `TestConfirmingRunsTheCommandThroughTheSameCodePath` runs the form in one
    workspace and the previewed argv through `run()` in another, then compares
    the whole hashed trees — identical. Nothing shells out; `Build` is pure and
    only constructs an `*Args`, normalized through the same `normalizeArgs`
    extracted out of `parse` so a form-built and a parser-built `*Args` cannot
    mean different things. **One documented carve-out to R-5.14**: `q` is a
    character while a free-text field has focus, because a form whose title
    cannot contain the letter q is not a form. `esc` and `ctrl+c` remain
    unconditional from every mode including a text field, the footer names them
    there specifically, and nothing is written at that point in any case — so no
    exit from a form can leave a partial write. `discover validate` is still
    wired nowhere: it rewrites the brief. If it is ever offered it belongs in
    this file, behind a preview, not in a browser. One real Bubble Tea trap
    found: `Update` takes `Model` by value, so the answer slice is shared across
    every model derived from the same parent — `setValue` copies on write, and
    `TestFormValuesDoNotLeakAcrossModels` branches two edits off one model to
    prove it.
  - why: the two commands a product owner actually authors. No forms are built for
    `workspace sync` or `scratchpad init` — nobody will scaffold infrastructure
    through a form instead of typing the command.
  - acceptance: R-5.5, R-5.8, R-5.9, R-5.11, R-5.12
  - verify: no write occurs before confirmation; cancellation before confirmation
    leaves the workspace untouched; scaffolding delegates to the same code path
    `init`'s `_prompt` wizard uses, so GPF-R-1.3/1.4 hold for both.

- [ ] 7.6 Product-owner journey verification (deps: 7.3, 7.5, est: ~2h) — **BLOCKED**
  - BLOCKED 2026-07-26. This is the only Phase 7 acceptance that no engineering
    session can satisfy, and it must not be marked done by proxy. It needs a
    HUMAN PRODUCT OWNER, and specifically one who is not a member of this
    project. What unblocks it, exactly:
    1. **A person** who has never used `company-os`, whose job is product rather
       than engineering, and who is not told how the TUI works beforehand.
    2. **A machine they own** — a fresh macOS box with no prior shell
       configuration, no Go toolchain, no `PATH` edits, no repo checkout.
    3. **An install line only** — `curl -fsSL …/install.sh | bash`, and nothing
       else: no repo, no toolchain, no instructions beyond the published docs.
       **This item was previously blocked on notarization and no longer is.**
       It read "a signed, notarized darwin artifact, which today does not
       exist… R-6.3 is a hard prerequisite for THIS task." That was true of a
       browser download, which hangs under `com.apple.quarantine`. Task 5.2
       established that `curl` never sets that attribute, so the unsigned
       artifact installs and runs unassisted. The prerequisite is gone; what
       remains is a person and a machine.
    4. **An observer** who records where they hesitate rather than helping —
       the finding is the hesitation.
    The journey to observe, unassisted: read workspace status, open a
    component's governance, and inspect a validate failure. Nothing about 7.1–7.5
    substitutes for it. If it is never run, the honest outcome is stated in the
    task's own `why:` line — strike the product-owner justification and
    re-justify the TUI on engineer value alone, which is defensible but has to be
    said rather than assumed.
  - why: not one Phase 1 success criterion involves a member of the audience the
    TUI is justified on. If nobody will run this test, the PO justification is
    struck and the TUI is re-justified on engineer value alone — which is
    defensible, but it has to be said rather than assumed.
  - acceptance: HLD Success Criterion 9
  - verify: a product owner, given only the published install line and no prior use, reads
    workspace status, opens a component's governance, and inspects a validate
    failure — unassisted, without prior shell configuration.

- [x] 7.7 Recovery menu, advisor, and `add team --repair` (deps: 7.5, est: ~6h)
  - DONE 2026-07-26. Three behaviours, one theme: the TUI stops being a viewer
    for workspaces that are already correct and starts being usable from the
    states people are actually in. New EARS: R-5.17 – R-5.22 and
    GPF-R-1.9a – GPF-R-1.9c.
  - **Recovery menu (R-5.17).** `tui` outside a workspace root used to return
    `RequireRoot`'s error and exit 3. Measured standing in
    `examples/banking/bank/workspaces/` — a directory holding *two* workspace
    roots — it refused and printed the root-resolution order, which is the wrong
    answer to "I am one directory too high" when both right answers are in the
    directory being read. It now opens a menu: roots found at or one level below,
    a previewed `init` here, and the old error's content shown rather than fatal.
    Depth is one level on purpose (`TestNearbyRootsStopsAtOneLevel`) — a
    recursive scan of `/` or `$HOME` is slow where it matters and surprising
    everywhere. **This exempts `tui` from R-4.4**, recorded as
    [Amendment 1](../ears/go-cli-tui-port.md#amendment-1--r-44-exemption-for-tui-2026-07-26);
    the visible surface is that quitting the menu exits 0, not 3. R-5.3 is
    untouched: the TTY gate still runs first, so no TTY is still exit 7 with no
    filesystem change, whether or not a root is present.
  - **Advisor (R-5.18 – R-5.22).** Heads the catalog, empty on a healthy
    workspace. Detects drifted generated blocks (via `graph.NodeGate`, the same
    check `validate` runs, so advisor and gate cannot disagree), a missing
    `effective-governance.yaml`, and missing scaffolded team files (via
    `scaffold.MissingTeamFiles`, which reads the same definition `RepairTeam`
    writes from — a detector answering a different question from the fixer is how
    an offer starts lying). Each offer is a field-less `Form` opening straight on
    the previewed command, so R-5.6/5.7/5.8 hold through the machinery 7.4
    already proved rather than through new promises. **A missing
    `workspace.yaml` is explained and never offered** (R-5.22): a manifest needs
    repo URLs and commit pins nothing can infer, and a form that writes a
    plausible-but-wrong one is worse than the missing file.
  - **`add team --repair` (GPF-R-1.9a–c).** GPF-R-1.9 refuses once the unit
    exists, which left a team missing one standards file with no path back short
    of hand-copying from another team. `--repair` writes only what is absent,
    from the same `teamFiles()` definition `scaffoldTeam` uses, and reports
    written and skipped per file — the skipped list *is* the evidence nothing was
    overwritten. A no-op repair does not rebuild generated blocks, which is what
    makes it safe to run speculatively.
  - **Four defects the tests caught, worth recording because three were in the
    new code and one was in the test:**
    1. `AllTeams()` returns directory paths, not ids — the offers were building
       `--team /abs/path/teams/core`, which parses and addresses nothing.
       `baseNames` is what the rest of the catalog already does.
    2. The round-trip test needed `tokens[1:]`: `screenCommand` includes the
       program name.
    3. Backing out of a field-less confirm returned to an empty form and panicked
       on an index out of range. Fixed in both `confirmKey` and `formKey`.
    4. `TestAFormWritesNothingUntilConfirmed` pressed enter on an offer and
       called it an R-5.8 violation — but on a confirmation, enter *is* the
       confirmation. The key set was narrowed for field-less forms rather than
       the property weakened.
  - **One claim from this session was found false and is retracted**, not
    inherited: that a fresh `init` leaves a drifted block behind. See task 6.11.
    `TestAdviceHeadsTheCatalog` now pins the true fresh-workspace state —
    `validate` exits 0, and the advisor is non-empty because governance is
    unresolved.
  - why: the two states a new user is most likely to be in — standing one
    directory too high, and holding a workspace whose derived files have not been
    generated — were the two the TUI handled worst. One exited 3 with a
    resolution order; the other required knowing which of eleven subcommands to
    run next.
  - acceptance: R-5.17, R-5.18, R-5.19, R-5.20, R-5.21, R-5.22, GPF-R-1.9a,
    GPF-R-1.9b, GPF-R-1.9c
  - verify: `tui` from a directory holding two roots offers both and exits 0 on
    quit; `tui` with no TTY still exits 7; the advisor is empty on
    `examples/workspace` and heads the catalog when it is not; every offered fix
    round-trips through the real parser into the same `*Args`; detection and
    preview leave the workspace tree hash unchanged; `--repair` restores a
    deleted file byte-identically and skips every file that exists.
