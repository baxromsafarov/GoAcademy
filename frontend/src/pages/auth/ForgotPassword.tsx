import { useState, type FormEvent } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { api } from "@/lib/api"
import { AuthShell } from "@/components/AuthShell"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function ForgotPassword() {
  const { t } = useTranslation()
  const [email, setEmail] = useState("")
  const [sent, setSent] = useState(false)
  const [busy, setBusy] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    try {
      // The endpoint always succeeds (silent on unknown emails).
      await api.post("/auth/forgot-password", { email }, false)
      setSent(true)
    } finally {
      setBusy(false)
    }
  }

  return (
    <AuthShell
      title={t("auth.resetTitle")}
      footer={
        <Link to="/login" className="text-primary hover:underline">
          {t("auth.backToSignIn")}
        </Link>
      }
    >
      {sent ? (
        <p className="text-sm text-muted-foreground">{t("auth.resetSent")}</p>
      ) : (
        <form onSubmit={onSubmit} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <Label htmlFor="email">{t("auth.email")}</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
            />
          </div>
          <Button type="submit" disabled={busy}>
            {busy ? t("auth.sending") : t("auth.sendResetLink")}
          </Button>
        </form>
      )}
    </AuthShell>
  )
}
