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
GET    /api/tickets/{ticket_id}
PATCH  /api/tickets/{ticket_id}
```

Use `POST` for actions that are not simple CRUD:

```text
POST /api/sprints/{id}/start
POST /api/sprints/{id}/complete
POST /api/cron-jobs/{id}/run
POST /api/webhooks/{id}/test
```

## JSON

- Requests and responses are JSON unless uploading/downloading attachments.
- Use snake_case JSON names.
- Use UTC RFC3339 timestamps.
- IDs are opaque strings in JSON, even if SQLite stores integers internally.

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
  "next_cursor": "opaque"
}
```

For the first implementation, offset pagination is acceptable only for small admin lists.

## Search

`POST /api/search` accepts:

```json
{
  "filter": "project == \"CORE\" && status != \"Done\"",
  "text": "login error",
  "sort": [{"field": "updated_at", "direction": "desc"}],
  "limit": 50,
  "cursor": ""
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
