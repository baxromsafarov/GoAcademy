import { useEffect, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { api, errorMessage } from "@/lib/api"
import { AuthShell } from "@/components/AuthShell"

export function VerifyEmail() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const token = params.get("token")
  const [message, setMessage] = useState(t("auth.verifying"))

  useEffect(() => {
    if (!token) {
      setMessage(t("auth.missingToken"))
      return
    }
    void (async () => {
      try {
        await api.post("/auth/verify-email", { token }, false)
        setMessage(t("auth.verified"))
      } catch (err) {
        setMessage(errorMessage(err))
      }
    })()
  }, [token, t])

  return (
    <AuthShell
      title={t("auth.verifyTitle")}
      footer={
        <Link to="/login" className="text-primary hover:underline">
          {t("auth.goToSignIn")}
        </Link>
      }
    >
      <p className="text-sm text-muted-foreground">{message}</p>
    </AuthShell>
  )
}
