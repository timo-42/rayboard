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

Swagger UI and the OpenAPI security schemes model bearer tokens as sufficient by themselves. Cookie-authenticated mutating requests are modeled as a separate `sessionCookie` plus `csrfToken` alternative. Bearer-token requests do not need CSRF, even if browser cookies are also present.

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

Ticket `labels` is a string array on create, update, get, list, board/backlog, and search-related ticket payloads. Labels are normalized to lowercase slugs, deduplicated, and stored directly on the ticket. Updating `labels` replaces the ticket's label set. The embedded browser UI accepts comma-separated labels on ticket create and ticket cards. There are no separate label CRUD endpoints in this slice.

Ticket `custom_fields` is an object keyed by project custom-field key. On create, all required project custom fields must be present. On update, omitting `custom_fields` leaves existing custom-field values unchanged; sending `custom_fields` replaces the ticket's custom-field values and revalidates required fields.

Ticket activity responses use a list resource with `metadata.count` and `status.items`. Each activity item uses `metadata.id`, `metadata.ticket_id`, `metadata.created_at`, `spec.activity_type`, optional `spec.data`, and `status.actor_id`. Common activity types include `ticket.created`, `ticket.updated`, `comment.created`, `comment.deleted`, `attachment.uploaded`, and `attachment.deleted`. The embedded issue page at `/issues/{ticket_id}` renders this activity history.

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

The sprint API supports sprint CRUD within a project, starting and completing sprints, and assigning or removing a ticket from a sprint. The embedded browser UI exposes basic sprint list/create/start/complete/delete and ticket assignment/removal for the selected project. Browser backlog planning, board drag/drop, detailed sprint editing, sprint filtering, and sprint report screens are **Planned**.

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

The components/versions API supports project component CRUD, project version/release CRUD, and assignment of tickets to a component or version through ticket create/update fields. The embedded browser UI exposes basic component/version list/create/delete, version release/archive state changes, and ticket component/version assignment. Release reports, roadmap timeline screens, detailed component/version editing/filtering, and advanced release planning UI are **Planned**.

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

The roadmap API lists epics and direct child-ticket progress. Epics are regular tickets with `type: "epic"`, optional `start_date` and `due_date`, and direct child tickets linked by `parent_ticket_id`. The embedded browser UI exposes a basic roadmap panel with epic schedule dates and child progress, plus ticket-form fields for epic creation, parent epic assignment, and roadmap dates. Rich browser roadmap timeline screens and drag/drop planning are **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/roadmap` | none |

Ticket roadmap dates use `YYYY-MM-DD` date strings or empty strings. Roadmap list items use `metadata` for the epic/project identity, `spec.epic` for the epic ticket resource, and `status.progress` for direct-child progress totals by status, with `done` counting children whose status is `done`. Search and saved views can filter, sort, and display `start_date` and `due_date` where the existing search API supports filters, sort specs, and saved-view columns.

## Custom Fields

The custom-field API supports project-scoped field definitions, select options, typed ticket values, server-side validation during ticket create/update, and CEL search filters through `custom.<field_key>`. The embedded browser UI exposes basic field list/create/delete controls and JSON ticket custom-field value editing. Browser field update forms, custom create page layouts, richer field schemas, and field-aware search/layout screens are **Planned**.

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

## Custom Ticket Create Pages

The custom create-page slice is backend/API-only. It lets project owners/admins define project-scoped ticket intake pages with a slug, optional target ticket type/status, a structured field layout, default ticket values, and optional Lua or OpenRouter AI form logic. Submissions create tickets through the normal backend path as the submitting user, so RBAC, ticket validation, custom-field validation, and before ticket hooks still apply. Browser rendering is **Planned**.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/ticket-create-pages` | Optional `include_disabled=true`, `limit`, `offset`; requires project `automations:manage`. |
| `POST` | `/api/projects/{project_id}/ticket-create-pages` | `{"spec":{"name":"Bug Intake","slug":"bug-intake","enabled":true,"target_type":"bug","target_status":"todo","field_layout":[{"key":"title","required":true}],"defaults":{"priority":"high"},"form_lua_script":"return { defaults = { priority = \"High\" } }","form_ai_prompt":"","form_ai_provider_id":"","owner_user_id":"user_..."}}` |
| `GET` | `/api/ticket-create-pages/{page_id}` | none; requires project `automations:manage`. |
| `PATCH` | `/api/ticket-create-pages/{page_id}` | `{"spec":{...}}` with any subset of `name`, `slug`, `description`, `enabled`, `target_type`, `target_status`, `field_layout`, `defaults`, `form_lua_script`, `form_ai_prompt`, `form_ai_provider_id`, `owner_user_id`. |
| `DELETE` | `/api/ticket-create-pages/{page_id}` | none |
| `GET` | `/api/projects/{project_id}/ticket-create-pages/{slug}/schema` | Resolves an enabled page for an authenticated project reader. |
| `POST` | `/api/projects/{project_id}/ticket-create-pages/{slug}/submit` | `{"spec":{"ticket":{"title":"Broken login","description":"...","labels":["customer"],"custom_fields":{}}}}` |

