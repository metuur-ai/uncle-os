---
title: Check your work against governance
---

# Check your work against governance

Four commands cover "what applies to me, and am I meeting it": `governance
resolve`, `governance explain`, `check ready` / `check done`, and `today`.

## See what governance a team is under

```bash
$ company-os governance resolve --team web
resolved governance for team 'web' (1 component(s))
wrote teams/web/generated/effective-governance.yaml
  online-ordering-app: platforms [ordering], 3 company + 3 platform requirement(s)
```

This merges company baseline requirements, the platform's requirements
(filtered to components the team actually owns or touches), and the team's
approved deviations, into `teams/<t>/generated/effective-governance.yaml`.
It's derived — never hand-edit it; re-run `resolve` instead. `--team` isn't
marked `required` by the CLI's argument parser, but in practice you always
need to pass it — there's no team to resolve without one.

## Ask why a specific rule applies

```bash
$ company-os governance explain online-ordering-app
component 'online-ordering-app' (team web):
  - order-confirmation-sla v1.0 (mandatory)
      applies because the component 'belongs-to' platform 'ordering'
  - prd-structure v1.0 (default) [deviation applied]
      applies because the component 'belongs-to' platform 'ordering'
```

Each line names the tier and, when a deviation is in play, flags it right
there — you never have to cross-reference `deviations.yaml` by hand to know
whether a rule is still fully in force.

## Check readiness or doneness for specific components

```bash
$ company-os check ready --team web --components online-ordering-app
== Team baseline (definition-of-ready.md) ==
...
== Applicable governance (online-ordering-app) ==
- [ ] ordering: order-confirmation-sla v1.0 (mandatory) — evidence:
...

$ company-os check done --team web --components online-ordering-app
```

Both `--team` and `--components` are required. `check` composes the team's
own `standards/definition-of-{ready,done}.md` with the resolved governance
checklist for exactly the components you name — comma-separate multiple
component ids. Nothing here is a static checklist file that goes stale;
it's regenerated from current ownership and current rules every time you
run it.

## See your role's daily view

```bash
$ company-os today --role developer
```

`--role` accepts `developer` (default), `team-lead`, `product-owner`,
`architect`, `vp-engineering`, or `director-of-product`. Each role sees a
different slice — a developer sees their components and open governance
items; a product owner sees active PRDs and outcome reviews coming due.

## When things go sideways

| Symptom | Cause | Fix |
|---|---|---|
| `governance resolve` reports 0 components for a team | team's `ownership/components.yaml` is empty, or doesn't match any descriptor | check `company-os ids list --team <team>` and the descriptor under `platforms/<p>/components/` |
| `governance explain` output doesn't mention a rule you expected | the component doesn't `belong-to` that platform, or the rule is scoped to a different `componentType` | re-check the component descriptor's platform relationship |
| `check ready`/`check done` errors about missing flags | `--team` or `--components` omitted — both required | supply both; comma-separate multiple component ids |
| `effective-governance.yaml` looks stale | it's generated, and something changed ownership/requirements since the last run | re-run `company-os governance resolve --team <team>` |
