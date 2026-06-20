# User Guide

The current browser UI is a small proof-of-concept website. Log in with the startup-generated admin credentials, then use `/` for the dashboard overview, `/projects` or `/projects/{project_id}` for project work, `/issues/{ticket_id}` for an issue detail page, `/profile` for account/API tokens, `/rbac` for administration, `/settings` for global/user settings, `/search` for saved searches, and `/automation` for the engine workbench and basic project ticket-hook management.

Implemented browser workflows:

- login and logout;
- dashboard summary with recently modified tickets, biggest projects, active sprints, unread notifications, and ticket/project counts;
- project list and project creation;
- project detail pages under `/projects/{project_id}`;
- issue detail pages under `/issues/{ticket_id}` with watch/unwatch controls, linked issues, comments, attachments, custom fields, delete actions, and activity history;
- ticket list for the selected project with label, component, and version filters;
- ticket creation and soft delete;
- ticket status changes;
- ticket label entry, display, update from ticket cards, project label filtering with ticket counts, and project label administration with description/color metadata;
- ticket watch/unwatch controls and watcher lists from ticket cards and issue detail pages;
- ticket link list, creation, and removal from ticket cards and issue detail pages;
- ticket comment list, creation, deletion, and `@username` mention notifications from each ticket card;
- ticket attachment list, upload, download, and delete from each ticket card;
- notification inbox listing with a persistent unread badge, unread filter, mention notification labels, read/unread toggle, refresh, and mark-all-read;
- sprint list, create, edit, state filtering, start, complete, delete, and ticket sprint assignment/removal for the selected project;
- component and version list/create/edit/delete, version release/archive state changes, ticket component/version assignment, project ticket filtering by component/version, and richer release reports with live-versus-released snapshot scope labels, release health summaries, release timeline summaries, scope-change summaries, estimate coverage, status breakdowns, priority breakdowns, issue-type breakdowns, component breakdowns, assignee workload summaries, point totals, and unassigned-component highlighting;
- roadmap scheduled epic timeline, aggregate capacity insights, read-only monthly capacity summary buckets, drag/drop timeline rescheduling, quick scheduling for unscheduled epics, inline schedule edits, child-ticket progress, dependency overview summaries, and dependency list/editing controls, plus ticket form fields for epics, parent epics, and roadmap dates;
- custom field list/create/update/delete with metadata summaries and a layout overview with type breakdowns for the selected project, plus field-aware custom-field controls on ticket create and ticket cards;
- workflow status replacement, board create/select/edit/delete with board definition metadata summaries and a selected-board column settings overview, selected-board status coverage overview, board ticket summary metrics, saved-view board filters, advisory board WIP limits with per-column count warnings, board-backed ticket columns, backlog summary metrics, backlog estimate coverage, backlog status breakdowns, backlog priority breakdowns, backlog assignee breakdowns, backlog sprint workload summaries, backlog sprint assignment and drag/drop reorder, and board card drag/drop status changes for the selected project;
- Profile/API Tokens page for viewing user metadata, token metadata, creating tokens, and revoking your own tokens;
- RBAC page for users, groups, roles, role-binding summaries, and effective-permission inspection when permitted;
- Settings page for global settings, OpenRouter provider management, Shoutrrr notification destination management and test-send, notification policy CRUD with scoped destination-name selection, notification delivery history/manual retry with loaded-delivery health and timing summaries, notification hook CRUD plus policy-aware preview/routing inspection, and security audit-log inspection when permitted, plus personal notification preferences for every signed-in user;
- text/CEL search with field-aware custom-field filter building, result pagination plus saved-view list pagination, saved-view list overview with scope breakdowns, saved-view metadata summaries, create, edit, apply, and delete, including query, sort, columns, display mode, grouping, and pin state, plus pinned project saved views in project navigation and saved-view filters in project boards;
- engine workbench tests for Lua, OpenRouter AI, and WASM automation engines;
- cron job list, create, inline edit, delete, enable/disable, manual run, run-output inspection, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, and trigger-type breakdowns for the selected project;
- project webhook list, create, inline edit, delete, enable/disable, incoming token rotation, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, run history, and outgoing delivery inspection for the selected project;
- project ticket-hook list, create, inline edit, delete, enable/disable, preview, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, and run-history controls for the selected project;
- custom ticket create-page list, create, inline edit, enable/disable, delete, schema preview, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, preview run-history, and intake-form links for the selected project;
- rendered custom ticket create-page intake forms under `/projects/{project_id}/create/{slug}` with safe section/help/group layout widgets and nested fields;
- compact selected-sprint reports with totals, status counts, sprint health summaries, story-point totals when estimates exist, velocity, remaining and burnup summaries, scope-change summaries, assignee workload summaries, ticket links, and live-versus-completed snapshot scope labels.