Create-page responses use `metadata` for page/project/owner/timestamps, `spec` for the page definition, and `status.deleted_at` for soft-delete state. Schema responses use `metadata.page_id`, `metadata.project_id`, and `metadata.slug`, return the renderable page definition in `spec`, and expose `status.enabled`. Disabled pages are not resolvable by slug. Submit responses return the normal ticket resource shape.

When `spec.form_lua_script` is set, schema resolution and submission run that Lua script in the shared JSON sandbox with `context`, `page`, `json`, `rayboard.json`, and `rayboard.log(message)`. The script may return a table containing `field_layout`, `defaults`, and/or `description`; returned raw HTML fields are rejected. The script does not receive filesystem, shell, network, database, or Rayboard API helpers.

When `spec.form_ai_prompt` is set, `spec.form_ai_provider_id` must reference an enabled OpenRouter provider. Schema resolution and submission call OpenRouter with the saved prompt plus page/user context, then apply the returned JSON object through the same form-output validator as Lua. AI output may set only `field_layout`, `defaults`, and/or `description`, and raw HTML fields are rejected. A page may use Lua form logic or AI form logic, not both. Schema responses expose the effective renderable form definition but redact `form_lua_script`, `form_ai_prompt`, and `form_ai_provider_id`.

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

Filters are CEL expressions parsed and type-checked with cel-go, then translated to parameterized SQLite. See CEL at https://cel.dev/ and cel-go at https://github.com/google/cel-go.

Current Rayboard CEL filter support:

- fields: `project`, `project_id`, `key`, `title`, `status`, `priority`, `type`, `reporter_id`, `assignee_id`, `parent_ticket_id`, `sprint_id`, `component_id`, `version_id`, `labels`, `start_date`, `due_date`, `created_at`, and `updated_at`;
- custom fields through `custom.<field_key>`, for example `custom.severity == "critical"` or `custom.impact >= 8`;
- operators: `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`, parentheses, and `in` against literal lists or labels;
- string helpers on string ticket fields: `contains`, `startsWith`, and `endsWith`;
- approved functions: `currentUser()`, `today()`, and `now()`.

Examples:

```cel
(status == "todo" || key in ["CORE-12", "CORE-13"]) && assignee_id == currentUser()
"backend" in labels && due_date <= today()
custom.severity == "critical" && custom.impact >= 8
title.contains("login")
```

CEL filters do not expose direct SQL, arbitrary host functions, or an unrestricted function registry. Sorting and pagination remain separate API inputs.

Supported sort fields are `created_at`, `updated_at`, `key`, `title`, `status`, `priority`, `start_date`, and `due_date`.

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

Admins and project notification managers can configure named Shoutrrr destinations for later reuse by notification hooks and delivery policies through the API or browser `/settings` page. Destination create/update uses the standard `spec` body. The Shoutrrr URL is write-only: create/update accepts `spec.shoutrrr_url`, but responses only expose `status.url_set` and `spec.type`.

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

