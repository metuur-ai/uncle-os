---
type: ears
id: ears-ux-simplification-phase-2
title: UX Simplification Phase 2 — EARS Specifications
status: draft
---

# UX Simplification Phase 2 — EARS Specifications

> **HELD — not locked.** Revised 2026-08-28 to apply product-owner, tech-lead and
> senior-engineer review. Held pending diagnosis validation
> (`.devlocal/research/2026-08-28-ux-diagnosis-validation-protocol.md`).
> Units 6 and 8 survived review unchanged and are cleared to land independently.

**Design:** `docs/hld/ux-simplification-phase-2.md`,
`docs/lld/ux-simplification-phase-2.md`
**Predecessor:** `docs/tasks/ux-simplification.md` (Phase 1, landed in full; the
complaint persisted — see HLD A5)

**Groups:** A = non-technical on-ramp (U1, U7, U9) — this change's success
condition. B = CLI/agent ergonomics (U2, U3, U5). C = contract documentation
(U6, U8).

---

## Unit 0: Global invariants

**Why:** Phase 1 established red lines this change inherits. They are stated
first because every later unit is subordinate to them.

| ID | EARS statement |
| --- | --- |
| R-0.1 | THE SYSTEM SHALL keep gates `[1/8]`–`[8/8]` and `[1/9]`–`[9/9]` semantically and textually identical. |
| R-0.2 | THE SYSTEM SHALL keep all exit codes unchanged. |
| R-0.3 | THE SYSTEM SHALL keep all five committed goldens byte-identical. |
| R-0.4 | THE SYSTEM SHALL NOT run `acceptance.sh --update` in any unit. |
| R-0.5 | IF a unit unexpectedly encounters a golden that pins its output, THE SYSTEM SHALL halt that unit and cite the specific golden file and line; a deferral without a cited line SHALL NOT be recorded as compliance. |
| R-0.6 | THE SYSTEM SHALL NOT introduce any check on how an artifact was produced. |
| R-0.7 | THE SYSTEM SHALL keep every command returning `[]model.GateResult`, with output written only by `cmd/` and `internal/render/`. |
| R-0.8 | THE SYSTEM SHALL NOT call `os.Exit` or print from any package below `cmd/`. |
| R-0.9 | THE SYSTEM SHALL NOT import package `main` from any `internal/` package. |
| R-0.10 | THE SYSTEM SHALL keep the `frontmatter()` parser contract untouched. |
| R-0.11 | WHERE a command mutates the workspace, THE SYSTEM SHALL print the next command in the workflow. |
| R-0.12 | WHEN `make check` runs, THE SYSTEM SHALL exit 0. |
| R-0.13 | WHERE inference is added, THE SYSTEM SHALL provide an explicit overriding flag, and SHALL behave byte-identically for every existing flag-carrying invocation. |
| R-0.14 | THE SYSTEM SHALL keep both copies of `TUTORIAL.md` in sync with each other AND accurate against real command output. |

---

## Unit 1: On-ramp documentation and boundary disclosure — Group A

**Why:** the primary audience is non-technical, and a menu-driven interface that
serves them has shipped unmentioned by any document. But it cannot complete a
unit of work: a brief created in the interface does not appear in the next
screen's picker, because that picker lists only validated briefs and the
validating command is unreachable from the UI. Documenting the interface without
that boundary would route the target audience into an unexplained dead end —
worse than the status quo, where they fail cheaply at minute two. This unit
therefore ships the interface and its edge together.

**✅ LANDED (disclosure half) 2026-08-28.** `make check` exit 0. The boundary is
stated in `company-os-starter/README.md` (new section) and in both `TUTORIAL.md`
copies (§0, before the first CLI invocation), each naming the four absent
commands, the empty-picker consequence and its rationale, and showing the
surface-crossing loop end to end.

**Amendment I — R-1.4 is unimplementable and is withdrawn.** It required
`tui --help` to name the screens and the gap. `help()` forbids a description
line: "none of the sub-parsers sets `description=`, so argparse prints none, and
adding one would add a human-facing line R-0.8 freezes"
(`cmd/company-os/main.go:214-217`). `tui` has no flags or positionals, so its
help body has nowhere else to put prose. A `long:` field was drafted and reverted
on discovering the rule; the reasoning now sits as a comment at the `tui`
commandSpec so the next reader does not retry it. The disclosure lives in the
docs instead — where someone deciding whether to rely on the UI is looking.

