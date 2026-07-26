---
type: hld
id: hld-go-cli-tui-port
title: Go CLI + TUI Port — High-Level Design
status: draft
tags: [kind/hld, status/draft]
---

# Go CLI + TUI Port — High-Level Design

## Overview

Replace `company-os-starter/bin/company-os` — 2781 lines of Python requiring an
interpreter and a vendored PyYAML — with a single statically linked Go binary
that a user copies into `~/.local/bin` and runs. The port carries all 16
existing commands forward with byte-identical output, adds `--json` output and
per-category exit codes for agents and CI, and — in a second, strictly
downstream phase — an interactive terminal UI behind `company-os tui`.

The port is not a rewrite of behavior. Every gate, every scaffold, every
validation rule stays exactly as it is today. What changes is the runtime
(interpreter → static binary), the internal shape of how findings are computed
(prose strings → records rendered at the edge), and the surfaces available on
top of those records.

Source analysis: `.devlocal/research/2026-07-26-cli-tui-and-agent-interface.md`.
Reviewed by Product Owner and Technical Lead; their MUST-FIX findings are
incorporated throughout and the material ones are named in
`docs/lld/go-cli-tui-port.md` under Key Decisions.

## Phasing

The change ships in two phases under one spec set. Phase 2 does not start until
Phase 1's parity gate is green.

**Phase 1 — the port.** All 16 commands at parity, `--json`, exit codes,
`--version`, Python retired. This is where all the value the change can defend
on evidence lives: one-file install, structured output for agents, distinguishable
exit codes for CI.

**Phase 2 — the TUI.** Read-only surfaces first, mutating forms after, each one
justified by an observed request rather than by completeness. Phase 2 is the only
unbounded element of the scope and the only part with no parity oracle, so it is
sequenced last and budgeted separately.

## Stakeholders & Impact

**Engineers** keep the flag CLI unchanged. Every command they run today runs
identically tomorrow, with the same output and the same exit-0 semantics. They
gain `--json` where they previously parsed prose, distinguishable non-zero exit
codes where everything was previously `1`, and an install that is one `cp`.

**Agents** are already first-class consumers and already run non-interactively —
all four shipped skills in `company-os-starter/skills/*/SKILL.md` instruct agents
to run specific commands. They do not need a new mode. They need machine-readable
findings and exit codes they can branch on without parsing English. Both arrive
in Phase 1.

**Adopters running the CLI in CI** see no break. `validate` still exits 0 on a
clean workspace and non-zero on a dirty one; the new codes are all non-zero, so
existing `if company-os validate` gates behave identically. Their install path
changes from "clone, `pip install pyyaml`, run `install.sh`" to "download one
binary."

**Product owners and business analysts** are the audience recorded as blocked:
`.devlocal/research/2026-07-22-simplicity-user-journey-review.md:96` records
"Terminal-only, zero interactivity… Blocking — POs/BAs must learn zsh, PATH, env
vars before step 1." This change removes the Python and pip prerequisites. It
does **not** remove PATH, environment variables, the terminal itself, or the
domain vocabulary — friction items F5 (35-40 domain terms) and F6 (canonical ID,
`@spec`, and EARS syntax typed exactly and unassisted) are untouched by anything
here and remain open. Phase 2 addresses the navigation half of the wall. Claiming
this change solves PO onboarding would be false, and Success Criterion 9 exists
to keep that claim honest.

**Maintainers** trade a language they can edit quickly for a distribution story
that works. The program is roughly 80% YAML and markdown transformation, which
is Python's strongest suit; Go is estimated at 4500–6000 lines against today's
2781. This is a real cost, accepted deliberately because distribution is the
binding constraint. The Go codebase needs a named owner before Phase 1 starts —
there is no Go precedent in this repository.

## Goals

1. **A user copies one binary and starts working.** No interpreter, no
   `pip install`, no vendored dependency directory, no generated bash launcher.
   `install.sh` and `company-os-starter/vendor/` are deleted.
2. **Behavior is preserved provably, not by inspection.** A cross-implementation
   differential harness runs both binaries over identical fixtures across all 16
   commands and diffs stdout, stderr, exit code, and the resulting filesystem
   tree. The golden snapshots alone are not sufficient evidence and are not
   treated as such.
