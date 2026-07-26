---
type: lld
id: lld-go-cli-tui-port
title: Go CLI + TUI Port — Low-Level Design
status: draft
tags: [kind/lld, status/draft]
---

# Go CLI + TUI Port — Low-Level Design

Source of truth for intent: `docs/hld/go-cli-tui-port.md`.
Behavioral requirements: `docs/ears/go-cli-tui-port.md`.
Research basis: `.devlocal/research/2026-07-26-cli-tui-and-agent-interface.md`.

Reviewed by Technical Lead, who verified the YAML risk empirically against
Go 1.25.7 + `gopkg.in/yaml.v3` v3.0.1 and the committed fixtures. That
verification changed this document materially — see "The YAML problem" below.

## Architecture

### The central rule: records before renderers

The Python CLI destroys structure at the point of computation, not at the point
of printing. `cmd_validate` (`bin/company-os:922-1107`) tracks findings in a bare
integer — `problems = 0` at `:923`, `problems += 1` at 15 sites — and prints an
English sentence at each one. `gather_prd_governance` (`:551-570`) builds a
markdown checklist string. `skill_conflicts` (`:837-866`) computes `skill`,
`conflicts_with`, and `reason` and immediately concatenates them into prose.
`federated_slice_problems` (`:2490-2526`), `core_field_errors`, `pointer_errors`,
and `identity_errors` all return `[]string` of finished sentences.

The consequence is that `--json` cannot be obtained by intercepting output. This
is the design constraint that shapes the whole port, and the reason the Go
implementation must be built this way from day one rather than retrofitted.

**The record model is `GateResult`, not a flat `[]Finding`.** A flat slice cannot
reproduce the golden: `examples/golden-validate.txt:11-12` is gate 3's header
followed by zero findings, and a renderer driven only by findings cannot know
gate 3 ran at all.

```go
type Severity int // SevOK, SevWarn, SevFail

type Report struct {
    Root  string // "validating workspace <root>" (:924) — the only underivable piece
    Gates []GateResult
}

type GateResult struct {
    Ordinal  int    // 1..N, rendered as [N/7] or [N/8]
    Slug     string // "ownership-reconciliation", stable, JSON-facing
    Title    string // "ownership reconciliation", the human header text
    Findings []Finding
}

type Finding struct {
    Severity Severity
    Code     string // stable machine identifier, "ownership.agrees"
    Subject  string // render-ready prefix token; see the prefix table
    Path     string // workspace-relative path, "" when not file-scoped
    Message  string // sentence body, no prefix, no indentation
    Fields   Fields // consumed by BOTH renderers, not JSON-only
}

type Fields map[string]any // Str/Int/Strs accessors, never panicking
```

**Two corrections from task 1.7, which measured this sketch against all five
goldens.**

*A bare `[]GateResult` is not the top of the model.* The oracle's first line is
`validating workspace <root>` (`:924`), and `Root` cannot be derived from the
gates. `Report` wraps them. Everything else stays derived rather than stored: the
`[N/M]` denominator is `len(Gates)` (gate 8 exists only in federated mode,
`:930`), and the `FAIL — N problem(s)` trailer is a count of `SevFail` findings —
warnings do not count, which
`examples/failing-workspace-golden-validate.txt` pins at 15 fails, 4 warns, and a
trailer reading 15.

*`Fields` cannot be `map[string]string`.* Two independent measurements rule it
out. `:990` renders an **ordered list** inside the sentence — `missing
frontmatter ['team', 'components', 'governanceSnapshot']` — and its element order
is what the human reads. And the counts in `[ok] communications: feature-index in
sync (1 component(s))`, `[ok] skills layered cleanly (2 canonical, 0 team; no
shadowing or dangling extends)`, and `[ok] federated slices match
workspace.lock.yaml (4 file(s) across 1 repo(s); no hand-edits)` must reach the
human output as numbers and JSON as `1`, not `"1"` — R-2.2 says *typed* fields.
`map[string]any` carries both. Key order is meaningless (the renderer addresses
fields by name; `encoding/json` sorts map keys, so JSON stays deterministic);
order *within* a slice value is load-bearing.

### The per-gate prefix policy

There is no uniform prefix rule in the Python CLI, and pretending there is one is
how a records refactor breaks the golden silently. The renderer carries an
explicit policy per gate:

