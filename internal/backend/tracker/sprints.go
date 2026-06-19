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
	SprintStatePlanned        = "planned"
	SprintStateActive         = "active"
	SprintStateCompleted      = "completed"
	SprintReportScopeCurrent  = "current"
	SprintReportScopeSnapshot = "completed_snapshot"
	dateOnlyLayout            = "2006-01-02"
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

func (s *Service) GetSprintReport(ctx context.Context, principal authz.Principal, sprintID string) (SprintReport, error) {
	sprint, err := s.getSprint(ctx, sprintID)
	if err != nil {
		return SprintReport{}, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(sprint.ProjectID)); err != nil {
		return SprintReport{}, err
	}

	scope := SprintReportScopeCurrent
	var snapshotAt *time.Time
	var tickets []Ticket
	if sprint.State == SprintStateCompleted {
		var snapshotExists bool
		tickets, snapshotAt, snapshotExists, err = s.listSprintSnapshotTickets(ctx, sprint.ID)
		if err != nil {
			return SprintReport{}, err
		}
		scope = SprintReportScopeSnapshot
		if snapshotAt == nil {
			snapshotAt = sprint.CompletedAt
		}
		if !snapshotExists {
			tickets = nil
		}
	}
	if tickets == nil {
		tickets, err = s.listSprintReportTickets(ctx, sprint.ID)
		if err != nil {
			return SprintReport{}, err
		}
	}
	tickets, err = s.attachTicketDetailsToTickets(ctx, tickets)
	if err != nil {
		return SprintReport{}, err
	}
	analytics, err := s.sprintAnalytics(ctx, sprint, tickets)
	if err != nil {
		return SprintReport{}, err
	}
	return SprintReport{
		Sprint:     sprint,
		Scope:      scope,
		SnapshotAt: snapshotAt,
		Progress:   sprintReportProgress(tickets),
		Analytics:  analytics,
		Tickets:    tickets,
	}, nil
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
	event := events.Event{
		Type:        "sprint.started",
		ActorID:     actorID(principal),
		ProjectID:   current.ProjectID,
		ObjectID:    current.ID,
		SubjectType: "sprint",
		SubjectID:   current.ID,
		At:          now,
		Data:        map[string]any{"name": current.Name},
	}
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			UPDATE sprints
			SET state = ?, started_at = ?, updated_at = ?
			WHERE id = ?
		`, current.State, formatTime(now), formatTime(now), current.ID); err != nil {
			return fmt.Errorf("start sprint: %w", err)
		}
		return s.appendDomainEvent(ctx, tx, event)
	}); err != nil {
		return Sprint{}, err
	}
	s.publish(ctx, event)
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
	event := events.Event{
		Type:        "sprint.completed",
		ActorID:     actorID(principal),
		ProjectID:   current.ProjectID,
		ObjectID:    current.ID,
		SubjectType: "sprint",
		SubjectID:   current.ID,
		At:          now,
		Data:        map[string]any{"name": current.Name},
	}
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			UPDATE sprints
			SET state = ?, completed_at = ?, updated_at = ?
			WHERE id = ?
		`, current.State, formatTime(now), formatTime(now), current.ID); err != nil {
			return fmt.Errorf("complete sprint: %w", err)
		}
		if err := s.snapshotSprintReportTickets(ctx, tx, current.ID, now); err != nil {
			return err
		}
		return s.appendDomainEvent(ctx, tx, event)
	}); err != nil {
		return Sprint{}, err
	}
	s.publish(ctx, event)
	return current, nil
}

func (s *Service) snapshotSprintReportTickets(ctx context.Context, tx *sql.Tx, sprintID string, capturedAt time.Time) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM sprint_report_tickets WHERE sprint_id = ?", sprintID); err != nil {
		return fmt.Errorf("delete sprint report snapshot: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sprint_report_snapshots (sprint_id, captured_at)
		VALUES (?, ?)
		ON CONFLICT(sprint_id) DO UPDATE SET captured_at = excluded.captured_at
	`, sprintID, formatTime(capturedAt)); err != nil {
		return fmt.Errorf("upsert sprint report snapshot: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM tickets
		WHERE sprint_id = ? AND deleted_at IS NULL
		ORDER BY status ASC, created_at DESC, key DESC
	`, sprintID)
	if err != nil {
		return fmt.Errorf("list sprint report snapshot members: %w", err)
	}
	defer rows.Close()

	position := 0
	for rows.Next() {
		var ticketID string
		if err := rows.Scan(&ticketID); err != nil {
			return fmt.Errorf("scan sprint report snapshot member: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO sprint_report_tickets (sprint_id, ticket_id, position)
			VALUES (?, ?, ?)
		`, sprintID, ticketID, position); err != nil {
			return fmt.Errorf("insert sprint report snapshot: %w", err)
		}
		position++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate sprint report snapshot members: %w", err)
	}
	return nil
}

func (s *Service) listSprintReportTickets(ctx context.Context, sprintID string) ([]Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, created_at, updated_at, deleted_at
		FROM tickets
		WHERE sprint_id = ? AND deleted_at IS NULL
		ORDER BY status ASC, created_at DESC, key DESC
	`, sprintID)
	if err != nil {
		return nil, fmt.Errorf("list sprint report tickets: %w", err)
	}
	defer rows.Close()

	tickets := []Ticket{}
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan sprint report ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sprint report tickets: %w", err)
	}
	return tickets, nil
}

