# Rayboard Multi-Agent Implementation Plan

This directory is the implementation plan for Rayboard. It is written for multiple agents with limited context. An agent should read this file, then the relevant shared contracts, then exactly one workstream file under `plan/agents/`.

The source of truth for product scope remains `REQUIREMENTS.md`. This plan turns that scope into separable implementation tracks, integration contracts, and acceptance criteria.

## Ground Rules

- Build in Go as one binary first, with internal frontend/backend boundaries.
- Use SQLite with a CGO-free driver, defaulting to `modernc.org/sqlite`, so `darwin/arm64` and `linux/amd64` cross-builds remain straightforward.
- Keep frontend and backend communicating through HTTP-compatible interfaces even in combined mode.
- Keep all authorization backend-owned and routed through the RBAC evaluator.
- Add migrations from day one; never rely on implicit schema creation in feature code.
- Prefer narrow vertical slices over broad untested scaffolding.
- Do not bypass public service/API paths for demo seed, cron, webhooks, Lua, AI, or frontend behavior.

## Directory Layout

```text
plan/
  README.md
  contracts/
    architecture.md
    api-conventions.md
    authorization.md
    automation-engines.md
    data-model.md
  agents/
    00-foundation-runtime.md
    01-storage-migrations.md
    02-auth-rbac.md
    03-core-tracking.md
    04-search-views.md
    05-frontend-ui.md
    06-automation-lua-ai.md
    07-notifications-webhooks.md
    08-admin-demo-ops.md
```

## Agent Assignments

- Agent 00: runtime, CLI, process modes, HTTP server wiring, build targets.
- Agent 01: SQLite, migrations, repositories, FTS, attachments storage foundations.
- Agent 02: auth, sessions, tokens, RBAC, groups, permissions.
- Agent 03: projects, tickets, boards, backlog, sprints, roadmap, custom fields.
- Agent 04: CEL search, SQLite FTS integration, saved views.
- Agent 05: embedded frontend, templates, HTMX, static assets, UI shells.
- Agent 06: automation engine, Lua, OpenRouter AI, cron, hooks, custom create pages.
- Agent 07: notifications, Shoutrrr, webhooks, delivery queues.
- Agent 08: admin settings, demo seed, audit/ops glue, release/build verification.

## Dependency Order

Work can start in parallel, but integration should follow this order:

1. Agent 00 and Agent 01 establish module, binary, config, DB, migrations, and test harness.
2. Agent 02 lands auth/RBAC primitives because most other agents need principals and permission checks.
3. Agent 03 lands core project/ticket APIs and events.
4. Agents 04, 06, and 07 integrate against core tickets/events.
5. Agent 05 builds UI pages once API shapes are stable enough.
6. Agent 08 ties settings, demo data, docs checks, and release verification together.

## Shared Contracts

Every agent must follow these:

- `contracts/architecture.md`: package boundaries and process modes.
- `contracts/api-conventions.md`: route style, JSON, errors, pagination, auth.
- `contracts/authorization.md`: RBAC checks and principal model.
- `contracts/data-model.md`: migration style, IDs, timestamps, storage rules.
- `contracts/automation-engines.md`: Lua/AI execution contracts and sandbox rules.

## Definition Of Done

A workstream is complete when:

- It has repository/service/API layers for its domain.
- It has backend tests for success, auth failure, validation failure, and RBAC denial.
- It updates frontend or leaves explicit UI stubs for Agent 05.
- It emits activity/audit events where applicable.
- It does not introduce direct DB access from frontend, Lua, AI, cron, or demo seed.
- `go test ./...` passes.

## Milestones

### Milestone 1: Executable Skeleton

- `rayboard combined`, `rayboard frontend`, `rayboard backend`.
- Health endpoints.
- SQLite open/migrate.
- Embedded frontend shell.
- Cross-build commands work for `darwin/arm64` and `linux/amd64`.

### Milestone 2: Secure Core

- Admin bootstrap password reset/logging.
- Password login, sessions, API tokens.
- RBAC with groups and role bindings.
- Project/ticket CRUD.
- Basic UI login and ticket list/detail.

### Milestone 3: Jira-Like POC

- Boards, backlog, sprints, epics/roadmap, components, releases.
- Custom fields.
- Attachments.
- CEL + FTS search.
- Saved views.

### Milestone 4: Automation And Integration

- Lua cron jobs.
- Lua ticket hooks.
- Custom create pages.
- Incoming/outgoing webhooks.
- OpenRouter AI alternative engine.
- Shoutrrr notifications and notification hooks.

### Milestone 5: Demo And Operations

- Admin/project/user settings.
- Demo seed command.
- Run history, delivery history, logs.
- Docs links and example snippets.
- Release/build verification.
