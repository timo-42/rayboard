CREATE TABLE ticket_watchers (
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (ticket_id, user_id)
);

CREATE INDEX ticket_watchers_user_id_idx ON ticket_watchers(user_id);
