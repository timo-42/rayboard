package tracker

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

func (s *Service) ListRoadmap(ctx context.Context, principal authz.Principal, projectID string) ([]RoadmapItem, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, story_points, created_at, updated_at, deleted_at
		FROM tickets
		WHERE project_id = ? AND type = 'epic' AND deleted_at IS NULL
		ORDER BY
			CASE WHEN start_date IS NULL OR start_date = '' THEN 1 ELSE 0 END ASC,
			start_date ASC,
			CASE WHEN due_date IS NULL OR due_date = '' THEN 1 ELSE 0 END ASC,
			due_date ASC,
			created_at DESC,
			key DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list roadmap epics: %w", err)
	}
	defer rows.Close()

	items := []RoadmapItem{}
	for rows.Next() {
		epic, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, RoadmapItem{Epic: epic})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roadmap epics: %w", err)
	}

	for index := range items {
		progress, err := s.roadmapProgress(ctx, items[index].Epic.ID)
		if err != nil {
			return nil, err
		}
		items[index].Progress = progress
	}

	epics := make([]Ticket, len(items))
	for index := range items {
		epics[index] = items[index].Epic
	}
	epics, err = s.attachTicketDetailsAndWatcherStatus(ctx, principal, epics)
	if err != nil {
		return nil, err
	}
	for index := range epics {
		items[index].Epic = epics[index]
	}
	return items, nil
}

func (s *Service) ListRoadmapDependencies(ctx context.Context, principal authz.Principal, projectID string) ([]RoadmapDependency, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		WITH roadmap_epics AS (
			SELECT id
			FROM tickets
			WHERE project_id = ? AND type = 'epic' AND deleted_at IS NULL
		),
		roadmap_tickets AS (
			SELECT id, id AS epic_id
			FROM tickets
			WHERE project_id = ? AND type = 'epic' AND deleted_at IS NULL
			UNION
			SELECT tickets.id, tickets.parent_ticket_id AS epic_id
			FROM tickets
			JOIN roadmap_epics ON roadmap_epics.id = tickets.parent_ticket_id
			WHERE tickets.project_id = ? AND tickets.deleted_at IS NULL
		)
		SELECT links.id, source_scope.epic_id, target_scope.epic_id
		FROM ticket_links links
		JOIN roadmap_tickets source_scope ON source_scope.id = links.source_ticket_id
		JOIN roadmap_tickets target_scope ON target_scope.id = links.target_ticket_id
		WHERE links.deleted_at IS NULL
		ORDER BY links.created_at ASC, links.id ASC
	`, projectID, projectID, projectID)
	if err != nil {
		return nil, fmt.Errorf("list roadmap dependency ids: %w", err)
	}
	defer rows.Close()

	dependencies := []RoadmapDependency{}
	for rows.Next() {
		var linkID string
		var sourceEpicID sql.NullString
		var targetEpicID sql.NullString
		if err := rows.Scan(&linkID, &sourceEpicID, &targetEpicID); err != nil {
			return nil, fmt.Errorf("scan roadmap dependency id: %w", err)
		}
		link, err := s.repo.getTicketLink(ctx, linkID)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, RoadmapDependency{
			Link:         link,
			SourceEpicID: nullString(sourceEpicID),
			TargetEpicID: nullString(targetEpicID),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roadmap dependency ids: %w", err)
	}

	tickets := make([]Ticket, 0, len(dependencies)*2)
	for _, dependency := range dependencies {
		tickets = append(tickets, dependency.Link.Source, dependency.Link.Target)
	}
	tickets, err = s.attachTicketDetailsAndWatcherStatus(ctx, principal, tickets)
	if err != nil {
		return nil, err
	}
	for index := range dependencies {
		dependencies[index].Link.Source = tickets[index*2]
		dependencies[index].Link.Target = tickets[index*2+1]
	}
	return dependencies, nil
}

func (s *Service) ScheduleRoadmap(ctx context.Context, principal authz.Principal, projectID string, input RoadmapScheduleInput) ([]RoadmapItem, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsWrite, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	if fields := s.roadmapScheduleFields(ctx, projectID, input); len(fields) > 0 {
		return nil, validationFailed(fields)
	}
	startDate := strings.TrimSpace(input.StartDate)
	dueDate := strings.TrimSpace(input.DueDate)
	if _, err := s.UpdateTicket(ctx, principal, strings.TrimSpace(input.TicketID), UpdateTicketInput{
		StartDate: &startDate,
		DueDate:   &dueDate,
	}); err != nil {
		return nil, err
	}
	return s.ListRoadmap(ctx, principal, projectID)
}

func (s *Service) ListRoadmapCapacityTargets(ctx context.Context, principal authz.Principal, projectID string) ([]RoadmapCapacityTarget, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT project_id, month, target_points, created_at, updated_at
		FROM roadmap_capacity_targets
		WHERE project_id = ?
		ORDER BY month ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list roadmap capacity targets: %w", err)
	}
	defer rows.Close()

	targets := []RoadmapCapacityTarget{}
	for rows.Next() {
		target, err := scanRoadmapCapacityTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roadmap capacity targets: %w", err)
	}
	return targets, nil
}

func (s *Service) SetRoadmapCapacityTarget(ctx context.Context, principal authz.Principal, projectID string, input RoadmapCapacityTargetInput) (RoadmapCapacityTarget, bool, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return RoadmapCapacityTarget{}, false, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsWrite, authz.ProjectScope(projectID)); err != nil {
		return RoadmapCapacityTarget{}, false, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return RoadmapCapacityTarget{}, false, err
	}
	month := strings.TrimSpace(input.Month)
	fields := validateRoadmapCapacityTarget(month, input.TargetPoints)
	if len(fields) > 0 {
		return RoadmapCapacityTarget{}, false, validationFailed(fields)
	}

	now := s.now()
	if input.TargetPoints <= 0 {
		if _, err := s.db.ExecContext(ctx, `
			DELETE FROM roadmap_capacity_targets
			WHERE project_id = ? AND month = ?
		`, projectID, month); err != nil {
			return RoadmapCapacityTarget{}, false, fmt.Errorf("delete roadmap capacity target: %w", err)
		}
		return RoadmapCapacityTarget{ProjectID: projectID, Month: month}, true, nil
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO roadmap_capacity_targets (project_id, month, target_points, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(project_id, month) DO UPDATE SET
			target_points = excluded.target_points,
			updated_at = excluded.updated_at
	`, projectID, month, input.TargetPoints, formatTime(now), formatTime(now)); err != nil {
		return RoadmapCapacityTarget{}, false, fmt.Errorf("set roadmap capacity target: %w", err)
	}
	target, err := getRoadmapCapacityTarget(ctx, s.db, projectID, month)
	if err != nil {
		return RoadmapCapacityTarget{}, false, err
	}
	return target, false, nil
}

