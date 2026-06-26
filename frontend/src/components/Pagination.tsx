import { useTranslation } from "react-i18next"
import { ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"

/**
 * Pagination renders Prev / page-indicator / Next for a content list. It hides
 * itself when everything fits on one page. The total comes from the list
 * response; the page is driven by the URL via useListParams.
 */
export function Pagination({
  page,
  pageSize,
  total,
  onPage,
}: {
  page: number
  pageSize: number
  total: number
  onPage: (page: number) => void
}) {
  const { t } = useTranslation()
  const pages = Math.max(1, Math.ceil(total / pageSize))
  if (pages <= 1) return null

  return (
    <div className="flex items-center justify-center gap-3 pt-2">
      <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => onPage(page - 1)}>
        <ChevronLeft className="size-4" /> {t("common.prev")}
      </Button>
      <span className="text-sm text-muted-foreground tabular-nums">
        {t("common.pageOf", { page, pages })}
      </span>
      <Button variant="outline" size="sm" disabled={page >= pages} onClick={() => onPage(page + 1)}>
        {t("common.next")} <ChevronRight className="size-4" />
      </Button>
    </div>
  )
}
