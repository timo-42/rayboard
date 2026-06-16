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
- shared Lua JSON module
- shared Go<->Lua table conversion layer
- table-to-table helper adapter package for Go-backed Lua functions
- OpenRouter client wrapper
- automation run history
- cron scheduler
- ticket hooks
- custom ticket create pages
- test/preview endpoints

## Package Tasks

1. Common automation:
   - shared nested `engine` object with `engine.type = lua|ai`
   - Lua `engine.script`
   - AI `engine.prompt` and `engine.provider_id`
   - run record creation
   - logs
   - limits
   - validated outputs
2. Lua:
   - GopherLua state per run
   - safe standard library subset
   - shared sandbox package used by cron, hooks, create pages, webhooks, and notification hooks
   - `json.encode`, `json.decode`, and `json.null`
   - `rayboard.json` alias for the same JSON module
   - Go<->Lua conversion helpers for plain tables, arrays, JSON null, strings, booleans, and numbers
   - explicit DTO adapters so `rayboard.*` helpers accept Lua tables and return Lua tables plus errors
   - shared helper result convention: `local value, err = rayboard.action({...})`
   - rejection for mixed-key tables, sparse arrays, recursive tables, functions, userdata, threads, non-finite numbers, raw Go pointers, and unsupported values
   - limits for JSON input bytes, JSON output bytes, and nesting depth
   - table-to-table wrappers for Rayboard API payloads returned by Go-backed helper functions
   - examples and docs updates for JSON, `json.null`, table conversion, helper errors, and safe transformations
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
- Agent 09 documents every public Lua helper, conversion rule, AI schema expectation, and example.

## Tests

- Lua sandbox denial tests.
- Lua JSON encode/decode tests.
- Lua `json.null` round-trip tests.
- Lua mixed/sparse/recursive table rejection tests.
- Go<->Lua conversion tests for Rayboard API helper payloads.
- Lua helper adapter tests verify API/service DTOs round-trip through plain Lua tables.
- Lua docs examples are represented in tests where practical.
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
- Every Lua surface uses the same JSON/table conversion layer.
- Lua scripts never receive raw Go pointers, DB handles, HTTP clients, Shoutrrr secrets, or OpenRouter secrets.
- `/docs` contains examples for every implemented Lua-capable surface before the surface is considered done.
