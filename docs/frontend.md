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
- issue pages under `/issues/{ticket_id}` for one ticket with metadata, labels, custom fields, planning controls, comments, attachments, and activity history;
- a profile page under `/profile` for current user metadata and self-service API token management;
- an RBAC page under `/rbac` for creating users and groups, managing group members, enabling/disabling/deleting users, creating/deleting role bindings, and inspecting effective permissions by user/scope when the signed-in user has permission;
- a settings page under `/settings` for global attachment/webhook/demo settings, OpenRouter provider management, Shoutrrr notification destination management and test-send, notification policy CRUD, project notification defaults, notification delivery history/manual retry, notification hook CRUD/preview/run inspection, security audit-log inspection when permitted, and personal notification preferences for every signed-in user;
- login/logout using backend API sessions;
- CSRF header handling from the `rayboard_csrf` cookie;
- an unread notification badge in the persistent Dashboard nav item;
- project listing and project creation;
- ticket listing per selected project;
- ticket creation;
- ticket status changes between `todo`, `in_progress`, and `done`;
- ticket label entry on create plus ticket-card label display, comma-separated label updates, and project label filtering with ticket counts;
- ticket comment listing, creation, and deletion on ticket cards;
- ticket attachment listing, upload, download, and delete controls on ticket cards;
- a notification inbox with unread filtering, read/unread toggles, refresh, and mark-all-read;
- a sprint panel for listing, creating, starting, completing, and deleting project sprints, plus ticket-card sprint assignment/removal;
- a release-planning panel for listing, creating, updating, and deleting components and versions, changing version state, and assigning tickets to components/versions;
- a roadmap panel that lists project epics, schedule dates, and child-ticket progress, plus ticket-form fields for epics, parent epics, and roadmap dates;
- a custom-fields panel for listing, creating, updating, and deleting project fields, plus ticket create/card JSON entry for typed custom-field values;
- an Account/API Tokens profile page where signed-in users can view token metadata, create API tokens with a one-time secret display, and revoke their own tokens;
- compact search with text/CEL filters plus saved-view list, create, edit, apply, and delete controls for query, sort, columns, display mode, grouping, and pinning;
- an engine workbench for testing Lua, OpenRouter AI, and WASM engines through `/api/engines/test`, with state/mode badges, explicit action-preview display, and raw JSON output;
- basic cron job management on `/automation`, including project-filtered list, create, delete, enable/disable, manual run, and recent run output through `/api/cron-jobs`;
- basic project webhook management on `/automation`, including list, create, delete, enable/disable, incoming token rotation, run history, and outgoing delivery inspection through `/api/projects/{project_id}/webhooks`;
- basic project ticket-hook management on `/automation`, including list, create, delete, enable/disable, and preview through `/api/projects/{project_id}/ticket-hooks` and `/api/ticket-hooks/{hook_id}/preview`.
- basic project workflow and board management on `/projects/{project_id}`, including status replacement, board creation/selection/deletion, and board-backed ticket columns through `/api/projects/{project_id}/statuses`, `/api/projects/{project_id}/boards`, and `/api/boards/{board_id}/tickets`.
- rendered custom ticket create-page intake forms under `/projects/{project_id}/create/{slug}`, using the resolved structured schema and submitting through `/api/projects/{project_id}/ticket-create-pages/{slug}/submit`.

Token secrets are shown only when created and are not listed later.

It does not currently expose all backend endpoints. Advanced search pagination and component/version filtering are API-only for now.

Drag/drop UI, richer board settings/editing beyond basic create/select/delete, richer backlog planning beyond basic up/down ordering, sprint report screens, release reports, richer roadmap timeline controls, component/version filtering and reporting, custom-field search/layout screens, richer create-page layout widgets, and advanced release planning are **Planned**.

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
- SortableJS for board/backlog drag-and-drop: https://sortablejs.github.io/Sortable/
- CodeMirror for Lua/CEL editors: https://codemirror.net/

These libraries are not currently vendored or used. When added, they should be embedded static assets so Rayboard remains a simple Go binary without a Node build pipeline.

## CSS Extension Points

Custom CSS editing is **Planned** and intentionally not implemented yet. Current CSS is the only active stylesheet system. Future project/board CSS customization should use stable wrapper classes, predictable data attributes, and CSS variables where practical.
