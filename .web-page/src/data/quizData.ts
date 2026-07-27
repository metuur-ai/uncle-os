import { QuizQuestion } from '../types';

export const QUIZ_QUESTIONS_DATA: QuizQuestion[] = [
  {
    id: 1,
    question: 'What is Company OS fundamentally built on?',
    options: [
      'A cloud PostgreSQL database and REST API server',
      'Git-based validated Markdown and YAML files with a single Go binary CLI',
      'A proprietary web app requiring Docker containers',
      'A Node.js microservices platform'
    ],
    correctIndex: 1,
    explanation: 'Company OS is a Git-based operating system where all governance, product, and engineering artifacts live as validated Markdown and YAML files in Git — no SaaS app or database required.',
    category: 'Core Concepts'
  },
  {
    id: 2,
    question: 'What is "Team OS"?',
    options: [
      'A separate SaaS product built by a different vendor',
      'A legacy Python web framework',
      'Company OS running with only the teams/<t>/ layer present, degrading gracefully',
      'A local SQLite database runner'
    ],
    correctIndex: 2,
    explanation: 'Team OS is not a separate product; it is what you get when a single team runs Company OS with only the teams/<t>/ directory present. All CLI subcommands degrade gracefully without requiring company or platform layers.',
    category: 'Architecture'
  },
  {
    id: 3,
    question: 'Which file is the single source of truth for component ownership in Company OS?',
    options: [
      'teams/<t>/ownership/components.yaml',
      'platforms/<p>/components/<comp>.yaml (The Component Descriptor)',
      'company-ontology/ids/registry.yaml',
      'workspace.yaml'
    ],
    correctIndex: 1,
    explanation: 'The component descriptor in platforms/<p>/components/<comp>.yaml is the single source of truth. Gate [1/N] in company-os validate enforces that team claims match this descriptor.',
    category: 'Governance'
  },
  {
    id: 4,
    question: 'Why does "company-os prd complete" refuse to complete a PRD if the component\'s reality doc has an updated: date older than the PRD created: date?',
    options: [
      'Because the server database is locked',
      'Because Company OS enforces the rule: "A change is not done until the Representation of Reality is updated"',
      'Because PRDs must be completed within 24 hours',
      'Because reality docs must be written in HTML'
    ],
    correctIndex: 1,
    explanation: 'In Company OS, reality docs (platforms/<p>/reality/components/<comp>.md) represent true current behavior. A change is not done until reality is updated to reflect the new state, throwing Exit Code 5 if neglected.',
    category: 'Change Lifecycle'
  },
  {
    id: 5,
    question: 'What exit code does company-os return when a gate refuses prd complete because the reality doc date was not updated?',
    options: [
      'Exit Code 0 (Success)',
      'Exit Code 1 (Validation Failed)',
      'Exit Code 5 (Precondition Failed)',
      'Exit Code 8 (Conflict)'
    ],
    correctIndex: 2,
    explanation: 'Exit Code 5 means Precondition Failed (a gate refused action, such as prd complete before reality doc update or prd new from an unvalidated brief).',
    category: 'CLI & Exit Codes'
  },
  {
    id: 6,
    question: 'Which governance tier allows a team to declare a "comply-or-explain" deviation with an automatic 180-day review date?',
    options: [
      'mandatory tier',
      'default tier',
      'guidance tier',
      'optional tier'
    ],
    correctIndex: 1,
    explanation: 'The default tier allows teams to declare a comply-or-explain deviation using "company-os deviation declare". Mandatory rules cannot be deviated — they require an expiring exception request.',
    category: 'Governance'
  },
  {
    id: 7,
    question: 'When is Validation Gate [8/8] (Federated Slice Integrity) executed by company-os validate?',
    options: [
      'Always in every workspace',
      'Only when a workspace.yaml federation manifest is present at the workspace root',
      'Only when running on macOS',
      'Only when Local Search is disabled'
    ],
    correctIndex: 1,
    explanation: 'Gate 8 is dynamic: monorepo workspaces run 7 gates, while presence of workspace.yaml switches validate to 8 gates to verify 0444 read-only slice hashes against workspace.lock.yaml.',
    category: 'Validation Gates'
  },
  {
    id: 8,
    question: 'How do synced knowledge docs in knowledge/ behave under company-os validate?',
    options: [
      'They are strictly validated for required frontmatter and fail if missing type:',
      'They are indexed by Local Search and checked for hash integrity (Gate 8), but skipped by governance gates 1-7',
      'They are deleted on every build',
      'They must be written in C++'
    ],
    correctIndex: 1,
    explanation: 'The knowledge catalog is indexed, not governed. Foreign docs from external component repos carry no Company OS frontmatter, so they are indexed for search and checked for Gate 8 hash integrity without being policed by gates 1-7.',
    category: 'Federation'
  }
];
