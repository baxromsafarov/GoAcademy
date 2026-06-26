import { useTranslation } from "react-i18next"
import { FolderKanban } from "lucide-react"
import { useProjects, useDeleteProject } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Meta } from "@/components/ContentCard"
import { AdminCard } from "@/components/admin/AdminCard"
import { AdminListShell } from "@/components/admin/AdminListShell"
import { SearchBox, SizeSelect } from "@/components/admin/AdminFilters"

export function AdminProjects() {
  const { t } = useTranslation()
  const lp = useListParams()
  const { data, isPending, isError } = useProjects({
    show_hidden: true,
    q: lp.get("q") || undefined,
    difficulty: lp.get("difficulty") || undefined,
    language: lp.language || undefined,
    limit: lp.pageSize,
    offset: lp.offset,
  })
  const remove = useDeleteProject()

  return (
    <AdminListShell
      titleKey="admin.projects"
      newTo="/admin/projects/new"
      emptyKey="projects.empty"
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
            value={lp.get("difficulty")}
            onChange={(v) => lp.setParam("difficulty", v)}
            options={difficultyOptions(t)}
            ariaLabel={t("videos.filterDifficulty")}
          />
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
      {data?.items.map((p) => (
        <AdminCard
          key={p.id}
          editTo={`/admin/projects/${p.id}/edit`}
          title={p.title}
          Icon={FolderKanban}
          accentClass="bg-gradient-to-br from-rose-500/25 via-rose-500/10 to-transparent"
          badges={
            <>
              <Meta>{t(`difficulty.${p.difficulty}`)}</Meta>
              <Meta>{p.language.toUpperCase()}</Meta>
            </>
          }
          hidden={p.tags.includes("hidden")}
          deleting={remove.isPending}
          onDelete={() => {
            if (confirm(t("admin.confirmDelete"))) remove.mutate(p.id)
          }}
        />
      ))}
    </AdminListShell>
  )
}
