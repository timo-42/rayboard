# Admin Guide

Rayboard is currently a POC. On every backend or combined startup, the `admin` password is randomized, written to SQLite as a password hash, and printed to stdout. Treat the printed password as a local development credential.

Admin workflows currently available through the API:

- add users;
- disable users;
- soft-delete users;
- create and revoke API tokens for the authenticated user;
- create groups;
- add and remove group members;
- list roles;
- create and delete role bindings.
- create, update, disable, and delete OpenRouter provider references for AI automation.
- read and update global settings for attachment policy, webhook allowlist metadata, demo warning, backup availability flag, and system health note.

The current browser UI does not expose admin screens yet.

Security/admin-sensitive actions are written to the SQLite `audit_log` table. Current audited events include login failures, session creation/logout, API token creation/revocation, user create/disable/enable/delete, group creation and membership changes, role binding create/delete, OpenRouter provider create/update/delete, and global settings updates. Audit payloads intentionally exclude generated passwords, plaintext API tokens, password hashes, session secrets, and future webhook/Shoutrrr/OpenRouter secrets.

## RBAC

RBAC is group-aware and deny-by-default. Role bindings can target users or groups and can be global or project-scoped. Built-in roles include `global_admin`, `global_user_manager`, `project_owner`, `project_admin`, `project_member`, `project_viewer`, `automation_manager`, and `notification_manager`.

See [Authentication and RBAC](auth-rbac.md) for the implemented model and current permissions.

## Settings

Browser admin, project, and board settings pages are **Planned**.

Global settings are currently API-only at `/api/settings` and require global `settings:manage`. The implemented global settings cover:

- attachment maximum size and allowed attachment content types;
- webhook allowed base URL metadata for future allowlist-driven delivery;
- demo warning visibility;
- backup availability flag and system health note.

Attachment uploads enforce the configured max size and content-type allowlist. An empty content-type allowlist permits all content types.

Future settings should cover:

- user, group, role, and token administration;
- project ownership and project settings;
- board settings;
- custom CSS override layers for projects and boards;
- automation, webhook, notification, and OpenRouter configuration;
- Shoutrrr destination definitions.

OpenRouter provider configuration is currently API-only at `/api/openrouter-providers` and requires global `ai:manage`. Provider API keys are write-only; responses return `status.api_key_set` instead of the key.

Project notification defaults are currently API-only at `/api/projects/{project_id}/notification-preferences` and require project `notifications:manage`.

Notification policy CRUD is currently API-only. Global policies live under `/api/notification-policies`; project policies live under `/api/projects/{project_id}/notification-policies`. Policies validate event types and destination visibility. Delivery history is available under `/api/notification-deliveries` and `/api/projects/{project_id}/notification-deliveries`, with manual retry at `/api/notification-deliveries/{delivery_id}/retry`.

Shoutrrr destination configuration is currently API-only. Global destinations live under `/api/notification-destinations` and require global `notifications:manage`; project destinations live under `/api/projects/{project_id}/notification-destinations` and require project `notifications:manage`. Destination URLs are write-only, can be rotated with `PATCH`, and can be verified with `POST /api/notification-destinations/{destination_id}/test-send`.

Incoming and outgoing webhook definitions are currently API-only. Project webhooks live under `/api/projects/{project_id}/webhooks` and require project `webhooks:manage`. Incoming webhook tokens are returned once on create or rotation, stored only as hashes, and accepted at `/api/webhooks/incoming/{webhook_id}` with `Authorization: Bearer <webhook-token>`. Outgoing webhook definitions do not have tokens; they subscribe to domain event names through `spec.event_types`, persist queued deliveries, shape outbound requests with Lua/AI, and send relative paths below the configured `--outgoing-webhook-base-url`. Queued delivery history is available at `/api/webhook-definitions/{webhook_id}/deliveries` and `/api/webhook-deliveries/{delivery_id}`; failed/canceled deliveries can be requeued at `/api/webhook-deliveries/{delivery_id}/retry`.

Custom CSS is planned as an override layer only. The first implementation should not allow arbitrary template changes.

## Notifications

The current notification implementation includes per-user in-app notification listing/read state, current-user notification preferences, project notification defaults, API-only notification policy CRUD, API-only Shoutrrr destination CRUD for global and project scopes, delivery history/manual retry, Lua/AI notification hooks, and a backend worker that sends due queued deliveries.

Notification hook preview, run history, and browser management screens are **Planned**.
