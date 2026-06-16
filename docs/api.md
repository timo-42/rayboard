# API Guide

Backend API routes live under `/api`. Requests and responses are JSON unless uploading or downloading attachments. JSON uses `snake_case`, UTC RFC3339 timestamps, and opaque string IDs.

The backend generates an OpenAPI document in-process with Huma and serves it from the same Rayboard binary. No generated spec file or external docs server is required.

JSON API endpoints use a Kubernetes-inspired object shape. Create/update/action requests use `{"spec": {...}}` when they accept JSON; JSON responses use `{"metadata": {...}, "spec": {...}, "status": {...}}` for resources and resource-like computed views. `metadata` holds identity/bookkeeping, `spec` holds desired user-controlled state or request intent, and `status` holds observed/computed server state. List responses use the same envelope with `metadata.count` and `status.items`; each item follows the resource object shape. Empty `204` responses and binary attachment downloads are the practical exceptions.

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| `GET` | `/api/openapi.json` | No | OpenAPI 3.1 JSON document. |
| `GET` | `/api/openapi.yaml` | No | OpenAPI 3.1 YAML document. |
| `GET` | `/api/docs` | No | Swagger UI with embedded static assets, reading `/api/openapi.json`. |
| `GET` | `/api/docs/redoc` | No | Redoc UI with embedded static assets, reading `/api/openapi.json`. |

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

Project, ticket, user/admin, saved-view, and notification list endpoints currently use optional `limit` and `offset` query parameters where implemented. `POST /api/search` uses `spec.limit` and an opaque `spec.cursor`; the response cursor is returned as `status.next_cursor`.

## Health

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| `GET` | `/api/health` | No | Returns backend health. |

Health response uses `metadata.id`, `spec.service`, and `status.state`.

## Auth, Users, Groups, Roles

| Method | Path | Body |
| --- | --- | --- |
| `POST` | `/api/login` | `{"spec":{"username":"admin","password":"..."}}` |
| `POST` | `/api/logout` | none |
| `GET` | `/api/me` | none |
| `GET` | `/api/me/effective-permissions` | Optional `scope=global` or `scope=project&project_id=...`. |
| `GET` | `/api/tokens` | none |
| `POST` | `/api/tokens` | `{"spec":{"name":"local-script"}}` |
| `DELETE` | `/api/tokens/{token_id}` | none |
| `GET` | `/api/users` | none |
| `POST` | `/api/users` | `{"spec":{"username":"alice","display_name":"Alice","password":"","disabled":false}}` |
| `GET` | `/api/users/{user_id}` | none |
| `GET` | `/api/users/{user_id}/effective-permissions` | Optional `scope=global` or `scope=project&project_id=...`; requires global `roles:read`. |
| `PATCH` | `/api/users/{user_id}` | `{"spec":{"disabled":true}}` |
| `DELETE` | `/api/users/{user_id}` | none |
| `GET` | `/api/groups` | none |
| `POST` | `/api/groups` | `{"spec":{"name":"engineering","display_name":"Engineering"}}` |
| `GET` | `/api/groups/{group_id}/members` | none |
| `POST` | `/api/groups/{group_id}/members/{user_id}` | none |
| `DELETE` | `/api/groups/{group_id}/members/{user_id}` | none |
| `GET` | `/api/roles` | none |
| `GET` | `/api/role-bindings` | none |
| `POST` | `/api/role-bindings` | `{"spec":{"role_name":"project_member","subject_type":"group","subject_id":"group_...","scope":"project","project_id":"project_..."}}` |
| `DELETE` | `/api/role-bindings/{binding_id}` | none |

Login and `/api/me` responses use `metadata.user_id`, `spec.username`, `spec.display_name`, and session/auth state under `status`. Effective-permission responses use `metadata.user_id`, requested scope under `spec`, and `status.permissions`. Token, user, group, role, and role-binding responses use `metadata`, `spec`, and `status`. Creating a user with an empty password generates a random password and returns it once in `status.password`. Created API token secrets are returned once in `status.token`.

