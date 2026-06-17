# Automation and Lua

Automation public surfaces are partially implemented. The current implementation includes a shared Lua JSON sandbox foundation in `internal/backend/luasandbox`, automation run history persistence in `internal/backend/automation`, a generic `/api/engines/test` dry-run endpoint, cron job CRUD/scheduler/manual run/history APIs, ticket-hook runner/CRUD/preview APIs, and incoming webhook CRUD/execution APIs.

Relevant upstream references:

- GopherLua: https://github.com/yuin/gopher-lua
- wazero: https://github.com/wazero/wazero
- robfig cron: https://pkg.go.dev/github.com/robfig/cron/v3
- robfig cron source: https://github.com/robfig/cron
- OpenRouter docs: https://openrouter.ai/docs
- Shoutrrr: https://github.com/containrrr/shoutrrr
- Shoutrrr docs: https://containrrr.dev/shoutrrr/

## Lua Sandbox Foundation

Lua surfaces use one shared sandbox runtime package and one shared Go/Lua conversion layer. Sandboxes must not expose filesystem access, shell execution, raw sockets, unrestricted HTTP, direct SQLite handles, raw Go pointers, backend store/service handles, Shoutrrr secrets, or OpenRouter keys.

Lua-capable surfaces:

- generic engine tests: dry-run Lua/AI execution with supplied JSON input, `context`, `input`, JSON helpers, `rayboard.log`, source redaction, and run history implemented;
- cron jobs: first API/scheduler slice implemented;
- ticket hooks: Lua runner, management API, and preview API implemented; UI still **Planned**;
- custom ticket create pages: static definition/submit API implemented; dynamic Lua form logic and UI still **Planned**;
- incoming webhooks: definition CRUD, token auth, Lua/AI validation/logging, constrained Rayboard actions, and run history implemented;
- outgoing webhooks: definition CRUD, event-triggered delivery persistence, Lua/AI request shaping, controlled outbound HTTP, retries, manual retry, and delivery history API implemented;
- notification hooks: API-only Lua/AI suppress/transform/route slice implemented; preview, run history, UI, and richer routing are **Planned**.

Every surface should enforce timeouts, max script size, max log size, max input/output size, max JSON input/output bytes, max table nesting depth, and max action count where actions exist. The current shared JSON defaults are 1 MiB max JSON input, 1 MiB max encoded JSON output, and 64 levels max nesting depth.

## Engine Workbench

`POST /api/engines/test` is the shared automation workbench endpoint. It accepts a Kubernetes-style request body:

```json
{
  "spec": {
    "surface": "ticket_hook_before",
    "project_id": "project_123",
    "actor_user_id": "user_123",
    "engine": {
      "type": "lua",
      "script": "rayboard.log(input.title); return { ok = true }"
    },
    "context": {
      "ticket_id": "ticket_123"
    },
    "input": {
      "title": "Preview"
    },
    "dry_run": true
  }
}
```

The endpoint requires `automations:manage` globally or for `spec.project_id`. It currently normalizes all test executions to `dry_run = true`; workbench runs do not persist ticket or project mutations. Lua receives the supplied `context` and `input` globals. `context` always includes normalized `surface`, `project_id`, `actor_user_id`, and `dry_run` fields, and user-supplied context fields such as `ticket_id` are preserved unless they conflict with normalized fields.

Responses use `metadata`, `spec`, and `status`. `spec.engine.script` and `spec.engine.prompt` are redacted from responses and run history. `status.output` contains the returned Lua table or AI JSON object, `status.logs` contains captured log lines, `status.duration_millis` reports elapsed execution time when available, `status.engine` contains redacted engine metadata, and `status.error` contains runtime failures.

## JSON Module

Every Lua surface exposes the same sandboxed JSON API:

```lua
local value = json.decode('{"title":"Bug","assignee_id":null}')
value.assignee_id = json.null
local encoded = json.encode(value)

local same = rayboard.json.decode(encoded)
```

`json.encode`, `json.decode`, `rayboard.json.encode`, and `rayboard.json.decode` are aliases for the same implementation. `json.null` is a stable sentinel for JSON null.

Decode rules:

- JSON objects become Lua tables with string keys.
- JSON arrays become 1-indexed Lua array tables.
- Strings, booleans, and numbers map directly.
- JSON null becomes `json.null`.

Encode rules:

- Lua array tables become JSON arrays.
- Lua string-key tables become JSON objects.
- `json.null` becomes JSON null.
- Mixed string/numeric-key tables are rejected.
- Sparse array tables are rejected.
- Recursive tables are rejected.
- Functions, userdata, threads, raw Go pointers, and unsupported values are rejected.
- Non-finite numbers are rejected.

Go-backed Rayboard functions exposed to Lua accept plain Lua tables and return plain Lua tables plus an error value using these same rules:

```lua
local ticket, err = rayboard.create_ticket({
  project_id = "project_...",
  title = "Investigate import failure",
  assignee_id = json.null
})

if err then
  rayboard.log("create failed: " .. err.message)
end
```

Helper calls use the same result convention everywhere:

```lua
local value, err = rayboard.action({ id = "..." })
if err ~= nil then
  return { error = err.message }
end
```

