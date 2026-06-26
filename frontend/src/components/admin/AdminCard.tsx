import type { ReactNode } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { EyeOff, Pencil, Trash2, type LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"

/**
 * AdminCard mirrors the public ContentCard (media banner + title + meta) but
 * adds edit and delete actions instead of being a single link — so admins get
 * the same at-a-glance layout when managing content.
 */
export function AdminCard({
  editTo,
  title,
  subtitle,
  thumbnail,
  Icon,
  accentClass,
  mediaBadge,
  badges,
  onDelete,
  deleting,
  hidden,
}: {
  editTo: string
  title: string
  subtitle?: string
  thumbnail?: string
  Icon?: LucideIcon
  accentClass?: string
  mediaBadge?: ReactNode
  badges?: ReactNode
  onDelete: () => void
  deleting?: boolean
  hidden?: boolean
}) {
  const { t } = useTranslation()
  return (
    <div
      className={cn(
        "group relative flex flex-col overflow-hidden rounded-xl border bg-card transition-all hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-lg hover:shadow-black/20",
        hidden && "opacity-60",
      )}
    >
      {hidden && (
        <span className="absolute z-10 m-2 flex items-center gap-1 rounded bg-black/70 px-1.5 py-0.5 text-xs font-medium text-white">
          <EyeOff className="size-3" /> {t("admin.hiddenBadge")}
        </span>
      )}
      <Link to={editTo} className="block">
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
      </Link>
      <div className="flex flex-1 flex-col gap-2 p-4">
        <Link to={editTo}>
          <h3 className="leading-snug font-semibold tracking-tight transition-colors group-hover:text-primary">
            {title}
          </h3>
        </Link>
        {subtitle && <p className="line-clamp-1 text-xs text-muted-foreground">{subtitle}</p>}
        <div className="mt-auto flex items-end justify-between gap-2 pt-1">
          <div className="flex flex-wrap gap-1.5">{badges}</div>
          <div className="flex shrink-0 items-center gap-1">
            <Link
              to={editTo}
              className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              aria-label={t("admin.edit")}
            >
              <Pencil className="size-4" />
            </Link>
            <button
              type="button"
              onClick={onDelete}
              disabled={deleting}
              className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-red-500/10 hover:text-red-500 disabled:opacity-50"
              aria-label={t("admin.delete")}
            >
              <Trash2 className="size-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
