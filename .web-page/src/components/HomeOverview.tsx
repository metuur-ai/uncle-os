import React from 'react';
import {
  Building2,
  FileCode2,
  ListCheck,
  ShieldCheck,
  GitBranch,
  Layers,
  ArrowRight,
  Rocket,
  HelpCircle,
  CheckCircle2,
  type LucideIcon,
} from 'lucide-react';
import { TabType } from '../types';
import { LESSONS, lessonNumber } from '../lessons';
import {
  HOME_HERO_CONTENT,
  DUAL_CORE_COMPARISON,
  LIFECYCLE_STEPS,
  QUICK_DIRECTORY_CARDS,
} from '../data/homeData';
import { WhyWhatHowCard } from './WhyWhatHowCard';
import { PageShell, Section, Card, Badge, Button } from './ui';

interface HomeOverviewProps {
  onNavigateTab: (tab: TabType) => void;
  onOpenGuide: () => void;
  isStandalone: boolean;
  setIsStandalone: (val: boolean) => void;
}

const LIFECYCLE_ICONS: Record<string, LucideIcon> = {
  FileCode2,
  ListCheck,
  ShieldCheck,
  Building2,
};

export const HomeOverview: React.FC<HomeOverviewProps> = ({
  onNavigateTab,
  onOpenGuide,
  isStandalone,
  setIsStandalone,
}) => {
  const tutorials = LESSONS.filter((l) => l.id !== 'home');

  return (
    <PageShell width="wide">
      {/* ================= HERO ============================================ */}
      <section className="relative overflow-hidden rounded-2xl border border-border bg-surface px-6 py-14 sm:px-10 lg:px-14 lg:py-20">
        {/* Decorative only — the grid never carries meaning. */}
        <div className="bg-grid pointer-events-none absolute inset-0 opacity-60" aria-hidden="true" />
        <div
          className="pointer-events-none absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-surface to-transparent"
          aria-hidden="true"
        />

        <div className="relative">
          <Badge tone="accent" icon={Rocket}>
            {HOME_HERO_CONTENT.badge}
          </Badge>

          <h1 className="measure-wide mt-5 text-4xl font-bold tracking-tight text-fg sm:text-5xl">
            {HOME_HERO_CONTENT.titlePrefix}
            <span className="text-accent-text">{HOME_HERO_CONTENT.companyTitle}</span>
            {HOME_HERO_CONTENT.titleConnector}
            <span className="text-scope-text">{HOME_HERO_CONTENT.teamTitle}</span>
          </h1>

          <p className="measure mt-5 text-lg leading-relaxed text-fg-muted">
            {HOME_HERO_CONTENT.description}
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:items-center">
            <Button size="lg" iconRight={ArrowRight} onClick={() => onNavigateTab('install')}>
              Start lesson 1 — Install &amp; Setup
            </Button>
            <Button variant="secondary" size="lg" icon={HelpCircle} onClick={onOpenGuide}>
              I&apos;m new — explain the jargon
            </Button>
          </div>

          <p className="mt-6 font-mono text-xs text-fg-subtle">
            9 interactive lessons · no signup · runs entirely in your browser
          </p>
        </div>
      </section>

      {/* ================= WHY / WHAT / HOW ================================ */}
      <WhyWhatHowCard onNavigate={onNavigateTab} onOpenGuide={onOpenGuide} />

      {/* ================= DUAL CORE ======================================= */}
      <Section
        title="Two cores, one workspace"
        description="Company OS holds the rules everyone inherits. Team OS is where a squad actually works. Pick a side below to switch the whole site into that point of view."
      >
        <div className="grid gap-4 lg:grid-cols-2">
          {DUAL_CORE_COMPARISON.map((core) => {
            const isTeam = core.id === 'team';
            const selected = isTeam === isStandalone;
            const Icon = isTeam ? GitBranch : Layers;

            return (
              <Card
                key={core.id}
                padding="lg"
                tone={selected ? (isTeam ? 'scope' : 'accent') : 'neutral'}
                className="flex flex-col"
              >
                <div className="flex flex-wrap items-center gap-3">
                  <span
                    className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-lg ${
                      isTeam ? 'bg-scope-soft text-scope-text' : 'bg-accent-soft text-accent-text'
                    }`}
                    aria-hidden="true"
                  >
                    <Icon className="h-5 w-5" />
                  </span>
                  <div className="min-w-0">
                    <h3 className="text-2xl font-semibold tracking-tight text-fg">{core.title}</h3>
                    <p className="font-mono text-xs text-fg-subtle">{core.yamlFile}</p>
                  </div>
                  <Badge tone={isTeam ? 'scope' : 'accent'} className="ml-auto">
                    {core.badge}
                  </Badge>
                </div>

                <p className="mt-4 leading-relaxed text-fg-muted">{core.summary}</p>

                <ul className="mt-5 space-y-3">
                  {core.keyPoints.map((point) => {
                    const [label, ...rest] = point.split(': ');
                    const body = rest.join(': ');
                    return (
                      <li key={point} className="flex gap-2.5">
                        <CheckCircle2
                          className={`mt-0.5 h-4 w-4 shrink-0 ${
                            isTeam ? 'text-scope-text' : 'text-accent-text'
                          }`}
                          aria-hidden="true"
                        />
                        <span className="text-sm leading-relaxed text-fg-muted">
                          <span className="font-semibold text-fg">{label}</span>
                          {body ? <>: {body}</> : null}
                        </span>
                      </li>
                    );
                  })}
                </ul>

                <div className="mt-auto flex flex-wrap items-center gap-2 pt-6">
                  <Button
                    variant={selected ? 'primary' : 'secondary'}
                    size="sm"
                    iconRight={ArrowRight}
                    onClick={() => {
                      setIsStandalone(isTeam);
                      onNavigateTab(core.targetTab);
                    }}
                  >
                    {core.actionText}
                  </Button>
                  {!selected && (
                    <Button variant="ghost" size="sm" onClick={() => setIsStandalone(isTeam)}>
                      Switch to this view
                    </Button>
                  )}
                  {selected && (
                    <span className="font-mono text-2xs uppercase tracking-widest text-fg-subtle">
                      Currently viewing
                    </span>
                  )}
                </div>

                <p className="mt-4 border-t border-border pt-4 font-mono text-xs text-fg-subtle">
                  {core.footerNote}
                </p>
              </Card>
            );
          })}
        </div>
      </Section>

      {/* ================= LIFECYCLE ======================================= */}
      <Section
        title="What a change looks like end to end"
        description="From an idea in a squad's head to a merged, indexed, searchable document — four stages."
      >
        <ol className="stagger grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {LIFECYCLE_STEPS.map((step) => {
            const Icon = LIFECYCLE_ICONS[step.iconName] ?? FileCode2;
            return (
              <li key={step.stepNumber}>
                <Card padding="md" className="h-full">
                  <div className="flex items-center gap-2.5">
                    <span
                      className="tabular flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-accent font-mono text-xs font-bold text-accent-fg"
                      aria-hidden="true"
                    >
                      {step.stepNumber}
                    </span>
                    <Icon className="h-4 w-4 text-fg-subtle" aria-hidden="true" />
                  </div>
                  <h3 className="mt-3 font-semibold text-fg">{step.title}</h3>
                  <p className="mt-1.5 text-sm leading-relaxed text-fg-muted">
                    {step.description}
                  </p>
                </Card>
              </li>
            );
          })}
        </ol>
      </Section>

      {/* ================= LESSON INDEX ==================================== */}
      <Section
        title="All 9 lessons"
        description="The lessons build on each other, but every one stands alone if you already know the basics."
      >
        <ol className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {tutorials.map((lesson) => (
            <li key={lesson.id}>
              <button
                type="button"
                onClick={() => onNavigateTab(lesson.id)}
                className="group flex h-full w-full cursor-pointer flex-col rounded-xl border border-border bg-surface p-4 text-left transition-colors duration-150 hover:border-accent-border hover:bg-surface-sunken"
              >
                <div className="flex items-center gap-2.5">
                  <span
                    className="tabular flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-surface-sunken font-mono text-2xs font-semibold text-fg-muted transition-colors group-hover:border-accent group-hover:bg-accent group-hover:text-accent-fg"
                    aria-hidden="true"
                  >
                    {lessonNumber(lesson.id)}
                  </span>
                  <lesson.icon className="h-4 w-4 shrink-0 text-fg-subtle" aria-hidden="true" />
                  <span className="truncate font-semibold text-fg">{lesson.label}</span>
                  <ArrowRight
                    className="ml-auto h-4 w-4 shrink-0 text-fg-faint transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-accent-text"
                    aria-hidden="true"
                  />
                </div>
                <p className="mt-2 text-sm leading-relaxed text-fg-muted">{lesson.whyText}</p>
              </button>
            </li>
          ))}
        </ol>
      </Section>

      {/* ================= QUICK DIRECTORY ================================= */}
      <Section
        title="Jump straight to a tool"
        description="Already know your way around? These are the four things people open most."
      >
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {QUICK_DIRECTORY_CARDS.map((card) => (
            <Card key={card.title} padding="md" className="flex h-full flex-col">
              <h3 className="font-semibold text-fg">{card.title}</h3>
              <p className="mt-1.5 text-sm leading-relaxed text-fg-muted">{card.description}</p>
              <Button
                variant="ghost"
                size="sm"
                iconRight={ArrowRight}
                onClick={() => onNavigateTab(card.targetTab)}
                className="mt-auto self-start pt-4"
              >
                {card.actionText}
              </Button>
            </Card>
          ))}
        </div>
      </Section>
    </PageShell>
  );
};
