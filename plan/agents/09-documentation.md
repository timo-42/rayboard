# Agent 09: Documentation

## Mission

Create and maintain proper project documentation under `/docs`. Document implemented behavior accurately, mark planned behavior clearly, and keep docs updated with every user-facing endpoint, CLI flag, config variable, Lua helper, automation surface, and frontend workflow.

Read first:

- `REQUIREMENTS.md`
- `plan/contracts/api-conventions.md`
- `plan/contracts/architecture.md`
- `plan/contracts/authorization.md`
- `plan/contracts/automation-engines.md`

## Deliverables

- `/docs/README.md` documentation index
- user guide
- admin guide
- API guide
- authentication and RBAC guide
- automation guide covering Lua, JSON helpers, AI, cron, hooks, create pages, webhooks, and notification hooks
- operations guide
- development guide
- frontend architecture guide
- demo seed guide
- docs link/checklist support for Agent 08
- docs coverage checklist that feature agents can update when landing user-facing behavior

## Package Tasks

1. Documentation structure:
   - create `/docs/README.md` as the index
   - split docs by audience: user, admin, API, automation, development, operations
   - keep docs concise and versioned with implemented behavior
   - label unimplemented planned features as planned
   - maintain a lightweight coverage checklist mapping implemented CLI/API/UI/automation features to docs pages
2. Runtime and deployment docs:
   - `combined`, `frontend`, and `backend` modes
   - config flags and environment variables
   - SQLite database behavior and migrations
   - admin bootstrap password behavior
   - split frontend/backend deployment model
3. Auth and RBAC docs:
   - browser sessions
   - CSRF
   - API bearer tokens
   - disabled users
   - groups, roles, permissions, role bindings, global/project scopes
   - built-in roles and effective permissions
4. API docs:
   - common JSON error format
   - authentication requirements
   - pagination and cursor conventions
   - implemented endpoints with request/response examples
   - attachment upload/download behavior
   - search and saved views behavior
5. Query docs:
   - link to CEL upstream docs: https://cel.dev/
   - link to CEL Go docs/examples: https://github.com/google/cel-go
   - document Rayboard's supported CEL subset and current implementation limitations
   - include examples for ticket search, saved views, current user, FTS text search, and sort/pagination
6. Lua and automation docs:
   - link to GopherLua: https://github.com/yuin/gopher-lua
   - document sandbox restrictions
   - document available `rayboard.*` helpers per surface as they are implemented
   - document the shared `json` and `rayboard.json` module
   - document `json.encode`, `json.decode`, and `json.null`
   - document Go<->Lua table conversion rules, rejected values, size/depth limits, and error behavior
   - document the helper result convention where Go-backed Lua functions return plain Lua tables plus errors
   - provide examples for cron jobs, ticket hooks, create pages, webhooks, notification hooks, JSON payload transformation, validation, `json.null`, and helper error handling
7. AI docs:
   - link to OpenRouter docs: https://openrouter.ai/docs
   - document provider settings, allowed models, prompt/output schema expectations, limits, and audit/run history
8. Notification docs:
   - link to Shoutrrr docs
   - document named destinations, secret redaction, inheritance, notification policies, hooks, delivery history, and troubleshooting
9. Frontend docs:
   - server-rendered templates
   - HTMX/vanilla JS asset policy
   - embedded static assets
   - no Node/npm build step
   - future CSS override extension points
10. Development docs:
   - local run commands
   - `go test ./...`
   - cross-build commands
   - migration rules
   - package boundaries
   - release checklist
11. Documentation gating:
   - require docs updates in the same PR/commit series for user-facing behavior
   - allow tracked Agent 09 follow-ups only for behavior that is intentionally hidden or still behind a planned/stub label
   - make docs checks fail when required index links or upstream links are missing

## Integration Points

- Agent 00 provides CLI/runtime behavior to document.
- Agent 02 provides auth/RBAC behavior to document.
- Agent 03 provides project/ticket/comment/attachment behavior to document.
- Agent 04 provides search/saved view behavior to document.
- Agent 05 provides frontend behavior and embedded asset policy to document.
- Agent 06 provides Lua/AI behavior and JSON bridge semantics to document.
- Agent 07 provides notification/webhook behavior to document.
- Agent 08 consumes docs for release checks.

## Tests

- docs index links all required topic docs.
- docs contain required upstream links for CEL, cel-go, GopherLua, robfig cron, HTMX, SortableJS, CodeMirror, OpenRouter, Shoutrrr, and SQLite FTS5.
- API docs include examples for every implemented public endpoint.
- automation docs include JSON encode/decode and Go<->Lua conversion examples.
- automation docs include at least one table-to-table helper example returning `value, err`.
- docs checks fail on broken local links.

## Acceptance Criteria

- `/docs/README.md` is the user-facing documentation entry point.
- Implemented features are documented without overstating planned features.
- Planned features are clearly labeled as planned.
- User-facing code changes include docs changes or an explicit tracked docs follow-up.
- New Lua helpers, AI schemas, endpoints, CLI flags, config variables, and frontend workflows are not done until documented.
- Documentation never includes real generated credentials, API keys, webhook tokens, OpenRouter keys, or Shoutrrr secrets.
