CREATE TABLE notification_hooks (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  scope_type TEXT NOT NULL CHECK (scope_type IN ('global', 'project')),
  scope_key TEXT NOT NULL,
  project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
  actor_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  event_types_json TEXT NOT NULL DEFAULT '[]',
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  engine_type TEXT NOT NULL CHECK (engine_type IN ('lua', 'ai')),
  lua_script TEXT,
  ai_prompt TEXT,
  ai_provider_id TEXT REFERENCES openrouter_providers(id) ON DELETE RESTRICT,
  last_error TEXT,
  deleted_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX notification_hooks_scope_name_idx
  ON notification_hooks(scope_type, scope_key, name)
  WHERE deleted_at IS NULL;

CREATE INDEX notification_hooks_scope_idx
  ON notification_hooks(scope_type, scope_key, enabled);
