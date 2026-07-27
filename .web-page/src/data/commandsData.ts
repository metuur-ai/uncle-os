import { CliCommand, ExitCodeInfo } from '../types';

export const EXIT_CODES_DATA: ExitCodeInfo[] = [
  {
    code: 0,
    meaning: 'Success',
    whenOccurs: 'The command executed and completed successfully.',
    recommendedAction: 'Proceed to next step in workflow.',
  },
  {
    code: 1,
    meaning: 'Validation Failed',
    whenOccurs: 'A company-os validate gate reported [FAIL], or discover validate / prd validate refused an artifact.',
    recommendedAction: 'Read the [FAIL] diagnostic lines in stdout and fix the artifact or derived content.',
  },
  {
    code: 2,
    meaning: 'Usage Error',
    whenOccurs: 'Unknown subcommand, bad flag, missing or invalid argument, or running company-os with no subcommand.',
    recommendedAction: 'Check flag syntax with company-os --help or correct command parameters.',
  },
  {
    code: 3,
    meaning: 'Workspace Error',
    whenOccurs: 'Not inside a workspace root, or a platform, team, component, brief, PRD, or manifest repo named does not exist.',
    recommendedAction: 'Verify path or run company-os init, or pass --root /path/to/workspace.',
  },
  {
    code: 4,
    meaning: 'Artifact Error',
    whenOccurs: 'A YAML file or frontmatter block is malformed, or workspace.yaml breaks its schema.',
    recommendedAction: 'Inspect the malformed YAML file for syntax errors or missing schema properties.',
  },
  {
    code: 5,
    meaning: 'Precondition Failed',
    whenOccurs: 'A gate refused action: prd complete before reality doc was updated, or prd new against an unvalidated brief.',
    recommendedAction: 'Perform the prerequisite work (e.g. edit platforms/.../reality/components/app.md and bump updated: date).',
  },
  {
    code: 6,
    meaning: 'External Tool Error',
    whenOccurs: 'Git missing or < 2.27, clone/sparse-checkout failed, or workspace sync --frozen could not reconstruct slice from lock and cache.',
    recommendedAction: 'Check git installation/network or run company-os workspace sync without --frozen first.',
  },
  {
    code: 7,
    meaning: 'Interactive-Mode Error',
    whenOccurs: 'A prompt was required and no terminal is attached (e.g. company-os init in CI without --company/--team/--platform).',
    recommendedAction: 'Pass explicit flags (--company, --team, --platform) when executing in automated or CI environments.',
  },
  {
    code: 8,
    meaning: 'Conflict',
    whenOccurs: 'Destination already exists and command refuses to overwrite (e.g. init in existing root, duplicate add, existing reality doc).',
    recommendedAction: 'Use an existing file/component name or choose a different target directory.',
  },
];

export const CLI_COMMANDS_DATA: CliCommand[] = [
  {
    id: 'init',
    name: 'init',
    syntax: 'company-os init [--company NAME] [--team ID] [--platform ID]',
    category: 'Scaffolding',
    description: 'Scaffold a brand-new workspace containing the four core roots: company-os/, platforms/<p>/, teams/<t>/, and company-ontology/. Refuses to run in an existing workspace.',
    flags: [
      { flag: '--company', type: 'string', required: false, description: 'Company name (defaults to "My Company")', defaultVal: 'My Company' },
      { flag: '--team', type: 'string', required: false, description: 'Initial team ID (defaults to "core")', defaultVal: 'core' },
      { flag: '--platform', type: 'string', required: false, description: 'Initial platform ID (defaults to "platform-1")', defaultVal: 'platform-1' },
    ],
    example: 'company-os init --company "Moonbeam Bakery" --team web --platform ordering',
    expectedOutput: `initialized workspace at /Users/you/moonbeam-os
  company: Moonbeam Bakery | first team: web | first platform: ordering
next: cd /Users/you/moonbeam-os && company-os discover new --team web "<discovery title>"`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "init",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "workspace-init",
      "title": "Initialization",
      "findings": [
        {"severity": "ok", "code": "workspace.initialized", "message": "initialized workspace at /Users/you/moonbeam-os", "fields": {"company": "Moonbeam Bakery", "team": "web", "platform": "ordering"}}
      ]
    }
  ],
  "guidance": ["cd /Users/you/moonbeam-os && company-os discover new --team web \\"<discovery title>\\""],
  "error": null
}`,
    exitCodesPossible: [0, 2, 7, 8],
  },
  {
    id: 'add',
    name: 'add',
    syntax: 'company-os add {platform|team|component} <name> [--platform ID]',
    category: 'Scaffolding',
    description: 'Grow an existing workspace by adding a new platform, team, or component. Adding a component requires --platform.',
    flags: [
      { flag: '--platform', type: 'string', required: false, description: 'Platform ID (REQUIRED when kind is component)' },
    ],
    example: 'company-os add component online-ordering-app --platform ordering',
    expectedOutput: `added component 'online-ordering-app' to platform 'ordering'
