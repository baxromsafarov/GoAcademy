import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Plus, Trash2 } from "lucide-react"
import { useTrack, useSaveTrack, type AdminTrackInput } from "@/lib/queries"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import { AdminFormShell, Field, TextArea } from "@/components/admin/AdminFormShell"

const langOptions = ["ru", "en", "uz", "ja"].map((l) => ({ value: l, label: l.toUpperCase() }))
const diffs = ["beginner", "intermediate", "advanced"]
const contentTypes = ["video", "article", "quiz", "problem", "project"].map((c) => ({
  value: c,
  label: c,
}))

export function TrackForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const editing = Boolean(id)
  const existing = useTrack(id ?? "")
  const save = useSaveTrack()
  const [form, setForm] = useState<AdminTrackInput | null>(null)

  const value: AdminTrackInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          description: existing.data.description,
          level: existing.data.level,
          position: existing.data.position,
          language: existing.data.language,
          items: existing.data.items.map((it) => ({
            content_type: it.content_type,
            content_id: it.content_id,
          })),
        }
      : { title: "", description: "", level: "beginner", position: 1, language: "en", items: [] })

  function set<K extends keyof AdminTrackInput>(k: K, v: AdminTrackInput[K]) {
    setForm({ ...value, [k]: v })
  }
  function setItem(i: number, patch: Partial<AdminTrackInput["items"][number]>) {
    set(
      "items",
      value.items.map((it, j) => (j === i ? { ...it, ...patch } : it)),
    )
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <AdminFormShell
      backTo="/admin/tracks"
      title={editing ? t("admin.editTrack") : t("admin.newTrack")}
      saving={save.isPending}
      isError={save.isError}
      onSubmit={(e) => {
        e.preventDefault()
        const input = { ...value, items: value.items.filter((it) => it.content_id.trim()) }
        save.mutate({ id, input }, { onSuccess: () => navigate("/admin/tracks") })
      }}
    >
      <Field label={t("admin.fTitle")}>
        <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
      </Field>
      <Field label={t("admin.fDescription")}>
        <TextArea value={value.description} onChange={(v) => set("description", v)} rows={3} />
      </Field>
      <div className="flex flex-wrap gap-4">
        <Field label={t("admin.fLevel")}>
          <Select
            value={value.level}
            onChange={(v) => set("level", v)}
            options={diffs.map((d) => ({ value: d, label: t(`difficulty.${d}`) }))}
            ariaLabel={t("admin.fLevel")}
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
        <Field label={t("admin.fPosition")}>
          <Input
            type="number"
            min={1}
            value={value.position}
            onChange={(e) => set("position", Number(e.target.value))}
            className="w-24"
          />
        </Field>
      </div>

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium">{t("admin.fItems")}</span>
        {value.items.map((it, i) => (
          <div key={i} className="flex items-center gap-2">
            <Select
              value={it.content_type}
              onChange={(v) => setItem(i, { content_type: v })}
              options={contentTypes}
              ariaLabel={t("admin.fContentType")}
              className="w-32"
            />
            <Input
              value={it.content_id}
              onChange={(e) => setItem(i, { content_id: e.target.value })}
              placeholder={t("admin.fContentId")}
            />
            <button
              type="button"
              onClick={() => set("items", value.items.filter((_, j) => j !== i))}
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
          onClick={() => set("items", [...value.items, { content_type: "article", content_id: "" }])}
        >
          <Plus className="size-4" /> {t("admin.addItem")}
        </Button>
      </div>
    </AdminFormShell>
  )
}
