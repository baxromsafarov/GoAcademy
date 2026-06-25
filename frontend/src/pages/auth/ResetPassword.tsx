import { useState, type FormEvent } from "react"
import { Link, useSearchParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { api, errorMessage } from "@/lib/api"
import { AuthShell } from "@/components/AuthShell"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function ResetPassword() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const token = params.get("token") ?? ""
  const [password, setPassword] = useState("")
  const [done, setDone] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await api.post("/auth/reset-password", { token, password }, false)
      setDone(true)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <AuthShell
      title={t("auth.newPasswordTitle")}
      footer={
        <Link to="/login" className="text-primary hover:underline">
          {t("auth.backToSignIn")}
        </Link>
      }
    >
      {done ? (
        <p className="text-sm text-muted-foreground">{t("auth.resetDone")}</p>
      ) : (
        <form onSubmit={onSubmit} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <Label htmlFor="password">{t("auth.newPassword")}</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoComplete="new-password"
            />
          </div>
          {!token && <p className="text-sm text-red-500">{t("auth.missingResetToken")}</p>}
          {error && <p className="text-sm text-red-500">{error}</p>}
          <Button type="submit" disabled={busy || !token}>
            {busy ? t("auth.saving") : t("auth.resetPassword")}
          </Button>
        </form>
      )}
    </AuthShell>
  )
}
