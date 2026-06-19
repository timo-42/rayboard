package tracker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type sqlRunner interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type rowScanner interface {
	Scan(...any) error
}

func (r *Repository) CreateProject(ctx context.Context, project Project) error {
	runner, err := r.runner()
	if err != nil {
		return err
	}
	return r.insertProject(ctx, runner, project)
}

func (r *Repository) GetProject(ctx context.Context, projectID string) (Project, error) {
	runner, err := r.runner()
	if err != nil {
		return Project{}, err
	}
	return r.getProject(ctx, runner, projectID)
}

func (r *Repository) GetProjectByKey(ctx context.Context, key string) (Project, error) {
	runner, err := r.runner()
	if err != nil {
		return Project{}, err
	}
	return r.getProjectByKey(ctx, runner, key)
}

func (r *Repository) ListProjects(ctx context.Context, input ListProjectsInput) ([]Project, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.listProjects(ctx, runner, input)
}

func (r *Repository) GetTicket(ctx context.Context, ticketID string) (Ticket, error) {
	runner, err := r.runner()
	if err != nil {
		return Ticket{}, err
	}
	return r.getTicket(ctx, runner, ticketID)
}

func (r *Repository) ListTickets(ctx context.Context, input ListTicketsInput) ([]Ticket, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.listTickets(ctx, runner, input)
}

func (r *Repository) ListTicketActivity(ctx context.Context, ticketID string) ([]TicketActivity, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.listTicketActivity(ctx, runner, ticketID)
}

func (r *Repository) ListTicketWatchers(ctx context.Context, ticketID string) ([]TicketWatcher, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.listTicketWatchers(ctx, runner, ticketID)
}

func (r *Repository) runner() (sqlRunner, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("tracker: nil database")
	}
	return r.db, nil
}

