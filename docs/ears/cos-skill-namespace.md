# `cos-` Skill Namespace and Claude Code Plugin — EARS Specifications

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

## Unit 1: Skill file naming

**Why:** One name per skill, consistent across every slot and every layer, so
the workspace-side names match the names the plugin publishes into a shared
agent namespace. The federation rationale this unit originally carried does not
hold — slice targets are disjoint by validator rule, so a merged skills
directory cannot exist and origin is always in the path. A shared file stem
also has to stay inside the one layout the discovery glob matches.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL name every shared skill file `cos-<name>.SKILL.md`, where `<name>` is the skill's former unprefixed stem. |
| R-1.2 | THE SYSTEM SHALL place every shared skill file belonging to a workspace root directly inside a `skills/` directory, one level deep, with no intervening subdirectory. This governs Company OS skills only; Claude Code plugin skills follow the opposite convention under R-7.3. |
| R-1.3 | WHERE a skill file previously existed at `skills/<name>/SKILL.md`, THE SYSTEM SHALL relocate it to `skills/cos-<name>.SKILL.md` and remove the emptied directory. |
| R-1.4 | THE SYSTEM SHALL leave personal-rule files under `scratchpad/personal-rules/` unrenamed, retaining their plain `.md` extension and unprefixed stem. |
| R-1.5 | WHEN `company-os skills list` runs against `examples/workspace`, THE SYSTEM SHALL report each discovered skill under its `cos-`-prefixed name. |
| R-1.6 | IF a skill file's stem is `cos-<name>`, THEN THE SYSTEM SHALL report its `Name()` as `cos-<name>`, with the `.SKILL.md` suffix stripped and the prefix retained. |
| R-1.7 | THE SYSTEM SHALL give every starter-kit skill file a `type: skill` frontmatter key. |
| R-1.8 | THE SYSTEM SHALL set every starter-kit skill file's `tags:` to what tag derivation produces for a skill at the company layer, removing any hand-written tag that derivation does not emit. |
| R-1.8a | THE SYSTEM SHALL record in the starter-kit files that derived tags include a location facet at the platform and team layers, so a reader copying one to a platform runs the graph build rather than assuming the committed tags are portable. |
| R-1.9 | WHEN a starter-kit skill file is copied unchanged into the `company-os/skills/` directory of a synthesized workspace that contains no skill of the same name, THE SYSTEM SHALL discover it, SHALL pass it through the prefix gate, and SHALL report no core-field or tag-drift finding against it. |

## Unit 2: Skill identifiers and extension references

**Why:** Gate 7 compares two skills for shadowing on `id:` first and file name
second. If only one slot carries the prefix, the two identities disagree about
what the skill is called and the shadowing verdict depends on which slot
happened to be edited. `extends:` addresses a file path, so it has to track the
filename exactly.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL express every shared skill's `id:` as `skill://<scope>/cos-<name>`, preserving the existing `<scope>` segment unchanged. |
| R-2.2 | THE SYSTEM SHALL express every `extends:` URI as `platform-skill://<platform>/cos-<name>`, preserving the existing `<platform>` segment unchanged. |
| R-2.3 | WHEN resolving `extends: platform-skill://<platform>/cos-<name>`, THE SYSTEM SHALL resolve it to `platforms/<platform>/skills/cos-<name>.SKILL.md`. |
| R-2.4 | IF an `extends:` URI names a base skill that does not exist, THEN THE SYSTEM SHALL report a dangling-extends finding quoting the URI verbatim, including its `cos-` prefix. |
| R-2.5 | THE SYSTEM SHALL keep a skill's `id:` name segment and its file stem identical to each other. |
| R-2.6 | THE SYSTEM SHALL NOT add any `skill://` entry to any `ids/registry.yaml`. |

## Unit 3: Prefix validation gate

