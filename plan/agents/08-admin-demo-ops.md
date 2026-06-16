# Agent 08: Admin, Demo, Ops, Release

## Mission

Implement system/project/user settings, demo seeding, audit/operational glue, documentation verification, and release verification.

Read first:

- all contracts under `plan/contracts/`
- `plan/README.md`

## Deliverables

- global admin settings
- project settings integration
- user settings integration
- demo seed command
- demo reset endpoint
- audit log
- docs build/link checks
- release/build verification

## Package Tasks

1. Settings:
   - auth/session policy
   - attachment limits
   - notification destinations
   - webhook allowlists
   - OpenRouter settings
   - demo warning
   - backup/export placeholders
   - system health
2. Demo seed:
   - HTTP-only against running backend
   - requires admin credentials
   - requires `--fresh-reset`
   - refuses destructive reset otherwise
   - logs generated demo user credentials once
3. Demo content:
   - users
   - groups and role bindings
   - projects
   - boards/backlog/sprints
   - epics/tickets/comments/activity
   - custom fields
   - attachments
   - saved views
   - FTS-searchable text
   - cron jobs
   - hooks
   - custom create pages
   - incoming/outgoing webhooks
   - notification examples
   - AI examples only when OpenRouter configured
4. Audit log:
   - keep audit entries separate from `ticket_activity` and `domain_events`
   - login failures
   - user/group/role changes
   - token/session revocation
   - settings changes
   - webhook token rotation
   - OpenRouter key changes
   - automation/webhook/notification configuration changes
   - demo reset
5. Release:
   - `go test ./...`
   - cross-build for mac arm and linux amd64
   - verify embedded static assets exist
   - docs link static check
   - verify `/docs/README.md` indexes required docs
   - verify user-facing endpoints/features added by other agents have matching docs updates or tracked docs follow-ups
   - verify Lua/AI automation docs include JSON, Go<->Lua conversion, helper result, limit, and secret-redaction behavior

## Integration Points

- Agent 00 provides demo command shell.
- Every feature agent exposes APIs that demo seed uses.
- Agent 05 exposes settings UI.
- Agent 02 RBAC protects settings/demo reset.
- Agent 09 owns documentation content; Agent 08 owns release-time documentation checks.

## Tests

- settings RBAC and secret redaction.
- demo refuses without `--fresh-reset`.
- demo seed uses HTTP API, not DB writes.
- demo passwords logged and stored hashed.
- repeated fresh reset produces clean dataset.
- audit entries for sensitive changes.
- audit entries are not used as the ticket timeline or durable event outbox.
- cross-build commands complete.
- docs links present.
- docs index covers user, admin, API, automation, development, and operations docs.
- automation docs cover every implemented Lua helper and AI-capable surface.

## Acceptance Criteria

- Demo seed can populate a fresh running instance end to end.
- No demo path bypasses backend authorization or validation.
- Release verification can be run from a clean checkout.
- Release checks fail when required documentation is missing.