## Projects and Tickets

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects` | Optional `include_archived=true`, `limit`, `offset`. |
| `POST` | `/api/projects` | `{"spec":{"key":"CORE","name":"Core","description":"Main project","lead_user_id":""}}` |
| `GET` | `/api/projects/{project_id}` | none |
| `GET` | `/api/projects/{project_id}/tickets` | Optional `status`, `assignee_id`, `sprint_id`, `component_id`, `version_id`, `label`, `limit`, `offset`. |
| `POST` | `/api/projects/{project_id}/tickets` | `{"spec":{"title":"Fix login","description":"...","status":"todo","priority":"High","type":"Bug","assignee_id":"","component_id":"","version_id":"","labels":["backend","auth"],"custom_fields":{}}}` |
| `GET` | `/api/tickets/{ticket_id}` | none |
| `PATCH` | `/api/tickets/{ticket_id}` | `{"spec":{...}}` with any subset of `title`, `description`, `status`, `priority`, `type`, `assignee_id`, `component_id`, `version_id`, `parent_ticket_id`, `rank`, `labels`, `custom_fields`. |
| `GET` | `/api/tickets/{ticket_id}/activity` | none |

Ticket statuses are stored as strings. Workflow status APIs define the ordered project-scoped status slugs available to a project.

Project responses use `metadata`, `spec`, and `status`. Project IDs are returned as `metadata.id`; project key, name, description, and lead user are returned in `spec`; archive/delete lifecycle fields are returned in `status`.

Project ticket list/create responses, backlog ticket collection responses, and direct ticket get/update responses use `metadata`, `spec`, and `status`. Ticket IDs and project IDs are returned in `metadata`; editable fields such as title, description, status, priority, assignee, labels, and custom fields are returned in `spec`; server-observed fields such as ticket key, reporter, and delete state are returned in `status`.

Ticket `component_id` and `version_id` assignments are optional. When present, the component or version must belong to the ticket's project. Clearing either field removes the assignment.

Ticket `labels` is a string array on create, update, get, list, board/backlog, and search-related ticket payloads. Labels are normalized to lowercase slugs, deduplicated, and stored directly on the ticket. Updating `labels` replaces the ticket's label set. There are no separate label CRUD endpoints in this slice.

Ticket `custom_fields` is an object keyed by project custom-field key. On create, all required project custom fields must be present. On update, omitting `custom_fields` leaves existing custom-field values unchanged; sending `custom_fields` replaces the ticket's custom-field values and revalidates required fields.

## Backlog

The first backlog API slice is backend/API-only. It lists project backlog tickets in stable backlog order and supports reordering those tickets by writing stable rank/order values. Browser drag/drop, board UI, richer backlog planning, and reports are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/backlog` | none |
| `PATCH` | `/api/projects/{project_id}/backlog` | `{"spec":{"ticket_ids":["ticket_2","ticket_1"]}}` |

Backlog responses use the same persisted ticket shape as project ticket lists, ordered by backlog rank and then deterministic tie-breakers. Reorder requests submit ticket IDs in desired order and only affect tickets in the addressed project. The backend validates that every submitted ticket belongs to the project, writes rank values atomically, and returns the reordered backlog slice.

## Boards and Workflows