next: company-os reality new --platform ordering online-ordering-app`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "add",
  "action": "component",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "add-component",
      "title": "Component Registry",
      "findings": [
        {"severity": "ok", "code": "component.added", "message": "added component 'online-ordering-app' to platform 'ordering'", "fields": {"component": "online-ordering-app", "platform": "ordering"}}
      ]
    }
  ],
  "guidance": ["company-os reality new --platform ordering online-ordering-app"],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3, 8],
  },
  {
    id: 'reality',
    name: 'reality new',
    syntax: 'company-os reality new <component> --platform ID',
    category: 'Scaffolding',
    description: 'Scaffold a component\'s reality doc describing true current behavior. Refuses to overwrite existing files.',
    flags: [
      { flag: '--platform', type: 'string', required: true, description: 'Platform ID owning the component' },
    ],
    example: 'company-os reality new online-ordering-app --platform ordering',
    expectedOutput: `scaffolded platforms/ordering/reality/components/online-ordering-app.md
next: edit representation of reality and set updated: date`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "reality",
  "action": "new",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "reality-new",
      "title": "Representation of Reality",
      "findings": [
        {"severity": "ok", "code": "reality.scaffolded", "path": "platforms/ordering/reality/components/online-ordering-app.md", "message": "scaffolded reality document"}
      ]
    }
  ],
  "guidance": ["edit platforms/ordering/reality/components/online-ordering-app.md"],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3, 8],
  },
  {
    id: 'discover',
    name: 'discover',
    syntax: 'company-os discover {new|validate} --team ID ["<title>" | <brief-id>]',
    category: 'Lifecycle',
    description: 'Manage team-private discovery briefs. `new` creates a brief under teams/<t>/product/discovery/; `validate` checks required sections.',
    flags: [
      { flag: '--team', type: 'string', required: true, description: 'Team ID owning the discovery brief' },
    ],
    example: 'company-os discover validate 2026-same-day-pickup-slots --team web',
    expectedOutput: `  [ok] brief '2026-same-day-pickup-slots' validated (status: validated)`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "discover",
  "action": "validate",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "discover-validate",
      "title": "Discovery Brief Contract",
      "findings": [
        {"severity": "ok", "code": "brief.validated", "path": "teams/web/product/discovery/2026-same-day-pickup-slots/brief.md", "message": "brief '2026-same-day-pickup-slots' validated (status: validated)"}
      ]
    }
  ],
  "guidance": ["company-os prd new --team web --platform ordering --components online-ordering-app --from-discovery 2026-same-day-pickup-slots"],
  "error": null
}`,
    exitCodesPossible: [0, 1, 2, 3, 4],
  },
  {
    id: 'prd',
    name: 'prd',
    syntax: 'company-os prd {new|validate|complete} --platform ID [--team ID] [--components ID] [--from-discovery BRIEF-ID]',
    category: 'Lifecycle',
    description: 'Manage platform-visible PRDs. `new` scaffolds with governance snapshot; `validate` checks fields; `complete` verifies reality doc update before archiving.',
    flags: [
      { flag: '--platform', type: 'string', required: true, description: 'Platform ID' },
      { flag: '--team', type: 'string', required: false, description: 'Team ID' },
      { flag: '--components', type: 'string', required: false, description: 'Comma-separated list of component IDs' },
      { flag: '--from-discovery', type: 'string', required: false, description: 'Source discovery brief ID' },
      { flag: '--force', type: 'boolean', required: false, description: 'Bypass done-check (exceptional cases only)' },
    ],
    example: 'company-os prd complete 2026-same-day-pickup-slots --platform ordering',
    expectedOutput: `archived -> platforms/ordering/archive/prds/2026-same-day-pickup-slots
outcome review scheduled (due in 90 days)
appended platforms/ordering/log.md`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "prd",
  "action": "complete",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "prd-complete",
      "title": "PRD Completion",
      "findings": [
        {"severity": "ok", "code": "prd.archived", "message": "archived -> platforms/ordering/archive/prds/2026-same-day-pickup-slots"},
        {"severity": "ok", "code": "outcome.scheduled", "message": "outcome review scheduled (due in 90 days)"}
      ]
    }
  ],
  "guidance": ["company-os validate"],
  "error": null
}`,
    exitCodesPossible: [0, 1, 2, 3, 4, 5],
  },
  {
    id: 'governance',
    name: 'governance',
    syntax: 'company-os governance {resolve|explain} [--team ID] [component-id]',
    category: 'Governance',
    description: '`resolve` merges company baseline, platform rules, and team deviations into effective-governance.yaml. `explain` details why rules apply to a component.',
    flags: [
      { flag: '--team', type: 'string', required: false, description: 'Team ID (functionally required for resolve)' },
    ],
    example: 'company-os governance resolve --team web',
    expectedOutput: `resolved governance for team 'web' (1 component(s))
