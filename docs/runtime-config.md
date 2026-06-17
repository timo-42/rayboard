# Runtime and Configuration

Rayboard runs as one binary with three runtime modes:

```bash
rayboard combined [flags]
rayboard backend [flags]
rayboard frontend [flags]
```

`combined` opens SQLite, runs migrations, bootstraps the POC admin user, starts the backend API, and starts the frontend server in one process. `backend` starts only the backend API and owns SQLite. `frontend` starts only the embedded frontend and proxies API calls to a backend URL; it does not open SQLite.

## Flags and Environment

Flags are bound for all three runtime commands. Environment variables provide defaults, and flags win.

| Flag | Environment | Default | Purpose |
| --- | --- | --- | --- |
| `--frontend-addr` | `RAYBOARD_FRONTEND_ADDR` | `127.0.0.1:8080` | Frontend listen address. |
| `--backend-addr` | `RAYBOARD_BACKEND_ADDR` | `127.0.0.1:8081` | Backend API listen address. |
| `--backend-url` | `RAYBOARD_BACKEND_URL` | `http://127.0.0.1:8081` | Backend base URL used by the frontend/proxy. |
| `--db` | `RAYBOARD_DB` | `rayboard.sqlite` | SQLite database path. |
| `--outgoing-webhook-base-url` | `RAYBOARD_OUTGOING_WEBHOOK_BASE_URL` | empty | Base URL allowed for outgoing webhook delivery. Lua/AI webhooks return only relative paths below this URL. |

Examples:

```bash
rayboard combined --db ./data/rayboard.sqlite
rayboard backend --backend-addr 0.0.0.0:8081 --db ./data/rayboard.sqlite
rayboard frontend --frontend-addr 0.0.0.0:8080 --backend-url http://127.0.0.1:8081
```

## SQLite and Migrations

The backend uses `modernc.org/sqlite`. Foreign keys are enabled on every connection, `busy_timeout` is set to 5000 ms, and WAL mode is enabled for file-backed databases.

Migrations are embedded under `internal/backend/migrations`. The schema includes users, sessions, API tokens, groups, group memberships, roles, role permissions, role bindings, projects, tickets, ticket labels, comments, activity, attachments, saved views, automation run records, ticket create pages, notifications, notification preferences, notification destinations, notification policies, notification deliveries, webhooks, outgoing webhook deliveries, domain events, and SQLite FTS5 virtual tables for ticket text, comment text, and attachment metadata.

In `combined` and `backend` modes, a lightweight automation-delivery worker first enqueues outgoing webhook deliveries for pending `domain_events`, then processes pending comment/ticket-update notifications, external notification deliveries, and due outgoing webhook deliveries. Successful domain-event rows are marked `processed`; rows that cannot be handled are marked `failed` with `last_error` and an incremented attempt count. Successful delivery rows are marked `delivered`; failed deliveries are retried with backoff until their retry budget is exhausted or a permanent validation/destination error marks them `failed`.

## Admin Bootstrap

On every `combined` or `backend` startup, Rayboard:

- seeds or updates built-in roles and role permissions;
- creates or updates the `admin` user;
- resets the `admin` password to a new random POC password;
- ensures `admin` has the built-in `global_admin` role;
- prints the generated credentials to stdout.

The printed password is for local POC use. Do not copy real credentials into docs, scripts, commits, or issue trackers.

## Split Deployment Model

The intended boundary is:

```text
browser UI -> frontend server -> backend HTTP API -> services -> repositories -> SQLite
```

In current `frontend` mode, `/api/*` requests are reverse-proxied to `--backend-url`. In `combined` mode, both servers run in the same process but still expose separate frontend and backend listeners.
