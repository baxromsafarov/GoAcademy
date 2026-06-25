import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Pencil, Plus, Trash2 } from "lucide-react"
import { useArticles, useDeleteArticle } from "@/lib/queries"

const newBtnClass =
  "inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:opacity-90"

export function AdminArticles() {
  const { t } = useTranslation()
  const { data, isPending, isError } = useArticles()
  const remove = useDeleteArticle()

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">{t("admin.articles")}</h1>
        <Link to="/admin/articles/new" className={newBtnClass}>
          <Plus className="size-4" /> {t("admin.new")}
        </Link>
      </div>

      {isPending && <div className="h-32 animate-pulse rounded-lg border bg-card" />}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.items.length === 0 ? (
          <p className="text-muted-foreground">{t("articles.empty")}</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {data.items.map((a) => (
              <li key={a.id} className="flex items-center gap-3 rounded-lg border bg-card p-3">
                <span className="flex-1 font-medium">{a.title}</span>
                <span className="text-xs text-muted-foreground">{a.slug}</span>
                <Link
                  to={`/admin/articles/${a.slug}/edit`}
                  className="rounded p-1 text-muted-foreground hover:bg-accent"
                  aria-label={t("admin.edit")}
                >
                  <Pencil className="size-4" />
                </Link>
                <button
                  type="button"
                  onClick={() => {
                    if (confirm(t("admin.confirmDelete"))) remove.mutate(a.id)
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