wrote teams/web/generated/effective-governance.yaml
  online-ordering-app: platforms [ordering], 3 company + 3 platform requirement(s)`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "governance",
  "action": "resolve",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "governance-resolve",
      "title": "Effective Governance Resolution",
      "findings": [
        {"severity": "ok", "code": "governance.resolved", "path": "teams/web/generated/effective-governance.yaml", "fields": {"team": "web", "componentsCount": 1, "requirementsCount": 6}}
      ]
    }
  ],
  "guidance": ["company-os check ready --team web --components online-ordering-app"],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3, 4],
  },
  {
    id: 'check',
    name: 'check',
    syntax: 'company-os check {ready|done} --team ID --components ID[,ID...]',
    category: 'Governance',
    description: 'Generate on-demand Definition of Ready or Definition of Done combining team baseline standards and resolved component governance checklist.',
    flags: [
      { flag: '--team', type: 'string', required: true, description: 'Team ID' },
      { flag: '--components', type: 'string', required: true, description: 'Comma-separated list of component IDs' },
    ],
    example: 'company-os check ready --team web --components online-ordering-app',
    expectedOutput: `== Team baseline (definition-of-ready.md) ==
- [x] Discovery brief validated
- [x] Architecture review completed
== Applicable governance (online-ordering-app) ==
- [ ] ordering: order-confirmation-sla v1.0 (mandatory) — evidence:`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "check",
  "action": "ready",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "check-ready",
      "title": "Definition of Ready Check",
      "findings": [
        {"severity": "ok", "code": "dor.checklist", "message": "generated DoR checklist for online-ordering-app"}
      ]
    }
  ],
  "guidance": [],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3],
  },
  {
    id: 'validate',
    name: 'validate',
    syntax: 'company-os validate',
    category: 'Validation',
    description: 'Run all workspace validation gates (7 in monorepo, 8 in federated workspace with workspace.yaml). Checks ownership, expiry, PRDs, tags, drift, skills, and federated slices.',
    flags: [],
    example: 'company-os validate',
    expectedOutput: `[1/7] ownership reconciliation
  [ok] online-ordering-app: registry and descriptor agree (ordering)
