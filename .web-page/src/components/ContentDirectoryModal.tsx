import React, { useState } from 'react';
import { 
  X, 
  Search, 
  Home,
  Download,
  FolderTree,
  Terminal,
  PlayCircle, 
  ShieldCheck, 
  CheckSquare, 
  BookOpen, 
  Award, 
  ChevronRight,
  ArrowRight,
  Bot
} from 'lucide-react';
import { TabType } from '../types';
import { DIRECTORY_CATEGORIES_DATA, DirectoryCategoryData } from '../data/directoryData';

interface ContentDirectoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab: (tab: TabType) => void;
}

export const ContentDirectoryModal: React.FC<ContentDirectoryModalProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
}) => {
  const [searchTerm, setSearchTerm] = useState('');

  if (!isOpen) return null;

  const getCategoryIcon = (iconName: string) => {
    switch (iconName) {
      case 'Home': return Home;
      case 'Download': return Download;
      case 'FolderTree': return FolderTree;
      case 'Terminal': return Terminal;
      case 'PlayCircle': return PlayCircle;
      case 'ShieldCheck': return ShieldCheck;
      case 'CheckSquare': return CheckSquare;
      case 'Award': return Award;
      case 'Bot': return Bot;
      default: return FolderTree;
    }
  };

  // Filter items based on search term
  const filteredCategories = DIRECTORY_CATEGORIES_DATA.map(cat => {
    const matchingItems = cat.items.filter(item => 
      item.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      item.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      cat.categoryName.toLowerCase().includes(searchTerm.toLowerCase())
    );
    return { ...cat, items: matchingItems };
  }).filter(cat => cat.items.length > 0);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-indigo-950/60 backdrop-blur-sm animate-fadeIn">
      <div 
        className="bg-white rounded-3xl shadow-2xl border border-slate-200 w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden text-slate-900"
        onClick={(e) => e.stopPropagation()}
      >
        
        {/* Header */}
        <div className="bg-gradient-to-r from-indigo-900 via-indigo-800 to-indigo-950 text-white p-5 sm:p-6 flex items-start justify-between gap-4 shrink-0">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="px-2.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-200 border border-indigo-400/30 text-[11px] font-bold uppercase font-mono tracking-wider">
                CONTENT DIRECTORY & SITE MAP
              </span>
              <span className="text-xs text-indigo-300 hidden sm:inline">• Full Catalog Index</span>
            </div>
            <h2 className="text-xl sm:text-2xl font-black tracking-tight text-white flex items-center gap-2">
              <BookOpen className="w-6 h-6 text-indigo-400" />
              <span>Company OS Content Directory</span>
            </h2>
            <p className="text-xs sm:text-sm text-indigo-100 max-w-2xl">
              Browse every section, tool, guide, command, and safety gate in one clear directory view.
            </p>
          </div>

          <button
            onClick={onClose}
            className="w-9 h-9 rounded-full bg-white/10 hover:bg-white/20 text-white flex items-center justify-center transition-colors shrink-0"
            aria-label="Close modal"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Static Search Filter Bar inside Directory */}
        <div className="p-4 bg-slate-50 border-b border-slate-200 shrink-0">
          <div className="relative max-w-xl mx-auto">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search content directory (e.g., PRD, Validate, Gates, Rules, Quiz)..."
              className="w-full pl-10 pr-4 py-2 bg-white rounded-xl border border-slate-300 text-xs sm:text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 shadow-sm"
              autoFocus
            />
            {searchTerm && (
              <button
                onClick={() => setSearchTerm('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-400 hover:text-slate-600 bg-slate-100 px-1.5 py-0.5 rounded"
              >
                Clear
              </button>
            )}
          </div>
        </div>

        {/* Directory Body Content */}
        <div className="p-5 sm:p-6 overflow-y-auto space-y-6 no-scrollbar flex-1 bg-slate-50/50">
          {filteredCategories.length === 0 ? (
            <div className="py-12 text-center text-slate-500 space-y-3">
              <Search className="w-10 h-10 text-slate-300 mx-auto" />
              <p className="text-sm font-semibold text-slate-700">No content directory items found matching "{searchTerm}"</p>
              <button
                onClick={() => setSearchTerm('')}
                className="text-xs text-indigo-600 hover:underline font-bold"
              >
                Reset search filter
              </button>
            </div>
          ) : (
            filteredCategories.map((cat, idx) => {
              const CategoryIcon = getCategoryIcon(cat.iconName);
              return (
                <div key={idx} className="bg-white border border-slate-200 rounded-2xl p-4 sm:p-5 shadow-sm space-y-3">
                  {/* Category Title */}
                  <div className="flex items-center gap-2.5 pb-2 border-b border-slate-100">
                    <div className="w-8 h-8 rounded-xl bg-indigo-50 text-indigo-700 flex items-center justify-center shrink-0">
                      <CategoryIcon className="w-4 h-4" />
                    </div>
                    <div>
                      <h3 className="font-bold text-slate-900 text-sm sm:text-base tracking-tight">
                        {cat.categoryName}
                      </h3>
                      <p className="text-xs text-slate-500">{cat.description}</p>
                    </div>
                  </div>

                  {/* Category Items Grid */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
                    {cat.items.map((item, itemIdx) => (
                      <div
                        key={itemIdx}
                        onClick={() => {
                          onClose();
                          onNavigateTab(item.targetTab);
                        }}
                        className="p-3.5 rounded-xl border border-slate-200 hover:border-indigo-300 hover:bg-indigo-50/40 transition-all cursor-pointer group flex flex-col justify-between space-y-2"
                      >
                        <div>
                          <div className="flex items-center justify-between gap-2 mb-1">
                            <span className="font-bold text-slate-900 text-xs sm:text-sm group-hover:text-indigo-700 transition-colors">
                              {item.title}
                            </span>
                            {item.tag && (
                              <span className="px-2 py-0.5 rounded-full bg-slate-100 text-slate-700 text-[11px] font-bold shrink-0 border border-slate-200">
                                {item.tag}
                              </span>
                            )}
                          </div>
                          <p className="text-slate-600 text-xs leading-relaxed">
                            {item.description}
                          </p>
                        </div>

                        <div className="flex items-center text-xs font-bold text-indigo-600 group-hover:text-indigo-800 transition-colors pt-1">
                          <span>Jump to item</span>
                          <ArrowRight className="w-3.5 h-3.5 ml-1 transition-transform group-hover:translate-x-1" />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Modal Footer */}
        <div className="p-4 bg-white border-t border-slate-200 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs shrink-0">
          <span className="text-slate-500 text-center sm:text-left">
            Tip: You can reopen this Directory anytime from the header menu.
          </span>
          <button
            onClick={onClose}
            className="px-5 py-2 rounded-xl bg-indigo-950 hover:bg-indigo-900 text-white font-bold text-xs transition-colors shadow-sm"
          >
            Close Directory
          </button>
        </div>

      </div>
    </div>
  );
};
