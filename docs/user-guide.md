# User Guide

The current browser UI is a small proof-of-concept website. Log in with the startup-generated admin credentials, then use `/` for the dashboard overview, `/projects` or `/projects/{project_id}` for project work, `/issues/{ticket_id}` for an issue detail page, `/profile` for account/API tokens, `/rbac` for administration, `/settings` for global/user settings, `/search` for saved searches, and `/automation` for the engine workbench and basic project ticket-hook management.

Implemented browser workflows:

- login and logout;
- dashboard summary with recently modified tickets, biggest projects, active sprints, unread notifications, and ticket/project counts;
- project list and project creation;
- project detail pages under `/projects/{project_id}`;
- issue detail pages under `/issues/{ticket_id}` with comments, attachments, custom fields, and activity history;
- ticket list for the selected project;
- ticket creation;
- ticket status changes;
- ticket label entry, display, and update from ticket cards;
- ticket comment list, creation, and deletion from each ticket card;
- ticket attachment list, upload, download, and delete from each ticket card;
- notification inbox listing with unread filter, read/unread toggle, refresh, and mark-all-read;
- sprint list, create, start, complete, delete, and ticket sprint assignment/removal for the selected project;
- component and version list/create/delete, version release/archive state changes, and ticket component/version assignment;
- roadmap epic list with schedule dates and child-ticket progress, plus ticket form fields for epics, parent epics, and roadmap dates;
- custom field list/create/delete for the selected project, plus JSON custom-field values on ticket create and ticket cards;
- Profile/API Tokens page for viewing user metadata, token metadata, creating tokens, and revoking your own tokens;
- RBAC page for users, groups, roles, and role-binding summaries when permitted;
- Settings page for global settings, OpenRouter provider management, Shoutrrr notification destination management and test-send, notification policy CRUD, and security audit-log inspection when permitted, plus personal notification preferences for every signed-in user;
- text/CEL search plus saved-view list, create, apply, and delete;
- engine workbench tests for Lua, OpenRouter AI, and WASM automation engines;
- basic cron job list, create, delete, enable/disable, manual run, and run-output inspection for the selected project;
- basic project webhook list, create, delete, enable/disable, incoming token rotation, run history, and outgoing delivery inspection for the selected project;
- basic project ticket-hook list, create, delete, enable/disable, and preview controls for the selected project.

API token secrets are shown once when created and are not listed later.

Implemented API-only user workflows:

- advanced saved-view metadata editing, grouping, and pagination;
- detailed sprint editing and sprint-state filtering;
- project workflow status list/replace APIs and board definition CRUD;
- board ticket listing by board definition;
- detailed component/version editing and filtering;
- custom field update APIs beyond browser delete/recreate;
- project notification defaults;
- Lua/AI notification hook CRUD and delivery history/manual retry;
- incoming webhook execution APIs and advanced outgoing delivery retry workflows;
- custom ticket create-page definition, schema resolution, and submission APIs.

See [API Guide](api.md) for endpoint details.

## Planned Jira-Like Workflows

Backlog list/reorder, workflow status APIs, board definition CRUD, board ticket listing, saved-view metadata, notification hooks, and custom create pages are currently API-only workflows. Rich backlog planning UI, board settings UI, board/backlog drag/drop, sprint report screens, burndown/velocity/burnup reports, release reports, richer roadmap timeline screens, richer component/version UI screens, label management UI beyond direct ticket editing, richer custom-field layout/search integration, advanced release planning, custom create-page rendering/settings screens, browser notification-hook automation screens, richer cron/webhook/ticket-hook editing/history screens, and richer saved-view UI are **Planned**. Lua-backed and OpenRouter AI-backed dynamic custom create pages must return structured form definitions and options, not raw HTML. Remaining OpenRouter AI surfaces and persisted WebAssembly automation are also **Planned**.

## Notifications

The browser inbox exposes notification listing, unread filtering, individual read/unread toggles, refresh, and mark-all-read. Authenticated users can update personal notification preferences in `/settings`. Notification managers can create, edit, enable/disable, rotate, test, and delete global or selected-project Shoutrrr destinations in `/settings`; destination URLs are write-only and are not shown after save. Project notification managers can set project notification defaults, inspect delivery history, and manually retry failed external deliveries through the API. `read_at` is `null` while a notification is unread. Notifications for comments and ticket updates are generated from durable backend events so pending notifications can be processed after restart.

Browser badge UI, browser notification hook screens, and richer hook routing controls are **Planned**.

## Search

Current search supports full-text search over ticket title, description, comments, and attachment metadata such as filename and content type with SQLite FTS5, plus CEL-backed ticket filters. Roadmap date fields `start_date` and `due_date` are available for search filters, sort specs, and saved-view columns. Filters support boolean operators, comparisons, label membership, selected string helpers, `currentUser()`, `today()`, `now()`, and `custom.<field_key>` access for typed custom fields. See:

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- SQLite FTS5: https://www.sqlite.org/fts5.html
