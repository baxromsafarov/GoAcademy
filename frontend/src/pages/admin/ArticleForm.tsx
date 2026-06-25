import { useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft } from "lucide-react"
import { useArticle, useSaveArticle, type AdminArticleInput } from "@/lib/queries"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const difficulties = ["beginner", "intermediate", "advanced"]
const langs = ["ru", "en", "uz", "ja"]
const selectClass =
  "h-9 rounded-md border bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"

export function ArticleForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { slug } = useParams()
  const editing = Boolean(slug)
  const existing = useArticle(slug ?? "")
  const save = useSaveArticle()

  const [form, setForm] = useState<AdminArticleInput | null>(null)

  const value: AdminArticleInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          slug: existing.data.slug,
          body_markdown: existing.data.body_markdown,
          difficulty: existing.data.difficulty,
          language: existing.data.language,
          tags: existing.data.tags,
        }
      : {
          title: "",
          slug: "",
          body_markdown: "",
          difficulty: "beginner",
          language: "en",
          tags: [],
        })

  function set<K extends keyof AdminArticleInput>(key: K, v: AdminArticleInput[K]) {
    setForm({ ...value, [key]: v })
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    // Update targets the article id (PATCH /admin/articles/{id}); create has none.
    save.mutate({ id: existing.data?.id, input: value }, { onSuccess: () => navigate("/admin/articles") })
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <div className="flex max-w-2xl flex-col gap-4">
      <Link
        to="/admin/articles"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>
      <h1 className="text-2xl font-semibold tracking-tight">
        {editing ? t("admin.editArticle") : t("admin.newArticle")}
      </h1>

      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label>{t("admin.fTitle")}</Label>
          <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label>{t("admin.fSlug")}</Label>
          <Input
            value={value.slug}
            onChange={(e) => set("slug", e.target.value)}
            required
            disabled={editing}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label>{t("admin.fBody")}</Label>
          <textarea
            value={value.body_markdown}
            onChange={(e) => set("body_markdown", e.target.value)}
            rows={12}
            className="w-full resize-y rounded-md border bg-transparent p-2 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
        <div className="flex flex-wrap gap-4">
          <div className="flex flex-col gap-1.5">
            <Label>{t("videos.filterDifficulty")}</Label>
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
          </div>
          <div className="flex flex-col gap-1.5">
            <Label>{t("videos.filterLanguage")}</Label>
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
          </div>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label>{t("admin.fTags")}</Label>
          <Input
            value={value.tags.join(", ")}
            onChange={(e) => set("tags", e.target.value.split(",").map((s) => s.trim()).filter(Boolean))}
            placeholder={t("admin.tagsHint")}
          />
        </div>

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
