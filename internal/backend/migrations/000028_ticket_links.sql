CREATE TABLE ticket_links (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	source_ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	target_ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	link_type TEXT NOT NULL CHECK (link_type IN ('blocks', 'is_blocked_by', 'relates_to')),
	created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	deleted_at TEXT,
	CHECK (source_ticket_id <> target_ticket_id)
);

CREATE INDEX ticket_links_project_id_idx ON ticket_links(project_id);
CREATE INDEX ticket_links_source_ticket_id_idx ON ticket_links(source_ticket_id);
CREATE INDEX ticket_links_target_ticket_id_idx ON ticket_links(target_ticket_id);
CREATE UNIQUE INDEX ticket_links_unique_active_idx ON ticket_links(source_ticket_id, target_ticket_id, link_type) WHERE deleted_at IS NULL;
