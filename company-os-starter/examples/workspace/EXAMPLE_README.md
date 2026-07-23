# Example: `workspace/` — the fully worked monorepo workspace

**What it demonstrates:** the complete Company OS / Platform OS / Team OS layout
in a single repository — the canonical reference every other example derives its
file shapes from. This is also the acceptance path: after changing CLI behavior,
run the lifecycle here and confirm `company-os validate` still exits 0
(see `docs/TUTORIAL.md` for the end-to-end walkthrough with real output).

**The cast:**
- Platform: `platforms/communications/` — component
  `customer-notification-service`, requirements with EARS-style clauses
  (`delivery-reliability` mandatory v2.1, `message-schema` mandatory,
  `prd-structure` default), reality doc, platform-layer skill
  (`skills/creating-prd.SKILL.md`).
- Team: `teams/customer-engagement/` — ownership registry, an approved
  deviation (`prd-structure`), a declared deviation (`estimation/story-points`),
  an expiring exception (`message-schema` on `legacy-fax-gateway`), DoR/DoD,
  generated `effective-governance.yaml`, and a git-ignored scratchpad with a
  personal rule (`maria-prd-style.md`).
- Company: `company-os/standards/company-baseline.yaml` (mandatory / default
  tiers) + onboarding.
- Ontology: `company-ontology/` — ID registry, bounded context with
  `ubiquitousLanguage`/`forbiddenTerms`, context map, concepts, taxonomy.
- A completed change: `platforms/communications/archive/prds/2026-per-channel-quiet-hours/`
  (PRD + outcome) with its originating discovery brief under the team.

**Try it:**
```bash
export PATH="$PWD/../../bin:$PATH"   # from this directory
company-os validate                  # the [1/7]..[7/7] gate, exits 0
company-os governance resolve --team customer-engagement
company-os today --role product-owner
company-os skills list
```
