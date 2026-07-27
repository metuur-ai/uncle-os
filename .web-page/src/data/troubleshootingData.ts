import { TroubleshootingItem } from '../types';

export const TROUBLESHOOTING_DATA: TroubleshootingItem[] = [
  {
    id: 'ts-1',
    tool: 'Company OS',
    category: 'Validation',
    symptom: 'validate gate [1/N] fails: ownership mismatch',
    cause: "A team's ownership/components.yaml claims accountable, but the component descriptor's ownership.accountableTeam disagrees.",
    fix: "Edit the component descriptor in platforms/<p>/components/<comp>.yaml — it is the single source of truth, not the team registry."
  },
  {
    id: 'ts-2',
    tool: 'Company OS',
    category: 'Validation',
    symptom: 'validate gate [2/N] fails: expired deviation or exception',
    cause: "A deviation's reviewDate or an exception's expires date is in the past.",
    fix: "Re-declare the deviation with company-os deviation declare or request an exception with a fresh future date."
  },
  {
    id: 'ts-3',
    tool: 'Company OS',
    category: 'PRD',
    symptom: 'prd complete refuses with done-check error (Exit Code 5)',
    cause: "A governance checklist item in the PRD is still unchecked (- [ ]), or the component's reality doc updated: date is older than the PRD created: date.",
    fix: "Update the component reality doc (platforms/<p>/reality/components/<comp>.md), bump updated: date to today, check off checklist boxes with evidence links, and retry."
  },
  {
    id: 'ts-4',
    tool: 'Company OS',
    category: 'Validation',
    symptom: 'validate gate [4/N], [5/N], or [6/N] fails: drift (run: company-os graph build)',
    cause: "Committed derived content (tags, CLAUDE.md context node, or feature-index) no longer matches a fresh graph build.",
    fix: "Run company-os graph build, review the git diff, and commit the changes."
  },
  {
    id: 'ts-5',
    tool: 'Company OS',
    category: 'Validation',
    symptom: 'validate gate [8/N] fails: federated slice integrity',
    cause: "A materialized slice under knowledge/ was hand-edited or modified.",
    fix: "Discard hand-edits, re-run company-os workspace sync. Slices are 0444 read-only content — fix upstream in the source repo."
  },
  {
    id: 'ts-6',
    tool: 'Company OS',
    category: 'CLI/Upgrade',
    symptom: 'company-os --version prints usage banner and exits 2, or --json is rejected',
    cause: "A leftover launcher from the old Python install.sh is earlier on PATH and is silently shadowing the new Go binary.",
    fix: "Run type -a company-os to list resolution order. Delete ~/.local/bin/company-os launcher and ~/.local/share/company-os kit directory, then verify with company-os --version."
  },
  {
    id: 'ts-7',
    tool: 'Company OS',
    category: 'Scaffolding',
    symptom: 'governance resolve or discover/prd/check dies with team not found error',
    cause: "--team argument was omitted or misspelled.",
    fix: "Pass --team <id>. Verify team ID with company-os ids list --prefix team://."
  },
  {
    id: 'ts-8',
    tool: 'Company OS',
    category: 'Federation',
    symptom: 'company-os workspace ... dies immediately',
    cause: "No workspace.yaml manifest present at workspace root.",
    fix: "Monorepo workspaces do not need workspace commands (use company-os validate directly), or create workspace.yaml per federation runbook."
  },
  {
    id: 'ts-9',
    tool: 'Local Search',
    category: 'Search',
    symptom: 'Local Search returns "No repos added yet" or search finds nothing',
    cause: "No repositories registered in Local Search SQLite indexer.",
    fix: "Run local-search repo add <folder> [name] and verify with local-search repo list and local-search scan."
  },
  {
    id: 'ts-10',
    tool: 'Local Search',
    category: 'Search',
    symptom: 'Local Search results look corrupt or throw database errors',
    cause: "Corrupted SQLite FTS5 specs.db cache file.",
    fix: "Run rm ~/.local-search/specs.db (safe — disposable cache) or local-search reset."
  }
];
