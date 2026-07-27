# Company OS Claude Code Plugin — High-Level Design

## Overview

The four canonical skills carry the layer of the methodology that no validator
can express. `creating-prd` does not restate what `company-os prd new` does — it
says *"run every command with `--json` and branch on the exit code"*, then tells
the agent what exit 5 means in context, which warnings on a **passing** run are
load-bearing and easy to miss, and, in as many words, *"do not hand-edit the
brief's status to get past this."*

That is judgment about method. The CLI owns the artifacts and the verdicts on
them; this is the reasoning between the verbs, and it is the reason "strict on
artifacts, flexible on process" is workable rather than merely stated.

Today an agent reaches it only by having the workspace checked out and knowing to
go read files. This change packages the four as a Claude Code plugin so they load
into context automatically, under names that stay unambiguous in a session where
they sit beside skills from a dozen unrelated publishers. That is what the `cos-`
prefix is for, and it is the only place a mixed skill namespace actually exists —
inside a workspace there is none, because every `.SKILL.md` a workspace can hold
is a Company OS artifact by construction.

## Stakeholders & Impact

**Agents operating a Company OS workspace.** They get the exit-code semantics,
the load-bearing warnings, and the forbidden shortcuts without being told to go
read `skills/`. This is the whole point of the change.

**Agent users with crowded skill namespaces.** A session commonly offers well
over a hundred skills from many sources, with live collisions between same-named
skills from different publishers. `creating-prd` there names nothing in
particular; `cos-creating-prd` does.

**Workspace authors.** Unaffected. No file in any workspace changes, no gate
changes, and the workspace-side skill names stay as they are.

**Anyone whose workspace customizes a canonical skill.** This is the risk the
change introduces. The workspace layering model — canonical over personal,
`extends:` resolution, shadowing detection — has no plugin-side equivalent, so
an agent could hold the plugin's generic procedure and a workspace's customized
one at the same time with nothing to say which wins. The plugin skills say so
themselves, in their own text, because there is no mechanism available to enforce
it.

## Goals

1. The four canonical skills are installable as a Claude Code plugin and load
   under `cos-`-prefixed names.
2. Each plugin skill names the canonical workspace file it derives from, so its
   source is one lookup away.
3. Each plugin skill states that a workspace's own layered skill outranks it.
4. The plugin's manifest is verifiable offline, by `make check`, without invoking
   an external tool.
5. Nothing the plugin adds is reachable by Company OS skill discovery.
6. No workspace file, gate, golden, or command changes.

## Non-Goals

- **No slash commands.** A command wrapping `company-os validate` restates an
  invocation the agent can already make, and Claude Code loads skills on
  description match anyway, so an explicit entry point is redundant with the
  mechanism that already reaches them.
- **No workspace-side renaming.** The `cos-` prefix applies to what this plugin
  publishes. Applying it inside a workspace is a separate change with a separate
  justification, currently deferred.
- **No generated wrappers.** Deriving the plugin skills from the canonical files
  would fit the repository's derived-not-authored habit, but the generator would
  be a new CLI subcommand, which is out of scope.
- **No agents, hooks, MCP servers, LSP servers, or themes** in the package.
- **No enforcement that the wrappers stay in sync** with their sources beyond
  citing them.

## Success Criteria

Observable when this ships:

- `claude plugin validate` reports no errors against the repository.
- A Go test asserts the manifest's required fields and their types, and fails on
  a malformed manifest without any external tool.
- With the plugin installed, the four skills are offered under their `cos-`
  names.
- Each plugin skill's body names its canonical source path, and that path exists.
- Each plugin skill's body states that a workspace's layered skill is
  authoritative over it.
- `company-os skills list` and `company-os validate` behave identically before
  and after installation, on every example workspace.
- `git diff --stat` touches no file under `examples/`, `internal/`, or
  `cmd/`.
