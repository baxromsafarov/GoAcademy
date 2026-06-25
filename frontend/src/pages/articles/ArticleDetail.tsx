import { lazy, Suspense } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft, CheckCircle2, Clock, FileText } from "lucide-react"
import { useArticle, useArticleReadStatus, useMarkArticleRead } from "@/lib/queries"
import { BookmarkButton, NoteComposer } from "@/components/ContentActions"
import { ReadingProgress } from "@/components/ReadingProgress"
import { Button } from "@/components/ui/button"

// Markdown carries react-markdown + highlight.js; load it as its own chunk so
// those deps don't weigh down the rest of the app.
const Markdown = lazy(() => import("@/components/Markdown").then((m) => ({ default: m.Markdown })))

/** Rough reading time at ~180 words/min, floored to one minute. */
function readingMinutes(markdown: string): number {
  const words = markdown.trim().split(/\s+/).filter(Boolean).length
  return Math.max(1, Math.round(words / 180))
}

export function ArticleDetail() {
  const { t } = useTranslation()
  const { slug = "" } = useParams()
  const article = useArticle(slug)
  const readStatus = useArticleReadStatus(slug)
  const markRead = useMarkArticleRead(slug)

  return (
    <div className="mx-auto w-full max-w-3xl">
      <ReadingProgress />

      <Link
        to="/articles"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {article.isPending && <div className="mt-6 h-64 w-full animate-pulse rounded-md bg-muted" />}
      {article.isError && <p className="mt-6 text-sm text-red-500">{t("common.error")}</p>}

      {article.data && (
        <article className="mt-4">
          <header className="border-b pb-6">
            <div className="mb-3 flex items-center gap-1.5 text-xs font-medium tracking-wide text-sky-500 uppercase">
              <FileText className="size-3.5" /> {t("nav.articles")}
            </div>
            <h1 className="text-3xl leading-tight font-bold tracking-tight md:text-4xl">
              {article.data.title}
            </h1>
            <div className="mt-4 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <Clock className="size-3.5" />
                {t("articles.readingTime", { min: readingMinutes(article.data.body_markdown) })}
              </span>
              <span className="text-muted-foreground/40">·</span>
              <span className="rounded-md border px-1.5 py-0.5">
                {t(`difficulty.${article.data.difficulty}`)}
              </span>
              <span className="rounded-md border px-1.5 py-0.5">
                {article.data.language.toUpperCase()}
              </span>
              {article.data.tags.map((tag) => (
                <span key={tag} className="rounded-md border px-1.5 py-0.5">
                  #{tag}
                </span>
              ))}
              <span className="ml-auto">
                <BookmarkButton contentType="article" contentId={article.data.id} />
              </span>
            </div>
          </header>

          <div className="py-2 text-[1.05rem] leading-relaxed">
            <Suspense fallback={<div className="h-32 w-full animate-pulse rounded-md bg-muted" />}>
              <Markdown>{article.data.body_markdown}</Markdown>
            </Suspense>
          </div>

          <footer className="mt-8 flex flex-col items-center gap-3 border-t pt-6">
            {readStatus.data?.read ? (
              <span className="flex items-center gap-2 text-sm font-medium text-primary">
                <CheckCircle2 className="size-5" /> {t("articles.read")}
              </span>
            ) : (
              <Button
                size="lg"
                className="w-full sm:w-auto"
                onClick={() => markRead.mutate()}
                disabled={markRead.isPending || readStatus.isPending}
              >
                <CheckCircle2 className="size-4" /> {t("articles.markRead")}
              </Button>
            )}
          </footer>

          <div className="mt-8">
            <NoteComposer contentType="article" contentId={article.data.id} />
          </div>
        </article>
      )}
    </div>
  )
}
