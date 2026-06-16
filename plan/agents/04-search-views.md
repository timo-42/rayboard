# Agent 04: Search And Saved Views

## Mission

Implement CEL filtering, SQLite FTS5 full-text search, and saved views.

Read first:

- `plan/contracts/api-conventions.md`
- `plan/contracts/data-model.md`
- `plan/contracts/authorization.md`

## Deliverables

- CEL expression parser/type checker with `github.com/google/cel-go`
- safe translation from supported CEL subset to parameterized SQL
- FTS5 indexing/querying
- `POST /api/search`
- saved views APIs

## Package Tasks

1. CEL:
   - define field model for tickets/projects/sprints/epics/custom fields
   - support approved functions: `currentUser()`, `today()`, `now()`
   - reject unsupported fields/functions/operators
2. SQL translation:
   - parameterized only
   - project/RBAC scoping always applied
   - custom field type validation
3. FTS:
   - index ticket title/description
   - index comments
   - index attachment metadata where useful
   - update on create/update/delete
4. Search endpoint:
   - accepts `filter`, `text`, `sort`, `limit`, `cursor`
   - combines CEL and FTS
   - returns tickets with pagination
5. Saved views:
   - personal views
   - project-shared views
   - pinned project views
   - display mode/columns/sort/grouping

## Integration Points

- Agent 03 emits indexable events or calls index service.
- Agent 05 search UI and saved view UI use these endpoints.
- Agent 06 Lua/AI cron and webhooks can call search through controlled APIs.

## Tests

- valid and invalid CEL.
- unsupported fields rejected.
- SQL translation parameterized.
- custom field filters by type.
- FTS create/update/delete.
- CEL + FTS combined results.
- saved view CRUD and RBAC.
- pinned project views.

## Acceptance Criteria

- No raw user query text is interpolated into SQL.
- Search always respects RBAC project/ticket visibility.
- Saved views store query/display config, not SQL.
