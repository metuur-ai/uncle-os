# Example: `standalone-team/` — a real, minimal, one-team workspace

**What it demonstrates:** the smallest workspace that actually validates: one
company baseline, one platform (`core`) with one component (`solo-service`),
and one team (`solo`) that owns it. `company-os validate --root
examples/standalone-team` exits 0 here, exactly as it does on the full
`examples/workspace` fixture — just with one of everything instead of many.

**Progressive disclosure:** nothing in this workspace requires knowing what a
federation is. There is one team and one platform; ownership, governance, and
search all work the same as they would with ten. Federation concepts —
multiple teams, multiple platforms, cross-team deviations — only show up when
a second team or platform is added with `company-os add platform|team`. Until
then, this is just: a team, a thing it owns, and the loop of discovery → PRD →
`prd complete`.

**Contents:** the full minimal shape — `company-os/` (company + baseline
control), `platforms/core/` (platform, governance requirements, the
`solo-service` component descriptor + reality doc), `teams/solo/` (team,
ownership claim, DoR/DoD, onboarding doc), `company-ontology/` (the id
registry), all fully derived (`generated/`, `CLAUDE.md` nodes, tags) so a
double build is a no-op. Compare with `../banking/small-company/` (a whole
small company in one repo) and `../banking/bank/` (full federation) for what
this grows into.

**Try it:**
```bash
export PATH="$PWD/../../bin:$PATH"
company-os --root examples/standalone-team validate                # PASS, 0 findings to fix
company-os --root examples/standalone-team next                    # what to do right now
company-os --root examples/standalone-team today --role developer  # governance owed today
company-os --root examples/standalone-team find solo-service        # everything about it
```
