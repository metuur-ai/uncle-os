---
type: lld
id: lld-okf-agent-memory-uptake
title: OKF Agent-Memory Uptake — Low-Level Design
status: draft
---

# OKF Agent-Memory Uptake — Low-Level Design

## Architecture

Seven units, landed as four slices (see *Sequencing*). Units 1–2 share one new
file in `internal/governance`; Unit 3 introduces one shared package used by eleven
call sites; Unit 4 touches `internal/scaffold` and `internal/skills`; Units 5–6 are
tests and docs; Unit 7 is a doc reconciliation plus one warn check. Nothing below
`cmd/` prints or exits; every new signal is a `model.Finding` rendered by
`internal/render`.

Paths are relative to `company-os-starter/` unless they start with `docs/`,
`examples/` or `.devlocal/`, which are repo-root.

```
internal/governance/provenance_gate.go     Unit 1+2: provenance shape + stale checks, called from ExpiryGate
internal/governance/gates.go               ExpiryGate gains one call at its tail
internal/logbook/logbook.go                Unit 3: Append() + Line() — the one writer and the one sentence
internal/validate/logformat.go             Unit 3: log.md shape check, called from frontmatterGate
internal/validate/versioncheck.go          Unit 7: companyOsVersion/profile check, called from frontmatterGate
internal/validate/validate.go              Run(ws) kept as wrapper over RunWith(ws, RunOptions)
internal/scaffold/contract.go              Unit 4: WriteContract(root) — AGENTS.md / CLAUDE.md smart-append
internal/scaffold/contract_agents.md       Unit 4: embedded Minimal Agent Contract (go:embed)
internal/skills/install.go                 Unit 4: Install() also calls scaffold.WriteContract
internal/model/codes.go                    new finding codes (see Key Decision 11 for prefixes)
cmd/company-os/args.go                     Stale bool beside Fix (:52), flag entry beside fix (:225)
cmd/company-os/scenarios_test.go           Unit 5: TC-S1…S6
docs/AGENT_TESTING.md                      Unit 5 pass criteria      (company-os-starter/docs/)
docs/SECURITY.md                           Unit 6                    (company-os-starter/docs/)
docs/FRONTMATTER-CORE.md                   stale_after; provenance consistency; "derivation is not logged"
docs/CONFORMANCE.md                        §6 version scheme + profile enum + log.md shape note (Unit 7, D3)
docs/01-flexibility-skills-and-role-views.md   :121 enum corrected (Unit 7)
templates/outcome-review.md                stale_after + generated placeholders (Unit 2)
```

## Measured facts (2026-09-08)

Anchors the design depends on, verified in source this session. Line numbers are
for orientation; symbol names are the contract.

