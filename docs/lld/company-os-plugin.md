# Company OS Claude Code Plugin — Low-Level Design

## Architecture

A new directory tree at the repository root plus one Go test. No existing file
changes.

```
.claude-plugin/plugin.json          manifest, name "company-os"
.claude-plugin/marketplace.json     marketplace entry pointing at this repo
skills/cos-creating-prd/SKILL.md
skills/cos-running-discovery/SKILL.md
skills/cos-completing-a-change/SKILL.md
skills/cos-requesting-an-exception/SKILL.md
```

`skills/` is Claude Code's default component location and must sit at the plugin
root, not inside `.claude-plugin/`. Component directories other than `skills/`
are not created.

### Why the wrappers cannot be the canonical files

The two formats have incompatible frontmatter contracts. A Company OS skill
carries `id`, `authority`, `appliesTo`, `inputs`, `outputs`; a Claude Code skill
carries `name` and `description`, where `name` must equal its directory name.
One file cannot satisfy both parsers, so each plugin skill is a short wrapper.

Each wrapper states the procedure and cites the workspace-relative path of the
canonical `.SKILL.md` it derives from, keeping the authoritative text in one
place and making drift visible on inspection.

### Isolation from Company OS discovery

`Discover()` (`internal/skills/skills.go:197-225`) globs exactly four shapes:
`<root>/company-os/skills`, `<platform>/skills`, `<team>/skills`, and
`<team>/scratchpad/personal-rules`. The repository root holds no `company-os/`,
`platforms/`, `teams/`, or `company-ontology/` directory, so it cannot be
resolved as a workspace root and the plugin's `skills/` can never be walked. The
isolation is structural, not conventional — worth stating, because "a root
`skills/` happens to be free" would be a much weaker guarantee.

The adjacency to `company-os-starter/skills/`, which means something entirely
different, is a readability cost accepted for the conventional layout. Custom
component paths are available in the manifest if it proves confusing; they are
not needed for correctness.

### Precedence, stated rather than enforced

A workspace resolves competing skills through a real model: canonical outranks
personal, `extends:` layers a team skill onto a platform base, and the
skills-layering gate rejects a team skill that shadows a canonical id or name.
None of that exists on the plugin side, and no mechanism available here can
extend it there — Claude Code has no visibility into a workspace's layering, and
the CLI has no visibility into loaded plugin skills.

So an agent with the plugin installed and a workspace checked out can hold two
procedures for the same task: the plugin's generic one and the workspace's
possibly customized one. The only lever available is the wrapper's own text,
which says the workspace's layered skill is authoritative and to prefer what
`company-os skills list` reports. That is weak, and it is the honest ceiling of
what this design can do.

## Constraints

**`--strict` is not a safe acceptance bar.** `claude plugin validate --strict`
promotes unrecognized-field warnings to errors, and the tool is not vendored,
not version-pinned, and not invoked by `make check` (gofmt + vet +
`go test ./...` + `examples/acceptance.sh`). A new warning in any release would
turn a green repository red with no code change. Every other gate in this
repository is hermetic and offline; the acceptance script even skips its
federation section rather than fail on an old git. The bar is "no errors", with
an offline Go test carrying the part that must hold in CI.

**Plugin component namespacing already exists.** Claude Code namespaces
components as `<plugin-name>:<component>`, so a plugin named `company-os` yields
`company-os:cos-creating-prd`. The `cos-` prefix is therefore partly redundant
at the fully-qualified name. It still earns its place on the bare name, which is
what a description match surfaces and what a user types — and it keeps the
plugin's names aligned with the workspace-side names should the deferred rename
ever land.

**Wrapper drift has no automated detection.** Citing the source path makes drift
visible to a reader, and nothing more. Generating the wrappers would fix this and
would require a new CLI subcommand, which is out of scope. If drift becomes real,
converting authored wrappers to generated output later is contained.

## Key Decisions

**Ship skills, not slash commands.** Two arguments, and the second is the one
that holds. First: a command wrapping `company-os validate` restates an
invocation the agent can already make and creates a second place for it to
drift. That argument is easy to counter — someone will propose a command that
walks a genuine multi-step workflow rather than one binary call. The stronger
argument is that Claude Code loads skills automatically on description match, so
an explicit command is a redundant entry point to something already reachable.
A workflow belongs in a skill's steps, not in a parallel command surface.

**`cos-` here, and only here.** This is the one place a mixed skill namespace
exists. Inside a workspace, every discoverable `.SKILL.md` is a Company OS
artifact by construction and slice targets are disjoint by validator rule, so
there is nothing to disambiguate against and a prefix applied to every member of
a population distinguishes nothing within it.

**Authored wrappers, source-cited.** See the constraint above. Small files,
visible provenance, contained upgrade path.

**Precedence is documented, not enforced.** Recorded as a known limitation rather
than papered over. It is the one thing this change genuinely risks breaking, and
the alternative — building a plugin-side layering model — is far larger than the
value of the package.

## Out of Scope

- Renaming any skill inside a workspace.
- Any validation gate, finding code, or change to `validate`.
- Slash commands, agents, hooks, MCP servers, LSP servers, themes, or monitors.
- Generating the wrappers from their canonical sources.
- Any automated wrapper-versus-source drift check.
- Extending the workspace layering model to cover plugin-provided skills.
