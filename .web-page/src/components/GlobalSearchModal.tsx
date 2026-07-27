import React, { useState, useEffect } from 'react';
import { 
  Search, 
  X, 
  ArrowRight, 
  FolderTree, 
  Terminal, 
  ShieldCheck, 
  PlayCircle, 
  BookOpen, 
  Award,
  Command,
  FileText,
  Layers,
  Bot
} from 'lucide-react';
import { TabType } from '../types';
import { STATIC_SEARCH_ITEMS, SearchResultItemData } from '../data/searchData';

interface GlobalSearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab: (tab: TabType) => void;
}

export const GlobalSearchModal: React.FC<GlobalSearchModalProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
}) => {
  const [query, setQuery] = useState('');

  const getItemIcon = (iconName: string) => {
    switch (iconName) {
      case 'Layers': return Layers;
      case 'FolderTree': return FolderTree;
      case 'Terminal': return Terminal;
      case 'ShieldCheck': return ShieldCheck;
      case 'PlayCircle': return PlayCircle;
      case 'BookOpen': return BookOpen;
      case 'FileText': return FileText;
      case 'Award': return Award;
      case 'Bot': return Bot;
      default: return Search;
    }
  };

  // Filter static items by query
  const searchResults = query.trim() === ''
    ? STATIC_SEARCH_ITEMS
    : STATIC_SEARCH_ITEMS.filter(item => {
        const q = query.toLowerCase();
        return (
          item.title.toLowerCase().includes(q) ||
          item.snippet.toLowerCase().includes(q) ||
          item.category.toLowerCase().includes(q) ||
          item.keywords.some(k => k.toLowerCase().includes(q))
        );
      });

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-16 sm:pt-24 px-4 bg-slate-900/60 backdrop-blur-sm animate-fadeIn">
      <div 
        className="bg-white rounded-3xl shadow-2xl border border-slate-200 w-full max-w-2xl overflow-hidden flex flex-col max-h-[80vh] text-slate-900"
        onClick={(e) => e.stopPropagation()}
      >
        
        {/* Search Header Input */}
        <div className="p-4 sm:p-5 border-b border-slate-200 flex items-center gap-3 bg-slate-50/80 shrink-0">
          <Search className="w-5 h-5 text-indigo-600 shrink-0" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Type to search all views, commands, gates, rules & quiz..."
            className="w-full bg-transparent text-sm sm:text-base font-medium text-slate-900 placeholder:text-slate-400 focus:outline-none"
            autoFocus
          />
          {query && (
            <button
              onClick={() => setQuery('')}
              className="px-2 py-1 rounded bg-slate-200 text-slate-600 text-xs font-bold hover:bg-slate-300 transition-colors"
            >
              Clear
            </button>
          )}
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-full bg-slate-200/80 hover:bg-slate-300 text-slate-700 flex items-center justify-center transition-colors shrink-0"
            aria-label="Close search"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Results Info Bar */}
        <div className="px-5 py-2.5 bg-indigo-50/60 border-b border-indigo-100/80 text-[11px] text-indigo-900 font-medium flex items-center justify-between shrink-0">
          <span>
            {query.trim() === '' ? 'Showing all indexed topics' : `Found ${searchResults.length} matching topics`}
          </span>
          <span className="font-mono text-slate-500 hidden sm:inline">Press ESC to close</span>
        </div>

        {/* Results List */}
        <div className="p-3 sm:p-4 overflow-y-auto space-y-2 flex-1 no-scrollbar bg-slate-50/30">
          {searchResults.length === 0 ? (
            <div className="py-12 text-center text-slate-500 space-y-2">
              <Search className="w-8 h-8 text-slate-300 mx-auto" />
              <p className="text-sm font-semibold text-slate-700">No static results for "{query}"</p>
              <p className="text-xs text-slate-400">Try searching "PRD", "CLI", "Gate", "Governance", "Quiz" or "Exit code"</p>
            </div>
          ) : (
            searchResults.map((item) => {
              const Icon = getItemIcon(item.iconName);
              return (
                <div
                  key={item.id}
                  onClick={() => {
                    onClose();
                    onNavigateTab(item.targetTab);
                  }}
                  className="p-3.5 rounded-2xl bg-white border border-slate-200 hover:border-indigo-300 hover:bg-indigo-50/50 transition-all cursor-pointer group flex items-start justify-between gap-3 shadow-xs"
                >
                  <div className="flex items-start gap-3">
                    <div className="w-8 h-8 rounded-xl bg-indigo-50 text-indigo-700 flex items-center justify-center shrink-0 mt-0.5">
                      <Icon className="w-4 h-4" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2 mb-0.5">
                        <span className="font-bold text-slate-900 text-xs sm:text-sm group-hover:text-indigo-700 transition-colors">
                          {item.title}
                        </span>
                        <span className="px-2 py-0.5 rounded-md bg-slate-100 text-slate-600 text-[10px] font-bold font-mono">
                          {item.category}
                        </span>
                      </div>
                      <p className="text-slate-600 text-xs leading-relaxed">
                        {item.snippet}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center text-xs font-bold text-indigo-600 shrink-0 self-center group-hover:translate-x-1 transition-transform">
                    <ArrowRight className="w-4 h-4" />
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="p-3 bg-white border-t border-slate-200 text-center text-slate-400 text-[11px] shrink-0">
          Click any result to instantly jump to that tab & tool.
        </div>

      </div>
    </div>
  );
};
