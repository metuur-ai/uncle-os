---
type: tasks
id: tasks-ux-simplification
title: UX Simplification — Tasks
status: draft
tags: [kind/tasks, status/draft]
---

# UX Simplification — Tasks

Source of the complaint: following the process — with or without an LLM agent —
requires holding too much in one's head. One ordinary change is ~9 commands
across 3 locations; validation requires knowing where an artifact lives before
you can ask for it; two re-derive commands (`governance resolve`, `graph build`)
must be remembered or `validate` fails at the worst moment; search is fragmented
across tags, `ids list`, per-directory indexes, CLAUDE.md nodes and
feature-indexes; and the docs describe capabilities the CLI does not implement.

**Design intent of this plan:** the complexity is not the standard — it is that
the standard currently requires memorizing too much to comply with it. Every
unit below is therefore **additive**: new commands and inference, never new
enforcement. Nothing a user does today stops working; the CLI just stops
requiring the user to already know what they are asking about.

## Red lines — what must NOT change

1. **The OKF gate contract.** `validate` gates `[1/8]`–`[8/8]` (and `[1/9]`–
   `[9/9]` federated) keep their exact semantics and output on the default
   path. Exit
   codes unchanged. New behavior is opt-in (`--fix`) or lives in new commands.
2. **Flexibility.** Strict on artifacts, flexible on process. No unit may add a
   check on *how* an artifact was produced. Units 1–4 add capability; they never
   gate.
3. **Local search.** `graph build`, the ids registry, per-directory `index.md`,
   CLAUDE.md context nodes and feature-indexes stay exactly as they are. Unit 4
   puts a single front door in front of them; it removes nothing.
4. **The house conventions.** Commands return `[]model.GateResult` records and
   errors; only `cmd/` and `internal/render/` write output — nothing below
   `cmd/` calls `os.Exit` or prints. Every mutating command prints the next
   command in the workflow (guidance chain). The `frontmatter()` parser contract
   is untouched.

## Global acceptance (must hold after every unit)

- `make check` exits 0 (gofmt + `go vet` + `go test ./...` + `acceptance.sh`).
- All five committed goldens byte-identical: `examples/golden-validate.txt`,
  `examples/federated-golden-validate.txt`,
  `examples/failing-workspace-golden-validate.txt`,
  `examples/failing-federated-golden-validate.txt`,
  `examples/failing-federated-nolock-golden-validate.txt`. A golden diff is a
  regression to diagnose, not a baseline to refresh. No unit runs
  `acceptance.sh --update`.
- `company-os validate` still exits 0 on `examples/workspace`.
- Existing flags keep working: every inference added in Unit 2 has an explicit
  flag that overrides it.

## Ordering rule

Units 1–4 all touch the dispatch surface (`cmd/company-os/commands.go`,
`args.go`, help text), so they land one at a time in numeric order. Unit 5
touches fixtures and scaffold only after the new commands exist, so the
minimal on-ramp fixture can demonstrate them. Unit 6 is docs-only and lands
last so it can reference what now exists. Each unit lands as its own commit
with the task box checked in the same commit.

---

## Phase 1 — Collapse the command chain (Units 1–2)

- [x] 1.1 `company-os next` — one command that names the single next action (est: ~3h)
  **LANDED 2026-08-26** — read-only `next` subcommand with priority-ranked scan (expiry → contract → done-check → outcome → empty), `--all` grouping, text + `--json` renderers; soon-due window is 14 days.
  - why: the loop is 9 memorized commands; the guidance chain already prints
    the next step *after* each command, but from a cold start — a fresh clone,
    a new teammate, an agent handed the repo — nothing answers "what do I do
    now?". `next` is the guidance chain run in reverse: stateless, scans the
    workspace, prints exactly one action and the exact command that performs it.
  - behavior: scan and report the highest-priority pending action:
    1. an expired or soon-due deviation `reviewDate` / exception `expires`
       (gate-2 data, already parsed by `internal/governance`);
    2. an active PRD failing its artifact contract (missing sections/fields —
       reuse `internal/product` checks, do not re-implement);
    3. an active PRD whose governance checklist has unchecked items and/or
       whose component reality doc is stale → "update reality + check items,
       then `prd complete`";
    4. a completed PRD with an outcome review due (`outcome.md` frontmatter);
    5. nothing pending → "no pending actions; start with
       `company-os discover new …`".
    One action by default; `--all` lists every pending item grouped by kind.
    Read-only: `next` mutates nothing and therefore prints no guidance chain of
    its own beyond the command it recommends.
  - acceptance: registered in `commands.go`, rendered through `internal/render`
    (text + `--json`), no output changes to any existing command, gates
    untouched. `next` on `examples/workspace` (post-tutorial state) surfaces
    the scheduled outcome review. `next` on a workspace with an active PRD
    surfaces the PRD's blocking item, not the outcome review.
  - verify: `go test ./internal/...` with unit tests covering priority order
    and the empty case; `company-os next --root examples/workspace` exits 0;
    `make check` green with all goldens byte-identical.