func (s *Service) roadmapScheduleFields(ctx context.Context, projectID string, input RoadmapScheduleInput) map[string]string {
	fields := map[string]string{}
	ticketID := strings.TrimSpace(input.TicketID)
	if ticketID == "" {
		fields["ticket_id"] = "Required"
	} else {
		ticket, err := s.repo.GetTicket(ctx, ticketID)
		if err != nil || ticket.ProjectID != projectID || ticket.Type != "epic" {
			fields["ticket_id"] = "Epic not found in project"
		}
	}
	startDate := strings.TrimSpace(input.StartDate)
	dueDate := strings.TrimSpace(input.DueDate)
	validateTicketDate(fields, "start_date", startDate)
	validateTicketDate(fields, "due_date", dueDate)
	if startDate != "" && dueDate != "" {
		dateFields := map[string]string{}
		validateTicketDates(dateFields, startDate, dueDate)
		if message, ok := dateFields["due_date"]; ok {
			fields["due_date"] = message
		}
	}
	return fields
}

func validateRoadmapCapacityTarget(month string, targetPoints float64) map[string]string {
	fields := map[string]string{}
	if month == "" {
		fields["month"] = "Required"
	} else if _, err := time.Parse("2006-01", month); err != nil {
		fields["month"] = "Use YYYY-MM"
	}
	if targetPoints < 0 {
		fields["target_points"] = "Must be greater than or equal to zero"
	}
	return fields
}

func getRoadmapCapacityTarget(ctx context.Context, q sqlRunner, projectID string, month string) (RoadmapCapacityTarget, error) {
	target, err := scanRoadmapCapacityTarget(q.QueryRowContext(ctx, `
		SELECT project_id, month, target_points, created_at, updated_at
		FROM roadmap_capacity_targets
		WHERE project_id = ? AND month = ?
	`, projectID, month))
	if err != nil {
		if err == sql.ErrNoRows {
			return RoadmapCapacityTarget{}, notFound("roadmap_capacity_target", month)
		}
		return RoadmapCapacityTarget{}, fmt.Errorf("get roadmap capacity target: %w", err)
	}
	return target, nil
}

func scanRoadmapCapacityTarget(row rowScanner) (RoadmapCapacityTarget, error) {
	var target RoadmapCapacityTarget
	var createdAt string
	var updatedAt string
	if err := row.Scan(&target.ProjectID, &target.Month, &target.TargetPoints, &createdAt, &updatedAt); err != nil {
		return RoadmapCapacityTarget{}, err
	}
	var err error
	target.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return RoadmapCapacityTarget{}, fmt.Errorf("parse roadmap capacity target created_at: %w", err)
	}
	target.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return RoadmapCapacityTarget{}, fmt.Errorf("parse roadmap capacity target updated_at: %w", err)
	}
	return target, nil
}

func (s *Service) roadmapProgress(ctx context.Context, epicID string) (RoadmapProgress, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM tickets
		WHERE parent_ticket_id = ? AND deleted_at IS NULL
		GROUP BY status
		ORDER BY status ASC
	`, epicID)
	if err != nil {
		return RoadmapProgress{}, fmt.Errorf("load roadmap progress: %w", err)
	}
	defer rows.Close()

	progress := RoadmapProgress{ByStatus: map[string]int{}}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return RoadmapProgress{}, fmt.Errorf("scan roadmap progress: %w", err)
		}
		progress.ByStatus[status] = count
		progress.Total += count
		if status == "done" {
			progress.Done += count
		}
	}
	if err := rows.Err(); err != nil {
		return RoadmapProgress{}, fmt.Errorf("iterate roadmap progress: %w", err)
	}
	return progress, nil
}
