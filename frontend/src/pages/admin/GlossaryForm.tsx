import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { useGlossary, useSaveGlossary, type AdminGlossaryInput } from "@/lib/queries"
import { Input } from "@/components/ui/input"
import { Select } from "@/components/ui/select"
import { AdminFormShell, Field, TextArea } from "@/components/admin/AdminFormShell"

const langOptions = ["ru", "en", "uz", "ja"].map((l) => ({ value: l, label: l.toUpperCase() }))

export function GlossaryForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const editing = Boolean(id)
  // The glossary has no single-item endpoint, so for editing we load the list
  // (it is small) and pick the term by id.
  const list = useGlossary({ limit: 96 })
  const found = editing ? list.data?.items.find((g) => g.id === id) : undefined
  const save = useSaveGlossary()
  const [form, setForm] = useState<AdminGlossaryInput | null>(null)

  const value: AdminGlossaryInput =
    form ??
    (found
      ? { term: found.term, definition_markdown: found.definition_markdown, language: found.language }
      : { term: "", definition_markdown: "", language: "en" })

  function set<K extends keyof AdminGlossaryInput>(k: K, v: AdminGlossaryInput[K]) {
    setForm({ ...value, [k]: v })
  }

  if (editing && list.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <AdminFormShell
      backTo="/admin/glossary"
      title={editing ? t("admin.editGlossary") : t("admin.newGlossary")}
      saving={save.isPending}
      isError={save.isError}
      onSubmit={(e) => {
        e.preventDefault()
        save.mutate({ id, input: value }, { onSuccess: () => navigate("/admin/glossary") })
      }}
    >
      <Field label={t("admin.fTerm")}>
        <Input value={value.term} onChange={(e) => set("term", e.target.value)} required />
      </Field>
      <Field label={t("admin.fDefinition")}>
        <TextArea
          value={value.definition_markdown}
          onChange={(v) => set("definition_markdown", v)}
          rows={6}
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
    </AdminFormShell>
  )
}
