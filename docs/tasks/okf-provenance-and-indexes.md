---
type: tasks
id: tasks-okf-provenance-and-indexes
title: OKF Provenance and Indexes — Tasks
status: draft
---

# OKF Provenance and Indexes — Tasks

Source of truth: `docs/ears/okf-provenance-and-indexes.md` (Units 1–6).
Architecture constraints: `docs/lld/okf-provenance-and-indexes.md`.
Parent: `docs/ears/okf-v02-conformance.md` §"Deferred — decided, not built"
(D1, D2), promoted here.

Targets: `company-os-starter/internal/{graph,scaffold,product,render}`,
`company-os-starter/cmd/company-os`, `company-os-starter/docs/`,
`company-os-starter/templates/`, and the `examples/workspace` +
`examples/standalone-team` fixtures.

**Precondition:** task 1.6 of the parent reserves `generated:` and `verified:` as
inert names in `FRONTMATTER-CORE.md`. If that has not landed, do it first — this
plan makes the names live and a rename afterwards is expensive.

## Four things every task inherits

1. **The gate is `make check`** — gofmt + `go vet` + `go test ./...` +
   `examples/acceptance.sh`. A task that runs only the harness verifies less than
   the project's own gate.
2. **Templates live in two places and must move together.** Every scaffolded
   document is emitted from a built-in string in `internal/scaffold/template.go`
   *and* has a peer file under `company-os-starter/templates/`. The repo
   `CLAUDE.md` makes the sync a hard constraint.
3. **The goldens move in this change, deliberately.** Unlike the parent (R-5.6),
   this change adds gate-5 output lines and fixture files. Re-baseline in the same
   commit as the cause, review the diff line by line, and never run
   `acceptance.sh --update` to clear a red build (R-6.7).
4. **No new blocking check** (R-5.5, invariant I1). A new warn line is permitted;
   a failure that did not exist before is not.

## Two measurements that changed the plan

- **R-2.6 is one site, not seven.** The EARS names two derived-artifact paths and
  D-2.2 feared more. Both `rebuildGenerated` (`cmd/company-os/scaffold.go:46`) and
  `rebuildSections` (`cmd/company-os/product.go:26`) funnel through
  `graph.Rebuild` (`internal/graph/graph.go:59`), which shares its `rebuild()`
  helper with `graph.Build` (`:21`). Index generation goes in `rebuild()` once and
  every entry point inherits it. Task 2.3 is sized accordingly.
- **R-2.10 needed no new machinery.** `rewriteGeneratedBlock`
  (`internal/graph/node.go:100-150`) already implements leave-alone-and-report for
  marker-less files, and gate 5 already renders it as
  `"hand-owned, no generated markers (-> pass)"` (`internal/graph/gates.go:291`).
  Task 2.5 reuses it rather than writing a parallel path.

## Ordering rule

Unit 1 before Unit 2 — the generator has nothing to render without
`description:`. Unit 3 after Unit 2 — you cannot drift-check a file no writer
produces. Units 4–5 are independent of 1–3 and may run in parallel by a second
person; Unit 6 is last and verifies the whole diff at once.

## Mutex tags

`cli` (`internal/` + `cmd/`), `templates` (the two-place sync), `fixtures`
(`examples/workspace` + `examples/standalone-team`), `goldens`.

---

## Unit 1 — `description:` as a documented, emitted field

- [ ] 1.1 Document `description:` and its rubric in `FRONTMATTER-CORE.md` (est: ~35m)
  - why: The index is the field's only consumer, and the parent deferred
    `description` precisely because shipping it without one creates a field
    nothing reads. Writing the rubric first is what stops the Unit 6 backfill from
    degrading into filename restatement — the failure mode no gate can catch.
  - acceptance: R-1.1 — documented as recommended on every document, non-blocking,
    with the index named as consumer; R-1.2 — rubric requires at least one fact not
    derivable from filename, `title:`, `type:`, or directory path; R-1.3 — a
    description must not read as true when pasted onto a sibling.
  - verify: the rubric is stated as a rule a reviewer can apply to a specific
    document and get a yes/no, not as advice.
  - landed:

- [ ] 1.2 Emit `description:` from the scaffolding templates (deps: 1.1, est: ~50m, mutex: cli, templates)
  - why: If the shipped scaffolding does not emit a field the contract recommends,
    every agent following it produces a document that needs backfilling later.
  - acceptance: R-1.5 — `prd new`, `discover new`, and `reality new` emit
    `description:`, with `internal/scaffold/template.go` (`DiscoveryTemplate:27`,
    `PRDTemplate:58`) and the peer files under `templates/` (`discovery-brief.md`,
    `prd.md`, `reality-component.md`) moving together; R-1.4 — the placeholder value
    is quoted, since a useful description commonly contains a colon and would
    otherwise be a YAML parse error.
  - verify: scaffold each of the three document types into a scratch workspace,
    confirm the field is present, quoted, and the result passes `validate`;
    `make check` green — that is what catches a built-in string drifting from its
    template peer.
  - landed:

- [ ] 1.3 Emit `description:` from the `outcome.md` writer (deps: 1.1, est: ~25m, mutex: cli, templates)
  - why: The fourth document-emitting path and the easiest to miss —
    `prd complete` writes `outcome.md` inline. Without this, Unit 6 backfills the
    fixture's committed copy and the very next `prd complete` emits one without the
    field. `outcomeDoc` (`internal/product/prd.go:569-574`) is the same writer task
    3.3 of the parent touches for `title:`; check whether that has landed and move
    both together if not.
  - acceptance: R-1.6 — the writer emits `description:`;
    `templates/outcome-review.md` moves with it.
  - verify: run `prd complete` on a scratch workspace; the emitted `outcome.md`
    validates and carries the field.
  - landed:

- [ ] 1.4 Confirm omission stays silent (deps: 1.2, 1.3, est: ~15m)
  - why: R-1.7 is the backward-compatibility guarantee for this unit, and it is
    cheap to assert now and expensive to discover violated in Unit 6.
  - acceptance: R-1.7 — a document omitting `description:` validates unchanged,
    producing neither an error nor a warning.
  - verify: strip `description:` from a scratch-workspace document; `validate`
    output is byte-identical to before the strip.
  - landed:

---

## Unit 2 — Per-directory `index.md` generation

- [ ] 2.1 Register the inert `index` type and extend `skipNames` (est: ~30m, mutex: cli)
  - why: These two must land in the same commit as each other and before the
    generator. A typed `index.md` that `IterGraphDocs` still yields would be
    counted toward its own directory's threshold and listed inside itself — the
    re-ingestion loop D-2.5 exists to prevent.
  - acceptance: R-2.4 — `index` registered in the type→tag vocabulary
    (`internal/graph/tags.go:24-28`) as inert, with no required-field gate;
    R-2.5 — `"index.md"` added to `skipNames` (`internal/graph/tags.go:41`,
    currently `log.md`, `README.md`, `CLAUDE.md`).
  - verify: a hand-placed `index.md` carrying `type: index` is absent from
    `IterGraphDocs` output, covered by a Go test asserting the skip.
  - note: the LLD records that this type earns nothing internally — skipped means
    no derived tags and no gate 4 coverage. It is for external consumers. Do not
    later "fix" the skip.
  - landed:

- [ ] 2.2 Write the index renderer (deps: 1.1, 2.1, est: ~75m, mutex: cli)
  - why: The progressive-disclosure payoff. An agent opening
    `platforms/communications/reality/` today must `ls` and open files to learn
    what is there.
  - acceptance: R-2.1 — an index is generated in any directory directly holding
    ≥2 graph documents and in none below that; R-2.2 — each entry renders title +
    `description:`, grouped by `type:`, reusing the fallback chain at
    `internal/graph/node.go:319-324`; R-2.3 — frontmatter carries `type: index`
    plus a `company-os:generated` marker block.
  - verify: unit test over a synthetic tree covering the 1-doc, 2-doc, and 3-doc
    cases; assert the threshold boundary explicitly, since off-by-one here is
    silent.
  - landed:

- [ ] 2.3 Wire generation into `rebuild()` (deps: 2.2, est: ~30m, mutex: cli)
  - why: R-2.6 exists so `company-os init` does not emit a workspace that fails its
    own `validate` — the D-2.2 failure. **Re-measured: this is one site, not
    seven.** `graph.Rebuild` (`internal/graph/graph.go:59`) and `graph.Build`
    (`:21`) share the `rebuild()` helper, and both `cmd` seams funnel through it.
  - acceptance: R-2.6 — every derived-artifact path generates indexes;
    R-2.9 — a directory dropping below two graph documents has its generated
    `index.md` deleted; R-2.11 — two consecutive `derive` runs leave the workspace
    byte-identical (I6).
  - verify: `examples/acceptance.sh` §4 double-build check (`s0 == s1 == s2`)
    stays green on both monorepo fixtures — that is the idempotency oracle.
    Separately assert the deletion path, which the double-build check does not
    cover.
  - landed:

- [ ] 2.4 Exclude slice roots and `knowledge/` (deps: 2.3, est: ~40m, mutex: cli)
  - why: Invariant I9. `examples/federated/` materializes graph documents inside a
    `0444` slice whose bytes are hashed into `workspace.lock.yaml`; a write there
    fails gate `[8/8]`. The drift check must skip the same paths or a slice
    directory that qualifies but cannot be written reports permanent drift.
  - acceptance: R-2.7 — no index generated or drift-checked beneath a
    manifest-declared slice root; R-2.8 — none anywhere under `knowledge/`, which
    `IterGraphDocs` already omits as a graph-docs root
    (`internal/graph/tags.go:191`).
  - verify: `company-os --root examples/federated derive` then
    `git diff --stat examples/federated` is empty; `validate` on that fixture
    exits 0 at `[8/8]`.
  - landed:

- [ ] 2.5 Honour a hand-written `index.md` (deps: 2.3, est: ~20m, mutex: cli)
  - why: **Corrected during pre-mortem.** The first spec draft would have failed
    gate 5 on a marker-less `index.md`. That contradicts the decision this repo
    already settled for `CLAUDE.md` and would have added a blocking check,
    violating R-5.5 and I1. Reuse the existing path rather than writing a parallel
    one.
  - acceptance: R-2.10 — a marker-less `index.md` is left untouched, reported as
    hand-owned at severity OK, and passes — via `rewriteGeneratedBlock`
    (`internal/graph/node.go:100-150`), rendered as at
    `internal/graph/gates.go:291`.
  - verify: place a hand-written `index.md` in a qualifying directory; `derive`
    leaves it byte-identical and `validate` exits 0 with a hand-owned line.
  - landed:

---

## Unit 3 — Index drift detection

- [ ] 3.1 Report index drift inside gate 5 (deps: 2.3, est: ~55m, mutex: cli, goldens)
  - why: An index is a generated artifact and I4 makes derived-vs-authored a hard
    boundary. A silently stale index is worse than none, because a reader trusts
    it. Gate 5 is the right home: adding a gate would renumber, violating I5.
  - acceptance: R-3.1 — reported inside gate 5, no gate added, none renumbered;
    R-3.2 — gate 5's header stays byte-identical to
    `[5/N] CLAUDE.md context node drift (fail-safe, absence-tolerant)`;
    R-3.3 — a committed index differing from a fresh derivation fails gate 5 and
    names the file; R-3.4 — a qualifying directory with no index fails gate 5 and
    names the directory; R-3.5 — one aggregate line per federation root.
  - verify: `NodeGate` (`internal/graph/gates.go:127`) is extended rather than
    duplicated; the header string is asserted byte-identical by test, not by eye —
    it is a frozen string under I5 and the header will now under-describe its
    contents by design.
  - landed:

- [ ] 3.2 Confirm absence tolerance survives (deps: 3.1, est: ~20m)
  - why: Invariant I2 is the standalone-team on-ramp promise. A drift check that
    assumes four federation roots breaks the fixture that proves three can be
    missing.
  - acceptance: R-3.6 — `examples/standalone-team/` keeps passing `validate` with
    three of four federation roots absent, and generates zero indexes.
  - verify: `company-os --root examples/standalone-team validate` exits 0;
    `find examples/standalone-team -name index.md` returns nothing.
  - landed:

---

## Unit 4 — `generated:` and `verified:` provenance

- [ ] 4.1 Define the field shapes and actor convention (est: ~40m)
  - why: The largest conceptual gap in the system. The whole `skills/` layer exists
    to direct agents producing PRDs and briefs, and nothing records that a document
    was agent-drafted rather than written by a person.
  - acceptance: R-4.1 — `generated: {by, at}` for origin and
    `verified: [{by, at}]` for confirmation, documented as distinct because an
    author need not be its confirmer; R-4.2 — actors are `<producer>/<version>`,
    `human:<id>`, or `process:<id>`.
  - verify: `FRONTMATTER-CORE.md` promotes both from the reserved-inert list the
    parent's task 1.6 created, and the two documents do not disagree about the
    shapes.
  - landed:

- [ ] 4.2 Prove nested-collection round-trip (deps: 4.1, est: ~50m, mutex: cli)
  - why: This is the one place the change can corrupt user data silently.
    `TestRewriteFrontmatterTagsPreservesUnknownKeys` uses a plain string scalar and
    does **not** cover a mapping or a sequence-of-mappings. Those re-layout through
    `PyDumpAutoFlow` (`internal/graph/tags.go:302`) — the exact site of the
    flow/block divergence recorded as R-0.7a(g) in the port.
  - acceptance: R-4.5 — a document carrying both fields validates and preserves
    them byte-identical through `derive`; R-4.6 — covered by a Go test written
    against the mapping and sequence-of-mappings shapes specifically.
  - verify: the test asserts byte equality of the frontmatter block before and
    after `derive`, not just key presence. Run it twice — idempotency here is what
    the double-build check would otherwise catch far too late.
  - landed:

- [ ] 4.3 Derive the trust tier (deps: 4.1, est: ~35m, mutex: cli)
  - why: The signal the fields exist to produce. Keeping it computed-at-read and
    unstored is what stops it becoming a fourth approval field that can disagree
    with the other three.
  - acceptance: R-4.3 — absent yields *unverified*, no `human:` actor yields
    *machine-confirmed*, any `human:` actor yields *human-reviewed*;
    R-4.4 — advisory, computed at read time, never stored, and no gate or access
    decision depends on it.
  - verify: grep confirms no gate consumes the tier; a table test covers the three
    tiers plus the empty-list edge case.
  - landed:

- [ ] 4.4 Emit `generated:` where the producer is known (deps: 4.1, 4.2, est: ~40m, mutex: cli, templates)
  - why: Without an emitter the field only ever appears if a human types it, which
    is precisely the population it is least useful for.
  - acceptance: R-4.7 — a producing command writing a document emits
    `generated: {by, at}`; R-4.8 — a document omitting both fields validates
    unchanged and resolves to *unverified*.
  - verify: scaffold a document and confirm the field is present and well-formed;
    strip both fields from another and confirm `validate` output is unchanged.
  - landed:

---

## Unit 5 — One-way derivation from existing approval fields

- [ ] 5.1 Derive `verified:` from `decisionOwner` / `approvedBy` (deps: 4.3, est: ~45m, mutex: cli)
  - why: Without it the trust tier reads *unverified* for every document in every
    fixture on the day it ships — a signal with no signal. Derivation makes it
    meaningful without touching approval semantics, which is what made full
    reconciliation too large to carry here.
  - acceptance: R-5.1 — a `verified:` entry is derived in memory for tier
    computation; R-5.2 — `decisionOwner:` and `approvedBy:` are never rewritten,
    migrated, or removed, and every gate consuming them is unchanged;
    R-5.3 — a literal `TODO` value derives nothing.
  - verify: `git diff` over the fixtures after a full `derive` shows no change to
    any approval field; a test asserts the `TODO` case derives an empty list, not
    an entry with a `TODO` actor.
  - landed:

- [ ] 5.2 Warn on a literal `TODO` in `approvedBy` (deps: 5.1, est: ~30m, mutex: cli, goldens)
  - why: `internal/scaffold/template.go:67` scaffolds `decisionOwner: TODO` and it
    hard-fails `prd validate`; `internal/governance/declare.go:121` scaffolds
    `approvedBy: 'TODO: rule owner'` and it passes. The asymmetry is real and
    currently invisible. Making it block would fail `examples/workspace` today.
  - acceptance: R-5.4 — a warning naming the file, and validation does not fail;
    R-5.5 — no new blocking check is introduced anywhere in this change.
  - verify: `validate` on `examples/workspace` still exits 0 and now carries one
    warn line for
    `teams/customer-engagement/governance/exceptions.yaml:8`; the golden moves by
    exactly that line and no other.
  - note: this is a deliberate departure from the parent's R-5.3 ("no new warn
    line"), which constrained a documentation-only change and does not bind this
    one. Recorded in the LLD as Key Decision 3.
  - landed:

---

## Unit 6 — Fixtures, compatibility, and verification

- [ ] 6.1 Backfill `description:` across the two monorepo fixtures (deps: 1.4, 2.2, est: ~80m, mutex: fixtures)
  - why: This is where the rubric earns its keep or the whole unit degrades into
    filename restatement. Do it with the index rendered, side by side — the test
    for a bad description is whether it still reads true pasted onto a sibling, and
    that only works when siblings are visible together.
  - acceptance: R-6.1 — every document in `examples/workspace/` and
    `examples/standalone-team/` carrying frontmatter `type:` has a `description:`
    satisfying R-1.2 and R-1.3.
  - verify: `grep -rL "^description:" examples/workspace examples/standalone-team
    --include='*.md'` returns only the untyped files (the 5 generated `CLAUDE.md`,
    2 `EXAMPLE_README.md`, and `platforms/communications/log.md` — the same
    allowlist the parent's task 3.4 established); then read every new description
    against its siblings and cut the ones that pass the grep but fail R-1.3.
  - landed:

- [ ] 6.2 Regenerate indexes and `CLAUDE.md` blocks in the same commit (deps: 6.1, est: ~25m, mutex: fixtures, goldens)
  - why: Adding descriptions changes what both the index renderer and
    `buildClaudeNode` emit, so committed generated state goes stale the moment 6.1
    lands. The harness's double-build check requires the committed workspace to be
    already fully derived — this cannot be a follow-up commit.
  - acceptance: R-6.8 — regenerated indexes and `CLAUDE.md` blocks committed
    alongside the backfill; R-2.12 — exactly three `index.md` in
    `examples/workspace/`, zero in `examples/standalone-team/`.
  - verify: `examples/acceptance.sh` §4 reports "committed state fully derived +
    idempotent" for both monorepo fixtures; `find examples -name index.md | wc -l`
    is 3.
  - landed:

- [ ] 6.3 Add the `init` → `validate` acceptance case (deps: 2.3, est: ~40m)
  - why: **Found in pre-mortem — this is a real coverage hole, not a nicety.**
    `examples/acceptance.sh` never runs `company-os init` followed by `validate`;
    its only `init` is `git init` for the federated source fixture (`:171`). The
    D-2.2 failure mode that R-2.6 exists to prevent would therefore ship
    undetected, because the double-build check covers committed fixtures only.
  - acceptance: R-6.10 — an acceptance case runs `company-os init` into a scratch
    directory then `company-os validate`, asserting exit 0.
  - verify: temporarily revert task 2.3's wiring and confirm the new case goes red.
    A test that cannot fail is not a test.
  - landed:

- [ ] 6.4 Re-baseline the goldens deliberately (deps: 3.1, 5.2, 6.2, est: ~35m, mutex: goldens)
  - why: This change legitimizes moving files the parent froze, which is exactly
    how a regression gets buried. The defence is that every moved line is
    predicted before it is accepted.
  - acceptance: R-6.7 — goldens re-baselined in the same commit as the change
    causing the move, diff reviewed line by line, and `acceptance.sh --update`
    never run to clear a failing build.
  - verify: before re-baselining, write down the expected diff — new gate-5 index
    lines, one warn line from 5.2, and nothing else. Any line outside that
    prediction is a regression to diagnose, not a baseline to refresh.
  - landed:

- [ ] 6.5 Backward-compatibility pass (deps: 6.3, 6.4, est: ~45m)
  - why: Every requirement here is a veto over the rest of the change. Verified
    once at the end against the whole diff rather than assumed per unit.
  - acceptance: R-6.2 — `git diff --stat examples/federated` empty;
    R-6.3 — federated `validate` exits 0 at `[8/8]` with lock hashes valid (I9);
    R-6.4 — `examples/banking/` untouched, as the harness does not exercise it;
    R-6.5 — any workspace passing before still passes, beyond one documented
    `derive` re-run (I1); R-6.6 — unknown fields and types preserved (I7) and the
    mandatory/default/guidance tier model untouched (I3).
  - verify: `make check` exits 0; `git log -p` for the change contains no
    `acceptance.sh --update`; `git diff` against the merge base shows nothing under
    `examples/federated/` or `examples/banking/`.
  - landed:
