---
type: tasks
slug: derived-drift-repair
status: draft
created: 2026-08-23
---

# Derived Drift Repair (`validate --fix`) — Tasks

Source of truth: `docs/ears/derived-drift-repair.md` (Units 0–4, locked).
Architecture constraints: `docs/lld/derived-drift-repair.md`.
Target: `company-os-starter/` — the Go module at `cmd/company-os` + `internal/`.

**Global acceptance (must hold after every phase):** `make check` passes, and
`company-os validate` still exits 0 on `examples/workspace` with byte-identical
output when `--fix` is absent (R-1.2).

**Hard sequencing:** Phase 0 lands alone, in two commits, before any user-visible
work. Two clauses in Units 2–4 are unimplementable against the shipped code
without it (LLD Sequencing, D5/P1, D9).

**Naming note:** `boolFlag("repair", …)` already exists on `add team`
(`args.go:128`). `--fix` on `validate` is a distinct verb surface; do not
consolidate them in this change.

---

## Phase 0a — `model.Error` carries a cause (Unit 0, prerequisite)

- [ ] 0.1 Add `Err` field and `Unwrap` to `model.Error` (est: ~25m)
  - why: `model.Errorf` formats with `fmt.Sprintf`, not `%w`, and `model.Error`
    has no `Unwrap` (`model.go:231`). `os.IsPermission` on any writer error
    returns false, always. Unit 3 cannot classify a permission failure until
    this exists — it is the whole reason Unit 3 is blocked.
  - acceptance: R-0.1 — THE SYSTEM SHALL allow an error constructed by
    `model.Errorf` to carry an underlying error, and SHALL expose it via `Unwrap`.
  - verify: unit test asserts `errors.Is` traverses a wrapped sentinel; assert
    `model.CodeOf` still resolves the code on a wrapped error.
  - landed:

- [ ] 0.2 Graph writers pass the `*os.PathError` through (deps: 0.1, est: ~30m)
  - why: the wrap is inert unless the three writers that can hit a `0444` slice
    actually supply the cause. These are the only sites Unit 3 reads.
  - acceptance: R-0.2 — WHEN a write performed by a graph writer fails, THE
    SYSTEM SHALL propagate the underlying filesystem error such that
    `errors.Is(err, fs.ErrPermission)` reports truthfully.
  - verify: test writes into a `t.TempDir()` chmod'd `0555` via each of
    `writeIfChanged`, `WriteFeatureIndexes`, `RewriteFrontmatterTags`; assert
    `errors.Is(err, fs.ErrPermission)` on all three.
  - landed:

- [ ] 0.3 Characterization test: rendered messages unchanged (deps: 0.1, 0.2, est: ~20m)
  - why: `model.Error`'s rendered text is what users and goldens see. The wrap
    must be invisible above the error boundary. Without this pinned, 0.1/0.2 are
    an unbounded change.
  - acceptance: R-0.3 — THE SYSTEM SHALL leave the rendered message text of every
    existing error unchanged by R-0.1 and R-0.2.
  - verify: table test over the existing `Errorf` call shapes asserting exact
    `.Error()` strings; `make check` golden tests pass unmodified.
  - landed:

---

## Phase 0b — `Rebuild` reports what it retagged (Unit 0, prerequisite)

- [ ] 0.4 `Rebuild` returns the tag section; `Build` re-expressed on top of it (est: ~40m)
  - why: `Rebuild` discards the `wrote` boolean (`graph.go:73`), so it cannot say
    which documents it retagged; only `Build` builds that section
    (`graph.go:26-43`). R-4.2 forbids re-deriving the list by a second walk,
    because a second derivation can disagree with what was actually written.
  - acceptance: R-0.4 — THE SYSTEM SHALL report, from the rebuild entry point,
    one section naming each document whose derived tags it rewrote.
    R-0.6 — THE SYSTEM SHALL NOT derive the set of rewritten files by any
    traversal separate from the traversal that performed the writes.
  - verify: `Build` is a call to `Rebuild` plus its summary record — assert by
    reading the diff, not by test; unit test asserts the section names exactly
    the drifted document on a two-doc fixture where one is clean.
  - landed:

- [ ] 0.5 Pin `graph build` output as byte-identical (deps: 0.4, est: ~15m)
  - why: `Build`'s output is pinned by the graph goldens; 0.4 restructures its
    internals. This is the checkpoint that the restructure was behaviour-neutral.
  - acceptance: R-0.5 — THE SYSTEM SHALL produce, from `graph build`, output
    byte-identical to that produced before R-0.4.
  - verify: existing graph golden tests pass unmodified; `graph build; graph build`
    on `examples/workspace` is still a no-op diff.
  - landed:

---

## Phase 1 — Flag surface and partitioning (Unit 1)

- [ ] 1.1 Declare the derived-condition code set as a constant (est: ~20m)
  - why: membership decided by string pattern at runtime is how
    `federation.slice-set-drift` (`codes.go:81`) gets silently repaired by
    punctuation luck. The set is a design decision and must be written down once.
  - acceptance: R-1.8 — THE SYSTEM SHALL define the derived-condition code set as
    a single declared constant, and SHALL NOT determine membership by string
    pattern at runtime.
  - verify: the set is exactly `{CodeTagsDrift, CodeNodeDrift,
    CodeFeatureIndexDrift, CodeNodeAbsent, CodeFeatureIndexAbsent}`, declared in
    one place, referenced everywhere.
  - landed:

- [ ] 1.2 Structural guard test against future drift codes (deps: 1.1, est: ~25m)
  - why: the failure mode is a future gate adding a derived condition that
    silently defaults into "never repaired," with nobody making that call. The
    guard converts that silence into a red test.
  - acceptance: R-1.9 — THE SYSTEM SHALL fail its own test suite if any finding
    code declared in `model/codes.go` contains `drift`, case-insensitively, and
    appears in neither the derived-condition code set nor a declared
    decision-class allowlist.
  - verify: add a fake `governance.resolve-drift` code locally in the test and
    confirm the guard fires; match on `drift` anywhere, case-insensitively — not
    a suffix test (D11).
  - landed:

- [ ] 1.3 `--fix` flag on `validate`, inert by default (est: ~30m)
  - why: CI must see byte-identical behaviour on a pipeline that never passes the
    flag. Default-inertness is the promise that makes this safe to ship.
  - acceptance: R-1.1, R-1.2, R-1.3 — accept a boolean `--fix` on `validate`;
    WHERE absent, produce output, workspace mutations, and an exit code identical
    to those produced before this change; SHALL NOT accept it on any other command.
  - verify: `golden_test.go` and `json_test.go` pass unmodified;
    `company-os graph build --fix` is a usage error; `args_test.go` case added.
  - landed:

- [ ] 1.4 Partition the first run and decide whether to repair (deps: 1.1, 1.3, est: ~35m)
  - why: repair must be triggered by evidence from a completed run, not by a
    guess made before one. A clean workspace must not take the write path at all.
  - acceptance: R-1.4, R-1.5, R-1.6, R-1.7 — execute the first run to completion
    before any repair pass; no derived-condition finding ⇒ no repair pass and the
    first run's results returned unchanged; at least one ⇒ exactly one repair
    pass; never more than one per invocation.
  - verify: fixture with only a decision-class failure takes the no-repair branch
    (assert zero writes); fixture with tag drift takes the repair branch exactly once.
  - landed:

- [ ] 1.5 Decision-class findings survive `--fix` untouched (deps: 1.4, est: ~25m)
  - why: this is the trust boundary of the whole feature. A reviewer must be able
    to believe that `--fix` never negotiated with a decision.
  - acceptance: R-1.10 — THE SYSTEM SHALL NOT alter, suppress, downgrade, reword,
    or re-order any decision-class finding as a consequence of `--fix` being supplied.
  - verify: SC2 — workspace with tag drift *and* a past-`reviewDate` deviation;
    assert the expiry finding under `--fix` is identical to plain `validate`'s,
    and the tags-drift finding is absent.
  - landed:

---

## Phase 2 — Repair execution (Unit 2)

- [ ] 2.1 Repair pass invokes `graph.Rebuild` (deps: 0.4, 1.4, est: ~30m)
  - why: `Rebuild` cannot disagree with the gates, because `WriteFeatureIndexes`
    and `FeatureIndexGate` share `sameCanonical` — the writer's "should I write"
    test and the gate's "has it drifted" test are the same function. A per-code
    repair path re-derives values a second way and can diverge (D2).
  - acceptance: R-2.1, R-2.2 — implement the repair pass by invoking
    `graph.Rebuild`; SHALL NOT contain any derivation of tags, context-node
    blocks, or feature indexes reachable only from the repair pass.
  - verify: read the diff — no new derivation functions; `go list -deps` shows no
    new import cycle from `internal/validate` onto `internal/graph`.
  - landed:

- [ ] 2.2 Hand-owned node guard (deps: 2.1, est: ~30m)
  - why: gate 5 calls a marker-less `CLAUDE.md` `node.hand-owned` at `SevOK`
    (`gates.go:157-160`); the writer appends a generated region to it
    (`node.go:96-103`). `--fix` moves that conversion from "you ran a build
    command" to "you fixed a tag typo." The guard sits in the repair path so
    `graph build` is unchanged for every existing caller (D10).
  - acceptance: R-2.11, R-2.12 — SHALL NOT write a `CLAUDE.md` the first run
    reported as hand-owned; WHERE one was skipped, report that skip, naming the file.
  - verify: fixture with a marker-less `CLAUDE.md` plus tag drift elsewhere —
    assert the node is byte-identical after `--fix` and the skip is reported.
  - landed:

- [ ] 2.3 Second run against the repaired workspace (deps: 2.1, est: ~25m)
  - why: repairing the feature index can change which references are unresolved.
    Filtering the first run's findings would report a state that never existed
    on disk.
  - acceptance: R-2.3, R-2.4 — WHEN a repair pass completes without error,
    execute the second run against the same workspace; derive reported gate
    results, for an invocation in which a repair pass completed, from the second
    run only.
  - verify: fixture where repair resolves a feature-index reference — assert the
    unresolved finding present in run one is absent from the output.
  - landed:

- [ ] 2.4 Idempotence and no-op guarantees (deps: 2.1, est: ~30m)
  - why: the writers already no-op when in sync (`PyEqual`, `writeIfChanged`,
    `sameCanonical`), so this is a property to *pin*, not to engineer. Pinning it
    is what stops a later "optimisation" from breaking it.
  - acceptance: R-2.6, R-2.7, R-2.8, R-2.10 — no write where content already
    equals its fresh derivation; byte-identical workspace when no
    derived-condition finding was produced; no bytes written on a second
    consecutive invocation; `workspace.lock.yaml` never modified.
  - verify: SC3 — clean example workspace under `--fix` leaves
    `git status --porcelain` empty; run `--fix` twice and assert zero writes on
    the second (mtime or write-counter fixture).
  - landed:

- [ ] 2.5 Equivalence with the manual path (deps: 2.2, 2.3, est: ~30m)
  - why: Goal 5 — `--fix` is a shortcut, not a second implementation. If the two
    paths can diverge, the feature is a new source of truth.
  - acceptance: R-2.9 — WHERE a workspace contains no hand-owned node, leave its
    tree byte-identical to the tree produced by `graph build` followed by
    `validate` from the same starting state. R-2.5 — no network access during a
    repair pass.
  - verify: SC4 — same drifted starting fixture down both paths, `diff -r` the
    trees; run under `--frozen` to confirm no network dependency.
  - landed:

---

## Phase 3 — Unrepairable writes (Unit 3)

- [ ] 3.1 Read-only slice fixture (deps: 0.2, est: ~45m)
  - why: `examples/acceptance.sh` has no `--fix` coverage and cannot grow it by
    accident. This fixture must run `workspace sync` and then mutate a slice —
    no existing fixture does. Every clause below is untestable until it exists.
  - acceptance: fixture yields a drifted derived file inside a `0444` slice
    target whose `localDirectory` lands under `platforms/` (per D6, a bare
    `knowledge/` target is already refused by `SliceRel`).
  - verify: fixture is reachable from `make check`; asserts the file is mode
    `0444` before the repair attempt.
  - landed:

- [ ] 3.2 Classify permission failures structurally (deps: 3.1, est: ~25m)
  - why: parsing a rendered message is the coupling Unit 0a exists to remove.
    Reintroducing it here would make Phase 0a pointless.
  - acceptance: R-3.8 — classify a write failure as a permission error by testing
    the returned error against `fs.ErrPermission`, and SHALL NOT classify it by
    parsing a rendered message.
  - verify: grep the repair path for message-parsing; test asserts classification
    survives a change to the rendered message text.
  - landed:

- [ ] 3.3 The named finding and its remedy (deps: 3.2, est: ~35m)
  - why: the person who hits this cannot act on "permission denied" for a file
    they did not know existed and are not supposed to edit. They can act on
    "re-sync the slice."
  - acceptance: R-3.1, R-3.2, R-3.3, R-3.4 — report a finding naming the
    workspace-relative path; state the remedy (change the source repo, update its
    pin, re-run `workspace sync`); leave the file's content and mode unchanged;
    SHALL NOT chmod to force the write through.
  - verify: SC5 — assert the path appears in the output, the remedy text is
    present, and the file's bytes and mode are unchanged after the run.
  - landed:

- [ ] 3.4 Abort semantics and exit code (deps: 3.3, est: ~30m)
  - why: a read-only slice is a single structural cause; naming the first file
    names the cause. Continuing would mean converting three writer loops to
    accumulate-and-continue, changing what `graph build` and every scaffolding
    caller do on partial failure — behaviour no fixture covers (D5a).
  - acceptance: R-3.5, R-3.6, R-3.7, R-3.9 — abort the repair pass and do not
    execute the second run; report the first run's gate results together with the
    abort finding; exit non-zero; for a non-permission abort, exit with the code
    that failure produces when `--fix` is absent.
  - verify: SC5 exits non-zero; assert the reported gate set is the first run's,
    and that the banner's `complete: false` meaning is unchanged.
  - landed:

---

## Phase 4 — Reporting and exit codes (Unit 4)

- [ ] 4.1 Repair section as a `GateResult` (deps: 0.4, 2.2, est: ~30m)
  - why: nothing below `cmd/` prints or exits. The section carries per-file
    findings so the renderer prints concrete paths rather than a count — and it
    is populated from what the repair pass returned, not from a second walk.
  - acceptance: R-4.1, R-4.2, R-4.12 — report one section, distinct from the gate
    list, enumerating each written file and each skipped hand-owned node;
    populate it from the repair pass's findings without re-deriving the list;
    omit the section rather than emitting it empty when nothing happened.
  - verify: assert the section is absent on a clean-workspace `--fix` run and
    names exactly the written files on a drifted one.
  - landed:

- [ ] 4.2 Renderer branch with no `[n/total]` prefix (deps: 4.1, est: ~35m,
      mutex: render-golden)
  - why: `render.Validate` partitions on exactly one rule — banner, else gate —
    and prints every gate as `[%d/%d]` against the banner-carried total
    (`text.go:40-53`, `:65`). A repair section entering that list prints `[0/8]`
    or `[9/8]`.
  - acceptance: R-4.3 — render the repair section without a `[n/total]` gate
    prefix, and SHALL NOT include it in the gate count.
  - verify: new golden fixture — a drifted workspace under `--fix`; assert gate
    numbering is unchanged and the repair section carries no prefix.
  - landed:

- [ ] 4.3 Exactly one banner, from the second run (deps: 2.3, 4.1, est: ~25m)
  - why: the `[n/total]` denominator is carried on the banner, not derived from
    the gate list. Two runs produce two banners; emitting both would make the
    denominator ambiguous and the first run's results visible.
  - acceptance: R-4.4, R-4.5, R-4.9 — report exactly one workspace banner record
    per invocation; WHEN a repair pass completed, report the second run's banner;
    SHALL NOT emit the first run's gate results in any output stream.
  - verify: assert one banner record in the returned set; grep the text and JSON
    output of a repaired run for the first run's finding codes — none present.
  - landed:

- [ ] 4.4 Exit codes derived from the second run (deps: 4.3, est: ~25m)
  - why: the exit code must remain a truthful statement about the workspace as it
    now exists on disk. A `--fix` invocation that parsed is never a usage error —
    the invocation was actionable.
  - acceptance: R-4.6, R-4.7, R-4.8 — second run green ⇒ exit zero; second run
    failing ⇒ exit with the same code a non-`--fix` run reporting those same gate
    results would produce; never the usage exit code for a successfully parsed
    `--fix`.
  - verify: SC1 exits 0, SC2 exits 1, SC6 — add `--fix` cases to
    `exitcode_test.go` asserting no path yields `ExitUsage` (2).
  - landed:

- [ ] 4.5 `--json` emits one document (deps: 4.1, 4.3, est: ~25m)
  - why: a machine consumer asking "is this workspace green" must get the answer
    for the workspace as it now exists on disk — not two documents, and not the
    pre-repair set (D8).
  - acceptance: R-4.10 — WHERE `--json` is supplied together with `--fix`, emit a
    single JSON document containing both the repair section and the reported gate
    results.
  - verify: `--fix --json` output parses as one document; assert the repair
    section and the second run's gates are both present in it.
  - landed:

- [ ] 4.6 Architecture guard: no output or exit below `cmd/` (deps: 4.2, est: ~15m)
  - why: this is a standing repo invariant, and the repair path is exactly the
    kind of change that tempts a print statement. It costs one test to keep
    honest.
  - acceptance: R-4.11 — THE SYSTEM SHALL emit no output and perform no exit from
    any package below `cmd/`.
  - verify: extend the existing architecture test to cover the repair path; grep
    for `fmt.Print`/`os.Exit` under `internal/validate` and `internal/graph`.
  - landed:

---

## Traceability

Every story's acceptance clause is bound to implementation and tests by
`@spec req://uncle-os/derived-drift-repair@0.1#R-n.m` markers. Per the ontology
rule, every mandatory clause carries at least one test-side marker; a story is
not done until its marker exists on the test side.
