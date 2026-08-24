# Team-Drafted PRDs and Promotion — High-Level Design

> Scope: **Ship 1** — team drafts and platform-level promotion only.
> Component-scoped and company-scoped change records are deferred; see [Non-Goals](#non-goals).

## Overview

Today a PRD only exists once it is a platform-visible change record under `platforms/<p>/change-records/active/`. There is nowhere to think. A team that wants to work up a proposal either creates a half-formed change record that the workspace gate immediately fails, or keeps it out of the OS entirely in a scratchpad the tooling cannot see.

This change gives a team a place to draft. A draft lives under `teams/<t>/product/change-records/draft/<id>/prd.md`, carries `status: draft`, and is deliberately cheap: it must satisfy the core frontmatter contract every graph document obeys, and nothing else. When the team decides the proposal is real, `company-os prd promote` copies it forward into a platform change record with `status: proposed` — and at that moment the full PRD contract binds. Both documents keep a pointer to the other, and the promoted draft is frozen against later edits.

The shape mirrors discovery, which is already team-private and already promotes forward via `prd new --from-discovery`. Drafting is the same idea applied one stage later, for teams whose proposal did not begin as a discovery brief.

## Stakeholders & Impact

**Teams** gain a first-class drafting stage. Today they choose between a red workspace and an invisible scratchpad; after this they get a directory the CLI scaffolds, `today` surfaces, and `validate` tolerates.

**Product owners** get a queryable answer to "what is this team working up?" — drafts appear in `today --role product-owner` alongside discovery briefs, so a proposal in progress is visible before it lands on a platform.

**Platform maintainers** are unaffected until promotion. Nothing new appears under `platforms/` until a team runs `prd promote`, and what appears then is indistinguishable from a hand-scaffolded change record.

**Everyone running `validate`** sees one additional gate and a re-numbered gate sequence. This is a deliberate, amendment-tracked break; see [Compatibility](#compatibility).

## Goals

1. A team can scaffold a PRD draft in one command and iterate on it without `validate` failing on its incompleteness.
2. Promotion is an explicit, auditable act — never implicit, never a side effect of editing frontmatter.
3. A promoted change record is byte-for-byte a normal change record: `prd validate`, `prd complete`, `check ready`, `check done` and the workspace PRD gate cannot tell it was promoted.
4. Both sides of a promotion point at each other, and the origin draft cannot be silently rewritten afterwards.
5. A draft that never becomes real can be retired without hand-editing or hand-deleting anything.
6. A team can see its own drafts without knowing the directory layout.
7. The canonical skills teach the draft stage, so a team following its skill never has to discover the new commands from a changelog.

## Non-Goals

- **Component-scoped change records.** Nesting a record under the component it changes requires component addressing on `prd validate`/`prd complete`, a second sweep root in `BuildFeatureIndex`, a component reality path in the done-check, and a component archive path. That is its own increment.
- **Company-scoped change records.** A change spanning several platforms has no exit today: the done-check resolves every component's reality doc under a single platform, so such a record could never complete. Deferred until it has a reality model.
- **Cross-platform reality reconciliation.** Unchanged.
- **A `prd demote` verb.** Retiring a draft is supported; un-promoting a change record is not.
- **Changing `prd new`, `prd complete`, `discover new`, `discover validate` or `--from-discovery` behaviour for any input that exists today.**
- **A configurable draft directory.** The location is fixed. If a team needs it elsewhere, that is a later request with a real motivating case.

## Success Criteria

1. `company-os prd new --team <t> --draft "<title>"` creates a draft, and `company-os validate` exits 0 on that workspace after `company-os graph build` has run — the same post-scaffold expectation `discover new` already sets.
2. `company-os prd promote --team <t> <draft-id>` refuses a draft that is missing any of the six process fields (`title`, `team`, `platform`, `components`, `governanceSnapshot`, `decisionOwner`) or any of the three required sections, naming every omission in one pass and writing no file.
3. Promotion of a complete draft produces `platforms/<p>/change-records/active/<id>/prd.md` with `status: proposed`, and `company-os prd validate --platform <p> <id>` exits 0 on it immediately.
4. `company-os validate` exits 0 immediately after a successful promotion, with no intervening `graph build`.
5. Editing a promoted draft causes the promotion-integrity gate to fail, naming the draft and the record it no longer matches.
6. `company-os prd abandon --team <t> <draft-id>` sets `status: abandoned`, and the draft is thereafter ignored by every gate.
7. `company-os today --role product-owner --team <t>` lists the team's open drafts.
8. `skills/creating-prd/SKILL.md` documents the draft → promote path end to end, `skills/completing-a-change/SKILL.md` states that a draft must be promoted before it can complete, and `company-os-starter/docs/user-guide/reference/company-os-cli.md` lists every new command and flag.
9. `make check` passes, and the only golden-file changes are gate-count and gate-ordinal lines authorised by the amendment.

## Compatibility

`validate` derives its banner and every `[n/total]` header from the length of its step list. Adding an authored gate therefore renumbers the sequence and rewrites the golden files — this cannot be avoided while also adding a gate, and the earlier claim that it could was wrong. This change follows the precedent set by [Amendment 1](../ears/federation-enrichment.md#amendment-1--r-74-partial-retirement-2026-07-26): the re-baseline is explicit, authorised, and landed as its own isolated commit so that no real regression hides inside it.

What does **not** change: every existing gate keeps its slug, title, finding codes and intra-gate finding order, and the dynamically appended federation gate remains last in the sequence.
