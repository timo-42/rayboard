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

All Lua surfaces must enforce:

- timeout
- max script size
- max log size
- max input/output size
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
