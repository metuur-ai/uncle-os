import React, { useCallback, useEffect, useId, useRef, useState } from 'react';
import {
  Check,
  Copy,
  Info,
  AlertTriangle,
  XCircle,
  CheckCircle2,
  Sparkles,
  ArrowRight,
  Loader2,
  type LucideIcon,
} from 'lucide-react';

/* ============================================================================
   Shared UI primitives. See DESIGN.md.
   Everything here consumes semantic tokens only — no raw palette utilities.
   ========================================================================= */

export const cx = (...parts: (string | false | null | undefined)[]) =>
  parts.filter(Boolean).join(' ');

/* --- Tone ---------------------------------------------------------------- */

export type Tone = 'neutral' | 'accent' | 'scope' | 'success' | 'warn' | 'danger' | 'info';

const toneSoft: Record<Tone, string> = {
  neutral: 'bg-surface-sunken text-fg-muted border-border',
  accent: 'bg-accent-soft text-accent-text border-accent-border',
  scope: 'bg-scope-soft text-scope-text border-scope-border',
  success: 'bg-success-soft text-success-text border-success-border',
  warn: 'bg-warn-soft text-warn-text border-warn-border',
  danger: 'bg-danger-soft text-danger-text border-danger-border',
  info: 'bg-info-soft text-info-text border-info-border',
};

const toneSolid: Record<Tone, string> = {
  neutral: 'bg-fg-muted text-fg-inverse',
  accent: 'bg-accent text-accent-fg',
  scope: 'bg-scope text-scope-fg',
  success: 'bg-success text-fg-inverse',
  warn: 'bg-warn text-fg-inverse',
  danger: 'bg-danger text-fg-inverse',
  info: 'bg-info text-fg-inverse',
};

const toneIcon: Record<Tone, LucideIcon> = {
  neutral: Info,
  accent: Sparkles,
  scope: Info,
  success: CheckCircle2,
  warn: AlertTriangle,
  danger: XCircle,
  info: Info,
};

/* --- Layout -------------------------------------------------------------- */

export const PageShell: React.FC<{
  children: React.ReactNode;
  width?: 'default' | 'wide';
  className?: string;
}> = ({ children, width = 'default', className }) => (
  <div
    className={cx(
      'mx-auto w-full px-4 sm:px-6 lg:px-8',
      width === 'wide' ? 'max-w-7xl' : 'max-w-6xl',
      className
    )}
  >
    <div className="space-y-10 lg:space-y-12">{children}</div>
  </div>
);

export const PageHeader: React.FC<{
  eyebrow?: string;
  title: string;
  lead?: string;
  icon?: LucideIcon;
  actions?: React.ReactNode;
}> = ({ eyebrow, title, lead, icon: Icon, actions }) => (
  <header className="animate-fade-rise">
    <div className="flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
      <div className="min-w-0">
        {eyebrow && (
          <p className="mb-2 flex items-center gap-2 font-mono text-xs font-medium uppercase tracking-widest text-accent-text">
            {Icon && <Icon className="h-3.5 w-3.5" aria-hidden="true" />}
            {eyebrow}
          </p>
        )}
        <h1 className="text-3xl font-bold tracking-tight text-fg">{title}</h1>
        {lead && <p className="measure mt-3 text-lg text-fg-muted">{lead}</p>}
      </div>
      {actions && <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div>}
    </div>
  </header>
);

export const Section: React.FC<{
  title?: string;
  description?: string;
  id?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}> = ({ title, description, id, actions, children, className }) => {
  const auto = useId();
  const headingId = id ? `${id}-heading` : auto;
  return (
    <section aria-labelledby={title ? headingId : undefined} id={id} className={cx('scroll-mt-36', className)}>
      {title && (
        <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div className="min-w-0">
            <h2 id={headingId} className="text-2xl font-semibold tracking-tight text-fg">
              {title}
            </h2>
            {description && <p className="measure mt-2 text-base text-fg-muted">{description}</p>}
          </div>
          {actions && <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div>}
        </div>
      )}
      {children}
    </section>
  );
};

