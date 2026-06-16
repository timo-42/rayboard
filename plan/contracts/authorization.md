# Authorization Contract

## Principal

Every authenticated operation resolves to a principal:

```text
principal:
  user_id
  auth_kind: session | api_token | cron | webhook | internal_admin_demo
  actor_user_id
  effective_permissions(scope)
```

Cron jobs, webhooks, Lua, and AI act as configured users. They never get extra privileges from the scheduler or automation subsystem.

## RBAC Model

Persist:

- users
- groups
- group memberships
- roles
- permissions
- role bindings

Role bindings:

- target: user or group
- scope: global or project
- role: built-in role for v1

Authorization is deny-by-default.

## Built-In Roles

Seed built-in roles and make them non-deletable:

- global admin
- global user manager
- project owner
- project admin
- project member
- project viewer
- automation manager
- notification manager

Custom roles can be added later unless needed to unblock implementation.

## Permission Namespaces

Use permission strings like:

```text
users:read
users:write
groups:read
groups:write
roles:read
roles:bind
projects:read
projects:write
tickets:read
tickets:write
comments:write
attachments:write
sprints:manage
boards:manage
fields:manage
views:manage
notifications:manage
webhooks:manage
automations:manage
settings:manage
demo:reset
ai:manage
```

Wildcard permissions such as `tickets:*` may be stored for built-in roles, but the evaluator must normalize checks through a single function.

## Required Evaluator API

Service code should call one evaluator rather than open-coding role checks:

```text
Can(principal, permission, scope) bool
Require(principal, permission, scope) error
EffectivePermissions(user_id, scope) []permission
```

Scope examples:

- global scope
- project scope
- ticket scope resolves to project scope
- board/sprint/component/version/custom field resolves to project scope

## Immediate Revocation

Group membership and role binding changes must affect:

- current sessions
- API tokens
- cron jobs
- webhook actors
- Lua hooks
- AI automations
- notification hooks

Do not cache effective permissions longer than one request/run without explicit invalidation.
