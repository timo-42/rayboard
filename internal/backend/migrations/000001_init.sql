CREATE TABLE users (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	password_hash TEXT,
	is_disabled INTEGER NOT NULL DEFAULT 0 CHECK (is_disabled IN (0, 1)),
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	deleted_at TEXT
);

CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL UNIQUE,
	expires_at TEXT NOT NULL,
	last_seen_at TEXT,
	revoked_at TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);

CREATE TABLE api_tokens (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	token_hash TEXT NOT NULL UNIQUE,
	scopes_json TEXT NOT NULL DEFAULT '[]',
	expires_at TEXT,
	last_used_at TEXT,
	revoked_at TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX api_tokens_user_id_idx ON api_tokens(user_id);

CREATE TABLE groups (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE group_memberships (
	group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (group_id, user_id)
);

CREATE INDEX group_memberships_user_id_idx ON group_memberships(user_id);

CREATE TABLE roles (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE role_permissions (
	role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	permission TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (role_id, permission)
);

CREATE TABLE role_bindings (
	id TEXT PRIMARY KEY,
	role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'group')),
	subject_id TEXT NOT NULL,
	resource_type TEXT,
	resource_id TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX role_bindings_role_id_idx ON role_bindings(role_id);
CREATE INDEX role_bindings_subject_idx ON role_bindings(subject_type, subject_id);
CREATE INDEX role_bindings_resource_idx ON role_bindings(resource_type, resource_id);

CREATE TABLE projects (
	id TEXT PRIMARY KEY,
	key TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT,
	lead_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	archived_at TEXT,
	deleted_at TEXT
);

CREATE INDEX projects_lead_user_id_idx ON projects(lead_user_id);

CREATE TABLE tickets (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	key TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL DEFAULT 'todo',
	priority TEXT,
	type TEXT,
	reporter_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	assignee_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	parent_ticket_id TEXT REFERENCES tickets(id) ON DELETE SET NULL,
	rank TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	deleted_at TEXT
);

CREATE INDEX tickets_project_id_idx ON tickets(project_id);
CREATE INDEX tickets_assignee_id_idx ON tickets(assignee_id);
CREATE INDEX tickets_status_idx ON tickets(status);

CREATE TABLE ticket_comments (
	id TEXT PRIMARY KEY,
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	author_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	body TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	deleted_at TEXT
);

CREATE INDEX ticket_comments_ticket_id_idx ON ticket_comments(ticket_id);
CREATE INDEX ticket_comments_author_id_idx ON ticket_comments(author_id);

CREATE TABLE ticket_activity (
	id TEXT PRIMARY KEY,
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	actor_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	activity_type TEXT NOT NULL,
	data_json TEXT NOT NULL DEFAULT '{}',
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX ticket_activity_ticket_id_idx ON ticket_activity(ticket_id);
CREATE INDEX ticket_activity_actor_id_idx ON ticket_activity(actor_id);

CREATE TABLE ticket_attachments (
	id TEXT PRIMARY KEY,
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	comment_id TEXT REFERENCES ticket_comments(id) ON DELETE SET NULL,
	uploader_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	file_name TEXT NOT NULL,
	content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
	size_bytes INTEGER NOT NULL CHECK (size_bytes >= 0),
	data BLOB NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	deleted_at TEXT
);

CREATE INDEX ticket_attachments_ticket_id_idx ON ticket_attachments(ticket_id);
CREATE INDEX ticket_attachments_comment_id_idx ON ticket_attachments(comment_id);

CREATE TABLE saved_views (
	id TEXT PRIMARY KEY,
	owner_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
	scope_type TEXT NOT NULL DEFAULT 'user' CHECK (scope_type IN ('user', 'project', 'global')),
	name TEXT NOT NULL,
	query_json TEXT NOT NULL DEFAULT '{}',
	sort_json TEXT NOT NULL DEFAULT '[]',
	columns_json TEXT NOT NULL DEFAULT '[]',
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX saved_views_owner_user_id_idx ON saved_views(owner_user_id);
CREATE INDEX saved_views_project_id_idx ON saved_views(project_id);

CREATE TABLE automation_runs (
	id TEXT PRIMARY KEY,
	trigger_type TEXT NOT NULL,
	trigger_ref TEXT,
	project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
	ticket_id TEXT REFERENCES tickets(id) ON DELETE SET NULL,
	status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'canceled')),
	input_json TEXT NOT NULL DEFAULT '{}',
	output_json TEXT NOT NULL DEFAULT '{}',
	error TEXT,
	started_at TEXT,
	finished_at TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX automation_runs_status_idx ON automation_runs(status);
CREATE INDEX automation_runs_ticket_id_idx ON automation_runs(ticket_id);
CREATE INDEX automation_runs_project_id_idx ON automation_runs(project_id);

CREATE TABLE notifications (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	subject_type TEXT,
	subject_id TEXT,
	body TEXT NOT NULL,
	data_json TEXT NOT NULL DEFAULT '{}',
	read_at TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX notifications_user_id_idx ON notifications(user_id);
CREATE INDEX notifications_unread_user_id_idx ON notifications(user_id, read_at);

CREATE VIRTUAL TABLE ticket_fts USING fts5(
	ticket_id UNINDEXED,
	title,
	description
);

CREATE VIRTUAL TABLE comment_fts USING fts5(
	comment_id UNINDEXED,
	body
);
