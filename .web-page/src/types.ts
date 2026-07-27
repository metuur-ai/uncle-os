export type TabType = 
  | 'home'
  | 'architecture' 
  | 'cli' 
  | 'workflows' 
  | 'governance' 
  | 'validation' 
  | 'search-agent' 
  | 'reference' 
  | 'quiz';

export interface CliCommand {
  id: string;
  name: string;
  syntax: string;
  category: 'Scaffolding' | 'Lifecycle' | 'Governance' | 'Validation' | 'Federation' | 'Utility';
  description: string;
  flags: { flag: string; type: string; required: boolean; description: string; defaultVal?: string }[];
  example: string;
  expectedOutput: string;
  jsonOutput: string;
  exitCodesPossible: number[];
}

export interface WorkspaceNode {
  id: string;
  name: string;
  type: 'folder' | 'file';
  path: string;
  layer: 'company-os' | 'platforms' | 'teams' | 'company-ontology' | 'knowledge' | 'root';
  description: string;
  writtenBy: string;
  validatorCheck?: string;
  content?: string;
  children?: WorkspaceNode[];
  standaloneIncluded?: boolean;
}

export interface ValidationGate {
  id: number;
  name: string;
  shortName: string;
  description: string;
  checks: string[];
  absenceTolerant: boolean;
  federatedOnly?: boolean;
  examplePass: string;
  exampleFail: string;
  fixAction: string;
}

export interface WorkflowStep {
  stepNumber: number;
  title: string;
  command: string;
  description: string;
  keyRule: string;
  fileAffected?: string;
  mockTerminalOutput: string;
  interactiveForm?: {
    fields: { key: string; label: string; placeholder: string; required?: boolean }[];
  };
}

export interface WorkflowScenario {
  id: string;
  title: string;
  subtitle: string;
  badge: string;
  description: string;
  steps: WorkflowStep[];
}

export interface ExitCodeInfo {
  code: number;
  meaning: string;
  whenOccurs: string;
  recommendedAction: string;
}

export interface TroubleshootingItem {
  id: string;
  symptom: string;
  tool: 'Company OS' | 'Local Search';
  cause: string;
  fix: string;
  category: 'Validation' | 'Scaffolding' | 'PRD' | 'Federation' | 'CLI/Upgrade' | 'Search';
}

export interface QuizQuestion {
  id: number;
  question: string;
  options: string[];
  correctIndex: number;
  explanation: string;
  category: string;
}