**Why:** The four starter-kit skills drifted into a layout their own discovery
glob cannot match, and nobody noticed, because the rule lives only in a
template's prose. A naming convention that nothing checks decays on the first
skill written in a repo whose author never read the template. Enforcement is
what makes the prefix a signal rather than a habit.

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL provide a validation gate that inspects every skill discovered at the company, platform, and team layers. |
| R-3.2 | IF a discovered shared skill's file stem does not begin with `cos-`, THEN THE SYSTEM SHALL emit a failing finding naming that skill's workspace-relative path. |
| R-3.3 | IF a discovered shared skill's `id:` is a string whose name segment does not begin with `cos-`, THEN THE SYSTEM SHALL emit a failing finding naming that skill's workspace-relative path and the offending id. |
| R-3.3a | IF a discovered shared skill has no `id:` key, THEN THE SYSTEM SHALL evaluate only its file stem under R-3.2 and SHALL NOT emit an id finding. |
| R-3.3b | IF a discovered shared skill's `id:` is present but not a string, THEN THE SYSTEM SHALL emit a failing finding naming the path and reporting the id as unusable, without attempting to parse a name segment. |
| R-3.3c | IF both a skill's file stem and its `id:` name segment lack the prefix, THEN THE SYSTEM SHALL emit exactly one finding naming both slots. |
| R-3.4 | WHILE a workspace contains no unprefixed shared skill, THE SYSTEM SHALL emit a single passing finding reporting the number of skills checked, counted on the same basis as the skills-layering gate's own totals so the two gates cannot disagree about one discovery result. |
| R-3.5 | THE SYSTEM SHALL exclude personal rules under `scratchpad/personal-rules/` from this gate. |
| R-3.6 | THE SYSTEM SHALL exclude the `knowledge/` root from this gate, consistent with the other non-federation gates. |
| R-3.7 | WHEN this gate emits any failing finding, THE SYSTEM SHALL cause `company-os validate` to exit non-zero. |
| R-3.7a | WHEN this gate reports an unprefixed skill, THE SYSTEM SHALL name the next command in the finding, including the graph rebuild a rename requires, honoring the guidance-chain convention rather than leaving that step to a how-to the reader is not looking at. |
| R-3.8 | THE SYSTEM SHALL emit this gate's findings under a finding code distinct from every code already present in `internal/model/codes.go`. |
| R-3.9 | THE SYSTEM SHALL NOT alter the rendered output of the existing skills-layering gate in any respect, including its finding codes, message text, ordering, and counts. |
| R-3.9a | WHERE the new gate composes finding text, THE SYSTEM SHALL do so by adding a case to the existing exported skills message switch, and SHALL NOT modify any case that gate serves today. |
| R-3.9b | THE SYSTEM SHALL have the validation composer resolve slice targets and pass them to the new gate, and SHALL NOT introduce a dependency from the skills package to the federation package. |
| R-3.10 | THE SYSTEM SHALL leave `internal/skills/testdata/gate7_reference.json` byte-for-byte unmodified. |
| R-3.11 | THE SYSTEM SHALL place the new gate immediately after the skills-layering gate in the validate sequence, leaving that gate at ordinal 7 and moving federated slice integrity from ordinal 8 to ordinal 9. |
| R-3.11a | THE SYSTEM SHALL NOT give the new gate an ordinal that differs between a monorepo workspace and a federated workspace. |
| R-3.12 | THE SYSTEM SHALL leave the gate ordinal and denominator passed by the frozen-oracle test at 7, and SHALL record in that test why those literals do not track the production gate total. |
| R-3.13 | THE SYSTEM SHALL leave every skill name synthesized by `internal/skills` test fixtures unprefixed, and SHALL record in `skills_test.go` that those names are test inputs shared with the frozen oracle rather than workspace artifacts. |

## Unit 4: Snapshot and test realignment

**Why:** Five golden files are diffed byte-for-byte by both the Go suite and the
acceptance script, and three of them quote skill paths and ids. Adding a gate
also shifts every `[N/M]` counter. None of this is optional; the gate is the
CI contract.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL update all five golden `.txt` snapshots so their skill paths and ids carry the `cos-` prefix. |
| R-4.2 | THE SYSTEM SHALL update all five golden `.txt` snapshots so their `[N/M]` gate counters reflect the added gate. |
| R-4.3 | WHEN `examples/acceptance.sh` runs, THE SYSTEM SHALL produce output matching every golden snapshot with no diff. |
| R-4.4 | WHEN `company-os validate` runs against `examples/workspace` or `examples/federated`, THE SYSTEM SHALL exit 0. |
| R-4.5 | WHEN `company-os validate` runs against `examples/failing-workspace`, THE SYSTEM SHALL report the same 15 failing findings as before this change, with skill paths and ids prefixed, and SHALL close with the same problem count. |
| R-4.6 | THE SYSTEM SHALL update every Go test assertion that names a skill in a committed `examples/` fixture so the assertion matches the prefixed value. |
| R-4.6a | THE SYSTEM SHALL NOT update any Go test assertion that names a skill synthesized by a test fixture, including those in `internal/skills/skills_test.go` and `internal/skills/gate_oracle_test.go`. |
| R-4.7 | THE SYSTEM SHALL provide test coverage that copies each of the four starter-kit skills into a synthesized workspace and asserts discovery finds it, the prefix gate passes it, and no core-field or tag-drift finding is raised against it. |
| R-4.8 | WHEN `make check` runs, THE SYSTEM SHALL complete gofmt, vet, `go test ./...`, and `examples/acceptance.sh` without failure. |

## Unit 5: Federated slice migration