- [x] 1.2 Context inference — validate/complete without `--platform`/`--team` (est: ~2h)
  **LANDED 2026-08-26** — `prd validate`/`prd complete` infer `--platform` from `platforms/*/change-records/active/` and `platforms/*/archive/prds/`; `discover validate` infers `--team` from `teams/*/product/discovery/`; unique match proceeds, ambiguity lists every candidate, absence reports not-found; explicit flags always win.
  - why: `discover validate` requires `--team` and `prd validate`/`prd complete`
    require `--platform`: the user must already know where the artifact lives —
    i.e. its lifecycle stage — before the tool will check it. The descriptor
    and directory layout already contain that information.
  - behavior: when `--platform` is omitted, search
    `platforms/*/change-records/active/` and `platforms/*/archive/prds/` for
    the given id; exactly one match → proceed as if the flag had been passed;
    several matches → a usage error listing every candidate with its platform;
    none → the existing "not found" error. Same shape for `discover validate`
    across `teams/*/product/discovery/`. Explicit flags always win. Help text
    marks both flags "required unless the id is unique across the workspace".
  - acceptance: `company-os prd validate <id>` and `company-os discover
    validate <id>` work flag-free on the example fixtures; every existing
    flag-carrying invocation behaves byte-identically; ambiguity and not-found
    paths are tested.
  - verify: unit tests for unique/ambiguous/absent inferences; re-run the
    tutorial command sequence both with and without flags; `make check` green,
    goldens byte-identical.

## Phase 2 — Kill the "remember to re-derive" tax (Unit 3)

- [x] 2.1 `company-os validate --fix` — regenerate derived state before gating (est: ~3h)
  **LANDED 2026-08-26** — `--fix` regenerates effective-governance, tags, indexes, and CLAUDE.md nodes through the same `governance.Resolve` + `graph.Rebuild` code paths before gating; summary line `validate --fix: N file(s) regenerated`; `--json` carries `fixRegenerated`; creation-path trailing-newline fix makes the derivation 1-cycle idempotent.
  - why: after a deviation declaration or a frontmatter edit the user must
    remember `governance resolve` and `graph build`, or gates 1/4–6 fail at
    CI time. The CLI itself prints "re-run: …" — it knows what is missing but
    will not do it. `--fix` makes the derived-state commands self-healing
    locally while CI keeps the strict diff check.
  - behavior: with `--fix`, before running gates, regenerate exactly the
    artifacts the acceptance double-build already proves are derived:
    `teams/*/generated/effective-governance.yaml`, frontmatter `tags:`,
    CLAUDE.md context nodes, per-directory `index.md`, and platform
    feature-indexes — reusing the existing `rebuild_generated`/derive paths,
    never new writers. Then run the normal gates against the refreshed tree.
    Output: the usual gate lines, plus one summary line
    `validate --fix: N file(s) regenerated` (0 when already clean). Default
    `validate` (no flag) is byte-identical to today.
  - acceptance: hand-drift a tag and an index in a scratch copy of
    `examples/workspace`, run `validate --fix` → exit 0, drift gone, second run
    reports 0 regenerated. Without `--fix` the same drift still fails the same
    gates it fails today. No golden changes.
  - verify: unit tests exercising the fix path against fixture copies;
    `make check` green; both committed goldens byte-identical.

## Phase 3 — One front door for local search (Unit 4)

