import {
  LayoutDashboard,
  Video,
  FileText,
  ListChecks,
  Code2,
  Route,
  Terminal,
  BookOpen,
  FolderKanban,
  BookA,
  Trophy,
  StickyNote,
  Bookmark,
  User,
  Shield,
  type LucideIcon,
} from "lucide-react"

export type Section = {
  path: string
  labelKey: string
  icon: LucideIcon
  adminOnly?: boolean
}

/** The top-level sections of the app, used by the sidebar and dashboard.
 * labelKey is an i18n key resolved at render time. */
export const sections: Section[] = [
  { path: "/", labelKey: "nav.dashboard", icon: LayoutDashboard },
  { path: "/videos", labelKey: "nav.videos", icon: Video },
  { path: "/articles", labelKey: "nav.articles", icon: FileText },
  { path: "/quizzes", labelKey: "nav.quizzes", icon: ListChecks },
  { path: "/problems", labelKey: "nav.problems", icon: Code2 },
  { path: "/tracks", labelKey: "nav.tracks", icon: Route },
  { path: "/sandbox", labelKey: "nav.sandbox", icon: Terminal },
  { path: "/cheatsheets", labelKey: "nav.cheatsheets", icon: BookOpen },
  { path: "/projects", labelKey: "nav.projects", icon: FolderKanban },
  { path: "/glossary", labelKey: "nav.glossary", icon: BookA },
  { path: "/leaderboard", labelKey: "nav.leaderboard", icon: Trophy },
  { path: "/notes", labelKey: "nav.notes", icon: StickyNote },
  { path: "/bookmarks", labelKey: "nav.bookmarks", icon: Bookmark },
  { path: "/profile", labelKey: "nav.profile", icon: User },
  { path: "/admin", labelKey: "nav.admin", icon: Shield, adminOnly: true },
]
