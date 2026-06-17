# User Guide

The current browser UI is a small proof-of-concept board shell. Log in with the startup-generated admin credentials, then use the page at `/` to create projects, create tickets, select a project, and move tickets between `todo`, `in_progress`, and `done`.

Implemented browser workflows:

- login and logout;
- project list and project creation;
- ticket list for the selected project;
- ticket creation;
- ticket status changes.

Implemented API-only user workflows:

- comments on tickets;
- attachments on tickets;
- saved views with display mode, grouping, and pinned project-view metadata;
- text and filter search;
- sprint CRUD, start/complete actions, and ticket sprint assignment/removal;
- project workflow status list/replace APIs and board definition CRUD;
- board ticket listing by board definition;
- component CRUD, version/release CRUD, and ticket component/version assignment;
- roadmap data for scheduled project tickets;
- ticket labels on ticket create/update/list/get/search payloads;
- custom field definition CRUD and typed ticket custom-field values;
- in-app notification inbox API, including unread filtering and read/unread state;
- current-user notification preferences and project notification defaults;
- global/project notification policy CRUD and delivery history/manual retry;
- Lua cron job management, manual runs, run history, and cron Lua helpers for search, ticket create/update, ticket lookup, comments, and logging.
- custom ticket create-page definition, schema resolution, and submission APIs.

See [API Guide](api.md) for endpoint details.

## Planned Jira-Like Workflows

Backlog list/reorder, sprint CRUD, sprint start/complete, ticket sprint assignment, workflow status APIs, board definition CRUD, board ticket listing, component/version CRUD, ticket component/version assignment, roadmap data, ticket labels, custom fields, saved-view metadata, notification policies, Shoutrrr/external notification delivery, incoming webhook definitions/execution, ticket-hook management/preview, and custom create pages are currently API-only workflows. Rich backlog planning UI, board settings UI, board/backlog drag/drop, sprint report screens, burndown/velocity/burnup reports, release reports, roadmap timeline screens, component/version UI screens, label management UI, custom-field UI/search integration, advanced release planning, custom create-page rendering/settings screens, ticket-hook management screens, and richer saved-view UI are **Planned**. Dynamic custom create pages should return structured form definitions and options, not raw HTML. Outgoing webhooks, notification hooks, remaining OpenRouter AI surfaces, and future WebAssembly automation are also **Planned**.

## Notifications

The current notification slice is API-only. Authenticated users can list their own notifications, filter to unread notifications, mark individual notifications read or unread, and update their notification preferences. Project notification managers can set project notification defaults, inspect delivery history, and manually retry failed external deliveries. `read_at` is `null` while a notification is unread. Notifications for comments and ticket updates are generated from durable backend events so pending notifications can be processed after restart.

Browser inbox and badge UI, outgoing webhooks, and AI/Lua notification hooks are **Planned**.

## Search

Current search supports full-text search over ticket title, description, comments, and attachment metadata such as filename and content type with SQLite FTS5, plus a constrained filter expression subset. Roadmap date fields `start_date` and `due_date` are available for search filters, sort specs, and saved-view columns. Full CEL-backed queries are **Planned**. See:

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- SQLite FTS5: https://www.sqlite.org/fts5.html
