import { useTranslation } from "react-i18next"
import { BookOpen, Search } from "lucide-react"
import { useCheatsheets } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { languageOptions } from "@/lib/filterOptions"
import { ContentCard, Meta } from "@/components/ContentCard"
import { Select } from "@/components/ui/select"
import { Pagination } from "@/components/Pagination"

export function CheatsheetList() {
  const { t } = useTranslation()
  const { get, language, offset, pageSize, page, setParam, setLanguage, setPage } = useListParams()
  const { data, isPending, isError } = useCheatsheets({
    q: get("q") || undefined,
    language: language || undefined,
    limit: pageSize,
    offset,
  })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.cheatsheets")}</h1>

      <div className="flex flex-wrap gap-2">
        <div className="relative">
          <Search className="absolute top-1/2 left-2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            value={get("q")}
            onChange={(e) => setParam("q", e.target.value)}
            placeholder={t("common.search")}
            className="h-9 rounded-md border bg-transparent pr-2 pl-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
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
          <p className="text-muted-foreground">{t("cheatsheets.empty")}</p>
        ) : (
          <>
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
              {data.items.map((c) => (
                <ContentCard
                  key={c.id}
                  to={`/cheatsheets/${c.id}`}
                  title={c.title}
                  Icon={BookOpen}
                  accentClass="bg-gradient-to-br from-cyan-500/25 via-cyan-500/10 to-transparent"
                  badges={
                    <>
                      <Meta>{c.category}</Meta>
                      <Meta>{c.language.toUpperCase()}</Meta>
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
