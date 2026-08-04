import React from 'react';
import { GitBranch, Layers, HelpCircle, Search, ListFilter } from 'lucide-react';
import { Tooltip } from './Tooltip';

interface HeaderProps {
  isStandalone: boolean;
  setIsStandalone: (val: boolean) => void;
  onOpenGuide: () => void;
  onOpenDirectory: () => void;
  onOpenSearch: () => void;
}

/**
 * Global header. Identity on the left, search in the middle, tools and the
 * Company OS / Team OS scope switch on the right.
 *
 * The "Easy Read" toggle that used to live here is gone: the base type scale is
 * now the accessible scale, so there is one type system instead of two.
 */
export const Header: React.FC<HeaderProps> = ({
  isStandalone,
  setIsStandalone,
  onOpenGuide,
  onOpenDirectory,
  onOpenSearch,
}) => {
  return (
    <header className="sticky top-0 z-50 border-b border-border bg-surface/85 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center gap-3 px-4 sm:px-6 lg:px-8">
        {/* --- Identity --- */}
        <a
          href="#main"
          className="flex shrink-0 items-center gap-2.5 rounded-lg"
          aria-label="Company OS and Team OS, version 1.4.0"
        >
          <span
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-accent font-mono text-xs font-bold tracking-tight text-accent-fg"
            aria-hidden="true"
          >
            OS
          </span>
          <span className="hidden min-w-0 sm:block">
            <span className="flex items-center gap-2">
              <span className="truncate text-base font-semibold tracking-tight text-fg">
                Company OS &amp; Team OS
              </span>
              <span className="tabular hidden shrink-0 rounded-md border border-accent-border bg-accent-soft px-1.5 py-0.5 font-mono text-2xs font-medium text-accent-text lg:inline">
                v1.4.0
              </span>
            </span>
            <span className="hidden truncate text-xs text-fg-subtle xl:block">
              Git-based governance, product &amp; engineering operating system
            </span>
          </span>
        </a>

        {/* --- Search --- */}
        <div className="mx-auto min-w-0 max-w-md flex-1">
          <Tooltip
            title="Why click search?"
            content="Instantly find any topic, CLI terminal command, validation gate, or glossary term across the entire platform."
            position="bottom"
            className="w-full"
          >
            <button
              type="button"
              onClick={onOpenSearch}
              className="group flex h-10 w-full cursor-pointer items-center justify-between gap-2 rounded-lg border border-border bg-surface-sunken px-3 text-left transition-colors duration-150 hover:border-border-strong hover:bg-surface"
            >
              <span className="flex min-w-0 items-center gap-2">
                <Search
                  className="h-4 w-4 shrink-0 text-fg-subtle transition-colors group-hover:text-accent-text"
                  aria-hidden="true"
                />
                <span className="truncate text-sm text-fg-subtle">
                  <span className="hidden sm:inline">Search topics, commands, gates &amp; rules…</span>
                  <span className="sm:hidden">Search…</span>
                </span>
              </span>
              <kbd className="hidden shrink-0 rounded border border-border bg-surface px-1.5 py-0.5 font-mono text-2xs text-fg-subtle sm:inline-block">
                Search
              </kbd>
            </button>
          </Tooltip>
        </div>

        {/* --- Tools + scope --- */}
        <div className="flex shrink-0 items-center gap-2">
          <Tooltip
            title="Why click directory?"
            content="Open a complete site map and quick index to navigate all 9 main sections, tools, and guides."
            position="bottom"
          >
            <button
              type="button"
              onClick={onOpenDirectory}
              aria-label="Open content directory"
              className="flex h-10 cursor-pointer items-center gap-2 rounded-lg border border-border-strong bg-surface px-2.5 text-sm font-medium text-fg transition-colors duration-150 hover:bg-surface-sunken sm:px-3"
            >
              <ListFilter className="h-4 w-4 shrink-0 text-fg-muted" aria-hidden="true" />
              <span className="hidden lg:inline">Directory</span>
            </button>
          </Tooltip>

          <Tooltip
            title="Why click glossary?"
            content="Read simple, jargon-free explanations for technical concepts like canonical, PRDs, and gates."
            position="bottom"
          >
            <button
              type="button"
              onClick={onOpenGuide}
              aria-label="Open plain-English glossary"
              className="flex h-10 cursor-pointer items-center gap-2 rounded-lg border border-border-strong bg-surface px-2.5 text-sm font-medium text-fg transition-colors duration-150 hover:bg-surface-sunken sm:px-3"
            >
              <HelpCircle className="h-4 w-4 shrink-0 text-fg-muted" aria-hidden="true" />
              <span className="hidden lg:inline">Glossary</span>
            </button>
          </Tooltip>

          {/* Scope switch. Two mutually exclusive views of the same workspace,
              so it is a radiogroup rather than two independent buttons. */}
          <Tooltip
            title="Why switch scope?"
            content="Toggle between Company OS (company-wide canonical view) and Team OS (single team workspace)."
            position="bottom"
          >
            <div
              role="radiogroup"
              aria-label="Workspace scope"
              className="flex h-10 shrink-0 items-center gap-0.5 rounded-lg border border-border bg-surface-sunken p-1"
            >
              <button
                type="button"
                role="radio"
                aria-checked={!isStandalone}
                onClick={() => setIsStandalone(false)}
                className={`flex h-8 cursor-pointer items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-colors duration-150 ${
                  !isStandalone
                    ? 'bg-accent text-accent-fg shadow-xs'
                    : 'text-fg-muted hover:text-fg'
                }`}
              >
                <Layers className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
                <span className="hidden sm:inline">Company OS</span>
                <span className="sm:hidden">Company</span>
              </button>
              <button
                type="button"
                role="radio"
                aria-checked={isStandalone}
                onClick={() => setIsStandalone(true)}
                className={`flex h-8 cursor-pointer items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-colors duration-150 ${
                  isStandalone ? 'bg-scope text-scope-fg shadow-xs' : 'text-fg-muted hover:text-fg'
                }`}
              >
                <GitBranch className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
                <span className="hidden sm:inline">Team OS</span>
                <span className="sm:hidden">Team</span>
              </button>
            </div>
          </Tooltip>
        </div>
      </div>
    </header>
  );
};