## Notification Hooks

Notification hooks are API-only global/project automation rules that run after a notification policy matches and before external Shoutrrr delivery rows are enqueued. Hooks use the shared `engine` object with Lua or OpenRouter AI. They receive notification plan context without raw Shoutrrr URLs or secrets, and may return `suppress`, `message`, `payload`, and `destination_ids`. `destination_ids` is constrained to the current policy's already-allowed destination IDs.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/notification-hooks` | Lists global notification hooks; requires global `notifications:manage`. |
| `POST` | `/api/notification-hooks` | Creates a global notification hook. |
| `GET` | `/api/projects/{project_id}/notification-hooks` | Lists project notification hooks; requires project `notifications:manage`. |
| `POST` | `/api/projects/{project_id}/notification-hooks` | Creates a project notification hook. |
| `GET` | `/api/notification-hooks/{hook_id}` | Gets one notification hook. |
| `PATCH` | `/api/notification-hooks/{hook_id}` | Updates a notification hook. |
| `POST` | `/api/notification-hooks/{hook_id}/preview` | Executes a saved hook against a supplied sample notification plan and returns the transformed plan plus run metadata. |
| `GET` | `/api/notification-hooks/{hook_id}/runs` | Lists saved hook preview and event-triggered execution history. |
| `DELETE` | `/api/notification-hooks/{hook_id}` | Soft-deletes and disables a notification hook. |

Preview requests use `{"spec":{"event_type":"ticket_assigned","message":"Assigned AUTO-1","payload":{"ticket_id":"ticket_123"},"destination_ids":["dest_123"]}}`. Hook run resources use `metadata` for run identity, `spec` for trigger/input context, and `status` for state, output, error, and timestamps. Dashboard/view notification policies, recipient rules, and richer destination-name routing are **Planned**.

## Webhooks

