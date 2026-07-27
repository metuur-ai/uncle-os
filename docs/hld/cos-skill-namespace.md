# `cos-` Skill Namespace and Claude Code Plugin — High-Level Design

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

## Overview

Company OS skills carry generic names. `creating-prd`, `running-discovery`, and
`completing-a-change` say what they do but not what they belong to. That is
fine inside a single workspace, where every `.SKILL.md` under `company-os/`,
`platforms/<p>/`, or `teams/<t>/` is by construction a Company OS artifact.

**A correction to the original justification.** This change was first written
against a federation driver: that `workspace sync` merges `skills/` slices from
several repos into one directory where generic names become ambiguous. That
directory cannot exist. Slice targets must be disjoint across a manifest, and
the validator rejects any overlap by name — so two repos can never contribute
skills to the same path. The banking fixture shows the real shape: three source
repos land at `company-os/skills/`, `platforms/fraud/skills/`, and
`platforms/payments/skills/`. Origin is always in the path, structurally, and
`skills list` already renders it. Worse, a fixed token applied to every member of
a population distinguishes nothing within it, and every `.SKILL.md` a workspace
can hold is a Company OS artifact by construction. The set the prefix would
disambiguate against, inside a workspace, is empty.

The real driver is a shared agent namespace. A Claude Code session routinely
offers well over a hundred skills from a dozen unrelated sources, with live
collisions between same-named skills from different publishers. There,
`creating-prd` is genuinely ambiguous and `cos-creating-prd` is not. That value
is delivered by the plugin. It is not delivered by renaming files inside a
workspace, where no mixed namespace exists.

This change gives every Company OS skill a `cos-` prefix across all three of
its name slots — file stem, `id:`, and any `extends:` URI addressing it — and
adds a validation gate so the convention holds in repos that were written after
this document. It also packages the four canonical starter-kit skills as a
distributable Claude Code plugin, so the same skills that govern a workspace are
invokable in an agent session under the same `cos-` names.

The four starter-kit skills under `company-os-starter/skills/` are flattened to
the same shape while we are in those files. They currently sit at
`<name>/SKILL.md`, which the discovery glob would never match. That is a
consistency problem, not a live defect: like the sibling `.md` files under
`templates/`, they are human reference copies that no code path reads. They are
not embedded in the binary, `company-os init` does not copy them, and
`company-os-starter/` is not a workspace root, so nothing discovers them from
either layout. They are also not currently valid workspace artifacts: none carries a `type:`
key, and all four hand-write tags that derivation does not produce, so copying
one into a workspace today yields validation failures. Flattening the layout,
adding `type: skill`, and correcting `tags:` together mean a reader who copies
one into a real workspace gets a file that works.

## Stakeholders & Impact

**Agent users with a mixed skill namespace.** This is the population that
actually gains. Their session offers Company OS skills beside skills from a
dozen unrelated publishers, and `creating-prd` there names nothing in
particular. The prefix makes it unambiguous at the point of invocation.

**Workspace authors.** They gain consistency and pay for it. Their skills read
uniformly, and the names in their workspace match the names their agent sees —
but no ambiguity is resolved inside the workspace itself, because there was none
to resolve.

**Teams writing their own skills.** A team skill that shadows a canonical one
is already a gate-7 failure. The prefix makes the intended relationship legible
before validation runs: a team writing `cos-creating-prd` is visibly reaching
for a Company OS name and should be extending, not replacing.

**Agents operating a workspace.** Four shipped skills instruct an agent to grep
CLI stderr for the `company-os ` prefix and tell it what each exit code means in
context — including which shortcuts past a gate are forbidden. That judgment is
the layer no validator can express, and today an agent reaches it only by
knowing to go read the files. Packaging those four as plugin skills loads it
automatically. The plugin ships no slash commands: wrapping a CLI the agent can
already invoke would add drift, not capability.

**Anyone with an existing workspace.** The rename is hard, with no aliases.
Every `.SKILL.md` filename, every skill `id:`, and every `extends:` URI in a
live workspace stops matching until it is renamed. The new gate reports exactly
which files are wrong, so the migration is mechanical, but it is not automatic
and it is not optional.

