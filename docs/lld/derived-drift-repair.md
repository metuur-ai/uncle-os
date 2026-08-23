---
type: lld
slug: derived-drift-repair
status: locked
created: 2026-08-23
---

# Derived Drift Repair (`validate --fix`) — Low-Level Design

## Architecture

`validate.Run(ws)` today loads the manifest, composes seven or eight gate closures,
runs every one, and returns `[]model.GateResult` with a leading banner. It never
early-returns on a gate failure — a failure is a finding, not an error.

`--fix` adds a second pass around that, entirely inside `internal/validate`:

```
Run(ws)                       -> first result set
  |
  +-- fix disabled ...........-> return as today
  |
  +-- fix enabled
        partition findings by Code
          |
          +-- no repairable code present -> return the first result set unchanged
          |
          +-- repairable codes present
                graph.Rebuild(ws)          (writes)
                Run(ws)                    (second, fresh result set)
                return: banner + repair section + second gate set
```

Three things fall out of that shape.

**Repair is `graph.Rebuild`, not a per-code repair function.** `Rebuild` re-derives
all three: it calls `RewriteFrontmatterTags` over every graph doc (`graph.go:72-77`),
then `WriteFeatureIndexes`, then `writeClaudeNodes`. It is the sanctioned convergence
path and is already exercised by every scaffolding command.

It is **not** the function `graph build` invokes — that is `graph.Build`
(`views.go:37`), which additionally reports each retagged file and appends a summary
section at `Ordinal: 4` (`graph.go:26-55`). Injecting that summary into a `validate`
run would collide with gate 4's ordinal. `Rebuild` is the correct callee, and it
needs one extension: see D9.

**The gates are re-run rather than filtered.** Repairing the feature index can
change which references are unresolved; repairing tags can change nothing else at
all. Only a fresh run states the truth about the repaired workspace. Filtering the
first run's findings would report a state that never existed on disk.

**The repair section is a `GateResult`, not print output.** Nothing below `cmd/`
writes to stdout or calls `os.Exit`. The section carries per-file findings
(`graph.tagged`, `graph.feature-index-written`, `graph.node-written`) so the renderer
prints concrete paths rather than a count.

The renderer does need a change for this, contrary to a first reading.
`render.Validate` partitions on exactly one rule — banner, else gate — and prints
every gate as `[%d/%d]` against the banner-carried total (`text.go:40-53`, `:65`). A
repair section entering that list prints `[0/8]` or `[9/8]`. It needs its own slug
and its own branch, ahead of the gate list, with no `[n/m]` prefix.

### Data flow

| Step | Input | Output |
|---|---|---|
| partition | first `[]GateResult` | set of repairable codes present |
| repair | `ws` | `[]GateResult` from `Rebuild` + any write error |
| re-validate | `ws` | second `[]GateResult` |
| assemble | both | banner, repair section, second gate set |

## Constraints

- **No package below `cmd/` prints or exits.** The repair section must be returned,
  not printed. The exit code is derived by `cmd/` from the returned records exactly
  as it is today.
- **`Run`'s existing signature is load-bearing for the golden tests.** The `--fix`
  entry point is a new exported function; `Run(ws)` keeps its signature and
  behaviour so `golden_test.go` and `json_test.go` stay valid unmodified.
- **The `[n/total]` denominator is carried on the banner, not derived** from the
  gate list. The second run produces its own banner with its own denominator; the
  assembled output must carry exactly one banner, and it must be the second run's.
- **`--frozen` / offline.** Repair performs no network I/O. `graph.Rebuild` is
  AST-and-filesystem only.
- **Writers already no-op when in sync.** `RewriteFrontmatterTags` returns early on
  `PyEqual`; `rewriteGeneratedBlock` goes through `writeIfChanged`;
  `WriteFeatureIndexes` compares with the same `sameCanonical` the gate uses. A
  blanket `Rebuild` on a clean workspace therefore writes zero bytes. This is what
  makes D2 and idempotence free rather than something to engineer.

## Key Decisions

**D1 — The flag goes on `validate`, not on a new `check` verb.**
`company-os check ready|done` already exists and means product readiness for a
component set. Putting `--fix` on `check` would attach a workspace-wide repair to a
command that is scoped to `--team` and `--components`. The backlog's shorthand
`uncle check --fix` describes a future verb surface, not today's CLI.

