import { Building2, Users, type LucideIcon } from 'lucide-react';
import { TabType } from '../types';
import { Tone } from '../components/ui/primitives';

/**
 * Copy for the two "explain it plainly" pages: Company OS and Team OS.
 *
 * Both pages answer the same four questions in the same order — what is it,
 * what does it do, how does it work, why do you need it — so a reader can hold
 * the two layers side by side. One data shape, one view component, two entries.
 */
export interface OsLayerPage {
  id: TabType;
  icon: LucideIcon;
  tone: Tone;
  eyebrow: string;
  title: string;
  /** One sentence. If a reader remembers nothing else, this is the sentence. */
  lead: string;
  /** The "what is it" section: plain prose, then the folders this layer owns. */
  what: {
    paragraphs: string[];
    ownsTitle: string;
    owns: { path: string; description: string }[];
  };
  /** The "what does it do" section — the jobs, in outcome language. */
  does: { title: string; description: string }[];
  /** The "how does it work" section — the loop, in order, with real commands. */
  how: { title: string; command?: string; description: string }[];
  /** The "why you need it" section — today's pain, and what replaces it. */
  why: { pain: string; instead: string }[];
  /** Common misreadings, corrected. */
  notThis: string[];
  /** The line that hands the reader to the other page. */
  handoff: { tab: TabType; label: string; description: string };
}

