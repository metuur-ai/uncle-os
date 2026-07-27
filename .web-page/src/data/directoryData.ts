import { TabType } from '../types';

export interface DirectoryItemData {
  title: string;
  description: string;
  targetTab: TabType;
  tag?: string;
}

export interface DirectoryCategoryData {
  id: string;
  categoryName: string;
  iconName: string;
  description: string;
  items: DirectoryItemData[];
}

export const DIRECTORY_CATEGORIES_DATA: DirectoryCategoryData[] = [
  {
    id: 'home',
    categoryName: '0. Index & Overview',
    iconName: 'Home',
    description: 'High-level introduction to Company OS vs Team OS dual-core framework.',
    items: [
      {
        title: 'Index & System Overview',
        description: 'Learn what Company OS and Team OS are, their key differences, and mode switching.',
        targetTab: 'home',
        tag: 'Home View',
      }
    ]
  },
  {
    id: 'arch',
    categoryName: '1. Architecture & Folder Structure',
    iconName: 'FolderTree',
    description: 'Explore the workspace directory, Git-based Markdown files, architectural layers, and multi-repo federation.',
    items: [
      {
        title: 'Workspace Architecture Explorer',
        description: 'Browse workspace tree (discovery/, prds/, reality/, teams/) with live inspector.',
        targetTab: 'architecture',
        tag: 'Core View',
      },
      {
        title: 'Federated Mode: One Tree from Several GitHub Repos',
        description: 'Compose the same directory tree from many pinned repositories with a workspace.yaml manifest. Includes the source-to-destination map, slices: for one repo feeding several targets, and sample sync output.',
        targetTab: 'architecture',
        tag: 'Multi-Repo',
      },
      {
        title: 'Read-Only Slices & workspace.lock.yaml',
        description: 'Why synced content is materialized 0444, how the lock records resolved commits plus per-file hashes, and the six rules the CLI enforces on a manifest.',
        targetTab: 'architecture',
      },
      {
        title: 'GitHub MCP Boundary: Read vs Write',
        description: 'Company OS ships no MCP server. What an agent may read, what only the CLI may write, and what no agent may ever touch.',
        targetTab: 'architecture',
        tag: 'Agent Tooling',
      },
      {
        title: 'Company OS Root Layer (company-os.yaml)',
        description: 'Company-wide metadata, governance versioning, and ontology specs.',
        targetTab: 'architecture',
      },
      {
        title: 'Teams Layer (teams/team-name)',
        description: 'Team-level specs, team-os.yaml, local rules, and team reality docs.',
        targetTab: 'architecture',
      },
      {
        title: 'Platform Specs (platforms/platform-name)',
        description: 'Shared technology platform rules, APIs, and infrastructure standards.',
        targetTab: 'architecture',
      }
    ]
  },
  {
    id: 'cli',
    categoryName: '2. Terminal CLI Commands & Tools',
    iconName: 'Terminal',
    description: 'Explore all 16 CLI subcommands, exit code contracts (0-8), and JSON envelopes.',
    items: [
      {
        title: 'CLI Terminal Explorer',
        description: 'Simulate company-os CLI commands in a live retro terminal window.',
        targetTab: 'cli',
        tag: 'Interactive Tool',
      },
      {
        title: 'Validation Command: company-os validate',
        description: 'Run pre-commit checks across all 8 validation gates locally.',
        targetTab: 'cli',
      },
      {
        title: 'PRD Scaffold Command: company-os prd create',
        description: 'Scaffold a new PRD specification in prds/ with standard metadata headers.',
        targetTab: 'cli',
      },
      {
        title: 'Rule Merging Command: company-os rules show',
        description: 'Inspect resolved inheritance rules (Mandatory vs Default vs Guidance).',
        targetTab: 'cli',
      }
    ]
  },
  {
    id: 'workflows',
    categoryName: '3. Interactive PRD Lifecycle Workflows',
    iconName: 'PlayCircle',
    description: 'Step-by-step simulations of feature lifecycles from brief to production release.',
    items: [
      {
        title: 'PRD Lifecycle Workflow Simulator',
        description: 'Walk through Discovery Brief -> Draft PRD -> Safety Gate Check -> Federated Release.',
        targetTab: 'workflows',
        tag: 'Simulator',
      },
      {
        title: 'Scenario 5: Multi-repo Federation & Knowledge Slices',
        description: 'Four steps: scout the source repo with the GitHub MCP, declare the workspace.yaml entry, run workspace sync, then verify slice integrity at Gate 8.',
        targetTab: 'workflows',
        tag: 'Federation',
      }
    ]
  },
  {
    id: 'validation',
    categoryName: '4. Validation Safety Gates (1 to 8)',
    iconName: 'ShieldCheck',
    description: 'Detailed inspection of all 8 automated quality and security safety checks.',
    items: [
      {
        title: 'Gate 1: Security & Leak Detection',
        description: 'Scans for hardcoded API keys, secrets, private certificates, and passwords.',
        targetTab: 'validation',
      },
      {
        title: 'Gate 2: Schema Validation',
        description: 'Verifies YAML/JSON syntax against company-os.schema.json.',
        targetTab: 'validation',
      },
      {
        title: 'Gate 3: Representation of Reality',
        description: 'Verifies reality.md matches code state and date freshness limits.',
        targetTab: 'validation',
      },
      {
        title: 'Gate 8: Federated Slice Integrity',
        description: 'Only runs when workspace.yaml is present. Re-hashes every synced slice against workspace.lock.yaml to catch hand-edits and moved targets.',
        targetTab: 'validation',
      }
    ]
  },
  {
    id: 'governance',
    categoryName: '5. Governance Tiers & Rule Engine',
    iconName: 'CheckSquare',
    description: 'Mandatory, Default, and Guidance rule tiers with clear deviation logging.',
    items: [
      {
        title: 'Governance Engine & Rule Tier Inspector',
        description: 'Explore Mandatory rules, Default rules, and Guidance tiers with deviation forms.',
        targetTab: 'governance',
        tag: 'Policy Center',
      }
    ]
  },
  {
    id: 'skills-search',
    categoryName: '6. Agent Skills & Local Search Engine',
    iconName: 'Bot',
    description: '4-layer agent skills (company, platform, team, personal) and offline SQLite BM25 search.',
    items: [
      {
        title: 'Agent Skills & Precedence Inspector',
        description: 'Explore the 4 skill layers, extends: resolution, step tiers (mandatory, default, guidance), and reference skills.',
        targetTab: 'search-agent',
        tag: 'AI Skills',
      },
      {
        title: 'skills list CLI Command',
        description: 'Inspect merged skill view or run company-os --json skills list for AI agents.',
        targetTab: 'search-agent',
      },
      {
        title: 'Local Search Engine (BM25 FTS5)',
        description: 'Instant offline Markdown search indexed into a local SQLite database for AI agents.',
        targetTab: 'search-agent',
        tag: 'Search Engine',
      },
      {
        title: 'External Agent Tooling: the GitHub MCP',
        description: 'How an MCP server relates to skills, the read/write split mapped onto skill://governance/syncing-knowledge, and a copyable MCP-assisted prompt that keeps the agent on the read half.',
        targetTab: 'search-agent',
        tag: 'MCP',
      }
    ]
  },
  {
    id: 'quiz',
    categoryName: '7. Knowledge Mastery Quiz',
    iconName: 'Award',
    description: '10 plain-English questions to verify your understanding of Company OS.',
    items: [
      {
        title: 'Company OS Mastery Check Quiz',
        description: 'Test your understanding with instant score reporting and explanations.',
        targetTab: 'quiz',
        tag: 'Assessment',
      }
    ]
  }
];
