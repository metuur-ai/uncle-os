# `skills install` — EARS Specifications

## Unit 1: Shipping the skills in the binary

**Why:** An installed CLI has no checkout. Skills that live only as files in the
repository cannot be written into a user's workspace by a binary on PATH, which
is why the copy step is manual today.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL embed every `*.SKILL.md` file at the root of `company-os-starter/skills/` into the binary. |
| R-1.2 | THE SYSTEM SHALL expose each embedded skill's name, declared `version:`, and bytes. |
| R-1.3 | THE SYSTEM SHALL return embedded skills in name order, so command output does not depend on filesystem iteration order. |
| R-1.4 | THE SYSTEM SHALL place the embed directive in a package at or above `company-os-starter/skills/`, and that package SHALL NOT import `internal/`. |
| R-1.5 | WHEN a skill file's frontmatter carries no readable `version:`, THE SYSTEM SHALL report its version as empty rather than assume a default. |

## Unit 2: The command

**Why:** The workspace `team.yaml` names `skills/` as the canonical path, and
`init` creates no such directory. One command closes the gap between the pointer
and the thing it points at.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL accept `install` as an action of the `skills` command. |
| R-2.2 | THE SYSTEM SHALL write embedded skills into `<root>/company-os/skills/`, creating the directory when absent. |
| R-2.3 | THE SYSTEM SHALL name each written file `<name>.SKILL.md`, the flat form skill discovery matches. |
| R-2.4 | THE SYSTEM SHALL NOT write to the platform layer, the team layer, or any path outside the destination directory. |
| R-2.5 | THE SYSTEM SHALL NOT delete any file. |
| R-2.6 | THE SYSTEM SHALL print the next command in the workflow, per the guidance chain (R-1.8 of the CLI port). |
| R-2.7 | WHEN the run completes, THE SYSTEM SHALL report how many skills it considered and how many it changed. |

## Unit 3: Version comparison

**Why:** A skill's exit-code semantics are matched to a CLI version, so a stale
skill tells an agent the wrong thing about the tool it is driving. A workspace
ahead of the binary is the opposite problem and must not be treated the same
way.

| ID | EARS statement |
| --- | --- |
| R-3.1 | WHEN no file exists at a skill's destination, THE SYSTEM SHALL write it and report `skills.installed`. |
| R-3.2 | WHEN an installed skill's version is lower than the embedded one's, THE SYSTEM SHALL overwrite it and report `skills.updated` naming both versions. |
| R-3.3 | WHEN an installed skill's version equals the embedded one's, THE SYSTEM SHALL leave the file unmodified and report `skills.unchanged`. |
| R-3.4 | WHEN an installed skill's version is higher than the embedded one's, THE SYSTEM SHALL leave the file unmodified, report `skills.locally-newer` at warn severity, and name the CLI as the component to upgrade. |
| R-3.5 | WHEN an installed file's version cannot be read, THE SYSTEM SHALL leave it unmodified and report `skills.unreadable` at warn severity. |
| R-3.6 | THE SYSTEM SHALL compare versions numerically per dot-separated segment, so `1.10` ranks above `1.9`. |
| R-3.7 | THE SYSTEM SHALL emit one finding code per outcome rather than one code carrying an outcome field. |

## Unit 4: Derived artifacts and the green-workspace guarantee

**Why:** Writing into `company-os/skills/` changes a graph-docs root, so the
company `CLAUDE.md` context node and the directory index go stale and gate 5
fails. A command that adds the files the CLI itself ships must not redden a
workspace that was green.

| ID | EARS statement |
| --- | --- |
| R-4.1 | WHEN `skills install` completes, THE SYSTEM SHALL rebuild derived artifacts before returning. |
| R-4.2 | THE SYSTEM SHALL rebuild even when no file changed. |
| R-4.3 | THE SYSTEM SHALL emit the rebuild's lines before the command's own output. |
| R-4.4 | WHEN `company-os validate` runs against a workspace that validated clean before the install, THE SYSTEM SHALL exit 0 afterwards with no intervening `graph build`. |
| R-4.5 | THE SYSTEM SHALL declare the rebuild seam in `internal/skills` and satisfy it in `cmd/`, so `internal/skills` does not import `internal/graph`. |

## Unit 5: Conformance of what is shipped

**Why:** The four legacy starter-kit skills carried no `type:` and hand-written
tags, so copying one into a workspace produced two gate-4 failures. A command
that writes them at scale makes that defect everyone's, not just a careful
reader's.

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL ship every canonical skill flat, as `skills/<name>.SKILL.md`. |
| R-5.2 | THE SYSTEM SHALL give every shipped skill a `type: skill` key. |
| R-5.3 | THE SYSTEM SHALL give every shipped skill only the tags derivation produces at the company layer. |
| R-5.4 | THE SYSTEM SHALL provide test coverage that copies every shipped skill into a synthesized workspace and asserts discovery finds it and no core-field or tag-drift finding is raised against it. |
| R-5.5 | THE SYSTEM SHALL provide test coverage asserting that the files `skills install` writes draw no core-field or tag-drift finding. |

## Unit 6: Non-regression

| ID | EARS statement |
| --- | --- |
| R-6.1 | THE SYSTEM SHALL leave `skills list` behavior unchanged. |
| R-6.2 | THE SYSTEM SHALL NOT alter skill discovery, the shadowing gate, or extends resolution. |
| R-6.3 | THE SYSTEM SHALL NOT add, remove, renumber, or alter any validation gate. |
| R-6.4 | THE SYSTEM SHALL update the frozen usage and invalid-choice assertions for the `skills` command only, leaving other commands' identical-looking assertions untouched. |
| R-6.5 | WHEN `make check` runs, THE SYSTEM SHALL complete gofmt, vet, `go test ./...`, and `examples/acceptance.sh` without failure. |
