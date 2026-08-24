# Team-Drafted PRDs and Promotion — Tasks

Spec set: `docs/hld/team-drafted-prd-promotion.md`, `docs/lld/team-drafted-prd-promotion.md`,
`docs/ears/team-drafted-prd-promotion.md`. Scope is **Ship 1** — team drafts and
platform-level promotion only.

Story IDs follow the EARS unit numbers. Every `acceptance:` line quotes the R-IDs
it discharges; the EARS table is the contract, this file is the slicing.

Order is bottom-up: residency and the digest helper first (they are what several
later stories share), then authoring, then promotion, then the readers, and the
golden re-baseline last — R-10.4 requires it to land as its own commit, so it
cannot be folded into the story that changes behaviour.

## Unit 1: Draft residency

- [x] 1.1 Draft path resolution and enumeration (est: ~1h)
  - why: R-1.1 fixes the location so promotion, abandonment, `today` and the integrity gate address one path instead of each re-deriving it. Absence has to be a non-event or every draft-less workspace grows a finding.
  - acceptance: R-1.1 — resolve `teams/<team>/product/change-records/draft`; R-1.2 — a draft lives at `<draft-dir>/<draft-id>/prd.md`; R-1.3 — IF the directory does not exist, treat the team as having no drafts and emit no finding
  - verify: `go test ./internal/product/... -run Draft`; `company-os validate` on today's `examples/workspace` (no draft dirs anywhere) still exits 0 with no new findings
  - landed: (uncommitted) — company-os-starter/internal/product/draft.go, company-os-starter/internal/product/draft_test.go

## Unit 2: Creating a draft

- [x] 2.1 `prd new --draft` scaffolds a draft (deps: 1.1, est: ~2h30m)
  - why: R-2.1's point is that drafting is as cheap as scaffolding a discovery brief. It reuses the existing template chain rather than a second one, so a team that overrode the PRD template does not get a stock draft.
  - acceptance: R-2.1 — `prd new --team <t> --draft "<title>"` writes `type: prd`, `status: draft`, a derived `id`, `created` = today; R-2.2 — no `level` key (that word means rule tier here); R-2.3 — body comes from `ResolveTemplate` with name `prd` and the team id, honouring team → platform → company → built-in precedence; R-2.4 — `title`, `platform`, `components`, `governanceSnapshot`, `decisionOwner` written as placeholders and `promoteTo.platform` as a placeholder; R-2.5 — WHERE `--draft --platform <p>` is supplied, set both `promoteTo.platform` and `platform`; R-2.10 — no hand-written `tags:` key
  - verify: `go test ./internal/product/... -run DraftNew`; `company-os prd new --team <t> --draft "<title>"` in `examples/workspace`, then `company-os graph build` is a no-op on the new file's `tags:`
  - landed: (uncommitted) — company-os-starter/internal/product/draft.go, company-os-starter/internal/product/prd.go, company-os-starter/internal/product/sections.go, company-os-starter/internal/model/codes.go, company-os-starter/internal/render/product.go, company-os-starter/cmd/company-os/args.go, company-os-starter/cmd/company-os/product.go, company-os-starter/cmd/company-os/args_test.go, company-os-starter/internal/product/draft_new_test.go

- [x] 2.2 Draft creation refusals, discovery origin, and guidance chain (deps: 2.1, est: ~1h30m)
  - why: without R-2.6 a second `prd new --draft` silently overwrites a draft someone is editing; R-2.8 is what keeps `validate` green without teaching every author to run `graph build`.
  - acceptance: R-2.6 — refuse on an existing derived id with a conflict error, no overwrite; R-2.7 — `--draft --from-discovery <brief-id>` requires the brief to be `status: validated` and copies its Problem and Success sections forward; R-2.8 — rebuild derived artifacts before returning so `validate` exits 0 with no separate `graph build`; R-2.9 — name the draft file as the next thing to edit and do NOT print `prd promote` as the immediate next command
  - verify: `go test ./internal/product/... -run DraftNew`; `make check`
  - landed: (uncommitted) — company-os-starter/internal/product/draft.go, company-os-starter/internal/product/prd.go, company-os-starter/internal/model/codes.go, company-os-starter/internal/render/product.go, company-os-starter/internal/product/draft_new_test.go

## Unit 3: Draft-time validation posture

