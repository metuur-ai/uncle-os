import React, { useState } from 'react';
import { X, BookOpen, Search, HelpCircle, UserCheck, GraduationCap, CheckCircle2, ChevronRight, Sparkles } from 'lucide-react';
import { GLOSSARY_ITEMS, ROLE_GUIDES } from '../data/glossaryData';

interface BeginnerGuideModalProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab: (tab: string) => void;
}

export const BeginnerGuideModal: React.FC<BeginnerGuideModalProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
}) => {
  const [activeSection, setActiveSection] = useState<'primer' | 'glossary' | 'guide'>('primer');
  const [searchTerm, setSearchTerm] = useState('');

  if (!isOpen) return null;

  const filteredGlossary = GLOSSARY_ITEMS.filter(
    item => item.term.toLowerCase().includes(searchTerm.toLowerCase()) || item.plain.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 overflow-y-auto">
      <div className="bg-white border border-slate-200 rounded-3xl shadow-2xl max-w-3xl w-full overflow-hidden flex flex-col max-h-[90vh]">
        
        {/* Header */}
        <div className="bg-gradient-to-r from-indigo-600 to-indigo-700 p-6 text-white flex items-center justify-between shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-white/10 rounded-2xl flex items-center justify-center border border-white/20">
              <BookOpen className="w-6 h-6 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] font-bold uppercase tracking-wider bg-white/20 px-2 py-0.5 rounded-full text-indigo-100 font-mono">
                  ACCESSIBLE GUIDE
                </span>
              </div>
              <h2 className="text-xl font-bold">Plain English Helper & Glossary</h2>
              <p className="text-xs text-indigo-100 mt-0.5">
                Simple explanations for interns, new hires, and non-technical teammates.
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="w-9 h-9 bg-white/10 hover:bg-white/20 rounded-full flex items-center justify-center text-white transition-colors"
            aria-label="Close modal"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="bg-slate-50 border-b border-slate-200 px-6 py-2.5 flex items-center gap-2 overflow-x-auto shrink-0">
          <button
            onClick={() => setActiveSection('primer')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 ${
              activeSection === 'primer'
                ? 'bg-indigo-600 text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-200/60'
            }`}
          >
            <HelpCircle className="w-4 h-4" />
            <span>1. How Company OS Works</span>
          </button>

          <button
            onClick={() => setActiveSection('glossary')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 ${
              activeSection === 'glossary'
                ? 'bg-indigo-600 text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-200/60'
            }`}
          >
            <Search className="w-4 h-4" />
            <span>2. Plain English Glossary</span>
          </button>

          <button
            onClick={() => setActiveSection('guide')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 ${
              activeSection === 'guide'
                ? 'bg-indigo-600 text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-200/60'
            }`}
          >
            <GraduationCap className="w-4 h-4" />
            <span>3. Interns & Non-Technical Managers Guide</span>
          </button>
        </div>

        {/* Modal Content */}
        <div className="p-6 overflow-y-auto space-y-6 text-slate-800 leading-relaxed text-sm">
          
          {/* Section 1: Primer */}
          {activeSection === 'primer' && (
            <div className="space-y-5">
              <div className="bg-indigo-50 border border-indigo-200 p-5 rounded-2xl space-y-2">
                <div className="flex items-center gap-2 text-indigo-900 font-bold text-base">
                  <Sparkles className="w-5 h-5 text-indigo-600" />
                  <span>What is Company OS in simple terms?</span>
                </div>
                <p className="text-slate-700 text-xs sm:text-sm leading-relaxed">
                  Imagine if every team at a big company organized their projects, rules, and documents in completely different ways. It would be total chaos! <strong>Company OS</strong> is a unified, easy-to-follow system that keeps everyone on the exact same page using clear, structured files and automatic quality checks.
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-white border border-slate-200 p-4 rounded-2xl shadow-sm space-y-2">
                  <span className="w-7 h-7 bg-indigo-100 text-indigo-800 rounded-lg flex items-center justify-center font-bold text-xs">1</span>
                  <h4 className="font-bold text-slate-900 text-sm">Organized Folders</h4>
                  <p className="text-slate-600 text-xs">
                    Every team has dedicated folders for discovery, requirements, and active projects. No lost files in emails!
                  </p>
                </div>

                <div className="bg-white border border-slate-200 p-4 rounded-2xl shadow-sm space-y-2">
                  <span className="w-7 h-7 bg-indigo-100 text-indigo-800 rounded-lg flex items-center justify-center font-bold text-xs">2</span>
                  <h4 className="font-bold text-slate-900 text-sm">Automated Safety Checks</h4>
                  <p className="text-slate-600 text-xs">
                    Before work is approved, <code>company-os validate</code> automatically scans for security, formatting, and missing info.
                  </p>
                </div>

                <div className="bg-white border border-slate-200 p-4 rounded-2xl shadow-sm space-y-2">
                  <span className="w-7 h-7 bg-indigo-100 text-indigo-800 rounded-lg flex items-center justify-center font-bold text-xs">3</span>
                  <h4 className="font-bold text-slate-900 text-sm">Always Accurate</h4>
                  <p className="text-slate-600 text-xs">
                    "Representation of Reality": When code changes, docs must change too. No outdated documentation!
                  </p>
                </div>
              </div>

              <div className="pt-2 flex justify-end">
                <button
                  onClick={() => {
                    onClose();
                    onNavigateTab('workflows');
                  }}
                  className="px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs flex items-center gap-2 shadow-md"
                >
                  <span>Try Interactive Workflow Playground</span>
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {/* Section 2: Glossary */}
          {activeSection === 'glossary' && (
            <div className="space-y-4">
              <div className="relative">
                <Search className="w-4 h-4 absolute left-3.5 top-3 text-slate-400" />
                <input
                  type="text"
                  placeholder="Search technical terms (e.g. PRD, Gate, Deviation)..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl pl-10 pr-4 py-2 text-xs text-slate-800 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="space-y-2.5 max-h-[350px] overflow-y-auto pr-1">
                {filteredGlossary.map((item, idx) => (
                  <div key={idx} className="p-3.5 bg-slate-50 border border-slate-200 rounded-xl space-y-1">
                    <span className="font-bold text-indigo-900 text-xs sm:text-sm block">{item.term}</span>
                    <p className="text-slate-600 text-xs leading-relaxed">{item.plain}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Section 3: Combined Intern & Non-Technical Manager Guide */}
          {activeSection === 'guide' && (
            <div className="space-y-6">
              
              {/* Part 1: Interns & New Engineers */}
              <div className="space-y-3">
                <div className="bg-emerald-50 border border-emerald-200 p-4 rounded-2xl text-emerald-900 text-xs space-y-1">
                  <div className="flex items-center gap-2">
                    <GraduationCap className="w-4 h-4 text-emerald-700" />
                    <span className="font-bold text-sm block">For 18yo Interns & New Engineers 🎓 (3-Step Cheat Sheet)</span>
                  </div>
                  <p className="text-slate-700 text-xs mt-0.5">
                    Company OS makes it impossible to break things secretly. Follow these 3 simple steps when working on a task:
                  </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <div className="p-3.5 bg-white border border-slate-200 rounded-2xl space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 font-bold text-xs flex items-center justify-center shrink-0">1</span>
                      <h4 className="font-bold text-slate-900 text-xs">Check specs first</h4>
                    </div>
                    <p className="text-slate-600 text-[11px] leading-relaxed">
                      Use <strong>Local Search</strong> to check existing components and specs before building anything new.
                    </p>
                  </div>

                  <div className="p-3.5 bg-white border border-slate-200 rounded-2xl space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 font-bold text-xs flex items-center justify-center shrink-0">2</span>
                      <h4 className="font-bold text-slate-900 text-xs">Run validation</h4>
                    </div>
                    <p className="text-slate-600 text-[11px] leading-relaxed">
                      Run <code>company-os validate</code> in the <strong>Terminal Explorer</strong>. If a gate fails, fix the reported error.
                    </p>
                  </div>

                  <div className="p-3.5 bg-white border border-slate-200 rounded-2xl space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 font-bold text-xs flex items-center justify-center shrink-0">3</span>
                      <h4 className="font-bold text-slate-900 text-xs">Update docs</h4>
                    </div>
                    <p className="text-slate-600 text-[11px] leading-relaxed">
                      Bump the <code>updated: YYYY-MM-DD</code> date in the reality document so docs match the code!
                    </p>
                  </div>
                </div>

                <div className="flex justify-end">
                  <button
                    onClick={() => {
                      onClose();
                      onNavigateTab('cli');
                    }}
                    className="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs flex items-center gap-1.5 transition-colors"
                  >
                    <span>Try Terminal Explorer</span>
                    <ChevronRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>

              <div className="border-t border-slate-200 my-2"></div>

              {/* Part 2: Non-Technical Managers */}
              <div className="space-y-3">
                <div className="bg-amber-50 border border-amber-200 p-4 rounded-2xl text-amber-900 text-xs space-y-1">
                  <div className="flex items-center gap-2">
                    <UserCheck className="w-4 h-4 text-amber-700" />
                    <span className="font-bold text-sm block">For Non-Technical Managers & Reviewers 💼</span>
                  </div>
                  <p className="text-slate-700 text-xs mt-0.5">
                    You don't need to write code to understand project health. Company OS gives you full transparent visibility:
                  </p>
                </div>

                <div className="space-y-2.5">
                  <div className="p-3.5 bg-white border border-slate-200 rounded-2xl space-y-1">
                    <span className="font-bold text-slate-900 text-xs block">1. How do I know if a project is safe?</span>
                    <p className="text-slate-600 text-xs leading-relaxed">
                      Look at the <strong>Validation Gates</strong> tab. If all 8 gates show PASS (green checkmarks), the project meets all compliance and quality standards.
                    </p>
                  </div>

                  <div className="p-3.5 bg-white border border-slate-200 rounded-2xl space-y-1">
                    <span className="font-bold text-slate-900 text-xs block">2. What if my team needs a rule exception?</span>
                    <p className="text-slate-600 text-xs leading-relaxed">
                      Visit the <strong>Governance Tiers</strong> tab. You can submit a transparent, time-boxed <em>Deviation</em> or <em>Exception</em> with rationale without breaking the system.
                    </p>
                  </div>

                  <div className="p-3.5 bg-white border border-slate-200 rounded-2xl space-y-1">
                    <span className="font-bold text-slate-900 text-xs block">3. Where do I test my knowledge?</span>
                    <p className="text-slate-600 text-xs leading-relaxed">
                      Try the 8-question <strong>Mastery Check</strong> quiz. It explains every answer in plain language so you feel confident discussing project governance.
                    </p>
                  </div>
                </div>

                <div className="flex justify-end gap-2 pt-1">
                  <button
                    onClick={() => {
                      onClose();
                      onNavigateTab('validation');
                    }}
                    className="px-3.5 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-800 font-bold text-xs flex items-center gap-1.5 border border-slate-200 transition-colors"
                  >
                    <span>View Safety Gates</span>
                  </button>

                  <button
                    onClick={() => {
                      onClose();
                      onNavigateTab('governance');
                    }}
                    className="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs flex items-center gap-1.5 transition-colors"
                  >
                    <span>Explore Governance Tiers</span>
                    <ChevronRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>

            </div>
          )}

        </div>

        {/* Footer */}
        <div className="bg-slate-100 border-t border-slate-200 px-6 py-3.5 flex justify-between items-center shrink-0 text-xs text-slate-500">
          <span>Tip: You can reopen this guide anytime from the top header menu.</span>
          <button
            onClick={onClose}
            className="px-4 py-1.5 rounded-xl bg-slate-900 text-white font-bold hover:bg-slate-800 transition-colors"
          >
            Got it, thanks!
          </button>
        </div>

      </div>
    </div>
  );
};
