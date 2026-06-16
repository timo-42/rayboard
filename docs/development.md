# Development

Common commands:

```bash
make test
make build
make build-cross
```

Equivalent direct commands:

```bash
go test ./...
go build -o dist/rayboard ./cmd/rayboard
GOOS=darwin GOARCH=arm64 go build -o dist/rayboard-darwin-arm64 ./cmd/rayboard
GOOS=linux GOARCH=amd64 go build -o dist/rayboard-linux-amd64 ./cmd/rayboard
```

Run locally:

```bash
go run ./cmd/rayboard combined --db ./rayboard.sqlite
go run ./cmd/rayboard backend --db ./rayboard.sqlite
go run ./cmd/rayboard frontend --backend-url http://127.0.0.1:8081
```

## Package Boundaries

The intended request path is:

```text
browser UI -> frontend server -> backend HTTP API -> services -> repositories -> SQLite
```

Current packages:

- `cmd/rayboard`: binary entry point.
- `internal/app`: CLI dispatch and demo seed.
- `internal/config`: defaults, environment, and flags.
- `internal/runtime`: runtime mode orchestration and graceful shutdown.
- `internal/backend`: HTTP handlers.
- `internal/backend/auth`, `authz`, `tracker`, `comments`, `attachments`, `search`: domain services.
- `internal/backend/store` and `migrations`: SQLite open/migrate behavior.
- `internal/frontend`: embedded template/static frontend and API proxy.

## Migration Rules

Migrations are embedded and applied at backend startup. New schema changes should be additive where possible and preserve existing POC data. Keep SQLite foreign keys enabled and avoid migrations that require manual shell access for normal upgrades.

## Release Checklist

- Run `go test ./...`.
- Build the local binary with `make build`.
- For distributable artifacts, run `make build-cross`.
- Start `rayboard combined` against a fresh database and confirm the admin bootstrap password is printed.
- Log in through the frontend and create a project and ticket.
- Run the demo seed against a backend and confirm it completes.
- Update docs for any user-facing CLI, API, auth, frontend, or automation behavior.

