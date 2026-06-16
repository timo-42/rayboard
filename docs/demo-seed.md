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
- binds the demo engineers group to `project_member` for that project;
- creates three demo tickets;
- moves one ticket to `in_progress`;
- prints generated demo usernames/passwords and seeded object summaries.

The command exercises normal backend validation, permissions, and activity behavior because it calls public API endpoints rather than writing directly to SQLite.

## Planned Expansion

The larger requirements target is **Planned**. Future demo seed work should add destructive reset semantics gated by `--fresh-reset`, more groups, global/project role bindings, components, versions, workflows, boards, backlog ordering, sprints, epics, comments, labels, custom fields, saved CEL query examples, Lua cron jobs, ticket hooks, and custom create pages.