export const Prose: React.FC<{ children: React.ReactNode; className?: string }> = ({
  children,
  className,
}) => <div className={cx('measure space-y-4 text-base text-fg-muted', className)}>{children}</div>;

/* --- Card ---------------------------------------------------------------- */

type CardProps = {
  children: React.ReactNode;
  className?: string;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  tone?: Tone;
} & (
  | { as?: 'div'; onClick?: never; interactive?: false }
  | { as: 'button'; onClick: () => void; interactive?: true }
);

export const Card: React.FC<CardProps> = ({
  children,
  className,
  padding = 'md',
  tone,
  as = 'div',
  onClick,
}) => {
  const pad = { none: '', sm: 'p-4', md: 'p-5', lg: 'p-6 sm:p-8' }[padding];
  const base = cx(
    'rounded-xl border text-left',
    tone ? toneSoft[tone] : 'border-border bg-surface',
    pad,
    className
  );

  if (as === 'button') {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cx(
          base,
          'group w-full cursor-pointer transition-all duration-200 hover:border-accent-border hover:shadow-md active:scale-[0.99]'
        )}
      >
        {children}
      </button>
    );
  }
  return <div className={base}>{children}</div>;
};

/* --- Button -------------------------------------------------------------- */

export const Button: React.FC<{
  children?: React.ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  icon?: LucideIcon;
  iconRight?: LucideIcon;
  disabled?: boolean;
  loading?: boolean;
  href?: string;
  type?: 'button' | 'submit';
  'aria-label'?: string;
  className?: string;
  fullWidth?: boolean;
}> = ({
  children,
  onClick,
  variant = 'primary',
  size = 'md',
  icon: Icon,
  iconRight: IconRight,
  disabled,
  loading,
  href,
  type = 'button',
  className,
  fullWidth,
  ...rest
}) => {
  const variants = {
    primary: 'bg-accent text-accent-fg hover:bg-accent-hover active:bg-accent-active shadow-xs',
    secondary: 'bg-surface text-fg border border-border-strong hover:bg-surface-sunken',
    ghost: 'text-fg-muted hover:bg-surface-sunken hover:text-fg',
    danger: 'bg-danger text-fg-inverse hover:opacity-90',
  }[variant];

  // Heights clear the 44px touch floor at md/lg; sm is for dense toolbars only.
  const sizes = {
    sm: 'h-9 px-3 text-xs gap-1.5',
    md: 'h-11 px-4 text-sm gap-2',
    lg: 'h-12 px-6 text-base gap-2',
  }[size];

  const cls = cx(
    'inline-flex items-center justify-center rounded-lg font-medium transition-colors duration-150',
    'cursor-pointer disabled:cursor-not-allowed disabled:opacity-50',
    variants,
    sizes,
    fullWidth && 'w-full',
    className
  );

  const inner = (
    <>
      {loading ? (
        <Loader2 className="h-4 w-4 shrink-0 animate-spin" aria-hidden="true" />
      ) : (
        Icon && <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
      )}
      {children}
      {IconRight && !loading && <IconRight className="h-4 w-4 shrink-0" aria-hidden="true" />}
    </>
  );

  if (href) {
    return (
      <a href={href} target="_blank" rel="noreferrer" className={cls} {...rest}>
        {inner}
      </a>
    );
  }
  return (
    <button type={type} onClick={onClick} disabled={disabled || loading} className={cls} {...rest}>
      {inner}
    </button>
  );
};

/* --- Badge --------------------------------------------------------------- */

