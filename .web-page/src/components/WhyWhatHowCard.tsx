import React from 'react';
import { TabType } from '../types';
import { WHY_CONTENT, WHAT_POINTS, HOW_STEPS } from '../data/whyWhatHowData';
import { Section, Card, Badge, Button, StepList, Step } from './ui';
import {
  Layers,
  CheckCircle2,
  Terminal,
  XCircle,
  ArrowRight,
  HelpCircle,
  BookOpen,
  type LucideIcon,
} from 'lucide-react';

interface WhyWhatHowCardProps {
  onNavigate: (tab: TabType) => void;
  onOpenGuide: () => void;
}

const WHAT_ICONS: Record<string, LucideIcon> = {
  layers: Layers,
  check: CheckCircle2,
  terminal: Terminal,
};

/**
 * The why / what / how narrative. This used to render above every tab, so all
 * nine lessons opened with the same marketing block. It now appears once, on
 * home, and is laid out as real sections instead of nested tabs.
 */
export const WhyWhatHowCard: React.FC<WhyWhatHowCardProps> = ({ onNavigate, onOpenGuide }) => {
  return (
    <>
      {/* --- WHY: the problem, then the answer ---------------------------- */}
      <Section
        title="Why this exists"
        description="Two states of the world, side by side. The left is what most teams live with; the right is what changes when specs and code share a repository."
        actions={
          <Button variant="secondary" size="sm" icon={HelpCircle} onClick={onOpenGuide}>
            Plain-English glossary
          </Button>
        }
      >
        <div className="grid gap-4 md:grid-cols-2">
          <Card tone="danger" padding="lg">
            <Badge tone="danger" icon={XCircle} className="mb-3">
              Before
            </Badge>
            <h3 className="text-xl font-semibold text-fg">{WHY_CONTENT.problemTitle}</h3>
            <p className="measure mt-3 leading-relaxed text-fg-muted">
              {WHY_CONTENT.problemDescription}
            </p>
          </Card>

          <Card tone="success" padding="lg">
            <Badge tone="success" icon={CheckCircle2} className="mb-3">
              After
            </Badge>
            <h3 className="text-xl font-semibold text-fg">{WHY_CONTENT.solutionTitle}</h3>
            <p className="measure mt-3 leading-relaxed text-fg-muted">
              {WHY_CONTENT.solutionDescription}
            </p>
          </Card>
        </div>
      </Section>

      {/* --- WHAT: the three moving parts ---------------------------------- */}
      <Section
        title="What you get"
        description="Three pieces do the work: a two-layer rule system, automated checks that run before merge, and a single CLI that ties them together."
      >
        <div className="stagger grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {WHAT_POINTS.map((point) => {
            const Icon = WHAT_ICONS[point.iconType] ?? BookOpen;
            return (
              <Card key={point.title} padding="lg" className="flex flex-col">
                <span
                  className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-accent-soft text-accent-text"
                  aria-hidden="true"
                >
                  <Icon className="h-5 w-5" />
                </span>
                <h3 className="text-xl font-semibold text-fg">{point.title}</h3>
                <p className="mt-1 font-mono text-xs uppercase tracking-widest text-accent-text">
                  {point.subtitle}
                </p>
                <p className="mt-3 leading-relaxed text-fg-muted">{point.description}</p>
              </Card>
            );
          })}
        </div>
      </Section>

      {/* --- HOW: the recommended path ------------------------------------- */}
      <Section
        title="How to work through this site"
        description="Four checkpoints, in order. Each one opens an interactive lesson rather than a page of prose."
      >
        <Card padding="lg">
          <StepList>
            {HOW_STEPS.map((step, i) => (
              <Step
                key={step.stepNumber}
                index={step.stepNumber}
                title={step.title}
                state={i === 0 ? 'current' : 'pending'}
                last={i === HOW_STEPS.length - 1}
              >
                <p className="leading-relaxed">{step.description}</p>
                <Button
                  variant="secondary"
                  size="sm"
                  iconRight={ArrowRight}
                  onClick={() => onNavigate(step.targetTab as TabType)}
                  className="mt-3"
                >
                  {step.actionLabel}
                </Button>
              </Step>
            ))}
          </StepList>
        </Card>
      </Section>
    </>
  );
};
