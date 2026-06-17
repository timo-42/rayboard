ALTER TABLE webhooks ADD COLUMN event_types_json TEXT NOT NULL DEFAULT '[]';

CREATE TABLE outgoing_webhook_deliveries (
  id TEXT PRIMARY KEY,
  webhook_id TEXT REFERENCES webhooks(id) ON DELETE SET NULL,
  webhook_name TEXT NOT NULL,
  domain_event_id TEXT REFERENCES domain_events(id) ON DELETE SET NULL,
  idempotency_key TEXT UNIQUE,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  subject_type TEXT,
  subject_id TEXT,
  payload_json TEXT NOT NULL DEFAULT '{}',
  status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'sending', 'delivered', 'failed', 'canceled')),
  attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
  max_attempts INTEGER NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
  next_attempt_at TEXT,
  last_attempt_at TEXT,
  delivered_at TEXT,
  last_error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX outgoing_webhook_deliveries_webhook_idx
  ON outgoing_webhook_deliveries(webhook_id, status, created_at);

CREATE INDEX outgoing_webhook_deliveries_project_idx
  ON outgoing_webhook_deliveries(project_id, status, created_at);

CREATE INDEX outgoing_webhook_deliveries_next_attempt_idx
  ON outgoing_webhook_deliveries(status, next_attempt_at, created_at);