export const Badge: React.FC<{
  children: React.ReactNode;
  tone?: Tone;
  icon?: LucideIcon;
  solid?: boolean;
  mono?: boolean;
  className?: string;
}> = ({ children, tone = 'neutral', icon: Icon, solid, mono, className }) => (
  <span
    className={cx(
      'inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium whitespace-nowrap',
      solid ? toneSolid[tone] : cx('border', toneSoft[tone]),
      mono && 'font-mono',
      className
    )}
  >
    {Icon && <Icon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />}
    {children}
  </span>
);

/* --- Callout ------------------------------------------------------------- */

export const Callout: React.FC<{
  tone?: Tone;
  title?: string;
  icon?: LucideIcon;
  children: React.ReactNode;
  className?: string;
}> = ({ tone = 'info', title, icon, children, className }) => {
  const Icon = icon ?? toneIcon[tone];
  return (
    <div className={cx('rounded-xl border p-4', toneSoft[tone], className)}>
      <div className="flex gap-3">
        <Icon className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
        <div className="min-w-0 flex-1">
          {title && <p className="mb-1 font-semibold">{title}</p>}
          <div className="measure text-sm leading-relaxed">{children}</div>
        </div>
      </div>
    </div>
  );
};

/* --- Code ---------------------------------------------------------------- */

export const InlineCode: React.FC<{ children: React.ReactNode; className?: string }> = ({
  children,
  className,
}) => (
  <code
    className={cx(
      'rounded-md border border-border bg-surface-sunken px-1.5 py-0.5 font-mono text-[0.9em] text-fg',
      className
    )}
  >
    {children}
  </code>
);

export const CopyButton: React.FC<{ value: string; label?: string; className?: string }> = ({
  value,
  label = 'Copy to clipboard',
  className,
}) => {
  const [copied, setCopied] = useState(false);

  const copy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
    } catch {
      /* Clipboard can be blocked by permissions; still confirm the intent. */
    }
    setCopied(true);
  }, [value]);

  useEffect(() => {
    if (!copied) return;
    const t = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(t);
  }, [copied]);

  return (
    <button
      type="button"
      onClick={copy}
      aria-label={copied ? 'Copied' : label}
      className={cx(
        'inline-flex h-9 w-9 shrink-0 cursor-pointer items-center justify-center rounded-md',
        'text-term-muted transition-colors hover:bg-term-bg-soft hover:text-term-fg',
        className
      )}
    >
      {copied ? (
        <Check className="h-4 w-4 text-term-success" aria-hidden="true" />
      ) : (
        <Copy className="h-4 w-4" aria-hidden="true" />
      )}
      <span className="sr-only" aria-live="polite">
        {copied ? 'Copied to clipboard' : ''}
      </span>
    </button>
  );
};

export const CodeBlock: React.FC<{
  code: string;
  label?: string;
  copyable?: boolean;
  className?: string;
}> = ({ code, label, copyable = true, className }) => (
  <div className={cx('overflow-hidden rounded-xl border border-border bg-term-bg', className)}>
    {(label || copyable) && (
      <div className="flex items-center justify-between border-b border-term-bg-soft py-1.5 pl-4 pr-1.5">
        <span className="font-mono text-2xs uppercase tracking-widest text-term-muted">
          {label ?? 'shell'}
        </span>
        {copyable && <CopyButton value={code} />}
      </div>
    )}
    <pre className="thin-scrollbar overflow-x-auto p-4 font-mono text-sm leading-relaxed text-term-fg">
      <code>{code}</code>
    </pre>
  </div>
);

/* Terminal surface — prompt + output, no chrome assumptions. */
export const Terminal: React.FC<{
  children: React.ReactNode;
  title?: string;
  className?: string;
}> = ({ children, title, className }) => (
  <div className={cx('overflow-hidden rounded-xl border border-border bg-term-bg', className)}>
    <div className="flex items-center gap-2 border-b border-term-bg-soft px-4 py-2.5">
      <span className="flex gap-1.5" aria-hidden="true">
        <span className="h-2.5 w-2.5 rounded-full bg-term-danger/70" />
        <span className="h-2.5 w-2.5 rounded-full bg-term-warn/70" />
        <span className="h-2.5 w-2.5 rounded-full bg-term-success/70" />
      </span>
      {title && (
        <span className="truncate font-mono text-2xs uppercase tracking-widest text-term-muted">
          {title}
        </span>
      )}
    </div>
    <div className="thin-scrollbar overflow-x-auto p-4 font-mono text-sm leading-relaxed text-term-fg">
      {children}
    </div>
  </div>
);

