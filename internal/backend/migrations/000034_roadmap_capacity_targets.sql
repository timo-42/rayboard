CREATE TABLE roadmap_capacity_targets (
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	month TEXT NOT NULL CHECK (month GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]'),
	target_points REAL NOT NULL CHECK (target_points > 0),
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (project_id, month)
);
