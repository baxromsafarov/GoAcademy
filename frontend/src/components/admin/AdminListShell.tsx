import type { ReactNode } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Plus } from "lucide-react"
import { Pagination } from "@/components/Pagination"

/**
 * AdminListShell is the common chrome for an admin content list: a title with a
 * "New" button, a filter toolbar, the responsive card grid (with loading and
 * empty states) and pagination. Pages supply the toolbar and the mapped cards.
 */
export function AdminListShell({
  titleKey,
  newTo,
  toolbar,
  isPending,
  isError,
  isEmpty,
  emptyKey,
  page,
  pageSize,
  total,
  onPage,
  children,
}: {
  titleKey: string
  newTo: string
  toolbar: ReactNode
  isPending: boolean
  isError: boolean
  isEmpty: boolean
  emptyKey: string
  page: number
  pageSize: number
  total: number
  onPage: (page: number) => void
  children: ReactNode
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold tracking-tight">{t(titleKey)}</h1>
        <Link
          to={newTo}
          className="inline-flex h-9 shrink-0 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:opacity-90"
        >
          <Plus className="size-4" /> {t("admin.new")}
        </Link>
      </div>

      <div className="flex flex-wrap gap-2">{toolbar}</div>

      {isPending && (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-60 animate-pulse rounded-xl border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {!isPending && !isError && isEmpty && <p className="text-muted-foreground">{t(emptyKey)}</p>}
      {!isPending && !isError && !isEmpty && (
        <>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">{children}</div>
          <Pagination page={page} pageSize={pageSize} total={total} onPage={onPage} />
        </>
      )}
    </div>
  )
}
