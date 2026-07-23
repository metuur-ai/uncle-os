---
title: Keep Local Search fresh
---

# Keep Local Search fresh

Local Search re-checks freshness on every query — git HEAD plus
staged/unstaged/untracked changes (respecting `.gitignore`) for git repos,
file timestamps for non-git ones. Most of the time you don't need to do
anything. This page covers the cases where you do.

## Force a full or single-repo rescan

```bash
$ local-search scan            # full rebuild, every registered repo
$ local-search scan moonbeam-os  # surgical rebuild of just one repo
```

Reach for `scan` when results look stale despite the automatic
change-detection, or after a bulk operation (a rebase, a branch switch with
many changed files) you'd rather index deterministically than rely on
query-time diffing for.

## Never think about it again: git hooks

```bash
$ local-search scan-hooks install
```

Installs managed git hooks that keep the index current automatically as you
work, instead of relying purely on query-time detection.

## Scope which repos a query searches

There are **two different** per-project scope files — don't confuse them:

| File | Managed by | Used by |
|---|---|---|
| `.agent/local-search-config.yaml` | `local-search init` / `setup` (`--add`/`--remove`/`--set`) | the Claude skill |
| `.local-search.toml` | `local-search scope show / set / clear / init` | `find` and `code` |

If you only ever use `search`/`read`/`related` and the Claude skill, you'll
only ever touch the first one. If you use `find`, use the second.

```bash
$ local-search scope set moonbeam-os        # .local-search.toml
$ local-search init --add moonbeam-os       # .agent/local-search-config.yaml
```

See [reference/configuration.md](../reference/configuration.md) for the
full comparison.

## What's indexed

Only `.md`, `.mdx`, and `.txt` files are indexed directly. Images and PDFs
are picked up via a companion `.md` sidecar file, not directly.

## When things go sideways

| Symptom | Cause | Fix |
|---|---|---|
| "No repos added yet" | nothing registered | `local-search repo add <folder> [name]` |
| A search finds nothing you expect | wrong repo paths, or the doc isn't actually indexed | check `local-search repo list` and `local-search stats`, broaden the query, then `local-search scan` |
| Results look stale | change detection missed something (e.g. very bulk changes) | `local-search scan` |
| Weird or corrupt-looking results, or errors reading the index | a corrupted cache DB | `rm ~/.local-search/specs.db` (safe — it's a disposable cache) or `local-search reset` (also clears registrations, prompts before doing so) |
| A file you know exists is never found | it's not `.md`/`.mdx`/`.txt` and has no sidecar | add a `.md` sidecar describing it, or convert it |
| Change detection isn't triggering for a repo | the repo needs at least 1 commit, and `git` needs to be on `PATH` | commit at least once; confirm `git` is installed and reachable |
