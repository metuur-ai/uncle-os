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
