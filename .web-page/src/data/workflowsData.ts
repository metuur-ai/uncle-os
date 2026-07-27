import { WorkflowScenario } from '../types';

export const WORKFLOW_SCENARIOS_DATA: WorkflowScenario[] = [
  {
    id: 'day-one',
    title: 'Scenario 1: Day 1 - From Zero to Validated Workspace',
    subtitle: 'First-time setup of Company OS workspace with platform, team, component, and initial validation.',
    badge: 'Onboarding & Scaffolding',
    description: 'Walk through initializing a workspace for Moonbeam Bakery, registering a web team and ordering platform, adding an online ordering component, scaffolding its reality doc, and running your first workspace validation gate.',
    steps: [
      {
        stepNumber: 1,
        title: 'Install CLI & Scaffold Workspace',
        command: 'company-os init --company "Moonbeam Bakery" --team web --platform ordering',
        description: 'Runs init to create the 4 core workspace roots: company-os/, platforms/ordering/, teams/web/, and company-ontology/.',
        keyRule: 'Single static binary. No runtime dependencies. Refuses to run inside an existing workspace root.',
        fileAffected: 'Creates /company-os, /platforms/ordering, /teams/web, /company-ontology',
        mockTerminalOutput: `initialized workspace at /Users/you/moonbeam-os
  company: Moonbeam Bakery | first team: web | first platform: ordering
next: cd /Users/you/moonbeam-os && company-os discover new --team web "<discovery title>"`
      },
      {
        stepNumber: 2,
        title: 'Register Component & Scaffold Reality Doc',
        command: 'company-os add component online-ordering-app --platform ordering\ncompany-os reality new online-ordering-app --platform ordering',
        description: 'Creates component descriptor in platforms/ordering/components/online-ordering-app.yaml and scaffolds its true reality doc.',
        keyRule: 'The component descriptor is the single source of truth for ownership.',
        fileAffected: 'platforms/ordering/reality/components/online-ordering-app.md',
        mockTerminalOutput: `added component 'online-ordering-app' to platform 'ordering'
scaffolded platforms/ordering/reality/components/online-ordering-app.md
next: edit representation of reality and set updated: date`
      },
      {
        stepNumber: 3,
        title: 'Resolve Team Governance',
        command: 'company-os governance resolve --team web',
        description: 'Merges company baseline, platform standards, and team deviations into teams/web/generated/effective-governance.yaml.',
        keyRule: 'Generated file — never hand-edit. Re-run resolve whenever ownership or rules change.',
        fileAffected: 'teams/web/generated/effective-governance.yaml',
        mockTerminalOutput: `resolved governance for team 'web' (1 component(s))
wrote teams/web/generated/effective-governance.yaml
  online-ordering-app: platforms [ordering], 3 company + 3 platform requirement(s)`
      },
      {
        stepNumber: 4,
        title: 'Validate Workspace',
        command: 'company-os validate',
        description: 'Executes all workspace validation gates.',
        keyRule: 'Returns exit code 0 if all gates pass. CI branches on exit code.',
        mockTerminalOutput: `[1/7] ownership reconciliation
  [ok] online-ordering-app: registry and descriptor agree (ordering)
[2/7] deviation and exception expiry
[3/7] active PRD contracts
[4/7] frontmatter core and tag derivation (interop contract)
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
[6/7] feature-index drift (derived component->artifact map)
[7/7] custom skills layering (shadowing + extends resolution)
PASS`
      }
    ]
  },
  {
    id: 'change-lifecycle',
    title: 'Scenario 2: Full Change Lifecycle (Discovery -> PRD -> Reality -> Done)',
    subtitle: 'Shipping a feature end-to-end: team discovery brief, platform PRD, reality update, and PRD completion.',
    badge: 'Core Change Engine',
    description: 'Experience how Company OS enforces that "a change is not done until the Representation of Reality is updated". See how prd complete blocks with Exit Code 5 if reality doc date is stale.',
    steps: [
      {
        stepNumber: 1,
        title: 'Create & Validate Discovery Brief',
        command: 'company-os discover new "Same-day pickup slots" --team web\ncompany-os discover validate 2026-same-day-pickup-slots --team web',
        description: 'Capture team-private problem signal, hypothesis, and success criteria in teams/web/product/discovery/2026-same-day-pickup-slots/brief.md.',
        keyRule: 'Discovery is team-private. Validation fails if Problem signal, Hypothesis, or Success criteria are empty.',
        fileAffected: 'teams/web/product/discovery/2026-same-day-pickup-slots/brief.md',
        mockTerminalOutput: `created teams/web/product/discovery/2026-same-day-pickup-slots/brief.md
  [ok] brief '2026-same-day-pickup-slots' validated (status: validated)`
      },
      {
        stepNumber: 2,
        title: 'Promote Brief to Platform PRD',
        command: 'company-os prd new --team web --platform ordering --components online-ordering-app --from-discovery 2026-same-day-pickup-slots',
        description: 'Scaffolds PRD under platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md with governance snapshot and applicable checklist.',
        keyRule: 'Auto-copies problem statement & success metrics from brief and stamps governanceSnapshot so PRD is judged against rules at creation time.',
        fileAffected: 'platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md',
        mockTerminalOutput: `created platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md
stamped governanceSnapshot with 2 applicable requirements`
      },
      {
        stepNumber: 3,
        title: 'Attempt Completion Without Reality Update (Fails)',
        command: 'company-os prd complete 2026-same-day-pickup-slots --platform ordering',
        description: 'Demonstrates Company OS enforcement gate: attempting to complete PRD before updating reality doc throws Exit Code 5.',
        keyRule: 'EXIT CODE 5 (Precondition Failed): A change is NOT done until Representation of Reality is updated!',
        mockTerminalOutput: `done-check failed — a change is not done until reality is updated:
  [FAIL] reality doc for 'online-ordering-app' not updated since PRD created
exit code: 5 (Precondition Failed)`
      },
      {
        stepNumber: 4,
        title: 'Update Reality Doc & Complete PRD',
        command: '# Update platforms/ordering/reality/components/online-ordering-app.md and set updated: 2026-07-26\ncompany-os prd complete 2026-same-day-pickup-slots --platform ordering',
        description: 'Update reality doc, check off governance checklist items with evidence links, and successfully archive PRD.',
        keyRule: 'Archives PRD to archive/prds/, schedules 90-day outcome review, and appends log.md.',
        fileAffected: 'platforms/ordering/archive/prds/2026-same-day-pickup-slots',
        mockTerminalOutput: `archived -> platforms/ordering/archive/prds/2026-same-day-pickup-slots
outcome review scheduled (due in 90 days)
appended platforms/ordering/log.md
PASS`
      }
    ]
  },
  {
    id: 'standalone-team',
    title: 'Scenario 3: Standalone Team OS Mode',
    subtitle: 'How a single team operates with no company or platform layer, degrading gracefully.',
    badge: 'Single Team Mode',
    description: 'Learn how "Team OS" is simply Company OS running with only the teams/<t>/ folder present. All CLI commands degrade gracefully without error.',
    steps: [
      {
        stepNumber: 1,
        title: 'Inspect Standalone Team Layout',
        command: 'ls -R teams/solo/',
        description: 'Only teams/solo/ exists. There is no company-os/, platforms/, or company-ontology/.',
        keyRule: 'Graceful degradation: company-os skills list and validate skip missing layers silently.',
        fileAffected: 'teams/solo/team.yaml, teams/solo/product/discovery/',
        mockTerminalOutput: `teams/solo/
├── team.yaml
├── governance/
│   ├── deviations.yaml
│   └── exceptions.yaml
├── product/
│   └── discovery/
└── standards/
    ├── definition-of-ready.md
    └── definition-of-done.md`
      },
      {
        stepNumber: 2,
        title: 'Run Validation in Standalone Mode',
        command: 'company-os validate',
        description: 'Validation passes cleanly because absence of platform/company layers is a valid state.',
        keyRule: 'Gates 1 and 3 skip cleanly when platforms are absent.',
        mockTerminalOutput: `[2/7] deviation and exception expiry
[4/7] frontmatter core and tag derivation (interop contract)
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
[7/7] custom skills layering (shadowing + extends resolution)
PASS`
      },
      {
        stepNumber: 3,
        title: 'Seamless Expansion into Federated Workspace',
        command: 'company-os add platform ordering\ncompany-os add component online-ordering-app --platform ordering',
        description: 'When the team grows, add platform ordering without restructuring or altering existing team files.',
        keyRule: 'Team directory structure never changes whether standalone or in a multi-team company.',
        mockTerminalOutput: `added platform 'ordering'
added component 'online-ordering-app' to platform 'ordering'
workspace expanded seamlessly!`
      }
    ]
  },
  {
    id: 'governance-tiers',
    title: 'Scenario 4: Governance Tiers, Deviations & Exceptions',
    subtitle: 'Handling mandatory vs default rules: comply-or-explain vs expiring exceptions.',
    badge: 'Rules & Compliance',
    description: 'Understand the three tiers: mandatory (guaranteed outcome), default (comply or explain), guidance (advisory). Practice declaring deviations and requesting exceptions.',
    steps: [
      {
        stepNumber: 1,
        title: 'Explain Component Governance',
        command: 'company-os governance explain online-ordering-app',
        description: 'Prints all applicable rules and their tiers for the component.',
        keyRule: 'mandatory = exception required; default = deviation allowed; guidance = advisory.',
        mockTerminalOutput: `component 'online-ordering-app' (team web):
  - order-confirmation-sla v1.0 (mandatory)
      applies because component belongs to platform 'ordering'
  - estimation/story-points v1.0 (default)
      applies because component belongs to company standard`
      },
      {
        stepNumber: 2,
        title: 'Declare Deviation from Default Rule',
        command: 'company-os deviation declare "company-standard://estimation/story-points" --team web --rationale "Team forecasts with cycle time instead of points."',
        description: 'Declares a comply-or-explain deviation. Automatically sets 180-day review date.',
        keyRule: 'Deviation declare ONLY accepts default-tier rules. Rejects mandatory rules.',
        fileAffected: 'teams/web/governance/deviations.yaml',
        mockTerminalOutput: `declared deviation from company-standard://estimation/story-points in teams/web/governance/deviations.yaml
review due 2027-01-19; re-run: company-os governance resolve --team web`
      },
      {
        stepNumber: 3,
        title: 'Request Expiring Exception for Mandatory Rule',
        command: 'company-os exception request "platform-standard://ordering/order-confirmation-sla" --team web --component legacy-pos-bridge --expires 2026-12-31 --reason "Legacy POS bridge cannot emit sync confirmations."',
        description: 'Drafts a time-boxed exception requiring rule owner signoff.',
        keyRule: 'No permanent exceptions! Requires explicit --expires date. Fails validation when expired.',
        fileAffected: 'teams/web/governance/exceptions.yaml',
        mockTerminalOutput: `exception drafted in teams/web/governance/exceptions.yaml (expires 2026-12-31)
note: mandatory rules require approval by the rule owner before this is valid.`
      }
    ]
  },
  {
    id: 'federation-sync',
    title: 'Scenario 5: Multi-repo Federation & Knowledge Slices',
    subtitle: 'Syncing documentation without source code using workspace.yaml and workspace.lock.yaml.',
    badge: 'Federation & Slices',
    description: 'Learn how Company OS pulls docs into knowledge/ as 0444 read-only slices, verified by Gate 8 hash checks.',
    steps: [
      {
        stepNumber: 1,
        title: 'Scout the Source Repo (optional, agent-assisted)',
        command: '# in your agent: "use the GitHub MCP to read acme/component-library"',
        description: 'Before writing the manifest, have an agent browse the source repo through the GitHub MCP: where the docs live, which tags exist, and what a pin bump would bring in. This is a read, and it happens entirely outside the workspace.',
        keyRule: 'Company OS ships no MCP server or client — this is your agent\'s tool. Reading is fine; the CLI still performs every write, and no MCP path may fetch content into the workspace or write the lock.',
        mockTerminalOutput: `# agent, via GitHub MCP (no company-os command runs here)
repo acme/component-library
  docs/sdd/            42 files   <- documentation lives here
  architecture/         8 files
  src/                 —          <- never cloned; not in paths:
  tags: v1.1.0, v1.1.4, v1.2.0 (latest, 2026-07-14)

diff v1.1.0..v1.2.0 — 6 doc files changed, 2 added

suggestion: paths: [docs/sdd], pin: {tag: v1.2.0}
-> now write it into workspace.yaml yourself; sync does the fetching.`
      },
      {
        stepNumber: 2,
        title: 'Define workspace.yaml Slices',
        command: 'cat workspace.yaml',
        description: 'Declare external git repository, commit/tag pin, and target localDirectory under knowledge/.',
        keyRule: 'paths is an allowlist. Code is never cloned or stored in workspace.',
        fileAffected: 'workspace.yaml',
        mockTerminalOutput: `version: 1
repos:
  - name: component-library
    url: https://github.com/acme/component-library.git
    pin: {tag: v1.2.0}
    slices:
      - {paths: [docs/sdd], localDirectory: knowledge/components/component-library}`
      },
      {
        stepNumber: 3,
        title: 'Synchronize Federated Workspace',
        command: 'company-os workspace sync',
        description: 'Downloads specified paths, writes 0444 read-only files, and records hashes in workspace.lock.yaml.',
        keyRule: 'workspace sync --frozen is used in CI to materialize from lock without network access.',
        fileAffected: 'workspace.lock.yaml, /knowledge/components/component-library/',
        mockTerminalOutput: `synced 1 repo(s) into workspace
wrote workspace.lock.yaml (sha256 verified)
next: company-os graph build && company-os validate`
      },
      {
        stepNumber: 4,
        title: 'Validate Federated Slice Integrity (Gate 8)',
        command: 'company-os validate',
        description: 'Gate 8 activates automatically when workspace.yaml is present to ensure no synced files were hand-edited.',
        keyRule: 'Gate [8/8] fails if any file under knowledge/ is hand-edited or lock is stale.',
        mockTerminalOutput: `[8/8] federated slice integrity (workspace.lock.yaml)
  [ok] 1 slice(s) hash-verified against lockfile
PASS`
      }
    ]
  }
];
