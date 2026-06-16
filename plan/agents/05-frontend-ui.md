# Agent 05: Embedded Frontend UI

## Mission

Build the server-rendered frontend with embedded templates/static assets, HTMX interactions, and feature pages driven by backend HTTP APIs.

Read first:

- `plan/contracts/architecture.md`
- `plan/contracts/api-conventions.md`
- `plan/contracts/authorization.md`

## Deliverables

- embedded templates/static filesystem
- base layout and navigation
- login/logout/account screens
- project/ticket/board/backlog/roadmap screens
- admin/project/user settings screens
- automation editors
- notification/search/saved view UI

## Frontend Stack

- Go `html/template`
- HTMX
- plain CSS
- small vanilla JS modules
- SortableJS for board/backlog drag-and-drop
- CodeMirror optional for Lua/CEL/AI prompt editors; textarea acceptable for first pass
- all assets vendored and embedded with `embed.FS`

## Package Tasks

1. Asset embedding:
   - templates
   - CSS
   - JS
   - vendored HTMX and SortableJS
2. Base UI:
   - login
   - nav
   - error surfaces
   - CSRF handling
3. Project/ticket UI:
   - project list/detail
   - ticket list/detail/create/edit
   - comments/attachments/activity
4. Board/backlog/sprint/roadmap UI:
   - drag/drop status/order updates
   - sprint assignment
   - epic timeline basics
5. Settings:
   - global admin settings
   - RBAC users/groups/roles/bindings
   - project settings
   - user settings
6. Automation UI:
   - cron
   - hooks
   - custom create pages
   - webhooks
   - Lua/AI selector
   - test panels and run history
7. Notifications/search:
   - notification inbox/badge
   - preferences
   - Shoutrrr destination settings
   - CEL + FTS search
   - saved views

## Integration Points

- Must call backend through configured backend URL.
- Requires API client layer in frontend package.
- Agent 00 provides frontend server shell.
- Feature agents provide APIs.

## Tests

- templates render.
- HTMX partials render.
- static assets embedded.
- frontend API client uses backend URL.
- unauthenticated pages redirect to login.
- forms include CSRF where required.
- board/backlog drag/drop requests hit correct backend routes.

## Acceptance Criteria

- No backend store/service imports from frontend.
- No runtime filesystem asset dependency.
- Text fits in compact admin/tool UI; avoid marketing-style landing pages.
