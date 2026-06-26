import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"
import {
  Award,
  Bookmark,
  ChevronRight,
  Code2,
  FileText,
  FolderKanban,
  Flame,
  HelpCircle,
  ListChecks,
  Route,
  TrendingUp,
  Trophy,
  Video,
  Zap,
  type LucideIcon,
} from "lucide-react"
import { Card, CardTitle } from "@/components/ui/card"
import { ActivityHeatmap } from "@/components/ActivityHeatmap"
import { useAuth } from "@/lib/auth-context"
import {
  useActivity,
  useBadges,
  useBookmarks,
  useLeaderboard,
  useMyTracks,
  useProgressSummary,
  useRecentCompletions,
  useStats,
} from "@/lib/queries"
import { contentPath } from "@/lib/contentPath"
import { cn } from "@/lib/utils"

/** Mirrors the backend curve level = 1 + floor(sqrt(xp/100)); level L starts at
 * 100*(L-1)^2 XP. Returns how far into the current level the user is. */
function levelProgress(totalXp: number, level: number) {
  const base = 100
  const curStart = base * (level - 1) ** 2
  const nextStart = base * level ** 2
  const span = nextStart - curStart
  const pct = span > 0 ? Math.min(100, Math.max(0, ((totalXp - curStart) / span) * 100)) : 0
  return { pct, toNext: Math.max(0, nextStart - totalXp) }
}

const recentIcon: Record<string, LucideIcon> = {
  video: Video,
  article: FileText,
  quiz: ListChecks,
  problem: Code2,
}

