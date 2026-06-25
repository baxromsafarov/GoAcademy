import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import {
  ArrowLeft,
  CheckCircle2,
  ChevronRight,
  Code2,
  FileText,
  FolderKanban,
  HelpCircle,
  Route,
  Video,
} from "lucide-react"
import { useTrack, useTrackProgress } from "@/lib/queries"
import type { TrackContentType } from "@/lib/types"
import { cn } from "@/lib/utils"

const typeIcon: Record<TrackContentType, typeof Video> = {
  video: Video,
  article: FileText,
  quiz: HelpCircle,
  problem: Code2,
  project: FolderKanban,
}

function itemPath(type: TrackContentType, id: string): string {
  switch (type) {
    case "video":
      return `/videos/${id}`
    case "article":
      return `/articles/${id}`
    case "quiz":
      return `/quizzes/${id}`
    case "problem":
      return `/problems/${id}`
    case "project":
      return `/projects/${id}`
  }
}

export function TrackDetail() {
  const { t } = useTranslation()
  const { id = "" } = useParams()
  const track = useTrack(id)
  const progress = useTrackProgress(id)

  const completedByKey = new Map<string, boolean>()
  progress.data?.items.forEach((it) =>
    completedByKey.set(`${it.content_type}:${it.content_id}`, it.completed),
  )

  // The next not-yet-done item is highlighted as the lesson to resume.
  const firstUndone = track.data?.items.find(
    (it) => !(completedByKey.get(`${it.content_type}:${it.content_id}`) ?? false),
  )

  return (
    <div className="flex flex-col gap-4">
      <Link
        to="/tracks"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {track.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {track.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {track.data && (
        <div className="flex flex-col gap-6">
          {/* Course header */}
          <div className="rounded-2xl border bg-gradient-to-br from-amber-500/10 via-card to-card p-6">
            <div className="mb-2 flex items-center gap-1.5 text-xs font-medium tracking-wide text-amber-500 uppercase">
              <Route className="size-3.5" /> {t("tracks.course")}
            </div>
            <div className="flex items-start justify-between gap-4">
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{track.data.title}</h1>
              {progress.data?.track_complete && (
                <span className="flex shrink-0 items-center gap-1 text-sm font-medium text-green-600 dark:text-green-400">
                  <CheckCircle2 className="size-5" /> {t("tracks.complete")}
                </span>
              )}
            </div>
            {track.data.description && (
              <p className="mt-2 max-w-2xl text-muted-foreground">{track.data.description}</p>
            )}
            <div className="mt-3 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
              <span className="rounded-md border px-1.5 py-0.5">
                {t(`difficulty.${track.data.level}`)}
              </span>
              <span className="rounded-md border px-1.5 py-0.5">
                {track.data.language.toUpperCase()}
              </span>
            </div>

            {/* Segmented sprint progress — one segment per lesson. */}
            {progress.data && (
              <div className="mt-5 flex flex-col gap-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">{t("tracks.progress")}</span>
                  <span className="font-medium">
                    {t("tracks.completedOf", {
                      completed: progress.data.completed,
                      total: progress.data.total,
                    })}
                  </span>
                </div>
                <div className="flex gap-1">
                  {track.data.items.map((item) => {
                    const done =
                      completedByKey.get(`${item.content_type}:${item.content_id}`) ?? false
                    return (
                      <div
                        key={`${item.content_type}:${item.content_id}`}
                        className={cn(
                          "h-1.5 flex-1 rounded-full transition-colors",
                          done ? "bg-primary" : "bg-muted",
                        )}
                      />
                    )
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Lesson path */}
          <section className="flex flex-col gap-3">
            <h2 className="text-lg font-semibold">{t("tracks.program")}</h2>
            <ol className="relative flex flex-col">
              {track.data.items.map((item, idx) => {
                const Icon = typeIcon[item.content_type]
                const key = `${item.content_type}:${item.content_id}`
                const done = completedByKey.get(key) ?? false
                const isNext = firstUndone === item
                const isLast = idx === track.data.items.length - 1
                return (
                  <li key={key} className="relative flex gap-4">
                    {/* Connector rail + node */}
                    <div className="flex flex-col items-center">
                      <div
                        className={cn(
                          "z-10 flex size-9 shrink-0 items-center justify-center rounded-full border-2 transition-colors",
                          done
                            ? "border-primary bg-primary text-primary-foreground"
                            : isNext
                              ? "border-primary bg-card text-primary"
                              : "border-muted bg-card text-muted-foreground",
                        )}
                      >
                        {done ? (
                          <CheckCircle2 className="size-5" />
                        ) : (
                          <span className="text-sm font-semibold">{item.position}</span>
                        )}
                      </div>
                      {!isLast && (
                        <div
                          className={cn("w-0.5 flex-1", done ? "bg-primary/40" : "bg-border")}
                        />
                      )}
                    </div>

                    {/* Lesson card */}
                    <Link
                      to={itemPath(item.content_type, item.content_id)}
                      className={cn(
                        "group mb-3 flex flex-1 items-center gap-3 rounded-xl border bg-card p-4 transition-all hover:-translate-y-0.5 hover:border-primary/60 hover:shadow-sm",
                        isNext && "border-primary/50 ring-1 ring-primary/20",
                      )}
                    >
                      <Icon className="size-5 shrink-0 text-muted-foreground transition-colors group-hover:text-primary" />
                      <div className="flex-1">
                        <div className="text-sm font-semibold">
                          {t("tracks.lesson", { n: item.position })}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {t(`tracks.type.${item.content_type}`)}
                        </div>
                      </div>
                      <ChevronRight className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
                    </Link>
                  </li>
                )
              })}
            </ol>
          </section>
        </div>
      )}
    </div>
  )
}
