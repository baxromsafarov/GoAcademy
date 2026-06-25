import { lazy, Suspense } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowLeft } from "lucide-react"
import { useCheatsheet } from "@/lib/queries"

const Markdown = lazy(() => import("@/components/Markdown").then((m) => ({ default: m.Markdown })))

export function CheatsheetDetail() {
  const { t } = useTranslation()
  const { id = "" } = useParams()
  const cheatsheet = useCheatsheet(id)

  return (
    <div className="flex flex-col gap-4">
      <Link
        to="/cheatsheets"
        className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:underline"
      >
        <ArrowLeft className="size-4" /> {t("common.back")}
      </Link>

      {cheatsheet.isPending && <div className="h-64 w-full animate-pulse rounded-md bg-muted" />}
      {cheatsheet.isError && <p className="text-sm text-red-500">{t("common.error")}</p>}

      {cheatsheet.data && (
        <article className="flex flex-col gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{cheatsheet.data.title}</h1>
            <div className="mt-2 flex flex-wrap gap-1.5 text-xs text-muted-foreground">
              <span className="rounded border px-1.5 py-0.5">{cheatsheet.data.category}</span>
              <span className="rounded border px-1.5 py-0.5">{cheatsheet.data.language.toUpperCase()}</span>
            </div>
          </div>
          <Suspense fallback={<div className="h-32 w-full animate-pulse rounded-md bg-muted" />}>
            <Markdown>{cheatsheet.data.body_markdown}</Markdown>
          </Suspense>
        </article>
      )}
    </div>
  )
}
