ALTER TABLE tickets ADD COLUMN start_date TEXT;
ALTER TABLE tickets ADD COLUMN due_date TEXT;

CREATE INDEX tickets_project_type_start_date_idx ON tickets(project_id, type, start_date);
CREATE INDEX tickets_project_type_due_date_idx ON tickets(project_id, type, due_date);
