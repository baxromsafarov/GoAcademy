import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { useCheatsheet, useSaveCheatsheet, type AdminCheatsheetInput } from "@/lib/queries"
import { Input } from "@/components/ui/input"
import { Select } from "@/components/ui/select"
import { AdminFormShell, Field, TextArea } from "@/components/admin/AdminFormShell"

const langOptions = ["ru", "en", "uz", "ja"].map((l) => ({ value: l, label: l.toUpperCase() }))

export function CheatsheetForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const editing = Boolean(id)
  const existing = useCheatsheet(id ?? "")
  const save = useSaveCheatsheet()
  const [form, setForm] = useState<AdminCheatsheetInput | null>(null)

  const value: AdminCheatsheetInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          category: existing.data.category,
          body_markdown: existing.data.body_markdown,
          language: existing.data.language,
        }
      : { title: "", category: "", body_markdown: "", language: "en" })

  function set<K extends keyof AdminCheatsheetInput>(k: K, v: AdminCheatsheetInput[K]) {
    setForm({ ...value, [k]: v })
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <AdminFormShell
      backTo="/admin/cheatsheets"
      title={editing ? t("admin.editCheatsheet") : t("admin.newCheatsheet")}
      saving={save.isPending}
      isError={save.isError}
      onSubmit={(e) => {
        e.preventDefault()
        save.mutate({ id, input: value }, { onSuccess: () => navigate("/admin/cheatsheets") })
      }}
    >
      <Field label={t("admin.fTitle")}>
        <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
      </Field>
      <Field label={t("admin.fCategory")}>
        <Input value={value.category} onChange={(e) => set("category", e.target.value)} required />
      </Field>
      <Field label={t("admin.fBody")}>
        <TextArea value={value.body_markdown} onChange={(v) => set("body_markdown", v)} rows={12} mono />
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
    </AdminFormShell>
  )
}
