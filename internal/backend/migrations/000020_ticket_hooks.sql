CREATE TABLE ticket_hooks (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  event TEXT NOT NULL CHECK (event IN ('ticket_create', 'ticket_update')),
  phase TEXT NOT NULL CHECK (phase IN ('before', 'after')),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  position INTEGER NOT NULL DEFAULT 100,
  engine_type TEXT NOT NULL CHECK (engine_type IN ('lua', 'ai')),
  lua_script TEXT,
  ai_prompt TEXT,
  ai_provider_id TEXT REFERENCES openrouter_providers(id) ON DELETE RESTRICT,
  last_error TEXT,
  deleted_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX ticket_hooks_project_event_phase_name_idx
  ON ticket_hooks(project_id, event, phase, name)
  WHERE deleted_at IS NULL;

CREATE INDEX ticket_hooks_project_event_phase_idx
  ON ticket_hooks(project_id, event, phase, enabled, position)
  WHERE deleted_at IS NULL;
