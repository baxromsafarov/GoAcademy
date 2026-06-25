import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Bookmark } from "lucide-react"
import {
  useBookmarks,
  useCreateBookmark,
  useDeleteBookmark,
  useCreateNote,
} from "@/lib/queries"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

/** BookmarkButton toggles a bookmark for a piece of content (create/delete). */
export function BookmarkButton({ contentType, contentId }: { contentType: string; contentId: string }) {
  const { t } = useTranslation()
  const { data } = useBookmarks()
  const create = useCreateBookmark()
  const remove = useDeleteBookmark()
  const existing = data?.bookmarks.find(
    (b) => b.content_type === contentType && b.content_id === contentId,
  )
  const busy = create.isPending || remove.isPending

  function toggle() {
    if (existing) remove.mutate(existing.id)
    else create.mutate({ content_type: contentType, content_id: contentId })
  }

  return (
    <Button variant="outline" onClick={toggle} disabled={busy}>
      <Bookmark className={cn("size-4", existing && "fill-current text-primary")} />
      {existing ? t("bookmarks.saved") : t("bookmarks.save")}
    </Button>
  )
}

/** NoteComposer adds a note attached to a piece of content. Read/edit/delete of
 * existing notes lives on the My Notes page. */
export function NoteComposer({ contentType, contentId }: { contentType: string; contentId: string }) {
  const { t } = useTranslation()
  const create = useCreateNote()
  const [body, setBody] = useState("")

  function add() {
    create.mutate(
      { content_type: contentType, content_id: contentId, body },
      { onSuccess: () => setBody("") },
    )
  }

  return (
    <section className="flex flex-col gap-2 border-t pt-4">
      <h2 className="text-lg font-semibold">{t("notes.addHeading")}</h2>
      <textarea
        value={body}
        onChange={(e) => setBody(e.target.value)}
        rows={3}
        placeholder={t("notes.placeholder")}
        className="w-full resize-y rounded-md border bg-transparent p-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
      />
      <div className="flex items-center gap-3">
        <Button onClick={add} disabled={!body.trim() || create.isPending}>
          {t("notes.add")}
        </Button>
        {create.isSuccess && (
          <span className="text-sm text-green-600 dark:text-green-400">{t("notes.added")}</span>
        )}
        {create.isError && <span className="text-sm text-red-500">{t("common.error")}</span>}
      </div>
    </section>
  )
}
