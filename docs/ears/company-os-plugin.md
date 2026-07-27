# Company OS Claude Code Plugin — EARS Specifications

## Unit 1: Package structure

**Why:** Claude Code discovers components at fixed default locations, and puts
them at the plugin root rather than inside the metadata directory. Getting the
layout wrong means the plugin loads with nothing in it.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL provide a plugin manifest at `.claude-plugin/plugin.json` declaring a `name` of `company-os`, plus a description, a version, and a repository. |
| R-1.2 | THE SYSTEM SHALL provide a marketplace entry at `.claude-plugin/marketplace.json` listing the plugin with its source path. |
| R-1.3 | THE SYSTEM SHALL place plugin skills at the repository root under `skills/<name>/SKILL.md`. |
| R-1.4 | THE SYSTEM SHALL NOT place `skills/` or any other component directory inside `.claude-plugin/`. |
| R-1.5 | THE SYSTEM SHALL NOT create any plugin component directory at the repository root other than `skills/`. |
| R-1.6 | THE SYSTEM SHALL NOT provide any slash command, and SHALL NOT create a `commands/` directory. |

## Unit 2: The skills

**Why:** These four carry the judgment the CLI cannot encode — what an exit code
means here, which warning on a passing run is load-bearing, which shortcut past a
gate is forbidden. Packaging them is the entire value of this change; the naming
is what keeps them findable in a crowded session.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL provide one plugin skill for each of the four canonical starter-kit skills. |
| R-2.2 | THE SYSTEM SHALL name each plugin skill's directory `cos-<name>`, where `<name>` is the canonical skill's name. |
| R-2.3 | THE SYSTEM SHALL give each plugin skill frontmatter containing a `name` equal to its directory name and a description. |
| R-2.4 | THE SYSTEM SHALL state in each plugin skill's body the workspace-relative path of the canonical skill file it derives from. |
| R-2.5 | THE SYSTEM SHALL keep each plugin skill's body a summary that defers to its cited source, and SHALL NOT let it become a second authoritative copy of the procedure. |
| R-2.6 | THE SYSTEM SHALL instruct in each plugin skill's body that every CLI command be run with JSON output and branched on by exit code, matching the canonical skills' own contract. |
| R-2.7 | THE SYSTEM SHALL NOT generate the plugin skills from their canonical sources, and SHALL NOT add a command to do so. |

## Unit 3: Precedence against a workspace

**Why:** A workspace resolves competing skills through canonical-over-personal
ordering, `extends:` layering, and a shadowing gate. The plugin has no equivalent
and no mechanism here can give it one, so an agent can hold the plugin's generic
procedure alongside a workspace's customized one with nothing to say which wins.
This is the one thing the change genuinely risks breaking.

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL state in each plugin skill's body that a workspace's own skill for the same task is authoritative over the plugin's. |
| R-3.2 | THE SYSTEM SHALL direct the reader in each plugin skill's body to the merged skills listing as the way to discover what a workspace actually layers. |
| R-3.3 | THE SYSTEM SHALL record this precedence gap as a known limitation in the plugin's documentation, rather than implying the workspace layering model extends to plugin-provided skills. |
| R-3.4 | THE SYSTEM SHALL NOT alter the workspace layering model, its shadowing detection, or its extension resolution. |

## Unit 4: Verifiability

**Why:** Every other gate in this repository is hermetic and offline. Binding the
build to an external tool's warning output would let a release of that tool turn
a green repository red with no code change.

| ID | EARS statement |
| --- | --- |
| R-4.1 | WHEN the plugin validator runs against the repository, THE SYSTEM SHALL report no errors. |
| R-4.2 | THE SYSTEM SHALL provide an offline test asserting the manifest's required fields and their types, so a malformed manifest fails the standard build without any external tool. |
| R-4.3 | THE SYSTEM SHALL NOT make any external tool's warning output a condition of the build passing. |
| R-4.4 | THE SYSTEM SHALL provide a test asserting that every canonical source path cited by a plugin skill exists. |

## Unit 5: Isolation and non-regression

**Why:** The plugin adds a root `skills/` directory to a repository whose own
`skills/` directories mean something else entirely. The separation has to be
structural, not a coincidence of the current layout.

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL keep the repository root free of the directories that would let it resolve as a workspace root, so the plugin's `skills/` can never be reached by Company OS skill discovery. |
| R-5.2 | THE SYSTEM SHALL provide a test asserting that skill discovery run against the repository root finds no plugin skill. |
| R-5.3 | THE SYSTEM SHALL NOT modify any file under `examples/`. |
| R-5.4 | THE SYSTEM SHALL NOT modify any Go source outside the new tests. |
| R-5.5 | THE SYSTEM SHALL NOT rename any skill inside any workspace. |
| R-5.6 | THE SYSTEM SHALL NOT add, remove, renumber, or alter any validation gate, finding code, message, or severity. |
| R-5.7 | THE SYSTEM SHALL NOT change any command's argument surface. |
| R-5.8 | THE SYSTEM SHALL NOT modify any golden snapshot. |
| R-5.9 | WHEN `make check` runs, THE SYSTEM SHALL complete gofmt, vet, `go test ./...`, and `examples/acceptance.sh` without failure. |
