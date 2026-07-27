export interface SkillLayer {
  layer: 'company' | 'platform' | 'team' | 'personal';
  title: string;
  location: string;
  description: string;
  authority: 'canonical' | 'team' | 'personal';
  overrideRule: string;
}

export interface SamplePrompt {
  id: string;
  category: string;
  title: string;
  description: string;
  promptText: string;
  targetSkillId?: string;
  skillName?: string;
}

export interface StepTierInfo {
  tier: 'mandatory' | 'default' | 'guidance';
  label: string;
  badgeColor: string;
  description: string;
  behavior: string;
}

export interface ReferenceSkill {
  id: string;
  title: string;
  scope: string;
  authority: string;
  summary: string;
  steps: { number: number; tier: 'mandatory' | 'default' | 'guidance'; text: string }[];
  location: string;
}

export const SKILL_LAYERS: SkillLayer[] = [
  {
    layer: 'company',
    title: 'Company Layer',
    location: 'company-os/skills/<name>.SKILL.md',
    description: 'Central company baseline rules. Applies to all platforms and teams in the organization.',
    authority: 'canonical',
    overrideRule: 'Cannot be overridden by team or personal layers for mandatory steps.',
  },
  {
    layer: 'platform',
    title: 'Platform Layer',
    location: 'platforms/<p>/skills/<name>.SKILL.md',
    description: 'Platform-specific engineering standards and authoritative authoring procedures.',
    authority: 'canonical',
    overrideRule: 'Outranks team/personal layers for mandatory steps.',
  },
  {
    layer: 'team',
    title: 'Team Layer',
    location: 'teams/<t>/skills/<name>.SKILL.md',
    description: 'Team-specific skill extensions using extends: platform-skill:// references.',
    authority: 'team',
    overrideRule: 'Extends canonical skills without replacing them. Cannot shadow canonical IDs/filenames.',
  },
  {
    layer: 'personal',
    title: 'Personal Layer (Scratchpad)',
    location: 'teams/<t>/scratchpad/personal-rules/*.md',
    description: 'Git-ignored personal preferences created via company-os scratchpad init.',
    authority: 'personal',
    overrideRule: 'Non-overriding. Outranks canonical default/guidance, but loses to canonical mandatory.',
  },
];

export const STEP_TIERS: StepTierInfo[] = [
  {
    tier: 'mandatory',
    label: '(mandatory)',
    badgeColor: 'bg-rose-100 text-rose-800 border-rose-200',
    description: 'Non-negotiable requirement enforced by automated CI validation gates.',
    behavior: 'Validator gate [7/N] or workflow commands will block until satisfied.',
  },
  {
    tier: 'default',
    label: '(default)',
    badgeColor: 'bg-indigo-100 text-indigo-800 border-indigo-200',
    description: 'Standard procedure expected of all teams unless a deviation is recorded.',
    behavior: 'Divergence requires logging a deviation via company-os deviation declare.',
  },
  {
    tier: 'guidance',
    label: '(guidance)',
    badgeColor: 'bg-slate-100 text-slate-700 border-slate-200',
    description: 'Advisory suggestion or recommended best practice.',
    behavior: 'Untracked by validation gates; team or author may ignore freely.',
  },
];

