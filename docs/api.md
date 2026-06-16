# API Guide

Backend API routes live under `/api`. Requests and responses are JSON unless uploading or downloading attachments. JSON uses `snake_case`, UTC RFC3339 timestamps, and opaque string IDs.

## Authentication

Protected routes accept either:

- browser session cookie plus CSRF header for mutating methods; or
- `Authorization: Bearer <api_token>`.

Unauthenticated API requests return `401`.

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

Common codes are `unauthenticated`, `forbidden`, `not_found`, `validation_failed`, `conflict`, and `internal_error`.

## Pagination

Project, ticket, user/admin, saved-view, and notification list endpoints currently use optional `limit` and `offset` query parameters where implemented. `POST /api/search` uses `limit` and an opaque `cursor` in the JSON body and returns `next_cursor`.

## Health

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| `GET` | `/api/health` | No | Returns backend health. |

## Auth, Users, Groups, Roles

| Method | Path | Body |
| --- | --- | --- |
| `POST` | `/api/login` | `{"username":"admin","password":"..."}` |
| `POST` | `/api/logout` | none |
| `GET` | `/api/me` | none |
| `GET` | `/api/tokens` | none |
| `POST` | `/api/tokens` | `{"name":"local-script"}` |
| `DELETE` | `/api/tokens/{token_id}` | none |
| `GET` | `/api/users` | none |
| `POST` | `/api/users` | `{"username":"alice","display_name":"Alice","password":"","disabled":false}` |
| `GET` | `/api/users/{user_id}` | none |
| `PATCH` | `/api/users/{user_id}` | `{"disabled":true}` |
| `DELETE` | `/api/users/{user_id}` | none |
| `GET` | `/api/groups` | none |
| `POST` | `/api/groups` | `{"name":"engineering","display_name":"Engineering"}` |
| `GET` | `/api/groups/{group_id}/members` | none |
| `POST` | `/api/groups/{group_id}/members/{user_id}` | none |
| `DELETE` | `/api/groups/{group_id}/members/{user_id}` | none |
| `GET` | `/api/roles` | none |
| `GET` | `/api/role-bindings` | none |
| `POST` | `/api/role-bindings` | `{"role_name":"project_member","subject_type":"group","subject_id":"group_...","scope":"project","project_id":"project_..."}` |
| `DELETE` | `/api/role-bindings/{binding_id}` | none |

Creating a user with an empty password generates a random password and returns it once in the response.

## Projects and Tickets

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects` | Optional `include_archived=true`, `limit`, `offset`. |
| `POST` | `/api/projects` | `{"key":"CORE","name":"Core","description":"Main project","lead_user_id":""}` |
| `GET` | `/api/projects/{project_id}` | none |
| `GET` | `/api/projects/{project_id}/tickets` | Optional `status`, `assignee_id`, `limit`, `offset`. |
| `POST` | `/api/projects/{project_id}/tickets` | `{"title":"Fix login","description":"...","status":"todo","priority":"High","type":"Bug","assignee_id":""}` |
| `GET` | `/api/tickets/{ticket_id}` | none |
| `PATCH` | `/api/tickets/{ticket_id}` | Any subset of `title`, `description`, `status`, `priority`, `type`, `assignee_id`, `parent_ticket_id`, `rank`. |
| `GET` | `/api/tickets/{ticket_id}/activity` | none |

Statuses are stored as strings. The current frontend uses `todo`, `in_progress`, and `done`.

## Backlog

The first backlog API slice is backend/API-only. It lists project backlog tickets in stable backlog order and supports reordering those tickets by writing stable rank/order values. Browser drag/drop, board UI, richer backlog planning, and reports are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/backlog` | none |
| `PATCH` | `/api/projects/{project_id}/backlog` | `{"ticket_ids":["ticket_2","ticket_1"]}` |

Backlog responses use the same persisted ticket shape as project ticket lists, ordered by backlog rank and then deterministic tie-breakers. Reorder requests submit ticket IDs in desired order and only affect tickets in the addressed project. The backend validates that every submitted ticket belongs to the project, writes rank values atomically, and returns the reordered backlog slice.

## Sprints