| Fact | Where |
|---|---|
| `ExpiryGate(ws, ordinal)` walks `ws.AllTeams()` only | `internal/governance/gates.go:142` |
| `before()` is the blocking ISO-date parser for `reviewDate`/`expires`; hard-errors on non-`YYYY-MM-DD` | `gates.go:372-383` |
| `today()` — two plain `time.Now()` wrappers, no seam, nothing pins them | `internal/governance/declare.go:184`, `internal/product/pysem.go:33` |
| `validate.Run(ws)` takes no options; `--fix` calls `FixDerived` before `Run` | `internal/validate/validate.go:60`, `cmd/company-os/validate.go:26-48` |
| `Run` callers | `cmd/company-os/validate.go:31,48`, `internal/validate/golden_test.go:52,143`, `internal/validate/draft_posture_test.go:107`, `internal/skills/starterkit_test.go:92` |
| `frontmatterGate` uses `IterGraphDocs`; `IterGraphDocs` walks `graphRoots` = company + platforms + teams + ontology — never `ws.Root`; skips docs with empty frontmatter; hard-errors on non-mapping frontmatter | `validate.go:153-155`, `internal/graph/tags.go:206,236-240,279-284` |
| `skipNames` = `log.md README.md CLAUDE.md index.md` | `tags.go:52` |
| `NodeRoots` = company + platforms + teams + ontology + knowledge; not `ws.Root` | `internal/graph/node.go:202` |
| generated markers `startRE`/`endRE` match only `company-os:generated:*` | `node.go:38-39` |
| `graph` imports `federation frontmatter model workspace yamlio`; `governance` imports `ids model workspace yamlio`; `federation` imports `model workspace yamlio` → `governance→graph` is cycle-free | import blocks |
| `product` imports `scaffold`, `graph`, `governance` → any of those importing `product` is a cycle | `internal/product/discover.go:15,17`, `checklist.go:22` |
| `Init`: refuse if `ws.IsRoot()`; `scaffoldWorkspace(staging…)` at `:121`; `moveRoots(staging, target)` at `:128`; compensated unwind `:164-181` | `internal/scaffold/commands.go:91-133` |
| `scaffoldWorkspace` creates company, one platform, one team — **no component** | `commands.go:139-159` |
| `outcomeDoc()` writes `generated:` + `due:`; `templates/outcome-review.md` has neither `generated:` nor `stale_after:` | `internal/product/prd.go:639-648`, `templates/outcome-review.md:1-8` |
| `appendLog` (func `prd.go:652`, call `:414`), `appendPromotionLog` (`promote.go:587`) — the two existing writers | |
| `DiscoverNew` on an existing brief → `ExitConflict` "`%s already exists`", writes nothing | `internal/product/discover.go:46-54` |
| `next:` chain actually printed: `discover new → discover validate` (`discover.go:79`), `→ prd new` (`:172`), `prd new → prd validate` (`prd.go:141`), `prd validate → prd complete` (`:329`), `prd complete → validate` (`:438`). `check ready` and `reality new` are on no `next:` line; `reality new` is offered as `fix:` (`CodeDoneFix`, `sections.go:193`) on the **failing** branch of `prd complete` (`prd.go:369-383`) | |
| `FieldNext = "next"` is a structured `--json` field | `internal/model/codes.go:707` |
| `runArgs(t, argv...) (code, stdout, stderr)` end-to-end helper | `cmd/company-os/selftest_test.go:30` |
| `skills.Install(ws, rebuild)` writes embedded text into `company-os/skills/` with version compare | `internal/skills/install.go:81` |
| `PRDAbandon`, `DiscoverValidate` producers | `internal/product/abandon.go:40`, `discover.go:96` |
| Parent R-1.7 enum `minimal \| standard \| strict`; `docs/01-flexibility…md:121` says `standard \| strict \| provisional`; `scaffoldPlatform` emits `2026.2`/`standard` | `docs/ears/okf-v02-conformance.md:84`, `scaffold.go:196-197` |
| `companyOsVersion` appears in **nine** fixture files; `profile: minimal` in two `examples/banking` files; `examples/federated` has **no `platform.yaml`** | `docs/CONFORMANCE.md:147`; glob |
| Fixtures: zero `generated:`/`verified:`/`stale_after:` frontmatter; all four `log.md` match `^- \d{4}-\d{2}-\d{2}: \S`; `examples/banking` not in `acceptance.sh` | grep |

### Unit 1 — provenance consistency (gate 2, warn)

**Where it hooks.** `governance.ExpiryGate` iterates teams for deviations/exceptions.
After that loop, before returning `g`, it calls `provenanceFindings(ws)` and appends
the result. Gate 2 was chosen by the user (research §9 Q1); ordinal, slug and title
are unchanged. Human readers of gate 2 will see `provenance`-worded findings under
an "expiry" header — accepted cost, recorded in Decision 1; the finding *message*
begins with the word `provenance:` so the text output is self-labelling.

**What it walks.** `graph.IterGraphDocs(ws)` — the same iterator gate 4 uses, so
`knowledge/`, `scratchpad/` and `skipNames` are excluded for free. Import is
cycle-free (Measured facts).

**Never aborts gate 2.** `IterGraphDocs` hard-errors on non-mapping frontmatter.
Today that abort happens in gate 4, after gates 1–3 have printed; moving the walk
into gate 2 would truncate output two gates earlier for a malformed workspace.
`provenanceFindings` therefore **swallows** the iterator error, emits nothing for
it, and lets gate 4 report it exactly as today. No fixture exercises this path
(`failing-workspace` gate 4 completes), so goldens are unaffected either way — the
rule exists so behaviour does not change silently.

