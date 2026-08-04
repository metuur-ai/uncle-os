import React, { useMemo, useState } from 'react';
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
  Terminal,
  StepList,
  Step,
  ProgressBar,
} from './ui';
import { WORKFLOW_SCENARIOS_DATA } from '../data/workflowsData';
import { WorkflowScenario, WorkflowStep } from '../types';
import { getLesson } from '../lessons';
import { PlayCircle, ArrowLeft, ArrowRight, RefreshCw, CheckCircle2 } from 'lucide-react';

type FormState = Record<string, string>;
type FormErrors = Record<string, string>;

export const WorkflowPlayground: React.FC = () => {
  const lesson = getLesson('workflows');
  const [scenario, setScenario] = useState<WorkflowScenario>(WORKFLOW_SCENARIOS_DATA[1]);
  const [stepIndex, setStepIndex] = useState(0);
  const [completed, setCompleted] = useState<Record<number, boolean>>({});
  const [formValues, setFormValues] = useState<FormState>({});
  const [formErrors, setFormErrors] = useState<FormErrors>({});
  const [running, setRunning] = useState(false);
  const [ranOutput, setRanOutput] = useState<string | null>(null);

  const step: WorkflowStep = scenario.steps[stepIndex];
  const isLastStep = stepIndex === scenario.steps.length - 1;

  const requiredFields = useMemo(
    () => step.interactiveForm?.fields.filter((f) => f.required) ?? [],
    [step]
  );
  const formValid = requiredFields.every((f) => (formValues[f.key] ?? '').trim().length > 0);

  const resetStepState = () => {
    setFormValues({});
    setFormErrors({});
    setRanOutput(null);
    setRunning(false);
  };

  const handleSelectScenario = (s: WorkflowScenario) => {
    setScenario(s);
    setStepIndex(0);
    setCompleted({});
    resetStepState();
  };

  const goToStep = (idx: number) => {
    setStepIndex(idx);
    resetStepState();
  };

  const handleBlurField = (key: string, required?: boolean) => {
    if (!required) return;
    const value = (formValues[key] ?? '').trim();
    const field = step.interactiveForm?.fields.find((f) => f.key === key);
    if (!value) {
      setFormErrors((prev) => ({
        ...prev,
        [key]: `${field?.label ?? key} is required — enter a value before running this step.`,
      }));
    } else {
      setFormErrors((prev) => {
        const next = { ...prev };
        delete next[key];
        return next;
      });
    }
  };

  const handleRun = () => {
    if (step.interactiveForm && !formValid) return;
    setRunning(true);
    setRanOutput(null);
    window.setTimeout(() => {
      setRunning(false);
      setRanOutput(step.mockTerminalOutput);
      setCompleted((prev) => ({ ...prev, [stepIndex]: true }));
    }, 600);
  };

  const handleNext = () => {
    if (!isLastStep) goToStep(stepIndex + 1);
  };

  const handleBack = () => {
    if (stepIndex > 0) goToStep(stepIndex - 1);
  };

  const handleRestart = () => {
    setStepIndex(0);
    setCompleted({});
    resetStepState();
  };

  return (
    <PageShell>
      <PageHeader eyebrow="Lesson 4 of 9" title={lesson.label} lead={lesson.whyText} icon={PlayCircle} />

      <Section title="Interactive Workflow Simulator">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
          {WORKFLOW_SCENARIOS_DATA.map((s) => {
            const isSelected = s.id === scenario.id;
            return (
              <Card
                key={s.id}
                as="button"
                onClick={() => handleSelectScenario(s)}
                padding="sm"
                tone={isSelected ? 'accent' : undefined}
              >
                <Badge tone={isSelected ? 'accent' : 'neutral'}>{s.badge}</Badge>
                <p className="mt-2 line-clamp-2 text-sm font-semibold text-fg">{s.title}</p>
                <p className="mt-1 line-clamp-2 text-xs text-fg-muted">{s.subtitle}</p>
              </Card>
            );
          })}
        </div>

        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-12">
          {/* Stepper navigation */}
          <div className="lg:col-span-4">
            <Card padding="sm" className="space-y-4">
              <div>
                <h3 className="text-sm font-bold text-fg">{scenario.title}</h3>
                <p className="mt-1 text-xs text-fg-muted">{scenario.description}</p>
              </div>

              <ProgressBar value={stepIndex + 1} max={scenario.steps.length} label="Progress" />

              <StepList>
                {scenario.steps.map((s, idx) => (
                  <Step
                    key={s.stepNumber}
                    index={s.stepNumber}
                    title={s.title}
                    state={idx === stepIndex ? 'current' : completed[idx] ? 'done' : 'pending'}
                    onSelect={() => goToStep(idx)}
                    last={idx === scenario.steps.length - 1}
                  >
                    <span className="font-mono">{s.command.split('\n')[0]}</span>
                  </Step>
                ))}
              </StepList>

              <Button variant="ghost" icon={RefreshCw} onClick={handleRestart} fullWidth>
                Reset Workflow
              </Button>
            </Card>
          </div>

          {/* Step workspace */}
          <div className="lg:col-span-8">
            <Card className="space-y-4">
              <div>
                <p className="font-mono text-xs font-semibold uppercase tracking-widest text-accent-text">
                  Step {step.stepNumber} of {scenario.steps.length}
                </p>
                <h3 className="mt-1 text-xl font-semibold text-fg">{step.title}</h3>
              </div>

              <p className="measure text-sm text-fg-muted">{step.description}</p>

              <Callout tone="warn" title="Governance Contract Rule:">
                {step.keyRule}
              </Callout>

              {step.fileAffected && (
                <p className="text-xs text-fg-subtle">
                  File affected: <InlineCode>{step.fileAffected}</InlineCode>
                </p>
              )}

              <CodeBlock code={step.command} label="command" />

              {step.interactiveForm && (
                <div className="space-y-3 rounded-xl border border-border bg-surface-sunken p-4">
                  {step.interactiveForm.fields.map((f) => (
                    <div key={f.key}>
                      <label
                        htmlFor={`field-${f.key}`}
                        className="mb-1 flex items-center gap-1 text-sm font-medium text-fg"
                      >
                        {f.label}
                        {f.required && (
                          <span className="text-danger-text" aria-hidden="true">
                            *
                          </span>
                        )}
                        {f.required && <span className="sr-only">(required)</span>}
                      </label>
                      <input
                        id={`field-${f.key}`}
                        type="text"
                        required={f.required}
                        placeholder={f.placeholder}
                        value={formValues[f.key] ?? ''}
                        onChange={(e) =>
                          setFormValues((prev) => ({ ...prev, [f.key]: e.target.value }))
                        }
                        onBlur={() => handleBlurField(f.key, f.required)}
                        aria-invalid={Boolean(formErrors[f.key])}
                        aria-describedby={formErrors[f.key] ? `field-${f.key}-error` : undefined}
                        className="h-11 w-full rounded-lg border border-border bg-surface px-3 text-sm text-fg focus-visible:outline-none"
                      />
                      {formErrors[f.key] && (
                        <p id={`field-${f.key}-error`} className="mt-1 text-xs text-danger-text">
                          {formErrors[f.key]}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              )}

              <div className="flex flex-wrap items-center gap-2">
                <Button
                  icon={PlayCircle}
                  onClick={handleRun}
                  loading={running}
                  disabled={Boolean(step.interactiveForm) && !formValid}
                >
                  Run Step Command
                </Button>
                {completed[stepIndex] && (
                  <Badge tone="success" icon={CheckCircle2}>
                    Done
                  </Badge>
                )}
              </div>

              <div aria-live="polite">
                {running ? (
                  <Terminal title={`${scenario.id} — step ${step.stepNumber}`}>
                    <p className="text-term-muted">Running…</p>
                  </Terminal>
                ) : ranOutput ? (
                  <Terminal title={`${scenario.id} — step ${step.stepNumber}`}>
                    <pre className="whitespace-pre-wrap">{ranOutput}</pre>
                  </Terminal>
                ) : null}
              </div>

              <div className="flex items-center justify-between gap-3 pt-2">
                <Button variant="secondary" icon={ArrowLeft} onClick={handleBack} disabled={stepIndex === 0}>
                  Back
                </Button>

                {isLastStep ? (
                  <Callout tone="success" title="Workflow Complete!" className="flex-1">
                    Scenario successfully completed.
                  </Callout>
                ) : (
                  <Button
                    variant="primary"
                    iconRight={ArrowRight}
                    onClick={handleNext}
                    disabled={!completed[stepIndex]}
                  >
                    Proceed to Step {stepIndex + 2}
                  </Button>
                )}
              </div>
            </Card>
          </div>
        </div>
      </Section>
    </PageShell>
  );
};