func (r *Repository) insertProject(ctx context.Context, q sqlRunner, project Project) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO projects (
			id, key, name, description, lead_user_id, created_by, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, project.ID, project.Key, project.Name, nullableString(project.Description), nullableString(project.LeadUserID), nullableString(project.CreatedBy), formatTime(project.CreatedAt), formatTime(project.UpdatedAt))
	if err != nil {
		if isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE) || isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY) {
			return conflict("project", "key", project.Key)
		}
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (r *Repository) bindProjectOwner(ctx context.Context, q sqlRunner, projectID string, userID string, createdAt time.Time) error {
	if strings.TrimSpace(userID) == "" {
		return nil
	}

	bindingID := "binding_" + projectID + "_owner_" + userID
	_, err := q.ExecContext(ctx, `
		INSERT INTO role_bindings (
			id, role_id, subject_type, subject_id, resource_type, resource_id, created_at
		)
		VALUES (?, ?, 'user', ?, 'project', ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, bindingID, string(authz.RoleProjectOwner), userID, projectID, formatTime(createdAt))
	if err != nil {
		return fmt.Errorf("bind project owner: %w", err)
	}
	return nil
}

func (r *Repository) getProject(ctx context.Context, q sqlRunner, projectID string) (Project, error) {
	project, err := scanProject(q.QueryRowContext(ctx, `
		SELECT id, key, name, description, lead_user_id, created_by, created_at, updated_at, archived_at, deleted_at
		FROM projects
		WHERE id = ? AND deleted_at IS NULL
	`, projectID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, notFound("project", projectID)
		}
		return Project{}, fmt.Errorf("get project: %w", err)
	}
	return project, nil
}

func (r *Repository) getProjectByKey(ctx context.Context, q sqlRunner, key string) (Project, error) {
	project, err := scanProject(q.QueryRowContext(ctx, `
		SELECT id, key, name, description, lead_user_id, created_by, created_at, updated_at, archived_at, deleted_at
		FROM projects
		WHERE key = ? AND deleted_at IS NULL
	`, key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, notFound("project", key)
		}
		return Project{}, fmt.Errorf("get project by key: %w", err)
	}
	return project, nil
}

func (r *Repository) listProjects(ctx context.Context, q sqlRunner, input ListProjectsInput) ([]Project, error) {
	limit, offset := normalizeListWindow(input.Limit, input.Offset)
	query := `
		SELECT id, key, name, description, lead_user_id, created_by, created_at, updated_at, archived_at, deleted_at
		FROM projects
		WHERE deleted_at IS NULL`
	if !input.IncludeArchived {
		query += " AND archived_at IS NULL"
	}
	query += " ORDER BY key ASC LIMIT ? OFFSET ?"

	rows, err := q.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (r *Repository) nextTicketKey(ctx context.Context, q sqlRunner, project Project) (string, error) {
	prefix := project.Key + "-"
	var maxSuffix sql.NullInt64
	if err := q.QueryRowContext(ctx, `
		SELECT MAX(CAST(SUBSTR(key, ?) AS INTEGER))
		FROM tickets
		WHERE project_id = ? AND key GLOB ?
	`, len(prefix)+1, project.ID, prefix+"[0-9]*").Scan(&maxSuffix); err != nil {
		return "", fmt.Errorf("find next ticket key: %w", err)
	}
	next := int64(1)
	if maxSuffix.Valid {
		next = maxSuffix.Int64 + 1
	}
	return prefix + strconv.FormatInt(next, 10), nil
}

func (r *Repository) insertTicket(ctx context.Context, q sqlRunner, ticket Ticket) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO tickets (
			id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ticket.ID, ticket.ProjectID, ticket.Key, ticket.Title, nullableString(ticket.Description), ticket.Status, nullableString(ticket.Priority), nullableString(ticket.Type), nullableString(ticket.ReporterID), nullableString(ticket.AssigneeID), nullableString(ticket.ParentTicketID), nullableString(ticket.SprintID), nullableString(ticket.ComponentID), nullableString(ticket.VersionID), nullableString(ticket.Rank), nullableString(ticket.StartDate), nullableString(ticket.DueDate), formatTime(ticket.CreatedAt), formatTime(ticket.UpdatedAt))
	if err != nil {
		if isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE) || isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY) {
			return conflict("ticket", "key", ticket.Key)
		}
		return fmt.Errorf("insert ticket: %w", err)
	}
	return nil
}

func (r *Repository) getTicket(ctx context.Context, q sqlRunner, ticketID string) (Ticket, error) {
	ticket, err := scanTicket(q.QueryRowContext(ctx, `
		SELECT id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, created_at, updated_at, deleted_at
		FROM tickets
		WHERE id = ? AND deleted_at IS NULL
	`, ticketID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Ticket{}, notFound("ticket", ticketID)
		}
		return Ticket{}, fmt.Errorf("get ticket: %w", err)
	}
	return ticket, nil
}

func (r *Repository) listTickets(ctx context.Context, q sqlRunner, input ListTicketsInput) ([]Ticket, error) {
	limit, offset := normalizeListWindow(input.Limit, input.Offset)
	query := `
		SELECT id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, created_at, updated_at, deleted_at
		FROM tickets
		WHERE project_id = ? AND deleted_at IS NULL`
	args := []any{input.ProjectID}
	if input.Status != "" {
		query += " AND status = ?"
		args = append(args, input.Status)
	}
	if input.AssigneeID != "" {
		query += " AND assignee_id = ?"
		args = append(args, input.AssigneeID)
	}
	if input.SprintID != "" {
		query += " AND sprint_id = ?"
		args = append(args, input.SprintID)
	}
	if input.ComponentID != "" {
		query += " AND component_id = ?"
		args = append(args, input.ComponentID)
	}
	if input.VersionID != "" {
		query += " AND version_id = ?"
		args = append(args, input.VersionID)
	}
	if input.Label != "" {
		query += ` AND id IN (
			SELECT ticket_id
			FROM ticket_labels
			WHERE label = ?
		)`
		args = append(args, input.Label)
	}
	query += " ORDER BY created_at DESC, key DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}
	return tickets, nil
}

func (r *Repository) updateTicket(ctx context.Context, q sqlRunner, ticket Ticket) error {
	result, err := q.ExecContext(ctx, `
		UPDATE tickets
		SET title = ?,
			description = ?,
			status = ?,
			priority = ?,
			type = ?,
			assignee_id = ?,
			parent_ticket_id = ?,
			sprint_id = ?,
			component_id = ?,
			version_id = ?,
			rank = ?,
			start_date = ?,
			due_date = ?,
			updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, ticket.Title, nullableString(ticket.Description), ticket.Status, nullableString(ticket.Priority), nullableString(ticket.Type), nullableString(ticket.AssigneeID), nullableString(ticket.ParentTicketID), nullableString(ticket.SprintID), nullableString(ticket.ComponentID), nullableString(ticket.VersionID), nullableString(ticket.Rank), nullableString(ticket.StartDate), nullableString(ticket.DueDate), formatTime(ticket.UpdatedAt), ticket.ID)
	if err != nil {
		return fmt.Errorf("update ticket: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated ticket rows: %w", err)
	}
	if updated == 0 {
		return notFound("ticket", ticket.ID)
	}
	return nil
}

