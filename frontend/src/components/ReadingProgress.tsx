import { useEffect, useState } from "react"

/**
 * ReadingProgress is the thin accent bar pinned under the header that fills as
 * the reader scrolls a long page — the same affordance Practicum uses on its
 * lesson screens. It tracks the document scroll (the app's scroll container).
 */
export function ReadingProgress() {
  const [pct, setPct] = useState(0)

  useEffect(() => {
    function update() {
      const el = document.documentElement
      const scrollable = el.scrollHeight - el.clientHeight
      setPct(scrollable > 0 ? Math.min(100, (el.scrollTop / scrollable) * 100) : 0)
    }
    update()
    window.addEventListener("scroll", update, { passive: true })
    window.addEventListener("resize", update)
    return () => {
      window.removeEventListener("scroll", update)
      window.removeEventListener("resize", update)
    }
  }, [])

  return (
    <div className="fixed inset-x-0 top-14 z-20 h-0.5 bg-transparent" aria-hidden>
      <div
        className="h-full bg-primary transition-[width] duration-75 ease-out"
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}
