import React, { useState } from 'react';
import { WorkspaceNode } from '../types';
import { COMPANY_OS_WORKSPACE_TREE } from '../data/workspaceData';
import { getLesson } from '../lessons';
import {
  PageShell,
  PageHeader,
  Section,
  Card,
  Callout,
  Badge,
  CodeBlock,
  EmptyState,
  cx,
  type Tone,
} from './ui';
import {
  FolderTree,
  Folder,
  FileText,
  ChevronRight,
  ChevronDown,
  Building2,
  Users,
  Info,
} from 'lucide-react';

const lesson = getLesson('architecture');

const LAYER_TONE: Record<WorkspaceNode['layer'], Tone> = {
  root: 'neutral',
  'company-os': 'accent',
  platforms: 'info',
  teams: 'scope',
  'company-ontology': 'warn',
  knowledge: 'success',
};

interface ArchitectureViewProps {
  isStandalone: boolean;
  setIsStandalone: (val: boolean) => void;
}

const findNode = (node: WorkspaceNode, id: string): WorkspaceNode | null => {
  if (node.id === id) return node;
  for (const child of node.children ?? []) {
    const found = findNode(child, id);
    if (found) return found;
  }
  return null;
};

const TEAMS_NODE_ID = 'teams-folder';
const COMPANY_OS_NODE_ID = 'company-os-folder';