func (r *Repository) deleteTicket(ctx context.Context, q sqlRunner, ticketID string, deletedAt time.Time) error {
	result, err := q.ExecContext(ctx, `
		UPDATE tickets
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(deletedAt), formatTime(deletedAt), ticketID)
	if err != nil {
		return fmt.Errorf("delete ticket: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted ticket rows: %w", err)
	}
	if deleted == 0 {
		return notFound("ticket", ticketID)
	}
	return nil
}

func (r *Repository) insertTicketActivity(ctx context.Context, q sqlRunner, activity TicketActivity) error {
	data := activity.Data
	if data == nil {
		data = map[string]any{}
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode ticket activity data: %w", err)
	}

	_, err = q.ExecContext(ctx, `
		INSERT INTO ticket_activity (id, ticket_id, actor_id, activity_type, data_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, activity.ID, activity.TicketID, nullableString(activity.ActorID), activity.ActivityType, string(encoded), formatTime(activity.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert ticket activity: %w", err)
	}
	return nil
}

func (r *Repository) listTicketActivity(ctx context.Context, q sqlRunner, ticketID string) ([]TicketActivity, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, ticket_id, actor_id, activity_type, data_json, created_at
		FROM ticket_activity
		WHERE ticket_id = ?
		ORDER BY created_at ASC, id ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list ticket activity: %w", err)
	}
	defer rows.Close()

	var activities []TicketActivity
	for rows.Next() {
		activity, err := scanTicketActivity(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket activity: %w", err)
		}
		activities = append(activities, activity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket activity: %w", err)
	}
	return activities, nil
}

func (r *Repository) listTicketWatchers(ctx context.Context, q sqlRunner, ticketID string) ([]TicketWatcher, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT tw.ticket_id, tw.user_id, u.username, u.display_name, tw.created_at
		FROM ticket_watchers tw
		JOIN users u ON u.id = tw.user_id
		WHERE tw.ticket_id = ? AND u.deleted_at IS NULL
		ORDER BY u.display_name ASC, u.username ASC, tw.user_id ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list ticket watchers: %w", err)
	}
	defer rows.Close()

	var watchers []TicketWatcher
	for rows.Next() {
		watcher, err := scanTicketWatcher(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket watcher: %w", err)
		}
		watchers = append(watchers, watcher)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket watchers: %w", err)
	}
	return watchers, nil
}

func (r *Repository) setTicketWatcher(ctx context.Context, q sqlRunner, ticketID string, userID string, createdAt time.Time) (bool, error) {
	result, err := q.ExecContext(ctx, `
		INSERT INTO ticket_watchers (ticket_id, user_id, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(ticket_id, user_id) DO NOTHING
	`, ticketID, userID, formatTime(createdAt))
	if err != nil {
		return false, fmt.Errorf("set ticket watcher: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read ticket watcher insert rows: %w", err)
	}
	return affected > 0, nil
}

func (r *Repository) deleteTicketWatcher(ctx context.Context, q sqlRunner, ticketID string, userID string) (bool, error) {
	result, err := q.ExecContext(ctx, `
		DELETE FROM ticket_watchers
		WHERE ticket_id = ? AND user_id = ?
	`, ticketID, userID)
	if err != nil {
		return false, fmt.Errorf("delete ticket watcher: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read ticket watcher delete rows: %w", err)
	}
	return affected > 0, nil
}

func (r *Repository) attachWatcherStatus(ctx context.Context, q sqlRunner, tickets []Ticket, userID string) ([]Ticket, error) {
	if len(tickets) == 0 {
		return tickets, nil
	}
	ids := make([]string, 0, len(tickets))
	indexByID := make(map[string]int, len(tickets))
	for index := range tickets {
		ids = append(ids, tickets[index].ID)
		indexByID[tickets[index].ID] = index
	}
	args := make([]any, 0, len(ids)*2+1)
	for _, id := range ids {
		args = append(args, id)
	}
	query := `
		SELECT ticket_id, COUNT(*)
		FROM ticket_watchers tw
		JOIN users u ON u.id = tw.user_id
		WHERE ticket_id IN (` + placeholders(len(ids)) + `) AND u.deleted_at IS NULL
		GROUP BY ticket_id
	`
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("count ticket watchers: %w", err)
	}
	for rows.Next() {
		var ticketID string
		var count int
		if err := rows.Scan(&ticketID, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan ticket watcher count: %w", err)
		}
		if index, ok := indexByID[ticketID]; ok {
			tickets[index].WatcherCount = count
		}
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close ticket watcher counts: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket watcher counts: %w", err)
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return tickets, nil
	}
	args = args[:0]
	for _, id := range ids {
		args = append(args, id)
	}
	args = append(args, userID)
	rows, err = q.QueryContext(ctx, `
		SELECT ticket_id
		FROM ticket_watchers
		WHERE ticket_id IN (`+placeholders(len(ids))+`) AND user_id = ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list current user ticket watches: %w", err)
	}
	for rows.Next() {
		var ticketID string
		if err := rows.Scan(&ticketID); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan current user ticket watch: %w", err)
		}
		if index, ok := indexByID[ticketID]; ok {
			tickets[index].Watching = true
		}
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close current user ticket watches: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current user ticket watches: %w", err)
	}
	return tickets, nil
}

func (r *Repository) listTicketLinks(ctx context.Context, ticketID string) ([]TicketLink, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	rows, err := runner.QueryContext(ctx, `
		SELECT
			links.id, links.project_id, links.link_type, links.created_by, links.created_at,
			source.id, source.project_id, source.key, source.title, source.description, source.status, source.priority, source.type,
			source.reporter_id, source.assignee_id, source.parent_ticket_id, source.sprint_id, source.component_id, source.version_id, source.rank, source.start_date, source.due_date, source.created_at, source.updated_at, source.deleted_at,
			target.id, target.project_id, target.key, target.title, target.description, target.status, target.priority, target.type,
			target.reporter_id, target.assignee_id, target.parent_ticket_id, target.sprint_id, target.component_id, target.version_id, target.rank, target.start_date, target.due_date, target.created_at, target.updated_at, target.deleted_at
		FROM ticket_links AS links
		JOIN tickets AS source ON source.id = links.source_ticket_id
		JOIN tickets AS target ON target.id = links.target_ticket_id
		WHERE links.source_ticket_id = ? AND links.deleted_at IS NULL AND source.deleted_at IS NULL AND target.deleted_at IS NULL
		ORDER BY links.created_at ASC, links.id ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list ticket links: %w", err)
	}
	defer rows.Close()

	var links []TicketLink
	for rows.Next() {
		link, err := scanTicketLink(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket link: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket links: %w", err)
	}
	return links, nil
}

func (r *Repository) getTicketLink(ctx context.Context, linkID string) (TicketLink, error) {
	runner, err := r.runner()
	if err != nil {
		return TicketLink{}, err
	}
	link, err := scanTicketLink(runner.QueryRowContext(ctx, `
		SELECT
			links.id, links.project_id, links.link_type, links.created_by, links.created_at,
			source.id, source.project_id, source.key, source.title, source.description, source.status, source.priority, source.type,
			source.reporter_id, source.assignee_id, source.parent_ticket_id, source.sprint_id, source.component_id, source.version_id, source.rank, source.start_date, source.due_date, source.created_at, source.updated_at, source.deleted_at,
			target.id, target.project_id, target.key, target.title, target.description, target.status, target.priority, target.type,
			target.reporter_id, target.assignee_id, target.parent_ticket_id, target.sprint_id, target.component_id, target.version_id, target.rank, target.start_date, target.due_date, target.created_at, target.updated_at, target.deleted_at
		FROM ticket_links AS links
		JOIN tickets AS source ON source.id = links.source_ticket_id
		JOIN tickets AS target ON target.id = links.target_ticket_id
		WHERE links.id = ? AND links.deleted_at IS NULL AND source.deleted_at IS NULL AND target.deleted_at IS NULL
	`, linkID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TicketLink{}, notFound("ticket_link", linkID)
		}
		return TicketLink{}, fmt.Errorf("get ticket link: %w", err)
	}
	return link, nil
}

func (r *Repository) insertTicketLink(ctx context.Context, q sqlRunner, link TicketLink) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO ticket_links (id, project_id, source_ticket_id, target_ticket_id, link_type, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, link.ID, link.ProjectID, link.Source.ID, link.Target.ID, link.LinkType, nullableString(link.CreatedBy), formatTime(link.CreatedAt))
	if err != nil {
		if isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE) {
			return conflict("ticket_link", "link_type", link.LinkType)
		}
		if isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_CHECK) || isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY) {
			return validationFailed(map[string]string{"target_ticket_id": "Invalid ticket link"})
		}
		return fmt.Errorf("insert ticket link: %w", err)
	}
	return nil
}

