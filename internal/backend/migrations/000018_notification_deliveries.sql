CREATE TABLE notification_deliveries (
	id TEXT PRIMARY KEY,
	domain_event_id TEXT REFERENCES domain_events(id) ON DELETE SET NULL,
	idempotency_key TEXT UNIQUE,
	scope_type TEXT NOT NULL CHECK (scope_type IN ('global', 'project')),
	scope_key TEXT NOT NULL,
	project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
	policy_id TEXT REFERENCES notification_policies(id) ON DELETE SET NULL,
	policy_name TEXT,
	destination_id TEXT REFERENCES notification_destinations(id) ON DELETE SET NULL,
	destination_name TEXT,
	destination_service TEXT,
	event_type TEXT NOT NULL,
	subject_type TEXT,
	subject_id TEXT,
	message TEXT NOT NULL,
	payload_json TEXT NOT NULL DEFAULT '{}',
	status TEXT NOT NULL CHECK (status IN ('queued', 'sending', 'delivered', 'failed', 'canceled')),
	attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
	max_attempts INTEGER NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
	next_attempt_at TEXT,
	last_attempt_at TEXT,
	delivered_at TEXT,
	last_error TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX notification_deliveries_scope_idx ON notification_deliveries(scope_type, scope_key, status, created_at);
CREATE INDEX notification_deliveries_destination_idx ON notification_deliveries(destination_id, status);
CREATE INDEX notification_deliveries_policy_idx ON notification_deliveries(policy_id, status);
CREATE INDEX notification_deliveries_next_attempt_idx ON notification_deliveries(status, next_attempt_at);
