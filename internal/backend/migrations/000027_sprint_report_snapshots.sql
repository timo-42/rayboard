CREATE TABLE sprint_report_snapshots (
	sprint_id TEXT PRIMARY KEY REFERENCES sprints(id) ON DELETE CASCADE,
	captured_at TEXT NOT NULL
);

CREATE TABLE sprint_report_tickets (
	sprint_id TEXT NOT NULL REFERENCES sprint_report_snapshots(sprint_id) ON DELETE CASCADE,
	ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	PRIMARY KEY (sprint_id, ticket_id)
);

CREATE INDEX sprint_report_tickets_ticket_id_idx ON sprint_report_tickets(ticket_id);

INSERT INTO sprint_report_snapshots (sprint_id, captured_at)
SELECT id, COALESCE(completed_at, updated_at, created_at)
FROM sprints
WHERE state = 'completed';

INSERT INTO sprint_report_tickets (sprint_id, ticket_id, position)
SELECT sprint_id, ticket_id, position
FROM (
	SELECT
		tickets.sprint_id,
		tickets.id AS ticket_id,
		ROW_NUMBER() OVER (
			PARTITION BY tickets.sprint_id
			ORDER BY tickets.status ASC, tickets.created_at DESC, tickets.key DESC
		) - 1 AS position
	FROM tickets
	JOIN sprint_report_snapshots snapshot ON snapshot.sprint_id = tickets.sprint_id
	WHERE tickets.deleted_at IS NULL
);
