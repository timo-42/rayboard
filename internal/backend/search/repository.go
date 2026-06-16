package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
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

type savedViewListQuery struct {
	OwnerUserID       string
	ProjectID         string
	VisibleProjectIDs []string
	PinnedOnly        bool
	Limit             int
	Offset            int
}

type ticketSearchQuery struct {
	ProjectIDs []string
	Filter     compiledFilter
	FTSQuery   string
	Sort       []SortSpec
	Limit      int
	Offset     int
}

func (r *Repository) CreateSavedView(ctx context.Context, view SavedView) error {
	runner, err := r.runner()
	if err != nil {
		return err
	}
	return r.insertSavedView(ctx, runner, view)
}

func (r *Repository) GetSavedView(ctx context.Context, id string) (SavedView, error) {
	runner, err := r.runner()
	if err != nil {
		return SavedView{}, err
	}
	return r.getSavedView(ctx, runner, id)
}

func (r *Repository) UpdateSavedView(ctx context.Context, view SavedView) error {
	runner, err := r.runner()
	if err != nil {
		return err
	}
	return r.updateSavedView(ctx, runner, view)
}

func (r *Repository) DeleteSavedView(ctx context.Context, id string) error {
	runner, err := r.runner()
	if err != nil {
		return err
	}
	return r.deleteSavedView(ctx, runner, id)
}

func (r *Repository) ListSavedViews(ctx context.Context, input savedViewListQuery) ([]SavedView, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.listSavedViews(ctx, runner, input)
}