[2/7] deviation and exception expiry
[3/7] active PRD contracts
[4/7] frontmatter core and tag derivation (interop contract)
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
[6/7] feature-index drift (derived component->artifact map)
[7/7] custom skills layering (shadowing + extends resolution)
PASS`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "validate",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {"ordinal": 1, "slug": "gate-1", "title": "ownership reconciliation", "findings": [{"severity": "ok", "code": "ownership.reconciled"}]},
    {"ordinal": 2, "slug": "gate-2", "title": "deviation and exception expiry", "findings": []},
    {"ordinal": 3, "slug": "gate-3", "title": "active PRD contracts", "findings": []},
    {"ordinal": 4, "slug": "gate-4", "title": "frontmatter core and tag derivation", "findings": []},
    {"ordinal": 5, "slug": "gate-5", "title": "CLAUDE.md context node drift", "findings": []},
    {"ordinal": 6, "slug": "gate-6", "title": "feature-index drift", "findings": []},
    {"ordinal": 7, "slug": "gate-7", "title": "custom skills layering", "findings": []}
  ],
  "guidance": [],
  "error": null
}`,
    exitCodesPossible: [0, 1, 3, 4, 6],
  },
  {
    id: 'deviation',
    name: 'deviation declare',
    syntax: 'company-os deviation declare <rule-uri> --team ID [--rationale TEXT]',
    category: 'Governance',
    description: 'Declare a comply-or-explain deviation from a default-tier rule. Rejects mandatory-tier rules; automatically sets 180-day review date in teams/<t>/governance/deviations.yaml.',
    flags: [
      { flag: '--team', type: 'string', required: true, description: 'Team ID' },
      { flag: '--rationale', type: 'string', required: false, description: 'Explanation for opting out of default rule' },
    ],
    example: 'company-os deviation declare "company-standard://estimation/story-points" --team web --rationale "Team forecasts with cycle time instead of points."',
    expectedOutput: `declared deviation from company-standard://estimation/story-points in teams/web/governance/deviations.yaml
review due 2027-01-19; re-run: company-os governance resolve --team web`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "deviation",
  "action": "declare",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "deviation-declare",
      "title": "Deviation Declaration",
      "findings": [
        {"severity": "ok", "code": "deviation.declared", "path": "teams/web/governance/deviations.yaml", "fields": {"rule": "company-standard://estimation/story-points", "reviewDate": "2027-01-19"}}
      ]
    }
  ],
  "guidance": ["company-os governance resolve --team web"],
  "error": null
}`,
    exitCodesPossible: [0, 1, 2, 3],
  },
  {
    id: 'exception',
    name: 'exception request',
    syntax: 'company-os exception request <rule-uri> --team ID --component ID --expires DATE [--reason TEXT]',
    category: 'Governance',
    description: 'Request a temporary expiring exception to a mandatory-tier rule. Requires approval by rule owner before valid.',
    flags: [
      { flag: '--team', type: 'string', required: true, description: 'Team ID' },
      { flag: '--component', type: 'string', required: true, description: 'Component ID affected' },
      { flag: '--expires', type: 'string', required: true, description: 'Expiry date YYYY-MM-DD (REQUIRED)' },
      { flag: '--reason', type: 'string', required: false, description: 'Technical or organizational reason' },
    ],
    example: 'company-os exception request "platform-standard://ordering/order-confirmation-sla" --team web --component legacy-pos-bridge --expires 2026-12-31 --reason "Legacy POS can\'t emit confirmations synchronously yet."',
    expectedOutput: `exception drafted in teams/web/governance/exceptions.yaml (expires 2026-12-31)
