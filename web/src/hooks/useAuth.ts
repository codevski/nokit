import { useEffect, useState } from "react"

type AuthState =
  | { status: "loading" }
  | { status: "authenticated" }
  | { status: "unauthenticated" }

export function useAuth(): AuthState {
  const [state, setState] = useState<AuthState>({ status: "loading" })

  useEffect(() => {
    fetch("/api/me")
      .then((r) => {
        if (r.ok) {
          setState({ status: "authenticated" })
        } else {
          setState({ status: "unauthenticated" })
        }
      })
      .catch(() => setState({ status: "unauthenticated" }))
  }, [])

  return state
}
