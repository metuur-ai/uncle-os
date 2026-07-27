import React, { useState } from 'react';
import { WORKFLOW_SCENARIOS_DATA } from '../data/workflowsData';
import { WorkflowScenario, WorkflowStep } from '../types';
import { PlayCircle, CheckCircle2, AlertCircle, ArrowRight, RefreshCw, Terminal, Check, ShieldAlert, Sparkles } from 'lucide-react';

export const WorkflowPlayground: React.FC = () => {
  const [selectedScenario, setSelectedScenario] = useState<WorkflowScenario>(WORKFLOW_SCENARIOS_DATA[1]); // Default to Scenario 2 (Change Lifecycle)
  const [currentStepIndex, setCurrentStepIndex] = useState<number>(0);
  const [completedSteps, setCompletedSteps] = useState<Record<number, boolean>>({});
  const [simulationOutput, setSimulationOutput] = useState<string | null>(null);

  // Form states for Scenario 2
  const [realityDateUpdated, setRealityDateUpdated] = useState<boolean>(false);
  const [checklistChecked, setChecklistChecked] = useState<boolean>(false);

  const currentStep: WorkflowStep = selectedScenario.steps[currentStepIndex];

  const handleSelectScenario = (scenario: WorkflowScenario) => {
    setSelectedScenario(scenario);
    setCurrentStepIndex(0);
    setCompletedSteps({});
    setSimulationOutput(null);
    setRealityDateUpdated(false);
    setChecklistChecked(false);
  };

  const handleRunStep = () => {
    // Check if step 3 of Scenario 2 (Change Lifecycle) and realityDateUpdated is false
    if (selectedScenario.id === 'change-lifecycle' && currentStepIndex === 2 && !realityDateUpdated) {
      setSimulationOutput(
        `done-check failed — a change is not done until reality is updated:\n  [FAIL] reality doc for 'online-ordering-app' not updated since PRD created\n\nEXIT CODE: 5 (Precondition Failed)`
      );
      return;
    }

    // Success simulation
    setSimulationOutput(currentStep.mockTerminalOutput);
    setCompletedSteps(prev => ({ ...prev, [currentStepIndex]: true }));
  };

  const handleNextStep = () => {
    if (currentStepIndex < selectedScenario.steps.length - 1) {
      setCurrentStepIndex(currentStepIndex + 1);
      setSimulationOutput(null);
    }
  };

  const handleResetScenario = () => {
    setCurrentStepIndex(0);
    setCompletedSteps({});
    setSimulationOutput(null);
    setRealityDateUpdated(false);
    setChecklistChecked(false);
  };

  return (
    <div className="space-y-6">
      
      {/* Header Banner */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                03 WORKFLOWS
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                LIFECYCLE SIMULATOR
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">Interactive Workflow Simulator</h2>
          </div>

          <button
            onClick={handleResetScenario}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-white hover:bg-slate-100 text-xs text-slate-700 font-semibold border border-slate-200 shadow-sm self-start md:self-auto transition-colors"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Reset Workflow</span>
          </button>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              To practice real-world team scenarios (like creating a PRD or updating docs) before doing it in real life.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              5 scenario cards above, and an interactive step-by-step simulator below.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Pick a scenario card, then click <strong>Execute Step</strong> to advance through each milestone!
            </p>
          </div>
        </div>
      </div>

      {/* Scenario Selector Tabs */}
      <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-5 gap-3">
        {WORKFLOW_SCENARIOS_DATA.map((scenario) => {
          const isSelected = selectedScenario.id === scenario.id;
          return (
            <button
              key={scenario.id}
              onClick={() => handleSelectScenario(scenario)}
              className={`p-3 rounded-2xl text-left border transition-all flex flex-col justify-between ${
                isSelected
                  ? 'bg-indigo-50 border-indigo-300 text-slate-900 shadow-sm font-semibold'
                  : 'bg-white border-slate-200 hover:bg-slate-100/70 text-slate-700'
              }`}
            >
              <div>
                <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold uppercase tracking-wider ${
                  isSelected ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-600'
                }`}>
                  {scenario.badge}
                </span>
                <h3 className="text-xs font-bold mt-2 text-slate-900 line-clamp-2">{scenario.title}</h3>
              </div>
              <p className="text-[10px] text-slate-500 mt-2 line-clamp-2">{scenario.subtitle}</p>
            </button>
          );
        })}
      </div>

      {/* Main Stepper & Playground Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">

        {/* Left: Stepper Navigation */}
        <div className="lg:col-span-4 bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
          <div className="border-b border-slate-100 pb-3">
            <h3 className="text-sm font-bold text-slate-900">{selectedScenario.title}</h3>
            <p className="text-xs text-slate-500 mt-1">{selectedScenario.description}</p>
          </div>

          <div className="space-y-2">
            {selectedScenario.steps.map((step, idx) => {
              const isCurrent = currentStepIndex === idx;
              const isCompleted = completedSteps[idx];

              return (
                <div
                  key={step.stepNumber}
                  onClick={() => {
                    setCurrentStepIndex(idx);
                    setSimulationOutput(null);
                  }}
                  className={`p-3 rounded-xl border cursor-pointer transition-all flex items-start gap-3 ${
                    isCurrent
                      ? 'bg-indigo-50 border-indigo-300 text-slate-900 shadow-sm'
                      : isCompleted
                      ? 'bg-emerald-50 border-emerald-200 text-slate-800'
                      : 'bg-slate-50/80 border-slate-200 text-slate-600 hover:bg-slate-100'
                  }`}
                >
                  <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold shrink-0 mt-0.5 ${
                    isCompleted
                      ? 'bg-emerald-600 text-white'
                      : isCurrent
                      ? 'bg-indigo-600 text-white'
                      : 'bg-slate-200 text-slate-600'
                  }`}>
                    {isCompleted ? <Check className="w-3.5 h-3.5" /> : step.stepNumber}
                  </div>

                  <div>
                    <h4 className="text-xs font-bold text-slate-800">{step.title}</h4>
                    <p className="text-[11px] text-slate-500 font-mono mt-0.5 line-clamp-1">{step.command}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right: Step Interactive Workspace & Terminal Simulator */}
        <div className="lg:col-span-8 bg-white border border-slate-200 shadow-sm rounded-2xl p-6 space-y-5">
          
          {/* Step Header */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
            <div>
              <span className="text-xs font-bold text-indigo-700 uppercase tracking-wider font-mono">
                Step {currentStep.stepNumber} of {selectedScenario.steps.length}
              </span>
              <h3 className="text-base font-bold text-slate-900 mt-0.5">{currentStep.title}</h3>
            </div>

            <button
              onClick={handleRunStep}
              className="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs flex items-center gap-2 shadow-md shadow-indigo-600/20 self-start sm:self-auto"
            >
              <Terminal className="w-4 h-4" />
              <span>Run Step Command</span>
            </button>
          </div>

          <p className="text-xs text-slate-600 leading-relaxed">{currentStep.description}</p>

          {/* Key Rule Callout */}
          <div className="p-3 rounded-xl bg-amber-50 border border-amber-200 flex items-start gap-2.5 text-xs text-amber-900">
            <ShieldAlert className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
            <div>
              <span className="font-bold text-amber-900 block">Governance Contract Rule:</span>
              <span>{currentStep.keyRule}</span>
            </div>
          </div>

          {/* Special Interactive Form controls for Scenario 2 (Change Lifecycle) Step 3 */}
          {selectedScenario.id === 'change-lifecycle' && currentStepIndex === 2 && (
            <div className="bg-slate-50 p-4 rounded-xl border border-slate-200 space-y-3 text-xs">
              <span className="font-bold text-indigo-800 block">Interactive Precondition Toggle</span>
              <p className="text-slate-700">
                Test how <code className="text-amber-900 bg-amber-100/70 px-1 py-0.5 rounded border border-amber-200">prd complete</code> enforces reality updates. Toggle the reality doc update below to see the difference between Exit Code 5 failure and successful archiving!
              </p>

              <div className="space-y-2 pt-1">
                <label className="flex items-center gap-2 cursor-pointer text-slate-800">
                  <input
                    type="checkbox"
                    checked={realityDateUpdated}
                    onChange={(e) => setRealityDateUpdated(e.target.checked)}
                    className="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500 bg-white"
                  />
                  <span>I updated <code className="text-slate-800 font-mono">platforms/ordering/reality/components/online-ordering-app.md</code> and bumped <code className="text-emerald-700 font-mono font-bold">updated: 2026-07-26</code></span>
                </label>
              </div>
            </div>
          )}

          {/* Command execution display */}
          <div className="bg-slate-900 p-3 rounded-xl border border-slate-800 font-mono text-xs text-indigo-300 shadow-inner">
            <span className="text-slate-500">$ </span>
            <span className="text-slate-100 font-bold">{currentStep.command}</span>
          </div>

          {/* Terminal Output */}
          {simulationOutput && (
            <div className="bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs space-y-2 shadow-md">
              <div className="text-slate-500 text-[11px] border-b border-slate-800/80 pb-1 flex justify-between">
                <span>Terminal Output</span>
                {simulationOutput.includes('FAIL') ? (
                  <span className="text-rose-400 font-bold">Execution Blocked [FAIL]</span>
                ) : (
                  <span className="text-emerald-400 font-bold">Execution Passed [PASS]</span>
                )}
              </div>
              <pre className={`whitespace-pre-wrap leading-relaxed ${
                simulationOutput.includes('FAIL') ? 'text-rose-300' : 'text-emerald-300'
              }`}>
                {simulationOutput}
              </pre>
            </div>
          )}

          {/* Stepper Next Navigation */}
          <div className="flex justify-end pt-2">
            {currentStepIndex < selectedScenario.steps.length - 1 ? (
              <button
                onClick={handleNextStep}
                disabled={!completedSteps[currentStepIndex]}
                className={`px-4 py-2 rounded-xl text-xs font-bold flex items-center gap-2 transition-all ${
                  completedSteps[currentStepIndex]
                    ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-600/20'
                    : 'bg-slate-100 text-slate-400 border border-slate-200 cursor-not-allowed'
                }`}
              >
                <span>Proceed to Step {currentStepIndex + 2}</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            ) : (
              <div className="p-3 bg-emerald-50 border border-emerald-200 rounded-xl text-emerald-800 text-xs flex items-center gap-2">
                <CheckCircle2 className="w-5 h-5 text-emerald-600" />
                <span className="font-bold">Workflow Complete! Scenario successfully completed.</span>
              </div>
            )}
          </div>

        </div>

      </div>
    </div>
  );
};