export const OS_LAYER_PAGES: OsLayerPage[] = [
  {
    id: 'company-os',
    icon: Building2,
    tone: 'accent',
    eyebrow: 'The shared layer',
    title: 'Company OS',
    lead:
      'The rules everyone inherits, and the catalog of what the company actually owns — written as Markdown and YAML in Git, and checked by a command instead of a meeting.',
    what: {
      paragraphs: [
        'Company OS is not an application you log into. It is a folder in a Git repository. Standards, platform requirements, the component catalog, the current state of production, and the canonical vocabulary all live there as plain files with YAML frontmatter — readable by a person, by an editor like Obsidian, and by an agent.',
        'Because the rules are files, they diff, they review, and they version like code. Because they are validated, they cannot quietly rot: company-os validate is a gate you can run locally and in CI, and it fails on stale ownership, expired exceptions, unchecked done-criteria, and hand-edited generated files.',
      ],
      ownsTitle: 'What lives in this layer',
      owns: [
        {
          path: 'company-os/standards/',
          description:
            'The company baseline. Controls that apply to everything, regardless of team or platform.',
        },
        {
          path: 'platforms/<platform>/',
          description:
            'The authoritative catalog: component descriptors, platform requirements, current-state reality, and live change records.',
        },
        {
          path: 'company-ontology/',
          description:
            'Canonical IDs and shared vocabulary — registered once, referenced by every other layer, never redefined.',
        },
        {
          path: 'knowledge/',
          description:
            'Read-only documentation slices synced from repos that are not workspaces. Indexed for search, never authored by hand.',
        },
      ],
    },
    does: [
      {
        title: 'Publishes one rulebook',
        description:
          'Every control is tagged mandatory, default, or guidance. Mandatory rules must be written as verifiable outcomes, never as implementations — that is the clause that keeps teams free to choose their own method.',
      },
      {
        title: 'Keeps one answer to "who owns this?"',
        description:
          'The component descriptor is the single source of truth for both platform relationships and the accountable team. Ownership registries are reconciled against it; disagreement fails validation.',
      },
      {
        title: 'Tracks reality, not intentions',
        description:
          'Each component carries a reality file describing what is actually running. A change record cannot be archived while its reality file is older than the change itself.',
      },
      {
        title: 'Makes the whole workspace searchable',
        description:
          'Frontmatter and canonical IDs become tags and graph edges. Local Search indexes the same files offline; Obsidian opens them as a vault. No export step, no second copy.',
      },
    ],
    how: [
      {
        title: 'Register meaning once',
        command: 'company-os ids list',
        description:
          'Concepts, components, capabilities, and requirements get canonical URL-style IDs in the ontology registry. Everything else points at those IDs rather than repeating a name.',
      },
      {
        title: 'Declare the rules by tier',
        command: 'company-os governance explain <component>',
        description:
          'Baseline controls plus platform requirements, each with a tier. Ask why any rule applies to any component and get the chain back to the file that declared it.',
      },
      {
        title: 'Resolve the merged rulebook per team',
        command: 'company-os governance resolve --team <team>',
        description:
          'Company baseline + platform requirements + that team\'s accepted deviations, written to a generated file. Generated files are derived, never hand-edited; CI regenerates and diffs.',
      },
      {
        title: 'Derive tags and graph edges',
        command: 'company-os derive',
        description:
          'Tags, wikilinks, and index blocks are computed from frontmatter, so the graph you browse in Obsidian and the graph the CLI validates are the same graph.',
      },
      {
        title: 'Gate on the artifacts',
        command: 'company-os validate',
        description:
          'Eight gates: ownership reconciliation, expiring deviations and exceptions, generated-file drift, frontmatter integrity, and hash integrity for synced slices. Exit code is the contract.',
      },
      {
        title: 'Federate read-only slices',
        command: 'company-os workspace sync',
        description:
          'Pull documentation from other repositories as pinned, read-only slices with a lock file. Edit the source repo, bump the pin, re-sync — never the slice.',
      },
    ],
    why: [
      {
        pain: 'The standard is a document nobody opens, so drift is invisible until an incident.',
        instead:
          'The standard is a file with a tier, and a command that fails when reality no longer matches it.',
      },
      {
        pain: 'Two systems disagree about who owns a service, and both are "authoritative".',
        instead:
          'One descriptor is authoritative. Everything else is reconciled against it, mechanically, on every run.',
      },
      {
        pain: 'Exceptions are granted in chat and never expire.',
        instead:
          'Exceptions carry an approver and an expiry date. A past expiry is a failing gate, not a forgotten thread.',
      },
      {
        pain: 'Nobody can reconstruct why a decision was made.',
        instead: 'The decision, its rationale, and its outcome review are commits. git log is the audit.',
      },
    ],
    notThis: [
      'Not a dashboard or a SaaS — there is no server, no account, and no database. It is your repository.',
      'Not a process police. Validators check outputs: schemas, links, ownership, expiries. They never check the method a team used to produce them.',
      'Not a rewrite of your tooling. It sits beside the code repos it describes and reaches them through pinned, read-only slices.',
    ],
    handoff: {
      tab: 'team-os',
      label: 'Team OS',
      description:
        'The other half: what a single squad owns, and how it moves fast inside these rules without asking permission.',
    },
  },
  {
    id: 'team-os',
    icon: Users,
    tone: 'scope',
    eyebrow: 'The squad layer',
    title: 'Team OS',
    lead:
      'One squad\'s working area: what you own, what you are discovering, what you are shipping, and where you say out loud that a default rule does not fit you.',
    what: {
      paragraphs: [
        'Team OS is the teams/<team>/ folder. It holds the work that belongs to one squad and nobody else — discovery briefs, your definition of ready and done, your ownership registry, and your recorded departures from the company defaults.',
        'It is a peer of the company layer, not a child of it. A team can run Team OS on its own repository from day one, with no platform and no company baseline in sight, and federate later. Nothing about the workflow changes when it does — the same commands, the same files, more rules merged in.',
      ],
      ownsTitle: 'What lives in this layer',
      owns: [
        {
          path: 'teams/<team>/product/discovery/',
          description:
            'Team-private discovery briefs. They move draft → validated before anything downstream can reference them.',
        },
        {
          path: 'teams/<team>/ownership/components.yaml',
          description:
            'What this team is accountable for, consulted on, or informed about. Reconciled against the platform catalog.',
        },
        {
          path: 'teams/<team>/governance/',
          description:
            'Deviations (comply-or-explain, for default-tier rules) and exceptions (approved and expiring, for mandatory ones).',
        },
        {
          path: 'teams/<team>/standards/',
          description:
            'Your definition of ready and definition of done — the checks this team applies to its own work.',
        },
      ],
    },
    does: [
      {
        title: 'Turns a question into a shippable change',
        description:
          'A discovery brief must be validated before it can become a PRD, and the PRD copies its problem and success sections forward. The thread from "why" to "shipped" is a file trail, not a memory.',
      },
      {
        title: 'Tells you the rules that apply to you',
        description:
          'Instead of reading every standard in the company, you resolve one merged rulebook for your team and read that. It already accounts for your accepted deviations.',
      },
      {
        title: 'Gives you a legitimate way to disagree',
        description:
          'Default-tier rules are comply-or-explain: declare a deviation with a rationale and a review date. Mandatory rules need an approved exception with an expiry. Neither is a silent workaround.',
      },
      {
        title: 'Refuses to call a change done too early',
        description:
          'Completion is blocked while any governance checklist item is unchecked or while the component\'s reality file is older than the change. Done means the current state was updated.',
      },
    ],
    how: [
      {
        title: 'See your day',
        command: 'company-os today --role developer',
        description:
          'A role-aware view of what is open, what is expiring, and what is waiting on you — assembled from the same files, no separate tracker.',
      },
      {
        title: 'Start with a question, not a ticket',
        command: 'company-os discover new --team <team> "<title>"',
        description:
          'Scaffold a discovery brief. Validate it when the problem and success criteria hold up. Only a validated brief can move forward.',
      },
      {
        title: 'Propose the change',
        command: 'company-os prd new --team <t> --platform <p> --from-discovery <id>',
        description:
          'The brief becomes a platform-visible change record. From this point the work is no longer team-private — the platform can see what is coming.',
      },
      {
        title: 'Check you are ready before you build',
        command: 'company-os check ready --team <t> --components <id,...>',
        description:
          'Your definition of ready, plus the governance that applies to those components, evaluated as a gate rather than remembered as a habit.',
      },
      {
        title: 'Say so when a rule does not fit',
        command: 'company-os deviation declare <rule> --team <t>',
        description:
          'Comply-or-explain, in the open, with a review date. For mandatory rules use exception request instead — it needs an approver and an expiry.',
      },
      {
        title: 'Update reality, then complete',
        command: 'company-os prd complete --platform <p> <prd-id>',
        description:
          'The change archives, an outcome review is scheduled 90 days out, and the log is appended. Refused while reality is stale — that refusal is the point.',
      },
    ],
    why: [
      {
        pain: 'Every new rule feels like a tax handed down from somewhere else.',
        instead:
          'Mandatory rules are outcomes, not implementations. How you meet them is yours, and disagreement with a default has a recorded, legitimate path.',
      },
      {
        pain: 'Work is "done" in the tracker but the docs describe last quarter.',
        instead: 'Completion is mechanically blocked until the reality file is newer than the change.',
      },
      {
        pain: 'You need the whole company to adopt this before you get any value.',
        instead:
          'A standalone team repo works alone on day one. Federation is a manifest you add later, not a prerequisite.',
      },
      {
        pain: 'Context lives in six tools and none of them are searchable together.',
        instead:
          'Briefs, PRDs, ownership, and standards are Markdown in one directory — grep-able, Obsidian-openable, indexed offline by Local Search, readable by an agent.',
      },
    ],
    notThis: [
      'Not a project management tool. There are no sprints, points, or assignees here — bring your own tracker.',
      'Not a place to copy company standards into. You reference canonical IDs; the merged rulebook is generated for you.',
      'Not blocked on a platform existing. company-os init gives a standalone team workspace that validates on its own.',
    ],
    handoff: {
      tab: 'company-os',
      label: 'Company OS',
      description:
        'The other half: where the rules, the component catalog, and the shared vocabulary this layer points at actually live.',
    },
  },
];

export const getOsLayerPage = (id: TabType): OsLayerPage | undefined =>
  OS_LAYER_PAGES.find((p) => p.id === id);
