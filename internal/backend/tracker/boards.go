package tracker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

var defaultProjectStatuses = []ProjectStatusInput{
	{Slug: "todo", Name: "Todo"},
	{Slug: "in_progress", Name: "In Progress"},
	{Slug: "done", Name: "Done"},
}

func (s *Service) ListProjectStatuses(ctx context.Context, principal authz.Principal, projectID string) ([]ProjectStatus, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	return s.listProjectStatuses(ctx, s.db, projectID)
}

func (s *Service) ReplaceProjectStatuses(ctx context.Context, principal authz.Principal, projectID string, input ReplaceProjectStatusesInput) ([]ProjectStatus, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionBoardsManage, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	statuses, err := s.buildProjectStatuses(projectID, input.Statuses)
	if err != nil {
		return nil, err
	}

	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.requireUsedTicketStatuses(ctx, tx, projectID, statuses); err != nil {
			return err
		}
		if err := s.replaceProjectStatuses(ctx, tx, projectID, statuses); err != nil {
			return err
		}
		return s.pruneInvalidBoardColumns(ctx, tx, projectID, statuses)
	}); err != nil {
		return nil, err
	}

	return s.ListProjectStatuses(ctx, principal, projectID)
}

func (s *Service) ListBoards(ctx context.Context, principal authz.Principal, projectID string) ([]Board, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, description, created_by, created_at, updated_at
		FROM boards
		WHERE project_id = ?
		ORDER BY name ASC, id ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}
	defer rows.Close()

	boards := []Board{}
	for rows.Next() {
		board, err := scanBoard(rows)
		if err != nil {
			return nil, err
		}
		board.Columns, err = s.listBoardColumns(ctx, s.db, board.ID)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate boards: %w", err)
	}
	return boards, nil
}

func (s *Service) CreateBoard(ctx context.Context, principal authz.Principal, input CreateBoardInput) (Board, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return Board{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionBoardsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return Board{}, err
	}
	if _, err := s.repo.GetProject(ctx, input.ProjectID); err != nil {
		return Board{}, err
	}

	board, err := s.buildBoard(principal, input)
	if err != nil {
		return Board{}, err
	}
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		columns, err := s.buildBoardColumns(ctx, tx, board.ID, input.ProjectID, input.StatusSlugs, board.CreatedAt)
		if err != nil {
			return err
		}
		board.Columns = columns
		return s.insertBoard(ctx, tx, board)
	}); err != nil {
		return Board{}, err
	}
	return board, nil
}

func (s *Service) GetBoard(ctx context.Context, principal authz.Principal, boardID string) (Board, error) {
	board, err := s.getBoard(ctx, boardID)
	if err != nil {
		return Board{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(board.ProjectID)); err != nil {
		return Board{}, err
	}
	board.Columns, err = s.listBoardColumns(ctx, s.db, board.ID)
	if err != nil {
		return Board{}, err
	}
	return board, nil
}

func (s *Service) UpdateBoard(ctx context.Context, principal authz.Principal, boardID string, input UpdateBoardInput) (Board, error) {
	current, err := s.getBoard(ctx, boardID)
	if err != nil {
		return Board{}, err
	}
	if err := s.require(principal, authz.PermissionBoardsManage, authz.ProjectScope(current.ProjectID)); err != nil {
		return Board{}, err
	}

	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		updated.Description = strings.TrimSpace(*input.Description)
	}
	if fields := boardFields(updated.Name, updated.Description); len(fields) > 0 {
		return Board{}, validationFailed(fields)
	}

	updated.UpdatedAt = s.now().UTC()
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if input.StatusSlugs != nil {
			columns, err := s.buildBoardColumns(ctx, tx, updated.ID, updated.ProjectID, *input.StatusSlugs, updated.UpdatedAt)
			if err != nil {
				return err
			}
			updated.Columns = columns
		}
		if err := s.updateBoard(ctx, tx, updated); err != nil {
			return err
		}
		if input.StatusSlugs != nil {
			return s.replaceBoardColumns(ctx, tx, updated.ID, updated.Columns)
		}
		return nil
	}); err != nil {
		return Board{}, err
	}
	return s.GetBoard(ctx, principal, boardID)
}

func (s *Service) DeleteBoard(ctx context.Context, principal authz.Principal, boardID string) error {
	board, err := s.getBoard(ctx, boardID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionBoardsManage, authz.ProjectScope(board.ProjectID)); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM boards WHERE id = ?", board.ID)
	if err != nil {
		return fmt.Errorf("delete board: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check board delete: %w", err)
	}
	if affected == 0 {
		return notFound("board", boardID)
	}
	return nil
}

