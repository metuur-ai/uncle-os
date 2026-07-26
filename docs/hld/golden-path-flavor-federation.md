# Golden Path, Flavor & Federation — High-Level Design

> Merges three research documents into one spec, framed simplicity-first:
> `.devlocal/research/2026-07-22-simplicity-user-journey-review.md` (umbrella),
> `.devlocal/research/2026-07-22-custom-templates-and-skills.md`,
> `.devlocal/research/2026-07-22-multi-repo-workspace-federation.md`.
>
> Related spec: `docs/{hld,lld,ears}/federation-enrichment.md` builds on the federation base defined here (Units 6–7). Requirement IDs in this spec carry the `GPF-` prefix to avoid collision with that spec's numbering.

## Overview

Company OS today is strict on artifacts and flexible on process, but the user journey is terminal-only, jargon-heavy, and assumes a monorepo. This change wraps the existing lifecycle in a **simplified golden path** — a guided `init` wizard, canonical-ID lookup, next-command guidance, and role-aware terminology — and delivers two capabilities *inside* that path: (1) **team/org flavor** via a template-override resolution chain and additive custom-skill layers that never weaken the validate contract, and (2) **manifest-optional workspace federation** so the same workspace can span multiple repos through SHA-pinned, governance-only slices (`workspace.yaml` + `workspace sync`), with zero behavior change when no manifest is present.

## Stakeholders & Impact

| Who | Pain today | After this ships |
|---|---|---|
| Non-technical users (PO / PM / product-owner role) | Terminal-only setup, ~13-command surface, must already know `component://` IDs, OS jargon (deviation, exception, EARS, ontology) | Guided setup wizard, one-document golden path, ID lookup from the CLI, plain-language display labels per role |
| Team members / developers | Templates are hardcoded `*_TEMPLATE` strings in `bin/company-os`; no sanctioned way to add team flavor | Override templates at team/platform/company layers; add team skills and personal rules additively |
| Platform owners / governance maintainers | Federation requires a monorepo; sharing governance across repos means copying or full-repo inclusion | `workspace.yaml` federates governance-relevant slices only, pinned to explicit SHAs, reproducible via lock file |
| CI (`company-os validate`) | — (must not regress) | Same gate; gains `--frozen` deterministic mode and new checks (heading contract on custom templates, skill-shadowing, read-only slices) |
| Agents / skills consumers | Skill discovery is single-layer | Deterministic 4-layer skill discovery (company + platform + team shared, personal local), canonical-wins conflict rule |

## Goals

When this ships, the following are observable:

1. A new user can go from environment prerequisites (download the binary, PATH) to a validated first PRD following **one document** and the CLI's printed next-command guidance, without prior knowledge of canonical IDs or OS terminology — including the reality-update step, which is scaffolded (`reality new`), never hand-created from scratch.
2. `company-os init` interactively scaffolds a workspace (company baseline, first team, first platform) and prints the first golden-path command.
3. Canonical IDs are discoverable from the CLI (lookup/list from `company-ontology/ids/registry.yaml`), and unknown-ID errors suggest close matches.
4. Role-aware views (`--role`) render plain-language labels alongside canonical terms — display-only; artifacts keep canonical vocabulary.
5. Scaffolding commands resolve templates through a first-found-wins chain (team → platform → company → built-in) while `validate` still enforces the required-headings contract on every artifact regardless of which template produced it.
6. Skills resolve additively across four layers; a canonical skill can be *extended* (`extends:` frontmatter) but never *shadowed* — shadowing is a `validate` error; canonical mandatory steps win over personal rules.
7. Dropping a `workspace.yaml` at the workspace root activates federated mode: `company-os workspace sync` materializes read-only, governance-relevant slices of remote repos at explicit SHA pins, recorded in `workspace.lock.yaml`; `workspace status` reports drift; `--frozen` runs offline from the lock.
8. A workspace with **no** `workspace.yaml` behaves byte-for-byte as today (monorepo mode unchanged).
9. `company-os add platform|team|component` grows an existing workspace from the same templates `init` uses — no hand-copied YAML — and workspace-scoped commands fail fast outside a workspace root instead of succeeding silently.

