# `cos-` Skill Namespace and Claude Code Plugin — Low-Level Design

> **STATUS: DEFERRED — specified, not scheduled.**
>
> This change was scoped against a federation driver that does not hold: slice
> targets are disjoint by validator rule, so the merged skills directory it
> assumed cannot exist, and every `.SKILL.md` a workspace can hold is a Company
> OS artifact by construction — leaving nothing inside a workspace for a prefix
> to disambiguate against.
>
> The value the prefix does carry belongs to a shared agent namespace, and is
> delivered by [the plugin](company-os-plugin.md). The starter-kit repair that
> was folded in here is now [its own change](starter-kit-skills.md) and ships
> first.
>
> The units below still contain the starter-kit and plugin material, now
> superseded by those two specs — read them there, not here. What this document
> uniquely still holds is the workspace-side rename and its enforcing gate. It is
> kept because the analysis is sound and expensive to redo — five golden
> regenerations, eighteen regenerated context nodes, a gate insertion moving
> federated slice integrity to `[9/9]`, a counter sweep across every tracked
> file, and a federated fixture lock the CLI provably cannot regenerate. Revisit
> it if a real mixed-namespace collision appears in a real workspace. Today the
> corpus says one structurally cannot.

## Architecture

The change has four parts. The first three are data and documentation edits
plus one new validation gate; the fourth adds a directory tree that no existing
code reads.

### Part 1 — the three name slots

A skill's identity lives in three independently-editable places. All three take
the prefix, and they must move together or the skill half-resolves.

| Slot | Before | After | Consumed by |
|---|---|---|---|
| file stem | `creating-prd.SKILL.md` | `cos-creating-prd.SKILL.md` | `Name()` (`internal/skills/skills.go:164-170`) → gate-7 name shadowing (`conflicts.go:127`); `ResolveExtends` path construction (`skills.go:423-424`); `skills list` `name` field |
| `id:` | `skill://product/creating-prd` | `skill://product/cos-creating-prd` | gate-7 id shadowing (`conflicts.go:124`); generated `CLAUDE.md` doc-index title (`internal/graph/node.go:319-327`) |
| `extends:` | `platform-skill://<p>/creating-prd` | `platform-skill://<p>/cos-creating-prd` | `extendsRE` (`skills.go:398`) → file path, not id |

No production Go changes are needed for the rename itself. `Name()` strips a
fixed `.SKILL.md` suffix and returns whatever precedes it. `extendsRE` captures
`(.+)` after the platform segment and appends `Suffix`. Both carry a prefixed
name through unchanged. Gate 7's shadowing comparisons are equality tests
between two skills' names or ids; a uniform prefix shifts both sides equally
and changes no verdict.

### Part 2 — files that move

Eleven shared skill files exist. Seven are already flat; four are in the
undiscoverable nested layout and are flattened as they are renamed.

Flat, rename in place:

```
examples/workspace/company-os/skills/syncing-knowledge.SKILL.md
examples/workspace/platforms/communications/skills/creating-prd.SKILL.md
examples/federated/platforms/communications/skills/creating-prd.SKILL.md
examples/banking/bank/repos/platform-payments/skills/creating-prd.SKILL.md
examples/failing-workspace/company-os/skills/creating-prd.SKILL.md
examples/failing-workspace/teams/ghost/skills/creating-prd.SKILL.md
examples/failing-workspace/teams/ghost/skills/reviewing-prd.SKILL.md
```

One of those seven is not an ordinary file.
`examples/federated/platforms/communications/skills/creating-prd.SKILL.md` is a
materialized federation slice: mode `0444`, its bytes hashed into
`examples/federated/workspace.lock.yaml`, and its path listed in the manifest's
`paths:` allowlist. Renaming it changes both the lock's path key and its hash,
so the fixture's lock must be rewritten as part of the rename, not left to
drift.

Nested, flatten and rename:

```
company-os-starter/skills/completing-a-change/SKILL.md      -> skills/cos-completing-a-change.SKILL.md
company-os-starter/skills/creating-prd/SKILL.md             -> skills/cos-creating-prd.SKILL.md
company-os-starter/skills/requesting-an-exception/SKILL.md  -> skills/cos-requesting-an-exception.SKILL.md
company-os-starter/skills/running-discovery/SKILL.md        -> skills/cos-running-discovery.SKILL.md
```

The four starter-kit files are reference copies with no reader.
`templates/embed.go` embeds exactly one file, `reality-component.md`, and states
that its siblings "are human reference copies with different placeholder text
that the CLI has never read". `company-os-starter/skills/` is in the same
position: nothing embeds it, `internal/scaffold` never copies it, and discovery
globs `<workspace-root>/company-os/skills/`, which `company-os-starter/` is not.
Flattening them is a consistency fix so that copying one into a workspace
produces a working file. It is not a defect repair, and it changes no runtime
behavior.

