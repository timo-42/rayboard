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
- incoming webhooks: **Planned**;
- outgoing webhooks: **Planned**;
- notification hooks: **Planned**.

Every surface should enforce timeouts, max script size, max log size, max input/output size, max JSON input/output bytes, max table nesting depth, and max action count where actions exist.

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

Go-backed Rayboard functions exposed to Lua should accept and return plain Lua tables using these same rules:

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

## Cron Jobs

Cron jobs use robfig cron for schedules and GopherLua for execution. The first public API/scheduler slice supports cron job CRUD, manual runs, and run history. Jobs act as their owner user, inherit the owner's current RBAC permissions at run time, and should not overlap by default.

Cron job management requires automation permissions. Disabled users cannot run owned cron jobs.

Run history uses the shared automation run-history model. A run record is used for scheduled and manual runs and may include trigger type, job identity, owner/project context when applicable, input/output summaries, logs, status, error details, start/finish timestamps, duration, and applied limits. Run history must not expose secrets.

Implemented cron Lua helpers:

- `rayboard.log(message)`
- `rayboard.search({ project_id, filter, text, sort, limit, cursor })`
- `rayboard.get_ticket(ticket_id)`
- `rayboard.create_ticket({ project_id, title, description, status, priority, type, reporter_id, assignee_id, parent_ticket_id, rank })`
- `rayboard.update_ticket(ticket_id, { title, description, status, priority, type, assignee_id, parent_ticket_id, rank })`
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
  rayboard.log(ticket.key .. ": " .. ticket.title)
end
```

Example create/comment shape:

```lua
local ticket, err = rayboard.create_ticket({
  project_id = "project_...",
  title = "Investigate recurring alert"
})
if err then return { error = err.message } end

local comment, comment_err = rayboard.comment({
  ticket_id = ticket.id,
  body = "Created by scheduled triage."
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

## Planned Webhooks

Incoming webhooks authenticate first, then Lua maps request data to allowed Rayboard actions using a configured actor user. Outgoing webhooks shape a controlled outbound request subject to allowlists, timeouts, max payload sizes, retries, and delivery history.

Incoming example shape:

```lua
local body = json.decode(request.body)
if body.title == nil then
  return { reject = { status = 400, message = "title is required" } }
end

return {
  actions = {
    { type = "create_ticket", input = { title = body.title, project_id = context.project_id } }
  }
}
```

## Planned Notifications and Shoutrrr

External notifications will use named Shoutrrr destinations. Hooks may filter, suppress, transform, enrich, and route notification plans by destination name, but must never receive raw Shoutrrr URLs or secrets.

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

## Planned OpenRouter AI Automation

AI automation will use OpenRouter only. Global admins configure the API key, default model, allowed models, timeout, and limits. Project users select only allowed models. AI output must be JSON matching a declared schema and must be validated before any effect is applied. AI output must never bypass RBAC, ticket validation, custom field validation, hooks, or API authorization.
