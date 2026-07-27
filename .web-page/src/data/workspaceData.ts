import { WorkspaceNode } from '../types';

export const COMPANY_OS_WORKSPACE_TREE: WorkspaceNode = {
  id: 'root',
  name: 'moonbeam-os (Workspace Root)',
  type: 'folder',
  path: '/',
  layer: 'root',
  description: 'Git-based Company OS workspace containing all governance, platform, team, ontology, and synced knowledge slices.',
  writtenBy: 'Human engineers & company-os CLI',
  children: [
    {
      id: 'workspace-yaml',
      name: 'workspace.yaml',
      type: 'file',
      path: '/workspace.yaml',
      layer: 'root',
      description: 'Federation manifest declaring pinned external repos and destination slice directories under canonical roots.',
      writtenBy: 'Human maintainer',
      validatorCheck: 'Gate [8/N]: Rejects overlapping slice targets or roots outside canonical folders.',
      standaloneIncluded: false,
      content: `version: 1
repos:
  - name: component-library
    url: https://github.com/acme/component-library.git
    pin: {tag: v1.2.0}
    slices:
      - {paths: [docs/sdd], localDirectory: knowledge/components/component-library}
      - {paths: [architecture], localDirectory: knowledge/architecture/component-library}`
    },
    {
      id: 'company-os-folder',
      name: 'company-os/',
      type: 'folder',
      path: '/company-os',
      layer: 'company-os',
      description: 'Company-wide baseline standards, mandatory security/compliance policies, and default engineering guidelines.',
      writtenBy: 'Governance / VP Engineering',
      validatorCheck: 'Gate [4/N]: Frontmatter validation on company standards.',
      standaloneIncluded: false,
      children: [
        {
          id: 'company-standards',
          name: 'standards/',
          type: 'folder',
          path: '/company-os/standards',
          layer: 'company-os',
          description: 'Enterprise standards categorized by domain.',
          writtenBy: 'Architecture Board',
          standaloneIncluded: false,
          children: [
            {
              id: 'story-points-std',
              name: 'estimation-story-points.yaml',
              type: 'file',
              path: '/company-os/standards/estimation-story-points.yaml',
              layer: 'company-os',
              description: 'Company default rule for agile planning using Fibonacci story points.',
              writtenBy: 'Agile Center of Excellence',
              validatorCheck: 'Gate [2/N]: Checks for declared deviations in team governance.',
              standaloneIncluded: false,
              content: `id: company-standard://estimation/story-points
title: Agile Story Point Estimation
tier: default
version: 1.0
summary: Teams should estimate effort using Fibonacci story points for sprint planning.`
            }
          ]
        }
      ]
    },
    {
      id: 'platforms-folder',
      name: 'platforms/',
      type: 'folder',
      path: '/platforms',
      layer: 'platforms',
      description: 'Platform directories containing platform governance, components, reality docs, and PRDs.',
      writtenBy: 'Platform Lead / Product Owners',
      standaloneIncluded: false,
      children: [
        {
          id: 'ordering-platform',
          name: 'ordering/',
          type: 'folder',
          path: '/platforms/ordering',
          layer: 'platforms',
          description: 'Online ordering platform.',
          writtenBy: 'Ordering Platform Lead',
          standaloneIncluded: false,
          children: [
            {
              id: 'ordering-components-folder',
              name: 'components/',
              type: 'folder',
              path: '/platforms/ordering/components',
              layer: 'platforms',
              description: 'Component descriptor files defining single source of truth for component ownership.',
              writtenBy: 'company-os add component',
              standaloneIncluded: false,
              children: [
                {
                  id: 'online-ordering-app-yaml',
                  name: 'online-ordering-app.yaml',
                  type: 'file',
                  path: '/platforms/ordering/components/online-ordering-app.yaml',
                  layer: 'platforms',
                  description: 'Single source of truth for online-ordering-app ownership and platform relationship.',
                  writtenBy: 'company-os add component online-ordering-app --platform ordering',
                  validatorCheck: 'Gate [1/N]: Must match ownership.accountableTeam in teams/web/ownership/components.yaml.',
                  standaloneIncluded: false,
                  content: `id: component://ordering/online-ordering-app
name: Online Ordering Web App
platform: ordering
ownership:
  accountableTeam: web
  techLead: dev-lead@moonbeam.bakery
componentType: web-application`
                }
              ]
            },
            {
              id: 'ordering-reality-folder',
              name: 'reality/',
              type: 'folder',
              path: '/platforms/ordering/reality',
              layer: 'platforms',
              description: 'Representation of Reality: true, current behavior of platform components.',
              writtenBy: 'company-os reality new & Engineers',
              standaloneIncluded: false,
              children: [
                {
                  id: 'online-ordering-app-reality-md',
                  name: 'components/online-ordering-app.md',
                  type: 'file',
                  path: '/platforms/ordering/reality/components/online-ordering-app.md',
                  layer: 'platforms',
                  description: 'Living documentation of online-ordering-app current state.',
                  writtenBy: 'Engineers on completion of PRDs',
                  validatorCheck: 'Precondition for prd complete: updated: date MUST be newer than PRD created: date.',
                  standaloneIncluded: false,
                  content: `---
type: reality-doc
component: component://ordering/online-ordering-app
updated: 2026-07-26
version: 2.1.0
---

# Representation of Reality: Online Ordering Web App

## Current Capabilities
- Guest and authenticated checkout
- Same-day pickup slot selection at checkout with real-time capacity checks
- Integration with payment gateway (Stripe) and bakery POS bridge`
                }
              ]
            },
            {
              id: 'ordering-change-records',
              name: 'change-records/active/',
              type: 'folder',
              path: '/platforms/ordering/change-records/active',
              layer: 'platforms',
              description: 'Active PRDs proposing changes to platform reality.',
              writtenBy: 'company-os prd new',
              standaloneIncluded: false,
              children: [
                {
                  id: 'prd-sample-md',
                  name: '2026-same-day-pickup-slots/prd.md',
                  type: 'file',
                  path: '/platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md',
                  layer: 'platforms',
                  description: 'Active PRD with governance snapshot and checklist.',
                  writtenBy: 'company-os prd new --from-discovery ...',
                  validatorCheck: 'Gate [3/N]: Requires title, team, components, governanceSnapshot in frontmatter.',
                  standaloneIncluded: false,
                  content: `---
type: prd
id: prd://ordering/2026-same-day-pickup-slots
title: Same-day Pickup Slots
team: web
components:
  - component://ordering/online-ordering-app
decisionOwner: product-owner@moonbeam.bakery
created: 2026-07-01
governanceSnapshot:
  snapshotDate: 2026-07-01
  rules:
    - id: platform-standard://ordering/order-confirmation-sla
      tier: mandatory
---

# PRD: Same-day Pickup Slots

## Proposed Change
Allow web customers to choose 15-minute pickup slots on the current day.

## Governance Checklist
- [x] ordering: order-confirmation-sla v1.0 — evidence: https://github.com/moonbeam/pos/pull/42`
                }
              ]
            }
          ]
        }
      ]
    },
    {
      id: 'teams-folder',
      name: 'teams/',
      type: 'folder',
      path: '/teams',
      layer: 'teams',
      description: 'Team directories containing ownership, discovery briefs, standards, governance deviations/exceptions, and generated views.',
      writtenBy: 'Team Engineers & Product Owners',
      standaloneIncluded: true,
      children: [
        {
          id: 'web-team',
          name: 'web/',
          type: 'folder',
          path: '/teams/web',
          layer: 'teams',
          description: 'Web development team folder.',
          writtenBy: 'Web Team',
          standaloneIncluded: true,
          children: [
            {
              id: 'team-yaml',
              name: 'team.yaml',
              type: 'file',
              path: '/teams/web/team.yaml',
              layer: 'teams',
              description: 'Team declaration including agentSkills precedence rules.',
              writtenBy: 'Team Lead',
              standaloneIncluded: true,
              content: `id: team://web
name: Web Team
agentSkills:
  canonicalPath: skills/
  personalPath: scratchpad/personal-rules/
  precedence: canonical-mandatory > personal > canonical-default > canonical-guidance
  onConflict: prefer-canonical-and-inform-user`
            },
            {
              id: 'team-discovery',
              name: 'product/discovery/',
              type: 'folder',
              path: '/teams/web/product/discovery',
              layer: 'teams',
              description: 'Team-private discovery briefs before PRD promotion.',
              writtenBy: 'company-os discover new',
              standaloneIncluded: true,
              children: [
                {
                  id: 'brief-md',
                  name: '2026-same-day-pickup-slots/brief.md',
                  type: 'file',
                  path: '/teams/web/product/discovery/2026-same-day-pickup-slots/brief.md',
                  layer: 'teams',
                  description: 'Validated team discovery brief.',
                  writtenBy: 'Product Manager / Engineers',
                  validatorCheck: 'company-os discover validate ensures Problem signal, Hypothesis, and Success criteria are filled.',
                  standaloneIncluded: true,
                  content: `---
type: discovery-brief
id: discovery://web/2026-same-day-pickup-slots
status: validated
created: 2026-06-15
---

# Discovery: Same-day Pickup Slots

## Problem signal
Customers abandon carts when pickup is limited to next-day.

## Hypothesis
Offering same-day 15-min slots will increase daily online sales by 18%.

## Success criteria
- Cart checkout conversion increases by >15%
- Zero POS slot double-booking incidents`
                }
              ]
            },
            {
              id: 'team-governance',
              name: 'governance/',
              type: 'folder',
              path: '/teams/web/governance',
              layer: 'teams',
              description: 'Team deviations from default rules and expiring exceptions for mandatory rules.',
              writtenBy: 'company-os deviation / exception',
              standaloneIncluded: true,
              children: [
                {
                  id: 'deviations-yaml',
                  name: 'deviations.yaml',
                  type: 'file',
                  path: '/teams/web/governance/deviations.yaml',
                  layer: 'teams',
                  description: 'Comply-or-explain deviations from default rules.',
                  writtenBy: 'company-os deviation declare',
                  validatorCheck: 'Gate [2/N]: Fails if reviewDate is in the past.',
                  standaloneIncluded: true,
                  content: `deviations:
  - rule: company-standard://estimation/story-points
    declaredDate: 2026-07-26
    reviewDate: 2027-01-22
    rationale: Team forecasts with cycle time instead of Fibonacci points.`
                }
              ]
            },
            {
              id: 'team-generated',
              name: 'generated/',
              type: 'folder',
              path: '/teams/web/generated',
              layer: 'teams',
              description: 'Derived governance files re-generated by CLI.',
              writtenBy: 'company-os governance resolve',
              validatorCheck: 'Verified in CI via git diff --exit-code teams/*/generated/.',
              standaloneIncluded: true,
              children: [
                {
                  id: 'eff-gov-yaml',
                  name: 'effective-governance.yaml',
                  type: 'file',
                  path: '/teams/web/generated/effective-governance.yaml',
                  layer: 'teams',
                  description: 'Merged view of baseline company standards, platform rules, and team deviations.',
                  writtenBy: 'company-os governance resolve --team web',
                  standaloneIncluded: true,
                  content: `team: web
resolvedDate: 2026-07-26
components:
  component://ordering/online-ordering-app:
    rules:
      - id: platform-standard://ordering/order-confirmation-sla
        tier: mandatory
      - id: company-standard://estimation/story-points
        tier: default
        status: deviated`
                }
              ]
            }
          ]
        }
      ]
    },
    {
      id: 'ontology-folder',
      name: 'company-ontology/',
      type: 'folder',
      path: '/company-ontology',
      layer: 'company-ontology',
      description: 'Central registry of canonical URIs (teams, platforms, components, requirements, contexts).',
      writtenBy: 'company-os init / add',
      standaloneIncluded: false,
      children: [
        {
          id: 'registry-yaml',
          name: 'ids/registry.yaml',
          type: 'file',
          path: '/company-ontology/ids/registry.yaml',
          layer: 'company-ontology',
          description: 'Canonical index of all workspace IDs.',
          writtenBy: 'CLI auto-update',
          validatorCheck: 'Gate [1/N] & Gate [4/N] query registry for ID resolution.',
          standaloneIncluded: false,
          content: `registry:
  - id: team://web
    kind: team
  - id: platform://ordering
    kind: platform
  - id: component://ordering/online-ordering-app
    kind: component`
        }
      ]
    },
    {
      id: 'knowledge-folder',
      name: 'knowledge/',
      type: 'folder',
      path: '/knowledge',
      layer: 'knowledge',
      description: 'Read-only synced knowledge catalog (synced docs from external component repos without source code).',
      writtenBy: 'company-os workspace sync',
      validatorCheck: 'Gate [8/N]: Hash integrity check. Indexed by Local Search, NOT governed by gates 1-7.',
      standaloneIncluded: false,
      children: [
        {
          id: 'knowledge-claude-md',
          name: 'CLAUDE.md',
          type: 'file',
          path: '/knowledge/CLAUDE.md',
          layer: 'knowledge',
          description: 'Auto-generated context index for AI agents reading the knowledge catalog.',
          writtenBy: 'company-os graph build',
          standaloneIncluded: false,
          content: `# Knowledge Catalog Context Node
- components/component-library (synced v1.2.0)
- architecture/component-library (synced v1.2.0)`
        }
      ]
    }
  ]
};