The first board/workflow API slice is backend/API-only. It defines ordered project workflow statuses and board definitions whose ordered columns map to those status slugs. Browser board settings UI and board drag/drop are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/statuses` | none |
| `PUT` | `/api/projects/{project_id}/statuses` | `{"spec":{"statuses":[{"slug":"todo","name":"To Do"},{"slug":"in_progress","name":"In Progress"},{"slug":"done","name":"Done"}]}}` |
| `GET` | `/api/projects/{project_id}/boards` | none |
| `POST` | `/api/projects/{project_id}/boards` | `{"spec":{"name":"Development","description":"Team delivery board","status_slugs":["todo","in_progress","done"]}}` |
| `GET` | `/api/boards/{board_id}` | none |
| `PATCH` | `/api/boards/{board_id}` | `{"spec":{...}}` with any subset of `name`, `description`, `status_slugs`. |
| `DELETE` | `/api/boards/{board_id}` | none |
| `GET` | `/api/boards/{board_id}/tickets` | none |

Status and board responses use `metadata`, `spec`, and `status`. Status slugs/names are returned in `spec`; board columns are returned in `status.columns`. Replacing a project's statuses validates slug uniqueness and preserves project ownership. Board columns are derived from the ordered `status_slugs` in the board's project; cross-project status mappings are invalid.

Board ticket responses use `metadata` for the board view identity, `spec.board` for the board definition resource, and `status.columns` for computed columns with ticket resources. Moving tickets between columns continues to use ticket status updates.

## Sprints

The first sprint API slice is backend/API-only. It supports sprint CRUD within a project, starting and completing sprints, and assigning or removing a ticket from a sprint. Browser backlog planning, board drag/drop, and sprint report screens are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/sprints` | Optional `state`. |
| `POST` | `/api/projects/{project_id}/sprints` | `{"spec":{"name":"Sprint 1","goal":"Ship auth fixes","start_date":"2026-06-16","end_date":"2026-06-30"}}` |
| `GET` | `/api/sprints/{sprint_id}` | none |
| `PATCH` | `/api/sprints/{sprint_id}` | `{"spec":{...}}` with any subset of `name`, `goal`, `start_date`, `end_date`. |
| `DELETE` | `/api/sprints/{sprint_id}` | none |
| `POST` | `/api/sprints/{sprint_id}/start` | Starts a planned sprint. |
| `POST` | `/api/sprints/{sprint_id}/complete` | Completes an active sprint. |
| `PUT` | `/api/tickets/{ticket_id}/sprint` | `{"spec":{"sprint_id":"sprint_..."}}` |
| `DELETE` | `/api/tickets/{ticket_id}/sprint` | Removes the ticket from its sprint. |

Sprint responses use `metadata`, `spec`, and `status`. Sprint states are returned in `status.state` and can be `planned`, `active`, or `completed`. Start and complete actions validate state transitions. Ticket assignment keeps the ticket in its existing project; cross-project sprint assignment is invalid.

Burndown, velocity, burnup, sprint report APIs, and other agile analytics are **Planned**.

## Components and Versions

The first components/versions API slice is backend/API-only. It supports project component CRUD, project version/release CRUD, and assignment of tickets to a component or version through ticket create/update fields. Release reports, roadmap timeline screens, component/version UI screens, and advanced release planning UI are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/components` | none |
| `POST` | `/api/projects/{project_id}/components` | `{"spec":{"name":"API","description":"Backend API surface","owner_user_id":"","default_assignee_id":""}}` |
| `GET` | `/api/components/{component_id}` | none |
| `PATCH` | `/api/components/{component_id}` | `{"spec":{...}}` with any subset of `name`, `description`, `owner_user_id`, `default_assignee_id`. |
| `DELETE` | `/api/components/{component_id}` | none |
| `GET` | `/api/projects/{project_id}/versions` | Optional `status`. |
| `POST` | `/api/projects/{project_id}/versions` | `{"spec":{"name":"2026.7","description":"July release","target_date":"2026-07-31","release_date":""}}` |
| `GET` | `/api/versions/{version_id}` | none |
| `PATCH` | `/api/versions/{version_id}` | `{"spec":{...}}` with any subset of `name`, `description`, `status`, `target_date`, `release_date`. |
| `DELETE` | `/api/versions/{version_id}` | none |

Component and version responses use `metadata`, `spec`, and `status`. Version lifecycle state is returned in `status.state`. Component names and version names are unique within a project. Component owner/default assignee IDs are optional user IDs. Version statuses are strings; the first slice accepts `planned`, `released`, and `archived`. Version `target_date` and `release_date` use `YYYY-MM-DD` date strings or empty strings.

Deleting a component or version does not delete tickets. SQLite foreign-key behavior clears affected ticket assignments.

Tickets are assigned to components and versions through the ticket API:

| Method | Path | Body |
| --- | --- | --- |
| `PATCH` | `/api/tickets/{ticket_id}` | `{"spec":{"component_id":"component_...","version_id":"version_..."}}` |

Ticket assignment keeps all records in one project. Cross-project component or version assignment is invalid.

## Roadmap

The roadmap slice is backend/API-only. Epics are regular tickets with `type: "epic"`, optional `start_date` and `due_date`, and direct child tickets linked by `parent_ticket_id`. Browser roadmap timeline screens and drag/drop planning are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/roadmap` | none |

