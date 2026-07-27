import React, { useState } from 'react';
import { WorkspaceNode } from '../types';
import { COMPANY_OS_WORKSPACE_TREE } from '../data/workspaceData';
import {
  FEDERATED_SOURCES,
  FEDERATION_MANIFEST_SAMPLE,
  FEDERATION_SYNC_SAMPLE,
  FEDERATION_LOCK_SAMPLE,
  FEDERATION_RULES,
  MCP_BOUNDARY,
  MCP_SYNC_GIT_VERBS
} from '../data/federationData';
import { Folder, FileText, ChevronRight, ChevronDown, Copy, Check, Info, ShieldAlert, Sparkles, Layers, GitBranch, Lock, ArrowRight, Github, Plug, Eye, PenLine, Ban } from 'lucide-react';

const MCP_COLUMN_STYLE: Record<string, { wrap: string; head: string; icon: React.ReactNode }> = {
  reads: {
    wrap: 'bg-emerald-50/60 border-emerald-200',
    head: 'text-emerald-900',
    icon: <Eye className="w-4 h-4 text-emerald-600 shrink-0" />
  },
  writes: {
    wrap: 'bg-indigo-50/60 border-indigo-200',
    head: 'text-indigo-900',
    icon: <PenLine className="w-4 h-4 text-indigo-600 shrink-0" />
  },
  never: {
    wrap: 'bg-rose-50/60 border-rose-200',
    head: 'text-rose-900',
    icon: <Ban className="w-4 h-4 text-rose-600 shrink-0" />
  }
};

type FederationTab = 'manifest' | 'sync' | 'lock';

const FEDERATION_TABS: { id: FederationTab; label: string; caption: string; source: string }[] = [
  {
    id: 'manifest',
    label: 'workspace.yaml',
    caption: 'The composition you author by hand — five repos, six destinations.',
    source: FEDERATION_MANIFEST_SAMPLE
  },
  {
    id: 'sync',
    label: 'sync + status',
    caption: 'What the CLI prints when it materializes those repos into one tree.',
    source: FEDERATION_SYNC_SAMPLE
  },
  {
    id: 'lock',
    label: 'workspace.lock.yaml',
    caption: 'The machine-owned proof of what was synced — gate [8/8] reads this.',
    source: FEDERATION_LOCK_SAMPLE
  }
];

interface ArchitectureViewProps {
  isStandalone: boolean;
  setIsStandalone: (val: boolean) => void;
}

