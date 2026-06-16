CREATE TABLE openrouter_providers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	default_model TEXT NOT NULL,
	api_key_secret TEXT NOT NULL,
	allowed_models_json TEXT NOT NULL DEFAULT '[]',
	default_timeout_seconds INTEGER NOT NULL DEFAULT 30 CHECK (default_timeout_seconds > 0),
	max_output_tokens INTEGER NOT NULL DEFAULT 2048 CHECK (max_output_tokens > 0),
	enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	deleted_at TEXT
);

CREATE INDEX openrouter_providers_enabled_idx ON openrouter_providers(enabled, deleted_at);
