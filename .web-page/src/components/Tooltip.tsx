import React, { useEffect, useId, useRef, useState } from 'react';
import { cx } from './ui';

interface TooltipProps {
  content: React.ReactNode;
  title?: string;
  position?: 'top' | 'bottom' | 'left' | 'right';
  /** Hover-in delay in ms. Keyboard focus always shows instantly. */
  delay?: number;
  children: React.ReactElement;
  className?: string;
  maxWidth?: string;
}

/**
 * Lightweight, accessible tooltip. Shows on hover AND keyboard focus, hides
 * on Escape/blur/mouse-leave. The bubble is `pointer-events-none` so it never
 * traps the pointer and never blocks an underlying tap on touch devices —
 * touch interaction never triggers it in the first place, since it only
 * listens for mouse/focus events.
 */
export const Tooltip: React.FC<TooltipProps> = ({
  content,
  title,
  position = 'bottom',
  delay = 150,
  children,
  className = '',
  maxWidth = 'max-w-xs',
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const tooltipId = useId();

  const show = (immediate: boolean) => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    if (immediate) {
      setIsVisible(true);
    } else {
      timeoutRef.current = setTimeout(() => setIsVisible(true), delay);
    }
  };

  const hide = () => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    setIsVisible(false);
  };

  useEffect(() => {
    if (!isVisible) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') hide();
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [isVisible]);

  useEffect(() => () => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
  }, []);

  const positionClasses = {
    top: 'bottom-full mb-2 left-1/2 -translate-x-1/2',
    bottom: 'top-full mt-2 left-1/2 -translate-x-1/2',
    left: 'right-full mr-2 top-1/2 -translate-y-1/2',
    right: 'left-full ml-2 top-1/2 -translate-y-1/2',
  };

  const arrowClasses = {
    top: 'top-full left-1/2 -translate-x-1/2 border-t-fg border-x-transparent border-b-transparent',
    bottom: 'bottom-full left-1/2 -translate-x-1/2 border-b-fg border-x-transparent border-t-transparent',
    left: 'left-full top-1/2 -translate-y-1/2 border-l-fg border-y-transparent border-r-transparent',
    right: 'right-full top-1/2 -translate-y-1/2 border-r-fg border-y-transparent border-l-transparent',
  };

  const trigger = React.cloneElement(children as React.ReactElement<{ 'aria-describedby'?: string }>, {
    'aria-describedby': isVisible ? tooltipId : undefined,
  });

  return (
    <div
      className={cx('relative inline-flex', className)}
      onMouseEnter={() => show(false)}
      onMouseLeave={hide}
      onFocus={() => show(true)}
      onBlur={hide}
    >
      {trigger}

      {isVisible && (
        <div
          id={tooltipId}
          role="tooltip"
          className={cx(
            'pointer-events-none absolute z-50 transition-opacity duration-200 animate-fade-in',
            positionClasses[position],
            maxWidth
          )}
        >
          <div className="rounded-xl border border-border-strong bg-fg p-2.5 text-left leading-snug text-fg-inverse shadow-lg">
            {title && (
              <div className="mb-1 flex items-center gap-1.5 font-mono text-2xs font-bold uppercase tracking-wider text-fg-inverse/80">
                <span className="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden="true" />
                <span>{title}</span>
              </div>
            )}
            <div className="text-xs leading-relaxed text-fg-inverse/90">{content}</div>
          </div>
          <div className={cx('absolute h-0 w-0 border-4', arrowClasses[position])} aria-hidden="true" />
        </div>
      )}
    </div>
  );
};
