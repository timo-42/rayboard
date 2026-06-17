# Demo Seed

The implemented demo seed command populates a running backend over HTTP:

```bash
rayboard demo seed \
  --backend-url http://127.0.0.1:8081 \
  --admin-user admin \
  --admin-password '<printed-admin-password>' \
  --fresh-reset
```

`--fresh-reset` is required, but the current implementation does not delete existing data. It creates new demo objects with a random suffix.

## Current Behavior

The command:

- logs in through `POST /api/login`;
- uses the returned session cookies and CSRF token;
- creates three demo users with generated passwords;
- creates two demo groups;
- adds demo users to those groups;
- creates one demo project;
- binds a demo lead as `project_owner`, the demo engineers group as `project_member`, and the demo stakeholders group as `project_viewer` for that project;
- replaces project workflow statuses and creates a delivery board;
- creates a component, version/release target, required custom field, sprint, Lua ticket hook, and custom ticket create page;
- submits one ticket through the custom ticket create page;
- creates an epic plus three child tickets with labels, component/version/sprint assignment, roadmap dates, and custom-field values;
- moves one ticket to `in_progress`;
- reorders the backlog;
- adds a comment and a small text attachment to the seeded epic;
- creates a pinned project saved view with a CEL label filter and FTS text query;
- runs one search request using the same CEL/FTS pattern;
- creates a disabled Lua cron job owned by the demo lead so the automation editor has realistic content without scheduling background work;
- creates one incoming Lua webhook and prints its one-time demo token;
- creates one disabled outgoing Lua webhook for ticket update events;
- creates a project Shoutrrr `logger://` notification destination plus disabled policy and hook examples;
- prints generated demo usernames/passwords and seeded object summaries.

The command exercises normal backend validation, permissions, and activity behavior because it calls public API endpoints rather than writing directly to SQLite.

## Planned Expansion

The larger requirements target is **Planned**. Future demo seed work should add destructive reset semantics gated by `--fresh-reset`, more groups, global role bindings, and AI-backed automation examples when OpenRouter is configured.
