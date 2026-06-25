import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import {
  ArrowLeft,
  CheckCircle2,
  ChevronRight,
  Circle,
  Code2,
  FileText,
  FolderKanban,
  HelpCircle,
  Video,
} from "lucide-react"
import { useTrack, useTrackProgress } from "@/lib/queries"
import type { TrackContentType } from "@/lib/types"

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
  progress.data?.items.forEach((it) => completedByKey.set(`${it.content_type}:${it.content_id}`, it.completed))

  return (
    <div className="flex flex-col gap-4">
      <Link
        to="/tracks"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {track.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {track.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {track.data && (
        <div className="flex flex-col gap-5">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">{track.data.title}</h1>
              {track.data.description && (
                <p className="mt-1 text-muted-foreground">{track.data.description}</p>
              )}
              <div className="mt-2 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                <span className="rounded border px-1.5 py-0.5">{t(`difficulty.${track.data.level}`)}</span>
                <span className="rounded border px-1.5 py-0.5">{track.data.language.toUpperCase()}</span>
              </div>
            </div>
            {progress.data?.track_complete && (
              <span className="flex shrink-0 items-center gap-1 text-sm text-green-600 dark:text-green-400">
                <CheckCircle2 className="size-5" /> {t("tracks.complete")}
              </span>
            )}
          </div>

          {progress.data && (
            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">{t("tracks.progress")}</span>
                <span className="font-medium">
                  {t("tracks.completedOf", {
                    completed: progress.data.completed,
                    total: progress.data.total,
                  })}
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-primary transition-all"
                  style={{ width: `${progress.data.percent}%` }}
                />
              </div>
            </div>
          )}

          <section className="flex flex-col gap-2">
            <h2 className="text-lg font-semibold">{t("tracks.program")}</h2>
            <ol className="flex flex-col gap-2">
              {track.data.items.map((item) => {
                const Icon = typeIcon[item.content_type]
                const done = completedByKey.get(`${item.content_type}:${item.content_id}`) ?? false
                return (
                  <li key={`${item.content_type}:${item.content_id}`}>
                    <Link
                      to={itemPath(item.content_type, item.content_id)}
                      className="flex items-center gap-3 rounded-lg border bg-card p-3 transition-colors hover:border-primary"
                    >
                      {done ? (
                        <CheckCircle2 className="size-5 shrink-0 text-green-600 dark:text-green-400" />
                      ) : (
                        <Circle className="size-5 shrink-0 text-muted-foreground" />
                      )}
                      <Icon className="size-4 shrink-0 text-muted-foreground" />
                      <span className="flex-1 text-sm font-medium">
                        {item.position}. {t(`tracks.type.${item.content_type}`)}
                      </span>
                      <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
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
