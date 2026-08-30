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
    cause: "A governance checklist item in the PRD is still unchecked (- [ ]), the component has no reality doc at all, or its reality doc updated: date is older than the PRD created: date.",
    fix: "The refusal names every reason and prints the exact command that fixes a missing doc. Scaffold with company-os reality new if it names one, edit the reality doc and bump updated: to today, check off checklist boxes with evidence links, and retry. Do not reach for --force: it overrides the rule that a change is not done until reality is updated, which is why the TUI's complete form has no field for it."
  },
  {
    id: 'ts-11',
    tool: 'Company OS',
    category: 'Scaffolding',
    symptom: "reality new refuses: \"component 'X' belongs to platform 'A', not 'B'\" (Exit Code 3)",
    cause: "The --platform you passed is not the one the component's descriptor names. The descriptor is authoritative for the component-to-platform relationship.",
    fix: "Re-run with the platform named in the message. Before 2026-08-29 this pair was written rather than refused — the doc landed where nothing reads it, so the command reported created while prd complete's done-check went on refusing, with nothing connecting the two. If you hit that, delete the misfiled doc under the wrong platform and re-scaffold under the right one."
  },
  {
    id: 'ts-12',
    tool: 'Company OS',
    category: 'CLI/Upgrade',
    symptom: 'The TUI does not offer a brief, PRD, or component I just created in the same session',
    cause: "A build older than 2026-08-29. Four lifecycle pickers, the component browser, and governance explain were built when the TUI launched rather than when the screen was opened, so nothing created in a session appeared in the next step.",
    fix: "Upgrade. On an older build the workaround is to quit and relaunch company-os tui between steps."
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
