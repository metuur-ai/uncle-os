# Team-Drafted PRDs and Promotion — EARS Specifications

> Scope: **Ship 1** — team drafts and platform-level promotion only.

## Unit 1: Draft residency

**Why:** A team needs a place to draft that the tooling knows about. The location is fixed rather than configurable, because no team has asked to move it and a draft that landed somewhere `validate` could not see would be promotable while unchecked.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL resolve a team's draft directory to `teams/<team>/product/change-records/draft`. |
| R-1.2 | THE SYSTEM SHALL store a draft at `<draft-dir>/<draft-id>/prd.md`. |
| R-1.3 | IF a team's draft directory does not exist, THE SYSTEM SHALL treat that team as having no drafts and SHALL NOT emit a finding for its absence. |

## Unit 2: Creating a draft

**Why:** Drafting must be as cheap as scaffolding a discovery brief, must reuse the template chain teams can already override, and must leave `validate` green — because `teams/` is a graph-docs root and the node gate sees a new document immediately.

| ID | EARS statement |
| --- | --- |
| R-2.1 | WHEN `prd new --team <t> --draft "<title>"` runs, THE SYSTEM SHALL create a draft whose frontmatter carries `type: prd`, `status: draft`, a derived `id`, a `created` date of the current day, and `team` set to `<t>`. |
| R-2.2 | THE SYSTEM SHALL NOT write a `level` key into a draft, because `level` denotes rule tier elsewhere in the workspace. |
| R-2.3 | WHEN creating a draft, THE SYSTEM SHALL obtain its body from `ResolveTemplate` with template name `prd` and the team's id, honouring the existing team → platform → company → built-in probe order. |
| R-2.4 | WHEN creating a draft, THE SYSTEM SHALL write the process fields `title`, `platform`, `components`, `governanceSnapshot` and `decisionOwner` as placeholders, and SHALL write `promoteTo.platform` as a placeholder. |
| R-2.5 | WHERE `--draft --platform <p>` is supplied, THE SYSTEM SHALL set both `promoteTo.platform` and the frontmatter `platform` field to `<p>`. |
| R-2.6 | IF a draft with the derived id already exists in the team's draft directory, THE SYSTEM SHALL refuse with a conflict error and SHALL NOT overwrite the existing draft. |
| R-2.7 | WHEN `prd new --team <t> --draft --from-discovery <brief-id>` runs, THE SYSTEM SHALL require the brief to be `status: validated` and SHALL copy its Problem and Success sections into the draft. |
| R-2.8 | WHEN a draft is created successfully, THE SYSTEM SHALL rebuild derived artifacts before returning, so that `validate` exits 0 without a separate `graph build`. |
| R-2.9 | WHEN a draft is created successfully, THE SYSTEM SHALL name the draft file as the next thing to edit, and SHALL NOT print `prd promote` as the immediate next command. |
| R-2.10 | THE SYSTEM SHALL NOT write a `tags:` key into a draft by hand; tag facets for drafts SHALL be produced by `graph build`. |

## Unit 3: Draft-time validation posture

**Why:** Drafting is thinking. Requiring the full process contract while a change is still being formed defeats the purpose of the stage. The core frontmatter contract still applies, because every graph document obeys it and a draft with no identity cannot be promoted or audited.

| ID | EARS statement |
| --- | --- |
| R-3.1 | WHILE a PRD carries `status: draft`, THE SYSTEM SHALL require only the core frontmatter fields — `type`, an `id` identity, and `status` — and SHALL NOT require `title`, `platform`, `components`, `governanceSnapshot` or `decisionOwner`. |
| R-3.2 | WHILE a PRD carries `status: draft`, THE SYSTEM SHALL NOT check its body for the PRD section headings. |
| R-3.3 | THE SYSTEM SHALL NOT include team drafts in the workspace PRD contract gate that sweeps active change records. |
| R-3.4 | WHEN `validate` runs over a workspace whose only additions are well-formed drafts and whose derived artifacts are current, THE SYSTEM SHALL exit 0. |
| R-3.5 | IF a draft omits `status`, THE SYSTEM SHALL emit a core-field failure for that document, because `prd` is a lifecycle type. |
| R-3.6 | THE SYSTEM SHALL derive a `team/<t>` location facet for a draft. |

## Unit 4: Resolving the promotion target

**Why:** Where a draft is going must be written down by the author and checked for coherence before anything is written, not inferred at promotion time.