**Checks (all `SevWarn`, one finding per problem, no `[ok]` lines):**

| # | Condition | Code | Fields |
|---|---|---|---|
| 1 | `generated:` is a mapping and `by` is falsy | `expiry.provenance-generated-no-by` | `path` |
| 2 | `generated:` present but not a mapping | `expiry.provenance-generated-not-mapping` | `path` |
| 3 | `verified[i].at` and `generated.at` both parse and `verified[i].at < generated.at` | `expiry.provenance-superseded-verification` | `path`, `index`, `verifiedAt`, `generatedAt` |
| 4 | any actor (`generated.by`, `verified[i].by`) fails the actor regex | `expiry.provenance-actor-format` | `path`, `actor` |
| 5 | actor has `<prefix>:` form and prefix ∉ {`human`, `process`} | `expiry.provenance-actor-prefix` | `path`, `actor`, `prefix` |

Actor regex is OKF's, verbatim: `^(?:[a-zA-Z][\w.-]*:\S+|[^\s/]+/[^\s/]+)$`
(OKF `pkg/okf/types.go:9-15`). `company-os/<version>` and `company-os/dev` match
the second branch.

**Date parsing.** A **new** `parseProvenanceDate(s) (time.Time, bool)` in
`provenance_gate.go` accepts `YYYY-MM-DD` or RFC3339 and returns `ok=false`
otherwise. It does **not** touch `before()`: `before` is the blocking parser for
`reviewDate`/`expires` and extending it would silently start accepting RFC3339
expiry dates that fail today. An unparseable provenance date is not a finding;
only check 3 needs both dates parsed.

`mappingsOf` (`internal/graph/provenance.go:120`) already normalises a scalar
`verified:` into a list; export it as `graph.VerifiedEntries` and reuse.

**Emission rule.** No per-document `[ok]`. Fixtures carry zero
`generated:`/`verified:`, so gate 2 output is byte-identical across all five
goldens (`golden-validate.txt`, `federated-`, `failing-workspace-`,
`failing-federated-`, `failing-federated-nolock-`).

### Unit 2 — `stale_after` (gate 2 warn, `--stale` fail)

**Field.** `stale_after: YYYY-MM-DD`, optional, any typed document. Documented in
`docs/FRONTMATTER-CORE.md` under Tier 2 next to `updated`, marked *optional*; the
Tier 2 heading ("Lifecycle — required per doc family") gains "unless marked
optional". Semantics: "on or after this date the document should be re-read;
nothing is deleted or archived because of it. Clear the signal by advancing or
removing the field once the re-read has happened."

**Check.** Same walk as Unit 1. If `stale_after` parses (via
`parseProvenanceDate`) and `stale_after <= today` → `expiry.stale` with fields
`path`, `staleAfter`, `today`; `SevWarn` unless `opts.Stale` → `SevFail`.
Unparseable → `expiry.stale-after-unparseable`, `SevWarn` always.

**Clock.** No injectable clock exists (Measured facts) and none is added. Tests use
far-dated values: `2000-01-01` for stale, `2099-01-01` for fresh — the same
technique the existing expiry fixtures use (`review 2035-01-15`). TC-S2's
`outcome.md` gets `stale_after` = today+90, never `<= today` during the test.

**Flag.** `validate --stale`. `cmd/company-os/args.go`: `Stale bool` beside `Fix`
(`:52`), flag entry beside `fix` (`:225`). `internal/validate`: add
`type RunOptions struct{ Stale bool }` and `RunWith(ws, RunOptions)`; keep
`Run(ws)` as `RunWith(ws, RunOptions{})` so the four existing test callers do not
churn. `cmd/company-os/validate.go:31,48` call `RunWith(ws, RunOptions{Stale:
args.Stale})`. `--stale` composes with `--fix`: `FixDerived` runs first, then
`RunWith` with stale promoted.

**Producer.** `outcomeDoc()` (`prd.go:639`) adds `stale_after: <due>` directly
under `due:`. `due:` stays — it is the review deadline the outcome process reads;
`stale_after` is the freshness signal the validator reads. Same value, two
consumers, one line each.

