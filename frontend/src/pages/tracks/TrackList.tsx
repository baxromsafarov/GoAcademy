import { useTranslation } from "react-i18next"
import { Route } from "lucide-react"
import { useTracks } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { ContentCard, Meta } from "@/components/ContentCard"
import { Select } from "@/components/ui/select"
import { Pagination } from "@/components/Pagination"
import { SearchBox } from "@/components/admin/AdminFilters"

export function TrackList() {
  const { t } = useTranslation()
  const { get, language, offset, pageSize, page, setParam, setLanguage, setPage } = useListParams()
  const { data, isPending, isError } = useTracks({
    q: get("q") || undefined,
    difficulty: get("difficulty") || undefined,
    language: language || undefined,
    limit: pageSize,
    offset,
  })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.tracks")}</h1>

      <div className="flex flex-wrap gap-2">
        <SearchBox value={get("q")} onChange={(v) => setParam("q", v)} />
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
          <p className="text-muted-foreground">{t("tracks.empty")}</p>
        ) : (
          <>
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
              {data.items.map((tr) => (
                <ContentCard
                  key={tr.id}
                  to={`/tracks/${tr.id}`}
                  title={tr.title}
                  description={tr.description}
                  Icon={Route}
                  accentClass="bg-gradient-to-br from-amber-500/25 via-amber-500/10 to-transparent"
                  badges={
                    <>
                      <Meta>{t(`difficulty.${tr.level}`)}</Meta>
                      <Meta>{tr.language.toUpperCase()}</Meta>
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