| ID | EARS statement |
| --- | --- |
| R-4.1 | THE SYSTEM SHALL read the promotion target from the draft's `promoteTo.platform`. |
| R-4.2 | IF `promoteTo.platform` is absent or still a placeholder, THE SYSTEM SHALL refuse promotion with an artifact error stating that the draft declares no target platform. |
| R-4.3 | IF `promoteTo.platform` names a platform that does not exist in the workspace, THE SYSTEM SHALL refuse promotion with a workspace error naming the unresolved platform. |
| R-4.4 | IF `promoteTo.platform` and the frontmatter `platform` field disagree, THE SYSTEM SHALL refuse promotion and SHALL report both values. |
| R-4.5 | THE SYSTEM SHALL reserve the key `promoteTo.scope` for later increments and SHALL neither write nor read it. |
| R-4.6 | THE SYSTEM SHALL evaluate every rule in this unit and in Unit 5's readiness rules before writing any file. |

## Unit 5: Promotion by copy-forward

**Why:** Promotion is the moment a change becomes platform-visible, so it is the moment the full contract must bind. Copying forward — rather than moving — keeps one document per graph root, preserves the team's record of what it proposed, and reuses the mechanism `--from-discovery` already established.

| ID | EARS statement |
| --- | --- |
| R-5.1 | WHEN `prd promote --team <t> <draft-id>` runs, THE SYSTEM SHALL create a change record at `platforms/<p>/change-records/active/<id>/prd.md`, seeded from the draft's body. |
| R-5.2 | WHEN writing the promoted record, THE SYSTEM SHALL set `status: proposed` and `created` to the current day. |
| R-5.3 | IF the draft does not carry all six process fields — `title`, `team`, `platform`, `components`, `governanceSnapshot`, `decisionOwner` — with values other than the placeholder, THE SYSTEM SHALL refuse promotion, SHALL report every missing field in one pass, and SHALL NOT write any file. |
| R-5.4 | IF the draft's body omits any of the sections `Problem statement`, `Success metrics` or `Proposed change`, THE SYSTEM SHALL refuse promotion and SHALL name each missing section in the same pass as R-5.3. |
| R-5.5 | WHEN refusing promotion for readiness, THE SYSTEM SHALL phrase the result as a not-yet-ready report naming what to supply, rather than as a malformed-artifact error. |
| R-5.6 | IF the draft's `status` is not `draft`, THE SYSTEM SHALL refuse promotion with a conflict error stating the status it found. |
| R-5.7 | IF a change record with the same id already exists at the target location, THE SYSTEM SHALL refuse promotion with a conflict error and SHALL NOT overwrite it, unless the resumption condition in R-5.11 holds. |
| R-5.8 | THE SYSTEM SHALL NOT delete, relocate or truncate the origin draft during promotion. |
| R-5.9 | WHEN promotion succeeds, THE SYSTEM SHALL perform its mutating effects in the order: write the target record, rewrite the draft, rebuild derived artifacts, compute the draft digest, patch the target with the origin pointer, append the team log. |
| R-5.10 | WHEN promotion succeeds, THE SYSTEM SHALL leave the workspace such that `validate` exits 0 with no intervening `graph build`. |
| R-5.11 | IF the target record exists, carries no `promotedFrom`, and the draft is still `status: draft`, THE SYSTEM SHALL treat the prior attempt as interrupted and SHALL resume from the write step rather than refusing. |
| R-5.12 | WHEN promotion succeeds, THE SYSTEM SHALL print `company-os prd validate --platform <p> <id>` as the next command in the guidance chain. |
| R-5.13 | WHEN `prd validate --platform <p> <id>` runs against a promoted record, THE SYSTEM SHALL produce the same result as for an equivalent hand-scaffolded record. |
| R-5.14 | WHEN `prd complete --platform <p> <id>` runs against a promoted record, THE SYSTEM SHALL produce the same result as for an equivalent hand-scaffolded record. |
| R-5.15 | WHEN `check ready` runs over a promoted record's components, THE SYSTEM SHALL produce the same result as for an equivalent hand-scaffolded record. |
| R-5.16 | WHEN `check done` runs over a promoted record's components, THE SYSTEM SHALL produce the same result as for an equivalent hand-scaffolded record. |
| R-5.17 | WHEN the workspace PRD contract gate sweeps active change records, THE SYSTEM SHALL treat a promoted record identically to a hand-scaffolded one. |

## Unit 6: Promotion audit trail

