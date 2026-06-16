CREATE TABLE project_statuses (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	slug TEXT NOT NULL,
	name TEXT NOT NULL,
	position INTEGER NOT NULL CHECK (position >= 0),
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(project_id, slug),
	UNIQUE(project_id, position)
);

CREATE INDEX project_statuses_project_id_idx ON project_statuses(project_id);

CREATE TABLE boards (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(project_id, name)
);

CREATE INDEX boards_project_id_idx ON boards(project_id);

CREATE TABLE board_columns (
	id TEXT PRIMARY KEY,
	board_id TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
	status_slug TEXT NOT NULL,
	name TEXT NOT NULL,
	position INTEGER NOT NULL CHECK (position >= 0),
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	UNIQUE(board_id, status_slug),
	UNIQUE(board_id, position)
);

CREATE INDEX board_columns_board_id_idx ON board_columns(board_id);

INSERT INTO project_statuses (id, project_id, slug, name, position)
SELECT 'status_' || id || '_todo', id, 'todo', 'Todo', 0
FROM projects
WHERE deleted_at IS NULL;

INSERT INTO project_statuses (id, project_id, slug, name, position)
SELECT 'status_' || id || '_in_progress', id, 'in_progress', 'In Progress', 1
FROM projects
WHERE deleted_at IS NULL;

INSERT INTO project_statuses (id, project_id, slug, name, position)
SELECT 'status_' || id || '_done', id, 'done', 'Done', 2
FROM projects
WHERE deleted_at IS NULL;

INSERT INTO boards (id, project_id, name, description, created_by)
SELECT 'board_' || id || '_default', id, 'Default Board', 'Default project workflow board', created_by
FROM projects
WHERE deleted_at IS NULL;

INSERT INTO board_columns (id, board_id, status_slug, name, position)
SELECT 'column_' || id || '_todo', 'board_' || id || '_default', 'todo', 'Todo', 0
FROM projects
WHERE deleted_at IS NULL;

INSERT INTO board_columns (id, board_id, status_slug, name, position)
SELECT 'column_' || id || '_in_progress', 'board_' || id || '_default', 'in_progress', 'In Progress', 1
FROM projects
WHERE deleted_at IS NULL;

INSERT INTO board_columns (id, board_id, status_slug, name, position)
SELECT 'column_' || id || '_done', 'board_' || id || '_default', 'done', 'Done', 2
FROM projects
WHERE deleted_at IS NULL;
