import type { ButtonHTMLAttributes } from "react"
import { cn } from "@/lib/utils"

type Variant = "default" | "outline" | "ghost"
type Size = "sm" | "default" | "lg"

const variants: Record<Variant, string> = {
  default: "bg-primary text-primary-foreground hover:opacity-90",
  outline: "border bg-transparent hover:bg-accent hover:text-accent-foreground",
  ghost: "hover:bg-accent hover:text-accent-foreground",
}

const sizes: Record<Size, string> = {
  sm: "h-8 gap-1.5 px-3 text-sm",
  default: "h-10 gap-2 px-4 text-sm",
  lg: "h-11 gap-2 px-6 text-base",
}

export function Button({
  className,
  variant = "default",
  size = "default",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant; size?: Size }) {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50",
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    />
  )
}
