# Agent 08: Admin, Demo, Ops, Release

## Mission

Implement system/project/user settings, demo seeding, audit/operational glue, documentation checks, and release verification.

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
- docs link checks
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
   - user/group/role changes
   - token/session revocation
   - settings changes
   - automation/webhook/notification configuration changes
   - demo reset
5. Release:
   - `go test ./...`
   - cross-build for mac arm and linux amd64
   - verify embedded static assets exist
   - docs link static check

## Integration Points

- Agent 00 provides demo command shell.
- Every feature agent exposes APIs that demo seed uses.
- Agent 05 exposes settings UI.
- Agent 02 RBAC protects settings/demo reset.

## Tests

- settings RBAC and secret redaction.
- demo refuses without `--fresh-reset`.
- demo seed uses HTTP API, not DB writes.
- demo passwords logged and stored hashed.
- repeated fresh reset produces clean dataset.
- audit entries for sensitive changes.
- cross-build commands complete.
- docs links present.

## Acceptance Criteria

- Demo seed can populate a fresh running instance end to end.
- No demo path bypasses backend authorization or validation.
- Release verification can be run from a clean checkout.
