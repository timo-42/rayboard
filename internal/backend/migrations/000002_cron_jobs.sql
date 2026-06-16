CREATE TABLE cron_jobs (
	id TEXT PRIMARY KEY,
	owner_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	schedule TEXT NOT NULL,
	timezone TEXT NOT NULL DEFAULT 'UTC',
	enabled INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
	engine TEXT NOT NULL DEFAULT 'lua' CHECK (engine IN ('lua', 'ai')),
	lua_source TEXT NOT NULL DEFAULT '',
	ai_prompt TEXT NOT NULL DEFAULT '',
	last_run_status TEXT,
	last_run_at TEXT,
	next_run_at TEXT,
	last_error TEXT,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX cron_jobs_owner_user_id_idx ON cron_jobs(owner_user_id);
CREATE INDEX cron_jobs_project_id_idx ON cron_jobs(project_id);
CREATE INDEX cron_jobs_enabled_idx ON cron_jobs(enabled);
