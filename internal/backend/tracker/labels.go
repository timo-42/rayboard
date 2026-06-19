package tracker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

const maxTicketLabels = 50

var labelColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

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

func (s *Service) ListProjectLabels(ctx context.Context, principal authz.Principal, projectID string) ([]ProjectLabel, error) {
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
		WITH ticket_counts AS (
			SELECT labels.label, COUNT(*) AS ticket_count
			FROM ticket_labels labels
			JOIN tickets ON tickets.id = labels.ticket_id
			WHERE tickets.project_id = ? AND tickets.deleted_at IS NULL
			GROUP BY labels.label
		),
		all_labels AS (
			SELECT label FROM project_labels WHERE project_id = ?
			UNION
			SELECT label FROM ticket_counts
		)
		SELECT
			all_labels.label,
			COALESCE(project_labels.description, ''),
			COALESCE(project_labels.color, ''),
			COALESCE(ticket_counts.ticket_count, 0),
			COALESCE(project_labels.created_at, ''),
			COALESCE(project_labels.updated_at, '')
		FROM all_labels
		LEFT JOIN project_labels ON project_labels.project_id = ? AND project_labels.label = all_labels.label
		LEFT JOIN ticket_counts ON ticket_counts.label = all_labels.label
		ORDER BY all_labels.label ASC
	`, projectID, projectID, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project labels: %w", err)
	}
	defer rows.Close()

	labels := []ProjectLabel{}
	for rows.Next() {
		label := ProjectLabel{ProjectID: projectID}
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&label.Label, &label.Description, &label.Color, &label.TicketCount, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan project label: %w", err)
		}
		if createdAt != "" {
			parsed, err := parseTime(createdAt)
			if err != nil {
				return nil, fmt.Errorf("parse project label created_at: %w", err)
			}
			label.CreatedAt = parsed
		}
		if updatedAt != "" {
			parsed, err := parseTime(updatedAt)
			if err != nil {
				return nil, fmt.Errorf("parse project label updated_at: %w", err)
			}
			label.UpdatedAt = parsed
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project labels: %w", err)
	}
	return labels, nil
}

func (s *Service) CreateProjectLabel(ctx context.Context, principal authz.Principal, input CreateProjectLabelInput) (ProjectLabel, error) {
	projectID := strings.TrimSpace(input.ProjectID)
	if projectID == "" {
		return ProjectLabel{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(projectID)); err != nil {
		return ProjectLabel{}, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return ProjectLabel{}, err
	}
	label, description, color, err := projectLabelFields(input.Label, input.Description, input.Color, true)
	if err != nil {
		return ProjectLabel{}, err
	}
	now := s.now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO project_labels (project_id, label, description, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, projectID, label, nullableString(description), nullableString(color), formatTime(now), formatTime(now)); err != nil {
		if isUniqueConstraint(err) {
			return ProjectLabel{}, conflict("project_label", "label", label)
		}
		return ProjectLabel{}, fmt.Errorf("insert project label: %w", err)
	}
	return s.getProjectLabelWithAccess(ctx, principal, projectID, label, authz.PermissionProjectsRead)
}

func (s *Service) UpdateProjectLabel(ctx context.Context, principal authz.Principal, projectID string, labelValue string, input UpdateProjectLabelInput) (ProjectLabel, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return ProjectLabel{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(projectID)); err != nil {
		return ProjectLabel{}, err
	}
	label, err := normalizeProjectLabel(labelValue)
	if err != nil {
		return ProjectLabel{}, err
	}
	current, err := s.getProjectLabel(ctx, projectID, label)
	if err != nil {
		return ProjectLabel{}, err
	}
	description := current.Description
	color := current.Color
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}
	if input.Color != nil {
		color = strings.TrimSpace(*input.Color)
	}
	_, description, color, err = projectLabelFields(label, description, color, false)
	if err != nil {
		return ProjectLabel{}, err
	}
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE project_labels
		SET description = ?, color = ?, updated_at = ?
		WHERE project_id = ? AND label = ?
	`, nullableString(description), nullableString(color), formatTime(now), projectID, label)
	if err != nil {
		return ProjectLabel{}, fmt.Errorf("update project label: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return ProjectLabel{}, fmt.Errorf("check project label update: %w", err)
	} else if affected == 0 {
		return ProjectLabel{}, notFound("project_label", label)
	}
	return s.getProjectLabelWithAccess(ctx, principal, projectID, label, authz.PermissionProjectsRead)
}

