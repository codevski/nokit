import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom"
import { useAuth } from "@/hooks/useAuth"
import { useServers } from "@/hooks/useServers"
import { LoginPage } from "@/pages/login-page"
import { ServerPage } from "@/pages/server-page"
import { Header } from "./components/header"
import { Layout } from "./components/layout"

function RequireAuth({ children }: { children: React.ReactNode }) {
  const auth = useAuth()
  if (auth.status === "loading") return null
  if (auth.status === "unauthenticated") {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function RootRedirect() {
  const state = useServers()
  if (state.status === "loading") return null
  if (state.status === "error")
    return <div className="p-4 text-destructive">{state.message}</div>
  if (state.servers.length === 0)
    return (
      <div className="flex flex-1 flex-col">
        <Header servers={[]} currentId="" onRefresh={state.refresh} />
        <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
          no servers configured - add one using the switcher above
        </div>
      </div>
    )
  return <Navigate to={`/servers/${state.servers[0].id}`} replace />
}
export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <RequireAuth>
                <RootRedirect />
              </RequireAuth>
            }
          />
          <Route
            path="/servers/:id"
            element={
              <RequireAuth>
                <ServerPage />
              </RequireAuth>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
