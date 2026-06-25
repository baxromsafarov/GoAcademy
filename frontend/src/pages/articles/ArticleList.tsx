import { useState } from "react"
import { useTranslation } from "react-i18next"
import { FileText } from "lucide-react"
import { useArticles, type ArticleFilters } from "@/lib/queries"
import { useContentLanguage } from "@/lib/useContentLanguage"
import { ContentCard, Meta } from "@/components/ContentCard"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

export function ArticleList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<ArticleFilters>({})
  const [language, setLanguage] = useContentLanguage()
  const { data, isPending, isError } = useArticles({ ...filters, language: language || undefined })

  function set(key: keyof ArticleFilters, value: string) {
    setFilters((f) => ({ ...f, [key]: value || undefined }))
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.articles")}</h1>

      <div className="flex flex-wrap gap-2">
        <select
          value={filters.difficulty ?? ""}
          onChange={(e) => set("difficulty", e.target.value)}
          className={selectClass}
          aria-label={t("videos.filterDifficulty")}
        >
          <option value="">
            {t("videos.filterDifficulty")}: {t("common.all")}
          </option>
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
          <option value="">
            {t("videos.filterLanguage")}: {t("common.all")}
          </option>
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
            <div key={i} className="h-56 animate-pulse rounded-xl border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("articles.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((a) => (
              <ContentCard
                key={a.id}
                to={`/articles/${a.slug}`}
                title={a.title}
                Icon={FileText}
                accentClass="bg-gradient-to-br from-sky-500/25 via-sky-500/10 to-transparent"
                badges={
                  <>
                    <Meta>{t(`difficulty.${a.difficulty}`)}</Meta>
                    <Meta>{a.language.toUpperCase()}</Meta>
                    {a.tags.slice(0, 2).map((tag) => (
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