Ticket roadmap dates use `YYYY-MM-DD` date strings or empty strings. Roadmap list items use `metadata` for the epic/project identity, `spec.epic` for the epic ticket resource, and `status.progress` for direct-child progress totals by status, with `done` counting children whose status is `done`. Search and saved views can filter, sort, and display `start_date` and `due_date` where the existing search API supports filters, sort specs, and saved-view columns.

## Custom Fields

The first custom-field API slice is backend/API-only. It supports project-scoped custom field definitions, select options, typed ticket values, and server-side validation during ticket create/update. Browser field management UI, custom-field search/CEL integration, custom create page layouts, and richer field schemas are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/custom-fields` | none |
| `POST` | `/api/projects/{project_id}/custom-fields` | `{"spec":{"key":"severity","name":"Severity","field_type":"single_select","required":true,"options":["Low","High"]}}` |
| `GET` | `/api/custom-fields/{field_id}` | none |
| `PATCH` | `/api/custom-fields/{field_id}` | `{"spec":{...}}` with any subset of `key`, `name`, `field_type`, `required`, `options`. |
| `DELETE` | `/api/custom-fields/{field_id}` | none |

Custom-field responses use `metadata`, `spec`, and `status`. User-provided option values are in `spec.options`; persisted option rows with IDs and positions are returned in `status.options`. Supported field types are `text`, `number`, `boolean`, `date`, `single_select`, `multi_select`, and `user`. `single_select` and `multi_select` fields require configured options. Dates use `YYYY-MM-DD`. User fields store user IDs and validate that the user exists.

Ticket custom-field values use field keys:

```json
{
  "custom_fields": {
    "severity": "High",
    "estimate": 3,
    "needs_review": true,
    "target_date": "2026-07-01",
    "reviewers": ["Alice", "Bob"],
    "owner": "user_..."
  }
}
```

## Comments and Attachments

| Method | Path | Body |
| --- | --- | --- |
| `GET` | `/api/tickets/{ticket_id}/comments` | none |
| `POST` | `/api/tickets/{ticket_id}/comments` | `{"spec":{"body":"Looks reproducible."}}` |
| `DELETE` | `/api/comments/{comment_id}` | none |
| `GET` | `/api/tickets/{ticket_id}/attachments` | none |
| `POST` | `/api/tickets/{ticket_id}/attachments` | multipart form field `file`. |
| `GET` | `/api/attachments/{attachment_id}/download` | binary download. |
| `DELETE` | `/api/attachments/{attachment_id}` | none |

Comment responses and attachment metadata responses use `metadata`, `spec`, and `status`. Comment body is returned in `spec.body`; author is returned in `status.author_id`. Attachment filename and content type are returned in `spec`; upload size and uploader are returned in `status`. Attachment bytes are stored in SQLite. The current maximum upload size is 10 MiB. Downloads set `Content-Type` from stored metadata and `Content-Disposition: attachment`.

## Search and Saved Views

Search endpoint:

```http
POST /api/search
Content-Type: application/json
```

```json
{
  "spec": {
    "project_id": "project_...",
    "filter": "status == \"todo\" && assignee_id == currentUser()",
    "text": "login error",
    "sort": [{"field": "updated_at", "direction": "desc"}],
    "limit": 50,
    "cursor": ""
  }
}
```

Search returns:

```json
{
  "metadata": {"generated_at": "..."},
  "spec": {
    "project_id": "project_...",
    "filter": "status == \"todo\" && assignee_id == currentUser()",
    "text": "login error",
    "sort": [{"field": "updated_at", "direction": "desc"}],
    "limit": 50,
    "cursor": ""
  },
  "status": {
    "items": [
      {
        "metadata": {"id": "ticket_...", "project_id": "project_...", "created_at": "...", "updated_at": "..."},
        "spec": {"title": "Fix login", "status": "todo"},
        "status": {"key": "CORE-12", "reporter_id": "user_..."}
      }
    ],
    "next_cursor": ""
  }
}
```

Current filter support is a constrained expression parser, not full CEL yet. Supported fields are `project`, `project_id`, `key`, `title`, `status`, `priority`, `type`, `reporter_id`, `assignee_id`, `parent_ticket_id`, `sprint_id`, `component_id`, `version_id`, `labels`, `start_date`, and `due_date`. Supported operators are `==`, `!=`, and `&&`. Values are string literals or `currentUser()`. Parentheses are only parsed as part of function calls.

Supported sort fields are `created_at`, `updated_at`, `key`, `title`, `status`, `priority`, `start_date`, and `due_date`.

Full CEL support is **Planned**. See CEL at https://cel.dev/ and cel-go at https://github.com/google/cel-go.

Full-text search uses SQLite FTS5 virtual tables for ticket title/description, comments, and attachment metadata such as filename and content type. See https://www.sqlite.org/fts5.html. Current text input is tokenized into quoted FTS terms; it is not raw FTS syntax.

Saved views:

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/saved-views` | Optional `project_id`, `pinned=true`, `limit`, `offset`. |
| `POST` | `/api/saved-views` | `{"spec":{"scope_type":"user","project_id":"","name":"My bugs","query":{"filter":"assignee_id == currentUser()","text":"bug"},"sort":[{"field":"updated_at","direction":"desc"}],"columns":["key","title","status"],"display_mode":"list","group_by":"","pinned":false}}` |
| `GET` | `/api/saved-views/{view_id}` | none |
| `PATCH` | `/api/saved-views/{view_id}` | `{"spec":{...}}` with any subset of `name`, `query`, `sort`, `columns`, `display_mode`, `group_by`, `pinned`. |
| `DELETE` | `/api/saved-views/{view_id}` | none |

