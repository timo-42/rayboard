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

The current browser UI does not expose admin screens yet.

Security/admin-sensitive actions are written to the SQLite `audit_log` table. Current audited events include login failures, session creation/logout, API token creation/revocation, user create/disable/enable/delete, group creation and membership changes, and role binding create/delete. Audit payloads intentionally exclude generated passwords, plaintext API tokens, password hashes, session secrets, and future webhook/Shoutrrr/OpenRouter secrets.

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

Custom CSS is planned as an override layer only. The first implementation should not allow arbitrary template changes.

## Notifications

The current notification implementation is limited to per-user in-app notification listing and read/unread state through the API. There are no admin notification settings in this slice.

Shoutrrr destinations, secret rotation, test-send, notification preferences, notification policies, external delivery queues, delivery history/retry, webhooks, and AI/Lua notification hooks are **Planned**.
