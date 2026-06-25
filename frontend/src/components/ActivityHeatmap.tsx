import { useMemo } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "@/lib/utils"
import type { ActivityDay } from "@/lib/types"

const levelColors = ["bg-muted", "bg-primary/30", "bg-primary/50", "bg-primary/75", "bg-primary"]

function intensity(count: number): number {
  if (count <= 0) return 0
  if (count <= 2) return 1
  if (count <= 5) return 2
  if (count <= 9) return 3
  return 4
}

function parseUTC(s: string): Date {
  return new Date(s + "T00:00:00Z")
}

function dateKey(d: Date): string {
  return d.toISOString().slice(0, 10)
}

type Cell = { key: string | null; count: number }

/** ActivityHeatmap renders a GitHub-style calendar: weeks as columns, days as
 * rows, cells shaded by activity count over the [from, to] window. Month labels
 * run along the top and weekday labels down the left so the dates are visible;
 * each cell also has a localized date + count tooltip. */
export function ActivityHeatmap({ from, to, days }: { from: string; to: string; days: ActivityDay[] }) {
  const { i18n, t } = useTranslation()
  const locale = i18n.resolvedLanguage ?? "en"

  const weeks = useMemo(() => {
    const counts = new Map<string, number>()
    for (const d of days) counts.set(d.day, d.count)

    const fromDate = parseUTC(from)
    const end = parseUTC(to)
    const cursor = parseUTC(from)
    cursor.setUTCDate(cursor.getUTCDate() - cursor.getUTCDay()) // back to Sunday

    const cells: Cell[] = []
    while (cursor <= end) {
      if (cursor < fromDate) {
        cells.push({ key: null, count: 0 })
      } else {
        const key = dateKey(cursor)
        cells.push({ key, count: counts.get(key) ?? 0 })
      }
      cursor.setUTCDate(cursor.getUTCDate() + 1)
    }
    const w: Cell[][] = []
    for (let i = 0; i < cells.length; i += 7) w.push(cells.slice(i, i + 7))
    return w
  }, [from, to, days])

  // A month label appears on the first week-column where that month begins.
  const monthFmt = useMemo(() => new Intl.DateTimeFormat(locale, { month: "short" }), [locale])
  const monthLabels = weeks.map((week, wi) => {
    const dated = week.find((c) => c.key)
    if (!dated) return ""
    const month = parseUTC(dated.key as string).getUTCMonth()
    const prev = weeks[wi - 1]?.find((c) => c.key)
    const prevMonth = prev ? parseUTC(prev.key as string).getUTCMonth() : -1
    return month !== prevMonth ? monthFmt.format(parseUTC(dated.key as string)) : ""
  })

  // Weekday labels (localized) for rows Mon, Wed, Fri.
  const dayFmt = useMemo(() => new Intl.DateTimeFormat(locale, { weekday: "short" }), [locale])
  const weekdayLabels = useMemo(() => {
    const sunday = new Date(Date.UTC(2024, 0, 7)) // a known Sunday
    return [0, 1, 2, 3, 4, 5, 6].map((i) => {
      const d = new Date(sunday)
      d.setUTCDate(d.getUTCDate() + i)
      return i === 1 || i === 3 || i === 5 ? dayFmt.format(d) : ""
    })
  }, [dayFmt])

  const dateFmt = useMemo(
    () => new Intl.DateTimeFormat(locale, { year: "numeric", month: "long", day: "numeric" }),
    [locale],
  )

  return (
    <div className="flex gap-1 overflow-x-auto">
      {/* weekday labels column (aligned below the month-label row) */}
      <div className="mt-[1.1rem] flex shrink-0 flex-col gap-1 pr-1 text-[10px] text-muted-foreground">
        {weekdayLabels.map((label, i) => (
          <div key={i} className="flex h-3 items-center">
            {label}
          </div>
        ))}
      </div>

      <div className="flex flex-col gap-1">
        {/* month labels */}
        <div className="flex gap-1 text-[10px] text-muted-foreground">
          {weeks.map((_, wi) => (
            <div key={wi} className="w-3 shrink-0 whitespace-nowrap">
              {monthLabels[wi]}
            </div>
          ))}
        </div>

        {/* day grid: each week is a column */}
        <div className="flex gap-1">
          {weeks.map((week, wi) => (
            <div key={wi} className="flex flex-col gap-1">
              {week.map((cell, di) => (
                <div
                  key={di}
                  title={
                    cell.key
                      ? `${dateFmt.format(parseUTC(cell.key))} — ${t("dashboard.activityCount", { n: cell.count })}`
                      : undefined
                  }
                  className={cn(
                    "size-3 rounded-sm",
                    cell.key ? levelColors[intensity(cell.count)] : "bg-transparent",
                  )}
                />
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
