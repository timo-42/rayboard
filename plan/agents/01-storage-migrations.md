# Agent 01: Storage And Migrations

## Mission

Provide SQLite storage, migrations, repository helpers, and low-level persistence support for other agents.

Read first:

- `plan/contracts/data-model.md`
- `plan/contracts/architecture.md`

## Deliverables

- DB open/close helper using `modernc.org/sqlite`
- migration runner
- base schema
- transaction helper
- repository test harness
- FTS and attachment storage primitives

## Package Tasks

1. Implement `store.Open(path)`:
   - enable foreign keys
   - configure WAL where appropriate
   - expose `*sql.DB`
2. Implement migration runner:
   - embedded migrations
   - schema version table
   - idempotent startup
3. Add first migrations for:
   - users/auth/RBAC skeleton
   - projects/tickets skeleton
   - automation/run history skeleton
   - notification skeleton
   - saved views
   - attachments
   - FTS tables
4. Add transaction helper:
   - context-aware
   - rollback on error/panic
5. Add repository test harness:
   - temporary DB
   - migrate
   - cleanup

## Integration Points

- Agent 02 depends on auth/RBAC tables.
- Agent 03 depends on project/ticket/custom field tables.
- Agent 04 depends on FTS and saved view tables.
- Agent 07 depends on notification/webhook/delivery tables.

## Tests

- migrations apply from empty DB.
- migrations are idempotent.
- foreign keys are enforced.
- transactions commit and rollback correctly.
- FTS tables can insert, update, delete, and query sample rows.
- attachment blob roundtrip works.

## Acceptance Criteria

- No feature package creates tables outside migrations.
- Tests can create isolated migrated DBs quickly.
- Cross-build remains CGO-free.
