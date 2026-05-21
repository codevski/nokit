import { useRef, useState, useEffect, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import type { ServerInfo } from "@/hooks/useServers"

type Props = {
  servers: ServerInfo[]
  currentId: string
  onRefresh: () => void
}

export function ServerSwitcher({ servers, currentId, onRefresh }: Props) {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [adding, setAdding] = useState(false)
  const [name, setName] = useState("")
  const [host, setHost] = useState("")
  const [pass, setPass] = useState("")
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState("")
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
        setAdding(false)
      }
    }
    document.addEventListener("mousedown", handler)
    return () => document.removeEventListener("mousedown", handler)
  }, [open])

  const current = servers.find((s) => s.id === currentId)

  const handleSelect = (id: string) => {
    setOpen(false)
    setAdding(false)
    navigate(`/servers/${id}`)
  }

  const handleRemove = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    if (!confirm(`Remove server "${id}"?`)) return
    await fetch(`/api/servers/${id}`, { method: "DELETE" })
    onRefresh()
    if (id === currentId && servers.length > 1) {
      const next = servers.find((s) => s.id !== id)
      if (next) navigate(`/servers/${next.id}`)
    }
  }

  const handleAdd = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setErr("")
    try {
      const r = await fetch("/api/servers", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, host, rcon_pass: pass }),
      })
      const body = await r.json()
      if (!r.ok) {
        setErr(body.error ?? "failed to add server")
        return
      }
      setName("")
      setHost("")
      setPass("")
      setAdding(false)
      onRefresh()
      navigate(`/servers/${body.id}`)
    } catch {
      setErr("network error")
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex h-8 items-center gap-2 rounded-md border border-border px-3 text-sm transition-colors hover:bg-accent"
      >
        <span className="h-2 w-2 flex-shrink-0 rounded-full bg-green-500" />
        <span className="max-w-[140px] truncate font-medium">
          {current?.name ?? currentId}
        </span>
        <svg
          className={`h-3.5 w-3.5 text-muted-foreground transition-transform ${open ? "rotate-180" : ""}`}
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path d="M4 6l4 4 4-4" />
        </svg>
      </button>

      {open && (
        <div className="absolute top-10 left-0 z-50 w-72 overflow-hidden rounded-lg border border-border bg-background shadow-lg">
          <div className="py-1">
            <p className="px-3 py-1.5 text-[11px] tracking-wider text-muted-foreground uppercase">
              servers
            </p>
            {servers.map((s) => (
              <div
                key={s.id}
                onClick={() => handleSelect(s.id)}
                className={`flex cursor-pointer items-center gap-3 px-3 py-2.5 transition-colors hover:bg-accent ${
                  s.id === currentId ? "bg-accent/60" : ""
                }`}
              >
                <span className="h-2 w-2 flex-shrink-0 rounded-full bg-green-500" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{s.name}</p>
                  <p className="truncate font-mono text-[11px] text-muted-foreground">
                    {s.host}
                  </p>
                </div>
                <button
                  onClick={(e) => handleRemove(e, s.id)}
                  className="rounded p-1 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                  title={`remove ${s.name}`}
                >
                  <svg
                    viewBox="0 0 16 16"
                    className="h-3.5 w-3.5"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="1.5"
                  >
                    <path d="M2 4h12M5 4V2h6v2M6 7v5M10 7v5M3 4l1 9h8l1-9" />
                  </svg>
                </button>
              </div>
            ))}
          </div>

          <div className="border-t border-border" />

          {!adding ? (
            <button
              onClick={() => setAdding(true)}
              className="flex w-full items-center gap-2 px-3 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            >
              <svg
                viewBox="0 0 16 16"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
              >
                <path d="M8 2v12M2 8h12" />
              </svg>
              add server
            </button>
          ) : (
            <form onSubmit={handleAdd} className="space-y-2 p-3">
              <p className="mb-2 text-[11px] tracking-wider text-muted-foreground uppercase">
                add server
              </p>
              <input
                autoFocus
                placeholder="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="h-8 w-full rounded border border-input bg-background px-2.5 font-sans text-sm outline-none focus:ring-1 focus:ring-ring"
                required
              />
              <input
                placeholder="host:port"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                className="h-8 w-full rounded border border-input bg-background px-2.5 font-mono text-sm outline-none focus:ring-1 focus:ring-ring"
                required
              />
              <input
                type="password"
                placeholder="rcon password"
                value={pass}
                onChange={(e) => setPass(e.target.value)}
                className="h-8 w-full rounded border border-input bg-background px-2.5 font-mono text-sm outline-none focus:ring-1 focus:ring-ring"
                required
              />
              {err && <p className="text-xs text-destructive">{err}</p>}
              <div className="flex gap-2 pt-1">
                <button
                  type="button"
                  onClick={() => {
                    setAdding(false)
                    setErr("")
                  }}
                  className="h-7 flex-1 rounded border border-border text-xs text-muted-foreground hover:bg-accent"
                >
                  cancel
                </button>
                <button
                  type="submit"
                  disabled={busy}
                  className="h-7 flex-1 rounded bg-foreground text-xs font-medium text-background disabled:opacity-50"
                >
                  {busy ? "adding…" : "add"}
                </button>
              </div>
            </form>
          )}
        </div>
      )}
    </div>
  )
}
