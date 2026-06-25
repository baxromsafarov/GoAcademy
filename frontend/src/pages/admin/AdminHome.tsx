import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { Video, FileText, Users } from "lucide-react"

const cards = [
  { to: "/admin/videos", labelKey: "admin.videos", icon: Video },
  { to: "/admin/articles", labelKey: "admin.articles", icon: FileText },
  { to: "/admin/users", labelKey: "admin.users", icon: Users },
]

export function AdminHome() {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("admin.title")}</h1>
      <p className="text-muted-foreground">{t("admin.intro")}</p>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {cards.map((c) => {
          const Icon = c.icon
          return (
            <Link
              key={c.to}
              to={c.to}
              className="flex items-center gap-3 rounded-lg border bg-card p-4 transition-colors hover:border-primary"
            >
              <Icon className="size-5 text-primary" />
              <span className="font-medium">{t(c.labelKey)}</span>
            </Link>
          )
        })}
      </div>
    </div>
  )
}
