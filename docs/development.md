# Development

Common commands:

```bash
make test
make verify-docs
make build
make build-cross
go run ./cmd/rayboard verify docs
```

Equivalent direct commands:

```bash
go test ./...
go build -o dist/rayboard ./cmd/rayboard
GOOS=darwin GOARCH=arm64 go build -o dist/rayboard-darwin-arm64 ./cmd/rayboard
GOOS=linux GOARCH=amd64 go build -o dist/rayboard-linux-amd64 ./cmd/rayboard
go run ./cmd/rayboard verify docs
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
- `internal/backend`: public backend server facade and option wiring.
- `internal/backend/httpapi`: Huma router, shared HTTP helpers, and resource route packages.
- `internal/backend/auth`, `authz`, `tracker`, `comments`, `attachments`, `search`: domain services.
- `internal/backend/store` and `migrations`: SQLite open/migrate behavior.
- `internal/frontend`: embedded template/static frontend and API proxy.

Backend HTTP route packages are resource-focused. Each package keeps `routes.go` for route registration/thin handlers, `schema.go` for Huma/OpenAPI request and response DTOs, and `provider.go` for dependency wiring. Huma DTOs are the source of the generated OpenAPI request/response body schemas served by the binary.

New JSON DTOs should follow the Rayboard resource object convention: create/update/action `Input` bodies contain `spec`, and JSON `Output` bodies contain `metadata`, `spec`, and `status` for resources or resource-like computed views. List outputs use the same envelope with `metadata.count` and `status.items`; each item should use the same resource object shape when it represents API state.

## Migration Rules

Migrations are embedded and applied at backend startup. New schema changes should be additive where possible and preserve existing POC data. Keep SQLite foreign keys enabled and avoid migrations that require manual shell access for normal upgrades.

## Release Checklist

- Run `go test ./...`.
- Run `make verify-docs` or `go run ./cmd/rayboard verify docs`.
- Build the local binary with `make build`.
- For distributable artifacts, run `make build-cross`.
- Start `rayboard combined` against a fresh database and confirm the admin bootstrap password is printed.
- Log in through the frontend and create a project and ticket.
- Run the demo seed against a backend and confirm it completes.
- Update docs for any user-facing CLI, API, auth, frontend, or automation behavior.