func (s *Service) ListBoardTickets(ctx context.Context, principal authz.Principal, boardID string) (BoardTickets, error) {
	board, err := s.GetBoard(ctx, principal, boardID)
	if err != nil {
		return BoardTickets{}, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(board.ProjectID)); err != nil {
		return BoardTickets{}, err
	}

	result := BoardTickets{Board: board, Columns: make([]BoardTicketsColumn, 0, len(board.Columns))}
	for _, column := range board.Columns {
		tickets, err := s.listTicketsForStatus(ctx, board.ProjectID, column.StatusSlug)
		if err != nil {
			return BoardTickets{}, err
		}
		tickets, err = s.attachTicketWatcherStatus(ctx, principal, tickets)
		if err != nil {
			return BoardTickets{}, err
		}
		result.Columns = append(result.Columns, BoardTicketsColumn{
			Column:  column,
			Tickets: tickets,
		})
	}
	return result, nil
}

func (s *Service) seedDefaultProjectWorkflow(ctx context.Context, q sqlRunner, project Project) error {
	statuses, err := s.buildProjectStatuses(project.ID, defaultProjectStatuses)
	if err != nil {
		return err
	}
	if err := s.replaceProjectStatuses(ctx, q, project.ID, statuses); err != nil {
		return err
	}

	boardID := "board_" + project.ID + "_default"
	now := project.CreatedAt
	board := Board{
		ID:          boardID,
		ProjectID:   project.ID,
		Name:        "Default Board",
		Description: "Default project workflow board",
		CreatedBy:   project.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	columns, err := s.boardColumnsFromStatuses(board.ID, statuses, now)
	if err != nil {
		return err
	}
	board.Columns = columns
	return s.insertBoard(ctx, q, board)
}

func (s *Service) buildProjectStatuses(projectID string, inputs []ProjectStatusInput) ([]ProjectStatus, error) {
	if len(inputs) == 0 {
		return nil, validationFailed(map[string]string{"statuses": "At least one status is required"})
	}
	if len(inputs) > 50 {
		return nil, validationFailed(map[string]string{"statuses": "Must contain 50 or fewer statuses"})
	}

	now := s.now().UTC()
	seen := map[string]struct{}{}
	statuses := make([]ProjectStatus, 0, len(inputs))
	fields := map[string]string{}
	for index, input := range inputs {
		slug := normalizeSlug(input.Slug)
		name := strings.TrimSpace(input.Name)
		if slug == "" {
			fields["statuses"] = "Status slugs are required"
		} else if !slugPattern.MatchString(slug) {
			fields["statuses"] = "Status slugs must be lowercase slugs"
		}
		if name == "" {
			fields["statuses"] = "Status names are required"
		} else if len(name) > 100 {
			fields["statuses"] = "Status names must be 100 characters or fewer"
		}
		if _, ok := seen[slug]; ok {
			fields["statuses"] = "Status slugs must be unique"
		}
		seen[slug] = struct{}{}
		statuses = append(statuses, ProjectStatus{
			ID:        fmt.Sprintf("status_%s_%s", projectID, slug),
			ProjectID: projectID,
			Slug:      slug,
			Name:      name,
			Position:  index,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	if len(fields) > 0 {
		return nil, validationFailed(fields)
	}
	return statuses, nil
}

func (s *Service) buildBoard(principal authz.Principal, input CreateBoardInput) (Board, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if fields := boardFields(name, description); len(fields) > 0 {
		return Board{}, validationFailed(fields)
	}
	id, err := newID("board")
	if err != nil {
		return Board{}, err
	}
	now := s.now().UTC()
	return Board{
		ID:          id,
		ProjectID:   strings.TrimSpace(input.ProjectID),
		Name:        name,
		Description: description,
		CreatedBy:   actorID(principal),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func boardFields(name string, description string) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(name) == "" {
		fields["name"] = "Required"
	} else if len(strings.TrimSpace(name)) > 200 {
		fields["name"] = "Must be 200 characters or fewer"
	}
	if len(strings.TrimSpace(description)) > 2000 {
		fields["description"] = "Must be 2000 characters or fewer"
	}
	return fields
}

func (s *Service) buildBoardColumns(ctx context.Context, q sqlRunner, boardID string, projectID string, statusSlugs []string, createdAt time.Time) ([]BoardColumn, error) {
	if len(statusSlugs) == 0 {
		statuses, err := s.listProjectStatuses(ctx, q, projectID)
		if err != nil {
			return nil, err
		}
		return s.boardColumnsFromStatuses(boardID, statuses, createdAt)
	}

	statusNames, err := s.projectStatusNames(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	if len(statusSlugs) > 50 {
		return nil, validationFailed(map[string]string{"status_slugs": "Must contain 50 or fewer statuses"})
	}
	seen := map[string]struct{}{}
	columns := make([]BoardColumn, 0, len(statusSlugs))
	for index, slug := range statusSlugs {
		slug = normalizeSlug(slug)
		if slug == "" {
			return nil, validationFailed(map[string]string{"status_slugs": "Status slugs must be non-empty"})
		}
		if _, ok := seen[slug]; ok {
			return nil, validationFailed(map[string]string{"status_slugs": "Status slugs must be unique"})
		}
		seen[slug] = struct{}{}
		name, ok := statusNames[slug]
		if !ok {
			return nil, validationFailed(map[string]string{"status_slugs": "Status is not configured for project"})
		}
		columns = append(columns, BoardColumn{
			ID:         fmt.Sprintf("column_%s_%06d_%s", boardID, index+1, slug),
			BoardID:    boardID,
			StatusSlug: slug,
			Name:       name,
			Position:   index,
			CreatedAt:  createdAt,
		})
	}
	return columns, nil
}

func (s *Service) boardColumnsFromStatuses(boardID string, statuses []ProjectStatus, createdAt time.Time) ([]BoardColumn, error) {
	columns := make([]BoardColumn, 0, len(statuses))
	for index, status := range statuses {
		columns = append(columns, BoardColumn{
			ID:         fmt.Sprintf("column_%s_%06d_%s", boardID, index+1, status.Slug),
			BoardID:    boardID,
			StatusSlug: status.Slug,
			Name:       status.Name,
			Position:   index,
			CreatedAt:  createdAt,
		})
	}
	return columns, nil
}

func (s *Service) requireUsedTicketStatuses(ctx context.Context, q sqlRunner, projectID string, statuses []ProjectStatus) error {
	allowed := map[string]struct{}{}
	for _, status := range statuses {
		allowed[status.Slug] = struct{}{}
	}

	rows, err := q.QueryContext(ctx, `
		SELECT DISTINCT status
		FROM tickets
		WHERE project_id = ? AND deleted_at IS NULL
	`, projectID)
	if err != nil {
		return fmt.Errorf("list used ticket statuses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return fmt.Errorf("scan used ticket status: %w", err)
		}
		if _, ok := allowed[status]; !ok {
			return validationFailed(map[string]string{"statuses": "All statuses used by tickets must remain configured"})
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate used ticket statuses: %w", err)
	}
	return nil
}

func (s *Service) replaceProjectStatuses(ctx context.Context, q sqlRunner, projectID string, statuses []ProjectStatus) error {
	if _, err := q.ExecContext(ctx, "DELETE FROM project_statuses WHERE project_id = ?", projectID); err != nil {
		return fmt.Errorf("delete project statuses: %w", err)
	}
	for _, status := range statuses {
		if _, err := q.ExecContext(ctx, `
			INSERT INTO project_statuses (id, project_id, slug, name, position, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, status.ID, status.ProjectID, status.Slug, status.Name, status.Position, formatTime(status.CreatedAt), formatTime(status.UpdatedAt)); err != nil {
			return fmt.Errorf("insert project status: %w", err)
		}
	}
	return nil
}

func (s *Service) listProjectStatuses(ctx context.Context, q sqlRunner, projectID string) ([]ProjectStatus, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, project_id, slug, name, position, created_at, updated_at
		FROM project_statuses
		WHERE project_id = ?
		ORDER BY position ASC, slug ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project statuses: %w", err)
	}
	defer rows.Close()
	statuses := []ProjectStatus{}
	for rows.Next() {
		status, err := scanProjectStatus(rows)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project statuses: %w", err)
	}
	return statuses, nil
}

func (s *Service) projectStatusNames(ctx context.Context, q sqlRunner, projectID string) (map[string]string, error) {
	statuses, err := s.listProjectStatuses(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	for _, status := range statuses {
		names[status.Slug] = status.Name
	}
	return names, nil
}

func (s *Service) pruneInvalidBoardColumns(ctx context.Context, q sqlRunner, projectID string, statuses []ProjectStatus) error {
	allowed := map[string]struct{}{}
	for _, status := range statuses {
		allowed[status.Slug] = struct{}{}
	}
	rows, err := q.QueryContext(ctx, `
		SELECT bc.id, bc.status_slug
		FROM board_columns bc
		JOIN boards b ON b.id = bc.board_id
		WHERE b.project_id = ?
	`, projectID)
	if err != nil {
		return fmt.Errorf("list board columns: %w", err)
	}
	defer rows.Close()
	var deleteIDs []string
	for rows.Next() {
		var id string
		var slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return fmt.Errorf("scan board column: %w", err)
		}
		if _, ok := allowed[slug]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate board columns: %w", err)
	}
	for _, id := range deleteIDs {
		if _, err := q.ExecContext(ctx, "DELETE FROM board_columns WHERE id = ?", id); err != nil {
			return fmt.Errorf("delete invalid board column: %w", err)
		}
	}
	return nil
}

func (s *Service) insertBoard(ctx context.Context, q sqlRunner, board Board) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO boards (id, project_id, name, description, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, board.ID, board.ProjectID, board.Name, nullableString(board.Description), nullableString(board.CreatedBy), formatTime(board.CreatedAt), formatTime(board.UpdatedAt))
	if err != nil {
		if isUniqueConstraint(err) {
			return conflict("board", "name", board.Name)
		}
		return fmt.Errorf("insert board: %w", err)
	}
	return s.replaceBoardColumns(ctx, q, board.ID, board.Columns)
}

func (s *Service) updateBoard(ctx context.Context, q sqlRunner, board Board) error {
	_, err := q.ExecContext(ctx, `
		UPDATE boards
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`, board.Name, nullableString(board.Description), formatTime(board.UpdatedAt), board.ID)
	if err != nil {
		if isUniqueConstraint(err) {
			return conflict("board", "name", board.Name)
		}
		return fmt.Errorf("update board: %w", err)
	}
	return nil
}

func (s *Service) replaceBoardColumns(ctx context.Context, q sqlRunner, boardID string, columns []BoardColumn) error {
	if _, err := q.ExecContext(ctx, "DELETE FROM board_columns WHERE board_id = ?", boardID); err != nil {
		return fmt.Errorf("delete board columns: %w", err)
	}
	for _, column := range columns {
		if _, err := q.ExecContext(ctx, `
			INSERT INTO board_columns (id, board_id, status_slug, name, position, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, column.ID, column.BoardID, column.StatusSlug, column.Name, column.Position, formatTime(column.CreatedAt)); err != nil {
			return fmt.Errorf("insert board column: %w", err)
		}
	}
	return nil
}

func (s *Service) getBoard(ctx context.Context, boardID string) (Board, error) {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return Board{}, validationFailed(map[string]string{"board_id": "Required"})
	}
	board, err := scanBoard(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, description, created_by, created_at, updated_at
		FROM boards
		WHERE id = ?
	`, boardID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Board{}, notFound("board", boardID)
		}
		return Board{}, fmt.Errorf("get board: %w", err)
	}
	return board, nil
}

func (s *Service) listBoardColumns(ctx context.Context, q sqlRunner, boardID string) ([]BoardColumn, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, board_id, status_slug, name, position, created_at
		FROM board_columns
		WHERE board_id = ?
		ORDER BY position ASC, status_slug ASC
	`, boardID)
	if err != nil {
		return nil, fmt.Errorf("list board columns: %w", err)
	}
	defer rows.Close()
	columns := []BoardColumn{}
	for rows.Next() {
		column, err := scanBoardColumn(rows)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate board columns: %w", err)
	}
	return columns, nil
}

func (s *Service) listTicketsForStatus(ctx context.Context, projectID string, status string) ([]Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, story_points, created_at, updated_at, deleted_at
		FROM tickets
		WHERE project_id = ? AND status = ? AND deleted_at IS NULL
		ORDER BY
			CASE WHEN rank IS NULL OR rank = '' THEN 1 ELSE 0 END ASC,
			rank ASC,
			created_at DESC,
			key DESC
	`, projectID, status)
	if err != nil {
		return nil, fmt.Errorf("list board tickets: %w", err)
	}
	defer rows.Close()
	tickets := []Ticket{}
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate board tickets: %w", err)
	}
	return s.attachTicketDetailsToTickets(ctx, tickets)
}

func scanProjectStatus(scanner rowScanner) (ProjectStatus, error) {
	var status ProjectStatus
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&status.ID, &status.ProjectID, &status.Slug, &status.Name, &status.Position, &createdAt, &updatedAt); err != nil {
		return ProjectStatus{}, err
	}
	var err error
	status.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return ProjectStatus{}, fmt.Errorf("parse status created_at: %w", err)
	}
	status.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return ProjectStatus{}, fmt.Errorf("parse status updated_at: %w", err)
	}
	return status, nil
}

func scanBoard(scanner rowScanner) (Board, error) {
	var board Board
	var description sql.NullString
	var createdBy sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&board.ID, &board.ProjectID, &board.Name, &description, &createdBy, &createdAt, &updatedAt); err != nil {
		return Board{}, err
	}
	var err error
	board.Description = nullString(description)
	board.CreatedBy = nullString(createdBy)
	board.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Board{}, fmt.Errorf("parse board created_at: %w", err)
	}
	board.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Board{}, fmt.Errorf("parse board updated_at: %w", err)
	}
	return board, nil
}

func scanBoardColumn(scanner rowScanner) (BoardColumn, error) {
	var column BoardColumn
	var createdAt string
	if err := scanner.Scan(&column.ID, &column.BoardID, &column.StatusSlug, &column.Name, &column.Position, &createdAt); err != nil {
		return BoardColumn{}, err
	}
	var err error
	column.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return BoardColumn{}, fmt.Errorf("parse board column created_at: %w", err)
	}
	return column, nil
}