| Gate | Line shape | `Subject` value |
|---|---|---|
| 1 ownership reconciliation | **three shapes** | team id (`:941`), **single-quoted** component id (`:946`), bare component id (`:951`) |
| 2 deviation/exception expiry | `<team>: <msg>` | team id |
| 3 active PRD contracts | `<platform>/<prd-id>: <msg>` | compound |
| 4 frontmatter core + tags | `<path>: <msg>` | `Path` |
| 5 CLAUDE.md node drift | **three shapes** | `<root>/team.yaml` (`:1025`), bare `<root>` (`:1029`), `<root>/CLAUDE.md` (`:1033`, `:1037`, `:1041`) |
| 6 feature-index drift | `<platform>: <msg>` | platform id |
| 7 skills layering | `<msg>` — no prefix | empty |
| 8 federated slice integrity | `<msg>` — no prefix | empty |

**Corrected by task 1.7.** The earlier version of this table said gate 1 was
uniformly `<component>: <msg>`. It is not: `:941` prefixes the *team* id, and
`:946` wraps the component id in single quotes — `[FAIL] 'svc-alpha': team
'ghost' claims accountable …`. Gate 1 has exactly the three-shape property this
table attributed uniquely to gate 5.

The same task removes the "per-code, not per-gate" prefix source for gate 5.
There is no renderer discriminator, per gate or per code. All seven shapes are
seven distinct `Subject` **values** under one uniform rule: emit `Subject`, then
`": "`, then `Message`, and `Message` alone when `Subject` is empty. The
consequence is that `Subject` is render-ready text and may carry punctuation, so
the clean machine value must **also** live in `Fields` — that is what JSON
consumers read.

Three further shapes, all of which turned out to need no field of their own:

- **Blank lines are a property of the gate header**, not of findings. Every gate
  header is printed as `\n[N/M] title` except gate 1, which has no leading blank.
  That is exactly `Ordinal > 1`, so it is derived, not stored.
  `examples/failing-federated-golden-validate.txt:3-4` confirms the blank
  survives a gate with zero findings.
- **Gate 4's `[ok]` is conditional on the absence of core errors** (`:1003-1008`):
  a document with core-field errors emits its `[FAIL]` lines and **no** `[ok]`
  line. Gate 6 has the same structure (`:1058`, and the `unresolved` branch).
  Neither needs model support — "this gate produced failures and deliberately no
  ok line" is already the absence of a record, and the producer simply does not
  append one.
- **The `warn` site at `:1013` is a loop, not a multi-line finding.** One
  document with four bad pointers produces four separate one-line findings
  (`examples/failing-workspace-golden-validate.txt:27-30`). The real constraint it
  imposes is on `GateResult.Findings`: those four warns sit between that
  document's `[ok]` and the next document's `[FAIL]`, so findings must stay one
  ordered slice in document order and can never be bucketed by severity.

### The YAML problem — three distinct risks, ranked by real severity

The pre-mortem named one YAML risk and ranked it first. Empirical verification
found it real but narrow, and found two others that outrank it.

**Risk A (highest) — scalar resolution: PyYAML is YAML 1.1, `yaml.v3` is YAML 1.2.**
This affects *parsing*, so it produces wrong behavior, not merely different bytes.
Verified:

```
examples/workspace/teams/customer-engagement/governance/deviations.yaml:8
  reviewDate: 2027-01-15        (unquoted)

PyYAML   → datetime.date  → str()  → "2027-01-15"
yaml.v3  → time.Time      → fmt    → "2027-01-15 00:00:00 +0000 UTC"
```

Golden line 7 becomes `(review 2027-01-15 00:00:00 +0000 UTC)`. Line 10 of the
same fixture has `reviewDate: '2027-01-14'` **quoted**, so both types must be
handled by one code path. The class is wider than dates — YAML 1.1 also resolves
`yes`/`no`/`on`/`off` as booleans, sexagesimals, and leading-zero octals.

**Risk B — emitter byte signature.** Measured against the committed artifacts,
the divergence set is exactly four mechanical deltas:

| Delta | PyYAML | yaml.v3 |
|---|---|---|
| Block sequence indent | `- item` at parent indent | `  - item`; `SetIndent` cannot separate the two |
| Quote style | `'3.0'` single | `"3.0"` double |
| Plain-scalar folding | wraps at width 80, 2-space continuation | no fold |
| Key sort, unicode handling | — | matches |

The pre-mortem claimed this fails gates 4, 5, 6 and both goldens. **It does not.**
Gate 4 (`:1002-1005`) compares parsed, sorted tag lists. Gate 5 (`:1035`) compares
a markdown block. Gate 6 (`:1053`) is
`canonical_yaml(committed) != canonical_yaml(fresh)` — both sides pass through the
same emitter, so the signature cancels algebraically. Neither golden contains a
byte of emitted YAML. And `rewrite_frontmatter_tags` (`:1352`) early-returns on a
semantic list compare, so in-sync `tags:` blocks are never rewritten.

