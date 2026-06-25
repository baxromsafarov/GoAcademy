import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Pencil, Plus, Trash2 } from "lucide-react"
import { useVideos, useDeleteVideo } from "@/lib/queries"

const newBtnClass =
  "inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:opacity-90"

export function AdminVideos() {
  const { t } = useTranslation()
  const { data, isPending, isError } = useVideos()
  const remove = useDeleteVideo()

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">{t("admin.videos")}</h1>
        <Link to="/admin/videos/new" className={newBtnClass}>
          <Plus className="size-4" /> {t("admin.new")}
        </Link>
      </div>

      {isPending && <div className="h-32 animate-pulse rounded-lg border bg-card" />}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("videos.empty")}</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {data.items.map((v) => (
              <li key={v.id} className="flex items-center gap-3 rounded-lg border bg-card p-3">
                <span className="flex-1 font-medium">{v.title}</span>
                <span className="text-xs text-muted-foreground">{t(`difficulty.${v.difficulty}`)}</span>
                <Link
                  to={`/admin/videos/${v.id}/edit`}
                  className="rounded p-1 text-muted-foreground hover:bg-accent"
                  aria-label={t("admin.edit")}
                >
                  <Pencil className="size-4" />
                </Link>
                <button
                  type="button"
                  onClick={() => {
                    if (confirm(t("admin.confirmDelete"))) remove.mutate(v.id)
                  }}
                  disabled={remove.isPending}
                  className="rounded p-1 text-muted-foreground hover:bg-accent"
                  aria-label={t("admin.delete")}
                >
                  <Trash2 className="size-4" />
                </button>
              </li>
            ))}
          </ul>
        ))}
    </div>
  )
}