**Amendment J — R-1.2/R-1.3 retarget to `company-os-starter/README.md`.** The
repo-root `README.md` documents an unrelated graph-explorer web tool and carries
no CLI content, exactly as Phase 1 recorded
(`docs/tasks/ux-simplification.md:230`). A TUI section there would have been the
only CLI text in a document about something else.

**Amendment K — R-1.10/R-1.11 withdrawn on the workload evidence.** They would
have added a `company-os tui` line to `next`'s empty state and `init`'s
completion guidance. The owner's answers (~40 changes/month, design for today)
make a menu the slower surface for the actual user, and protocol §0.2 records
U1's demotion from centerpiece to cheap truth-telling. Advertising the UI in
command output no longer follows from the workload; disclosing its boundary in
the docs still does. Withdrawn rather than deferred, since the evidence against
them is on record.

| ID | EARS statement |
| --- | --- |
| R-1.1 | THE SYSTEM SHALL name `company-os tui` in both copies of `TUTORIAL.md` before the first CLI invocation is taught. |
| R-1.2 | THE SYSTEM SHALL name `company-os tui` in `README.md`. |
| R-1.3 | THE SYSTEM SHALL document the four capabilities reachable only through the interface: workspace overview, component browser, PRD browser, discovery browser. |
| R-1.4 | WHEN `company-os tui --help` runs, THE SYSTEM SHALL name its six read-only screens, five mutating flows and three one-key repairs. |
| R-1.5 | THE SYSTEM SHALL keep `company-os tui` free of flags. |
| R-1.6 | THE SYSTEM SHALL state which lifecycle steps the interface performs and which require the CLI, naming `discover validate`, `prd validate` and `prd complete` explicitly. |
| R-1.7 | THE SYSTEM SHALL state that a brief created in the interface will not appear in the PRD form's `from-discovery` picker until it is validated from the CLI. |
| R-1.8 | THE SYSTEM SHALL show the exact CLI commands that close the loop from an interface-created brief to a completed PRD. |
| R-1.9 | WHEN a reader follows only the documented path, THE SYSTEM SHALL permit completion of one full discovery → PRD → complete lifecycle. |
| R-1.10 | WHEN `company-os next` finds no pending action, THE SYSTEM SHALL offer `company-os tui` in addition to its existing guidance. |
| R-1.11 | WHEN `company-os init` completes, THE SYSTEM SHALL offer `company-os tui` in addition to its existing guidance. |
| R-1.12 | THE SYSTEM SHALL document that `company-os tui` may be run from outside a workspace and offers a recovery menu there. |
| R-1.13 | THE SYSTEM SHALL NOT add, remove or alter any interface screen in this change. |
| R-1.14 | THE SYSTEM SHALL NOT deprecate or remove any CLI capability in favour of the interface. |
| R-1.15 | THE SYSTEM SHALL locate the `init` guidance emitter before implementation; the previously cited `internal/scaffold/scaffold.go:74-75` is inside `Slugify()` and is not it. |

---

## Unit 2: Infer context for `prd new` — Group B

**Why:** after Phase 1 this is the only lifecycle step that still refuses to run
without IDs, and it is the step where the user knows least. In a one-platform
workspace every flag it demands has exactly one legal value.

**✅ LANDED 2026-08-28.** `internal/product/infer.go` — `InferPlatformForNew`,
`InferTeamForNew`, `InferComponentsForNew`; wired in `cmd/company-os/product.go`;
`a.Action == "new"` added to the suspension list at `cmd/company-os/args.go`.
R-2.13 verified by byte-diff: the flag-free and fully-flagged PRDs are identical.
`make check` exit 0, five goldens byte-identical. Four dispatch-level tests added
to `cmd/company-os/inference_test.go`.

