# Data Model Contract

## IDs And Time

- Public JSON IDs are strings.
- SQLite can use integer primary keys internally, but API clients should not depend on that.
- Store timestamps as UTC.
- Prefer `created_at`, `updated_at`, `deleted_at` or domain-specific revocation fields.

## Migrations

Use explicit migrations from day one:

```text
internal/backend/migrations/
  000001_init.sql
  000002_auth_rbac.sql
```

The migration runner must:

- create a schema migrations table
- apply migrations in order
- run inside a transaction when SQLite allows
- fail startup on migration error
- be callable from tests

## SQLite

- Use WAL mode where appropriate.
- Enable foreign keys.
- Use parameterized SQL only.
- Store attachment bytes in SQLite for v1.
- Use FTS5 virtual tables for full-text search.

## Core Tables By Domain

Foundation:

- migrations
- system_settings
- audit_log

Auth/RBAC:

- users
- sessions
- api_tokens
- groups
- group_memberships
- roles
- role_permissions
- role_bindings

Projects:

- projects
- project_components
- project_versions
- project_statuses
- project_workflows
- boards
- board_columns
- sprints
- epics

Tickets:

- tickets
- ticket backlog rank/order is persisted on tickets unless a later board-specific ordering table is needed
- ticket_labels
- ticket_comments
- ticket_activity
- ticket_attachments
- ticket_custom_field_values
- ticket_links/dependencies if needed later

Custom fields:

- custom_field_definitions
- custom_field_options
- custom_field_values

Search/views:

- saved_views
- ticket_fts
- comment_fts if separate

Automation:

- cron_jobs
- automation_runs
- ticket_hooks
- create_pages
- webhooks
- webhook_deliveries
- notification_hooks

Notifications:

- notifications
- notification_preferences
- notification_destinations
- notification_policies
- notification_deliveries

AI/OpenRouter:

- ai_settings
- ai_run_metadata

Demo:

- demo_markers or generic tagged metadata where cleanup needs it

## Soft Delete

Prefer soft delete for records referenced by history:

- users
- tickets
- comments
- attachments
- automation definitions
- notification destinations

Hard delete is acceptable for ephemeral run/delivery history only if retention policy says so.