func (r *Repository) deleteTicketLink(ctx context.Context, q sqlRunner, linkID string, deletedAt time.Time) error {
	result, err := q.ExecContext(ctx, `
		UPDATE ticket_links
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(deletedAt), linkID)
	if err != nil {
		return fmt.Errorf("delete ticket link: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted ticket link rows: %w", err)
	}
	if deleted == 0 {
		return notFound("ticket_link", linkID)
	}
	return nil
}

func (r *Repository) userExists(ctx context.Context, q sqlRunner, userID string) (bool, error) {
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists == 1, nil
}

func (r *Repository) ticketExistsInProject(ctx context.Context, q sqlRunner, ticketID string, projectID string) (bool, error) {
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tickets
		WHERE id = ? AND project_id = ? AND deleted_at IS NULL
	`, ticketID, projectID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check parent ticket exists: %w", err)
	}
	return exists == 1, nil
}

func (r *Repository) epicExistsInProject(ctx context.Context, q sqlRunner, ticketID string, projectID string) (bool, error) {
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tickets
		WHERE id = ? AND project_id = ? AND type = 'epic' AND deleted_at IS NULL
	`, ticketID, projectID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check epic exists: %w", err)
	}
	return exists == 1, nil
}

func scanProject(scanner rowScanner) (Project, error) {
	var project Project
	var description sql.NullString
	var leadUserID sql.NullString
	var createdBy sql.NullString
	var createdAt string
	var updatedAt string
	var archivedAt sql.NullString
	var deletedAt sql.NullString

	if err := scanner.Scan(&project.ID, &project.Key, &project.Name, &description, &leadUserID, &createdBy, &createdAt, &updatedAt, &archivedAt, &deletedAt); err != nil {
		return Project{}, err
	}

	var err error
	project.Description = nullString(description)
	project.LeadUserID = nullString(leadUserID)
	project.CreatedBy = nullString(createdBy)
	project.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Project{}, fmt.Errorf("parse project created_at: %w", err)
	}
	project.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Project{}, fmt.Errorf("parse project updated_at: %w", err)
	}
	project.ArchivedAt = parseNullableTime(archivedAt)
	project.DeletedAt = parseNullableTime(deletedAt)
	return project, nil
}

func scanTicket(scanner rowScanner) (Ticket, error) {
	var ticket Ticket
	var description sql.NullString
	var priority sql.NullString
	var ticketType sql.NullString
	var reporterID sql.NullString
	var assigneeID sql.NullString
	var parentTicketID sql.NullString
	var sprintID sql.NullString
	var componentID sql.NullString
	var versionID sql.NullString
	var rank sql.NullString
	var startDate sql.NullString
	var dueDate sql.NullString
	var createdAt string
	var updatedAt string
	var deletedAt sql.NullString

	if err := scanner.Scan(&ticket.ID, &ticket.ProjectID, &ticket.Key, &ticket.Title, &description, &ticket.Status, &priority, &ticketType, &reporterID, &assigneeID, &parentTicketID, &sprintID, &componentID, &versionID, &rank, &startDate, &dueDate, &createdAt, &updatedAt, &deletedAt); err != nil {
		return Ticket{}, err
	}

	var err error
	ticket.Description = nullString(description)
	ticket.Priority = nullString(priority)
	ticket.Type = nullString(ticketType)
	ticket.ReporterID = nullString(reporterID)
	ticket.AssigneeID = nullString(assigneeID)
	ticket.ParentTicketID = nullString(parentTicketID)
	ticket.SprintID = nullString(sprintID)
	ticket.ComponentID = nullString(componentID)
	ticket.VersionID = nullString(versionID)
	ticket.Rank = nullString(rank)
	ticket.StartDate = nullString(startDate)
	ticket.DueDate = nullString(dueDate)
	ticket.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Ticket{}, fmt.Errorf("parse ticket created_at: %w", err)
	}
	ticket.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Ticket{}, fmt.Errorf("parse ticket updated_at: %w", err)
	}
	ticket.DeletedAt = parseNullableTime(deletedAt)
	return ticket, nil
}

func scanTicketActivity(scanner rowScanner) (TicketActivity, error) {
	var activity TicketActivity
	var actorID sql.NullString
	var dataJSON string
	var createdAt string

	if err := scanner.Scan(&activity.ID, &activity.TicketID, &actorID, &activity.ActivityType, &dataJSON, &createdAt); err != nil {
		return TicketActivity{}, err
	}

	activity.ActorID = nullString(actorID)
	if dataJSON == "" {
		dataJSON = "{}"
	}
	if err := json.Unmarshal([]byte(dataJSON), &activity.Data); err != nil {
		return TicketActivity{}, fmt.Errorf("decode ticket activity data: %w", err)
	}
	var err error
	activity.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return TicketActivity{}, fmt.Errorf("parse ticket activity created_at: %w", err)
	}
	return activity, nil
}

func scanTicketWatcher(scanner rowScanner) (TicketWatcher, error) {
	var watcher TicketWatcher
	var createdAt string
	if err := scanner.Scan(&watcher.TicketID, &watcher.UserID, &watcher.Username, &watcher.DisplayName, &createdAt); err != nil {
		return TicketWatcher{}, err
	}
	var err error
	watcher.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return TicketWatcher{}, fmt.Errorf("parse ticket watcher created_at: %w", err)
	}
	return watcher, nil
}

func scanTicketLink(scanner rowScanner) (TicketLink, error) {
	var link TicketLink
	var createdBy sql.NullString
	var createdAt string
	var sourceDescription sql.NullString
	var sourcePriority sql.NullString
	var sourceType sql.NullString
	var sourceReporterID sql.NullString
	var sourceAssigneeID sql.NullString
	var sourceParentTicketID sql.NullString
	var sourceSprintID sql.NullString
	var sourceComponentID sql.NullString
	var sourceVersionID sql.NullString
	var sourceRank sql.NullString
	var sourceStartDate sql.NullString
	var sourceDueDate sql.NullString
	var sourceCreatedAt string
	var sourceUpdatedAt string
	var sourceDeletedAt sql.NullString
	var targetDescription sql.NullString
	var targetPriority sql.NullString
	var targetType sql.NullString
	var targetReporterID sql.NullString
	var targetAssigneeID sql.NullString
	var targetParentTicketID sql.NullString
	var targetSprintID sql.NullString
	var targetComponentID sql.NullString
	var targetVersionID sql.NullString
	var targetRank sql.NullString
	var targetStartDate sql.NullString
	var targetDueDate sql.NullString
	var targetCreatedAt string
	var targetUpdatedAt string
	var targetDeletedAt sql.NullString
	if err := scanner.Scan(
		&link.ID,
		&link.ProjectID,
		&link.LinkType,
		&createdBy,
		&createdAt,
		&link.Source.ID,
		&link.Source.ProjectID,
		&link.Source.Key,
		&link.Source.Title,
		&sourceDescription,
		&link.Source.Status,
		&sourcePriority,
		&sourceType,
		&sourceReporterID,
		&sourceAssigneeID,
		&sourceParentTicketID,
		&sourceSprintID,
		&sourceComponentID,
		&sourceVersionID,
		&sourceRank,
		&sourceStartDate,
		&sourceDueDate,
		&sourceCreatedAt,
		&sourceUpdatedAt,
		&sourceDeletedAt,
		&link.Target.ID,
		&link.Target.ProjectID,
		&link.Target.Key,
		&link.Target.Title,
		&targetDescription,
		&link.Target.Status,
		&targetPriority,
		&targetType,
		&targetReporterID,
		&targetAssigneeID,
		&targetParentTicketID,
		&targetSprintID,
		&targetComponentID,
		&targetVersionID,
		&targetRank,
		&targetStartDate,
		&targetDueDate,
		&targetCreatedAt,
		&targetUpdatedAt,
		&targetDeletedAt,
	); err != nil {
		return TicketLink{}, err
	}
	link.CreatedBy = nullString(createdBy)
	link.Source.Description = nullString(sourceDescription)
	link.Source.Priority = nullString(sourcePriority)
	link.Source.Type = nullString(sourceType)
	link.Source.ReporterID = nullString(sourceReporterID)
	link.Source.AssigneeID = nullString(sourceAssigneeID)
	link.Source.ParentTicketID = nullString(sourceParentTicketID)
	link.Source.SprintID = nullString(sourceSprintID)
	link.Source.ComponentID = nullString(sourceComponentID)
	link.Source.VersionID = nullString(sourceVersionID)
	link.Source.Rank = nullString(sourceRank)
	link.Source.StartDate = nullString(sourceStartDate)
	link.Source.DueDate = nullString(sourceDueDate)
	link.Source.DeletedAt = parseNullableTime(sourceDeletedAt)
	link.Target.Description = nullString(targetDescription)
	link.Target.Priority = nullString(targetPriority)
	link.Target.Type = nullString(targetType)
	link.Target.ReporterID = nullString(targetReporterID)
	link.Target.AssigneeID = nullString(targetAssigneeID)
	link.Target.ParentTicketID = nullString(targetParentTicketID)
	link.Target.SprintID = nullString(targetSprintID)
	link.Target.ComponentID = nullString(targetComponentID)
	link.Target.VersionID = nullString(targetVersionID)
	link.Target.Rank = nullString(targetRank)
	link.Target.StartDate = nullString(targetStartDate)
	link.Target.DueDate = nullString(targetDueDate)
	link.Target.DeletedAt = parseNullableTime(targetDeletedAt)
	var err error
	link.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return TicketLink{}, fmt.Errorf("parse ticket link created_at: %w", err)
	}
	link.Source.CreatedAt, err = parseTime(sourceCreatedAt)
	if err != nil {
		return TicketLink{}, fmt.Errorf("parse source ticket created_at: %w", err)
	}
	link.Source.UpdatedAt, err = parseTime(sourceUpdatedAt)
	if err != nil {
		return TicketLink{}, fmt.Errorf("parse source ticket updated_at: %w", err)
	}
	link.Target.CreatedAt, err = parseTime(targetCreatedAt)
	if err != nil {
		return TicketLink{}, fmt.Errorf("parse target ticket created_at: %w", err)
	}
	link.Target.UpdatedAt, err = parseTime(targetUpdatedAt)
	if err != nil {
		return TicketLink{}, fmt.Errorf("parse target ticket updated_at: %w", err)
	}
	return link, nil
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func normalizeListWindow(limit int, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	parsed, err := parseTime(value.String)
	if err != nil {
		return nil
	}
	return &parsed
}

func isSQLiteCode(err error, code int) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}
	return sqliteErr.Code() == code
}
