import { useState, type FormEvent } from "react"
import { Link, useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { useAuth } from "@/lib/auth-context"
import { errorMessage } from "@/lib/api"
import { AuthShell } from "@/components/AuthShell"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const localeOptions = [
  { value: "ru", label: "Русский" },
  { value: "en", label: "English" },
  { value: "uz", label: "Oʻzbekcha" },
  { value: "ja", label: "日本語" },
]

export function Register() {
  const { t } = useTranslation()
  const { register } = useAuth()
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: "", password: "", display_name: "", locale: "en" })
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  function update(key: keyof typeof form) {
    return (e: { target: { value: string } }) => setForm((f) => ({ ...f, [key]: e.target.value }))
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await register(form)
      navigate("/login", { replace: true, state: { registered: true } })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <AuthShell
      title={t("auth.createAccount")}
      footer={
        <>
          {t("auth.haveAccount")}{" "}
          <Link to="/login" className="text-primary hover:underline">
            {t("auth.signIn")}
          </Link>
        </>
      }
    >
      <form onSubmit={onSubmit} className="flex flex-col gap-3">
        <div className="flex flex-col gap-1">
          <Label htmlFor="display_name">{t("auth.displayName")}</Label>
          <Input id="display_name" value={form.display_name} onChange={update("display_name")} required />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="email">{t("auth.email")}</Label>
          <Input id="email" type="email" value={form.email} onChange={update("email")} required autoComplete="email" />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="password">{t("auth.password")}</Label>
          <Input
            id="password"
            type="password"
            value={form.password}
            onChange={update("password")}
            required
            autoComplete="new-password"
          />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="locale">{t("auth.language")}</Label>
          <select
            id="locale"
            value={form.locale}
            onChange={update("locale")}
            className="h-10 rounded-md border bg-transparent px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            {localeOptions.map((l) => (
              <option key={l.value} value={l.value}>
                {l.label}
              </option>
            ))}
          </select>
        </div>
        {error && <p className="text-sm text-red-500">{error}</p>}
        <Button type="submit" disabled={busy}>
          {busy ? t("auth.creating") : t("auth.createAccount")}
        </Button>
      </form>
    </AuthShell>
  )
}
