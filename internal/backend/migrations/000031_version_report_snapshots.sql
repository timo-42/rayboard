CREATE TABLE version_report_snapshots (
	version_id TEXT PRIMARY KEY REFERENCES project_versions(id) ON DELETE CASCADE,
	captured_at TEXT NOT NULL
);

CREATE TABLE version_report_tickets (
	version_id TEXT NOT NULL REFERENCES version_report_snapshots(version_id) ON DELETE CASCADE,
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	PRIMARY KEY (version_id, ticket_id)
);

CREATE INDEX version_report_tickets_ticket_id_idx ON version_report_tickets(ticket_id);

INSERT INTO version_report_snapshots (version_id, captured_at)
SELECT
	id,
	COALESCE(
		CASE
			WHEN release_date IS NOT NULL AND release_date != '' THEN release_date || 'T00:00:00Z'
			ELSE NULL
		END,
		updated_at,
		created_at
	)
FROM project_versions
WHERE status = 'released';

INSERT INTO version_report_tickets (version_id, ticket_id, position)
SELECT version_id, ticket_id, position
FROM (
	SELECT
		tickets.version_id,
		tickets.id AS ticket_id,
		ROW_NUMBER() OVER (
			PARTITION BY tickets.version_id
			ORDER BY tickets.status ASC, tickets.created_at DESC, tickets.key DESC
		) - 1 AS position
	FROM tickets
	JOIN version_report_snapshots snapshot ON snapshot.version_id = tickets.version_id
	WHERE tickets.deleted_at IS NULL
);
