# Frontend Architecture

The current frontend is embedded in the Go binary with `embed.FS` and served by `internal/frontend`.

Implemented routes:

- `GET /`: renders the signed-in dashboard shell from `templates/index.html`.
- `GET /projects` and `GET /projects/{project_id}`: render the embedded project page shell.
- `GET /issues/{ticket_id}`: renders the embedded issue detail page shell.
- `GET /profile`: renders the profile/API token page shell.
- `GET /rbac` and `GET /admin/rbac`: render the RBAC administration page shell.
- `GET /settings`: renders the settings page shell.
- `GET /search`: renders the search/saved-views page shell.
- `GET /automation`: renders the engine workbench and basic project ticket-hook management shell.
- `GET /1` through `GET /5`: render the same embedded application shell with distinct design variants selected.
- `GET /docs` and `GET /docs/{page}`: render embedded markdown documentation as HTML.
- `GET /health`: returns frontend health JSON.
- `GET /static/*`: serves embedded static assets.
- `/api/*` for `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`: reverse-proxies to `--backend-url`.

## Current UI

The current UI is a small vanilla JavaScript website shell. It supports:

- a production dashboard at `/`, with five embedded design previews still available under `/1`, `/2`, `/3`, `/4`, and `/5`;
- a persistent app navigation for Dashboard, Projects, Search, Automation, RBAC, Profile, Docs, Swagger, and Redoc;
- a dashboard overview at `/` with project/ticket summary metrics, recently modified tickets, biggest projects, active sprints, and notifications;
- project pages under `/projects` and `/projects/{project_id}` for project-scoped tickets, sprints, components, versions, roadmap epics, and custom fields;
- issue pages under `/issues/{ticket_id}` for one ticket with metadata, watch/unwatch controls, labels, custom fields, planning controls, linked issues, comments, attachments, and activity history;
- a profile page under `/profile` for current user metadata and self-service API token management;
- an RBAC page under `/rbac` for creating users and groups, managing group members, enabling/disabling/deleting users, creating/deleting role bindings, and inspecting effective permissions by user/scope when the signed-in user has permission;
- a settings page under `/settings` for global attachment/webhook/demo settings, OpenRouter provider management, Shoutrrr notification destination management and test-send, notification policy CRUD with scoped destination-name selection, project notification defaults, notification delivery history/manual retry with loaded-delivery health, timing, event, destination, and retry-pressure summaries, notification hook CRUD plus policy-aware preview/routing inspection with compact routing summaries, security audit-log inspection when permitted, and personal notification preferences for every signed-in user;
- login/logout using backend API sessions;
- CSRF header handling from the `rayboard_csrf` cookie;
- an unread notification badge in the persistent Dashboard nav item;
- project listing and project creation;
- ticket listing per selected project, including label, component, and version filters backed by the project ticket API;
- ticket creation and soft delete from ticket cards or issue detail pages;
- ticket status changes between `todo`, `in_progress`, and `done`;
- ticket label entry on create plus ticket-card label display, comma-separated label updates, project label filtering with ticket counts, and project label administration with description/color metadata;
- ticket watch/unwatch controls and watcher lists on ticket cards and issue detail pages;
- ticket link dependency summaries, listing, creation, and removal on ticket cards and issue detail pages;
- ticket comment listing, creation, deletion, and `@username` mention notification support on ticket cards;
- ticket attachment listing, upload, download, and delete controls on ticket cards;
- a notification inbox with unread filtering, mention notification labels, read/unread toggles, refresh, and mark-all-read;
- a sprint panel for listing, state-filtering, creating, editing, starting, completing, and deleting project sprints, plus ticket-card and backlog sprint assignment/removal;
- a backlog panel with planning summary metrics, estimate coverage, status breakdowns, priority breakdowns, issue-type breakdowns, label breakdowns, component breakdowns, version breakdowns, parent epic breakdowns, assignee workload breakdowns, reporter breakdowns, sprint workload progress summaries, start-date breakdowns, due-date breakdowns, readiness summaries, risk summaries, attention summaries, ticket-age breakdowns, update-freshness breakdowns, current sprint labels, sprint assignment/removal controls, up/down controls, and native drag/drop ticket reordering through `/api/projects/{project_id}/backlog`;
- a release-planning panel for listing, creating, updating, and deleting components and versions, showing component ownership coverage, component visible-ticket status summaries, version visible-ticket status, priority, issue-type, label, and assignee summaries, grouping versions by lifecycle state, date coverage, target-date health, and release timing variance, assigning tickets to components/versions, and viewing compact live reports for the selected version;
- a roadmap panel with a scheduled epic timeline, aggregate capacity insights, read-only monthly capacity summary buckets, selected-month capacity drilldowns, persisted monthly point target indicators, drag/drop timeline rescheduling, quick scheduling for unscheduled epics, inline schedule edits, child-ticket progress, dependency overview summaries, read-only dependency graph visualization, and dependency list/editing controls, plus ticket-form fields for epics, parent epics, and roadmap dates through `/api/projects/{project_id}/roadmap`, `/api/projects/{project_id}/roadmap/capacity-targets`, and `/api/projects/{project_id}/roadmap/schedule`;
- a custom-fields panel for listing, creating, updating, and deleting project fields with metadata summaries, a layout overview with type breakdowns plus requirement/configuration insight counts, usage coverage summaries, select option usage summaries, and unmodeled value summaries across current project tickets, plus field-aware ticket create/card controls for typed custom-field values;
- an Account/API Tokens profile page where signed-in users can view token metadata, create API tokens with a one-time secret display, and revoke their own tokens;
- compact search with text/CEL filters, field-aware custom-field filter building, cursor-paginated results, applied saved-view column and grouping presentation, saved-view list pagination, list overview with scope breakdowns, display-mode summaries, configuration insight counts, field usage summaries, metadata summaries, create, edit, apply, and delete controls for query, sort, columns, display mode, grouping, and pinning, plus pinned project saved views in project navigation and saved-view filters in project boards;
- an engine workbench for testing Lua, OpenRouter AI, and WASM engines through `/api/engines/test`, with state/mode badges, explicit action-preview display, and raw JSON output;
- cron job management on `/automation`, including project-filtered list, create, inline edit, delete, enable/disable, manual run, recent run output, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, and failure-cause breakdowns through `/api/cron-jobs`;
- project webhook management on `/automation`, including list, create, inline edit, delete, enable/disable, incoming token rotation, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, failure-cause breakdowns, run history, and outgoing delivery inspection through `/api/projects/{project_id}/webhooks`, `/api/webhook-definitions/{webhook_id}`, and `/api/webhook-definitions/{webhook_id}/deliveries`;
- project ticket-hook management on `/automation`, including list, create, inline edit, delete, enable/disable, preview, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, failure-cause breakdowns, and run-history inspection through `/api/projects/{project_id}/ticket-hooks`, `/api/ticket-hooks/{hook_id}`, `/api/ticket-hooks/{hook_id}/preview`, and `/api/ticket-hooks/{hook_id}/runs`.
- custom ticket create-page management on `/automation`, including list, create, inline edit, enable/disable, delete, structured flat field-layout builder with raw JSON fallback, schema preview, loaded run-status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, failure-cause breakdowns, preview run-history inspection, and intake-form links through `/api/projects/{project_id}/ticket-create-pages`, `/api/ticket-create-pages/{page_id}`, and `/api/ticket-create-pages/{page_id}/runs`;
- project workflow and board management on `/projects/{project_id}`, including status replacement, board creation/selection/edit/deletion with board definition metadata summaries, a selected-board column settings overview, selected-board status coverage overview, board-backed ticket columns, board ticket summary metrics, selected-board flow-balance summaries, selected-board priority breakdowns, selected-board due-date breakdowns, selected-board issue-type breakdowns, selected-board label breakdowns, selected-board component breakdowns, selected-board version breakdowns, selected-board parent epic breakdowns, selected-board sprint breakdowns, selected-board assignee workload summaries, selected-board reporter breakdowns, selected-board estimate coverage, selected-board capacity utilization overviews, selected-board attention summaries, selected-board risk overviews for blocked/overdue/stale/high-priority work, saved-view board filtering, per-column WIP limit warnings, and drag/drop card moves through `/api/projects/{project_id}/statuses`, `/api/projects/{project_id}/boards`, `/api/boards/{board_id}`, and `/api/boards/{board_id}/tickets`.
- rendered custom ticket create-page intake forms under `/projects/{project_id}/create/{slug}`, using the resolved structured schema, safe section/help/group layout widgets with nested fields, and submissions through `/api/projects/{project_id}/ticket-create-pages/{slug}/submit`.
- compact sprint reports with live current-assignment and completed snapshot scope markers, sprint health summaries from sprint dates/state, story-point velocity when estimates exist, ticket-count fallback analytics, remaining count, burnup summary, scope-change summaries, status breakdowns, start-date breakdowns, due-date breakdowns, ticket-age breakdowns, update freshness breakdowns, readiness summaries, risk summaries, attention summaries, priority breakdowns, issue-type breakdowns, label breakdowns, estimate coverage, component breakdowns, version breakdowns, parent epic breakdowns, reporter breakdowns, assignee workload summaries, and ticket links for the selected project.
- richer version release reports with live current-assignment and released snapshot scope markers, release health summaries from target/release dates, release timeline summaries, release analytics, scope-change summaries, progress percentage, estimate coverage, story-point totals when estimates exist, total/done/open counts, status breakdowns, start-date breakdowns, due-date breakdowns, ticket-age breakdowns, update freshness breakdowns, readiness summaries, risk summaries, attention summaries, priority breakdowns, issue-type breakdowns, label breakdowns, component breakdowns, parent epic breakdowns, sprint breakdowns, assignee workload summaries, reporter breakdowns, unassigned-component highlighting, and ticket links for the selected version.

