import { useTranslation } from "react-i18next"
import { Search } from "lucide-react"
import { PAGE_SIZES } from "@/lib/useListParams"
import { Select } from "@/components/ui/select"

/** Title search box shared by the admin lists. */
export function SearchBox({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const { t } = useTranslation()
  return (
    <div className="relative">
      <Search className="absolute top-1/2 left-2 size-4 -translate-y-1/2 text-muted-foreground" />
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={t("common.search")}
        className="h-9 rounded-md border bg-transparent pr-2 pl-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
      />
    </div>
  )
}

/** Page-size picker shared by the admin lists. */
export function SizeSelect({
  pageSize,
  onChange,
}: {
  pageSize: number
  onChange: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <Select
      value={String(pageSize)}
      onChange={onChange}
      options={PAGE_SIZES.map((n) => ({ value: String(n), label: t("common.perPage", { n }) }))}
      ariaLabel={t("common.perPage", { n: pageSize })}
      className="w-28"
    />
  )
}
