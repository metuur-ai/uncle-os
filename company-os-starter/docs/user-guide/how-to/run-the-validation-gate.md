---
title: Run the validation gate
---

# Run the validation gate

`company-os validate` is the one command that checks the whole workspace at
once. Run it locally before opening a PR, and wire it into CI so nothing
merges that would fail it.

```bash
$ company-os validate
```

## The gates

The gate count is dynamic — 7 gates in a monorepo workspace, 8 the moment a
`workspace.yaml` federation manifest exists (see
[reference/configuration.md](../reference/configuration.md)). Gates 5, 6,
and 7 are absence-tolerant: they pass when the artifact they'd check simply
isn't there yet.

| Gate | Checks |
|---|---|
| `[1/N]` ownership reconciliation | a team's `ownership/components.yaml` claim matches the component descriptor's `ownership.accountableTeam` |
| `[2/N]` deviation and exception expiry | no deviation's `reviewDate`, and no exception's `expires`, has passed |
| `[3/N]` active PRD contracts | every active PRD has `title`, `team`, `components`, `governanceSnapshot` in frontmatter |
| `[4/N]` frontmatter core and tag derivation | every doc's core frontmatter fields are present, and committed `tags:` match what `graph build` would derive |
| `[5/N]` CLAUDE.md context node drift | a generated `CLAUDE.md` context block matches a fresh render (absent CLAUDE.md → pass) |
| `[6/N]` feature-index drift | each platform's `generated/feature-index.yaml` matches its derivation, and every reference in it resolves (absent → pass) |
| `[7/N]` custom skills layering | no team/personal skill shadows a canonical skill's id/name, and every `extends:` resolves (absent/no conflicts → pass) |
| `[8/N]` federated slice integrity *(federated workspaces only)* | materialized slices still match `workspace.lock.yaml` — a content-hash mismatch means someone hand-edited derived content, and a slice-set mismatch means a `localDirectory:`/`paths:` moved without a re-sync |

## Reading a real run

```text
$ company-os validate
[1/7] ownership reconciliation
  [ok] online-ordering-app: registry and descriptor agree (ordering)
[2/7] deviation and exception expiry
[3/7] active PRD contracts
[4/7] frontmatter core and tag derivation (interop contract)
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
[6/7] feature-index drift (derived component->artifact map)
[7/7] custom skills layering (shadowing + extends resolution)
PASS
```

Any failure prints as `[FAIL] ...` inline under its gate, exits non-zero, and
ends the run with `FAIL — N problem(s)` instead of `PASS` — that exit code
is what CI acts on.

## Wire it into CI

```yaml
# .github/workflows/os-validate.yml (any OS repo)
- run: pip install pyyaml
- run: bin/company-os --root . validate
- run: bin/company-os governance resolve --team <team> && git diff --exit-code teams/*/generated/
```

The second step is not redundant with `validate` — it proves
`effective-governance.yaml` is truly derived. If someone hand-edited it, the
freshly regenerated file differs from what's committed and `git diff
--exit-code` fails the build.

## When things go sideways

| Symptom | Cause | Fix |
|---|---|---|
| Gate `[1/N]` fails: "claims accountable but descriptor says..." | team registry and component descriptor disagree on the accountable team | edit the component descriptor's `ownership.accountableTeam` — it's authoritative, not the team's registry |
| Gate `[2/N]` fails: expired deviation or exception | `reviewDate`/`expires` passed | re-declare the deviation or re-request the exception with a fresh date, or remove it if it's no longer needed |
| Gate `[4/N]`, `[5/N]`, or `[6/N]` fails: "drifted... run: company-os graph build" | committed derived content (tags, CLAUDE.md node, feature-index) is stale | run `company-os graph build`, review the diff, commit it |
| Gate `[7/N]` fails: skill shadowing or dangling `extends` | a team/personal skill reuses a canonical skill's id/name, or `extends: platform-skill://...` points nowhere | rename the conflicting skill, or fix/remove the broken `extends` |
| Gate `[8/N]` fails: federated slice integrity | a materialized governance slice was hand-edited instead of coming from `workspace sync` | discard the hand-edit and re-run `company-os workspace sync` |