**Why:** `skills/` is a synced slice path. Slices are materialized read-only and
hash-locked, and gate `[8/8]` fails on any hand-edit. A consumer therefore has no
way to rename a skill it received, which means the migration has an order and
adopting the gate before the sources rename produces a workspace nobody can
repair.

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL NOT rename any file inside a materialized slice by editing it in place. |
| R-5.2 | WHERE a skill reaches a workspace through a slice, THE SYSTEM SHALL require the rename to occur in the source repository, followed by a pin bump and a re-sync. |
| R-5.3 | WHEN a slice is re-synced after a skill rename, THE SYSTEM SHALL write a lock recording the new file path, that file's new hash, and the slice's recomputed aggregate hash, and SHALL remove the former path's entry. |
| R-5.3a | THE SYSTEM SHALL specify the procedure used to produce the federated fixture's rewritten lock, given that the fixture pins an unreachable remote and that frozen sync re-checks the aggregate hash it would need to regenerate. |
| R-5.4 | WHEN `company-os validate` runs against the federated fixture after its slice is renamed and its lock rewritten, THE SYSTEM SHALL report slice integrity clean and exit 0. |
| R-5.4a | WHEN frozen sync runs against the federated fixture after its lock is rewritten, THE SYSTEM SHALL report no aggregate-hash mismatch. |
| R-5.5 | THE SYSTEM SHALL document the source-first migration order — rename sources, bump pins, re-sync consumers — in the federation how-to. |
| R-5.6 | IF a workspace contains an unprefixed skill whose path falls inside a manifest-declared slice target, THEN THE SYSTEM SHALL emit a warning rather than a failure, and the warning SHALL name the source repository as the place to fix it. |
| R-5.6a | THE SYSTEM SHALL determine slice targets from the manifest alone, so the downgrade still applies in a workspace whose lock is missing or malformed. |
| R-5.7 | WHILE a consumer workspace's only unprefixed skills are inside manifest-declared slice targets, THE SYSTEM SHALL exit 0, so that a consumer is never left in a state it has no write access to repair. |
| R-5.8 | WHERE a skill is evaluated in the repository that authors it, THE SYSTEM SHALL apply R-3.2 at full failing severity, so that enforcement lands on the party able to act. |
| R-5.9 | THE SYSTEM SHALL leave `examples/banking` free of materialized slices and lock files. |

## Unit 6: Generated artifact regeneration

**Why:** Four generated `CLAUDE.md` context nodes embed a skill `id:` as a
doc-index link title, and gate `[5/N]` fails on drift between a generated block
and what the builder would produce. Hand-editing generated files is a standing
prohibition, so these have to be rebuilt rather than patched.

| ID | EARS statement |
| --- | --- |
| R-6.1 | WHEN a skill's `id:` changes, THE SYSTEM SHALL regenerate every `CLAUDE.md` context node whose doc index cites that id. |
| R-6.2 | THE SYSTEM SHALL produce regenerated context nodes via `company-os graph build` rather than by hand-editing content inside a `company-os:generated:start` block. |
| R-6.3 | WHILE regenerated context nodes are committed, THE SYSTEM SHALL report no drift from the context-node gate. |
| R-6.4 | THE SYSTEM SHALL NOT introduce any skill-name-derived or skill-id-derived value into tag derivation. Verified true today at `internal/graph/tags.go:69-117`; stated as a boundary the change must not cross, not as new behavior. |

## Unit 7: Claude Code plugin package

**Why:** The canonical skills carry the judgment the CLI cannot encode — what an
exit code means for your situation, which passing-run warning is load-bearing,
which shortcut past a gate is forbidden. Today an agent only reaches that
judgment by having the workspace checked out and knowing to go read the files. A
plugin loads it into context automatically.

The package ships skills and no slash commands. A command wrapping
`company-os validate` would restate a binary the agent can already invoke, add a
second place for the invocation to drift, and put nothing in context that the
CLI's own output does not already carry. The judgment layer is the only thing
worth packaging.

| ID | EARS statement |
| --- | --- |
| R-7.1 | THE SYSTEM SHALL provide a plugin manifest at `.claude-plugin/plugin.json` declaring a `name` of `company-os`, plus `description`, `version`, and `repository`. |
| R-7.2 | THE SYSTEM SHALL provide a marketplace entry at `.claude-plugin/marketplace.json` listing the `company-os` plugin with its source path. |
| R-7.3 | THE SYSTEM SHALL provide one Claude Code skill per canonical starter-kit skill, at `skills/cos-<name>/SKILL.md`. |
| R-7.4 | THE SYSTEM SHALL give every plugin skill frontmatter containing `name` and `description`, where `name` equals the directory name. |
| R-7.5 | THE SYSTEM SHALL state in every plugin skill's body the workspace-relative path of the canonical `.SKILL.md` it derives from. |
| R-7.6 | THE SYSTEM SHALL NOT provide any slash command that wraps a `company-os` subcommand. |
| R-7.7 | THE SYSTEM SHALL NOT create a `commands/` directory at the repository root. |
| R-7.8 | WHEN `claude plugin validate` runs against the repository, THE SYSTEM SHALL report no errors. |
| R-7.8a | THE SYSTEM SHALL provide an offline Go test asserting the plugin manifest's required fields and their types, so that plugin correctness is verifiable by `make check` without invoking an external tool. |
| R-7.8b | THE SYSTEM SHALL NOT make any external tool's warning output a condition of the build passing. |
| R-7.9 | THE SYSTEM SHALL NOT place `skills/` or any other component directory inside `.claude-plugin/`. |
| R-7.9a | THE SYSTEM SHALL NOT create any plugin component directory at the repository root other than `skills/`. |
| R-7.11 | WHERE a plugin skill and a workspace skill of the same name are both reachable, THE SYSTEM SHALL state in the plugin skill's body that the workspace's own layered skill is authoritative, since the plugin carries no equivalent of the workspace shadowing and extends resolution. |
| R-7.10 | THE SYSTEM SHALL keep the repository root free of `company-os/`, `platforms/`, `teams/`, and `company-ontology/` directories, so that it is structurally incapable of being resolved as a workspace root and the plugin's `skills/` directory can never be reached by Company OS skill discovery. |

## Unit 8: Documentation realignment

**Why:** 23 `.SKILL.md` mentions, 67 `skill://` references, and 19
`platform-skill://` references sit in tracked markdown, concentrated in the
agent-skills how-to. A rename that leaves the docs describing the old names
teaches the old names.

| ID | EARS statement |
| --- | --- |
| R-8.1 | THE SYSTEM SHALL update `templates/SKILL-template.md` so its stated naming rule requires `cos-<name>.SKILL.md`. |
| R-8.2 | THE SYSTEM SHALL update `docs/ONTOLOGY-GUIDE.md` so its `skill://` and `platform-skill://` examples carry the prefix. |
| R-8.3 | THE SYSTEM SHALL update every tracked markdown occurrence of a skill filename, `skill://` id, or `platform-skill://` URI to its prefixed form. |
| R-8.4 | THE SYSTEM SHALL document the new validation gate in the validation-gate how-to and the troubleshooting reference, including what the failure means, how to fix it, and why a slice-resident skill warns instead of failing. |
| R-8.5 | THE SYSTEM SHALL document plugin installation in the repository README. |
| R-8.6 | THE SYSTEM SHALL update every literal `[N/M]` gate counter in every tracked file, including markdown, Go doc comments, YAML fixture comments, and shell script prose, to reflect the added gate and the federation gate's move to ordinal 9. |
| R-8.6a | THE SYSTEM SHALL NOT alter any `[N/M]` literal inside a test that constructs its own gate total, since those are self-consistent and independent of the production sequence. |
| R-8.6b | THE SYSTEM SHALL regenerate every golden snapshot by running the binary, and SHALL NOT hand-edit or hand-merge golden text. |
| R-8.7 | THE SYSTEM SHALL update the gate-count prose in the repository-root `CLAUDE.md` invariants, which is the file that teaches every agent the gate numbering. |
| R-8.8 | THE SYSTEM SHALL NOT alter any documented `company-os <subcommand>` invocation. |
| R-8.9 | WHERE documentation states the skill naming rule using `<name>` as a metavariable, THE SYSTEM SHALL rewrite the rule by hand rather than by textual substitution, so the prefix lands outside the metavariable. |

## Unit 9: Non-regression of unaffected surfaces

**Why:** The research behind this design found 559 tracked-doc occurrences of
`company-os `, 17 subcommands pinned byte-for-byte by tests, and 29 finding-code
prefixes — none of which derive from a skill name. Stating the boundary makes an
accidental widening of the diff a spec violation rather than a judgment call.

| ID | EARS statement |
| --- | --- |
| R-9.1 | THE SYSTEM SHALL leave all 17 CLI subcommand names unchanged. |
| R-9.2 | THE SYSTEM SHALL leave every CLI action verb, usage string, and `--help` banner unchanged. |
| R-9.3 | THE SYSTEM SHALL leave all 131 finding-code literals in `internal/model/codes.go` unchanged, other than adding the code required by R-3.8. |
| R-9.4 | THE SYSTEM SHALL leave the binary name, release artifact names, install script tool name, and `COMPANY_OS_WORKSPACE_ROOT` unchanged. |
| R-9.5 | THE SYSTEM SHALL leave every guidance-chain next-command string unchanged. |
| R-9.6 | THE SYSTEM SHALL NOT provide an alias, redirect, or deprecation warning for any pre-rename skill name or id. |
| R-9.7 | IF a workspace contains a skill under a pre-rename name, THEN THE SYSTEM SHALL treat it as an ordinary unprefixed skill and fail it per R-3.2, with no special-case message. |
