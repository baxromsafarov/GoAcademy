import { useState } from "react"
import { NavLink, Outlet, useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { GraduationCap, LogOut, Menu, Moon, Sun } from "lucide-react"
import { sections } from "@/lib/sections"
import { useAuth } from "@/lib/auth-context"
import { useTheme } from "@/lib/theme"
import { cn } from "@/lib/utils"
import { LanguageSwitcher } from "@/components/LanguageSwitcher"

/** Layout is the app shell: a top bar, a section sidebar, and the routed page. */
export function Layout() {
  const { t } = useTranslation()
  const { theme, toggle } = useTheme()
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)

  async function onLogout() {
    await logout()
    navigate("/login", { replace: true })
  }

  return (
    <div className="flex min-h-svh flex-col">
      <header className="sticky top-0 z-20 flex h-14 items-center gap-3 border-b bg-card px-4">
        <button
          className="rounded p-2 hover:bg-accent md:hidden"
          onClick={() => setMenuOpen((o) => !o)}
          aria-label="Toggle navigation"
        >
          <Menu className="size-5" />
        </button>
        <div className="flex items-center gap-2 font-semibold">
          <GraduationCap className="size-6 text-primary" />
          <span>GoAcademy</span>
        </div>
        <div className="ml-auto flex items-center gap-2">
          {user && (
            <span className="hidden text-sm text-muted-foreground sm:inline">{user.display_name}</span>
          )}
          <LanguageSwitcher />
          <button className="rounded p-2 hover:bg-accent" onClick={toggle} aria-label="Toggle theme">
            {theme === "dark" ? <Sun className="size-5" /> : <Moon className="size-5" />}
          </button>
          <button className="rounded p-2 hover:bg-accent" onClick={onLogout} aria-label={t("common.signOut")}>
            <LogOut className="size-5" />
          </button>
        </div>
      </header>

      <div className="flex flex-1">
        <aside
          className={cn(
            "w-60 shrink-0 border-r bg-card p-3 md:block",
            menuOpen ? "block" : "hidden",
          )}
        >
          <nav className="flex flex-col gap-1">
            {sections
              .filter((s) => !s.adminOnly || user?.role === "admin")
              .map((s) => {
              const Icon = s.icon
              return (
                <NavLink
                  key={s.path}
                  to={s.path}
                  end={s.path === "/"}
                  onClick={() => setMenuOpen(false)}
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
          </nav>
        </aside>

        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
