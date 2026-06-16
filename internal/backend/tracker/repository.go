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
			reporter_id, assignee_id, parent_ticket_id, rank, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ticket.ID, ticket.ProjectID, ticket.Key, ticket.Title, nullableString(ticket.Description), ticket.Status, nullableString(ticket.Priority), nullableString(ticket.Type), nullableString(ticket.ReporterID), nullableString(ticket.AssigneeID), nullableString(ticket.ParentTicketID), nullableString(ticket.Rank), formatTime(ticket.CreatedAt), formatTime(ticket.UpdatedAt))
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
			reporter_id, assignee_id, parent_ticket_id, rank, created_at, updated_at, deleted_at
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
			reporter_id, assignee_id, parent_ticket_id, rank, created_at, updated_at, deleted_at
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
			rank = ?,
			updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, ticket.Title, nullableString(ticket.Description), ticket.Status, nullableString(ticket.Priority), nullableString(ticket.Type), nullableString(ticket.AssigneeID), nullableString(ticket.ParentTicketID), nullableString(ticket.Rank), formatTime(ticket.UpdatedAt), ticket.ID)
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
	var rank sql.NullString
	var createdAt string
	var updatedAt string
	var deletedAt sql.NullString

	if err := scanner.Scan(&ticket.ID, &ticket.ProjectID, &ticket.Key, &ticket.Title, &description, &ticket.Status, &priority, &ticketType, &reporterID, &assigneeID, &parentTicketID, &rank, &createdAt, &updatedAt, &deletedAt); err != nil {
		return Ticket{}, err
	}

	var err error
	ticket.Description = nullString(description)
	ticket.Priority = nullString(priority)
	ticket.Type = nullString(ticketType)
	ticket.ReporterID = nullString(reporterID)
	ticket.AssigneeID = nullString(assigneeID)
	ticket.ParentTicketID = nullString(parentTicketID)
	ticket.Rank = nullString(rank)
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