## Non-Goals

Invariants that must NOT change:

- **Strict on artifacts, flexible on process** — validators check outputs, never method.
- The validate contract does not weaken: required headings, frontmatter, tier rules, ownership reconciliation, expiries all still gate.
- IDs stay canonical; tags and wikilinks stay derived (`graph build`); generated files stay machine-owned.
- CLI-owned artifact parts (frontmatter, IDs, status fields) stay CLI-owned regardless of template.

Explicitly out of scope (decided):

- **Git submodules / full-repo inclusion** — rejected; federation syncs governance-relevant slices only (Option B).
- **Per-machine path mapping** (the former federation Option C) — rejected alongside submodules; `workspace.yaml` + lock file is the only federation mechanism.
- **Templatable YAML row seeds** for `deviations.yaml` / `exceptions.yaml` — stay built-in (deferred until a real need).
- **Localized heading aliases** — required section headings remain English even in localized-flavor templates (deferred).
- Web/GUI surfaces; renaming canonical terminology inside artifacts; any new runtime dependency beyond `pyyaml` + git.

> **Scoping note, 2026-07-26 (R-8.6, R-8.7).** The last bullet above has been
> read as forbidding two things it does not forbid. Recorded here rather than
> rewritten, so the original decision stays legible:
>
> - **"Web/GUI surfaces" does not cover a terminal UI shipped inside the
>   binary.** The rejected class is a surface with a *runtime* of its own — a
>   browser, a server, a windowing toolkit, a separate thing to install, run,
>   and keep alive. `company-os tui` renders into the terminal the user already
>   invoked the CLI from, is compiled into the same static binary, adds no
>   runtime dependency, and reaches no capability the flags do not. It is a
>   second renderer over the same records, alongside text and `--json` — not a
>   second surface. It remains subject to every other invariant above: no TUI
>   action writes anything the equivalent flag invocation would not, and each
>   one is previewed as a reproducible command line.
> - **"Any new runtime dependency beyond `pyyaml` + git" governs the user's
>   machine.** Build-time Go modules linked into a static binary are not
>   runtime prerequisites. The Go binary satisfies this bullet's intent more
>   completely than the implementation it was written against: `pyyaml` and the
>   interpreter it needs are both gone, and git is still required only for
>   federation. See
>   [Amendment 1](../ears/federation-enrichment.md#amendment-1--r-74-partial-retirement-2026-07-26).

## Success Criteria

- `company-os validate` exits 0 on `examples/workspace` before and after, with no artifact diffs from the flavor/federation machinery when defaults are used.
- Wizard path: `init` → `discover new` → `prd new` → `reality new` → `prd complete` succeeds end-to-end in a fresh directory using only printed next-command guidance.
- `company-os ids list` returns every ID in `company-ontology/ids/registry.yaml` and nothing else; a lifecycle command run with a misspelled ID names the unknown ID and suggests up to 3 closest registered matches.
- `today --role product-owner` renders a plain-language label alongside at least *deviation*, *exception*, and *EARS*; artifacts and validator output are byte-identical with and without `--role`.
- A team template override missing a required heading is caught by `validate` naming the missing heading; the same artifact from the built-in template passes.
- A team skill that shadows a canonical platform skill ID fails `validate`; the same skill with `extends:` passes and appears layered in discovery.
- In a two-repo fixture: `workspace sync` materializes only governance-relevant paths, writes SHA pins to `workspace.lock.yaml`, and a second `sync --frozen` run with the network disabled reproduces the identical tree.
- Deleting `workspace.yaml` restores exact monorepo behavior (regression fixture diff is empty).