The webhook slice implements project-scoped incoming and outgoing webhook definitions. Incoming webhooks have hashed bearer tokens, one-time token display, token rotation, Lua or AI execution for authenticated incoming requests, constrained Rayboard action helpers, and shared automation run history. Outgoing webhook definitions can be created, listed, updated, and deleted with `spec.event_types`, and the backend has a durable queued-delivery model for matching domain events. Due outgoing deliveries can be shaped by Lua or AI, sent through a controlled HTTP client to the configured outgoing webhook base URL, retried with backoff, inspected through the API, and manually requeued after failure.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/webhooks` | Optional `direction`, `limit`, `offset`; requires project `webhooks:manage`. |
| `POST` | `/api/projects/{project_id}/webhooks` | `{"spec":{"name":"github","direction":"incoming","enabled":true,"actor_user_id":"user_...","event_types":["ticket.updated"],"engine":{"type":"lua","script":"return { ok = true }"}}}`; `direction` may be `incoming` or `outgoing`. |
| `GET` | `/api/webhook-definitions/{webhook_id}` | Webhook definition; token is never returned. |
| `PATCH` | `/api/webhook-definitions/{webhook_id}` | Any subset of `name`, `enabled`, `actor_user_id`, `event_types`, and `engine`. |
| `POST` | `/api/webhook-definitions/{webhook_id}/rotate-token` | Rotates an incoming webhook bearer token and returns the new token once in `status.token`. |
| `GET` | `/api/webhook-definitions/{webhook_id}/runs` | Lists Lua/AI run history for an incoming webhook. |
| `GET` | `/api/webhook-definitions/{webhook_id}/deliveries` | Lists queued outgoing webhook deliveries for one webhook definition. |
| `GET` | `/api/webhook-deliveries/{delivery_id}` | Gets one outgoing webhook delivery record. |
| `POST` | `/api/webhook-deliveries/{delivery_id}/retry` | Requeues a failed or canceled outgoing webhook delivery. |
| `DELETE` | `/api/webhook-definitions/{webhook_id}` | Soft-deletes the webhook and clears its token hash. |
| `POST` | `/api/webhooks/incoming/{webhook_id}` | Authenticates with `Authorization: Bearer <webhook-token>`, accepts `{"spec":{"payload":{...},"headers":{},"query":{}}}`, executes the webhook engine, and returns a run resource. |

Webhook definition responses use `metadata` for IDs/timestamps, `spec` for direction, actor user, enabled state, `event_types`, and engine configuration, and `status` for `token_set`, `token_rotated_at`, and `last_error`. Incoming create and rotate responses are the only responses that include `status.token`. Outgoing webhook definitions do not have bearer tokens, so `status.token_set` is false and token rotation returns a validation error. Outgoing webhooks require at least one `event_types` entry; values are domain event names such as `ticket.updated` or `comment.created`.

Outgoing delivery responses use `metadata` for delivery ID, webhook snapshot, domain event ID, idempotency key, project ID, and timestamps; `spec` for event type, subject, payload, and max attempts; and `status` for queue state, attempt count, attempt timestamps, delivery timestamp, and last error. Delivery inspection and retry require project `webhooks:manage` through the owning webhook.

Incoming webhook Lua helpers are `rayboard.log`, `rayboard.search`, `rayboard.get_ticket`, `rayboard.create_ticket`, `rayboard.update_ticket`, and `rayboard.comment`. AI incoming webhooks use the same actor/RBAC path through a JSON `actions` array with action types `search`, `get_ticket`, `create_ticket`, `update_ticket`, and `comment`; AI can also return `reject` to fail the webhook before actions are applied. Disabled or deleted actor users cause incoming execution to fail before Lua or AI runs. Outgoing Lua scripts receive `event`, `webhook`, and `delivery` tables and must return a request object with `method`, relative `path`, optional `query`, optional `headers`, and optional JSON `body`. AI outgoing webhooks receive the same context in the prompt and must return the same request object. Outgoing delivery accepts only relative paths and sends them under `--outgoing-webhook-base-url`; arbitrary hosts, URL credentials, unsupported methods, `Host`, and `Content-Length` headers are rejected.

## Ticket Hooks

Project ticket hooks are project-scoped automation resources that run during ticket create/update. Before hooks can validate or transform the pending ticket payload; after hooks run after commit and cannot roll back committed changes. The `/automation` browser UI provides basic project ticket-hook list, create, delete, enable/disable, and preview controls; the API remains the source of truth for the full resource shape.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/projects/{project_id}/ticket-hooks` | Optional `event`, `phase`, `limit`, `offset`; requires project `automations:manage`. |
| `POST` | `/api/projects/{project_id}/ticket-hooks` | `{"spec":{"name":"normalize-title","event":"ticket_create","phase":"before","enabled":true,"position":100,"engine":{"type":"lua","script":"return { ticket = ticket }"}}}` |
| `GET` | `/api/ticket-hooks/{hook_id}` | Ticket hook definition. |
| `PATCH` | `/api/ticket-hooks/{hook_id}` | Any subset of `name`, `event`, `phase`, `enabled`, `position`, and `engine`. |
| `POST` | `/api/ticket-hooks/{hook_id}/preview` | `{"spec":{"ticket":{"title":"Example"},"current":{}}}`; executes one saved hook without changing tickets or `last_error`. |
| `DELETE` | `/api/ticket-hooks/{hook_id}` | Soft-deletes and disables the hook. |

Ticket hook responses use `metadata` for IDs/project/timestamps, `spec` for event, phase, order, enabled state, and engine configuration, and `status.last_error` for the last persisted execution error. Preview responses use `metadata.hook_id`, echo the preview input in `spec`, and return `status.output`, transformed `status.ticket` when present, `status.logs`, and `status.error` for reject/runtime errors. The OpenAPI schema represents `spec.engine` as a discriminated `oneOf` object. Lua hooks require `engine.script`; AI hooks require `engine.prompt` and `engine.provider_id`. AI hook prompts receive hook context and ticket input, must return a JSON object, and use the same top-level `ticket` transform and `reject` response contract as Lua hooks. AI hooks call OpenRouter through the configured provider without exposing API keys.

## OpenRouter Providers

