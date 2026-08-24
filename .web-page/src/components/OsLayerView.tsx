import React from 'react';
import { ArrowRight, CircleSlash, FolderOpen } from 'lucide-react';
import { TabType } from '../types';
import { getOsLayerPage } from '../data/osLayersData';
import {
  Badge,
  Callout,
  Card,
  InlineCode,
  PageHeader,
  PageShell,
  Prose,
  Section,
  Step,
  StepList,
} from './ui/primitives';

interface OsLayerViewProps {
  /** Which layer to render — 'company-os' or 'team-os'. */
  tab: TabType;
  onNavigateTab: (tab: TabType) => void;
}

/**
 * One page, two entries. Company OS and Team OS are explained with the same
 * four questions in the same order — what is it, what does it do, how does it
 * work, why do you need it — so the layers can be compared without re-reading.
 */
export const OsLayerView: React.FC<OsLayerViewProps> = ({ tab, onNavigateTab }) => {
  const page = getOsLayerPage(tab);
  if (!page) return null;

  return (
    <PageShell>
      <PageHeader eyebrow={page.eyebrow} title={page.title} lead={page.lead} icon={page.icon} />

      {/* --- What is it ---------------------------------------------------- */}
      <Section title="What it is" description="In plain terms, before any commands.">
        <div className="grid items-start gap-6 lg:grid-cols-2">
          <Prose>
            {page.what.paragraphs.map((p) => (
              <p key={p.slice(0, 40)}>{p}</p>
            ))}
          </Prose>

          <Card padding="lg">
            <p className="mb-4 flex items-center gap-2 font-mono text-xs font-medium uppercase tracking-widest text-fg-subtle">
              <FolderOpen className="h-3.5 w-3.5" aria-hidden="true" />
              {page.what.ownsTitle}
            </p>
            <ul className="space-y-4">
              {page.what.owns.map((o) => (
                <li key={o.path}>
                  <InlineCode>{o.path}</InlineCode>
                  <p className="mt-1.5 text-sm text-fg-muted">{o.description}</p>
                </li>
              ))}
            </ul>
          </Card>
        </div>
      </Section>

      {/* --- Why you need it ------------------------------------------------ */}
      <Section
        title="Why you need it"
        description="Left: what happens without it. Right: what replaces it."
      >
        <div className="space-y-3">
          {page.why.map((w) => (
            <Card key={w.pain} padding="lg">
              <div className="grid gap-4 sm:grid-cols-2 sm:items-center">
                <div className="flex gap-3">
                  <Badge tone="danger">Today</Badge>
                  <p className="min-w-0 text-sm leading-relaxed text-fg-muted">{w.pain}</p>
                </div>
                <div className="flex gap-3 border-t border-border pt-4 sm:border-l sm:border-t-0 sm:pl-6 sm:pt-0">
                  <Badge tone="success">Instead</Badge>
                  <p className="min-w-0 text-sm leading-relaxed text-fg">{w.instead}</p>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </Section>

      {/* --- What does it do ----------------------------------------------- */}
      <Section title="What it does" description="The jobs it takes off your hands.">
        <div className="grid gap-4 sm:grid-cols-2">
          {page.does.map((d) => (
            <Card key={d.title} padding="lg">
              <h3 className="text-base font-semibold text-fg">{d.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-fg-muted">{d.description}</p>
            </Card>
          ))}
        </div>
      </Section>

      {/* --- How does it work ---------------------------------------------- */}
      <Section
        title="How it works"
        description="The loop, in order. Every step is a command you can run today."
      >
        <StepList>
          {page.how.map((h, i) => (
            <Step
              key={h.title}
              index={i + 1}
              title={h.title}
              state={i === 0 ? 'current' : 'pending'}
              last={i === page.how.length - 1}
            >
              {h.command && (
                <p className="mb-2">
                  <InlineCode>{h.command}</InlineCode>
                </p>
              )}
              <p className="measure leading-relaxed">{h.description}</p>
            </Step>
          ))}
        </StepList>
      </Section>

      {/* --- What it is not -------------------------------------------------- */}
      <Section title="What it is not" description="The three things people assume, corrected.">
        <div className="space-y-3">
          {page.notThis.map((n) => (
            <Callout key={n.slice(0, 40)} tone="warn" icon={CircleSlash}>
              {n}
            </Callout>
          ))}
        </div>
      </Section>

      {/* --- Handoff to the other layer -------------------------------------- */}
      <Card
        as="button"
        onClick={() => onNavigateTab(page.handoff.tab)}
        padding="lg"
        tone={page.tone === 'accent' ? 'scope' : 'accent'}
      >
        <div className="flex items-center justify-between gap-4">
          <div className="min-w-0 text-left">
            <p className="font-mono text-xs font-medium uppercase tracking-widest text-fg-subtle">
              The other layer
            </p>
            <p className="mt-1.5 text-xl font-semibold text-fg">{page.handoff.label}</p>
            <p className="measure mt-1 text-sm text-fg-muted">{page.handoff.description}</p>
          </div>
          <ArrowRight className="h-6 w-6 shrink-0 text-fg-muted" aria-hidden="true" />
        </div>
      </Card>
    </PageShell>
  );
};
