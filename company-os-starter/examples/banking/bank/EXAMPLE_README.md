# Example: `bank/` — a ~250-person federated bank

**What it demonstrates:** patterns P1-P7 and P9 (index in `../README.md`)
across three zones:

## `repos/` — who owns which repo
Simulated source repositories, one per organizational unit:
- `company-os/` — baseline tiers: `data-residency` + `security-service-baseline`
  (mandatory, outcome-phrased), `estimation-story-points` (default),
  `adr-format` (guidance).
- `company-ontology/` — the linchpin: 21-entry ID registry; bounded contexts
  where **Payments forbids "Transaction"** (means Payment Order/Settlement
  there) while **Fraud owns "Transaction"**; translation lives in
  `context-maps/fraud-to-payments.md` (published-language + ACL).
- `platform-{payments,cards,fraud,identity,customer-service,lending}/` — one
  Platform OS each. `identity` is the shared platform (P4): its only mandatory
  requirement targets `relationships: [consumes]`. `payments` catalogs two
  components in two code repos under one team (P3).
- `team-cards-issuing/` — the accountable side of P2 (its descriptor claim is
  reconciled by validate [1/7]).
- `code-transaction-screening/` — a code repo: invisible to the CLI, bound to
  governance only via grep-able `@spec req://…#R<n>` markers in tests.

## `workspaces/` — where people actually work
Each team clones **its own repo as the workspace root** and syncs everything
else read-only:
- `team-fraud-detection/` — `workspace.yaml` pinning company-os, ontology, and
  four platforms; ownership mixing `accountable` + three `consuming` claims
  (P2/P4); a deviation on a default rule and an **expiring exception** on
  mandatory `payments/settlement-finality` (P6); Priya's git-ignored personal
  rule (P1). `SYNC-NOTE.md` explains what `workspace sync` would materialize
  and why no lock file is committed here.
- `team-payments-rails/` — second workspace; Priya appears on both rosters (P7).

## `initiatives/` — the known gap, made concrete
`instant-refunds/initiative.md` models the **proposed** cross-platform
initiative artifact (gap G2): today the CLI scopes PRDs to one platform, so
multi-domain work is three unlinked PRDs; this file shows the lightweight
linking doc the research suggests.

**Caveats:** commit pins are illustrative placeholders; lock hashes cannot be
faked, so none are committed; synced slices are not materialized inside the
workspaces (see `team-fraud-detection/SYNC-NOTE.md`).