export const REFERENCE_SKILLS: ReferenceSkill[] = [
  {
    id: 'skill://product/running-discovery',
    title: 'Running Discovery',
    scope: 'product',
    authority: 'canonical',
    summary: 'Guides product managers and engineers from initial problem signal to a validated brief.',
    location: 'company-os/skills/running-discovery.SKILL.md',
    steps: [
      { number: 1, tier: 'mandatory', text: 'Scaffold brief with `company-os discover new --team <id> "<title>"`.' },
      { number: 2, tier: 'mandatory', text: 'Document problem signal, user evidence, and measurable success metrics.' },
      { number: 3, tier: 'default', text: 'Validate brief against Gate 3 specs before drafting a PRD.' },
    ],
  },
  {
    id: 'skill://product/creating-prd',
    title: 'Creating a PRD',
    scope: 'product',
    authority: 'canonical',
    summary: 'Transforms a validated discovery brief into an active PRD specification.',
    location: 'platforms/<p>/skills/creating-prd.SKILL.md',
    steps: [
      { number: 1, tier: 'mandatory', text: 'Scaffold PRD with `company-os prd new --from-discovery <brief-id>`.' },
      { number: 2, tier: 'mandatory', text: 'Ensure governanceSnapshot block is auto-injected with current rules.' },
      { number: 3, tier: 'default', text: 'Follow platform PRD section structure and list affected component IDs.' },
      { number: 4, tier: 'guidance', text: 'Attach user flow diagrams or architectural sketches if available.' },
    ],
  },
  {
    id: 'skill://product/completing-a-change',
    title: 'Completing a Change',
    scope: 'product',
    authority: 'canonical',
    summary: 'Archives an active PRD and synchronizes reality documentation upon feature release.',
    location: 'platforms/<p>/skills/completing-a-change.SKILL.md',
    steps: [
      { number: 1, tier: 'mandatory', text: 'Update component reality doc in `platforms/<p>/reality/components/<id>.md`.' },
      { number: 2, tier: 'mandatory', text: 'Verify all governance checklist items in the PRD are checked.' },
      { number: 3, tier: 'mandatory', text: 'Execute `company-os prd complete --platform <p> <prd-id>`.' },
    ],
  },
  {
    id: 'skill://governance/requesting-an-exception',
    title: 'Requesting an Exception',
    scope: 'governance',
    authority: 'canonical',
    summary: 'Procedure for requesting a time-boxed exception to a mandatory governance rule.',
    location: 'company-os/skills/requesting-an-exception.SKILL.md',
    steps: [
      { number: 1, tier: 'mandatory', text: 'Specify target mandatory rule URI, component ID, and explicit expiration date.' },
      { number: 2, tier: 'mandatory', text: 'Run `company-os exception request <rule> --team <t> --component <c> --expires <date>`.' },
      { number: 3, tier: 'default', text: 'Record technical debt remediation plan prior to exception expiry.' },
    ],
  },
  {
    id: 'skill://governance/syncing-knowledge',
    title: 'Syncing Knowledge Catalog',
    scope: 'governance',
    authority: 'canonical',
    summary: 'Pulls external documentation repositories into knowledge/ pinned and hash-locked.',
    location: 'company-os/skills/syncing-knowledge.SKILL.md',
    steps: [
      { number: 1, tier: 'mandatory', text: 'Declare repository, commit/tag pin, and slice paths in `workspace.yaml`.' },
      { number: 2, tier: 'mandatory', text: 'Execute `company-os workspace sync` to write hash locks.' },
      { number: 3, tier: 'mandatory', text: 'Run `company-os graph build` to regenerate context nodes in `knowledge/CLAUDE.md`.' },
    ],
  },
];

export const AUTHORING_RULES = [
  {
    rule: 'Exact File Extension Constraint',
    detail: 'Skill files MUST end with `.SKILL.md` (e.g., `creating-prd.SKILL.md`). Files named `SKILL.md` inside subdirectories are invisible to discovery.',
  },
  {
    rule: 'Frontmatter Schema',
    detail: 'Must carry `type: skill`, a unique `id` (`skill://<scope>/<name>`), `version: "1.0"`, `authority`, and `appliesTo`. Never hand-write `tags:` list — `company-os graph build` derives tags.',
  },
  {
    rule: 'No Shadowing Rule (Gate [7/N])',
    detail: 'You cannot override a canonical skill by redefining its ID or filename. Team skills MUST use a distinct ID and declare `extends: platform-skill://<p>/<name>`. Violation fails Gate 7.',
  },
  {
    rule: 'Precedence & Resolution',
    detail: 'Configured in `teams/<t>/team.yaml`: `canonical-mandatory > personal > canonical-default > canonical-guidance`. Canonical mandatory steps always win over personal rules.',
  },
  {
    rule: 'No Execution Authority',
    detail: 'Skills guide humans and AI agents on *how to author* artifacts correctly. A skill cannot grant an agent permission to bypass mandatory gates, hand-edit `generated/` files, or modify `0444` synced slices.',
  },
];

