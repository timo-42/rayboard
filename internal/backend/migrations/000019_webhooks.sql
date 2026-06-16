CREATE TABLE webhooks (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  direction TEXT NOT NULL CHECK (direction IN ('incoming', 'outgoing')),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  actor_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  engine_type TEXT NOT NULL CHECK (engine_type IN ('lua', 'ai')),
  lua_script TEXT,
  ai_prompt TEXT,
  ai_provider_id TEXT REFERENCES openrouter_providers(id) ON DELETE RESTRICT,
  token_hash TEXT,
  token_rotated_at TEXT,
  last_error TEXT,
  deleted_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX webhooks_project_direction_name_idx
  ON webhooks(project_id, direction, name)
  WHERE deleted_at IS NULL;

CREATE INDEX webhooks_project_idx ON webhooks(project_id, direction, enabled);
CREATE INDEX webhooks_token_idx ON webhooks(token_hash) WHERE token_hash IS NOT NULL AND deleted_at IS NULL;
