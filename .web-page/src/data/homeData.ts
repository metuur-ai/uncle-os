import { 
  Building2, 
  Users, 
  FolderTree, 
  Terminal, 
  ShieldCheck, 
  Award, 
  FileCode2, 
  ListCheck, 
  GitBranch 
} from 'lucide-react';
import { TabType } from '../types';

export interface DualCoreSystemInfo {
  id: 'company' | 'team';
  title: string;
  badge: string;
  yamlFile: string;
  summary: string;
  keyPoints: string[];
  footerNote: string;
  actionText: string;
  targetTab: TabType;
}

export interface LifecycleStep {
  stepNumber: number;
  title: string;
  description: string;
  iconName: string;
}

export interface QuickDirectoryCard {
  title: string;
  description: string;
  actionText: string;
  targetTab: TabType;
}

export const HOME_HERO_CONTENT = {
  badge: 'Dual-Core Engineering Operating System',
  titlePrefix: 'Welcome to ',
  companyTitle: 'Company OS',
  titleConnector: ' & ',
  teamTitle: 'Team OS',
  description: 'A Git-native operating framework designed to bring speed to squads and clarity to leadership. Learn how root company governance and autonomous team workspaces harmonize without bureaucracy.',
};

export const DUAL_CORE_COMPARISON: DualCoreSystemInfo[] = [
  {
    id: 'company',
    title: 'Company OS',
    badge: 'Federated Governance',
    yamlFile: 'company-os.yaml',
    summary: "Company OS is the organization's central source of truth. It defines the root standards, company ontology, compliance rules, and security gates that all product engineering teams inherit.",
    keyPoints: [
      'Root Metadata & Ontology: Central definitions for environments, terminology, and platforms in company-os.yaml.',
      'Quality Safety Gates (1-8): Automated compliance verification for security, schema, tags, and date freshness.',
      'Federated Multi-Repo Sync: Ensures all squads remain aligned without breaking central build pipelines.',
      'Target Audience: VPs of Engineering, Security Officers, Platform Architects & Tech Leads.',
    ],
    footerNote: 'Central alignment & zero surprises',
    actionText: 'View Governance Rules',
    targetTab: 'governance',
  },
  {
    id: 'team',
    title: 'Team OS',
    badge: 'Squad Autonomy',
    yamlFile: 'team-os.yaml',
    summary: 'Team OS is the local workspace operating environment for individual product squads, feature teams, or standalone projects. It empowers teams to iterate fast without waiting for central approvals.',
    keyPoints: [
      'Local PRD & Discovery Docs: Scaffold specifications in prds/ and exploratory notes in discovery/.',
      'Instant CLI Feedback: Run company-os validate in terminal for fast pre-commit check.',
      'Standalone Autonomy: Can run as an independent repository or federate seamlessly up to Company OS.',
      'Target Audience: Product Managers, Software Engineers, Interns & Designers.',
    ],
    footerNote: 'High velocity & local execution',
    actionText: 'Try Terminal Commands',
    targetTab: 'cli',
  },
];

export const LIFECYCLE_STEPS: LifecycleStep[] = [
  {
    stepNumber: 1,
    title: 'Ideate & Write PRD',
    description: 'Squad uses company-os prd create to scaffold Markdown specs in prds/.',
    iconName: 'FileCode2',
  },
  {
    stepNumber: 2,
    title: 'Local Validation',
    description: 'Developer runs company-os validate in terminal to test against Gates 1 through 8.',
    iconName: 'ListCheck',
  },
  {
    stepNumber: 3,
    title: 'Rule Resolution',
    description: 'Inheritance engine merges Mandatory, Default, and Guidance rules from Company & Team layers.',
    iconName: 'ShieldCheck',
  },
  {
    stepNumber: 4,
    title: 'Federated Release',
    description: 'PR is merged into Git main. Docs auto-index into BM25 SQLite engine for instant AI & team search.',
    iconName: 'Building2',
  },
];

export const QUICK_DIRECTORY_CARDS: QuickDirectoryCard[] = [
  {
    title: 'Workspace Architecture',
    description: 'Inspect folder trees, YAML specs, and reality documents.',
    actionText: 'View Folder Tree',
    targetTab: 'architecture',
  },
  {
    title: 'CLI Terminal Explorer',
    description: 'Simulate company-os CLI commands with live terminal logs.',
    actionText: 'Run CLI Commands',
    targetTab: 'cli',
  },
  {
    title: 'Validation Gates (1-8)',
    description: 'See pass/fail examples for all 8 quality safety checks.',
    actionText: 'Review Safety Gates',
    targetTab: 'validation',
  },
  {
    title: 'Mastery Check Quiz',
    description: 'Test your knowledge with 10 plain-English questions.',
    actionText: 'Start Quiz',
    targetTab: 'quiz',
  },
];
