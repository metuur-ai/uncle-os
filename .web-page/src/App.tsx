import React, { useState } from 'react';
import { TabType } from './types';
import { Header } from './components/Header';
import { Navbar } from './components/Navbar';
import { HomeOverview } from './components/HomeOverview';
import { InstallSetupView } from './components/InstallSetupView';
import { ArchitectureView } from './components/ArchitectureView';
import { CliTerminalView } from './components/CliTerminalView';
import { WorkflowPlayground } from './components/WorkflowPlayground';
import { GovernanceTiersView } from './components/GovernanceTiersView';
import { ValidationGatesView } from './components/ValidationGatesView';
import { LocalSearchAgentView } from './components/LocalSearchAgentView';
import { ReferenceMatrix } from './components/ReferenceMatrix';
import { QuizSection } from './components/QuizSection';
import { BeginnerGuideModal } from './components/BeginnerGuideModal';
import { ContentDirectoryModal } from './components/ContentDirectoryModal';
import { GlobalSearchModal } from './components/GlobalSearchModal';
import { WhyWhatHowCard } from './components/WhyWhatHowCard';
import { Coffee, Github } from 'lucide-react';

const VERSION = '1.4.0';
const GITHUB_URL = 'https://github.com/metuur-ai/uncle-os';
const X_URL = 'https://x.com/javierhbr';
const COFFEE_URL = 'https://buymeacoffee.com/javierhbr';

const footerLinkClass = 'text-slate-500 hover:text-indigo-600 transition-colors';

// lucide-react has no X (Twitter) brand mark, so it is inlined here.
const XMark = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className={className}>
    <path d="M18.9 1.153h3.68l-8.04 9.19L24 22.846h-7.406l-5.8-7.584-6.638 7.584H.474l8.6-9.83L0 1.154h7.594l5.243 6.932zm-1.29 19.49h2.039L6.486 3.24H4.298z" />
  </svg>
);

export default function App() {
  const [activeTab, setActiveTab] = useState<TabType>('home');
  const [isStandalone, setIsStandalone] = useState<boolean>(false);
  const [isGuideOpen, setIsGuideOpen] = useState<boolean>(false);
  const [isDirectoryOpen, setIsDirectoryOpen] = useState<boolean>(false);
  const [isSearchOpen, setIsSearchOpen] = useState<boolean>(false);
  const [easyReadMode, setEasyReadMode] = useState<boolean>(false);

  return (
    <div className={`min-h-screen bg-slate-50 text-slate-900 font-sans antialiased selection:bg-indigo-500 selection:text-white transition-all ${
      easyReadMode ? 'text-base leading-relaxed font-medium' : 'text-sm'
    }`}>
      
      {/* Top Header */}
      <Header
        isStandalone={isStandalone}
        setIsStandalone={setIsStandalone}
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onOpenGuide={() => setIsGuideOpen(true)}
        onOpenDirectory={() => setIsDirectoryOpen(true)}
        onOpenSearch={() => setIsSearchOpen(true)}
        easyReadMode={easyReadMode}
        setEasyReadMode={setEasyReadMode}
      />

      {/* Tab Navigation */}
      <Navbar activeTab={activeTab} setActiveTab={setActiveTab} />

      {/* Main Content Body */}
      <main className="max-w-[95%] mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        
        {/* Core Why, What, How Card */}
        <WhyWhatHowCard
          onNavigate={(tab) => setActiveTab(tab)}
          onOpenGuide={() => setIsGuideOpen(true)}
        />

        {/* Dynamic Views */}
        {activeTab === 'home' && (
          <HomeOverview
            onNavigateTab={(tab) => setActiveTab(tab)}
            onOpenGuide={() => setIsGuideOpen(true)}
            isStandalone={isStandalone}
            setIsStandalone={setIsStandalone}
          />
        )}

        {activeTab === 'install' && <InstallSetupView />}

        {activeTab === 'architecture' && (
          <ArchitectureView isStandalone={isStandalone} setIsStandalone={setIsStandalone} />
        )}

        {activeTab === 'cli' && <CliTerminalView />}

        {activeTab === 'workflows' && <WorkflowPlayground />}

        {activeTab === 'governance' && <GovernanceTiersView />}

        {activeTab === 'validation' && <ValidationGatesView />}

        {activeTab === 'search-agent' && <LocalSearchAgentView />}

        {activeTab === 'reference' && <ReferenceMatrix />}

        {activeTab === 'quiz' && <QuizSection />}
      </main>

      {/* Accessible Plain English Modal */}
      <BeginnerGuideModal
        isOpen={isGuideOpen}
        onClose={() => setIsGuideOpen(false)}
        onNavigateTab={(tab) => setActiveTab(tab as TabType)}
      />

      {/* Content Directory Modal */}
      <ContentDirectoryModal
        isOpen={isDirectoryOpen}
        onClose={() => setIsDirectoryOpen(false)}
        onNavigateTab={(tab) => setActiveTab(tab)}
      />

      {/* Static Global Search Modal */}
      <GlobalSearchModal
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
        onNavigateTab={(tab) => setActiveTab(tab)}
      />

      {/* Bento Grid Footer */}
      <footer className="border-t border-slate-200 bg-white py-6 text-slate-500">
        <div className="max-w-[95%] mx-auto px-4 sm:px-6 lg:px-8 flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-[11px] text-slate-500 font-mono">
            © 2026 uncle-os v{VERSION} · made by{' '}
            <a
              href={X_URL}
              target="_blank"
              rel="noreferrer"
              className="text-slate-700 font-bold hover:text-indigo-600 transition-colors"
            >
              @javierhbr
            </a>
          </p>

          <div className="flex items-center gap-5">
            <a href={GITHUB_URL} target="_blank" rel="noreferrer" aria-label="GitHub repository" className={footerLinkClass}>
              <Github className="w-4 h-4" />
            </a>

            <a href={X_URL} target="_blank" rel="noreferrer" aria-label="Author on X" className={footerLinkClass}>
              <XMark className="w-3.5 h-3.5" />
            </a>

            <a
              href={COFFEE_URL}
              target="_blank"
              rel="noreferrer"
              className={`${footerLinkClass} flex items-center gap-1.5`}
            >
              <Coffee className="w-4 h-4" />
              <span className="text-[11px] font-bold uppercase tracking-wider font-mono">Buy me a coffee</span>
            </a>
          </div>
        </div>
      </footer>

    </div>
  );
}

