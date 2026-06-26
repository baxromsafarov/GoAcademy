import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Code2 } from "lucide-react"
import { useProblems, type ProblemFilters } from "@/lib/queries"
import { useContentLanguage } from "@/lib/useContentLanguage"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { ContentCard, Meta } from "@/components/ContentCard"
import { Select } from "@/components/ui/select"

export function ProblemList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<ProblemFilters>({})
  const [language, setLanguage] = useContentLanguage()
  const { data, isPending, isError } = useProblems({ ...filters, language: language || undefined })

  function set(key: keyof ProblemFilters, value: string) {
    setFilters((f) => ({ ...f, [key]: value || undefined }))
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.problems")}</h1>

      <div className="flex flex-wrap gap-2">
        <Select
          value={filters.difficulty ?? ""}
          onChange={(v) => set("difficulty", v)}
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
            <div key={i} className="h-56 animate-pulse rounded-xl border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("problems.empty")}</p>
        ) : (
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
        ))}
    </div>
  )
}