note: mandatory rules require approval by the rule owner before this is valid.`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "exception",
  "action": "request",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "exception-request",
      "title": "Exception Request",
      "findings": [
        {"severity": "ok", "code": "exception.drafted", "path": "teams/web/governance/exceptions.yaml", "fields": {"rule": "platform-standard://ordering/order-confirmation-sla", "expires": "2026-12-31"}}
      ]
    }
  ],
  "guidance": ["company-os governance resolve --team web"],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3],
  },
  {
    id: 'graph',
    name: 'graph build',
    syntax: 'company-os graph build',
    category: 'Utility',
    description: 'Re-derive tags and generated aggregates (feature-index, CLAUDE.md context nodes) from frontmatter across the whole workspace.',
    flags: [],
    example: 'company-os graph build',
    expectedOutput: `graph build: 12 doc(s) scanned, 3 updated`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "graph",
  "action": "build",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "graph-build",
      "title": "Knowledge Graph Synthesis",
      "findings": [
        {"severity": "ok", "code": "graph.built", "fields": {"scanned": 12, "updated": 3}}
      ]
    }
  ],
  "guidance": ["company-os validate"],
  "error": null
}`,
    exitCodesPossible: [0, 3, 4],
  },
  {
    id: 'today',
    name: 'today',
    syntax: 'company-os today [--role ROLE]',
    category: 'Utility',
    description: 'Role-aware daily view summarizing open governance items, active PRDs, or outcome reviews due.',
    flags: [
      { flag: '--role', type: 'string', required: false, description: 'Role (developer, team-lead, product-owner, architect, vp-engineering, director-of-product)', defaultVal: 'developer' },
    ],
    example: 'company-os today --role product-owner',
    expectedOutput: `== Today's Overview (Role: product-owner) ==
- Active PRDs: 1 (2026-same-day-pickup-slots in ordering)
- Outcomes due for review: 0
- Expiring deviations: 0`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "today",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "today-overview",
      "title": "Daily Summary",
      "findings": [
        {"severity": "ok", "code": "today.summary", "fields": {"role": "product-owner", "activePRDs": 1, "outcomesDue": 0}}
      ]
    }
  ],
  "guidance": [],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3],
  },
  {
    id: 'ids',
    name: 'ids list',
    syntax: 'company-os ids list [--team ID] [--platform ID] [--prefix TEXT] [--role ROLE]',
    category: 'Utility',
    description: 'List canonical IDs from company-ontology/ids/registry.yaml with optional role glosssary.',
    flags: [
      { flag: '--prefix', type: 'string', required: false, description: 'Filter by ID prefix (e.g. component://, team://, platform://)' },
      { flag: '--team', type: 'string', required: false, description: 'Filter by team' },
      { flag: '--platform', type: 'string', required: false, description: 'Filter by platform' },
      { flag: '--role', type: 'string', required: false, description: 'Print plain-language glossary for role' },
    ],
    example: 'company-os ids list --prefix component://',
    expectedOutput: `component://ordering/online-ordering-app
component://loyalty/crumb-club-app`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "ids",
  "action": "list",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "ids-list",
      "title": "Ontology ID Registry",
      "findings": [
        {"severity": "ok", "code": "ids.item", "fields": {"id": "component://ordering/online-ordering-app"}},
        {"severity": "ok", "code": "ids.item", "fields": {"id": "component://loyalty/crumb-club-app"}}
      ]
    }
  ],
  "guidance": [],
  "error": null
}`,
    exitCodesPossible: [0, 2, 3],
  },
  {
    id: 'skills',
    name: 'skills list',
    syntax: 'company-os skills list',
    category: 'Utility',
    description: 'List merged agent skills across all 4 layers (company, platform, team, personal scratchpad). Never errors if layers are absent.',
    flags: [],
    example: 'company-os skills list',
    expectedOutput: `== Company Skills ==
  - company-skill://standards/prds
== Platform Skills ==
  - platform-skill://ordering/creating-prd
== Team Skills ==
  - team-skill://web/discovery
