import React, { useState } from 'react';
import { Search, Layers } from 'lucide-react';
import { EXIT_CODES_DATA } from '../data/commandsData';
import { TROUBLESHOOTING_DATA } from '../data/troubleshootingData';
import { PageShell, PageHeader, Section, Card, Badge, Tabs, EmptyState } from './ui';

interface PrecedenceLayer {
  order: number;
  name: string;
  status: 'wired' | 'unwired';
  path: string;
}

const PRECEDENCE_LAYERS: PrecedenceLayer[] = [
  { order: 1, name: 'CLI Flag (Implemented)', status: 'wired', path: 'company-os --root /abs/path ...' },
  {
    order: 2,
    name: 'Environment Variable (Implemented)',
    status: 'wired',
    path: 'export COMPANY_OS_WORKSPACE_ROOT=/abs/path',
  },
  {
    order: 3,
    name: 'Repo-Local Override (Specified in Spec)',
    status: 'unwired',
    path: '.company-os.local.yaml (git-ignored)',
  },
  {
    order: 4,
    name: 'User-Level Config (Specified in Spec)',
    status: 'unwired',
    path: '~/.company-os/config.yaml',
  },
  {
    order: 5,
    name: 'Committed Shared Config (Specified in Spec)',
    status: 'unwired',
    path: 'config/repositories.yaml',
  },
  { order: 6, name: 'Built-in Default (Implemented)', status: 'wired', path: 'Current working directory (cwd)' },
];

export const ReferenceMatrix: React.FC = () => {
  const [activeSection, setActiveSection] = useState<'troubleshooting' | 'precedence' | 'exitcodes'>(
    'troubleshooting'
  );
  const [searchQuery, setSearchQuery] = useState('');

  const filteredTroubleshooting = TROUBLESHOOTING_DATA.filter(
    (item) =>
      item.symptom.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.cause.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.fix.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <PageShell width="wide">
      <PageHeader
        eyebrow="Lesson 8 of 9"
        title="Configuration Reference & Troubleshooting"
        lead="A fast lookup for fixing errors and understanding configuration order: a searchable troubleshooting directory, the workspace-root precedence hierarchy, and the full exit code contract."
        icon={Layers}
      />

      <Tabs
        label="Reference matrix section"
        tabs={[
          { id: 'troubleshooting', label: `Troubleshooting (${TROUBLESHOOTING_DATA.length})` },
          { id: 'precedence', label: 'Precedence layers' },
          { id: 'exitcodes', label: 'Exit codes' },
        ]}
        active={activeSection}
        onChange={(id) => setActiveSection(id as typeof activeSection)}
      />

      {activeSection === 'troubleshooting' && (
        <Section
          title="Troubleshooting directory"
          description="Type any symptom, error message, or tool name to filter — e.g. &quot;expired deviation&quot;, &quot;reality doc&quot;, or &quot;python&quot;."
        >
          <div className="space-y-4">
            <div className="relative">
              <label htmlFor="troubleshoot-search" className="sr-only">
                Search symptom, error message, or fix
              </label>
              <Search
                className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-fg-subtle"
                aria-hidden="true"
              />
              <input
                id="troubleshoot-search"
                type="text"
                placeholder="Search symptom, error message, or fix…"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-11 w-full rounded-xl border border-border bg-surface pl-9 pr-4 font-mono text-sm text-fg shadow-xs focus-visible:outline-none"
              />
            </div>

            {filteredTroubleshooting.length === 0 ? (
              <EmptyState
                icon={Search}
                title="No matching symptoms"
                description={`Nothing in the troubleshooting directory matches "${searchQuery}".`}
              />
            ) : (
              <div className="space-y-3">
                {filteredTroubleshooting.map((item) => (
                  <Card key={item.id} className="space-y-2">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <Badge tone="danger" mono>
                        {item.symptom}
                      </Badge>
                      <Badge>
                        {item.tool} • {item.category}
                      </Badge>
                    </div>
                    <div className="space-y-1 text-xs">
                      <p>
                        <span className="font-bold text-fg">Root cause: </span>
                        <span className="text-fg-muted">{item.cause}</span>
                      </p>
                      <p>
                        <span className="font-bold text-success-text">Recommended fix: </span>
                        <span className="font-mono font-medium text-success-text">{item.fix}</span>
                      </p>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </Section>
      )}

      {activeSection === 'precedence' && (
        <Section
          title="Workspace root precedence layers"
          description="Company OS resolves the workspace root in six specified layers (highest precedence wins)."
        >
          <div className="thin-scrollbar overflow-x-auto rounded-xl border border-border">
            <table className="w-full min-w-[560px] text-sm">
              <thead className="sticky top-0 bg-surface-sunken">
                <tr className="border-b border-border">
                  <th className="w-12 px-3 py-2.5 text-left font-semibold text-fg">#</th>
                  <th className="px-3 py-2.5 text-left font-semibold text-fg">Layer & path</th>
                  <th className="w-28 px-3 py-2.5 text-left font-semibold text-fg">Status</th>
                </tr>
              </thead>
              <tbody>
                {PRECEDENCE_LAYERS.map((layer) => (
                  <tr key={layer.order} className="border-b border-border last:border-0">
                    <td className="tabular px-3 py-3 align-top font-mono text-fg-subtle">{layer.order}</td>
                    <td className="px-3 py-3 align-top">
                      <p className="font-semibold text-fg">{layer.name}</p>
                      <code className="mt-0.5 block break-all font-mono text-xs text-fg-muted">{layer.path}</code>
                    </td>
                    <td className="px-3 py-3 align-top">
                      <Badge tone={layer.status === 'wired' ? 'success' : 'neutral'}>
                        {layer.status === 'wired' ? 'Wired' : 'Unwired'}
                      </Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}

      {activeSection === 'exitcodes' && (
        <Section
          title="Full exit code contract matrix"
          description="Every subcommand exits with one of eight codes. The exit code meaning is stable across reworded messages."
        >
          <div className="thin-scrollbar overflow-x-auto rounded-xl border border-border">
            <table className="w-full min-w-[720px] text-sm">
              <thead className="sticky top-0 bg-surface-sunken">
                <tr className="border-b border-border">
                  <th className="w-16 px-3 py-2.5 text-left font-semibold text-fg">Code</th>
                  <th className="px-3 py-2.5 text-left font-semibold text-fg">Meaning</th>
                  <th className="px-3 py-2.5 text-left font-semibold text-fg">When it occurs</th>
                  <th className="px-3 py-2.5 text-left font-semibold text-fg">Recommended action</th>
                </tr>
              </thead>
              <tbody>
                {EXIT_CODES_DATA.map((ec) => (
                  <tr key={ec.code} className="border-b border-border last:border-0">
                    <td className="px-3 py-3 align-top">
                      <Badge tone={ec.code === 0 ? 'success' : 'danger'} mono className="tabular">
                        {ec.code}
                      </Badge>
                    </td>
                    <td className="px-3 py-3 align-top font-semibold text-fg">{ec.meaning}</td>
                    <td className="px-3 py-3 align-top text-fg-muted">{ec.whenOccurs}</td>
                    <td className="px-3 py-3 align-top font-mono text-xs text-success-text">
                      {ec.recommendedAction}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}
    </PageShell>
  );
};
