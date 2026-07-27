# Starter-Kit Skills Repair — EARS Specifications

## Unit 1: File layout

**Why:** Skill discovery globs `*.SKILL.md` one directory level deep. The four
starter-kit skills sit at `<name>/SKILL.md`, a layout that glob can never match
— so the reference implementation demonstrates a shape its own tool cannot see,
and `templates/SKILL-template.md` already documents the opposite as required.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL place each starter-kit skill at `skills/<name>.SKILL.md`, directly inside the skills directory with no intervening subdirectory. |
| R-1.2 | THE SYSTEM SHALL remove each directory emptied by that move. |
| R-1.3 | THE SYSTEM SHALL preserve each skill's existing name, changing only its position and extension. |
| R-1.4 | THE SYSTEM SHALL leave the body of every starter-kit skill unchanged. |
| R-1.5 | THE SYSTEM SHALL NOT widen the discovery glob to accommodate the nested layout. |

## Unit 2: Frontmatter validity

**Why:** None of the four carries a `type:` key, which the frontmatter gate
requires of every artifact, and all four hand-write tags that derivation does not
emit. Copying one into a workspace — the only thing a starter kit is for —
produces two gate failures on a file the project itself shipped.

| ID | EARS statement |
| --- | --- |
| R-2.1 | THE SYSTEM SHALL give each starter-kit skill a `type:` key identifying it as a skill. |
| R-2.2 | THE SYSTEM SHALL set each starter-kit skill's `tags:` to exactly what tag derivation produces for a skill at the company layer. |
| R-2.3 | THE SYSTEM SHALL remove every hand-written tag facet that derivation does not emit. |
| R-2.4 | THE SYSTEM SHALL record in each starter-kit skill that derived tags gain a location facet at the platform and team layers, so a reader copying one below the company layer runs the graph build rather than assuming the committed tags are portable. |
| R-2.5 | THE SYSTEM SHALL NOT add any entry to the kind-tag map, and SHALL NOT change the derived tags of any artifact outside these four files. |

## Unit 3: The promise, made testable

**Why:** These files are read by no code path, so no golden, no gate, and no
existing test would notice them rotting again. The test added here is the entire
safety net.

| ID | EARS statement |
| --- | --- |
| R-3.1 | THE SYSTEM SHALL provide test coverage that, for each of the four starter-kit skills, copies it unchanged into the `company-os/skills/` directory of a synthesized workspace holding no skill of the same name. |
| R-3.2 | WHEN that copy is made, THE SYSTEM SHALL discover the file. |
| R-3.3 | WHEN that copy is validated, THE SYSTEM SHALL report no core-field finding against it. |
| R-3.4 | WHEN that copy is validated, THE SYSTEM SHALL report no tag-drift finding against it. |
| R-3.5 | THE SYSTEM SHALL assert the above for all four files, so a future edit to any one of them is covered. |

## Unit 4: Non-regression

**Why:** The value of this change is that it is small and provably contained.
Anything that moves a golden or touches the CLI belongs to a different change.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL NOT modify any file under `examples/`. |
| R-4.2 | THE SYSTEM SHALL NOT modify any golden snapshot. |
| R-4.3 | THE SYSTEM SHALL NOT add, remove, renumber, or alter any validation gate. |
| R-4.4 | THE SYSTEM SHALL NOT alter any finding code, message, or severity. |
| R-4.5 | THE SYSTEM SHALL NOT change any command's argument surface. |
| R-4.6 | THE SYSTEM SHALL NOT apply any name prefix to these or any other skills. |
| R-4.7 | WHEN `make check` runs, THE SYSTEM SHALL complete gofmt, vet, `go test ./...`, and `examples/acceptance.sh` without failure and without regenerating a golden. |
| R-4.8 | THE SYSTEM SHALL confine its diff to `company-os-starter/skills/` and the single new test file. |