== Personal Skills ==
  - personal-skill://scratchpad/testing`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "skills",
  "action": "list",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "skills-list",
      "title": "Agent Skills Layer Hierarchy",
      "findings": [
        {"severity": "ok", "code": "skill.layer", "fields": {"layer": "platform", "id": "platform-skill://ordering/creating-prd"}}
      ]
    }
  ],
  "guidance": [],
  "error": null
}`,
    exitCodesPossible: [0, 3],
  },
  {
    id: 'workspace',
    name: 'workspace',
    syntax: 'company-os workspace {sync|status} [--frozen] [--only NAME]',
    category: 'Federation',
    description: 'Federated multi-repo sync/status. Requires workspace.yaml manifest. `sync --frozen` materializes strictly from workspace.lock.yaml without network in CI.',
    flags: [
      { flag: '--frozen', type: 'string', required: false, description: 'Offline CI mode strictly using workspace.lock.yaml' },
      { flag: '--only', type: 'string', required: false, description: 'Limit sync to specific repository name' },
    ],
    example: 'company-os workspace sync',
    expectedOutput: `synced 3 repo(s) into workspace
updated workspace.lock.yaml
next: company-os graph build && company-os validate`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "workspace",
  "action": "sync",
  "root": "/Users/you/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "workspace-sync",
      "title": "Federated Slices Synchronization",
      "findings": [
        {"severity": "ok", "code": "federation.synced", "path": "workspace.lock.yaml", "fields": {"syncedCount": 3}}
      ]
    }
  ],
  "guidance": ["company-os graph build && company-os validate"],
  "error": null
}`,
    exitCodesPossible: [0, 1, 2, 3, 4, 6],
  },
  {
    id: 'scratchpad',
    name: 'scratchpad init',
    syntax: 'company-os scratchpad init [--repo PATH]',
    category: 'Utility',
    description: 'Initialize local-only, git-ignored scratchpad in any repo. Exempt from workspace root requirement.',
    flags: [
      { flag: '--repo', type: 'string', required: false, description: 'Path to target repository (defaults to current dir)' },
    ],
    example: 'company-os scratchpad init --repo teams/web',
    expectedOutput: `scaffolded scratchpad/{drafts,brainstorms,personal-rules,experiments,inbox}/
updated .gitignore`,
    jsonOutput: `{
  "schemaVersion": 1,
  "build": {"version": "1.4.0", "commit": "a1b2c3d", "goVersion": "go1.25.7", "platform": "darwin/arm64"},
  "command": "scratchpad",
  "action": "init",
  "root": "/Users/you/moonbeam-os/teams/web",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "scratchpad-init",
      "title": "Personal Scratchpad",
      "findings": [
        {"severity": "ok", "code": "scratchpad.scaffolded"}
      ]
    }
  ],
  "guidance": [],
  "error": null
}`,
    exitCodesPossible: [0, 2],
  },
  {
    id: 'tui',
    name: 'tui',
    syntax: 'company-os tui',
    category: 'Utility',
    description: 'Interactive terminal UI over the current workspace: 10 read-only screens and 5 forms that scaffold artifacts. Every form previews the exact company-os command it will run and writes nothing until you confirm.',
    flags: [],
    example: 'company-os tui',
    expectedOutput: `company-os — workspace overview

  > workspace overview
    today (role view)
    validate results
    component browser
    PRD browser
    discovery browser
    governance explain
    skills list
    ids list
    workspace status
    new discovery brief (writes)
    new PRD (writes)
    add team (writes)
    add platform (writes)
    add component (writes)

  ↑↓/kj move · enter open · esc back · q quit`,
    jsonOutput: `// company-os tui has no --json.
//
// It is a human surface and is deliberately outside the agent contract.
// Agents and scripts call the underlying subcommands, where the structured
// envelope and the differentiated exit codes live.
//
// The TUI runs those same commands: each read-only screen executes the
// equivalent subcommand, and each (writes) form previews the exact
// flag-complete invocation before running it.`,
    exitCodesPossible: [0, 2, 7],
  },
];