**Template sync (CLAUDE.md convention).** `templates/outcome-review.md` gains
`stale_after: <YYYY-MM-DD>` **and** the `generated:` placeholder block it already
lacks (a pre-existing drift found during this review; fixing one without the other
leaves the template still out of sync). No fixture runs `prd complete`, so no
golden moves; `internal/product` tests asserting `outcome.md` bytes are updated.

**Standing consequence.** Every archived `outcome.md` begins warning 90 days after
completion and, under `--stale`, fails. That is the intended nudge; the reviewer
clears it by advancing or removing `stale_after` when the outcome review is done.
Stated in `FRONTMATTER-CORE.md` and in HLD Non-Goals (`--stale` never blocks
`check`/`prd complete`).

### Unit 3 — log discipline

**Package.** New `internal/logbook`, importing only `model`:

```go
// Append writes "- <date>: <text>\n" with O_APPEND|O_CREATE|O_WRONLY, 0o666.
func Append(path, date, text string) error
// Line renders the human sentence for model.CodeLogAppended from its Fields.
func Line(f model.Fields) string
```

`logbook` owns **both** the writer and the sentence. This is forced, not
stylistic: `CodeLogAppended`'s sentence lives in `internal/product/sections.go:200`,
and `product` imports `scaffold`, `graph` and `governance` — so `scaffold`,
`federation` (via `graph`) and `governance` cannot import `product` without a
cycle. The code **constant** stays in `model` (no cycle); every renderer that can
emit it calls `logbook.Line`. `product`'s existing `case` delegates to
`logbook.Line` so the sentence has one home.

`appendLog` (`prd.go:652`) and `appendPromotionLog` (`promote.go:587`) are
**re-pointed at `logbook.Append` with their line text unchanged**, so
`examples/workspace/platforms/communications/log.md` and the promote/complete tests
keep their bytes.

**Call sites and destinations** (nearest node root owning the mutated artifact).
Line vocabulary follows the committed corpus (`examples/banking/.../log.md` uses
`PRD \`<id>\` proposed (from discovery <id>)`):

| Command | Producer | log.md | Line text |
|---|---|---|---|
| `deviation declare` | `governance.Declare` (`declare.go:59`) | `teams/<t>/log.md` | `deviation declared for rule \`<rule>\`; review <reviewDate>` |
| `exception request` | `governance.Request` (`declare.go:101`) | `teams/<t>/log.md` | `exception requested for rule \`<rule>\` on \`<component>\`; expires <expires>` |
| `discover new` | `product.DiscoverNew` (`discover.go:27`) | `teams/<t>/log.md` | `discovery brief \`<id>\` created` |
| `discover validate` | `product.DiscoverValidate` (`discover.go:96`) | `teams/<t>/log.md` | `discovery brief \`<id>\` validated` |
| `prd new` | `product.PRDNew` (`prd.go:43`) | `platforms/<p>/log.md` | `PRD \`<id>\` proposed` + ` (from discovery <brief>)` when set |
| `prd new --draft` | `product.DraftNew` (`draft.go:88`; reached via the `--draft` flag, `args.go:191`) | `teams/<t>/log.md` | `PRD draft \`<id>\` created` + same suffix rule |
| `prd abandon` | `product.PRDAbandon` (`abandon.go:40`) | `teams/<t>/log.md` | `PRD draft \`<id>\` abandoned` |
| `add platform\|team\|component` | `scaffold.Add` (`commands.go:367`) | created/owning root's `log.md` | `<kind> \`<name>\` added` |
| `reality new` | `scaffold.RealityNew` (`commands.go:474`) | `platforms/<p>/log.md` | `reality doc for \`<component>\` scaffolded` |
| `workspace sync` | `federation.Sync` (`sync.go:93`) | `<root>/log.md` | `workspace synced: <n> repo(s), lock updated` |
| `prd promote` / `prd complete` | existing | unchanged | unchanged |

`discover validate` and `prd abandon` were added on the user's decision (2026-09-08):
every command that mutates an artifact's status or existence logs.