func (r *Repository) RefreshFTSIndex(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("search: nil database")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin search index refresh: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM ticket_fts`); err != nil {
		return fmt.Errorf("clear ticket fts: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_fts (ticket_id, title, description)
		SELECT id, title, COALESCE(description, '')
		FROM tickets
		WHERE deleted_at IS NULL
	`); err != nil {
		return fmt.Errorf("populate ticket fts: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM comment_fts`); err != nil {
		return fmt.Errorf("clear comment fts: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO comment_fts (comment_id, body)
		SELECT c.id, c.body
		FROM ticket_comments c
		JOIN tickets t ON t.id = c.ticket_id
		WHERE c.deleted_at IS NULL
		  AND t.deleted_at IS NULL
	`); err != nil {
		return fmt.Errorf("populate comment fts: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit search index refresh: %w", err)
	}
	return nil
}

func (r *Repository) SearchTickets(ctx context.Context, input ticketSearchQuery) ([]Ticket, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.searchTickets(ctx, runner, input)
}

func (r *Repository) ProjectExists(ctx context.Context, projectID string) (bool, error) {
	runner, err := r.runner()
	if err != nil {
		return false, err
	}
	return r.projectExists(ctx, runner, projectID)
}

func (r *Repository) ListActiveProjectIDs(ctx context.Context) ([]string, error) {
	runner, err := r.runner()
	if err != nil {
		return nil, err
	}
	return r.listActiveProjectIDs(ctx, runner)
}

func (r *Repository) runner() (sqlRunner, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("search: nil database")
	}
	return r.db, nil
}

func (r *Repository) insertSavedView(ctx context.Context, q sqlRunner, view SavedView) error {
	queryJSON, sortJSON, columnsJSON, err := encodeSavedViewConfig(view)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `
		INSERT INTO saved_views (
			id, owner_user_id, project_id, scope_type, name, query_json,
			sort_json, columns_json, display_mode, group_by, is_pinned, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, view.ID, nullableString(view.OwnerUserID), nullableString(view.ProjectID), view.ScopeType, view.Name, queryJSON, sortJSON, columnsJSON, view.DisplayMode, nullableString(view.GroupBy), boolInt(view.Pinned), formatTime(view.CreatedAt), formatTime(view.UpdatedAt))
	if err != nil {
		return fmt.Errorf("insert saved view: %w", err)
	}
	return nil
}

func (r *Repository) getSavedView(ctx context.Context, q sqlRunner, id string) (SavedView, error) {
	view, err := scanSavedView(q.QueryRowContext(ctx, `
		SELECT id, owner_user_id, project_id, scope_type, name, query_json,
			sort_json, columns_json, display_mode, group_by, is_pinned, created_at, updated_at
		FROM saved_views
		WHERE id = ?
	`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SavedView{}, notFound("saved_view", id)
		}
		return SavedView{}, fmt.Errorf("get saved view: %w", err)
	}
	return view, nil
}

func (r *Repository) updateSavedView(ctx context.Context, q sqlRunner, view SavedView) error {
	queryJSON, sortJSON, columnsJSON, err := encodeSavedViewConfig(view)
	if err != nil {
		return err
	}
	result, err := q.ExecContext(ctx, `
		UPDATE saved_views
		SET name = ?,
			query_json = ?,
			sort_json = ?,
			columns_json = ?,
			display_mode = ?,
			group_by = ?,
			is_pinned = ?,
			updated_at = ?
		WHERE id = ?
	`, view.Name, queryJSON, sortJSON, columnsJSON, view.DisplayMode, nullableString(view.GroupBy), boolInt(view.Pinned), formatTime(view.UpdatedAt), view.ID)
	if err != nil {
		return fmt.Errorf("update saved view: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated saved view rows: %w", err)
	}
	if updated == 0 {
		return notFound("saved_view", view.ID)
	}
	return nil
}

func (r *Repository) deleteSavedView(ctx context.Context, q sqlRunner, id string) error {
	result, err := q.ExecContext(ctx, `DELETE FROM saved_views WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete saved view: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted saved view rows: %w", err)
	}
	if deleted == 0 {
		return notFound("saved_view", id)
	}
	return nil
}

func (r *Repository) listSavedViews(ctx context.Context, q sqlRunner, input savedViewListQuery) ([]SavedView, error) {
	where, args := savedViewListWhere(input)
	args = append(args, input.Limit, input.Offset)
	rows, err := q.QueryContext(ctx, `
		SELECT id, owner_user_id, project_id, scope_type, name, query_json,
			sort_json, columns_json, display_mode, group_by, is_pinned, created_at, updated_at
		FROM saved_views
		WHERE `+where+`
		ORDER BY
			is_pinned DESC,
			CASE scope_type WHEN 'user' THEN 0 WHEN 'project' THEN 1 ELSE 2 END,
			name COLLATE NOCASE ASC,
			id ASC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list saved views: %w", err)
	}
	defer rows.Close()

	var views []SavedView
	for rows.Next() {
		view, err := scanSavedView(rows)
		if err != nil {
			return nil, fmt.Errorf("scan saved view: %w", err)
		}
		views = append(views, view)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate saved views: %w", err)
	}
	return views, nil
}

func savedViewListWhere(input savedViewListQuery) (string, []any) {
	parts := []string{"(scope_type = 'user' AND owner_user_id = ?)"}
	args := []any{input.OwnerUserID}
	if input.ProjectID != "" {
		parts[0] = "(scope_type = 'user' AND owner_user_id = ? AND (project_id IS NULL OR project_id = ?))"
		args = append(args, input.ProjectID)
	}

	projectIDs := input.VisibleProjectIDs
	if input.ProjectID != "" {
		projectIDs = []string{input.ProjectID}
	}
	if len(projectIDs) > 0 {
		placeholders := make([]string, len(projectIDs))
		for i, projectID := range projectIDs {
			placeholders[i] = "?"
			args = append(args, projectID)
		}
		parts = append(parts, "(scope_type = 'project' AND project_id IN ("+strings.Join(placeholders, ", ")+"))")
	}

	parts = append(parts, "(scope_type = 'global')")
	where := "(" + strings.Join(parts, " OR ") + ")"
	if input.PinnedOnly {
		where += " AND is_pinned = 1"
	}
	return where, args
}

func (r *Repository) searchTickets(ctx context.Context, q sqlRunner, input ticketSearchQuery) ([]Ticket, error) {
	if len(input.ProjectIDs) == 0 {
		return nil, nil
	}

	where := []string{"t.deleted_at IS NULL", "p.deleted_at IS NULL"}
	args := []any{}

	projectClause, projectArgs := inClause("t.project_id", input.ProjectIDs)
	where = append(where, projectClause)
	args = append(args, projectArgs...)

	for _, part := range input.Filter.Parts {
		where = append(where, part.SQL)
		args = append(args, part.Args...)
	}

	if input.FTSQuery != "" {
		where = append(where, `(
			t.id IN (
				SELECT ticket_id
				FROM ticket_fts
				WHERE ticket_fts MATCH ?
			)
			OR t.id IN (
				SELECT c.ticket_id
				FROM comment_fts cf
				JOIN ticket_comments c ON c.id = cf.comment_id
				WHERE comment_fts MATCH ?
				  AND c.deleted_at IS NULL
			)
		)`)
		args = append(args, input.FTSQuery, input.FTSQuery)
	}

	orderBy := ticketOrderBy(input.Sort)
	args = append(args, input.Limit, input.Offset)
	rows, err := q.QueryContext(ctx, `
		SELECT t.id, t.project_id, t.key, t.title, t.description, t.status,
			t.priority, t.type, t.reporter_id, t.assignee_id, t.parent_ticket_id,
			t.sprint_id, t.component_id, t.version_id, t.rank, t.start_date, t.due_date,
			t.created_at, t.updated_at, t.deleted_at
		FROM tickets t
		JOIN projects p ON p.id = t.project_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY `+orderBy+`
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("search tickets: %w", err)
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

func (r *Repository) projectExists(ctx context.Context, q sqlRunner, projectID string) (bool, error) {
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM projects
		WHERE id = ? AND deleted_at IS NULL
	`, projectID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check project exists: %w", err)
	}
	return exists == 1, nil
}

func (r *Repository) listActiveProjectIDs(ctx context.Context, q sqlRunner) ([]string, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id
		FROM projects
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list project ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan project id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project ids: %w", err)
	}
	return ids, nil
}

func encodeSavedViewConfig(view SavedView) (string, string, string, error) {
	queryJSON, err := json.Marshal(view.Query)
	if err != nil {
		return "", "", "", fmt.Errorf("encode saved view query: %w", err)
	}
	sort := view.Sort
	if sort == nil {
		sort = []SortSpec{}
	}
	sortJSON, err := json.Marshal(sort)
	if err != nil {
		return "", "", "", fmt.Errorf("encode saved view sort: %w", err)
	}
	columns := view.Columns
	if columns == nil {
		columns = []string{}
	}
	columnsJSON, err := json.Marshal(columns)
	if err != nil {
		return "", "", "", fmt.Errorf("encode saved view columns: %w", err)
	}
	return string(queryJSON), string(sortJSON), string(columnsJSON), nil
}

func scanSavedView(scanner rowScanner) (SavedView, error) {
	var view SavedView
	var ownerUserID sql.NullString
	var projectID sql.NullString
	var queryJSON string
	var sortJSON string
	var columnsJSON string
	var groupBy sql.NullString
	var isPinned int
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&view.ID, &ownerUserID, &projectID, &view.ScopeType, &view.Name, &queryJSON, &sortJSON, &columnsJSON, &view.DisplayMode, &groupBy, &isPinned, &createdAt, &updatedAt); err != nil {
		return SavedView{}, err
	}
	view.OwnerUserID = nullString(ownerUserID)
	view.ProjectID = nullString(projectID)
	view.GroupBy = nullString(groupBy)
	view.Pinned = isPinned != 0
	if view.DisplayMode == "" {
		view.DisplayMode = SavedViewDisplayList
	}
	if queryJSON == "" {
		queryJSON = "{}"
	}
	if sortJSON == "" {
		sortJSON = "[]"
	}
	if columnsJSON == "" {
		columnsJSON = "[]"
	}
	if err := json.Unmarshal([]byte(queryJSON), &view.Query); err != nil {
		return SavedView{}, fmt.Errorf("decode saved view query: %w", err)
	}
	if err := json.Unmarshal([]byte(sortJSON), &view.Sort); err != nil {
		return SavedView{}, fmt.Errorf("decode saved view sort: %w", err)
	}
	if err := json.Unmarshal([]byte(columnsJSON), &view.Columns); err != nil {
		return SavedView{}, fmt.Errorf("decode saved view columns: %w", err)
	}
	var err error
	view.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return SavedView{}, fmt.Errorf("parse saved view created_at: %w", err)
	}
	view.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return SavedView{}, fmt.Errorf("parse saved view updated_at: %w", err)
	}
	return view, nil
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

func inClause(field string, values []string) (string, []any) {
	placeholders := make([]string, len(values))
	args := make([]any, len(values))
	for i, value := range values {
		placeholders[i] = "?"
		args[i] = value
	}
	return field + " IN (" + strings.Join(placeholders, ", ") + ")", args
}

func ticketOrderBy(sort []SortSpec) string {
	if len(sort) == 0 {
		return "t.updated_at DESC, t.id ASC"
	}
	parts := make([]string, 0, len(sort)+1)
	for _, spec := range sort {
		field := ticketSortFields[spec.Field]
		direction := "ASC"
		if spec.Direction == SortDirectionDesc {
			direction = "DESC"
		}
		parts = append(parts, field+" "+direction)
	}
	parts = append(parts, "t.id ASC")
	return strings.Join(parts, ", ")
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
