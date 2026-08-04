import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  cx,
  PageShell,
  PageHeader,
  Section,
  Card,
  Badge,
  Button,
  CodeBlock,
  InlineCode,
  Terminal,
  Tabs,
  EmptyState,
} from './ui';
import { CLI_COMMANDS_DATA, EXIT_CODES_DATA } from '../data/commandsData';
import { CliCommand } from '../types';
import { getLesson } from '../lessons';
import {
  Terminal as TerminalIcon,
  Search,
  Play,
  CheckCircle2,
  XCircle,
  AlertTriangle,
} from 'lucide-react';

const CATEGORIES = Array.from(
  new Set(CLI_COMMANDS_DATA.map((c) => c.category))
) as CliCommand['category'][];

type CategoryFilter = 'all' | CliCommand['category'];

export const CliTerminalView: React.FC = () => {
  const lesson = getLesson('cli');
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const [selected, setSelected] = useState<CliCommand>(CLI_COMMANDS_DATA[7]);
  const [outputMode, setOutputMode] = useState<'text' | 'json'>('text');
  const [running, setRunning] = useState(false);
  const [output, setOutput] = useState<string | null>(selected.expectedOutput);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return CLI_COMMANDS_DATA.filter((c) => {
      const inCategory = category === 'all' || c.category === category;
      const inQuery =
        !q ||
        c.name.toLowerCase().includes(q) ||
        c.description.toLowerCase().includes(q) ||
        c.syntax.toLowerCase().includes(q);
      return inCategory && inQuery;
    });
  }, [query, category]);

  const grouped = useMemo(() => {
    const map = new Map<string, CliCommand[]>();
    filtered.forEach((c) => {
      const list = map.get(c.category) ?? [];
      list.push(c);
      map.set(c.category, list);
    });
    return map;
  }, [filtered]);

  const selectedRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    selectedRef.current?.scrollIntoView({ block: 'nearest' });
  }, [selected]);

  const handleSelect = (cmd: CliCommand) => {
    setSelected(cmd);
    setOutput(null);
    setRunning(false);
  };

  const handleRun = () => {
    setRunning(true);
    setOutput(null);
    window.setTimeout(() => {
      setRunning(false);
      setOutput(outputMode === 'json' ? selected.jsonOutput : selected.expectedOutput);
    }, 600);
  };

  return (
    <PageShell>
      <PageHeader eyebrow="Lesson 3 of 9" title={lesson.label} lead={lesson.whyText} icon={TerminalIcon} />

      <Section title="company-os CLI Terminal Simulator">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
          {/* Command list */}
          <div className="lg:col-span-4">
            <Card padding="sm" className="flex h-full flex-col gap-3">
              <h3 className="px-1 text-xs font-bold uppercase tracking-wider text-fg-subtle">
                Subcommand Reference ({CLI_COMMANDS_DATA.length})
              </h3>

              <div>
                <label htmlFor="cli-search" className="mb-1 block text-sm font-medium text-fg-muted">
                  Search
                </label>
                <div className="relative">
                  <Search
                    className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-fg-subtle"
                    aria-hidden="true"
                  />
                  <input
                    id="cli-search"
                    type="search"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="validate, prd new, graph build…"
                    className="h-11 w-full rounded-lg border border-border bg-surface pl-9 pr-3 text-sm text-fg placeholder:text-fg-subtle focus-visible:border-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus"
                  />
                </div>
              </div>

              <Tabs
                label="Filter by category"
                wrap
                active={category}
                onChange={(id) => setCategory(id as CategoryFilter)}
                tabs={[{ id: 'all', label: 'All' }, ...CATEGORIES.map((c) => ({ id: c, label: c }))]}
              />

              <div className="thin-scrollbar max-h-[520px] flex-1 overflow-y-auto pr-1">
                {filtered.length === 0 ? (
                  <EmptyState icon={Search} title="No commands match" />
                ) : (
                  Array.from(grouped.entries()).map(([cat, cmds]) => (
                    <div key={cat} className="mb-4 last:mb-0">
                      <h4 className="mb-2 px-1 font-mono text-2xs font-medium uppercase tracking-widest text-fg-subtle">
                        {cat}
                      </h4>
                      <div className="space-y-1.5">
                        {cmds.map((cmd) => {
                          const isSelected = selected.id === cmd.id;
                          return (
                            <button
                              key={cmd.id}
                              ref={isSelected ? selectedRef : undefined}
                              type="button"
                              onClick={() => handleSelect(cmd)}
                              aria-current={isSelected}
                              className={cx(
                                'w-full cursor-pointer rounded-lg border p-3 text-left text-sm transition-colors duration-150',
                                isSelected
                                  ? 'border-accent-border bg-accent-soft'
                                  : 'border-border bg-surface hover:bg-surface-sunken'
                              )}
                            >
                              <span className="block font-mono font-semibold text-fg">{cmd.name}</span>
                              <span className="mt-0.5 line-clamp-2 block text-xs text-fg-muted">
                                {cmd.description}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </Card>
          </div>

          {/* Detail */}
          <div className="lg:col-span-8 space-y-5">
            <Card className="space-y-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <Badge tone="accent">{selected.category}</Badge>
                  <p className="mt-2 break-words font-mono text-lg font-semibold text-fg">{selected.syntax}</p>
                </div>
              </div>
              <p className="measure text-sm text-fg-muted">{selected.description}</p>

              {selected.flags.length > 0 && (
                <div>
                  <p className="mb-2 text-xs font-bold uppercase tracking-wider text-fg-subtle">Command Flags</p>
                  <div className="overflow-x-auto rounded-lg border border-border">
                    <table className="w-full min-w-[560px] text-left text-sm">
                      <thead className="bg-surface-sunken text-xs uppercase tracking-wide text-fg-subtle">
                        <tr>
                          <th scope="col" className="px-3 py-2 font-medium">Flag</th>
                          <th scope="col" className="px-3 py-2 font-medium">Type</th>
                          <th scope="col" className="px-3 py-2 font-medium">Required</th>
                          <th scope="col" className="px-3 py-2 font-medium">Description</th>
                          <th scope="col" className="px-3 py-2 font-medium">Default</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {selected.flags.map((f) => (
                          <tr key={f.flag}>
                            <td className="px-3 py-2">
                              <InlineCode>{f.flag}</InlineCode>
                            </td>
                            <td className="px-3 py-2 font-mono text-xs text-fg-muted">{f.type}</td>
                            <td className="px-3 py-2">
                              {f.required ? (
                                <Badge tone="warn" icon={AlertTriangle}>Required</Badge>
                              ) : (
                                <Badge tone="neutral">Optional</Badge>
                              )}
                            </td>
                            <td className="px-3 py-2 text-fg-muted">{f.description}</td>
                            <td className="px-3 py-2 font-mono text-xs text-fg-subtle">{f.defaultVal ?? '—'}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              <div>
                <p className="mb-1.5 text-xs font-bold uppercase tracking-wider text-fg-subtle">Example</p>
                <CodeBlock code={selected.example} label={selected.name} />
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <span className="text-xs text-fg-subtle">Exit codes:</span>
                {selected.exitCodesPossible.map((code) => {
                  const info = EXIT_CODES_DATA.find((e) => e.code === code);
                  return (
                    <Badge
                      key={code}
                      tone={code === 0 ? 'success' : 'danger'}
                      icon={code === 0 ? CheckCircle2 : XCircle}
                      mono
                    >
                      {code}
                      {info ? ` · ${info.meaning}` : ''}
                    </Badge>
                  );
                })}
              </div>
            </Card>

            <Card className="space-y-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <Tabs
                  label="Output format"
                  active={outputMode}
                  onChange={(id) => {
                    setOutputMode(id);
                    setOutput(null);
                  }}
                  tabs={[
                    { id: 'text', label: 'Text output' },
                    { id: 'json', label: 'JSON output' },
                  ]}
                />
                <Button icon={Play} onClick={handleRun} loading={running}>
                  Execute
                </Button>
              </div>

              <Terminal title={`${selected.name} — output`}>
                <div aria-live="polite">
                  {running ? (
                    <p className="text-term-muted">Running…</p>
                  ) : output ? (
                    <pre className="whitespace-pre-wrap">{output}</pre>
                  ) : (
                    <p className="text-term-muted">Press Execute to run this command.</p>
                  )}
                </div>
              </Terminal>
            </Card>
          </div>
        </div>
      </Section>
    </PageShell>
  );
};
