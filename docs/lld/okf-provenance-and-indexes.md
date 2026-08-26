---
type: lld
id: lld-okf-provenance-and-indexes
title: OKF Provenance and Indexes — Low-Level Design
status: draft
---

# OKF Provenance and Indexes — Low-Level Design

## Architecture

Two units. Unit A (`description:` + `index.md`) lands first and carries the
generation plumbing; Unit B (provenance) rides the same contract work.

### Unit A — the index generator

**Where it hooks.** Index generation is a derive-path concern and must run from
every path that produces derived artifacts, not only from `graph build`. The seam
is `scaffold.Rebuild`, implemented at `cmd/company-os/scaffold.go:46` as
`rebuildGenerated` and passed into `scaffold.Add` (`:117`), `scaffold.RealityNew`
(`:146`), `scaffold.RepairTeam` (`:216`), and the init path (`:69`). A second
records-returning twin serves `prd complete` (`cmd/company-os/product.go:26`).
Both must gain index generation, or `company-os init` emits a workspace that fails
its own `validate` — the D-2.2 failure.

**What it walks.** `graph.IterGraphDocs` (`internal/graph/tags.go:194`) already
yields every typed document with its `Rel` path. Group by `filepath.Dir(Rel)`;
any directory with ≥2 entries qualifies. `knowledge/` is excluded for free —
`IterGraphDocs` deliberately omits it as a graph-docs root
(`internal/graph/tags.go:191`).

**What it writes.** `index.md` carrying frontmatter plus a `company-os:generated`
marker block. The body reuses the shape `buildClaudeNode` already renders —
`- [title](rel)` grouped by type — extended with each entry's `description:`. The
node builder's title fallback chain (`internal/graph/node.go:319-324`: `title` →
`id` → filename) is reused unchanged.

**The re-ingestion loop.** A generated `index.md` carrying frontmatter would be
picked up by the next `IterGraphDocs` walk, counted toward its own directory's
threshold, and listed inside itself. `"index.md"` is added to `skipNames`
(`internal/graph/tags.go:41`, currently `log.md`, `README.md`, `CLAUDE.md`) in the
same commit as the generator — never after. This is exactly how `CLAUDE.md` is
already handled, and the comment at `:39-40` states the reason.

**The drift check.** Gate 5 (`[5/7]` monorepo, `[5/8]` federated) already reports
`CLAUDE.md` context node drift. Index drift reports beneath the same gate. Its
header string stays byte-identical per I5 — it is not renamed to mention indexes,
because renaming it changes a frozen printed string. New `[ok]` / `[fail]` lines
appear under it.

**The federated exclusion.** Before writing, the generator resolves each
manifest-declared slice root from `workspace.yaml` and skips any path beneath one.
`examples/federated/` materializes two graph documents inside a `0444` slice whose
bytes are hashed into `workspace.lock.yaml`; a write there fails gate `[8/8]`.
The drift check skips the same paths, or a slice directory that qualifies but
cannot be written reports permanent drift.

### Unit B — provenance fields

**Shapes.** `generated:` is a mapping `{by, at}`. `verified:` is a sequence of
mappings `[{by, at}]`. Both are producer-authored; neither is written by `derive`.

**Actor convention.** `<producer>/<version>` for agents, `human:<id>` for people,
`process:<id>` for automated processes.

**Trust tier.** Derived from `verified:` at read time, never stored: absent means
unverified; entries with no `human:` actor mean machine-confirmed; any `human:`
actor means human-reviewed. Advisory only — no gate consumes it, no access
control depends on it.

**One-way derivation.** Where a document carries `decisionOwner:` (PRDs) or
`approvedBy:` (deviations, exceptions), a `verified:` entry is derived in memory
for tier computation. The source fields are not rewritten, not migrated, and keep
their current gates. A literal `TODO` value derives nothing.

**Round-trip risk.** Nested collections pass through `PyDumpAutoFlow`
(`internal/graph/tags.go:302`), the site of the flow/block re-layout divergence
recorded as R-0.7a(g) in the port. The existing
`TestRewriteFrontmatterTagsPreservesUnknownKeys` uses a plain string scalar and
does **not** cover this. The preservation test is written against the mapping and
sequence-of-mappings shapes specifically.

## Constraints

- Go module, single static binary, no runtime dependency.
- Nothing below `cmd/` may call `os.Exit` or print. Commands return
  `[]model.GateResult`; only `internal/render/` and `cmd/` write output.
- The `frontmatter()` parser contract expects `^---\n...\n---\n` exactly.
- Every mutating command prints the next command in the workflow (guidance chain).
- Templates live in **four** places that must move together, not two (measured
  while implementing story 1.2): the built-in string in
  `internal/scaffold/template.go`; its peer file under
  `company-os-starter/templates/`; the frozen Python oracle under
  `internal/scaffold/testdata/<name>-template.python.txt`; and the divergence note
  on `TestBuiltinsMatchPythonModuleStrings`. That test pins `DiscoveryTemplate`
  and `PRDTemplate` byte-for-byte to the oracles, so adding any frontmatter field
  makes it go red by design. Editing the oracle is the sanctioned path
  (`internal/scaffold/template_test.go:41-44`, "a deliberate act"), but the file
  is named `.python.txt` and carries a regeneration recipe that no longer
  reproduces it — so the note moves with the bytes or the next reader is misled.
  `reality-component` has no Go constant; it is `//go:embed`-ed from disk and
  `TestEmbeddedRealityTemplateMatchesDisk` keeps the two in step automatically.
