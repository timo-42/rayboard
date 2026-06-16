# Architecture Contract

## Goal

Rayboard starts as one Go binary but is internally split so the frontend can later run separately and talk to the backend over HTTP.

```text
browser UI -> frontend server -> backend HTTP API -> services -> repositories -> SQLite
```

Combined mode may run frontend and backend in one process, but frontend code must still use the backend HTTP client boundary. Do not import backend service packages from frontend handlers.

## Package Layout

Target package shape:

```text
cmd/rayboard/
internal/app/
internal/config/
internal/runtime/
internal/backend/
internal/backend/httpapi/
internal/backend/service/
internal/backend/store/
internal/backend/migrations/
internal/frontend/
internal/frontend/templates/
internal/frontend/static/
internal/shared/
```

Use more domain packages under `internal/backend/service/` and `internal/backend/store/` as the implementation grows.

## Runtime Modes

### `rayboard combined`

- Opens SQLite.
- Runs migrations.
- Starts backend API at `--backend-addr`.
- Starts frontend server at `--frontend-addr`.
- Configures frontend backend URL to the local backend URL.
- Exposes both UI and API.
- Handles one coordinated shutdown path.

### `rayboard backend`

- Opens SQLite.
- Runs migrations.
- Starts only backend API.
- Runs scheduler/queue workers that belong to backend behavior.

### `rayboard frontend`

- Starts only frontend server.
- Requires `--backend-url` or uses default.
- Serves embedded templates/static assets.
- Calls backend over HTTP.
- Does not open SQLite.

## Config Defaults

- `--frontend-addr`: `127.0.0.1:8080`
- `--backend-addr`: `127.0.0.1:8081`
- `--backend-url`: `http://127.0.0.1:8081`
- `--db`: `rayboard.sqlite`

Environment variables may mirror flags, but flags win.

## Build Rules

- Default to CGO-free dependencies.
- SQLite driver: `modernc.org/sqlite` unless implementation proves a better pure-Go driver.
- Target builds:

```bash
GOOS=darwin GOARCH=arm64 go build -o dist/rayboard-darwin-arm64 ./cmd/rayboard
GOOS=linux GOARCH=amd64 go build -o dist/rayboard-linux-amd64 ./cmd/rayboard
```

## Event Boundary

Service methods that mutate domain data should emit internal events for:

- activity history
- notifications
- webhooks
- FTS updates where needed
- audit/run history where needed

Events are internal Go values, not external queue contracts in v1.
