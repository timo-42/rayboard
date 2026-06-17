CREATE TABLE IF NOT EXISTS ticket_create_pages (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  description TEXT,
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  target_type TEXT,
  target_status TEXT,
  field_layout_json TEXT NOT NULL DEFAULT '[]',
  defaults_json TEXT NOT NULL DEFAULT '{}',
  owner_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  updated_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  deleted_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ticket_create_pages_project_slug_idx
  ON ticket_create_pages(project_id, slug)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ticket_create_pages_project_enabled_idx
  ON ticket_create_pages(project_id, enabled, slug)
  WHERE deleted_at IS NULL;
