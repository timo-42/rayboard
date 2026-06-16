CREATE TABLE notification_preferences (
	id TEXT PRIMARY KEY,
	scope_type TEXT NOT NULL CHECK (scope_type IN ('user', 'project')),
	scope_key TEXT NOT NULL,
	user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
	project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
	in_app_enabled INTEGER NOT NULL DEFAULT 1 CHECK (in_app_enabled IN (0, 1)),
	external_enabled INTEGER NOT NULL DEFAULT 1 CHECK (external_enabled IN (0, 1)),
	assignment_enabled INTEGER NOT NULL DEFAULT 1 CHECK (assignment_enabled IN (0, 1)),
	comment_enabled INTEGER NOT NULL DEFAULT 1 CHECK (comment_enabled IN (0, 1)),
	status_change_enabled INTEGER NOT NULL DEFAULT 1 CHECK (status_change_enabled IN (0, 1)),
	sprint_change_enabled INTEGER NOT NULL DEFAULT 1 CHECK (sprint_change_enabled IN (0, 1)),
	release_change_enabled INTEGER NOT NULL DEFAULT 1 CHECK (release_change_enabled IN (0, 1)),
	automation_failure_enabled INTEGER NOT NULL DEFAULT 1 CHECK (automation_failure_enabled IN (0, 1)),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	UNIQUE(scope_type, scope_key),
	CHECK (
		(scope_type = 'user' AND user_id IS NOT NULL AND project_id IS NULL) OR
		(scope_type = 'project' AND user_id IS NULL AND project_id IS NOT NULL)
	)
);

CREATE INDEX notification_preferences_user_id_idx ON notification_preferences(user_id);
CREATE INDEX notification_preferences_project_id_idx ON notification_preferences(project_id);