**Excluded:** `init`, `scratchpad init`, `governance resolve`, `graph build`/`derive`,
`validate --fix`, `skills install`. These are derivation, not authored mutation;
logging them would flood `log.md` with lines carrying no decision. Boundary stated
in `FRONTMATTER-CORE.md`.

**Surfacing.** Producers that return `[]model.GateResult` add one
`okFinding(model.CodeLogAppended, …, Fields{"path": relLog})` beside their
`next:` finding. `Declare`/`Request` return plain structs (`Declared`, `Requested`)
rendered by `internal/render/governance.go`; each gains a `Log string` field,
rendered as one `log -> <path>` line in text and as a `"log"` key in `--json`.

**`workspace sync` destination** = `<root>/log.md`. Decided: `acceptance.sh` never
runs `sync`, so no fixture grows a root `log.md`; `knowledge/` is read-only.

**Shape check (gate 4, warn).** `internal/validate/logformat.go`: for `ws.Root` and
every `NodeRoots(ws)` root, if `log.md` exists, read it as text; each non-blank
line must match `^- \d{4}-\d{2}-\d{2}: \S`. Non-matching line →
`frontmatter.log-line-format` (`SevWarn`, fields `path`, `line`, `text` ≤ 80 chars).
Called at the tail of `frontmatterGate` so it renders after document findings.
`log.md` stays in `skipNames`.

### Unit 4 — the agent contract

**Files** at `ws.Root` (not a `NodeRoots` entry; the fixture has no root
`CLAUDE.md` and gate 5 must not start expecting one):

- `AGENTS.md` — Minimal Agent Contract, embedded via `go:embed` from
  `internal/scaffold/contract_agents.md`, wrapped in
  `<!-- company-os:agent-contract:start -->` … `<!-- company-os:agent-contract:end -->`.
- `CLAUDE.md` — inside the same delimiters: `@AGENTS.md` plus one comment line:
  `<!-- Claude Code resolves @AGENTS.md; other runtimes read AGENTS.md directly. Content below this block is yours. -->`.
  User decision (2026-09-08): import only, no second copy.

**Contract text** — nine rules, each traceable to a skill or invariant:

1. Search before you write: `company-os find "<query>"` (unified local search
   over IDs, tags, indexes and the graph — `args.go:299`); never scan
   `knowledge/` by listing directories.
2. Never edit anything under `generated/`, inside `company-os:generated` markers,
   or under `knowledge/` — run `company-os validate --fix` to regenerate.
3. Never hand-write `tags:`; set source fields and run `company-os derive`
   (the command every validate finding prints; `graph build` is its alias).
4. Run `company-os validate` before claiming a change is done; `prd complete`
   refuses until `reality/components/<id>.md` is newer than the PRD.
5. Follow the printed next command; every mutating command prints one.
6. Record decisions in the artifact, not in chat: discovery → PRD → reality.
7. Never persist secrets, credentials, PII or production payloads
   (`docs/SECURITY.md`).
8. Do not claim human verification: `verified: [{by: human:…}]` is written by a
   person; agents write `process:` or `<tool>/<version>`.
9. Skills: `company-os skills list`; install with `company-os skills install`.

**Writer.** `scaffold.WriteContract(root string) (ContractResult, error)`, one
smart-append rule per file, mirroring OKF `bootstrap.go:16-54`: absent → write;
present with start marker → byte-for-byte untouched; present without → append
`\n` + block. Never truncate. The contract is **not versioned or compared** (unlike
skills): a marker means "leave it".

**Two hosts:**

1. `scaffold.Init` (`commands.go:91`) — a **new step after `moveRoots(staging,
   target)` at `:128`**, not inside `scaffoldWorkspace` (which only sees the staging
   dir). `Init` is allowed in a non-empty non-workspace directory, so the target may
   already hold `AGENTS.md`/`CLAUDE.md` — hence smart-append. This step sits outside
   `moveRoots`' compensated unwind: if it fails, the four roots are already in place
   and a re-run is refused. So a failure here is surfaced as a **warn finding**
   (`scaffold.contract-not-written`, fields `path`, `error`) naming
   `company-os skills install` as the repair, not as an error.
