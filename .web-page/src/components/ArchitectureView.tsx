import React, { useState } from 'react';
import { WorkspaceNode } from '../types';
import { COMPANY_OS_WORKSPACE_TREE } from '../data/workspaceData';
import { Folder, FileText, ChevronRight, ChevronDown, Copy, Check, Info, ShieldAlert, Sparkles, Layers } from 'lucide-react';

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
    </div>
  );
};
