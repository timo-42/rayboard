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
docs/
```

Use more domain packages under `internal/backend/service/` and `internal/backend/store/` as the implementation grows.

Backend HTTP API packages should stay resource-focused as the route surface grows:

```text
internal/backend/httpapi/
  router.go
  shared/
  projects/
    routes.go
    schema.go
    provider.go
  tickets/
    routes.go
    schema.go
    provider.go
  boards/
    routes.go
    schema.go
    provider.go
```

- `router.go` assembles resource route providers and shared middleware only.
- `routes.go` registers routes and contains thin HTTP handlers.
- `schema.go` contains Huma/OpenAPI-facing DTOs; request DTOs use `Input`, response DTOs use `Output`.
- `provider.go` wires dependencies and local route helpers.
- HTTP packages decode/auth/call services/encode; domain services and repositories own validation, transactions, and SQLite access.
- Do not put raw DB CRUD in HTTP route packages.
- Put `/api/projects` and closely project-scoped settings under the project route package.
- Give top-level resources such as `/api/tickets/{ticket_id}`, `/api/boards/{board_id}`, and `/api/sprints/{sprint_id}` their own route packages.
- Avoid deeply recursive package trees unless a subresource becomes large enough to justify one.
- Migrate incrementally; existing domain services can remain while handlers move.

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

## Event And History Boundary

Service methods that mutate domain data should record durable domain events for:

- activity history
- notifications
- webhooks
- FTS updates where needed
- audit/run history where needed

Use three separate records:

- `ticket_activity` is the user-facing timeline for tickets and epics.
- `domain_events` is the durable append-only backend event/outbox stream for notifications, webhooks, automations, search refreshes, integrations, and retryable async processing.
- `audit_log` is the security/admin history for sensitive operational actions.

The in-memory Go event bus may dispatch after durable writes, but it is not the source of truth and must not be the only record of meaningful mutations.

## Documentation Boundary

User-facing behavior is documented under `/docs`.

- `/docs/README.md` is the documentation index.
- Feature docs must describe implemented behavior accurately and mark planned behavior clearly.
- API, CLI, config, auth/RBAC, Lua/AI automation, frontend, operations, and development docs are separate audience-focused documents.
- Code changes that add user-facing behavior should include docs updates or a tracked Agent 09 follow-up.