The one real exposure is `examples/acceptance.sh:76-89`, the double-build no-op:
`snapshot()` shasums the fixture tree and asserts `s0 == s1`, where `s0` is
Python-emitted bytes. `write_feature_indexes` (`:1530-1537`) guards on
`out.read_text() != new` — a **byte** compare — so Go rewrites every
`feature-index.yaml` on first build and `s0 != s1`.

**Mitigation is four lines, not a custom emitter.** Change that guard to a
semantic compare (`canonical_yaml(load(out)) != new`), matching what gate 6
already does. Go then leaves Python-emitted files alone, `s0 == s1` holds, no
re-baseline is needed, and the system becomes permanently immune to emitter drift
— including a future `yaml.v3` upgrade, an exposure Python had too and nobody
noticed. This is a behavior change and takes an explicit R-0.7a carve-out.

A PyYAML-compatible custom emitter is feasible (~200 lines for this restricted
data shape; the folding algorithm is the only fiddly part) but is **not
warranted**. Nothing requires byte-parity with PyYAML once the guard is semantic.

**Risk C — Go map iteration is randomized, and it reaches byte-frozen output.**
Python dicts are insertion-ordered and `safe_dump(sort_keys=False)` preserves
authored order. Three concrete sites:

- `workspace.lock.yaml` (`:2614`) embeds a `files: {path: sha256}` map. Under Go
  it re-emits in random order every sync — unreviewable diffs, and it breaks the
  lock's own reproducibility claim.
- **Gate 8 renders finding order from map iteration**: `for rel, want in
  (lr.get("files") or {}).items()` (`:2521`). Go randomizes the order of `[FAIL]`
  lines in `validate` stdout.
- Gate 6's `feature_index_unresolved` (`:1519`) has the same shape.

Every map-driven emission and every map-driven finding loop is explicitly
ordered — but **not by sorting**. Task 1.4 measured all three sites against the
committed fixtures and the prescription "explicitly sorted" is wrong for two of
them:

- The `files:` map is built by `_materialize_all` (`:2454`) as a nested walk —
  manifest slice order, then each slice's `paths:` list order, then
  `sorted(src.rglob("*"))` *within* each subtree — and emitted with
  `sort_keys=False`. `examples/federated/workspace.lock.yaml` records
  `governance/…` before `components/…`, which is the manifest's `paths:` order
  and the reverse of alphabetical. Rebuilding it through the real `hash_tree`
  reproduces the committed file exactly; a global `sorted()` does not.
- Gate 8 iterates `safe_load` of that lock, so its `[FAIL]` order is the lock's
  DOCUMENT order. `examples/failing-federated-golden-validate.txt` freezes
  `governance/requirements.yaml` before `components/svc-sliced.yaml`; swapping
  the two lines in a copy of the lock swaps the two `[FAIL]` lines. Sorting here
  breaks the golden.
- Only gate 6 is sorted, and incidentally — `build_feature_index` (`:1440`)
  iterates `cids = sorted(...)`, so insertion order happens to equal sorted
  order. The loop still does not sort.

Two further orderings share the same key set and must not be conflated:
`aggregate_hash` (`:2436`) runs `sorted(files)`, a plain **string** sort, while
`sorted(src.rglob("*"))` is CPython's **component-wise** `PurePath` order, under
which `sdd/adr/a.md` precedes `sdd/adr-x.md` and Go's `sort.Strings` does not
apply. `internal/yamlio` supplies `MapPairs` (document order), `OrderedMap`
(insertion order, `dict` assignment semantics) and `PathLess`/`SortPaths`
(`PurePath` order, differentially tested against CPython).

### `internal/yamlio` — `yaml.Node` end to end

There is no struct-unmarshal path anywhere in this codebase. Three independent
reasons, and all three are mandatory:

1. **Preserve-unknown** (`examples/selftest.py:44-50`, R-1.5). Unknown frontmatter
   keys must survive a tag rewrite. Struct unmarshal drops them.
2. **Authored key order.** See Risk C.
3. **Style and quote fidelity** on read-modify-write of hand-authored files —
   `deviation declare` (`:1121`), `exception request` (`:1136`).