API token secrets are shown once when created and are not listed later.

Implemented API-only user workflows:

- incoming webhook execution APIs and advanced outgoing delivery retry workflows;
- custom ticket create-page schema resolution and submission APIs.

See [API Guide](api.md) for endpoint details.

## Planned Jira-Like Workflows

Rich backlog planning beyond summary metrics, estimate coverage, status breakdowns, priority breakdowns, assignee breakdowns, sprint workload summaries, sprint assignment, reorder controls, and drag/drop, richer board settings beyond inline edits, definition metadata summaries, selected-board column settings overview, selected-board status coverage overview, summary metrics, saved-view filters, and advisory WIP warnings, richer sprint reporting beyond compact selected-sprint summaries, health summaries, point/ticket-count analytics, scope-change summaries, and assignee workload summaries, roadmap capacity planning beyond aggregate insights, read-only monthly summary buckets, and dependency overview summaries, richer custom-field layout integration beyond metadata summaries, layout overview type breakdowns, ticket controls, and search filter building, advanced release planning beyond version drilldowns, release health summaries, release timeline summaries, scope-change summaries, priority breakdowns, issue-type breakdowns, and assignee workload summaries, richer notification delivery analytics beyond loaded Settings health/timing summaries and policy-aware routing previews, richer automation run-history analytics beyond loaded status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, and raw cron, webhook, ticket-hook, and create-page rows, and richer saved-view UI beyond the current paginated list, list overview scope breakdowns, metadata summaries, pinned project navigation, and board filters are **Planned**. Lua-backed and OpenRouter AI-backed dynamic custom create pages must return structured form definitions and options, not raw HTML. Remaining OpenRouter AI surfaces and persisted WebAssembly automation are also **Planned**.

## Notifications

The browser inbox exposes notification listing, unread filtering, mention notification labels, individual read/unread toggles, refresh, and mark-all-read. Authenticated users can update personal notification preferences in `/settings`. Notification managers can create, edit, enable/disable, rotate, test, and delete global or selected-project Shoutrrr destinations in `/settings`; destination URLs are write-only and are not shown after save. Project notification managers can set project notification defaults, inspect delivery history, and manually retry failed external deliveries in `/settings` or through the API. `read_at` is `null` while a notification is unread. Notifications for comments, mentions, and ticket updates are generated from durable backend events so pending notifications can be processed after restart. Watched tickets include the watcher in comment, status, sprint, and release-change in-app notifications while excluding the actor and deduplicating reporter or assignee overlaps.

Richer hook routing controls are **Planned**.

## Search

Current search supports full-text search over ticket title, description, comments, and attachment metadata such as filename and content type with SQLite FTS5, plus CEL-backed ticket filters. Roadmap date fields `start_date` and `due_date`, and estimate field `story_points`, are available for search filters, sort specs, and saved-view columns. Filters support boolean operators, comparisons, label membership, selected string helpers, `currentUser()`, `today()`, `now()`, and `custom.<field_key>` access for typed custom fields. When a project is selected, `/search` includes a custom-field builder that appends safe `custom.<field_key>` CEL clauses to the raw filter before saving or running a search. See:

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- SQLite FTS5: https://www.sqlite.org/fts5.html
