import React, { useState } from 'react';
import { Search, Code, Cpu, BookOpen, CheckCircle2, ArrowRight, Layers, Bot, Terminal, ShieldAlert, Sparkles, FileCode, CheckSquare, GitBranch, Info, MessageSquareCode, Copy, Check } from 'lucide-react';
import { Tooltip } from './Tooltip';
import { 
  SKILL_LAYERS, 
  STEP_TIERS, 
  REFERENCE_SKILLS, 
  AUTHORING_RULES, 
  MOCK_SKILLS_CLI_TEXT, 
  MOCK_SKILLS_JSON_TEXT,
  SAMPLE_PROMPTS,
  PROMPTING_RULES
} from '../data/skillsData';

export const LocalSearchAgentView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'search' | 'skills'>('skills');
  const [searchQuery, setSearchQuery] = useState('same-day pickup');
  const [searchScope, setSearchScope] = useState('moonbeam-os');
  const [cliFormat, setCliFormat] = useState<'text' | 'json'>('text');
  const [selectedReferenceSkill, setSelectedReferenceSkill] = useState<string>(REFERENCE_SKILLS[1].id);
  const [copiedPromptId, setCopiedPromptId] = useState<string | null>(null);

  const handleCopyPrompt = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedPromptId(id);
    setTimeout(() => setCopiedPromptId(null), 2000);
  };

  const activeRefSkill = REFERENCE_SKILLS.find(s => s.id === selectedReferenceSkill) || REFERENCE_SKILLS[1];

  return (
    <div className="space-y-6">
      
      {/* Header Banner */}
      <div className="bg-white p-5 sm:p-6 rounded-xl border border-slate-200 shadow-xs space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                06 AGENT SKILLS & LOCAL SEARCH
              </span>
              <span className="px-2 py-0.5 bg-indigo-50 rounded text-[10px] font-bold text-indigo-800 border border-indigo-100 font-mono">
                AUTHORING & READ-SIDE ARCHITECTURE
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">Agent Skills & Workspace Search Engine</h2>
          </div>

          <div className="flex items-center gap-2 bg-slate-100 p-1 rounded-xl border border-slate-200 shrink-0">
            <Tooltip title="Why click this?" content="Explore the 4 skill layers (Company, Platform, Team, Personal) that guide AI agents in authoring artifacts." position="bottom">
              <button
                onClick={() => setActiveTab('skills')}
                className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 ${
                  activeTab === 'skills' ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                <Bot className="w-4 h-4" />
                <span>Agent Skills (4 Layers)</span>
              </button>
            </Tooltip>

            <Tooltip title="Why click this?" content="Test the offline BM25 search engine for instant document indexing and keyword discovery." position="bottom">
              <button
                onClick={() => setActiveTab('search')}
                className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 ${
                  activeTab === 'search' ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                <Search className="w-4 h-4" />
                <span>Local Search Engine</span>
              </button>
            </Tooltip>
          </div>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-slate-100">
          <div className="bg-slate-50 p-3 rounded-lg border border-slate-200 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Skills teach AI agents and humans <strong>how to author</strong> artifacts, while Local Search enables instant <strong>offline discovery</strong>.
            </p>
          </div>
          <div className="bg-slate-50 p-3 rounded-lg border border-slate-200 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              A 4-layer skills resolution system (Company, Platform, Team, Personal) merged by <code>company-os skills list</code>.
            </p>
          </div>
          <div className="bg-slate-50 p-3 rounded-lg border border-slate-200 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Explore reference skills, review tier rules, or switch format to view the <code>--json</code> agent envelope!
            </p>
          </div>
        </div>
      </div>

      {activeTab === 'skills' ? (
        <div className="space-y-6">
          
          {/* Dual Cooperation Architecture Overview */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-white border border-indigo-100 rounded-2xl p-5 shadow-xs space-y-2.5">
              <div className="flex items-center gap-2 text-indigo-700">
                <Bot className="w-5 h-5" />
                <h3 className="font-bold text-slate-900 text-sm">Company OS Agent Skills (Authoring Side)</h3>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Versioned Markdown files (<code>*.SKILL.md</code>) that teach agents and humans <strong>how to author</strong> compliant discovery briefs, PRDs, and reality docs. Enforced during validation.
              </p>
              <div className="pt-2 text-[11px] text-indigo-900 font-mono bg-indigo-50/70 px-3 py-1.5 rounded-lg border border-indigo-100">
                $ company-os skills list
              </div>
            </div>

            <div className="bg-white border border-cyan-100 rounded-2xl p-5 shadow-xs space-y-2.5">
              <div className="flex items-center gap-2 text-cyan-700">
                <Search className="w-5 h-5" />
                <h3 className="font-bold text-slate-900 text-sm">Local Search Claude Skill (Read Side)</h3>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Installed via <code>local-search install-skill</code> into <code>~/.claude/skills/local-search</code>. Teaches agents <strong>how to find</strong> existing workspace docs offline via SQLite FTS5.
              </p>
              <div className="pt-2 text-[11px] text-cyan-900 font-mono bg-cyan-50/70 px-3 py-1.5 rounded-lg border border-cyan-100">
                $ local-search init --json
              </div>
            </div>
          </div>

          {/* Precedence Hierarchy Staircase */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Layers className="w-5 h-5 text-indigo-600" />
                <h3 className="text-base font-bold text-slate-900">4-Layer Skills Precedence Hierarchy</h3>
              </div>
              <span className="text-xs font-mono text-indigo-700 font-bold bg-indigo-50 px-2.5 py-1 rounded-lg border border-indigo-100">
                teams/&lt;t&gt;/team.yaml
              </span>
            </div>

            <p className="text-xs text-slate-600 leading-relaxed">
              When an AI agent operates on a workspace, rules and skills are resolved across 4 layers. Canonical mandatory steps always outrank personal preferences, ensuring company governance remains unbreakable:
            </p>

            <div className="bg-slate-900 p-3.5 rounded-xl border border-slate-800 font-mono text-xs text-indigo-300 space-y-1 shadow-inner">
              <div className="text-emerald-400"># Configured precedence contract in team.yaml</div>
              <div>precedence: canonical-mandatory &gt; personal &gt; canonical-default &gt; canonical-guidance</div>
              <div className="text-slate-400">onConflict: prefer-canonical-and-inform-user</div>
            </div>

            {/* Visual Precedence Staircase */}
            <div className="grid grid-cols-1 sm:grid-cols-4 gap-3 text-xs pt-1">
              <div className="bg-rose-50 border border-rose-200 p-3.5 rounded-xl space-y-1">
                <span className="text-rose-700 font-bold block text-[10px] uppercase font-mono">1. TOP AUTHORITY</span>
                <span className="text-slate-900 font-bold block text-sm">canonical-mandatory</span>
                <p className="text-slate-600 text-[11px] leading-tight">Company & platform security/compliance gates. Unbreakable.</p>
              </div>

              <div className="bg-cyan-50 border border-cyan-200 p-3.5 rounded-xl space-y-1">
                <span className="text-cyan-800 font-bold block text-[10px] uppercase font-mono">2. SECOND LEVEL</span>
                <span className="text-slate-900 font-bold block text-sm">personal scratchpad</span>
                <p className="text-slate-600 text-[11px] leading-tight">Personal workflow rules in <code>scratchpad/personal-rules/</code>.</p>
              </div>

              <div className="bg-amber-50 border border-amber-200 p-3.5 rounded-xl space-y-1">
                <span className="text-amber-800 font-bold block text-[10px] uppercase font-mono">3. THIRD LEVEL</span>
                <span className="text-slate-900 font-bold block text-sm">canonical-default</span>
                <p className="text-slate-600 text-[11px] leading-tight">Standard process unless a formal deviation is logged.</p>
              </div>

              <div className="bg-slate-100 border border-slate-200 p-3.5 rounded-xl space-y-1">
                <span className="text-slate-600 font-bold block text-[10px] uppercase font-mono">4. FOURTH LEVEL</span>
                <span className="text-slate-900 font-bold block text-sm">canonical-guidance</span>
                <p className="text-slate-600 text-[11px] leading-tight">Advisory best practices. Untracked and optional.</p>
              </div>
            </div>
          </div>

          {/* 4 Skill Layers Detail Cards */}
          <div className="space-y-3">
            <h3 className="text-sm font-bold text-slate-900">4 Skill Storage Layers</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {SKILL_LAYERS.map((layer) => (
                <div key={layer.layer} className="bg-white border border-slate-200 rounded-2xl p-4 shadow-xs space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-slate-900 text-sm">{layer.title}</span>
                    <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold font-mono uppercase ${
                      layer.authority === 'canonical' ? 'bg-indigo-100 text-indigo-800' :
                      layer.authority === 'team' ? 'bg-cyan-100 text-cyan-800' : 'bg-slate-100 text-slate-700'
                    }`}>
                      {layer.authority}
                    </span>
                  </div>
                  <div className="text-xs font-mono text-indigo-700 bg-slate-50 px-2.5 py-1 rounded-lg border border-slate-200 truncate">
                    {layer.location}
                  </div>
                  <p className="text-xs text-slate-600 leading-relaxed">
                    {layer.description}
                  </p>
                  <p className="text-[11px] text-slate-500 italic pt-1 border-t border-slate-100">
                    {layer.overrideRule}
                  </p>
                </div>
              ))}
            </div>
          </div>

          {/* Reference Skills Interactive Inspector */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h3 className="text-base font-bold text-slate-900">Reference Agent Skills Explorer</h3>
                <p className="text-xs text-slate-600">
                  Pre-packaged skills shipped with Company OS to guide agents through lifecycle milestones:
                </p>
              </div>
              <span className="text-xs font-mono text-slate-500 font-semibold bg-slate-100 px-2.5 py-1 rounded-lg">
                {REFERENCE_SKILLS.length} Reference Skills
              </span>
            </div>

            <div className="flex flex-wrap gap-2">
              {REFERENCE_SKILLS.map((skill) => (
                <button
                  key={skill.id}
                  onClick={() => setSelectedReferenceSkill(skill.id)}
                  className={`px-3 py-1.5 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5 ${
                    selectedReferenceSkill === skill.id
                      ? 'bg-indigo-600 text-white shadow-sm'
                      : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                  }`}
                >
                  <FileCode className="w-3.5 h-3.5" />
                  <span>{skill.title}</span>
                </button>
              ))}
            </div>

            {/* Selected Skill Viewer */}
            <div className="p-4 bg-slate-50 border border-slate-200 rounded-2xl space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-200 pb-2">
                <div>
                  <span className="text-[10px] font-mono text-indigo-700 font-bold uppercase">{activeRefSkill.id}</span>
                  <h4 className="font-bold text-slate-900 text-sm">{activeRefSkill.title}</h4>
                </div>
                <span className="text-[11px] font-mono text-slate-500 bg-white px-2 py-0.5 rounded border border-slate-200 self-start sm:self-center">
                  {activeRefSkill.location}
                </span>
              </div>

              <p className="text-xs text-slate-600 leading-relaxed">
                {activeRefSkill.summary}
              </p>

              <div className="space-y-2 pt-1">
                <span className="text-xs font-bold text-slate-900 uppercase font-mono tracking-wider">Step Execution Guidance:</span>
                <div className="space-y-2">
                  {activeRefSkill.steps.map((st) => (
                    <div key={st.number} className="flex items-start gap-2 bg-white p-2.5 rounded-xl border border-slate-200 text-xs">
                      <span className="w-5 h-5 rounded-full bg-slate-100 text-slate-800 font-bold flex items-center justify-center font-mono shrink-0 text-[11px]">
                        {st.number}
                      </span>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-bold font-mono uppercase shrink-0 ${
                        st.tier === 'mandatory' ? 'bg-rose-100 text-rose-800 border border-rose-200' :
                        st.tier === 'default' ? 'bg-indigo-100 text-indigo-800 border border-indigo-200' :
                        'bg-slate-100 text-slate-700 border border-slate-200'
                      }`}>
                        ({st.tier})
                      </span>
                      <span className="text-slate-800 font-mono leading-relaxed">{st.text}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Authoring Rules & Gate [7/N] Constraints */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
            <div className="flex items-center gap-2">
              <ShieldAlert className="w-5 h-5 text-indigo-600" />
              <h3 className="text-base font-bold text-slate-900">Authoring Rules & Gate [7/N] Constraints</h3>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {AUTHORING_RULES.map((r, idx) => (
                <div key={idx} className="p-3.5 bg-slate-50 border border-slate-200 rounded-xl space-y-1">
                  <div className="flex items-center gap-2 text-indigo-900 font-bold text-xs">
                    <CheckCircle2 className="w-3.5 h-3.5 text-indigo-600 shrink-0" />
                    <span>{r.rule}</span>
                  </div>
                  <p className="text-xs text-slate-600 leading-relaxed pl-5">
                    {r.detail}
                  </p>
                </div>
              ))}
            </div>
          </div>

          {/* Sample Prompts for AI Agents */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="flex items-center gap-2">
                  <MessageSquareCode className="w-5 h-5 text-indigo-600" />
                  <h3 className="text-base font-bold text-slate-900">Sample Prompts for AI Agents</h3>
                </div>
                <p className="text-xs text-slate-600 mt-1">
                  Ready-to-use structured prompts for instructing AI agents to follow Company OS skills with strict validation boundaries:
                </p>
              </div>
              <span className="text-xs font-mono text-indigo-700 font-bold bg-indigo-50 px-2.5 py-1 rounded-lg border border-indigo-100 self-start sm:self-center shrink-0">
                {SAMPLE_PROMPTS.length} Agent Templates
              </span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {SAMPLE_PROMPTS.map((prompt) => (
                <div key={prompt.id} className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-2.5 flex flex-col justify-between">
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between gap-2 flex-wrap">
                      <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded border border-indigo-100">
                        {prompt.category}
                      </span>
                      {prompt.skillName && (
                        <span className="text-[11px] font-mono font-bold text-indigo-950 bg-indigo-100/80 border border-indigo-200 px-2 py-0.5 rounded flex items-center gap-1">
                          <FileCode className="w-3 h-3 text-indigo-600" />
                          <span>{prompt.skillName}</span>
                        </span>
                      )}
                    </div>
                    <div>
                      <h4 className="font-bold text-slate-900 text-sm">{prompt.title}</h4>
                      {prompt.targetSkillId && (
                        <div className="mt-1 text-[10px] font-mono text-slate-500 bg-slate-100/90 px-2 py-0.5 rounded border border-slate-200/70 inline-block max-w-full overflow-x-auto whitespace-nowrap">
                          ID: <span className="text-slate-700 font-semibold">{prompt.targetSkillId}</span>
                        </div>
                      )}
                    </div>
                    <p className="text-xs text-slate-600">{prompt.description}</p>
                    <div className="p-3 bg-slate-900 rounded-lg text-xs font-mono text-slate-200 leading-relaxed border border-slate-800 shadow-inner overflow-x-auto max-h-36">
                      {prompt.promptText}
                    </div>
                  </div>

                  <button
                    onClick={() => handleCopyPrompt(prompt.id, prompt.promptText)}
                    className={`w-full py-1.5 px-3 rounded-lg text-xs font-semibold flex items-center justify-center gap-1.5 transition-all ${
                      copiedPromptId === prompt.id
                        ? 'bg-emerald-600 text-white'
                        : 'bg-white hover:bg-slate-100 text-slate-800 border border-slate-200 shadow-2xs'
                    }`}
                  >
                    {copiedPromptId === prompt.id ? (
                      <>
                        <Check className="w-3.5 h-3.5" />
                        <span>Prompt Copied!</span>
                      </>
                    ) : (
                      <>
                        <Copy className="w-3.5 h-3.5 text-indigo-600" />
                        <span>Copy Prompt for AI Agent</span>
                      </>
                    )}
                  </button>
                </div>
              ))}
            </div>

            {/* Prompting Best Practices */}
            <div className="pt-3 border-t border-slate-200 space-y-3">
              <div className="flex items-center gap-2 text-xs font-bold text-slate-900 uppercase font-mono tracking-wider">
                <Sparkles className="w-4 h-4 text-indigo-600" />
                <span>Prompting Best Practices (What Makes These Work):</span>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-xs">
                {PROMPTING_RULES.map((pr, idx) => (
                  <div key={idx} className="p-3 bg-indigo-50/50 border border-indigo-100/80 rounded-xl space-y-1">
                    <div className="font-bold text-indigo-950 text-xs">{pr.rule}</div>
                    <p className="text-slate-600 text-[11px] leading-relaxed">{pr.detail}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* CLI Terminal Output Simulator */}
          <div className="bg-slate-950 border border-slate-800 shadow-xl rounded-2xl p-5 space-y-3 font-mono text-xs">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-3">
              <div className="flex items-center gap-2">
                <Terminal className="w-4 h-4 text-emerald-400" />
                <span className="text-slate-200 font-bold">company-os skills list CLI Simulator</span>
              </div>

              <div className="flex items-center gap-2 bg-slate-900 p-1 rounded-lg border border-slate-800">
                <button
                  onClick={() => setCliFormat('text')}
                  className={`px-2.5 py-1 rounded text-[11px] font-semibold transition-all ${
                    cliFormat === 'text' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Text Format
                </button>
                <button
                  onClick={() => setCliFormat('json')}
                  className={`px-2.5 py-1 rounded text-[11px] font-semibold transition-all ${
                    cliFormat === 'json' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  --json Envelope
                </button>
              </div>
            </div>

            <div className="text-slate-400 text-[11px]">
              Command: <code className="text-emerald-300">company-os {cliFormat === 'json' ? '--json ' : ''}skills list</code>
            </div>

            <pre className="p-4 bg-slate-900 rounded-xl text-slate-200 overflow-x-auto text-[11px] leading-relaxed border border-slate-800/80 max-h-80">
              {cliFormat === 'text' ? MOCK_SKILLS_CLI_TEXT : MOCK_SKILLS_JSON_TEXT}
            </pre>
          </div>

        </div>
      ) : (
        <div className="space-y-6">
          
          {/* Architecture comparison: Authoring vs Read */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-3">
              <span className="text-xs font-bold text-indigo-700 uppercase tracking-wider font-mono">
                Authoring & Validation Side
              </span>
              <h3 className="text-base font-bold text-slate-900">company-os CLI</h3>
              <p className="text-xs text-slate-600 leading-relaxed">
                Scaffolds discovery briefs, PRDs, reality docs, resolves governance, builds graph tags, and runs validation gates. Never queries external servers or databases.
              </p>
              <div className="p-2.5 rounded-xl bg-slate-900 border border-slate-800 text-xs font-mono text-indigo-300 shadow-inner">
                $ company-os validate
              </div>
            </div>

            <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-3">
              <span className="text-xs font-bold text-cyan-700 uppercase tracking-wider font-mono">
                Read & Query Side
              </span>
              <h3 className="text-base font-bold text-slate-900">Local Search (SQLite FTS5)</h3>
              <p className="text-xs text-slate-600 leading-relaxed">
                Single Go binary that registers workspaces, scans Markdown files, and provides BM25 hybrid search for terminal users and AI agents.
              </p>
              <div className="p-2.5 rounded-xl bg-slate-900 border border-slate-800 text-xs font-mono text-cyan-300 shadow-inner">
                $ local-search search "same-day pickup"
              </div>
            </div>
          </div>

          {/* Interactive Search Simulator */}
          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-5 space-y-4">
            <h3 className="text-sm font-bold text-slate-900">Local Search Query Simulator</h3>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Search Query</label>
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
                />
              </div>

              <div>
                <label className="text-slate-700 block mb-1 font-semibold">Registered Scope</label>
                <input
                  type="text"
                  value={searchScope}
                  onChange={(e) => setSearchScope(e.target.value)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-slate-800 font-mono"
                />
              </div>

              <div className="flex items-end">
                <div className="w-full bg-slate-900 p-2 rounded-lg border border-slate-800 font-mono text-cyan-300 text-xs truncate shadow-inner">
                  $ local-search search "{searchQuery}"
                </div>
              </div>
            </div>

            <div className="bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs space-y-2 shadow-md">
              <div className="text-slate-500 text-[10px] border-b border-slate-800 pb-1">
                Local Search BM25 Ranked Results (2 matches)
              </div>
              <div className="text-emerald-300 space-y-2">
                <div>
                  <span className="text-slate-400">[1] score: 0.94 | repo: {searchScope}</span>
                  <div className="text-white font-bold">platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md</div>
                  <div className="text-slate-400 text-[11px]">...Allow web customers to choose 15-minute same-day pickup slots on current day...</div>
                </div>

                <div className="pt-2 border-t border-slate-800/80">
                  <span className="text-slate-400">[2] score: 0.81 | repo: {searchScope}</span>
                  <div className="text-white font-bold">teams/web/product/discovery/2026-same-day-pickup-slots/brief.md</div>
                  <div className="text-slate-400 text-[11px]">...Problem signal: Customers abandon carts when pickup is limited to next-day...</div>
                </div>
              </div>
            </div>
          </div>

        </div>
      )}

    </div>
  );
};

