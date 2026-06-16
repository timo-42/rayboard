# API Contract

## Route Style

All backend routes live under `/api` except incoming webhook endpoints, which still use `/api/webhooks/incoming/{id}`.

Use nouns and nested project resources where it clarifies authorization:

```text
GET    /api/me
POST   /api/login
POST   /api/logout
GET    /api/projects
POST   /api/projects
GET    /api/projects/{project_id}
POST   /api/projects/{project_id}/tickets
GET    /api/projects/{project_id}/backlog
PATCH  /api/projects/{project_id}/backlog
GET    /api/projects/{project_id}/sprints
POST   /api/projects/{project_id}/sprints
GET    /api/projects/{project_id}/components
POST   /api/projects/{project_id}/components
GET    /api/projects/{project_id}/versions
POST   /api/projects/{project_id}/versions
GET    /api/projects/{project_id}/custom-fields
POST   /api/projects/{project_id}/custom-fields
GET    /api/tickets/{ticket_id}
PATCH  /api/tickets/{ticket_id}
PUT    /api/tickets/{ticket_id}/sprint
DELETE /api/tickets/{ticket_id}/sprint
GET    /api/components/{component_id}
PATCH  /api/components/{component_id}
DELETE /api/components/{component_id}
GET    /api/versions/{version_id}
PATCH  /api/versions/{version_id}
DELETE /api/versions/{version_id}
GET    /api/custom-fields/{field_id}
PATCH  /api/custom-fields/{field_id}
DELETE /api/custom-fields/{field_id}
```

Use `POST` for actions that are not simple CRUD:

```text
POST /api/sprints/{sprint_id}/start
POST /api/sprints/{sprint_id}/complete
POST /api/cron-jobs/{id}/run
POST /api/webhooks/{id}/test
```

## Backlog Ordering

The first backlog route shape is:

```text
GET   /api/projects/{project_id}/backlog
PATCH /api/projects/{project_id}/backlog
```

`GET` lists tickets for the project in stable backlog order. `PATCH` accepts ticket IDs in the desired order and updates persisted ranks atomically. Reorder handlers must validate project ownership for every ticket, keep ranks stable across repeated reads, and use a deterministic secondary sort when ranks collide.

## Components And Versions

The first component/version route shape is:

```text
GET    /api/projects/{project_id}/components
POST   /api/projects/{project_id}/components
GET    /api/components/{component_id}
PATCH  /api/components/{component_id}
DELETE /api/components/{component_id}
GET    /api/projects/{project_id}/versions
POST   /api/projects/{project_id}/versions
GET    /api/versions/{version_id}
PATCH  /api/versions/{version_id}
DELETE /api/versions/{version_id}
PATCH  /api/tickets/{ticket_id}
```

Components and versions are project-scoped resources. Nested project collection routes make create/list authorization explicit; item routes resolve the owning project before checking permissions. Ticket component/version assignment uses `component_id` and `version_id` fields on ticket create/update and must reject cross-project assignment.

Release reports, roadmap timeline screens, component/version UI screens, and advanced release planning are planned work outside this first backend/API slice.

## Custom Fields

The first custom-field route shape is:

```text
GET    /api/projects/{project_id}/custom-fields
POST   /api/projects/{project_id}/custom-fields
GET    /api/custom-fields/{field_id}
PATCH  /api/custom-fields/{field_id}
DELETE /api/custom-fields/{field_id}
POST   /api/projects/{project_id}/tickets
PATCH  /api/tickets/{ticket_id}
```

Custom fields are project-scoped resources. Definitions support `text`, `number`, `boolean`, `date`, `single_select`, `multi_select`, and `user`. Ticket values are submitted as a `custom_fields` object keyed by custom-field key. Create validates all required fields. Update treats an omitted `custom_fields` object as no change, and a provided `custom_fields` object as a full replacement.

Custom-field CEL filtering, UI field management screens, custom create page integration, and advanced field layouts are planned follow-up work.

## JSON

- Requests and responses are JSON unless uploading/downloading attachments.
- Use snake_case JSON names.
- Use UTC RFC3339 timestamps.
- IDs are opaque strings in JSON, even if SQLite stores integers internally.

## Resource Objects

JSON API resources and resource-like computed views use a Kubernetes-inspired envelope, adapted for Rayboard rather than copied wholesale.

Create, update, and JSON action requests put desired state or command intent under `spec`:

```json
{
  "spec": {
    "title": "Fix login",
    "description": "Session cookie is not persisted",
    "assignee_id": "user_123"
  }
}
```

Resource responses use `metadata`, `spec`, and `status`:

```json
{
  "metadata": {
    "id": "ticket_123",
    "key": "CORE-42",
    "created_at": "2026-06-16T10:00:00Z",
    "updated_at": "2026-06-16T11:00:00Z"
  },
  "spec": {
    "title": "Fix login",
    "description": "Session cookie is not persisted",
    "assignee_id": "user_123",
    "status": "in_progress"
  },
  "status": {
    "state": "open",
    "comment_count": 2,
    "attachment_count": 1
  }
}
```

- `metadata` is identity/bookkeeping: IDs, keys, slugs/names where they identify the object, timestamps, generation/resource-version fields, labels, annotations, archive/delete markers.
- `spec` is user-controlled desired state or command intent and is the only resource object section accepted by create/update/action handlers.
- `status` is observed/computed/server-controlled state and is never accepted in create/update input.
- List endpoints return arrays of resource objects, and each item uses `metadata`, `spec`, and `status` when it represents API state.
- Huma schema DTOs must expose concrete `spec` and `status` object fields so `/api/openapi.json`, Swagger UI, and Redoc show useful request/response bodies.
- Binary attachment downloads and empty `204` responses are the practical exceptions. JSON command/action endpoints should still use `spec` inputs and resource-like `metadata`/`spec`/`status` outputs when they return a body.

## Errors

All API errors use:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Human readable message",
    "fields": {
      "title": "Required"
    }
  }
}
```

Recommended codes:

- `unauthenticated`
- `forbidden`
- `not_found`
- `validation_failed`
- `conflict`
- `rate_limited`
- `automation_failed`
- `external_delivery_failed`
- `internal_error`

## Pagination

Use cursor pagination for lists that may grow:

```json
{
  "items": [],
  "status": {"next_cursor": "opaque"}
}
```

For the first implementation, offset pagination is acceptable only for small admin lists.

## Search

`POST /api/search` accepts:

```json
{
  "spec": {
    "filter": "project == \"CORE\" && status != \"Done\"",
    "text": "login error",
    "sort": [{"field": "updated_at", "direction": "desc"}],
    "limit": 50,
    "cursor": ""
  }
}
```

- `filter` is CEL.
- `text` is SQLite FTS input.
- Backend combines both safely.
- Sorting and pagination are not CEL syntax.

## Authentication

Browser UI:

- session cookie
- CSRF required for mutating requests

Scripts/API:

- `Authorization: Bearer <api_token>`
- no CSRF

Incoming webhooks:

- `Authorization: Bearer <webhook_token>`

## Frontend Requests

Frontend handlers should call the backend HTTP API through an internal client. They should not access stores or services directly.

HTMX partial routes can exist in frontend, but they should still use backend API responses as their data source.
