# Schema Reference (contract summary)

Machine validation is performed by `company-os validate` and the per-artifact
`validate` subcommands. This file is the human-readable contract.

## Rule tiers (used everywhere)
- `mandatory` — must be satisfied; only escapable via an approved, expiring exception.
- `default`  — comply-or-explain; teams may deviate via governance/deviations.yaml.
- `guidance` — recommended; no tracking.

## Component descriptor — platforms/<p>/components/<id>.yaml
Authoritative for: component↔platform relationships AND accountable team.
Required: metadata.id, metadata.type, ownership.accountableTeam, platformRelationships[].
The team ownership registry is validated against this file (single-source rule).

## Platform requirements — platforms/<p>/governance/requirements.yaml
Required per requirement: id, level (tier), version.
Optional: appliesTo.componentTypes, appliesTo.relationships, effectiveFrom,
supersedes, migrationDeadline, verification.evidence, verification.checklist.
Mandatory requirements must be written as OUTCOMES, not implementations.

## Team ownership — teams/<t>/ownership/components.yaml
Required per component: id, relationship (accountable | maintainer), repository.

## Deviations — teams/<t>/governance/deviations.yaml
Required: rule, tier(default), status, rationale, reviewDate.
Expired reviewDate ⇒ validate FAILS.

## Exceptions — teams/<t>/governance/exceptions.yaml
Required: rule, component, reason, compensatingControls, approvedBy, expires.
Missing/expired expires ⇒ validate FAILS.

## Discovery brief / PRD / reality / outcome
See templates/ — frontmatter fields marked mandatory are enforced by
`discover validate`, `prd validate`, and `prd complete`.

## Federation manifest — workspace.yaml (workspace root, optional)
Its presence switches the workspace into federated mode (validate grows an
8th gate). Top level: `repos:`, a non-empty list.
Required per repo: name (a plain [A-Za-z0-9._-] label — it keys the git cache
dir), url, pin, and a destination.
`pin:` — EXACTLY ONE of commit:/tag:. Branches and bare refs ⇒ FAILS.
Destination — EITHER top-level `localDirectory:` + optional `paths:`, OR a
`slices:` list of {localDirectory, paths}. Setting both forms ⇒ FAILS.
`paths:` defaults to the governance allowlist; it is an include-only
allowlist (no globs, no excludes).
`localDirectory:` must be relative and land under company-os/, platforms/,
teams/, company-ontology/, or knowledge/. Under knowledge/ it must name an
area (depth >= 2), not the catalog root. Targets must be disjoint across the
whole manifest — equal or nested ⇒ FAILS.

## Generated files
teams/<t>/generated/effective-governance.yaml is DERIVED by
`company-os governance resolve`. Never edit it by hand; CI should run
`governance resolve` and fail if the committed file differs.
workspace.lock.yaml is DERIVED by `company-os workspace sync` (resolved SHAs,
slice list, per-file hashes). Committed, machine-owned, never hand-edited.