3. **Findings exist as data before they exist as text.** Every gate, conflict,
   and validation error is a record. Human text and JSON are two renderers over
   the same records.
4. **Exit codes distinguish failure categories**, under a documented contract,
   with every failure code remaining non-zero.
5. **The Python CLI is deleted** once parity is proven by the differential
   harness. One implementation, not two.
6. **An interactive TUI covers the workspace** (Phase 2), launched only by
   `company-os tui`, read-only surfaces first. Every mutating action it performs
   is expressible as a flag-complete command, and the TUI shows that command
   before running it.

## Non-Goals

- **Behavior changes of any kind**, except the carve-outs enumerated in
  `docs/ears/go-cli-tui-port.md` R-0.7a. Where the Go binary differs from the
  Python one outside that list, the Go binary is wrong.
- **OKF v0.2 Phases 1–3.** They move fixture frontmatter and `validate` output,
  and moving the golden baseline and the language at once makes a parity failure
  ambiguous. **OKF Phase 0 is explicitly carved out of this deferral** — it is a
  30-minute done-gate date-parsing fix that moves neither golden, and deferring
  it means porting a known bug and re-fixing it later.
- **Merging with `local-search`.** The two stay separate binaries composing by
  convention. Merging remains possible later.
- **Windows.** Release targets are `darwin/arm64`, `darwin/amd64`, and
  `linux/amd64`. Go cross-compiles to Windows for free, but nothing here is
  verified there, so nothing is claimed.
- **A web or GUI surface.** `docs/hld/golden-path-flavor-federation.md:47-52`
  decided against these and that stands. A terminal UI in the same binary, with
  no runtime dependency, is ruled inside the line by this document.
- **Solving the PO vocabulary wall.** Friction F5 and F6 are out of scope and
  stay open.
- **A transition period in which both implementations ship.**

## Success Criteria

**Phase 1**

1. The cross-implementation differential harness reports zero divergence between
   the Python and Go binaries across all 16 commands — stdout, stderr, exit code,
   and resulting filesystem tree — over the `workspace`, `standalone-team`, and
   `federated` fixtures, on both the passing and the failing paths.
2. `examples/acceptance.sh` passes against the Go binary with zero edits to the
   script and zero diff against either golden file, including the failure-path
   goldens captured under R-0.9.
3. Every one of the 85 real assertions in `examples/selftest.py` has a named Go
   test, tracked against `.devlocal/go-port/selftest-inventory.md`, before
   `selftest.py` is deleted — noting that selftest covers only 7 of 16
   subcommands, so this is necessary but not sufficient for coverage.
4. **Downloaded** release artifacts — not locally built binaries — install and run
   on a clean macOS arm64 machine and a clean Linux amd64 machine, neither with
   Python installed, following only the published install documentation.
5. For `validate` on both fixtures, the set of `{gate, code, severity, subject}`
   tuples in `--json` output is exactly the set the text renderer reports,
   asserted by test.
6. `company-os --version` reports a version and build identifier, and that
   identifier appears in every `--json` payload.
7. `docs/ears/federation-enrichment.md:149` (R-7.4) has its single-file and
   output-helper clauses formally retired, with the retirement recorded rather
   than silently overridden, and its frontmatter-parser and guidance-chain clauses
   left in force.
8. `company-os-starter/bin/company-os`, `install.sh`, `vendor/`, and
   `examples/selftest.py` no longer exist in the tree, and the final Python commit
   is tagged for rollback.

**Phase 2**

9. Given only a download link and no prior use of the tool, a product owner on a
   fresh macOS machine reads workspace status, opens a component's governance, and
   inspects a validate failure — unassisted and without prior shell configuration.
   If nobody will run this test, Phase 2's PO justification is struck and the TUI
   is re-justified on engineer value alone.
10. `company-os tui` with no TTY attached exits 7, prints an explanatory message,
    and changes nothing.
11. Every TUI screen is enumerated in the spec and asserted by test. "Reaches all
    16 commands" is not a measurement and is not used as one.
