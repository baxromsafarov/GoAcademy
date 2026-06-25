import { useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { useVideos, type VideoFilters } from "@/lib/queries"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

export function VideoList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<VideoFilters>({})
  const { data, isPending, isError } = useVideos(filters)

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
          value={filters.language ?? ""}
          onChange={(e) => set("language", e.target.value)}
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
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-28 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("videos.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((v) => (
              <Link
                key={v.id}
                to={`/videos/${v.id}`}
                className="rounded-lg border bg-card p-4 transition-colors hover:border-primary"
              >
                <div className="font-medium">{v.title}</div>
                <div className="mt-1 line-clamp-2 text-sm text-muted-foreground">{v.description}</div>
                <div className="mt-3 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                  <span className="rounded border px-1.5 py-0.5">{t(`difficulty.${v.difficulty}`)}</span>
                  <span className="rounded border px-1.5 py-0.5">{v.language.toUpperCase()}</span>
                  {v.tags.map((tag) => (
                    <span key={tag} className="rounded border px-1.5 py-0.5">
                      #{tag}
                    </span>
                  ))}
                </div>
              </Link>
            ))}
          </div>
        ))}
    </div>
  )
}
