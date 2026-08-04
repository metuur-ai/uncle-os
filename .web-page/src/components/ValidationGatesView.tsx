import React, { useId, useState } from 'react';
import { VALIDATION_GATES_DATA } from '../data/validationGatesData';
import { ValidationGate } from '../types';
import {
  CheckSquare,
  CheckCircle2,
  XCircle,
  Wrench,
  Globe,
  Ghost,
} from 'lucide-react';
import { PageShell, PageHeader, Section, Card, Callout, Badge, Terminal } from './ui';
import { getLesson } from '../lessons';

export const ValidationGatesView: React.FC = () => {
  const lesson = getLesson('validation');
  const [selectedGate, setSelectedGate] = useState<ValidationGate>(VALIDATION_GATES_DATA[0]);
  const detailId = useId();

  return (
    <PageShell>
      <PageHeader
        eyebrow="Lesson 6 of 9"
        title={lesson.label}
        lead={lesson.whyText}
        icon={CheckSquare}
      />

      <Section
        title="Pick a gate to inspect"
        description="company-os validate runs these gates in order. A plain monorepo runs 7; a federated workspace (workspace.yaml present) runs all 8."
      >
        <div
          role="group"
          aria-label="Validation gates"
          className="grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-8"
        >
          {VALIDATION_GATES_DATA.map((gate) => {
            const isSelected = selectedGate.id === gate.id;
            return (
              <button
                key={gate.id}
                type="button"
                onClick={() => setSelectedGate(gate)}
                aria-pressed={isSelected}
                aria-controls={detailId}
                className={
                  'flex h-28 min-h-[44px] cursor-pointer flex-col justify-between rounded-xl border p-2.5 text-left transition-colors duration-150 ' +
                  (isSelected
                    ? 'border-accent-border bg-accent-soft text-fg shadow-xs'
                    : 'border-border bg-surface text-fg-muted hover:border-accent-border hover:bg-surface-sunken')
                }
              >
                <span className="tabular font-mono text-xs font-bold text-accent-text">
                  Gate {gate.id}
                </span>
                <span className="line-clamp-2 text-xs font-semibold leading-tight text-fg">
                  {gate.shortName}
                </span>
                {gate.federatedOnly && (
                  <Badge tone="scope" icon={Globe} className="w-fit">
                    Federated
                  </Badge>
                )}
              </button>
            );
          })}
        </div>
      </Section>

      <Section id={detailId} title={`Gate ${selectedGate.id}: ${selectedGate.shortName}`}>
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
          <Card className="space-y-4 lg:col-span-5">
            <div className="flex flex-wrap items-center gap-2 border-b border-border pb-3">
              <span className="tabular font-mono text-xs font-bold uppercase tracking-wider text-accent-text">
                [{selectedGate.id}/N] {selectedGate.name}
              </span>
            </div>

            <div className="flex flex-wrap gap-2">
              <Badge
                tone={selectedGate.absenceTolerant ? 'info' : 'neutral'}
                icon={selectedGate.absenceTolerant ? Ghost : CheckSquare}
              >
                {selectedGate.absenceTolerant ? 'Absence tolerant' : 'Strict contract'}
              </Badge>
              {selectedGate.federatedOnly && (
                <Badge tone="scope" icon={Globe}>
                  Federated workspaces only
                </Badge>
              )}
            </div>

            <p className="text-sm leading-relaxed text-fg-muted">{selectedGate.description}</p>

            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-fg">Validator rule checks</h3>
              <ul className="space-y-1.5 text-sm text-fg-muted">
                {selectedGate.checks.map((chk, idx) => (
                  <li key={idx} className="flex items-start gap-2">
                    <CheckCircle2 className="mt-0.5 h-3.5 w-3.5 shrink-0 text-accent-text" aria-hidden="true" />
                    <span>{chk}</span>
                  </li>
                ))}
              </ul>
            </div>
          </Card>

          <div className="space-y-4 lg:col-span-7">
            <Terminal title="example pass">
              <div className="mb-2 flex items-center gap-2 font-mono text-xs font-bold uppercase tracking-widest text-term-success">
                <CheckCircle2 className="h-4 w-4" aria-hidden="true" />
                <span>Pass</span>
              </div>
              <pre className="whitespace-pre-wrap leading-relaxed text-term-success">
                {selectedGate.examplePass}
              </pre>
            </Terminal>

            <Terminal title="example fail">
              <div className="mb-2 flex items-center gap-2 font-mono text-xs font-bold uppercase tracking-widest text-term-danger">
                <XCircle className="h-4 w-4" aria-hidden="true" />
                <span>Fail</span>
              </div>
              <pre className="whitespace-pre-wrap leading-relaxed text-term-danger">
                {selectedGate.exampleFail}
              </pre>
            </Terminal>

            <Callout tone="success" title="Recovery: fix action" icon={Wrench}>
              {selectedGate.fixAction}
            </Callout>
          </div>
        </div>
      </Section>
    </PageShell>
  );
};
