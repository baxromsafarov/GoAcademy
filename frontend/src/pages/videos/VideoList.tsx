import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useVideos, type VideoFilters } from "@/lib/queries"
import { useContentLanguage } from "@/lib/useContentLanguage"
import { ContentCard, Meta } from "@/components/ContentCard"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

function fmtDuration(s: number): string {
  return `${Math.floor(s / 60)}:${(s % 60).toString().padStart(2, "0")}`
}

export function VideoList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<VideoFilters>({})
  const [language, setLanguage] = useContentLanguage()
  const { data, isPending, isError } = useVideos({ ...filters, language: language || undefined })

  function set(key: keyof VideoFilters, value: string) {
    setFilters((f) => ({ ...f, [key]: value || undefined }))
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.videos")}</h1>

      <div className="flex flex-wrap gap-2">
        <select
          value={filters.difficulty ?? ""}
          onChange={(e) => set("difficulty", e.target.value)}
          className={selectClass}
          aria-label={t("videos.filterDifficulty")}
        >
          <option value="">{t("videos.filterDifficulty")}: {t("common.all")}</option>
          {difficulties.map((d) => (
            <option key={d} value={d}>
              {t(`difficulty.${d}`)}
            </option>
          ))}
        </select>
        <select
          value={language}
          onChange={(e) => setLanguage(e.target.value)}
          className={selectClass}
          aria-label={t("videos.filterLanguage")}
        >
          <option value="">{t("videos.filterLanguage")}: {t("common.all")}</option>
          {langs.map((l) => (
            <option key={l} value={l}>
              {l.toUpperCase()}
            </option>
          ))}
        </select>
      </div>

      {isPending && (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-64 animate-pulse rounded-xl border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("videos.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((v) => (
              <ContentCard
                key={v.id}
                to={`/videos/${v.id}`}
                title={v.title}
                description={v.description}
                thumbnail={`https://img.youtube.com/vi/${v.youtube_id}/hqdefault.jpg`}
                mediaBadge={v.duration_seconds > 0 ? fmtDuration(v.duration_seconds) : undefined}
                badges={
                  <>
                    <Meta>{t(`difficulty.${v.difficulty}`)}</Meta>
                    <Meta>{v.language.toUpperCase()}</Meta>
                    {v.tags.slice(0, 2).map((tag) => (
                      <Meta key={tag}>#{tag}</Meta>
                    ))}
                  </>
                }
              />
            ))}
          </div>
        ))}
    </div>
  )
}
