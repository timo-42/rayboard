# Agent 03: Core Tracking Domain

## Mission

Implement projects, tickets, workflows, boards, backlog, sprints, roadmap epics, components, releases, custom fields, comments, activity, and attachments integration.

Read first:

- `plan/contracts/data-model.md`
- `plan/contracts/authorization.md`
- `plan/contracts/api-conventions.md`

## Deliverables

- project CRUD and settings domain
- tickets CRUD
- comments/activity history
- workflow/statuses
- boards and backlog ordering
- sprints
- epics/roadmap basics
- components and versions/releases
- custom fields
- attachment endpoints

## Package Tasks

1. Projects:
   - key/name/description
   - owner role binding
   - members via RBAC bindings
2. Workflow:
   - project statuses
   - board columns map to statuses
   - status changes emit activity/events
3. Tickets:
   - key generation per project
   - title/description/status/priority/assignee/reporter
   - labels, dates, parent epic
   - comments
   - activity log
4. Boards/backlog:
   - list by project
   - first backend backlog API slice: `GET /api/projects/{project_id}/backlog` and `PATCH /api/projects/{project_id}/backlog`
   - reorder backlog tickets using stable persisted rank/order values
   - stable ordering fields with deterministic tie-breakers
   - board reorder can follow later once board APIs/UI are introduced
5. Sprints:
   - CRUD
   - start/complete
   - goal/date range
   - assign tickets
6. Roadmap:
   - epics with start/due dates
   - child ticket rollup by status
7. Components/releases:
   - first backend API slice: project component CRUD under `/api/projects/{project_id}/components` and `/api/components/{component_id}`
   - first backend API slice: project version/release CRUD under `/api/projects/{project_id}/versions` and `/api/versions/{version_id}`
   - optional owner/default assignee
   - assign tickets through `component_id` and `version_id` fields on ticket create/update
   - release reports, roadmap timeline screens, component/version UI screens, and advanced release planning are planned follow-up work
8. Custom fields:
   - definitions/options per project
   - typed values
   - required validation
9. Attachments:
   - upload/download/delete/list
   - checksum/size/content-type enforcement
   - activity entries

## Integration Points

- Agent 04 indexes tickets/comments for FTS and search.
- Agent 06 ticket hooks run before/after ticket create/update.
- Agent 07 listens to ticket/project events for notifications/webhooks.
- Agent 05 builds UI on top of these APIs.

## Tests

- project create with owner binding.
- ticket create/update/delete/list.
- status transition and activity entry.
- comment create and mention event.
- board/backlog ordering.
- sprint start/complete and ticket assignment.
- component/version create/update/delete and ticket assignment.
- epic rollup.
- custom field validation and typed storage.
- attachment permission/checksum/limits.
- RBAC denial for every mutating operation.

## Acceptance Criteria

- Mutations emit domain events.
- All project-scoped objects resolve authorization through project scope.
- Ticket create/update path is single and reusable by UI, API, demo seed, cron, hooks, and webhooks.
