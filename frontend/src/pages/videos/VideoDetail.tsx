import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, CheckCircle2, Clock, PlayCircle } from "lucide-react"
import { usePostVideoProgress, useVideo, useVideoProgress } from "@/lib/queries"
import { YouTubePlayer } from "@/components/YouTubePlayer"
import { BookmarkButton, NoteComposer } from "@/components/ContentActions"
import { Button } from "@/components/ui/button"

function fmtDuration(s: number): string {
  return `${Math.floor(s / 60)}:${(s % 60).toString().padStart(2, "0")}`
}

export function VideoDetail() {
  const { t } = useTranslation()
  const { id = "" } = useParams()
  const video = useVideo(id)
  const progress = useVideoProgress(id)
  const post = usePostVideoProgress(id)

  return (
    <div className="mx-auto w-full max-w-3xl">
      <Link
        to="/videos"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {video.isPending && (
        <div className="mt-4 aspect-video w-full animate-pulse rounded-xl bg-muted" />
      )}
      {video.isError && <p className="mt-4 text-sm text-red-500">{t("common.error")}</p>}

      {video.data && !progress.isPending && (
        <div className="mt-4 flex flex-col gap-5">
          <div className="overflow-hidden rounded-xl border shadow-sm shadow-black/10">
            <YouTubePlayer
              videoId={video.data.youtube_id}
              startSeconds={progress.data?.last_position_seconds ?? 0}
              onProgress={(percent, position) => post.mutate({ percent, position })}
            />
          </div>

          <div>
            <div className="mb-2 flex items-center gap-1.5 text-xs font-medium tracking-wide text-primary uppercase">
              <PlayCircle className="size-3.5" /> {t("nav.videos")}
            </div>
            <div className="flex items-start justify-between gap-4">
              <h1 className="text-2xl leading-tight font-bold tracking-tight">{video.data.title}</h1>
              <span className="shrink-0">
                <BookmarkButton contentType="video" contentId={video.data.id} />
              </span>
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              {video.data.duration_seconds > 0 && (
                <span className="flex items-center gap-1">
                  <Clock className="size-3.5" /> {fmtDuration(video.data.duration_seconds)}
                </span>
              )}
              <span className="rounded-md border px-1.5 py-0.5">
                {t(`difficulty.${video.data.difficulty}`)}
              </span>
              <span className="rounded-md border px-1.5 py-0.5">
                {video.data.language.toUpperCase()}
              </span>
              {video.data.tags.map((tag) => (
                <span key={tag} className="rounded-md border px-1.5 py-0.5">
                  #{tag}
                </span>
              ))}
            </div>
            {video.data.description && (
              <p className="mt-4 leading-relaxed text-muted-foreground">{video.data.description}</p>
            )}
          </div>

          <div className="flex items-center justify-center border-t pt-5">
            {progress.data?.completed ? (
              <span className="flex items-center gap-2 text-sm font-medium text-primary">
                <CheckCircle2 className="size-5" /> {t("videos.watched")}
              </span>
            ) : (
              <Button
                size="lg"
                className="w-full sm:w-auto"
                onClick={() =>
                  post.mutate({
                    percent: progress.data?.watched_percent ?? 0,
                    position: progress.data?.last_position_seconds ?? 0,
                    completed: true,
                  })
                }
                disabled={post.isPending}
              >
                <CheckCircle2 className="size-4" /> {t("videos.markWatched")}
              </Button>
            )}
          </div>

          <NoteComposer contentType="video" contentId={video.data.id} />
        </div>
      )}
    </div>
  )
}
