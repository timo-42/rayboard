CREATE TABLE ticket_labels (
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	label TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (ticket_id, label)
);

CREATE INDEX ticket_labels_label_idx ON ticket_labels(label);