**Anyone consuming skills through a federated slice.** This is the hard case,
and it is the same population the change exists to serve. `skills/` slices are
materialized at `0444` and their bytes are hash-locked in `workspace.lock.yaml`;
gate `[8/8]` fails on a hand-edit. A consumer therefore cannot rename a synced
skill at all. The rename has to happen in the source repo, followed by a pin
bump and a re-sync in every consumer. Until a consumer re-syncs, its workspace
contains unprefixed skills it has no write access to, and the new gate fails it.
Ordering that migration is part of this change, not an afterthought to it, and
the gate warns rather than fails on a skill inside a locked slice so no consumer
is ever stranded in a state it cannot write to.

**Not affected:** the `company-os` binary name, all 17 CLI subcommands, action
verbs, finding-code prefixes, `COMPANY_OS_WORKSPACE_ROOT`, and release artifact
names. None of these change.

## Goals

1. Every shared skill file in the repository is named `cos-<name>.SKILL.md` and
   sits directly inside a `skills/` directory, including the four starter-kit
   reference copies.
2. Every skill `id:` reads `skill://<scope>/cos-<name>`, and every `extends:`
   URI reads `platform-skill://<platform>/cos-<name>`, resolving to the renamed
   file.
3. `company-os validate` fails a workspace containing an authored shared skill
   whose file stem or `id:` name segment lacks the prefix, names the offending
   file, and downgrades to a warning when that file is inside a read-only slice
   the workspace cannot edit.
4. The four canonical skills are installable as a Claude Code plugin, exposing
   `cos-`-named skills and no slash commands.
5. `make check` passes: gofmt, vet, `go test ./...`, and
   `examples/acceptance.sh` against all five example workspaces.
6. `company-os validate` exits 0 against `examples/workspace` and
   `examples/federated`, and produces exactly the planted failures against the
   three failing fixtures.

## Non-Goals

- **The CLI's own naming is untouched.** No subcommand, action verb, finding
  code, gate slug, section slug, environment variable, binary name, or release
  artifact is renamed. The guidance chain keeps emitting `company-os <cmd>`.
- **No aliases and no deprecation window.** Old skill names stop resolving the
  moment this ships. There is no compatibility layer to build, document, or
  later remove.
- **The frozen gate-7 oracle stays byte-identical.** Its 22 answers were
  recovered from a deleted Python implementation and cannot be re-derived
  in-tree. Nothing in this change may alter gate 7's rendered output.
- **No `skill://` entries are added to `ids/registry.yaml`.** Skill ids are not
  registered canonical IDs today and do not become them here.
- **No id format validation beyond the prefix.** The `skill://<scope>/<name>`
  shape stays a documented convention; only the `cos-` segment becomes
  enforced.
- **Personal rules are exempt.** Files under `scratchpad/personal-rules/` are
  git-ignored, never federate, and keep plain `.md` names with no prefix.

## Success Criteria

Observable when this ships:

- `find . -name '*.SKILL.md'` returns only `cos-`-prefixed basenames, and
  `company-os-starter/skills/` contains flat `cos-<name>.SKILL.md` files rather
  than `<name>/SKILL.md` directories.
- `company-os skills list` in `examples/workspace` reports every skill under its
  prefixed name.
- Copying a starter-kit skill unchanged into a workspace's `company-os/skills/`
  yields a file that discovery finds and the new gate passes.
- The federated fixture's renamed skill slice validates clean, proving the
  source-rename plus pin-bump plus re-sync path produces a consistent lock.
- Introducing an unprefixed shared skill into `examples/workspace` makes
  `company-os validate` exit non-zero with a finding naming that file; removing
  it restores exit 0.
- `claude plugin validate` reports no errors, an offline Go test covers the
  manifest's shape so `make check` catches a malformed one, and the plugin's
  skills appear under the `cos-` names in a Claude Code session.
- The five golden `.txt` snapshots and `examples/acceptance.sh` agree with the
  binary, and `internal/skills/testdata/gate7_reference.json` is unmodified.
