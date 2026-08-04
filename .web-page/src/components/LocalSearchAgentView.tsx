import React, { useState } from 'react';
import {
  Search,
  Bot,
  Layers,
  FileCode,
  ShieldAlert,
  Plug,
  MessageSquareCode,
  GitBranch,
  XCircle,
  Sparkles,
  Terminal as TerminalIcon,
  Loader2,
  CheckCircle2,
} from 'lucide-react';
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
  Tabs,
  EmptyState,
  cx,
} from './ui';
import {
  SKILL_LAYERS,
  STEP_TIERS,
  REFERENCE_SKILLS,
  AUTHORING_RULES,
  MOCK_SKILLS_CLI_TEXT,
  MOCK_SKILLS_JSON_TEXT,
  SAMPLE_PROMPTS,
  PROMPTING_RULES,
} from '../data/skillsData';
import { MCP_SKILL_STEPS, MCP_ASSISTED_PROMPT, MCP_BOUNDARY } from '../data/federationData';

const PRECEDENCE_STEPS = [
  {
    tone: 'danger' as const,
    label: '1. Top authority',
    name: 'canonical-mandatory',
    detail: 'Company & platform security/compliance gates. Unbreakable.',
  },
  {
    tone: 'scope' as const,
    label: '2. Second level',
    name: 'personal scratchpad',
    detail: 'Personal workflow rules in scratchpad/personal-rules/.',
  },
  {
    tone: 'warn' as const,
    label: '3. Third level',
    name: 'canonical-default',
    detail: 'Standard process unless a formal deviation is logged.',
  },
  {
    tone: 'neutral' as const,
    label: '4. Fourth level',
    name: 'canonical-guidance',
    detail: 'Advisory best practices. Untracked and optional.',
  },
];

const layerTone = (authority: 'canonical' | 'team' | 'personal') =>
  authority === 'canonical' ? 'accent' : authority === 'team' ? 'scope' : 'neutral';

const tierTone = (tier: 'mandatory' | 'default' | 'guidance') =>
  tier === 'mandatory' ? 'danger' : tier === 'default' ? 'accent' : 'neutral';

interface MockDoc {
  id: string;
  score: number;
  path: string;
  snippet: string;
  keywords: string[];
}

const MOCK_DOCS: MockDoc[] = [
  {
    id: 'doc-prd',
    score: 0.94,
    path: 'platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md',
    snippet:
      '...Allow web customers to choose 15-minute same-day pickup slots on current day...',
    keywords: ['same-day', 'same day', 'pickup', 'slots', 'prd', 'ordering'],
  },
  {
    id: 'doc-brief',
    score: 0.81,
    path: 'teams/web/product/discovery/2026-same-day-pickup-slots/brief.md',
    snippet: '...Problem signal: Customers abandon carts when pickup is limited to next-day...',
    keywords: ['cart', 'abandon', 'pickup', 'brief', 'discovery', 'next-day'],
  },
];

