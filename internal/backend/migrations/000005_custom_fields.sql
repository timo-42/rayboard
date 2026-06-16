CREATE TABLE custom_field_definitions (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	key TEXT NOT NULL,
	name TEXT NOT NULL,
	field_type TEXT NOT NULL CHECK (field_type IN ('text', 'number', 'boolean', 'date', 'single_select', 'multi_select', 'user')),
	required INTEGER NOT NULL DEFAULT 0 CHECK (required IN (0, 1)),
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(project_id, key),
	UNIQUE(project_id, name)
);

CREATE INDEX custom_field_definitions_project_id_idx ON custom_field_definitions(project_id);

CREATE TABLE custom_field_options (
	id TEXT PRIMARY KEY,
	field_id TEXT NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
	value TEXT NOT NULL,
	position INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(field_id, value)
);

CREATE INDEX custom_field_options_field_id_idx ON custom_field_options(field_id);

CREATE TABLE ticket_custom_field_values (
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	field_id TEXT NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
	value_json TEXT NOT NULL,
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (ticket_id, field_id)
);

CREATE INDEX ticket_custom_field_values_field_id_idx ON ticket_custom_field_values(field_id);
