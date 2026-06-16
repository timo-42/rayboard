# Automation Engines Contract

## Engines

Automation-capable surfaces support:

```text
engine = lua | ai
```

Lua uses GopherLua. AI uses OpenRouter only in v1.

Surfaces:

- cron jobs
- ticket hooks
- custom ticket create pages
- incoming webhooks
- outgoing webhooks
- notification hooks

## Common Run Record

Every automation run records:

- automation id/type
- engine
- actor/owner user
- project scope when applicable
- input summary
- validated output
- logs
- error
- started/finished time
- duration
- limits applied

AI runs also record:

- OpenRouter model
- usage metadata where available
- no API key or secrets

## Lua Rules

Do not expose:

- filesystem
- shell
- raw sockets
- direct SQLite
- unrestricted HTTP
- full backend service/store handles

Expose only surface-specific helpers.

All Lua-capable surfaces must reuse one shared sandbox package:

- one GopherLua state setup path per run
- one Go<->Lua conversion layer
- one JSON module implementation
- one limit/error model for JSON/table conversion

The shared JSON module is available as both:

```lua
json.encode(value)
json.decode(text)
rayboard.json.encode(value)
rayboard.json.decode(text)
```

JSON decode rules:

- objects become Lua tables with string keys
- arrays become 1-indexed Lua array tables
- strings, booleans, and numbers map directly
- JSON null becomes a stable `json.null` sentinel

JSON encode rules:

- Lua array tables become JSON arrays
- Lua string-key tables become JSON objects
- `json.null` becomes JSON null
- mixed string/numeric-key tables are rejected
- sparse array tables are rejected
- recursive tables are rejected
- functions, userdata, threads, raw Go pointers, and unsupported values are rejected
- non-finite numbers are rejected

Go-backed Rayboard helper functions exposed to Lua must accept and return plain Lua tables using the same conversion rules. Example:

```lua
local ticket, err = rayboard.create_ticket({
  title = "Bug",
  assignee_id = json.null
})
```

All Lua surfaces must enforce:

- timeout
- max script size
- max log size
- max input/output size
- max JSON input bytes
- max JSON output bytes
- max table/JSON nesting depth
- max action count where actions exist

## AI Rules

AI prompts receive the same context that Lua receives for the same surface.

AI output must:

- be JSON
- match the surface-specific schema
- be validated before any effect
- never bypass RBAC, ticket validation, custom field validation, hooks, or API authorization

Free-form text is never applied directly.

## Surface Effects

Cron:

- may return allowed Rayboard actions
- acts as owner user
- no overlapping runs by default

Ticket hooks:

- `before` may reject or transform pending payload
- `after` may inspect/log only
- no general API client

Custom create pages:

- return form schema/defaults/options
- never raw HTML

Incoming webhooks:

- authenticate first
- Lua/AI validates and maps request to allowed actions
- acts as configured actor user

Outgoing webhooks:

- event enqueues delivery
- Lua/AI shapes outbound request
- destination allowlist enforced
- does not block originating operation

Notification hooks:

- transform/suppress/route notification plans
- route by named destination only
- never see Shoutrrr URLs or secrets