- [ ] 3.1 `company-os find <query>` — unified local search (est: ~3h)
  - why: local search is currently five mechanisms: derived tags (Obsidian),
    `ids list`, per-directory `index.md`, CLAUDE.md context nodes, and
    feature-indexes. A human without Obsidian and an agent without prior graph
    context have no single "everything about X". The repo owner's own note in
    TUTORIAL.md asks to integrate graphify for exactly this.
  - behavior: `company-os find <query>` (case-insensitive substring; exact-id
    match ranked first) fans out over: canonical ids in `ids/registry.yaml`,
    derived tags, frontmatter `title:`/`id:`, per-directory index entries, and
    feature-index component maps. Output grouped by match kind with path,
    title (when present) and why it matched. Deterministic ordering (path).
    Graphify hook: if a `graphify` binary is on PATH **and**
    `graphify-out/graph.json` exists under the workspace root, append a
    `graphify query "<query>"` section with the tool's output; if either is
    missing, print one quiet hint line (`graphify not detected — install it for
    graph search`) and succeed anyway. `--no-graphify` skips the hook.
    Read-only, no gates affected.
  - acceptance: `find customer-notification-service` returns the descriptor,
    the reality doc, PRDs (active or archived), the feature-index entry and the
    registry id — every hit carrying its match reason; a nonsense query exits 0
    with "no matches" (search is not a gate); `--json` emits the same records.
  - verify: fixture-based unit tests per source (ids/tags/title/index/feature)
    and a test that the graphify path is skipped cleanly when the binary is
    absent; `make check` green, goldens byte-identical.

## Phase 4 — A real one-team on-ramp (Unit 5)

- [ ] 4.1 `examples/standalone-team` becomes a working minimal workspace (est: ~2h)
  - why: the small-team on-ramp is nominal — `standalone-team` today holds an
    onboarding README, not a workspace, while `init` always scaffolds the full
    four-root federation. Every adopter pays federation cognitive cost on day
    one. The minimal loop (one team, one platform, one component, discovery →
    PRD → complete) must be demonstrable as a real, validating workspace.
  - behavior: rebuild `examples/standalone-team` as a complete workspace —
    company baseline, one platform with requirements + one component
    descriptor, one team with ownership + generated artifacts — such that
    `company-os validate --root examples/standalone-team` exits 0 and
    `company-os next`/`today` work there. Keep the existing onboarding doc
    inside it. A short README in the fixture states the progressive-disclosure
    rule: nothing in this workspace requires knowing what a federation is;
    those concepts appear only when a second team or platform is added.
    Acceptance's existing double-build for this fixture must stay green, so the
    fixture is committed fully derived (tags, indexes, CLAUDE.md nodes,
    effective-governance all generated, not hand-written).
  - acceptance: `validate` exits 0 there; `make check` green (including the
    existing standalone-team double-build step); no change to any other fixture
    or golden.
  - verify: run every Phase-1–3 command against the new fixture (`next`,
    flag-free `prd validate` on nothing active → clean empty behavior, `find`
    for its component); `make check` green.

## Phase 5 — Docs face reality (Unit 6)

- [ ] 5.1 Cut or clearly mark everything the CLI does not implement (est: ~1h)
  - why: readers currently carry the spec-vs-reality gap in their heads —
    TUTORIAL §0.5 teaches six path-resolution layers of which the CLI
    implements three, and ONTOLOGY-GUIDE describes `validate --ontology` and
    `spec trace` which do not exist. Docs must describe what exists; the
    roadmap belongs in one marked place, not scattered through the walkthrough.
  - behavior: in both copies of TUTORIAL.md (repo root and
    `company-os-starter/docs/`): collapse §0.5 to the implemented layers
    (flag → env var → cwd) with a one-line pointer to the proposal for the
    rest; replace the trailing handwritten note ("I want to integrate the
    graphify…") with a pointer to `company-os find`. In
    `company-os-starter/docs/ONTOLOGY-GUIDE.md`: mark `validate --ontology` and
    `spec trace` explicitly as roadmap (they are named as such in CLAUDE.md
    already — make the guide agree). No description of implemented behavior is
    removed.
  - acceptance: the two TUTORIAL copies stay in sync with each other; every
    command named in the edited docs runs as documented; `make check` green.
  - verify: `grep` the edited docs for layer-3–5 and `--ontology`/`spec trace`
    mentions and confirm each is either gone or explicitly labelled roadmap;
    re-run every command quoted in the edited sections.

---

## Execution notes

- Units execute sequentially through sub-agents, one unit per agent, in numeric
  order; each agent's exit condition is its own acceptance list **plus** the
  global acceptance. Shared dispatch files are why units do not run in
  parallel.
- Each unit lands as one commit: implementation + tests + the checked box in
  this file, together.
- When a unit discovers a requirement that contradicts shipped behavior, it
  records an amendment here rather than silently adapting — the same rule the
  OKF conformance plan applies to itself.
