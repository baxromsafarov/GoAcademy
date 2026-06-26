import { useMemo, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, Check, ListChecks, X } from "lucide-react"
import { useQuiz, useSubmitQuiz } from "@/lib/queries"
import type { QuizQuestion, QuizQuestionReview } from "@/lib/types"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function QuizDetail() {
  const { t } = useTranslation()
  const { id = "" } = useParams()
  const quiz = useQuiz(id)
  const submit = useSubmitQuiz(id)
  const [answers, setAnswers] = useState<Record<string, string[]>>({})

  const result = submit.data
  const reviewByQ = useMemo(() => {
    const m = new Map<string, QuizQuestionReview>()
    result?.review.forEach((r) => m.set(r.question_id, r))
    return m
  }, [result])

  function selectSingle(qid: string, oid: string) {
    setAnswers((a) => ({ ...a, [qid]: [oid] }))
  }
  function toggleMultiple(qid: string, oid: string) {
    setAnswers((a) => {
      const cur = a[qid] ?? []
      return { ...a, [qid]: cur.includes(oid) ? cur.filter((x) => x !== oid) : [...cur, oid] }
    })
  }
  function retry() {
    submit.reset()
    setAnswers({})
  }

  const questions = quiz.data?.questions ?? []
  const answeredCount = questions.filter((q) => (answers[q.id]?.length ?? 0) > 0).length
  const allAnswered = questions.length > 0 && answeredCount === questions.length

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <Link
        to="/quizzes"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {quiz.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {quiz.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {quiz.data && (
        <div className="flex flex-col gap-5">
          <div>
            <div className="mb-2 flex items-center gap-1.5 text-xs font-medium tracking-wide text-violet-500 uppercase">
              <ListChecks className="size-3.5" /> {t("nav.quizzes")}
            </div>
            <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{quiz.data.title}</h1>
            {quiz.data.description && (
              <p className="mt-2 text-muted-foreground">{quiz.data.description}</p>
            )}
            <div className="mt-3 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
              <span className="rounded-md border px-1.5 py-0.5">{t(`difficulty.${quiz.data.difficulty}`)}</span>
              <span className="rounded-md border px-1.5 py-0.5">{quiz.data.language.toUpperCase()}</span>
              <span className="rounded-md border px-1.5 py-0.5">
                {t("quizzes.passThreshold", { pct: quiz.data.pass_threshold })}
              </span>
            </div>
          </div>

          {result && (
            <div
              className={cn(
                "flex items-center justify-between gap-4 rounded-lg border p-4",
                result.passed
                  ? "border-green-500/40 bg-green-500/10"
                  : "border-amber-500/40 bg-amber-500/10",
              )}
            >
              <div>
                <div className="text-lg font-semibold">
                  {result.passed ? t("quizzes.passed") : t("quizzes.failed")}
                </div>
                <div className="text-sm text-muted-foreground">
                  {t("quizzes.score", { score: result.score })}
                </div>
              </div>
              <Button variant="outline" onClick={retry}>
                {t("quizzes.retry")}
              </Button>
            </div>
          )}

          <ol className="flex flex-col gap-5">
            {questions.map((q, idx) => (
              <QuestionBlock
                key={q.id}
                index={idx + 1}
                question={q}
                selected={answers[q.id] ?? []}
                review={reviewByQ.get(q.id)}
                onSingle={selectSingle}
                onMultiple={toggleMultiple}
              />
            ))}
          </ol>

          {!result && (
            <div className="flex items-center gap-4">
              <Button onClick={() => submit.mutate(answers)} disabled={!allAnswered || submit.isPending}>
                {submit.isPending ? t("quizzes.submitting") : t("quizzes.submit")}
              </Button>
              <span className="text-sm text-muted-foreground">
                {t("quizzes.progress", { answered: answeredCount, total: questions.length })}
              </span>
              {submit.isError && <span className="text-sm text-red-500">{t("common.error")}</span>}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function QuestionBlock({
  index,
  question,
  selected,
  review,
  onSingle,
  onMultiple,
}: {
  index: number
  question: QuizQuestion
  selected: string[]
  review?: QuizQuestionReview
  onSingle: (qid: string, oid: string) => void
  onMultiple: (qid: string, oid: string) => void
}) {
  const { t } = useTranslation()
  const reviewed = review !== undefined

  return (
    <li className="rounded-lg border bg-card p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="font-medium">
          {index}. {question.prompt}
        </div>
        {reviewed &&
          (review.correct ? (
            <span className="flex shrink-0 items-center gap-1 text-sm text-green-600 dark:text-green-400">
              <Check className="size-4" /> {t("quizzes.correct")}
            </span>
          ) : (
            <span className="flex shrink-0 items-center gap-1 text-sm text-red-600 dark:text-red-400">
              <X className="size-4" /> {t("quizzes.incorrect")}
            </span>
          ))}
      </div>
      <div className="mt-1 text-xs text-muted-foreground">
        {question.type === "multiple" ? t("quizzes.multiple") : t("quizzes.single")}
      </div>

      <div className="mt-3 flex flex-col gap-2">
        {question.options.map((o) => {
          const isSelected = selected.includes(o.id)
          const isCorrect = review?.correct_option_ids.includes(o.id) ?? false
          return (
            <label
              key={o.id}
              className={cn(
                "flex items-center gap-2 rounded-md border p-2 text-sm",
                !reviewed && "cursor-pointer hover:bg-accent",
                reviewed && isCorrect && "border-green-500/50 bg-green-500/10",
                reviewed && isSelected && !isCorrect && "border-red-500/50 bg-red-500/10",
              )}
            >
              <input
                type={question.type === "single" ? "radio" : "checkbox"}
                name={question.id}
                checked={isSelected}
                disabled={reviewed}
                onChange={() =>
                  question.type === "single" ? onSingle(question.id, o.id) : onMultiple(question.id, o.id)
                }
                className="size-4 accent-primary"
              />
              <span className="flex-1">{o.text}</span>
              {reviewed && isCorrect && (
                <Check className="size-4 text-green-600 dark:text-green-400" />
              )}
              {reviewed && isSelected && !isCorrect && (
                <X className="size-4 text-red-600 dark:text-red-400" />
              )}
            </label>
          )
        })}
      </div>
    </li>
  )
}
