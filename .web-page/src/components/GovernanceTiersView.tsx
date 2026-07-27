import React, { useState } from 'react';
import { ShieldCheck, AlertCircle, FileSpreadsheet, Check, Sparkles, Clock, AlertTriangle, ArrowRight } from 'lucide-react';

export const GovernanceTiersView: React.FC = () => {
  // Interactive Deviation State
  const [devRule, setDevRule] = useState('company-standard://estimation/story-points');
  const [devTeam, setDevTeam] = useState('web');
  const [devRationale, setDevRationale] = useState('Team forecasts with cycle time instead of points.');
  const [devOutput, setDevOutput] = useState<string | null>(null);

  // Interactive Exception State
  const [excRule, setExcRule] = useState('platform-standard://ordering/order-confirmation-sla');
  const [excTeam, setExcTeam] = useState('web');
  const [excComp, setExcComp] = useState('legacy-pos-bridge');
  const [excExpires, setExcExpires] = useState('2026-12-31');
  const [excReason, setExcReason] = useState('Legacy POS bridge cannot emit confirmations synchronously.');
  const [excOutput, setExcOutput] = useState<string | null>(null);

  const handleDeclareDeviation = () => {
    setDevOutput(
      `declared deviation from ${devRule} in teams/${devTeam}/governance/deviations.yaml\nreview due 2027-01-22 (180 days out); re-run: company-os governance resolve --team ${devTeam}`
    );
  };

  const handleRequestException = () => {
    if (!excExpires) {
      setExcOutput(`Error: --expires DATE is required for mandatory exceptions!`);
      return;
    }
    setExcOutput(
      `exception drafted in teams/${excTeam}/governance/exceptions.yaml (expires ${excExpires})\nnote: mandatory rules require approval by the rule owner before this is valid.`
    );
  };

  return (
    <div className="space-y-6">
      
      {/* Intro Header */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                04 GOVERNANCE TIERS
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                FLEXIBILITY & STANDARDS
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">Governance Rule Tiers, Deviations & Exceptions</h2>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              So teams can follow company standards while having a safe, official way to request exceptions when needed.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              The 3 rule tiers (Mandatory, Default, Guidance) and forms to declare a rule deviation or exception.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Read the 3 rule cards below, or fill out a quick form to test declaring a team rule exception!
            </p>
          </div>
        </div>
      </div>

      {/* 3 Tiers Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        
        {/* Tier 1: Mandatory */}
        <div className="bg-white border border-rose-200 shadow-sm rounded-2xl p-5 space-y-3 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-rose-700 uppercase tracking-wider font-mono bg-rose-50 px-2 py-0.5 rounded border border-rose-200">
              Mandatory Tier
            </span>
            <AlertCircle className="w-5 h-5 text-rose-600" />
          </div>
          <h3 className="text-base font-bold text-slate-900">Guaranteed Outcome</h3>
          <p className="text-xs text-slate-600 leading-relaxed">
            Non-negotiable outcomes (e.g. security SLAs, compliance checks). Cannot be opted out of via deviations. Escapable ONLY via an approved, time-boxed <strong className="text-rose-700">Exception</strong>.
          </p>
          <div className="pt-2 text-[11px] text-slate-500 border-t border-slate-100">
            <strong>Escape Hatch:</strong> <code className="text-rose-700 font-mono">company-os exception request</code> (Requires expiry date & owner signoff)
          </div>
        </div>

        {/* Tier 2: Default */}
        <div className="bg-white border border-amber-200 shadow-sm rounded-2xl p-5 space-y-3 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-amber-800 uppercase tracking-wider font-mono bg-amber-50 px-2 py-0.5 rounded border border-amber-200">
              Default Tier
            </span>
            <Clock className="w-5 h-5 text-amber-600" />
          </div>
          <h3 className="text-base font-bold text-slate-900">Comply or Explain</h3>
          <p className="text-xs text-slate-600 leading-relaxed">
            Standard guidelines (e.g. story points estimation). Teams can opt out by declaring a <strong className="text-amber-800">Deviation</strong> with rationale. Auto-sets a 180-day review date.
          </p>
          <div className="pt-2 text-[11px] text-slate-500 border-t border-slate-100">
            <strong>Escape Hatch:</strong> <code className="text-amber-800 font-mono">company-os deviation declare</code> (Review date 180 days out)
          </div>
        </div>

        {/* Tier 3: Guidance */}
        <div className="bg-white border border-emerald-200 shadow-sm rounded-2xl p-5 space-y-3 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-emerald-700 uppercase tracking-wider font-mono bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
              Guidance Tier
            </span>
            <FileSpreadsheet className="w-5 h-5 text-emerald-600" />
          </div>
          <h3 className="text-base font-bold text-slate-900">Advisory Recommendation</h3>
          <p className="text-xs text-slate-600 leading-relaxed">
            Best practices and suggested templates. Fully advisory, untracked by validators, and requiring no deviations or approval.
          </p>
          <div className="pt-2 text-[11px] text-slate-500 border-t border-slate-100">
            <strong>Escape Hatch:</strong> None needed — completely voluntary
          </div>
        </div>

      </div>

      {/* Interactive Builders: Deviation vs Exception */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

        {/* Deviation Builder */}
        <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="border-b border-slate-100 pb-3">
            <span className="text-xs font-bold text-amber-800 font-mono uppercase tracking-wider">Default-Tier Rule</span>
            <h3 className="text-base font-bold text-slate-900 mt-0.5">Declare Deviation Builder</h3>
            <p className="text-xs text-slate-500 mt-1">
              Opt out of a default rule by recording team rationale. Automatically sets a 180-day reviewDate in teams/&lt;t&gt;/governance/deviations.yaml.
            </p>
          </div>

          <div className="space-y-3 text-xs">
            <div>
              <label className="text-slate-700 block mb-1 font-semibold">Rule URI (Default Tier Only)</label>
              <input
                type="text"
                value={devRule}
                onChange={(e) => setDevRule(e.target.value)}
                className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Team ID</label>
                <input
                  type="text"
                  value={devTeam}
                  onChange={(e) => setDevTeam(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
                />
              </div>

              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Review Period</label>
                <input
                  type="text"
                  value="180 Days (Auto-Calculated)"
                  disabled
                  className="w-full bg-slate-100 border border-slate-200 rounded-lg px-3 py-2 text-slate-400 font-mono cursor-not-allowed"
                />
              </div>
            </div>

            <div>
              <label className="text-slate-700 block mb-1 font-semibold">Rationale / Explanation</label>
              <input
                type="text"
                value={devRationale}
                onChange={(e) => setDevRationale(e.target.value)}
                className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
              />
            </div>

            <button
              onClick={handleDeclareDeviation}
              className="w-full py-2.5 rounded-xl bg-amber-600 hover:bg-amber-500 text-white font-bold transition-all shadow-md shadow-amber-600/20"
            >
              Simulate company-os deviation declare
            </button>

            {devOutput && (
              <div className="p-3 bg-slate-900 border border-slate-800 rounded-xl font-mono text-emerald-300 text-xs whitespace-pre-wrap shadow-inner">
                {devOutput}
              </div>
            )}
          </div>
        </div>

        {/* Exception Builder */}
        <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="border-b border-slate-100 pb-3">
            <span className="text-xs font-bold text-rose-700 font-mono uppercase tracking-wider">Mandatory-Tier Rule</span>
            <h3 className="text-base font-bold text-slate-900 mt-0.5">Request Exception Builder</h3>
            <p className="text-xs text-slate-500 mt-1">
              Request temporary relief for a specific component. Requires an explicit expiry date and rule owner signoff.
            </p>
          </div>

          <div className="space-y-3 text-xs">
            <div>
              <label className="text-slate-700 block mb-1 font-semibold">Rule URI (Mandatory Tier)</label>
              <input
                type="text"
                value={excRule}
                onChange={(e) => setExcRule(e.target.value)}
                className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
              />
            </div>

            <div className="grid grid-cols-3 gap-2">
              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Team ID</label>
                <input
                  type="text"
                  value={excTeam}
                  onChange={(e) => setExcTeam(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
                />
              </div>

              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Component ID</label>
                <input
                  type="text"
                  value={excComp}
                  onChange={(e) => setExcComp(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
                />
              </div>

              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Expires Date *</label>
                <input
                  type="date"
                  value={excExpires}
                  onChange={(e) => setExcExpires(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-lg px-2 py-2 text-slate-800 font-mono"
                />
              </div>
            </div>

            <div>
              <label className="text-slate-700 block mb-1 font-semibold">Technical Reason</label>
              <input
                type="text"
                value={excReason}
                onChange={(e) => setExcReason(e.target.value)}
                className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
              />
            </div>

            <button
              onClick={handleRequestException}
              className="w-full py-2.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white font-bold transition-all shadow-md shadow-rose-600/20"
            >
              Simulate company-os exception request
            </button>

            {excOutput && (
              <div className="p-3 bg-slate-900 border border-slate-800 rounded-xl font-mono text-xs whitespace-pre-wrap text-emerald-300 shadow-inner">
                {excOutput}
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
};
