import React, { useId, useState } from 'react';
import {
  ShieldCheck,
  ShieldAlert,
  Clock,
  Info,
  Layers,
  FileWarning,
  Send,
} from 'lucide-react';
import {
  PageShell,
  PageHeader,
  Section,
  Card,
  Callout,
  Badge,
  Button,
  Terminal,
  Prose,
  type Tone,
} from './ui';
import { STEP_TIERS } from '../data/skillsData';
import { SKILL_LAYERS } from '../data/skillsData';
import { getLesson } from '../lessons';

const TIER_TONE: Record<(typeof STEP_TIERS)[number]['tier'], Tone> = {
  mandatory: 'danger',
  default: 'warn',
  guidance: 'info',
};

const TIER_ICON: Record<(typeof STEP_TIERS)[number]['tier'], typeof ShieldAlert> = {
  mandatory: ShieldAlert,
  default: Clock,
  guidance: Info,
};

const TIER_ESCAPE_HATCH: Record<(typeof STEP_TIERS)[number]['tier'], { label: string; command: string | null }> = {
  mandatory: {
    label: 'Exception (time-boxed, requires owner signoff)',
    command: 'company-os exception request <rule> --team <t> --component <c> --expires <date>',
  },
  default: {
    label: 'Deviation (rationale, auto-reviewed)',
    command: 'company-os deviation declare <rule> --team <t>',
  },
  guidance: {
    label: 'None needed — fully voluntary',
    command: null,
  },
};

const AUTHORITY_TONE: Record<(typeof SKILL_LAYERS)[number]['authority'], Tone> = {
  canonical: 'accent',
  team: 'scope',
  personal: 'neutral',
};

