import { useNavigate } from "react-router-dom"
import type { ServerInfo } from "@/hooks/useServers"
import { ServerSwitcher } from "./server-switcher"

type Props = {
  servers: ServerInfo[]
  currentId: string
  onRefresh: () => void
}

export function Header({ servers, currentId, onRefresh }: Props) {
  const navigate = useNavigate()

  const handleLogout = async () => {
    await fetch("/api/logout", { method: "POST" })
    navigate("/login", { replace: true })
  }

  return (
    <header className="border-b border-border bg-background">
      <div className="mx-auto flex h-13 max-w-5xl items-center gap-3 px-4 py-3">
        <div className="flex-shrink-0 font-mono text-lg font-semibold tracking-tight">
          nokit
        </div>

        <div className="h-4 w-px flex-shrink-0 bg-border" />

        <ServerSwitcher
          servers={servers}
          currentId={currentId}
          onRefresh={onRefresh}
        />

        <div className="ml-auto">
          <button
            onClick={handleLogout}
            className="flex h-7 items-center gap-1.5 rounded border border-border px-2.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          >
            <svg
              viewBox="0 0 16 16"
              className="h-3.5 w-3.5"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="M6 2H3a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3M10 11l3-3-3-3M5 8h8" />
            </svg>
            sign out
          </button>
        </div>
      </div>
    </header>
  )
}