The first sprint API slice is backend/API-only. It supports sprint CRUD within a project, starting and completing sprints, and assigning or removing a ticket from a sprint. Browser backlog planning, board drag/drop, and sprint report screens are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/sprints` | Optional `state`. |
| `POST` | `/api/projects/{project_id}/sprints` | `{"name":"Sprint 1","goal":"Ship auth fixes","start_date":"2026-06-16","end_date":"2026-06-30"}` |
| `GET` | `/api/sprints/{sprint_id}` | none |
| `PATCH` | `/api/sprints/{sprint_id}` | Any subset of `name`, `goal`, `start_date`, `end_date`. |
| `DELETE` | `/api/sprints/{sprint_id}` | none |
| `POST` | `/api/sprints/{sprint_id}/start` | Starts a planned sprint. |
| `POST` | `/api/sprints/{sprint_id}/complete` | Completes an active sprint. |
| `PUT` | `/api/tickets/{ticket_id}/sprint` | `{"sprint_id":"sprint_..."}` |
| `DELETE` | `/api/tickets/{ticket_id}/sprint` | Removes the ticket from its sprint. |

Sprint states are `planned`, `active`, and `completed`. Start and complete actions validate state transitions. Ticket assignment keeps the ticket in its existing project; cross-project sprint assignment is invalid.

Burndown, velocity, burnup, sprint report APIs, and other agile analytics are **Planned**.

## Comments and Attachments

| Method | Path | Body |
| --- | --- | --- |
| `GET` | `/api/tickets/{ticket_id}/comments` | none |
| `POST` | `/api/tickets/{ticket_id}/comments` | `{"body":"Looks reproducible."}` |
| `DELETE` | `/api/comments/{comment_id}` | none |
| `GET` | `/api/tickets/{ticket_id}/attachments` | none |
| `POST` | `/api/tickets/{ticket_id}/attachments` | multipart form field `file`. |
| `GET` | `/api/attachments/{attachment_id}/download` | binary download. |
| `DELETE` | `/api/attachments/{attachment_id}` | none |

Attachment bytes are stored in SQLite. The current maximum upload size is 10 MiB. Downloads set `Content-Type` from stored metadata and `Content-Disposition: attachment`.

## Search and Saved Views

Search endpoint:

```http
POST /api/search
Content-Type: application/json
```

```json
{
  "project_id": "project_...",
  "filter": "status == \"todo\" && assignee_id == currentUser()",
  "text": "login error",
  "sort": [{"field": "updated_at", "direction": "desc"}],
  "limit": 50,
  "cursor": ""
}
```

Search returns:

```json
{
  "items": [],
  "next_cursor": ""
}
```

Current filter support is a constrained expression parser, not full CEL yet. Supported fields are `project`, `project_id`, `key`, `title`, `status`, `priority`, `type`, `reporter_id`, `assignee_id`, and `parent_ticket_id`. Supported operators are `==`, `!=`, and `&&`. Values are string literals or `currentUser()`. Parentheses are only parsed as part of function calls.

Full CEL support is **Planned**. See CEL at https://cel.dev/ and cel-go at https://github.com/google/cel-go.

Full-text search uses SQLite FTS5 virtual tables for ticket title/description and comments. See https://www.sqlite.org/fts5.html. Current text input is tokenized into quoted FTS terms; it is not raw FTS syntax.

Saved views:

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/saved-views` | Optional `project_id`, `limit`, `offset`. |
| `POST` | `/api/saved-views` | `{"scope_type":"user","project_id":"","name":"My bugs","query":{"filter":"assignee_id == currentUser()","text":"bug"},"sort":[{"field":"updated_at","direction":"desc"}],"columns":["key","title","status"]}` |
| `GET` | `/api/saved-views/{view_id}` | none |
| `PATCH` | `/api/saved-views/{view_id}` | Any subset of `name`, `query`, `sort`, `columns`. |
| `DELETE` | `/api/saved-views/{view_id}` | none |

Saved-view scopes are `user`, `project`, and `global`. Managing project/global views requires the matching `views:manage` permission.

## Notifications

The first notification API slice is in-app only. It lists notifications for the authenticated user and supports read/unread state. It does not send external messages.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/notifications` | Optional `unread=true`, `limit`, `offset`. |
| `POST` | `/api/notifications/{notification_id}/read` | Marks one of the current user's notifications read. |
| `POST` | `/api/notifications/{notification_id}/unread` | Marks one of the current user's notifications unread. |
| `POST` | `/api/notifications/read-all` | Marks all current user's unread notifications read. |

Notification responses use the persisted notification shape:

```json
{
  "items": [
    {
      "id": "notification_...",
      "user_id": "user_...",
      "type": "ticket_assigned",
      "subject_type": "ticket",
      "subject_id": "ticket_...",
      "body": "You were assigned CORE-12",
      "data": {},
      "read_at": null,
      "created_at": "2026-06-16T10:30:00Z"
    }
  ]
}
```

Read state is owned by the notification row. `read_at: null` means unread; marking read sets `read_at`, and marking unread clears it. Users may only list or mutate their own notifications.

Shoutrrr destinations, notification preferences, global/project/dashboard notification policies, notification hooks, external delivery queues, delivery history/retry, webhooks, and AI/Lua notification hooks are **Planned**.

## Cron Jobs

The first cron automation API/scheduler slice exposes Lua cron job management, manual execution, and run history. Cron jobs execute as their owner user and use the owner's current effective RBAC permissions at run time.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/cron-jobs` | Optional list filters where implemented. |
| `POST` | `/api/cron-jobs` | Cron job definition. |
| `GET` | `/api/cron-jobs/{cron_job_id}` | none |
| `PATCH` | `/api/cron-jobs/{cron_job_id}` | Cron job definition updates. |
| `DELETE` | `/api/cron-jobs/{cron_job_id}` | none |
| `POST` | `/api/cron-jobs/{cron_job_id}/run` | Starts a manual run. |
| `GET` | `/api/cron-jobs/{cron_job_id}/runs` | Run history for the job. |

Cron job CRUD and manual runs require automation management permissions. Run history uses the shared automation run-history model and must not expose secrets. The implemented cron slice is Lua-only.

Implemented cron Lua helpers are `rayboard.log`, `rayboard.search`, `rayboard.get_ticket`, `rayboard.create_ticket`, `rayboard.update_ticket`, and `rayboard.comment`. Helpers execute through normal backend service/RBAC paths as the cron job owner. OpenRouter AI automation, ticket hooks, custom create pages, webhooks, and notification hooks are **Planned**.
