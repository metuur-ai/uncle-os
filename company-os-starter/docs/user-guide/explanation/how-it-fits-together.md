---
title: How it fits together
---

# How it fits together

Five things show up across this guide: Company OS, Team OS, Local Search,
Obsidian, and Claude skills. It's easy to read them as five separate tools.
They're not — they're layers over one shared substrate.

The short version: **Company OS writes, Local Search and Obsidian read.**
Nothing on the read side is a plugin, an export, or a sync job — both of
them open the same files the CLI just wrote.

## The shared substrate is files

Everything Company OS produces — discovery briefs, PRDs, reality docs,
governance requirements, component descriptors, ADRs — is Markdown with
YAML frontmatter, committed to Git. Nothing lives in a database only the
CLI can read. That single decision is what makes the rest of this diagram
possible: any tool that can read a file can participate.

```text
                    company-os CLI          ← the only writer
         (governance resolve, discover, prd, check,
                validate, derive, today, sync)
                             │
                     writes & validates
                             │
                             ▼
        ┌────────────────────────────────────────────┐
        │        Company OS workspace                │
        │        Markdown + YAML, in Git             │
        │                                            │
        │   company-os/   platforms/*/               │
        │   teams/*/      company-ontology/          │
        │   knowledge/    (synced catalog)           │
        │                                            │
        │   every file: frontmatter core             │
        │               + derived tags:              │
        │               + generated CLAUDE.md blocks │
        └──────────┬──────────────────────┬──────────┘
                   │                      │
           indexed by                 opened by
                   │                      │
                   ▼                      ▼
    ┌─────────────────────────┐   ┌──────────────────────┐
    │      Local Search       │   │       Obsidian       │
    │  SQLite FTS5 + vector   │   │  the vault IS the    │
    │  re-rank, graph edges   │   │  workspace directory │
    │  from frontmatter,      │   │                      │
    │  offline, no server     │   │  graph view, backlinks│
    └────┬───────────────┬────┘   │  tag panes, editing  │
         │               │        └──────────────────────┘
  local-search CLI  local-search ui        no plugin,
  (search/find/     (web, :8787)           no export,
   graph/json)                             no sync step
         │
 consumed by an agent via
         │
  ~/.claude/skills/local-search
  (local-search install-skill)
```

## Company OS: the authoring and validation side

`company-os` is a guide, not a gatekeeper in the punitive sense — its
subcommands scaffold artifacts, inject the governance that applies to them,
and then `validate` checks the result against the shared contract
(frontmatter core, tag derivation, ownership reconciliation, expiring
deviations/exceptions). Rules are tiered — mandatory, default, guidance — so
teams keep real flexibility in *how* they work while the company still gets
guaranteed outcomes where it actually matters. See
[reference/company-os-cli.md](../reference/company-os-cli.md).

## Team OS: the same system, one layer

"Team OS" isn't a second CLI or a second file format — it's Company OS with
only the `teams/<t>/` layer present, as demonstrated by
`examples/standalone-team/`. A team can start there and grow into a full
federation later without restructuring anything, because the team-layer
shape never changes. See
[tutorials/02-running-a-standalone-team.md](../tutorials/02-running-a-standalone-team.md).

## The read side: Local Search and Obsidian

Company OS never queries anything — it only writes files and validates them.
Two tools read what it wrote, and they read it *differently*, which is the
point: one is for asking questions, the other is for wandering around.

### `derive` is the seam

Both readers depend on one command. `company-os derive` walks the workspace
and rewrites, from the frontmatter core, two things it owns:

- **`tags:`** — faceted, nested (`#kind/prd`, `#platform/ordering`,
  `#tier/mandatory`). Local Search indexes these as first-class facets;
  Obsidian shows them in the tag pane and search.
- **Generated `CLAUDE.md` blocks** — a doc index and a federation index per
  layer, written between markers as *relative Markdown links*. Obsidian
  follows them for backlinks and the graph view; Local Search turns them
  into link edges. Reference frontmatter (`components:`, `dependsOn:`,
  `upstream:`, `fromDiscovery:`) gives Local Search typed edges on top.

Hand-writing either is pointless — the next `derive` overwrites them, and
`validate` gates `[4/7]`–`[6/7]` fail on the drift in the meantime. Change
the frontmatter; run `derive`; both readers update. See
[FRONTMATTER-CORE.md](../../FRONTMATTER-CORE.md) and
[obsidian-and-local-search.md](obsidian-and-local-search.md).

### Local Search: the question-answering side

