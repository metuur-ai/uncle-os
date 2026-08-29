# Company OS — Conformance

**What this document is for.** Company OS is built on the Open Knowledge Format
(OKF), and almost everything it enforces is *not* OKF. Without that line written
down, every rule in the system reads as mandated by the standard and therefore
immovable. This document draws the line: what OKF requires, what Company OS
chose, and why the difference matters when deciding what may change.

**Status.** Partial delivery of R-1.1 of `docs/ears/okf-v02-conformance.md`.
The Conformance clause, MUST-NOT-reject list and "considered and deferred"
sections below are complete. Terminology and Versioning are stubs pointing at
their source of truth; see §7 for the parent clauses still outstanding. This
document records existing behavior and changes none of it.

---

## 1. Goals

1. State the OKF-required floor precisely enough that a reader can tell whether a
   given rule is negotiable.
2. Cite an implementation site for every claim about what this tooling does.
3. Name the additions Company OS makes on top of OKF, so they can be argued about
   individually rather than defended as a bloc.
4. Record known inaccuracies elsewhere in the documentation without fixing them
   here.

## 2. Non-Goals

1. **Changing any conformance behavior.** Nothing here alters a gate, a field
   check, or an exit code.
2. **Fixing the inaccuracies recorded in §6.** They belong to the parent
   conformance change.
3. **Defining the `profile:` enum.** R-1.7 of the parent change owns it. Note
   `internal/scaffold/scaffold.go:196-197` already writes `profile: standard`
   into every new platform, against an enum defined nowhere.
4. **Claiming full OKF v0.2 conformance.** This document describes what is true
   today.

## 3. Terminology

Defined normatively in `docs/FRONTMATTER-CORE.md` and the root `CLAUDE.md`. The
parent change's R-1.10 owes this section full definitions of *reality*,
*deviation*, *exception*, *component*, *canonical*, *authority*, *tier*,
*federation root*, *generated artifact*, *absence tolerance* and `knowledge/`.
Two are load-bearing here and are stated now:

- **Absence tolerance** — a gate that iterates rows produces zero findings when
  the file holding those rows is missing. `loadOr(path, pyMap{})`
  (`internal/governance/gates.go:47`) makes a missing file and an empty list
  behave identically. This is why a workspace can omit whole roots and still
  validate.
- **Generated artifact** — content produced by `company-os derive` and
  authoritative in the tree, but never hand-edited: frontmatter `tags:`,
  `teams/<t>/generated/effective-governance.yaml`,
  `platforms/<p>/generated/feature-index.yaml`, per-directory `index.md`, and the
  marked block inside each `CLAUDE.md`.

## 4. Conformance clause

Keywords per RFC 2119.

### 4.1 What OKF requires — the floor

A document **MUST** carry a non-empty `type:` field.
*Enforced:* `internal/product/contract.go:79-81` (`CodeCoreTypeMissing`).

Tooling **MUST** preserve frontmatter keys it does not recognise.
*Honoured:* `docs/FRONTMATTER-CORE.md:16-20` states the rule and attributes it to
OKF; covered by `TestRewriteFrontmatterTagsPreservesUnknownKeys`
(`internal/graph/selftest_test.go:68`).

Documents **MUST** be plain files: Markdown with YAML frontmatter, no database,
no service.

Tooling **MUST NOT** reject a document for a broken cross-link.

`title`, `description` and `resource` are **RECOMMENDED**, not required.
*Consistent with implementation:* `description` never blocks
(`docs/FRONTMATTER-CORE.md` "What validates what" — "No, ever").

`index.md` is a reserved filename that **MAY** be generated.
*Company OS generates it and gates its drift* — an addition, see §5.

**That is the entire floor.** Roughly six lines of enforcement.

### 4.2 What Company OS adds

Every item below is a local decision with no OKF counterpart.

