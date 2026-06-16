# Agent 07: Notifications And Webhooks

## Mission

Implement incoming/outgoing webhooks, in-app notifications, Shoutrrr destinations, notification policies, notification hooks, and delivery queues.

Read first:

- `plan/contracts/automation-engines.md`
- `plan/contracts/authorization.md`
- `plan/contracts/api-conventions.md`

## Deliverables

- incoming webhook endpoints
- outgoing webhook event triggers
- webhook delivery queue/history
- in-app notifications
- Shoutrrr destinations
- notification policies
- Lua/AI notification hooks
- delivery retries/backoff

## Package Tasks

1. Incoming webhooks:
   - project scoped
   - bearer token creation/rotation/hash storage
   - stable endpoint
   - actor user
   - Lua/AI validation/action mapping
   - shared Lua `json`/table conversion from Agent 06 for request payloads
2. Outgoing webhooks:
   - subscriptions consume durable `domain_events`
   - queued delivery
   - Lua/AI request shaping
   - shared Lua `json`/table conversion from Agent 06 for outbound payloads
   - destination allowlist
   - timeout/max payload
   - retry/backoff/manual redelivery
3. In-app notifications:
   - assignment
   - mentions
   - comments
   - watched tickets
   - status/sprint/release changes
   - automation failures
4. Shoutrrr destinations:
   - global/project/dashboard scopes
   - named destinations
   - secret redaction/rotation
   - test-send
5. Notification policies:
   - event selectors
   - recipients
   - allowed destinations
   - global/project/dashboard scopes
6. Notification hooks:
   - Lua/AI plan transformation
   - shared Lua `json`/table conversion from Agent 06 for notification plans
   - route to named destinations only
   - no raw Shoutrrr URL or secret exposure

## Integration Points

- Agent 03 writes durable domain events for ticket/project mutations.
- Agent 06 provides automation engine wrappers and the shared Lua JSON/table conversion layer.
- Agent 02 provides RBAC and actor principals.
- Agent 05 builds destination/policy/history UI.
- Agent 08 keeps security/admin audit separate from webhook and notification delivery history.

## Tests

- incoming token one-time display/hash/rotation.
- incoming Lua/AI action validation and RBAC.
- outgoing domain-event enqueue.
- outgoing delivery allowlist/retry/history.
- in-app notification read/unread/preferences.
- Shoutrrr secret redaction/test-send.
- destination inheritance global -> project -> dashboard/view.
- notification hooks suppress/transform/route.
- no secrets in logs/run history.

## Acceptance Criteria

- Outgoing delivery never blocks ticket/project mutations.
- Hooks cannot call Shoutrrr directly.
- Delivery queue resolves destination config at processing time.
- Webhook/notification processing can resume from durable domain events after restart.
