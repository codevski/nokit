import { Navigate, useParams } from "react-router-dom"
import { useServers } from "@/hooks/useServers"
import { Header } from "@/components/header"
import { Console } from "@/components/console"

export function ServerPage() {
  const { id } = useParams<{ id: string }>()
  const state = useServers()

  if (state.status === "loading") return null
  if (state.status === "error")
    return <div className="p-4 text-destructive">{state.message}</div>

  if (state.servers.length === 0) return <Navigate to="/" replace />

  const known = state.servers.some((s) => s.id === id)
  if (!known) return <Navigate to="/" replace />

  return (
    <div className="flex flex-1 flex-col">
      <Header
        servers={state.servers}
        currentId={id!}
        onRefresh={state.refresh}
      />
      <main className="mx-auto w-full max-w-5xl flex-1 p-4">
        <Console serverId={id!} />
      </main>
    </div>
  )
}
