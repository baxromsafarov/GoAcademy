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

export type NavGroup = {
  /** i18n key for the group heading; omitted for the top (overview) group. */
  titleKey?: string
  items: Section[]
}

/** The sidebar navigation, grouped into themed sections so the menu reads as a
 * balanced, modern layout rather than one long flat list. */
export const navGroups: NavGroup[] = [
  {
    items: [{ path: "/", labelKey: "nav.dashboard", icon: LayoutDashboard }],
  },
  {
    titleKey: "navGroup.learn",
    items: [
      { path: "/videos", labelKey: "nav.videos", icon: Video },
      { path: "/articles", labelKey: "nav.articles", icon: FileText },
      { path: "/tracks", labelKey: "nav.tracks", icon: Route },
      { path: "/cheatsheets", labelKey: "nav.cheatsheets", icon: BookOpen },
      { path: "/glossary", labelKey: "nav.glossary", icon: BookA },
    ],
  },
  {
    titleKey: "navGroup.practice",
    items: [
      { path: "/quizzes", labelKey: "nav.quizzes", icon: ListChecks },
      { path: "/problems", labelKey: "nav.problems", icon: Code2 },
      { path: "/sandbox", labelKey: "nav.sandbox", icon: Terminal },
      { path: "/projects", labelKey: "nav.projects", icon: FolderKanban },
    ],
  },
  {
    titleKey: "navGroup.progress",
    items: [
      { path: "/leaderboard", labelKey: "nav.leaderboard", icon: Trophy },
      { path: "/notes", labelKey: "nav.notes", icon: StickyNote },
      { path: "/bookmarks", labelKey: "nav.bookmarks", icon: Bookmark },
    ],
  },
  {
    titleKey: "navGroup.account",
    items: [
      { path: "/profile", labelKey: "nav.profile", icon: User },
      { path: "/admin", labelKey: "nav.admin", icon: Shield, adminOnly: true },
    ],
  },
]

/** Flat list of every section, kept for any consumer that needs the whole set. */
export const sections: Section[] = navGroups.flatMap((g) => g.items)
