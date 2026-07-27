import { ValidationGate } from '../types';

export const VALIDATION_GATES_DATA: ValidationGate[] = [
  {
    id: 1,
    name: '[1/N] Ownership Reconciliation',
    shortName: 'Ownership Reconciliation',
    description: 'Ensures that every team claiming accountability for a component in teams/<t>/ownership/components.yaml matches the accountableTeam declared in the component descriptor (platforms/<p>/components/<comp>.yaml).',
    checks: [
      'Component descriptor exists in platforms/<p>/components/<comp>.yaml',
      'ownership.accountableTeam matches team ID',
      'No orphaned components or unregistered team claims'
    ],
    absenceTolerant: false,
    examplePass: `[1/7] ownership reconciliation
  [ok] online-ordering-app: registry and descriptor agree (ordering)`,
    exampleFail: `[1/7] ownership reconciliation
  [FAIL] online-ordering-app: team 'web' claims accountable, but descriptor says 'ordering-platform-lead'`,
    fixAction: 'Edit the component descriptor platforms/<p>/components/<comp>.yaml — it is the single source of truth for component ownership.'
  },
  {
    id: 2,
    name: '[2/N] Deviation and Exception Expiry',
    shortName: 'Deviation/Exception Expiry',
    description: 'Verifies that no team deviation reviewDate and no rule exception expires date is in the past or missing an expiry date.',
    checks: [
      'reviewDate in teams/<t>/governance/deviations.yaml is in the future',
      'expires date in teams/<t>/governance/exceptions.yaml is present and in the future',
      'Deviations only reference default-tier rules'
    ],
    absenceTolerant: true,
    examplePass: `[2/7] deviation and exception expiry
  [ok] 1 active deviation(s) checked, 0 expired`,
    exampleFail: `[2/7] deviation and exception expiry
  [FAIL] teams/web/governance/deviations.yaml: deviation for 'estimation/story-points' expired on 2026-01-19`,
    fixAction: 'Re-declare the deviation with company-os deviation declare to reset review date 180 days out, or remove it if no longer needed.'
  },
  {
    id: 3,
    name: '[3/N] Active PRD Contracts',
    shortName: 'Active PRD Contracts',
    description: 'Validates frontmatter fields on active PRDs under platforms/<p>/change-records/active/<id>/prd.md.',
    checks: [
      'title is present and non-empty',
      'team ID resolves to registered team',
      'components list contains valid component URIs',
      'governanceSnapshot block is stamped with rules and tier definitions'
    ],
    absenceTolerant: true,
    examplePass: `[3/7] active PRD contracts
  [ok] 1 active PRD(s) verified`,
    exampleFail: `[3/7] active PRD contracts
  [FAIL] platforms/ordering/change-records/active/2026-same-day-pickup/prd.md: missing 'governanceSnapshot'`,
    fixAction: 'Re-run company-os prd new --from-discovery to auto-inject governanceSnapshot, or restore missing frontmatter fields.'
  },
  {
    id: 4,
    name: '[4/N] Frontmatter Core & Tag Derivation',
    shortName: 'Frontmatter & Tags',
    description: 'Verifies that all Markdown documents carry core frontmatter (type, id) and that committed tags match tags derived from content.',
    checks: [
      'type and id frontmatter present on governed docs',
      'Committed tags: array matches graph build output exactly',
      'No syntax errors in YAML frontmatter blocks'
    ],
    absenceTolerant: false,
    examplePass: `[4/7] frontmatter core and tag derivation (interop contract)
  [ok] 14 document(s) tag-verified`,
    exampleFail: `[4/7] frontmatter core and tag derivation (interop contract)
  [FAIL] platforms/ordering/reality/components/online-ordering-app.md: derived tags drift (run: company-os graph build)`,
    fixAction: 'Execute company-os graph build to update derived tags, then commit the diff.'
  },
  {
    id: 5,
    name: '[5/N] CLAUDE.md Context Node Drift',
    shortName: 'CLAUDE.md Drift',
    description: 'Checks if generated CLAUDE.md context nodes across platforms, teams, and knowledge match a fresh graph derivation.',
    checks: [
      'CLAUDE.md matches graph build render',
      'Passes if CLAUDE.md is absent (absence-tolerant)'
    ],
    absenceTolerant: true,
    examplePass: `[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
  [ok] CLAUDE.md up to date`,
    exampleFail: `[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
  [FAIL] teams/web/CLAUDE.md drifted from derived state (run: company-os graph build)`,
    fixAction: 'Run company-os graph build to regenerate CLAUDE.md files and commit the diff.'
  },
  {
    id: 6,
    name: '[6/N] Feature-Index Drift',
    shortName: 'Feature-Index Drift',
    description: 'Validates platform feature indexes platforms/<p>/generated/feature-index.yaml against component and artifact mappings.',
    checks: [
      'feature-index.yaml matches graph build output',
      'All referenced component IDs resolve in ontology'
    ],
    absenceTolerant: true,
    examplePass: `[6/7] feature-index drift (derived component->artifact map)
  [ok] feature index up to date`,
    exampleFail: `[6/7] feature-index drift (derived component->artifact map)
  [FAIL] platforms/ordering/generated/feature-index.yaml drifted (run: company-os graph build)`,
    fixAction: 'Execute company-os graph build to re-derive feature indexes.'
  },
  {
    id: 7,
    name: '[7/N] Custom Skills Layering',
    shortName: 'Skills Layering',
    description: 'Validates that custom personal or team skills do not shadow canonical skill IDs/names, and resolves extends: platform-skill:// references.',
    checks: [
      'No name or ID collision between personal/team skills and canonical skills',
      'All extends: URIs point to existing canonical skills'
    ],
    absenceTolerant: true,
    examplePass: `[7/7] custom skills layering (shadowing + extends resolution)
  [ok] 4 skill(s) resolved with zero conflicts`,
    exampleFail: `[7/7] custom skills layering (shadowing + extends resolution)
  [FAIL] teams/web/skills/creating-prd.md shadows canonical platform skill 'creating-prd'`,
    fixAction: 'Rename the custom skill or use extends: platform-skill://ordering/creating-prd instead of re-declaring the same ID.'
  },
  {
    id: 8,
    name: '[8/N] Federated Slice Integrity',
    shortName: 'Federated Slice Integrity',
    description: 'Only runs when workspace.yaml is present. Compares materialized slices on disk with hashes in workspace.lock.yaml.',
    checks: [
      'Content hashes of files under synced slices match workspace.lock.yaml',
      'Target directories match localDirectory in workspace.yaml',
      'No hand-edited files in read-only 0444 synced slices'
    ],
    absenceTolerant: true,
    federatedOnly: true,
    examplePass: `[8/8] federated slice integrity (workspace.lock.yaml)
  [ok] 3 slice(s) hash-verified against lockfile`,
    exampleFail: `[8/8] federated slice integrity (workspace.lock.yaml)
  [FAIL] knowledge/components/component-library/docs/sdd/arch.md: content modified (hand-edit detected in 0444 slice)`,
    fixAction: 'Discard hand-edits and run company-os workspace sync. To make permanent changes, update the upstream repo and release a new tag/pin.'
  }
];
