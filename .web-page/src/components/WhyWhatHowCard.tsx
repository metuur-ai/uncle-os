import React, { useState } from 'react';
import { Target, HelpCircle, Compass, PlayCircle, FolderTree, Terminal, CheckSquare, ChevronRight, BookOpen, Layers } from 'lucide-react';
import { TabType } from '../types';
import { WHY_CONTENT, WHAT_POINTS, HOW_STEPS } from '../data/whyWhatHowData';
import { Tooltip } from './Tooltip';

interface WhyWhatHowCardProps {
  onNavigate: (tab: TabType) => void;
  onOpenGuide: () => void;
}

export const WhyWhatHowCard: React.FC<WhyWhatHowCardProps> = ({ onNavigate, onOpenGuide }) => {
  const [activeTab, setActiveTab] = useState<'why' | 'what' | 'how'>('why');

  const getWhatIcon = (iconType: string) => {
    switch (iconType) {
      case 'layers': return <Layers className="w-4 h-4 text-indigo-600" />;
      case 'check': return <CheckSquare className="w-4 h-4 text-indigo-600" />;
      case 'terminal': return <Terminal className="w-4 h-4 text-indigo-600" />;
      default: return <BookOpen className="w-4 h-4 text-indigo-600" />;
    }
  };

  return (
    <div className="bg-white border border-slate-200 rounded-2xl shadow-xs overflow-hidden">
      {/* Top Banner Bar */}
      <div className="bg-slate-900 p-5 sm:p-6 text-white flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="px-2 py-0.5 rounded bg-slate-800 text-indigo-300 border border-slate-700 text-[10px] font-bold uppercase tracking-wider font-mono">
              SIMPLE GUIDE
            </span>
            <span className="text-xs text-slate-400 font-medium hidden sm:inline">
              For Interns, Engineers & Managers
            </span>
          </div>
          <h2 className="text-xl font-bold tracking-tight text-white">
            Company OS & Team OS Explained Simply
          </h2>
          <p className="text-xs sm:text-sm text-slate-300 mt-1 max-w-2xl leading-relaxed">
            Understand Company OS in 3 straightforward questions: <strong className="text-white font-bold">Why</strong> it exists, <strong className="text-white font-bold">What</strong> it is, and <strong className="text-white font-bold">How</strong> to use it.
          </p>
        </div>

        {/* Quick Action Button for Glossary */}
        <Tooltip
          title="Why click glossary?"
          content="Open simple plain English definitions for technical concepts like PRDs, canonical rules, and validation gates."
          position="left"
        >
          <button
            onClick={onOpenGuide}
            className="self-start md:self-auto px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-100 font-semibold text-xs flex items-center gap-2 border border-slate-700 transition-all shrink-0"
          >
            <BookOpen className="w-4 h-4 text-indigo-400" />
            <span>Plain English Glossary</span>
          </button>
        </Tooltip>
      </div>

      {/* Why / What / How Tab Bar */}
      <div className="bg-slate-50 border-b border-slate-200 px-4 sm:px-6 pt-3 flex items-center gap-2 overflow-x-auto no-scrollbar">
        <Tooltip title="Why click this?" content="Understand the core organizational problem Company OS solves." position="bottom">
          <button
            onClick={() => setActiveTab('why')}
            className={`px-4 py-2 rounded-t-xl text-xs sm:text-sm font-bold transition-all flex items-center gap-2 border-t border-x ${
              activeTab === 'why'
                ? 'bg-white border-slate-200 text-indigo-700 -mb-px pb-2.5'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <HelpCircle className="w-4 h-4 text-rose-600" />
            <span>1. WHY does this exist?</span>
          </button>
        </Tooltip>

        <Tooltip title="Why click this?" content="See the 3 fundamental pillars that define Company OS architecture." position="bottom">
          <button
            onClick={() => setActiveTab('what')}
            className={`px-4 py-2 rounded-t-xl text-xs sm:text-sm font-bold transition-all flex items-center gap-2 border-t border-x ${
              activeTab === 'what'
                ? 'bg-white border-slate-200 text-indigo-700 -mb-px pb-2.5'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Target className="w-4 h-4 text-indigo-600" />
            <span>2. WHAT is Company OS?</span>
          </button>
        </Tooltip>

        <Tooltip title="Why click this?" content="Get quick guidance on how to explore and use this interactive app." position="bottom">
          <button
            onClick={() => setActiveTab('how')}
            className={`px-4 py-2 rounded-t-xl text-xs sm:text-sm font-bold transition-all flex items-center gap-2 border-t border-x ${
              activeTab === 'how'
                ? 'bg-white border-slate-200 text-indigo-700 -mb-px pb-2.5'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Compass className="w-4 h-4 text-emerald-600" />
            <span>3. HOW do I use this app?</span>
          </button>
        </Tooltip>
      </div>

      {/* Content Area */}
      <div className="p-5 sm:p-6 bg-white">
        
        {/* WHY TAB */}
        {activeTab === 'why' && (
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              
              {/* The Old Problem */}
              <div className="bg-rose-50/60 border border-rose-200 p-4 sm:p-5 rounded-xl space-y-2">
                <div className="flex items-center gap-2 text-rose-900 font-bold text-sm">
                  <span className="w-5 h-5 rounded bg-rose-200 text-rose-800 flex items-center justify-center text-xs font-mono">✕</span>
                  <span>{WHY_CONTENT.problemTitle}</span>
                </div>
                <p className="text-slate-700 text-xs sm:text-sm leading-relaxed">
                  {WHY_CONTENT.problemDescription}
                </p>
              </div>

              {/* The Solution */}
              <div className="bg-emerald-50/60 border border-emerald-200 p-4 sm:p-5 rounded-xl space-y-2">
                <div className="flex items-center gap-2 text-emerald-900 font-bold text-sm">
                  <span className="w-5 h-5 rounded bg-emerald-200 text-emerald-800 flex items-center justify-center text-xs font-mono">✓</span>
                  <span>{WHY_CONTENT.solutionTitle}</span>
                </div>
                <p className="text-slate-700 text-xs sm:text-sm leading-relaxed">
                  {WHY_CONTENT.solutionDescription}
                </p>
              </div>

            </div>

            <div className="bg-indigo-50/60 border border-indigo-100 p-3.5 rounded-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs">
              <div className="flex items-center gap-2 text-slate-700 font-medium">
                <Compass className="w-4 h-4 text-indigo-600 shrink-0" />
                <span>Want to see what is inside? Click to see the 3 core pillars.</span>
              </div>
              <button
                onClick={() => setActiveTab('what')}
                className="px-3.5 py-1.5 rounded-lg bg-indigo-600 text-white font-bold text-xs flex items-center gap-1.5 shrink-0 hover:bg-indigo-500 transition-colors"
              >
                <span>Next: What is it?</span>
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* WHAT TAB */}
        {activeTab === 'what' && (
          <div className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              {WHAT_POINTS.map((pt, idx) => (
                <div key={idx} className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="w-6 h-6 rounded bg-indigo-100 text-indigo-800 flex items-center justify-center font-bold text-xs font-mono">
                      {idx + 1}
                    </span>
                    {getWhatIcon(pt.iconType)}
                  </div>
                  <h3 className="font-bold text-slate-900 text-sm">{pt.title}</h3>
                  <p className="text-xs font-semibold text-indigo-600">{pt.subtitle}</p>
                  <p className="text-slate-600 text-xs leading-relaxed">
                    {pt.description}
                  </p>
                </div>
              ))}
            </div>

            <div className="bg-indigo-50/60 border border-indigo-100 p-3.5 rounded-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs">
              <div className="flex items-center gap-2 text-slate-700 font-medium">
                <Compass className="w-4 h-4 text-indigo-600 shrink-0" />
                <span>Ready to try it out? Learn HOW to navigate this interactive app!</span>
              </div>
              <button
                onClick={() => setActiveTab('how')}
                className="px-3.5 py-1.5 rounded-lg bg-indigo-600 text-white font-bold text-xs flex items-center gap-1.5 shrink-0 hover:bg-indigo-500 transition-colors"
              >
                <span>Next: How do I use it?</span>
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* HOW TAB */}
        {activeTab === 'how' && (
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {HOW_STEPS.map((step) => (
                <div key={step.stepNumber} className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-3 flex flex-col justify-between">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold font-mono text-indigo-800 uppercase">Step {step.stepNumber}</span>
                      <FolderTree className="w-4 h-4 text-indigo-600" />
                    </div>
                    <h3 className="font-bold text-slate-900 text-sm">{step.title}</h3>
                    <p className="text-slate-600 text-xs leading-relaxed">
                      {step.description}
                    </p>
                  </div>
                  <Tooltip title="Why click action?" content={`Jump directly to the ${step.targetTab} view.`} position="top" className="w-full">
                    <button
                      onClick={() => onNavigate(step.targetTab)}
                      className="w-full py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold text-xs transition-colors mt-2"
                    >
                      {step.actionLabel}
                    </button>
                  </Tooltip>
                </div>
              ))}
            </div>

            <div className="p-4 bg-slate-900 rounded-xl text-white flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs">
              <div>
                <span className="font-bold text-indigo-300 block">Want an interactive guided simulation?</span>
                <span className="text-slate-300">Run through a step-by-step PRD lifecycle from discovery brief to production release!</span>
              </div>
              <Tooltip title="Why launch workflow?" content="Opens the interactive step-by-step PRD lifecycle simulation." position="left">
                <button
                  onClick={() => onNavigate('workflows')}
                  className="px-3.5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs flex items-center gap-2 shrink-0 transition-colors"
                >
                  <PlayCircle className="w-4 h-4" />
                  <span>Launch Workflow Simulator</span>
                </button>
              </Tooltip>
            </div>
          </div>
        )}

      </div>
    </div>
  );
};
