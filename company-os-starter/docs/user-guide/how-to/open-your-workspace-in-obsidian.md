---
type: doc
tags: [doc/company-os-starter, kind/how-to]
---

# Open your workspace in Obsidian

There is no plugin, no export, and no sync step. The vault *is* the
workspace directory. This takes about a minute.

## 1. Open the folder as a vault

In Obsidian: **Open folder as vault** → select your workspace root (the
directory containing `company-os/`, `platforms/`, `teams/`, and
`company-ontology/`).

Do not point Obsidian at a copy. The whole benefit is that the files you
edit are the files Git tracks and `company-os validate` checks.

## 2. Ignore Obsidian's own state in Git

Obsidian writes workspace layout, open panes, and plugin settings into
`.obsidian/`. That is per-person noise. Add it to the workspace's
`.gitignore`:

```
.obsidian/
```

If you want to share editor settings across the team, commit
`.obsidian/app.json` and `.obsidian/appearance.json` and ignore the rest —
`.obsidian/workspace.json` in particular changes every time anyone moves a
pane.

## 3. Run `derive` once, so the graph has edges

```bash
company-os derive
```

This writes the frontmatter `tags:` Obsidian's tag pane groups by, and the
context-index blocks in each `CLAUDE.md` whose relative Markdown links
Obsidian's backlinks and graph view follow. Before you run it, a fresh
workspace opens as a pile of unlinked files.

## What you should see

- **Tag pane** — a nested tree: `kind/` → `prd`, `discovery`, `reality`;
  `team/` → your teams; `status/` → `draft`, `validated`, `completed`.
  Those come from YAML `tags:`; the `/` is Obsidian's nesting separator.
- **Graph view** — each `CLAUDE.md` sits at the centre of its layer, linking
  down to that layer's documents and across to sibling roots. The visible
  shape is the federation.
- **Backlinks** — on any PRD or reality doc, the platform `CLAUDE.md` that
  indexes it.

## Two things not to touch

**Never hand-edit `tags:`.** They are derived. Change the source field —
`type:`, `team:`, `platform:`, `status:` — and re-run `company-os derive`.
Gate `[4/7]` fails on the difference.

**Never type between the generated markers** in a `CLAUDE.md`:

```markdown
<!-- company-os:generated:start — do not edit; run `company-os derive` -->
...
<!-- company-os:generated:end -->
```

Gate `[5/7]` fails on an edited block. Everything *outside* the markers is
hand-owned and survives every rebuild — put your own notes there.

If you edited either by accident:

```bash
company-os validate --fix
```

That repairs the drift in place instead of only reporting it.

## Files that will look broken (they aren't)

- **`knowledge/`** — synced documentation slices are written `0444`.
  Obsidian will refuse to save edits there. Correct: change the source repo,
  bump the pin, re-run `company-os workspace sync`.
- **`generated/feature-index.yaml`** — YAML, so Obsidian shows it as an
  unopenable file. It is derived state for the CLI's gates, not a note.

## Using Obsidian and Local Search together

They are independent and can be attached in either order. Obsidian is for
wandering — follow a link, notice a neighbour, edit in place. Local Search
answers a question you can phrase but can't locate:

```bash
local-search repo add company-os .
local-search search "how do we handle delivery retries"
```

Local Search builds its graph from your frontmatter reference fields
(`components:`, `dependsOn:`, `upstream:`, `fromDiscovery:`), not from
Obsidian's link syntax — so neither tool needs the other to be installed.

See [../explanation/obsidian-and-local-search.md](../explanation/obsidian-and-local-search.md)
for what each tool reads and why, and
[keep-search-fresh.md](keep-search-fresh.md) for keeping the index current.