One behavior inversion falls out and needs a ruling rather than a silent
improvement: PyYAML's `safe_load` drops comments, so `deviation declare` destroys
them today. `yaml.Node` preserves them. Go would be *better*, which R-0.7
forbids. Comment preservation is added to the R-0.7a carve-out list as a
sanctioned improvement.

### Package layout

```
company-os-starter/
  go.mod                        module github.com/<org>/company-os
  cmd/company-os/main.go        arg parsing, dispatch, exit-code mapping, --version
  internal/yamlio/              yaml.Node helpers, YAML-1.1 scalar resolution,
                                canonical emit, deterministic map ordering
  internal/workspace/           root resolution, path helpers, MANIFEST_NAME/LOCK_NAME
  internal/frontmatter/         the ^---\n…\n---\n parser
  internal/model/               GateResult, Finding, Severity, exit codes
  internal/governance/          resolve, explain, deviations, exceptions
  internal/product/             discover, prd, check
  internal/graph/               tags, feature-index, CLAUDE.md nodes, rebuildGenerated
  internal/skills/              four-layer merge, shadowing, extends
  internal/scaffold/            init, add, reality, scratchpad  (+ embedded templates)
  internal/federation/          manifest, sparse-checkout, slices, lock
  internal/ids/                 the canonical ID registry, `ids list`, difflib suggestions
  internal/roles/               the role glossary, `today`
  internal/validate/            the 7/8 gates, each returning GateResult
  internal/render/              text.go, json.go, ids.go, today.go, glossary.go
  internal/tui/                 Phase 2: Bubble Tea program, screens, command preview
  templates/                    //go:embed
```

Four structural rulings the naive port gets wrong:

- **`Workspace` is not a mechanical transliteration of `:211-263`.** `require_root`
  (`:230`), `platform_dir` (`:238`), and `team_dir` (`:244`) all call `die()`.
  R-2.5 forbids any package below dispatch from exiting or writing to stdout, so
  all three become error-returning and the change ripples into every caller.
- **`MANIFEST_NAME` and `LOCK_NAME` live in `internal/workspace/`, not
  `internal/federation/`.** `Workspace.is_root` (`:221`) reads `MANIFEST_NAME`
  while `internal/federation/` needs `internal/workspace/` for path resolution.
  Putting the constant in `federation` produces an import cycle that will not
  compile. `is_root` needs the constant, not `load_manifest`.
- **`rebuild_generated` (`:1803`, 6 call sites) lives in `internal/graph/`**, with
  a one-way `scaffold → graph` dependency. It is the mandatory bridge between the
  write path and the derive path, and research §4c flags it as the one function
  that resists any split. Placing it in `scaffold` — the natural wrong guess —
  creates a cycle.
- **`ids` and `today` are two packages, not one, and neither is named after its
  command.** *(Added by task 2.8; the layout above listed neither.)* The Python
  groups both under one comment banner (`:1208`, "ID registry & role views"),
  which invites a single package, and the task-0.4 inventory guessed
  `internal/ids` + `internal/today`. Both guesses are wrong for the same measured
  reason: `role_glossary_lines` (`:1260`) has **two** callers, `cmd_today`
  (`:1171`) and `cmd_ids` (`:1277`), while the registry has **three** callers in
  other clusters — `suggest_ids` from `governance explain` (`:365`) and
  `register_id` from `init`/`add` (`:1950`, `:1951`, `:2008`, `:2015`, `:2025`).
  One package would force `governance` to import the role glossary to reach the
  registry; a package called `today` could not hold the glossary without
  `internal/ids` importing `internal/today`, which inverts what the two things
  are. Splitting on the two fan-in sets and naming the second for the *role*
  concept rather than the `today` command gives a one-way `ids → roles` edge and
  leaves each shared function in the package its callers actually want.
  `register_id` belongs in `internal/ids` when it lands, guarded per R-0.7c, on
  the same `scaffold → ids` reasoning as `rebuild_generated`.

The remaining cluster boundaries follow the coupling in research §4c: federation
is most self-contained (2 external callers), skills next (2), then
governance/product, then scaffolding, with graph/tags having the highest fan-in.
`validate` reaches into six clusters and sits above all of them.

### Command surface — unchanged

16 subcommands, ported verbatim from `main()` (`bin/company-os:2661-2781`):

