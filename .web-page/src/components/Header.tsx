import React from 'react';
import { GitBranch, Layers, HelpCircle, Eye, Search, ListFilter } from 'lucide-react';
import { Tooltip } from './Tooltip';

interface HeaderProps {
  isStandalone: boolean;
  setIsStandalone: (val: boolean) => void;
  activeTab: string;
  setActiveTab: (tab: any) => void;
  onOpenGuide: () => void;
  onOpenDirectory: () => void;
  onOpenSearch: () => void;
  easyReadMode: boolean;
  setEasyReadMode: (val: boolean) => void;
}

export const Header: React.FC<HeaderProps> = ({
  isStandalone,
  setIsStandalone,
  onOpenGuide,
  onOpenDirectory,
  onOpenSearch,
  easyReadMode,
  setEasyReadMode,
}) => {
  return (
    <header className="bg-white/95 border-b border-slate-200 text-slate-900 sticky top-0 z-50 backdrop-blur-md">
      <div className="max-w-[95%] mx-auto px-2 sm:px-4 lg:px-6">
        <div className="flex flex-col lg:flex-row items-center justify-between py-2.5 gap-2.5">
          
          {/* Logo & Title */}
          <div className="flex items-center space-x-2.5 shrink-0">
            <div className="w-8 h-8 sm:w-9 sm:h-9 bg-indigo-600 rounded-xl flex items-center justify-center font-bold text-white shadow-md shadow-indigo-500/20 shrink-0">
              <span className="font-mono text-xs tracking-tighter">OS</span>
            </div>
            <div>
              <div className="flex items-center gap-1.5">
                <h1 className="text-sm sm:text-base font-bold tracking-tight text-slate-900 whitespace-nowrap">
                  Company OS & Team OS
                </h1>
                <span className="px-2 py-0.5 bg-indigo-50 rounded-full text-[10px] font-semibold text-indigo-700 uppercase tracking-wider border border-indigo-200 shrink-0">
                  v1.4.0
                </span>
              </div>
              <p className="text-[11px] text-slate-500 hidden xl:block whitespace-nowrap">
                Git-based governance, product & engineering operating system
              </p>
            </div>
          </div>

          {/* Search Button (Triggers Global Search Modal) */}
          <div className="w-full lg:w-auto flex-1 max-w-md mx-0 lg:mx-3">
            <Tooltip 
              title="Why click search?" 
              content="Instantly find any topic, CLI terminal command, validation gate, or glossary term across the entire platform."
              position="bottom"
              className="w-full"
            >
              <button
                onClick={onOpenSearch}
                className="w-full flex items-center justify-between px-3 py-1.5 bg-slate-100 hover:bg-slate-200/80 rounded-xl border border-slate-200 text-xs text-slate-500 transition-all shadow-xs group"
              >
                <div className="flex items-center gap-2 min-w-0 pr-2">
                  <Search className="w-3.5 h-3.5 text-slate-400 group-hover:text-indigo-600 transition-colors shrink-0" />
                  <span className="truncate text-slate-600 font-medium">Search topics, commands, gates & rules...</span>
                </div>
                <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] font-mono bg-white text-slate-500 rounded border border-slate-200 shrink-0">
                  Search
                </kbd>
              </button>
            </Tooltip>
          </div>

          {/* Controls & Directory Modal Buttons */}
          <div className="flex items-center gap-1.5 sm:gap-2 flex-wrap justify-center lg:justify-end shrink-0 w-full lg:w-auto">
            
            {/* Content Directory Modal Button */}
            <Tooltip
              title="Why click directory?"
              content="Open a complete site map and quick index to navigate all 9 main sections, tools, and guides."
              position="bottom"
            >
              <button
                onClick={onOpenDirectory}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold transition-all shadow-sm whitespace-nowrap"
              >
                <ListFilter className="w-3.5 h-3.5 text-indigo-300 shrink-0" />
                <span>Content Directory</span>
              </button>
            </Tooltip>

            {/* Beginner Guide Button */}
            <Tooltip
              title="Why click glossary?"
              content="Read simple, jargon-free explanations for technical concepts like canonical, PRDs, and gates."
              position="bottom"
            >
              <button
                onClick={onOpenGuide}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-50 hover:bg-indigo-100 text-indigo-800 text-xs font-bold border border-indigo-200 transition-all whitespace-nowrap"
              >
                <HelpCircle className="w-3.5 h-3.5 text-indigo-600 shrink-0" />
                <span>Glossary</span>
              </button>
            </Tooltip>

            {/* Easy Read Mode Toggle */}
            <Tooltip
              title="Why click easy read?"
              content="Enlarge typography and spacing across the entire app for high legibility and comfort."
              position="bottom"
            >
              <button
                onClick={() => setEasyReadMode(!easyReadMode)}
                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl text-xs font-semibold border transition-all whitespace-nowrap ${
                  easyReadMode
                    ? 'bg-amber-100 text-amber-900 border-amber-300'
                    : 'bg-slate-100 text-slate-700 border-slate-200 hover:bg-slate-200/70'
                }`}
              >
                <Eye className="w-3.5 h-3.5 text-amber-700 shrink-0" />
                <span className="hidden sm:inline">{easyReadMode ? 'Easy Read: ON' : 'Easy Read'}</span>
              </button>
            </Tooltip>

            {/* Mode Switcher */}
            <Tooltip
              title="Why switch scope?"
              content="Toggle between Company OS (company-wide canonical view) and Team OS (single team workspace)."
              position="bottom"
            >
              <div className="bg-slate-100 p-1 rounded-xl border border-slate-200 flex items-center text-xs shrink-0">
                <button
                  onClick={() => setIsStandalone(false)}
                  className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 whitespace-nowrap ${
                    !isStandalone
                      ? 'bg-indigo-600 text-white shadow-xs'
                      : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  <Layers className="w-3 h-3 shrink-0" />
                  <span className="hidden sm:inline">Company OS</span>
                  <span className="sm:hidden">Company</span>
                </button>
                <button
                  onClick={() => setIsStandalone(true)}
                  className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 whitespace-nowrap ${
                    isStandalone
                      ? 'bg-cyan-600 text-white shadow-xs'
                      : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  <GitBranch className="w-3 h-3 shrink-0" />
                  <span className="hidden sm:inline">Team OS</span>
                  <span className="sm:hidden">Team</span>
                </button>
              </div>
            </Tooltip>

          </div>

        </div>
      </div>
    </header>
  );
};


