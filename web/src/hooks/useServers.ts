import { useCallback, useEffect, useState } from "react"

export type ServerInfo = {
  id: string
  name: string
  host: string
}

type State =
  | { status: "loading" }
  | { status: "ready"; servers: ServerInfo[] }
  | { status: "error"; message: string }

export function useServers(): State & { refresh: () => void } {
  const [state, setState] = useState<State>({ status: "loading" })
  const [tick, setTick] = useState(0)

  const refresh = useCallback(() => setTick((t) => t + 1), [])

  useEffect(() => {
    let cancelled = false
    fetch("/api/servers")
      .then(async (r) => {
        if (!r.ok) throw new Error(`servers ${r.status}`)
        return (await r.json()) as ServerInfo[]
      })
      .then((servers) => {
        if (!cancelled) setState({ status: "ready", servers })
      })
      .catch((err: Error) => {
        if (!cancelled) setState({ status: "error", message: err.message })
      })
    return () => {
      cancelled = true
    }
  }, [tick])

  return { ...state, refresh }
}
