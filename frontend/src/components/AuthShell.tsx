import type { ReactNode } from "react"
import { GraduationCap } from "lucide-react"

/** AuthShell centers a card for the unauthenticated auth screens. */
export function AuthShell({
  title,
  children,
  footer,
}: {
  title: string
  children: ReactNode
  footer?: ReactNode
}) {
  return (
    <div className="flex min-h-svh items-center justify-center p-4">
      <div className="w-full max-w-sm rounded-lg border bg-card p-6">
        <div className="mb-1 flex items-center gap-2 font-semibold">
          <GraduationCap className="size-5 text-primary" />
          GoAcademy
        </div>
        <h1 className="mb-4 text-xl font-semibold tracking-tight">{title}</h1>
        {children}
        {footer && <div className="mt-4 text-center text-sm text-muted-foreground">{footer}</div>}
      </div>
    </div>
  )
}
