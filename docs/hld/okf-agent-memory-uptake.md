---
type: hld
id: hld-okf-agent-memory-uptake
title: OKF Agent-Memory Uptake — High-Level Design
status: draft
---

# OKF Agent-Memory Uptake — High-Level Design

## Overview

Takes seven items from `.devlocal/research/2026-09-07-okf-agent-memory-vs-uncle-os.md`
§7 (T1–T7) and makes them buildable. The source project, `okf-agent-memory`, is a
small Go CLI plus a behavioural convention that turns an OKF v0.2 bundle into an
agent's long-term memory. uncle-os already shares its format layer (`description:`,
per-directory `index.md`, `generated:`/`verified:`, three trust tiers — all landed by
`docs/ears/okf-provenance-and-indexes.md` on 2026-08-26). What it does **not** share
is the behavioural layer: the fields are parsed but never checked, mutations leave
no trail, a fresh workspace tells an agent nothing about how to behave, and no test
proves an agent can get from `init` to `prd complete` using only the repository.

This change closes that layer. It adds **no blocking gate** and **no new gate
ordinal**. Every new signal is warn-tier, or opt-in fail behind a flag, in the same
way `validate --fix` is opt-in repair.

T8 (an MCP shim) is **not** in scope. The research lists it for visibility; the
standing decision at `company-os-starter/docs/user-guide/explanation/github-mcp-and-automation.md:5`
("Company OS ships no MCP server and no MCP client") holds. The user's note on §9 Q4
confirms the only MCP in play is the GitHub MCP for remote repository operations,
which is outside this CLI.

**Relationship to prior changes.** `okf-provenance-and-indexes` (completed) shipped
the fields this change validates; its `FRONTMATTER-CORE.md` rule "no gate consumes
the *tier*" is preserved — a shape/ordering check on `generated:`/`verified:` does
not read the tier. `okf-v02-conformance` (locked, 1/20 tasks) still owns task 1.3
(version scheme + `profile:` enum); Unit 7 here **completes** that task rather than
re-specifying it, and the parent's tasks file is updated to point here.

## Stakeholders & Impact

Two audiences, weighted equally (elicited 2026-09-07).

**Coding agents continuing a workspace from a fresh clone.** Today an agent that
opens a workspace scaffolded by `company-os init` finds four roots and no file
saying: search with `company-os find` before drafting; never hand-edit anything
under `generated/`; run `company-os validate` before claiming done; `prd complete`
refuses until `reality/` is updated. Those rules exist, spread across five skills the
agent must first know to install. After this ships, `init` writes an `AGENTS.md`
carrying a short Minimal Agent Contract and a `CLAUDE.md` that imports it, so the
first file any agent runtime reads states the rules. An executable scenario suite
(`AGENT_TESTING.md`) proves the documented skill chain works from `init` alone.

**Workspace maintainers running `validate`.** Today a document can carry
`verified: [{by: human:x, at: 2020-01-01}]` and `generated: {at: 2026-…}` and pass;
a `component-reality` doc has no freshness signal after its PRD completes;
`deviation declare`, `exception request`, `add …`, `reality new`, `discover new`,
`prd new` and `workspace sync` mutate the tree and write nothing to `log.md`;
`companyOsVersion: "2026.2"` is committed by nine fixtures against a version defined
nowhere, and `profile:` has three contradictory definitions (parent EARS,
`docs/01-flexibility…md:121`, the scaffolder). After this ships, each of those surfaces a `[warn]` line (or a fail under
`--stale`), every mutating command leaves one dated line in the nearest `log.md`,
and the version and profile values are defined and checked.

**Secondary consumers.** `examples/acceptance.sh` goldens (must not move except
where named in the LLD); the five canonical skills (unchanged; the contract file
points at them); `docs/TUTORIAL.md` (gains one paragraph on the contract file and
the log trail); the future `company-os-plugin` HLD (the contract text is the natural
plugin-side `CLAUDE.md` seed — noted, not built).

## Goals

- **G1** `validate` gate 2 warns on ill-formed or inconsistent provenance:
  `generated:` without `by`; any `verified[i].at` earlier than `generated.at`;
  actor strings outside the `human:`/`process:`/`<producer>/<version>` convention.
- **G2** A `stale_after: YYYY-MM-DD` frontmatter field exists, is documented in
  `FRONTMATTER-CORE.md` Tier 2, warns in gate 2 once `<= today`, and fails under a
  new `validate --stale` flag. `prd complete` writes it into `outcome.md` so the
  field has a real producer on day one — a deliberate extension of research T2,
  which defined the field but named no producer.
- **G3** Every mutating command appends one `- YYYY-MM-DD: …` line to the nearest
  `log.md`; gate 4 warns on any `log.md` line that does not match that shape.
- **G4** `init` and `skills install` write a workspace-root `AGENTS.md` (Minimal
  Agent Contract) and `CLAUDE.md` (`@AGENTS.md` import only). If either already
  exists, the contract is appended inside a delimited block, never overwriting
  hand-owned text; a present marker leaves the file untouched. `skills install` is
  how every pre-existing workspace gets the contract.