**D2 — Repair via `graph.Rebuild`, not targeted per-code repair.**
Rejected alternative: repair only the files named in the failing findings. It is
more code, it re-derives values a second way, and it can disagree with the gate.
`Rebuild` cannot disagree, because `WriteFeatureIndexes` and `FeatureIndexGate`
share `sameCanonical` — the writer's "should I write" test and the gate's "has it
drifted" test are the same function. Blanket rebuild costs nothing on the untouched
majority because the writers no-op.

**D3 — Re-run the gates after repair; do not mutate the first result set.**
See Architecture. The corollary is that `--fix` runs the gate set twice on a drifted
workspace and once on a clean one. Gate runs are filesystem walks over a workspace,
not a cost worth optimising against correctness.

**D4 — Exit code comes from the second run, unmodified.**
Repaired everything → the second run is green → 0. Repaired some things but a
decision-class finding remains → the second run fails → `ExitValidation` (1). A
`--fix` invocation that parsed is never `ExitUsage` (2); the invocation was
actionable. The banner from the aborted-run case (`complete: false`) keeps its
existing meaning.

**D5 — A write refused by a read-only slice is a named finding, not a raw
`cannot write` error — and this requires a prerequisite change to `internal/model`.**
Today every writer wraps a failed `os.WriteFile` into
`model.Errorf(model.ExitArtifact, "cannot write %s: %v", …)`. Under `--fix` that
surfaces to a vault editor as an opaque permission error on a file they are not
supposed to know exists.

Detecting the permission class is not currently possible. `model.Errorf` formats with
`fmt.Sprintf`, not `%w`, and `model.Error` has no `Unwrap` — only `QuietError`
(`model.go:283`) and `UsageError` (`model.go:331`) do. `os.IsPermission` on any
writer error returns false, always.

So this ships in two steps, and the second is not startable before the first:

- **P1 (prerequisite):** give `model.Error` an `Err error` field and an `Unwrap`,
  and have the three graph writers pass the `*os.PathError` through. Verify
  `model.CodeOf` still resolves. Land alone, with tests.
- **P2:** the repair path inspects the unwrapped error, and emits a finding naming
  the workspace-relative path plus the remedy — change the source repo, bump the
  pin, re-sync. The file is left unmodified.

**D5a — the repair pass aborts at the first refused write; it does not continue.**
Continuing would mean converting three writer loops from `(result, error)` to
accumulate-and-continue (`graph.go:73-75`, `node.go:457-459`, `featureindex.go:439`).
That changes what `graph build` and every scaffolding caller do on partial failure —
behaviour no fixture covers. A read-only slice is a single structural cause; naming
the first file names the cause. Not worth the blast radius.

**D6 — `knowledge/CLAUDE.md` is not a hazard, and this is already guaranteed.**
Two existing guards, verified in source, mean the catalog root's context node stays
writable. `SliceRel` refuses a bare `knowledge/` target precisely because
"targeting it bare would chmod that node 0444 via `makeReadonly`, after which graph
build cannot rewrite it." And `hashTree` is scoped to the slice's `paths`
allowlist rather than walking the target, precisely so "a derived artifact that
appeared under the target (a `generated/` aggregate, a `CLAUDE.md` node) would
[not] be absorbed into the lock and frozen 0444." Repair inherits both. D5 therefore
covers the remaining real case, which is a slice whose `localDirectory` lands under
`company-os/`, `company-ontology/`, `platforms/` or `teams/` — all of which
`SliceRel` permits at depth 1 or deeper, and which `makeReadonly` does chmod.

**D7 — The repairable code set is a declared constant with a structural guard.**
The set is exactly `{CodeTagsDrift, CodeNodeDrift, CodeFeatureIndexDrift,
CodeNodeAbsent, CodeFeatureIndexAbsent}` — every condition `Rebuild` writes for, per
the HLD's Overview. It is declared once, and a test asserts that no drift-named code
in `model/codes.go` falls outside that set without appearing in the decision-class
allowlist (D11). A future gate that adds a derived condition then fails that test and
forces an explicit derived-versus-decision call, rather than silently defaulting into
"never repaired." This is the same guard shape used for the advisory rule in the
retrieval adapter spec.

