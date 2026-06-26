import { useTranslation } from "react-i18next"
import { BookA } from "lucide-react"
import { useGlossary, useDeleteGlossary } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Meta } from "@/components/ContentCard"
import { AdminCard } from "@/components/admin/AdminCard"
import { AdminListShell } from "@/components/admin/AdminListShell"
import { SearchBox, SizeSelect } from "@/components/admin/AdminFilters"

export function AdminGlossary() {
  const { t } = useTranslation()
  const lp = useListParams()
  const { data, isPending, isError } = useGlossary({
    q: lp.get("q") || undefined,
    language: lp.language || undefined,
    limit: lp.pageSize,
    offset: lp.offset,
  })
  const remove = useDeleteGlossary()

  return (
    <AdminListShell
      titleKey="admin.glossary"
      newTo="/admin/glossary/new"
      emptyKey="glossary.empty"
      isPending={isPending}
      isError={isError}
      isEmpty={!!data && data.items.length === 0}
      page={lp.page}
      pageSize={lp.pageSize}
      total={data?.total ?? 0}
      onPage={lp.setPage}
      toolbar={
        <>
          <SearchBox value={lp.get("q")} onChange={(v) => lp.setParam("q", v)} />
          <Select
            value={lp.language}
            onChange={lp.setLanguage}
            options={languageOptions(t)}
            ariaLabel={t("videos.filterLanguage")}
          />
          <SizeSelect pageSize={lp.pageSize} onChange={lp.setSize} />
        </>
      }
    >
      {data?.items.map((g) => (
        <AdminCard
          key={g.id}
          editTo={`/admin/glossary/${g.id}/edit`}
          title={g.term}
          subtitle={g.definition_markdown}
          Icon={BookA}
          accentClass="bg-gradient-to-br from-teal-500/25 via-teal-500/10 to-transparent"
          badges={<Meta>{g.language.toUpperCase()}</Meta>}
          deleting={remove.isPending}
          onDelete={() => {
            if (confirm(t("admin.confirmDelete"))) remove.mutate(g.id)
          }}
        />
      ))}
    </AdminListShell>
  )
}
