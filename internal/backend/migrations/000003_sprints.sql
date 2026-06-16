CREATE TABLE sprints (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	goal TEXT,
	state TEXT NOT NULL DEFAULT 'planned' CHECK (state IN ('planned', 'active', 'completed')),
	start_date TEXT,
	end_date TEXT,
	started_at TEXT,
	completed_at TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX sprints_project_id_idx ON sprints(project_id);
CREATE INDEX sprints_state_idx ON sprints(state);

ALTER TABLE tickets ADD COLUMN sprint_id TEXT REFERENCES sprints(id) ON DELETE SET NULL;
CREATE INDEX tickets_sprint_id_idx ON tickets(sprint_id);