Token secrets are shown only when created and are not listed later.

It does not currently expose all backend endpoints.

Richer board settings beyond inline edits, definition metadata summaries, the selected-board column settings overview, selected-board status coverage overview, summary metrics, flow-balance summaries, selected-board priority, due-date, issue-type, label, component, version, parent epic, and sprint breakdowns, assignee workload summaries, reporter breakdowns, estimate coverage, capacity utilization overviews, selected-board attention summaries, selected-board risk overviews, saved-view filters, and WIP warnings, richer backlog planning beyond summary metrics, estimate coverage, status breakdowns, priority breakdowns, issue-type breakdowns, label breakdowns, component breakdowns, version breakdowns, parent epic breakdowns, assignee workload breakdowns, reporter breakdowns, sprint workload progress summaries, start-date breakdowns, due-date breakdowns, readiness summaries, risk summaries, attention summaries, ticket-age breakdowns, update-freshness breakdowns, sprint assignment, reorder controls, and drag/drop, richer sprint reporting beyond compact selected-sprint summaries, health summaries, point/ticket-count analytics, scope-change summaries, ticket-age breakdowns, update freshness breakdowns, readiness summaries, risk summaries, attention summaries, priority breakdowns, issue-type breakdowns, label breakdowns, estimate coverage, component breakdowns, version breakdowns, and assignee workload summaries, richer roadmap capacity planning beyond aggregate insights, read-only monthly summary buckets, selected-month drilldowns, and persisted target indicators, richer custom-field layout screens beyond metadata summaries, layout overview type breakdowns, requirement/configuration insight counts, usage coverage summaries, select option usage summaries, unmodeled value summaries, ticket controls, search filter building, and flat create-page layout building, richer saved-view UI beyond metadata summaries, list overview scope breakdowns, display-mode summaries, configuration insight counts, field usage summaries, applied list columns/grouping, pinned project navigation, and board filters, richer notification delivery analytics beyond loaded Settings health/timing/event/destination/retry summaries and policy-aware routing previews, richer automation run-history analytics beyond loaded status summaries, completion/failure rates, state filters, duration metrics, oldest/newest run timestamps, trigger-type breakdowns, and raw cron, webhook, ticket-hook, and create-page rows, and advanced release planning beyond version drilldowns, release health summaries, release timeline summaries, release analytics, scope-change summaries, start-date breakdowns, due-date breakdowns, ticket-age breakdowns, update freshness breakdowns, readiness summaries, risk summaries, attention summaries, priority breakdowns, issue-type breakdowns, label breakdowns, parent epic breakdowns, sprint breakdowns, reporter breakdowns, and assignee workload summaries are **Planned**.

## Design Variants

The selector under `/` exposes five embedded visual directions without changing backend API behavior or the current JavaScript contract:

- `/1`: Operations, a dense admin shell for repeated project and ticket work.
- `/2`: Planning, a backlog-forward treatment with stronger prioritization cues.
- `/3`: Automation, a darker workspace for hooks, scripts, and run feedback.
- `/4`: Triage, a compact high-contrast queue style.
- `/5`: Executive, a quieter overview style for roadmap and delivery status review.

The variants are currently CSS-scoped explorations on the same HTML template. They are not separate frontend applications.

## Asset Policy

Current assets are plain HTML, CSS, vanilla JavaScript, and the markdown files under `/docs`, all embedded into the Go binary. There is no Node/npm build step.

Planned frontend dependencies:

- HTMX for server-rendered partial workflows: https://htmx.org/
- CodeMirror for Lua/CEL editors: https://codemirror.net/

These libraries are not currently vendored or used. When added, they should be embedded static assets so Rayboard remains a simple Go binary without a Node build pipeline.

## CSS Extension Points

Custom CSS editing is **Planned** and intentionally not implemented yet. Current CSS is the only active stylesheet system. Future project/board CSS customization should use stable wrapper classes, predictable data attributes, and CSS variables where practical.
