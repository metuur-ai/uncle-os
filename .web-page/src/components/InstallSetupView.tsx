import React, { useState } from 'react';
import {
  INSTALL_STEPS,
  INSTALL_ARTIFACTS,
  INSTALL_OPTIONS,
  INSTALL_PATHS,
  INSTALL_FAQS,
  COMPANION_TOOL,
  INSTALL_ONE_LINER,
} from '../data/installData';
import {
  Download,
  Terminal,
  Copy,
  Check,
  Cpu,
  Sliders,
  HelpCircle,
  Search,
  ChevronDown,
  ExternalLink,
  ShieldCheck,
} from 'lucide-react';

export const InstallSetupView: React.FC = () => {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [activePath, setActivePath] = useState<string>(INSTALL_PATHS[0].id);
  const [openFaq, setOpenFaq] = useState<string | null>(INSTALL_FAQS[0].id);

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const selectedPath =
    INSTALL_PATHS.find((p) => p.id === activePath) || INSTALL_PATHS[0];

  return (
    <div className="space-y-6">

      {/* Intro Header */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[11px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                01 INSTALL &amp; SETUP
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[11px] font-semibold text-indigo-800 border border-indigo-200">
                START HERE
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">
              Install the company-os CLI
            </h2>
          </div>

          <div className="flex items-center gap-2 text-xs shrink-0">
            <span className="px-3 py-1.5 rounded-xl bg-white text-slate-700 border border-slate-200 font-semibold shadow-sm">
              Single Static Binary
            </span>
            <span className="px-3 py-1.5 rounded-xl bg-indigo-600 text-white font-semibold shadow-md shadow-indigo-600/20 border border-indigo-500/30">
              Zero Runtime Deps
            </span>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[11px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[12px] leading-relaxed">
              Everything else on this site assumes you can run <code className="font-mono">company-os</code>. This page gets you there in one line.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[11px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[12px] leading-relaxed">
              Four numbered steps from empty machine to validated workspace, plus install options, platform artifacts, and the common failure modes.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[11px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[12px] leading-relaxed">
              Copy the command in the black box below and paste it into a terminal. Then work down the steps in order.
            </p>
          </div>
        </div>
      </div>

      {/* The one-liner, front and center */}
      <div className="bg-indigo-950 rounded-2xl border border-indigo-900 shadow-lg overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2 border-b border-indigo-900 bg-[#171340]/50">
          <div className="flex items-center gap-2">
            <Download className="w-3.5 h-3.5 text-indigo-400" />
            <span className="text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400">
              The one-line install
            </span>
          </div>
          <span className="text-[11px] font-mono text-slate-500">macOS · Linux</span>
        </div>
        <div className="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center gap-3">
          <code className="flex-1 font-mono text-[12px] sm:text-xs text-emerald-300 break-all leading-relaxed">
            <span className="text-slate-500 select-none">$ </span>
            {INSTALL_ONE_LINER}
          </code>
          <button
            onClick={() => handleCopy('one-liner', INSTALL_ONE_LINER)}
            className="shrink-0 px-3 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs flex items-center justify-center gap-1.5 transition-colors border border-indigo-500/30 shadow-md shadow-indigo-600/20"
          >
            {copiedId === 'one-liner' ? (
              <><Check className="w-3.5 h-3.5" /> Copied</>
            ) : (
              <><Copy className="w-3.5 h-3.5" /> Copy</>
            )}
          </button>
        </div>
      </div>

      {/* Numbered steps */}
      <div className="space-y-4">
        {INSTALL_STEPS.map((step) => (
          <div
            key={step.stepNumber}
            className="bg-white border border-slate-200 shadow-sm rounded-2xl overflow-hidden"
          >
            <div className="grid grid-cols-1 lg:grid-cols-12">

              {/* Left: what and why */}
              <div className="lg:col-span-5 p-5 space-y-3 border-b lg:border-b-0 lg:border-r border-slate-100">
                <div className="flex items-start gap-3">
                  <span className="shrink-0 w-7 h-7 rounded-xl bg-indigo-600 text-white font-mono font-bold text-xs flex items-center justify-center shadow-md shadow-indigo-600/20">
                    {step.stepNumber}
                  </span>
                  <div>
                    <h3 className="text-base font-bold text-slate-900 leading-tight">{step.title}</h3>
                    <p className="text-[12px] text-slate-600 leading-relaxed mt-1.5">{step.description}</p>
                  </div>
                </div>

                <div className="bg-amber-50 border border-amber-200 rounded-xl p-3 flex gap-2">
                  <ShieldCheck className="w-3.5 h-3.5 text-amber-700 shrink-0 mt-0.5" />
                  <div>
                    <span className="block font-bold text-amber-900 font-mono uppercase text-[10px] tracking-wider">
                      Key rule
                    </span>
                    <p className="text-[12px] text-amber-900/90 leading-relaxed mt-0.5">{step.keyRule}</p>
                  </div>
                </div>
              </div>

              {/* Right: the command and its output */}
              <div className="lg:col-span-7 bg-indigo-950 p-4 sm:p-5 space-y-3">
                <div className="flex items-start justify-between gap-3">
                  <pre className="flex-1 font-mono text-[12px] text-emerald-300 whitespace-pre-wrap break-all leading-relaxed">
{step.command.split('\n').map((line, i) => (
  <React.Fragment key={i}>
    <span className="text-slate-500 select-none">$ </span>{line}{'\n'}
  </React.Fragment>
))}
                  </pre>
                  <button
                    onClick={() => handleCopy(`step-${step.stepNumber}`, step.command)}
                    aria-label={`Copy command for step ${step.stepNumber}`}
                    className="shrink-0 p-1.5 rounded-lg bg-indigo-900 hover:bg-indigo-800 text-slate-300 border border-indigo-800 transition-colors"
                  >
                    {copiedId === `step-${step.stepNumber}` ? (
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                    ) : (
                      <Copy className="w-3.5 h-3.5" />
                    )}
                  </button>
                </div>

                <div className="border-t border-indigo-900 pt-3">
                  <span className="block text-[10px] font-mono font-bold uppercase tracking-widest text-slate-500 mb-1.5">
                    Expected output
                  </span>
                  <pre className="font-mono text-[11.5px] text-slate-400 whitespace-pre-wrap leading-relaxed">
                    {step.mockTerminalOutput}
                  </pre>
                </div>
              </div>

            </div>
          </div>
        ))}
      </div>

      {/* Install paths + platform artifacts */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">

        {/* Alternative install paths */}
        <div className="lg:col-span-7 bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="flex items-center gap-2 border-b border-slate-100 pb-3">
            <Terminal className="w-4 h-4 text-indigo-600" />
            <h3 className="text-base font-bold text-slate-900">Three ways to install</h3>
          </div>

          <div className="flex flex-wrap gap-2">
            {INSTALL_PATHS.map((path) => {
              const isActive = activePath === path.id;
              return (
                <button
                  key={path.id}
                  onClick={() => setActivePath(path.id)}
                  className={`px-3 py-2 rounded-xl text-xs font-semibold border transition-all ${
                    isActive
                      ? 'bg-indigo-50 border-indigo-300 text-slate-900 shadow-sm'
                      : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-100/70'
                  }`}
                >
                  {path.label}
                </button>
              );
            })}
          </div>

          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="px-2 py-0.5 rounded-full bg-indigo-100/70 text-[11px] font-bold text-indigo-800 border border-indigo-200 font-mono">
                {selectedPath.badge}
              </span>
              <span className="text-[12px] text-slate-500">
                Requires: {selectedPath.requirement}
              </span>
            </div>

            <p className="text-xs text-slate-600 leading-relaxed">{selectedPath.summary}</p>

            <div className="bg-indigo-950 rounded-xl p-4 flex items-start justify-between gap-3">
              <pre className="flex-1 font-mono text-[12px] text-emerald-300 whitespace-pre-wrap break-all leading-relaxed">
{selectedPath.commands.map((line, i) => (
  <React.Fragment key={i}>
    <span className="text-slate-500 select-none">$ </span>{line}{'\n'}
  </React.Fragment>
))}
              </pre>
              <button
                onClick={() => handleCopy(`path-${selectedPath.id}`, selectedPath.commands.join('\n'))}
                aria-label={`Copy ${selectedPath.label} commands`}
                className="shrink-0 p-1.5 rounded-lg bg-indigo-900 hover:bg-indigo-800 text-slate-300 border border-indigo-800 transition-colors"
              >
                {copiedId === `path-${selectedPath.id}` ? (
                  <Check className="w-3.5 h-3.5 text-emerald-400" />
                ) : (
                  <Copy className="w-3.5 h-3.5" />
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Platform artifacts */}
        <div className="lg:col-span-5 bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="flex items-center gap-2 border-b border-slate-100 pb-3">
            <Cpu className="w-4 h-4 text-indigo-600" />
            <h3 className="text-base font-bold text-slate-900">Release artifacts</h3>
          </div>

          <div className="space-y-2">
            {INSTALL_ARTIFACTS.map((artifact) => (
              <div
                key={artifact.filename}
                className="bg-slate-50 border border-slate-200 rounded-xl p-3 space-y-1"
              >
                <code className="block font-mono text-[12px] font-bold text-indigo-700 break-all">
                  {artifact.filename}
                </code>
                <div className="flex items-center justify-between gap-2">
                  <span className="text-[12px] text-slate-700">{artifact.runsOn}</span>
                  <span className="text-[11px] font-mono text-slate-500 shrink-0">
                    {artifact.detectedAs}
                  </span>
                </div>
              </div>
            ))}
          </div>

          <p className="text-[12px] text-slate-600 leading-relaxed bg-slate-50 border border-slate-200 rounded-xl p-3">
            The names carry <strong>no version</strong> — the installer resolves
            <code className="font-mono text-[11px] mx-1">/releases/latest/download/&lt;name&gt;</code>
            and that URL needs a fixed filename. The version lives in the release tag and is
            stamped into the binary.
          </p>
        </div>
      </div>

      {/* Installer options */}
      <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
        <div className="flex items-center gap-2 border-b border-slate-100 pb-3">
          <Sliders className="w-4 h-4 text-indigo-600" />
          <h3 className="text-base font-bold text-slate-900">Installer options</h3>
          <span className="text-[12px] text-slate-500">— set as environment variables</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {INSTALL_OPTIONS.map((opt) => (
            <div
              key={opt.envVar}
              className="bg-slate-50 border border-slate-200 rounded-xl p-3 space-y-2"
            >
              <div className="flex items-center justify-between gap-2">
                <code className="font-mono text-xs font-bold text-indigo-700">{opt.envVar}</code>
                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-white text-slate-600 border border-slate-200 shrink-0">
                  {opt.defaultVal}
                </span>
              </div>
              <p className="text-[12px] text-slate-600 leading-relaxed">{opt.description}</p>
              <code className="block font-mono text-[11px] text-slate-500 bg-white border border-slate-200 rounded-lg p-2 break-all">
                {opt.example}
              </code>
            </div>
          ))}
        </div>
      </div>

      {/* FAQ / failure modes */}
      <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
        <div className="flex items-center gap-2 border-b border-slate-100 pb-3">
          <HelpCircle className="w-4 h-4 text-indigo-600" />
          <h3 className="text-base font-bold text-slate-900">
            Gatekeeper, upgrades, and the ways this goes wrong
          </h3>
        </div>

        <div className="space-y-2">
          {INSTALL_FAQS.map((faq) => {
            const isOpen = openFaq === faq.id;
            return (
              <div
                key={faq.id}
                className={`rounded-xl border transition-all ${
                  isOpen ? 'border-indigo-200 bg-indigo-50/40' : 'border-slate-200 bg-white'
                }`}
              >
                <button
                  onClick={() => setOpenFaq(isOpen ? null : faq.id)}
                  aria-expanded={isOpen}
                  className="w-full flex items-center justify-between gap-3 p-3.5 text-left"
                >
                  <span className="text-xs font-bold text-slate-900">{faq.question}</span>
                  <ChevronDown
                    className={`w-4 h-4 text-slate-500 shrink-0 transition-transform ${
                      isOpen ? 'rotate-180' : ''
                    }`}
                  />
                </button>

                {isOpen && (
                  <div className="px-3.5 pb-3.5 space-y-3">
                    <p className="text-[12px] text-slate-600 leading-relaxed">{faq.answer}</p>
                    {faq.command && (
                      <div className="bg-indigo-950 rounded-xl p-3 flex items-start justify-between gap-3">
                        <pre className="flex-1 font-mono text-[11.5px] text-emerald-300 whitespace-pre-wrap break-all leading-relaxed">
                          {faq.command}
                        </pre>
                        <button
                          onClick={() => handleCopy(`faq-${faq.id}`, faq.command as string)}
                          aria-label={`Copy command for: ${faq.question}`}
                          className="shrink-0 p-1.5 rounded-lg bg-indigo-900 hover:bg-indigo-800 text-slate-300 border border-indigo-800 transition-colors"
                        >
                          {copiedId === `faq-${faq.id}` ? (
                            <Check className="w-3.5 h-3.5 text-emerald-400" />
                          ) : (
                            <Copy className="w-3.5 h-3.5" />
                          )}
                        </button>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Companion tool */}
      <div className="bg-gradient-to-br from-cyan-50 via-white to-slate-50 border border-cyan-100 shadow-sm rounded-2xl p-5 space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-cyan-100/80 pb-3">
          <div className="flex items-center gap-2">
            <Search className="w-4 h-4 text-cyan-700" />
            <h3 className="text-base font-bold text-slate-900">
              Optional companion: <code className="font-mono text-sm">{COMPANION_TOOL.name}</code>
            </h3>
            <span className="px-2 py-0.5 rounded-full bg-cyan-100/70 text-[11px] font-bold text-cyan-800 border border-cyan-200 font-mono">
              SEPARATE BINARY
            </span>
          </div>
          <a
            href={COMPANION_TOOL.repo}
            target="_blank"
            rel="noopener noreferrer"
            className="text-[12px] font-semibold text-cyan-700 hover:text-cyan-900 flex items-center gap-1 shrink-0"
          >
            View repo <ExternalLink className="w-3 h-3" />
          </a>
        </div>

        <p className="text-xs text-slate-600 leading-relaxed">{COMPANION_TOOL.summary}</p>

        <div className="space-y-2">
          {COMPANION_TOOL.steps.map((step, i) => (
            <div
              key={step.command}
              className="bg-white/80 border border-slate-200/80 rounded-xl p-3 space-y-2"
            >
              <p className="text-[12px] text-slate-600 leading-relaxed">{step.description}</p>
              <div className="bg-indigo-950 rounded-lg p-2.5 flex items-start justify-between gap-3">
                <code className="flex-1 font-mono text-[11.5px] text-cyan-300 break-all leading-relaxed">
                  <span className="text-slate-500 select-none">$ </span>{step.command}
                </code>
                <button
                  onClick={() => handleCopy(`companion-${i}`, step.command)}
                  aria-label={`Copy local-search command ${i + 1}`}
                  className="shrink-0 p-1.5 rounded-lg bg-indigo-900 hover:bg-indigo-800 text-slate-300 border border-indigo-800 transition-colors"
                >
                  {copiedId === `companion-${i}` ? (
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                  ) : (
                    <Copy className="w-3.5 h-3.5" />
                  )}
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

    </div>
  );
};
