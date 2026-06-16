# Agent 00: Foundation Runtime

## Mission

Create the Go module, binary entrypoint, runtime modes, config, process lifecycle, and cross-build foundation.

Read first:

- `plan/contracts/architecture.md`
- `plan/contracts/api-conventions.md`

## Deliverables

- `go.mod`
- `cmd/rayboard/main.go`
- config parsing for shared flags/env
- `combined`, `frontend`, `backend`, and `demo seed` command shells
- backend HTTP server skeleton
- frontend HTTP server skeleton
- health endpoints
- graceful shutdown
- build script or Makefile targets

## Package Tasks

1. Create module and base package layout.
2. Implement config struct:
   - frontend addr
   - backend addr
   - backend URL
   - DB path
   - mode
3. Implement CLI parser using standard library `flag` or a small dependency.
4. Implement runtime startup:
   - backend mode opens DB and starts API
   - frontend mode starts UI only
   - combined mode starts backend then frontend
5. Add `GET /api/health` on backend.
6. Add `GET /` shell page on frontend.
7. Add shutdown on SIGINT/SIGTERM.
8. Add cross-build targets:
   - `darwin/arm64`
   - `linux/amd64`

## Integration Points

- Agent 01 provides DB open/migrate function.
- Agent 05 replaces frontend shell with real templates.
- Agent 08 fills in demo command implementation.

## Tests

- CLI parses modes and flags.
- backend health endpoint responds.
- frontend root responds.
- combined mode starts both servers on test ports and shuts down.
- cross-build command compiles after dependencies land.

## Acceptance Criteria

- `go test ./...` passes.
- `go run ./cmd/rayboard backend --backend-addr 127.0.0.1:0` can start in tests.
- No frontend package imports backend service/store packages.