| Command | Actions / args |
|---|---|
| `init` | `--company --team --platform` |
| `add` | `platform\|team\|component <name> --platform` |
| `reality` | `new <component> --platform` |
| `discover` | `new\|validate --team [title\|id]` |
| `prd` | `new\|validate\|complete [id] --team --platform --components --title --from-discovery --force` |
| `governance` | `resolve\|explain [component] --team` |
| `check` | `ready\|done --team --components` |
| `validate` | (no args) |
| `deviation` | `declare <rule> --team --rationale` |
| `exception` | `request <rule> --team --component --expires --reason` |
| `scratchpad` | `init --repo` |
| `today` | `--role developer\|team-lead\|product-owner\|architect\|vp-engineering\|director-of-product` |
| `graph` | `build` |
| `ids` | `list --team --platform --prefix --role` |
| `skills` | `list` |
| `workspace` | `sync\|status --frozen --only` |

Plus the global `--root` and the root-requirement exemption for `init` and
`scratchpad` (`:2774-2776`). New in Go: global `--json`, global `--version`, and
in Phase 2 one new subcommand, `tui`.

### The dispatch seam

```go
type Command func(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error)
```

**The `io.Writer` was added during implementation and is unavoidable.** The
original signature could not carry prose, and every mutating command prints prose
— the next-command guidance chain (R-1.8) is not a finding and never will be.
Records still flow back through the return value and are rendered by a per-command
renderer map in `cmd/company-os`; `out` carries only what was already unstructured
narration in Python.

Commands return results and an error. `main` maps them to output and an exit
code. Nothing below `cmd/` calls `os.Exit` or writes to stdout — this is what
makes both the TUI and `--json` possible without a second code path, and it is
the structural difference from Python, where `die()` (`:41-55`) exits from
anywhere in the call tree. Verified count: **52 `die()` call sites plus 5
`sys.exit` statements — 56 distinct failure paths.** (Earlier drafts said 53;
that figure counted `def die(msg):` at `:41` as a call site.)

### Exit codes

**Correcting a premise repeated in three documents.** Research §1.5, and the
first draft of this section, stated that every non-zero path returns `1`. That is
false: argparse already exits **2** for an unknown subcommand, a bad flag, and a
bare invocation. Code 2 documents existing behavior rather than introducing it.

