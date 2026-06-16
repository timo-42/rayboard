CREATE TABLE notification_destinations (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	scope_type TEXT NOT NULL CHECK (scope_type IN ('global', 'project', 'dashboard')),
	scope_key TEXT NOT NULL,
	project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
	dashboard_id TEXT,
	service TEXT NOT NULL,
	shoutrrr_url_secret TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
	last_delivery_status TEXT,
	last_delivery_at TEXT,
	last_error TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	deleted_at TEXT,
	UNIQUE(scope_type, scope_key, name)
);

CREATE INDEX notification_destinations_scope_idx ON notification_destinations(scope_type, scope_key, enabled);
