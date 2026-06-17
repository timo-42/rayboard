# Frontend Architecture

The current frontend is embedded in the Go binary with `embed.FS` and served by `internal/frontend`.

Implemented routes:

- `GET /`: renders `templates/index.html`.
- `GET /1` through `GET /5`: render the same embedded application shell with distinct design variants selected.
- `GET /docs` and `GET /docs/{page}`: render embedded markdown documentation as HTML.
- `GET /health`: returns frontend health JSON.
- `GET /static/*`: serves embedded static assets.
- `/api/*` for `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`: reverse-proxies to `--backend-url`.

## Current UI

The current UI is a small vanilla JavaScript board shell. It supports:

- a root UI selector linking to five embedded design variants under `/1`, `/2`, `/3`, `/4`, and `/5`;
- login/logout using backend API sessions;
- CSRF header handling from the `rayboard_csrf` cookie;
- project listing and project creation;
- ticket listing per selected project;
- ticket creation;
- ticket status changes between `todo`, `in_progress`, and `done`;
- ticket label entry on create plus ticket-card label display and comma-separated label updates;
- ticket comment listing, creation, and deletion on ticket cards;
- ticket attachment listing, upload, download, and delete controls on ticket cards;
- a notification inbox with unread filtering, read/unread toggles, refresh, and mark-all-read;
- a sprint panel for listing, creating, starting, completing, and deleting project sprints, plus ticket-card sprint assignment/removal;
- a release-planning panel for listing, creating, deleting components and versions, changing version state, and assigning tickets to components/versions;
- a roadmap panel that lists project epics, schedule dates, and child-ticket progress, plus ticket-form fields for epics, parent epics, and roadmap dates;
- an Account/API Tokens panel where signed-in users can view token metadata, create API tokens with a one-time secret display, and revoke their own tokens;
- compact search with text/CEL filters plus saved-view list, create, apply, and delete controls;
- an engine workbench for testing Lua, OpenRouter AI, and WASM engines through `/api/engines/test`.

Token secrets are shown only when created and are not listed later.

It does not currently expose all backend endpoints. User/group/RBAC administration beyond self-service API tokens, advanced saved-view editing, advanced search pagination, backlog list/reorder endpoints, project workflow status APIs, board definition CRUD, board ticket listing, detailed component/version editing, custom field management, ticket custom-field values, and saved automation management screens are API-only for now.

Drag/drop UI, board settings UI, board UI beyond the current simple status shell, richer backlog planning, sprint report screens, release reports, richer roadmap timeline controls, richer component/version screens, custom-field screens, and advanced release planning are **Planned**.

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
