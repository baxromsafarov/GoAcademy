import { useState } from "react"
import { useTranslation } from "react-i18next"
import { ListChecks } from "lucide-react"
import { useQuizzes, type QuizFilters } from "@/lib/queries"
import { useContentLanguage } from "@/lib/useContentLanguage"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { ContentCard, Meta } from "@/components/ContentCard"
import { Select } from "@/components/ui/select"

export function QuizList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<QuizFilters>({})
  const [language, setLanguage] = useContentLanguage()
  const { data, isPending, isError } = useQuizzes({ ...filters, language: language || undefined })

  function set(key: keyof QuizFilters, value: string) {
    setFilters((f) => ({ ...f, [key]: value || undefined }))
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.quizzes")}</h1>

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
            <div key={i} className="h-64 animate-pulse rounded-xl border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("quizzes.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((q) => (
              <ContentCard
                key={q.id}
                to={`/quizzes/${q.id}`}
                title={q.title}
                description={q.description}
                Icon={ListChecks}
                accentClass="bg-gradient-to-br from-violet-500/25 via-violet-500/10 to-transparent"
                badges={
                  <>
                    <Meta>{t(`difficulty.${q.difficulty}`)}</Meta>
                    <Meta>{q.language.toUpperCase()}</Meta>
                    <Meta>{t("quizzes.passThreshold", { pct: q.pass_threshold })}</Meta>
                  </>
                }
              />
            ))}
          </div>
        ))}
    </div>
  )
}
