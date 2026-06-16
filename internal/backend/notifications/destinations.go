package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
)

const (
	DestinationScopeGlobal    = "global"
	DestinationScopeProject   = "project"
	DestinationScopeDashboard = "dashboard"
)

type Destination struct {
	ID                 string
	Name               string
	ScopeType          string
	ProjectID          string
	DashboardID        string
	Service            string
	URLSet             bool
	Enabled            bool
	LastDeliveryStatus string
	LastDeliveryAt     *time.Time
	LastError          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListDestinationsInput struct {
	ScopeType   string
	ProjectID   string
	DashboardID string
}

type CreateDestinationInput struct {
	Name        string
	ScopeType   string
	ProjectID   string
	DashboardID string
	ShoutrrrURL string
	Enabled     bool
}

type UpdateDestinationInput struct {
	Name        *string
	ShoutrrrURL *string
	Enabled     *bool
}

type TestDestinationInput struct {
	Message string
}

func (s *Service) CreateDestination(ctx context.Context, input CreateDestinationInput) (Destination, error) {
	id, err := newID("dest")
	if err != nil {
		return Destination{}, err
	}
	destination := Destination{
		ID:          id,
		Name:        normalizeDestinationName(input.Name),
		ScopeType:   strings.TrimSpace(input.ScopeType),
		ProjectID:   strings.TrimSpace(input.ProjectID),
		DashboardID: strings.TrimSpace(input.DashboardID),
		Enabled:     input.Enabled,
		CreatedAt:   s.now().UTC(),
		UpdatedAt:   s.now().UTC(),
	}
	if destination.ScopeType == "" {
		destination.ScopeType = DestinationScopeGlobal
	}
	if err := validateDestinationScope(destination.ScopeType, destination.ProjectID, destination.DashboardID); err != nil {
		return Destination{}, err
	}
	service, rawURL, err := validateShoutrrrURL(input.ShoutrrrURL)
	if err != nil {
		return Destination{}, err
	}
	destination.Service = service
	if err := validateDestination(destination, true); err != nil {
		return Destination{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_destinations (
			id, name, scope_type, scope_key, project_id, dashboard_id, service,
			shoutrrr_url_secret, enabled, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, destination.ID, destination.Name, destination.ScopeType, destinationScopeKey(destination.ScopeType, destination.ProjectID, destination.DashboardID),
		nullableString(destination.ProjectID), nullableString(destination.DashboardID), destination.Service,
		rawURL, destination.Enabled, formatTime(destination.CreatedAt), formatTime(destination.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return Destination{}, fmt.Errorf("%w: destination name already exists in scope", ErrValidation)
		}
		return Destination{}, fmt.Errorf("insert notification destination: %w", err)
	}
	destination.URLSet = true
	return destination, nil
}

func (s *Service) ListDestinations(ctx context.Context, input ListDestinationsInput) ([]Destination, error) {
	scopeType := strings.TrimSpace(input.ScopeType)
	if scopeType == "" {
		scopeType = DestinationScopeGlobal
	}
	projectID := strings.TrimSpace(input.ProjectID)
	dashboardID := strings.TrimSpace(input.DashboardID)
	if err := validateDestinationScope(scopeType, projectID, dashboardID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, scope_type, project_id, dashboard_id, service, shoutrrr_url_secret,
			enabled, COALESCE(last_delivery_status, ''), last_delivery_at,
			COALESCE(last_error, ''), created_at, updated_at
		FROM notification_destinations
		WHERE deleted_at IS NULL AND scope_type = ? AND scope_key = ?
		ORDER BY name ASC
	`, scopeType, destinationScopeKey(scopeType, projectID, dashboardID))
	if err != nil {
		return nil, fmt.Errorf("list notification destinations: %w", err)
	}
	defer rows.Close()

	var destinations []Destination
	for rows.Next() {
		destination, err := scanDestination(rows)
		if err != nil {
			return nil, err
		}
		destinations = append(destinations, destination)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification destinations: %w", err)
	}
	return destinations, nil
}

func (s *Service) GetDestination(ctx context.Context, destinationID string) (Destination, error) {
	return s.getDestination(ctx, destinationID)
}

func (s *Service) UpdateDestination(ctx context.Context, destinationID string, input UpdateDestinationInput) (Destination, error) {
	current, err := s.getDestination(ctx, destinationID)
	if err != nil {
		return Destination{}, err
	}
	rawURL := ""
	if input.Name != nil {
		current.Name = normalizeDestinationName(*input.Name)
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	if input.ShoutrrrURL != nil {
		service, value, err := validateShoutrrrURL(*input.ShoutrrrURL)
		if err != nil {
			return Destination{}, err
		}
		current.Service = service
		rawURL = value
	}
	current.UpdatedAt = s.now().UTC()
	if err := validateDestination(current, false); err != nil {
		return Destination{}, err
	}

	query := `
		UPDATE notification_destinations
		SET name = ?, service = ?, enabled = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	args := []any{current.Name, current.Service, current.Enabled, formatTime(current.UpdatedAt), destinationID}
	if input.ShoutrrrURL != nil {
		query = `
			UPDATE notification_destinations
			SET name = ?, service = ?, shoutrrr_url_secret = ?, enabled = ?, updated_at = ?
			WHERE id = ? AND deleted_at IS NULL
		`
		args = []any{current.Name, current.Service, rawURL, current.Enabled, formatTime(current.UpdatedAt), destinationID}
	}
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueConstraint(err) {
			return Destination{}, fmt.Errorf("%w: destination name already exists in scope", ErrValidation)
		}
		return Destination{}, fmt.Errorf("update notification destination: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Destination{}, fmt.Errorf("check notification destination update: %w", err)
	}
	if affected == 0 {
		return Destination{}, ErrNotFound
	}
	return s.getDestination(ctx, destinationID)
}

func (s *Service) DeleteDestination(ctx context.Context, destinationID string) error {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_destinations
		SET deleted_at = ?, enabled = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(now), formatTime(now), destinationID)
	if err != nil {
		return fmt.Errorf("delete notification destination: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check notification destination delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) TestDestination(ctx context.Context, destinationID string, input TestDestinationInput) (Destination, error) {
	destination, rawURL, err := s.getDestinationWithSecret(ctx, destinationID)
	if err != nil {
		return Destination{}, err
	}
	message := strings.TrimSpace(input.Message)
	if message == "" {
		message = "Rayboard notification destination test"
	}
	if len(message) > 1000 {
		return Destination{}, fmt.Errorf("%w: test message must be 1000 characters or fewer", ErrValidation)
	}

	now := s.now().UTC()
	sender, err := shoutrrr.CreateSender(rawURL)
	if err != nil {
		return Destination{}, fmt.Errorf("%w: invalid Shoutrrr service URL", ErrValidation)
	}
	if err := firstSendError(sender.Send(message, nil)); err != nil {
		if updateErr := s.updateDestinationDeliveryStatus(ctx, destinationID, "failed", now, "Shoutrrr delivery failed"); updateErr != nil {
			return Destination{}, updateErr
		}
		return Destination{}, fmt.Errorf("%w: Shoutrrr delivery failed", ErrDelivery)
	}
	if err := s.updateDestinationDeliveryStatus(ctx, destinationID, "delivered", now, ""); err != nil {
		return Destination{}, err
	}
	destination, err = s.getDestination(ctx, destinationID)
	if err != nil {
		return Destination{}, err
	}
	return destination, nil
}

func (s *Service) getDestination(ctx context.Context, destinationID string) (Destination, error) {
	destination, err := scanDestination(s.db.QueryRowContext(ctx, `
		SELECT id, name, scope_type, project_id, dashboard_id, service, shoutrrr_url_secret,
			enabled, COALESCE(last_delivery_status, ''), last_delivery_at,
			COALESCE(last_error, ''), created_at, updated_at
		FROM notification_destinations
		WHERE id = ? AND deleted_at IS NULL
	`, destinationID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Destination{}, ErrNotFound
		}
		return Destination{}, fmt.Errorf("get notification destination: %w", err)
	}
	return destination, nil
}

func (s *Service) getDestinationWithSecret(ctx context.Context, destinationID string) (Destination, string, error) {
	var secret string
	destination, err := scanDestinationWithSecret(s.db.QueryRowContext(ctx, `
		SELECT id, name, scope_type, project_id, dashboard_id, service, shoutrrr_url_secret,
			enabled, COALESCE(last_delivery_status, ''), last_delivery_at,
			COALESCE(last_error, ''), created_at, updated_at
		FROM notification_destinations
		WHERE id = ? AND deleted_at IS NULL
	`, destinationID), &secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Destination{}, "", ErrNotFound
		}
		return Destination{}, "", fmt.Errorf("get notification destination: %w", err)
	}
	if secret == "" {
		return Destination{}, "", fmt.Errorf("%w: destination URL is not configured", ErrValidation)
	}
	return destination, secret, nil
}

func scanDestination(scanner interface{ Scan(...any) error }) (Destination, error) {
	var secret string
	return scanDestinationWithSecret(scanner, &secret)
}

func scanDestinationWithSecret(scanner interface{ Scan(...any) error }, secret *string) (Destination, error) {
	var destination Destination
	var projectID sql.NullString
	var dashboardID sql.NullString
	var lastDeliveryAt sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&destination.ID,
		&destination.Name,
		&destination.ScopeType,
		&projectID,
		&dashboardID,
		&destination.Service,
		secret,
		&destination.Enabled,
		&destination.LastDeliveryStatus,
		&lastDeliveryAt,
		&destination.LastError,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Destination{}, err
	}
	destination.ProjectID = nullString(projectID)
	destination.DashboardID = nullString(dashboardID)
	destination.URLSet = *secret != ""
	destination.LastDeliveryAt = parseNullableTime(lastDeliveryAt)
	created, err := parseTime(createdAt)
	if err != nil {
		return Destination{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Destination{}, err
	}
	destination.CreatedAt = created
	destination.UpdatedAt = updated
	return destination, nil
}

func (s *Service) updateDestinationDeliveryStatus(ctx context.Context, destinationID string, status string, at time.Time, message string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_destinations
		SET last_delivery_status = ?, last_delivery_at = ?, last_error = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, status, formatTime(at), nullableString(message), formatTime(at), destinationID)
	if err != nil {
		return fmt.Errorf("update notification destination delivery status: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check notification destination delivery status update: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func validateDestination(destination Destination, requireURL bool) error {
	fields := map[string]string{}
	if destination.Name == "" {
		fields["name"] = "Required"
	}
	if len(destination.Name) > 80 {
		fields["name"] = "Must be 80 characters or fewer"
	}
	if destination.Service == "" {
		fields["shoutrrr_url"] = "Required"
	}
	if requireURL && !destination.URLSet && destination.Service == "" {
		fields["shoutrrr_url"] = "Required"
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid notification destination", ErrValidation)
	}
	return nil
}

func validateDestinationScope(scopeType string, projectID string, dashboardID string) error {
	switch scopeType {
	case DestinationScopeGlobal:
		if projectID != "" || dashboardID != "" {
			return fmt.Errorf("%w: global destinations cannot include project or dashboard ids", ErrValidation)
		}
	case DestinationScopeProject:
		if projectID == "" || dashboardID != "" {
			return fmt.Errorf("%w: project destinations require project_id only", ErrValidation)
		}
	case DestinationScopeDashboard:
		if projectID == "" || dashboardID == "" {
			return fmt.Errorf("%w: dashboard destinations require project_id and dashboard_id", ErrValidation)
		}
	default:
		return fmt.Errorf("%w: invalid destination scope", ErrValidation)
	}
	return nil
}

func validateShoutrrrURL(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("%w: shoutrrr_url is required", ErrValidation)
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return "", "", fmt.Errorf("%w: shoutrrr_url must be a valid service URL", ErrValidation)
	}
	if _, err := shoutrrr.CreateSender(raw); err != nil {
		return "", "", fmt.Errorf("%w: invalid Shoutrrr service URL", ErrValidation)
	}
	return strings.ToLower(parsed.Scheme), raw, nil
}

func destinationScopeKey(scopeType string, projectID string, dashboardID string) string {
	switch scopeType {
	case DestinationScopeGlobal:
		return "global"
	case DestinationScopeProject:
		return projectID
	case DestinationScopeDashboard:
		return projectID + ":" + dashboardID
	default:
		return ""
	}
}

func normalizeDestinationName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func isUniqueConstraint(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unique constraint") || strings.Contains(text, "constraint failed")
}

func firstSendError(errs []error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