`examples/acceptance.sh:62` tests only zero against non-zero, so distinct
non-zero codes are safe by construction.

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | validation failed — one or more gates reported `[FAIL]` |
| 2 | usage error — bad flags, missing required argument, unknown subcommand (**already today's behavior**) |
| 3 | workspace error — not a workspace root, or a required workspace object does not exist (`:238` platform, `:244` team, `:367` component, `:430`, `:584`, `:636`, `:676`, `:2561`) |
| 4 | artifact error — malformed YAML/frontmatter, schema violation |
| 5 | precondition failed — done-gate refusal, unvalidated discovery brief, deviation aimed at a mandatory rule |
| 6 | external tool error — git missing, git < 2.27, sparse-checkout or clone failure, `--frozen` lock reconciliation failure (`:2564`-`:2594`) |
| 7 | interactive-mode error — `tui` or a wizard prompt with no TTY (`:1961`, `_prompt`) |
| 8 | conflict — refusing to overwrite an existing artifact (`:417`, `:610`, `:1797`, `:1971`, `:2037`) |

`1` stays reserved for `validate` failing its gates, which keeps every existing
CI gate reading identically. Not-found lookups resolve to **3**, not 2 — the
invocation was well-formed; the workspace did not contain the object.

The full classification of all 56 failure paths lives at
`.devlocal/go-port/exit-code-map.md`, produced before implementation rather than
inferred during it. Four findings from that classification change the contract:

- **Code 4's headline case has zero sites today.** All 22 code-4 sites are
  `workspace.yaml` schema violations. `load_yaml` (`:58`) and `frontmatter`
  (`:76`) call `yaml.safe_load` with no `try`, so genuinely malformed YAML raises
  an uncaught `YAMLError` and exits 1 through a Python traceback, never through
  `die`. The Go port must **add** those sites in `internal/yamlio`, which is an
  intended divergence the differential harness will surface — R-0.7a(e).
- **Code 5's third example does not exist as an exit site.**
  `resolve_team_governance` (`:317-319`) records `deviationRejected` into
  generated governance and continues; the refusal surfaces as a `validate` gate
  failure (exit 1). `cmd_deviation` validates nothing and always exits 0. The
  contract wording is amended rather than the behavior, because enforcing at
  declare time would break R-9.1 parity.
- **`:601` and `:2021` move from 1 to 2**, a genuine behavior change. They are
  hand-rolled conditional-requirement checks, not argparse, so the claim that
  "code 2 documents existing behavior" holds for argparse only — R-0.7a(f).
- **`:2318` is code 4, not 6.** It sits inside the git block (`_fetch_pinned`)
  but both `git fetch` and `rev-parse` succeeded; the check catches an
  abbreviated SHA in `workspace.yaml`. Artifact fault, not tool fault. Likewise
  `:2547` ("no `workspace.yaml`") is 3, not 4 — the file is absent, not
  malformed, and absence is legal in monorepo mode.

One further contradiction: `:2041` ("reality template not found") **ceases to
exist** under `//go:embed`, which R-1.11 mandates and R-0.7 forbids. It is
R-0.7a(c).

### The TUI (Phase 2)

Bubble Tea (`charmbracelet/bubbletea`) plus Bubbles and Lipgloss.

The four-times-stated "no dependencies" policy (`bin/company-os:13`,
`install.sh:5-8`, `docs/GOLDEN-PATH.md:24`,
`docs/lld/okf-v02-conformance.md:203-205`) is a policy about *runtime*
prerequisites on the user's machine — it exists because the user had to install
Python and PyYAML before running anything. A statically linked Go binary has zero
runtime dependencies regardless of how many modules it compiled from. The policy's
intent is satisfied more completely by this port than by the status quo, and
build-time Go modules do not violate it.

Structure:

- Model-view-update, one root model holding current screen, workspace, and a
  cached result set.
- **Read-only screens ship first**: workspace overview, `today --role`, validate
  results, component/PRD/discovery browsers, `governance explain`, `skills list`,
  `ids list`, `workspace status`. This subset needs none of the mutation
  machinery and eliminates the entire class of "the TUI wrote the wrong thing"
  defect.
- **Mutating forms ship second, one at a time, each justified by an observed
  request.** `discover new` and `prd new` first — the two a PO actually authors.
  No forms are built for `workspace sync` or `scratchpad init`.
- **Command preview is mandatory on every mutating screen**, and it is **derived
  from the same args structure the command executes**, never a hand-written string
  per screen. A hand-written preview drifts from what runs, which destroys the one
  property justifying interactive mutation at all.
- The TUI calls the same `Command` functions in-process. It does not shell out to
  itself and does not parse its own text output.
- ANSI escapes exist only inside `tui`. Every other subcommand stays ANSI-free —
  the golden files depend on it.

**Cancellation is scoped to pre-execution.** "A cancelled action leaves the
workspace exactly as it was" is unimplementable as a general guarantee: only
`init` is atomic today (staging directory, `:1982`). The TUI guarantees that no
write begins until the previewed command is confirmed. Mid-execution failure
leaves the same partial state the flag CLI would.

### Distribution

`install.sh` (`company-os-starter/install.sh:69-93`) and
`company-os-starter/vendor/` (536K of PyYAML) are deleted.

Release artifacts: `company-os_<version>_darwin_arm64`,
`company-os_<version>_darwin_amd64`, `company-os_<version>_linux_amd64`, built
with `CGO_ENABLED=0`, plus checksums.

**macOS Gatekeeper is a first-class distribution problem, not a footnote.** A
downloaded, unsigned, un-notarized binary carries `com.apple.quarantine` and is
killed on first exec with "cannot be opened because the developer cannot be
verified." The workaround — `xattr -d com.apple.quarantine` — is strictly harder
than `pip install pyyaml` for the audience this change is justified on. Darwin
artifacts are Developer-ID-signed and notarized; if that is not done, the
quarantine workaround ships in the install documentation and the HLD records it
as an accepted cost. Verification is against a **downloaded** artifact, never a
locally built one.

Templates are embedded with `//go:embed`, closing the one place the Python CLI
reads a file from beside the binary (`_builtin_template`, `:526-529`). Discovery
and PRD templates are already module strings (`:522-525`). User overrides are
unaffected: `resolve_template` (`:533-548`) probes only workspace-relative paths.

### Testing

**The differential harness is the parity oracle, not `acceptance.sh`.** The
first draft asserted that `go test` passing plus `acceptance.sh` passing means
parity. That is false twice over: `acceptance.sh` byte-freezes `validate` only,
and `go test` is written against the Go implementation and therefore cannot
detect a Go/Python divergence by construction.

The harness runs both binaries over identical fixtures across all 16 commands and
diffs stdout, stderr, exit code, and the resulting filesystem tree. The Python
deletion gate hangs on the harness.

`examples/selftest.py` carries **86 `check()` call sites, 85 of them real
assertions** (one is a skip sentinel), inventoried at
`.devlocal/go-port/selftest-inventory.md`. Its `SourceFileLoader` mechanism
(`:11-15`) cannot survive any multi-file layout in any language, so it goes — but
not before each assertion has a named Go test. A promise of "coverage for every
package" is not a port.

**The coverage is far narrower than assumed, and that changes what the
differential harness must carry.** Verified: selftest drives exactly **7 of 16**
subcommands through subprocess — `init`, `add`, `reality new`, `prd new`,
`prd validate`, `validate`, `workspace sync`. `discover`, `deviation`,
`exception`, `check`, `governance`, `today`, `graph`, `ids`, `skills`, and
`scratchpad` have **zero** subprocess coverage today. Deleting selftest loses
nothing there because nothing was ever there. The gap belongs to the differential
harness, and it means porting the inventory alone does not satisfy R-7.4.

Distribution skews hard: 38 of 85 assertions (44%) are federation, 14 scaffold,
11 skills, 7 graph. `internal/governance`, `internal/validate`, `internal/render`,
and `internal/model` have zero inherited coverage and are written from scratch.
Fourteen assertions test `die()`/`SystemExit` and must invert to error-returning
under R-2.10; eight test private Python functions and port as unexported unit
tests; three have no Go analogue at all and need an explicit ruling rather than a
silent drop.

The goldens cover only the all-pass path: zero `[FAIL]` and zero `[warn]` lines
across both files, against 15 failure sites in `cmd_validate`. Failure-path
goldens are captured **from Python, before deletion**.

`examples/acceptance.sh` survives unmodified; it shells out to a path (`:12`).

## Constraints

**C1 — `validate` stdout is byte-frozen.** `examples/acceptance.sh:26-56` diffs
full stdout against two goldens; `normalize()` (`:20`) strips only the absolute
path on line 1.

**C2 — gate count is fixture-dependent.** Non-federated workspaces print `[N/7]`;
manifest-bearing workspaces print `[N/8]`. The renderer computes the denominator.

**C3 — flag-complete equivalence.** `docs/ears/golden-path-flavor-federation.md:8-21`
(GPF-R-1.3/1.4) requires every wizard answer to have a flag equivalent, or CI
cannot reproduce it. Phase 2 multiplies the interactive surfaces and this rule
applies to all of them. It also means `init`'s existing `_prompt` wizard
(`:1954-1965`) and the TUI become two interactive paths to the same scaffold, both
bound by GPF-R-1.4 — the TUI delegates to the same scaffold code path rather than
reimplementing it.

**C4 — R-7.4 must be retired, not ignored.** `docs/ears/federation-enrichment.md:149`
mandates keeping "the single-file CLI, the `die/ok/warn/fail` helpers, the
`frontmatter()` parser, and the next-command guidance chain intact," repeated at
`CLAUDE.md:102`, `docs/lld/golden-path-flavor-federation.md:49`,
`docs/lld/federation-enrichment.md:12`, and `docs/lld/okf-v02-conformance.md:204-206`.
No document states why. This change retires the single-file and helper clauses
explicitly and **keeps** the frontmatter and guidance-chain clauses. A project
whose product *is* the methodology must retire its own locked requirement with
the same ceremony it demands of adopters.

**C5 — Windows is not a target.** The POSIX separator normalization at `:1747-1748`
is ported; nothing else about Windows is claimed or tested.

**C6 — no behavior drift**, except the R-0.7a carve-outs.

**C7 — the acceptance oracle expires on 2026-12-31.** `TODAY = dt.date.today()`
(`:31`) is compared against `examples/workspace/teams/customer-engagement/governance/exceptions.yaml:9`
`expires: '2026-12-31'`. On 2027-01-01 golden line 9 flips from `[ok] … valid
until 2026-12-31` to `[FAIL] … expired`, `PASS` becomes `FAIL — 1 problem(s)`,
both goldens break, and both fixtures fail `acceptance.sh` §3. The deviations
follow on 2027-01-14. Today is 2026-07-26: **the parity-gated cutover rests on an
oracle with 158 days of life.** If the port slips past new year, the goldens must
be re-baselined mid-port, creating precisely the "was it the port or the rule?"
ambiguity cited as the reason for deferring OKF. Pushing the fixture dates out and
re-baselining **from Python** is task zero.

**C8 — the drift gates compare semantically and are blind to emitter divergence.**
Gate 6 at `:1053`, gate 4 at `:1002-1005`. The golden files are therefore not
evidence on the emitter axis. See Risk B.

## Key Decisions

**Go over Python and TypeScript.** The deciding input is the stated goal: copy a
binary into `~/.local/bin` and start working. Python cannot deliver it — there is
always an interpreter on the other side. TypeScript is worse: Node is a heavier
prerequisite for this audience, bundling yields 50–100MB against Go's ~10MB, and
there is no `package.json` anywhere in the repo. Precedent supports Go:
`local-search`, the sibling tool `company-os` composes with, is already a single
Go binary with no runtime dependencies. Note that this is precedent for the
ecosystem, not evidence this team can maintain 4500–6000 lines of Go — hence the
named-owner requirement in the HLD.

**Rejected: `zipapp`.** Python's stdlib `zipapp` bundles the CLI plus vendored
PyYAML into one ~600KB executable `.pyz` — roughly a day of work, deleting
`install.sh`, `vendor/`, and the launcher. It is the better trade *if the goal
softens to "users have Python 3.9+."* It does not meet the goal as stated.
Recorded because the cost difference is a day against weeks.

**Rejected: a Go TUI wrapping the Python CLI.** The user still needs Python, so
the distribution goal is not met and the wrapper becomes permanent scaffolding.

**Rejected: keeping Python as a fallback.** Two implementations in parity is a
maintenance tax with no expiry date.

**Rejected: doing the records-then-renderers refactor in Python first.** The right
first move *if the CLI stays Python*. Given a committed port, doing it twice is
waste.

**Reversed by measurement: a PyYAML-compatible emitter is mandatory after all.**
This document previously rejected it, on the reasoning that the four-line semantic
guard removed the exposure permanently. That reasoning holds for *derived*
artifacts and is now R-0.7c. It does not hold everywhere, because task 2.1 found
a write path the guard cannot cover: `register_id` (`:1815`) re-dumps the **whole**
`ids/registry.yaml` through `safe_dump` on every call, and on `examples/workspace`
a single `add` rewrites seven flow-style entries into block style. That output is
compared by the differential harness, so it must match PyYAML's bytes.

An approximation is not sufficient either. "Wrap at 80 columns" is wrong on
`team.yaml`, whose 85-column `precedence:` line PyYAML does **not** fold, because
there is no space past column 80 to break at. The implementation transliterates
`analyze_scalar` and the writer primitives from `vendor/yaml/emitter.py` and is
verified against a live PyYAML oracle.

It currently lives at `internal/scaffold/pyemit.go` only because `internal/yamlio`
was under concurrent edit when it was written. **It belongs in `internal/yamlio`**
— `internal/governance` (`deviation declare`, `exception request`) and
`internal/federation` (the lock) both need it, and neither should import
`internal/scaffold` to get it.

**TUI split into Phase 2, read-only first.** The user chose full interactive
coverage of all 16 commands over the cheaper read-only dashboard. That choice is
kept — the TUI is not cancelled or reduced in eventual scope. What changed after
review is sequencing and ordering: Phase 2 is strictly downstream of the parity
gate, and read-only screens ship before mutating forms. Both reviewers converged
on this independently. `R-5.4`'s original "all 16 subcommands reachable" was an
unbounded requirement dressed as a checkbox — roughly 30 form screens, none
covered by the parity oracle — and it is replaced by an enumerated, testable
screen list.

**OKF Phase 0 carved out of the deferral.** Phases 1–3 move fixture frontmatter
and `validate` output and stay deferred. Phase 0
(`docs/tasks/okf-v02-conformance.md:44-57`) fixes `prd complete`'s done-gate
comparing ISO dates as raw strings — "correct for well-formed ISO dates by lexical
accident, silently wrong for `18/07/2026`, an empty value, or a YAML-parsed
`datetime.date`." It is a live correctness bug in the gate enforcing invariant #4
of the methodology, roughly 30 minutes of work, and it moves neither golden.
Deferring it means R-0.7 compels the Go port to faithfully reproduce a bug whose
fix is already written.

**Two binaries, not one.** `company-os` and `local-search` compose by convention
(`how-it-fits-together.md:100-103`). Merging would couple two release cycles for
no benefit this change needs.

## Out of Scope

- OKF v0.2 Phases 1–3 (Phase 0 is **in** scope).
- Merging with `local-search`.
- Windows builds, Windows CI, `windows/amd64` artifacts.
- Any web or GUI surface.
- New commands, new gates, new validation rules, or reworded human output.
- The PO vocabulary wall — friction F5 (35-40 domain terms) and F6 (canonical ID,
  `@spec`, EARS syntax) stay open and are not addressed by anything here.
- A transition period in which both implementations ship.
