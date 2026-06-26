import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Plus, Trash2 } from "lucide-react"
import { useQuiz, useSaveQuiz, type AdminQuizInput } from "@/lib/queries"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import { AdminFormShell, Field, TextArea, VisibilityField } from "@/components/admin/AdminFormShell"

const langOptions = ["ru", "en", "uz", "ja"].map((l) => ({ value: l, label: l.toUpperCase() }))
const diffs = ["beginner", "intermediate", "advanced"]

export function QuizForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const editing = Boolean(id)
  const existing = useQuiz(id ?? "")
  const save = useSaveQuiz()
  const [form, setForm] = useState<AdminQuizInput | null>(null)

  const value: AdminQuizInput =
    form ??
    (editing && existing.data
      ? {
          title: existing.data.title,
          description: existing.data.description,
          pass_threshold: existing.data.pass_threshold,
          difficulty: existing.data.difficulty,
          language: existing.data.language,
          tags: existing.data.tags,
          // The public detail hides which option is correct, so seed is_correct
          // as false; the admin re-marks them (see the answersHidden note).
          questions: existing.data.questions.map((q) => ({
            prompt: q.prompt,
            type: q.type,
            options: q.options.map((o) => ({ text: o.text, is_correct: false })),
          })),
        }
      : {
          title: "",
          description: "",
          pass_threshold: 60,
          difficulty: "beginner",
          language: "en",
          tags: [],
          questions: [
            { prompt: "", type: "single", options: [{ text: "", is_correct: true }, { text: "", is_correct: false }] },
          ],
        })

  function set<K extends keyof AdminQuizInput>(k: K, v: AdminQuizInput[K]) {
    setForm({ ...value, [k]: v })
  }
  function setQuestion(qi: number, patch: Partial<AdminQuizInput["questions"][number]>) {
    set("questions", value.questions.map((q, j) => (j === qi ? { ...q, ...patch } : q)))
  }
  function setOption(qi: number, oi: number, patch: Partial<{ text: string; is_correct: boolean }>) {
    setQuestion(qi, {
      options: value.questions[qi].options.map((o, j) => (j === oi ? { ...o, ...patch } : o)),
    })
  }

  if (editing && existing.isPending) return <div className="h-64 animate-pulse rounded-md bg-muted" />

  return (
    <AdminFormShell
      backTo="/admin/quizzes"
      title={editing ? t("admin.editQuiz") : t("admin.newQuiz")}
      saving={save.isPending}
      isError={save.isError}
      onSubmit={(e) => {
        e.preventDefault()
        save.mutate({ id, input: value }, { onSuccess: () => navigate("/admin/quizzes") })
      }}
    >
      <Field label={t("admin.fTitle")}>
        <Input value={value.title} onChange={(e) => set("title", e.target.value)} required />
      </Field>
      <Field label={t("admin.fDescription")}>
        <TextArea value={value.description} onChange={(v) => set("description", v)} rows={2} />
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
        <Field label={t("admin.fPassThreshold")}>
          <Input
            type="number"
            min={0}
            max={100}
            value={value.pass_threshold}
            onChange={(e) => set("pass_threshold", Number(e.target.value))}
            className="w-24"
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

      <div className="flex flex-col gap-4">
        <span className="text-sm font-medium">{t("admin.fQuestions")}</span>
        {value.questions.map((q, qi) => (
          <div key={qi} className="flex flex-col gap-2 rounded-lg border bg-card p-3">
            <div className="flex items-center gap-2">
              <Input
                value={q.prompt}
                onChange={(e) => setQuestion(qi, { prompt: e.target.value })}
                placeholder={t("admin.fPrompt")}
              />
              <Select
                value={q.type}
                onChange={(v) => setQuestion(qi, { type: v })}
                options={[
                  { value: "single", label: t("quizzes.single") },
                  { value: "multiple", label: t("quizzes.multiple") },
                ]}
                ariaLabel={t("admin.fType")}
                className="w-36"
              />
              <button
                type="button"
                onClick={() => set("questions", value.questions.filter((_, j) => j !== qi))}
                className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
                aria-label={t("admin.remove")}
              >
                <Trash2 className="size-4" />
              </button>
            </div>
            {q.options.map((o, oi) => (
              <div key={oi} className="flex items-center gap-2 pl-3">
                <input
                  type="checkbox"
                  checked={o.is_correct}
                  onChange={(e) => setOption(qi, oi, { is_correct: e.target.checked })}
                  className="size-4 accent-primary"
                  title={t("admin.fCorrect")}
                />
                <Input value={o.text} onChange={(e) => setOption(qi, oi, { text: e.target.value })} />
                <button
                  type="button"
                  onClick={() => setQuestion(qi, { options: q.options.filter((_, j) => j !== oi) })}
                  className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
                  aria-label={t("admin.remove")}
                >
                  <Trash2 className="size-4" />
                </button>
              </div>
            ))}
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="w-fit"
              onClick={() => setQuestion(qi, { options: [...q.options, { text: "", is_correct: false }] })}
            >
              <Plus className="size-4" /> {t("admin.addOption")}
            </Button>
          </div>
        ))}
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="w-fit"
          onClick={() =>
            set("questions", [
              ...value.questions,
              { prompt: "", type: "single", options: [{ text: "", is_correct: true }, { text: "", is_correct: false }] },
            ])
          }
        >
          <Plus className="size-4" /> {t("admin.addQuestion")}
        </Button>
      </div>
    </AdminFormShell>
  )
}