**D8 — `--json` emits one document containing the repair section and the
post-repair gates.** Not two documents, not the pre-repair set. A machine consumer
asking "is this workspace green" must get the answer for the workspace as it now
exists on disk.

**D9 — `Rebuild` is extended to report retagged files, and `Build` is re-expressed
in terms of it.** `Rebuild` discards the `wrote` boolean today (`graph.go:73`), so it
cannot say which documents it retagged; only `Build` builds that section
(`graph.go:26-43`). Repair must report concrete paths, and must not re-derive that
list by walking the workspace a second time — a second derivation can disagree with
what was written. `Rebuild` therefore gains the tag section, and `Build` becomes
`Rebuild` plus its summary record. This keeps one derivation and leaves `Build`'s
output byte-identical, which the graph goldens already pin.

**D10 — nothing converts a hand-owned `CLAUDE.md`; the guard lives in
`rewriteGeneratedBlock`, not in the repair path.** Gate 5 calls a marker-less node
`node.hand-owned` at `SevOK` and continues (`gates.go:157-160`); the writer used to
append a generated region to it anyway. The first draft of this design put a guard
in the repair path — collect the roots gate 5 reported as hand-owned, skip those
nodes — on the reasoning that changing the writer would change `graph build` for
every existing caller.

That was reversed. Changing `graph build` is the point: it converts hand-owned files
today, silently, and the gate has always said it should not. A repair-path guard
would have fixed the symptom on one path, left the other path converting files
behind the user's back, and forced `--fix` to consult first-run findings to decide
what to skip — coupling the repair pass to the report. Instead `rewriteGeneratedBlock`
returns the input unchanged when no markers are present (R-2.11), so `Build`,
`Rebuild`, and the repair pass all inherit the refusal from the single function they
share. The repair path holds no list of hand-owned roots and needs none, and the
equivalence claim in SC5 holds without a carve-out.

**D11 — the repairable-condition set is not a suffix match.** `CodeSliceSetDrift =
"federation.slice-set-drift"` (`codes.go:81`) is a decision-class code that ends in
`-drift`; it passes a naive `.drift` suffix test by punctuation luck. The D7 guard
matches case-insensitively on `drift` anywhere in the code string, against an
explicit allowlist of known decision-class drift codes, so a future
`federation.lock-drift` or `governance.resolve-drift` cannot slip through silently.

## Sequencing

Unit 0 is not optional preparation — two clauses elsewhere are unimplementable
without it, and neither is visible to a user.

1. **Unit 0a** — `model.Error` wraps and unwraps; writers pass the `*os.PathError`
   through. Rendered messages unchanged. Lands alone.
2. **Unit 0b** — `Rebuild` reports retagged files; `Build` re-expressed on top of it,
   output byte-identical. Lands alone.
3. **Unit 1–2** — flag, partition, repair, re-run. This is the first commit a user
   can observe.
4. **Unit 3** — permission classification and the read-only finding.
5. **Unit 4** — repair section, renderer branch, new golden fixture.

## Test surface this needs and does not have

`examples/acceptance.sh` has no `--fix` coverage and cannot grow it by accident.
Three fixtures are required, and each maps to a Success Criterion:

- clean workspace under `--fix` — zero writes, exit 0 (SC3)
- drifted workspace under `--fix` — repair, green second run, exit 0 (SC1)
- workspace with a drifted file inside a `0444` slice target — named finding,
  file unmodified, non-zero exit (SC4)
- workspace containing a marker-less `CLAUDE.md` — byte-identical under both
  `--fix` and `graph build && validate` (SC5)

The third needs a fixture that runs `workspace sync` and then mutates a slice, which
no existing fixture does. The fourth is cheap: drop a marker-less `CLAUDE.md` into
the drifted fixture and assert it survives both paths unchanged.

## Out of Scope

- The governance-resolve drift gate (prerequisite; separate change).
- Any renaming of `validate` or `check`.
- Link escaping in node emission sites.
- Repairing gate 8, gate 1, gate 2, gate 3's decisions, gate 4's core-field errors,
  gate 5's identity errors, gate 6's unresolved references, or gate 7.
- A `--dry-run` that shows what would be repaired. `validate` without `--fix`
  already is that.