Validation and limit failures return or raise messages such as `JSON input exceeds 1048576 bytes`, `encoded JSON exceeds 1048576 bytes`, `JSON depth exceeds 64`, `mixed string and array keys`, `sparse array`, `recursive table`, or `use json.null for JSON null`.

## Cron Jobs

Cron jobs use robfig cron for schedules, GopherLua for Lua execution, and OpenRouter for AI execution. The public API/scheduler slice supports cron job CRUD, manual runs, and run history. Jobs act as their owner user, inherit the owner's current RBAC permissions at run time, and should not overlap by default.

Cron job management requires automation permissions. Disabled users cannot run owned cron jobs.

Run history uses the shared automation run-history model. A run record is used for scheduled and manual runs and may include trigger type, job identity, owner/project context when applicable, input/output summaries, logs, status, error details, start/finish timestamps, duration, and applied limits. Run history must not expose secrets.

Implemented cron Lua helpers:

- `rayboard.log(message)`
- `rayboard.search({ project_id, filter, text, sort, limit, cursor })`
- `rayboard.get_ticket({ ticket_id })`
- `rayboard.create_ticket({ project_id, title, description, status, priority, type, reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, labels, custom_fields })`
- `rayboard.update_ticket({ ticket_id, title, description, status, priority, type, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, labels, custom_fields })`
- `rayboard.comment({ ticket_id, body })`

These helpers return `value, nil` on success and `nil, { message = "..." }` on failure. Each helper goes through the normal backend service and RBAC path using the cron job owner as an `AuthKindCron` principal.

AI cron jobs use `engine.type = "ai"`, `engine.prompt`, and `engine.provider_id`. The provider references `/api/openrouter-providers`. At run time Rayboard calls OpenRouter Chat Completions with the provider default model, timeout, max output tokens, and JSON-object response mode. The assistant response must be a JSON object. The validated object is stored under run output, with provider/model/usage metadata and without the prompt or API key. AI cron jobs do not apply returned actions yet.

Example shape:

```lua
local tickets, err = rayboard.search({
  filter = 'status == "todo" && assignee_id == currentUser()',
  limit = 10
})
if err then return { error = err.message } end

for _, ticket in ipairs(tickets.items) do
  local fetched, get_err = rayboard.get_ticket({ ticket_id = ticket.id })
  if get_err then return { error = get_err.message } end
  rayboard.log(fetched.key .. ": " .. fetched.title)
end
```

Example create/comment shape:

```lua
local ticket, err = rayboard.create_ticket({
  project_id = "project_...",
  title = "Investigate recurring alert",
  labels = {"automation", "triage"}
})
if err then return { error = err.message } end

local matching, search_err = rayboard.search({
  project_id = "project_...",
  filter = '"automation" in labels && due_date <= today()',
  limit = 10
})
if search_err then return { error = search_err.message } end

local comment, comment_err = rayboard.comment({
  ticket_id = ticket.id,
  body = "Created by scheduled triage for " .. tostring(#matching.items) .. " matching tickets."
})
if comment_err then return { error = comment_err.message } end
```

## Ticket Hooks

The backend ticket hook runner is implemented in the tracker service. Project-scoped Lua and AI hooks can run before ticket create/update to validate or transform the pending payload. After hooks run after commit, may inspect/log, and do not roll back committed ticket changes if they fail. Hook CRUD and single-hook preview are available through the API; UI is still **Planned**.

Hook Lua receives `context`, `ticket`, and for update hooks `current`. The preview API uses the same globals for one saved hook without changing tickets or persisting `last_error`. The only Rayboard helper exposed in this first hook sandbox is `rayboard.log(message)`.

AI ticket hooks use `engine.type = "ai"`, `engine.prompt`, and `engine.provider_id`. Rayboard appends hook context and ticket input to the prompt, calls OpenRouter with JSON-object response mode, and requires the response to be a JSON object. Before hooks may return `{"ticket": {...}}` to transform the pending payload or `{"reject": {"message": "..."}}` to reject. After hook output is recorded as hook output only and cannot change committed tickets. OpenRouter API keys are never exposed to prompts, preview responses, hook output, or errors.

Example validation shape:

```lua
if ticket.priority == "High" and ticket.description == "" then
  return {
    reject = {
      message = "High priority tickets need a description",
      fields = { description = "Required" }
    }
  }
end

return { ticket = ticket }
```

## Custom Create Pages

Custom create pages currently expose static project-scoped definitions and submit tickets through the normal backend ticket-create path. The planned Lua layer will compute validated form schema/defaults/options. Lua must never return raw HTML.

```lua
return {
  sections = {
    {
      title = "Request",
      fields = {
        { name = "title", type = "text", required = true },
        { name = "priority", type = "single_select", options = {"Low", "Medium", "High"} }
      }
    }
  }
}
```

## Webhooks