export const MOCK_SKILLS_CLI_TEXT = `$ company-os skills list
== agent skills (merged view across 4 layers) ==

layers (origin-labeled):
  [company] <none>
  [platform:ordering] creating-prd  id=skill://product/creating-prd authority=canonical
  [team:web] creating-prd-web  id=skill://product/creating-prd-web authority=team extends=platform-skill://ordering/creating-prd
  [personal:web] my-style  (teams/web/scratchpad/personal-rules/my-style.md)

merged guidance (canonical steps first; personal rules last, non-overriding):

  creating-prd [platform:ordering, authority=canonical]
      1. (mandatory) Scaffold with \`company-os prd new --from-discovery <id>\`.
      2. (default) Follow the platform PRD structure.

  creating-prd-web [team:web, authority=team]
    layered on base platform-skill://ordering/creating-prd:
      [base] 1. (mandatory) Scaffold with \`company-os prd new --from-discovery <id>\`.
      [base] 2. (default) Follow the platform PRD structure.
      1. (default) Attach Figma frame frame to every UI-visible change.

  personal rules (non-overriding — canonical mandatory steps always win):
    [personal:web] my-style

3 skill(s) across 3 populated layer(s).`;

export const MOCK_SKILLS_JSON_TEXT = `{
  "schemaVersion": 1,
  "command": "skills",
  "action": "list",
  "root": "/Users/developer/moonbeam-os",
  "exitCode": 0,
  "sections": [
    {
      "ordinal": 1,
      "slug": "skills-layers",
      "title": "Agent Skills Layers",
      "findings": [
        {
          "severity": "ok",
          "code": "skills.layer-entry",
          "message": "creating-prd-web [team:web, authority=team]",
          "fields": {
            "id": "skill://product/creating-prd-web",
            "authority": "team",
            "extends": "platform-skill://ordering/creating-prd",
            "layer": "team",
            "name": "creating-prd-web"
          }
        }
      ]
    }
  ],
  "guidance": ["company-os --json skills list"]
}`;

