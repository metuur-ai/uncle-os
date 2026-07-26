---
title: Sync a knowledge catalog
---

# Sync a knowledge catalog

Component repos hold documentation you want an agent to read — `docs/sdd`,
`specs/`, `architecture/` — mixed in with source code you do not. A knowledge
slice pulls only the documentation directories into `knowledge/`, pinned and
hash-locked, without cloning or indexing the rest of the repo.

This is the same machinery as governance federation, pointed at a different
root. If you have not federated before, read
[FEDERATION-RUNBOOK.md](../../FEDERATION-RUNBOOK.md) first.

## Add a repo to the catalog

Declare it in `workspace.yaml` at the workspace root:

```yaml
version: 1
repos:
  - name: component-library
    url: https://github.com/acme/component-library.git
    pin: {tag: v1.2.0}
    slices:
      - {paths: [docs/sdd],       localDirectory: knowledge/components/component-library}
      - {paths: [architecture],   localDirectory: knowledge/architecture/component-library}
      - {paths: [.claude/skills], localDirectory: knowledge/skills/component-library}
```

One repo, one clone, one cache, three destinations. A repo contributing a
single area can use the flat form instead — `localDirectory:` and `paths:`
directly on the entry, no `slices:` key.

Then:

```bash
$ company-os workspace sync
$ company-os graph build     # refresh knowledge/CLAUDE.md so agents can find it
$ company-os validate
```

Commit three things together: the manifest, the materialized slices, and
`workspace.lock.yaml`. The lock diff is the record of which content moved.

## What lands, and where

`paths:` is an **allowlist**, not a filter — there is no exclude form. Anything
not named is never copied, which is what keeps source code out. The path below
the destination mirrors the source, so `paths: [docs/sdd]` into
`knowledge/components/component-library` lands at:

```text
knowledge/components/component-library/docs/sdd/...
```

There is no rename. If you want a different shape locally, change it upstream.

Two rules the CLI enforces:

- **Target an area, not the catalog root.** `localDirectory: knowledge` is
  refused — `knowledge/` itself holds the generated context node.
- **Targets must not overlap.** Equal or nested destinations are refused across
  the whole manifest, because the outer slice's read-only pass would freeze the
  inner one and break the next sync.

## What the catalog is, and is not

`knowledge/` is **indexed, not governed**. That distinction is deliberate:

| | |
|---|---|
| It gets | a generated `CLAUDE.md` context node listing every area and document, cross-links from every sibling root, and gate `[8/8]` hash integrity |
| It skips | `graph build` tag derivation and validate gates `[1/8]`–`[7/8]` |

The reason is that knowledge slices come from repos that are not Company OS
workspaces. Their docs carry no `type:`/`id:` frontmatter, so the frontmatter
gate would reject every one of them; and the materialized slice is `0444`, so
tag rewriting would fail against read-only files. The catalog is there for an
agent to read, not for the governance gates to police.

If you want a document *governed* — owned, reviewed, expiring — it belongs in a
platform or team root as a normal artifact, not in the catalog.

## Keeping it current

```bash
$ company-os workspace status
```

Run this before every sync. It reports three kinds of drift:

- **pin drift** — the manifest pin no longer matches the lock (you edited the
  pin; sync to apply it)
- **slice-set drift** — a target or allowlist changed without a re-sync. Worth
  knowing about because the old files still exist and still hash clean, so the
  file-level check alone would report green.
- **missing / drifted slices** — files absent, or hand-edited

To take a new upstream release: bump `pin:`, `sync`, `graph build`, commit.

**Never edit a synced file.** Slices are `0444` derived content and gate
`[8/8]` fails on any hand-edit, naming the file. The fix is always upstream —
change the source repo, cut a new pin, re-sync.

## In CI

The committed slices and lock are already on disk after checkout, and gate
`[8/8]` is a pure hash check with no network and no git cache. So CI needs
nothing more than:

```yaml
- run: pip install pyyaml
- run: bin/company-os --root . validate
```

`sync --frozen` rebuilds from the local `.company-os/` cache, which is
git-ignored and therefore absent on a clean runner — prefer validating the
committed slices directly. See
[FEDERATION-RUNBOOK.md](../../FEDERATION-RUNBOOK.md) §5 for the source-repo
variant.

## Related

- [FEDERATION-RUNBOOK.md](../../FEDERATION-RUNBOOK.md) — the full federation model
- [reference/company-os-cli.md](../reference/company-os-cli.md) — `workspace` command reference
- [how-to/run-the-validation-gate.md](run-the-validation-gate.md) — what each gate checks
- `skill://governance/syncing-knowledge` — the canonical skill for this workflow
