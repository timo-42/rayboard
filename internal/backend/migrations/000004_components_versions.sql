CREATE TABLE project_components (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	owner_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	default_assignee_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(project_id, name)
);

CREATE INDEX project_components_project_id_idx ON project_components(project_id);

CREATE TABLE project_versions (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	status TEXT NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'released', 'archived')),
	target_date TEXT,
	release_date TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(project_id, name)
);

CREATE INDEX project_versions_project_id_idx ON project_versions(project_id);
CREATE INDEX project_versions_status_idx ON project_versions(status);

ALTER TABLE tickets ADD COLUMN component_id TEXT REFERENCES project_components(id) ON DELETE SET NULL;
ALTER TABLE tickets ADD COLUMN version_id TEXT REFERENCES project_versions(id) ON DELETE SET NULL;
CREATE INDEX tickets_component_id_idx ON tickets(component_id);
CREATE INDEX tickets_version_id_idx ON tickets(version_id);
