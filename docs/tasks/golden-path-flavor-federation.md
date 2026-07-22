---
type: tasks
id: tasks-golden-path-flavor-federation
title: Golden Path, Flavor & Federation — Tasks
status: draft
tags: [kind/tasks, status/draft]
---

# Golden Path, Flavor & Federation — Tasks

Source of truth: `docs/ears/golden-path-flavor-federation.md` (Units 1–8, 37 statements).
Architecture constraints: `docs/lld/golden-path-flavor-federation.md`.
Target: single-file CLI `company-os-starter/bin/company-os`.

**Global acceptance (must hold after every phase):** `company-os validate`
exits 0 on `examples/workspace` **and** `examples/standalone-team`, the
double-build no-op check stays clean, **and** — with no `workspace.yaml`
present — monorepo behavior is byte-for-byte identical to the frozen golden
snapshot (federation is invisible until opted into).

**Deferred — NOT planned here (LLD Out of Scope):** Option C full-repo
federation / git submodules; localized required-headings aliases in the
validate contract (localized templates keep English headings); template row
seeds for `deviations.yaml`/`exceptions.yaml`; any Web/GUI wizard.

**Ordering rule:** Phase 0 (Unit 8 + shared constants) lands before any
feature — every later task assumes the single `frontmatter()` contract,
fail-fast root resolution, and `*_SECTIONS` single-sourcing. Phase 1
(Units 1–2) is the golden-path MVP. Phase 2 (Units 3–4) adds flavor
(templates + role views), Phase 3 (Unit 5) skills layering, Phase 4
(Units 6–7) federation. Tasks reference GPF-R ids, not literal `[n/m]`
gate strings.

---

## Phase 0 — Contract hardening & single-sourcing (Unit 8, prereqs)

- [x] 0.1 Single `frontmatter()` contract sweep (est: ~40m)
  - why: Every later feature (template resolution, skill discovery, slice
    validation) parses frontmatter; two parsers means two behaviors and
    unreproducible bugs. Consolidate before growing the surface.
  - acceptance: GPF-R-8.2 — all artifact types parse YAML frontmatter through
    the one `frontmatter()` contract (`^---\n…\n---\n`), identically.
  - verify: grep the CLI for ad-hoc `---` splitting outside `frontmatter()`
    — none remain; existing validate output vs golden snapshot unchanged.

- [x] 0.2 Fail-fast workspace-root resolution (deps: none, est: ~30m)
  - why: Commands that silently succeed outside a workspace produce artifacts
    in the wrong place — the worst first-run experience possible, and the
    exact spot federation detection (4.1) will later hook into.
  - acceptance: GPF-R-8.1 — workspace-scoped commands outside a root fail
    fast, naming the resolution order (`--root` → `$COMPANY_OS_WORKSPACE_ROOT`
    → cwd) instead of succeeding silently.
  - verify: run a workspace command from `/tmp` — non-zero exit, message names
    all three resolution steps; inside a workspace, behavior unchanged.

- [x] 0.3 Extract `*_SECTIONS` constants shared by templates + validate (deps: 0.1, est: ~45m)
  - why: The LLD makes heading drift impossible by construction: required
    headings live once (e.g. `DISCOVERY_SECTIONS`, `PRD_SECTIONS`) and feed
    **both** the built-in `*_TEMPLATE` strings and the validate greps. This is
    the enabler for custom templates (Phase 2) and must be behavior-preserving.
  - acceptance: supports GPF-R-4.3 — default scaffolds stay byte-identical;
    validate keeps grepping the same headings, now sourced from the constants.
  - verify: scaffold each artifact type before/after the refactor — byte-equal
    output; validate golden snapshot unchanged.

- [x] 0.4 Golden-path characterization baseline (deps: 0.1–0.3, est: ~20m)
  - why: Reuse the federation-enrichment harness (golden validate snapshot +
    double-build no-op) as the oracle for "monorepo untouched" that every
    federation task must hold against.
  - acceptance: committed script run passes on the untouched repo after 0.1–0.3.
  - verify: snapshot matches, no-op diff clean, exit 0.

