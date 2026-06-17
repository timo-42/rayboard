# Automation Engines Contract

## Engines

Automation-capable surfaces support:

```json
{
  "engine": {
    "type": "lua",
    "script": "rayboard.log(\"hello\")"
  }
}
```

`engine.type` is the discriminator. Lua uses `engine.script`. AI uses `engine.prompt` plus `engine.provider_id`, where the provider ID references an admin-managed OpenRouter configuration with model, API key/secret material, limits, and defaults. OpenAPI schemas must represent this as `oneOf` with a discriminator on `type`, so `lua` requires `script` and `ai` requires `prompt` plus `provider_id`.

Lua uses GopherLua. AI uses OpenRouter only in v1.

Surfaces:

- cron jobs
- ticket hooks, using `ticket_hook_before` and `ticket_hook_after` in engine test contexts
- custom ticket create pages
- incoming webhooks, using `incoming_webhook` in engine test contexts
- outgoing webhooks, using `outgoing_webhook` in engine test contexts
- notification hooks

## Engine Workbench

The generic engine workbench endpoint is:

```text
POST /api/engines/test
```

It accepts a Kubernetes-style request body:

```json
{
  "spec": {
    "engine": {
      "type": "lua",
      "script": "return { ok = true }"
    },
    "surface": "scratch",
    "context": {},
    "input": {
      "title": "Preview"
    },
    "dry_run": true
  }
}
```

The response uses `metadata`, `spec`, and `status`. The response `spec` must redact Lua source and AI prompts. The response `status` contains the execution state, validated output, error when present, and timing metadata available from the run record.

Use `surface: "scratch"` for a playground run that only validates generic JSON input/output. Use concrete surfaces such as `ticket_hook_before`, `ticket_hook_after`, `custom_create_page`, `incoming_webhook`, `outgoing_webhook`, or `notification_hook` to test against a surface-specific contract. Missing `surface` defaults to `scratch`.

For `custom_create_page`, workbench output must be structured form data and include at least one of `field_layout`, `defaults`, or `description`. `field_layout` must be an array of objects, nested `fields` arrays are validated recursively, `defaults` must be an object, `description` must be a string, and raw `html` fields are rejected. Invalid surface output is a failed run response, not a transport-level API error.

Workbench execution must use the same engine discriminator, JSON/table conversion, RBAC model, actor resolution, timeouts, logs, payload limits, secret redaction, and run-history persistence as the corresponding real surface. In the current POC the endpoint is always mutation-free and normalizes `dry_run` to `true`.

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
- one table-to-table adapter path for Go-backed Rayboard helper payloads
- one documentation source for helper behavior and examples under `/docs`

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

The conversion layer must be explicit and testable. Do not expose raw Go structs, pointers, `map[string]any` values with inconsistent semantics, database handles, service handles, HTTP clients, or unrestricted userdata to Lua. Surface adapters should translate from Lua tables into the same DTOs used by API/service code, then translate responses back to Lua tables.

Common helper result shape:

```lua
local result, err = rayboard.some_action({ id = "123" })
if err ~= nil then
  reject(err.message)
end
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

## Documentation Contract

Every automation surface must document:

- supported `engine` values
- owner/actor and RBAC behavior
- allowed Lua helpers and denied capabilities
- JSON encode/decode behavior
- Go<->Lua table conversion rules
- input/output limits
- AI prompt and structured output schema behavior
- run history and secret redaction behavior
