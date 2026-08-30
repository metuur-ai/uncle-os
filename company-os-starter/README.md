---
type: doc
tags: [doc/company-os-starter, kind/readme]
---

# company-os Starter Kit

Reference implementation of the Federated Company / Platform / Team OS:
tiered governance (mandatory / default / guidance), canonical skills,
a guiding CLI, validation gates, and a fully worked example.

## Contents

```text
cmd/company-os        CLI entry point (Go); `internal/` holds the packages
Makefile              make build | install | release | check
templates/            Example artifact formats — not contracts (see
                      templates/README.md); discovery, PRD, ADR, outcome,
                      reality doc, deviations, exceptions, SKILL template
skills/               Canonical skills: running-discovery, creating-prd,
                      completing-a-change, reality-from-prds,
                      requesting-an-exception
schemas/SCHEMAS.md    Human-readable artifact contracts
docs/FRONTMATTER-CORE.md  The minimal frontmatter core — the shared interop
                      contract for teams, tools, and Obsidian
docs/TUTORIAL.md      End-to-end walkthrough with real command outputs
../examples/workspace/  Populated company + platform + team (in the repo root, not shipped), including a
                      completed PRD, deviations, an exception, and a
                      personal-rules example in scratchpad/
```

## Quick start

The CLI is a single static binary with no runtime dependency — download a
release artifact, `chmod +x`, and put it on your `PATH`. From a source
checkout, `make install` does the same thing (`PREFIX=` to change where it
lands); the Go toolchain is a build-time requirement only.

```bash
make install                 # -> ~/.local/bin/company-os
export PATH="$HOME/.local/bin:$PATH"
cd ../examples/workspace
company-os governance resolve --team customer-engagement
company-os today --role product-owner
company-os validate
```

Then follow `docs/TUTORIAL.md` for the full discovery → PRD → complete loop.

## Browsing without memorizing commands: `company-os tui`

`company-os tui` opens a menu-driven terminal UI. It is safe to run as your
first command — it is one of the few that works outside a workspace, where it
offers to scaffold one.

It is the only way to reach four views:

- **workspace overview** — what exists, at a glance
- **component browser** — descriptors and their reality docs
- **PRD browser** — active change records and archived ones
- **discovery browser** — briefs by team

It also mirrors read-only commands (`today`, `validate`, `governance explain`,
`skills list`, `ids list`, `workspace status`) and gives one-key repairs for the
two things people forget: regenerating derived state (`derive`) and re-resolving
governance after a deviation.

### The whole loop, without leaving the menu

Every step of a change has a guided form. Each one previews the exact
`company-os` command it is about to run and does nothing until you confirm it,
so the menu teaches the CLI rather than hiding it.

| Step | Screen | Command it runs |
|---|---|---|
| Write a discovery brief | new discovery brief | `discover new` |
| Validate it | validate discovery brief | `discover validate` |
| Write the PRD from it | new PRD | `prd new` |
| Check the PRD | validate PRD | `prd validate` |
| Describe current state | new reality doc | `reality new` |
| Finish the change | complete PRD | `prd complete` |

Plus `add team|platform|component` for growing the federation.

```bash
company-os tui        # and that is the whole loop
```

Two things the menu deliberately will not do:

- **`prd complete --force` has no field.** It overrides the check that a change
  is not done until reality is updated. That flag stays where using it is a
  deliberate act — typed, at a terminal.
- **No browsing screen writes anything.** The browsers list; the forms write and
  say so in their titles. `discover validate` rewrites the brief it is given, so
  it lives with the forms, never in the discovery browser.

Still unreachable from the menu, and staying that way: `workspace sync` and
`scratchpad init`. Both need a value that is not in the workspace — a repo URL
and a commit pin, a path outside the tree — and a form that writes a
plausible-but-wrong value is worse than a missing one.

If you run the loop often, the CLI is still the faster surface: `discover
validate`, `prd validate`, `prd complete` and `prd new` all infer `--team`,
`--platform` and `--components` when the workspace admits one answer, so the
flag-free forms are complete commands rather than abbreviations.

## Design rules encoded here

1. Strict on process and structure, flexible on document formats — validators
   check the shared contract (frontmatter core: identity, lifecycle, references,
   derived tags; see docs/FRONTMATTER-CORE.md). Section structure is team-local
   guidance unless a team opts in via standards/doc-formats.yaml. A company or a
   single team can adopt the OS jointly, independently, or alongside other tools.
2. Rules have tiers; mandatory rules are outcomes, not implementations.
3. Deviations (default rules) and exceptions (mandatory rules) are explicit,
   expiring, and validated in CI.
4. Component descriptors are the single source for platform links and ownership;
   everything else is reconciled against them.
5. A change is done only when the Representation of Reality is updated —
   `prd complete` enforces it.
6. `generated/` files are derived, never hand-edited; CI regenerates and diffs.
7. Personal skills live in git-ignored `scratchpad/personal-rules/` and layer
   on top of canonical skills; mandatory steps always win.
