# User Guide

The current browser UI is a small proof-of-concept board shell. Log in with the startup-generated admin credentials, then use the page at `/` to create projects, create tickets, select a project, and move tickets between `todo`, `in_progress`, and `done`.

Implemented browser workflows:

- login and logout;
- project list and project creation;
- ticket list for the selected project;
- ticket creation;
- ticket status changes;
- ticket comment list, creation, and deletion from each ticket card;
- ticket attachment list, upload, download, and delete from each ticket card;
- notification inbox listing with unread filter, read/unread toggle, refresh, and mark-all-read;
- sprint list, create, start, complete, delete, and ticket sprint assignment/removal for the selected project;
- component and version list/create/delete, version release/archive state changes, and ticket component/version assignment;
- roadmap epic list with schedule dates and child-ticket progress, plus ticket form fields for epics, parent epics, and roadmap dates;
- Account/API Tokens panel for viewing token metadata, creating tokens, and revoking your own tokens;
- text/CEL search plus saved-view list, create, apply, and delete;
- engine workbench tests for Lua, OpenRouter AI, and WASM automation engines.

API token secrets are shown once when created and are not listed later.

Implemented API-only user workflows:

- advanced saved-view metadata editing, grouping, and pagination;
- detailed sprint editing and sprint-state filtering;
- project workflow status list/replace APIs and board definition CRUD;
- board ticket listing by board definition;
- detailed component/version editing and filtering;
- ticket labels on ticket create/update/list/get/search payloads;
- custom field definition CRUD and typed ticket custom-field values;
- in-app notification inbox API, including unread filtering and read/unread state;
- current-user notification preferences and project notification defaults;
- global/project notification policy CRUD, Lua/AI notification hook CRUD, and delivery history/manual retry;
- incoming/outgoing webhook definition APIs, incoming execution, outgoing delivery history, and outgoing delivery retry;
- Lua cron job management, manual runs, run history, and cron Lua helpers for search, ticket create/update, ticket lookup, comments, and logging;
- custom ticket create-page definition, schema resolution, and submission APIs.

See [API Guide](api.md) for endpoint details.

## Planned Jira-Like Workflows

Backlog list/reorder, workflow status APIs, board definition CRUD, board ticket listing, ticket labels, custom fields, saved-view metadata, notification policies/hooks, Shoutrrr/external notification delivery, incoming/outgoing webhook workflows, ticket-hook management/preview, and custom create pages are currently API-only workflows. Rich backlog planning UI, board settings UI, board/backlog drag/drop, sprint report screens, burndown/velocity/burnup reports, release reports, richer roadmap timeline screens, richer component/version UI screens, label management UI, custom-field UI/search integration, advanced release planning, custom create-page rendering/settings screens, ticket-hook management screens, browser notification hook screens, and richer saved-view UI are **Planned**. Lua-backed and OpenRouter AI-backed dynamic custom create pages must return structured form definitions and options, not raw HTML. Remaining OpenRouter AI surfaces and persisted WebAssembly automation are also **Planned**.

## Notifications

The browser inbox exposes notification listing, unread filtering, individual read/unread toggles, refresh, and mark-all-read. Authenticated users can also update notification preferences through the API. Project notification managers can set project notification defaults, inspect delivery history, and manually retry failed external deliveries through the API. `read_at` is `null` while a notification is unread. Notifications for comments and ticket updates are generated from durable backend events so pending notifications can be processed after restart.

Browser badge UI, browser notification hook screens, and richer hook routing controls are **Planned**.

## Search

Current search supports full-text search over ticket title, description, comments, and attachment metadata such as filename and content type with SQLite FTS5, plus CEL-backed ticket filters. Roadmap date fields `start_date` and `due_date` are available for search filters, sort specs, and saved-view columns. Filters support boolean operators, comparisons, label membership, selected string helpers, `currentUser()`, `today()`, `now()`, and `custom.<field_key>` access for typed custom fields. See:

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- SQLite FTS5: https://www.sqlite.org/fts5.html