export const LocalSearchAgentView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'skills' | 'search'>('skills');
  const [selectedReferenceSkill, setSelectedReferenceSkill] = useState<string>(REFERENCE_SKILLS[1].id);
  const [cliFormat, setCliFormat] = useState<'text' | 'json'>('text');

  const [searchQuery, setSearchQuery] = useState('same-day pickup');
  const [searchScope, setSearchScope] = useState('moonbeam-os');
  const [isSearching, setIsSearching] = useState(false);
  const [results, setResults] = useState<MockDoc[]>(MOCK_DOCS);
  const [hasSearched, setHasSearched] = useState(true);

  const activeRefSkill = REFERENCE_SKILLS.find((s) => s.id === selectedReferenceSkill) || REFERENCE_SKILLS[1];

  const runSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSearching(true);
    const q = searchQuery.trim().toLowerCase();
    window.setTimeout(() => {
      const matched = q
        ? MOCK_DOCS.filter((doc) => doc.keywords.some((k) => q.includes(k) || k.includes(q)))
        : [];
      setResults(matched);
      setHasSearched(true);
      setIsSearching(false);
    }, 500);
  };

  return (
    <PageShell>
      <PageHeader
        eyebrow="Lesson 7 of 9"
        title="Agent Skills & Workspace Search Engine"
        lead="Skills teach AI agents and humans how to author artifacts. Local Search teaches them how to find what already exists — offline, instantly."
        icon={Bot}
      />

      <Tabs
        label="Agent skills and local search"
        tabs={[
          { id: 'skills', label: 'Agent Skills (4 Layers)', icon: Bot },
          { id: 'search', label: 'Local Search Engine', icon: Search },
        ]}
        active={activeTab}
        onChange={(id) => setActiveTab(id as typeof activeTab)}
      />

      {activeTab === 'skills' ? (
        <>
          <Section
            title="Authoring side vs. read side"
            description="Two independent systems that cooperate: skills teach agents how to write compliant artifacts; Local Search teaches them how to find existing ones."
          >
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Card className="space-y-2.5">
                <div className="flex items-center gap-2 text-accent-text">
                  <Bot className="h-5 w-5" aria-hidden="true" />
                  <h3 className="text-sm font-bold text-fg">Company OS Agent Skills (Authoring Side)</h3>
                </div>
                <p className="text-xs leading-relaxed text-fg-muted">
                  Versioned Markdown files (<InlineCode>*.SKILL.md</InlineCode>) that teach agents and humans{' '}
                  <strong>how to author</strong> compliant discovery briefs, PRDs, and reality docs. Enforced during
                  validation.
                </p>
                <InlineCode>company-os skills list</InlineCode>
              </Card>

              <Card className="space-y-2.5">
                <div className="flex items-center gap-2 text-scope-text">
                  <Search className="h-5 w-5" aria-hidden="true" />
                  <h3 className="text-sm font-bold text-fg">Local Search Claude Skill (Read Side)</h3>
                </div>
                <p className="text-xs leading-relaxed text-fg-muted">
                  Installed via <InlineCode>local-search install-skill</InlineCode> into{' '}
                  <InlineCode>~/.claude/skills/local-search</InlineCode>. Teaches agents <strong>how to find</strong>{' '}
                  existing workspace docs offline via SQLite FTS5.
                </p>
                <InlineCode>local-search init --json</InlineCode>
              </Card>
            </div>
          </Section>

          <Section
            title="4-layer skills precedence hierarchy"
            description="When an AI agent operates on a workspace, rules and skills are resolved across 4 layers. Canonical mandatory steps always outrank personal preferences."
          >
            <div className="space-y-4">
              <CodeBlock
                label="teams/<t>/team.yaml"
                copyable={false}
                code={[
                  '# Configured precedence contract in team.yaml',
                  'precedence: canonical-mandatory > personal > canonical-default > canonical-guidance',
                  'onConflict: prefer-canonical-and-inform-user',
                ].join('\n')}
              />

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
                {PRECEDENCE_STEPS.map((step) => (
                  <Card key={step.name} tone={step.tone} padding="sm" className="space-y-1">
                    <span className="block font-mono text-2xs font-bold uppercase tracking-wider">{step.label}</span>
                    <span className="block text-sm font-bold text-fg">{step.name}</span>
                    <p className="text-xs leading-tight">{step.detail}</p>
                  </Card>
                ))}
              </div>

              <div className="flex flex-wrap gap-2">
                {STEP_TIERS.map((t) => (
                  <Badge key={t.tier} tone={tierTone(t.tier)}>
                    {t.label} — {t.description}
                  </Badge>
                ))}
              </div>
            </div>
          </Section>

          <Section title="4 skill storage layers">
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {SKILL_LAYERS.map((layer) => (
                <Card key={layer.layer} className="space-y-2">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-bold text-fg">{layer.title}</span>
                    <Badge tone={layerTone(layer.authority)}>{layer.authority}</Badge>
                  </div>
                  <InlineCode className="block w-fit max-w-full truncate">{layer.location}</InlineCode>
                  <p className="text-xs leading-relaxed text-fg-muted">{layer.description}</p>
                  <p className="border-t border-border pt-1 text-2xs italic text-fg-subtle">{layer.overrideRule}</p>
                </Card>
              ))}
            </div>
          </Section>

          <Section
            title="Reference agent skills explorer"
            description="Pre-packaged skills shipped with Company OS to guide agents through lifecycle milestones."
            actions={<Badge>{REFERENCE_SKILLS.length} reference skills</Badge>}
          >
            <div className="space-y-4">
              <div className="flex flex-wrap gap-2" role="group" aria-label="Select a reference skill">
                {REFERENCE_SKILLS.map((skill) => {
                  const on = skill.id === selectedReferenceSkill;
                  return (
                    <button
                      key={skill.id}
                      type="button"
                      aria-pressed={on}
                      onClick={() => setSelectedReferenceSkill(skill.id)}
                      className={cx(
                        'flex h-11 cursor-pointer items-center gap-1.5 rounded-lg px-3 text-xs font-semibold transition-colors duration-150',
                        on ? 'bg-accent text-accent-fg' : 'bg-surface-sunken text-fg-muted hover:text-fg'
                      )}
                    >
                      <FileCode className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
                      <span>{skill.title}</span>
                    </button>
                  );
                })}
              </div>

              <Card tone="neutral" className="space-y-3">
                <div className="flex flex-col gap-2 border-b border-border pb-2 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <span className="block font-mono text-2xs font-bold uppercase text-accent-text">
                      {activeRefSkill.id}
                    </span>
                    <h4 className="text-sm font-bold text-fg">{activeRefSkill.title}</h4>
                  </div>
                  <InlineCode className="self-start text-2xs sm:self-center">{activeRefSkill.location}</InlineCode>
                </div>

                <p className="text-xs leading-relaxed text-fg-muted">{activeRefSkill.summary}</p>

                <div className="space-y-2">
                  <span className="text-xs font-bold uppercase tracking-wider text-fg">Step execution guidance</span>
                  {activeRefSkill.steps.map((st) => (
                    <div
                      key={st.number}
                      className="flex items-start gap-2 rounded-xl border border-border bg-surface p-2.5 text-xs"
                    >
                      <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-surface-sunken font-mono text-2xs font-bold text-fg">
                        {st.number}
                      </span>
                      <Badge tone={tierTone(st.tier)} className="shrink-0">
                        {st.tier}
                      </Badge>
                      <span className="font-mono leading-relaxed text-fg-muted">{st.text}</span>
                    </div>
                  ))}
                </div>
              </Card>
            </div>
          </Section>

          <Section title="Authoring rules & gate [7/N] constraints">
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              {AUTHORING_RULES.map((r, idx) => (
                <Card key={idx} padding="sm" className="space-y-1">
                  <div className="flex items-center gap-2 text-xs font-bold text-fg">
                    <ShieldAlert className="h-3.5 w-3.5 shrink-0 text-accent-text" aria-hidden="true" />
                    <span>{r.rule}</span>
                  </div>
                  <p className="pl-5 text-xs leading-relaxed text-fg-muted">{r.detail}</p>
                </Card>
              ))}
            </div>
          </Section>

          <Section
            title="External agent tooling: the GitHub MCP"
            description="How MCP servers relate to skills — and why the skill layer is the only door they come through."
            actions={
              <Badge tone="danger" icon={Plug}>
                Ships 0 MCP servers
              </Badge>
            }
          >
            <div className="space-y-4">
              <Callout tone="info" icon={Plug}>
                <p>
                  <strong>Company OS ships no MCP server and no MCP client.</strong> There is no{' '}
                  <InlineCode>.mcp.json</InlineCode> anywhere in the repo, and no{' '}
                  <InlineCode>company-os</InlineCode> command talks to the GitHub API —{' '}
                  <InlineCode>workspace sync</InlineCode> is eight plain git invocations with no{' '}
                  <InlineCode>push</InlineCode>, <InlineCode>commit</InlineCode>, or HTTP call anywhere.
                </p>
                <p className="mt-2">
                  So the GitHub MCP is <strong>your agent&apos;s</strong> tool, configured in your agent; Company OS
                  neither knows nor cares that it exists. It becomes part of a governed workflow only when a{' '}
                  <strong>skill</strong> tells the agent when to reach for it. That inherits the constraint from{' '}
                  <em>No Execution Authority</em> above: no skill — and therefore no MCP call it suggests — can grant
                  permission to bypass a mandatory gate, edit <InlineCode>generated/</InlineCode>, or modify a{' '}
                  <InlineCode>0444</InlineCode> synced slice.
                </p>
              </Callout>

              <div className="space-y-2">
                <div className="flex items-center gap-2 text-xs font-bold text-fg">
                  <GitBranch className="h-4 w-4 text-accent-text" aria-hidden="true" />
                  <span>
                    Applied to <InlineCode>skill://governance/syncing-knowledge</InlineCode>
                  </span>
                </div>
                <p className="text-xs text-fg-muted">
                  The skill&apos;s three mandatory steps are all CLI. MCP helps <em>around</em> them, never inside
                  them:
                </p>

                <div className="thin-scrollbar overflow-x-auto rounded-xl border border-border">
                  <table className="w-full min-w-[640px] text-xs">
                    <thead className="sticky top-0 bg-surface-sunken">
                      <tr className="border-b border-border">
                        <th className="w-[26%] px-3 py-2 text-left font-bold text-fg">Moment</th>
                        <th className="w-[37%] px-3 py-2 text-left font-bold text-success-text">GitHub MCP (read)</th>
                        <th className="w-[37%] px-3 py-2 text-left font-bold text-accent-text">
                          company-os CLI (write)
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {MCP_SKILL_STEPS.map((step) => (
                        <tr key={step.id} className="border-b border-border last:border-0 even:bg-surface-sunken/50">
                          <td className="px-3 py-2 align-top font-semibold leading-relaxed text-fg">{step.moment}</td>
                          <td
                            className={cx(
                              'px-3 py-2 align-top leading-relaxed',
                              step.mcp === '—' ? 'text-center text-fg-subtle' : 'text-fg-muted'
                            )}
                          >
                            {step.mcp}
                          </td>
                          <td
                            className={cx(
                              'px-3 py-2 align-top leading-relaxed',
                              step.cli === '—' ? 'text-center text-fg-subtle' : 'text-fg-muted'
                            )}
                          >
                            {step.cli}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2 text-xs font-bold text-danger-text">
                  <XCircle className="h-4 w-4" aria-hidden="true" />
                  <span>{MCP_BOUNDARY[2].title}</span>
                </div>
                <div className="grid grid-cols-1 gap-2.5 md:grid-cols-3">
                  {MCP_BOUNDARY[2].items.map((item, i) => (
                    <Card key={i} tone="danger" padding="sm" className="text-2xs leading-relaxed">
                      {item}
                    </Card>
                  ))}
                </div>
              </div>

              <div className="space-y-2 pt-1">
                <div className="flex items-center gap-2 text-xs font-bold text-fg">
                  <MessageSquareCode className="h-4 w-4 text-accent-text" aria-hidden="true" />
                  <span>MCP-assisted variant of the Sync Knowledge Catalog prompt</span>
                </div>
                <p className="text-xs text-fg-muted">
                  Same skill, same gates — the prompt just uses MCP for the read half and names the boundary out loud
                  so the agent cannot drift across it:
                </p>
                <CodeBlock label="prompt" code={MCP_ASSISTED_PROMPT} />
              </div>

              <Callout tone="warn" icon={ShieldAlert}>
                <strong className="text-fg">
                  An agent may read anything, but the CLI writes the workspace and the lock is the oracle.
                </strong>{' '}
                An MCP server that respects that boundary is a convenience. One that crosses it produces a workspace
                whose validation results are no longer evidence of anything.
              </Callout>
            </div>
          </Section>

          <Section
            title="Sample prompts for AI agents"
            description="Ready-to-use structured prompts for instructing AI agents to follow Company OS skills with strict validation boundaries."
            actions={<Badge>{SAMPLE_PROMPTS.length} agent templates</Badge>}
          >
            <div className="space-y-4">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                {SAMPLE_PROMPTS.map((prompt) => (
                  <Card key={prompt.id} className="flex flex-col justify-between space-y-2.5">
                    <div className="space-y-1.5">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <Badge tone="accent">{prompt.category}</Badge>
                        {prompt.skillName && (
                          <Badge tone="accent" icon={FileCode} mono>
                            {prompt.skillName}
                          </Badge>
                        )}
                      </div>
                      <div>
                        <h4 className="text-sm font-bold text-fg">{prompt.title}</h4>
                        {prompt.targetSkillId && (
                          <p className="mt-1 font-mono text-2xs text-fg-subtle">
                            ID: <span className="font-semibold text-fg-muted">{prompt.targetSkillId}</span>
                          </p>
                        )}
                      </div>
                      <p className="text-xs text-fg-muted">{prompt.description}</p>
                      <CodeBlock code={prompt.promptText} label="prompt" className="max-h-36" />
                    </div>
                  </Card>
                ))}
              </div>

              <div className="space-y-3 border-t border-border pt-3">
                <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-fg">
                  <Sparkles className="h-4 w-4 text-accent-text" aria-hidden="true" />
                  <span>Prompting best practices (what makes these work)</span>
                </div>
                <div className="grid grid-cols-1 gap-3 text-xs sm:grid-cols-2 lg:grid-cols-3">
                  {PROMPTING_RULES.map((pr, idx) => (
                    <Card key={idx} tone="accent" padding="sm" className="space-y-1">
                      <div className="text-xs font-bold">{pr.rule}</div>
                      <p className="leading-relaxed">{pr.detail}</p>
                    </Card>
                  ))}
                </div>
              </div>
            </div>
          </Section>

          <Section title="company-os skills list CLI simulator">
            <div className="space-y-3">
              <Tabs
                label="CLI output format"
                tabs={[
                  { id: 'text', label: 'Text format' },
                  { id: 'json', label: '--json envelope' },
                ]}
                active={cliFormat}
                onChange={(id) => setCliFormat(id as typeof cliFormat)}
              />
              <Terminal title={`company-os ${cliFormat === 'json' ? '--json ' : ''}skills list`}>
                <pre className="whitespace-pre-wrap">{cliFormat === 'text' ? MOCK_SKILLS_CLI_TEXT : MOCK_SKILLS_JSON_TEXT}</pre>
              </Terminal>
            </div>
          </Section>
        </>
      ) : (
        <>
          <Section
            title="Two engines, two jobs"
            description="company-os writes and validates the workspace. Local Search only reads it — a single Go binary backed by SQLite FTS5."
          >
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Card className="space-y-3">
                <Badge tone="accent">Authoring & validation side</Badge>
                <h3 className="text-sm font-bold text-fg">company-os CLI</h3>
                <p className="text-xs leading-relaxed text-fg-muted">
                  Scaffolds discovery briefs, PRDs, reality docs, resolves governance, builds graph tags, and runs
                  validation gates. Never queries external servers or databases.
                </p>
                <InlineCode>company-os validate</InlineCode>
              </Card>

              <Card className="space-y-3">
                <Badge tone="scope">Read & query side</Badge>
                <h3 className="text-sm font-bold text-fg">Local Search (SQLite FTS5)</h3>
                <p className="text-xs leading-relaxed text-fg-muted">
                  Single Go binary that registers workspaces, scans Markdown files, and provides BM25 hybrid search
                  for terminal users and AI agents.
                </p>
                <InlineCode>local-search search &quot;same-day pickup&quot;</InlineCode>
              </Card>
            </div>
          </Section>

          <Section
            title="Local search query simulator"
            description="Run a query against the offline BM25 index and see ranked results, exactly as an agent would receive them."
          >
            <Card className="space-y-4">
              <form onSubmit={runSearch} className="grid grid-cols-1 gap-3 sm:grid-cols-[2fr_1fr_auto] sm:items-end">
                <div>
                  <label htmlFor="search-query" className="mb-1 block text-xs font-semibold text-fg">
                    Search query
                  </label>
                  <input
                    id="search-query"
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="h-11 w-full rounded-lg border border-border bg-surface-sunken px-3 font-mono text-sm text-fg focus-visible:outline-none"
                  />
                </div>

                <div>
                  <label htmlFor="search-scope" className="mb-1 block text-xs font-semibold text-fg">
                    Registered scope
                  </label>
                  <input
                    id="search-scope"
                    type="text"
                    value={searchScope}
                    onChange={(e) => setSearchScope(e.target.value)}
                    className="h-11 w-full rounded-lg border border-border bg-surface-sunken px-3 font-mono text-sm text-fg focus-visible:outline-none"
                  />
                </div>

                <Button type="submit" icon={Search} loading={isSearching}>
                  Run search
                </Button>
              </form>

              <div aria-live="polite">
                {isSearching ? (
                  <div className="flex items-center gap-2 rounded-xl border border-border bg-surface-sunken p-6 text-sm text-fg-muted">
                    <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
                    <span>Querying local BM25 index…</span>
                  </div>
                ) : hasSearched && results.length === 0 ? (
                  <EmptyState
                    icon={Search}
                    title="No matches in the local index"
                    description={`No documents in "${searchScope}" match "${searchQuery}". Try broader keywords.`}
                  />
                ) : (
                  <Terminal title={`Local Search BM25 ranked results (${results.length} matches)`}>
                    <div className="space-y-3">
                      {results.map((doc, i) => (
                        <div key={doc.id} className={i > 0 ? 'border-t border-term-bg-soft pt-2' : undefined}>
                          <div className="flex items-center gap-2 text-term-muted">
                            <span className="tabular">
                              [{i + 1}] score: {doc.score.toFixed(2)} | repo: {searchScope}
                            </span>
                          </div>
                          <div className="font-bold text-term-fg">{doc.path}</div>
                          <div className="text-term-muted">{doc.snippet}</div>
                        </div>
                      ))}
                    </div>
                  </Terminal>
                )}
              </div>
            </Card>
          </Section>
        </>
      )}
    </PageShell>
  );
};
