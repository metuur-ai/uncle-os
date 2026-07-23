# Agent OS Graph Explorer

A single-file, zero-build web tool for exploring a knowledge base as an interactive force-directed graph. It renders files, and the projects, repos, and tags that connect them, so you can navigate the structure of an Agent OS workspace visually.

Everything lives in [`os-graph-explorer.html`](./os-graph-explorer.html) — no install, no bundler, no dependencies to vendor. Tailwind and the [`force-graph`](https://github.com/vasturiano/force-graph) library are loaded from a CDN at runtime.

## Quick start

The page fetches `sample.json` on load, and browsers block `fetch()` from `file://` URLs — so serve the folder over HTTP rather than double-clicking the file:

```bash
cd uncle-os
python3 -m http.server 8000
# open http://localhost:8000/os-graph-explorer.html
```

If you open the file directly, the auto-load is skipped (harmless) — use **Upload JSON** to load a graph, or **Test Scale (30k)** to generate a synthetic one.

## Loading data

Three ways to get a graph on screen:

| Action | Source |
|---|---|
| Auto-load | `sample.json` next to the HTML file, fetched on page load |
| **Upload JSON** | Any `.json` file from your machine |
| **Test Scale (30k)** | Generates 30k synthetic nodes to stress-test performance |

### Accepted formats

**1. Flat array of nodes** (the primary format — see [`sample-hub.json`](./sample-hub.json)). Community nodes for `repo`, `project`, and each `tag` are synthesized automatically, and files are linked to them:

```json
[
  {
    "id": "platforms/communications/platform.yaml",
    "name": "Communications Platform",
    "title": "Communications Platform",
    "type": "platform",
    "repo": "communications",
    "project": null,
    "tags": ["kind/platform", "platform/communications"]
  }
]
```

Fields: `id` (falls back to an index), `name`, `title`, `type` (`file` if omitted), `repo`, `project`, and `tags` (an array, or a comma-separated string). Only `id`/`name` are really needed; the rest enrich filtering and the inspector.

**2. Pre-built graph** — an object with `nodes` and either `links` or `edges`. Used as-is without synthesis:

```json
{ "nodes": [ ... ], "links": [ { "source": "a", "target": "b" } ] }
```

## Features

- **Search** — live full-text match over node names and titles (debounced for large graphs).
- **Filters** — multi-select dropdowns for type, repo, project, and tag, plus name/title contains-boxes. Filters combine; links are pruned to surviving nodes.
- **Node inspector** — click any node for its properties and a list of its connections; the view centers and zooms on it.
- **Hover highlight** — O(1) neighbor lookup lights up a node's immediate connections and dims the rest.
- **Labels** — level-of-detail rendering shows labels only when zoomed in or highlighted; toggle **Show all labels** to force them.
- **Node colors** — files (slate), repos (sky), projects (orange), tags (violet).

### Physics controls (bottom center)

Zoom in/out, fit-to-view, and a **Physics** toggle. On load the layout runs a short warmup and then settles automatically. Toggling physics re-heats the layout and lets it settle again; dragging any node re-heats it on demand.

## Performance notes

The graph is tuned to stay responsive at tens of thousands of nodes:

- The charge (repulsion) force uses `distanceMax(300)` so far-apart nodes are skipped — the largest per-tick cost on big graphs.
- `d3AlphaDecay(0.05)` + `d3VelocityDecay(0.5)` make the layout converge in fewer ticks, so the simulation (and its per-frame repaint) stops sooner instead of burning CPU.
- The simulation settles via a bounded `cooldownTicks` rather than running indefinitely.
- Node/neighbor/link lookups are pre-computed into `Map`s for constant-time hover and inspection.
- Filter inputs are debounced to avoid re-layout thrash while typing.

## Generating input from a workspace

Node data typically comes from a graphify export of a repo or an Agent OS workspace — a flat list of files with their repo/project/tags. Point the tool at that JSON via **Upload JSON**, or drop it in as `sample.json` for auto-load.
