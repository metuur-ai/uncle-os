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
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-indigo-600 shadow-[0_0_8px_rgba(79,70,229,0.5)]" />
              <span className="text-[10px] text-slate-600 font-bold uppercase tracking-wider font-mono">Node: London-01</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
              <span className="text-[10px] text-slate-600 font-bold uppercase tracking-wider font-mono">Encryption: AES-256</span>
            </div>
            <div className="hidden sm:flex items-center gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-cyan-600" />
              <span className="text-[10px] text-slate-600 font-bold uppercase tracking-wider font-mono">BM25 Local Engine</span>
            </div>
          </div>
          <p className="text-[10px] text-slate-500 font-mono">
            COMPANY OS & TEAM OS • GIT-BASED DUAL-CORE OPERATING SYSTEM v1.4.0
          </p>
        </div>
      </footer>

    </div>
  );
}

