---
type: doc
tags: [doc/company-os-starter, kind/frontmatter-core]
---

# The Minimal Frontmatter Core

The smallest field set a Markdown document needs to participate in the
operating system. This core — not the document's section structure — is the
whole interop contract between teams, the company layer, `graph build`,
`validate`, Obsidian, and any other tool that reads Markdown + YAML
frontmatter.

**Strict on process and structure, flexible on document formats.** A company
or a single team can adopt the OS jointly, independently, or alongside other
tools: any producer that emits the core participates in the graph. Everything
below the frontmatter is yours — ADR, RFC, free-form, house style. This is the
OKF rule applied: producers may extend metadata, and consumers must preserve
unknown fields rather than reject the document
(`docs/00-original-proposal.md:190`).

## Tier 1 — Identity (required on every doc)

```yaml
type: prd            # doc kind; known kinds derive a #kind/* tag
id: 2026-faster-webhooks   # stable, unique; outcome reviews may use `prd:` instead
```

## Tier 2 — Lifecycle (required per doc family)

```yaml
status: proposed     # discovery/prd/adr/outcome: draft|validated|proposed|…|completed
updated: 2026-07-18  # reality docs only — the `prd complete` done-gate reads it
authority: canonical # reality docs only
```

## Tier 3 — References (any that apply; these ARE the graph)

```yaml
team: customer-engagement
platform: communications
components: [customer-notification-service]
boundedContext: context://communications   # optional
fromDiscovery: 2026-faster-webhooks        # traceability edge
prd: 2026-faster-webhooks                  # outcome reviews → their PRD
```

Every tag and wikilink Obsidian sees is derived from these fields by
`company-os graph build` (Ontology Guide §2.2). Hand-written tags are
overwritten; change the frontmatter instead.

## Tier 4 — Process accountability (required by specific gates)

```yaml
created: 2026-07-18        # PRD; compared against reality `updated:`
governanceSnapshot: 2026-07-18
decisionOwner: Ada (PM)    # `prd validate` refuses TODO
due: 2026-10-16            # outcome reviews
```

## Pointers — external references (any doc)

`pointers:` is an optional list of references to systems outside the OS. It is
valid on `team.yaml`, component descriptors, PRDs, reality docs, and any other
document. Each entry carries a `label`, a `system`, and at least one of `url`
or `id`:

```yaml
pointers:
  - {label: Service repository, system: github, url: 'https://...'}
  - {label: On-call rotation, system: pagerduty, id: PD-1234}
```

Well-formedness is **guidance-tier**: malformed entries warn but do not block —
*except* where a gate consumes a specific pointer, in which case that gate
blocks. The system stores references only; it **never** fetches or mirrors the
external content. Pointers are collected into each platform's derived
`feature-index.yaml` under `externalPointers`.

## Team identity blocks (`teams/<t>/team.yaml`)

Three optional blocks describe a team. All are optional — a `team.yaml` without
them still validates:

```yaml
roster:                              # list of people
  - {name: Ada, role: product-owner}
channels:                            # list of comms channels
  - {name: team-notifs, id: C012345, system: slack}   # system optional
pointers:                            # as above
  - {label: Team wiki, system: confluence, url: 'https://...'}
```

These feed the identity summary in the generated team `CLAUDE.md` context node.

## Onboarding guides (`type: onboarding-guide`)

A doc `type` for role-scoped onboarding material. It derives a `kind/onboarding`
tag and a `role/<role>` tag, requires `id` and `role`, and does **not** require
`status`:

```yaml
type: onboarding-guide
id: developer-onboarding
role: developer
```

Guides live at `company-os/onboarding/<role>.md` (company scope) and
`teams/<t>/onboarding/<role>.md` (team scope). `company-os today --role <r>`
prints a pointer to the matching guide, preferring team scope over company
scope.

## Reserved doc types (inert)

`account-context`, `customer-call`, and `data-catalog` are **reserved** type
names. They are inert today — no required-field gate runs against them until
their consumer ships — but the names are claimed so producers can start emitting
them without collision.

## Generated — never hand-written

```yaml
tags: [kind/prd, platform/communications, team/customer-engagement, status/proposed]
```

Two derived artifacts also exist now, both produced by `graph build`,
drift-checked by `validate`, and never hand-edited:

- `platforms/<p>/generated/feature-index.yaml` — the derived component→artifact
  map (includes `externalPointers` collected from `pointers:`).
- the `company-os:generated` marker block inside each federation root's
  `CLAUDE.md` — the generated context node.

## Everything else is format

Titles, section headings, prose structure, extra metadata — team-local.
Unknown fields are preserved, never rejected.

## What validates what

| Field tier | Checked by | Blocking? |
| --- | --- | --- |
| Identity + lifecycle | `validate` gate 4, `discover validate`, `prd validate` | Yes, everywhere |
| References → tag derivation | `validate` gate 4 (drift vs `graph build`) | Yes, everywhere |
| Process accountability | `prd validate` / `prd complete` gates | Yes, at that gate |
| Section structure (templates/) | `discover validate`, `prd validate` | No — warnings, unless the team opts in |

A team that wants blocking section checks for its own docs opts in with
`teams/<team>/standards/doc-formats.yaml`:

```yaml
schemaVersion: "1.0"
enforce: true
```
