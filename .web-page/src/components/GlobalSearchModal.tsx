import React, { useEffect, useId, useMemo, useState } from 'react';
import {
  Search,
  ArrowRight,
  Download,
  FolderTree,
  Terminal,
  ShieldCheck,
  PlayCircle,
  BookOpen,
  Award,
  FileText,
  Layers,
  Bot,
  GitBranch,
  Plug,
  type LucideIcon,
} from 'lucide-react';
import { Modal } from './ui/Modal';
import { EmptyState, cx } from './ui';
import { TabType } from '../types';
import { STATIC_SEARCH_ITEMS, SearchResultItemData } from '../data/searchData';

interface GlobalSearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab: (tab: TabType) => void;
}

const ICONS: Record<string, LucideIcon> = {
  Layers,
  Download,
  FolderTree,
  Terminal,
  ShieldCheck,
  PlayCircle,
  BookOpen,
  FileText,
  Award,
  Bot,
  GitBranch,
  Plug,
};

const getItemIcon = (iconName: string): LucideIcon => ICONS[iconName] ?? Search;

export const GlobalSearchModal: React.FC<GlobalSearchModalProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
}) => {
  const [query, setQuery] = useState('');
  const [highlighted, setHighlighted] = useState(0);
  const listboxId = useId();

  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setHighlighted(0);
    }
  }, [isOpen]);

  const results = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (q === '') return STATIC_SEARCH_ITEMS;
    return STATIC_SEARCH_ITEMS.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        item.snippet.toLowerCase().includes(q) ||
        item.category.toLowerCase().includes(q) ||
        item.keywords.some((k) => k.toLowerCase().includes(q))
    );
  }, [query]);

  const groups = useMemo(() => {
    const order: string[] = [];
    const byCategory = new Map<string, SearchResultItemData[]>();
    for (const item of results) {
      if (!byCategory.has(item.category)) {
        byCategory.set(item.category, []);
        order.push(item.category);
      }
      byCategory.get(item.category)!.push(item);
    }
    return order.map((category) => ({ category, items: byCategory.get(category)! }));
  }, [results]);

  useEffect(() => {
    setHighlighted(0);
  }, [query]);

  const select = (item: SearchResultItemData) => {
    onClose();
    onNavigateTab(item.targetTab);
  };

  const onInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (results.length === 0) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setHighlighted((i) => (i + 1) % results.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlighted((i) => (i - 1 + results.length) % results.length);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const item = results[highlighted];
      if (item) select(item);
    }
  };

  const activeItemId = results[highlighted] ? `${listboxId}-${results[highlighted].id}` : undefined;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Search Company OS"
      description="Search all views, commands, gates, rules & the quiz."
      icon={Search}
      size="lg"
      footer={
        <p className="text-center text-2xs text-fg-subtle">
          Use <kbd className="font-mono">Up</kbd>/<kbd className="font-mono">Down</kbd> to move,{' '}
          <kbd className="font-mono">Enter</kbd> to select, <kbd className="font-mono">Esc</kbd> to close.
        </p>
      }
    >
      <div className="space-y-4">
        <div>
          <label htmlFor={listboxId} className="mb-1.5 block text-sm font-medium text-fg">
            Search query
          </label>
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-fg-subtle"
              aria-hidden="true"
            />
            <input
              id={listboxId}
              type="text"
              role="combobox"
              aria-expanded="true"
              aria-controls={`${listboxId}-list`}
              aria-activedescendant={activeItemId}
              autoComplete="off"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={onInputKeyDown}
              placeholder="Type to search all views, commands, gates, rules & quiz..."
              className={cx(
                'h-11 w-full rounded-lg border border-border bg-surface pl-10 pr-4 text-sm text-fg',
                'placeholder:text-fg-subtle focus:outline-none'
              )}
              data-modal-autofocus
            />
          </div>
          <p className="mt-1.5 text-xs text-fg-subtle" aria-live="polite">
            {query.trim() === ''
              ? 'Showing all indexed topics'
              : `Found ${results.length} matching ${results.length === 1 ? 'topic' : 'topics'}`}
          </p>
        </div>

        {results.length === 0 ? (
          <EmptyState
            icon={Search}
            title={`No results for "${query}"`}
            description='Try "PRD", "CLI", "Gate", "Governance", "Federation", "MCP", "Quiz" or "Exit code".'
          />
        ) : (
          <div id={`${listboxId}-list`} role="listbox" aria-label="Search results" className="space-y-5">
            {groups.map((group) => (
              <div key={group.category}>
                <h3 className="mb-2 font-mono text-2xs font-semibold uppercase tracking-widest text-fg-subtle">
                  {group.category}
                </h3>
                <div className="space-y-2">
                  {group.items.map((item) => {
                    const Icon = getItemIcon(item.iconName);
                    const index = results.indexOf(item);
                    const isActive = index === highlighted;
                    return (
                      <button
                        key={item.id}
                        id={`${listboxId}-${item.id}`}
                        type="button"
                        role="option"
                        aria-selected={isActive}
                        onMouseEnter={() => setHighlighted(index)}
                        onClick={() => select(item)}
                        className={cx(
                          'group flex w-full cursor-pointer items-start justify-between gap-3 rounded-lg border p-3.5 text-left',
                          'transition-colors duration-150',
                          isActive
                            ? 'border-accent-border bg-accent-soft'
                            : 'border-border bg-surface hover:border-accent-border hover:bg-accent-soft'
                        )}
                      >
                        <div className="flex min-w-0 items-start gap-3">
                          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-accent-soft text-accent-text">
                            <Icon className="h-4 w-4" aria-hidden="true" />
                          </span>
                          <div className="min-w-0">
                            <span className="block text-sm font-semibold text-fg group-hover:text-accent-text">
                              {item.title}
                            </span>
                            <p className="mt-0.5 text-xs leading-relaxed text-fg-muted">{item.snippet}</p>
                          </div>
                        </div>
                        <ArrowRight
                          className="h-4 w-4 shrink-0 self-center text-accent-text transition-transform duration-150 group-hover:translate-x-1"
                          aria-hidden="true"
                        />
                      </button>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Modal>
  );
};