Saved-view responses use `metadata`, `spec`, and `status`. The view ID and timestamps are in `metadata`; scope, project, query, sort, columns, display mode, grouping, and pinned state are in `spec`. Saved-view scopes are `user`, `project`, and `global`. Managing project/global views requires the matching `views:manage` permission. Display modes are `list`, `board`, and `backlog`. Supported grouping fields are `status`, `assignee_id`, `sprint_id`, `component_id`, `version_id`, `priority`, and `type`. Saved-view columns can include built-in ticket fields, including `labels`, `start_date`, and `due_date`. Only project-scoped views can be pinned.

## Notifications

The first notification API slice lists in-app notifications for the authenticated user and supports read/unread state. Runtime notification generation consumes durable `domain_events`, so pending comment and ticket-update notifications can be processed after restart. External Shoutrrr destinations, policies, delivery history, and the background delivery worker are API/backend-only.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/notifications` | Optional `unread=true`, `limit`, `offset`. |
| `POST` | `/api/notifications/{notification_id}/read` | Marks one of the current user's notifications read. |
| `POST` | `/api/notifications/{notification_id}/unread` | Marks one of the current user's notifications unread. |
| `POST` | `/api/notifications/read-all` | Marks all current user's unread notifications read. |

Notification list responses use the standard list envelope; notification resources use `metadata`, `spec`, and `status`:

```json
{
  "metadata": {"count": 1},
  "spec": {},
  "status": {
    "items": [
      {
        "metadata": {
          "id": "notification_...",
          "user_id": "user_...",
          "created_at": "2026-06-16T10:30:00Z"
        },
        "spec": {
          "type": "ticket_assigned",
          "subject_type": "ticket",
          "subject_id": "ticket_...",
          "body": "You were assigned CORE-12",
          "data": {}
        },
        "status": {
          "read_at": null
        }
      }
    ]
  }
}
```

Read state is owned by the notification row. `read_at: null` means unread; marking read sets `read_at`, and marking unread clears it. Users may only list or mutate their own notifications.

## Notification Preferences

Authenticated users can manage their own notification preferences. Project notification managers can manage project notification defaults. Preferences are stored as reusable resource objects; missing rows use enabled defaults and return `status.customized: false`.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/me/notification-preferences` | Current user's notification preferences. |
| `PATCH` | `/api/me/notification-preferences` | `{"spec":{"external_enabled":false,"status_change_enabled":false}}` |
| `GET` | `/api/projects/{project_id}/notification-preferences` | Project defaults; requires project `notifications:manage`. |
| `PATCH` | `/api/projects/{project_id}/notification-preferences` | `{"spec":{"comment_enabled":false}}`; requires project `notifications:manage`. |

