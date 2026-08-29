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
                      completing-a-change, requesting-an-exception
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
`skills list`, `ids list`, `workspace status`), offers guided forms for
`discover new`, `prd new` and `add team|platform|component`, and gives one-key
repairs for the two things people forget: regenerating derived state (`derive`)
and re-resolving governance after a deviation.

### What the TUI cannot do — read this before relying on it

**The TUI can start a unit of work; it cannot finish one.** These steps have no
menu entry and must be run from the CLI:

| Step | Command |
|---|---|
| Validate a discovery brief | `company-os discover validate <brief-id>` |
| Validate a PRD | `company-os prd validate <prd-id>` |
| Scaffold a reality doc | `company-os reality new --platform <p> <component>` |
| Complete a PRD | `company-os prd complete <prd-id>` |

The consequence you would otherwise hit without warning: **a brief created in
the TUI does not appear in the TUI's own "new PRD" form.** That form lists only
briefs with `status: validated`, and the command that validates one is not
reachable from the UI. This is deliberate — `discover validate` rewrites the
brief, so it is a mutation, and the TUI refuses to hide mutations behind
browsing — but the effect is a gap you have to step around.

The full loop, mixing both surfaces:

```bash
company-os tui                                  # create the brief
company-os discover validate <brief-id>         # CLI — the TUI cannot
company-os tui                                  # create the PRD (brief now listed)
company-os prd validate <prd-id>                # CLI
# ... deliver the change, update the reality doc ...
company-os prd complete <prd-id>                # CLI
```

If you run the loop often, the CLI is the faster surface anyway: `discover
validate` and `prd new` both infer `--team`, `--platform` and `--components`
when the workspace admits one answer, so the flag-free forms above are complete
commands, not abbreviations.

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
