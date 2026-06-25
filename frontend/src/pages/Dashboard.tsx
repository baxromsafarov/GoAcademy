import { useTranslation } from "react-i18next"
import { Card, CardTitle } from "@/components/ui/card"
import { ActivityHeatmap } from "@/components/ActivityHeatmap"
import { useActivity, useBadges, useProgressSummary, useStats } from "@/lib/queries"

export function Dashboard() {
  const { t } = useTranslation()
  const stats = useStats()
  const progress = useProgressSummary()
  const badges = useBadges()
  const activity = useActivity()

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t("dashboard.title")}</h1>
        <p className="mt-1 text-muted-foreground">{t("dashboard.welcome")}</p>
      </div>

      {/* XP / level / streaks */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.isPending &&
          [0, 1, 2, 3].map((i) => <Card key={i} className="h-[5.5rem] animate-pulse" />)}
        {stats.isError && (
          <p className="col-span-full text-sm text-red-500">{t("common.error")}</p>
        )}
        {stats.data && (
          <>
            <Stat label={t("dashboard.xp")} value={stats.data.total_xp} />
            <Stat label={t("dashboard.level")} value={stats.data.level} />
            <Stat
              label={t("dashboard.currentStreak")}
              value={`${stats.data.current_streak} ${t("dashboard.days")}`}
            />
            <Stat
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
            <ul className="mt-3 flex flex-col gap-2 text-sm">
              <ProgressRow label={t("nav.videos")} value={progress.data.videos_completed} />
              <ProgressRow label={t("nav.articles")} value={progress.data.articles_read} />
              <ProgressRow label={t("nav.quizzes")} value={progress.data.quizzes_passed} />
              <ProgressRow label={t("nav.problems")} value={progress.data.problems_solved} />
              <ProgressRow label={t("nav.projects")} value={progress.data.projects_completed} />
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
                    className="flex items-center gap-1 rounded-md border px-2 py-1 text-sm"
                  >
                    <span aria-hidden>{b.icon}</span>
                    {b.title}
                  </div>
                ))}
              </div>
            ))}
        </Card>
      </div>

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

function Stat({ label, value }: { label: string; value: number | string }) {
  return (
    <Card>
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-1 text-2xl font-semibold">{value}</div>
    </Card>
  )
}

function ProgressRow({ label, value }: { label: string; value: number }) {
  return (
    <li className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </li>
  )
}
