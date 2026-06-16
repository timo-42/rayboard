# Agent 06: Automation, Lua, AI, Cron, Hooks, Forms

## Mission

Implement Lua and OpenRouter AI automation surfaces: cron jobs, ticket hooks, custom create pages, and common run history/sandbox behavior.

Read first:

- `plan/contracts/automation-engines.md`
- `plan/contracts/authorization.md`
- `plan/contracts/api-conventions.md`

## Deliverables

- automation engine abstraction
- GopherLua runtime wrapper
- OpenRouter client wrapper
- automation run history
- cron scheduler
- ticket hooks
- custom ticket create pages
- test/preview endpoints

## Package Tasks

1. Common automation:
   - `engine = lua|ai`
   - run record creation
   - logs
   - limits
   - validated outputs
2. Lua:
   - GopherLua state per run
   - safe standard library subset
   - no filesystem/shell/socket/DB/unrestricted HTTP
   - surface-specific helper registration
3. OpenRouter:
   - global settings
   - API key storage/redaction
   - allowed models
   - JSON schema/output validation
   - timeout and usage metadata
4. Cron:
   - robfig cron scheduler
   - owner user
   - no overlap
   - manual run
   - run history
5. Ticket hooks:
   - project scoped
   - before/after create/update
   - transform/reject before
   - inspect/log after
6. Custom create pages:
   - form definitions
   - Lua/AI schema/default/options logic
   - server-rendered controls only
   - submit through normal ticket create path

## Integration Points

- Agent 02 provides principal/RBAC.
- Agent 03 provides ticket create/update service and events.
- Agent 04 search is callable through constrained helper.
- Agent 05 builds editors/test panels.
- Agent 07 owns webhook and notification hook surfaces but reuses common engine pieces.

## Tests

- Lua sandbox denial tests.
- AI structured JSON validation.
- missing/disabled OpenRouter behavior.
- cron schedule/manual/no-overlap.
- disabled owner cannot run automation.
- before hook transform/reject.
- after hook cannot mutate committed data.
- custom form schema/default/options.
- run history contains no secrets.

## Acceptance Criteria

- Automation effects always pass through normal service/API authorization.
- AI output is never applied without schema validation.
- Lua and AI share limits and run history where possible.
