import { Navigate, Outlet } from "react-router-dom"
import { useAuth } from "@/lib/auth-context"

/** AdminRoute gates the admin section: only users with the admin role pass;
 * everyone else is redirected to the dashboard. The backend independently
 * enforces RequireRole('admin') on every /admin/* endpoint. */
export function AdminRoute() {
  const { user, loading } = useAuth()
  if (loading) return null
  if (!user || user.role !== "admin") return <Navigate to="/" replace />
  return <Outlet />
}
