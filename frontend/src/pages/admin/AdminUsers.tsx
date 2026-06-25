import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Search, ShieldCheck, Shield, Ban, CircleCheck } from "lucide-react"
import { useAdminUsers, useUpdateAdminUser } from "@/lib/queries"
import { useAuth } from "@/lib/auth-context"

const PAGE = 20

export function AdminUsers() {
  const { t } = useTranslation()
  const { user: me } = useAuth()
  const [q, setQ] = useState("")
  const [offset, setOffset] = useState(0)
  const { data, isPending, isError } = useAdminUsers({ q: q || undefined, limit: PAGE, offset })
  const update = useUpdateAdminUser()

  const total = data?.total ?? 0

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-semibold tracking-tight">{t("admin.users")}</h1>

      <div className="relative w-fit">
        <Search className="absolute top-1/2 left-2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          value={q}
          onChange={(e) => {
            setQ(e.target.value)
            setOffset(0)
          }}
          placeholder={t("common.search")}
          className="h-9 rounded-md border bg-transparent pr-2 pl-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </div>

      {isPending && <div className="h-40 animate-pulse rounded-lg border bg-card" />}
      {isError && <p className="text-sm text-red-500">{t("common.error")}</p>}
      {data && (
        <>
          <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-left">
                <tr>
                  <th className="px-3 py-2 font-medium">{t("admin.uUser")}</th>
                  <th className="px-3 py-2 font-medium">{t("admin.uRole")}</th>
                  <th className="px-3 py-2 font-medium">{t("admin.uStatus")}</th>
                  <th className="px-3 py-2 text-right font-medium">{t("admin.uActions")}</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((u) => {
                  const isSelf = me?.id === u.id
                  const busy = update.isPending
                  return (
                    <tr key={u.id} className="border-t">
                      <td className="px-3 py-2">
                        <div className="font-medium">{u.display_name}</div>
                        <div className="text-xs text-muted-foreground">{u.email}</div>
                      </td>
                      <td className="px-3 py-2">
                        <span className="rounded border px-1.5 py-0.5 text-xs">{u.role}</span>
                      </td>
                      <td className="px-3 py-2">
                        {u.is_blocked ? (
                          <span className="text-xs text-red-500">{t("admin.uBlocked")}</span>
                        ) : (
                          <span className="text-xs text-green-600 dark:text-green-400">
                            {t("admin.uActive")}
                          </span>
                        )}
                      </td>
                      <td className="px-3 py-2">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            type="button"
                            disabled={isSelf || busy}
                            title={t("admin.toggleRole")}
                            onClick={() =>
                              update.mutate({ id: u.id, role: u.role === "admin" ? "student" : "admin" })
                            }
                            className="rounded p-1 text-muted-foreground hover:bg-accent disabled:opacity-40"
                          >
                            {u.role === "admin" ? (
                              <ShieldCheck className="size-4" />
                            ) : (
                              <Shield className="size-4" />
                            )}
                          </button>
                          <button
                            type="button"
                            disabled={isSelf || busy}
                            title={t("admin.toggleBlock")}
                            onClick={() => update.mutate({ id: u.id, is_blocked: !u.is_blocked })}
                            className="rounded p-1 text-muted-foreground hover:bg-accent disabled:opacity-40"
                          >
                            {u.is_blocked ? (
                              <CircleCheck className="size-4" />
                            ) : (
                              <Ban className="size-4" />
                            )}
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">
              {t("admin.uTotal", { total })}
            </span>
            <div className="flex gap-2">
              <button
                type="button"
                disabled={offset === 0}
                onClick={() => setOffset(Math.max(0, offset - PAGE))}
                className="rounded-md border px-3 py-1 disabled:opacity-40"
              >
                {t("admin.prev")}
              </button>
              <button
                type="button"
                disabled={offset + PAGE >= total}
                onClick={() => setOffset(offset + PAGE)}
                className="rounded-md border px-3 py-1 disabled:opacity-40"
              >
                {t("admin.next")}
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
