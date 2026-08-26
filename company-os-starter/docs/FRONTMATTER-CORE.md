---
type: doc
tags: [doc/company-os-starter, kind/frontmatter-core]
---

# The Minimal Frontmatter Core

The smallest field set a Markdown document needs to participate in the
operating system. This core — not the document's section structure — is the
whole interop contract between teams, the company layer, `derive`,
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

## Recommended on every document

```yaml
description: "Fans out webhook delivery; retries are best-effort and unbounded."
```

`description:` is recommended on every document and **blocks nothing**. Its
consumer is the generated per-directory `index.md`, which renders each document's
title and description so a reader — human or agent — can decide what to open
without opening anything.

An index of titles alone is `ls` with extra steps. Two rules keep it from becoming
that. Both are tests you run against a specific document, not advice:

1. **Carry at least one fact not derivable from the document's filename,
   `title:`, `type:`, or directory path.** If a reader could reconstruct your
   sentence from the four things the index already displays, the sentence costs
   space and earns nothing.
2. **Do not let it read as true when pasted onto a sibling document in the same
   directory.** Copy the sentence onto the file next to it. If it still reads
   true, it describes the directory rather than the document, and it needs
   rewriting.

Applied to `platforms/communications/reality/customer-notification-service.md`:

| Description | Verdict |
| --- | --- |
| `"Reality doc for the customer notification service"` | Fails rule 1 — every word already appears in the filename, the `type:`, and the path |
| `"Current state of the customer notification service"` | Fails rule 2 — reads true pasted onto any reality doc in that directory |
| `"Fans out webhook delivery; retries are best-effort and unbounded."` | Passes — the retry semantics appear in neither the filename nor the type |

**Quote the value.** A useful description usually contains a colon, and
`description: One sentence: what this says` is a YAML parse error.

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

Every tag Obsidian shows, and every context-index link it follows, is
derived from these fields by `company-os derive`. Local Search reads the
same fields into typed graph edges — `components:` → `has_component`,
`dependsOn:` → `depends_on`, and so on. Hand-written tags are overwritten;
change the frontmatter instead. See
[user-guide/explanation/obsidian-and-local-search.md](user-guide/explanation/obsidian-and-local-search.md).

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

Two derived artifacts also exist now, both produced by `derive`,
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
| `description` (recommended) | Nothing — consumed by the generated `index.md` | No, ever |
| References → tag derivation | `validate` gate 4 (drift vs `derive`) | Yes, everywhere |
| Process accountability | `prd validate` / `prd complete` gates | Yes, at that gate |
| Section structure (templates/) | `discover validate`, `prd validate` | No — warnings, unless the team opts in |

A team that wants blocking section checks for its own docs opts in with
`teams/<team>/standards/doc-formats.yaml`:

```yaml
schemaVersion: "1.0"
enforce: true
```
