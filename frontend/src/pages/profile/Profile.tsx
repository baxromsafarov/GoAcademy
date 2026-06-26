import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Upload } from "lucide-react"
import { useAuth } from "@/lib/auth-context"
import { useUpdateProfile, useUploadAvatar } from "@/lib/queries"
import { setLang, type Lang } from "@/i18n"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select } from "@/components/ui/select"

const locales = ["ru", "en", "uz", "ja"]
const localeOptions = locales.map((l) => ({ value: l, label: l.toUpperCase() }))

export function Profile() {
  const { t } = useTranslation()
  const { user, setUser } = useAuth()
  const update = useUpdateProfile()
  const avatar = useUploadAvatar()
  const fileRef = useRef<HTMLInputElement>(null)
  const [saved, setSaved] = useState(false)

  const [form, setForm] = useState({
    display_name: user?.display_name ?? "",
    bio: user?.bio ?? "",
    location: user?.location ?? "",
    locale: user?.locale ?? "en",
    is_public: user?.is_public ?? false,
  })

  if (!user) return null

  function set<K extends keyof typeof form>(key: K, value: (typeof form)[K]) {
    setForm((f) => ({ ...f, [key]: value }))
    setSaved(false)
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    update.mutate(form, {
      onSuccess: (updated) => {
        setUser(updated)
        setLang(updated.locale as Lang) // saving the profile is an explicit language choice
        setSaved(true)
      },
    })
  }

  function onAvatarChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    avatar.mutate(file, { onSuccess: (updated) => setUser(updated) })
  }

  return (
    <div className="flex max-w-xl flex-col gap-6">
      <h1 className="text-2xl font-semibold tracking-tight">{t("nav.profile")}</h1>

      {/* Avatar */}
      <div className="flex items-center gap-4">
        {user.avatar_url ? (
          <img src={user.avatar_url} alt="" className="size-16 rounded-full object-cover" />
        ) : (
          <span className="flex size-16 items-center justify-center rounded-full bg-muted text-lg font-medium">
            {user.display_name.slice(0, 2).toUpperCase()}
          </span>
        )}
        <div className="flex flex-col gap-1">
          <input
            ref={fileRef}
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            onChange={onAvatarChange}
            className="hidden"
          />
          <Button
            type="button"
            variant="outline"
            onClick={() => fileRef.current?.click()}
            disabled={avatar.isPending}
          >
            <Upload className="size-4" /> {avatar.isPending ? t("profile.uploading") : t("profile.changeAvatar")}
          </Button>
          {avatar.isError && <span className="text-xs text-red-500">{t("profile.avatarError")}</span>}
        </div>
      </div>

      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="email">{t("auth.email")}</Label>
          <Input id="email" value={user.email} disabled />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="display_name">{t("auth.displayName")}</Label>
          <Input
            id="display_name"
            value={form.display_name}
            onChange={(e) => set("display_name", e.target.value)}
            required
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="bio">{t("profile.bio")}</Label>
          <textarea
            id="bio"
            value={form.bio}
            onChange={(e) => set("bio", e.target.value)}
            rows={3}
            className="w-full resize-y rounded-md border bg-transparent p-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="location">{t("profile.location")}</Label>
          <Input
            id="location"
            value={form.location}
            onChange={(e) => set("location", e.target.value)}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="locale">{t("auth.language")}</Label>
          <Select
            value={form.locale}
            onChange={(v) => set("locale", v)}
            options={localeOptions}
            ariaLabel={t("auth.language")}
            className="w-40"
          />
        </div>

        <label className="flex w-fit cursor-pointer items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={form.is_public}
            onChange={(e) => set("is_public", e.target.checked)}
            className="size-4 accent-primary"
          />
          {t("profile.isPublic")}
        </label>

        <div className="flex items-center gap-3">
          <Button type="submit" disabled={update.isPending}>
            {update.isPending ? t("profile.saving") : t("profile.save")}
          </Button>
          {saved && <span className="text-sm text-green-600 dark:text-green-400">{t("profile.saved")}</span>}
          {update.isError && <span className="text-sm text-red-500">{t("common.error")}</span>}
        </div>
      </form>
    </div>
  )
}
