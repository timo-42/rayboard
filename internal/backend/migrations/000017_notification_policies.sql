CREATE TABLE notification_policies (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	scope_type TEXT NOT NULL CHECK (scope_type IN ('global', 'project')),
	scope_key TEXT NOT NULL,
	project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
	event_types_json TEXT NOT NULL DEFAULT '[]',
	destination_ids_json TEXT NOT NULL DEFAULT '[]',
	enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	deleted_at TEXT,
	UNIQUE(scope_type, scope_key, name)
);

CREATE INDEX notification_policies_scope_idx ON notification_policies(scope_type, scope_key, enabled);
CREATE INDEX notification_policies_project_id_idx ON notification_policies(project_id);