2. `skills.Install` (`install.go:81`) — calls `WriteContract(ws.Root)` after the
   skills loop and before `rebuild`. This is how every pre-existing workspace
   (including `examples/workspace` and the tutorial's) gets the contract: the
   existing precedent for "write embedded canonical text into a workspace". User
   decision (2026-09-08) over a copy-paste story or a new `contract install`
   command. `InstallResult` gains `Contract ContractResult`; render prints one
   line per file (`written`/`appended`/`unchanged`).

**Graph interaction.** `IterGraphDocs` never walks `ws.Root` and skips
frontmatter-less files, so a root `AGENTS.md` is already invisible to gate 4.
`AGENTS.md` is still added to `skipNames` as belt-and-braces — it also hides an
`AGENTS.md` someone drops *inside* a node root, which is the case the walk would
otherwise reach. `graph build` and gate 5 ignore both root files. Acceptance step 4
(double-build SHA) holds because neither file is derived.

**`add` does not touch the contract.**

### Unit 5 — scenario tests

**Placement.** `cmd/company-os/scenarios_test.go`, built on `runArgs`
(`selftest_test.go:30`). `next:`/`fix:` commands are read from `--json`
(`FieldNext = "next"`, `CodeDoneFix` fields), never parsed from prose.

| ID | Name | Drives | Pass criterion |
|---|---|---|---|
| TC-S1 | FreshWorkspaceContract | `init` | `AGENTS.md` and `CLAUDE.md` exist with both delimiters; `AGENTS.md` contains an anchor phrase from each of the nine rules; `CLAUDE.md` contains `@AGENTS.md`; `validate` exits 0 |
| TC-S2 | SkillChainFromInit | `init → add component → discover new → discover validate → prd new --from-discovery --components → prd validate → prd complete (①) → reality new → prd complete (②) → validate` | Steps joined by a printed `next:` are driven by it; `add component` and the retry ② come from the skill text (documented in `AGENT_TESTING.md`). ① **must** exit non-zero with the done-check and print `fix: company-os reality new …`, which the test executes verbatim. ② exits 0; final `validate` is `PASS`; `archive/prds/<id>/outcome.md` carries `stale_after: <due>` equal to its `due:` |
| TC-S3 | AntiDuplication | `discover new` twice, same title | second call exits `ExitConflict`, stderr names the existing path, first `brief.md` byte-identical, no other file created |
| TC-S4 | HumanCorrectionPreserved | PRD with `decisionOwner: Ada` → complete | archived PRD still has `decisionOwner: Ada`; `generated:` present in `outcome.md` only |
| TC-S5 | LogTrail | TC-S2 chain | one `- YYYY-MM-DD: ` line per mutating step in the `log.md` named in Unit 3 (① adds none); gate 4 emits zero `frontmatter.log-line-format` |
| TC-S6 | StaleAndProvenanceSignals | fixture with `stale_after: 2000-01-01` and `verified[0].at < generated.at` | `validate` exit 0 with exactly one `expiry.stale` warn and one `expiry.provenance-superseded-verification` warn; `validate --stale` exit ≠ 0 with `expiry.stale` at `fail` and no other difference |

**`docs/AGENT_TESTING.md`** lists the six scenarios and pass criteria, the
definition of done borrowed from OKF (`ROADMAP.md:516-538`): *"a completely new
agent can continue the workspace using the repository alone"*, states which TC-S2
steps are skill-driven rather than `next:`-driven, and requires a row here for any
new scenario.

### Unit 6 — `docs/SECURITY.md`

Unchanged from the first draft: never-persist table; knowledge integrity (gate 9,
`internal/federation/lock.go:226`); poisoning (no `sources[]` — untraceable content
stays in `scratchpad/`); audit trail (`log.md` + `git log`, gate 4 line-format
warn); deletion. Linked from contract rule 7 and `README.md`.

### Unit 7 — version handshake and enum reconciliation

**Definitions** in `docs/CONFORMANCE.md §6`:

- `companyOsVersion` scheme `YYYY.N`, compared **numerically** (year, then N).
  `2026.2` is the first and only defined value. Forward-only bump (parent R-1.6).
- `profile` enum **`minimal | standard | strict`** — the parent's R-1.7, adopted
  on the user's decision (2026-09-08). `standard` = all gates; `minimal` = gates
  1–4; `strict` = reserved for a future "all warns are fails" posture. Gating by
  profile is **not** implemented here (parent task 4.2).
- Compatibility: a validator built for `V` accepts declared `≤ V` silently, warns
  on `> V` or unparseable. Today, with `V = 2026.2` and no earlier version, that
  means: `2026.2` silent, anything else warns.
- `log.md` shape note (Decision 3): "uses dated bullets; OKF §9 heading form is not
  used".

**Three sync points**, all edited in this unit: `CONFORMANCE.md §6` (definition),
`docs/01-flexibility-skills-and-role-views.md:121` (comment becomes
`# minimal | standard | strict`; `provisional` is dropped — never used by any
fixture), `internal/scaffold/scaffold.go:196-197` (already emits `2026.2`/`standard`;
cited as the writer, no change).

**Check** (`internal/validate/versioncheck.go`, tail of `frontmatterGate`): for
each `ws.AllPlatforms()`, if `platform.yaml` is **absent** → nothing
(`examples/federated` has none). If present with no `conformance` block → nothing.
`companyOsVersion` unparseable or `> V` → `frontmatter.conformance-version`
(`path`, `declared`, `supported`). `profile` ∉ enum →
`frontmatter.conformance-profile` (`path`, `declared`).

**Location deviation from research T7** (Decision 12): OKF reads the root
`index.md`; uncle-os reads `platform.yaml`, because that is where the nine fixtures
already commit the value.

`docs/tasks/okf-v02-conformance.md` task 1.3 gets an "**Update 2026-09-08.**
Completed by `okf-agent-memory-uptake` Unit 7" note, matching its 2026-08-26 style.

## Constraints

- **Finding codes are a contract** (`codes.go:5-21`); new codes follow the
  per-gate prefix convention (Decision 11) and are added in the gate-2 and gate-4
  blocks. Rendered by exactly one `case` **per renderer that can emit them**
  (`log-appended` is emitted by product, governance, scaffold and federation
  renderers; each case calls `logbook.Line`).
- **Renderer-only output.** No `fmt.Print` below `cmd/`.
- **`--json` parity.** Each finding's `Fields` carries every value its sentence
  shows; `Declared`/`Requested` expose `Log` as a JSON key.
- **Determinism.** `IterGraphDocs` order; `log.md` top-to-bottom; `AllPlatforms()`
  order.
- **Golden freeze.** All five goldens unchanged. A diff is a bug in the emission
  rule, never a reason for `--update`.
- **Template/emitter sync** (`CLAUDE.md`): `templates/outcome-review.md` moves
  with `outcomeDoc()`.
- **No marker collision.** `company-os:agent-contract:*` vs `company-os:generated:*`
  (`startRE`/`endRE` `node.go:38-39`).
- **`before()` untouched.**

## Key Decisions

1. **Gate 2 hosts Units 1–2, gate 4 hosts Units 3 and 7.** User decision for 1–2;
   gate 4 is the document-contract gate for the rest. Cost: gate 2 gains a document
   walk and its header reads "expiry" over provenance findings. Mitigations: the
   walk never aborts gate 2; finding text starts with `provenance:`.
2. **Warn-only, no per-item `[ok]`.** Departure from house style to keep five
   goldens byte-identical while fixture adoption is zero. A summary ok-line can be a
   later decision.
3. **`- date:` bullets kept** (user, §9 Q2). One line in `CONFORMANCE.md` records
   the OKF §9 divergence.
4. **`AGENTS.md` is the source; `CLAUDE.md` is `@AGENTS.md`** (user, 2026-09-08).
   Two copies drift; the import is Claude-Code-specific, which the in-block comment
   says.
5. **`skills install` also writes the contract** (user, 2026-09-08). Reuses the
   embedded-text precedent; gives every pre-existing workspace a path; no new
   command; no copy-paste tutorial step.
6. **`outcome.md` is the first `stale_after` producer** — an explicit extension of
   research T2 (HLD G2 names it). `due:` stays.
7. **Actor regex adopted from OKF verbatim.**
8. **Derivation commands do not log; status transitions do** (user, 2026-09-08 for
   `discover validate`/`prd abandon`).
9. **Unit 7 completes parent task 1.3** by adopting the parent's enum and editing
   all three sync points. No parent amendment needed.
10. **Scenarios live in `cmd/`** on `runArgs`, driving from `--json` fields.
11. **New codes use the host gate's prefix** — `expiry.*` for gate 2,
    `frontmatter.*` for gate 4, `scaffold.*` for the init warn — because
    `codes.go` exists so "an inconsistent prefix" is reviewable. `provenance.*`
    was rejected for that reason even though it reads better.
12. **Version handshake reads `platform.yaml`, not a root index** (deviation from
    research T7; the value is already committed there nine times).
13. **No clock seam.** Far-dated fixtures, as existing tests do. A shared
    `var Now` would touch `governance` and `product` for no test that needs it.
14. **`logbook` owns the log sentence**, forced by the `product` import graph.
15. **`Run(ws)` stays as a wrapper** over `RunWith`; four test callers untouched.

## Sequencing

Four slices, each green on `make check` alone:

| Slice | Units | Why here |
|---|---|---|
| A | 7 + 6 | Docs + one read-only gate-4 check; zero fixture risk; closes parent debt first |
| B | 1 + 2 (+ TC-S6) | New gate-2 walk, `--stale`, `RunWith`, `outcome.md`/template |
| C | 3 (+ TC-S5) | **Highest blast radius**: re-points two byte-exact writers, adds nine writers across four packages, two struct/renderer changes, one gate-4 check. TC-S5 lands in the same PR as its guard |
| D | 4 + 5 (rest) | `init`/`skills install` contract + TC-S1–S4; last because TC-S2 needs B and C |

## Out of Scope

- MCP shim (T8 / N1); `sources[]`/attestation (N2); `TrustTier()` changes (I6).
- Gating by `profile` (parent 4.2); `strict` semantics beyond reservation.
- Backfilling `stale_after`/`generated:` into fixtures.
- Skill file changes or `version:` bumps.
- LLM-in-the-loop testing (N9).
- A shared clock seam (Decision 13).
- `--stale` influencing `check ready|done` or `prd complete`.

## Open Questions

1. ~~`discover new` duplicate-title behaviour~~ **Resolved (measured):** refuses
   with `ExitConflict`, names the existing absolute path, writes nothing
   (`discover.go:46-54`). TC-S3 asserts exactly that.
2. ~~`governance`→`graph` cycle~~ **Resolved: none.**
3. ~~`workspace sync` log destination~~ **Resolved:** `<root>/log.md`
   (acceptance never runs `sync`; `knowledge/` is read-only).

None open.

## Pre-mortem findings folded in

- **"Golden moved; someone ran `--update`."** No ok-lines; gate-2 walk never
  aborts; `platform.yaml` absence is silent; I8.
- **"`init` in a repo with `CLAUDE.md` overwrites it."** Smart-append; marker →
  untouched.
- **"`init` half-fails after roots land; re-run refused."** Contract failure is a
  warn naming `skills install` as the repair.
- **"Gate 5 demands a root `CLAUDE.md`."** Root not in `NodeRoots`.
- **"Double `graph build` changes bytes."** Neither root file is derived.
- **"Agents write `human:` verifications."** Rule 8 + check 5.
- **"`stale_after` becomes deletion."** N8; `FRONTMATTER-CORE` says re-read.
- **"Every `outcome.md` warns forever."** Intended nudge; clearing = advance/remove
  the field; stated in docs.
- **"RFC3339 leaks into expiry parsing."** Separate parser; `before()` untouched.
- **"Import cycle from reusing product's sentence."** `logbook.Line`.
- **"TC-S2 asserts a chain the CLI does not print."** Chain rewritten from measured
  `next:`/`fix:` lines; skill-driven steps named.
- **"Two enum definitions after 'completing' 1.3."** Three sync points edited.
- **"Template drifts from `outcomeDoc()`."** Template updated in the same slice.
- **"Malformed frontmatter now aborts in gate 2."** Swallowed; gate 4 reports.
