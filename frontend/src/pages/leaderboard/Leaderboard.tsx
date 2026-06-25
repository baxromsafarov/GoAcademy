import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Trophy } from "lucide-react"
import { useLeaderboard } from "@/lib/queries"
import { useAuth } from "@/lib/auth-context"
import { cn } from "@/lib/utils"

const periods = ["all", "week", "month"] as const

export function Leaderboard() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const [period, setPeriod] = useState<(typeof periods)[number]>("all")
  const { data, isPending, isError } = useLeaderboard(period)

  return (
    <div className="flex flex-col gap-4">
      <h1 className="flex items-center gap-2 text-2xl font-semibold tracking-tight">
        <Trophy className="size-6" /> {t("nav.leaderboard")}
      </h1>

      <div className="flex gap-1 rounded-md border p-1 w-fit">
        {periods.map((p) => (
          <button
            key={p}
            type="button"
            onClick={() => setPeriod(p)}
            className={cn(
              "rounded px-3 py-1 text-sm transition-colors",
              period === p ? "bg-primary text-primary-foreground" : "hover:bg-accent",
            )}
          >
            {t(`leaderboard.${p}`)}
          </button>
        ))}
      </div>

      {isPending && (
        <div className="flex flex-col gap-2">
          {[0, 1, 2, 3, 4].map((i) => (
            <div key={i} className="h-12 animate-pulse rounded-lg border bg-card" />
          ))}
        </div>
      )}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data &&
        (data.entries.length === 0 ? (
          <p className="text-muted-foreground">{t("leaderboard.empty")}</p>
        ) : (
          <ol className="flex flex-col gap-2">
            {data.entries.map((e) => {
              const isMe = user?.id === e.user_id
              return (
                <li
                  key={e.user_id}
                  className={cn(
                    "flex items-center gap-3 rounded-lg border bg-card p-3",
                    isMe && "border-primary ring-1 ring-primary",
                  )}
                >
                  <span className="w-8 text-center font-semibold tabular-nums text-muted-foreground">
                    {e.rank}
                  </span>
                  {e.avatar_url ? (
                    <img src={e.avatar_url} alt="" className="size-8 rounded-full object-cover" />
                  ) : (
                    <span className="flex size-8 items-center justify-center rounded-full bg-muted text-xs font-medium">
                      {e.display_name.slice(0, 2).toUpperCase()}
                    </span>
                  )}
                  <span className="flex-1 font-medium">
                    {e.display_name}
                    {isMe && <span className="ml-2 text-xs text-muted-foreground">({t("leaderboard.you")})</span>}
                  </span>
                  <span className="font-semibold tabular-nums">{e.xp} XP</span>
                </li>
              )
            })}
          </ol>
        ))}
    </div>
  )
}