Preference responses use `metadata`, `spec`, and `status`. The scope and subject IDs are in `metadata`; booleans such as `in_app_enabled`, `external_enabled`, `assignment_enabled`, `comment_enabled`, `status_change_enabled`, `sprint_change_enabled`, `release_change_enabled`, and `automation_failure_enabled` are in `spec`; `status.customized` shows whether a persisted override exists.

## Notification Destinations

Admins and project notification managers can configure named Shoutrrr destinations for later reuse by notification hooks and delivery policies. Destination create/update uses the standard `spec` body. The Shoutrrr URL is write-only: create/update accepts `spec.shoutrrr_url`, but responses only expose `status.url_set` and `spec.type`.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/notification-destinations` | List global destinations; requires global `notifications:manage`. |
| `POST` | `/api/notification-destinations` | `{"spec":{"name":"ops","shoutrrr_url":"logger://","enabled":true}}` |
| `GET` | `/api/projects/{project_id}/notification-destinations` | List project destinations; requires project `notifications:manage`. |
| `POST` | `/api/projects/{project_id}/notification-destinations` | `{"spec":{"name":"team","shoutrrr_url":"logger://","enabled":true}}` |
| `GET` | `/api/notification-destinations/{destination_id}` | none |
| `PATCH` | `/api/notification-destinations/{destination_id}` | `{"spec":{"name":"ops-primary","shoutrrr_url":"logger://","enabled":false}}`; omit `shoutrrr_url` to keep the existing secret. |
| `POST` | `/api/notification-destinations/{destination_id}/test-send` | `{"spec":{"message":"Rayboard test"}}`; message is optional and defaults to a Rayboard test message. |
| `DELETE` | `/api/notification-destinations/{destination_id}` | soft-deletes and disables the destination. |

Destination responses use `metadata`, `spec`, and `status`. Scope identity and timestamps are in `metadata`; name, Shoutrrr service type, and enabled state are in `spec`; URL presence, last delivery state, and last error are in `status`. Test-send updates `status.last_delivery_status`, `status.last_delivery_at`, and `status.last_error`. CRUD and test-send write security audit entries without storing the URL in audit payloads.

## Notification Policies

Notification policies define which event types should route to named destinations. Policy CRUD is API-only in this slice. When the durable notification processor handles supported domain events, enabled matching global and project policies enqueue idempotent delivery rows for their destinations. Delivery queue rows are represented separately so workers can process or retry them without exposing destination secrets.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/notification-policies` | List global policies; requires global `notifications:manage`. |
| `POST` | `/api/notification-policies` | `{"spec":{"name":"ops","event_types":["ticket_assigned"],"destination_ids":["dest_..."],"enabled":true}}` |
| `GET` | `/api/projects/{project_id}/notification-policies` | List project policies; requires project `notifications:manage`. |
| `POST` | `/api/projects/{project_id}/notification-policies` | `{"spec":{"name":"team","event_types":["comment_added"],"destination_ids":["dest_..."],"enabled":true}}` |
| `GET` | `/api/notification-policies/{policy_id}` | none |
| `PATCH` | `/api/notification-policies/{policy_id}` | `{"spec":{"enabled":false}}` or any subset of name, event types, destination IDs, and enabled state. |
| `DELETE` | `/api/notification-policies/{policy_id}` | soft-deletes and disables the policy. |

Supported policy event types are `ticket_assigned`, `comment_added`, `ticket_status_changed`, `sprint_changed`, `release_changed`, and `automation_failed`. Global policies may use global destinations. Project policies may use global destinations and destinations from the same project. Policy responses use `metadata`, `spec`, and `status`; destination details and Shoutrrr URLs are not embedded in policy responses.