export const ArchitectureView: React.FC<ArchitectureViewProps> = ({ isStandalone, setIsStandalone }) => {
  const [selectedId, setSelectedId] = useState<string>(COMPANY_OS_NODE_ID);
  const [expandedFolders, setExpandedFolders] = useState<Record<string, boolean>>({
    root: true,
    [COMPANY_OS_NODE_ID]: true,
    'platforms-folder': true,
    'ordering-platform': true,
    [TEAMS_NODE_ID]: true,
    'web-team': true,
    'ontology-folder': true,
    'knowledge-folder': true,
  });

  const selectedNode =
    findNode(COMPANY_OS_WORKSPACE_TREE, selectedId) ?? COMPANY_OS_WORKSPACE_TREE;

  const toggleFolder = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setExpandedFolders((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const handleScopeChange = (standalone: boolean) => {
    setIsStandalone(standalone);
    setSelectedId(standalone ? TEAMS_NODE_ID : COMPANY_OS_NODE_ID);
  };

  const renderTree = (node: WorkspaceNode, depth: number) => {
    if (isStandalone && node.standaloneIncluded === false) return null;

    const isFolder = node.type === 'folder';
    const isExpanded = expandedFolders[node.id];
    const isSelected = selectedNode.id === node.id;

    return (
      <div key={node.id} className="my-0.5">
        <div
          className={cx(
            'flex items-center gap-1.5 rounded-lg pr-2 text-sm transition-colors duration-150',
            isSelected ? 'bg-accent-soft' : 'hover:bg-surface-sunken'
          )}
          style={{ paddingLeft: `${depth * 1.25}rem` }}
        >
          {isFolder ? (
            <button
              type="button"
              onClick={(e) => toggleFolder(node.id, e)}
              aria-label={isExpanded ? `Collapse ${node.name}` : `Expand ${node.name}`}
              aria-expanded={isExpanded}
              className="flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md text-fg-subtle transition-colors duration-150 hover:bg-surface-sunken hover:text-fg"
            >
              {isExpanded ? (
                <ChevronDown className="h-3.5 w-3.5" aria-hidden="true" />
              ) : (
                <ChevronRight className="h-3.5 w-3.5" aria-hidden="true" />
              )}
            </button>
          ) : (
            <span className="w-7 shrink-0" aria-hidden="true" />
          )}

          <button
            type="button"
            onClick={() => setSelectedId(node.id)}
            aria-current={isSelected ? 'true' : undefined}
            className={cx(
              'flex min-h-9 min-w-0 flex-1 cursor-pointer items-center gap-1.5 rounded-md py-1 text-left transition-colors duration-150',
              isSelected ? 'font-semibold text-accent-text' : 'text-fg'
            )}
          >
            {isFolder ? (
              <Folder className="h-4 w-4 shrink-0 text-fg-subtle" aria-hidden="true" />
            ) : (
              <FileText className="h-4 w-4 shrink-0 text-fg-subtle" aria-hidden="true" />
            )}
            <span className="truncate font-mono text-xs">{node.name}</span>
          </button>

          <Badge tone={LAYER_TONE[node.layer]} className="shrink-0">
            {node.layer}
          </Badge>
        </div>

        {isFolder && isExpanded && node.children && (
          <div className="ml-3.5 border-l border-border pl-1">
            {node.children.map((child) => renderTree(child, depth))}
          </div>
        )}
      </div>
    );
  };

  return (
    <PageShell width="wide">
      <PageHeader eyebrow="Lesson 2 of 9" title={lesson.label} lead={lesson.whyText} icon={FolderTree} />

      <Section
        title="Choose a scope"
        description="The workspace tree is the same on disk either way — this toggle changes which parts of it apply to you."
      >
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div
            role="group"
            aria-label="Workspace scope"
            className="inline-flex gap-1 rounded-lg border border-border bg-surface-sunken p-1"
          >
            <button
              type="button"
              onClick={() => handleScopeChange(false)}
              aria-pressed={!isStandalone}
              className={cx(
                'flex h-11 cursor-pointer items-center gap-2 rounded-md px-4 text-sm font-medium transition-colors duration-150',
                !isStandalone
                  ? 'bg-accent text-accent-fg shadow-xs'
                  : 'text-fg-muted hover:bg-surface hover:text-fg'
              )}
            >
              <Building2 className="h-4 w-4 shrink-0" aria-hidden="true" />
              Full Company OS
            </button>
            <button
              type="button"
              onClick={() => handleScopeChange(true)}
              aria-pressed={isStandalone}
              className={cx(
                'flex h-11 cursor-pointer items-center gap-2 rounded-md px-4 text-sm font-medium transition-colors duration-150',
                isStandalone
                  ? 'bg-scope text-scope-fg shadow-xs'
                  : 'text-fg-muted hover:bg-surface hover:text-fg'
              )}
            >
              <Users className="h-4 w-4 shrink-0" aria-hidden="true" />
              Team OS only
            </button>
          </div>

          <Callout tone={isStandalone ? 'scope' : 'accent'} className="flex-1">
            {isStandalone
              ? 'Standalone mode: only teams/ is shown. Company-os/, platforms/, company-ontology/ and knowledge/ do not exist yet for a team running on its own.'
              : 'Full mode: every layer is shown — company-os/, platforms/, teams/, company-ontology/ and knowledge/.'}
          </Callout>
        </div>
      </Section>

      <Section title="Workspace directory tree" description="Select a folder or file to inspect it on the right.">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
          <Card padding="sm" className="flex h-[32rem] flex-col lg:col-span-5">
            <div className="mb-2 flex items-center justify-between border-b border-border pb-2">
              <span className="text-xs font-semibold text-fg">Directory tree</span>
              <span className="font-mono text-2xs text-fg-subtle">
                {isStandalone ? 'teams/ only' : '/ (root)'}
              </span>
            </div>
            <div className="thin-scrollbar min-h-0 flex-1 overflow-auto">
              {renderTree(COMPANY_OS_WORKSPACE_TREE, 0)}
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5 border-t border-border pt-2">
              {(Object.keys(LAYER_TONE) as WorkspaceNode['layer'][]).map((layer) => (
                <Badge key={layer} tone={LAYER_TONE[layer]}>
                  {layer}
                </Badge>
              ))}
            </div>
          </Card>

          <Card padding="sm" className="flex h-[32rem] flex-col lg:col-span-7">
            <div className="flex items-start justify-between gap-3 border-b border-border pb-3">
              <div className="min-w-0">
                <Badge tone="neutral" mono className="mb-1.5">
                  {selectedNode.type}
                </Badge>
                <h3 className="truncate font-mono text-sm font-semibold text-fg">{selectedNode.path}</h3>
              </div>
            </div>

            <div className="my-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div className="rounded-lg border border-border bg-surface-sunken p-3">
                <p className="text-2xs font-semibold uppercase tracking-widest text-fg-subtle">
                  Written / scaffolded by
                </p>
                <p className="mt-1 text-sm font-medium text-fg">{selectedNode.writtenBy}</p>
              </div>
              <div className="rounded-lg border border-border bg-surface-sunken p-3">
                <p className="text-2xs font-semibold uppercase tracking-widest text-fg-subtle">
                  Validator gate check
                </p>
                <p className="mt-1 text-sm font-medium text-warn-text">
                  {selectedNode.validatorCheck ?? 'Standard frontmatter & structure sanity check'}
                </p>
              </div>
            </div>

            <p className="measure mb-3 text-sm text-fg-muted">{selectedNode.description}</p>

            <div className="min-h-0 flex-1 overflow-auto">
              {selectedNode.content ? (
                <CodeBlock code={selectedNode.content} label={selectedNode.name} />
              ) : (
                <EmptyState
                  icon={Info}
                  title="No file preview here"
                  description="This directory holds the subfolders and files listed in the tree. Select a nested file to preview its Markdown or YAML source."
                />
              )}
            </div>

            {selectedNode.standaloneIncluded && (
              <Callout tone="scope" className="mt-3">
                Included in Standalone Team OS mode — this directory operates identically with or without
                the company and platform layers.
              </Callout>
            )}
          </Card>
        </div>
      </Section>
    </PageShell>
  );
};
