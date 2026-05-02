# Defuse

Self-hosted web admin panel for Counter-Strike 2 dedicated servers.

RCON console, live log streaming, player management, map control, and multi-server support — all in a single Go binary you can run alongside your CS2 container.

> ⚠️ Early development. Not yet production-ready.

## Features

- **RCON console** with autocomplete for commands and CVARs
- **Live logs** streamed over Server-Sent Events
- **Player management** — search, kick, SteamID resolution
- **Map control** — standard maps, workshop IDs, browser-stored favorites
- **CVAR presets** — pill buttons for common server configs
- **Multi-server** support with header dropdown switching
- **Auth** — sessions by default, optional reverse-proxy SSO pass-through

## Screenshots

_Coming soon._

## Requirements

- A CS2 dedicated server reachable over RCON (typically a Docker container which I have used joedwards32/cs2)
- Go 1.22+ (for building from source)
- Bun 1.x (for building the frontend from source)

## Quick start

```bash
git clone https://github.com/codevski/defuse
cd defuse
cp .env.example .env       # set PANEL_PASSWORD, RCON_PASSWORD, STEAM_GSLT
docker compose up -d
```

Open `http://localhost:8080`.

## Configuration

Servers are configured in `config.yml`. Secrets come from environment variables.

See [`config.example.yml`](./config.example.yaml) and [`.env.example`](./.env.example).

## License

MIT
