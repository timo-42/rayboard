package tracker

import (
	"context"
	"fmt"
	"strings"

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
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, created_at, updated_at, deleted_at
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
	epics, err = s.attachTicketDetailsToTickets(ctx, epics)
	if err != nil {
		return nil, err
	}
	for index := range epics {
		items[index].Epic = epics[index]
	}
	return items, nil
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
	for _, item := range input.Items {
		startDate := strings.TrimSpace(item.StartDate)
		dueDate := strings.TrimSpace(item.DueDate)
		if _, err := s.UpdateTicket(ctx, principal, strings.TrimSpace(item.TicketID), UpdateTicketInput{
			StartDate: &startDate,
			DueDate:   &dueDate,
		}); err != nil {
			return nil, err
		}
	}
	return s.ListRoadmap(ctx, principal, projectID)
}

func (s *Service) roadmapScheduleFields(ctx context.Context, projectID string, input RoadmapScheduleInput) map[string]string {
	fields := map[string]string{}
	if len(input.Items) == 0 {
		fields["items"] = "Required"
		return fields
	}
	if len(input.Items) > 100 {
		fields["items"] = "Must contain 100 or fewer schedule updates"
	}
	seen := map[string]bool{}
	for index, item := range input.Items {
		prefix := fmt.Sprintf("items[%d].", index)
		ticketID := strings.TrimSpace(item.TicketID)
		if ticketID == "" {
			fields[prefix+"ticket_id"] = "Required"
		} else if seen[ticketID] {
			fields[prefix+"ticket_id"] = "Ticket IDs must be unique"
		} else {
			seen[ticketID] = true
			ticket, err := s.repo.GetTicket(ctx, ticketID)
			if err != nil || ticket.ProjectID != projectID || ticket.Type != "epic" {
				fields[prefix+"ticket_id"] = "Epic not found in project"
			}
		}
		startDate := strings.TrimSpace(item.StartDate)
		dueDate := strings.TrimSpace(item.DueDate)
		validateTicketDate(fields, prefix+"start_date", startDate)
		validateTicketDate(fields, prefix+"due_date", dueDate)
		if startDate != "" && dueDate != "" {
			dateFields := map[string]string{}
			validateTicketDates(dateFields, startDate, dueDate)
			if message, ok := dateFields["due_date"]; ok {
				fields[prefix+"due_date"] = message
			}
		}
	}
	return fields
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
