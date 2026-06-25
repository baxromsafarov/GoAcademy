import { useEffect, useRef } from "react"

interface YTPlayer {
  getCurrentTime(): number
  getDuration(): number
  seekTo(seconds: number, allowSeekAhead: boolean): void
  destroy(): void
}

interface YTPlayerOptions {
  videoId: string
  width?: string
  height?: string
  events?: {
    onReady?: (e: { target: YTPlayer }) => void
    onStateChange?: (e: { data: number; target: YTPlayer }) => void
  }
}

interface YTNamespace {
  Player: new (el: HTMLElement, opts: YTPlayerOptions) => YTPlayer
  PlayerState: { PLAYING: number; ENDED: number }
}

declare global {
  interface Window {
    YT?: YTNamespace
    onYouTubeIframeAPIReady?: () => void
  }
}

let apiPromise: Promise<void> | null = null

function loadApi(): Promise<void> {
  if (window.YT?.Player) return Promise.resolve()
  if (apiPromise) return apiPromise
  apiPromise = new Promise((resolve) => {
    const previous = window.onYouTubeIframeAPIReady
    window.onYouTubeIframeAPIReady = () => {
      previous?.()
      resolve()
    }
    const tag = document.createElement("script")
    tag.src = "https://www.youtube.com/iframe_api"
    document.head.appendChild(tag)
  })
  return apiPromise
}

/** YouTubePlayer embeds a YouTube video and reports watch progress (~every 10s
 * while playing) as a percent and a position in seconds. */
export function YouTubePlayer({
  videoId,
  startSeconds,
  onProgress,
}: {
  videoId: string
  startSeconds?: number
  onProgress: (percent: number, position: number) => void
}) {
  const containerRef = useRef<HTMLDivElement>(null)
  const onProgressRef = useRef(onProgress)
  onProgressRef.current = onProgress

  useEffect(() => {
    let cancelled = false
    let player: YTPlayer | null = null
    let interval: number | undefined

    void loadApi().then(() => {
      if (cancelled || !containerRef.current || !window.YT) return
      player = new window.YT.Player(containerRef.current, {
        videoId,
        width: "100%",
        height: "100%",
        events: {
          onReady: (e) => {
            if (startSeconds && startSeconds > 0) e.target.seekTo(startSeconds, true)
          },
          onStateChange: (e) => {
            const YT = window.YT
            if (!YT) return
            if (e.data === YT.PlayerState.PLAYING) {
              interval = window.setInterval(() => {
                if (!player) return
                const dur = player.getDuration()
                const cur = player.getCurrentTime()
                if (dur > 0) onProgressRef.current(Math.round((cur / dur) * 100), Math.round(cur))
              }, 10000)
            } else if (interval) {
              window.clearInterval(interval)
              interval = undefined
            }
          },
        },
      })
    })

    return () => {
      cancelled = true
      if (interval) window.clearInterval(interval)
      player?.destroy()
    }
  }, [videoId, startSeconds])

  return (
    <div className="aspect-video w-full overflow-hidden rounded-md bg-muted">
      <div ref={containerRef} className="size-full" />
    </div>
  )
}
