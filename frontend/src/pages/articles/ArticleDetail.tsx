import { lazy, Suspense } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, CheckCircle2 } from "lucide-react"
import { useArticle, useArticleReadStatus, useMarkArticleRead } from "@/lib/queries"
import { BookmarkButton, NoteComposer } from "@/components/ContentActions"
import { Button } from "@/components/ui/button"

// Markdown carries react-markdown + highlight.js; load it as its own chunk so
// those deps don't weigh down the rest of the app.
const Markdown = lazy(() => import("@/components/Markdown").then((m) => ({ default: m.Markdown })))

export function ArticleDetail() {
  const { t } = useTranslation()
  const { slug = "" } = useParams()
  const article = useArticle(slug)
  const readStatus = useArticleReadStatus(slug)
  const markRead = useMarkArticleRead(slug)

  return (
    <div className="flex flex-col gap-4">
      <Link
        to="/articles"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {article.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {article.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {article.data && (
        <article className="flex flex-col gap-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">{article.data.title}</h1>
              <div className="mt-2 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
                <span className="rounded border px-1.5 py-0.5">
                  {t(`difficulty.${article.data.difficulty}`)}
                </span>
                <span className="rounded border px-1.5 py-0.5">
                  {article.data.language.toUpperCase()}
                </span>
                {article.data.tags.map((tag) => (
                  <span key={tag} className="rounded border px-1.5 py-0.5">
                    #{tag}
                  </span>
                ))}
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <BookmarkButton contentType="article" contentId={article.data.id} />
              {readStatus.data?.read ? (
                <span className="flex items-center gap-1 text-sm text-primary">
                  <CheckCircle2 className="size-5" /> {t("articles.read")}
                </span>
              ) : (
                <Button
                  variant="outline"
                  onClick={() => markRead.mutate()}
                  disabled={markRead.isPending || readStatus.isPending}
                >
                  {t("articles.markRead")}
                </Button>
              )}
            </div>
          </div>

          <Suspense fallback={<div className="h-32 w-full animate-pulse rounded-md bg-muted" />}>
            <Markdown>{article.data.body_markdown}</Markdown>
          </Suspense>

          <NoteComposer contentType="article" contentId={article.data.id} />
        </article>
      )}
    </div>
  )
}
