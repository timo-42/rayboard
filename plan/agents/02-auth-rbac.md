# Agent 02: Auth And RBAC

## Mission

Implement users, password auth, sessions, API tokens, groups, roles, role bindings, and the single authorization evaluator.

Read first:

- `plan/contracts/authorization.md`
- `plan/contracts/api-conventions.md`
- `plan/contracts/data-model.md`

## Deliverables

- admin bootstrap
- password hashing
- login/logout/me endpoints
- session cookies and CSRF
- API token CRUD and bearer auth
- users/groups/roles/bindings APIs
- RBAC evaluator
- middleware for backend API

## Package Tasks

1. Implement password hashing with Argon2id or bcrypt.
2. On backend startup:
   - ensure `admin` user exists
   - generate random password
   - hash/store it
   - assign global admin role
   - log plaintext once
3. Implement sessions:
   - opaque random token
   - hashed storage
   - cookie flags
   - expiry/revocation
4. Implement CSRF for cookie-authenticated mutations.
5. Implement API tokens:
   - show plaintext once
   - store hash
   - revoke/list
   - optional expiry
6. Implement RBAC:
   - built-in roles
   - permissions
   - group memberships
   - role bindings
   - effective permission evaluation
7. Add middleware:
   - session auth
   - bearer auth
   - principal injection
   - require permission helper

## Integration Points

- Every agent uses `Require(principal, permission, scope)`.
- Agent 06 automation actors resolve through this evaluator.
- Agent 07 webhooks/notifications use actor effective permissions.
- Agent 05 login/admin UI uses these endpoints.

## Tests

- admin password reset updates DB and logs plaintext once.
- disabled users cannot login, use sessions, use tokens, or run automations.
- session cookie flags and logout behavior.
- token one-time display, hash storage, bearer auth, revoke.
- direct role, group role, global/project scope, deny-by-default.
- role binding changes affect existing sessions/tokens on next request.
- CSRF required for mutating cookie requests.

## Acceptance Criteria

- No feature code checks user IDs or role names directly except through RBAC evaluator.
- Existing sessions/tokens do not cache stale permissions.
- API returns `401` for unauthenticated and `403` for unauthorized.
