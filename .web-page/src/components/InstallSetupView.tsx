import React, { useState } from 'react';
import {
  INSTALL_STEPS,
  INSTALL_ARTIFACTS,
  INSTALL_OPTIONS,
  INSTALL_PATHS,
  INSTALL_FAQS,
  COMPANION_TOOL,
  INSTALL_ONE_LINER,
} from '../data/installData';
import { getLesson } from '../lessons';
import {
  PageShell,
  PageHeader,
  Section,
  Card,
  Callout,
  Badge,
  Button,
  CodeBlock,
  InlineCode,
  StepList,
  Step,
  ProgressBar,
  Tabs,
  Prose,
} from './ui';
import {
  Download,
  Cpu,
  Sliders,
  HelpCircle,
  Search,
  ChevronDown,
  ExternalLink,
  CheckCircle2,
} from 'lucide-react';

const lesson = getLesson('install');
const verifyStep = INSTALL_STEPS.find((s) => s.stepNumber === 3) ?? INSTALL_STEPS[2];

export const InstallSetupView: React.FC = () => {
  const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set());
  const [activePath, setActivePath] = useState<string>(INSTALL_PATHS[0].id);
  const [openFaq, setOpenFaq] = useState<string | null>(INSTALL_FAQS[0].id);

  const firstIncomplete = INSTALL_STEPS.find((s) => !completedSteps.has(s.stepNumber))?.stepNumber;

  const toggleStep = (stepNumber: number) => {
    setCompletedSteps((prev) => {
      const next = new Set(prev);
      if (next.has(stepNumber)) next.delete(stepNumber);
      else next.add(stepNumber);
      return next;
    });
  };

  const selectedPath = INSTALL_PATHS.find((p) => p.id === activePath) ?? INSTALL_PATHS[0];

  return (
    <PageShell>
      <PageHeader eyebrow="Lesson 1 of 9" title={lesson.label} lead={lesson.whyText} icon={Download} />

      <Section title="The one-line install" description="Copy it, paste it, run it.">
        <CodeBlock code={INSTALL_ONE_LINER} label="install.sh" />
      </Section>

      <Section
        title="Step-by-step"
        description="Work down the list in order. Mark a step done once you've run it."
        actions={
          <ProgressBar
            value={completedSteps.size}
            max={INSTALL_STEPS.length}
            label="Progress"
            className="w-40"
          />
        }
      >
        <StepList>
          {INSTALL_STEPS.map((step, i) => {
            const state = completedSteps.has(step.stepNumber)
              ? 'done'
              : step.stepNumber === firstIncomplete
                ? 'current'
                : 'pending';
            return (
              <Step
                key={step.stepNumber}
                index={step.stepNumber}
                title={step.title}
                state={state}
                last={i === INSTALL_STEPS.length - 1}
              >
                <div className="space-y-3">
                  <p className="measure">{step.description}</p>
                  <CodeBlock code={step.command} label={`step ${step.stepNumber}`} />
                  <Callout tone="warn" title="Key rule">
                    {step.keyRule}
                  </Callout>
                  <CodeBlock code={step.mockTerminalOutput} label="expected output" copyable={false} />
                  <Button
                    variant={state === 'done' ? 'secondary' : 'primary'}
                    size="sm"
                    icon={state === 'done' ? CheckCircle2 : undefined}
                    onClick={() => toggleStep(step.stepNumber)}
                  >
                    {state === 'done' ? 'Marked done' : 'Mark step done'}
                  </Button>
                </div>
              </Step>
            );
          })}
        </StepList>
      </Section>

      <Callout tone="success" title="Verify it worked">
        Run <InlineCode>{verifyStep.command}</InlineCode> — you should see:
        <CodeBlock
          code={verifyStep.mockTerminalOutput}
          label="company-os --version"
          className="mt-3"
        />
      </Callout>

      <Section
        title="Three ways to install"
        description="Pick whichever matches how you got here."
      >
        <div className="space-y-4">
          <Tabs
            tabs={INSTALL_PATHS.map((p) => ({ id: p.id, label: p.label }))}
            active={activePath}
            onChange={setActivePath}
            label="Install method"
          />
          <Card>
            <div className="flex flex-wrap items-center gap-2">
              <Badge tone="accent" mono>
                {selectedPath.badge}
              </Badge>
              <span className="text-sm text-fg-subtle">Requires: {selectedPath.requirement}</span>
            </div>
            <Prose className="mt-3">
              <p>{selectedPath.summary}</p>
            </Prose>
            <CodeBlock code={selectedPath.commands.join('\n')} label={selectedPath.label} className="mt-3" />
          </Card>
        </div>
      </Section>

      <Section title="Release artifacts" description="One binary per platform, resolved by the installer.">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {INSTALL_ARTIFACTS.map((artifact) => (
            <Card key={artifact.filename} padding="sm" className="space-y-2">
              <div className="flex items-center gap-2 text-fg-subtle">
                <Cpu className="h-4 w-4 shrink-0" aria-hidden="true" />
                <span className="text-2xs font-mono uppercase tracking-widest">{artifact.detectedAs}</span>
              </div>
              <InlineCode className="block w-full break-all">{artifact.filename}</InlineCode>
              <p className="text-sm text-fg-muted">{artifact.runsOn}</p>
            </Card>
          ))}
        </div>
      </Section>

      <Section
        title="Installer options"
        description="Set as environment variables before the install command."
      >
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          {INSTALL_OPTIONS.map((opt) => (
            <Card key={opt.envVar} padding="sm" className="space-y-2">
              <div className="flex items-center justify-between gap-2">
                <InlineCode>{opt.envVar}</InlineCode>
                <Badge tone="neutral" mono>
                  {opt.defaultVal}
                </Badge>
              </div>
              <p className="text-sm text-fg-muted">{opt.description}</p>
              <CodeBlock code={opt.example} copyable={false} />
            </Card>
          ))}
        </div>
      </Section>

      <Section
        title="Gatekeeper, upgrades, and the ways this goes wrong"
        description="Common failure modes and their fix."
      >
        <div className="space-y-2">
          {INSTALL_FAQS.map((faq) => {
            const isOpen = openFaq === faq.id;
            const panelId = `faq-panel-${faq.id}`;
            return (
              <div
                key={faq.id}
                className={
                  isOpen
                    ? 'rounded-xl border border-accent-border bg-accent-soft transition-colors duration-150'
                    : 'rounded-xl border border-border bg-surface transition-colors duration-150'
                }
              >
                <button
                  type="button"
                  onClick={() => setOpenFaq(isOpen ? null : faq.id)}
                  aria-expanded={isOpen}
                  aria-controls={panelId}
                  className="flex min-h-11 w-full cursor-pointer items-center justify-between gap-3 p-4 text-left"
                >
                  <span className="flex items-center gap-2 text-sm font-semibold text-fg">
                    <HelpCircle className="h-4 w-4 shrink-0 text-fg-subtle" aria-hidden="true" />
                    {faq.question}
                  </span>
                  <ChevronDown
                    className={
                      isOpen
                        ? 'h-4 w-4 shrink-0 rotate-180 text-fg-subtle transition-transform duration-200'
                        : 'h-4 w-4 shrink-0 text-fg-subtle transition-transform duration-200'
                    }
                    aria-hidden="true"
                  />
                </button>
                {isOpen && (
                  <div id={panelId} className="space-y-3 px-4 pb-4">
                    <p className="measure text-sm text-fg-muted">{faq.answer}</p>
                    {faq.command && <CodeBlock code={faq.command} label="fix" />}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </Section>

      <Section
        title={`Optional companion: ${COMPANION_TOOL.name}`}
        description={COMPANION_TOOL.summary}
        actions={
          <Button
            variant="ghost"
            size="sm"
            iconRight={ExternalLink}
            href={COMPANION_TOOL.repo}
          >
            View repo
          </Button>
        }
      >
        <div className="flex items-center gap-2">
          <Search className="h-4 w-4 shrink-0 text-fg-subtle" aria-hidden="true" />
          <Badge tone="scope" mono>
            SEPARATE BINARY
          </Badge>
        </div>
        <StepList className="mt-4">
          {COMPANION_TOOL.steps.map((step, i) => (
            <Step
              key={step.command}
              index={i + 1}
              title={step.description}
              state="pending"
              last={i === COMPANION_TOOL.steps.length - 1}
            >
              <CodeBlock code={step.command} />
            </Step>
          ))}
        </StepList>
      </Section>

      <Callout tone="info" icon={Sliders} title="Not sure which path is right?">
        Start with the installer script above — it covers the common cases. The build-from-source and
        offline-checkout paths in <InlineCode>Three ways to install</InlineCode> exist for CI images and
        air-gapped environments.
      </Callout>
    </PageShell>
  );
};