Incoming webhook definition CRUD, one-time bearer token creation/rotation, hashed token storage, the stable `POST /api/webhooks/incoming/{id}` endpoint, Lua/AI validation/logging, constrained Rayboard actions, and run history are implemented. Incoming webhook scripts receive `request.headers`, `request.query`, and `request.payload`, may call `rayboard.log(message)`, and may return a table that is stored in the automation run output. `rayboard.search`, `rayboard.get_ticket`, `rayboard.create_ticket`, `rayboard.update_ticket`, and `rayboard.comment` run as the configured actor user through normal RBAC. AI incoming webhooks receive the same request context in the prompt and may return `reject` or an `actions` array using action types `search`, `get_ticket`, `create_ticket`, `update_ticket`, and `comment`; returned actions are capped and execute as the configured actor user through normal service/RBAC paths. Disabled or deleted actor users cause execution to fail before Lua or AI runs. Outgoing webhook definitions are implemented with `event_types`, and matching domain events are persisted as queued outgoing delivery rows that snapshot event and webhook context. Delivery history is available through `GET /api/webhook-definitions/{webhook_id}/deliveries` and `GET /api/webhook-deliveries/{delivery_id}`; failed/canceled deliveries can be requeued with `POST /api/webhook-deliveries/{delivery_id}/retry`.

Outgoing Lua scripts receive `event`, `webhook`, and `delivery` tables and return a controlled outbound request shape:

```lua
return {
  method = "POST",
  path = "/events",
  query = { source = "rayboard" },
  headers = { ["X-Rayboard-Event"] = event.type },
  body = {
    event = event,
    webhook_id = webhook.id,
    delivery_id = delivery.id
  }
}
```

Outgoing AI webhooks receive the same context in the prompt and must return the same JSON request object. Outgoing requests are sent only under the configured `--outgoing-webhook-base-url`; relative paths are required, unsupported methods are rejected, request bodies are JSON, and `Host`/`Content-Length` headers cannot be set by automation.

Incoming example shape:

```lua
if request.payload.title == nil then
  rayboard.log("missing title")
  return { reject = { status = 400, message = "title is required" } }
end

local ticket, err = rayboard.create_ticket({
  project_id = request.payload.project_id,
  title = request.payload.title,
  labels = {"webhook"}
})
if err then error(err.message) end

return {
  accepted = true,
  ticket_id = ticket.id,
  key = ticket.key
}
```

## Planned Notifications and Shoutrrr

The first notification API slice lets users list their own notifications and mark them read or unread. In-app notification generation for comments and ticket updates is driven by durable `domain_events`, with processed/failed state stored on the event row. External notification delivery uses named Shoutrrr destinations, durable delivery rows, and the backend notification worker.

Notification policies and the delivery history/manual retry API are implemented as the queue foundation. Enabled matching global and project policies enqueue delivery rows when durable notification events are processed. Notification hooks can run before delivery rows are created. They may suppress a policy plan, replace the outbound message, replace the payload, or reduce `destination_ids` to a subset of the policy's already-allowed destinations. Hooks never receive raw Shoutrrr URLs or secrets.

Lua notification hooks receive a `notification` table with `context`, `policy`, `plan`, and `instructions`. AI notification hooks receive the same context in the prompt and must return the same validated JSON object. Hook output must not bypass RBAC, user preferences, destination visibility, or backend validation.

Example route shape:

```lua
if notification.plan.event_type == "ticket_assigned" then
  return {
    message = "Assigned: " .. notification.plan.message,
    destination_ids = { notification.policy.destination_ids[1] },
    payload = notification.plan.payload
  }
end

return { suppress = true }
```

## OpenRouter Providers and Planned AI Automation

AI automation will use OpenRouter only. Global admins can manage provider references through `/api/openrouter-providers`; these records contain a name, default model, allowed model list, timeout/output limits, enabled state, and a write-only API key. Responses never include the API key and only expose `status.api_key_set`. Provider create/update/delete changes are written to the security audit log without secret values.

Automation surfaces use the same nested `engine` object: `engine.type` is `lua` or `ai`, Lua uses `engine.script`, and AI uses `engine.prompt` plus `engine.provider_id`. The provider ID references the admin-managed OpenRouter configuration. Project users select only allowed provider/model configurations. AI output must be JSON matching a declared schema and must be validated before any effect is applied. AI output must never bypass RBAC, ticket validation, custom field validation, hooks, or API authorization.

AI execution is implemented for cron jobs, ticket hooks, incoming webhooks, outgoing webhook request shaping, and notification hook plan shaping as validated JSON-object output. AI cron action execution, custom create pages, notification hook preview, and notification hook run history are still **Planned**.

## Future WebAssembly Engine

WebAssembly automation is planned as a later engine after Lua and OpenRouter AI behavior are complete. The planned runtime is wazero (`https://github.com/wazero/wazero`) so Rayboard can keep a pure-Go, single-binary deployment model. WASM modules will use the same owner/actor, RBAC, timeout, memory, log, payload-size, and action-count limits as the equivalent Lua/AI surface, and will only receive constrained Rayboard host functions. Filesystem, shell, raw sockets, unrestricted HTTP, direct SQLite, raw Go pointers, Shoutrrr secrets, and OpenRouter keys must remain unavailable. Outputs must validate against the same structured schema before Rayboard applies ticket transformations, actions, outbound requests, or form definitions.
