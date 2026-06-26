import { useTranslation } from "react-i18next"
import { Code2 } from "lucide-react"
import { useProblems } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { ContentCard, Meta } from "@/components/ContentCard"
import { Select } from "@/components/ui/select"
import { Pagination } from "@/components/Pagination"

export function ProblemList() {
  const { t } = useTranslation()
  const { get, language, offset, pageSize, page, setParam, setLanguage, setPage } = useListParams()
  const { data, isPending, isError } = useProblems({
    difficulty: get("difficulty") || undefined,
    tag: get("tag") || undefined,
    language: language || undefined,
    limit: pageSize,
    offset,
  })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.problems")}</h1>

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
            <div key={i} className="h-56 animate-pulse rounded-xl border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("problems.empty")}</p>
        ) : (
          <>
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
              {data.items.map((p) => (
                <ContentCard
                  key={p.id}
                  to={`/problems/${p.slug}`}
                  title={p.title}
                  Icon={Code2}
                  accentClass="bg-gradient-to-br from-emerald-500/25 via-emerald-500/10 to-transparent"
                  badges={
                    <>
                      <Meta>{t(`difficulty.${p.difficulty}`)}</Meta>
                      <Meta>{p.language.toUpperCase()}</Meta>
                      {p.tags.slice(0, 2).map((tag) => (
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
