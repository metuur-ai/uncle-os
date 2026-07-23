---
title: "Observer: the knowledge-graph roadmap"
---

> **Not yet available.** Everything on this page is a design vision, not a
> shipped tool. No commands below can be run today. If you came here
> looking for something to install, you want
> [tutorials/03-search-your-workspace.md](../tutorials/03-search-your-workspace.md)
> instead — that's real, working software.

# Observer: the knowledge-graph roadmap

"Observer" is the working name for a future capability: a shared knowledge
graph over everything Company OS, Local Search, and related tools already
know about — not a new binary to install, but an evolution of Local Search
itself. This page summarizes the current thinking, drawn from internal
research (`2026-07-22-knowledge-graph-recommendation.md`), so you know
what's coming without mistaking it for what's here.

## The problems it's meant to solve

- **"I know we wrote this down somewhere."** Specs, PRDs, ADRs, and reality
  docs live across many repos with no shared parent — finding the right one
  today means grepping several trees or asking a teammate.
- **"The agent doesn't know what the company knows."** An agent working in
  one repo has no way to discover that another repo's ADR constrains its
  task.
- **"I can see the doc but not the shape."** Relationships between
  documents — what implements what, which PRD produced which component —
  exist only as frontmatter fields and prose a human has to piece together.
- **"I can't trust what I found."** Without provenance (which repo, which
  commit, how fresh, what authority), nothing can be trust-ranked.

## The key architectural decision already made

There is **no separate `kg` tool**. The graph is a derived, disposable index
over files that remain the source of truth — the same principle Company OS
already follows (IDs are canonical; tags and links are derived, never the
reverse) and the same principle Local Search's storage layer already states
in its own schema comments. Local Search is the vehicle this evolves on top
of: one Go binary, SQLite underneath, no separate server.

## What's actually done vs. proposed

| Phase | Status |
|---|---|
| 1. Registry + scan + keyword search + provenance | **Done** — this is Local Search today, as covered in [tutorials/03](../tutorials/03-search-your-workspace.md). |
| 2. Graph edges + traversal + a graph UI pane | **Started.** `@spec` marker and wikilink extraction into searchable tags landed 2026-07-22; the `nodes`/`edges`/`aliases` schema, canonical-ID entity merging across repos, and a `walk` traversal command do not exist yet. |
| 3. Semantic tier + task-shaped agent context (`context --budget`) | Not started. |
| 4. Hygiene + integrations (EARS clause nodes, conflict/entity-resolution lens, Obsidian projection, optional MCP shim, a Company OS `validate` hook, Observer/session ingestion) | Not started. |

## What it would look like for an agent, eventually

These are proposed, not real, commands — shown only so the shape of the
vision is concrete:

```text
local-search node component://online-ordering-app        # (future)
local-search edges component://online-ordering-app --out  # (future)
local-search walk req://ordering/order-confirmation-sla --depth 2  # (future)
local-search context "add retry to the ordering webhook" --budget 4000  # (future)
```

Every result would carry `{source_repo, path, line_span, commit, updated,
authority}` so an agent could quote, open, and trust-rank instead of
receiving an unsourced fact.

## What it would look like for a human, eventually

`local-search ui` exists today as a search web UI. The vision grows it into
a graph view: a force-directed graph, click-to-inspect nodes with rendered
docs and provenance, a freshness lens (stale reality docs glow), an orphan
lens, and a conflict lens for candidate-duplicate content — navigational
only, never an editor. Edits would still happen in the files; the graph
would catch up within seconds.

## Why this matters for Company OS specifically

The research explicitly calls out Company OS's ontology (`component://`,
`capability://`, `req://`, `context://` canonical IDs, `graph build`'s
derived tags) as the richest input this vision has seen — Company OS
workspaces would plausibly become a first-class extraction profile before
generic repos do. None of that changes anything about how you use
`company-os` or Local Search today.
