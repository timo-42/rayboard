# Rayboard Documentation

Rayboard is a single Go binary for a Jira-like project board proof of concept. The current implementation includes runtime modes, SQLite persistence, admin bootstrap, browser/API authentication, RBAC, projects, tickets, comments, attachments, search, saved views, a first sprint API slice, an in-app notification API slice with read/unread state, Lua cron job API/scheduler basics, a small embedded frontend, and a demo seed command.

Planned features are marked **Planned** in these docs. Do not treat planned sections as implemented API behavior.

## Guides

- [Runtime and Configuration](runtime-config.md): CLI modes, flags, environment variables, SQLite behavior, and admin bootstrap.
- [Authentication and RBAC](auth-rbac.md): sessions, CSRF, bearer tokens, users, groups, roles, permissions, and role bindings.
- [User Guide](user-guide.md): current browser workflows, API-only workflows, and planned rich board/backlog/sprint UI and reports.
- [Admin Guide](admin-guide.md): bootstrap credentials, user/group administration, RBAC, settings, and planned admin surfaces.
- [API Guide](api.md): JSON conventions, errors, auth requirements, implemented endpoints, search, saved views, comments, attachments, notifications, and cron jobs.
- [Frontend Architecture](frontend.md): embedded templates/static assets, browser workflow, reverse proxy behavior, and planned frontend dependencies.
- [Demo Seed](demo-seed.md): current `rayboard demo seed` behavior and planned seed expansion.
- [Automation and Lua](automation-lua.md): current status plus Lua JSON/table conversion, cron jobs, and planned hooks, webhooks, notifications, and OpenRouter AI design.
- [Development](development.md): local commands, tests, builds, migrations, package boundaries, and release checks.
- [Operations](operations.md): deployment modes, backups, logs, upgrades, and operational checks.

## Required Upstream References

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- GopherLua: https://github.com/yuin/gopher-lua
- robfig cron: https://pkg.go.dev/github.com/robfig/cron/v3
- robfig cron source: https://github.com/robfig/cron
- HTMX: https://htmx.org/
- SortableJS: https://sortablejs.github.io/Sortable/
- CodeMirror: https://codemirror.net/
- OpenRouter: https://openrouter.ai/docs
- Shoutrrr: https://github.com/containrrr/shoutrrr
- Shoutrrr docs: https://containrrr.dev/shoutrrr/
- SQLite FTS5: https://www.sqlite.org/fts5.html