func (s *Service) listSprintSnapshotTickets(ctx context.Context, sprintID string) ([]Ticket, *time.Time, bool, error) {
	var capturedAt string
	if err := s.db.QueryRowContext(ctx, `
		SELECT captured_at
		FROM sprint_report_snapshots
		WHERE sprint_id = ?
	`, sprintID).Scan(&capturedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, false, nil
		}
		return nil, nil, false, fmt.Errorf("get sprint report snapshot: %w", err)
	}
	snapshotAt, err := parseTime(capturedAt)
	if err != nil {
		return nil, nil, false, fmt.Errorf("parse sprint report snapshot captured_at: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT tickets.id, tickets.project_id, tickets.key, tickets.title, tickets.description, tickets.status, tickets.priority, tickets.type,
			tickets.reporter_id, tickets.assignee_id, tickets.parent_ticket_id, tickets.sprint_id, tickets.component_id, tickets.version_id,
			tickets.rank, tickets.start_date, tickets.due_date, tickets.created_at, tickets.updated_at, tickets.deleted_at
		FROM sprint_report_tickets snapshot
		JOIN tickets ON tickets.id = snapshot.ticket_id
		WHERE snapshot.sprint_id = ? AND tickets.deleted_at IS NULL
		ORDER BY snapshot.position ASC
	`, sprintID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("list sprint report snapshot tickets: %w", err)
	}
	defer rows.Close()

	tickets := []Ticket{}
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, nil, false, fmt.Errorf("scan sprint report snapshot ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, fmt.Errorf("iterate sprint report snapshot tickets: %w", err)
	}
	return tickets, &snapshotAt, true, nil
}

func sprintReportProgress(tickets []Ticket) SprintReportProgress {
	progress := SprintReportProgress{ByStatus: map[string]int{}}
	for _, ticket := range tickets {
		progress.Total++
		progress.ByStatus[ticket.Status]++
		if ticket.Status == "done" {
			progress.Done++
		}
	}
	return progress
}

func (s *Service) sprintAnalytics(ctx context.Context, sprint Sprint, tickets []Ticket) (SprintAnalytics, error) {
	start, end := sprintAnalyticsWindow(sprint, tickets, s.now().UTC())
	doneDates := map[string]time.Time{}
	for _, ticket := range tickets {
		doneAt, err := s.ticketDoneAt(ctx, ticket)
		if err != nil {
			return SprintAnalytics{}, err
		}
		if doneAt != nil {
			doneDates[ticket.ID] = dateOnly(*doneAt)
		}
	}

	days := daysBetween(start, end)
	burndown := make([]SprintBurndownPoint, 0, len(days))
	burnup := make([]SprintBurnupPoint, 0, len(days))
	completed := 0
	for _, day := range days {
		total := 0
		done := 0
		for _, ticket := range tickets {
			if !dateOnly(ticket.CreatedAt).After(day) {
				total++
			}
			if doneAt, ok := doneDates[ticket.ID]; ok && !doneAt.After(day) {
				done++
			}
		}
		date := day.Format(dateOnlyLayout)
		burndown = append(burndown, SprintBurndownPoint{Date: date, Remaining: total - done})
		burnup = append(burnup, SprintBurnupPoint{Date: date, Total: total, Done: done})
		completed = done
	}
	return SprintAnalytics{
		Burndown: burndown,
		Burnup:   burnup,
		Velocity: SprintVelocity{Completed: completed, Unit: "tickets"},
	}, nil
}

func sprintAnalyticsWindow(sprint Sprint, tickets []Ticket, now time.Time) (time.Time, time.Time) {
	start := parseSprintDateOrZero(sprint.StartDate)
	if start.IsZero() {
		for _, ticket := range tickets {
			created := dateOnly(ticket.CreatedAt)
			if start.IsZero() || created.Before(start) {
				start = created
			}
		}
	}
	if start.IsZero() {
		start = dateOnly(now)
	}

	end := parseSprintDateOrZero(sprint.EndDate)
	if end.IsZero() && sprint.CompletedAt != nil {
		end = dateOnly(*sprint.CompletedAt)
	}
	if end.IsZero() {
		end = dateOnly(now)
	}
	if end.Before(start) {
		end = start
	}
	return start, end
}

func (s *Service) ticketDoneAt(ctx context.Context, ticket Ticket) (*time.Time, error) {
	if ticket.Status != "done" {
		return nil, nil
	}
	activities, err := s.repo.ListTicketActivity(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("list ticket activity for sprint analytics: %w", err)
	}
	for _, activity := range activities {
		if ticketActivityStatusNew(activity) == "done" {
			doneAt := activity.CreatedAt
			return &doneAt, nil
		}
	}
	doneAt := ticket.UpdatedAt
	return &doneAt, nil
}

func ticketActivityStatusNew(activity TicketActivity) string {
	changes, ok := activity.Data["changes"].(map[string]any)
	if !ok {
		return ""
	}
	statusChange, ok := changes["status"].(map[string]any)
	if !ok {
		return ""
	}
	next, _ := statusChange["new"].(string)
	return next
}

func parseSprintDateOrZero(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(dateOnlyLayout, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func daysBetween(start time.Time, end time.Time) []time.Time {
	days := []time.Time{}
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		days = append(days, day)
	}
	return days
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
		if err := s.repo.insertTicketActivity(ctx, tx, TicketActivity{
			ID:           activityID,
			TicketID:     updated.ID,
			ActorID:      actorID(principal),
			ActivityType: activityTicketUpdated,
			Data:         data,
			CreatedAt:    updated.UpdatedAt,
		}); err != nil {
			return err
		}
		return s.appendDomainEvent(ctx, tx, events.Event{
			Type:        activityTicketUpdated,
			ActorID:     actorID(principal),
			ProjectID:   updated.ProjectID,
			ObjectID:    updated.ID,
			SubjectType: "ticket",
			SubjectID:   updated.ID,
			RelatedType: "sprint",
			RelatedID:   sprintID,
			At:          updated.UpdatedAt,
			Data:        data,
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
