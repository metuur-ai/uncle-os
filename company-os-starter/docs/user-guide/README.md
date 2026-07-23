# The Company OS user guide

Company OS is a Git-based operating system for running governance, product,
and engineering work as validated Markdown and YAML — not a SaaS app, not a
database, just files, a CLI that guides you through them, and a validator
that keeps everyone honest. Rules come in tiers (mandatory, default,
guidance), so a company or a single team can adopt it without losing the
flexibility to work their own way. And "Team OS" — you'll see the name in
older docs — isn't a separate product; it's what you get when one team runs
Company OS on its own, with no platform or company layer above it.

This folder is organized so you can jump straight to what you need instead
of reading front to back.

## Never touched Company OS before?

Start with **[tutorials/01-first-day-with-company-os.md](tutorials/01-first-day-with-company-os.md)**
— install the CLI, scaffold a workspace, and walk one real change from
discovery through to an updated reality doc.

## I'm a solo team, not a whole company

Read **[tutorials/02-running-a-standalone-team.md](tutorials/02-running-a-standalone-team.md)**.
This is the "Team OS" story: one team, no platform, no company layer, and a
CLI that degrades gracefully around the pieces that aren't there yet.

## I want my docs to be searchable from the terminal or an AI agent

**[tutorials/03-search-your-workspace.md](tutorials/03-search-your-workspace.md)**
walks through installing [Local Search](https://github.com/metuur-ai/local-search),
registering your Company OS workspace, and wiring up the Claude skill.

## I know roughly what I want to do

Jump to a how-to:

| I need to... | Page |
|---|---|
| Scaffold a new workspace, or add a platform/team/component to one | [how-to/grow-a-workspace.md](how-to/grow-a-workspace.md) |
| Ship a change: discovery → PRD → done | [how-to/take-a-change-from-discovery-to-done.md](how-to/take-a-change-from-discovery-to-done.md) |
| Deviate from a rule, or ask for an exception | [how-to/handle-a-deviation-or-exception.md](how-to/handle-a-deviation-or-exception.md) |
| Check my work against governance before I claim it's ready/done | [how-to/check-your-work-against-governance.md](how-to/check-your-work-against-governance.md) |
| Understand why `company-os validate` failed, or wire it into CI | [how-to/run-the-validation-gate.md](how-to/run-the-validation-gate.md) |
| Keep Local Search's index fresh and scoped correctly | [how-to/keep-search-fresh.md](how-to/keep-search-fresh.md) |

## I want the bigger picture

- **[explanation/how-it-fits-together.md](explanation/how-it-fits-together.md)** — how Company OS, Team OS, Local Search, and Claude skills cooperate.
- **[explanation/observer-roadmap.md](explanation/observer-roadmap.md)** — where the knowledge-graph vision (codename Observer) is headed. **Not shipped yet** — read this as roadmap, not a manual.

## I need the exact facts

- **[reference/company-os-cli.md](reference/company-os-cli.md)** — every `company-os` subcommand, its flags, and defaults.
- **[reference/configuration.md](reference/configuration.md)** — every config file and env var either tool actually reads.
- **[reference/troubleshooting.md](reference/troubleshooting.md)** — symptom → cause → fix, for both tools.