An artifact **MUST** carry `id:` — or `prd:` for outcome reviews — in addition to
`type:`. *(`internal/product/contract.go:83-85`. OKF's identity is path-minus-`.md`.)*

Lifecycle types (`discovery-brief`, `prd`, `adr`, `outcome-review`) **MUST**
carry `status:`; `component-reality` **MUST** carry `updated:`;
`onboarding-guide` **MUST** carry `role:`. *(`contract.go:87-99`.)*

`tags:` **MUST** be derived, never authored. `company-os derive` overwrites them
from the source fields. *(`internal/graph/tags.go:80-128`. OKF's `tags:` is
authored.)*

Canonical IDs **MUST** be registered once in `company-ontology/ids/registry.yaml`.
*(`internal/ids/ids.go:75`.)*

A component descriptor is the **single source of truth** for its accountable
team, and a team's ownership registry **MUST** agree with it.
*(Gate 1, `internal/governance/gates.go:100-120`.)*

Deviations **MUST** carry an unexpired `reviewDate`; exceptions **MUST** carry an
unexpired `expires`. *(Gate 2, `gates.go:163-232`.)*

Generated artifacts **MUST** match a fresh derivation; drift is a failure.
*(Gates 4, 5, 6.)*

Federated slices **MUST** hash-match `workspace.lock.yaml`, and are materialised
read-only. *(Gate 9, `internal/federation/lock.go`.)*

A change **MUST NOT** complete while a governance checklist item is unchecked or
a component's reality doc predates the PRD. *(`prd complete`.)*

**Only gate 4 is the OKF surface** — labelled "the OKF/Obsidian interop contract"
at `internal/validate/validate.go:140`. Gates 1–3 and 5–9 are house enforcement.

## 5. MUST-NOT-reject list

Tooling **MUST NOT** reject a workspace or document for any of the following.
Each cites the site that already honours it.

| # | Must not reject for | Honoured at |
|---|---|---|
| 1 | An unknown `type:` value | `internal/graph/tags.go:23-28` — `kindTag` returns no facet for an unrecognised type; nothing raises |
| 2 | Unknown frontmatter keys | `internal/graph/selftest_test.go:68` |
| 3 | A broken cross-link | no gate resolves wikilinks; pointer well-formedness is WARN only (`internal/validate/validate.go:195-198`) |
| 4 | A missing optional document | gates iterate rows and emit nothing when the file is absent (`internal/governance/gates.go:47`) |
| 5 | A missing federation root | `IsRoot()` returns true if **any** canonical root is present (`internal/workspace/workspace.go:149-159`) — a two-root workspace validates clean |
| 6 | A missing `index.md` where a directory does not qualify | `indexThreshold = 2` (`internal/graph/index.go:21`) |
| 7 | A hand-owned `CLAUDE.md` with no generated markers | `internal/graph/gates.go:156-161` treats it as authored and passes |
| 8 | An empty section, unless the team opted in | `internal/product/contract.go:111-156` — guidance tier by default |

This list binds the `--json` surface identically to human output.

**Consequence worth stating plainly:** items 4 and 5 together mean the four-root
federation is a *convention of the scaffolder*, not a requirement of the
validator.

## 6. Versioning

`companyOsVersion: "2026.2"` appears in nine fixture files. **The version it names
is defined nowhere**, and this document does not define it — R-1.5 and R-1.6 of
the parent change own that, along with the forward-only bump rule. Recorded here
so the gap is visible rather than implied.

## 7. Considered and deferred

Recorded, not fixed (§2.2).

| Condition | Evidence | Owner |
|---|---|---|
| `docs/00-original-proposal.md:3` claims **OKF v0.1** while the EARS targets v0.2 | verified | parent R-1.8/R-1.9, task 1.4 |
| `companyOsVersion: "2026.2"` in 9 fixtures against a version defined nowhere | verified | parent R-1.5, task 1.3 |
| `resource:` appears in **no** example fixture despite being specified | `grep "^resource:" examples/` returns nothing | parent R-3.1–R-3.5, tasks 3.1–3.5 |
| No single interop-contract claimant; `docs/FRONTMATTER-CORE.md` and this document both describe the contract | — | parent R-1.14 |
| `profile: standard` is written into every new platform against an undefined enum | `internal/scaffold/scaffold.go:196-197` | parent R-1.7, task 1.3 |

**Parent clauses still outstanding after this document:** full Terminology
(R-1.10), the versioning definition (R-1.5, R-1.6), the `profile:` enum (R-1.7),
per-field adopt/supersede/decline against OKF (R-1.8), recording HLD N1–N8
(R-1.11), and the single-claimant amendment (R-1.14).

---

## Why this document exists — the short version

The OKF floor is `type:` plus key preservation, plain files, and link tolerance.
Everything else in Company OS — the eight gates, the required `id:`, derived tags,
the ID registry, tiering, deviations, exceptions, federation, and
generation-with-drift-as-failure — is this project's own choice.

Choices can be revisited. That is the point of writing them down.
