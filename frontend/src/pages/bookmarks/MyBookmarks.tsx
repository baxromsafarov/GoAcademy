import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Trash2, ChevronRight } from "lucide-react"
import { useBookmarks, useDeleteBookmark } from "@/lib/queries"
import { contentPath } from "@/lib/contentPath"

export function MyBookmarks() {
  const { t } = useTranslation()
  const { data, isPending, isError } = useBookmarks()
  const remove = useDeleteBookmark()

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.bookmarks")}</h1>

      {isPending && (
        <div className="flex flex-col gap-2">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-14 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.bookmarks.length === 0 ? (
          <p className="text-muted-foreground">{t("bookmarks.empty")}</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {data.bookmarks.map((b) => (
              <li
                key={b.id}
                className="flex items-center gap-3 rounded-lg border bg-card p-3"
              >
                <span className="rounded border px-1.5 py-0.5 text-xs text-muted-foreground">
                  {b.content_type}
                </span>
                <Link
                  to={contentPath(b.content_type, b.content_id)}
                  className="flex flex-1 items-center gap-1 text-sm font-medium hover:underline"
                >
                  {t("bookmarks.open")}
                  <ChevronRight className="size-4" />
                </Link>
                <button
                  type="button"
                  onClick={() => remove.mutate(b.id)}
                  disabled={remove.isPending}
                  className="rounded p-1 text-muted-foreground hover:bg-accent"
                  aria-label={t("bookmarks.remove")}
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
