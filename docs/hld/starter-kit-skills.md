# Starter-Kit Skills Repair — High-Level Design

## Overview

`company-os-starter/skills/` holds four canonical skills — creating a PRD,
running discovery, completing a change, requesting an exception. They are the
reference implementations of the methodology's own procedures, and the thing a
new workspace is meant to copy from.

Copy one into a workspace today and it fails validation twice.

Each file sits at `skills/<name>/SKILL.md`. Skill discovery globs `*.SKILL.md`
one directory level deep, so the nested layout is invisible to it — the layout
the starter kit demonstrates is one the tool cannot see. Separately, none of the
four carries a `type:` key, which the frontmatter gate requires of every
artifact, and all four hand-write `tags:` containing facets that tag derivation
does not emit, which the tags-in-sync check rejects.

None of this is currently observable, because nothing reads these files. They
are reference copies with no reader: not embedded in the binary, not copied by
`company-os init`, and not on any discovery path, exactly like the sibling
`.md` files under `templates/` that the embed declaration describes as "human
reference copies with different placeholder text that the CLI has never read."

That is why the defect survived. The starter kit is only exercised by a human
copying from it, and a human who does gets two gate failures and no explanation.

This change fixes the four files. Nothing else.

## Stakeholders & Impact

**Anyone starting from the starter kit.** The intended path — copy a canonical
skill into `company-os/skills/`, adapt it — currently produces a workspace that
fails `validate` on an artifact the project itself shipped. After this change it
produces a clean workspace.

**Readers learning the conventions.** The starter kit teaches by example, and
the example currently contradicts the rules stated in `templates/SKILL-template.md`
and enforced by the frontmatter gate. Fixing the files makes the demonstration
and the rule agree.

**Not affected:** every workspace under `examples/`, the CLI, every gate, every
golden snapshot, and every test. These four files are read by nothing, so
changing them changes no behavior anywhere.

## Goals

1. Each of the four skills lives at `skills/<name>.SKILL.md` — flat, one level
   inside `skills/`, matching the layout discovery actually globs.
2. Each carries a `type:` key, satisfying the frontmatter gate's core-field
   requirement.
3. Each carries exactly the tags that derivation produces, with no hand-written
   facet derivation does not emit.
4. Copying any of the four unchanged into a workspace's `company-os/skills/`
   yields a file that discovery finds and validation passes.
5. `make check` passes unchanged, and no golden snapshot moves.

## Non-Goals

- **No renaming beyond the layout flattening.** The skills keep the names they
  have. Any prefix is a separate decision, specified separately.
- **No new validation.** Nothing is added to enforce this going forward; that is
  a different change with a different cost.
- **No change to skill content.** The procedures these files describe are
  correct. Only their frontmatter and file position move.
- **No change to `templates/`.** Its reference copies have the same
  no-reader status and the same latitude; they are out of scope here.
- **Nothing under `examples/`.** Those skills are already flat and already
  valid.

## Success Criteria

Observable when this ships:

- `ls company-os-starter/skills/` lists four `.SKILL.md` files and no
  directories.
- Copying any one of them into the `company-os/skills/` directory of a
  workspace that holds no skill of the same name, then running `validate`,
  reports the file as in sync and exits 0.
- A test performs exactly that copy for each of the four and asserts it.
- `git diff --stat` touches only files under `company-os-starter/skills/` and
  the new test.
- `make check` passes with no golden regeneration.
