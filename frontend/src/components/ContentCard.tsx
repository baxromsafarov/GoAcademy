import type { ReactNode } from "react"
import { Link } from "react-router-dom"
import type { LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"

/**
 * ContentCard is the shared rich card for content lists (videos, articles,
 * quizzes, …): a 16:9 media banner (a real thumbnail when available, otherwise a
 * tinted gradient with the content-type icon) above a title, description and a
 * meta row. It lifts and highlights on hover.
 */
export function ContentCard({
  to,
  title,
  description,
  thumbnail,
  Icon,
  accentClass,
  mediaBadge,
  badges,
  footer,
}: {
  to: string
  title: string
  description?: string
  thumbnail?: string
  Icon?: LucideIcon
  accentClass?: string
  mediaBadge?: ReactNode
  badges?: ReactNode
  footer?: ReactNode
}) {
  return (
    <Link
      to={to}
      className="group flex flex-col overflow-hidden rounded-xl border bg-card transition-all hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-lg hover:shadow-black/20"
    >
      <div
        className={cn(
          "relative flex aspect-video items-center justify-center overflow-hidden",
          accentClass ?? "bg-gradient-to-br from-primary/25 via-primary/10 to-transparent",
        )}
      >
        {thumbnail ? (
          <img
            src={thumbnail}
            alt=""
            loading="lazy"
            className="size-full object-cover transition duration-300 group-hover:scale-105"
          />
        ) : Icon ? (
          <Icon className="size-14 text-primary/60 transition group-hover:scale-110" strokeWidth={1.5} />
        ) : null}
        {mediaBadge && (
          <span className="absolute right-2 bottom-2 rounded bg-black/65 px-1.5 py-0.5 text-xs font-medium text-white">
            {mediaBadge}
          </span>
        )}
      </div>
      <div className="flex flex-1 flex-col gap-2 p-4">
        <h3 className="leading-snug font-semibold tracking-tight transition-colors group-hover:text-primary">
          {title}
        </h3>
        {description && (
          <p className="line-clamp-2 text-sm text-muted-foreground">{description}</p>
        )}
        {(badges || footer) && (
          <div className="mt-auto flex flex-col gap-2 pt-1">
            {badges && <div className="flex flex-wrap gap-1.5">{badges}</div>}
            {footer}
          </div>
        )}
      </div>
    </Link>
  )
}

/** Meta is a small pill used inside a card's badge row. */
export function Meta({ children }: { children: ReactNode }) {
  return (
    <span className="rounded-md border bg-muted/40 px-1.5 py-0.5 text-xs text-muted-foreground">
      {children}
    </span>
  )
}
