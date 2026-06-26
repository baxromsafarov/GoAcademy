import { useTranslation } from "react-i18next"
import { BookOpen } from "lucide-react"
import { useCheatsheets, useDeleteCheatsheet } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Meta } from "@/components/ContentCard"
import { AdminCard } from "@/components/admin/AdminCard"
import { AdminListShell } from "@/components/admin/AdminListShell"
import { SearchBox, SizeSelect } from "@/components/admin/AdminFilters"

export function AdminCheatsheets() {
  const { t } = useTranslation()
  const lp = useListParams()
  const { data, isPending, isError } = useCheatsheets({
    q: lp.get("q") || undefined,
    language: lp.language || undefined,
    limit: lp.pageSize,
    offset: lp.offset,
  })
  const remove = useDeleteCheatsheet()

  return (
    <AdminListShell
      titleKey="admin.cheatsheets"
      newTo="/admin/cheatsheets/new"
      emptyKey="cheatsheets.empty"
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
      {data?.items.map((c) => (
        <AdminCard
          key={c.id}
          editTo={`/admin/cheatsheets/${c.id}/edit`}
          title={c.title}
          subtitle={c.category}
          Icon={BookOpen}
          accentClass="bg-gradient-to-br from-cyan-500/25 via-cyan-500/10 to-transparent"
          badges={
            <>
              <Meta>{c.category}</Meta>
              <Meta>{c.language.toUpperCase()}</Meta>
            </>
          }
          deleting={remove.isPending}
          onDelete={() => {
            if (confirm(t("admin.confirmDelete"))) remove.mutate(c.id)
          }}
        />
      ))}
    </AdminListShell>
  )
}
