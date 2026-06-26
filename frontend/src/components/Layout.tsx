import { useState } from "react"
import { Link, NavLink, Outlet, useLocation } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { GraduationCap, Menu, Moon, Sun } from "lucide-react"
import { navGroups } from "@/lib/sections"
import { useAuth } from "@/lib/auth-context"
import { useTheme } from "@/lib/theme"
import { cn } from "@/lib/utils"
import { LanguageSwitcher } from "@/components/LanguageSwitcher"
import { UserMenu } from "@/components/UserMenu"

/** Layout is the app shell: a top bar, a collapsible section sidebar, and the
 * routed page. The sidebar is toggled by the menu button on every screen size;
 * the open/closed choice is remembered. On phones it opens as an overlay. */
export function Layout() {
  const { t } = useTranslation()
  const { theme, toggle } = useTheme()
  const { user } = useAuth()
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(() => {
    const saved = localStorage.getItem("sidebarOpen")
    return saved === null ? true : saved === "true"
  })

  function toggleSidebar() {
    setSidebarOpen((o) => {
      localStorage.setItem("sidebarOpen", String(!o))
      return !o
    })
  }

  // On phones the sidebar is an overlay, so close it after navigating.
  function onNavigate() {
    if (window.matchMedia("(max-width: 767px)").matches) {
      setSidebarOpen(false)
    }
  }

  return (
    <div className="flex min-h-svh flex-col">
      <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b bg-card/80 px-4 backdrop-blur">
        <button
          className="rounded p-2 hover:bg-accent"
          onClick={toggleSidebar}
          aria-label="Toggle navigation"
          aria-expanded={sidebarOpen}
        >
          <Menu className="size-5" />
        </button>
        <Link
          to="/"
          className="flex items-center gap-2 rounded-md font-semibold transition-opacity hover:opacity-80"
          aria-label={t("nav.dashboard")}
        >
          <GraduationCap className="size-6 text-primary" />
          <span>GoAcademy</span>
        </Link>
        <div className="ml-auto flex items-center gap-2">
          <LanguageSwitcher />
          <button className="rounded p-2 hover:bg-accent" onClick={toggle} aria-label="Toggle theme">
            {theme === "dark" ? <Sun className="size-5" /> : <Moon className="size-5" />}
          </button>
          {user && <UserMenu />}
        </div>
      </header>

      <div className="flex flex-1">
        {/* Dim backdrop behind the overlay sidebar on phones. */}
        {sidebarOpen && (
          <div
            className="fixed inset-x-0 top-14 bottom-0 z-10 bg-black/40 md:hidden"
            onClick={toggleSidebar}
            aria-hidden
          />
        )}
        <aside
          className={cn(
            "w-60 shrink-0 border-r bg-card p-3",
            // On md+ the sidebar sticks under the header and scrolls on its own,
            // independently of the routed page.
            "md:sticky md:top-14 md:h-[calc(100dvh-3.5rem)] md:self-start md:overflow-y-auto",
            // Overlay on phones; in-flow column on md+.
            "max-md:fixed max-md:top-14 max-md:bottom-0 max-md:left-0 max-md:z-20 max-md:overflow-y-auto",
            sidebarOpen ? "block" : "hidden",
          )}
        >
          <nav className="flex flex-col gap-4">
            {navGroups.map((group, gi) => {
              const items = group.items.filter((s) => !s.adminOnly || user?.role === "admin")
              if (items.length === 0) return null
              return (
                <div key={group.titleKey ?? gi} className="flex flex-col gap-1">
                  {group.titleKey && (
                    <span className="px-3 pb-1 text-xs font-semibold tracking-wider text-muted-foreground/70 uppercase">
                      {t(group.titleKey)}
                    </span>
                  )}
                  {items.map((s) => {
                    const Icon = s.icon
                    return (
                      <NavLink
                        key={s.path}
                        to={s.path}
                        end={s.path === "/"}
                        onClick={onNavigate}
                        className={({ isActive }) =>
                          cn(
                            "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground",
                            isActive &&
                              "bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground",
                          )
                        }
                      >
                        <Icon className="size-4" />
                        {t(s.labelKey)}
                      </NavLink>
                    )
                  })}
                </div>
              )
            })}
          </nav>
        </aside>

        <main className="flex-1 overflow-x-hidden">
          {/* Keyed on the path so routed content re-plays the enter animation. */}
          <div
            key={location.pathname}
            className="animate-page mx-auto w-full max-w-5xl px-4 py-6 md:px-8 md:py-8"
          >
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