**Amendment A — R-2.15 was exercised, then reversed on evidence.** Component
inference was first dropped, on the argument that `--components` has a legal
empty default and so omitting it was an existing working invocation protected by
R-0.13. **That argument was wrong, and testing showed it.** Omitting the flag
yields `components: []`, which fails *both* `prd validate`'s process contract and
gate 3 (`[FAIL] core/2026-…: missing frontmatter ['components']`). There was no
working behavior to protect: the flag-free path produced an artifact the tool
itself rejects. Inference therefore repairs a defect rather than altering a
success, and R-2.5 is implemented. Without it R-2.13 was unsatisfiable, since the
fully-flagged invocation supplies `--components`.

**Amendment B — components are inferred from effective governance, not from
`ownership/components.yaml`.** `PRDNew` calls `Gather` immediately afterwards,
and `Gather` resolves components against the generated file. Inferring from the
authored registry could propose a component `Gather` then reports as "not in any
generated governance file" — the tool contradicting itself inside one command.
Same source in, same source out; a stale generated file suppresses inference
rather than corrupting it.

**Amendment C — two parser tests repointed, not deleted.**
`TestArgumentErrorDiagnostics/missing_--platform` and
`…/surplus_positional_loses_to_the_required_check` both pinned `prd new` failing
at parse time, which is exactly what this unit changes. They document
*precedence*, which still needs a command whose flag the parser enforces, so both
were repointed to `reality new` — the last create-command with an unconditional
`--platform`. Related observation: with `--platform` suspended, `prd new id1
extra` now reports the surplus positional under the top-level usage line, which
is precisely what `prd validate id1 extra` and `prd promote id1 extra` have done
since Phase 1. Consistency with the already-suspended actions, not a regression.

**Amendment D — `ws.AllPlatforms()` returns directory PATHS, not ids**
(`internal/workspace/workspace.go:220-233`, `subdirs` joins onto the parent).
Passing them straight to `PlatformDir` produced
`platform '/abs/path/platforms/core' not found`. `filepath.Base` is applied
before use and before reporting candidates.

**Not implemented, deliberately:** team inference when `--from-discovery` is
absent. R-2.4 scopes team resolution to the brief's directory — a lookup, not a
guess. Inferring a lone team in a no-discovery `prd new` would exceed the
requirement; it is recorded here as a candidate, not silently added.

| ID | EARS statement |
| --- | --- |
| R-2.1 | WHEN `prd new` runs without `--platform` and exactly one platform is a candidate, THE SYSTEM SHALL proceed as if the flag had been supplied. |
| R-2.2 | IF more than one platform is a candidate, THE SYSTEM SHALL emit a usage error naming every candidate and exit 2. |
| R-2.3 | IF no platform is a candidate, THE SYSTEM SHALL emit the existing not-found error unchanged. |
| R-2.4 | WHEN `prd new` runs with `--from-discovery` and without `--team`, THE SYSTEM SHALL resolve the team from the discovery brief's directory. |
| R-2.5 | WHEN `prd new` runs without `--components` and the resolved team owns exactly one component, THE SYSTEM SHALL proceed as if the flag had been supplied. |
| R-2.6 | WHEN an explicit flag is supplied, THE SYSTEM SHALL use it and SHALL NOT infer. |
| R-2.7 | THE SYSTEM SHALL infer only on a unique match, and SHALL NOT select a best guess. |
| R-2.8 | THE SYSTEM SHALL place the new resolver in an `internal/` package, so that Unit 3 can call it. |
| R-2.9 | THE SYSTEM SHALL resolve the platform by enumerating `ws.AllPlatforms()`, NOT by `ws.FindPRD`, which requires a PRD that does not yet exist. |
| R-2.10 | THE SYSTEM SHALL suspend the `--platform` requirement by adding `a.Action == "new"` to the existing per-action list at `cmd/company-os/args.go:677-681`. |
| R-2.11 | THE SYSTEM SHALL NOT implement the suspension by setting `required: false`, which brackets the flag in every `prd` usage line (pinned at `cmd/company-os/args_test.go:636`). |
| R-2.12 | THE SYSTEM SHALL NOT change the usage or error text of `prd validate`, `prd complete`, `prd promote` or `prd abandon`. |
| R-2.13 | WHEN `prd new --from-discovery <id> "<title>"` runs on `examples/standalone-team` with no other flags, THE SYSTEM SHALL create a PRD byte-identical to the fully-flagged invocation. |
| R-2.14 | THE SYSTEM SHALL emit the help text `required unless uniquely inferable from the workspace` for `--platform`, `--team` and `--components` on `prd new`. |
| R-2.15 | IF `--components` inference cannot be shown to resolve uniquely under fixture testing, THE SYSTEM SHALL ship R-2.1–R-2.4 without R-2.5 and record the deferral. |

