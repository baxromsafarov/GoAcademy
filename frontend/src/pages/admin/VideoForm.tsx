import { useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft } from "lucide-react"
import { useVideo, useSaveVideo, type AdminVideoInput } from "@/lib/queries"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

export function VideoForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const editing = Boolean(id)
  const existing = useVideo(id ?? "")
  const save = useSaveVideo()

  const [form, setForm] = useState<AdminVideoInput | null>(null)

  // Initialise from the loaded video once (edit) or with blanks (create).
  const value: AdminVideoInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          description: existing.data.description,
          youtube_id: existing.data.youtube_id,
          duration_seconds: existing.data.duration_seconds,
          difficulty: existing.data.difficulty,
          language: existing.data.language,
          tags: existing.data.tags,
        }
      : {
          title: "",
          description: "",
          youtube_id: "",
          duration_seconds: 0,
          difficulty: "beginner",
          language: "en",
          tags: [],
        })

  function set<K extends keyof AdminVideoInput>(key: K, v: AdminVideoInput[K]) {
    setForm({ ...value, [key]: v })
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    save.mutate({ id, input: value }, { onSuccess: () => navigate("/admin/videos") })
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <div className="flex max-w-2xl flex-col gap-4">
      <Link
        to="/admin/videos"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>
      <h1 className="text-2xl font-semibold tracking-tight">
        {editing ? t("admin.editVideo") : t("admin.newVideo")}
      </h1>

      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <Field label={t("admin.fTitle")}>
          <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
        </Field>
        <Field label={t("admin.fDescription")}>
          <textarea
            value={value.description}
            onChange={(e) => set("description", e.target.value)}
            rows={3}
            className="w-full resize-y rounded-md border bg-transparent p-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </Field>
        <Field label={t("admin.fYoutubeId")}>
          <Input value={value.youtube_id} onChange={(e) => set("youtube_id", e.target.value)} required />
        </Field>
        <Field label={t("admin.fDuration")}>
          <Input
            type="number"
            min={0}
            value={value.duration_seconds}
            onChange={(e) => set("duration_seconds", Number(e.target.value))}
          />
        </Field>
        <div className="flex flex-wrap gap-4">
          <Field label={t("videos.filterDifficulty")}>
            <select
              value={value.difficulty}
              onChange={(e) => set("difficulty", e.target.value)}
              className={selectClass}
            >
              {difficulties.map((d) => (
                <option key={d} value={d}>
                  {t(`difficulty.${d}`)}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t("videos.filterLanguage")}>
            <select
              value={value.language}
              onChange={(e) => set("language", e.target.value)}
              className={selectClass}
            >
              {langs.map((l) => (
                <option key={l} value={l}>
                  {l.toUpperCase()}
                </option>
              ))}
            </select>
          </Field>
        </div>
        <Field label={t("admin.fTags")}>
          <Input
            value={value.tags.join(", ")}
            onChange={(e) => set("tags", e.target.value.split(",").map((s) => s.trim()).filter(Boolean))}
            placeholder={t("admin.tagsHint")}
          />
        </Field>

        <div className="flex items-center gap-3">
          <Button type="submit" disabled={save.isPending}>
            {save.isPending ? t("admin.saving") : t("admin.save")}
          </Button>
          {save.isError && <span className="text-sm text-red-500">{t("admin.saveError")}</span>}
        </div>
      </form>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label>{label}</Label>
      {children}
    </div>
  )
}
