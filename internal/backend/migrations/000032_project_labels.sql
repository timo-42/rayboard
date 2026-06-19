CREATE TABLE project_labels (
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	label TEXT NOT NULL CHECK (label != ''),
	description TEXT,
	color TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (project_id, label)
);

CREATE INDEX project_labels_project_id_idx ON project_labels(project_id);

INSERT INTO project_labels (project_id, label, created_at, updated_at)
SELECT
	tickets.project_id,
	labels.label,
	MIN(labels.created_at),
	MIN(labels.created_at)
FROM ticket_labels labels
JOIN tickets ON tickets.id = labels.ticket_id
WHERE tickets.deleted_at IS NULL
GROUP BY tickets.project_id, labels.label;
