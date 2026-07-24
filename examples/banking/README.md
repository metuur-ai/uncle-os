# Banking Example — Multi-Repo & Multi-User Pattern Catalog

Companion to `.devlocal/research/2026-07-22-multi-repo-multi-user-use-cases.md`.
Two org profiles built from the same substrate (same layout, same commands):

- `small-company/` — ~10-person startup, single monorepo workspace, no federation.
- `bank/` — ~250-person bank, federated: simulated source repos + team workspace roots.

## Pattern index

| # | Pattern | Where to look |
|---|---------|---------------|
| P1 | Single user contributing to multiple repositories | `bank/workspaces/team-fraud-detection/teams/fraud-detection/scratchpad/personal-rules/priya-review-checklist.md` + same person in `bank/workspaces/team-payments-rails/teams/payments-rails/team.yaml` roster |
| P2 | Multiple teams collaborating on the same platform | `bank/repos/platform-cards/components/card-issuing.yaml` (accountable: team://cards-issuing) + `bank/workspaces/team-fraud-detection/teams/fraud-detection/ownership/components.yaml` (consuming) |
| P3 | Platform composed of many code repos owned by different teams | `bank/repos/platform-payments/components/` (two components, two `repo://` pointers, different accountable teams) + `bank/repos/code-transaction-screening/` |
| P4 | Shared repositories/platforms used across multiple platforms | `bank/repos/platform-identity/` — every team consumes `component://auth-service` |
| P5 | Cross-platform initiative across business domains | `bank/initiatives/instant-refunds/initiative.md` (references one PRD per platform — this artifact type is a documented gap G2, modeled here as a proposal) |
| P6 | Company-wide standards + team-level customization | `bank/repos/company-os/standards/company-baseline.yaml` (tiers) + `bank/workspaces/team-fraud-detection/teams/fraud-detection/governance/{deviations,exceptions}.yaml` |
| P7 | Users belonging to multiple teams simultaneously | Priya Shah in both team rosters (`team.yaml` of fraud-detection and payments-rails) |
| P8 | Small-company profile (~10 people, monorepo) | `small-company/` — no `workspace.yaml`, one platform, one team |
| P9 | Growing monorepo → federated | Compare `small-company/` layout with `bank/workspaces/team-fraud-detection/workspace.yaml` — same canonical roots, platform dirs become pinned read-only slices |

## What is real vs illustrative

- File shapes follow `examples/workspace/` and `examples/federated/` exactly (schemas, tags, tiers, EARS-style requirements).
- `workspace.lock.yaml` hashes here are **illustrative placeholders** — a real lock is written by `company-os workspace sync` against real repos (validate gate [8/8] compares slice bytes to lock hashes).
- Synced slices are **not materialized** inside `bank/workspaces/*` — in real use, `workspace sync` would materialize `company-os/`, `company-ontology/`, and `platforms/*` there read-only (0444/0555) from `bank/repos/*`.
- The ontology enforces vocabulary per bounded context; Payments forbids "Transaction" (Fraud owns it) — see `bank/repos/company-ontology/contexts/`.

## Content artifacts (beyond the YAML)

Real lifecycle documents, following `examples/workspace/` shapes exactly:

- Discovery briefs: `small-company/teams/core/product/discovery/2026-instant-statements/brief.md`, `bank/workspaces/team-fraud-detection/teams/fraud-detection/product/discovery/2026-alert-triage-queues/brief.md`
- Active PRDs (P5 sister set): `bank/repos/platform-{payments,cards,customer-service}/change-records/active/2026-instant-refunds-*/prd.md` + `small-company/platforms/product/change-records/active/2026-instant-statements/prd.md`
- Completed change: `bank/repos/platform-fraud/archive/prds/2026-alert-triage-queues/{prd.md,outcome.md}` — checklist shows a deviation note AND an exception note in action; `log.md` entries in both platforms
- Skills: `bank/repos/platform-payments/skills/creating-prd.SKILL.md`
- Ontology concept: `bank/repos/company-ontology/concepts/capability--instant-transfer.md`
- Team standards & context: DoR/DoD for `core` and `fraud-detection`, fraud team `CLAUDE.md` with the generated-block convention, company onboarding guide
