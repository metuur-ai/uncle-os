---
title: Running a standalone team (Team OS)
---

# Running a standalone team ("Team OS")

Before Moonbeam Bakery was five locations, an online ordering platform, and a
`web` team reporting into it, it was one shop. No company layer, no
platform, just a couple of people baking and shipping code. This tutorial is
that shop's story — and it's also the origin story of the whole system.

## Where "Team OS" comes from

The founding proposal for this system (`docs/00-original-proposal.md`)
describes three cooperating layers: Company OS owns enterprise-wide
requirements, Platform OS owns reality/capabilities/components, and

> "The Team OS defines how one team operates" — charter, ownership,
> Definition of Ready/Done, workflows, AI-agent instructions.

**Team OS was never a separate tool.** It's what you get by running Company
OS with only the team layer present — no `company-os/`, no `platforms/`, no
`company-ontology/`. That's exactly what `examples/standalone-team/` in this
repo demonstrates, and it's what you're about to build for Moonbeam's
original shop.

> **Note:** if you've read older internal docs that talk about "Team OS" as
> if it ships separately, that's the same system described before the
> federated CLI existed. Today it's one CLI, one mode.

## 1. Look at the worked example first

The starter kit ships a complete, working standalone team at
`examples/standalone-team/`. Its own README calls out what matters:

> "One team (`teams/solo/`) uses the methodology with **no** `company-os/`,
> **no** `platforms/`, **no** `company-ontology/` present. Every command
> degrades gracefully."

Open it up:

```bash
$ cd examples/standalone-team
$ export PATH="$PWD/../../bin:$PATH"
$ company-os validate
```

`validate` only checks what exists — with no platforms and no active PRDs,
gates 1 and 3 simply have nothing to report, and it still ends in `PASS`.

```bash
$ company-os skills list
```

Company and platform skill layers show as empty; only the team and personal
layers render. Nothing errors — absence is a valid state, not a failure one.

Look at `teams/solo/team.yaml`. This is the piece worth copying verbatim for
Moonbeam's shop — the `agentSkills` block that governs how an AI agent
should weigh personal preferences against canonical rules:

```yaml
agentSkills: {canonicalPath: skills/, personalPath: scratchpad/personal-rules/,
    precedence: canonical-mandatory > personal > canonical-default > canonical-guidance,
    onConflict: prefer-canonical-and-inform-user}
```

`team.yaml` also carries optional identity blocks — `roster`, `channels`,
`pointers` — that render into the generated `CLAUDE.md` context node and are
queryable across teams later, if Moonbeam ever has more than one.

## 2. Scaffold Moonbeam's original shop the same way

You already know the command — the difference is you never run `add
platform` or touch `company-os/` or `company-ontology/` at all:

```bash
$ mkdir moonbeam-downtown && cd moonbeam-downtown
$ company-os init --company "Moonbeam Bakery" --team downtown --platform ordering
```

Wait — `init` always scaffolds one platform alongside the team. If you want
the *pure* standalone shape (team only, nothing else), delete the platform
and company roots it created and keep just `teams/downtown/` and
`company-ontology/`, mirroring `examples/standalone-team/`'s layout — or
simply copy that example's `teams/solo/` structure directly and rename it to
`downtown`. Either path gets you the same thing: a team directory that is
byte-for-byte identical in shape to what it will be inside a full workspace
later.

## 3. Work exactly like tutorial 01, minus the platform

Everything in [tutorials/01](01-first-day-with-company-os.md) that touches
`teams/<t>/` still applies here — `discover new`, `discover validate`, the
team's own `standards/definition-of-{ready,done}.md`, `scratchpad init` for
personal rules. What you *can't* do without a platform is `prd new`
(a PRD is a platform-visible change record by definition) — a standalone
team's discovery briefs stay team-private until there's a platform to attach
a change record to.

```bash
$ company-os discover new "Extend weekend hours" --team downtown
$ company-os discover validate 2026-extend-weekend-hours --team downtown
$ company-os validate
```

## 4. Joining the federation later, without restructuring

This is the entry ramp, not a dead end. When Moonbeam opens a second
location and needs a real `ordering` platform behind both shops, the
`teams/downtown/` directory doesn't move or change shape — you just run
`company-os add platform ordering` (or join a federated `workspace.yaml`,
see [docs/FEDERATION-RUNBOOK.md](../../../docs/FEDERATION-RUNBOOK.md)) and
point ownership at it. Compare this to `examples/banking/small-company/`
(one company, one repo) and `examples/banking/bank/` (full federation) in
this repo — same team-layer shape at every scale.

## Where to next

- [how-to/grow-a-workspace.md](../how-to/grow-a-workspace.md) — the exact
  commands for adding a platform once you outgrow standalone.
- [explanation/how-it-fits-together.md](../explanation/how-it-fits-together.md)
  — how the layers cooperate once more than one exists.
