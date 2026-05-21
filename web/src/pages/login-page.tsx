import { useState, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"

export function LoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setError("")
    try {
      const r = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      })
      if (r.ok) {
        navigate("/", { replace: true })
      } else {
        const body = await r.json().catch(() => ({}))
        setError(body.error ?? "invalid credentials")
      }
    } catch {
      setError("could not reach server")
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-sm">
        <div className="mb-8">
          <h1 className="font-mono text-2xl font-semibold tracking-tight">
            defuse
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">cs2 server panel</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-sm text-muted-foreground" htmlFor="username">
              username
            </label>
            <input
              id="username"
              type="text"
              autoComplete="username"
              autoFocus
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-1 focus:ring-ring"
              required
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-sm text-muted-foreground" htmlFor="password">
              password
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-1 focus:ring-ring"
              required
            />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <button
            type="submit"
            disabled={busy}
            className="h-9 w-full rounded-md bg-foreground text-sm font-medium text-background disabled:opacity-50"
          >
            {busy ? "signing in…" : "sign in"}
          </button>
        </form>
      </div>
    </div>
  )
}