- [x] 3.1 Drafts carry the core contract only (deps: 2.1, est: ~1h30m)
  - why: R-3.1–R-3.3 are the invariant that makes the stage worth having. Gate 3 already sweeps only `platforms/*/change-records/active` (`internal/product/check.go:96-105`) and `CoreFieldErrors` already stops at type/identity/status for `type: prd`, so most of this story is pinning that split with tests before the next refactor breaks it.
  - acceptance: R-3.1 — WHILE `status: draft`, require only `type`, an `id` identity, and `status`; R-3.2 — do not check the body for PRD section headings; R-3.3 — team drafts are not included in the workspace PRD contract gate; R-3.5 — a draft omitting `status` still fails core-field validation, because `prd` is a lifecycle type
  - verify: `go test ./internal/validate/... ./internal/product/...` against a fixture workspace holding a deliberately incomplete draft
  - landed: (uncommitted) — behaviour was already correct; no logic changed. New `company-os-starter/internal/validate/draft_posture_test.go` pins R-3.1/R-3.2/R-3.3/R-3.5 against hand-written incomplete drafts (not DraftNew output, so the tests fail if validate tightens even when the CLI still writes complete drafts). Rationale comments added at the two refactor sites: `internal/product/check.go` (why the gate-3 corpus stays `platforms/*/change-records/active`) and `internal/product/contract.go` (why `CoreFieldErrors` stops at type/identity/status, and why `status` is still mandatory).

- [x] 3.2 Draft location facet and green-workspace guarantee (deps: 3.1, est: ~1h)
  - why: R-3.6 is what makes a draft findable through the same tag vocabulary as everything else; it comes from the parent-directory inference at `internal/graph/tags.go:201-207`, so the work is confirming it fires for the new depth, not writing new derivation.
  - acceptance: R-3.4 — `validate` exits 0 over a workspace whose only additions are well-formed drafts with current derived artifacts; R-3.6 — a `team/<t>` location facet is derived for a draft
  - verify: `go test ./internal/graph/...`; `company-os graph build` twice in `examples/workspace` leaves the draft's `tags:` byte-identical
  - landed: (uncommitted) — derivation already fired at the new depth; no change to `derive_tags`. New `company-os-starter/internal/graph/draft_facet_test.go` proves the location facet survives a draft's four-deep path using a fixture that omits `team:` frontmatter (so the directory inference is the only possible source), plus a build-twice idempotence test. Confirmed against `examples/workspace`: a draft tags `[kind/prd, platform/TODO, status/draft, team/customer-engagement]`, byte-identical across two builds, `validate` exits 0, gate 3 does not list it and gate 4 does. `platform/TODO` is correct derived output for an unrouted draft and does not affect R-3.6.

## Unit 4: Resolving the promotion target

- [x] 4.1 Promotion target resolution and refusals (deps: 1.1, est: ~1h30m)
  - why: R-4.1–R-4.4 make the destination something the author wrote down and the tool checked, rather than something inferred while files are being written.
  - acceptance: R-4.1 — read the target from `promoteTo.platform`; R-4.2 — IF absent or still a placeholder, refuse with an artifact error stating the draft declares no target platform; R-4.3 — IF it names a platform not in the workspace, refuse with a workspace error naming it; R-4.4 — IF `promoteTo.platform` and the frontmatter `platform` disagree, refuse and report both values; R-4.5 — reserve `promoteTo.scope` for a later increment, neither written nor read
  - verify: `go test ./internal/product/... -run PromoteTarget`
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_test.go`

## Unit 5: Promotion by copy-forward

- [x] 5.1 Promotion readiness preflight (deps: 4.1, est: ~2h)
  - why: R-4.6 requires every rule in Unit 4 and every readiness rule here to be evaluated before a single file is written — a half-promoted workspace is the one state re-running cannot obviously repair. R-5.5 is the tone rule: a draft that is not ready is not a malformed artifact, and reporting it as one trains people to ignore the message.
  - acceptance: R-4.6 — evaluate all target and readiness rules before writing any file; R-5.3 — refuse unless all six process fields (`title`, `team`, `platform`, `components`, `governanceSnapshot`, `decisionOwner`) carry non-placeholder values; R-5.4 — refuse when the body omits `Problem statement`, `Success metrics` or `Proposed change`, naming each missing section; R-5.5 — phrase readiness refusals as a not-yet-ready report naming what to supply; R-5.6 — IF `status` is not `draft`, refuse with a conflict error stating the status found; R-5.7 — refuse on an existing target id without overwriting, except under the R-5.11 resume condition
  - verify: `go test ./internal/product/... -run PromotePreflight`; after any refusal, `git status --porcelain` in `examples/workspace` is empty
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_copy_test.go`, `company-os-starter/cmd/company-os/product.go`, `company-os-starter/cmd/company-os/args.go`, `company-os-starter/internal/render/product.go`, `company-os-starter/internal/model/codes.go`

