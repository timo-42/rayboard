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
- component CRUD, version/release CRUD, and ticket component/version assignment;
- roadmap data for scheduled project tickets;
- custom field definition CRUD and typed ticket custom-field values;
- in-app notification inbox API, including unread filtering and read/unread state;
- Lua cron job management, manual runs, run history, and cron Lua helpers for search, ticket create/update, ticket lookup, comments, and logging.

See [API Guide](api.md) for endpoint details.

## Planned Jira-Like Workflows

Backlog list/reorder, sprint CRUD, sprint start/complete, ticket sprint assignment, component/version CRUD, ticket component/version assignment, roadmap data, custom fields, and saved-view metadata are currently API-only workflows. Rich backlog planning UI, board UI, board/backlog drag/drop, sprint report screens, burndown/velocity/burnup reports, release reports, roadmap timeline screens, component/version UI screens, custom-field UI/search integration, advanced release planning, workflows, labels, custom create pages, and richer saved-view UI are **Planned**. Custom create pages should return structured form definitions and options, not raw HTML. Ticket hooks, webhooks, notification preferences, Shoutrrr/external notification delivery, notification policies, notification hooks, and OpenRouter AI automation are also **Planned**.

## Notifications

The current notification slice is API-only and in-app only. Authenticated users can list their own notifications, filter to unread notifications, and mark individual notifications read or unread. `read_at` is `null` while a notification is unread.

Browser inbox and badge UI, user/project notification preferences, external delivery, Shoutrrr destination settings, notification policies, delivery queues, webhooks, and AI/Lua notification hooks are **Planned**.

## Search

Current search supports full-text search over ticket title, description, and comments with SQLite FTS5, plus a constrained filter expression subset. Roadmap date fields `start_date` and `due_date` are available for search filters, sort specs, and saved-view columns. Full CEL-backed queries are **Planned**. See:

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- SQLite FTS5: https://www.sqlite.org/fts5.html
