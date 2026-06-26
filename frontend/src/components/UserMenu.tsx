import { useEffect, useRef, useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ChevronDown, LogOut, Shield, User as UserIcon } from "lucide-react"
import { useAuth } from "@/lib/auth-context"

/** Round avatar: the user's image, or their initials when none is set. */
function Avatar({ url, name, className }: { url: string; name: string; className: string }) {
  if (url) {
    return <img src={url} alt="" className={`${className} rounded-full object-cover`} />
  }
  return (
    <span
      className={`${className} flex items-center justify-center rounded-full bg-primary/15 font-semibold text-primary`}
    >
      {name.slice(0, 2).toUpperCase()}
    </span>
  )
}

/**
 * UserMenu is the header avatar button. Clicking it opens a dropdown with the
 * account summary, a link to the profile/settings page, and sign out.
 */
export function UserMenu() {
  const { t } = useTranslation()
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    function onDoc(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false)
    }
    document.addEventListener("mousedown", onDoc)
    document.addEventListener("keydown", onKey)
    return () => {
      document.removeEventListener("mousedown", onDoc)
      document.removeEventListener("keydown", onKey)
    }
  }, [open])

  if (!user) return null

  async function onLogout() {
    setOpen(false)
    await logout()
    navigate("/login", { replace: true })
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-2 rounded-full border py-0.5 pr-2 pl-0.5 transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={t("nav.profile")}
      >
        <Avatar url={user.avatar_url} name={user.display_name} className="size-8 text-xs" />
        <span className="hidden max-w-32 truncate text-sm font-medium sm:inline">
          {user.display_name}
        </span>
        <ChevronDown
          className={`hidden size-3.5 opacity-60 transition-transform sm:inline ${open ? "rotate-180" : ""}`}
        />
      </button>

      {open && (
        <div
          role="menu"
          className="animate-pop absolute right-0 z-50 mt-2 w-60 overflow-hidden rounded-lg border bg-card text-foreground shadow-lg"
        >
          <div className="flex items-center gap-3 border-b px-3 py-3">
            <Avatar url={user.avatar_url} name={user.display_name} className="size-10 text-sm" />
            <div className="min-w-0">
              <div className="truncate text-sm font-semibold">{user.display_name}</div>
              <div className="truncate text-xs text-muted-foreground">{user.email}</div>
            </div>
          </div>
          <div className="flex flex-col py-1">
            <Link
              to="/profile"
              role="menuitem"
              onClick={() => setOpen(false)}
              className="flex items-center gap-2.5 px-3 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              <UserIcon className="size-4" /> {t("nav.profile")}
            </Link>
            {user.role === "admin" && (
              <Link
                to="/admin"
                role="menuitem"
                onClick={() => setOpen(false)}
                className="flex items-center gap-2.5 px-3 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
              >
                <Shield className="size-4" /> {t("nav.admin")}
              </Link>
            )}
            <button
              type="button"
              role="menuitem"
              onClick={onLogout}
              className="flex items-center gap-2.5 px-3 py-2 text-left text-sm text-red-600 transition-colors hover:bg-accent dark:text-red-400"
            >
              <LogOut className="size-4" /> {t("common.signOut")}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
