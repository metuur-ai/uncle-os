import React, { useState } from 'react';
import { VALIDATION_GATES_DATA } from '../data/validationGatesData';
import { ValidationGate } from '../types';
import { CheckSquare, AlertTriangle, ShieldCheck, CheckCircle2, Bug, ArrowRight, Sparkles, RefreshCw } from 'lucide-react';

export const ValidationGatesView: React.FC = () => {
  const [selectedGate, setSelectedGate] = useState<ValidationGate>(VALIDATION_GATES_DATA[0]);
  const [simulatedFailure, setSimulatedFailure] = useState<ValidationGate | null>(null);
  const [isFixed, setIsFixed] = useState<boolean>(false);

  const handleSimulateFailure = (gate: ValidationGate) => {
    setSelectedGate(gate);
    setSimulatedFailure(gate);
    setIsFixed(false);
  };

  const handleApplyFix = () => {
    setIsFixed(true);
  };

  return (
    <div className="space-y-6">
      
      {/* Intro Header */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                05 VALIDATION GATES
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                QUALITY CONTRACT
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">Validation Safety Gates (1 to 8)</h2>
          </div>

          <div className="flex items-center gap-2 text-xs shrink-0">
            <span className="px-3 py-1.5 rounded-xl bg-white text-slate-700 border border-slate-200 font-semibold shadow-sm">
              Monorepo Mode: 7 Gates
            </span>
            <span className="px-3 py-1.5 rounded-xl bg-indigo-600 text-white font-semibold shadow-md shadow-indigo-600/20 border border-indigo-500/30">
              Federated Mode: 8 Gates
            </span>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              To automatically verify that every document is properly formatted, security rules are respected, and dates match.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              8 numbered gate buttons across the top. Selecting a gate displays its exact rules and interactive failure simulation.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Click any gate button (1-8) to inspect what it tests, or click <strong>Simulate Gate Failure</strong> to test fixing errors!
            </p>
          </div>
        </div>
      </div>

      {/* Gate Selection Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-2">
        {VALIDATION_GATES_DATA.map((gate) => {
          const isSelected = selectedGate.id === gate.id;
          return (
            <button
              key={gate.id}
              onClick={() => setSelectedGate(gate)}
              className={`p-2.5 rounded-xl text-left border transition-all flex flex-col justify-between h-24 ${
                isSelected
                  ? 'bg-indigo-50 border-indigo-300 text-slate-900 shadow-sm font-semibold'
                  : 'bg-white border-slate-200 hover:bg-slate-100/70 text-slate-700'
              }`}
            >
              <div className="flex items-center justify-between">
                <span className="font-mono font-bold text-xs text-indigo-700">Gate {gate.id}</span>
                {gate.federatedOnly && (
                  <span className="text-[9px] px-1 py-0.2 rounded bg-cyan-50 text-cyan-700 border border-cyan-200 font-mono">
                    FED
                  </span>
                )}
              </div>
              <h4 className="text-[11px] font-bold line-clamp-2 leading-tight text-slate-800 mt-1">
                {gate.shortName}
              </h4>
              <span className="text-[9px] text-slate-500">
                {gate.absenceTolerant ? 'Absence Tolerant' : 'Strict Contract'}
              </span>
            </button>
          );
        })}
      </div>

      {/* Gate Detail & Interactive Failure Debugger */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">

        {/* Left: Selected Gate Specs */}
        <div className="lg:col-span-5 bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="border-b border-slate-100 pb-3 flex items-center justify-between">
            <div>
              <span className="text-xs font-mono font-bold text-indigo-700 uppercase tracking-wider">
                Gate [{selectedGate.id}/N] Specification
              </span>
              <h3 className="text-base font-bold text-slate-900 mt-0.5">{selectedGate.name}</h3>
            </div>

            <button
              onClick={() => handleSimulateFailure(selectedGate)}
              className="px-3 py-1.5 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 font-semibold text-xs flex items-center gap-1.5 transition-colors"
            >
              <Bug className="w-3.5 h-3.5" />
              <span>Simulate Failure</span>
            </button>
          </div>

          <p className="text-xs text-slate-600 leading-relaxed">{selectedGate.description}</p>

          <div className="space-y-2">
            <span className="text-xs font-bold text-slate-800 block">Validator Rule Checks:</span>
            <ul className="space-y-1 text-xs text-slate-600">
              {selectedGate.checks.map((chk, idx) => (
                <li key={idx} className="flex items-start gap-2">
                  <CheckCircle2 className="w-3.5 h-3.5 text-indigo-600 shrink-0 mt-0.5" />
                  <span>{chk}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="p-3 bg-slate-900 border border-slate-800 rounded-xl space-y-2 text-xs shadow-inner">
            <span className="font-bold text-emerald-400 block font-mono text-[11px]">Example PASS Output:</span>
            <pre className="font-mono text-[11px] text-emerald-300 whitespace-pre-wrap">{selectedGate.examplePass}</pre>
          </div>
        </div>

        {/* Right: Live Interactive Debugger & Fixer */}
        <div className="lg:col-span-7 bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="border-b border-slate-100 pb-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Bug className="w-5 h-5 text-rose-600" />
              <h3 className="text-base font-bold text-slate-900">Interactive Gate Failure Debugger</h3>
            </div>

            <span className={`text-xs px-2.5 py-1 rounded-full font-bold uppercase tracking-wider ${
              isFixed ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-rose-50 text-rose-700 border border-rose-200'
            }`}>
              {isFixed ? 'STATUS: RESOLVED [PASS]' : 'STATUS: FAILING [FAIL]'}
            </span>
          </div>

          {/* Failure Output Console */}
          <div className="bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs space-y-2 shadow-md">
            <div className="text-slate-500 text-[10px] uppercase font-bold border-b border-slate-800 pb-1">
              company-os validate console output
            </div>

            {isFixed ? (
              <pre className="text-emerald-300 whitespace-pre-wrap leading-relaxed">
{`$ company-os validate
Gate [${selectedGate.id}/N] ${selectedGate.shortName}...
  [ok] All checks passed successfully!
PASS`}
              </pre>
            ) : (
              <pre className="text-rose-300 whitespace-pre-wrap leading-relaxed">
{`$ company-os validate
Gate [${selectedGate.id}/N] ${selectedGate.shortName}...
${selectedGate.exampleFail}
FAIL — 1 problem(s) found (Exit Code 1)`}
              </pre>
            )}
          </div>

          {/* Root Cause & One-Click Fix */}
          <div className="bg-slate-50 p-4 rounded-xl border border-slate-200 space-y-3 text-xs">
            <div>
              <span className="font-bold text-amber-800 block text-[11px] uppercase tracking-wider">Root Cause Analysis</span>
              <p className="text-slate-700 mt-1">{selectedGate.description}</p>
            </div>

            <div className="pt-2 border-t border-slate-200">
              <span className="font-bold text-indigo-700 block text-[11px] uppercase tracking-wider">Recommended Fix Action</span>
              <p className="text-slate-800 mt-1 font-mono text-[11px]">{selectedGate.fixAction}</p>
            </div>

            {!isFixed && (
              <button
                onClick={handleApplyFix}
                className="w-full py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold transition-all shadow-lg shadow-emerald-600/20 flex items-center justify-center gap-2"
              >
                <Sparkles className="w-4 h-4" />
                <span>Simulate Recommended Fix & Re-run Validation</span>
              </button>
            )}
          </div>

        </div>

      </div>
    </div>
  );
};