---

## Unit 3: Guidance chain emits resolved values — Group B

**Why:** the guidance chain is the product's strongest feature and it currently
prints `<platform-id>` in a workspace with one platform. Without this unit, Unit
2's benefit is invisible to anyone following the printed advice.

**✅ LANDED 2026-08-28.** `make check` exit 0, five goldens byte-identical. On a
single-platform workspace:

```
next: company-os prd new --team solo --from-discovery 2026-quiet-hours \
      --platform core --components solo-service
```

Verified runnable by executing the printed command verbatim. Ambiguous
workspaces still print both placeholders. Two tests added.

**Amendment E — resolution happens at build time, not render time.** R-3.3 named
`sections.go:98-101` and `discover.go:157` as the two sites, implying a change in
each. `sections.go` is a **pure renderer over `model.Fields`** with no workspace
handle, so it cannot resolve anything. The split that respects the existing
layering: `DiscoverValidate` (which has `ws`) resolves and puts `platform` and
`components` into the finding's Fields; `Message` renders them, falling back to
the original placeholders when a field is absent. Because `FieldNext` is produced
by calling `Message` over those same Fields, the rendered text and the `--json`
value stay in sync by construction rather than by a second substitution.

**Amendment F — inference errors are swallowed here, deliberately.** The Unit 2
resolvers return a usage error on ambiguity. That is right for `prd new`, which
must refuse; it would be wrong for `discover validate`, which would start failing
on workspaces it passes today. Ambiguity means only that there is nothing to
substitute, so the error is discarded and the placeholder stands. A test pins
this: ambiguity must not turn `discover validate` into a usage error.

**Testing note.** The first draft of both tests passed *vacuously* — the brief
fixture omitted the required `## Hypothesis` heading, so validation failed before
any guidance was emitted, and an assertion that no placeholder appeared was
satisfied by an error message containing no guidance at all. The fixture now
carries all three `DiscoverySections`, and both assertions require the guidance
line to be present.

| ID | EARS statement |
| --- | --- |
| R-3.1 | WHEN the `discover validate` guidance would print a placeholder for which inference resolves uniquely, THE SYSTEM SHALL print the resolved value. |
| R-3.2 | IF inference does not resolve uniquely, THE SYSTEM SHALL print the placeholder exactly as today. |
| R-3.3 | THE SYSTEM SHALL apply R-3.1 to both the rendered text (`internal/product/sections.go:98-101`) and the `FieldNext` value (`internal/product/discover.go:157`). |
| R-3.4 | THE SYSTEM SHALL call the Unit 2 resolver and SHALL NOT introduce a second inference implementation. |
| R-3.5 | THE SYSTEM SHALL NOT call from `internal/` into package `main` to reach a resolver. |
| R-3.6 | WHEN `company-os discover validate <id>` runs on a single-platform workspace, THE SYSTEM SHALL emit a `prd new` command containing no `<placeholder>` tokens. |
| R-3.7 | THE SYSTEM SHALL verify its output against all five committed goldens; verification is expected to pass, as no golden pins `discover validate` output. |
| R-3.8 | THE SYSTEM SHALL NOT be implemented before Unit 2 has landed. |

*(The first draft's requirement to substitute into the per-component reality-doc
hint is withdrawn: `sections.go:109-116` already interpolates resolved values.)*

---

## ~~Unit 4: Close the scaffold gap~~ — WITHDRAWN

