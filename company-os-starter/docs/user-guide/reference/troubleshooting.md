---
title: Troubleshooting
---

# Troubleshooting

Symptom → cause → fix, for both tools. If your symptom isn't here, check the
task-specific "When things go sideways" table on the relevant how-to page
first — this page collects the cross-cutting ones.

## Company OS

| Symptom | Cause | Fix |
|---|---|---|
| `validate` gate `[1/N]` fails: ownership mismatch | a team's `ownership/components.yaml` claims `accountable`, but the component descriptor's `ownership.accountableTeam` disagrees | edit the component descriptor — it's the single source of truth, not the team registry |
| `validate` gate `[2/N]` fails: expired deviation or exception | a deviation's `reviewDate`, or an exception's `expires`, is in the past | re-declare / re-request with a fresh date, or remove it if it no longer applies |
| `validate` gate `[3/N]` fails: PRD contract | an active PRD is missing `title`, `team`, `components`, or `governanceSnapshot` in frontmatter | these are written automatically by `prd new` — if missing, the PRD was likely hand-created or hand-edited; restore the fields |
| `validate` gate `[4/N]`, `[5/N]`, or `[6/N]` fails: drift ("run: company-os graph build") | committed derived content (tags, `CLAUDE.md` context node, or a platform's feature-index) no longer matches what would be freshly derived | `company-os graph build`, review the diff, commit it |
| `validate` gate `[7/N]` fails: skill shadowing or dangling `extends` | a team/personal skill reuses a canonical skill's id or name, or its `extends: platform-skill://...` points at nothing | rename the conflicting skill, or fix/remove the broken `extends` |
| `validate` gate `[8/N]` fails: federated slice integrity | a materialized governance slice under the federation cache was hand-edited instead of coming from `workspace sync` | discard the hand-edit, re-run `company-os workspace sync` |
| `prd complete` refuses with a done-check error | a governance checklist item in the PRD is still `- [ ]`, or a listed component's reality doc has an `updated:` date older than the PRD's `created:` date | check the remaining boxes for real, or update the reality doc and its `updated:` date — this gate exists specifically so "done" always means reality changed |
| `governance resolve` or `discover`/`prd`/`check` dies with a team-not-found-style error | `--team` omitted or misspelled — most of these commands only look optional in `--help`, but functionally need a real team id | pass `--team <id>`; confirm the id with `company-os ids list --prefix team://` (or check `teams/`) |
| `company-os workspace ...` dies immediately | no `workspace.yaml` manifest at the workspace root | either this workspace is a plain monorepo (no federation needed — just use `validate` directly), or add a manifest per `docs/FEDERATION-RUNBOOK.md` |
| `workspace sync --frozen` dies | no `workspace.lock.yaml` yet, or the lock doesn't cover a repo in the manifest | run `workspace sync` once without `--frozen` (needs network + git ≥2.27) to produce a lock, then `--frozen` works offline |

## Local Search

| Symptom | Cause | Fix |
|---|---|---|
| "No repos added yet" | nothing registered | `local-search repo add <folder> [name]` |
| A search finds nothing you expect | wrong repo paths, or the doc isn't actually indexed | check `local-search repo list` and `local-search stats`, broaden the query, then `local-search scan` |
| Results look stale | automatic change detection missed something (e.g. a large bulk change) | `local-search scan` (or `local-search scan <repo>` for just one) |
| Weird/corrupt-looking results, or errors reading the index | a corrupted cache database | `rm ~/.local-search/specs.db` (safe — disposable cache) or `local-search reset` (also clears repo registrations; prompts first) |
| A file you know exists is never found | only `.md`, `.mdx`, `.txt` are indexed directly; other files need a `.md` sidecar | add a sidecar describing it, or convert the file |
| Change detection isn't triggering for a repo | git-based freshness needs the repo to have at least 1 commit and `git` on `PATH` | commit at least once; confirm `git` is installed and reachable |

## See also

- [reference/company-os-cli.md](company-os-cli.md)
- [reference/configuration.md](configuration.md)
- [how-to/run-the-validation-gate.md](../how-to/run-the-validation-gate.md)
- [how-to/keep-search-fresh.md](../how-to/keep-search-fresh.md)
