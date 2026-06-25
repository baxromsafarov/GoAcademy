import { useCallback, useEffect, useState, type ReactNode } from "react"
import { api, refresh, setAccessToken, setOnAuthFailure } from "@/lib/api"
import { applyProfileLocale } from "@/i18n"
import type { User } from "@/lib/types"
import { AuthContext, type RegisterInput } from "@/lib/auth-context"

interface LoginResponse {
  access_token: string
  user: User
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  // Clear the session if a refresh ultimately fails mid-flight.
  useEffect(() => {
    setOnAuthFailure(() => setUser(null))
    return () => setOnAuthFailure(null)
  }, [])

  // Bootstrap: if a valid refresh cookie exists, restore the session.
  useEffect(() => {
    let cancelled = false
    void (async () => {
      if (await refresh()) {
        try {
          const me = await api.get<User>("/me")
          if (!cancelled) {
            setUser(me)
            applyProfileLocale(me.locale)
          }
        } catch {
          if (!cancelled) setUser(null)
        }
      }
      if (!cancelled) setLoading(false)
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.post<LoginResponse>("/auth/login", { email, password }, false)
    setAccessToken(res.access_token)
    setUser(res.user)
    applyProfileLocale(res.user.locale)
  }, [])

  const register = useCallback(async (input: RegisterInput) => {
    await api.post("/auth/register", input, false)
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.post("/auth/logout", undefined, false)
    } catch {
      // ignore — clear locally regardless
    }
    setAccessToken(null)
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, setUser }}>
      {children}
    </AuthContext.Provider>
  )
}