## Notification Deliveries

Notification deliveries are the durable queue/history foundation for external notification sending. They are enqueued from enabled notification policies while processing durable domain events. They snapshot policy and destination names/service at queue time, keep delivery state in `status`, and never expose raw Shoutrrr URLs. The background notification worker sends queued deliveries, resolves the destination secret at processing time, retries transient Shoutrrr failures with backoff, and marks permanent destination errors as failed. Manual retry requeues failed or canceled deliveries for immediate processing.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/notification-deliveries` | Global delivery history; optional `status`, `policy_id`, `destination_id`, `limit`, `offset`; requires global `notifications:manage`. |
| `GET` | `/api/projects/{project_id}/notification-deliveries` | Project delivery history; optional `status`, `policy_id`, `destination_id`, `limit`, `offset`; requires project `notifications:manage`. |
| `GET` | `/api/notification-deliveries/{delivery_id}` | Delivery resource; requires notification management permission for that delivery scope. |
| `POST` | `/api/notification-deliveries/{delivery_id}/retry` | Requeue a failed or canceled delivery. |

Delivery resources use `metadata` for queue identity, scope, policy snapshot, and destination snapshot; `spec` for event/message payload and retry budget; and `status` for current state, attempt counts, timestamps, and last error.

Dashboard/view notification policies, recipient rules, notification hooks, webhooks, and AI/Lua notification hooks are **Planned**.

## OpenRouter Providers

Global admins configure OpenRouter provider references for future AI automation. Provider CRUD requires global `ai:manage`; providers hold global secrets and are not project-scoped. API keys are write-only: create/update accepts `spec.api_key`, but responses only return `status.api_key_set`.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/openrouter-providers` | none |
| `POST` | `/api/openrouter-providers` | `{"spec":{"name":"default","default_model":"openai/gpt-4.1-mini","api_key":"sk-or-...","allowed_models":["openai/gpt-4.1-mini"],"default_timeout_seconds":30,"max_output_tokens":2048,"enabled":true}}` |
| `GET` | `/api/openrouter-providers/{provider_id}` | none |
| `PATCH` | `/api/openrouter-providers/{provider_id}` | `{"spec":{...}}`; omit `api_key` to keep the existing key, provide a non-empty `api_key` to rotate it. |
| `DELETE` | `/api/openrouter-providers/{provider_id}` | soft-deletes and disables the provider. |

Provider responses use `metadata`, `spec`, and `status`. `spec` contains name, default model, allowed model list, limits, and enabled state. `status.api_key_set` reports whether a secret is configured. OpenRouter provider create/update/delete writes security audit entries without storing the key in audit payloads.

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

Cron job CRUD and manual runs require automation management permissions. Run history uses the shared automation run-history model and must not expose secrets. Run resources use `metadata` for run identity/timestamps, `spec` for trigger and input context, and `status` for state, output, error, and start/finish timestamps. The implemented cron slice is Lua-only.

Cron job engine configuration is nested and reusable across future hooks/webhooks:

```json
{
  "spec": {
    "name": "Daily triage",
    "schedule": "0 9 * * *",
    "timezone": "UTC",
    "enabled": true,
    "engine": {
      "type": "lua",
      "script": "rayboard.log(\"triage started\")"
    }
  }
}
```

The OpenAPI schema represents `spec.engine` as a discriminated `oneOf` object. `{"type":"lua"}` requires `script`; `{"type":"ai"}` requires `prompt` and `provider_id`.

The planned AI form uses the same `engine` object with an OpenRouter provider reference:

```json
{
  "spec": {
    "engine": {
      "type": "ai",
      "prompt": "Return JSON actions for stale tickets.",
      "provider_id": "ai_provider_default"
    }
  }
}
```

Implemented cron Lua helpers are `rayboard.log`, `rayboard.search`, `rayboard.get_ticket`, `rayboard.create_ticket`, `rayboard.update_ticket`, and `rayboard.comment`. Helpers execute through normal backend service/RBAC paths as the cron job owner. OpenRouter AI automation, ticket hooks, custom create pages, webhooks, and notification hooks are **Planned**.
