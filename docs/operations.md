# Operations

Rayboard can run as one combined process or as split frontend/backend processes:

```bash
rayboard combined --db ./data/rayboard.sqlite
rayboard backend --backend-addr 0.0.0.0:8081 --db ./data/rayboard.sqlite
rayboard frontend --frontend-addr 0.0.0.0:8080 --backend-url http://127.0.0.1:8081
```

Only the backend opens SQLite. In split mode, run exactly one backend process against a given SQLite database file.

## Health Checks

Current health endpoints:

- frontend: `GET /health`
- backend: `GET /api/health`

## Backups

For file-backed SQLite databases, back up the database file and its WAL/shm sidecar files while Rayboard is stopped, or use SQLite backup tooling when online backups are added. Online backup commands and restore automation are **Planned**.

## Logs and Secrets

Startup logs include the randomized POC admin password. Demo seed logs include randomized demo user passwords. Do not ship those logs to shared systems unless that is intentional for the demo environment.

OpenRouter keys, Shoutrrr destination URLs, webhook secrets, and generated tokens must never be exposed to Lua scripts, browser code, or ordinary API responses.

## Security Audit Log

Rayboard stores security/admin-sensitive events in the SQLite `audit_log` table, separate from user-facing ticket activity and durable domain events. The current audit slice records password login failures, session creation/logout, API token creation/revocation, user creation/disable/enable/delete, group creation, group membership changes, role binding changes, OpenRouter provider changes, Shoutrrr destination changes, and global settings changes.

Audit payloads are JSON metadata for operations and must not include plaintext API tokens, generated passwords, password hashes, session secrets, Shoutrrr URLs, webhook tokens, or OpenRouter keys. Global admins can inspect recent entries with `GET /api/audit-log`; the endpoint supports `limit`, `event_type`, `actor_user_id`, `subject_type`, `subject_id`, and `outcome` query filters.

## Upgrades

Migrations run at backend startup. For now:

- back up SQLite before upgrading;
- run `go test ./...`, `make verify-docs`, and `make build-cross` before producing release artifacts;
- start a fresh `combined` instance and verify login, project creation, ticket creation, and demo seed;
- review [Development](development.md) release checks before tagging.

## Planned Operational Work

Production hardening is **Planned** and should include structured logging, configurable session TTLs, TLS/reverse-proxy guidance, online backup/restore docs, metrics, audit-log UI/export tools, rate limits, webhook delivery history, automation run inspection, and notification delivery inspection.
