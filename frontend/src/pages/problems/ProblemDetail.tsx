import { lazy, Suspense, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, Check, CheckCircle2, Info, X } from "lucide-react"
import { useProblem, useProblemSolution, useSubmitProblem } from "@/lib/queries"
import type { JudgeVerdict, ProblemSample } from "@/lib/types"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

// Markdown carries react-markdown + highlight.js — load it as its own chunk.
const Markdown = lazy(() => import("@/components/Markdown").then((m) => ({ default: m.Markdown })))

const languages = ["go", "python", "javascript", "rust", "java"]

export function ProblemDetail() {
  const { t } = useTranslation()
  const { slug = "" } = useParams()
  const problem = useProblem(slug)
  const solution = useProblemSolution(slug)
  const submit = useSubmitProblem(slug)

  const [code, setCode] = useState("")
  const [language, setLanguage] = useState("go")
  const [markSolved, setMarkSolved] = useState(false)

  const referenceSolution = solution.data?.reference_solution_markdown ?? null
  const solved = referenceSolution !== null

  return (
    <div className="flex flex-col gap-4">
      <Link
        to="/problems"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {problem.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {problem.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {problem.data && (
        <div className="flex flex-col gap-5">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">{problem.data.title}</h1>
              <div className="mt-2 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                <span className="rounded border px-1.5 py-0.5">
                  {t(`difficulty.${problem.data.difficulty}`)}
                </span>
                <span className="rounded border px-1.5 py-0.5">
                  {problem.data.language.toUpperCase()}
                </span>
                {problem.data.tags.map((tag) => (
                  <span key={tag} className="rounded border px-1.5 py-0.5">
                    #{tag}
                  </span>
                ))}
              </div>
            </div>
            {solved && (
              <span className="flex shrink-0 items-center gap-1 text-sm text-green-600 dark:text-green-400">
                <CheckCircle2 className="size-5" /> {t("problems.solved")}
              </span>
            )}
          </div>

          <Suspense fallback={<div className="h-32 w-full animate-pulse rounded-md bg-muted" />}>
            <Markdown>{problem.data.statement_markdown}</Markdown>
          </Suspense>

          {problem.data.sample_io.length > 0 && (
            <section className="flex flex-col gap-3">
              <h2 className="text-lg font-semibold">{t("problems.examples")}</h2>
              {problem.data.sample_io.map((ex, i) => (
                <SampleExample key={i} index={i + 1} sample={ex} />
              ))}
            </section>
          )}

          <section className="flex flex-col gap-3">
            <h2 className="text-lg font-semibold">{t("problems.yourSolution")}</h2>
            <div className="flex flex-wrap items-center gap-2">
              <label className="text-sm text-muted-foreground" htmlFor="lang">
                {t("problems.language")}
              </label>
              <select
                id="lang"
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
                className="h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {languages.map((l) => (
                  <option key={l} value={l}>
                    {l}
                  </option>
                ))}
              </select>
            </div>
            <textarea
              value={code}
              onChange={(e) => setCode(e.target.value)}
              spellCheck={false}
              rows={12}
              placeholder={t("problems.codePlaceholder")}
              className="w-full resize-y rounded-md border bg-card p-3 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />

            <label className="flex w-fit cursor-pointer items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={markSolved}
                onChange={(e) => setMarkSolved(e.target.checked)}
                className="size-4 accent-primary"
              />
              {t("problems.markSolved")}
            </label>

            <div className="flex flex-wrap items-center gap-3">
              <Button
                onClick={() => submit.mutate({ code, language, solved: markSolved })}
                disabled={code.trim().length === 0 || submit.isPending}
              >
                {submit.isPending ? t("problems.submitting") : t("problems.submit")}
              </Button>
              {submit.isSuccess &&
                !submit.data.verdict &&
                (submit.data.status === "solved" ? (
                  <span className="text-sm text-green-600 dark:text-green-400">
                    {t("problems.submittedSolved")}
                  </span>
                ) : (
                  <span className="text-sm text-muted-foreground">{t("problems.submittedSaved")}</span>
                ))}
              {submit.isError && <span className="text-sm text-red-500">{t("common.error")}</span>}
            </div>

            {submit.data?.verdict && <VerdictPanel verdict={submit.data.verdict} />}

            {!submit.data?.verdict && (
              <p className="flex items-start gap-1.5 text-xs text-muted-foreground">
                <Info className="mt-0.5 size-3.5 shrink-0" />
                {t("problems.judgeNote")}
              </p>
            )}
          </section>

          {solved && referenceSolution && (
            <section className="flex flex-col gap-3 border-t pt-4">
              <h2 className="text-lg font-semibold">{t("problems.referenceSolution")}</h2>
              <Suspense fallback={<div className="h-24 w-full animate-pulse rounded-md bg-muted" />}>
                <Markdown>{referenceSolution}</Markdown>
              </Suspense>
            </section>
          )}
        </div>
      )}
    </div>
  )
}