export const GovernanceTiersView: React.FC = () => {
  const lesson = getLesson('governance');

  // Deviation builder state
  const devRuleId = useId();
  const devTeamId = useId();
  const devRationaleId = useId();
  const [devRule, setDevRule] = useState('company-standard://estimation/story-points');
  const [devTeam, setDevTeam] = useState('web');
  const [devRationale, setDevRationale] = useState(
    'Team forecasts with cycle time instead of points.'
  );
  const [devOutput, setDevOutput] = useState<string | null>(null);

  // Exception builder state
  const excRuleId = useId();
  const excTeamId = useId();
  const excCompId = useId();
  const excExpiresId = useId();
  const excReasonId = useId();
  const [excRule, setExcRule] = useState('platform-standard://ordering/order-confirmation-sla');
  const [excTeam, setExcTeam] = useState('web');
  const [excComp, setExcComp] = useState('legacy-pos-bridge');
  const [excExpires, setExcExpires] = useState('2026-12-31');
  const [excReason, setExcReason] = useState(
    'Legacy POS bridge cannot emit confirmations synchronously.'
  );
  const [excOutput, setExcOutput] = useState<string | null>(null);

  const handleDeclareDeviation = () => {
    setDevOutput(
      `$ company-os deviation declare ${devRule} --team ${devTeam}\ndeclared deviation in teams/${devTeam}/governance/deviations.yaml\nreview due 2027-01-22 (180 days out); rationale: "${devRationale}"\nnext: company-os governance resolve --team ${devTeam}`
    );
  };

  const handleRequestException = () => {
    if (!excExpires) {
      setExcOutput('Error: --expires DATE is required for mandatory-tier exceptions!');
      return;
    }
    setExcOutput(
      `$ company-os exception request ${excRule} --team ${excTeam} --component ${excComp} --expires ${excExpires}\nexception drafted in teams/${excTeam}/governance/exceptions.yaml (expires ${excExpires})\nnote: mandatory rules require approval by the rule owner before this is valid.`
    );
  };

  const inputCls =
    'w-full rounded-md border border-border bg-surface-sunken px-3 py-2 font-mono text-sm text-fg placeholder:text-fg-subtle';
  const labelCls = 'mb-1 block text-xs font-semibold text-fg-muted';

  return (
    <PageShell>
      <PageHeader
        eyebrow="Lesson 5 of 9"
        title={lesson.label}
        lead={lesson.whyText}
        icon={ShieldCheck}
      />

      <Section
        title="The three rule tiers"
        description="Every requirement or control in Company OS is written as one of three tiers. The tier decides what it permits, what escape hatch exists, and what happens when a rule is violated — never a colour alone."
      >
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          {STEP_TIERS.map((tier) => {
            const tone = TIER_TONE[tier.tier];
            const Icon = TIER_ICON[tier.tier];
            const hatch = TIER_ESCAPE_HATCH[tier.tier];
            return (
              <Card key={tier.tier} tone={tone} className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <Badge tone={tone} icon={Icon}>
                    {tier.label}
                  </Badge>
                </div>
                <h3 className="text-xl font-semibold capitalize text-fg">{tier.tier}</h3>

                <div className="space-y-1">
                  <p className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
                    What it permits
                  </p>
                  <p className="text-sm leading-relaxed">{tier.description}</p>
                </div>

                <div className="space-y-1 border-t border-border pt-3">
                  <p className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
                    On conflict
                  </p>
                  <p className="text-sm leading-relaxed">{tier.behavior}</p>
                </div>

                <div className="space-y-1 border-t border-border pt-3">
                  <p className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
                    Escape hatch
                  </p>
                  <p className="text-sm font-medium">{hatch.label}</p>
                  {hatch.command && (
                    <p className="truncate font-mono text-2xs text-fg-subtle">{hatch.command}</p>
                  )}
                </div>
              </Card>
            );
          })}
        </div>
      </Section>

      <Section
        title="Canonical, team and personal layers"
        description="Rules do not live in one file. Four layers can each define behavior for the same step — this is how they resolve against each other."
      >
        <div className="thin-scrollbar overflow-x-auto rounded-xl border border-border">
          <table className="w-full min-w-[640px] border-collapse text-sm">
            <thead>
              <tr className="border-b border-border bg-surface-sunken text-left">
                <th scope="col" className="px-4 py-2.5 font-semibold text-fg">
                  Layer
                </th>
                <th scope="col" className="px-4 py-2.5 font-semibold text-fg">
                  Location
                </th>
                <th scope="col" className="px-4 py-2.5 font-semibold text-fg">
                  Authority
                </th>
                <th scope="col" className="px-4 py-2.5 font-semibold text-fg">
                  Override rule
                </th>
              </tr>
            </thead>
            <tbody>
              {SKILL_LAYERS.map((layer, idx) => (
                <tr
                  key={layer.layer}
                  className={idx < SKILL_LAYERS.length - 1 ? 'border-b border-border' : ''}
                >
                  <td className="px-4 py-3 align-top font-medium text-fg">{layer.title}</td>
                  <td className="px-4 py-3 align-top font-mono text-2xs text-fg-subtle">
                    {layer.location}
                  </td>
                  <td className="px-4 py-3 align-top">
                    <Badge tone={AUTHORITY_TONE[layer.authority]}>{layer.authority}</Badge>
                  </td>
                  <td className="px-4 py-3 align-top text-fg-muted">{layer.overrideRule}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <Callout tone="info" title="Reading the precedence order" icon={Layers} className="mt-4">
          <Prose className="measure-tight text-sm">
            Company and platform layers are canonical: they win on mandatory steps no matter what a
            team or personal layer says. Team layers can only extend a canonical skill, never
            replace it. Personal layers outrank canonical default and guidance steps for the
            author's own workflow, but still lose to canonical mandatory steps.
          </Prose>
        </Callout>
      </Section>

      <Section
        title="Try it: declare a deviation or request an exception"
        description="Fill out either builder to see the exact command and file it writes. Default-tier rules take a deviation; mandatory-tier rules take an exception."
      >
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <Card className="space-y-4">
            <div className="border-b border-border pb-3">
              <Badge tone="warn" icon={Clock}>
                Default-tier rule
              </Badge>
              <h3 className="mt-2 text-lg font-semibold text-fg">Declare deviation builder</h3>
              <p className="mt-1 text-sm text-fg-muted">
                Opt out of a default rule by recording team rationale. Sets a 180-day reviewDate in
                teams/&lt;t&gt;/governance/deviations.yaml.
              </p>
            </div>

            <div className="space-y-3">
              <div>
                <label htmlFor={devRuleId} className={labelCls}>
                  Rule URI (default tier only)
                </label>
                <input
                  id={devRuleId}
                  type="text"
                  value={devRule}
                  onChange={(e) => setDevRule(e.target.value)}
                  className={inputCls}
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label htmlFor={devTeamId} className={labelCls}>
                    Team ID
                  </label>
                  <input
                    id={devTeamId}
                    type="text"
                    value={devTeam}
                    onChange={(e) => setDevTeam(e.target.value)}
                    className={inputCls}
                  />
                </div>
                <div>
                  <span className={labelCls}>Review period</span>
                  <input
                    type="text"
                    value="180 days (auto-calculated)"
                    disabled
                    aria-label="Review period, automatically calculated at 180 days"
                    className="w-full cursor-not-allowed rounded-md border border-border bg-surface-sunken px-3 py-2 font-mono text-sm text-fg-subtle opacity-70"
                  />
                </div>
              </div>

              <div>
                <label htmlFor={devRationaleId} className={labelCls}>
                  Rationale / explanation
                </label>
                <input
                  id={devRationaleId}
                  type="text"
                  value={devRationale}
                  onChange={(e) => setDevRationale(e.target.value)}
                  className={inputCls}
                />
              </div>

              <Button
                onClick={handleDeclareDeviation}
                variant="secondary"
                icon={FileWarning}
                fullWidth
              >
                Simulate company-os deviation declare
              </Button>

              {devOutput && (
                <Terminal title="deviation output">
                  <p className="whitespace-pre-wrap text-term-success" aria-live="polite">
                    {devOutput}
                  </p>
                </Terminal>
              )}
            </div>
          </Card>

          <Card className="space-y-4">
            <div className="border-b border-border pb-3">
              <Badge tone="danger" icon={ShieldAlert}>
                Mandatory-tier rule
              </Badge>
              <h3 className="mt-2 text-lg font-semibold text-fg">Request exception builder</h3>
              <p className="mt-1 text-sm text-fg-muted">
                Request temporary relief for a specific component. Requires an explicit expiry date
                and rule-owner signoff.
              </p>
            </div>

            <div className="space-y-3">
              <div>
                <label htmlFor={excRuleId} className={labelCls}>
                  Rule URI (mandatory tier)
                </label>
                <input
                  id={excRuleId}
                  type="text"
                  value={excRule}
                  onChange={(e) => setExcRule(e.target.value)}
                  className={inputCls}
                />
              </div>

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <div>
                  <label htmlFor={excTeamId} className={labelCls}>
                    Team ID
                  </label>
                  <input
                    id={excTeamId}
                    type="text"
                    value={excTeam}
                    onChange={(e) => setExcTeam(e.target.value)}
                    className={inputCls}
                  />
                </div>
                <div>
                  <label htmlFor={excCompId} className={labelCls}>
                    Component ID
                  </label>
                  <input
                    id={excCompId}
                    type="text"
                    value={excComp}
                    onChange={(e) => setExcComp(e.target.value)}
                    className={inputCls}
                  />
                </div>
                <div>
                  <label htmlFor={excExpiresId} className={labelCls}>
                    Expires date *
                  </label>
                  <input
                    id={excExpiresId}
                    type="date"
                    value={excExpires}
                    onChange={(e) => setExcExpires(e.target.value)}
                    className={inputCls}
                  />
                </div>
              </div>

              <div>
                <label htmlFor={excReasonId} className={labelCls}>
                  Technical reason
                </label>
                <input
                  id={excReasonId}
                  type="text"
                  value={excReason}
                  onChange={(e) => setExcReason(e.target.value)}
                  className={inputCls}
                />
              </div>

              <Button onClick={handleRequestException} variant="danger" icon={Send} fullWidth>
                Simulate company-os exception request
              </Button>

              {excOutput && (
                <Terminal title="exception output">
                  <p
                    className={
                      excOutput.startsWith('Error')
                        ? 'whitespace-pre-wrap text-term-danger'
                        : 'whitespace-pre-wrap text-term-success'
                    }
                    aria-live="polite"
                  >
                    {excOutput}
                  </p>
                </Terminal>
              )}
            </div>
          </Card>
        </div>
      </Section>
    </PageShell>
  );
};
