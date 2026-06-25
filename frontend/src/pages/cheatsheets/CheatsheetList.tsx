import { useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Search } from "lucide-react"
import { useCheatsheets, type CheatsheetFilters } from "@/lib/queries"

const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

export function CheatsheetList() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<CheatsheetFilters>({})
  const { data, isPending, isError } = useCheatsheets(filters)

  function set(key: keyof CheatsheetFilters, value: string) {
    setFilters((f) => ({ ...f, [key]: value || undefined }))
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.cheatsheets")}</h1>

      <div className="flex flex-wrap gap-2">
        <div className="relative">
          <Search className="absolute top-1/2 left-2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            value={filters.q ?? ""}
            onChange={(e) => set("q", e.target.value)}
            placeholder={t("common.search")}
            className="h-9 rounded-md border bg-transparent pr-2 pl-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
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
            <div key={i} className="h-20 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("cheatsheets.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((c) => (
              <Link
                key={c.id}
                to={`/cheatsheets/${c.id}`}
                className="rounded-lg border bg-card p-4 transition-colors hover:border-primary"
              >
                <div className="font-medium">{c.title}</div>
                <div className="mt-3 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                  <span className="rounded border px-1.5 py-0.5">{c.category}</span>
                  <span className="rounded border px-1.5 py-0.5">{c.language.toUpperCase()}</span>
                </div>
              </Link>
            ))}
          </div>
        ))}
    </div>
  )
}
