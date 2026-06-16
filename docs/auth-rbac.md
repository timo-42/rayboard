# Authentication and RBAC

Rayboard authentication is backend-owned. Browser users authenticate with sessions; scripts can authenticate with bearer API tokens.

## Browser Sessions

`POST /api/login` accepts a username and password. On success it returns the user and sets:

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

## API Tokens

Authenticated users can create, list, and revoke their own API tokens:

```bash
curl -b cookies.txt -H "X-CSRF-Token: $CSRF" \
  -H "Content-Type: application/json" \
  -d '{"name":"local-script"}' \
  http://127.0.0.1:8081/api/tokens
```

The token secret is returned once on creation as `token`. SQLite stores only a hash. Use it as:

```bash
curl -H "Authorization: Bearer $RAYBOARD_TOKEN" \
  http://127.0.0.1:8081/api/me
```

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

Common implemented permission checks include `users:read`, `users:write`, `groups:read`, `groups:write`, `roles:read`, `roles:bind`, `projects:read`, `projects:write`, `tickets:read`, `tickets:write`, `comments:write`, `attachments:write`, and `views:manage`.

## Current Limitations

There is no dedicated effective-permissions endpoint yet. Inspect roles with `GET /api/roles` and bindings with `GET /api/role-bindings`. Project-scoped role assignment is implemented through the generic role binding endpoint.

