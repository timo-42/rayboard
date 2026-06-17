# Authentication and RBAC

Rayboard authentication is backend-owned. Browser users authenticate with sessions; scripts can authenticate with bearer API tokens.

## Browser Sessions

`POST /api/login` accepts `{"spec":{"username":"...","password":"..."}}`. On success it returns a session resource and sets:

- `rayboard_session`: opaque `HttpOnly` session cookie.
- `rayboard_csrf`: readable CSRF cookie used by the frontend for mutating requests.

Both cookies use `SameSite=Lax` and `Secure` when served over HTTPS. Sessions are stored hashed in SQLite and expire after the service TTL, currently 24 hours.

`POST /api/logout` revokes the current session when present and clears both cookies.

## CSRF

For cookie-authenticated requests, mutating methods require:

```http
X-CSRF-Token: <value from rayboard_csrf cookie>
```

Bearer-token requests do not require CSRF because browsers do not attach bearer credentials automatically.

Swagger UI authorization models bearer API tokens as sufficient by themselves. Cookie-authenticated requests are modeled with `sessionCookie`; mutating session requests still require `X-CSRF-Token`, and the embedded Swagger UI sends that header automatically from the readable CSRF cookie when present. CSRF is tied to browser session auth and not to API tokens.

## API Tokens

Authenticated users can create, list, and revoke their own API tokens:

```bash
curl -b cookies.txt -H "X-CSRF-Token: $CSRF" \
  -H "Content-Type: application/json" \
  -d '{"spec":{"name":"local-script"}}' \
  http://127.0.0.1:8081/api/tokens
```

The token secret is returned once on creation as `status.token`. SQLite stores only a hash. Use it as:

```bash
curl -H "Authorization: Bearer $RAYBOARD_TOKEN" \
  http://127.0.0.1:8081/api/me
```

## Effective Permissions

Users can inspect their own effective permissions:

```bash
curl -b cookies.txt \
  'http://127.0.0.1:8081/api/me/effective-permissions?scope=project&project_id=project_...'
```

RBAC admins with global `roles:read` can inspect another user:

```bash
curl -H "Authorization: Bearer $RAYBOARD_TOKEN" \
  'http://127.0.0.1:8081/api/users/user_.../effective-permissions?scope=global'
```

Responses use the standard envelope. The requested scope is in `spec`; computed grants are in `status.permissions`.

## Disabled and Deleted Users

Disabled users cannot log in or authenticate with existing sessions or API tokens. Disabling a user revokes that user's active sessions and API tokens. Deleting a user is a soft delete that renames the username, disables the user, and revokes active credentials.

## RBAC Model

Authorization is deny-by-default and evaluates effective permissions from:

- direct user role bindings;
- group memberships plus group role bindings;
- global or project scope.

Implemented role binding subjects are `user` and `group`. Implemented scopes are `global` and `project`.

Built-in roles:

| Role | Purpose |
| --- | --- |
| `global_admin` | Grants `*`. |
| `global_user_manager` | User, group, and role administration. |
| `project_owner` | Full implemented project permissions plus project settings/automation/notification permissions. |
| `project_admin` | Project read, tickets, comments, attachments, sprints, boards, fields, views, notifications, webhooks, automations. |
| `project_member` | Project read, ticket read/write, comments, attachments. |
| `project_viewer` | Project read and ticket read. |
| `automation_manager` | Project read, automations, webhooks. |
| `notification_manager` | Project read, notifications. |

Common implemented permission checks include `users:read`, `users:write`, `groups:read`, `groups:write`, `roles:read`, `roles:bind`, `projects:read`, `projects:write`, `tickets:read`, `tickets:write`, `comments:write`, `attachments:write`, `views:manage`, `notifications:manage`, and `ai:manage`.

Global `ai:manage` is required to manage OpenRouter provider references because those records hold global secret material.

Personal notification preferences require authentication but no RBAC permission because users can only access their own preference resource.

`notifications:manage` is required to manage project notification defaults, notification policies, Shoutrrr notification destinations, delivery history, and manual delivery retry. Global policy, destination, and delivery APIs require the permission at global scope; project preference/default, policy, destination, and delivery APIs require it at the target project scope.

`webhooks:manage` is required to create, list, update, delete, rotate, and inspect run history for project webhook definitions. Incoming webhook receiver calls authenticate with a webhook-specific bearer token rather than a user session or API token. Lua execution records run history and logs, and constrained Rayboard helpers run as the configured actor user. Disabled or deleted actor users cannot execute incoming webhook scripts.

`automations:manage` is required to create, list, update, and delete project ticket hooks and custom ticket create pages. Hooks execute during normal ticket service calls and receive the authenticated principal that triggered the ticket operation. Custom create-page submissions use the submitter's normal ticket permissions.

## Current Limitations

Inspect roles with `GET /api/roles`, bindings with `GET /api/role-bindings`, and computed grants with the effective-permissions endpoints. Project-scoped role assignment is implemented through the generic role binding endpoint.
