import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Plus, Trash2 } from "lucide-react"
import { useProblem, useSaveProblem, type AdminProblemInput } from "@/lib/queries"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import { AdminFormShell, Field, TextArea, VisibilityField } from "@/components/admin/AdminFormShell"

const langOptions = ["ru", "en", "uz", "ja"].map((l) => ({ value: l, label: l.toUpperCase() }))
const diffs = ["beginner", "intermediate", "advanced"]

export function ProblemForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { slug } = useParams()
  const editing = Boolean(slug)
  const existing = useProblem(slug ?? "")
  const save = useSaveProblem()
  const [form, setForm] = useState<AdminProblemInput | null>(null)

  const value: AdminProblemInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          slug: existing.data.slug,
          statement_markdown: existing.data.statement_markdown,
          reference_solution_markdown: "",
          difficulty: existing.data.difficulty,
          language: existing.data.language,
          tags: existing.data.tags,
          sample_io: [],
          // Only the sample cases are exposed publicly; hidden tests and the
          // reference solution are not pre-filled (see answersHidden note).
          test_cases: existing.data.sample_io.map((s) => ({
            input: String(s.input ?? ""),
            expected_output: String(s.output ?? ""),
            is_sample: true,
          })),
        }
      : {
          title: "",
          slug: "",
          statement_markdown: "",
          reference_solution_markdown: "",
          difficulty: "beginner",
          language: "en",
          tags: [],
          sample_io: [],
          test_cases: [{ input: "", expected_output: "", is_sample: true }],
        })

  function set<K extends keyof AdminProblemInput>(k: K, v: AdminProblemInput[K]) {
    setForm({ ...value, [k]: v })
  }
  function setCase(i: number, patch: Partial<AdminProblemInput["test_cases"][number]>) {
    set("test_cases", value.test_cases.map((c, j) => (j === i ? { ...c, ...patch } : c)))
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <AdminFormShell
      backTo="/admin/problems"
      title={editing ? t("admin.editProblem") : t("admin.newProblem")}
      saving={save.isPending}
      isError={save.isError}
      onSubmit={(e) => {
        e.preventDefault()
        const cases = value.test_cases.filter((c) => c.input.trim() || c.expected_output.trim())
        const input: AdminProblemInput = {
          ...value,
          test_cases: cases,
          // Derive the public sample I/O from the cases flagged as samples.
          sample_io: cases
            .filter((c) => c.is_sample)
            .map((c) => ({ input: c.input, output: c.expected_output })),
        }
        save.mutate({ id: existing.data?.id, input }, { onSuccess: () => navigate("/admin/problems") })
      }}
    >
      <Field label={t("admin.fTitle")}>
        <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
      </Field>
      <Field label={t("admin.fSlug")}>
        <Input
          value={value.slug}
          onChange={(e) => set("slug", e.target.value)}
          required
          disabled={editing}
        />
      </Field>
      <Field label={t("admin.fStatement")}>
        <TextArea
          value={value.statement_markdown}
          onChange={(v) => set("statement_markdown", v)}
          rows={6}
          mono
        />
      </Field>
      <Field label={t("admin.fSolution")}>
        <TextArea
          value={value.reference_solution_markdown}
          onChange={(v) => set("reference_solution_markdown", v)}
          rows={6}
          mono
        />
      </Field>
      <div className="flex flex-wrap gap-4">
        <Field label={t("videos.filterDifficulty")}>
          <Select
            value={value.difficulty}
            onChange={(v) => set("difficulty", v)}
            options={diffs.map((d) => ({ value: d, label: t(`difficulty.${d}`) }))}
            ariaLabel={t("videos.filterDifficulty")}
            className="w-44"
          />
        </Field>
        <Field label={t("videos.filterLanguage")}>
          <Select
            value={value.language}
            onChange={(v) => set("language", v)}
            options={langOptions}
            ariaLabel={t("videos.filterLanguage")}
            className="w-28"
          />
        </Field>
      </div>
      <Field label={t("admin.fTags")}>
        <Input
          value={value.tags.join(", ")}
          onChange={(e) => set("tags", e.target.value.split(",").map((s) => s.trim()).filter(Boolean))}
          placeholder={t("admin.tagsHint")}
        />
      </Field>
      <VisibilityField tags={value.tags} onChange={(tags) => set("tags", tags)} />

      {editing && <p className="text-xs text-amber-600 dark:text-amber-400">{t("admin.answersHidden")}</p>}

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium">{t("admin.fTestCases")}</span>
        {value.test_cases.map((c, i) => (
          <div key={i} className="flex flex-col gap-2 rounded-lg border bg-card p-3">
            <div className="grid gap-2 sm:grid-cols-2">
              <div className="flex flex-col gap-1">
                <span className="text-xs text-muted-foreground">{t("admin.fInput")}</span>
                <TextArea value={c.input} onChange={(v) => setCase(i, { input: v })} rows={2} mono />
              </div>
              <div className="flex flex-col gap-1">
                <span className="text-xs text-muted-foreground">{t("admin.fExpected")}</span>
                <TextArea
                  value={c.expected_output}
                  onChange={(v) => setCase(i, { expected_output: v })}
                  rows={2}
                  mono
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <label className="flex cursor-pointer items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={c.is_sample}
                  onChange={(e) => setCase(i, { is_sample: e.target.checked })}
                  className="size-4 accent-primary"
                />
                {t("admin.fSample")}
              </label>
              <button
                type="button"
                onClick={() => set("test_cases", value.test_cases.filter((_, j) => j !== i))}
                className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
                aria-label={t("admin.remove")}
              >
                <Trash2 className="size-4" />
              </button>
            </div>
          </div>
        ))}
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="w-fit"
          onClick={() =>
            set("test_cases", [...value.test_cases, { input: "", expected_output: "", is_sample: false }])
          }
        >
          <Plus className="size-4" /> {t("admin.addTestCase")}
        </Button>
      </div>
    </AdminFormShell>
  )
}