`repo add` registers a workspace; the index tracks git state so results stay
fresh without a daemon; `search`/`find` return ranked hits with enough
provenance (repo, path, freshness) to trust. It combines SQLite FTS5 with
vector re-ranking, so "what did we decide about refund windows" works as
well as an exact term. Because it also reads reference frontmatter into
typed graph edges — and body links, `@spec` markers, and any `[[wikilinks]]`
you write by hand — it can answer relationship questions, not just "where is
X" but "what links to X".

This is what an agent uses. It is also what makes multi-repo work: Local
Search can hold several registered repos at once, so a query spans a
workspace *and* the component repos beside it.

### Obsidian: the wandering side

There is no Obsidian plugin to install and no export step. **Point Obsidian
at the workspace directory and it is a vault** — because the workspace was
already Markdown with YAML frontmatter and relative links, which is exactly
Obsidian's native format. Graph view, backlinks, the tag pane, and local
editing all work on day one. Setup:
[../how-to/open-your-workspace-in-obsidian.md](../how-to/open-your-workspace-in-obsidian.md).

This matters for the humans who won't use a CLI. A product owner can browse
the governance that applies to their platform by clicking links; an engineer
can see everything that points at a reality doc before changing it. Edits
made in Obsidian are ordinary edits to tracked files — `validate` checks
them like any other, and `derive` reclaims the fields it owns.

### Why they stay two separate binaries

`company-os` and `local-search` ship independently on purpose. Company OS
shouldn't need to know how you search your docs, and Local Search shouldn't
need to know Company OS's schema to be useful — it indexes the derived tags
for free, and it is just as happy indexing repos that have never heard of
Company OS. Obsidian isn't a dependency at all; it's a consequence of the
file format. Merging them would trade that away for nothing: the coupling
that matters is the file format, and it's already shared.

The one place the two are *planned* to meet is registration — having
`company-os` notice that the workspace isn't registered with an installed
Local Search and print the command to fix it. That is specified but not
shipped; see [What's not in this picture yet](#whats-not-in-this-picture-yet).
Today you run `local-search repo add <workspace>` yourself, once.

## Claude skills: the agent side

Two different skill mechanisms show up in this guide, and they're worth
telling apart:

- **Company OS's canonical skills** (`skills/*/SKILL.md`, layered with a
  team's personal rules in `scratchpad/personal-rules/`) teach an agent *how
  to author* an artifact correctly — e.g. "create a PRD" resolves to the
  platform's canonical `creating-prd` skill.
- **Local Search's Claude skill** (`~/.claude/skills/local-search`, from
  `local-search install-skill`) teaches an agent *how to find* what's
  already written, resolving the current project's search scope via
  `local-search init --json` before every query.

Together: an agent asked to work on Moonbeam's ordering platform can find
the relevant discovery briefs and reality docs (Local Search), then follow
the canonical skill for producing a compliant PRD (Company OS) — without a
human relaying context between two disconnected tools.

`company-os skills list` shows the merged four-layer view; see
[how-to/use-the-agent-skills.md](../how-to/use-the-agent-skills.md) for the
precedence rule, the `extends:` mechanism, and the `--json` shape an agent
consumes.

## What's not in this picture yet

Two pieces of the read-side integration are **specified but not shipped**.
The specs are written and reviewed; the code is not. Treat them as roadmap,
not as behaviour you can rely on today:

| Spec | What it would add | Status |
|---|---|---|
| **Retrieval adapter handshake** | `company-os` detects an installed Local Search, checks the workspace is registered and the index isn't stale, and prints the exact command to fix it — as an advisory warning, never a failing gate | Drafted; review found the "already registered" case unimplementable against today's `local-search` exit codes. Blocked on engine-side work |
| **Derived drift repair** | `company-os validate --fix` repairs derived drift in place — the tags and generated blocks `derive` owns — instead of only reporting it | Drafted and revised; not implemented |

Specs live in the `uncle-os` repo (outside this starter package) under
`docs/hld/`, `docs/lld/`, and `docs/ears/`, one file per slug:
`retrieval-adapter-handshake` and `derived-drift-repair`.

Both are deliberately advisory-shaped. Neither would make the governance
lifecycle depend on a search index being present or fresh — that separation
is the [reason the two binaries stay
separate](#why-they-stay-two-separate-binaries), and a working search index
is a convenience, never a gate.

Observer — a shared knowledge graph across all of this, with typed edges
instead of just search results — is further out still, and vision only. See
[explanation/observer-roadmap.md](observer-roadmap.md).

Also absent, and on purpose rather than pending: any live integration with an
external service. There is no MCP server, no MCP client, and no GitHub API call
in the CLI — `workspace sync` is plain git, and it is the only command that
touches the network at all. See
[explanation/github-mcp-and-automation.md](github-mcp-and-automation.md) for what
that rules out, and how to use the GitHub MCP server alongside a workspace
without breaking the guarantees.