**Why:** A change record that appears on a platform with no trace of where it came from loses the reasoning that produced it. Both sides must point at each other, and the team tree must keep its history rather than being emptied by a successful promotion.

| ID | EARS statement |
| --- | --- |
| R-6.1 | WHEN promotion succeeds, THE SYSTEM SHALL write `promotedFrom` onto the promoted record carrying the origin team id, the draft id, and a digest of the draft. |
| R-6.2 | WHEN promotion succeeds, THE SYSTEM SHALL rewrite the draft's `status` to `promoted` and SHALL write `promotedTo` carrying the target platform and record id. |
| R-6.3 | THE SYSTEM SHALL compute the digest as a SHA-256 over a normalized form of the draft file — LF line endings, trailing whitespace stripped, derived `tags:` excluded — taken after derived artifacts have been rebuilt. |
| R-6.4 | THE SYSTEM SHALL compute the digest at promotion time and recompute it at validation time through one shared helper, so the two cannot diverge. |
| R-6.5 | WHEN promotion succeeds, THE SYSTEM SHALL append a dated entry to `teams/<t>/log.md` naming the draft id and the target record, creating the file if absent. |
| R-6.6 | THE SYSTEM SHALL NOT require `promotedFrom` or `promotedTo` on records that were never promoted. |

## Unit 7: Promotion-integrity gate

**Why:** A promoted draft is the origin of record for something other teams read. Editing it afterwards rewrites history the platform record claims to descend from. This is the failure mode the synced-slice hash map exists to catch, and it gets the same answer: the recorded digest is the oracle.

| ID | EARS statement |
| --- | --- |
| R-7.1 | THE SYSTEM SHALL run a promotion-integrity gate as part of `validate`, positioned after the skills gate and before the dynamically appended federation gate. |
| R-7.2 | WHILE a draft carries `status: promoted`, THE SYSTEM SHALL recompute its digest and compare it to the `promotedFrom.digest` recorded on the record named by `promotedTo`. |
| R-7.3 | IF a promoted draft's recomputed digest differs from the recorded digest, THE SYSTEM SHALL emit a failure naming the draft and the change record it no longer matches. |
| R-7.4 | IF the record named by a draft's `promotedTo` exists at neither the active nor the archive location, THE SYSTEM SHALL emit a failure naming the unresolved target. |
| R-7.5 | IF the record named by a draft's `promotedTo` exists but carries no `promotedFrom`, THE SYSTEM SHALL report it as an interrupted promotion naming the command to re-run, and SHALL NOT report it as drift. |
| R-7.6 | WHERE a promoted record has been archived by `prd complete`, THE SYSTEM SHALL resolve it under `archive/prds/<id>/` and SHALL treat a digest match there as passing. |
| R-7.7 | WHILE a draft carries `status: draft` or `status: abandoned`, THE SYSTEM SHALL emit no record from this gate. |
| R-7.8 | WHEN a workspace contains no promoted drafts, THE SYSTEM SHALL render this gate's header followed by no findings. |

## Unit 8: Retiring a draft

**Why:** Not every draft becomes real. Without a retirement path a workspace accumulates half-thoughts, and the only way to remove one is `rm` — hand-repair of tooling state, which this repo forbids everywhere else.

| ID | EARS statement |
| --- | --- |
| R-8.1 | WHEN `prd abandon --team <t> <draft-id>` runs, THE SYSTEM SHALL set the draft's `status` to `abandoned`. |
| R-8.2 | IF the draft's `status` is not `draft`, THE SYSTEM SHALL refuse with a conflict error stating the status it found. |
| R-8.3 | WHILE a draft carries `status: abandoned`, THE SYSTEM SHALL exclude it from every gate, from promotion, and from role views. |
| R-8.4 | THE SYSTEM SHALL provide no transition out of `abandoned`. |
| R-8.5 | WHEN a draft is abandoned successfully, THE SYSTEM SHALL rebuild derived artifacts before returning. |

## Unit 9: Draft discoverability

**Why:** A drafting stage nobody can find is a scratchpad with extra steps. `today` is this workspace's answer to "what is mine", and it must answer it for drafts as it already does for discovery briefs.

| ID | EARS statement |
| --- | --- |
| R-9.1 | WHEN `today --team <t>` runs, THE SYSTEM SHALL list the team's drafts carrying `status: draft`, naming each draft's id, title and target platform. |
| R-9.2 | THE SYSTEM SHALL exclude drafts carrying `status: promoted` or `status: abandoned` from that listing. |
| R-9.3 | WHERE a draft's title or target platform is still a placeholder, THE SYSTEM SHALL render it as unset rather than printing the placeholder. |
| R-9.4 | WHEN a team has no open drafts, THE SYSTEM SHALL omit the drafts section rather than rendering an empty one. |
| R-9.5 | THE SYSTEM SHALL list only the drafts of the team the view is scoped to. |

