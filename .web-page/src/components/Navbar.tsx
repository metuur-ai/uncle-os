import React, { useRef, useState, useEffect } from 'react';
import { TabType } from '../types';
import { Tooltip } from './Tooltip';
import { 
  Home,
  FolderTree, 
  Terminal, 
  PlayCircle, 
  ShieldCheck, 
  CheckSquare, 
  Search, 
  BookOpen, 
  Award,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';

interface NavbarProps {
  activeTab: TabType;
  setActiveTab: (tab: TabType) => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activeTab, setActiveTab }) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(false);

  const tabs: { 
    id: TabType; 
    label: string; 
    icon: React.FC<{ className?: string }>;
    whyText: string;
  }[] = [
    { 
      id: 'home', 
      label: 'Index & Overview', 
      icon: Home,
      whyText: 'Start here for a high-level summary of how Company OS connects product, engineering, and governance.'
    },
    { 
      id: 'architecture', 
      label: 'Workspace Architecture', 
      icon: FolderTree,
      whyText: 'Explore the workspace directory structure, file hierarchy, and folder layouts for Company OS and Team OS.'
    },
    { 
      id: 'cli', 
      label: 'CLI Terminal Explorer', 
      icon: Terminal,
      whyText: 'Run interactive terminal commands like validate, graph build, and prd new in a live CLI simulator.'
    },
    { 
      id: 'workflows', 
      label: 'Interactive Workflows', 
      icon: PlayCircle,
      whyText: 'Simulate a complete PRD lifecycle step-by-step from discovery brief to production release.'
    },
    { 
      id: 'governance', 
      label: 'Governance Tiers', 
      icon: ShieldCheck,
      whyText: 'Learn how Canonical, Team, and Personal rules coordinate without confusion or broken builds.'
    },
    { 
      id: 'validation', 
      label: 'Validation Gates (1-8)', 
      icon: CheckSquare,
      whyText: 'Inspect automated compliance gates (1-8) that verify your PRDs, dependencies, and workspace rules.'
    },
    { 
      id: 'search-agent', 
      label: 'Local Search & BM25', 
      icon: Search,
      whyText: 'Search offline workspace docs and discover AI Agent Skills for authoring standardized artifacts.'
    },
    { 
      id: 'reference', 
      label: 'Config & Matrix', 
      icon: BookOpen,
      whyText: 'Compare YAML configurations, schema rules, and precedence resolution matrices side-by-side.'
    },
    { 
      id: 'quiz', 
      label: 'Mastery Check', 
      icon: Award,
      whyText: 'Test your knowledge with an interactive 10-question quiz and earn your Company OS certificate!'
    },
  ];

  const checkScrollability = () => {
    if (scrollRef.current) {
      const { scrollLeft, scrollWidth, clientWidth } = scrollRef.current;
      setCanScrollLeft(scrollLeft > 4);
      setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 4);
    }
  };

  useEffect(() => {
    checkScrollability();
    window.addEventListener('resize', checkScrollability);
    return () => window.removeEventListener('resize', checkScrollability);
  }, []);

  const handleScroll = (direction: 'left' | 'right') => {
    if (scrollRef.current) {
      const scrollAmount = direction === 'left' ? -260 : 260;
      scrollRef.current.scrollBy({ left: scrollAmount, behavior: 'smooth' });
    }
  };

  return (
    <nav className="bg-white/90 border-b border-slate-200 sticky top-[56px] z-40 backdrop-blur-md">
      <div className="max-w-[95%] mx-auto px-4 sm:px-6 lg:px-8 relative">
        <div className="relative flex items-center group py-2">
          
          {/* Left Scroll Chevron Button */}
          {canScrollLeft && (
            <div className="absolute left-0 top-0 bottom-0 z-10 flex items-center pr-6 bg-gradient-to-r from-white via-white/90 to-transparent">
              <Tooltip content="Scroll left through tabs" position="right">
                <button
                  onClick={() => handleScroll('left')}
                  aria-label="Scroll left"
                  className="w-8 h-8 rounded-full bg-white border border-slate-200 text-slate-700 hover:text-indigo-600 hover:bg-indigo-50 hover:border-indigo-200 shadow-md flex items-center justify-center transition-all duration-150"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
              </Tooltip>
            </div>
          )}

          {/* Tab List Container */}
          <div
            ref={scrollRef}
            onScroll={checkScrollability}
            className="flex overflow-x-auto space-x-1.5 py-1 no-scrollbar scroll-smooth w-full"
          >
            {tabs.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <Tooltip
                  key={tab.id}
                  title="Why click this?"
                  content={tab.whyText}
                  position="bottom"
                  maxWidth="max-w-xs"
                >
                  <button
                    onClick={() => setActiveTab(tab.id)}
                    className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all duration-150 shrink-0 ${
                      isActive
                        ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20 border border-indigo-500/30'
                        : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100 border border-transparent'
                    }`}
                  >
                    <Icon className={`w-4 h-4 ${isActive ? 'text-white' : 'text-slate-500'}`} />
                    <span>{tab.label}</span>
                  </button>
                </Tooltip>
              );
            })}
          </div>

          {/* Right Scroll Chevron Button */}
          {canScrollRight && (
            <div className="absolute right-0 top-0 bottom-0 z-10 flex items-center pl-6 bg-gradient-to-l from-white via-white/90 to-transparent">
              <Tooltip content="Scroll right through tabs" position="left">
                <button
                  onClick={() => handleScroll('right')}
                  aria-label="Scroll right"
                  className="w-8 h-8 rounded-full bg-white border border-slate-200 text-slate-700 hover:text-indigo-600 hover:bg-indigo-50 hover:border-indigo-200 shadow-md flex items-center justify-center transition-all duration-150"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </Tooltip>
            </div>
          )}

        </div>
      </div>
    </nav>
  );
};


