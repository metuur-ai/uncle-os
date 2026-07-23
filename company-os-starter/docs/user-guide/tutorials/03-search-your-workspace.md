---
title: Search your workspace with Local Search
---

# Search your workspace with Local Search

Moonbeam's `moonbeam-os` workspace is growing — discovery briefs, PRDs,
reality docs, ADRs. [Local Search](https://github.com/metuur-ai/local-search)
is a fast, offline spec registry built for exactly this: a single Go binary
with no runtime dependencies that indexes your project documentation across
multiple repos with SQLite FTS5 and BM25 ranking. No MCP server — that's a
deliberate choice, in favor of a CLI plus a Claude skill.

This tutorial only covers how it plugs into a Company OS workflow. For
everything else — the full command reference, architecture, flags — go to
[Local Search's own repo and README](https://github.com/metuur-ai/local-search).

## 1. Install it

```bash
curl -fsSL https://raw.githubusercontent.com/metuur-ai/local-search/main/install.sh | bash
```

This installs the CLI to `~/.local/bin`, a Claude skill to
`~/.claude/skills/local-search`, and a web UI to
`~/.local/share/local-search/web`.

> **Note:** the web UI needs Node ≥ 18. If it's missing, the installer skips
> that piece with a warning and everything else still works — you can add
> the web UI later once Node is available. Pre-built binaries are also on
> the project's GitHub releases if you'd rather skip the script.

## 2. Register the Moonbeam Bakery workspace

Local Search indexes whatever repos you register — it auto-scans just the
folder you point it at:

```bash
$ local-search repo add ~/moonbeam-os moonbeam-os
```

Check it landed:

```bash
$ local-search repo list
```

`repo remove <name>` un-registers a repo just as surgically, without
touching anything else in the index.

## 3. Run your first searches

```bash
$ local-search search "same-day pickup"
```

That's a plain hybrid search (FTS5 + graph-aware ranking) across every
registered repo. Add `--repos moonbeam-os` to scope it, or `--semantic` for
the embedding-assisted path.

> **Warning:** `local-search find` is a **different command**, not an alias
> of `search`. `find` does unified scoped search across both specs and code
> graphs:
> ```bash
> $ local-search find "same-day pickup" --scope moonbeam-os
> ```
> Reach for `search` for a quick keyword lookup and `find` when you want
> specs and code considered together in one scoped query.

Once you've found something:

```bash
$ local-search read <name>
```

`related`, `tags`, `recent`, and `stats` round out discovery once you're
past the first search — see the project README for the full set.

## 4. Wire it up for Claude

```bash
$ local-search install-skill --global
```

This installs the skill so any Claude Code session can search your
registered repos directly. It refuses to overwrite an existing skill install
unless you pass `--force`. Prefer a project-local skill instead of a global
one? Use `install-skill --local`, which writes to `./.claude/skills`.

Before every query, the skill resolves which repos are in scope for the
*current* project by running `local-search init --json` — worth knowing
before you go looking for the config file yourself (see
[reference/configuration.md](../reference/configuration.md) for exactly
where that scope lives and how it differs from the `.local-search.toml`
scope file used by `find`).

## 5. Keep it fresh

Local Search auto-detects changes on every query (git HEAD plus
staged/unstaged/untracked files, respecting `.gitignore`), so day-to-day
editing of Moonbeam's discovery briefs and PRDs just works. For the full
story on forcing a rescan, installing git hooks so it never goes stale, and
scoping which repos a given project searches, see
[how-to/keep-search-fresh.md](../how-to/keep-search-fresh.md).

## Where to next

- [how-to/keep-search-fresh.md](../how-to/keep-search-fresh.md) — scanning,
  scan-hooks, and the two different scope files.
- [reference/configuration.md](../reference/configuration.md) — every file
  and env var both tools read, side by side.
- [reference/troubleshooting.md](../reference/troubleshooting.md) — "no
  repos added yet," stale results, a corrupt index, and more.