**Withdrawn 2026-08-28. Premise falsified.** The unit claimed `add team` produces
a workspace its own validator rejects. Verified live: `add team newteam` →
`validate` → gates 1 and 2 clean, exit 0. Gates are absence-tolerant
(`internal/governance/gates.go:47`, `loadOr(path, pyMap{})`), so a missing file
and an empty list are byte-identical, and `examples/acceptance.sh:151-162` has
asserted a clean `init` → `validate` since it was written. Its acceptance
requirement was already true and its "12 → 9 files" metric counted files that
gate nothing. See HLD A1. The unit number is retained, not reused.

---

## Unit 5: Remove `graph` — Group B

**Why:** `derive` and `graph build` dispatch to the same handler with identical
help strings, forcing every reader to discover there is no difference. `graph`
existed only for a Python differential harness that was deleted. Nothing depends
on it but our own test strings.

> ### ⚠ BLOCKING HAZARD — read before implementing
>
> **`examples/acceptance.sh:123` and `:125` invoke `graph build` with
> `>/dev/null 2>&1` and NO exit-code assertion.** They are the two writes inside
> the double-build idempotency check (`acceptance.sh:119-132`). Remove `graph`
> from the dispatch table without touching them and both invocations exit 2,
> write nothing, and `s0 == s1 == s2` becomes **trivially true forever** — the
> suite stays GREEN while the idempotency gate is silently dead, letting every
> future derivation regression through. This is the exact failure class
> `acceptance.sh:112-118` exists to catch, and it is worse than a red build
> because nothing announces it. R-5.4 and R-5.5 are mandatory, not advisory.

| ID | EARS statement |
| --- | --- |
| R-5.1 | THE SYSTEM SHALL remove `graph` from `commandSpecs` and from the dispatch table. |
| R-5.2 | WHEN `company-os --help` runs, THE SYSTEM SHALL list 19 commands and SHALL NOT list `graph`. |
| R-5.3 | WHEN `company-os derive` runs, THE SYSTEM SHALL behave exactly as today. |
| R-5.4 | THE SYSTEM SHALL change `examples/acceptance.sh:123` and `:125` from `graph build` to `derive` **in the same commit** that removes `graph`. |
| R-5.5 | THE SYSTEM SHALL add an exit-code assertion to both invocations, so that a future silent failure of the derivation command cannot leave the double-build check trivially passing. |
| R-5.6 | THE SYSTEM SHALL update every remaining test invoking `graph build`: `cmd/company-os/ansi_test.go:216,235`; `exitcode_test.go:67`; `json_test.go:125`; `args_test.go:20,159,163`; `tuiadvise_test.go:95`. |
| R-5.7 | THE SYSTEM SHALL update the tests pinning `commandNames()`-derived strings at `cmd/company-os/args_test.go:266`, `:312` and `:319`. |
| R-5.8 | THE SYSTEM SHALL NOT change the behavior of `derive`. |
| R-5.9 | AFTER removal, THE SYSTEM SHALL verify that the double-build check still detects a deliberately introduced non-idempotent derivation, and SHALL NOT rely on a green suite as evidence that it does. |
| R-5.10 | THE SYSTEM SHALL record in the commit that `graph build` is removed, so an existing script invoking it fails loudly rather than silently. |

*(The first draft kept `graph` dispatching for a differential harness. That
harness does not exist — `company-os-starter/Makefile:5-6` records its retirement
with R-9.3. See HLD A3.)*

---

## Unit 6: Document the PRD contract split — Group C

**Why:** gate 3 enforces four PRD fields and `prd validate` enforces six, so
`validate` can pass a file `prd validate` fails. The reason is real — one is the
active-record floor, the other an author-time pre-flight — and recorded nowhere,
so it reads as a bug. Trust in the validator is the method; a user who sees the
tool contradict itself stops believing either result. Documenting it also
forecloses the natural "fix" of widening gate 3, which would change goldens and
break workspaces in the field.

**✅ LANDED 2026-08-28.** `check.go:32-36` was verified complete and left
untouched, so R-6.1 discharged as a no-op — the existing comment already states
the sweep-vs-author-time distinction. The unit's real content was R-6.2: a note
added to both `TUTORIAL.md` copies immediately after the `[FAIL] process
contract field 'decisionOwner'` example, where a reader actually meets the
mismatch. `make check` exit 0, five goldens byte-identical.

