# Automation and Lua

Automation public surfaces are mostly **Planned**. The current implementation includes a shared Lua JSON sandbox foundation in `internal/backend/luasandbox`, automation run history persistence in `internal/backend/automation`, and the first cron jobs API/scheduler slice for Lua cron job CRUD, manual runs, and run history.

Relevant upstream references:

- GopherLua: https://github.com/yuin/gopher-lua
- robfig cron: https://pkg.go.dev/github.com/robfig/cron/v3
- robfig cron source: https://github.com/robfig/cron
- OpenRouter docs: https://openrouter.ai/docs
- Shoutrrr: https://github.com/containrrr/shoutrrr
- Shoutrrr docs: https://containrrr.dev/shoutrrr/

## Lua Sandbox Foundation

Lua surfaces use one shared sandbox runtime package and one shared Go/Lua conversion layer. Sandboxes must not expose filesystem access, shell execution, raw sockets, unrestricted HTTP, direct SQLite handles, raw Go pointers, backend store/service handles, Shoutrrr secrets, or OpenRouter keys.

Lua-capable surfaces:

- cron jobs: first API/scheduler slice implemented;
- ticket hooks: **Planned**;
- custom ticket create pages: **Planned**;
- incoming webhooks: definition CRUD, token auth, Lua validation/logging, and run history implemented;
- outgoing webhooks: **Planned**;
- notification hooks: **Planned**.

Every surface should enforce timeouts, max script size, max log size, max input/output size, max JSON input/output bytes, max table nesting depth, and max action count where actions exist. The current shared JSON defaults are 1 MiB max JSON input, 1 MiB max encoded JSON output, and 64 levels max nesting depth.

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

Cron jobs use robfig cron for schedules and GopherLua for execution. The first public API/scheduler slice supports cron job CRUD, manual runs, and run history. Jobs act as their owner user, inherit the owner's current RBAC permissions at run time, and should not overlap by default.

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
  filter = 'labels == "automation"',
  limit = 10
})
if search_err then return { error = search_err.message } end

local comment, comment_err = rayboard.comment({
  ticket_id = ticket.id,
  body = "Created by scheduled triage for " .. tostring(#matching.items) .. " matching tickets."
})
if comment_err then return { error = comment_err.message } end
```

## Planned Ticket Hooks

Ticket hooks are project-scoped. `before` hooks may validate or transform pending payloads; `after` hooks may inspect and log but cannot alter committed data.

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

## Planned Custom Create Pages

Custom create pages return validated form schema/defaults/options. Lua must never return raw HTML.

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

Incoming webhook definition CRUD, one-time bearer token creation/rotation, hashed token storage, the stable `POST /api/webhooks/incoming/{id}` endpoint, Lua validation/logging, and run history are implemented. Incoming webhook scripts receive `request.headers`, `request.query`, and `request.payload`, may call `rayboard.log(message)`, and may return a table that is stored in the automation run output. Ticket-mutating Rayboard action helpers for incoming webhooks are still **Planned** and will run as the configured actor user through normal RBAC. Outgoing webhooks are **Planned** and will shape a controlled outbound request subject to allowlists, timeouts, max payload sizes, retries, and delivery history.

Incoming example shape:

```lua
if request.payload.title == nil then
  rayboard.log("missing title")
  return { reject = { status = 400, message = "title is required" } }
end

return {
  accepted = true,
  title = request.payload.title
}
```

## Planned Notifications and Shoutrrr

The first notification API slice lets users list their own notifications and mark them read or unread. In-app notification generation for comments and ticket updates is driven by durable `domain_events`, with processed/failed state stored on the event row. External notification delivery uses named Shoutrrr destinations, durable delivery rows, and the backend notification worker.

Notification policies and the delivery history/manual retry API are implemented as the queue foundation. Enabled matching global and project policies enqueue delivery rows when durable notification events are processed. Notification hooks are **Planned** and may filter, suppress, transform, enrich, and route notification plans by destination name, but must never receive raw Shoutrrr URLs or secrets.

Webhooks and AI/Lua notification hooks are **Planned**. AI notification hooks must use the same validated notification-plan shape as Lua hooks and must not bypass RBAC, user preferences, destination visibility, or backend validation.

Example route shape:

```lua
if event.type == "ticket.assigned" then
  return {
    deliveries = {
      { destination = "team-chat", title = message.title, body = message.body }
    }
  }
end

return { suppress = true }
```

## OpenRouter Providers and Planned AI Automation

AI automation will use OpenRouter only. Global admins can manage provider references through `/api/openrouter-providers`; these records contain a name, default model, allowed model list, timeout/output limits, enabled state, and a write-only API key. Responses never include the API key and only expose `status.api_key_set`. Provider create/update/delete changes are written to the security audit log without secret values.

Automation surfaces use the same nested `engine` object: `engine.type` is `lua` or `ai`, Lua uses `engine.script`, and AI uses `engine.prompt` plus `engine.provider_id`. The provider ID references the admin-managed OpenRouter configuration. Project users select only allowed provider/model configurations. AI output must be JSON matching a declared schema and must be validated before any effect is applied. AI output must never bypass RBAC, ticket validation, custom field validation, hooks, or API authorization.

Actual AI execution for cron jobs, ticket hooks, custom create pages, webhooks, and notification hooks is still **Planned**.
