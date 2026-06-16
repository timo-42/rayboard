package tracker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
)

const (
	SprintStatePlanned   = "planned"
	SprintStateActive    = "active"
	SprintStateCompleted = "completed"
	dateOnlyLayout       = "2006-01-02"
)

func (s *Service) ListSprints(ctx context.Context, principal authz.Principal, projectID string, state string) ([]Sprint, error) {
	projectID = strings.TrimSpace(projectID)
	state = normalizeSlug(state)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if state != "" && !validSprintState(state) {
		return nil, validationFailed(map[string]string{"state": "Invalid sprint state"})
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, project_id, name, goal, state, start_date, end_date, started_at, completed_at, created_at, updated_at
		FROM sprints
		WHERE project_id = ?`
	args := []any{projectID}
	if state != "" {
		query += " AND state = ?"
		args = append(args, state)
	}
	query += " ORDER BY created_at DESC, id DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}
	defer rows.Close()
	var sprints []Sprint
	for rows.Next() {
		sprint, err := scanSprint(rows)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, sprint)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sprints: %w", err)
	}
	return sprints, nil
}

func (s *Service) CreateSprint(ctx context.Context, principal authz.Principal, input CreateSprintInput) (Sprint, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return Sprint{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionSprintsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return Sprint{}, err
	}
	if _, err := s.repo.GetProject(ctx, input.ProjectID); err != nil {
		return Sprint{}, err
	}
	sprint, err := s.buildSprint(input)
	if err != nil {
		return Sprint{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO sprints (id, project_id, name, goal, state, start_date, end_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, sprint.ID, sprint.ProjectID, sprint.Name, nullableString(sprint.Goal), sprint.State, nullableString(sprint.StartDate), nullableString(sprint.EndDate), formatTime(sprint.CreatedAt), formatTime(sprint.UpdatedAt)); err != nil {
		return Sprint{}, fmt.Errorf("insert sprint: %w", err)
	}
	return sprint, nil
}

func (s *Service) GetSprint(ctx context.Context, principal authz.Principal, sprintID string) (Sprint, error) {
	sprint, err := s.getSprint(ctx, sprintID)
	if err != nil {
		return Sprint{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(sprint.ProjectID)); err != nil {
		return Sprint{}, err
	}
	return sprint, nil
}

func (s *Service) UpdateSprint(ctx context.Context, principal authz.Principal, sprintID string, input UpdateSprintInput) (Sprint, error) {
	current, err := s.getSprint(ctx, sprintID)
	if err != nil {
		return Sprint{}, err
	}
	if err := s.require(principal, authz.PermissionSprintsManage, authz.ProjectScope(current.ProjectID)); err != nil {
		return Sprint{}, err
	}
	if current.State == SprintStateCompleted {
		return Sprint{}, &ConflictError{Resource: "sprint", Field: "state", Value: current.State, Message: "completed sprints cannot be edited"}
	}
	updated, err := s.applySprintUpdate(current, input)
	if err != nil {
		return Sprint{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE sprints
		SET name = ?, goal = ?, start_date = ?, end_date = ?, updated_at = ?
		WHERE id = ?
	`, updated.Name, nullableString(updated.Goal), nullableString(updated.StartDate), nullableString(updated.EndDate), formatTime(updated.UpdatedAt), updated.ID); err != nil {
		return Sprint{}, fmt.Errorf("update sprint: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteSprint(ctx context.Context, principal authz.Principal, sprintID string) error {
	sprint, err := s.getSprint(ctx, sprintID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionSprintsManage, authz.ProjectScope(sprint.ProjectID)); err != nil {
		return err
	}
	if sprint.State == SprintStateActive {
		return &ConflictError{Resource: "sprint", Field: "state", Value: sprint.State, Message: "active sprints cannot be deleted"}
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM sprints WHERE id = ?", sprint.ID)
	if err != nil {
		return fmt.Errorf("delete sprint: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check sprint delete: %w", err)
	}
	if affected == 0 {
		return notFound("sprint", sprintID)
	}
	return nil
}

func (s *Service) StartSprint(ctx context.Context, principal authz.Principal, sprintID string) (Sprint, error) {
	current, err := s.getSprint(ctx, sprintID)
	if err != nil {
		return Sprint{}, err
	}
	if err := s.require(principal, authz.PermissionSprintsManage, authz.ProjectScope(current.ProjectID)); err != nil {
		return Sprint{}, err
	}
	if current.State != SprintStatePlanned {
		return Sprint{}, &ConflictError{Resource: "sprint", Field: "state", Value: current.State, Message: "only planned sprints can be started"}
	}
	if err := s.requireNoActiveSprint(ctx, current.ProjectID, current.ID); err != nil {
		return Sprint{}, err
	}
	now := s.now().UTC()
	current.State = SprintStateActive
	current.StartedAt = &now
	current.UpdatedAt = now
	if _, err := s.db.ExecContext(ctx, `
		UPDATE sprints
		SET state = ?, started_at = ?, updated_at = ?
		WHERE id = ?
	`, current.State, formatTime(now), formatTime(now), current.ID); err != nil {
		return Sprint{}, fmt.Errorf("start sprint: %w", err)
	}
	s.publish(ctx, events.Event{
		Type:      "sprint.started",
		ActorID:   actorID(principal),
		ProjectID: current.ProjectID,
		ObjectID:  current.ID,
		At:        now,
		Data:      map[string]any{"name": current.Name},
	})
	return current, nil
}

func (s *Service) CompleteSprint(ctx context.Context, principal authz.Principal, sprintID string) (Sprint, error) {
	current, err := s.getSprint(ctx, sprintID)
	if err != nil {
		return Sprint{}, err
	}
	if err := s.require(principal, authz.PermissionSprintsManage, authz.ProjectScope(current.ProjectID)); err != nil {
		return Sprint{}, err
	}
	if current.State != SprintStateActive {
		return Sprint{}, &ConflictError{Resource: "sprint", Field: "state", Value: current.State, Message: "only active sprints can be completed"}
	}
	now := s.now().UTC()
	current.State = SprintStateCompleted
	current.CompletedAt = &now
	current.UpdatedAt = now
	if _, err := s.db.ExecContext(ctx, `
		UPDATE sprints
		SET state = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`, current.State, formatTime(now), formatTime(now), current.ID); err != nil {
		return Sprint{}, fmt.Errorf("complete sprint: %w", err)
	}
	s.publish(ctx, events.Event{
		Type:      "sprint.completed",
		ActorID:   actorID(principal),
		ProjectID: current.ProjectID,
		ObjectID:  current.ID,
		At:        now,
		Data:      map[string]any{"name": current.Name},
	})
	return current, nil
}

func (s *Service) SetTicketSprint(ctx context.Context, principal authz.Principal, ticketID string, sprintID string) (Ticket, error) {
	ticketID = strings.TrimSpace(ticketID)
	sprintID = strings.TrimSpace(sprintID)
	if ticketID == "" {
		return Ticket{}, validationFailed(map[string]string{"ticket_id": "Required"})
	}
	current, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}
	if err := s.require(principal, authz.PermissionSprintsManage, authz.ProjectScope(current.ProjectID)); err != nil {
		return Ticket{}, err
	}
	if err := s.requireSprint(ctx, s.db, "sprint_id", sprintID, current.ProjectID); err != nil {
		return Ticket{}, err
	}
	if current.SprintID == sprintID {
		return current, nil
	}
	updated := current
	updated.SprintID = sprintID
	updated.UpdatedAt = s.now().UTC()
	data := map[string]any{"changes": map[string]ticketFieldChange{
		"sprint_id": {Old: current.SprintID, New: updated.SprintID},
	}}
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.updateTicket(ctx, tx, updated); err != nil {
			return err
		}
		activityID, err := newID("activity")
		if err != nil {
			return err
		}
		return s.repo.insertTicketActivity(ctx, tx, TicketActivity{
			ID:           activityID,
			TicketID:     updated.ID,
			ActorID:      actorID(principal),
			ActivityType: activityTicketUpdated,
			Data:         data,
			CreatedAt:    updated.UpdatedAt,
		})
	}); err != nil {
		return Ticket{}, err
	}
	s.publish(ctx, events.Event{
		Type:      activityTicketUpdated,
		ActorID:   actorID(principal),
		ProjectID: updated.ProjectID,
		ObjectID:  updated.ID,
		At:        updated.UpdatedAt,
		Data:      data,
	})
	return updated, nil
}

