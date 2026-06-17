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
- create, update, disable, and delete OpenRouter provider references for future AI automation.

The current browser UI does not expose admin screens yet.

Security/admin-sensitive actions are written to the SQLite `audit_log` table. Current audited events include login failures, session creation/logout, API token creation/revocation, user create/disable/enable/delete, group creation and membership changes, role binding create/delete, and OpenRouter provider create/update/delete. Audit payloads intentionally exclude generated passwords, plaintext API tokens, password hashes, session secrets, and future webhook/Shoutrrr/OpenRouter secrets.

## RBAC

RBAC is group-aware and deny-by-default. Role bindings can target users or groups and can be global or project-scoped. Built-in roles include `global_admin`, `global_user_manager`, `project_owner`, `project_admin`, `project_member`, `project_viewer`, `automation_manager`, and `notification_manager`.

See [Authentication and RBAC](auth-rbac.md) for the implemented model and current permissions.

## Settings

Admin, project, and board settings pages are **Planned**. Future settings should cover:

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

Incoming and outgoing webhook definitions are currently API-only. Project webhooks live under `/api/projects/{project_id}/webhooks` and require project `webhooks:manage`. Incoming webhook tokens are returned once on create or rotation, stored only as hashes, and accepted at `/api/webhooks/incoming/{webhook_id}` with `Authorization: Bearer <webhook-token>`. Outgoing webhook definitions do not have tokens; they subscribe to domain event names through `spec.event_types` and can be persisted as queued deliveries. Queued delivery history is available at `/api/webhook-definitions/{webhook_id}/deliveries` and `/api/webhook-deliveries/{delivery_id}`. Delivery processing, retries, and outbound HTTP are planned follow-up work.

Custom CSS is planned as an override layer only. The first implementation should not allow arbitrary template changes.

## Notifications

The current notification implementation includes per-user in-app notification listing/read state, current-user notification preferences, project notification defaults, API-only notification policy CRUD, API-only Shoutrrr destination CRUD for global and project scopes, delivery history/manual retry, and a backend worker that sends due queued deliveries.

Outgoing webhook delivery and AI/Lua notification hooks are **Planned**.