Global admins configure OpenRouter provider references for AI automation through the API or browser `/settings` page. Provider CRUD requires global `ai:manage`; providers hold global secrets and are not project-scoped. API keys are write-only: create/update accepts `spec.api_key`, but responses only return `status.api_key_set`.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/openrouter-providers` | none |
| `POST` | `/api/openrouter-providers` | `{"spec":{"name":"default","default_model":"openai/gpt-4.1-mini","api_key":"sk-or-...","allowed_models":["openai/gpt-4.1-mini"],"default_timeout_seconds":30,"max_output_tokens":2048,"enabled":true}}` |
| `GET` | `/api/openrouter-providers/{provider_id}` | none |
| `PATCH` | `/api/openrouter-providers/{provider_id}` | `{"spec":{...}}`; omit `api_key` to keep the existing key, provide a non-empty `api_key` to rotate it. |
| `DELETE` | `/api/openrouter-providers/{provider_id}` | soft-deletes and disables the provider. |

Provider responses use `metadata`, `spec`, and `status`. `spec` contains name, default model, allowed model list, limits, and enabled state. `status.api_key_set` reports whether a secret is configured. OpenRouter provider create/update/delete writes security audit entries without storing the key in audit payloads.

## Global Settings

Global settings require global `settings:manage` and are available through the API and the browser `/settings` page.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/settings` | none |
| `PATCH` | `/api/settings` | `{"spec":{"attachment_max_size_bytes":10485760,"attachment_allowed_content_types":["text/plain"],"webhook_allowed_base_urls":["https://example.com/hooks"],"demo_warning_enabled":true,"backup_enabled":false,"system_health_note":"green"}}` |

Responses use `metadata`, `spec`, and `status`. `metadata` contains the global settings ID, update time, and last updater when set. `spec` contains attachment policy, webhook allowlist metadata, demo warning, backup flag, and health note. `status` reports whether attachment policy, webhook allowlist, demo warning, and backup availability are active.

Attachment uploads enforce `attachment_max_size_bytes` and `attachment_allowed_content_types`. An empty content-type list permits all content types. Outgoing webhook delivery enforces `webhook_allowed_base_urls` when the list is non-empty; if no process-level outgoing webhook base URL is configured, the first allowed base URL is used. Allowed webhook base URLs must be absolute `http` or `https` URLs without credentials, query strings, or fragments. Settings updates write `settings.updated` audit entries with changed field names only.

## Audit Log