func (s *Service) buildSprint(input CreateSprintInput) (Sprint, error) {
	fields := sprintFields(input.Name, input.Goal, input.StartDate, input.EndDate)
	if len(fields) > 0 {
		return Sprint{}, validationFailed(fields)
	}
	id, err := newID("sprint")
	if err != nil {
		return Sprint{}, err
	}
	now := s.now().UTC()
	return Sprint{
		ID:        id,
		ProjectID: input.ProjectID,
		Name:      strings.TrimSpace(input.Name),
		Goal:      strings.TrimSpace(input.Goal),
		State:     SprintStatePlanned,
		StartDate: strings.TrimSpace(input.StartDate),
		EndDate:   strings.TrimSpace(input.EndDate),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Service) applySprintUpdate(current Sprint, input UpdateSprintInput) (Sprint, error) {
	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.Goal != nil {
		updated.Goal = strings.TrimSpace(*input.Goal)
	}
	if input.StartDate != nil {
		updated.StartDate = strings.TrimSpace(*input.StartDate)
	}
	if input.EndDate != nil {
		updated.EndDate = strings.TrimSpace(*input.EndDate)
	}
	if fields := sprintFields(updated.Name, updated.Goal, updated.StartDate, updated.EndDate); len(fields) > 0 {
		return Sprint{}, validationFailed(fields)
	}
	updated.UpdatedAt = s.now().UTC()
	return updated, nil
}

func sprintFields(name string, goal string, startDate string, endDate string) map[string]string {
	fields := map[string]string{}
	name = strings.TrimSpace(name)
	if name == "" {
		fields["name"] = "Required"
	}
	if len(name) > 200 {
		fields["name"] = "Must be 200 characters or fewer"
	}
	if len(strings.TrimSpace(goal)) > 2000 {
		fields["goal"] = "Must be 2000 characters or fewer"
	}
	validateSprintDate(fields, "start_date", startDate)
	validateSprintDate(fields, "end_date", endDate)
	if strings.TrimSpace(startDate) != "" && strings.TrimSpace(endDate) != "" {
		start, _ := time.Parse(dateOnlyLayout, strings.TrimSpace(startDate))
		end, _ := time.Parse(dateOnlyLayout, strings.TrimSpace(endDate))
		if !end.IsZero() && end.Before(start) {
			fields["end_date"] = "Must be on or after start_date"
		}
	}
	return fields
}

func validateSprintDate(fields map[string]string, field string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if _, err := time.Parse(dateOnlyLayout, value); err != nil {
		fields[field] = "Must use YYYY-MM-DD"
	}
}

func (s *Service) getSprint(ctx context.Context, sprintID string) (Sprint, error) {
	sprintID = strings.TrimSpace(sprintID)
	if sprintID == "" {
		return Sprint{}, validationFailed(map[string]string{"sprint_id": "Required"})
	}
	sprint, err := scanSprint(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, goal, state, start_date, end_date, started_at, completed_at, created_at, updated_at
		FROM sprints
		WHERE id = ?
	`, sprintID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Sprint{}, notFound("sprint", sprintID)
		}
		return Sprint{}, fmt.Errorf("get sprint: %w", err)
	}
	return sprint, nil
}

func (s *Service) requireNoActiveSprint(ctx context.Context, projectID string, exceptSprintID string) error {
	var activeID string
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM sprints
		WHERE project_id = ? AND state = ? AND id != ?
		LIMIT 1
	`, projectID, SprintStateActive, exceptSprintID).Scan(&activeID)
	if err == nil {
		return &ConflictError{Resource: "sprint", Field: "state", Value: activeID, Message: "project already has an active sprint"}
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return fmt.Errorf("check active sprint: %w", err)
}

func scanSprint(scanner rowScanner) (Sprint, error) {
	var sprint Sprint
	var goal sql.NullString
	var startDate sql.NullString
	var endDate sql.NullString
	var startedAt sql.NullString
	var completedAt sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&sprint.ID, &sprint.ProjectID, &sprint.Name, &goal, &sprint.State, &startDate, &endDate, &startedAt, &completedAt, &createdAt, &updatedAt); err != nil {
		return Sprint{}, err
	}
	sprint.Goal = nullString(goal)
	sprint.StartDate = nullString(startDate)
	sprint.EndDate = nullString(endDate)
	sprint.StartedAt = parseNullableTime(startedAt)
	sprint.CompletedAt = parseNullableTime(completedAt)
	var err error
	sprint.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Sprint{}, fmt.Errorf("parse sprint created_at: %w", err)
	}
	sprint.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Sprint{}, fmt.Errorf("parse sprint updated_at: %w", err)
	}
	return sprint, nil
}

func validSprintState(state string) bool {
	switch state {
	case SprintStatePlanned, SprintStateActive, SprintStateCompleted:
		return true
	default:
		return false
	}
}
