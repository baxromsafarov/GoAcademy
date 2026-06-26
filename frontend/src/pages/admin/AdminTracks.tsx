import { useTranslation } from "react-i18next"
import { Route } from "lucide-react"
import { useTracks, useDeleteTrack } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Meta } from "@/components/ContentCard"
import { AdminCard } from "@/components/admin/AdminCard"
import { AdminListShell } from "@/components/admin/AdminListShell"
import { SearchBox, SizeSelect } from "@/components/admin/AdminFilters"

export function AdminTracks() {
  const { t } = useTranslation()
  const lp = useListParams()
  const { data, isPending, isError } = useTracks({
    q: lp.get("q") || undefined,
    difficulty: lp.get("difficulty") || undefined,
    language: lp.language || undefined,
    limit: lp.pageSize,
    offset: lp.offset,
  })
  const remove = useDeleteTrack()

  return (
    <AdminListShell
      titleKey="admin.tracks"
      newTo="/admin/tracks/new"
      emptyKey="tracks.empty"
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
      {data?.items.map((tr) => (
        <AdminCard
          key={tr.id}
          editTo={`/admin/tracks/${tr.id}/edit`}
          title={tr.title}
          subtitle={tr.description}
          Icon={Route}
          accentClass="bg-gradient-to-br from-amber-500/25 via-amber-500/10 to-transparent"
          badges={
            <>
              <Meta>{t(`difficulty.${tr.level}`)}</Meta>
              <Meta>{tr.language.toUpperCase()}</Meta>
            </>
          }
          deleting={remove.isPending}
          onDelete={() => {
            if (confirm(t("admin.confirmDelete"))) remove.mutate(tr.id)
          }}
        />
      ))}
    </AdminListShell>
  )
}
