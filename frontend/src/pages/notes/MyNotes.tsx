import { useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Pencil, Trash2, ExternalLink } from "lucide-react"
import { useNotes, useUpdateNote, useDeleteNote } from "@/lib/queries"
import { contentPath } from "@/lib/contentPath"
import { Button } from "@/components/ui/button"

export function MyNotes() {
  const { t } = useTranslation()
  const { data, isPending, isError } = useNotes()
  const update = useUpdateNote()
  const remove = useDeleteNote()
  const [editing, setEditing] = useState<string | null>(null)
  const [draft, setDraft] = useState("")

  function startEdit(id: string, body: string) {
    setEditing(id)
    setDraft(body)
  }

  function save(id: string) {
    update.mutate({ id, body: draft }, { onSuccess: () => setEditing(null) })
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.notes")}</h1>

      {isPending && (
        <div className="flex flex-col gap-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-24 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.notes.length === 0 ? (
          <p className="text-muted-foreground">{t("notes.empty")}</p>
        ) : (
          <ul className="flex flex-col gap-3">
            {data.notes.map((n) => (
              <li key={n.id} className="rounded-lg border bg-card p-4">
                <div className="mb-2 flex items-center justify-between gap-2">
                  <Link
                    to={contentPath(n.content_type, n.content_id)}
                    className="flex items-center gap-1 text-xs text-muted-foreground hover:underline"
                  >
                    <span className="rounded border px-1.5 py-0.5">{n.content_type}</span>
                    <ExternalLink className="size-3" />
                  </Link>
                  <div className="flex gap-1">
                    {editing !== n.id && (
                      <button
                        type="button"
                        onClick={() => startEdit(n.id, n.body)}
                        className="rounded p-1 text-muted-foreground hover:bg-accent"
                        aria-label={t("notes.edit")}
                      >
                        <Pencil className="size-4" />
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={() => remove.mutate(n.id)}
                      disabled={remove.isPending}
                      className="rounded p-1 text-muted-foreground hover:bg-accent"
                      aria-label={t("notes.delete")}
                    >
                      <Trash2 className="size-4" />
                    </button>
                  </div>
                </div>

                {editing === n.id ? (
                  <div className="flex flex-col gap-2">
                    <textarea
                      value={draft}
                      onChange={(e) => setDraft(e.target.value)}
                      rows={3}
                      className="w-full resize-y rounded-md border bg-transparent p-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    />
                    <div className="flex gap-2">
                      <Button onClick={() => save(n.id)} disabled={update.isPending || !draft.trim()}>
                        {t("notes.save")}
                      </Button>
                      <Button variant="outline" onClick={() => setEditing(null)}>
                        {t("notes.cancel")}
                      </Button>
                    </div>
                  </div>
                ) : (
                  <p className="text-sm whitespace-pre-line">{n.body}</p>
                )}
              </li>
            ))}
          </ul>
        ))}
    </div>
  )
}
