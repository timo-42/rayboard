CREATE TABLE domain_events (
	id TEXT PRIMARY KEY,
	event_type TEXT NOT NULL,
	actor_id TEXT,
	project_id TEXT,
	subject_type TEXT NOT NULL,
	subject_id TEXT NOT NULL,
	related_type TEXT,
	related_id TEXT,
	payload_json TEXT NOT NULL DEFAULT '{}',
	occurred_at TEXT NOT NULL,
	processing_status TEXT NOT NULL DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'processed', 'failed')),
	attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
	next_attempt_at TEXT,
	processed_at TEXT,
	last_error TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX domain_events_event_type_idx ON domain_events(event_type);
CREATE INDEX domain_events_project_id_idx ON domain_events(project_id);
CREATE INDEX domain_events_subject_idx ON domain_events(subject_type, subject_id);
CREATE INDEX domain_events_processing_idx ON domain_events(processing_status, next_attempt_at, occurred_at);
