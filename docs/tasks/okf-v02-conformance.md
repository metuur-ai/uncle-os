---
type: tasks
id: tasks-okf-v02-conformance
title: OKF v0.2 Conformance — Tasks
status: draft
tags: [kind/tasks, status/draft]
---

# OKF v0.2 Conformance — Tasks

Source of truth: `docs/ears/okf-v02-conformance.md` (Units 0–3, 5–6 committed).
Architecture constraints: `docs/lld/okf-v02-conformance.md`.
Targets: `company-os-starter/docs/`, `company-os-starter/templates/`, the
built-in template strings and done-gate in `company-os-starter/internal/`, and the
`examples/workspace` + `examples/standalone-team` fixtures.

## Re-plan against the Go binary (2026-07-27)

**Authority:** R-9.8 of `docs/ears/go-cli-tui-port.md`. This plan was written
while `company-os-starter/bin/company-os` was the implementation; R-9.3 deleted
it. The requirement dispositions and the full Python→Go anchor map live in
[Amendment 2](../ears/okf-v02-conformance.md#amendment-2--re-plan-against-the-go-binary-2026-07-27)
of the EARS. What follows is what changes for *execution*.

**Phase 0 is closed** (task 0.1, 2026-07-27) — it was the only phase the port
actually completed, via task 2.6 of the port.

**Four things every remaining task inherits:**

1. **The gate is `make check`, not `bash examples/acceptance.sh` alone.**
   `make check` is gofmt + `go vet` + `go test ./...` + `acceptance.sh`. The
   global acceptance below still holds, but a task that runs only the harness now
   verifies less than the project's own gate. Where a task says "run
   `acceptance.sh`", run `make check`.
2. **"No CLI diff" is now a Go diff.** Tasks 1.6 and 4.1 verify with
   `git diff --stat company-os-starter/bin/company-os`, which can only ever be
   empty now. The equivalent assertion is
   `git diff --stat company-os-starter/internal company-os-starter/cmd`.
3. **Templates live in two places and must move together.** Every scaffolded
   document is emitted from a built-in string in
   `internal/scaffold/template.go` *and* has a peer file under
   `company-os-starter/templates/`. The repo `CLAUDE.md` requires them to stay in
   sync with the section names the corresponding `validate` greps for. Task 3.2
   was planned against one file and now touches two — this is the single largest
   sizing change in the re-plan.
4. **Gate numbering survived the port; the gate *count* did not.** `validate` is
   `[1/7]`…`[7/7]` on a monorepo and `[1/8]`…`[8/8]` federated. Gates 3, 4, 5 and
   8 still mean what this plan says they mean (active PRD contracts, frontmatter
   core, CLAUDE.md drift, slice integrity), so every gate reference below is
   still correct — verified against both committed goldens, not assumed.

**Two tasks are resized by re-measurement** (full table in Amendment 2):

- **3.4** — the backfill is **14 of 16** frontmatter documents, not 12 of 17. The
  port added typed documents. Larger, not smaller.
- **1.1** — the audit is cheaper than planned. Its evidence table was to be built
  by reading `bin/company-os`; the Go port carried every cited Python line
  forward as a source comment, so the map already exists and the audit becomes a
  verification pass over Amendment 2's anchor map rather than a fresh read.

**One task is now partly satisfied.** 3.3 (`outcome.md` emits `title:`) —
`outcomeDoc` (`internal/product/prd.go:569-574`) already threads the PRD's title
through and renders it as the H1. The frontmatter `title:` is still absent, so the
task stands, but it is an additive one-line change to a writer that already has
the value in hand, not a plumbing exercise.

**One unresolved risk is unchanged and still blocks 2.3.** The federated fixture
pins `https://git.example.com/acme/platform-communications.git`, which does not
resolve, and `examples/federated/workspace.yaml:5` states the fixture is designed
to pass with no network and no source repo. "Correct at source and re-sync" (R-2.5)
is still not executable as written. Decide before editing anything.

**What this re-plan did NOT do.** It did not add requirements. Amendment 2 records
three things the conformance document arguably now owes an adopter — the exit-code
contract, the `--json` surface, and the `knowledge/` root, none of which existed
when Unit 1 was locked — but records them as findings, not clauses. R-9.8
authorized a re-plan, not a re-scope. Deciding whether they enter R-1.1's section
list is a scoping call for whoever picks up 1.2.

**Global acceptance (must hold after every phase):**
`bash examples/acceptance.sh` exits 0 (R-6.1), **and** both
`examples/golden-validate.txt` and `examples/federated-golden-validate.txt` are
byte-identical to their committed state (R-5.6). `acceptance.sh --update` is not
run at any point in this change — a golden diff is a regression to diagnose, not a
baseline to refresh.

**Deferred — NOT planned here:** EARS Units D1 (provenance/trust — `generated`,
`verified`, actor convention, trust tiers) and D2 (`description:` + per-directory
`index.md`). D2 additionally blocked on two design answers: excluding
manifest-declared slice roots, and wiring generation through `rebuild_generated`'s
seven call sites.

**Ordering rule:** Phase 0 is standalone and lands first (one line, no dependents).
Phase 1 writes the contract that Phases 2–3 implement against — 1.1 is a read-only
audit and must precede the clause it feeds. Phase 3 is the only phase touching
fixture frontmatter, and its CLAUDE.md regeneration (3.6) must land in the same
commit as the title backfill (3.4) or the harness's double-build check goes red.

---

## Phase 0 — Done-gate date correctness (Unit 0)

- [x] 0.1 Parse dates in the `prd complete` done-gate (est: ~30m)
  - **CLOSED 2026-07-27.** The task recorded its own exit condition — open "until
    either the Python side is fixed or R-9.3 deletes it." R-9.3 deleted it:
    `examples/selftest.py` and `company-os-starter/bin/company-os` are both gone
    (tasks 6.5, 6.10). The only clause holding this open named a file that no
    longer exists, so it was retired as
    [Amendment 1](../ears/okf-v02-conformance.md#amendment-1--r-04-partial-retirement-2026-07-27)
    rather than silently treated as satisfied. R-0.4's other half — `acceptance.sh`
    exits 0 — is unchanged and still enforced by `make check`.
  - **Re-verified against the binary before closing, not taken from the note
    below.** `TestParseDate` and `TestDoneGateNamesAMalformedDate` pass, and a
    reality doc carrying `updated: 18/07/2026` makes `prd complete` refuse with
    `cannot compare dates: … has 'updated: 18/07/2026', which is not an ISO-8601
    date (YYYY-MM-DD)` — file named, value named, no traceback, no silent pass.
    R-0.1, R-0.2 and R-0.3 are met on the shipped implementation.
  - The original PARTIALLY DONE note is kept below verbatim, because it is the
    record of *why* this sat open for a day: the divergence it describes between
    the Go and Python gates was real at the time and was resolved by deleting one
    of the two, not by reconciling them.
  - PARTIALLY DONE 2026-07-26, in Go only. Task 2.6 of the Go port landed R-0.1,
    R-0.2 and R-0.3 in `company-os-starter/internal/product/prd.go` (`parseDate`,
    `CodeDoneRealityDateInvalid`): both values are parsed as ISO-8601 and
    compared as dates, and a malformed or absent one produces a done-check error
    naming the file and the offending value. `bin/company-os:679-682` is
    UNCHANGED — the port's constraints forbid touching it — so the two
    implementations now differ here by design; the difference is invisible to
    `examples/differential.py` because no corpus fixture carries a malformed
    date (measured: 6 reality docs, 6 PRDs, all well-formed ISO). R-0.4's
    `examples/selftest.py` check is therefore still open, and so is this task
    until either the Python side is fixed or R-9.3 deletes it. The Go-side
    coverage is `TestParseDate`, `TestDoneGateNamesAMalformedDate`,
    `TestDoneGateAcceptsAFreshRealityDoc` and `TestDoneGateStaleRealityRefuses`
    in `internal/product/product_test.go`.
  - why: Invariant #4 of the whole methodology is "a change is done only when
    reality is updated", and the gate enforcing it compares dates as raw strings
    (`bin/company-os:679-682`) — correct for well-formed ISO dates by lexical
    accident, silently wrong for `18/07/2026`, an empty value, or a YAML-parsed
    `datetime.date`. Documenting this gate in a normative spec while it
    misbehaves is worse than not writing the spec.
  - acceptance: R-0.1 — parse both `updated:` and `created:` as ISO-8601 dates and
    compare parsed values; R-0.2 — a malformed/absent value produces a done-check
    error naming the file and the offending value, never an unhandled exception;
    R-0.3 — well-formed input yields the same accept/refuse outcome as today;
    R-0.4 — an automated test covers R-0.2 (~~a check in `examples/selftest.py`~~,
    retired 2026-07-27 with the file itself).
  - verify: `go test ./internal/product/` passes including
    `TestDoneGateNamesAMalformedDate`; feed a reality doc `updated: 18/07/2026`
    and confirm a named error rather than a traceback or a silent pass;
    `bash examples/acceptance.sh` exits 0 with both goldens unchanged.

---

## Phase 1 — Conformance and versioning contract (Unit 1)

- [ ] 1.1 Audit every blocking field check in the CLI (est: ~40m)
  - why: A conformance clause that contradicts shipped behaviour is worse than
    none. `title` cannot be written as a plain SHOULD — it blocks today for
    `type: prd` at gate 3 (`:975`) and `prd validate` (`:627-630`). This audit
    produces the evidence table that 1.2's clause is written against, so it has to
    come first. Read-only; no files change.
  - **Re-planned 2026-07-27 — cheaper than estimated.** The evidence table was to
    be built by reading `bin/company-os`, which no longer exists. The port carried
    every cited Python line forward as a Go source comment, so this becomes a
    verification pass over Amendment 2's anchor map rather than a fresh read.
    Re-estimate ~20m. Confirm each Go site says what the map claims; a site whose
    comment cites a Python line but implements something else is the finding this
    task exists to catch.
  - acceptance: R-1.13 — every field that blocks anywhere is identified across
    `CoreFieldErrors` (`internal/product/contract.go:65`), gate 3
    (`internal/product/check.go:89`), and the `prd validate` required-field list
    (`internal/product/prd.go:23`); R-1.4 (as amended) — every candidate
    MUST-NOT-reject item has a Go package, symbol and line that already honours it.
  - verify: the audit table names a Go file:line for each entry, and each cited
    line has been opened and confirmed to say what the table claims — an uncited or
    unverified item is a bug in the claim, not a formatting gap. Record in
    `.devlocal/<user>/1.1/scratchpad.md`.

- [ ] 1.2 Write `CONFORMANCE.md` — clause, MUST-NOT-reject list, terminology (deps: 1.1, est: ~90m)
  - why: Company OS has never defined "conformant". Adopters — especially the
    solo team in `examples/standalone-team/` — have no way to answer "am I doing
    this right?", which is what makes the on-ramp promise an assertion rather than
    a specification.
  - acceptance: R-1.1 — the document exists with Goals, Non-Goals, Terminology,
    conformance clause, MUST-NOT-reject list, versioning, and considered-and-
    deferred as distinct sections; R-1.2 — RFC-2119 clause declaring `type` + an
    identity field (`id`, or `prd` for outcome reviews) as the floor, `title` and
    `description` recommended on all documents, and `resource` recommended **only**
    on documents describing an external asset; R-1.3 — MUST-NOT-reject covers
    unknown types, unknown keys, broken cross-links, missing optional documents,
    and missing federation roots; R-1.10 — Terminology defines *reality*,
    *deviation*, *exception*, *component*, *canonical*, *authority*, *tier*,
    *federation root*, *generated artifact*, *absence tolerance*.
  - verify: every MUST-NOT-reject item carries a file:line from 1.1; `title` is
    documented as conditionally required with its two blocking sites cited; the
    document stays under ~150 lines — past that it is drifting into
    `FRONTMATTER-CORE.md`'s territory and should shed a section rather than sprawl.

- [ ] 1.3 Define the version scheme and reconcile the `profile:` enum (deps: 1.2, est: ~45m)
  - why: Eight fixtures commit `companyOsVersion: '2026.2'` against a version
    defined nowhere, and two commit `profile: minimal` against a documented enum
    (`docs/01-flexibility-skills-and-role-views.md:121`) that lists only
    `standard | strict | provisional`. Writing the contract against the documented
    enum would make two shipped fixtures non-conformant on the day it ships — the
    exact failure this change exists to prevent.
  - acceptance: R-1.5 — `2026.2` is the first version carrying a conformance
    clause, everything prior is `unversioned`, no retroactive `2026.1` is
    reconstructed; R-1.6 — forward-only bump rule (the version moves when the
    clause changes what tooling MUST accept or reject), and this change does not
    bump it; R-1.7 — the enum becomes `minimal | standard | strict`, `provisional`
    is dropped as unused, and `01-flexibility-skills-and-role-views.md:121` is
    corrected.
  - verify: `grep -rn "conformance:" examples/` — all eight fixtures conformant
    against the new document with **zero fixture edits**; `grep -rn "provisional"`
    returns nothing outside history.

- [ ] 1.4 State the OKF relationship and retire the stale v0.1 claims (deps: 1.2, est: ~40m)
  - why: The repo declares "Documentation standard: OKF v0.1" in its founding
    proposal while OKF has published v0.2, which supersedes it and retires
    `timestamp:`. Leaving the claim unqualified is how the drift went unnoticed for
    this long.
  - acceptance: R-1.8 — the targeted OKF version is declared, with a per-field
    statement of whether Company OS adopts, supersedes, or declines each
    OKF-recommended field; R-1.9 — `docs/00-original-proposal.md:3` and
    `docs/01-flexibility-skills-and-role-views.md:8` are updated or annotated as
    historical.
  - verify: `grep -rn "OKF v0.1" company-os-starter/` returns only lines that are
    explicitly marked historical.

- [ ] 1.5 Write the considered-and-deferred section (deps: 1.2, est: ~30m)
  - why: A decided "no" that isn't written down reads as an oversight and gets
    re-proposed every quarter. This is where provenance/trust, `sources[]`,
    attested computation, `stale_after`, the ontology roadmap, and the GitHub-sync
    request go to stop being open questions.
  - acceptance: R-1.11 — N1–N8 from the HLD each recorded with one sentence of
    reasoning, including N8 (repository directory sync) pointing at
    `.devlocal/research/2026-07-25-knowledge-catalog-directory-sync.md`.
  - verify: each of N1–N8 appears with a reason a reader could act on, not just a
    restatement of the title.

- [ ] 1.6 Reserve `generated:`/`verified:` and hand off the contract claim (deps: 1.2, est: ~40m)
  - why: Provenance is the next change, and reserving the names now means it lands
    without a rename. Separately, `FRONTMATTER-CORE.md:8-12` currently calls itself
    "the whole interop contract" — leaving that in place alongside `CONFORMANCE.md`
    produces two contracts and one confused reader.
  - acceptance: R-1.12 — both names reserved as inert in `FRONTMATTER-CORE.md`
    following the reserved-doc-types pattern at `:113-118`, with intended meaning
    and a pointer to the deferred change; R-1.14 — `:8-12` amended so only one
    document claims to be the contract, with a link to `CONFORMANCE.md`;
    R-1.15 (as amended) — a document carrying both fields validates and survives
    `graph build` unchanged, covered by an automated Go test.
  - verify: `make check` passes including the new preserve test; R-1.16 —
    `git diff --stat company-os-starter/internal company-os-starter/cmd` is empty
    for this phase apart from that test, and both goldens are unchanged.

---

## Phase 2 — Roadmap and shipped capability separated (Unit 2)

- [ ] 2.1 Split the unshipped ontology material into `ONTOLOGY-ROADMAP.md` (est: ~60m)
  - why: `ONTOLOGY-GUIDE.md` reads as reference documentation for six capabilities
    that do not exist — grep of `company-os-starter/internal` and `cmd` returns
    **zero** hits for `spec trace`, `@spec`, `validate --ontology` and `ears:`
    (re-verified against the Go CLI 2026-07-27; the original claim was measured
    against `bin/company-os`, and it survives the port unchanged). A reader
    following it writes `@spec` markers no
    scanner will ever read. A banner on a section inside a 368-line document is
    easy to scroll past; a separate file cannot be read by accident.
  - acceptance: R-2.1 — `ears:` blocks, `@spec` markers, `spec trace`,
    `validate --ontology`, vocabulary linting, `## Graph` wikilink blocks, and
    per-clause PRD checklists all move out; R-2.2 — the roadmap opens with a
    not-yet-available banner in the style of `observer-roadmap.md:5-9`, and
    `ONTOLOGY-GUIDE.md` links forward to it.
  - verify: everything remaining in `ONTOLOGY-GUIDE.md` is runnable with the
    shipped CLI (canonical IDs, `ids/registry.yaml`, tag derivation, `graph build`).

- [ ] 2.2 Correct the false capability claim in the workspace fixture (est: ~15m)
  - why: `examples/workspace/company-ontology/contexts/communications.md:20`
    asserts "`company-os validate --ontology` flags forbidden terms in canonical
    docs" as shipped fact. Fixtures are what adopters copy, so a false claim there
    is worse than one in the guide. The banking fixture already gets this right at
    `payments.md:20` — "(`validate --ontology`, roadmap)" — so the wording exists.
  - acceptance: R-2.3 — the workspace fixture uses the banking wording.
  - verify: `bash examples/acceptance.sh` exits 0; gate 4's per-document output is
    path-only so the golden stays unchanged — confirm it did.

- [ ] 2.3 Correct the same claim in the federated slice (deps: 2.2, est: ~45m, mutex: federated-fixture)
  - why: `examples/federated/company-ontology/contexts/communications.md:20`
    carries the identical false claim, but it lives inside a read-only materialized
    slice whose bytes are hashed into `workspace.lock.yaml`. Editing it in place
    breaks gate 8 (I9).
  - acceptance: R-2.5 — the claim is corrected without invalidating lock hashes.
  - verify: `company-os --root examples/federated validate` exits 0 at `[8/8]`;
    `examples/federated-golden-validate.txt` unchanged.
  - **risk — resolve before starting:** the fixture pins
    `https://git.example.com/acme/platform-communications.git`
    (`examples/federated/workspace.yaml:9`), which does not resolve, so
    "correct at source and re-sync" is not executable as written. Options are to
    regenerate the fixture and its lock wholesale, or to leave the federated copy
    and scope R-2.4's check to exclude it with a recorded reason. Decide before
    editing anything — do not hand-edit the slice.

- [ ] 2.4 Repo-wide check that no unshipped capability is asserted as fact (deps: 2.1, 2.2, 2.3, est: ~15m)
  - why: The point of the split is that a reader cannot reach a `spec trace`
    example without first being told it does not exist. That is only true if the
    check covers fixtures, not just `docs/`.
  - acceptance: R-2.4 — every hit from
    `grep -rn "validate --ontology\|spec trace\|@spec" company-os-starter/ examples/`
    sits inside a bannered document/section or is worded as roadmap;
    `docs/{hld,lld,ears}/` are exempt as they legitimately discuss deferred work.
  - verify: run the grep and walk every hit.

---

## Phase 3 — `title` and `resource` (Unit 3)

- [ ] 3.1 Document `title` and `resource` in `FRONTMATTER-CORE.md` (est: ~30m)
  - why: `resource` applying to exactly one document type reads as inconsistency
    unless the scoping rule is stated. OKF §4.1 defines it as the URI of the asset
    the document *describes* — a reality doc describes a running service; a
    discovery brief describes nothing external, it *is* the artifact, and its `id:`
    already identifies it.
  - acceptance: R-3.1 — both fields documented with tier and consumer, `title`
    non-blocking except where the process contract already requires it, `resource`
    scoped to documents describing an external asset; R-3.3 — `resource` is stated
    producer-authored, never written by `graph build`, checked by no gate.
  - verify: the scoping rule is stated as a rule, not implied by examples.

- [ ] 3.2 Emit `title:` and `resource:` from the scaffolding templates (deps: 3.1, est: ~45m, mutex: cli)
  - why: If the shipped scaffolding doesn't emit a field the contract recommends,
    agents following it produce non-conformant documents by construction.
  - **Re-planned 2026-07-27 — the largest sizing change in the re-plan.** Planned
    against one emitter; there are now two per document type that must move
    together: the built-in string in `internal/scaffold/template.go` and its peer
    file under `company-os-starter/templates/`. The repo `CLAUDE.md` makes the sync
    a hard constraint — the section names the built-in emits are what `validate`
    greps for. Re-estimate ~75m.
  - acceptance: R-3.5 — `prd new`, `discover new`, and `reality new` emit `title:`,
    with `reality new` additionally emitting `resource:`; `internal/scaffold/template.go`
    and `templates/*.md` stay in sync per the repo `CLAUDE.md`.
  - verify: scaffold each document type into a scratch workspace and confirm the
    fields are present and the result passes `validate`; `make check` green, which
    is what catches a built-in string drifting from its template peer.

- [ ] 3.3 Emit `title:` from the `outcome.md` writer (deps: 3.1, est: ~25m, mutex: cli)
  - why: The fourth document-emitting path is easy to miss — `prd complete` writes
    `outcome.md` inline with no `title:` and no `id:`. Without this, 3.4 backfills
    the fixture's committed copy and the very next `prd complete` emits a
    non-conformant one. It is also what makes an `archive/prds/<id>/` index worth
    generating in the deferred change.
  - **Re-planned 2026-07-27 — partly satisfied already.** `outcomeDoc`
    (`internal/product/prd.go:569-574`, called from `:359-360`) already receives
    the PRD's title and renders it as the H1. What is missing is the frontmatter
    field, so this is an additive change to a writer that already holds the value —
    not the plumbing exercise the estimate assumed. `templates/outcome-review.md`
    must move with it (see re-plan note 3).
  - acceptance: R-3.6 — the writer emits `title:`, keeps `prd:` as the identity
    field per `CoreFieldErrors` (`internal/product/contract.go:65`), and no
    title-fallback path assumes `id:` exists.
  - verify: run `prd complete` on a scratch workspace and confirm the emitted
    `outcome.md` validates and carries a title.

- [ ] 3.4 Backfill `title:` across the two fixtures (deps: 3.2, 3.3, est: ~50m, mutex: fixtures)
  - why: **14 of 16** fixture documents carrying frontmatter `type:` lack `title:`
    (re-measured 2026-07-27; was "12 of 17" against the pre-port tree), so the
    node builder falls back to `id` or filename
    (`internal/graph/node.go:319-324`) and the generated context nodes read as a
    file listing. This is the change with a consumer today. Re-estimate ~60m.
  - acceptance: R-3.2 — every document in `examples/workspace/` and
    `examples/standalone-team/` that lacks a title has one.
  - verify: `grep -rL "^title:" examples/workspace examples/standalone-team
    --include='*.md'` returns only generated/reserved files.

- [ ] 3.5 Add `resource:` to reality documents (deps: 3.4, est: ~20m, mutex: fixtures)
  - why: Restores the field the founding proposal used
    (`00-original-proposal.md:170`) and gives every reality doc a
    machine-followable pointer instead of requiring filename-to-component
    inference.
  - acceptance: R-3.4 — every `type: component-reality` document in
    `examples/workspace/` and `examples/standalone-team/` carries
    `resource: component://<component-id>`; **no federated slice is edited** (I9).
  - verify: `git diff --stat examples/federated` is empty.

- [ ] 3.6 Regenerate and commit the `CLAUDE.md` generated blocks (deps: 3.4, 3.5, est: ~20m, mutex: fixtures)
  - why: Adding titles changes what `build_claude_node` renders, so the committed
    blocks go stale the moment 3.4 lands. The harness's double-build check
    (`s0 == s1 == s2`) requires the committed workspace to be already fully
    derived — this must be in the same commit, not a follow-up.
  - acceptance: R-3.7 — affected `CLAUDE.md` blocks regenerated via
    `company-os graph build` and committed alongside the title backfill.
  - verify: `bash examples/acceptance.sh` §4 reports
    "committed state fully derived + idempotent" for both monorepo fixtures.

- [ ] 3.7 Confirm the golden snapshots did not move (deps: 3.6, est: ~15m)
  - why: This is the phase most likely to produce an unexpected diff, and the
    prediction that it produces none is itself the test. Validate's stdout carries
    document paths and per-root status but never a title — if a golden moves here,
    something else changed and the reflex to re-baseline would bury it.
  - acceptance: R-3.8 — both goldens byte-identical; a diff is treated as a
    regression to diagnose, not re-baselined; R-3.9 — omitting `title` or
    `resource` produces no new error or warning beyond today's blocking checks.
  - verify: `git diff --exit-code examples/golden-validate.txt
    examples/federated-golden-validate.txt` is clean.

---

## Phase 4 — Cross-cutting verification (Units 5–6)

- [ ] 4.1 Backward-compatibility pass (deps: 0.1, 1.6, 2.4, 3.7, est: ~40m)
  - why: The CI adopter has no pain today and this unit exists so that stays true.
    Every requirement here is a veto over the rest of the change, so it is verified
    once at the end against the whole diff rather than assumed per phase.
  - acceptance: R-5.1 — a workspace passing before still passes, with no migration
    beyond the documented `graph build` re-run; R-5.2 — no new blocking check;
    R-5.3 — no new warn line; R-5.4 — unknown fields and types still preserved;
    R-5.5 — the tier model untouched; R-5.6 — both goldens byte-identical across
    the entire change and `--update` never run; R-5.7 — nothing written into a
    materialized slice.
  - verify: `make check` exits 0 (gofmt, vet, `go test ./...`, and the harness on
    all three fixtures); `git log -p` for the change contains no
    `acceptance.sh --update`; `git diff` against the merge base shows no change
    under `examples/federated/` except what 2.3 resolved, and none under
    `company-os-starter/internal` or `cmd` beyond the emitters 3.2 and 3.3 touch.

- [ ] 4.2 Decide the `examples/banking/` fixture question (deps: 1.3, est: ~30m)
  - why: Banking is the largest fixture (38 markdown files) and the acceptance
    harness does not exercise it at all. 1.3 reconciles its `profile: minimal`
    values by definition rather than by edit, so the question is whether anything
    else there needs touching — and editing untested content is how a silent
    regression enters.
  - acceptance: R-6.2 — if banking is edited beyond the enum reconciliation, it is
    first added to the harness's validate-exit-code fixture loop; R-6.3 — if that
    proves too large, banking is left untouched and `CONFORMANCE.md` records it as
    a prior-version fixture not yet backfilled.
  - verify: either `examples/acceptance.sh` §3 includes banking and passes, or
    `CONFORMANCE.md` carries the recorded exemption. One or the other, not neither.
