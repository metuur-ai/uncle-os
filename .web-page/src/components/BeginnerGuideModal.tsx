import React, { useId, useMemo, useState } from 'react';
import {
  BookOpen,
  Search,
  HelpCircle,
  UserCheck,
  GraduationCap,
  Sparkles,
  ChevronRight,
} from 'lucide-react';
import { Modal } from './ui/Modal';
import { Callout, Card, Tabs, InlineCode, EmptyState, Button, Badge, cx } from './ui';
import { GLOSSARY_ITEMS } from '../data/glossaryData';

interface BeginnerGuideModalProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab: (tab: string) => void;
}

type Section = 'primer' | 'glossary' | 'guide';

export const BeginnerGuideModal: React.FC<BeginnerGuideModalProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
}) => {
  const [activeSection, setActiveSection] = useState<Section>('primer');
  const [searchTerm, setSearchTerm] = useState('');
  const searchLabelId = useId();

  const filteredGlossary = useMemo(
    () =>
      GLOSSARY_ITEMS.filter(
        (item) =>
          item.term.toLowerCase().includes(searchTerm.toLowerCase()) ||
          item.plain.toLowerCase().includes(searchTerm.toLowerCase())
      ),
    [searchTerm]
  );

  const go = (tab: string) => {
    onClose();
    onNavigateTab(tab);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Plain English Helper & Glossary"
      description="Simple explanations for interns, new hires, and non-technical teammates."
      icon={BookOpen}
      size="lg"
    >
      <div className="space-y-5">
        <Tabs
          label="Guide sections"
          tabs={[
            { id: 'primer', label: 'How It Works', icon: HelpCircle },
            { id: 'glossary', label: 'Glossary', icon: Search },
            { id: 'guide', label: 'Role Guides', icon: GraduationCap },
          ]}
          active={activeSection}
          onChange={(id) => setActiveSection(id as Section)}
        />

        {activeSection === 'primer' && (
          <div className="space-y-5">
            <Callout tone="accent" icon={Sparkles} title="What is Company OS in simple terms?">
              Imagine if every team at a big company organized their projects, rules, and
              documents in completely different ways. It would be total chaos!{' '}
              <strong>Company OS</strong> is a unified, easy-to-follow system that keeps
              everyone on the exact same page using clear, structured files and automatic
              quality checks.
            </Callout>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <Card padding="sm" className="space-y-2">
                <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent-soft text-xs font-bold text-accent-text">
                  1
                </span>
                <h4 className="text-sm font-semibold text-fg">Organized Folders</h4>
                <p className="text-xs text-fg-muted">
                  Every team has dedicated folders for discovery, requirements, and active
                  projects. No lost files in emails!
                </p>
              </Card>

              <Card padding="sm" className="space-y-2">
                <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent-soft text-xs font-bold text-accent-text">
                  2
                </span>
                <h4 className="text-sm font-semibold text-fg">Automated Safety Checks</h4>
                <p className="text-xs text-fg-muted">
                  Before work is approved, <InlineCode>company-os validate</InlineCode>{' '}
                  automatically scans for security, formatting, and missing info.
                </p>
              </Card>

              <Card padding="sm" className="space-y-2">
                <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent-soft text-xs font-bold text-accent-text">
                  3
                </span>
                <h4 className="text-sm font-semibold text-fg">Always Accurate</h4>
                <p className="text-xs text-fg-muted">
                  "Representation of Reality": When code changes, docs must change too. No
                  outdated documentation!
                </p>
              </Card>
            </div>

            <div className="flex justify-end">
              <Button onClick={() => go('workflows')} iconRight={ChevronRight}>
                Try Interactive Workflow Playground
              </Button>
            </div>
          </div>
        )}

        {activeSection === 'glossary' && (
          <div className="space-y-4">
            <div>
              <label htmlFor={searchLabelId} className="mb-1.5 block text-sm font-medium text-fg">
                Search glossary terms
              </label>
              <div className="relative">
                <Search
                  className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-fg-subtle"
                  aria-hidden="true"
                />
                <input
                  id={searchLabelId}
                  type="text"
                  placeholder="e.g. PRD, Gate, Deviation..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className={cx(
                    'h-11 w-full rounded-lg border border-border bg-surface pl-10 pr-4 text-sm text-fg',
                    'placeholder:text-fg-subtle focus:outline-none'
                  )}
                />
              </div>
              <p className="sr-only" aria-live="polite">
                {searchTerm
                  ? `${filteredGlossary.length} matching terms`
                  : `${GLOSSARY_ITEMS.length} terms available`}
              </p>
            </div>

            {filteredGlossary.length === 0 ? (
              <EmptyState
                icon={Search}
                title="No matching terms"
                description={`Nothing matches "${searchTerm}". Try a different word.`}
                action={
                  <Button variant="secondary" size="sm" onClick={() => setSearchTerm('')}>
                    Clear search
                  </Button>
                }
              />
            ) : (
              <dl className="space-y-2.5">
                {filteredGlossary.map((item) => (
                  <div key={item.term} className="rounded-lg border border-border bg-surface-sunken p-3.5">
                    <dt className="text-sm font-semibold text-fg">{item.term}</dt>
                    <dd className="mt-1 text-xs leading-relaxed text-fg-muted">{item.plain}</dd>
                  </div>
                ))}
              </dl>
            )}
          </div>
        )}

        {activeSection === 'guide' && (
          <div className="space-y-6">
            <div className="space-y-3">
              <Callout tone="success" icon={GraduationCap} title="For interns & new engineers (3-step cheat sheet)">
                Company OS makes it impossible to break things secretly. Follow these 3 simple
                steps when working on a task:
              </Callout>

              <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
                <Card padding="sm" className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Badge tone="accent">1</Badge>
                    <h4 className="text-xs font-semibold text-fg">Check specs first</h4>
                  </div>
                  <p className="text-2xs leading-relaxed text-fg-muted">
                    Use <strong>Local Search</strong> to check existing components and specs
                    before building anything new.
                  </p>
                </Card>

                <Card padding="sm" className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Badge tone="accent">2</Badge>
                    <h4 className="text-xs font-semibold text-fg">Run validation</h4>
                  </div>
                  <p className="text-2xs leading-relaxed text-fg-muted">
                    Run <InlineCode>company-os validate</InlineCode> in the{' '}
                    <strong>Terminal Explorer</strong>. If a gate fails, fix the reported error.
                  </p>
                </Card>

                <Card padding="sm" className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Badge tone="accent">3</Badge>
                    <h4 className="text-xs font-semibold text-fg">Update docs</h4>
                  </div>
                  <p className="text-2xs leading-relaxed text-fg-muted">
                    Bump the <InlineCode>updated: YYYY-MM-DD</InlineCode> date in the reality
                    document so docs match the code!
                  </p>
                </Card>
              </div>

              <div className="flex justify-end">
                <Button size="sm" onClick={() => go('cli')} iconRight={ChevronRight}>
                  Try Terminal Explorer
                </Button>
              </div>
            </div>

            <div className="border-t border-border" />

            <div className="space-y-3">
              <Callout tone="warn" icon={UserCheck} title="For non-technical managers & reviewers">
                You don't need to write code to understand project health. Company OS gives you
                full transparent visibility:
              </Callout>

              <div className="space-y-2.5">
                <Card padding="sm" className="space-y-1">
                  <p className="text-xs font-semibold text-fg">1. How do I know if a project is safe?</p>
                  <p className="text-xs leading-relaxed text-fg-muted">
                    Look at the <strong>Validation Gates</strong> tab. If all 8 gates show PASS
                    (green checkmarks), the project meets all compliance and quality standards.
                  </p>
                </Card>

                <Card padding="sm" className="space-y-1">
                  <p className="text-xs font-semibold text-fg">2. What if my team needs a rule exception?</p>
                  <p className="text-xs leading-relaxed text-fg-muted">
                    Visit the <strong>Governance Tiers</strong> tab. You can submit a
                    transparent, time-boxed <em>Deviation</em> or <em>Exception</em> with
                    rationale without breaking the system.
                  </p>
                </Card>

                <Card padding="sm" className="space-y-1">
                  <p className="text-xs font-semibold text-fg">3. Where do I test my knowledge?</p>
                  <p className="text-xs leading-relaxed text-fg-muted">
                    Try the 8-question <strong>Mastery Check</strong> quiz. It explains every
                    answer in plain language so you feel confident discussing project governance.
                  </p>
                </Card>
              </div>

              <div className="flex justify-end gap-2 pt-1">
                <Button variant="secondary" size="sm" onClick={() => go('validation')}>
                  View Safety Gates
                </Button>
                <Button size="sm" onClick={() => go('governance')} iconRight={ChevronRight}>
                  Explore Governance Tiers
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
};
