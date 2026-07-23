---
title: How it fits together
---

# How it fits together

Four things show up across this guide: Company OS, Team OS, Local Search,
and Claude skills. It's easy to read them as four separate tools. They're
not — they're layers over one shared substrate.

## The shared substrate is files

Everything Company OS produces — discovery briefs, PRDs, reality docs,
governance requirements, component descriptors, ADRs — is Markdown with
YAML frontmatter, committed to Git. Nothing lives in a database only the
CLI can read. That single decision is what makes the rest of this diagram
possible: any tool that can read a file can participate.

```text
                     ┌─────────────────────────────┐
                     │   Company OS workspace        │
                     │   (Markdown + YAML, in Git)   │
                     │                               │
                     │  company-os/  platforms/*/    │
                     │  teams/*/     company-ontology/│
                     └───────────────┬───────────────┘
                                     │
                 authored & validated by
                                     │
                         company-os CLI
              (governance resolve, discover, prd,
               check, validate, graph build, today)
                                     │
                 read & indexed by (separately)
                                     │
                      ┌──────────────▼──────────────┐
                      │        Local Search           │
                      │  SQLite FTS5 index, one repo  │
                      │  or many, offline, no server  │
                      └───────┬───────────────┬───────┘
                              │               │
                     local-search CLI   local-search ui
                     (search/find/json)  (web, port 8787)
                              │
                    consumed by an agent via
                              │
                    ~/.claude/skills/local-search
                      (installed by
                       local-search install-skill)
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

## Local Search: the read side

Company OS never queries anything — it only writes files and validates
them. Local Search is the tool that makes those files *findable*: `repo add`
registers a workspace, its index tracks git state so results stay fresh
without a daemon, and `search`/`find` return ranked results with enough
provenance (repo, path, freshness) to trust. It's a separate project on
purpose — Company OS shouldn't need to know how you search your docs, and
Local Search shouldn't need to know Company OS's schema to be useful (though
it indexes Company OS's derived tags for free, since `graph build` writes
them straight into frontmatter).

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

## What's not in this picture yet

Observer — a shared knowledge graph across all of this, with typed edges
instead of just search results — is vision only. See
[explanation/observer-roadmap.md](observer-roadmap.md).