Security audit log listing requires global `settings:manage` and is available through the API and the browser `/settings` page.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/audit-log` | optional query: `limit`, `event_type`, `actor_user_id`, `subject_type`, `subject_id`, `outcome` |

Responses use `metadata`, `spec`, and `status`. The list metadata contains the returned entry count. Each item includes `metadata.id` and `metadata.occurred_at`; `spec` includes event type, actor user ID, auth kind, subject type/ID, outcome, and payload metadata; `status.security_event` is true for these entries. Payloads must not contain plaintext generated passwords, API tokens, password hashes, session secrets, Shoutrrr URLs, webhook tokens, or OpenRouter keys.

## Engine Test

The generic engine test endpoint executes an inline Lua, OpenRouter AI, or initial WASI WebAssembly engine against supplied JSON input before attaching that engine to cron jobs, hooks, webhooks, notification hooks, or custom create pages. Use `surface: "scratch"` for a generic playground run that is not tied to a real automation surface.

| Method | Path | Body or Query |
| --- | --- | --- |
| `POST` | `/api/engines/test` | `{"spec":{"surface":"scratch","project_id":"","actor_user_id":"","engine":{"type":"lua","script":"return { ok = true, input = input }"},"context":{},"input":{"title":"Preview"},"dry_run":true}}` |

Engine tests require `automations:manage` globally or for `spec.project_id`. Browser-session requests require CSRF like other mutating API calls. `actor_user_id` defaults to the current user and must reference an enabled user.

The current endpoint is dry-run and mutation-free; `spec.dry_run` is normalized to `true`, and missing `spec.surface` defaults to `scratch`. Lua receives `context`, `input`, `json`, `rayboard.json`, and `rayboard.log(message)`, but it does not receive ticket/comment/search helper functions. `spec.context` is merged with normalized `surface`, `project_id`, `actor_user_id`, and `dry_run` values before execution. AI tests append the supplied context/input to the prompt and call the selected OpenRouter provider in JSON-object mode. WASM tests use `engine.type = "wasm"` with `engine.module_base64`, a base64-encoded WASI command module. Rayboard writes `surface`, `context`, `input`, and `dry_run` JSON to stdin; the module must write one JSON object to stdout. Stderr lines are captured as logs. No filesystem mounts, environment variables, shell, raw sockets, direct SQLite handles, or Rayboard secrets are exposed.

Responses use `metadata`, `spec`, and `status`. The response redacts `spec.engine.script`, `spec.engine.prompt`, and `spec.engine.module_base64`; run history stores the engine type, actor, surface, normalized context, dry-run flag, and supplied input, not raw source, prompt, or module bytes. `status.output` contains the validated engine output, `status.logs` contains captured log lines, `status.duration_millis` reports elapsed execution time when available, and `status.engine` contains redacted engine metadata. Runtime failures return a normal resource response with `status.state = "failed"` and `status.error`.

When `spec.surface` is `custom_create_page`, the workbench validates returned form data before marking the run successful. Output must include at least one of `field_layout`, `defaults`, or `description`; `field_layout` must be an array of objects; nested `fields` arrays are checked recursively; `defaults` must be an object; `description` must be a string; and raw `html` fields are rejected. Invalid surface output is recorded as a failed run and returned with HTTP `200`, `status.state = "failed"`, and `status.error`.

## Cron Jobs

The cron automation API/scheduler slice exposes Lua and OpenRouter AI cron job management, manual execution, and run history. Cron jobs execute as their owner user and use the owner's current effective RBAC permissions at run time.

| Method | Path | Body or Query |
| --- | --- | --- |
| `GET` | `/api/cron-jobs` | Optional list filters where implemented. |
| `POST` | `/api/cron-jobs` | Cron job definition. |
| `GET` | `/api/cron-jobs/{cron_job_id}` | none |
| `PATCH` | `/api/cron-jobs/{cron_job_id}` | Cron job definition updates. |
| `DELETE` | `/api/cron-jobs/{cron_job_id}` | none |
| `POST` | `/api/cron-jobs/{cron_job_id}/run` | Starts a manual run. |
| `GET` | `/api/cron-jobs/{cron_job_id}/runs` | Run history for the job. |

Cron job CRUD and manual runs require automation management permissions. Run history uses the shared automation run-history model and must not expose secrets. Run resources use `metadata` for run identity/timestamps, `spec` for trigger and input context, and `status` for state, output, error, and start/finish timestamps. Lua cron jobs execute scripts with constrained Rayboard helpers. AI cron jobs call the configured OpenRouter provider, require JSON-object output, and store the validated object plus model/usage metadata in run history without exposing prompts or API keys.

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

AI cron jobs use the same `engine` object with an OpenRouter provider reference:

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

Implemented cron Lua helpers are `rayboard.log`, `rayboard.search`, `rayboard.get_ticket`, `rayboard.create_ticket`, `rayboard.update_ticket`, and `rayboard.comment`. Helpers execute through normal backend service/RBAC paths as the cron job owner. Incoming webhook Lua exposes the same constrained action helper set as the webhook actor. Outgoing webhook Lua/AI shapes controlled outbound HTTP requests from domain-event context. Notification hooks use Lua/AI to suppress, transform, or narrow destination routing for notification plans before delivery rows are enqueued, with saved-hook preview and run history APIs. The backend ticket-hook runner and preview API expose `context`, `ticket`, optional `current`, and `rayboard.log`. Custom create-page Lua exposes `context`, `page`, JSON helpers, and `rayboard.log` for structured schema/default transformation. Custom create-page AI uses OpenRouter to return the same validated structured schema/default output. AI cron action execution is **Planned**.
