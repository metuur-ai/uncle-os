# Golden Path, Flavor & Federation — Low-Level Design

Sources merged: `.devlocal/research/2026-07-22-simplicity-user-journey-review.md`,
`.devlocal/research/2026-07-22-custom-templates-and-skills.md`,
`.devlocal/research/2026-07-22-multi-repo-workspace-federation.md`.

Related spec: `docs/{hld,lld,ears}/federation-enrichment.md` builds on subsystem D below; requirement IDs in this spec use the `GPF-` prefix to avoid collision.

## Architecture

All changes live in the CLI plus workspace files. Four subsystems, layered so each is independently shippable:

> **Amended 2026-07-26.** This line originally read "the single self-contained CLI (`company-os-starter/bin/company-os`)". The single-file constraint was retired by [Amendment 1](../ears/federation-enrichment.md#amendment-1--r-74-partial-retirement-2026-07-26); the CLI is now the Go module at `company-os-starter/cmd/company-os` over `internal/`. Read every `bin/company-os` reference below as naming the CLI, not a file.

### A. Golden path & UX layer (umbrella)

- **`company-os init` wizard** — new command. Interactive prompts (company name, first team, first platform) scaffold the four peer roots (`company-os/`, `platforms/<p>/`, `teams/<t>/`, `company-ontology/`) from the same templates the rest of the CLI uses, seed `ids/registry.yaml`, then print the first golden-path command. Refuses to run in a non-empty workspace. Non-interactive flags (`--company/--team/--platform`) for scripted use.
- **`company-os add platform|team|component`** — grows an existing workspace from the same scaffolding templates `init` uses; refuses to overwrite anything existing. Closes friction F2 (hand-copying YAML for the second team/platform/component).
- **`company-os reality new --platform <p> <component-id>`** — scaffolds `reality/components/<id>.md` from `templates/reality-component.md`; `prd new` prints it as the next command when a target component's reality doc is missing, keeping the guidance chain unbroken through `prd complete` (friction F3).
- **Fail-fast root check & parser consistency** — workspace-scoped commands fail fast outside a workspace root, naming the resolution order (`--root` → `$COMPANY_OS_WORKSPACE_ROOT` → cwd) instead of succeeding silently (F8); all frontmatter parsing goes through the single `frontmatter()` contract (F11).
- **Golden-path document** — one doc (extends `docs/GETTING-STARTED.md`) covering environment prerequisites (download the binary, PATH) → setup → discovery → PRD → reality update → completion; every step matches the CLI's printed guidance chain. The existing "every mutating command prints the next command" convention is the backbone and must be preserved and extended to the new commands.
- **ID lookup** — new `company-os ids list [--team|--platform|--prefix]` reading `company-ontology/ids/registry.yaml` (the canonical source; never a parallel index). Unknown-ID failures in existing commands gain closest-match suggestions (stdlib `difflib`; no new dependency).
- **Role views & terminology** — extend the existing `today --role` mechanism: a static translation table in the CLI maps canonical terms → plain-language labels per role (e.g. product-owner sees "promise with an expiry date" next to *exception*). Display-only: artifacts, IDs, tags, and validators always use canonical vocabulary.

### B. Template overrides (flavor, not contract)

- **Resolution chain, first found wins:** `teams/<t>/templates/<name>.md` → `platforms/<p>/templates/<name>.md` → `company-os/templates/<name>.md` → built-in `*_TEMPLATE` string in `bin/company-os`. One helper (`resolve_template(name, team, platform)`) used by every scaffolding command (`discover new`, `prd new`, `scratchpad init`, …).
- **Template format:** plain markdown body with the CLI's placeholder substitutions. Templates control body/flavor only.
- **CLI-owned regardless of template:** YAML frontmatter (`type:`, IDs, `tags:` derivation inputs), status fields, and the `deviations.yaml`/`exceptions.yaml` row seeds — injected by the CLI, never sourced from a template (seeds templatable = explicitly deferred).
- **Contract enforcement:** `validate` already greps artifacts for required section headings; it stays the single arbiter. To make drift impossible by construction, the required headings live in shared `*_SECTIONS` constants (e.g. `DISCOVERY_SECTIONS`, `PRD_SECTIONS`) consumed by **both** the built-in `*_TEMPLATE` strings and the validate greps — one source, two consumers. A custom template that drops a required heading produces artifacts that fail validate with the missing heading named — the template itself is not validated, its outputs are (strict on artifacts, flexible on process).

### C. Custom skills (four additive layers)

- **Layers:** company (`company-os/skills/`), platform (`platforms/<p>/skills/`), team (`teams/<t>/skills/`) — all shared/versioned — plus personal `scratchpad/personal-rules/` (git-ignored). Discovery merges all four; each skill's origin layer is shown.
- **Shadowing is a `validate` error:** a lower layer may not reuse a canonical skill's ID/name. Layering is explicit via `extends: platform-skill://…` frontmatter; the merged view presents base steps plus extension steps. The `platform-skill://` scheme is registered as a canonical ID scheme in the ontology conventions, alongside `component://`, `capability://`, `req://`, and `context://`.
- **Conflict rule:** canonical `(mandatory)` steps always win over personal rules, and the agent/CLI must say so when it happens (existing convention, now enforced in the merged skill view).

### D. Manifest-optional workspace federation (Option B)

- **Activation:** presence of `workspace.yaml` at the workspace root activates federated mode. Absence ⇒ monorepo mode, byte-for-byte unchanged. Detection sits in the existing workspace-root resolution (`--root` → `$COMPANY_OS_WORKSPACE_ROOT` → cwd) — one of only two CLI touch points; the other is path resolution for federated slices.
- **Manifest:** `workspace.yaml` lists remote repos (e.g. `acme/company-os.git`, `acme/platform-communications.git`, `acme/team-customer-engagement.git`) with **explicit pins**: `commit: <sha>` or `tag: v2.1.0` (tag is resolved to a SHA at sync). No floating `ref:`/branch pins.
- **`company-os workspace sync`:** for each repo, fetch only governance-relevant paths (`governance/`, `components/`, `requirements`, `reality/`, `skills/`, `templates/`) via sparse/filtered shallow git (`--filter`, sparse-checkout, `--depth`), materialize as **read-only** slices in the local workspace layout, and write resolved SHAs to `workspace.lock.yaml`.
- **`workspace status`:** reports drift between manifest pins, lock file, and materialized slices.
- **`--frozen` (CI):** no network; materialize strictly from `workspace.lock.yaml`; fail if lock is missing or doesn't cover the manifest.
- **Cross-layer interaction:** template/skill resolution (B, C) walks the same materialized layout, so federated platform templates and skills resolve identically to monorepo ones. Materialized slices are derived content: hand-edits fail `validate` (same rule as `generated/`).
- **Migration:** monorepo → federated is additive — split repos, add `workspace.yaml`, `sync`; no artifact format changes.

## Constraints

- No runtime dependency on the user's machine; git CLI required for federation (guarded — federation commands fail with a clear message if git is absent; monorepo mode never needs it). *(Written against the Python reference as "Python 3.9+, `pyyaml` only"; restated after the Go port. The policy governs runtime prerequisites, not build-time modules — see [Amendment 1](../ears/federation-enrichment.md#amendment-1--r-74-partial-retirement-2026-07-26).)*
- Preserve the `frontmatter()` parser contract (`^---\n...\n---\n`). *(The "one self-contained file" and `die`/`ok`/`warn`/`fail`-helper constraints were retired by that same amendment; the frontmatter contract was not.)*
- Templates in `templates/` and `*_TEMPLATE` strings must stay in sync with the headings `validate` greps for.
- No behavior change without `workspace.yaml`; no full-repo content ever fetched (privacy/size requirement that drove Option B).
- `workspace.lock.yaml` and materialized slices are machine-owned (never hand-edited), like `generated/`.

## Key Decisions

| Decision | Why | Rejected alternatives |
|---|---|---|
| Simplicity-first umbrella framing | UX findings showed jargon + ID gap + terminal-only entry block adoption; new capabilities must reduce, not add, surface | Shipping flavor/federation as standalone features (adds surface without a path) |
| Option B: sparse governance-slice sync with explicit SHA/tag pins | Reproducible, minimal, no full-repo inclusion | Git submodules (full-repo, UX-hostile — explicitly rejected); floating `ref:` (ambiguous, non-reproducible) |
| Manifest-optional activation | Zero migration cost; monorepo remains the default and the test baseline | Config flag in a settings file (invisible state); always-federated (breaks existing workspaces) |
| Validate outputs, not templates | Preserves "strict on artifacts, flexible on process"; one arbiter | Linting template files themselves (second contract to keep in sync) |
| Extend-only skills; shadowing = validate error | Keeps canonical skills authoritative and conflicts impossible by construction | Override/precedence rules (silent divergence between teams) |
| Terminology translation is display-only | Artifacts stay grep-able and canonical; ubiquitous language intact | Renaming terms in artifacts (breaks IDs, tags, `@spec` traceability) |
| Defer templatable YAML seeds & heading localization | No demonstrated need; both widen the contract surface | Speculative flexibility now |

## Out of Scope

- Templatable `deviations.yaml` / `exceptions.yaml` row seeds (deferred).
- Localized required-heading aliases in the validate contract (deferred; localized templates keep English headings).
- Git submodules or any full-repo federation mode.
- Web/GUI wizard; changes to the `@spec`/ontology roadmap items (`validate --ontology`, `spec trace`) — CLI source remains ground truth for what exists.
