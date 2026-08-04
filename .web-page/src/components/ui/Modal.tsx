import React, { useEffect, useId, useRef } from 'react';
import { X, type LucideIcon } from 'lucide-react';
import { cx } from './primitives';

/* ============================================================================
   Shared modal primitive. See DESIGN.md section 6 (Accessibility floor) and
   section 8 (component inventory). Every modal in the app renders through
   this — it owns focus trapping, Escape-to-close, scroll lock, and the
   backdrop/panel chrome so individual modals only supply content.
   ========================================================================= */

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])';

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  icon?: LucideIcon;
  size?: 'md' | 'lg' | 'xl';
  children: React.ReactNode;
  footer?: React.ReactNode;
  className?: string;
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  description,
  icon: Icon,
  size = 'md',
  children,
  footer,
  className,
}) => {
  const titleId = useId();
  const descId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    triggerRef.current = document.activeElement as HTMLElement | null;

    const panel = panelRef.current;
    const autoFocusTarget = panel?.querySelector<HTMLElement>('[data-modal-autofocus]');
    const firstFocusable = panel?.querySelector<HTMLElement>(FOCUSABLE_SELECTOR);
    (autoFocusTarget ?? firstFocusable ?? panel)?.focus();

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
        return;
      }
      if (e.key !== 'Tab') return;

      const nodes = panel?.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR);
      if (!nodes || nodes.length === 0) return;
      const list = Array.from(nodes);
      const first = list[0];
      const last = list[list.length - 1];

      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('keydown', onKeyDown);
      document.body.style.overflow = previousOverflow;
      triggerRef.current?.focus?.();
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const sizes = { md: 'max-w-md', lg: 'max-w-2xl', xl: 'max-w-4xl' }[size];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-overlay p-4 backdrop-blur-sm animate-fade-in"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={description ? descId : undefined}
        tabIndex={-1}
        className={cx(
          'flex max-h-[85vh] w-full flex-col overflow-hidden rounded-xl border border-border bg-surface shadow-lg animate-scale-in',
          sizes,
          className
        )}
      >
        <div className="flex shrink-0 items-start justify-between gap-4 border-b border-border p-5 sm:p-6">
          <div className="flex min-w-0 items-start gap-3">
            {Icon && (
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-accent-soft text-accent-text">
                <Icon className="h-5 w-5" aria-hidden="true" />
              </span>
            )}
            <div className="min-w-0">
              <h2 id={titleId} className="text-xl font-semibold tracking-tight text-fg">
                {title}
              </h2>
              {description && (
                <p id={descId} className="mt-1 text-sm text-fg-muted">
                  {description}
                </p>
              )}
            </div>
          </div>

          <button
            type="button"
            onClick={onClose}
            aria-label="Close dialog"
            className={cx(
              'flex h-11 w-11 shrink-0 cursor-pointer items-center justify-center rounded-lg',
              'text-fg-muted transition-colors duration-150 hover:bg-surface-sunken hover:text-fg'
            )}
          >
            <X className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>

        <div className="thin-scrollbar min-h-0 flex-1 overflow-y-auto p-5 sm:p-6">{children}</div>

        {footer && <div className="shrink-0 border-t border-border p-4 sm:p-5">{footer}</div>}
      </div>
    </div>
  );
};