---

## Phase 1 — Golden path MVP: growth commands + guidance chain (Units 1–2)

- [ ] 1.1 `company-os add platform|team|component` (deps: 0.2, 0.3, est: ~75m)
  - why: Closes friction F2 — the second team/platform/component today means
    hand-copying YAML. Growth must reuse the exact scaffolding templates
    `init` uses so there is one shape, not two.
  - acceptance: GPF-R-1.x (growth) — `add` scaffolds from the same templates
    as `init`, refuses to overwrite anything existing, and prints the next
    command in the chain; non-interactive flags (`--company/--team/--platform`)
    mirror the wizard.
  - verify: in a fresh workspace, `add team`/`add platform`/`add component`
    each yield validate-green artifacts; re-running the same `add` fails
    without mutating; piped/non-TTY invocation works via flags.

- [ ] 1.2 `company-os reality new` + `templates/reality-component.md` (deps: 0.3, est: ~45m)
  - why: Closes friction F3 — the guidance chain currently dead-ends when a
    target component has no reality doc; the user is left to invent structure.
  - acceptance: GPF-R-2.1/2.2 — `reality new --platform <p> <component-id>`
    scaffolds `reality/components/<id>.md` from `templates/reality-component.md`;
    output validates green.
  - verify: run against a component without a reality doc → file created,
    validate exits 0; against an existing doc → refuses, no mutation.

- [ ] 1.3 Guidance-chain extension through the reality step (deps: 1.1, 1.2, est: ~30m)
  - why: The chain is the product: "every mutating command prints the next
    command." `prd new` must point at `reality new` when the reality doc is
    missing, and `prd complete` must include the reality-update step —
    otherwise the new commands are undiscoverable.
  - acceptance: GPF-R-1.x/2.3 — `prd new` prints the exact `reality new`
    invocation when the target component's reality doc is missing;
    `prd complete` guidance includes the reality-update step; chain unbroken
    init → discovery → prd → reality → complete.
  - verify: walk the full chain in a fresh workspace copying only printed
    next-commands — no step requires out-of-band knowledge.

- [ ] 1.4 Init hardening: refuse re-init, terminal-only wizard parity (deps: 0.2, est: ~25m)
  - why: `init` inside an existing workspace must not clobber; and every
    wizard answer needs a flag equivalent or CI/scripting can't reproduce it.
  - acceptance: GPF-R-1.x — `init` in an existing workspace refuses and exits
    non-zero without mutating; every interactive prompt has a documented
    non-interactive flag.
  - verify: `init` twice in the same dir → second run exits non-zero, `git
    status` clean; scripted init with flags only produces a validate-green
    workspace with no TTY attached.

- [ ] 1.5 Golden-path document (deps: 1.1–1.4, est: ~40m)
  - why: One doc, extending `docs/GETTING-STARTED.md`, that matches the CLI's
    printed chain exactly — including environment prerequisites (Python,
    pyyaml, PATH), which is where real first runs actually fail.
  - acceptance: GPF-R-1.x — doc covers prerequisites → setup → discovery →
    PRD → reality update → completion; every command shown is the one the CLI
    prints.
  - verify: execute the doc top-to-bottom in a clean checkout — no step
    deviates from the CLI's own guidance output.

---

## Phase 2 — Flavor: custom templates + role views (Units 3–4)

- [ ] 2.1 Template resolution, first-found-wins (deps: 0.3, est: ~60m)
  - why: Teams flavor their process without forking the CLI. Resolution order
    `teams/<t>/templates/` → `platforms/<p>/templates/` → `company-os/templates/`
    → built-in makes override locality predictable.
  - acceptance: GPF-R-4.1/4.2 — scaffolding commands resolve templates
    first-found-wins through the four layers; the resolved template source is
    recorded (provenance), and template files themselves are not validated —
    their *outputs* are (strict on artifacts, flexible on process).
  - verify: place a team-level override → scaffold uses it and records the
    source; remove it → falls through to platform/company/built-in in order.

