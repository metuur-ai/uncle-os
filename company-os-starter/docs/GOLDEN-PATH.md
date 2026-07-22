---
type: doc
tags: [doc/company-os-starter, kind/golden-path]
---

# Golden Path: from an empty directory to a completed change

This is the shortest route through `company-os`, start to finish, on a workspace
**you** scaffold — no example repo required. Every command below was executed
against the reference CLI in `bin/company-os`; the `next:` lines are quoted
verbatim from what the CLI prints, so you can follow the tool instead of this
page once you are moving.

**The principle behind everything:** strict on artifacts, flexible on process.
The CLI scaffolds and guides; the validators enforce only the contract (schemas,
lifecycle gates, ownership, expiries, tag derivation). How you draft the content
is your business.

Already know the shape and want the fully-populated example instead? Read
[TUTORIAL.md](TUTORIAL.md).

---

## 0. Environment prerequisites (this is where first runs fail)

The CLI is one self-contained Python file. It needs Python 3.9+ and PyYAML —
nothing else.

```bash
python3 --version            # 3.9 or newer
pip install pyyaml           # the only dependency
export PATH="$PWD/company-os-starter/bin:$PATH"   # so `company-os` is on PATH
company-os --help            # sanity check
```

**Workspace-root resolution.** Every command except `init` and `scratchpad`
operates on a *workspace root* and will fail fast if pointed at the wrong
directory. The root is resolved in this order (highest wins):

1. `--root /abs/path` on the command line
2. `$COMPANY_OS_WORKSPACE_ROOT`
3. the current working directory

Run a workspace command outside a root and the CLI tells you exactly this:

```text
$ company-os discover new --team solo "Faster checkout"
error: '/.../ws' is not a workspace root: none of company-os/, platforms/,
  teams/, company-ontology/ found here.
  workspace root resolution order: --root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory
```

Throughout this guide we pass `--root "$W"` explicitly so the commands are
copy-pasteable from anywhere. In day-to-day use you would instead `cd` into the
workspace or export `$COMPANY_OS_WORKSPACE_ROOT` once.

```bash
W=$(mktemp -d)/ws; mkdir -p "$W"    # a throwaway workspace for this walkthrough
```

---

## 1. Setup — scaffold the workspace

`init` writes the four peer roots (`company-os/`, `platforms/`, `teams/`,
`company-ontology/`) and derives their generated artifacts so a fresh workspace
validates green. Pass `--company/--team/--platform` to run non-interactively
(required when there is no terminal, e.g. CI); omit them and the CLI prompts.

```bash
company-os --root "$W" init --company Acme --team solo --platform core
```

```text
  wrote index platforms/core/generated/feature-index.yaml
  node company-os/CLAUDE.md
  node platforms/core/CLAUDE.md
  node teams/solo/CLAUDE.md
  node company-ontology/CLAUDE.md
initialized workspace at /.../ws
  company: Acme | first team: solo | first platform: core
next: cd /.../ws && company-os discover new --team solo "<discovery title>"
```

`init` **refuses to run inside an existing workspace**, mutating nothing:

```text
$ company-os --root "$W" init --company Acme --team solo --platform core
error: '/.../ws' is already a workspace root (company-os/, platforms/, teams/,
  company-ontology/ present) — refusing to re-init
```

---

## 2. Discovery — capture the problem, then validate it

Discovery is team-private. `discover new` scaffolds a brief; the brief id is
`<year>-<slugified-title>`.

```bash
company-os --root "$W" discover new --team solo "Faster checkout"
```

```text
created teams/solo/product/discovery/2026-faster-checkout/brief.md
next: fill Problem signal, Hypothesis, Success criteria, then run: company-os discover validate --team solo 2026-faster-checkout
```

Fill the three sections in the brief, then validate. Section-emptiness is
**format guidance, not a gate** — the CLI `warn`s but still validates unless your
team opts into enforcement via `standards/doc-formats.yaml`. Validation flips the
brief to `status: validated`.

```bash
company-os --root "$W" discover validate --team solo 2026-faster-checkout
```

```text
  [warn] section 'Problem signal' is empty — format guidance only; the team may use its own structure (opt in via standards/doc-formats.yaml)
  ...
  [ok] brief '2026-faster-checkout' validated (status: validated)
next: company-os prd new --team solo --from-discovery 2026-faster-checkout --platform <platform-id> --components <comp-id,...>
```

---

## 3. PRD — turn the validated brief into a change record

