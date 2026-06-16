package tracker

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
)

type ReorderBacklogInput struct {
	TicketIDs []string `json:"ticket_ids"`
}

func (s *Service) ListBacklog(ctx context.Context, principal authz.Principal, projectID string) ([]Ticket, error) {
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
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY
			CASE WHEN rank IS NULL OR rank = '' THEN 1 ELSE 0 END ASC,
			rank ASC,
			created_at DESC,
			key DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list backlog: %w", err)
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate backlog: %w", err)
	}
	return s.attachCustomFieldsToTickets(ctx, tickets)
}

func (s *Service) ReorderBacklog(ctx context.Context, principal authz.Principal, projectID string, input ReorderBacklogInput) ([]Ticket, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionBoardsManage, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if len(input.TicketIDs) == 0 {
		return nil, validationFailed(map[string]string{"ticket_ids": "Required"})
	}
	if len(input.TicketIDs) > 500 {
		return nil, validationFailed(map[string]string{"ticket_ids": "Must contain 500 or fewer tickets"})
	}
	seen := map[string]struct{}{}
	ticketIDs := make([]string, 0, len(input.TicketIDs))
	for _, id := range input.TicketIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, validationFailed(map[string]string{"ticket_ids": "Ticket IDs must be non-empty"})
		}
		if _, ok := seen[id]; ok {
			return nil, validationFailed(map[string]string{"ticket_ids": "Ticket IDs must be unique"})
		}
		seen[id] = struct{}{}
		ticketIDs = append(ticketIDs, id)
	}

	now := s.now().UTC()
	var published []events.Event
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		for index, ticketID := range ticketIDs {
			current, err := s.repo.getTicket(ctx, tx, ticketID)
			if err != nil {
				return err
			}
			if current.ProjectID != projectID {
				return validationFailed(map[string]string{"ticket_ids": "All tickets must belong to the project"})
			}
			rank := backlogRank(index)
			if current.Rank == rank {
				continue
			}
			updated := current
			updated.Rank = rank
			updated.UpdatedAt = now
			if err := s.repo.updateTicket(ctx, tx, updated); err != nil {
				return err
			}
			activityID, err := newID("activity")
			if err != nil {
				return err
			}
			data := map[string]any{"changes": map[string]ticketFieldChange{
				"rank": {Old: current.Rank, New: updated.Rank},
			}}
			if err := s.repo.insertTicketActivity(ctx, tx, TicketActivity{
				ID:           activityID,
				TicketID:     updated.ID,
				ActorID:      actorID(principal),
				ActivityType: activityTicketUpdated,
				Data:         data,
				CreatedAt:    now,
			}); err != nil {
				return err
			}
			published = append(published, events.Event{
				Type:      activityTicketUpdated,
				ActorID:   actorID(principal),
				ProjectID: updated.ProjectID,
				ObjectID:  updated.ID,
				At:        now,
				Data:      data,
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for _, event := range published {
		s.publish(ctx, event)
	}
	return s.ListBacklog(ctx, principal, projectID)
}

func backlogRank(index int) string {
	return fmt.Sprintf("%06d", index+1)
}