- [ ] 2.2 Custom-template outputs still face the validate contract (deps: 2.1, 0.3, est: ~30m)
  - why: Flexibility must not erode the contract: a custom template that drops
    a required heading should fail loudly at the artifact, naming the heading.
  - acceptance: GPF-R-4.3/4.4 — default templates scaffold byte-identically to
    today; an artifact from a heading-dropping custom template fails validate
    with the missing heading named.
  - verify: craft a template missing a `*_SECTIONS` heading → scaffold, then
    validate FAILs naming exactly that heading; default-template scaffolds
    byte-match Phase 0 baselines.

- [ ] 2.3 Role views: display-only terminology translation (deps: 0.1, est: ~45m)
  - why: Different audiences read the same graph; translation must never touch
    canonical terms on disk or the single vocabulary fractures.
  - acceptance: GPF-R-3.1–3.3 — translation is display-only; canonical terms
    in artifacts are unchanged; a canonical term with no translation entry for
    the active role displays unchanged (no error).
  - verify: render a role view with a partial translation map — translated
    terms swap, unmapped terms pass through, and `git diff` shows zero file
    changes after rendering.

---

## Phase 3 — Custom skills layering (Unit 5)

- [ ] 3.1 Four-layer skill discovery + merged view (deps: 0.1, est: ~60m)
  - why: Teams and individuals add process guidance additively; the merge
    order is the whole model — company, platform, team, personal.
  - acceptance: GPF-R-5.1 — skills discovered and merged across
    `company-os/skills/` → `platforms/<p>/skills/` → `teams/<t>/skills/` →
    personal; GPF-R-5.4 — canonical (mandatory) steps ordered above personal
    ones in the merged view.
  - verify: fixture with one skill per layer — merged view lists all four,
    canonical steps first.

- [ ] 3.2 Shadowing is a validate error (deps: 3.1, est: ~30m)
  - why: Conflicts must be impossible by construction, not adjudicated — a
    lower layer silently replacing a canonical skill is the failure mode that
    destroys trust in the merged view.
  - acceptance: GPF-R-5.2 — a team/personal skill reusing a canonical skill's
    ID or name fails validate identifying **both** files.
  - verify: duplicate a canonical skill id at team layer → validate FAIL names
    both paths; rename it → green.

- [ ] 3.3 `extends: platform-skill://…` layering + scheme registration (deps: 3.1, est: ~40m)
  - why: The sanctioned alternative to shadowing: extend, don't replace. The
    scheme must be a first-class canonical ID scheme or references can't be
    checked.
  - acceptance: GPF-R-5.3 — a skill with `extends: platform-skill://…` renders
    layered (base steps plus extension steps); `platform-skill://` documented
    as a canonical ID scheme in the ontology conventions alongside
    `component://`, `capability://`, `req://`, `context://`.
  - verify: extend a base skill in a fixture → merged render shows base then
    extension; a dangling `extends:` target fails validate naming the URI.

---

## Phase 4 — Federation Option B: manifest + sync + lock (Units 6–7)

- [ ] 4.1 `workspace.yaml` detection in root resolution (deps: 0.2, est: ~30m)
  - why: Activation must ride the existing root-resolution touch point —
    presence of the manifest is the *only* switch; absence must leave monorepo
    mode byte-for-byte untouched.
  - acceptance: GPF-R-6.x — manifest present ⇒ federated mode; absent ⇒
    monorepo mode identical to the golden snapshot; detection lives in the
    `--root` → `$COMPANY_OS_WORKSPACE_ROOT` → cwd resolution (touch point 1 of 2).
  - verify: run the 0.4 characterization script with no manifest — snapshot
    identical; add an empty-repos manifest — federated code path activates.

