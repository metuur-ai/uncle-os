# `skills install` — Low-Level Design

## Architecture

One embed package, one internal function, one new action on an existing
command.

```
skills/embed.go                     //go:embed *.SKILL.md, package skillfiles
internal/skills/install.go          Install() — placement, versions, rebuild
cmd/company-os/skills.go            `install` action, sentences, records
internal/render/skills.go           seven new codes
internal/model/codes.go             the codes themselves
cmd/company-os/args.go              action choices: {list, install}
```

### Why the embed package sits at `skills/`

A `//go:embed` directive may only name files at or below its own package
directory, so nothing under `internal/` can embed `company-os-starter/skills/`.
The directive lives beside the files, exactly as `templates/embed.go` does for
`reality-component.md` and for the same structural reason. The alternative —
copying the `.SKILL.md` files down beside a deeper package — would fork the one
source of text between the reference copy a reader browses and the bytes the
command writes.

`skillfiles` therefore sits *above* `internal/` and must not import it. That is
why it re-declares `Suffix` rather than importing `internal/skills.Suffix`, and
why its `version()` scan is duplicated in `install.go` rather than shared: one
reads bytes the binary carries, the other reads a file a team may have edited,
and the two agreeing today is not a reason to couple them.

### The company layer is the only destination

All five skills declare `company://all-platforms` or `company://all-teams`, and
`DeriveTags` gives a company-layer skill exactly `[authority/canonical]` — which
is what these files carry after the starter-kit repair. A platform layer would
add `platform/<p>`, so the same bytes would raise `frontmatter.tags-drift` the
moment they landed. The destination is not a preference; it is the only layer
where the shipped tags are already in sync.

### The five dispositions

| Installed | Action | Code |
|---|---|---|
| absent | write | `skills.installed` |
| older | overwrite | `skills.updated` |
| same version | leave | `skills.unchanged` |
| newer | leave | `skills.locally-newer` |
| version unreadable | leave | `skills.unreadable` |

One code per outcome rather than one code with an outcome field: an agent's
whole decision after this command is "did anything change, and is anything
wrong", and codes are the contract that answers it. `locally-newer` and
`unreadable` are the two an agent must not read as success.

Same-version files are left alone rather than rewritten. Byte differences at an
unchanged version are a team's local edit, and a version that did not move is
not a claim to overwrite them.

### Version comparison is numeric per segment

`compareVersions` splits on `.` and compares each segment as an integer, so
`1.10 > 1.9`. A string comparison gets this wrong — `"1.10" < "1.9"` — and would
overwrite a newer skill with an older one the first time a skill reached its
tenth revision. A non-numeric segment sorts below every number, so a malformed
version reads as older; the file is still protected, because an unreadable
version is caught earlier and never reaches the comparison.

### The rebuild is not optional

Writing into `company-os/skills/` changes a graph-docs root: the company
`CLAUDE.md` context node gains entries and the directory becomes index-eligible.
Gate 5 fails on both until derivation runs. `Install` therefore takes a
`Rebuild` — the same `scaffold -> graph` seam `reality new` and `add` use,
declared in `internal/skills` and satisfied in `cmd/` so this package never
imports `internal/graph` — and returns its lines for the caller to print first.

It runs unconditionally, including on a no-change run: a workspace whose derived
artifacts were already stale should not be left that way by a command that just
walked the same directory.

## Constraints

**This was caught by acceptance, not by unit tests.** Every unit test passed
against an `Install` that skipped the rebuild, because each asserted its own
disposition and none ran `validate` over the whole workspace. Running the built
binary against a fresh `init` workspace failed immediately: baseline `validate`
exit 0, post-install exit 1, two gate-5 failures. The regression test added
afterwards asserts the rebuild ran and its lines were returned, not that
specific files exist, so a future derivation this command must trigger is
covered without being named.

**The usage string is frozen output.** Adding `install` to the action choices
changes `usage: company-os skills [-h] {list}` to `{list,install}` and the
invalid-choice diagnostic to `(choose from list, install)`, both pinned by
`TestArgumentErrorDiagnostics`. Those assertions are updated. The `ids`
command's identical-looking assertion is NOT: it still has one action.

**No Python oracle.** The reference CLI shipped no such command, so R-0.8 does
not freeze these sentences. They are composed at the command like every other
scaffolding command's, and the renderer passes `Message` through rather than
recomposing from fields — the opposite of `skills list`, whose every line is a
transcription and is therefore composed in the renderer.

## Key Decisions

**Extend `skills`, not a new top-level `setup`.** The CLI has one command group
per concept and `skills` already exists. A `setup` command that also laid down
personal rules would duplicate `scratchpad init`, and a command that installs
skills is not a command that sets up a workspace — `init` is.

**Auto-update stale skills rather than refuse.** The alternative — refuse and
require `--update` — protects local edits at the cost of leaving workspaces on
stale skills by default, which is the failure this whole change exists to fix. A
skill's exit-code semantics are matched to a CLI version; a stale skill tells an
agent the wrong thing about the tool it is driving. Local edits are still
protected in the case that matters: a team that bumps the version keeps its
file.

**Never downgrade, and say why.** A workspace ahead of the binary is a stale
CLI, not a stale workspace. The warn names the CLI as the thing to upgrade,
because an agent reading "kept" without a reason would treat it as a failure to
install.

**Version, not content hash.** A hash would detect any edit and could refuse to
touch modified files. It would also make "unchanged" impossible to distinguish
from "edited", and every local edit would become a permanent block on updates.
The version is a declaration of intent; a hash is an observation about bytes.

## Out of Scope

- Personal rules, `AGENTS.md`, or any other agent-facing file that is not a
  canonical skill.
- Uninstalling, merging, or diffing against a modified local skill.
- Installing to the platform or team layer.
- Any change to discovery, gate 7, or the layering model.
- Packaging the skills for an agent runtime — that is the plugin.
