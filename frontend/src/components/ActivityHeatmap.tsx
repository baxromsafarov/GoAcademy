import { useMemo } from "react"
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

/** ActivityHeatmap renders a GitHub-style calendar: weeks as columns, days as
 * rows, cells shaded by activity count over the [from, to] window. */
export function ActivityHeatmap({ from, to, days }: { from: string; to: string; days: ActivityDay[] }) {
  const cells = useMemo(() => {
    const counts = new Map<string, number>()
    for (const d of days) counts.set(d.day, d.count)

    const fromDate = parseUTC(from)
    const end = parseUTC(to)
    // Start at the Sunday on or before `from` so columns align to weekdays.
    const cursor = parseUTC(from)
    cursor.setUTCDate(cursor.getUTCDate() - cursor.getUTCDay())

    const out: { key: string | null; count: number }[] = []
    while (cursor <= end) {
      if (cursor < fromDate) {
        out.push({ key: null, count: 0 })
      } else {
        const key = dateKey(cursor)
        out.push({ key, count: counts.get(key) ?? 0 })
      }
      cursor.setUTCDate(cursor.getUTCDate() + 1)
    }
    return out
  }, [from, to, days])

  return (
    <div className="overflow-x-auto">
      <div
        className="inline-grid gap-1"
        style={{ gridTemplateRows: "repeat(7, 0.75rem)", gridAutoFlow: "column" }}
      >
        {cells.map((cell, i) => (
          <div
            key={i}
            title={cell.key ? `${cell.key}: ${cell.count}` : undefined}
            className={cn(
              "size-3 rounded-sm",
              cell.key ? levelColors[intensity(cell.count)] : "bg-transparent",
            )}
          />
        ))}
      </div>
    </div>
  )
}