- [ ] 4.2 `company-os workspace sync`: sparse fetch + read-only slices + lock (deps: 4.1, est: ~90m)
  - why: The crown jewel of Option B — one governance view over many repos
    without submodules, fetching only governance-relevant paths.
  - acceptance: GPF-R-7.1/7.2 — sync fetches only `governance/`, `components/`,
    `requirements`, `reality/`, `skills/`, `templates/` via sparse/filtered
    shallow git (`--filter`, sparse-checkout, `--depth`); materializes
    read-only slices in the workspace layout; writes resolved repo SHAs and
    per-slice content hashes to `workspace.lock.yaml`.
  - verify: two-repo fixture — sync pulls only listed paths, slices are
    read-only, lock records SHAs + hashes; re-sync with no upstream change is
    a no-op diff.

- [ ] 4.3 `--frozen` CI mode (deps: 4.2, est: ~40m)
  - why: CI must be reproducible and offline; the lock is the contract.
  - acceptance: GPF-R-7.x — `--frozen` performs no network access,
    materializes strictly from `workspace.lock.yaml`, and fails if the lock is
    missing or doesn't cover the manifest.
  - verify: run `--frozen` with network access stubbed out/unavailable →
    succeeds from lock; delete a lock entry → fails naming the uncovered repo.

- [ ] 4.4 Hand-edit detection on slices (deps: 4.2, est: ~35m)
  - why: Slices are derived content, like `generated/` — a hand-edit is drift
    that must be caught, not silently synced over or shipped.
  - acceptance: GPF-R-7.x — a slice file whose content hash differs from
    `workspace.lock.yaml` fails validate naming the edited path with a
    re-sync remedy.
  - verify: hand-edit a slice file → validate FAIL names the path; re-sync →
    green.

- [ ] 4.5 Slice-aware path resolution (touch point 2 of 2) (deps: 4.2, est: ~45m)
  - why: Aggregates and validate must see federated slices exactly as they see
    local docs — one resolution seam keeps the rest of the CLI
    federation-ignorant.
  - acceptance: GPF-R-6.x/7.x — doc iteration resolves slice paths uniformly
    with local paths in federated mode; in monorepo mode the code path is
    unreachable (snapshot unchanged).
  - verify: federated fixture — validate and aggregates include slice docs;
    monorepo snapshot from 0.4 still byte-identical.

- [ ] 4.6 Federation runbook (deps: 4.2–4.5, est: ~30m)
  - why: Operational failure modes (pin bump, desync, removing a repo) are
    where federation dies in practice; document the recovery paths the CLI's
    error messages point at.
  - acceptance: GPF-R-7.x — runbook covers pin bump, lock/manifest desync
    recovery, and removing a federated repo; every remedy string printed by
    sync/validate references a runbook section that exists.
  - verify: grep CLI remedy strings → each named section present in the runbook.

---

## Phase 5 — Acceptance & backward compatibility (throughout)

- [ ] 5.1 Fixtures per unit + committed exemplars (deps: 1.3, 2.2, 3.3, 4.5, est: ~60m)
  - why: The only acceptance path is a green validate on fixtures that
    exercise every unit — golden-path e2e, custom-template team, layered
    skills, and a two-repo federated workspace.
  - acceptance: validate exits 0 on `examples/workspace`,
    `examples/standalone-team`, and the new federated fixture on fresh
    checkout, no `sync`/`graph build` needed for the committed state.
  - verify: fresh clone → `validate` exits 0 on all fixtures.

- [ ] 5.2 Freeze existing gates + CLI conventions (deps: all phases, est: ~30m)
  - why: New surface must not perturb existing gates, the `die/ok/warn/fail`
    helpers, or the guidance chain adopters depend on.
  - acceptance: existing gate output identical (modulo gate-count headers) vs
    the golden snapshot; `frontmatter()` (0.1) and next-command guidance
    conventions intact; docs mentioning command counts/flows corrected in this
    change.
  - verify: diff validate output against the 0.4 snapshot — only new-gate
    lines added; grep docs for stale command flows returns nothing.
