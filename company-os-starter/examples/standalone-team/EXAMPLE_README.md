# Example: `standalone-team/` — a single team adopting Team OS alone

**What it demonstrates:** absence tolerance. One team (`teams/solo/`) uses the
methodology with **no** `company-os/`, **no** `platforms/`, **no**
`company-ontology/` present. Every command degrades gracefully: missing layers
are simply skipped (e.g. `skills list` discovers zero company/platform skills
and still renders a valid merged view; `validate` checks only what exists).

**Why it matters:** this is the entry ramp. A team can start here — ownership
claims, DoR/DoD, scratchpad, personal rules — and later join a federation
without restructuring: the `teams/<t>/` layout is identical inside a full
workspace. Compare with `../banking/small-company/` (one step up: a whole
small company in one repo) and `../banking/bank/` (full federation).

**Contents:** `teams/solo/team.yaml` (with the standard `agentSkills`
precedence contract), `onboarding/developer.md`, `CLAUDE.md`.

**Try it:**
```bash
export PATH="$PWD/../../bin:$PATH"
company-os validate        # passes with the reduced check set
company-os skills list     # layers shown as <none> except team/personal
```
