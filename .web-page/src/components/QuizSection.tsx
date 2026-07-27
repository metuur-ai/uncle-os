import React, { useState } from 'react';
import { QUIZ_QUESTIONS_DATA } from '../data/quizData';
import { Award, CheckCircle2, XCircle, RefreshCw, Sparkles, HelpCircle } from 'lucide-react';

export const QuizSection: React.FC = () => {
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [selectedAnswers, setSelectedAnswers] = useState<Record<number, number>>({});
  const [showResults, setShowResults] = useState(false);

  const currentQuestion = QUIZ_QUESTIONS_DATA[currentQuestionIndex];
  const isAnswered = selectedAnswers[currentQuestionIndex] !== undefined;
  const isCorrect = selectedAnswers[currentQuestionIndex] === currentQuestion.correctIndex;

  const handleSelectOption = (index: number) => {
    if (showResults) return;
    setSelectedAnswers(prev => ({ ...prev, [currentQuestionIndex]: index }));
  };

  const handleNext = () => {
    if (currentQuestionIndex < QUIZ_QUESTIONS_DATA.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1);
    } else {
      setShowResults(true);
    }
  };

  const handleRestart = () => {
    setCurrentQuestionIndex(0);
    setSelectedAnswers({});
    setShowResults(false);
  };

  const calculateScore = () => {
    let score = 0;
    QUIZ_QUESTIONS_DATA.forEach((q, idx) => {
      if (selectedAnswers[idx] === q.correctIndex) {
        score += 1;
      }
    });
    return score;
  };

  const score = calculateScore();
  const percentage = Math.round((score / QUIZ_QUESTIONS_DATA.length) * 100);

  return (
    <div className="space-y-6">
      
      {/* Intro Header */}
      <div className="bg-gradient-to-br from-indigo-50 via-white to-slate-50 p-6 rounded-2xl border border-indigo-100 shadow-sm space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-bold text-indigo-700 uppercase tracking-widest font-mono">
                08 MASTERY CHECK
              </span>
              <span className="px-2 py-0.5 bg-indigo-100/70 rounded-full text-[10px] font-semibold text-indigo-800 border border-indigo-200">
                KNOWLEDGE ASSESSMENT
              </span>
            </div>
            <h2 className="text-xl font-bold text-slate-900 tracking-tight">Company OS Knowledge Assessment</h2>
          </div>

          <button
            onClick={handleRestart}
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-white hover:bg-slate-100 text-xs text-slate-700 font-semibold border border-slate-200 shadow-sm self-start md:self-auto transition-colors"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Restart Quiz</span>
          </button>
        </div>

        {/* Why What How Quick Guide */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-xs border-t border-indigo-100/80">
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHY IS THIS HERE?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              To help interns, managers, and engineers test their understanding of Company OS principles in 5 minutes.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">WHAT AM I LOOKING AT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              A 10-question multiple-choice interactive quiz with instant feedback and explanations.
            </p>
          </div>
          <div className="bg-white/80 p-3 rounded-xl border border-slate-200/80 space-y-1">
            <span className="font-bold text-indigo-900 block font-mono uppercase text-[10px]">HOW DO I USE IT?</span>
            <p className="text-slate-600 text-[11px] leading-relaxed">
              Select an answer below to get immediate explanation feedback, then click <strong>Next Question</strong>!
            </p>
          </div>
        </div>
      </div>

      {!showResults ? (
        <div className="max-w-3xl mx-auto bg-white border border-slate-200 shadow-sm rounded-2xl p-6 space-y-5">
          
          {/* Question Progress Header */}
          <div className="flex items-center justify-between border-b border-slate-100 pb-3">
            <span className="text-xs font-mono font-bold text-indigo-700 uppercase tracking-wider">
              Question {currentQuestionIndex + 1} of {QUIZ_QUESTIONS_DATA.length}
            </span>
            <span className="text-xs px-2.5 py-0.5 rounded-full bg-slate-100 text-slate-700 border border-slate-200 font-medium">
              {currentQuestion.category}
            </span>
          </div>

          {/* Question Text */}
          <h3 className="text-base font-bold text-slate-900 leading-snug">{currentQuestion.question}</h3>

          {/* Options */}
          <div className="space-y-2.5 pt-2">
            {currentQuestion.options.map((option, idx) => {
              const isSelected = selectedAnswers[currentQuestionIndex] === idx;
              let btnStyle = 'bg-slate-50 border-slate-200 text-slate-700 hover:bg-slate-100';

              if (isAnswered) {
                if (idx === currentQuestion.correctIndex) {
                  btnStyle = 'bg-emerald-50 border-emerald-300 text-emerald-900 font-bold';
                } else if (isSelected) {
                  btnStyle = 'bg-rose-50 border-rose-300 text-rose-900';
                }
              } else if (isSelected) {
                btnStyle = 'bg-indigo-50 border-indigo-300 text-slate-900 font-bold';
              }

              return (
                <button
                  key={idx}
                  onClick={() => handleSelectOption(idx)}
                  className={`w-full text-left p-3.5 rounded-xl border text-xs transition-all flex items-center justify-between gap-3 ${btnStyle}`}
                >
                  <span className="leading-relaxed">{option}</span>
                  {isAnswered && idx === currentQuestion.correctIndex && (
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  )}
                  {isAnswered && isSelected && idx !== currentQuestion.correctIndex && (
                    <XCircle className="w-4 h-4 text-rose-600 shrink-0" />
                  )}
                </button>
              );
            })}
          </div>

          {/* Instant Explanation Feedback */}
          {isAnswered && (
            <div className={`p-4 rounded-xl text-xs space-y-1.5 ${
              isCorrect ? 'bg-emerald-50 border border-emerald-200 text-emerald-900' : 'bg-amber-50 border border-amber-200 text-amber-900'
            }`}>
              <div className="font-bold flex items-center gap-1.5">
                {isCorrect ? <CheckCircle2 className="w-4 h-4 text-emerald-600" /> : <XCircle className="w-4 h-4 text-amber-600" />}
                <span>{isCorrect ? 'Correct!' : 'Explanation:'}</span>
              </div>
              <p className="leading-relaxed">{currentQuestion.explanation}</p>
            </div>
          )}

          {/* Next Button */}
          <div className="flex justify-end pt-2">
            <button
              onClick={handleNext}
              disabled={!isAnswered}
              className={`px-5 py-2.5 rounded-xl text-xs font-bold transition-all ${
                isAnswered
                  ? 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/20'
                  : 'bg-slate-100 text-slate-400 border border-slate-200 cursor-not-allowed'
              }`}
            >
              {currentQuestionIndex < QUIZ_QUESTIONS_DATA.length - 1 ? 'Next Question' : 'View Final Score'}
            </button>
          </div>

        </div>
      ) : (
        /* Final Score Certificate */
        <div className="max-w-2xl mx-auto bg-white border border-slate-200 shadow-sm rounded-2xl p-8 text-center space-y-6">
          <div className="w-16 h-16 bg-gradient-to-tr from-indigo-600 to-cyan-500 rounded-2xl flex items-center justify-center mx-auto shadow-xl shadow-indigo-500/20">
            <Award className="w-8 h-8 text-white" />
          </div>

          <div className="space-y-2">
            <h3 className="text-xl font-bold text-slate-900">Quiz Completed!</h3>
            <p className="text-xs text-slate-600">
              You scored <span className="text-indigo-700 font-bold text-sm">{score}</span> out of {QUIZ_QUESTIONS_DATA.length} ({percentage}%)
            </p>
          </div>

          <div className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-2 text-xs">
            {percentage >= 80 ? (
              <div className="text-emerald-800 font-bold flex items-center justify-center gap-2">
                <Sparkles className="w-4 h-4 text-emerald-600" />
                <span>Mastery Achieved! You are ready to run Company OS and Team OS in production.</span>
              </div>
            ) : (
              <div className="text-amber-800 font-bold">
                Good effort! Review the Architecture, CLI Terminal, and Validation Gates tabs to polish your skills.
              </div>
            )}
          </div>

          <button
            onClick={handleRestart}
            className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-indigo-600/20"
          >
            Try Quiz Again
          </button>
        </div>
      )}

    </div>
  );
};
