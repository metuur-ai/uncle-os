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
| `validate` gate `[8/N]` fails: federated slice integrity | a materialized slice was hand-edited instead of coming from `workspace sync` | discard the hand-edit, re-run `company-os workspace sync`. Slices are `0444` derived content — fix the source repo, bump the pin, re-sync |
| `validate` gate `[8/N]` fails: "slice set ... differs" | a `localDirectory:` or `paths:` changed in the manifest without a re-sync. The old files are still on disk and still hash clean, so only this check catches it | `company-os workspace sync`, then commit the moved slices and the new lock |
| `prd complete` refuses with a done-check error | a governance checklist item in the PRD is still `- [ ]`, or a listed component's reality doc has an `updated:` date older than the PRD's `created:` date | check the remaining boxes for real, or update the reality doc and its `updated:` date — this gate exists specifically so "done" always means reality changed |
| `governance resolve` or `discover`/`prd`/`check` dies with a team-not-found-style error | `--team` omitted or misspelled — most of these commands only look optional in `--help`, but functionally need a real team id | pass `--team <id>`; confirm the id with `company-os ids list --prefix team://` (or check `teams/`) |
| `company-os workspace ...` dies immediately | no `workspace.yaml` manifest at the workspace root | either this workspace is a plain monorepo (no federation needed — just use `validate` directly), or add a manifest per `docs/FEDERATION-RUNBOOK.md` |
| `workspace sync --frozen` dies | no `workspace.lock.yaml` yet, or the lock doesn't cover a repo in the manifest | run `workspace sync` once without `--frozen` (needs network + git ≥2.27) to produce a lock, then `--frozen` works offline |
| manifest rejected: "uses `root:` — renamed to `localDirectory:`" | the destination key was renamed | rename `root:` to `localDirectory:` on the repo entry (or inside each `slices:` entry) |
| manifest rejected: "cannot set both `slices:` and top-level ..." | a repo declares `slices:` *and* a top-level `localDirectory:`/`paths:`, so the top-level key would apply to nothing | move it into a `slices:` entry, or drop `slices:` and keep the flat form |
| manifest rejected: "slice target ... overlaps ..." | two destinations are equal or nested, anywhere in the manifest | give each slice a disjoint target. Nested targets are refused because the outer slice's read-only pass freezes the inner one and breaks the next sync |
| manifest rejected: "must name an area under `knowledge/`" | `localDirectory: knowledge` targets the catalog root itself, which holds the generated context node | target an area beneath it, e.g. `knowledge/components/<repo>` |
| synced docs don't appear in `knowledge/CLAUDE.md` | the index is generated, not live | run `company-os graph build` after each sync |
| a knowledge doc isn't validated / has no derived tags | by design — `knowledge/` is indexed, not governed. Foreign docs have no `type:` frontmatter and the slice is read-only | if the doc needs governing, it belongs in a platform or team root as a normal artifact — see [how-to/sync-a-knowledge-catalog.md](../how-to/sync-a-knowledge-catalog.md) |
| `company-os --version` prints a `usage: company-os [-h] [--root ROOT]` banner and exits 2, instead of a version line | you are not running the binary. A leftover launcher from the old Python `install.sh` is earlier on your `PATH` and is shadowing it — the binary may be installed and still never run | `type -a company-os` to see every match in resolution order, then delete the one that is not your binary and its kit root — full procedure: [how-to/release-and-upgrade.md](../how-to/release-and-upgrade.md#migrating-off-the-python-kit) |
| a new binary was installed but behavior did not change; `--json` is rejected as an unknown flag; exit codes are only ever `0` or `1` | same cause — the stale Python launcher is winning on `PATH`. It has no `--json` and no exit-code contract, so anything written against either fails against it | as above: `type -a company-os`, remove the launcher and `$COMPANY_OS_PREFIX/share/company-os/` |
| `company-os` dies with `python3: can't open file '.../share/company-os/bin/company-os': [Errno 2] No such file or directory`, exit 2 | the old kit root was deleted but its generated launcher was left on your `PATH` — a half-finished migration | `rm -f ~/.local/bin/company-os` (or `$COMPANY_OS_PREFIX/bin/company-os`), then install the binary |

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
