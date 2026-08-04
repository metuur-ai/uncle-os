import React, { useRef, useState, useEffect, useCallback } from 'react';
import { TabType } from '../types';
import { LESSONS, lessonNumber } from '../lessons';
import { Check, ChevronLeft, ChevronRight } from 'lucide-react';

interface NavbarProps {
  activeTab: TabType;
  setActiveTab: (tab: TabType) => void;
  /** Tabs the learner has already opened — drives the completion ticks. */
  visited: Set<TabType>;
}

/**
 * Lesson rail. Sticks directly under the header and doubles as a progress
 * indicator: each tutorial shows its number until it has been visited, then a
 * tick. Implemented as a real tablist so arrow keys work.
 */
export const Navbar: React.FC<NavbarProps> = ({ activeTab, setActiveTab, visited }) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(false);

  const measure = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setCanScrollLeft(el.scrollLeft > 4);
    setCanScrollRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 4);
  }, []);

  useEffect(() => {
    measure();
    const el = scrollRef.current;
    if (!el) return;
    el.addEventListener('scroll', measure, { passive: true });
    window.addEventListener('resize', measure);
    return () => {
      el.removeEventListener('scroll', measure);
      window.removeEventListener('resize', measure);
    };
  }, [measure]);

  // Keep the active tab in view when navigation happens from elsewhere
  // (home cards, lesson footers, the directory modal).
  useEffect(() => {
    const i = LESSONS.findIndex((l) => l.id === activeTab);
    tabRefs.current[i]?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' });
  }, [activeTab]);

  const nudge = (dir: 'left' | 'right') => {
    scrollRef.current?.scrollBy({ left: dir === 'left' ? -280 : 280, behavior: 'smooth' });
  };

  const onKeyDown = (e: React.KeyboardEvent) => {
    const i = LESSONS.findIndex((l) => l.id === activeTab);
    let next = -1;
    if (e.key === 'ArrowRight') next = (i + 1) % LESSONS.length;
    if (e.key === 'ArrowLeft') next = (i - 1 + LESSONS.length) % LESSONS.length;
    if (e.key === 'Home') next = 0;
    if (e.key === 'End') next = LESSONS.length - 1;
    if (next < 0) return;
    e.preventDefault();
    setActiveTab(LESSONS[next].id);
    tabRefs.current[next]?.focus();
  };

  const arrowBase =
    'absolute top-0 bottom-0 z-10 flex w-10 cursor-pointer items-center justify-center text-fg-muted transition-colors hover:text-fg';

  return (
    <nav className="sticky top-16 z-40 border-b border-border bg-canvas/90 backdrop-blur-md">
      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {canScrollLeft && (
          <>
            <button
              type="button"
              onClick={() => nudge('left')}
              aria-label="Scroll lessons left"
              className={`${arrowBase} left-0`}
            >
              <ChevronLeft className="h-5 w-5" aria-hidden="true" />
            </button>
            <div
              className="pointer-events-none absolute inset-y-0 left-10 z-10 w-8 bg-gradient-to-r from-canvas to-transparent"
              aria-hidden="true"
            />
          </>
        )}

        <div
          ref={scrollRef}
          role="tablist"
          aria-label="Lessons"
          onKeyDown={onKeyDown}
          className="no-scrollbar flex items-stretch gap-1 overflow-x-auto"
        >
          {LESSONS.map((lesson, i) => {
            const active = lesson.id === activeTab;
            const n = lessonNumber(lesson.id);
            const done = visited.has(lesson.id) && !active && n > 0;

            return (
              <button
                key={lesson.id}
                ref={(el) => {
                  tabRefs.current[i] = el;
                }}
                role="tab"
                aria-selected={active}
                aria-controls="main"
                tabIndex={active ? 0 : -1}
                onClick={() => setActiveTab(lesson.id)}
                title={lesson.whyText}
                className={`group relative flex h-13 shrink-0 cursor-pointer items-center gap-2 whitespace-nowrap px-5 py-3 text-sm font-medium transition-colors duration-150 ${
                  active ? 'text-accent-text' : 'text-fg-muted hover:text-fg'
                }`}
              >
                {/* Step marker: number → tick once visited. Never colour alone. */}
                <span
                  className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border text-2xs font-semibold transition-colors duration-150 ${
                    active
                      ? 'border-accent bg-accent text-accent-fg'
                      : done
                        ? 'border-success-border bg-success-soft text-success-text'
                        : 'border-border bg-surface text-fg-subtle'
                  }`}
                  aria-hidden="true"
                >
                  {n === 0 ? (
                    <lesson.icon className="h-3 w-3" />
                  ) : done ? (
                    <Check className="h-3 w-3" />
                  ) : (
                    <span className="tabular">{n}</span>
                  )}
                </span>

                <span className="hidden md:inline">{lesson.label}</span>
                <span className="md:hidden">{lesson.shortLabel}</span>

                {done && <span className="sr-only">(visited)</span>}

                {/* Active underline. Transform-only so it cannot shift layout. */}
                <span
                  className={`absolute inset-x-2 bottom-0 h-0.5 rounded-full bg-accent transition-transform duration-200 ${
                    active ? 'scale-x-100' : 'scale-x-0'
                  }`}
                  aria-hidden="true"
                />
              </button>
            );
          })}
        </div>

        {canScrollRight && (
          <>
            <div
              className="pointer-events-none absolute inset-y-0 right-10 z-10 w-8 bg-gradient-to-l from-canvas to-transparent"
              aria-hidden="true"
            />
            <button
              type="button"
              onClick={() => nudge('right')}
              aria-label="Scroll lessons right"
              className={`${arrowBase} right-0`}
            >
              <ChevronRight className="h-5 w-5" aria-hidden="true" />
            </button>
          </>
        )}
      </div>
    </nav>
  );
};
