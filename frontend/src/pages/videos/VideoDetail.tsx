import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, CheckCircle2 } from "lucide-react"
import { usePostVideoProgress, useVideo, useVideoProgress } from "@/lib/queries"
import { YouTubePlayer } from "@/components/YouTubePlayer"
import { BookmarkButton, NoteComposer } from "@/components/ContentActions"
import { Button } from "@/components/ui/button"

export function VideoDetail() {
  const { t } = useTranslation()
  const { id = "" } = useParams()
  const video = useVideo(id)
  const progress = useVideoProgress(id)
  const post = usePostVideoProgress(id)

  return (
    <div className="flex flex-col gap-4">
      <Link
        to="/videos"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {video.isPending && <div className="aspect-video w-full animate-pulse rounded-md bg-muted" />}
      {video.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {video.data && !progress.isPending && (
        <>
          <YouTubePlayer
            videoId={video.data.youtube_id}
            startSeconds={progress.data?.last_position_seconds ?? 0}
            onProgress={(percent, position) => post.mutate({ percent, position })}
          />
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-xl font-semibold tracking-tight">{video.data.title}</h1>
              <p className="mt-1 text-muted-foreground">{video.data.description}</p>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <BookmarkButton contentType="video" contentId={video.data.id} />
              {progress.data?.completed ? (
                <span className="flex items-center gap-1 text-sm text-primary">
                  <CheckCircle2 className="size-5" /> {t("videos.watched")}
                </span>
              ) : (
                <Button
                  variant="outline"
                  onClick={() =>
                    post.mutate({
                      percent: progress.data?.watched_percent ?? 0,
                      position: progress.data?.last_position_seconds ?? 0,
                      completed: true,
                    })
                  }
                  disabled={post.isPending}
                >
                  {t("videos.markWatched")}
                </Button>
              )}
            </div>
          </div>

          <NoteComposer contentType="video" contentId={video.data.id} />
        </>
      )}
    </div>
  )
}
