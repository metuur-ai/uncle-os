export interface ProblemSolutionComparison {
  problemTitle: string;
  problemDescription: string;
  solutionTitle: string;
  solutionDescription: string;
}

export interface WhatPoint {
  title: string;
  subtitle: string;
  description: string;
  iconType: 'layers' | 'check' | 'terminal';
}

export interface HowStep {
  stepNumber: number;
  title: string;
  description: string;
  targetTab: 'architecture' | 'cli' | 'workflows' | 'validation' | 'governance' | 'quiz';
  actionLabel: string;
}

export const WHY_CONTENT: ProblemSolutionComparison = {
  problemTitle: 'THE OLD PROBLEM (Without Company OS)',
  problemDescription: 'Companies write documents and code in completely different places. Within a month, documents become outdated ("documentation drift"). Team members lose hours asking "Where is the latest spec?" and AI tools output wrong code because there are no strict rules.',
  solutionTitle: 'THE SOLUTION (Git-Based Operating System)',
  solutionDescription: 'Company OS puts all company rules, specifications (PRDs), and code in Git right next to each other in standard Markdown files. Everything is checked by automated quality gates BEFORE code gets merged.',
};

export const WHAT_POINTS: WhatPoint[] = [
  {
    title: 'Dual-Core OS',
    subtitle: 'Company OS + Team OS',
    description: 'Company OS sets global rules for security & compliance. Team OS gives individual squads complete freedom to run fast without central bureaucracy.',
    iconType: 'layers',
  },
  {
    title: '8 Safety Gates',
    subtitle: 'Automated CI/CD Validation',
    description: 'Every pull request is automatically tested for security leaks, correct schemas, fresh dates, and required metadata before merging.',
    iconType: 'check',
  },
  {
    title: '1-CLI Terminal',
    subtitle: 'Simulated Developer CLI',
    description: 'Developers run simple terminal commands like "company-os validate" or "company-os prd create" to stay aligned effortlessly.',
    iconType: 'terminal',
  },
];

export const HOW_STEPS: HowStep[] = [
  {
    stepNumber: 1,
    title: 'Explore Workspace Architecture',
    description: 'Click "Workspace Architecture" to inspect actual folder trees, YAML files, and PRD Markdown specs.',
    targetTab: 'architecture',
    actionLabel: 'View Architecture',
  },
  {
    stepNumber: 2,
    title: 'Run Simulated CLI Commands',
    description: 'Click "CLI Terminal Explorer" to execute live commands and watch pre-commit safety gates pass or fail.',
    targetTab: 'cli',
    actionLabel: 'Try CLI Simulator',
  },
  {
    stepNumber: 3,
    title: 'Test Quality Safety Gates',
    description: 'Click "Validation Gates" to see real examples of pass vs. fail checks across all 8 security & governance tiers.',
    targetTab: 'validation',
    actionLabel: 'Review Safety Gates',
  },
  {
    stepNumber: 4,
    title: 'Take the Mastery Quiz',
    description: 'Click "Mastery Check Quiz" to test your knowledge with 10 interactive plain-English questions.',
    targetTab: 'quiz',
    actionLabel: 'Take 10-Q Quiz',
  },
];