## Unit 10: Compatibility

**Why:** This repo's output is a contract — goldens fix gate ordinals, finding order and sentences. Adding a gate necessarily breaks that contract, so the break is declared, bounded and authorised rather than denied.

| ID | EARS statement |
| --- | --- |
| R-10.1 | THE SYSTEM SHALL preserve the slug, title, finding codes and intra-gate finding order of every gate that exists today. |
| R-10.2 | THE SYSTEM SHALL keep the dynamically appended federation gate last in the gate sequence. |
| R-10.3 | THE SYSTEM SHALL accept that adding an authored gate changes the reported gate count and the ordinal of every gate after the insertion point, and this change SHALL be recorded as an amendment following the precedent of `docs/ears/federation-enrichment.md` Amendment 1. |
| R-10.4 | THE SYSTEM SHALL land the golden-file re-baseline as an isolated commit containing no behavioural change. |
| R-10.5 | THE SYSTEM SHALL leave `prd new` without `--draft`, `prd complete`, `discover new`, `discover validate` and `--from-discovery` behaviour unchanged for all inputs that exist today. |
| R-10.6 | THE SYSTEM SHALL write every document produced here with frontmatter matching the `^---\n...\n---\n` parser contract exactly. |
| R-10.7 | THE SYSTEM SHALL confine all printing and process exit to `cmd/` and `internal/render/`, returning `[]model.GateResult` from every command added here. |
| R-10.8 | THE SYSTEM SHALL keep the built-in scaffolding templates in sync with the section names the promotion contract checks, so that a scaffolded draft satisfies R-5.4 without edits to its headings. |
| R-10.9 | THE SYSTEM SHALL introduce no runtime dependency beyond the Go standard library and the module's existing imports. |

## Unit 11: Skill and reference guidance

**Why:** Skills are how a team learns the workflow — a command nobody's skill mentions does not exist in practice. The canonical skills currently teach a platform-first PRD flow; leaving them untouched would make the draft stage invisible and make `prd complete` look reachable from a draft.

| ID | EARS statement |
| --- | --- |
| R-11.1 | THE SYSTEM SHALL document in `skills/creating-prd/SKILL.md` the draft-first path: `prd new --team <t> --draft "<title>"`, `prd validate --team <t> <draft-id>`, and `prd promote <draft-id> --team <t> --platform <p>`. |
| R-11.2 | THE SYSTEM SHALL state in that skill that a draft carries no platform, components or governance snapshot, and that those fields are required by the promotion contract, not by draft-time validation. |
| R-11.3 | THE SYSTEM SHALL state in that skill that promotion is a precondition of `prd complete`, and SHALL keep the platform-level `prd new --platform <p>` path documented as the direct alternative. |
| R-11.4 | THE SYSTEM SHALL document `prd abandon <draft-id> --team <t>` as the retirement path for a draft that will not be promoted. |
| R-11.5 | THE SYSTEM SHALL state in `skills/completing-a-change/SKILL.md` that a draft cannot be completed and must be promoted first. |
| R-11.6 | THE SYSTEM SHALL state in `skills/running-discovery/SKILL.md` that a validated brief can be carried forward by either `prd new --from-discovery` at platform level or `prd new --draft --from-discovery`. |
| R-11.7 | THE SYSTEM SHALL tag every skill step added here `(mandatory)`, `(default)` or `(guidance)`, matching the tiering convention already used in that file. |
| R-11.8 | THE SYSTEM SHALL document every command, flag and exit code added here in `company-os-starter/docs/user-guide/reference/company-os-cli.md`, and SHALL name the draft branch in `company-os-starter/docs/user-guide/how-to/take-a-change-from-discovery-to-done.md`. |
| R-11.9 | WHERE an example workspace ships a skill that shadows the canonical `skill://product/creating-prd`, THE SYSTEM SHALL update that copy so it does not contradict the canonical draft-first guidance. |
| R-11.10 | THE SYSTEM SHALL leave the canonical id, `appliesTo` and `precedence` frontmatter of every skill touched here unchanged, so the skills-layering gate reports no new shadowing or dangling `extends`. |
