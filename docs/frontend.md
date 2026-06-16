# Frontend Architecture

The current frontend is embedded in the Go binary with `embed.FS` and served by `internal/frontend`.

Implemented routes:

- `GET /`: renders `templates/index.html`.
- `GET /health`: returns frontend health JSON.
- `GET /static/*`: serves embedded static assets.
- `/api/*` for `GET`, `POST`, `PATCH`, and `DELETE`: reverse-proxies to `--backend-url`.

## Current UI

The current UI is a small vanilla JavaScript board shell. It supports:

- login/logout using backend API sessions;
- CSRF header handling from the `rayboard_csrf` cookie;
- project listing and project creation;
- ticket listing per selected project;
- ticket creation;
- ticket status changes between `todo`, `in_progress`, and `done`.

It does not currently expose all backend endpoints. User/group/RBAC administration, comments, attachments, saved views, advanced search, backlog list/reorder endpoints, component CRUD, version/release CRUD, and ticket component/version assignment are API-only for now.

Sprint CRUD, start/complete actions, and ticket sprint assignment/removal are also API-only for now. Drag/drop UI, board UI beyond the current simple status shell, richer backlog planning, sprint/report screens, release reports, roadmap timeline screens, component/version UI screens, and advanced release planning are **Planned**.

## Asset Policy

Current assets are plain HTML, CSS, and vanilla JavaScript. There is no Node/npm build step.

Planned frontend dependencies:

- HTMX for server-rendered partial workflows: https://htmx.org/
- SortableJS for board/backlog drag-and-drop: https://sortablejs.github.io/Sortable/
- CodeMirror for Lua/CEL editors: https://codemirror.net/

These libraries are not currently vendored or used. When added, they should be embedded static assets so Rayboard remains a simple Go binary without a Node build pipeline.

## CSS Extension Points

Custom CSS editing is **Planned** and intentionally not implemented yet. Current CSS is the only active stylesheet system. Future project/board CSS customization should use stable wrapper classes, predictable data attributes, and CSS variables where practical.
