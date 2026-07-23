---
title: Grow a workspace
---

# Grow a workspace

Moonbeam Bakery started with one platform (`ordering`) and one team
(`web`). Six months later there's a loyalty program, a kitchen-ops team, and
three more components. None of that requires restructuring anything you
already built — `add` grows the same four peer roots `init` created.

## Add a platform

```bash
$ company-os add platform loyalty
added platform 'loyalty'
next: company-os add component --platform loyalty <component-id>
```

This scaffolds `platforms/loyalty/` with the standard shape
(`components/`, `governance/requirements.yaml`, `reality/`,
`change-records/active/`, `archive/prds/`, `skills/`) and registers
`platform://loyalty` in the ontology's `ids/registry.yaml`.

## Add a team

```bash
$ company-os add team kitchen-ops
added team 'kitchen-ops'
next: company-os discover new --team kitchen-ops "<discovery title>"
```

Scaffolds `teams/kitchen-ops/` — ownership, governance, product/discovery,
standards, generated/ — and registers `team://kitchen-ops`.

## Add a component to a platform

Components live under a platform and belong to a team. `--platform` is
required:

```bash
$ company-os add component crumb-club-app --platform loyalty
added component 'crumb-club-app' to platform 'loyalty'
next: company-os reality new --platform loyalty crumb-club-app
```

This writes `platforms/loyalty/components/crumb-club-app.yaml` — the
descriptor that is the **single source of truth** for both platform links
and the accountable team. Set `ownership.accountableTeam` there; a team's
own `ownership/components.yaml` must agree, or gate `[1/7]` in `company-os
validate` fails.

> **Tip:** running `add component` without `--platform` fails fast with
> `add component requires --platform <platform-id>` — there's no such thing
> as a component that doesn't belong to a platform.

## Scaffold the component's reality doc

Every component needs a reality doc before any change can complete against
it (see [how-to/take-a-change-from-discovery-to-done.md](take-a-change-from-discovery-to-done.md)):

```bash
$ company-os reality new crumb-club-app --platform loyalty
```

Writes `platforms/loyalty/reality/components/crumb-club-app.md` from the
template. It refuses to overwrite a file that already exists.

## Re-derive everything after growing

`add` already rebuilds generated content for you, but if you've hand-edited
anything (you shouldn't — see
[reference/configuration.md](../reference/configuration.md)) or want to be
sure, run:

```bash
$ company-os graph build
$ company-os validate
```

## When things go sideways

| Symptom | Cause | Fix |
|---|---|---|
| `add component requires --platform <platform-id>` | forgot `--platform` | pass an existing platform id — check with `company-os ids list --prefix platform://` |
| `validate` gate `[1/7]` fails after adding a component | `ownership.accountableTeam` in the descriptor doesn't match the team's `ownership/components.yaml` | edit one of the two to agree — the descriptor is authoritative |
| `reality new` refuses with "already exists" | the reality doc was already scaffolded | edit the existing file instead of re-running `reality new` |
| Newly added platform/team/component missing from `company-os today` | generated views are stale | run `company-os graph build` |
