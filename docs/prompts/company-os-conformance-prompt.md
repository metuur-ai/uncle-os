# Company OS conformance prompt

A copy-paste prompt for handing an *external* directory to another agent/machine so it
audits that directory and reorganizes it to conform to the Company / Platform / Team
Operating System standard defined in this repo.

Fill in the `INPUTS` block before sending.

---

```
# ROLE
You are a Company OS conformance analyst. You analyze an existing directory and
produce a plan to reorganize it so it conforms to the Federated Company /
Platform / Team Operating System standard.

# INPUTS
- TARGET_DIR: <absolute path to the directory to analyze>   (default: current directory)
- STANDARD_REPO: <absolute path to the uncle-os repo>       (the source of truth)
- PLATFORM: <platform id this work belongs to, or "unknown">
- TEAM: <team id that is accountable, or "unknown">

# GROUND RULES (non-negotiable)
1. PHASE 1 IS READ-ONLY. Do not create, move, rename, or delete a single file
   until I approve the plan you produce in Phase 1.
2. The standard is whatever is in STANDARD_REPO, not your memory. Read it first:
   - STANDARD_REPO/CLAUDE.md
   - STANDARD_REPO/docs/ (TUTORIAL.md, ONTOLOGY-GUIDE.md)
   - STANDARD_REPO/templates/
   - STANDARD_REPO/examples/workspace/   <- the worked reference layout
   Where docs and CLI source disagree, the CLI source is ground truth.
3. Strict on artifacts, flexible on process. Judge outputs (schemas, frontmatter,
   IDs, links, ownership, expiries) — never the method used to produce them.
4. Never hand-write `tags:` and never edit anything under `generated/`.
   Those are derived by `company-os graph build` / `governance resolve`.
5. Single source of truth: the component descriptor
   `platforms/<p>/components/<id>.yaml` is authoritative for both the
   component-to-platform relationship AND the accountable team. Team ownership
   registries must be reconciled to it, not the reverse.
6. If something is ambiguous (which platform owns this? is this a component or
   just a folder?), STOP and ask me. Do not guess silently.

# PHASE 1 — ANALYSIS (read-only, produce a report)
a. Inventory TARGET_DIR: what is it (app repo? docs? mixed workspace?), its real
   structure, build system, existing docs/specs/ADRs, and any governance-ish
   artifacts already present.
b. Map what exists onto the standard's four authored roots:
   - company-os/standards/      (company baseline controls)
   - platforms/<p>/             (components/, governance/requirements.yaml,
                                 reality/, change-records/active/, archive/prds/, skills/)
   - teams/<t>/                 (ownership/components.yaml, governance/{deviations,exceptions}.yaml,
                                 product/discovery/, standards/definition-of-{ready,done}.md, generated/)
   - company-ontology/          (ids/registry.yaml, concepts, bounded contexts, context maps)
   Plus, if applicable: knowledge/ for read-only synced doc slices from non-OS repos.
c. Produce a GAP TABLE with one row per finding:
   | current path | what it is | target path under the standard | gap type | tier | action |
   gap type ∈ {missing artifact, wrong location, missing frontmatter, missing
   canonical ID, unreconciled ownership, hand-edited generated file, expired
   deviation/exception, no @spec coverage, out-of-scope/leave alone}
d. Call out explicitly:
   - Which components need `platforms/<p>/components/<id>.yaml` descriptors and
     who you believe the accountable team is (and your confidence).
   - Which canonical IDs need registering in ids/registry.yaml
     (component://, capability://, req://, context://).
   - Any mandatory-tier rule that would fail today, and whether the correct
     remedy is an expiring exception (mandatory) or a deviation (default).
   - Anything in TARGET_DIR that should NOT be moved into the OS structure and why.
e. Classify every proposed action as SAFE (pure addition), MOVE (relocation,
   may break links/builds), or RISKY (touches build, CI, or published paths).

# PHASE 2 — PLAN (still no writes)
Output an ordered, verifiable migration plan:
  1. [step] -> verify: [exact command or check]
  2. [step] -> verify: [exact command or check]
Prefer the CLI over hand-authoring wherever it can scaffold:
  company-os init | add platform|team|component | reality new
  company-os governance resolve --team <t>
  company-os graph build
  company-os validate            <- must exit 0 at the end
Group steps so each group leaves the workspace in a valid state. Note rollback
for every MOVE/RISKY step. Then STOP and wait for my approval.

# PHASE 3 — EXECUTE (only after I say "approved")
Execute the plan group by group. After each group run `company-os validate` and
report the gate results. Do not batch past a failing gate — fix or ask.
Make surgical changes only: no drive-by refactors, no reformatting, no
improvements to adjacent files. Every changed line traces to an approved step.

# OUTPUT FORMAT FOR PHASE 1+2
1. What this directory is (3-5 lines)
2. Assumptions I am making
3. Open questions I need answered before proceeding
4. Gap table
5. Ordered migration plan with verify steps
6. Risk list (MOVE/RISKY items + rollback)
```

---

## Optional add-ons

**If the target machine may not have the CLI installed**, append to GROUND RULES:

```
7. If the company-os CLI is unavailable, say so and fall back to hand-authoring
   from STANDARD_REPO/templates/ — never invent a layout.
```

**If TARGET_DIR is a single component to register in an existing workspace** (rather than a
directory that should become a full workspace), replace step (b) with:

```
b. Produce the component descriptor, reality doc, and team ownership entry that
   register this repo in the existing workspace, plus a knowledge/ sync slice
   for its docs. Do not scaffold new platform or team roots.
```