- **G5** Agent-behaviour scenario tests exist and `docs/AGENT_TESTING.md` states each
  scenario's pass criterion; the suite runs under `go test ./...` and therefore
  `make check`.
- **G6** `docs/SECURITY.md` states what must never enter `reality/`,
  `product/discovery/`, `scratchpad/`, and how `knowledge/` integrity (gate 9) plus
  `log.md` + git history serve as the poisoning defence and audit trail.
- **G7** `companyOsVersion` has a defined scheme and `profile:` a defined enum in
  `CONFORMANCE.md §6`; gate 4 warns when a `platform.yaml` declares a value outside
  them. The enum is the parent's (`minimal | standard | strict`) and all three
  places that state it are reconciled. This closes `okf-v02-conformance` task 1.3.

## Non-Goals

- **N1** No MCP server or client (T8). Decision stands.
- **N2** `sources[]`, credibility signals, attested computation — still no consumer
  (parent N2/N3). Unchanged.
- **N3** Overwriting `generated:` on update (OKF `mutate.go:134-138`). uncle-os
  writes it once; that is the better behaviour and is kept.
- **N4** Agent-maintained index bullets (`UpdateParentIndex`). Indexes stay derived
  and drift-gated.
- **N5** Domain-neutral `type` vocabulary; "type is the only required field".
  Declined twice already.
- **N6** Changing `log.md` to OKF's `## YYYY-MM-DD` heading shape. §9 Q2 answer:
  keep `- date:` bullets; the OKF shape is not needed to satisfy "append-only".
- **N7** Parsing `log.md` into the graph. It stays in `skipNames`; the gate 4 check
  reads it as text.
- **N8** Auto-deleting or archiving stale documents. `stale_after` is a signal only
  (OKF convention §14 agrees).
- **N9** Running the scenario suite against a real LLM. Scenarios drive the CLI
  exactly as the skills tell an agent to; no model in the loop.
- **N10** Making any new check blocking by default. The only fail path is
  `--stale`, opt-in.
- **N11** `stale_after`/`--stale` influencing `check ready|done` or `prd complete`.
  A stale reality doc never blocks completion; the done-check's own `updated:`
  rule does that.
- **N12** Versioning or comparing the contract text (unlike skills). A present
  marker means "leave it"; the text changes only through a re-init of a fresh
  workspace or a hand edit.
- **N13** A clock seam. Tests use far-dated values as the existing expiry fixtures
  do.

## Success Criteria

- `company-os --root examples/workspace validate` output is **byte-identical** to
  today's `examples/golden-validate.txt` except for lines the LLD names (none are
  expected: fixtures carry zero `generated:`/`verified:`/`stale_after:`, four
  conformant `log.md` files, and `companyOsVersion: "2026.2"` with `profile`
  `standard` or `minimal` in every `platform.yaml`). The same holds for the other
  four goldens.
- A fixture with `verified[0].at < generated.at` produces exactly one `[warn]` in
  gate 2 and `validate` still exits 0.
- A fixture with `stale_after: 2000-01-01` exits 0 without `--stale` and non-zero
  with it, the difference being that one finding's severity.
- After `init`, the target contains `AGENTS.md` and `CLAUDE.md`; `validate` passes;
  a second `graph build` changes no bytes (acceptance step 4 idempotency holds).
- Running the full chain `init → add component → discover new → discover validate
  → prd new --from-discovery → prd validate → prd complete (refused: reality
  stale) → reality new → prd complete → validate` from a temp dir, driven by the
  printed `next:` and `fix:` guidance where the CLI prints one and by the skill
  text where it does not, ends in `PASS` and leaves one `log.md` line per
  mutating step. This is TC-S2 in `AGENT_TESTING.md`.
- `make check` green; `acceptance.sh --update` **not** used.

## Invariants

- **I1 Frozen gate numbering.** `[1/8]…[8/8]` (+ `[9]` federated) keep their
  ordinals, slugs and titles. New checks live inside gates 2 and 4.
- **I2 No new default-blocking behaviour.** Every new finding is `SevWarn`. Only
  `--stale` promotes one class of finding to `SevFail`.
- **I3 Generated content is never hand-edited; hand-owned content is never
  generated over.** The contract block in `AGENTS.md`/`CLAUDE.md` sits inside its
  own delimiters distinct from `company-os:generated:*`; `graph build` does not
  touch it; `init` appends rather than overwrites.
- **I4 `log.md` is append-only.** No command rewrites, reorders or deletes lines.
  `O_APPEND|O_CREATE` only, as `appendPromotionLog` already does.
- **I5 Zero new runtime dependencies.** Standard library only; the binary stays
  static.
- **I6 "No gate consumes the trust tier."** `TrustTier()` remains advisory. Gate 2
  reads `generated`/`verified` shapes and dates, never the derived tier.
- **I7 The frontmatter parser contract and the guidance chain are unchanged.**
  `frontmatter.go:35` fence regex untouched; every mutating command still prints
  its next command — the log append is added beside that print, not instead of it.
- **I8 Goldens move only deliberately.** Any diff in `examples/*golden-validate.txt`
  must be traceable to a line in the LLD; `--update` is never used to clear a
  red build.
