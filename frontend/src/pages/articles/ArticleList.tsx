import { useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { useArticles, type ArticleFilters } from "@/lib/queries"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

export function ArticleList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<ArticleFilters>({})
  const { data, isPending, isError } = useArticles(filters)

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
          value={filters.language ?? ""}
          onChange={(e) => set("language", e.target.value)}
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
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-24 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("articles.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((a) => (
              <Link
                key={a.id}
                to={`/articles/${a.slug}`}
                className="rounded-lg border bg-card p-4 transition-colors hover:border-primary"
              >
                <div className="font-medium">{a.title}</div>
                <div className="mt-3 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                  <span className="rounded border px-1.5 py-0.5">{t(`difficulty.${a.difficulty}`)}</span>
                  <span className="rounded border px-1.5 py-0.5">{a.language.toUpperCase()}</span>
                  {a.tags.map((tag) => (
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