The one live `extends:` is
`examples/failing-workspace/teams/ghost/skills/reviewing-prd.SKILL.md`, whose
`platform-skill://alpha/no-such-base` is a deliberately dangling target. It
becomes `platform-skill://alpha/cos-no-such-base` so the fixture still exercises
a dangling reference against a plausible name.

### Part 3 — the prefix gate

A new gate walks the same `Discover()` result gate 7 uses and fails any shared
skill whose file stem or `id:` name segment lacks the `cos-` prefix. It lives
in `internal/skills/` beside `conflicts.go`, is wired into
`internal/validate/validate.go` after the existing skills gate, and emits under
a new `skills.` finding code.

Gate numbering is a caller-supplied parameter (`skills.Gate(ws, 7)`), and the
rendered `[N/M]` totals are computed rather than hard-coded, so the counts move
from `[N/7]`/`[N/8]` to `[N/8]`/`[N/9]` across the five goldens.

**Placement is fixed, not "last".** `steps` in `internal/validate/validate.go:82-95`
is a seven-element slice, and the federation gate is appended to it only when a
manifest exists. Appending the new gate after that block would give it ordinal 8
in a monorepo and 9 in a federated workspace — a gate with no stable number,
which no golden or doc could describe. It is therefore inserted at index 7,
immediately after the skills-layering gate. Two consequences follow and are
accepted: the skills-layering gate keeps ordinal 7, and **federated slice
integrity moves from `[8/8]` to `[9/9]`**.

The skills-layering gate's ordinal is the first character of all 22 frozen
answers, which is why it must not move.

The gate composes its own finding text through the existing exported
`skills.Message(code, Fields)` switch (`conflicts.go:221-237`), which gate 7 and
`skills list` already share. Adding a case is additive and cannot reach gate 7's
three existing cases. Setting `Message:` on its own findings is also what keeps
the gate inside the AST-enforced rule that nothing under `internal/` prints or
exits (`cmd/company-os/architecture_test.go:24-42`) — `internal/render` has no
validate-specific dispatcher to extend.

Scope of enforcement:

- **In:** every file discovered by the `*.SKILL.md` glob at the company,
  platform, and team layers, whatever its `authority:`.
- **Warned, not failed:** a skill whose path falls inside a locked slice target.
  `validate.Run` already holds the manifest, and a consumer has no write access
  to a `0444` slice, so failing it would create a state with no legal local
  remedy. `SevWarn` findings are excluded from `Problems()`
  (`internal/model/model.go:148-160`), so the consumer stays green while the
  source repo's own `validate` fails the same file at full severity. Enforcement
  lands on whoever can act on it.
- **Out:** personal rules under `scratchpad/personal-rules/`. They are
  git-ignored, never enter a federated slice, and are already excluded from
  gate 7's totals.
- **Out:** `knowledge/` slices. Gates `[1/N]`–`[7/N]` already skip that root
  because foreign docs carry no `type:` frontmatter; the new gate skips it for
  the same reason.

### Part 4 — the Claude Code plugin

New tree at the repository root, matching the documented default locations:

```
.claude-plugin/plugin.json          plugin manifest, name "company-os"
.claude-plugin/marketplace.json     marketplace entry pointing at this repo
skills/cos-creating-prd/SKILL.md    Claude Code skill wrappers
skills/cos-running-discovery/SKILL.md
skills/cos-completing-a-change/SKILL.md
skills/cos-requesting-an-exception/SKILL.md
```

Claude Code skills use a different frontmatter contract from Company OS skills
(`name:`/`description:` versus `id:`/`authority:`/`appliesTo:`/`inputs:`), so
the two cannot be the same file. Each plugin skill is a short wrapper that
states the procedure and cites its canonical source path, keeping the
authoritative text in one place.

There are no slash commands. The three surfaces divide by question type: the
CLI owns verbs and verdicts on artifacts, the TUI owns browsing those same
verdicts, and skills own the judgment between the verbs — what an exit code
means here, which warning on a *passing* run is load-bearing, which shortcut is
forbidden. A `/cos-validate` that shells out to `company-os validate` sits in
none of those three. It restates an invocation the agent can already make,
creates a second place for that invocation to drift, and puts nothing in context
the CLI's own output does not carry.

## Constraints

**The frozen gate-7 oracle cannot be regenerated.**
`internal/skills/testdata/gate7_reference.json` holds 22 answers recovered from
tag `python-cli-final` after the Python implementation was deleted;
`gate_oracle_test.go:62-73` compares gate 7's rendered block against them with
exact string equality. Its own header records that regenerating means checking
out the tag. Any change to gate 7's output would force a hand-edit of a file
whose entire value is that it was not hand-written. This constraint is what
decides Part 3's placement.

