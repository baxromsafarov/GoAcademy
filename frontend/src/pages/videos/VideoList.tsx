import { useTranslation } from "react-i18next"
import { useVideos } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { ContentCard, Meta } from "@/components/ContentCard"
import { Select } from "@/components/ui/select"
import { Pagination } from "@/components/Pagination"

function fmtDuration(s: number): string {
  return `${Math.floor(s / 60)}:${(s % 60).toString().padStart(2, "0")}`
}

export function VideoList() {
  const { t } = useTranslation()
  const { get, language, offset, pageSize, page, setParam, setLanguage, setPage } = useListParams()
  const { data, isPending, isError } = useVideos({
    difficulty: get("difficulty") || undefined,
    tag: get("tag") || undefined,
    language: language || undefined,
    limit: pageSize,
    offset,
  })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.videos")}</h1>

      <div className="flex flex-wrap gap-2">
        <Select
          value={get("difficulty")}
          onChange={(v) => setParam("difficulty", v)}
          options={difficultyOptions(t)}
          ariaLabel={t("videos.filterDifficulty")}
        />
        <Select
          value={language}
          onChange={setLanguage}
          options={languageOptions(t)}
          ariaLabel={t("videos.filterLanguage")}
        />
        <input
          value={get("tag")}
          onChange={(e) => setParam("tag", e.target.value)}
          placeholder={t("common.filterTag")}
          className="h-9 w-28 rounded-md border bg-transparent px-2.5 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
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
          <>
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
            <Pagination page={page} pageSize={pageSize} total={data.total} onPage={setPage} />
          </>
        ))}
    </div>
  )
}
