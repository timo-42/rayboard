package tracker

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

const maxTicketLabels = 50

func normalizeTicketLabels(input []string) ([]string, error) {
	if len(input) > maxTicketLabels {
		return nil, validationFailed(map[string]string{"labels": "Must contain 50 or fewer labels"})
	}

	seen := map[string]struct{}{}
	labels := make([]string, 0, len(input))
	for _, value := range input {
		label := normalizeSlug(value)
		if label == "" {
			return nil, validationFailed(map[string]string{"labels": "Labels must be non-empty lowercase slugs"})
		}
		if !slugPattern.MatchString(label) {
			return nil, validationFailed(map[string]string{"labels": "Labels must be lowercase slugs"})
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels, nil
}

func (s *Service) replaceTicketLabels(ctx context.Context, q sqlRunner, ticketID string, labels []string, createdAt time.Time) error {
	if _, err := q.ExecContext(ctx, "DELETE FROM ticket_labels WHERE ticket_id = ?", ticketID); err != nil {
		return fmt.Errorf("delete ticket labels: %w", err)
	}
	for _, label := range labels {
		if _, err := q.ExecContext(ctx, `
			INSERT INTO ticket_labels (ticket_id, label, created_at)
			VALUES (?, ?, ?)
		`, ticketID, label, formatTime(createdAt)); err != nil {
			return fmt.Errorf("insert ticket label: %w", err)
		}
	}
	return nil
}

func (s *Service) loadTicketLabels(ctx context.Context, ticketID string) ([]string, error) {
	return s.loadTicketLabelsFrom(ctx, s.db, ticketID)
}

func (s *Service) loadTicketLabelsFrom(ctx context.Context, q sqlRunner, ticketID string) ([]string, error) {
	labelsByTicket, err := s.loadTicketLabelsForTicketsFrom(ctx, q, []string{ticketID})
	if err != nil {
		return nil, err
	}
	return labelsByTicket[ticketID], nil
}

func (s *Service) loadTicketLabelsForTickets(ctx context.Context, ticketIDs []string) (map[string][]string, error) {
	return s.loadTicketLabelsForTicketsFrom(ctx, s.db, ticketIDs)
}

func (s *Service) loadTicketLabelsForTicketsFrom(ctx context.Context, q sqlRunner, ticketIDs []string) (map[string][]string, error) {
	result := make(map[string][]string, len(ticketIDs))
	if len(ticketIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(ticketIDs))
	args := make([]any, len(ticketIDs))
	for index, id := range ticketIDs {
		placeholders[index] = "?"
		args[index] = id
		result[id] = nil
	}

	rows, err := q.QueryContext(ctx, `
		SELECT ticket_id, label
		FROM ticket_labels
		WHERE ticket_id IN (`+strings.Join(placeholders, ", ")+`)
		ORDER BY ticket_id ASC, label ASC
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("load ticket labels: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ticketID string
		var label string
		if err := rows.Scan(&ticketID, &label); err != nil {
			return nil, fmt.Errorf("scan ticket label: %w", err)
		}
		result[ticketID] = append(result[ticketID], label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket labels: %w", err)
	}
	return result, nil
}

func (s *Service) attachTicketLabels(ctx context.Context, ticket Ticket) (Ticket, error) {
	labels, err := s.loadTicketLabels(ctx, ticket.ID)
	if err != nil {
		return Ticket{}, err
	}
	if len(labels) > 0 {
		ticket.Labels = labels
	}
	return ticket, nil
}

func (s *Service) attachTicketLabelsToTickets(ctx context.Context, tickets []Ticket) ([]Ticket, error) {
	ids := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		ids = append(ids, ticket.ID)
	}
	labels, err := s.loadTicketLabelsForTickets(ctx, ids)
	if err != nil {
		return nil, err
	}
	for index := range tickets {
		if len(labels[tickets[index].ID]) > 0 {
			tickets[index].Labels = labels[tickets[index].ID]
		}
	}
	return tickets, nil
}

func (s *Service) attachTicketDetails(ctx context.Context, ticket Ticket) (Ticket, error) {
	var err error
	ticket, err = s.attachTicketLabels(ctx, ticket)
	if err != nil {
		return Ticket{}, err
	}
	return s.attachCustomFields(ctx, ticket)
}

func (s *Service) attachTicketDetailsToTickets(ctx context.Context, tickets []Ticket) ([]Ticket, error) {
	var err error
	tickets, err = s.attachTicketLabelsToTickets(ctx, tickets)
	if err != nil {
		return nil, err
	}
	return s.attachCustomFieldsToTickets(ctx, tickets)
}

func equalStringSlices(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
