import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import {
  Video,
  FileText,
  ListChecks,
  Code2,
  Route,
  FolderKanban,
  BookOpen,
  BookA,
  Users,
  type LucideIcon,
} from "lucide-react"

const cards: { to: string; labelKey: string; icon: LucideIcon }[] = [
  { to: "/admin/videos", labelKey: "admin.videos", icon: Video },
  { to: "/admin/articles", labelKey: "admin.articles", icon: FileText },
  { to: "/admin/quizzes", labelKey: "admin.quizzes", icon: ListChecks },
  { to: "/admin/problems", labelKey: "admin.problems", icon: Code2 },
  { to: "/admin/tracks", labelKey: "admin.tracks", icon: Route },
  { to: "/admin/projects", labelKey: "admin.projects", icon: FolderKanban },
  { to: "/admin/cheatsheets", labelKey: "admin.cheatsheets", icon: BookOpen },
  { to: "/admin/glossary", labelKey: "admin.glossary", icon: BookA },
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
              className="group flex items-center gap-3 rounded-xl border bg-card p-4 transition-all hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-sm"
            >
              <span className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <Icon className="size-5" />
              </span>
              <span className="font-medium transition-colors group-hover:text-primary">
                {t(c.labelKey)}
              </span>
            </Link>
          )
        })}
      </div>
    </div>
  )
}