**The oracle shares its fixture builder with the ordinary skills tests.**
`fourLayerWorkspace` (`internal/skills/skills_test.go:41`) is used at nine sites
in `skills_test.go` and again at `gate_oracle_test.go:192`, and the frozen
answers quote its synthesized paths verbatim. Those synthesized skill names are
test inputs, not workspace artifacts, and they must stay unprefixed. A blanket
"update every test asserting a literal skill name" instruction would rename them
and destroy the oracle — the precise outcome the new-gate decision exists to
avoid. Only assertions against committed `examples/` fixtures move.

**The frozen answers hard-code `[7/7]`, and that is correct.**
`gateMatchesReference` calls `skills.Gate(ws, 7)` and `renderGate(g, 7)` with
test-local constants, so every frozen string opens `[7/7]` regardless of how
many gates production runs. After this change those literals will look stale
next to goldens reading `[7/8]` and `[7/9]`. They are not. "Updating" them to
match is the single action that destroys this file irreversibly, because the
implementation that could regenerate the answers no longer exists.

**A federated skill slice cannot be renamed in place.** `skills/` is the fifth
entry in `DefaultSlicePaths`; slices are materialized `0444` with per-file
hashes in `workspace.lock.yaml`, and the slice-integrity gate fails both on a
hand-edit and on a slice-set change made without a re-sync. The rename order for
any consumer is therefore: rename in the source repo, bump the pin,
`workspace sync`, commit the new lock. There is no consumer-local fix, and the
standing rule is explicit about it: never edit a slice.

**The federated fixture's lock cannot be regenerated by the CLI.**
`examples/federated/workspace.yaml` pins a fictional
`https://git.example.com/…` remote, so `workspace sync` cannot fetch, and
`sync --frozen` re-materializes *from* the lock while re-checking
`AggregateHash(files)` against the stored `sliceHash` (`internal/federation/sync.go:231`),
so it cannot regenerate either. Renaming the fixture's skill mutates three lock
fields, not two: the per-file path key, that file's hash, and the slice's
`sliceHash` (`sync.go:319`, `materialize.go:318`). The production procedure has
to be specified rather than left to the implementer, or the fixture ships with a
wrong `sliceHash` that nothing catches until someone runs `--frozen`.

**`examples/banking` is unaffected, and should stay that way.** No
`workspace.lock.yaml` exists anywhere beneath it; its workspaces hold only
`workspace.yaml`, `teams/`, and `SYNC-NOTE.md`, with no materialized slices. Its
one skill, `bank/repos/platform-payments/skills/creating-prd.SKILL.md`, is on the
source side and is an ordinary rename. `acceptance.sh` does not validate either
banking workspace.

**Golden snapshots are byte-for-byte.** Five `.txt` files are diffed by
`internal/validate/golden_test.go:78-133` and by `examples/acceptance.sh` at
lines 41, 59, and 84. Three of them name skill files or ids directly. All five
carry `[N/M]` gate counters that shift when a gate is added.

**Generated context nodes are drift-checked.** Eighteen `CLAUDE.md` files carry
the `company-os:generated:start` marker and are verified by gate `[5/N]`. Four
embed a skill `id:` as a doc-index link title, so `graph build` must be re-run
and its output committed, per the standing rule that generated files are never
hand-edited.

**The frontmatter parser contract is fixed.** `frontmatter()` expects
`^---\n...\n---\n` exactly. Plugin skill wrappers must satisfy it if they are
ever discovered by Company OS tooling, and must satisfy Claude Code's own
parser regardless.

**The starter-kit skills are not valid workspace artifacts today.** None of the
four carries a `type:` key, and all four hand-write
`tags: [authority/canonical, kind/skill, process/<x>]`. `CoreFieldErrors`
(`internal/product/contract.go:65-68`) fails an artifact with no `type:`, and
neither `kind/skill` nor `process/*` appears in `kindTag` or `curatedFacets`
(`internal/graph/tags.go:22-35`), so `DeriveTags` would produce
`[authority/canonical]` alone and the tags-in-sync check fails too. Copying one
into a workspace unchanged yields two gate-4 failures, not a working file. Since
the rename already edits all four, `type: skill` is added and `tags:` corrected
to what `graph build` derives — otherwise the flattening claim is false and the
new test asserting it is a false assurance.

**Directory ordering is load-bearing and survives the prefix.** `globSorted`
(`skills.go:249-269`) returns name-ordered entries, and that order fixes gate 7's
emission sequence. A uniform prefix preserves relative order within a directory.
The only multi-skill directory in the repository is
`examples/failing-workspace/teams/ghost/skills/`, where `cos-creating-prd` still
precedes `cos-reviewing-prd`. No golden reorders. Stated rather than assumed.

**The starter-kit skills appear in no Go test.** `completing-a-change`,
`requesting-an-exception`, and `running-discovery` are absent from every test
file. Since nothing reads them either, flattening them is caught by nothing at
all. Coverage has to be created rather than repaired: a test that copies each
one into a synthesized workspace and asserts discovery finds it and the prefix
gate passes it.

**Nothing derives a skill name from anything else.** `DeriveTags`
(`internal/graph/tags.go:69-117`) never reads `name` or `id`; no
`ids/registry.yaml` contains a `skill://` entry; no finding code is constructed
from a skill name. The rename therefore has no derived-value fallout beyond the
three slots listed above.

## Key Decisions

**Enforcement goes in a new gate, not in gate 7.** Gate 7 is the natural home —
it already walks every discovered skill and knows each one's layer and
authority. It is also the one gate whose output is pinned to an oracle that
cannot be rebuilt. Extending it would invalidate all 22 frozen answers and
force a hand-edit that destroys their provenance. A separate gate costs a
golden-counter update across five files, which is mechanical and re-derivable
because the goldens are now compared Go-against-Go.

**Enforcement applies to all shared skills, not only canonical ones.** The
prefix's job is to make a Company OS skill recognizable in a federated
directory listing. A rule that exempted team skills would produce exactly the
mixed listing the change exists to prevent. Personal rules are exempt because
they never leave the machine that wrote them.

**Prefix all three slots rather than relying on the `skill://<scope>/` segment.**
The scope segment already namespaces the id, so `skill://product/creating-prd`
is arguably unambiguous. The filename is not, and the filename is what a human
browsing a synced slice sees. Gate 7 compares both id and name, so prefixing
one and not the other would leave the two identities disagreeing about what the
skill is called.

**Plugin skills are authored wrappers, not generated files.** Generating them
from the canonical `.SKILL.md` files would fit the repository's
derived-not-authored philosophy, but the generator would be a new CLI
subcommand, and CLI subcommands are explicitly out of scope. Each wrapper is a
few lines and cites its source, so drift is visible on inspection. If the
wrappers do drift, converting them to generated output later is a contained
change.

**The plugin's skills live at the repository root, not under
`company-os-starter/`.** Claude Code's default location is a root `skills/`
directory, and a root `skills/` is free today. The isolation from Company OS
discovery is structural rather than conventional: `Discover()`
(`skills.go:197-225`) globs only `<root>/company-os/skills`,
`<platform>/skills`, `<team>/skills`, and `<team>/scratchpad/personal-rules`,
and the repository root contains no `company-os/`, `platforms/`, `teams/`, or
`company-ontology/`, so it is not a workspace root and never can be walked as
one. The adjacency to `company-os-starter/skills/`, which means something else
entirely, is a readability cost accepted in exchange for the conventional
layout. Custom component paths in `plugin.json` remain available if it confuses
readers, but they are not needed for correctness.

**The migration is source-first, and the gate ships behind the fixtures.**
Because a consumer cannot edit a synced slice, a consumer that upgrades before
its sources do gets a failing workspace with no legal remedy. The sequence is
fixed: rename in producing repos, bump pins, re-sync consumers, and only then
adopt the binary carrying the gate. The federated fixture exercises exactly this
path, which is why its lock is rewritten rather than regenerated by hand-waving.

**Rejected: convention-only, documented in the template.**
`templates/SKILL-template.md` already states the naming rules and nothing
enforces them, which is how the four starter-kit skills ended up in a layout
their own discovery glob cannot see. A convention that only lives in prose is
the failure mode this change is repairing, not a design to repeat.

**Observed, not acted on: the plugin prefix is partly redundant.** Claude Code
namespaces plugin components as `<plugin-name>:<component>`, so a plugin named
`company-os` already yields `company-os:cos-creating-prd`. The `cos-` prefix
earns its place on the workspace side, where no such namespacing exists; on the
plugin side it is carried for consistency with the workspace names, deliberately
rather than by oversight.

## Out of Scope

- Renaming any CLI subcommand, action verb, finding-code prefix, gate slug,
  section slug, environment variable, binary, or release artifact.
- Compatibility aliases, deprecation warnings, or any migration tooling. The
  new gate reports what to rename; renaming it is manual.
- Registering `skill://` ids in `ids/registry.yaml`.
- Validating the `skill://<scope>/<name>` shape beyond the `cos-` segment.
- Surveying external consumers outside this repository. The research this
  design follows recorded that as unexamined, and it stays unexamined.
- Shell completion, plugin hooks, MCP servers, LSP servers, or agents in the
  plugin package.
