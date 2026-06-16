CREATE TABLE audit_log (
	id TEXT PRIMARY KEY,
	event_type TEXT NOT NULL,
	actor_id TEXT,
	auth_kind TEXT,
	subject_type TEXT NOT NULL,
	subject_id TEXT,
	outcome TEXT NOT NULL DEFAULT 'success' CHECK (outcome IN ('success', 'failure')),
	payload_json TEXT NOT NULL DEFAULT '{}',
	occurred_at TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX audit_log_event_type_idx ON audit_log(event_type, occurred_at);
CREATE INDEX audit_log_actor_idx ON audit_log(actor_id, occurred_at);
CREATE INDEX audit_log_subject_idx ON audit_log(subject_type, subject_id, occurred_at);