export function Dashboard() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const stats = useStats()
  const progress = useProgressSummary()
  const badges = useBadges()
  const activity = useActivity()
  const leaderboard = useLeaderboard("all", 100)
  const myTracks = useMyTracks()
  const bookmarks = useBookmarks()
  const recent = useRecentCompletions()
  const rank = leaderboard.data?.entries.find((e) => e.user_id === user?.id)?.rank

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">
          {user ? t("dashboard.greeting", { name: user.display_name }) : t("dashboard.title")}
        </h1>
        <p className="mt-1 text-muted-foreground">{t("dashboard.welcome")}</p>
      </div>

      {/* Level banner with progress to the next level */}
      {stats.isPending && <Card className="h-28 animate-pulse" />}
      {stats.data &&
        (() => {
          const { pct, toNext } = levelProgress(stats.data.total_xp, stats.data.level)
          return (
            <Card className="flex flex-col gap-3 border-primary/20 bg-gradient-to-br from-primary/10 via-card to-card p-5">
              <div className="flex items-center gap-4">
                <div className="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-primary/15 text-primary">
                  <TrendingUp className="size-7" />
                </div>
                <div className="flex-1">
                  <div className="text-sm text-muted-foreground">{t("dashboard.level")}</div>
                  <div className="text-3xl font-bold tracking-tight">{stats.data.level}</div>
                </div>
                <div className="text-right">
                  <div className="flex items-center justify-end gap-1 text-2xl font-bold text-primary">
                    <Zap className="size-5" />
                    {stats.data.total_xp}
                  </div>
                  <div className="text-xs text-muted-foreground">{t("dashboard.xp")}</div>
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-primary transition-all"
                    style={{ width: `${pct}%` }}
                  />
                </div>
                <p className="text-xs text-muted-foreground">
                  {t("dashboard.levelProgress", { xp: toNext, level: stats.data.level + 1 })}
                </p>
              </div>
            </Card>
          )
        })()}

      {/* Streak stats + clickable rank */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.isPending &&
          [0, 1, 2, 3].map((i) => <Card key={i} className="h-[5.5rem] animate-pulse" />)}
        {stats.isError && <p className="col-span-full text-sm text-red-500">{t("common.error")}</p>}
        {stats.data && (
          <>
            <Stat
              icon={Trophy}
              tone="violet"
              label={t("dashboard.rank")}
              value={rank ? `#${rank}` : "—"}
              to="/leaderboard"
            />
            <Stat icon={Zap} tone="amber" label={t("dashboard.xp")} value={stats.data.total_xp} />
            <Stat
              icon={Flame}
              tone="orange"
              label={t("dashboard.currentStreak")}
              value={`${stats.data.current_streak} ${t("dashboard.days")}`}
            />
            <Stat
              icon={Award}
              tone="emerald"
              label={t("dashboard.longestStreak")}
              value={`${stats.data.longest_streak} ${t("dashboard.days")}`}
            />
          </>
        )}
      </div>

      {/* progress summary + badges */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardTitle>{t("dashboard.progress")}</CardTitle>
          {progress.isPending && <div className="mt-3 h-28 animate-pulse rounded bg-muted" />}
          {progress.isError && <p className="mt-2 text-sm text-red-500">{t("common.error")}</p>}
          {progress.data && (
            <ul className="mt-3 flex flex-col gap-2.5 text-sm">
              <ProgressRow icon={Video} label={t("nav.videos")} value={progress.data.videos_completed} />
              <ProgressRow icon={FileText} label={t("nav.articles")} value={progress.data.articles_read} />
              <ProgressRow icon={HelpCircle} label={t("nav.quizzes")} value={progress.data.quizzes_passed} />
              <ProgressRow icon={Code2} label={t("nav.problems")} value={progress.data.problems_solved} />
              <ProgressRow icon={FolderKanban} label={t("nav.projects")} value={progress.data.projects_completed} />
            </ul>
          )}
        </Card>

        <Card>
          <CardTitle>{t("dashboard.badges")}</CardTitle>
          {badges.isPending && <div className="mt-3 h-28 animate-pulse rounded bg-muted" />}
          {badges.isError && <p className="mt-2 text-sm text-red-500">{t("common.error")}</p>}
          {badges.data &&
            (badges.data.badges.length === 0 ? (
              <p className="mt-3 text-sm text-muted-foreground">{t("dashboard.noBadges")}</p>
            ) : (
              <div className="mt-3 flex flex-wrap gap-2">
                {badges.data.badges.map((b) => (
                  <div
                    key={b.code}
                    title={b.description}
                    className="flex items-center gap-2 rounded-lg border bg-muted/30 px-3 py-1.5 text-sm font-medium"
                  >
                    <span className="text-base" aria-hidden>
                      {b.icon}
                    </span>
                    {b.title}
                  </div>
                ))}
              </div>
            ))}
        </Card>
      </div>

      {/* My tracks + saved content */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <div className="flex items-center justify-between">
            <CardTitle>{t("dashboard.myTracks")}</CardTitle>
            <Link to="/tracks" className="text-xs text-primary hover:underline">
              {t("common.all")}
            </Link>
          </div>
          {myTracks.isPending && <div className="mt-3 h-20 animate-pulse rounded bg-muted" />}
          {myTracks.data &&
            (myTracks.data.items.length === 0 ? (
              <p className="mt-3 text-sm text-muted-foreground">{t("dashboard.noTracks")}</p>
            ) : (
              <ul className="mt-3 flex flex-col gap-2">
                {myTracks.data.items.map((tr) => (
                  <li key={tr.id}>
                    <Link
                      to={`/tracks/${tr.id}`}
                      className="flex items-center gap-2.5 rounded-lg border bg-card p-2.5 text-sm transition-colors hover:border-primary/50"
                    >
                      <Route className="size-4 shrink-0 text-amber-500" />
                      <span className="flex-1 truncate font-medium">{tr.title}</span>
                      <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
                    </Link>
                  </li>
                ))}
              </ul>
            ))}
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <CardTitle>{t("dashboard.saved")}</CardTitle>
            <Link to="/bookmarks" className="text-xs text-primary hover:underline">
              {t("common.all")}
            </Link>
          </div>
          {bookmarks.isPending && <div className="mt-3 h-20 animate-pulse rounded bg-muted" />}
          {bookmarks.data &&
            (bookmarks.data.bookmarks.length === 0 ? (
              <p className="mt-3 text-sm text-muted-foreground">{t("dashboard.noSaved")}</p>
            ) : (
              <ul className="mt-3 flex flex-col gap-2">
                {bookmarks.data.bookmarks.slice(0, 6).map((b) => (
                  <li key={b.id}>
                    <Link
                      to={contentPath(b.content_type, b.content_id)}
                      className="flex items-center gap-2.5 rounded-lg border bg-card p-2.5 text-sm transition-colors hover:border-primary/50"
                    >
                      <Bookmark className="size-4 shrink-0 text-primary" />
                      <span className="flex-1 truncate">{b.title || b.content_type}</span>
                      <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
                    </Link>
                  </li>
                ))}
              </ul>
            ))}
        </Card>
      </div>

      {/* Recently completed content */}
      <Card>
        <CardTitle>{t("dashboard.recent")}</CardTitle>
        {recent.isPending && <div className="mt-3 h-20 animate-pulse rounded bg-muted" />}
        {recent.data &&
          (recent.data.items.length === 0 ? (
            <p className="mt-3 text-sm text-muted-foreground">{t("dashboard.noRecent")}</p>
          ) : (
            <ul className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
              {recent.data.items.map((it) => {
                const Icon = recentIcon[it.content_type] ?? FileText
                return (
                  <li key={`${it.content_type}:${it.content_id}`}>
                    <Link
                      to={contentPath(it.content_type, it.content_id)}
                      className="flex items-center gap-2.5 rounded-lg border bg-card p-2.5 text-sm transition-colors hover:border-primary/50"
                    >
                      <Icon className="size-4 shrink-0 text-muted-foreground" />
                      <span className="flex-1 truncate">{it.title}</span>
                      <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
                    </Link>
                  </li>
                )
              })}
            </ul>
          ))}
      </Card>

      {/* activity heatmap */}
      <Card>
        <CardTitle>{t("dashboard.activity")}</CardTitle>
        {activity.isPending && <div className="mt-3 h-24 animate-pulse rounded bg-muted" />}
        {activity.isError && <p className="mt-2 text-sm text-red-500">{t("common.error")}</p>}
        {activity.data && (
          <div className="mt-3">
            <ActivityHeatmap
              from={activity.data.from}
              to={activity.data.to}
              days={activity.data.days}
            />
          </div>
        )}
      </Card>
    </div>
  )
}

