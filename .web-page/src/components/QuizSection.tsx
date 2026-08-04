import React, { useState } from 'react';
import { Award, CheckCircle2, XCircle, RefreshCw, Sparkles } from 'lucide-react';
import { QUIZ_QUESTIONS_DATA } from '../data/quizData';
import { PageShell, PageHeader, Section, Card, Badge, Button, ProgressBar, cx } from './ui';

export const QuizSection: React.FC = () => {
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [selectedAnswers, setSelectedAnswers] = useState<Record<number, number>>({});
  const [showResults, setShowResults] = useState(false);

  const total = QUIZ_QUESTIONS_DATA.length;
  const currentQuestion = QUIZ_QUESTIONS_DATA[currentQuestionIndex];
  const isAnswered = selectedAnswers[currentQuestionIndex] !== undefined;
  const isCorrect = selectedAnswers[currentQuestionIndex] === currentQuestion.correctIndex;

  const handleSelectOption = (index: number) => {
    if (isAnswered) return;
    setSelectedAnswers((prev) => ({ ...prev, [currentQuestionIndex]: index }));
  };

  const handleNext = () => {
    if (currentQuestionIndex < total - 1) {
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

  const score = QUIZ_QUESTIONS_DATA.reduce(
    (acc, q, idx) => (selectedAnswers[idx] === q.correctIndex ? acc + 1 : acc),
    0
  );
  const answeredCount = Object.keys(selectedAnswers).length;
  const percentage = Math.round((score / total) * 100);

  return (
    <PageShell>
      <PageHeader
        eyebrow="Lesson 9 of 9"
        title="Company OS Knowledge Assessment"
        lead="A multiple-choice check of Company OS principles — core concepts, governance, the CLI contract, and federation. Answer each question to see immediate feedback with an explanation."
        icon={Award}
        actions={
          <Button variant="secondary" icon={RefreshCw} onClick={handleRestart}>
            Restart quiz
          </Button>
        }
      />

      {!showResults ? (
        <Section title={`Question ${currentQuestionIndex + 1} of ${total}`}>
          <Card padding="lg" className="mx-auto max-w-3xl space-y-5">
            <div className="space-y-2">
              <div className="flex items-center justify-between gap-3">
                <Badge tone="accent" mono>
                  {currentQuestion.category}
                </Badge>
                <span className="tabular font-mono text-xs text-fg-subtle">
                  Score: {score}/{answeredCount}
                </span>
              </div>
              <ProgressBar
                value={answeredCount}
                max={total}
                label="Progress"
                showValue={false}
              />
            </div>

            <h3 className="text-lg font-semibold leading-snug text-fg">{currentQuestion.question}</h3>

            <div role="radiogroup" aria-label="Answer options" className="space-y-2.5">
              {currentQuestion.options.map((option, idx) => {
                const isSelected = selectedAnswers[currentQuestionIndex] === idx;
                const isCorrectOption = idx === currentQuestion.correctIndex;

                let style = 'border-border bg-surface-sunken text-fg hover:border-border-strong';
                if (isAnswered && isCorrectOption) {
                  style = 'border-success-border bg-success-soft text-success-text';
                } else if (isAnswered && isSelected) {
                  style = 'border-danger-border bg-danger-soft text-danger-text';
                } else if (!isAnswered && isSelected) {
                  style = 'border-accent-border bg-accent-soft text-fg';
                }

                return (
                  <button
                    key={idx}
                    type="button"
                    role="radio"
                    aria-checked={isSelected}
                    disabled={isAnswered}
                    onClick={() => handleSelectOption(idx)}
                    className={cx(
                      'flex min-h-[3rem] w-full cursor-pointer items-center justify-between gap-3 rounded-xl border px-4 py-3 text-left text-sm font-medium transition-colors duration-150',
                      'disabled:cursor-not-allowed',
                      style
                    )}
                  >
                    <span className="leading-relaxed">{option}</span>
                    {isAnswered && isCorrectOption && (
                      <span className="flex shrink-0 items-center gap-1 text-xs font-bold">
                        <CheckCircle2 className="h-4 w-4" aria-hidden="true" />
                        Correct
                      </span>
                    )}
                    {isAnswered && isSelected && !isCorrectOption && (
                      <span className="flex shrink-0 items-center gap-1 text-xs font-bold">
                        <XCircle className="h-4 w-4" aria-hidden="true" />
                        Your answer
                      </span>
                    )}
                  </button>
                );
              })}
            </div>

            <div aria-live="polite">
              {isAnswered && (
                <div
                  className={cx(
                    'space-y-1.5 rounded-xl border p-4 text-sm',
                    isCorrect
                      ? 'border-success-border bg-success-soft text-success-text'
                      : 'border-warn-border bg-warn-soft text-warn-text'
                  )}
                >
                  <div className="flex items-center gap-1.5 font-bold">
                    {isCorrect ? (
                      <CheckCircle2 className="h-4 w-4" aria-hidden="true" />
                    ) : (
                      <XCircle className="h-4 w-4" aria-hidden="true" />
                    )}
                    <span>{isCorrect ? 'Correct' : 'Not quite'}</span>
                  </div>
                  <p className="leading-relaxed">{currentQuestion.explanation}</p>
                </div>
              )}
            </div>

            <div className="flex justify-end pt-2">
              <Button onClick={handleNext} disabled={!isAnswered}>
                {currentQuestionIndex < total - 1 ? 'Next question' : 'View final score'}
              </Button>
            </div>
          </Card>
        </Section>
      ) : (
        <Section title="Results">
          <Card padding="lg" className="mx-auto max-w-2xl space-y-6 text-center">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-accent shadow-md">
              <Award className="h-8 w-8 text-accent-fg" aria-hidden="true" />
            </div>

            <div className="space-y-2">
              <h3 className="text-xl font-bold text-fg">Quiz completed</h3>
              <p className="text-sm text-fg-muted">
                You scored <span className="tabular text-base font-bold text-accent-text">{score}</span> out of{' '}
                {total} ({percentage}%)
              </p>
            </div>

            <div
              className={cx(
                'rounded-xl border p-4 text-sm font-semibold',
                percentage >= 80
                  ? 'border-success-border bg-success-soft text-success-text'
                  : 'border-warn-border bg-warn-soft text-warn-text'
              )}
            >
              {percentage >= 80 ? (
                <span className="flex items-center justify-center gap-2">
                  <Sparkles className="h-4 w-4" aria-hidden="true" />
                  Mastery achieved! You are ready to run Company OS and Team OS in production.
                </span>
              ) : (
                <span>Good effort! Review the earlier lessons to polish your skills.</span>
              )}
            </div>

            <Button size="lg" icon={RefreshCw} onClick={handleRestart}>
              Try quiz again
            </Button>
          </Card>
        </Section>
      )}
    </PageShell>
  );
};