function VerdictPanel({ verdict }: { verdict: JudgeVerdict }) {
  const { t } = useTranslation()
  const ok = verdict.verdict === "OK"
  return (
    <div
      className={cn(
        "flex flex-col gap-3 rounded-lg border p-4",
        ok ? "border-green-500/40 bg-green-500/10" : "border-amber-500/40 bg-amber-500/10",
      )}
    >
      <div className="flex items-center justify-between gap-4">
        <span className="text-lg font-semibold">{t(`problems.verdict.${verdict.verdict}`)}</span>
        <span className="text-sm text-muted-foreground">
          {t("problems.passedOf", { passed: verdict.passed, total: verdict.total })}
        </span>
      </div>

      {verdict.compile_error && (
        <pre className="overflow-x-auto rounded bg-muted p-2 font-mono text-xs text-red-600 dark:text-red-400">
          {verdict.compile_error}
        </pre>
      )}

      {verdict.cases.length > 0 && (
        <ul className="flex flex-col gap-1">
          {verdict.cases.map((c) => {
            const caseOk = c.verdict === "OK"
            return (
              <li key={c.index} className="flex items-center gap-2 text-sm">
                {caseOk ? (
                  <Check className="size-4 text-green-600 dark:text-green-400" />
                ) : (
                  <X className="size-4 text-red-600 dark:text-red-400" />
                )}
                <span className="font-medium">
                  {t("problems.case")} {c.index + 1}
                </span>
                <span className="rounded border px-1.5 py-0.5 text-xs text-muted-foreground">
                  {c.is_sample ? t("problems.sample") : t("problems.hidden")}
                </span>
                <span className="text-xs text-muted-foreground">{t(`problems.verdict.${c.verdict}`)}</span>
                <span className="ml-auto text-xs text-muted-foreground">{c.duration_ms} ms</span>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

function SampleExample({ index, sample }: { index: number; sample: ProblemSample }) {
  const { t } = useTranslation()
  const hasIO = sample.input !== undefined || sample.output !== undefined
  return (
    <div className="rounded-md border bg-card p-3">
      <div className="mb-2 text-sm font-medium">
        {t("problems.example")} {index}
      </div>
      {hasIO ? (
        <div className="grid gap-3 sm:grid-cols-2">
          <div>
            <div className="mb-1 text-xs text-muted-foreground">{t("problems.input")}</div>
            <pre className="overflow-x-auto rounded bg-muted p-2 font-mono text-xs">
              {String(sample.input ?? "")}
            </pre>
          </div>
          <div>
            <div className="mb-1 text-xs text-muted-foreground">{t("problems.output")}</div>
            <pre className="overflow-x-auto rounded bg-muted p-2 font-mono text-xs">
              {String(sample.output ?? "")}
            </pre>
          </div>
        </div>
      ) : (
        <pre className="overflow-x-auto rounded bg-muted p-2 font-mono text-xs">
          {JSON.stringify(sample, null, 2)}
        </pre>
      )}
    </div>
  )
}