export const ArchitectureView: React.FC<ArchitectureViewProps> = ({ isStandalone, setIsStandalone }) => {
  const [selectedNode, setSelectedNode] = useState<WorkspaceNode>(
    COMPANY_OS_WORKSPACE_TREE.children![1] // default to company-os/
  );
  const [expandedFolders, setExpandedFolders] = useState<Record<string, boolean>>({
    'root': true,
    'company-os-folder': true,
    'platforms-folder': true,
    'ordering-platform': true,
    'teams-folder': true,
    'web-team': true,
    'ontology-folder': true,
    'knowledge-folder': true,
  });
  const [copied, setCopied] = useState(false);
  const [federationTab, setFederationTab] = useState<FederationTab>('manifest');
  const [activeSource, setActiveSource] = useState<string | null>(null);

  const toggleFolder = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setExpandedFolders(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const renderTree = (node: WorkspaceNode) => {
    // If standalone mode, filter node
    if (isStandalone && node.standaloneIncluded === false) {
      return null;
    }

    const isFolder = node.type === 'folder';
    const isExpanded = expandedFolders[node.id];
    const isSelected = selectedNode.id === node.id;

    const layerBadgeColor: Record<string, string> = {
      'company-os': 'bg-blue-50 text-blue-700 border-blue-200',
      'platforms': 'bg-purple-50 text-purple-700 border-purple-200',
      'teams': 'bg-emerald-50 text-emerald-700 border-emerald-200',
      'company-ontology': 'bg-amber-50 text-amber-800 border-amber-200',
      'knowledge': 'bg-cyan-50 text-cyan-700 border-cyan-200',
      'root': 'bg-slate-100 text-slate-700 border-slate-300'
    };

    return (
      <div key={node.id} className="select-none my-0.5">
        <div
          onClick={() => setSelectedNode(node)}
          className={`flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs cursor-pointer transition-colors ${
            isSelected
              ? 'bg-indigo-50 text-indigo-900 border border-indigo-300 font-semibold'
              : 'hover:bg-slate-100 text-slate-700'
          }`}
        >
          <div className="flex items-center gap-1.5 overflow-hidden">
            {isFolder ? (
              <button
                onClick={(e) => toggleFolder(node.id, e)}
                className="p-0.5 hover:bg-slate-200 rounded text-slate-500"
              >
                {isExpanded ? (
                  <ChevronDown className="w-3.5 h-3.5" />
                ) : (
                  <ChevronRight className="w-3.5 h-3.5" />
                )}
              </button>
            ) : (
              <span className="w-4" />
            )}

            {isFolder ? (
              <Folder className={`w-4 h-4 shrink-0 ${isSelected ? 'text-indigo-600' : 'text-amber-500'}`} />
            ) : (
              <FileText className={`w-4 h-4 shrink-0 ${isSelected ? 'text-indigo-600' : 'text-slate-400'}`} />
            )}

            <span className="font-mono truncate">{node.name}</span>
          </div>

          <span className={`text-[10px] px-1.5 py-0.5 rounded border uppercase tracking-wider font-mono ${layerBadgeColor[node.layer] || 'bg-slate-100 text-slate-600 border-slate-200'}`}>
            {node.layer}
          </span>
        </div>

        {isFolder && isExpanded && node.children && (
          <div className="ml-4 pl-2 border-l border-slate-200 space-y-0.5">
            {node.children.map(child => renderTree(child))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-6">
      {/* Intro Bento Hero Card */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm relative overflow-hidden space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 relative z-10">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                01 ARCHITECTURE
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                UNIFIED DUAL-CORE
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">
              Workspace Directory & Folder Explorer
            </h2>
          </div>

          <div className="flex items-center gap-2 bg-white p-1.5 rounded-xl border border-slate-200 shadow-sm shrink-0 text-xs">
            <span className="text-slate-500 px-2 font-medium">Layout Mode:</span>
            <button
              onClick={() => setIsStandalone(false)}
              className={`px-3 py-1 rounded-lg font-semibold transition-all ${!isStandalone ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Full Company OS
            </button>
            <button
              onClick={() => setIsStandalone(true)}
              className={`px-3 py-1 rounded-lg font-semibold transition-all ${isStandalone ? 'bg-cyan-600 text-white shadow-md shadow-cyan-600/20' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Team OS Only
            </button>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              To give every file a standardized location so team members never ask "Where is the latest document?"
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              The left panel shows the folder tree. The right panel displays the specs and rules for whatever file you select.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Click any folder or file on the left (e.g. <code>prds/</code> or <code>company-os.yaml</code>) to inspect its exact contents!
            </p>
          </div>
        </div>
      </div>

      {/* Main Split View: Tree Explorer vs File Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        
        {/* Left: Tree Explorer */}
        <div className="lg:col-span-5 bg-white border border-slate-200 shadow-sm rounded-2xl p-4 flex flex-col h-[650px]">
          <div className="flex items-center justify-between pb-3 mb-3 border-b border-slate-100">
            <div className="flex items-center gap-2 text-xs font-bold text-slate-800">
              <Layers className="w-4 h-4 text-indigo-600" />
              <span>Workspace Directory Tree</span>
            </div>
            <span className="text-[11px] text-slate-500 font-mono">
              {isStandalone ? 'teams/ solo' : '/ (root)'}
            </span>
          </div>

          <div className="flex-1 overflow-y-auto pr-1 no-scrollbar">
            {renderTree(COMPANY_OS_WORKSPACE_TREE)}
          </div>

          {/* Quick legend footer */}
          <div className="mt-3 pt-3 border-t border-slate-100 flex flex-wrap gap-2 text-[10px]">
            <span className="px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200">company-os</span>
            <span className="px-1.5 py-0.5 rounded bg-purple-50 text-purple-700 border border-purple-200">platforms</span>
            <span className="px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200">teams</span>
            <span className="px-1.5 py-0.5 rounded bg-amber-50 text-amber-800 border border-amber-200">ontology</span>
            <span className="px-1.5 py-0.5 rounded bg-cyan-50 text-cyan-700 border border-cyan-200">knowledge</span>
          </div>
        </div>

        {/* Right: Selected File Inspector */}
        <div className="lg:col-span-7 bg-white border border-slate-200 shadow-sm rounded-2xl p-5 flex flex-col h-[650px]">
          <div className="flex items-center justify-between pb-3 border-b border-slate-100">
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono px-2 py-0.5 rounded bg-indigo-50 text-indigo-700 border border-indigo-200 font-semibold">
                  {selectedNode.type.toUpperCase()}
                </span>
                <h3 className="text-sm font-bold text-slate-900 font-mono">{selectedNode.path}</h3>
              </div>
              <p className="text-xs text-slate-500 mt-1">{selectedNode.description}</p>
            </div>

            {selectedNode.content && (
              <button
                onClick={() => handleCopy(selectedNode.content!)}
                className="flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium border border-slate-200 transition-colors"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copied ? 'Copied' : 'Copy'}</span>
              </button>
            )}
          </div>

          {/* Details & Metadata Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 my-3 text-xs">
            <div className="bg-slate-50 p-2.5 rounded-xl border border-slate-200">
              <span className="text-slate-500 block text-[10px] uppercase font-bold tracking-wider">Written / Scaffolded By</span>
              <span className="font-semibold text-slate-800 mt-0.5 block">{selectedNode.writtenBy}</span>
            </div>

            <div className="bg-slate-50 p-2.5 rounded-xl border border-slate-200">
              <span className="text-slate-500 block text-[10px] uppercase font-bold tracking-wider">Validator Gate Check</span>
              <span className="font-semibold text-amber-800 mt-0.5 block">
                {selectedNode.validatorCheck || 'Standard frontmatter & structure sanity check'}
              </span>
            </div>
          </div>

          {/* Code/Content Preview */}
          <div className="flex-1 bg-slate-900 rounded-xl border border-slate-800 p-4 overflow-y-auto font-mono text-xs text-slate-200 relative">
            {selectedNode.content ? (
              <pre className="whitespace-pre-wrap leading-relaxed">{selectedNode.content}</pre>
            ) : (
              <div className="h-full flex flex-col items-center justify-center text-slate-400 gap-2">
                <Info className="w-6 h-6 text-slate-500" />
                <span>This directory contains subfolders/files listed on the left tree menu.</span>
                <span className="text-[11px] text-slate-500">Select any nested file to preview its Markdown or YAML source.</span>
              </div>
            )}
          </div>

          {/* Standalone note */}
          {selectedNode.standaloneIncluded && (
            <div className="mt-3 p-2.5 rounded-xl bg-emerald-50 border border-emerald-200 flex items-center gap-2 text-xs text-emerald-800 font-medium">
              <ShieldAlert className="w-4 h-4 shrink-0 text-emerald-600" />
              <span>Included in Standalone Team OS mode. This directory operates identically with or without company/platform layers.</span>
            </div>
          )}
        </div>

      </div>

      {/* ============================================================= */}
      {/* Federated Mode: one tree, many repos                          */}
      {/* ============================================================= */}
      <div className="bg-white border border-slate-200 shadow-sm rounded-2xl overflow-hidden">

        {/* Section header + the core explanation */}
        <div className="bg-gradient-to-br from-cyan-50 via-white to-slate-50 p-6 border-b border-slate-200 space-y-4">
          <div className="flex items-center gap-2">
            <span className="text-[10px] font-bold text-cyan-800 uppercase tracking-widest font-mono">
              01b FEDERATED MODE
            </span>
            <span className="px-2 py-0.5 bg-cyan-100/70 rounded-full text-[10px] font-semibold text-cyan-900 border border-cyan-200">
              MULTI-REPO
            </span>
          </div>

          <div className="flex items-start gap-3">
            <GitBranch className="w-5 h-5 text-cyan-700 shrink-0 mt-0.5" />
            <div className="space-y-2">
              <h3 className="text-lg font-bold text-slate-900 tracking-tight">
                When the local directory comes from several GitHub repositories
              </h3>
              <p className="text-sm text-slate-600 leading-relaxed max-w-4xl">
                Everything in the explorer above assumed <strong>one repo</strong> — a monorepo where every
                folder is hand-edited in place. That is not how most companies are shaped. Governance lives in
                a governance repo, each platform ships on its own release train, and product documentation
                lives with the product code.
              </p>
              <p className="text-sm text-slate-600 leading-relaxed max-w-4xl">
                Federated mode composes <em>the exact same directory tree</em> out of many pinned Git
                repositories. You author one <code className="font-mono text-[12px] px-1 py-0.5 rounded bg-slate-100 border border-slate-200">workspace.yaml</code>,
                run <code className="font-mono text-[12px] px-1 py-0.5 rounded bg-slate-100 border border-slate-200">company-os workspace sync</code>,
                and the CLI clones each source, copies only the allow-listed directories, and drops them
                read-only at the destination you named. No submodules, no copy-paste.
              </p>
              <p className="text-sm text-slate-800 leading-relaxed max-w-4xl font-medium bg-white/70 border border-cyan-200 rounded-xl p-3">
                The switch is the file's existence, nothing else. No{' '}
                <code className="font-mono text-[12px]">workspace.yaml</code> ⇒ monorepo mode and{' '}
                <code className="font-mono text-[12px]">validate</code> runs 7 gates. Add the manifest and an
                eighth gate appears: <strong>[8/8] federated slice integrity</strong>. Every command you
                learned above behaves identically either way — the directory layout is the contract, the
                repo boundary is not.
              </p>
            </div>
          </div>
        </div>

        {/* Source → destination map */}
        <div className="p-6 border-b border-slate-100 space-y-3">
          <div className="flex items-center justify-between flex-wrap gap-2">
            <div className="flex items-center gap-2 text-xs font-bold text-slate-800">
              <Github className="w-4 h-4 text-slate-700" />
              <span>Where each directory actually comes from</span>
            </div>
            <span className="text-[11px] text-slate-500 font-mono">5 repos → 6 local directories</span>
          </div>
          <p className="text-xs text-slate-500 leading-relaxed">
            Click a row to see why that repo is separate. Note the last one: a single repo can feed{' '}
            <strong>several</strong> destinations at once.
          </p>

          <div className="space-y-2">
            {FEDERATED_SOURCES.map(src => {
              const isOpen = activeSource === src.id;
              return (
                <div
                  key={src.id}
                  onClick={() => setActiveSource(isOpen ? null : src.id)}
                  className={`rounded-xl border cursor-pointer transition-colors ${
                    isOpen ? 'border-cyan-300 bg-cyan-50/50' : 'border-slate-200 bg-white hover:bg-slate-50'
                  }`}
                >
                  <div className="p-3 flex flex-col lg:flex-row lg:items-center gap-3 text-xs">
                    {/* Source repo */}
                    <div className="lg:w-[38%] min-w-0">
                      <div className="flex items-center gap-1.5">
                        <Github className="w-3.5 h-3.5 text-slate-500 shrink-0" />
                        <span className="font-mono font-semibold text-slate-800 truncate">{src.repoUrl}</span>
                      </div>
                      <div className="flex items-center gap-1.5 mt-1 pl-5 flex-wrap">
                        <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-900 text-slate-100">
                          {src.pin}
                        </span>
                        <span className="text-[10px] text-slate-500 font-mono truncate">
                          paths: {src.paths.join(' ')}
                        </span>
                      </div>
                    </div>

                    <ArrowRight className="w-4 h-4 text-cyan-600 shrink-0 hidden lg:block" />

                    {/* Destinations */}
                    <div className="flex-1 min-w-0 space-y-1">
                      {src.targets.map(t => (
                        <div key={t} className="flex items-center gap-1.5">
                          <Lock className="w-3 h-3 text-amber-600 shrink-0" />
                          <span className="font-mono text-slate-800 truncate">{t}</span>
                          <span className={`text-[10px] px-1.5 py-0.5 rounded border uppercase tracking-wider font-mono shrink-0 ${
                            {
                              'company-os': 'bg-blue-50 text-blue-700 border-blue-200',
                              'platforms': 'bg-purple-50 text-purple-700 border-purple-200',
                              'teams': 'bg-emerald-50 text-emerald-700 border-emerald-200',
                              'company-ontology': 'bg-amber-50 text-amber-800 border-amber-200',
                              'knowledge': 'bg-cyan-50 text-cyan-700 border-cyan-200'
                            }[src.layer]
                          }`}>
                            {src.layer}
                          </span>
                        </div>
                      ))}
                    </div>

                    <ChevronDown
                      className={`w-4 h-4 text-slate-400 shrink-0 transition-transform ${isOpen ? 'rotate-180' : ''}`}
                    />
                  </div>

                  {isOpen && (
                    <div className="px-3 pb-3 pt-0 text-[11px] text-slate-600 leading-relaxed border-t border-cyan-200/70 mt-1 pl-3">
                      <p className="pt-2">{src.note}</p>
                      <p className="mt-1 text-slate-500">
                        Owned by <strong className="text-slate-700">{src.owner}</strong> — they merge and tag
                        in that repo; this workspace only ever reads the pinned result.
                      </p>
                    </div>
                  )}
                </div>
              );
            })}

            {/* The native counter-example */}
            <div className="rounded-xl border border-dashed border-emerald-300 bg-emerald-50/40 p-3 flex items-start gap-2 text-xs">
              <FileText className="w-3.5 h-3.5 text-emerald-700 shrink-0 mt-0.5" />
              <p className="text-emerald-900 leading-relaxed">
                <code className="font-mono font-semibold">teams/web/</code> has no source repo —{' '}
                <strong>it is native</strong>. This workspace owns it, humans edit it, and{' '}
                <code className="font-mono">governance resolve</code> writes its{' '}
                <code className="font-mono">generated/</code> folder. A federated workspace is a thin
                composition repo: your own team directory, plus read-only slices of everyone else's.
              </p>
            </div>
          </div>
        </div>

        {/* Code samples */}
        <div className="p-6 border-b border-slate-100 space-y-3">
          <div className="flex items-center gap-2 flex-wrap">
            {FEDERATION_TABS.map(tab => (
              <button
                key={tab.id}
                onClick={() => setFederationTab(tab.id)}
                className={`px-3 py-1.5 rounded-lg text-xs font-mono font-semibold transition-all border ${
                  federationTab === tab.id
                    ? 'bg-slate-900 text-white border-slate-900 shadow-sm'
                    : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-50'
                }`}
              >
                {tab.label}
              </button>
            ))}
            <span className="text-[11px] text-slate-500 ml-1">
              {FEDERATION_TABS.find(t => t.id === federationTab)!.caption}
            </span>
          </div>

          <div className="relative">
            <button
              onClick={() => handleCopy(FEDERATION_TABS.find(t => t.id === federationTab)!.source)}
              className="absolute top-3 right-3 z-10 flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 font-medium border border-slate-700 transition-colors"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copied ? 'Copied' : 'Copy'}</span>
            </button>
            <div className="bg-slate-900 rounded-xl border border-slate-800 p-4 overflow-auto max-h-[460px] font-mono text-xs text-slate-200">
              <pre className="whitespace-pre leading-relaxed">
                {FEDERATION_TABS.find(t => t.id === federationTab)!.source}
              </pre>
            </div>
          </div>
        </div>

        {/* Rules that bite */}
        <div className="p-6 space-y-3">
          <div className="flex items-center gap-2 text-xs font-bold text-slate-800">
            <ShieldAlert className="w-4 h-4 text-amber-600" />
            <span>Six rules the CLI enforces (and what happens when you break them)</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {FEDERATION_RULES.map(r => (
              <div key={r.id} className="bg-slate-50 border border-slate-200 rounded-xl p-3 space-y-1.5 text-xs">
                <div className="font-semibold text-slate-900 leading-snug">{r.rule}</div>
                <p className="text-[11px] text-slate-600 leading-relaxed">{r.why}</p>
                <p className="text-[11px] text-amber-900 bg-amber-50 border border-amber-200 rounded-lg p-2 leading-relaxed">
                  {r.breaksWith}
                </p>
              </div>
            ))}
          </div>

          <div className="mt-1 p-3 rounded-xl bg-indigo-50 border border-indigo-200 flex items-start gap-2 text-xs text-indigo-900">
            <Sparkles className="w-4 h-4 shrink-0 text-indigo-600 mt-0.5" />
            <span className="leading-relaxed">
              <strong>The day-2 loop is four steps:</strong>{' '}
              <code className="font-mono">bump the pin</code> →{' '}
              <code className="font-mono">workspace sync</code> →{' '}
              <code className="font-mono">governance resolve --team &lt;t&gt;</code> →{' '}
              commit the manifest, the slice, the lock and the regenerated{' '}
              <code className="font-mono">generated/</code> in one PR. The lock diff is your audit trail of
              which governance SHAs moved.
            </span>
          </div>
        </div>

        {/* Agent tooling boundary: where the GitHub MCP fits */}
        <div className="p-6 bg-slate-50/70 border-t border-slate-200 space-y-4">
          <div className="flex items-start gap-3">
            <Plug className="w-5 h-5 text-slate-700 shrink-0 mt-0.5" />
            <div className="space-y-2">
              <h4 className="text-sm font-bold text-slate-900">
                Where the GitHub MCP fits — and where it must not
              </h4>
              <p className="text-xs text-slate-600 leading-relaxed max-w-4xl">
                Since these directories come from GitHub, the obvious question is whether an agent should
                drive the sync through the <strong>GitHub MCP server</strong>. The answer is a boundary, not
                a yes or no. <strong>Company OS ships no MCP server and no MCP client</strong> — there is no{' '}
                <code className="font-mono text-[11px] px-1 rounded bg-slate-200/70">.mcp.json</code> anywhere
                in the repo, and no <code className="font-mono text-[11px]">company-os</code> command talks to
                the GitHub API. The MCP server is configured in <em>your agent</em>, and it enters this
                workflow through exactly one door: <strong>a skill</strong>. A skill tells an agent how to
                author an artifact — it can never grant permission to bypass a gate.
              </p>
            </div>
          </div>

          {/* The eight git verbs */}
          <div className="bg-slate-900 rounded-xl border border-slate-800 p-3.5 space-y-1.5">
            <span className="text-[10px] font-mono uppercase tracking-wider text-slate-400">
              The whole of workspace sync — no push, no commit, no merge, no HTTP API call
            </span>
            <pre className="font-mono text-[11px] text-slate-200 leading-relaxed overflow-x-auto">
              {MCP_SYNC_GIT_VERBS.join('\n')}
            </pre>
          </div>

          {/* Read / write / never */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
            {MCP_BOUNDARY.map(col => {
              const style = MCP_COLUMN_STYLE[col.id];
              return (
                <div key={col.id} className={`rounded-xl border p-3.5 space-y-2 ${style.wrap}`}>
                  <div className={`flex items-center gap-1.5 text-xs font-bold ${style.head}`}>
                    {style.icon}
                    <span>{col.title}</span>
                  </div>
                  <p className="text-[11px] text-slate-600 leading-relaxed">{col.subtitle}</p>
                  <ul className="space-y-1.5">
                    {col.items.map((item, i) => (
                      <li key={i} className="text-[11px] text-slate-700 leading-relaxed flex gap-1.5">
                        <span className="text-slate-400 shrink-0">•</span>
                        <span>{item}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              );
            })}
          </div>

          <div className="p-3 rounded-xl bg-slate-900 border border-slate-800 flex items-start gap-2 text-xs text-slate-200">
            <ShieldAlert className="w-4 h-4 shrink-0 text-amber-400 mt-0.5" />
            <span className="leading-relaxed">
              <strong className="text-white">One rule keeps this coherent:</strong> an agent may read
              anything, but the CLI writes the workspace and the lock is the oracle. An MCP server that
              respects that boundary is a convenience. One that crosses it produces a workspace whose
              validation results are no longer evidence of anything.
            </span>
          </div>

          <p className="text-[11px] text-slate-500 leading-relaxed">
            The skill-side mechanics — which skill, which step, and the prompt that keeps MCP on the read
            half — are in <strong className="text-slate-700">06 Agent Skills &amp; Local Search</strong>.
          </p>
        </div>
      </div>
    </div>
  );
};
