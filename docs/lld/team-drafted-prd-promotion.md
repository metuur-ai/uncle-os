# Team-Drafted PRDs and Promotion — Low-Level Design

> Scope: **Ship 1** — team drafts and platform-level promotion only.

## Architecture

### Filesystem shape

```
teams/<t>/product/change-records/draft/<draft-id>/prd.md    # NEW
teams/<t>/log.md                                            # appended on promote
platforms/<p>/change-records/active/<id>/prd.md             # unchanged shape
```

The draft directory is fixed. `teams/<t>/product/discovery/` is untouched and remains the earlier stage.

### Frontmatter

A draft on creation:

```yaml
---
type: prd
id: <draft-id>
status: draft
created: <today>
promoteTo:
  platform: TODO
title: TODO
team: <t>
platform: TODO
components: []
governanceSnapshot: TODO
decisionOwner: TODO
---
```

Two deliberate choices:

- **No `level:` key.** `level` already means *rule tier* in this workspace — `schemas/SCHEMAS.md` declares it as `mandatory|default|guidance` and `internal/product/checklist.go` reads it that way. A second meaning in the same vocabulary is exactly what the bounded-context discipline exists to prevent. The promotion target is `promoteTo.platform`, a platform id and nothing else. When component and company scopes arrive they add `promoteTo.scope`; that key is reserved now and unused.
- **`platform: TODO` on the draft is the process field**, distinct from `promoteTo.platform` which is the routing field. They are required to agree at promotion time (see step 3). This keeps the draft's process contract identical to a change record's, so promotion is a copy rather than a translation.

After promotion the draft gains:

```yaml
status: promoted
promotedTo:
  platform: <p>
  record: <id>
```

and the promoted record gains:

```yaml
promotedFrom:
  team: <t>
  draft: <draft-id>
  digest: sha256:<hex>
```

## Components

### `internal/product` — draft creation

`PRDNew` gains a `--draft` path. It reuses `ResolveTemplate("prd", teamID)` so the existing team → platform → company → built-in probe order applies unchanged, and reuses the `--from-discovery` section copy-forward verbatim. It writes into the team draft dir instead of the platform active dir, and stamps `status: draft`.

It calls the same `Rebuild` injection `PRDComplete` already performs at the end of its effect sequence. This is a change to `prd new`'s behaviour and is therefore scoped strictly to the `--draft` branch: `prd new` without `--draft` is untouched, satisfying the no-regression constraint while keeping Success Criterion 1 honest for drafts.

### `internal/product` — promotion

New `PRDPromote(ws, team, draftID)` returning `[]model.GateResult`. Nothing in it prints or exits.

### `internal/product` — abandonment

New `PRDAbandon(ws, team, draftID)`. Sets `status: abandoned`, runs `Rebuild`, returns. Refuses if status is not `draft`.

### `internal/validate` — promotion-integrity gate

A new authored gate appended after the skills gate and before the dynamically appended federation gate. It walks each team's draft dir, and for every draft with `status: promoted` recomputes the digest and compares it to the record named by `promotedTo`.

### `internal/roles` — `today`

The role view gains a drafts section listing each open draft's id, title and `promoteTo.platform`, rendered next to the existing discovery section. Drafts with `status: promoted` or `abandoned` are excluded.

### `skills/` and `company-os-starter/docs/user-guide/` — guidance surface

Prose, not code, and the only part of this change a team reads before using it. `skills/creating-prd/SKILL.md` gains the draft branch (`prd new --draft` → iterate → `prd validate --team` → `prd promote`) alongside the existing platform-first branch, with each new step tiered `(mandatory)`/`(default)`/`(guidance)` like the steps around it; `skills/completing-a-change/SKILL.md` gains the precondition that a draft must be promoted first; `skills/running-discovery/SKILL.md` gains the second carry-forward target. Skill frontmatter (`id`, `appliesTo`, `precedence`) is untouched, so `internal/skills` shadowing and `extends` resolution see no change. The example workspaces ship shadowing copies of `skill://product/creating-prd` (`examples/workspace`, `examples/federated`, `examples/banking/bank/repos/platform-payments`); these are updated for consistency, not because any gate reads their body.

## Key Decisions

### Copy-forward, not move

Moving the draft would empty the team's record of what it proposed and would move a document between two graph roots mid-lifecycle. Copying keeps one document per root and mirrors `--from-discovery`, which already copies forward rather than relocating. The cost is two documents that can drift; the digest gate is the answer to that cost.

### Digest is taken **after** the rebuild, over normalized bytes

This is the correction to the original ordering. `DeriveTags` emits a `status/` facet, and promotion flips `status` from `draft` to `promoted` — so `RewriteFrontmatterTags` is *guaranteed* to rewrite the draft after the status change. Digesting before the rebuild would fail the gate on every successful promotion, without exception.

The digest is therefore computed as the last step, and over a normalized form: LF line endings, trailing whitespace stripped, and the derived `tags:` block excluded. Excluding derived tags means a later `graph build` that re-derives the same block byte-differently cannot produce a false drift. `PRDPromote` and the gate share one exported normalize-and-hash helper, so the compute and the recompute cannot diverge.

### Effect ordering and crash recovery

Steps 1–6 are read-only. Steps 7 onward mutate:

```
1.  load draft, assert status == draft
2.  resolve promoteTo.platform, assert platform exists
3.  assert promoteTo.platform == frontmatter platform
4.  assert six process fields present and != TODO      -> collect all failures
5.  assert three required sections present             -> collect all failures
6.  assert target record id is free                    -> unless resumable (below)
7.  write platforms/<p>/change-records/active/<id>/prd.md, status: proposed
8.  rewrite draft: status -> promoted, add promotedTo
9.  Rebuild (graph build) over the workspace
10. compute normalized digest of the draft as it now stands on disk
11. patch the promoted record with promotedFrom{team, draft, digest}
12. append dated entry to teams/<t>/log.md, creating it if absent
```

Steps 4 and 5 accumulate rather than short-circuit, so one run reports every omission.

A crash between 7 and 11 leaves a target record with no `promotedFrom`. Re-running `prd promote` detects exactly that state — target exists, target has no `promotedFrom`, draft still `status: draft` — and resumes from step 7 rather than returning `ExitConflict`. Any other collision at step 6 is a genuine conflict and refuses. The gate treats a target that exists but carries no `promotedFrom` as an interrupted promotion and reports it as such, not as drift.

### Retiring a draft

`status: abandoned` is a terminal state excluded from every gate and from `today`. It exists so that a draft which never becomes real is retired by a command rather than by `rm`, and so that a workspace does not accumulate permanently-red half-thoughts. There is no path back from `abandoned`; create a new draft.

## Constraints

- Frontmatter is written to the `^---\n...\n---\n` parser contract exactly.
- Nothing below `cmd/` prints or calls `os.Exit`; every new command returns `[]model.GateResult`.
- `tags:` is never hand-written; all facets come from `DeriveTags`.
- No new module dependencies.
- The Go CLI's fidelity to the Python oracle is preserved for all existing behaviour. The one intentional divergence — the added gate and its renumbering — is authorised by the amendment named in the HLD and lands in an isolated commit.

## Out of Scope

- Component-scoped and company-scoped records, and everything they imply: `--component` addressing, a second `BuildFeatureIndex` sweep root, component reality paths in the done-check, component archive paths.
- `prd demote`.
- Draft expiry or review dates. Unlike deviations and exceptions, a draft does not go stale on a clock.
- Surfacing another team's drafts. `today` shows the caller's team only.
