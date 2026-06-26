import { useTranslation } from "react-i18next"
import { Search } from "lucide-react"
import { useGlossary } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Pagination } from "@/components/Pagination"

export function Glossary() {
  const { t } = useTranslation()
  const { get, language, offset, pageSize, page, setParam, setLanguage, setPage } = useListParams()
  const { data, isPending, isError } = useGlossary({
    q: get("q") || undefined,
    language: language || undefined,
    limit: pageSize,
    offset,
  })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.glossary")}</h1>

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
        <div className="flex flex-col gap-3">
          {[0, 1, 2, 3].map((i) => (
            <div key={i} className="h-16 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("glossary.empty")}</p>
        ) : (
          <>
            <dl className="flex flex-col gap-3">
              {data.items.map((g) => (
                <div key={g.id} className="rounded-lg border bg-card p-4">
                  <dt className="font-semibold">{g.term}</dt>
                  <dd className="mt-1 text-sm whitespace-pre-line text-muted-foreground">
                    {g.definition_markdown}
                  </dd>
                </div>
              ))}
            </dl>
            <Pagination page={page} pageSize={pageSize} total={data.total} onPage={setPage} />
          </>
        ))}
    </div>
  )
}
