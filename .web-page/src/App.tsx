import React, { useCallback, useEffect, useRef, useState } from 'react';
import { TabType } from './types';
import { LESSONS, nextLesson } from './lessons';
import { hashForTab, tabFromHash } from './routing';
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
import { LessonFooter, PageShell } from './components/ui';
import { Coffee, Github } from 'lucide-react';

const VERSION = '1.4.0';
const GITHUB_URL = 'https://github.com/metuur-ai/uncle-os';
const X_URL = 'https://x.com/javierhbr';
const COFFEE_URL = 'https://buymeacoffee.com/javierhbr';

const footerLinkClass =
  'inline-flex h-10 items-center gap-1.5 rounded-lg px-2 text-fg-subtle transition-colors duration-150 hover:text-accent-text cursor-pointer';

// lucide-react has no X (Twitter) brand mark, so it is inlined here.
const XMark = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className={className}>
    <path d="M18.9 1.153h3.68l-8.04 9.19L24 22.846h-7.406l-5.8-7.584-6.638 7.584H.474l8.6-9.83L0 1.154h7.594l5.243 6.932zm-1.29 19.49h2.039L6.486 3.24H4.298z" />
  </svg>
);

export default function App() {
  const [activeTab, setActiveTab] = useState<TabType>(() => tabFromHash() ?? 'home');
  const [isStandalone, setIsStandalone] = useState<boolean>(false);
  const [isGuideOpen, setIsGuideOpen] = useState<boolean>(false);
  const [isDirectoryOpen, setIsDirectoryOpen] = useState<boolean>(false);
  const [isSearchOpen, setIsSearchOpen] = useState<boolean>(false);
  const [visited, setVisited] = useState<Set<TabType>>(
    () => new Set<TabType>(['home', tabFromHash() ?? 'home'])
  );

  const mainRef = useRef<HTMLElement>(null);
  const isFirstRender = useRef(true);

  /**
   * Navigation goes through the URL, never straight to state: the hashchange
   * listener below is the single place that applies a route, so a click and a
   * browser Back button take the exact same path and cannot disagree.
   */
  const goToTab = useCallback((tab: TabType) => {
    window.location.hash = hashForTab(tab);
  }, []);

  useEffect(() => {
    const applyRoute = () => {
      const tab = tabFromHash();
      if (!tab) return; // in-page anchor or unknown id — leave the view alone
      setActiveTab(tab);
      setVisited((prev) => (prev.has(tab) ? prev : new Set(prev).add(tab)));
    };
    window.addEventListener('hashchange', applyRoute);
    return () => window.removeEventListener('hashchange', applyRoute);
  }, []);

  // Each lesson is its own history entry, so it needs its own title.
  useEffect(() => {
    const label = LESSONS.find((l) => l.id === activeTab)?.label;
    document.title = label && activeTab !== 'home' ? `${label} · Company OS` : 'Company OS & Team OS';
  }, [activeTab]);

  /**
   * A tab change is a navigation. Reset scroll and move focus to the new view
   * so keyboard and screen-reader users land on the content rather than being
   * stranded mid-page on the nav they just left.
   */
  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    window.scrollTo({ top: 0, behavior: 'smooth' });
    mainRef.current?.focus({ preventScroll: true });
  }, [activeTab]);

  // Cmd/Ctrl+K opens search — the shortcut people already expect.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setIsSearchOpen(true);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const upcoming = nextLesson(activeTab);
  const activeLabel = LESSONS.find((l) => l.id === activeTab)?.label ?? '';

  return (
    <div className="min-h-dvh bg-canvas font-sans text-fg">
      <a
        href="#main"
        className="skip-link rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-fg shadow-md"
      >
        Skip to content
      </a>

      <Header
        isStandalone={isStandalone}
        setIsStandalone={setIsStandalone}
        onOpenGuide={() => setIsGuideOpen(true)}
        onOpenDirectory={() => setIsDirectoryOpen(true)}
        onOpenSearch={() => setIsSearchOpen(true)}
      />

      <Navbar activeTab={activeTab} setActiveTab={goToTab} visited={visited} />

      <main
        id="main"
        ref={mainRef}
        tabIndex={-1}
        role="tabpanel"
        aria-label={activeLabel}
        className="py-10 outline-none lg:py-14"
      >
        {/* Each view supplies its own PageShell / PageHeader. */}
        {activeTab === 'home' && (
          <HomeOverview
            onNavigateTab={goToTab}
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

        {/* Lessons chain to the next one so the site reads as a sequence.
            Home has its own calls to action and does not need this. */}
        {activeTab !== 'home' && (
          <div className="mt-12 lg:mt-16">
            <PageShell>
              <LessonFooter
                onNext={() => goToTab(upcoming.id)}
                nextLabel={upcoming.label}
                nextDescription={upcoming.whyText}
              />
            </PageShell>
          </div>
        )}
      </main>

      <BeginnerGuideModal
        isOpen={isGuideOpen}
        onClose={() => setIsGuideOpen(false)}
        onNavigateTab={(tab) => goToTab(tab as TabType)}
      />

      <ContentDirectoryModal
        isOpen={isDirectoryOpen}
        onClose={() => setIsDirectoryOpen(false)}
        onNavigateTab={goToTab}
      />

      <GlobalSearchModal
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
        onNavigateTab={goToTab}
      />

      <footer className="border-t border-border bg-surface py-8">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 px-4 sm:px-6 md:flex-row lg:px-8">
          <p className="tabular font-mono text-xs text-fg-subtle">
            © 2026 uncle-os v{VERSION} · made by{' '}
            <a
              href={X_URL}
              target="_blank"
              rel="noreferrer"
              className="font-semibold text-fg-muted transition-colors duration-150 hover:text-accent-text"
            >
              @javierhbr
            </a>
          </p>

          <div className="flex items-center gap-2">
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noreferrer"
              aria-label="GitHub repository"
              className={footerLinkClass}
            >
              <Github className="h-4 w-4" aria-hidden="true" />
            </a>

            <a
              href={X_URL}
              target="_blank"
              rel="noreferrer"
              aria-label="Author on X"
              className={footerLinkClass}
            >
              <XMark className="h-3.5 w-3.5" />
            </a>

            <a href={COFFEE_URL} target="_blank" rel="noreferrer" className={footerLinkClass}>
              <Coffee className="h-4 w-4" aria-hidden="true" />
              <span className="font-mono text-xs font-medium">Buy me a coffee</span>
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