`prd new --from-discovery` requires the brief to be `validated` and copies its
Problem/Success sections forward into a platform change record. It also injects
the applicable-governance checklist and, if a named component has no reality doc
yet, prints the exact command to scaffold one.

```bash
company-os --root "$W" prd new --team solo --platform core \
  --from-discovery 2026-faster-checkout --components checkout-service
```

```text
  [warn] component 'checkout-service' has no resolved governance (not in team ownership? run governance resolve)
created platforms/core/change-records/active/2026-faster-checkout/prd.md
  note: component 'checkout-service' has no reality doc yet — scaffold it: company-os reality new --platform core checkout-service
next: fill Proposed change + decisionOwner, then: company-os prd validate --platform core 2026-faster-checkout
```

> The governance warning is expected on a brand-new workspace: `checkout-service`
> isn't in the platform catalog or the team's ownership registry yet, so no
> platform rules resolve for it. To make a component a first-class catalog
> citizen, use `company-os add component --platform core checkout-service` and
> register ownership; for this golden path we proceed with the ad-hoc component.

Now fill `decisionOwner` and the `## Proposed change` section (the CLI told you
to), then validate. `prd validate` **fails** on a missing/`TODO` process-contract
field (`decisionOwner`, `governanceSnapshot`, …) and only `warn`s on empty body
sections:

```bash
company-os --root "$W" prd validate --platform core 2026-faster-checkout
```

```text
  [ok] PRD '2026-faster-checkout' passes the process contract
next: deliver the change, update the reality doc for each component, then: company-os prd complete --platform core 2026-faster-checkout
```

---

## 4. Reality update — the step that used to be a dead-end

A change is not done until the **Representation of Reality** reflects it. Scaffold
the reality doc for each component the PRD touches. This is the same command the
CLI already pointed you at in step 3.

```bash
company-os --root "$W" reality new --platform core checkout-service
```

```text
created platforms/core/reality/components/checkout-service.md
next: fill in Business rules / Current limitations, then continue: company-os prd complete --platform core <prd-id>
```

Fill in the reality doc's `## Business rules` / `## Current limitations` and make
sure its `updated:` date is on or after the PRD's `created:` date — `prd complete`
checks exactly that.

---

## 5. Completion — archive the change

`prd complete` enforces the done-check before archiving: every `- [ ]`
governance-checklist item must be checked off (with linked evidence), and each
component's reality doc must be newer than the PRD. On success it moves the PRD to
`archive/prds/`, writes an `outcome.md` due in 90 days, and appends `log.md`.

```bash
company-os --root "$W" prd complete --platform core 2026-faster-checkout
```

```text
archived -> platforms/core/archive/prds/2026-faster-checkout
outcome review scheduled (due 2026-10-20)
appended platforms/core/log.md
  wrote index platforms/core/generated/feature-index.yaml
  node platforms/core/CLAUDE.md
next: company-os validate
```

`prd complete` re-derives the workspace's generated artifacts for you (feature
index + CLAUDE.md nodes) as its last step, so the tree is left in sync and its
guidance points straight at `validate` — no manual `graph build` needed.

If the done-check fails, the CLI lists each reason and, for any missing reality
doc, prints `fix: company-os reality new --platform core <cid>` — so you are never
left guessing.

---

## 6. Run the CI gate

Run the workspace CI gate. It exits non-zero on any problem and is the same
gate CI runs:

```bash
company-os --root "$W" validate
```

```text
[1/7] ownership reconciliation
[2/7] deviation and exception expiry
[3/7] active PRD contracts
[4/7] frontmatter core and tag derivation (interop contract)
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
[6/7] feature-index drift (derived component->artifact map)
[7/7] custom skills layering (shadowing + extends resolution)

PASS
```

(A federated workspace — one with a `workspace.yaml` manifest — adds an eighth
gate, `[8/8] federated slice integrity`; monorepo mode shows the seven above.)

`PASS` (exit 0) means the whole loop closed: discovery validated, PRD archived,
reality updated, outcome review scheduled, and every derived artifact in sync.

---

## The chain at a glance

```text
init  ──▶ discover new ──▶ discover validate ──▶ prd new --from-discovery
      ──▶ (fill + prd validate) ──▶ reality new ──▶ prd complete
      ──▶ validate  ▶  PASS
```

Each command prints the next one. When in doubt, do what the last `next:` /
`fix:` line told you — the chain runs unbroken from `init` all the way to a
green `validate`.