const tones = {
  amber: "bg-amber-500/15 text-amber-600 dark:text-amber-400",
  orange: "bg-orange-500/15 text-orange-600 dark:text-orange-400",
  emerald: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400",
  violet: "bg-violet-500/15 text-violet-600 dark:text-violet-400",
}

function Stat({
  icon: Icon,
  tone,
  label,
  value,
  to,
}: {
  icon: LucideIcon
  tone: keyof typeof tones
  label: string
  value: number | string
  to?: string
}) {
  const inner = (
    <>
      <div className={cn("flex size-10 shrink-0 items-center justify-center rounded-xl", tones[tone])}>
        <Icon className="size-5" />
      </div>
      <div className="min-w-0">
        <div className="truncate text-sm text-muted-foreground">{label}</div>
        <div className="text-xl font-semibold">{value}</div>
      </div>
    </>
  )
  if (to) {
    return (
      <Link to={to}>
        <Card className="flex items-center gap-3 transition-all hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-sm">
          {inner}
        </Card>
      </Link>
    )
  }
  return <Card className="flex items-center gap-3">{inner}</Card>
}

function ProgressRow({ icon: Icon, label, value }: { icon: LucideIcon; label: string; value: number }) {
  return (
    <li className="flex items-center gap-2.5">
      <Icon className="size-4 shrink-0 text-muted-foreground" />
      <span className="flex-1 text-muted-foreground">{label}</span>
      <span className="font-semibold tabular-nums">{value}</span>
    </li>
  )
}