**Scope correction 2026-08-28:** `internal/product/check.go:32-36` **already
carries the rationale** — "deliberately SHORTER than `prd validate`'s six: the
gate is a workspace sweep and does not demand a decisionOwner or a platform of a
document whose location already states it." The code half of this unit is
essentially done; what is missing is the user-facing half, since a user hits the
mismatch at the CLI and never reads `check.go`. R-6.1 is therefore a
verify-and-top-up, and the unit's real content is R-6.2.

| ID | EARS statement |
| --- | --- |
| R-6.1 | THE SYSTEM SHALL verify the existing rationale at `internal/product/check.go:32-36` is complete and extend it only if it omits the active-record-vs-pre-flight framing. |
| R-6.2 | THE SYSTEM SHALL document the same distinction in `TUTORIAL.md`. |
| R-6.3 | THE SYSTEM SHALL describe gate 3 as the active-record floor and `prd validate` as the author-time pre-flight. |
| R-6.4 | THE SYSTEM SHALL NOT change the fields gate 3 enforces. |
| R-6.5 | THE SYSTEM SHALL NOT change the fields `prd validate` enforces. |
| R-6.6 | THE SYSTEM SHALL NOT change any gate output. |

---

## Unit 7: Aggregate empty-section warnings — Group A

**Why:** walking the tutorial, a user meets six near-identical warning paragraphs
before completing a single artifact — the tool complaining about work they have
not had a chance to do. Worse than the noise is what it hides: routine warnings
on the happy path train users to skim `[warn]`, so one that matters gets ignored.

**✅ LANDED 2026-08-28.** `make check` exit 0, five goldens byte-identical.

```
  [warn] 3 sections are empty: Problem signal, Hypothesis, Success criteria — format guidance only; the team may use its own structure (opt in via standards/doc-formats.yaml)
```

Three paragraphs → one line; the opt-in pointer appears once. Aggregation is in
`applyFormatPolicy` (`internal/product/contract.go`), rendering in `Message`.
Four tests added covering all three paths.

**Amendment G — one code, not a new one.** The aggregated finding reuses
`model.CodeSectionEmpty` with a plural `sections` field rather than introducing
a code. This follows the rule the code's own declaration states
(`internal/model/codes.go:521-526`): a single code renders several sentences off
its fields "rather than splitting one check in two". It also avoids touching
`internal/render/product.go`'s severity case list.

**Amendment H — a stale transcript found and fixed, beyond R-7.7's scope.**
R-7.7 named the two three-warning blocks. Both tutorials ALSO showed, in §2:

```
$ company-os discover validate 2026-per-channel-quiet-hours --team customer-engagement
  [FAIL] section 'Problem signal' is empty
  [FAIL] section 'Hypothesis' is empty
  [FAIL] section 'Success criteria' is empty
```

introduced by the sentence "Try validating the empty brief — **the contract
pushes back**". That is the *enforced* rendering, and
`examples/workspace/teams/customer-engagement/` has no `doc-formats.yaml` — so
the team is not enforced and never was in this fixture. Verified live: an
untouched brief **validates**, exit 0, with one warning. The tutorial was
teaching that empty sections block, which is the opposite of the
guidance-tier design it elsewhere sells as the product's core principle.

Both copies now show the real output and explain the opt-in. Caught only because
U7 changed the neighbouring lines; recorded here because it is a behavior claim,
not a formatting one, and R-0.14 requires transcripts to be accurate rather than
merely mutually consistent.

| ID | EARS statement |
| --- | --- |
| R-7.1 | WHEN more than one section of a document is empty and the team has not opted into enforcement, THE SYSTEM SHALL emit one finding naming every empty section. |
| R-7.2 | THE SYSTEM SHALL state the `doc-formats.yaml` opt-in pointer once per document. |
| R-7.3 | THE SYSTEM SHALL keep these findings at warning severity. |
| R-7.4 | THE SYSTEM SHALL NOT allow an empty-section finding to block, unless the team has opted in. |
| R-7.5 | WHERE a team has opted in via `doc-formats.yaml`, THE SYSTEM SHALL keep emitting one finding per empty section. |
| R-7.6 | WHEN exactly one section is empty, THE SYSTEM SHALL emit the finding as today. |
| R-7.7 | THE SYSTEM SHALL update the captured three-warning transcripts in both copies of `TUTORIAL.md`. |
| R-7.8 | WHEN a freshly scaffolded discovery brief is validated, THE SYSTEM SHALL emit one empty-section warning line, not three. |
| R-7.9 | THE SYSTEM SHALL verify against all five committed goldens; verification is expected to pass, as `CodeSectionEmpty` never reaches `validate`. |

