CREATE TABLE system_settings (
	key TEXT PRIMARY KEY,
	value_json TEXT NOT NULL,
	updated_by TEXT REFERENCES users(id) ON DELETE SET NULL,
	updated_at TEXT NOT NULL
);

INSERT INTO system_settings (key, value_json, updated_at)
VALUES (
	'global',
	'{"attachment_max_size_bytes":10485760,"attachment_allowed_content_types":[],"webhook_allowed_base_urls":[],"demo_warning_enabled":true,"backup_enabled":false,"system_health_note":""}',
	strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
);

