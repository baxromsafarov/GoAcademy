import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Plus, Trash2 } from "lucide-react"
import { useProject, useSaveProject, type AdminProjectInput } from "@/lib/queries"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import { AdminFormShell, Field, TextArea, VisibilityField } from "@/components/admin/AdminFormShell"

const langOptions = ["ru", "en", "uz", "ja"].map((l) => ({ value: l, label: l.toUpperCase() }))
const diffs = ["beginner", "intermediate", "advanced"]

export function ProjectForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const editing = Boolean(id)
  const existing = useProject(id ?? "")
  const save = useSaveProject()
  const [form, setForm] = useState<AdminProjectInput | null>(null)

  const value: AdminProjectInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          description_markdown: existing.data.description_markdown,
          difficulty: existing.data.difficulty,
          language: existing.data.language,
          tags: existing.data.tags,
          steps: existing.data.steps.map((s) => ({ text: s.text })),
        }
      : {
          title: "",
          description_markdown: "",
          difficulty: "beginner",
          language: "en",
          tags: [],
          steps: [{ text: "" }],
        })

  function set<K extends keyof AdminProjectInput>(k: K, v: AdminProjectInput[K]) {
    setForm({ ...value, [k]: v })
  }
  function setStep(i: number, text: string) {
    set(
      "steps",
      value.steps.map((s, j) => (j === i ? { text } : s)),
    )
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <AdminFormShell
      backTo="/admin/projects"
      title={editing ? t("admin.editProject") : t("admin.newProject")}
      saving={save.isPending}
      isError={save.isError}
      onSubmit={(e) => {
        e.preventDefault()
        const input = { ...value, steps: value.steps.filter((s) => s.text.trim()) }
        save.mutate({ id, input }, { onSuccess: () => navigate("/admin/projects") })
      }}
    >
      <Field label={t("admin.fTitle")}>
        <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
      </Field>
      <Field label={t("admin.fBody")}>
        <TextArea
          value={value.description_markdown}
          onChange={(v) => set("description_markdown", v)}
          rows={8}
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

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium">{t("admin.fSteps")}</span>
        {value.steps.map((s, i) => (
          <div key={i} className="flex items-center gap-2">
            <Input value={s.text} onChange={(e) => setStep(i, e.target.value)} />
            <button
              type="button"
              onClick={() => set("steps", value.steps.filter((_, j) => j !== i))}
              className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
              aria-label={t("admin.remove")}
            >
              <Trash2 className="size-4" />
            </button>
          </div>
        ))}
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="w-fit"
          onClick={() => set("steps", [...value.steps, { text: "" }])}
        >
          <Plus className="size-4" /> {t("admin.addStep")}
        </Button>
      </div>
    </AdminFormShell>
  )
}
