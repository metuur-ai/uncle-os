import React, { useState } from 'react';
import { CLI_COMMANDS_DATA, EXIT_CODES_DATA } from '../data/commandsData';
import { CliCommand } from '../types';
import { Terminal, Copy, Check, Play, FileCode, CheckCircle2, AlertTriangle, HelpCircle, Code } from 'lucide-react';

export const CliTerminalView: React.FC = () => {
  const [selectedCommand, setSelectedCommand] = useState<CliCommand>(CLI_COMMANDS_DATA[7]); // default to 'validate'
  const [isJsonMode, setIsJsonMode] = useState<boolean>(false);
  const [customRoot, setCustomRoot] = useState<string>('/Users/you/moonbeam-os');
  const [selectedFlags, setSelectedFlags] = useState<Record<string, string>>({});
  const [terminalOutput, setTerminalOutput] = useState<string>(selectedCommand.expectedOutput);
  const [copied, setCopied] = useState<boolean>(false);

  const handleSelectCommand = (cmd: CliCommand) => {
    setSelectedCommand(cmd);
    setSelectedFlags({});
    setTerminalOutput(isJsonMode ? cmd.jsonOutput : cmd.expectedOutput);
  };

  const toggleJsonFormat = (val: boolean) => {
    setIsJsonMode(val);
    setTerminalOutput(val ? selectedCommand.jsonOutput : selectedCommand.example);
  };

  const handleRunCommand = () => {
    // Generate terminal output
    if (isJsonMode) {
      setTerminalOutput(selectedCommand.jsonOutput);
    } else {
      setTerminalOutput(selectedCommand.expectedOutput);
    }
  };

  const fullCommandLine = `${isJsonMode ? 'company-os --json' : 'company-os'} --root ${customRoot} ${selectedCommand.name} ${
    Object.entries(selectedFlags)
      .filter(([_, v]) => v)
      .map(([k, v]) => `${k} "${v}"`)
      .join(' ')
  }`.trim();

  const handleCopyCommand = () => {
    navigator.clipboard.writeText(fullCommandLine);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-6">
      
      {/* Intro Bento Header */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                02 CLI EXPLORER
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                GO BINARY v1.4.0
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">company-os CLI Terminal Simulator</h2>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <button
              onClick={() => toggleJsonFormat(!isJsonMode)}
              className={`px-3.5 py-2 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition-all ${
                isJsonMode
                  ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30 border border-indigo-500/30'
                  : 'bg-white text-slate-700 hover:bg-slate-100 border border-slate-200 shadow-sm'
              }`}
            >
              <Code className="w-3.5 h-3.5" />
              <span>{isJsonMode ? 'JSON Mode (--json)' : 'Human Text Mode'}</span>
            </button>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              So you can run automated quality checks from your computer terminal without needing any website or cloud server.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              A list of CLI commands on the left, and a simulated terminal terminal window on the right that shows output.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Click <code>validate</code> or <code>resolve-governance</code> on the left, then click the blue <strong>Run Command</strong> button!
            </p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">

        {/* Left: Command Palette / Subcommands list */}
        <div className="lg:col-span-4 bg-white border border-slate-200 shadow-sm rounded-2xl p-4 flex flex-col h-[650px]">
          <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 pb-2 mb-2 border-b border-slate-100">
            Subcommand Reference ({CLI_COMMANDS_DATA.length})
          </h3>

          <div className="flex-1 overflow-y-auto space-y-1.5 pr-1 no-scrollbar">
            {CLI_COMMANDS_DATA.map((cmd) => {
              const isSelected = selectedCommand.id === cmd.id;
              const categoryColor: Record<string, string> = {
                'Scaffolding': 'text-blue-700 bg-blue-50 border-blue-200',
                'Lifecycle': 'text-emerald-700 bg-emerald-50 border-emerald-200',
                'Governance': 'text-amber-800 bg-amber-50 border-amber-200',
                'Validation': 'text-purple-700 bg-purple-50 border-purple-200',
                'Federation': 'text-cyan-700 bg-cyan-50 border-cyan-200',
                'Utility': 'text-slate-700 bg-slate-100 border-slate-200',
              };

              return (
                <button
                  key={cmd.id}
                  onClick={() => handleSelectCommand(cmd)}
                  className={`w-full text-left p-2.5 rounded-xl text-xs transition-all flex flex-col gap-1 border ${
                    isSelected
                      ? 'bg-indigo-50 border-indigo-300 text-slate-900 shadow-sm font-semibold'
                      : 'bg-slate-50/70 border-slate-200 hover:bg-slate-100 text-slate-700'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-mono font-bold text-sm text-indigo-700">{cmd.name}</span>
                    <span className={`text-[10px] px-1.5 py-0.5 rounded border font-semibold ${categoryColor[cmd.category]}`}>
                      {cmd.category}
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-500 line-clamp-2">{cmd.description}</p>
                </button>
              );
            })}
          </div>
        </div>

        {/* Right: Interactive Terminal Simulator & Command Controls */}
        <div className="lg:col-span-8 space-y-4">
          
          {/* Command Configuration Card */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-3 border-b border-slate-100">
              <div>
                <span className="text-xs text-indigo-700 font-mono font-semibold uppercase tracking-wider">
                  {selectedCommand.category} Subcommand
                </span>
                <h3 className="text-lg font-bold text-slate-900 font-mono">{selectedCommand.syntax}</h3>
              </div>

              <div className="flex items-center gap-2">
                <span className="text-xs text-slate-500">Exit codes:</span>
                <div className="flex gap-1">
                  {selectedCommand.exitCodesPossible.map(code => (
                    <span key={code} className="px-1.5 py-0.5 rounded bg-amber-50 text-amber-800 border border-amber-200 text-[10px] font-mono font-bold">
                      {code}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <p className="text-xs text-slate-600 leading-relaxed">{selectedCommand.description}</p>

            {/* Flag Inputs if any */}
            {selectedCommand.flags.length > 0 && (
              <div className="bg-slate-50 p-3 rounded-xl border border-slate-200 space-y-2">
                <span className="text-[11px] font-bold text-slate-600 uppercase tracking-wider">Command Flags</span>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                  {selectedCommand.flags.map(flag => (
                    <div key={flag.flag} className="flex flex-col gap-1">
                      <label className="text-slate-700 font-mono flex items-center gap-1 font-medium">
                        <span>{flag.flag}</span>
                        {flag.required && <span className="text-rose-600 text-[10px]">*req</span>}
                      </label>
                      <input
                        type="text"
                        placeholder={flag.defaultVal || flag.description}
                        value={selectedFlags[flag.flag] || ''}
                        onChange={(e) => setSelectedFlags(prev => ({ ...prev, [flag.flag]: e.target.value }))}
                        className="bg-white border border-slate-200 rounded-lg px-2.5 py-1.5 text-xs text-slate-800 focus:outline-none focus:border-indigo-500 font-mono"
                      />
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Generated Executable Command Bar */}
            <div className="bg-slate-900 p-3 rounded-xl border border-slate-800 flex items-center justify-between gap-3 font-mono text-xs text-indigo-300 shadow-inner">
              <div className="flex items-center gap-2 overflow-x-auto no-scrollbar">
                <span className="text-slate-500">$</span>
                <span className="whitespace-nowrap font-bold text-indigo-200">{fullCommandLine}</span>
              </div>

              <div className="flex items-center gap-2 shrink-0">
                <button
                  onClick={handleCopyCommand}
                  className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs flex items-center gap-1 border border-slate-700"
                  title="Copy command"
                >
                  {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
                <button
                  onClick={handleRunCommand}
                  className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs flex items-center gap-1 shadow-md shadow-indigo-600/20"
                >
                  <Play className="w-3.5 h-3.5" />
                  <span>Execute</span>
                </button>
              </div>
            </div>
          </div>

          {/* Interactive Shell / Terminal Window */}
          <div className="bg-slate-950 border border-slate-800 rounded-2xl overflow-hidden shadow-xl flex flex-col h-[350px]">
            {/* Window titlebar */}
            <div className="bg-slate-900 px-4 py-2 border-b border-slate-800 flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <div className="flex gap-1.5">
                  <div className="w-3 h-3 rounded-full bg-rose-500/80" />
                  <div className="w-3 h-3 rounded-full bg-amber-500/80" />
                  <div className="w-3 h-3 rounded-full bg-emerald-500/80" />
                </div>
                <span className="font-mono text-slate-400 text-[11px] ml-2">bash - company-os cli</span>
              </div>

              <span className="text-[10px] text-emerald-400 font-mono bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                Exit Code: 0 (OK)
              </span>
            </div>

            {/* Output terminal body */}
            <div className="flex-1 p-4 font-mono text-xs text-slate-200 overflow-y-auto leading-relaxed whitespace-pre-wrap select-text">
              <div className="text-slate-500 mb-2">$ {fullCommandLine}</div>
              {isJsonMode ? (
                <span className="text-cyan-300">{selectedCommand.jsonOutput}</span>
              ) : (
                <span className="text-emerald-300">{terminalOutput}</span>
              )}
            </div>
          </div>

          {/* Exit Codes Contract Matrix preview */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-4">
            <h4 className="text-xs font-bold text-slate-800 uppercase tracking-wider mb-2 flex items-center gap-1.5">
              <AlertTriangle className="w-3.5 h-3.5 text-amber-600" />
              <span>CLI Exit Code Contract</span>
            </h4>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-[11px]">
              {EXIT_CODES_DATA.slice(0, 8).map(ec => (
                <div key={ec.code} className="bg-slate-50 p-2 rounded-lg border border-slate-200 flex items-center gap-2">
                  <span className={`px-1.5 py-0.5 rounded font-mono font-bold ${ec.code === 0 ? 'bg-emerald-100 text-emerald-800 border border-emerald-200' : 'bg-rose-100 text-rose-800 border border-rose-200'}`}>
                    {ec.code}
                  </span>
                  <span className="text-slate-700 font-medium truncate">{ec.meaning}</span>
                </div>
              ))}
            </div>
          </div>

        </div>

      </div>
    </div>
  );
};