func (s *Service) DeleteProjectLabel(ctx context.Context, principal authz.Principal, projectID string, labelValue string) error {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(projectID)); err != nil {
		return err
	}
	label, err := normalizeProjectLabel(labelValue)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM project_labels WHERE project_id = ? AND label = ?", projectID, label)
	if err != nil {
		return fmt.Errorf("delete project label: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check project label delete: %w", err)
	}
	if affected == 0 {
		return notFound("project_label", label)
	}
	return nil
}

func (s *Service) getProjectLabelWithAccess(ctx context.Context, principal authz.Principal, projectID string, labelValue string, permission authz.Permission) (ProjectLabel, error) {
	if err := s.require(principal, permission, authz.ProjectScope(projectID)); err != nil {
		return ProjectLabel{}, err
	}
	return s.getProjectLabel(ctx, projectID, labelValue)
}

func (s *Service) getProjectLabel(ctx context.Context, projectID string, labelValue string) (ProjectLabel, error) {
	label, err := normalizeProjectLabel(labelValue)
	if err != nil {
		return ProjectLabel{}, err
	}
	var result ProjectLabel
	var createdAt string
	var updatedAt string
	if err := s.db.QueryRowContext(ctx, `
		WITH ticket_counts AS (
			SELECT labels.label, COUNT(*) AS ticket_count
			FROM ticket_labels labels
			JOIN tickets ON tickets.id = labels.ticket_id
			WHERE tickets.project_id = ? AND tickets.deleted_at IS NULL AND labels.label = ?
			GROUP BY labels.label
		)
		SELECT
			project_labels.project_id,
			project_labels.label,
			COALESCE(project_labels.description, ''),
			COALESCE(project_labels.color, ''),
			COALESCE(ticket_counts.ticket_count, 0),
			project_labels.created_at,
			project_labels.updated_at
		FROM project_labels
		LEFT JOIN ticket_counts ON ticket_counts.label = project_labels.label
		WHERE project_labels.project_id = ? AND project_labels.label = ?
	`, projectID, label, projectID, label).Scan(&result.ProjectID, &result.Label, &result.Description, &result.Color, &result.TicketCount, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProjectLabel{}, notFound("project_label", label)
		}
		return ProjectLabel{}, fmt.Errorf("get project label: %w", err)
	}
	result.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return ProjectLabel{}, fmt.Errorf("parse project label created_at: %w", err)
	}
	result.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return ProjectLabel{}, fmt.Errorf("parse project label updated_at: %w", err)
	}
	return result, nil
}

func projectLabelFields(labelValue string, description string, color string, requireLabel bool) (string, string, string, error) {
	label := ""
	var err error
	if requireLabel || strings.TrimSpace(labelValue) != "" {
		label, err = normalizeProjectLabel(labelValue)
		if err != nil {
			return "", "", "", err
		}
	}
	description = strings.TrimSpace(description)
	color = strings.TrimSpace(color)
	fields := map[string]string{}
	if len(description) > 500 {
		fields["description"] = "Must be 500 characters or fewer"
	}
	if color != "" && !labelColorPattern.MatchString(color) {
		fields["color"] = "Must be a #RRGGBB color"
	}
	if len(fields) > 0 {
		return "", "", "", validationFailed(fields)
	}
	return label, description, strings.ToLower(color), nil
}

func normalizeProjectLabel(input string) (string, error) {
	labels, err := normalizeTicketLabels([]string{input})
	if err != nil {
		return "", err
	}
	return labels[0], nil
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
