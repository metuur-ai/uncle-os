import React, { useState } from 'react';
import { EXIT_CODES_DATA } from '../data/commandsData';
import { TROUBLESHOOTING_DATA } from '../data/troubleshootingData';
import { BookOpen, Search, AlertTriangle, Layers, Key, Shield, HelpCircle } from 'lucide-react';

export const ReferenceMatrix: React.FC = () => {
  const [activeSection, setActiveSection] = useState<'troubleshooting' | 'precedence' | 'exitcodes'>('troubleshooting');
  const [searchQuery, setSearchQuery] = useState('');

  const filteredTroubleshooting = TROUBLESHOOTING_DATA.filter(item =>
    item.symptom.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.cause.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.fix.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      
      {/* Intro Header */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                07 CONFIG MATRIX
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                DIAGNOSTICS & RULES
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">Configuration Reference & Troubleshooting</h2>
          </div>

          <div className="flex items-center gap-2 bg-slate-100 p-1.5 rounded-xl border border-slate-200 shrink-0">
            <button
              onClick={() => setActiveSection('troubleshooting')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                activeSection === 'troubleshooting' ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20' : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Troubleshooting ({TROUBLESHOOTING_DATA.length})
            </button>
            <button
              onClick={() => setActiveSection('precedence')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                activeSection === 'precedence' ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20' : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Precedence Layers
            </button>
            <button
              onClick={() => setActiveSection('exitcodes')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                activeSection === 'exitcodes' ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20' : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Exit Codes
            </button>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              To give engineers and managers a fast lookup for fixing errors and understanding configuration order.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              A searchable troubleshooting directory, precedence hierarchy map, and exit code reference contract.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Type any error or symptom (e.g., "missing frontmatter" or "exit code 5") into the search bar!
            </p>
          </div>
        </div>
      </div>

      {activeSection === 'troubleshooting' && (
        <div className="space-y-4">
          
          {/* Search bar */}
          <div className="relative">
            <Search className="w-4 h-4 absolute left-3 top-3 text-slate-400" />
            <input
              type="text"
              placeholder="Search symptom, error message, or fix (e.g. 'expired deviation', 'reality doc', 'python')..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full bg-white border border-slate-200 shadow-sm rounded-xl pl-9 pr-4 py-2.5 text-xs text-slate-800 focus:outline-none focus:border-indigo-500 font-mono"
            />
          </div>

          {/* Troubleshooting Table / Cards */}
          <div className="space-y-3">
            {filteredTroubleshooting.map((item) => (
              <div key={item.id} className="bg-white border border-slate-200 shadow-sm rounded-2xl p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-mono font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-rose-50 text-rose-700 border border-rose-200">
                    {item.symptom}
                  </span>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-100 text-slate-600 border border-slate-200">
                    {item.tool} • {item.category}
                  </span>
                </div>

                <div className="text-xs space-y-1">
                  <div>
                    <span className="font-bold text-slate-700">Root Cause: </span>
                    <span className="text-slate-800">{item.cause}</span>
                  </div>
                  <div>
                    <span className="font-bold text-emerald-800">Recommended Fix: </span>
                    <span className="text-emerald-700 font-mono font-medium">{item.fix}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>

        </div>
      )}

      {activeSection === 'precedence' && (
        <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <h3 className="text-sm font-bold text-slate-900">Workspace Root Precedence Layers</h3>
          <p className="text-xs text-slate-600">
            Company OS resolves the workspace root in six specified layers (highest precedence wins):
          </p>

          <div className="space-y-2 text-xs">
            <div className="p-3 bg-indigo-50 border border-indigo-200 rounded-xl flex items-center justify-between">
              <div>
                <span className="font-bold text-indigo-800 block font-mono">1. CLI Flag (Implemented)</span>
                <code className="text-slate-900 font-semibold">company-os --root /abs/path ...</code>
              </div>
              <span className="px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 border border-emerald-200 text-[10px] font-bold">WIRED</span>
            </div>

            <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center justify-between">
              <div>
                <span className="font-bold text-indigo-800 block font-mono">2. Environment Variable (Implemented)</span>
                <code className="text-slate-900 font-semibold">export COMPANY_OS_WORKSPACE_ROOT=/abs/path</code>
              </div>
              <span className="px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 border border-emerald-200 text-[10px] font-bold">WIRED</span>
            </div>

            <div className="p-3 bg-slate-50/60 border border-slate-200 rounded-xl flex items-center justify-between text-slate-500">
              <div>
                <span className="font-bold block font-mono text-slate-700">3. Repo-Local Override (Specified in Spec)</span>
                <code className="text-slate-600">.company-os.local.yaml (git-ignored)</code>
              </div>
              <span className="px-2 py-0.5 rounded bg-slate-100 text-slate-500 border border-slate-200 text-[10px] font-bold">UNWIRED</span>
            </div>

            <div className="p-3 bg-slate-50/60 border border-slate-200 rounded-xl flex items-center justify-between text-slate-500">
              <div>
                <span className="font-bold block font-mono text-slate-700">4. User-Level Config (Specified in Spec)</span>
                <code className="text-slate-600">~/.company-os/config.yaml</code>
              </div>
              <span className="px-2 py-0.5 rounded bg-slate-100 text-slate-500 border border-slate-200 text-[10px] font-bold">UNWIRED</span>
            </div>

            <div className="p-3 bg-slate-50/60 border border-slate-200 rounded-xl flex items-center justify-between text-slate-500">
              <div>
                <span className="font-bold block font-mono text-slate-700">5. Committed Shared Config (Specified in Spec)</span>
                <code className="text-slate-600">config/repositories.yaml</code>
              </div>
              <span className="px-2 py-0.5 rounded bg-slate-100 text-slate-500 border border-slate-200 text-[10px] font-bold">UNWIRED</span>
            </div>

            <div className="p-3 bg-indigo-50 border border-indigo-200 rounded-xl flex items-center justify-between">
              <div>
                <span className="font-bold text-indigo-800 block font-mono">6. Built-in Default (Implemented)</span>
                <code className="text-slate-900 font-semibold">Current working directory (cwd)</code>
              </div>
              <span className="px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 border border-emerald-200 text-[10px] font-bold">WIRED</span>
            </div>
          </div>
        </div>
      )}

      {activeSection === 'exitcodes' && (
        <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <h3 className="text-sm font-bold text-slate-900">Full Exit Code Contract Matrix</h3>
          <p className="text-xs text-slate-600">
            Every subcommand exits with one of eight codes. The exit code meaning is stable across reworded messages.
          </p>

          <div className="space-y-2 text-xs">
            {EXIT_CODES_DATA.map((ec) => (
              <div key={ec.code} className="bg-slate-50 border border-slate-200 p-3 rounded-xl flex items-start gap-3">
                <span className={`px-2 py-1 rounded font-mono font-bold text-xs shrink-0 ${
                  ec.code === 0 ? 'bg-emerald-100 text-emerald-800 border border-emerald-200' : 'bg-rose-100 text-rose-800 border border-rose-200'
                }`}>
                  Exit Code {ec.code}
                </span>

                <div className="space-y-1">
                  <h4 className="font-bold text-slate-900 text-xs">{ec.meaning}</h4>
                  <p className="text-slate-600 text-[11px]">{ec.whenOccurs}</p>
                  <p className="text-emerald-800 font-mono text-[10px]">{ec.recommendedAction}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

    </div>
  );
};
