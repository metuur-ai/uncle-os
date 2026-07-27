import React, { useState } from 'react';
import { 
  Building2, 
  Users, 
  GitBranch, 
  ShieldCheck, 
  Zap, 
  Layers, 
  Terminal, 
  FolderTree, 
  Award, 
  CheckCircle2, 
  ArrowRight, 
  HelpCircle,
  Compass,
  RefreshCw,
  FileCode2,
  ListCheck
} from 'lucide-react';
import { TabType } from '../types';
import { Tooltip } from './Tooltip';
import { 
  HOME_HERO_CONTENT, 
  DUAL_CORE_COMPARISON, 
  LIFECYCLE_STEPS, 
  QUICK_DIRECTORY_CARDS 
} from '../data/homeData';

interface HomeOverviewProps {
  onNavigateTab: (tab: TabType) => void;
  onOpenGuide: () => void;
  isStandalone: boolean;
  setIsStandalone: (val: boolean) => void;
}

export const HomeOverview: React.FC<HomeOverviewProps> = ({
  onNavigateTab,
  onOpenGuide,
  isStandalone,
  setIsStandalone,
}) => {
  const [activeComparisonTab, setActiveComparisonTab] = useState<'all' | 'company' | 'team'>('all');

  const getStepIcon = (iconName: string) => {
    switch (iconName) {
      case 'FileCode2': return <FileCode2 className="w-4 h-4 text-slate-400" />;
      case 'ListCheck': return <ListCheck className="w-4 h-4 text-slate-400" />;
      case 'ShieldCheck': return <ShieldCheck className="w-4 h-4 text-slate-400" />;
      case 'Building2': return <Building2 className="w-4 h-4 text-slate-400" />;
      default: return <FileCode2 className="w-4 h-4 text-slate-400" />;
    }
  };

  const getQuickCardIcon = (index: number) => {
    switch (index) {
      case 0: return <FolderTree className="w-5 h-5 text-indigo-600" />;
      case 1: return <Terminal className="w-5 h-5 text-indigo-600" />;
      case 2: return <ShieldCheck className="w-5 h-5 text-indigo-600" />;
      case 3: return <Award className="w-5 h-5 text-indigo-600" />;
      default: return <FolderTree className="w-5 h-5 text-indigo-600" />;
    }
  };

  return (
    <div className="space-y-8">
      
      {/* Hero Banner Section */}
      <div className="bg-slate-900 text-white rounded-2xl p-6 sm:p-8 md:p-9 border border-slate-800 shadow-sm relative">
        <div className="max-w-4xl space-y-4">
          <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-md bg-slate-800 border border-slate-700 text-indigo-300 font-mono text-[11px] font-bold uppercase tracking-wider">
            <span>{HOME_HERO_CONTENT.badge}</span>
          </div>

          <h1 className="text-2xl sm:text-3xl md:text-4xl font-bold tracking-tight text-white leading-tight">
            {HOME_HERO_CONTENT.titlePrefix}
            <span className="text-indigo-300"> {HOME_HERO_CONTENT.companyTitle} </span>
            {HOME_HERO_CONTENT.titleConnector}
            <span className="text-cyan-300"> {HOME_HERO_CONTENT.teamTitle}</span>
          </h1>

          <p className="text-slate-300 text-sm sm:text-base leading-relaxed max-w-3xl">
            {HOME_HERO_CONTENT.description}
          </p>

          <div className="pt-2 flex flex-wrap gap-3 items-center">
            <Tooltip title="Why click this?" content="View the full visual tree map of files, folders, and repository layouts." position="bottom">
              <button
                onClick={() => onNavigateTab('architecture')}
                className="px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs sm:text-sm transition-all flex items-center gap-2 shadow-xs"
              >
                <FolderTree className="w-4 h-4" />
                <span>Explore Architecture</span>
                <ArrowRight className="w-4 h-4 ml-0.5" />
              </button>
            </Tooltip>

            <Tooltip title="Why click this?" content="Test interactive terminal commands like company-os validate in a live CLI simulator." position="bottom">
              <button
                onClick={() => onNavigateTab('cli')}
                className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-100 font-semibold text-xs sm:text-sm border border-slate-700 transition-all flex items-center gap-2"
              >
                <Terminal className="w-4 h-4 text-cyan-400" />
                <span>Try CLI Simulator</span>
              </button>
            </Tooltip>

            <Tooltip title="Why click this?" content="Open plain English explanations for technical terms without confusing jargon." position="bottom">
              <button
                onClick={onOpenGuide}
                className="px-4 py-2.5 rounded-xl bg-slate-800/80 hover:bg-slate-800 text-slate-200 font-semibold text-xs sm:text-sm border border-slate-700 transition-all flex items-center gap-2"
              >
                <HelpCircle className="w-4 h-4 text-indigo-400" />
                <span>Beginner Glossary</span>
              </button>
            </Tooltip>
          </div>
        </div>
      </div>

      {/* Core Explanation Cards: Company OS vs Team OS */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-2 border-b border-slate-200">
          <div>
            <h2 className="text-lg font-bold text-slate-900 tracking-tight flex items-center gap-2">
              <Layers className="w-5 h-5 text-indigo-600" />
              <span>Understanding the Dual-Core System</span>
            </h2>
            <p className="text-xs text-slate-500">
              Why operating rules are separated into central company standards and local team autonomy.
            </p>
          </div>

          <div className="bg-slate-100 p-1 rounded-xl border border-slate-200 flex items-center text-xs self-start sm:self-auto">
            <button
              onClick={() => setActiveComparisonTab('all')}
              className={`px-3 py-1 rounded-lg font-semibold transition-all ${
                activeComparisonTab === 'all' ? 'bg-white text-slate-900 shadow-xs' : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Side-by-Side
            </button>
            <button
              onClick={() => setActiveComparisonTab('company')}
              className={`px-3 py-1 rounded-lg font-semibold transition-all ${
                activeComparisonTab === 'company' ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Company OS
            </button>
            <button
              onClick={() => setActiveComparisonTab('team')}
              className={`px-3 py-1 rounded-lg font-semibold transition-all ${
                activeComparisonTab === 'team' ? 'bg-cyan-600 text-white shadow-xs' : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Team OS
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {DUAL_CORE_COMPARISON.map((info) => {
            const isCompany = info.id === 'company';
            const isVisible = activeComparisonTab === 'all' || activeComparisonTab === info.id;
            if (!isVisible) return null;

            return (
              <div 
                key={info.id}
                className={`bg-white border rounded-2xl p-5 shadow-xs space-y-4 transition-all ${
                  activeComparisonTab === info.id 
                    ? (isCompany ? 'border-indigo-600 md:col-span-2' : 'border-cyan-600 md:col-span-2')
                    : 'border-slate-200 hover:border-slate-300'
                }`}
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-xl text-white flex items-center justify-center font-bold shrink-0 ${
                      isCompany ? 'bg-indigo-600' : 'bg-cyan-600'
                    }`}>
                      {isCompany ? <Building2 className="w-5 h-5" /> : <Users className="w-5 h-5" />}
                    </div>
                    <div>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-bold font-mono uppercase tracking-wider ${
                        isCompany ? 'bg-indigo-50 text-indigo-800 border border-indigo-100' : 'bg-cyan-50 text-cyan-800 border border-cyan-100'
                      }`}>
                        {info.badge}
                      </span>
                      <h3 className="text-lg font-bold text-slate-900 tracking-tight mt-0.5">
                        {info.title}
                      </h3>
                    </div>
                  </div>
                  <span className="text-xs font-mono font-semibold px-2.5 py-1 rounded-lg bg-slate-100 text-slate-700 border border-slate-200 shrink-0">
                    {info.yamlFile}
                  </span>
                </div>

                <p className="text-slate-600 text-xs leading-relaxed">
                  {info.summary}
                </p>

                <div className="space-y-2 pt-2 border-t border-slate-100">
                  <span className="text-[11px] font-bold text-slate-900 uppercase font-mono tracking-wider block">Key Focus Areas:</span>
                  <ul className="space-y-1.5 text-xs text-slate-700">
                    {info.keyPoints.map((point, idx) => (
                      <li key={idx} className="flex items-start gap-2">
                        <CheckCircle2 className={`w-4 h-4 shrink-0 mt-0.5 ${isCompany ? 'text-indigo-600' : 'text-cyan-600'}`} />
                        <span>{point}</span>
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="pt-2 flex items-center justify-between text-xs">
                  <span className="text-slate-500 italic">{info.footerNote}</span>
                  <button
                    onClick={() => onNavigateTab(info.targetTab)}
                    className={`px-3 py-1.5 rounded-xl font-bold transition-colors flex items-center gap-1 ${
                      isCompany ? 'bg-indigo-50 hover:bg-indigo-100 text-indigo-700' : 'bg-cyan-50 hover:bg-cyan-100 text-cyan-800'
                    }`}
                  >
                    <span>{info.actionText}</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Interactive Mode Demonstration Widget */}
      <div className="bg-slate-900 text-white rounded-2xl p-6 border border-slate-800 space-y-5">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <RefreshCw className="w-4 h-4 text-indigo-400" />
              <span className="text-xs font-bold text-indigo-300 font-mono uppercase tracking-wider">
                Live Operating Mode Preview
              </span>
            </div>
            <h3 className="text-lg font-bold text-white tracking-tight">
              Switching Operating Modes: Standalone vs. Federated
            </h3>
            <p className="text-slate-400 text-xs">
              Toggle mode to experience how validation rules & workspace requirements adapt.
            </p>
          </div>

          <div className="bg-slate-800 p-1 rounded-xl border border-slate-700 flex items-center shrink-0">
            <button
              onClick={() => setIsStandalone(false)}
              className={`px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-2 ${
                !isStandalone
                  ? 'bg-indigo-600 text-white'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Layers className="w-3.5 h-3.5" />
              <span>Company OS (Federated)</span>
            </button>
            <button
              onClick={() => setIsStandalone(true)}
              className={`px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-2 ${
                isStandalone
                  ? 'bg-cyan-600 text-white'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <GitBranch className="w-3.5 h-3.5" />
              <span>Team OS (Standalone)</span>
            </button>
          </div>
        </div>

        {/* Dynamic Mode Feedback Box */}
        <div className={`p-4 rounded-xl border text-xs leading-relaxed ${
          !isStandalone 
            ? 'bg-indigo-950/40 border-indigo-800/80 text-indigo-100' 
            : 'bg-cyan-950/40 border-cyan-800/80 text-cyan-100'
        }`}>
          <div className="flex items-start gap-3">
            <div className={`p-2 rounded-lg shrink-0 ${
              !isStandalone ? 'bg-indigo-600/30 text-indigo-300' : 'bg-cyan-600/30 text-cyan-300'
            }`}>
              {!isStandalone ? <Building2 className="w-4 h-4" /> : <GitBranch className="w-4 h-4" />}
            </div>
            <div className="space-y-1">
              <h4 className="font-bold text-sm text-white">
                Active Mode: {!isStandalone ? 'Company OS (Federated Mode)' : 'Team OS (Standalone Mode)'}
              </h4>
              <p className="text-slate-300 text-xs">
                {!isStandalone ? (
                  <>In <strong>Federated Mode</strong>, workspace validation enforces root <code>company-os.yaml</code>, cross-team platform dependencies, global compliance rules, and root ontology definitions. Gate 8 (Federation) is strictly enforced.</>
                ) : (
                  <>In <strong>Standalone Mode</strong>, the team repository operates independently using <code>team-os.yaml</code>. Central company dependencies are treated as optional for local squad velocity.</>
                )}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* How It Works Flow: 4 Steps */}
      <div className="space-y-4">
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight flex items-center gap-2">
            <Zap className="w-5 h-5 text-indigo-600" />
            <span>The 4-Step Feature Lifecycle</span>
          </h2>
          <p className="text-xs text-slate-500">
            How changes move from local ideation to verified release.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {LIFECYCLE_STEPS.map((step) => (
            <div key={step.stepNumber} className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs space-y-2">
              <div className="flex items-center justify-between">
                <span className="w-6 h-6 rounded bg-slate-100 text-slate-800 font-bold text-xs flex items-center justify-center font-mono">
                  {step.stepNumber}
                </span>
                {getStepIcon(step.iconName)}
              </div>
              <h4 className="font-bold text-slate-900 text-sm">{step.title}</h4>
              <p className="text-slate-600 text-xs leading-relaxed">{step.description}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Feature Navigation Grid (Direct Jump to Tools) */}
      <div className="space-y-4">
        <div>
          <h2 className="text-lg font-bold text-slate-900 tracking-tight flex items-center gap-2">
            <Compass className="w-5 h-5 text-indigo-600" />
            <span>Interactive Explorer Directory</span>
          </h2>
          <p className="text-xs text-slate-500">
            Jump directly to any interactive simulator or reference tool.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {QUICK_DIRECTORY_CARDS.map((card, idx) => (
            <div
              key={idx}
              onClick={() => onNavigateTab(card.targetTab)}
              className="p-4 bg-white rounded-xl border border-slate-200 hover:border-indigo-300 transition-all cursor-pointer group shadow-xs space-y-2"
            >
              <div className="w-8 h-8 rounded-lg bg-indigo-50 text-indigo-700 flex items-center justify-center">
                {getQuickCardIcon(idx)}
              </div>
              <h3 className="font-bold text-slate-900 text-sm group-hover:text-indigo-700 transition-colors">
                {card.title}
              </h3>
              <p className="text-slate-600 text-xs">
                {card.description}
              </p>
              <div className="pt-1 text-xs font-bold text-indigo-600 flex items-center gap-1">
                <span>{card.actionText}</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </div>
            </div>
          ))}
        </div>
      </div>

    </div>
  );
};