- [x] 5.2 Copy-forward write and effect ordering (deps: 5.1, 6.1, est: ~2h30m)
  - why: R-5.9 pins the effect order because the digest in R-6.3 is taken over the draft *after* it has been rewritten and derived artifacts rebuilt — do the steps in any other order and the recorded digest can never match what the gate recomputes.
  - acceptance: R-5.1 — `prd promote --team <t> <draft-id>` creates `platforms/<p>/change-records/active/<id>/prd.md` seeded from the draft body; R-5.2 — the promoted record carries `status: proposed` and `created` = today; R-5.8 — the origin draft is not deleted, relocated or truncated; R-5.9 — effects occur in the order: write target record, rewrite draft, rebuild derived artifacts, compute digest, (record it); R-5.10 — `validate` exits 0 afterwards with no intervening `graph build`; R-5.12 — print `company-os prd validate --platform <p> <id>` as the next command
  - verify: `go test ./internal/product/... -run PromoteCopy`; promote in `examples/workspace`, then `company-os validate` exits 0 without running `graph build`
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_copy_test.go`

- [x] 5.3 Interrupted-promotion resume (deps: 5.2, est: ~1h30m)
  - why: R-5.11 is the whole reason the collision check in 5.1 has an exception — re-running the command has to be the repair for a promotion that died between step one and step two.
  - acceptance: R-5.11 — IF the target record exists, carries no `promotedFrom`, and the draft is still `status: draft`, treat the prior attempt as interrupted and resume rather than refusing
  - verify: `go test ./internal/product/... -run PromoteResume`; a run interrupted after the target write, then re-run, reaches the same tree as an uninterrupted promotion
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_copy_test.go`

- [x] 5.4 Downstream parity for promoted records (deps: 5.2, est: ~1h30m)
  - why: R-5.13–R-5.17 are the payoff of copying forward rather than inventing a new record shape — if any downstream command needs a special case for promoted PRDs, the design failed.
  - acceptance: R-5.13 — `prd validate` produces the same result as for a hand-scaffolded record; R-5.14 — same for `prd complete`; R-5.15 — same for `check ready`; R-5.16 — same for `check done`; R-5.17 — the workspace PRD contract gate treats a promoted record identically
  - verify: `go test ./internal/product/... ./internal/validate/...`; run all four commands against one promoted and one hand-scaffolded record and diff the output modulo ids
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_copy_test.go`

## Unit 6: Promotion audit trail

- [x] 6.1 Provenance fields and the shared digest helper (deps: 4.1, est: ~2h)
  - why: R-6.4 is the load-bearing one — promotion-time and validation-time digests must come from one helper, because two implementations of "normalized form" will diverge and turn gate 7 into a source of false failures. 5.2 and 7.1 both consume this, so it lands before either.
  - acceptance: R-6.1 — write `promotedFrom` carrying origin team id, draft id, and a digest of the draft; R-6.2 — rewrite the draft's `status` to `promoted` and write `promotedTo` carrying target platform and record id; R-6.3 — the digest is SHA-256 over a normalized form (LF line endings, trailing whitespace stripped, derived `tags:` excluded) taken after the rewrite; R-6.4 — one shared helper computes it at promotion time and recomputes it at validation time; R-6.6 — never require `promotedFrom`/`promotedTo` on records that were never promoted
  - verify: `go test ./internal/product/... -run Digest`; the same draft digests identically through the promotion path and the gate path
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_test.go`

- [x] 6.2 Promotion log entry (deps: 5.2, 6.1, est: ~1h)
  - why: R-6.5 is what keeps promotion auditable after the draft is rewritten — the record's own frontmatter can be edited, the dated log line is append-only history.
  - acceptance: R-6.5 — WHEN promotion succeeds, append a dated entry to `teams/<t>/log.md` naming the draft id and the target record, creating the file if absent
  - verify: `go test ./internal/product/... -run PromotionLog`; two promotions in one team produce two ordered entries and no rewrite of the first
  - landed: `company-os-starter/internal/product/promote.go`, `company-os-starter/internal/product/promote_copy_test.go`

## Unit 7: Promotion-integrity gate

