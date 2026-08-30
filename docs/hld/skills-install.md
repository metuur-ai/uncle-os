# `skills install` — High-Level Design

## Overview

The CLI installs as a single binary and nothing else. `make install` and
`install.sh` both put `company-os` on PATH; neither writes a skill, a rule, or
any other agent-facing file into the directory the user then works in. The five
canonical skills ship as reference copies under `company-os-starter/skills/`,
which `company-os init` does not copy and the binary never reads.

So a workspace created by `company-os init` has an `agentSkills` block in its
`team.yaml` naming `skills/` as the canonical path, and no `skills/` directory.
The pointer is real; the thing it points at is not there. Getting the skills in
means finding the starter kit, knowing the nested `<name>/SKILL.md` layout is
undiscoverable, and renaming each file on the way in — three steps documented in
prose and enforced by nothing.

This change embeds the skills in the binary and adds one command that writes
them where discovery looks.

## Stakeholders & Impact

**Users setting up a workspace.** One command replaces a copy-and-rename
procedure they currently have to read about first. This is the whole point.

**Agents operating a workspace.** A session's merged skill view stops depending
on whether someone remembered to copy files. `skills list` is the same command
before and after; it just has something to report.

**Teams that edited an installed skill.** This is the risk the change
introduces. A skill is a starting point teams are expected to edit, and an
update overwrites the file. Version comparison bounds it — same version means
untouched, newer means left alone — but a team that edits without bumping the
version loses those edits on the next update. The command says what it replaced
and with which versions; it does not merge.

**Workspace authors on an older CLI.** A binary older than the workspace leaves
every skill alone and warns. The stale component is named as the CLI, because
that is what it is.

## Goals

1. The canonical skills ship inside the binary, so an installed CLI can write
   them without a checkout.
2. One command installs them into `company-os/skills/`, where discovery looks.
3. An installed skill older than the binary's is updated; the same version is
   left alone; a newer one is never downgraded.
4. A file whose version cannot be read is never overwritten.
5. What the command writes leaves the workspace validating clean, including the
   derived artifacts the new files change.
6. Running it twice reports no change the second time.

## Non-Goals

- **No personal-rules scaffolding.** `scratchpad init` already creates
  `personal-rules/`, and duplicating it here would give two commands one job.
- **No platform or team layer.** Every canonical skill declares company-wide
  applicability, and derivation gives a company-layer skill exactly the tags
  these files carry. Installing to a platform would put them in tag drift the
  moment they landed.
- **No merge.** A team's edits and a new upstream version are reconciled by the
  team, not by the tool. The command reports; it does not resolve.
- **No uninstall.** Nothing here removes a file.
- **No new frontmatter field.** Version comparison uses the `version:` key the
  skills already carry.

## Success Criteria

Observable when this ships:

- `company-os skills install` on a fresh `init` workspace writes five skills and
  exits 0.
- `company-os validate` exits 0 immediately afterwards, with no `graph build` in
  between.
- A second run reports every skill unchanged and 0 changed.
- An installed skill at an older version is replaced, and the finding names both
  versions.
- An installed skill at a newer version survives, with a warn naming the CLI as
  the stale component.
- An installed file with unreadable frontmatter survives byte-for-byte.
- `company-os skills list` reports the installed skills at the company layer.
- `company-os --json skills install` carries one finding code per skill and the
  next command in `guidance`.