- Template placeholders must contain no `{` or `}`. `formatTemplate`
  (`internal/product/pysem.go:136`) implements Python `str.format` semantics and
  reads a bare brace as a substitution field or an error.
- `make check` is the gate: gofmt + `go vet` + `go test ./...` +
  `examples/acceptance.sh`.
- Gate 1–7 numbering and printed strings are frozen (I5).
- Nothing may write into a materialized slice or invalidate lock hashes (I9).

## Key Decisions

**1. `index.md` carries frontmatter with a new inert `type: index`.**
Consistent with the rule that every document is typed, and it makes indexes
visible to any consumer that filters on `type:`. Cost: a new entry in the
`kind/*` vocabulary (`internal/graph/tags.go:24-28`) and a mandatory same-commit
addition to `skipNames`. Rejected alternative: a bare marker block like
`CLAUDE.md`, which reuses a proven skip-by-name path and adds no type, but
diverges from the typed-document rule for a file adopters will read constantly.

**2. Provenance is additive; approval semantics are untouched.**
`generated:`/`verified:` ship as new advisory fields and a `verified:` entry is
*derived* one-way from `decisionOwner`/`approvedBy` so the trust tier is
meaningful on day one without editing fixture approval data. Rejected: full
reconciliation into `verified:`, which changes what blocks and is the reason D1
was deferred; and add-only with no derivation, which ships a trust tier that
reads "unverified" for every document in every fixture.

**3. The exception `TODO` gap warns, never blocks.**
`internal/governance/declare.go:121` scaffolds `approvedBy: 'TODO: rule owner'`
and it passes validation, while `internal/scaffold/template.go:67` scaffolds
`decisionOwner: TODO` and it hard-fails. Making the second block would fail
`examples/workspace` today, violating I1. A warn line makes the gap visible at
zero compatibility cost. **This is a deliberate departure from the parent
change's R-5.3 ("no new warn line"), which constrained a documentation-only
change and does not bind this one.** Recorded here rather than left implicit.

**4. Index drift reports under gate 5 with an unchanged header.**
Adding a gate would renumber, violating I5. Renaming gate 5's header to mention
indexes would change a frozen printed string, also violating I5. So the header
stays exactly as it is and index lines appear beneath it. The header will
under-describe its contents; that is the accepted cost of the freeze, and the
alternative is worse.

**5. The goldens move, once, deliberately.**
New gate-5 lines and three new fixture files make a golden diff unavoidable. The
parent's R-5.6 froze them because its units were documentation-only. Here the
diff is reviewed line by line and re-baselined in the same commit as the change
that causes it. `acceptance.sh --update` is not run on a red build to make a
failure disappear.

**6. `description:` and `index.md` ship together.**
Recorded in the parent's deferral note: `description`'s only consumer is the
index, so shipping the field alone creates a field nothing reads. Its quality is
also unreviewable without a rendered index — the test for a bad description is
whether it still reads true pasted onto a sibling, which only works side by side.

## Out of Scope

- Reconciling `decisionOwner`/`approvedBy` into `verified:` (parent D-1.4).
- Making the exception `TODO` a blocking failure.
- `stale_after:`, OKF `sources[]`, Attested Computation.
- Indexes under `knowledge/`.
- Any change to gate numbering or to gates 1–4, 6, 7, 8.
- The `examples/banking/` fixture, which the acceptance harness does not exercise.
- Making any of `title`, `description`, `resource` required.

## Open Questions

Carried from D-2.8 and answered here except where noted.

| Question | Resolution |
|---|---|
| Does `index.md` carry frontmatter, and what `type`? | Yes; `type: index`, inert (no required-field gate) |
| A qualifying directory with no index | Gate 5 reports drift; `derive` creates it |
| A pre-existing hand-written `index.md` | **Resolved in pre-mortem.** Follow the settled `CLAUDE.md` precedent: leave it alone, report hand-owned at SevOK, pass (`internal/graph/node.go:126-145`, `internal/graph/gates.go:291`). Failing would add a blocking check and violate I1. |
| Index removal when a directory drops below threshold | `derive` deletes the generated file; a marker-less file is left alone per the row above |

## Pre-mortem findings folded in

- **R-2.10 corrected** — see the row above. The first draft would have added a
  blocking check and contradicted shipped behaviour.
- **`init` → `validate` is not covered by the harness.** `examples/acceptance.sh`
  never runs `company-os init` followed by `validate`; its only `init` is
  `git init` for the federated source fixture (`:171`). The D-2.2 failure mode —
  `init` emitting a workspace that fails its own validate — would therefore ship
  undetected. The double-build check (`s0 == s1 == s2`, `:112-130`) covers the
  committed fixtures only. A harness case is required, not optional.
- **`type: index` earns nothing internally.** R-2.5 puts `index.md` in
  `skipNames`, so `IterGraphDocs` never yields it, so no `tags:` are derived for
  it and gate 4 never checks it. The type is for external consumers only. This is
  an accepted cost of decision 1, recorded so it is not later mistaken for a bug.