---

## Unit 8: Record the OKF conformance boundary — Group C

**Why:** the belief that this complexity is mandated by the standard is what
protects it from review. The real floor is `type:` plus key preservation, and
establishing that took a full source audit that would otherwise be repeated.
Written down, every gate becomes a local decision arguable on its merits. It is
also the prerequisite for any conceptual-reduction work the validation protocol
may point to.

**✅ LANDED 2026-08-28** at `company-os-starter/docs/CONFORMANCE.md`. Delivered
as an explicitly *partial* discharge of the parent change's R-1.1: the
conformance clause, MUST-NOT-reject list and considered-and-deferred sections
are complete with per-claim citations; Terminology and Versioning are stubs
naming their owning parent clauses, and §7 lists the six parent clauses still
outstanding. Every claim was verified against source before writing.

Two findings worth surfacing, both recorded in the document:
- The MUST-NOT-reject list's items 4 and 5 together mean **the four-root
  federation is a convention of the scaffolder, not a requirement of the
  validator** — `IsRoot()` accepts any single canonical root
  (`internal/workspace/workspace.go:149-159`).
- `internal/scaffold/scaffold.go:196-197` writes `profile: standard` into every
  new platform against an enum defined nowhere. Recorded, not fixed (N7).

| ID | EARS statement |
| --- | --- |
| R-8.1 | THE SYSTEM SHALL provide `company-os-starter/docs/CONFORMANCE.md`. |
| R-8.2 | THE SYSTEM SHALL state that OKF requires `type:` present and non-empty, and cite the enforcing code. |
| R-8.3 | THE SYSTEM SHALL state that OKF requires preservation of unknown keys and tolerance of unknown types, and cite the enforcing code and test. |
| R-8.4 | THE SYSTEM SHALL state that OKF requires tolerance of broken cross-links and a markdown-plus-YAML substrate. |
| R-8.5 | THE SYSTEM SHALL state that `title`, `description` and `resource` are recommended by OKF, not required. |
| R-8.6 | THE SYSTEM SHALL state that `index.md` is a reserved name OKF permits but does not require to be generated. |
| R-8.7 | THE SYSTEM SHALL enumerate as Company-OS additions: `id:` as a required field, the gates and exit-1 posture, lifecycle-typed vocabulary, derived tags, the canonical URI registry, tiering with deviations and exceptions, federation slice hashing, and mandatory generation with drift-as-failure. |
| R-8.8 | THE SYSTEM SHALL cite a source location for every claim about implemented behavior; this is a review criterion, not an automated test. |
| R-8.9 | THE SYSTEM SHALL record that `docs/00-original-proposal.md` claims OKF v0.1 while the EARS targets v0.2. |
| R-8.10 | THE SYSTEM SHALL record that `companyOsVersion: "2026.2"` appears in fixtures against a version defined nowhere. |
| R-8.11 | THE SYSTEM SHALL record that `resource:` appears in no example fixture. |
| R-8.12 | THE SYSTEM SHALL record that no single interop-contract claimant exists. |
| R-8.13 | THE SYSTEM SHALL NOT fix the conditions recorded in R-8.9–R-8.12. |
| R-8.14 | THE SYSTEM SHALL NOT change any conformance behavior. |

---

## Unit 9: Doc-truth fixes — Group A

**Why:** two documented claims are false, and both mislead the audience least
able to detect the error. A user comparing real output to a tutorial that
disagrees with it concludes the tool is broken — the same "reads as failure"
problem Unit 7 addresses, at near-zero cost. Both live in files Unit 1 already
opens.

