import { TabType } from '../types';

export interface SearchResultItemData {
  id: string;
  title: string;
  category: string;
  snippet: string;
  targetTab: TabType;
  iconName: string;
  keywords: string[];
}

export const STATIC_SEARCH_ITEMS: SearchResultItemData[] = [
  {
    id: 'home-index',
    title: 'Index & Home Overview: What is Company OS & Team OS?',
    category: 'Overview & Index',
    snippet: 'Learn the core dual-core system architecture, differences, and mode switching between Company OS and Team OS.',
    targetTab: 'home',
    iconName: 'Layers',
    keywords: ['home', 'index', 'company os', 'team os', 'overview', 'dual core', 'federated', 'standalone', 'what is'],
  },
  {
    id: 'arch-tree',
    title: 'Workspace Folder Architecture Tree',
    category: 'Directory Architecture',
    snippet: 'Interactive file tree of discovery/, prds/, reality/, and teams/ directories.',
    targetTab: 'architecture',
    iconName: 'FolderTree',
    keywords: ['folder', 'tree', 'directory', 'architecture', 'file', 'yaml', 'reality', 'prd', 'discovery'],
  },
  {
    id: 'cli-validate',
    title: 'company-os validate CLI Command',
    category: 'CLI Terminal',
    snippet: 'Executes all 8 automated safety gates before code or doc release.',
    targetTab: 'cli',
    iconName: 'Terminal',
    keywords: ['validate', 'cli', 'terminal', 'gate', 'safety', 'command', 'exit code', 'check'],
  },
  {
    id: 'cli-prd-create',
    title: 'company-os prd create / release',
    category: 'CLI Terminal',
    snippet: 'Scaffolds new PRDs and manages feature lifecycle state transitions.',
    targetTab: 'cli',
    iconName: 'Terminal',
    keywords: ['prd', 'create', 'release', 'scaffold', 'lifecycle', 'draft', 'archived'],
  },
  {
    id: 'gates-all',
    title: 'Validation Safety Gates (1 to 8)',
    category: 'Validation Safety',
    snippet: '8 automated compliance gates covering schema, mandatory tags, security rules, and dates.',
    targetTab: 'validation',
    iconName: 'ShieldCheck',
    keywords: ['validation', 'gate', 'security', 'schema', 'date', 'compliance', 'pass', 'fail'],
  },
  {
    id: 'gov-tiers',
    title: 'Governance Rule Tiers (Mandatory, Default, Guidance)',
    category: 'Governance & Rules',
    snippet: '3-tier rule hierarchy balancing team velocity with strict company outcomes.',
    targetTab: 'governance',
    iconName: 'ShieldCheck',
    keywords: ['governance', 'tier', 'mandatory', 'default', 'guidance', 'rules', 'deviation', 'exception'],
  },
  {
    id: 'sim-workflows',
    title: 'Interactive Workflow Simulator',
    category: 'Workflows',
    snippet: 'Step-by-step simulations of real scenarios (PRD creation, API changes).',
    targetTab: 'workflows',
    iconName: 'PlayCircle',
    keywords: ['workflow', 'simulator', 'scenario', 'prd', 'api', 'step', 'simulation'],
  },
  {
    id: 'skills-list',
    title: 'company-os skills list & 4-Layer Agent Skills',
    category: 'Agent Skills & AI',
    snippet: 'Inspect merged agent skills across Company, Platform, Team, and Personal scratchpad layers.',
    targetTab: 'search-agent',
    iconName: 'Bot',
    keywords: ['skills', 'agent skills', 'company-os skills list', 'claude', 'precedence', 'layer', 'canonical', 'extends'],
  },
  {
    id: 'skills-tiers',
    title: 'Skill Step Tiers: Mandatory, Default, Guidance',
    category: 'Agent Skills & Governance',
    snippet: 'Understand (mandatory), (default), and (guidance) step tiers inside *.SKILL.md authoring files.',
    targetTab: 'search-agent',
    iconName: 'ShieldCheck',
    keywords: ['mandatory', 'default', 'guidance', 'step tier', 'skill tier', 'skill.md', 'authoring'],
  },
  {
    id: 'search-bm25',
    title: 'Local Search & BM25 Engine (Read-Side Discovery)',
    category: 'AI & Knowledge Search',
    snippet: 'Instant offline Markdown search indexed into a local SQLite database for AI agents and developers.',
    targetTab: 'search-agent',
    iconName: 'BookOpen',
    keywords: ['search', 'bm25', 'sqlite', 'fts5', 'local search', 'ai', 'agent', 'skill', 'offline', 'index'],
  },
  {
    id: 'troubleshoot-ref',
    title: 'Troubleshooting & Exit Code Matrix',
    category: 'Reference',
    snippet: 'Quick lookup for symptom-cause-fix diagnostics and exit code contracts (0-8).',
    targetTab: 'reference',
    iconName: 'FileText',
    keywords: ['troubleshoot', 'exit code', 'diagnostic', 'fix', 'error', 'matrix', 'reference'],
  },
  {
    id: 'quiz-mastery',
    title: 'Company OS Mastery Check Quiz',
    category: 'Assessment',
    snippet: '10-question assessment to test understanding of Company OS principles.',
    targetTab: 'quiz',
    iconName: 'Award',
    keywords: ['quiz', 'test', 'mastery', 'questions', 'assessment', 'score', 'learning'],
  }
];
