export interface GlossaryTerm {
  term: string;
  plain: string;
}

export interface RoleGuide {
  role: string;
  badge: string;
  focusText: string;
  keySteps: string[];
}

export const GLOSSARY_ITEMS: GlossaryTerm[] = [
  { term: 'Company OS', plain: 'The overall digital rulebook and blueprint for all teams in the company.' },
  { term: 'Team OS', plain: 'A smaller version of the rulebook tailored for a single independent team.' },
  { term: 'PRD (Product Requirement)', plain: 'A document describing a new feature or change before anyone builds it.' },
  { term: 'Discovery Brief', plain: 'An early research document exploring a problem before deciding to solve it.' },
  { term: 'Representation of Reality', plain: 'Keeping documentation 100% up-to-date with actual working software.' },
  { term: 'CLI (Command Line Interface)', plain: 'A simple text box where you type short commands like "company-os validate" instead of clicking menus.' },
  { term: 'Validation Gate', plain: 'Automated safety checks (like airport security) that test your work before it gets published.' },
  { term: 'Mandatory Rule', plain: 'A strict rule (like security) that everyone must follow with zero exceptions unless approved.' },
  { term: 'Deviation', plain: 'An official explanation recorded when a team chooses not to follow a default guideline.' },
  { term: 'Monorepo vs Federated', plain: 'Monorepo means everything is stored in one single folder; Federated means projects are stored in separate folders linked together.' },
  { term: 'workspace.yaml (Manifest)', plain: 'A shopping list saying which other GitHub repositories to pull folders from, and where each one lands on your computer. Having this file at all is what turns on federated mode.' },
  { term: 'Slice', plain: 'A copy of a few folders taken from someone else\'s repository. It arrives locked as read-only, so you fix problems in the original repo, never in the copy.' },
  { term: 'Pin', plain: 'The exact version (a tag or a commit) of another repo you agreed to use. Pinning means nobody can change your rules behind your back — you upgrade on purpose.' },
  { term: 'workspace.lock.yaml (Lock)', plain: 'An automatic receipt listing exactly which files arrived and their fingerprints. Validation re-checks the fingerprints, so an accidental edit to a copied file is caught immediately.' },
  { term: 'MCP (Model Context Protocol)', plain: 'A way for an AI assistant to plug into outside tools like GitHub. Company OS does not ship one — it belongs to your assistant, and it may read things, but only the company-os command is allowed to change your workspace.' },
];

export const ROLE_GUIDES: RoleGuide[] = [
  {
    role: 'Product Managers & Designers',
    badge: 'Planning & Specs',
    focusText: 'You write product specs in Markdown inside the prds/ directory. The system ensures your spec has an owner, status, and clean formatting.',
    keySteps: [
      'Create a new PRD using company-os prd create.',
      'Fill in problem, solution, and acceptance criteria.',
      'Run company-os validate before submitting a Pull Request.',
    ],
  },
  {
    role: 'Software Engineers & Interns',
    badge: 'Build & Ship',
    focusText: 'You write code alongside Markdown docs. Pre-commit hooks run 8 automated gates to catch missing dates or broken links before you merge.',
    keySteps: [
      'Check existing PRDs and discovery briefs in prds/.',
      'Run company-os validate locally to verify 8 safety gates.',
      'Merge PRs cleanly with instant AI indexing.',
    ],
  },
  {
    role: 'Tech Leads & VPs of Engineering',
    badge: 'Governance & Security',
    focusText: 'You define central company governance in company-os.yaml and enforce security compliance across all product squads automatically.',
    keySteps: [
      'Define Mandatory vs Default rules in Governance Engine.',
      'Review deviations logged by squads in team-os.yaml.',
      'Monitor company-wide compliance in real-time.',
    ],
  },
];
