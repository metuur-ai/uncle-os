import React, { useId, useMemo, useState } from 'react';
import { BookOpen, Search, ArrowRight } from 'lucide-react';
import { Modal } from './ui/Modal';
import { Badge, EmptyState, Button, cx } from './ui';
import { TabType } from '../types';
import { DIRECTORY_CATEGORIES_DATA } from '../data/directoryData';
import { getLesson } from '../lessons';

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
  const searchLabelId = useId();

  const filteredCategories = useMemo(() => {
    const q = searchTerm.toLowerCase();
    return DIRECTORY_CATEGORIES_DATA.map((cat) => ({
      ...cat,
      items: cat.items.filter(
        (item) =>
          item.title.toLowerCase().includes(q) ||
          item.description.toLowerCase().includes(q) ||
          cat.categoryName.toLowerCase().includes(q)
      ),
    })).filter((cat) => cat.items.length > 0);
  }, [searchTerm]);

  const go = (tab: TabType) => {
    onClose();
    onNavigateTab(tab);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Company OS Content Directory"
      description="Browse every section, tool, guide, command, and safety gate in one clear directory view."
      icon={BookOpen}
      size="xl"
      footer={
        <p className="text-center text-xs text-fg-subtle">
          Tip: you can reopen this directory anytime from the header menu.
        </p>
      }
    >
      <div className="space-y-5">
        <div>
          <label htmlFor={searchLabelId} className="mb-1.5 block text-sm font-medium text-fg">
            Search content directory
          </label>
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-fg-subtle"
              aria-hidden="true"
            />
            <input
              id={searchLabelId}
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="e.g., PRD, Validate, Gates, Rules, Quiz..."
              className={cx(
                'h-11 w-full rounded-lg border border-border bg-surface pl-10 pr-4 text-sm text-fg',
                'placeholder:text-fg-subtle focus:outline-none'
              )}
              data-modal-autofocus
            />
          </div>
          <p className="sr-only" aria-live="polite">
            {searchTerm
              ? `${filteredCategories.reduce((n, c) => n + c.items.length, 0)} matching entries`
              : 'Showing the full directory'}
          </p>
        </div>

        {filteredCategories.length === 0 ? (
          <EmptyState
            icon={Search}
            title="No directory items found"
            description={`Nothing matches "${searchTerm}".`}
            action={
              <Button variant="secondary" size="sm" onClick={() => setSearchTerm('')}>
                Reset search
              </Button>
            }
          />
        ) : (
          <div className="space-y-6">
            {filteredCategories.map((cat) => {
              const CategoryIcon = getLesson(cat.items[0].targetTab).icon;
              return (
                <section key={cat.id} aria-labelledby={`${cat.id}-heading`} className="space-y-3">
                  <div className="flex items-center gap-2.5 border-b border-border pb-2">
                    <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-accent-soft text-accent-text">
                      <CategoryIcon className="h-4 w-4" aria-hidden="true" />
                    </span>
                    <div className="min-w-0">
                      <h3 id={`${cat.id}-heading`} className="text-base font-semibold tracking-tight text-fg">
                        {cat.categoryName}
                      </h3>
                      <p className="text-xs text-fg-subtle">{cat.description}</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                    {cat.items.map((item, itemIdx) => (
                      <button
                        key={itemIdx}
                        type="button"
                        onClick={() => go(item.targetTab)}
                        className={cx(
                          'group flex cursor-pointer flex-col justify-between gap-2 rounded-lg border border-border p-3.5 text-left',
                          'transition-colors duration-150 hover:border-accent-border hover:bg-accent-soft'
                        )}
                      >
                        <div>
                          <div className="mb-1 flex items-center justify-between gap-2">
                            <span className="text-sm font-semibold text-fg group-hover:text-accent-text">
                              {item.title}
                            </span>
                            {item.tag && <Badge className="shrink-0">{item.tag}</Badge>}
                          </div>
                          <p className="text-xs leading-relaxed text-fg-muted">{item.description}</p>
                        </div>

                        <div className="flex items-center gap-1 pt-1 text-xs font-semibold text-accent-text">
                          <span>Jump to item</span>
                          <ArrowRight
                            className="h-3.5 w-3.5 transition-transform duration-150 group-hover:translate-x-1"
                            aria-hidden="true"
                          />
                        </div>
                      </button>
                    ))}
                  </div>
                </section>
              );
            })}
          </div>
        )}
      </div>
    </Modal>
  );
};
