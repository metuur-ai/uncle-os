---
type: hld
slug: derived-drift-repair
status: locked
created: 2026-08-23
---

# Derived Drift Repair (`validate --fix`) — High-Level Design

## Overview

`company-os validate` composes eight gates and reports every finding as pass or fail.
Some of its codes are not decisions — they are recomputations. The system already
knows the correct value, computed it once, and can compute it again:

| Code | Gate | Severity today | Condition |
|---|---|---|---|
| `frontmatter.tags-drift` | 4 | fail | a document's `tags:` no longer equals the derivation from its own frontmatter |
| `node.drift` | 5 | fail | a `CLAUDE.md` generated block no longer equals the freshly built block |
| `feature-index.drift` | 6 | fail | `platforms/<p>/generated/feature-index.yaml` no longer equals the fresh derivation |
| `node.absent` | 5 | ok | a federation root has no `CLAUDE.md` at all |
| `feature-index.absent` | 6 | ok | a platform has no committed feature index |

The last two are included deliberately, and the reason is not cosmetic. `graph build`
writes those files (`node.go:69-81`, `featureindex.go:449-460`). If `--fix` triggered
only on the three failing codes, then a workspace with an absent node would end up in
a *different* state than the sanctioned `graph build && validate` sequence — and the
equivalence claim below would be false. The trigger set is therefore "every condition
the rebuild would write for," not "every condition that is red."

Every other failing code in the eight gates encodes a decision a human has to make:
which side of an ownership conflict is right, what the new expiry date should be,
whether the reality doc is actually current, whether a foreign slice was tampered
with. Those must stay red.

This change adds one flag, `--fix`, that re-derives exactly the conditions above,
re-runs the whole gate set against the repaired workspace, and reports what it
repaired. Nothing else changes.

### A disagreement this change resolved at the source

Gate 5 reports a `CLAUDE.md` that has no generated markers as `node.hand-owned`,
severity ok, and moves on — it treats a hand-written context node as legitimate
(`gates.go:157-160`). The writer used to disagree: `rewriteGeneratedBlock` appended
a generated region to any file lacking markers, so `graph build` silently converted
hand-owned nodes into managed ones.

`--fix` would have made that disagreement matter, by moving the conversion from "you
explicitly ran a build command" to "you fixed a tag typo." Rather than guard the
repair path against a writer we did not trust, the writer was corrected: it now
leaves marker-less files byte-identical, and both `Build` and `Rebuild` inherit that
refusal from the one function they share.

The consequence for this design is that the guard is structural rather than
procedural. `--fix` needs no list of hand-owned roots and never consults the first
run's findings to decide what to skip — a file without markers is simply never
written, by anything. R-2.11 states the property; nothing in the repair path
implements it.

## Stakeholders & Impact

**People editing artifacts in an Obsidian vault** are the population this is for.
Company OS's invariants say never hand-edit `tags:` and never hand-edit `generated/`.
Obsidian's tag autocomplete does exactly the first thing, and a person editing in a
vault has no way to tell which files are derived. Today their only feedback is a red
gate and an instruction to go run a different command. After this ships they run the
command they were already running with one flag and the workspace is green.

**Agents** get the same benefit mechanically: a repair that is a deterministic
function of authored input stops consuming a turn.

**CI** is unaffected by default. `--fix` is opt-in; a pipeline that never passes it
sees byte-identical behaviour, including exit codes and golden output.

**Nobody's decisions are touched.** A reviewer looking at a `--fix` run sees every
decision-class finding it could not repair, unchanged and still failing.

## Goals

1. A workspace whose only failures are derived-condition codes reaches exit 0 in one
   command.
2. No decision-class finding is ever repaired, suppressed, downgraded, or reworded
   by `--fix`.
3. Repair is idempotent: a second `--fix` on the same workspace writes no bytes.
4. When repair is attempted but impossible — the derived file lives inside a
   read-only synced slice — the run says which file and what the remedy is, and
   changes nothing.
5. `--fix` produces the same on-disk result as the sanctioned manual sequence
   (`graph build`, then `validate`) — with no exceptions, now that neither path
   adopts hand-owned nodes. It is a shortcut, not a second implementation.

## Non-Goals

- **No new derivation logic.** Repair reuses the writers `graph build` already runs.
  If a value cannot be derived today, `--fix` does not derive it.
- **`generated/effective-governance.yaml` is out of scope.** Verified against source:
  there is no gate for it. `validate` never resolves governance and never diffs the
  derived file; staleness surfaces only as an advisory `[warn]`, and only when the
  file is missing entirely. `--fix` cannot repair a gate that does not exist. Adding
  that gate is a separate, sequenced change; only once it exists does the row belong
  in the derived class.
- **Gate 8 (federation slice hash integrity) is never repaired.** It is the
  tamper-evidence gate. Re-hashing on a hand-edit would destroy the only signal it
  carries. This is a hard invariant, not a default.
- **The `check` verb rename is not part of this.** `company-os check` today means
  `check ready|done` — product readiness — and is a different command. This change
  puts the flag on `validate`. If the verb surface is later reshaped, the flag rides
  along with whatever `validate` becomes.
- **Nothing adopts hand-owned context nodes — including `graph build`.** A
  `CLAUDE.md` without generated markers is left exactly as written, by every writer.
  Reconciling the gate and the writer was originally scoped as a separate change;
  it was pulled in here instead, because guarding only the repair path would have
  left `graph build` converting files behind the user's back and would have made the
  equivalence claim in SC5 false by construction.

- **No interactive prompting, no dry-run diff rendering, no partial selection.**
  `--fix` repairs every derived condition it is triggered by, or none.

## Success Criteria

Each is reachable by a reviewer with a checkout and no knowledge of the
implementation.

- **SC1 — the headline case.** Take a green example workspace, hand-edit one `tags:`
  line in one document. `validate` exits 1 and reports `frontmatter.tags-drift`.
  `validate --fix` exits 0, reports one repaired document, and leaves the file with
  its derived tags restored.

- **SC2 — decisions survive repair.** From that same workspace, additionally set a
  deviation's `reviewDate` to a past date. `validate --fix` exits 1. The
  `frontmatter.tags-drift` finding is absent from the output; the expiry finding is
  present and identical to what plain `validate` reports.

- **SC3 — idempotence.** On a workspace where `graph build` writes nothing,
  `validate --fix` exits 0 and leaves `git status --porcelain` empty. On any
  workspace, a second consecutive `validate --fix` writes no bytes.

- **SC4 — read-only refusal.** Given a workspace with a synced slice whose target
  contains a drifted derived file, `validate --fix` exits non-zero, names that file,
  states the remedy (change the source repo, bump the pin, re-sync), and leaves the
  file unmodified with its permissions unchanged.

- **SC5 — equivalence with the manual path.** From any workspace — including one
  containing a hand-owned context node — `validate --fix` and the sequence `graph
  build && validate` leave the workspace tree in identical states, byte for byte.
  The hand-owned file is untouched on both paths.

- **SC6 — no exit-code 2.** No `--fix` invocation that parsed successfully exits 2.
  The invocation was actionable; a residual failure is a validation result, not a
  usage error.
