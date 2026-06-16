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
2. Outgoing webhooks:
   - event subscriptions
   - queued delivery
   - Lua/AI request shaping
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
   - route to named destinations only
   - no raw Shoutrrr URL or secret exposure

## Integration Points

- Agent 03 emits events.
- Agent 06 provides automation engine wrappers.
- Agent 02 provides RBAC and actor principals.
- Agent 05 builds destination/policy/history UI.

## Tests

- incoming token one-time display/hash/rotation.
- incoming Lua/AI action validation and RBAC.
- outgoing event enqueue.
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
