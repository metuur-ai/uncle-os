---
title: Configuration reference
---

# Configuration reference

Every path and env var these tools actually read — split by tool, and by
what's implemented today versus specified-but-not-wired.

## Company OS: workspace-root precedence

`docs/TUTORIAL.md` §0.5 specifies six layers, highest wins:

```text
1. CLI flag              company-os --root /abs/path ...
2. Environment variable  $COMPANY_OS_WORKSPACE_ROOT
3. Repo-local override   .company-os.local.yaml        (git-ignored)
4. User-level config     ~/.company-os/config.yaml     (outside every repo)
5. Committed shared      config/repositories.yaml      (relative dirs only)
6. Built-in default      current working directory
```

The CLI **only implements layers 1, 2, and 6** — `--root`,
`$COMPANY_OS_WORKSPACE_ROOT`, and the cwd fallback.
Layers 3–5 (`.company-os.local.yaml` merging, `~/.company-os/config.yaml`,
and a `workspace sync` that clones repos into relative directories from a
committed `config/repositories.yaml`) are specified in the design but not
yet wired into the CLI — this is the CLI's own tutorial admitting the gap,
not a discrepancy we're pointing out from outside. If you write those files
today, nothing reads them.

**In practice:** for a single machine, set `COMPANY_OS_WORKSPACE_ROOT` once
and run `company-os` from anywhere:

```bash
export COMPANY_OS_WORKSPACE_ROOT=/Users/you/work/moonbeam-os
company-os governance resolve --team web
```

Or pass `--root` inline without exporting anything —
`company-os --root /Users/you/work/moonbeam-os validate`.

## Company OS: federation manifest and lock (implemented)

Unlike the layers above, multi-repo federation (Phase 4) is real and wired
in. Two files at the workspace root:

| File | Written by | Purpose |
|---|---|---|
| `workspace.yaml` | you, by hand | declares pinned repos/refs and where each lands; its mere presence switches `validate` to 8 gates and enables `company-os workspace ...` |
| `workspace.lock.yaml` | `company-os workspace sync` | generated — records exactly what was materialized, checked against on every `validate` (gate `[8/N]`) and every `workspace status` |

Each repo names its destination with `localDirectory:` and what to pull with
`paths:`. A repo contributing several areas uses a `slices:` list of
`{paths, localDirectory}` pairs instead, sharing one clone and one cache.
Destinations must land under a canonical root — `company-os/`, `platforms/`,
`teams/`, `company-ontology/`, or `knowledge/` — and must not overlap each
other.

`company-os workspace sync` also writes a git-ignored cache at
`.company-os/federation-cache/`. Full walkthrough:
[docs/FEDERATION-RUNBOOK.md](../../../docs/FEDERATION-RUNBOOK.md); the
catalog specifically:
[how-to/sync-a-knowledge-catalog.md](../how-to/sync-a-knowledge-catalog.md).

## Installation: nothing to configure

`company-os` is a single static binary. It reads no installation config,
resolves nothing relative to itself, and has no kit root — the templates it
scaffolds from are compiled in. Installing it is: download the artifact for
your platform, `chmod +x`, move it onto your `PATH`. Uninstalling it is
deleting that one file.

Building from source instead? `make install` from `company-os-starter/` honors
one variable:

```bash
PREFIX=/some/where make install      # defaults to $HOME/.local
```

which puts the binary at `$PREFIX/bin/company-os` and nothing anywhere else.
`make uninstall` removes it.

> **Migrating from the Python kit.** The old `install.sh` wrote a bash
> launcher at `$COMPANY_OS_PREFIX/bin/company-os` that `exec`'d a kit root
> under `$COMPANY_OS_PREFIX/share/company-os/`. Both are gone. Delete the
> stranded launcher and the kit directory, then install the binary:
>
> ```bash
> rm -f  ~/.local/bin/company-os          # or $COMPANY_OS_PREFIX/bin/company-os
> rm -rf ~/.local/share/company-os        # the old kit root
> ```
>
> `COMPANY_OS_PREFIX` is no longer read by anything.
>
> Do not skip this because the old command still works — it will. The kit was
> installed as a self-contained copy, so a leftover launcher keeps running the
> Python CLI forever and can silently shadow the binary on your `PATH`. Full
> procedure and how to verify it took:
> [how-to/release-and-upgrade.md](../how-to/release-and-upgrade.md#migrating-off-the-python-kit).

## Local Search: two different scope files

Don't confuse these — they're managed by different subcommands and read by
different consumers:

| File | Managed by | Used by | Format |
|---|---|---|---|
| `.agent/local-search-config.yaml` | `local-search init` / `setup` (`--add`/`--remove`/`--set`) | the Claude skill | YAML `repositories:` |
| `.local-search.toml` | `local-search scope show / set / clear / init` | `find` and `code` | TOML `scope = []` |

If your workflow is `search`/`read`/`related` plus the Claude skill, you'll
only ever touch `.agent/local-search-config.yaml`. If you use `find`, you
need `.local-search.toml` too. See
[how-to/keep-search-fresh.md](../how-to/keep-search-fresh.md) for the
commands that write each one.

## Local Search: data directory

```text
~/.local-search/
  repos            # registry of repos added via `local-search repo add`
  specs.db         # SQLite FTS5 cache — disposable, safe to delete
```

Deleting `specs.db` (or running `local-search reset`, which also clears repo
registrations and prompts first) never loses source data — it's a rebuildable
index over files that remain the source of truth.

## See also

- [reference/company-os-cli.md](company-os-cli.md)
- [reference/troubleshooting.md](troubleshooting.md)
- [how-to/keep-search-fresh.md](../how-to/keep-search-fresh.md)
