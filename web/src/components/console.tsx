import { useState, useRef, useEffect, type FormEvent } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

type LogLine = {
  id: number
  kind: "cmd" | "output" | "error"
  text: string
}

export function Console({ serverId }: { serverId: string }) {
  const [lines, setLines] = useState<LogLine[]>([])
  const [command, setCommand] = useState("")
  const [busy, setBusy] = useState(false)
  const nextId = useRef(0)
  const outputRef = useRef<HTMLDivElement>(null)

  const append = (kind: LogLine["kind"], text: string) => {
    setLines((prev) => [...prev, { id: nextId.current++, kind, text }])
  }

  useEffect(() => {
    outputRef.current?.scrollTo({ top: outputRef.current.scrollHeight })
  }, [lines])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const cmd = command.trim()
    if (!cmd || busy) return

    append("cmd", cmd)
    setCommand("")
    setBusy(true)

    try {
      const res = await fetch(`/api/servers/${serverId}/rcon`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ command: cmd }),
      })
      const body = await res.json()
      if (!res.ok) {
        append("error", body.error || `request failed (${res.status})`)
      } else {
        append("output", body.output.trimEnd() || "(no output)")
      }
    } catch (err) {
      append(
        "error",
        `network error: ${err instanceof Error ? err.message : "unknown"}`
      )
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="overflow-hidden rounded-lg border">
      <div
        ref={outputRef}
        className="h-96 overflow-y-auto bg-muted/50 p-3 font-mono text-xs leading-relaxed"
      >
        {lines.length === 0 ? (
          <div className="text-muted-foreground">
            No commands run yet. Try <code>status</code> or{" "}
            <code>mp_warmup_end</code>.
          </div>
        ) : (
          lines.map((line) => (
            <div key={line.id} className="whitespace-pre-wrap">
              {line.kind === "cmd" && (
                <span>
                  <span className="text-muted-foreground">{"> "}</span>
                  {line.text}
                </span>
              )}
              {line.kind === "output" && (
                <span className="text-foreground/80">{line.text}</span>
              )}
              {line.kind === "error" && (
                <span className="text-destructive">! {line.text}</span>
              )}
            </div>
          ))
        )}
      </div>

      <form
        onSubmit={handleSubmit}
        className="flex gap-2 border-t bg-background p-2"
      >
        <Input
          value={command}
          onChange={(e) => setCommand(e.target.value)}
          placeholder="status, say, mp_restartgame..."
          className="font-mono text-xs"
          disabled={busy}
          autoFocus
        />
        <Button type="submit" disabled={busy || !command.trim()} size="sm">
          {busy ? "sending..." : "send"}
        </Button>
      </form>
    </div>
  )
}