- [x] 7.1 Promotion-integrity gate (deps: 6.1, est: ~3h)
  - why: R-7.2–R-7.5 are the reason the digest exists at all. A promoted record is the origin of record other teams read; without recomputation, a hand-edit to either side rots the link silently. R-7.7's silence on draft/abandoned is what keeps the gate quiet in the common case.
  - acceptance: R-7.2 — WHILE a draft carries `status: promoted`, recompute its digest and compare to `promotedFrom.digest` on the record named by `promotedTo`; R-7.3 — on mismatch, fail naming the draft and the record it no longer matches; R-7.4 — IF the named record exists at neither the active nor archive location, fail naming the unresolved target; R-7.5 — IF it exists but carries no `promotedFrom`, report an interrupted promotion naming the command to re-run; R-7.6 — resolve an archived record under `archive/prds/<id>/` and treat a digest match there as passing; R-7.7 — emit nothing for drafts carrying `status: draft` or `status: abandoned`; R-7.8 — with no promoted drafts, render the gate header and no findings
  - verify: `go test ./internal/validate/... -run Promotion` with fixtures for each of match / mismatch / unresolved / interrupted / archived / empty
  - landed: `company-os-starter/internal/validate/promotion.go`, `company-os-starter/internal/validate/promotion_test.go`

- [x] 7.2 Gate registration and ordinal placement (deps: 7.1, est: ~1h)
  - why: R-7.1 places the gate after skills and before the conditional federation gate. `internal/validate/validate.go:82-95` decides the denominator before gate 1 runs, so the abort path needs no change — but the five goldens hard-code 7 and 8 (`golden_test.go:86-90`) and will fail until 10.1.
  - acceptance: R-7.1 — the gate runs as part of `validate`, positioned after the skills gate and before the dynamically appended federation gate; R-10.2 — the federation gate stays last in the sequence
  - verify: `go test ./internal/validate/... -run Order`; `company-os validate` on `examples/workspace` reads `[8/8]`, and `[9/9]` on the federated fixture
  - landed: `company-os-starter/internal/validate/validate.go`, `company-os-starter/internal/validate/promotion_test.go`

## Unit 8: Retiring a draft

- [ ] 8.1 `prd abandon` (deps: 1.1, est: ~1h30m)
  - why: R-8.3 is the point — not every draft becomes real, and without a terminal state dead drafts pollute the gate in Unit 7 and the listing in Unit 9. R-8.4 keeps it terminal on purpose: a resurrected draft would have provenance nobody can reason about.
  - acceptance: R-8.1 — `prd abandon --team <t> <draft-id>` sets `status: abandoned`; R-8.2 — IF `status` is not `draft`, refuse with a conflict error stating the status found; R-8.3 — WHILE `abandoned`, exclude the draft from every gate, from promotion, and from role views; R-8.4 — provide no transition out of `abandoned`; R-8.5 — rebuild derived artifacts before returning
  - verify: `go test ./internal/product/... -run Abandon`; after abandoning, `company-os validate` exits 0 and `company-os today --team <t>` no longer lists the draft
  - landed:

## Unit 9: Draft discoverability

- [ ] 9.1 Drafts in `today --team` (deps: 3.2, 8.1, est: ~2h)
  - why: R-9.1 — a drafting stage nobody can find is a scratchpad with extra steps. R-9.3 matters more than it looks: the scaffold writes placeholders, so without it the common case renders a listing full of literal placeholder text.
  - acceptance: R-9.1 — `today --team <t>` lists drafts carrying `status: draft` with each draft's id, title and target platform; R-9.2 — exclude `promoted` and `abandoned` drafts; R-9.3 — WHERE title or target platform is still a placeholder, render it as unset rather than printing the placeholder; R-9.4 — with no open drafts, omit the section rather than rendering an empty one; R-9.5 — list only the scoped team's drafts
  - verify: `go test ./internal/roles/...`; `company-os today --team <t>` in `examples/workspace` before and after creating a draft
  - landed:

## Unit 10: Compatibility

- [ ] 10.1 Golden re-baseline as an isolated commit (deps: 7.2, 9.1, est: ~1h30m)
  - why: R-10.4 requires the re-baseline to carry no behavioural change, so it must be the last commit and nothing else may ride along. Adding gate 8 shifts the ordinal of everything after the insertion point (R-10.3), which is exactly the kind of diff that hides a regression if it is mixed with real edits.
  - acceptance: R-10.1 — slug, title, finding codes and intra-gate finding order of every existing gate are unchanged; R-10.3 — the count and ordinal shift is accepted and recorded as declared in the spec; R-10.4 — the re-baseline lands as an isolated commit containing no behavioural change
  - verify: `make golden`, then review the diff line by line — every change is an ordinal, a denominator, or a new gate-8 line; `make check`; `examples/acceptance.sh` passes
  - landed:

- [ ] 10.2 Non-regression and architecture constraints (deps: 10.1, est: ~1h30m)
  - why: R-10.5 is the promise to everyone not using drafts; R-10.7 and R-10.9 are the standing repo constraints this change is most likely to break, since it adds commands that want to print and a digest that wants a library.
  - acceptance: R-10.5 — `prd new` without `--draft`, `prd complete`, `discover new`, `discover validate` and `--from-discovery` behave unchanged for all inputs that exist today; R-10.6 — every document written here matches the `^---\n...\n---\n` frontmatter parser contract exactly; R-10.7 — printing and process exit stay in `cmd/` and `internal/render/`, every added command returns `[]model.GateResult`; R-10.8 — built-in scaffolding templates stay in sync with the section names the promotion contract checks, so a scaffolded draft satisfies R-5.4 without edits to the template; R-10.9 — no runtime dependency beyond the standard library and existing imports
  - verify: `make check`; `grep -rn "fmt.Print\|os.Exit" internal/` shows no new sites; `go list -deps ./... | grep -v '^\(internal/\|company-os\)'` unchanged from `main`
  - landed:

## Unit 11: Skill and reference guidance

- [x] 11.1 Canonical skills teach the draft stage (deps: 8.1, est: ~2h)
  - why: R-11.1 through R-11.7 — a team meets this feature through its skill, not through the source. Until `creating-prd` names the draft branch, `prd new --draft` is a command only its author knows about, and `completing-a-change` still implies a draft can reach `prd complete`.
  - acceptance: R-11.1 — `skills/creating-prd/SKILL.md` documents `prd new --team <t> --draft`, `prd validate --team <t> <draft-id>` and `prd promote <draft-id> --team <t> --platform <p>`; R-11.2 — it states that platform, components and governance snapshot are required by promotion, not by draft validation; R-11.3 — it states that promotion precedes `prd complete` and keeps the platform-first path documented; R-11.4 — it documents `prd abandon`; R-11.5 — `skills/completing-a-change/SKILL.md` states a draft must be promoted first; R-11.6 — `skills/running-discovery/SKILL.md` names both carry-forward targets; R-11.7 — every added step carries a `(mandatory)`/`(default)`/`(guidance)` tag
  - verify: `company-os skills list` in `examples/workspace` reports no new shadowing or dangling `extends` (R-11.10); every command line quoted in the three skills runs as written against `examples/workspace`; `company-os validate` exits 0
  - landed: company-os-starter/skills/creating-prd/SKILL.md, company-os-starter/skills/completing-a-change/SKILL.md, company-os-starter/skills/running-discovery/SKILL.md

- [x] 11.2 CLI reference and example-workspace skill copies (deps: 11.1, est: ~1h)
  - why: R-11.8 and R-11.9 — the reference page is where someone checks a flag they half-remember, and an example workspace that shadows `skill://product/creating-prd` with pre-draft guidance teaches the old flow to anyone reading the worked example.
  - acceptance: R-11.8 — `company-os-starter/docs/user-guide/reference/company-os-cli.md` lists `prd new --draft`, `prd validate --team`, `prd promote`, `prd abandon`, their flags and their exit codes, and `how-to/take-a-change-from-discovery-to-done.md` names the draft branch; R-11.9 — the shadowing copies under `examples/workspace`, `examples/federated` and `examples/banking` no longer contradict the canonical guidance; R-11.10 — their frontmatter (`id`, `appliesTo`, `precedence`) is unchanged
  - verify: `grep -rn "creating-prd" examples/ --include=*.md` — every hit reviewed; `git diff` on the example skills shows body-only changes; `make check`; `examples/acceptance.sh` passes
  - landed: company-os-starter/docs/user-guide/reference/company-os-cli.md, company-os-starter/docs/user-guide/how-to/take-a-change-from-discovery-to-done.md, examples/workspace/platforms/communications/skills/creating-prd.SKILL.md, examples/federated/platforms/communications/skills/creating-prd.SKILL.md, examples/federated/workspace.lock.yaml, examples/banking/bank/repos/platform-payments/skills/creating-prd.SKILL.md
