ALTER TABLE saved_views ADD COLUMN display_mode TEXT NOT NULL DEFAULT 'list' CHECK (display_mode IN ('list', 'board', 'backlog'));
ALTER TABLE saved_views ADD COLUMN group_by TEXT;
ALTER TABLE saved_views ADD COLUMN is_pinned INTEGER NOT NULL DEFAULT 0 CHECK (is_pinned IN (0, 1));
CREATE INDEX saved_views_project_pinned_idx ON saved_views(project_id, is_pinned);
