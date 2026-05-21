export function Footer() {
  return (
    <footer className="border-t border-border px-4 py-3">
      <div className="mx-auto flex max-w-5xl items-center justify-between">
        <span className="font-mono text-xs text-muted-foreground">nokit</span>
        {/* TODO: pull from /api/version */}
        <span className="font-mono text-xs text-muted-foreground">v0.1.0</span>
      </div>
    </footer>
  )
}