/* --- Progress ------------------------------------------------------------ */

export const ProgressBar: React.FC<{
  value: number;
  max?: number;
  label?: string;
  tone?: Tone;
  showValue?: boolean;
  className?: string;
}> = ({ value, max = 100, label, tone = 'accent', showValue = true, className }) => {
  const pct = max > 0 ? Math.min(100, Math.max(0, (value / max) * 100)) : 0;
  return (
    <div className={className}>
      {(label || showValue) && (
        <div className="mb-1.5 flex items-baseline justify-between gap-3">
          {label && <span className="text-sm font-medium text-fg-muted">{label}</span>}
          {showValue && (
            <span className="tabular font-mono text-xs text-fg-subtle">
              {value}/{max}
            </span>
          )}
        </div>
      )}
      <div
        role="progressbar"
        aria-valuenow={value}
        aria-valuemin={0}
        aria-valuemax={max}
        aria-label={label}
        className="h-2 w-full overflow-hidden rounded-full bg-surface-sunken"
      >
        <div
          className={cx('h-full rounded-full transition-all duration-500', toneSolid[tone])}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
};

/* --- Steps --------------------------------------------------------------- */

export const StepList: React.FC<{ children: React.ReactNode; className?: string }> = ({
  children,
  className,
}) => <ol className={cx('relative space-y-4', className)}>{children}</ol>;

export const Step: React.FC<{
  index: number;
  title: string;
  state?: 'pending' | 'current' | 'done';
  children?: React.ReactNode;
  onSelect?: () => void;
  last?: boolean;
}> = ({ index, title, state = 'pending', children, onSelect, last }) => {
  const marker = {
    pending: 'border-border bg-surface text-fg-subtle',
    current: 'border-accent bg-accent text-accent-fg',
    done: 'border-success bg-success text-fg-inverse',
  }[state];

  const Wrapper = onSelect ? 'button' : 'div';

  return (
    <li className="relative flex gap-4">
      <div className="flex flex-col items-center">
        <span
          className={cx(
            'flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2',
            'tabular font-mono text-xs font-semibold transition-colors duration-200',
            marker
          )}
        >
          {state === 'done' ? <Check className="h-4 w-4" aria-hidden="true" /> : index}
        </span>
        {!last && <span className="mt-1 w-0.5 flex-1 rounded-full bg-border" aria-hidden="true" />}
      </div>

      <Wrapper
        {...(onSelect ? { type: 'button' as const, onClick: onSelect } : {})}
        className={cx(
          'min-w-0 flex-1 pb-6 text-left',
          onSelect && 'cursor-pointer rounded-lg transition-colors hover:opacity-80'
        )}
      >
        <p
          className={cx(
            'text-base font-semibold',
            state === 'pending' ? 'text-fg-muted' : 'text-fg'
          )}
        >
          {title}
          {state === 'current' && <span className="sr-only"> (current step)</span>}
          {state === 'done' && <span className="sr-only"> (completed)</span>}
        </p>
        {children && <div className="mt-2 text-sm text-fg-muted">{children}</div>}
      </Wrapper>
    </li>
  );
};

/* --- Tabs ---------------------------------------------------------------- */

export function Tabs<T extends string>({
  tabs,
  active,
  onChange,
  label,
  wrap = false,
  className,
}: {
  tabs: { id: T; label: string; icon?: LucideIcon }[];
  active: T;
  onChange: (id: T) => void;
  label: string;
  /** Wrap onto multiple lines instead of scrolling horizontally — use in narrow columns. */
  wrap?: boolean;
  className?: string;
}) {
  const refs = useRef<(HTMLButtonElement | null)[]>([]);

  const onKeyDown = (e: React.KeyboardEvent) => {
    const i = tabs.findIndex((t) => t.id === active);
    let next = -1;
    if (e.key === 'ArrowRight') next = (i + 1) % tabs.length;
    if (e.key === 'ArrowLeft') next = (i - 1 + tabs.length) % tabs.length;
    if (e.key === 'Home') next = 0;
    if (e.key === 'End') next = tabs.length - 1;
    if (next < 0) return;
    e.preventDefault();
    onChange(tabs[next].id);
    refs.current[next]?.focus();
  };

  return (
    <div
      role="tablist"
      aria-label={label}
      onKeyDown={onKeyDown}
      className={cx(
        'flex',
        // A segmented control that wraps reads as a broken box, so in wrap mode
        // drop the enclosing track and render the options as filter chips.
        wrap
          ? 'flex-wrap gap-1.5'
          : 'gap-1 rounded-lg border border-border bg-surface-sunken p-1 no-scrollbar overflow-x-auto',
        className
      )}
    >
      {tabs.map((t, i) => {
        const on = t.id === active;
        return (
          <button
            key={t.id}
            ref={(el) => {
              refs.current[i] = el;
            }}
            role="tab"
            aria-selected={on}
            tabIndex={on ? 0 : -1}
            onClick={() => onChange(t.id)}
            className={cx(
              'flex h-9 shrink-0 cursor-pointer items-center gap-1.5 text-sm font-medium',
              'transition-colors duration-150',
              wrap
                ? on
                  ? 'rounded-full border border-accent-border bg-accent-soft px-4 text-accent-text'
                  : 'rounded-full border border-border bg-surface px-4 text-fg-muted hover:bg-surface-sunken hover:text-fg'
                : on
                  ? 'rounded-md bg-surface px-4 text-fg shadow-xs'
                  : 'rounded-md px-4 text-fg-muted hover:bg-surface/60 hover:text-fg'
            )}
          >
            {t.icon && <t.icon className="h-4 w-4 shrink-0" aria-hidden="true" />}
            {t.label}
          </button>
        );
      })}
    </div>
  );
}

/* --- Empty state --------------------------------------------------------- */

export const EmptyState: React.FC<{
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}> = ({ icon: Icon = Info, title, description, action, className }) => (
  <div
    className={cx(
      'flex flex-col items-center justify-center rounded-xl border border-dashed border-border-strong',
      'bg-surface px-6 py-12 text-center',
      className
    )}
  >
    <Icon className="mb-3 h-8 w-8 text-fg-faint" aria-hidden="true" />
    <p className="text-base font-semibold text-fg">{title}</p>
    {description && <p className="measure-tight mt-1.5 text-sm text-fg-muted">{description}</p>}
    {action && <div className="mt-5">{action}</div>}
  </div>
);

/* --- Lesson footer ------------------------------------------------------- */

export const LessonFooter: React.FC<{
  onNext: () => void;
  nextLabel: string;
  nextDescription?: string;
}> = ({ onNext, nextLabel, nextDescription }) => (
  <Card as="button" onClick={onNext} padding="lg" className="!border-accent-border bg-accent-soft">
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <p className="font-mono text-xs font-medium uppercase tracking-widest text-accent-text">
          Next lesson
        </p>
        <p className="mt-1.5 text-xl font-semibold text-fg">{nextLabel}</p>
        {nextDescription && (
          <p className="measure mt-1 text-sm text-fg-muted">{nextDescription}</p>
        )}
      </div>
      <ArrowRight
        className="h-6 w-6 shrink-0 text-accent-text transition-transform duration-200 group-hover:translate-x-1"
        aria-hidden="true"
      />
    </div>
  </Card>
);
