import { lazy, Suspense } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, CheckCircle2, Circle, FolderKanban } from "lucide-react"
import { useProject, useProjectProgress, useToggleProjectStep } from "@/lib/queries"

const Markdown = lazy(() => import("@/components/Markdown").then((m) => ({ default: m.Markdown })))

export function ProjectDetail() {
  const { t } = useTranslation()
  const { id = "" } = useParams()
  const project = useProject(id)
  const progress = useProjectProgress(id)
  const toggle = useToggleProjectStep(id)

  const doneSet = new Set(progress.data?.completed_step_ids ?? [])
  const total = progress.data?.total ?? project.data?.steps.length ?? 0
  const completed = progress.data?.completed ?? 0
  const percent = total > 0 ? Math.round((completed / total) * 100) : 0

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <Link
        to="/projects"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {project.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {project.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {project.data && (
        <div className="flex flex-col gap-5">
          <div className="flex items-start justify-between gap-4">
            <div>
              <div className="mb-2 flex items-center gap-1.5 text-xs font-medium tracking-wide text-rose-500 uppercase">
                <FolderKanban className="size-3.5" /> {t("nav.projects")}
              </div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{project.data.title}</h1>
              <div className="mt-3 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                <span className="rounded-md border px-1.5 py-0.5">
                  {t(`difficulty.${project.data.difficulty}`)}
                </span>
                <span className="rounded-md border px-1.5 py-0.5">
                  {project.data.language.toUpperCase()}
                </span>
                {project.data.tags.map((tag) => (
                  <span key={tag} className="rounded-md border px-1.5 py-0.5">
                    #{tag}
                  </span>
                ))}
              </div>
            </div>
            {progress.data?.project_complete && (
              <span className="flex shrink-0 items-center gap-1 text-sm text-green-600 dark:text-green-400">
                <CheckCircle2 className="size-5" /> {t("projects.complete")}
              </span>
            )}
          </div>

          <Suspense fallback={<div className="h-24 w-full animate-pulse rounded-md bg-muted" />}>
            <Markdown>{project.data.description_markdown}</Markdown>
          </Suspense>

          <section className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">{t("projects.checklist")}</h2>
              <span className="text-sm text-muted-foreground">
                {t("projects.completedOf", { completed, total })}
              </span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${percent}%` }}
              />
            </div>

            <ul className="flex flex-col gap-2">
              {project.data.steps.map((step) => {
                const done = doneSet.has(step.id)
                return (
                  <li key={step.id}>
                    <button
                      type="button"
                      onClick={() => toggle.mutate(step.id)}
                      disabled={toggle.isPending || progress.isPending}
                      className="flex w-full items-center gap-3 rounded-lg border bg-card p-3 text-left transition-colors hover:border-primary disabled:opacity-60"
                    >
                      {done ? (
                        <CheckCircle2 className="size-5 shrink-0 text-green-600 dark:text-green-400" />
                      ) : (
                        <Circle className="size-5 shrink-0 text-muted-foreground" />
                      )}
                      <span className={`flex-1 text-sm ${done ? "text-muted-foreground line-through" : ""}`}>
                        {step.text}
                      </span>
                    </button>
                  </li>
                )
              })}
            </ul>
          </section>
        </div>
      )}
    </div>
  )
}