**✅ LANDED 2026-08-28.** All nine files corrected; `grep` for `[1/7]`/`[2/7]`/
`[7/7]` now returns nothing outside the spec documents themselves. The three
captured transcripts gained the missing `[8/8] promotion integrity` line and
`01-first-day-with-company-os.md:252` was corrected from "Seven gates today" to
"Eight gates today … when a ninth appears". The three `EXAMPLE_README.md` files
were confirmed frontmatter-free before editing (R-9.5) and `make check` passed
afterwards, including the `examples/` checksum and both double-build fixtures.

**Scope, established by exhaustive grep 2026-08-28** — nine files, not the four
first specified. Both earlier scopings were wrong; `handle-a-deviation-or-exception.md`
was missed by every reviewer, and `CLAUDE.md` — the project's own agent
instruction file — states a false gate count to every session that reads it.

| File | Stale refs |
|---|---|
| `CLAUDE.md` | `:80` `[1/7]`, `:86` `[2/7]` |
| `company-os-starter/docs/user-guide/how-to/run-the-validation-gate.md` | `:38`, `:40`, `:45` (transcript) |
| `…/how-to/handle-a-deviation-or-exception.md` | `:27`, `:52`, `:77`, `:78` |
| `…/how-to/grow-a-workspace.md` | `:50`, `:86` |
| `…/how-to/use-the-agent-skills.md` | `:146` (transcript) |
| `…/tutorials/01-first-day-with-company-os.md` | `:241`, `:243`, `:248` (transcript) |
| `examples/workspace/EXAMPLE_README.md` | `:30` |
| `examples/banking/bank/EXAMPLE_README.md` | `:20` |
| `examples/failing-workspace/EXAMPLE_README.md` | `:5` |

All three fixtures emit `[1/8]`…`[8/8]`, verified live and against
`examples/golden-validate.txt` and `examples/failing-workspace-golden-validate.txt`.
The denominator moves to 9 only when a federation manifest is present, so a
uniform 7→8 correction is right for every site above.

| ID | EARS statement |
| --- | --- |
| R-9.1 | THE SYSTEM SHALL remove `CLAUDE.md` context nodes from the list of sources `find` fronts, at `TUTORIAL.md:472` and `company-os-starter/docs/TUTORIAL.md:475`, because no code path in `internal/find/find.go` reads a `CLAUDE.md` (`searchIndexes` matches base name `index.md` only, `find.go:338`). |
| R-9.2 | THE SYSTEM SHALL NOT alter `TUTORIAL.md:388` or `company-os-starter/docs/TUTORIAL.md:391`, which name CLAUDE.md context nodes as artifacts `validate --fix` REGENERATES — an accurate statement about a different mechanism. |
| R-9.3 | THE SYSTEM SHALL correct every stale gate reference listed in the scope table above to `[1/8]`…`[8/8]`. |
| R-9.4 | WHERE a site is a captured transcript, THE SYSTEM SHALL regenerate it by running the command against the workspace shape that file describes, rather than hand-editing the digits, so that gate titles are corrected too (gate 5 gained a directory-index check). |
| R-9.5 | THE SYSTEM SHALL verify that the three `EXAMPLE_README.md` files carry no YAML frontmatter before editing them, because they sit inside `examples/`, which `examples/acceptance.sh:120` checksums, and `skipNames` (`internal/graph/tags.go:52-53`) contains `README.md` but NOT `EXAMPLE_README.md`. |
| R-9.6 | THE SYSTEM SHALL NOT alter any description of behavior that is accurate. |
| R-9.7 | THE SYSTEM SHALL treat `docs/tasks/ux-simplification.md:230` as superseded: it locates the stale gate transcript in `TUTORIAL.md` §8, where it does not appear. |

---

## Landing order

`U8 → U6 → U9 → U1(docs) → U5 → U2 → U3 → U1(output lines) → U7`

U2 → U3 is a hard dependency (U3 needs U2's `internal/` resolver). U5 precedes
U2 because both mutate the `commandSpecs` literal and U5 is the smaller diff.
U6, U8 and U9 are independent of everything, including the validation protocol's
outcome.