export const SAMPLE_PROMPTS: SamplePrompt[] = [
  {
    id: 'orient',
    category: 'Session Orientation',
    title: 'Orient Before Acting',
    description: 'First prompt in an unfamiliar workspace to discover applicable skills, layers, and mandatory steps.',
    skillName: 'skills list',
    promptText: `Run \`company-os --json skills list\` from the workspace root and tell me which skills apply to team customer-engagement, which layer each comes from, and which steps are mandatory. Don't do any work yet.`
  },
  {
    id: 'discovery',
    category: 'Product Lifecycle',
    title: 'Run Discovery',
    description: 'Guides the agent from problem signal to a validated discovery brief.',
    targetSkillId: 'skill://product/running-discovery',
    skillName: '/running-discovery',
    promptText: `Follow skill://product/running-discovery to open a discovery brief for team customer-engagement on this problem: notification volume spikes overnight and users can't mute per channel. Run each command with --json, follow guidance[0], and stop at the first non-zero exit code and show me the failing finding codes.`
  },
  {
    id: 'prd',
    category: 'Product Lifecycle',
    title: 'Create a PRD from a Validated Brief',
    description: 'Scaffolds a PRD specification from a discovery brief with governance snapshot.',
    targetSkillId: 'skill://product/creating-prd',
    skillName: '/creating-prd',
    promptText: `Follow skill://product/creating-prd to scaffold a PRD for platform communications from discovery 2026-per-channel-quiet-hours, components customer-notification-service. Treat the two non-failure findings as load-bearing: if you see prd.governance-unresolved or prd.reality-note, stop and tell me rather than proceeding. Do not hand-edit the brief's status to get past a step 1 failure.`
  },
  {
    id: 'change',
    category: 'Product Lifecycle',
    title: 'Complete a Change & Update Reality',
    description: 'Ensures reality docs, governance checklists, and PRD archiving are completed.',
    targetSkillId: 'skill://product/completing-a-change',
    skillName: '/completing-a-change',
    promptText: `Follow skill://product/completing-a-change for PRD <id> on platform communications. Reality first: update reality/components/customer-notification-service.md and its \`updated:\` date to reflect what actually shipped, then check off the governance checklist items that are genuinely satisfied, with evidence links, then run prd complete. Never pass --force. If it exits 5, report the done.* codes that fired and stop — don't tick a box to get past one.`
  },
  {
    id: 'exception',
    category: 'Governance & Exception',
    title: 'Request a Governance Exception',
    description: 'Drafts a time-boxed exception for a rule a component cannot satisfy.',
    targetSkillId: 'skill://governance/requesting-an-exception',
    skillName: '/requesting-an-exception',
    promptText: `customer-notification-service can't move to the current message envelope this quarter. Follow skill://governance/requesting-an-exception: confirm the tier with governance explain first, and if message-schema really is mandatory, draft the exception with a real --expires and --reason plus compensatingControls. Leave approvedBy empty — that's the rule owner's signature, not yours. Show me the entry before you write it.`
  },
  {
    id: 'knowledge',
    category: 'Knowledge Catalog',
    title: 'Sync Knowledge Catalog',
    description: 'Pins external documentation repos to knowledge/ without direct hand-edits.',
    targetSkillId: 'skill://governance/syncing-knowledge',
    skillName: '/syncing-knowledge',
    promptText: `Follow skill://governance/syncing-knowledge to add github.com/acme/component-library's docs/sdd to the catalog at knowledge/components/component-library, pinned to tag v1.2.0. Write the workspace.yaml entry, run workspace status, then sync, then graph build, then validate. Do not edit anything under knowledge/ directly.`
  },
  {
    id: 'extend-skill',
    category: 'Skill Authoring',
    title: 'Author or Extend a Team Skill',
    description: 'Creates a team skill that extends a canonical base skill rather than shadowing it.',
    targetSkillId: 'skill://product/creating-prd-web',
    skillName: '/creating-prd-web',
    promptText: `Our team always attaches the Figma frame to UI-visible PRDs. Write that as a team skill under teams/customer-engagement/skills/ that extends the platform's creating-prd rather than redefining it: distinct id and file name, \`extends: platform-skill://communications/creating-prd\`, \`type: skill\`, empty tags. Then run graph build and validate and show me gate [7/N].`
  },
  {
    id: 'troubleshoot-gate',
    category: 'Gate Enforcement',
    title: 'When a Mandatory Step Blocks You',
    description: 'Enforces strict cause-based diagnosis instead of illegitimate workarounds.',
    skillName: '/validate-gate',
    promptText: `validate is failing. Diagnose it and fix the cause, not the symptom. You may not hand-edit anything under generated/, any file in a synced slice, or a frontmatter status field to make a gate pass. If the only fix is a governance decision, stop and tell me which one.`
  }
];

export const PROMPTING_RULES = [
  {
    rule: 'Name the skill by ID',
    detail: 'Use exact URI like `skill://product/creating-prd`. Avoid generic phrases like "write a PRD" which invite the agent to invent ad-hoc structures.'
  },
  {
    rule: 'Supply the declared inputs',
    detail: 'Check the skill frontmatter for `inputs:` and `outputs:`. Providing inputs directly in the prompt eliminates agent guesswork.'
  },
  {
    rule: 'Demand --json and exit code branching',
    detail: 'Require the agent to run CLI commands with `--json` and branch on status codes or finding codes instead of paraphrasing errors.'
  },
  {
    rule: 'Specify stop conditions',
    detail: 'Tell the agent explicitly where to pause (e.g. "stop at first non-zero exit code"), preventing unauthorized automated workarounds.'
  },
  {
    rule: 'Explicitly forbid shortcuts',
    detail: 'Prohibit passing `--force`, hand-editing `generated/` files, modifying synced `0444` slices, or flipping `status:` fields to bypass gates.'
  },
  {
    rule: 'Respect personal rules hierarchy',
    detail: 'Personal scratchpad rules in `scratchpad/personal-rules/` apply to workflow style, but never override canonical mandatory steps.'
  }
];

