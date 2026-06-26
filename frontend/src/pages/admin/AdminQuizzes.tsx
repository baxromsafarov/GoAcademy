import { useTranslation } from "react-i18next"
import { ListChecks } from "lucide-react"
import { useQuizzes, useDeleteQuiz } from "@/lib/queries"
import { useListParams } from "@/lib/useListParams"
import { difficultyOptions, languageOptions } from "@/lib/filterOptions"
import { Select } from "@/components/ui/select"
import { Meta } from "@/components/ContentCard"
import { AdminCard } from "@/components/admin/AdminCard"
import { AdminListShell } from "@/components/admin/AdminListShell"
import { SearchBox, SizeSelect } from "@/components/admin/AdminFilters"

export function AdminQuizzes() {
  const { t } = useTranslation()
  const lp = useListParams()
  const { data, isPending, isError } = useQuizzes({
    show_hidden: true,
    q: lp.get("q") || undefined,
    difficulty: lp.get("difficulty") || undefined,
    language: lp.language || undefined,
    limit: lp.pageSize,
    offset: lp.offset,
  })
  const remove = useDeleteQuiz()

  return (
    <AdminListShell
      titleKey="admin.quizzes"
      newTo="/admin/quizzes/new"
      emptyKey="quizzes.empty"
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
      {data?.items.map((q) => (
        <AdminCard
          key={q.id}
          editTo={`/admin/quizzes/${q.id}/edit`}
          title={q.title}
          subtitle={q.description}
          Icon={ListChecks}
          accentClass="bg-gradient-to-br from-violet-500/25 via-violet-500/10 to-transparent"
          badges={
            <>
              <Meta>{t(`difficulty.${q.difficulty}`)}</Meta>
              <Meta>{q.language.toUpperCase()}</Meta>
            </>
          }
          hidden={q.tags.includes("hidden")}
          deleting={remove.isPending}
          onDelete={() => {
            if (confirm(t("admin.confirmDelete"))) remove.mutate(q.id)
          }}
        />
      ))}
    </AdminListShell>
  )
}
