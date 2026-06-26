import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Plus, Search } from "lucide-react"
import { useVideos, useDeleteVideo } from "@/lib/queries"
import { useListParams, PAGE_SIZES } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Pagination } from "@/components/Pagination"
import { Meta } from "@/components/ContentCard"
import { AdminCard } from "@/components/admin/AdminCard"

const newBtnClass =
  "inline-flex h-9 shrink-0 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:opacity-90"

function fmtDuration(s: number): string {
  return `${Math.floor(s / 60)}:${(s % 60).toString().padStart(2, "0")}`
}

export function AdminVideos() {
  const { t } = useTranslation()
  const { get, language, offset, pageSize, page, setParam, setLanguage, setPage, setSize } =
    useListParams()
  const { data, isPending, isError } = useVideos({
    show_hidden: true,
    q: get("q") || undefined,
    difficulty: get("difficulty") || undefined,
    language: language || undefined,
    limit: pageSize,
    offset,
  })
  const remove = useDeleteVideo()

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold tracking-tight">{t("admin.videos")}</h1>
        <Link to="/admin/videos/new" className={newBtnClass}>
          <Plus className="size-4" /> {t("admin.new")}
        </Link>
      </div>

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
        <Select
          value={String(pageSize)}
          onChange={setSize}
          options={PAGE_SIZES.map((n) => ({ value: String(n), label: t("common.perPage", { n }) }))}
          ariaLabel={t("common.perPage", { n: pageSize })}
          className="w-28"
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
                <AdminCard
                  key={v.id}
                  editTo={`/admin/videos/${v.id}/edit`}
                  title={v.title}
                  thumbnail={`https://img.youtube.com/vi/${v.youtube_id}/hqdefault.jpg`}
                  mediaBadge={v.duration_seconds > 0 ? fmtDuration(v.duration_seconds) : undefined}
                  badges={
                    <>
                      <Meta>{t(`difficulty.${v.difficulty}`)}</Meta>
                      <Meta>{v.language.toUpperCase()}</Meta>
                    </>
                  }
                  hidden={v.tags.includes("hidden")}
                  deleting={remove.isPending}
                  onDelete={() => {
                    if (confirm(t("admin.confirmDelete"))) remove.mutate(v.id)
                  }}
                />
              ))}
            </div>
            <Pagination page={page} pageSize={pageSize} total={data.total} onPage={setPage} />
          </>
        ))}
    </div>
  )
}
