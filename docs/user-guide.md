# User Guide

The current browser UI is a small proof-of-concept board shell. Log in with the startup-generated admin credentials, then use the page at `/` to create projects, create tickets, select a project, and move tickets between `todo`, `in_progress`, and `done`.

Implemented browser workflows:

- login and logout;
- project list and project creation;
- ticket list for the selected project;
- ticket creation;
- ticket status changes.

Implemented API-only user workflows:

- comments on tickets;
- attachments on tickets;
- saved views;
- text and filter search;
- Lua cron job management, manual runs, and run history.

See [API Guide](api.md) for endpoint details.

## Planned Jira-Like Workflows

Backlog, boards, sprints, roadmaps, custom fields, workflows, components, versions, labels, custom create pages, and richer saved views are **Planned**. Custom create pages should return structured form definitions and options, not raw HTML. Ticket hooks, webhooks, notification hooks, and OpenRouter AI automation are also **Planned**.

## Search

Current search supports full-text search over ticket title, description, and comments with SQLite FTS5, plus a constrained filter expression subset. Full CEL-backed queries are **Planned**. See:

- CEL: https://cel.dev/
- cel-go: https://github.com/google/cel-go
- SQLite FTS5: https://www.sqlite.org/fts5.html
