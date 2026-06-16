package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
)

const (
	DeliveryStatusQueued    = "queued"
	DeliveryStatusSending   = "sending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"
	DeliveryStatusCanceled  = "canceled"
)

type Delivery struct {
	ID                 string
	DomainEventID      string
	IdempotencyKey     string
	ScopeType          string
	ProjectID          string
	PolicyID           string
	PolicyName         string
	DestinationID      string
	DestinationName    string
	DestinationService string
	EventType          string
	SubjectType        string
	SubjectID          string
	Message            string
	Payload            map[string]any
	Status             string
	AttemptCount       int
	MaxAttempts        int
	NextAttemptAt      *time.Time
	LastAttemptAt      *time.Time
	DeliveredAt        *time.Time
	LastError          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type EnqueueDeliveryInput struct {
	DomainEventID  string
	IdempotencyKey string
	PolicyID       string
	DestinationID  string
	EventType      string
	SubjectType    string
	SubjectID      string
	Message        string
	Payload        map[string]any
	MaxAttempts    int
}

type ListDeliveriesInput struct {
	ScopeType     string
	ProjectID     string
	Status        string
	PolicyID      string
	DestinationID string
	Limit         int
	Offset        int
}

type ProcessDeliveriesInput struct {
	Limit int
}

func (s *Service) EnqueueDelivery(ctx context.Context, input EnqueueDeliveryInput) (Delivery, error) {
	policy, err := s.GetPolicy(ctx, strings.TrimSpace(input.PolicyID))
	if err != nil {
		return Delivery{}, err
	}
	destination, err := s.GetDestination(ctx, strings.TrimSpace(input.DestinationID))
	if err != nil {
		return Delivery{}, err
	}
	if err := s.validatePolicyDestinations(ctx, policy.ScopeType, policy.ProjectID, []string{destination.ID}); err != nil {
		return Delivery{}, err
	}
	delivery := Delivery{
		DomainEventID:      strings.TrimSpace(input.DomainEventID),
		IdempotencyKey:     strings.TrimSpace(input.IdempotencyKey),
		PolicyID:           policy.ID,
		PolicyName:         policy.Name,
		DestinationID:      destination.ID,
		DestinationName:    destination.Name,
		DestinationService: destination.Service,
		ScopeType:          policy.ScopeType,
		ProjectID:          policy.ProjectID,
		EventType:          strings.TrimSpace(input.EventType),
		SubjectType:        strings.TrimSpace(input.SubjectType),
		SubjectID:          strings.TrimSpace(input.SubjectID),
		Message:            strings.TrimSpace(input.Message),
		Payload:            nonNilMap(input.Payload),
		Status:             DeliveryStatusQueued,
		MaxAttempts:        input.MaxAttempts,
		CreatedAt:          s.now().UTC(),
		UpdatedAt:          s.now().UTC(),
	}
	if delivery.MaxAttempts == 0 {
		delivery.MaxAttempts = 3
	}
	if err := validateDelivery(delivery); err != nil {
		return Delivery{}, err
	}
	if delivery.IdempotencyKey != "" {
		existing, err := s.getDeliveryByIdempotencyKey(ctx, delivery.IdempotencyKey)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return Delivery{}, err
		}
		if err == nil {
			return existing, nil
		}
	}
	id, err := newID("delivery")
	if err != nil {
		return Delivery{}, err
	}
	delivery.ID = id
	payloadJSON, err := marshalMap(delivery.Payload)
	if err != nil {
		return Delivery{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_deliveries (
			id, domain_event_id, idempotency_key, scope_type, scope_key,
			project_id, policy_id, policy_name, destination_id, destination_name, destination_service,
			event_type, subject_type, subject_id, message, payload_json, status,
			attempt_count, max_attempts, next_attempt_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, delivery.ID, nullableString(delivery.DomainEventID), nullableString(delivery.IdempotencyKey),
		delivery.ScopeType, deliveryScopeKey(delivery.ScopeType, delivery.ProjectID),
		nullableString(delivery.ProjectID), delivery.PolicyID, nullableString(delivery.PolicyName),
		delivery.DestinationID, nullableString(delivery.DestinationName), nullableString(delivery.DestinationService),
		delivery.EventType, nullableString(delivery.SubjectType), nullableString(delivery.SubjectID),
		delivery.Message, payloadJSON, delivery.Status, delivery.AttemptCount, delivery.MaxAttempts,
		formatTime(delivery.CreatedAt), formatTime(delivery.CreatedAt), formatTime(delivery.UpdatedAt)); err != nil {
		return Delivery{}, fmt.Errorf("insert notification delivery: %w", err)
	}
	return delivery, nil
}

func (s *Service) ListDeliveries(ctx context.Context, input ListDeliveriesInput) ([]Delivery, error) {
	scopeType := strings.TrimSpace(input.ScopeType)
	if scopeType == "" {
		scopeType = PolicyScopeGlobal
	}
	projectID := strings.TrimSpace(input.ProjectID)
	if err := validatePolicyScope(scopeType, projectID); err != nil {
		return nil, err
	}
	limit, offset, err := normalizeList(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	where := []string{"scope_type = ?", "scope_key = ?"}
	args := []any{scopeType, deliveryScopeKey(scopeType, projectID)}
	if status := strings.TrimSpace(input.Status); status != "" {
		if !validDeliveryStatus(status) {
			return nil, fmt.Errorf("%w: invalid delivery status", ErrValidation)
		}
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if policyID := strings.TrimSpace(input.PolicyID); policyID != "" {
		where = append(where, "policy_id = ?")
		args = append(args, policyID)
	}
	if destinationID := strings.TrimSpace(input.DestinationID); destinationID != "" {
		where = append(where, "destination_id = ?")
		args = append(args, destinationID)
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scope_type, project_id, policy_id, destination_id, event_type,
			domain_event_id, idempotency_key, policy_name, destination_name, destination_service,
			subject_type, subject_id, message, payload_json, status, attempt_count,
			max_attempts, next_attempt_at, last_attempt_at, delivered_at, COALESCE(last_error, ''),
			created_at, updated_at
		FROM notification_deliveries
		WHERE `+joinAnd(where)+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list notification deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		delivery, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, delivery)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification deliveries: %w", err)
	}
	return deliveries, nil
}

func (s *Service) GetDelivery(ctx context.Context, deliveryID string) (Delivery, error) {
	delivery, err := scanDelivery(s.db.QueryRowContext(ctx, `
		SELECT id, scope_type, project_id, policy_id, destination_id, event_type,
			domain_event_id, idempotency_key, policy_name, destination_name, destination_service,
			subject_type, subject_id, message, payload_json, status, attempt_count,
			max_attempts, next_attempt_at, last_attempt_at, delivered_at, COALESCE(last_error, ''),
			created_at, updated_at
		FROM notification_deliveries
		WHERE id = ?
	`, deliveryID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Delivery{}, ErrNotFound
		}
		return Delivery{}, fmt.Errorf("get notification delivery: %w", err)
	}
	return delivery, nil
}

func (s *Service) getDeliveryByIdempotencyKey(ctx context.Context, key string) (Delivery, error) {
	delivery, err := scanDelivery(s.db.QueryRowContext(ctx, `
		SELECT id, scope_type, project_id, policy_id, destination_id, event_type,
			domain_event_id, idempotency_key, policy_name, destination_name, destination_service,
			subject_type, subject_id, message, payload_json, status, attempt_count,
			max_attempts, next_attempt_at, last_attempt_at, delivered_at, COALESCE(last_error, ''),
			created_at, updated_at
		FROM notification_deliveries
		WHERE idempotency_key = ?
	`, key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Delivery{}, ErrNotFound
		}
		return Delivery{}, fmt.Errorf("get notification delivery by idempotency key: %w", err)
	}
	return delivery, nil
}

func (s *Service) RetryDelivery(ctx context.Context, deliveryID string) (Delivery, error) {
	delivery, err := s.GetDelivery(ctx, deliveryID)
	if err != nil {
		return Delivery{}, err
	}
	if delivery.Status != DeliveryStatusFailed && delivery.Status != DeliveryStatusCanceled {
		return Delivery{}, fmt.Errorf("%w: only failed or canceled deliveries can be retried", ErrValidation)
	}
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_deliveries
		SET status = ?, next_attempt_at = ?, last_error = NULL, updated_at = ?
		WHERE id = ?
	`, DeliveryStatusQueued, formatTime(now), formatTime(now), deliveryID)
	if err != nil {
		return Delivery{}, fmt.Errorf("retry notification delivery: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Delivery{}, fmt.Errorf("check notification delivery retry: %w", err)
	}
	if affected == 0 {
		return Delivery{}, ErrNotFound
	}
	return s.GetDelivery(ctx, deliveryID)
}

func (s *Service) ProcessPendingDeliveries(ctx context.Context, input ProcessDeliveriesInput) (int, error) {
	if s == nil {
		return 0, nil
	}
	limit := input.Limit
	if limit == 0 {
		limit = 25
	}
	if limit < 0 || limit > 100 {
		return 0, fmt.Errorf("%w: delivery process limit must be between 1 and 100", ErrValidation)
	}
	now := s.now().UTC()
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM notification_deliveries
		WHERE status = ? AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
		ORDER BY next_attempt_at ASC, created_at ASC, id ASC
		LIMIT ?
	`, DeliveryStatusQueued, formatTime(now), limit)
	if err != nil {
		return 0, fmt.Errorf("list pending notification deliveries: %w", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("scan pending notification delivery: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return 0, fmt.Errorf("close pending notification deliveries: %w", err)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate pending notification deliveries: %w", err)
	}

	processed := 0
	var firstErr error
	for _, id := range ids {
		ok, err := s.claimDelivery(ctx, id)
		if err != nil {
			return processed, err
		}
		if !ok {
			continue
		}
		if err := s.processDelivery(ctx, id); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
		processed++
	}
	return processed, firstErr
}

func (s *Service) claimDelivery(ctx context.Context, deliveryID string) (bool, error) {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_deliveries
		SET status = ?, updated_at = ?
		WHERE id = ? AND status = ? AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
	`, DeliveryStatusSending, formatTime(now), deliveryID, DeliveryStatusQueued, formatTime(now))
	if err != nil {
		return false, fmt.Errorf("claim notification delivery: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check notification delivery claim: %w", err)
	}
	return affected > 0, nil
}

func (s *Service) processDelivery(ctx context.Context, deliveryID string) error {
	delivery, err := s.GetDelivery(ctx, deliveryID)
	if err != nil {
		return err
	}
	destination, rawURL, err := s.getDestinationWithSecret(ctx, delivery.DestinationID)
	if err != nil {
		failure := fmt.Errorf("destination is not available")
		if markErr := s.markDeliveryFailed(ctx, delivery, failure, true); markErr != nil {
			return markErr
		}
		return fmt.Errorf("%w: %v", ErrDelivery, failure)
	}
	if !destination.Enabled {
		failure := fmt.Errorf("destination is disabled")
		if markErr := s.markDeliveryFailed(ctx, delivery, failure, true); markErr != nil {
			return markErr
		}
		return fmt.Errorf("%w: %v", ErrDelivery, failure)
	}
	sender, err := shoutrrr.CreateSender(rawURL)
	if err != nil {
		failure := fmt.Errorf("invalid Shoutrrr service URL")
		if markErr := s.markDeliveryFailed(ctx, delivery, failure, true); markErr != nil {
			return markErr
		}
		return fmt.Errorf("%w: %v", ErrDelivery, failure)
	}
	if err := firstSendError(sender.Send(delivery.Message, nil)); err != nil {
		failure := fmt.Errorf("Shoutrrr delivery failed")
		if markErr := s.markDeliveryFailed(ctx, delivery, failure, false); markErr != nil {
			return markErr
		}
		return fmt.Errorf("%w: %v", ErrDelivery, failure)
	}
	return s.markDeliveryDelivered(ctx, delivery)
}

func (s *Service) markDeliveryDelivered(ctx context.Context, delivery Delivery) error {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_deliveries
		SET status = ?, attempt_count = attempt_count + 1, last_attempt_at = ?,
			delivered_at = ?, next_attempt_at = NULL, last_error = NULL, updated_at = ?
		WHERE id = ?
	`, DeliveryStatusDelivered, formatTime(now), formatTime(now), formatTime(now), delivery.ID)
	if err != nil {
		return fmt.Errorf("mark notification delivery delivered: %w", err)
	}
	if err := requireRowsAffected(result, "notification delivery delivered"); err != nil {
		return err
	}
	return s.updateDestinationDeliveryStatus(ctx, delivery.DestinationID, DeliveryStatusDelivered, now, "")
}

func (s *Service) markDeliveryFailed(ctx context.Context, delivery Delivery, failure error, final bool) error {
	now := s.now().UTC()
	attemptCount := delivery.AttemptCount + 1
	status := DeliveryStatusQueued
	var nextAttemptAt any = formatTime(nextDeliveryAttemptAt(now, attemptCount))
	if final || attemptCount >= delivery.MaxAttempts {
		status = DeliveryStatusFailed
		nextAttemptAt = nil
	}
	message := failure.Error()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_deliveries
		SET status = ?, attempt_count = ?, last_attempt_at = ?, next_attempt_at = ?,
			last_error = ?, updated_at = ?
		WHERE id = ?
	`, status, attemptCount, formatTime(now), nextAttemptAt, message, formatTime(now), delivery.ID)
	if err != nil {
		return fmt.Errorf("mark notification delivery failed: %w", err)
	}
	if err := requireRowsAffected(result, "notification delivery failed"); err != nil {
		return err
	}
	if status == DeliveryStatusFailed {
		return s.updateDestinationDeliveryStatus(ctx, delivery.DestinationID, DeliveryStatusFailed, now, message)
	}
	return s.updateDestinationDeliveryStatus(ctx, delivery.DestinationID, "retrying", now, message)
}

func nextDeliveryAttemptAt(now time.Time, attemptCount int) time.Time {
	switch {
	case attemptCount <= 1:
		return now.Add(1 * time.Minute)
	case attemptCount == 2:
		return now.Add(5 * time.Minute)
	default:
		return now.Add(15 * time.Minute)
	}
}

func requireRowsAffected(result sql.Result, action string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check %s update: %w", action, err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func validateDelivery(delivery Delivery) error {
	fields := map[string]string{}
	if delivery.PolicyID == "" {
		fields["policy_id"] = "Required"
	}
	if delivery.DestinationID == "" {
		fields["destination_id"] = "Required"
	}
	if delivery.EventType == "" {
		fields["event_type"] = "Required"
	}
	if delivery.Message == "" {
		fields["message"] = "Required"
	}
	if len(delivery.Message) > 4000 {
		fields["message"] = "Must be 4000 characters or fewer"
	}
	if delivery.MaxAttempts <= 0 {
		fields["max_attempts"] = "Must be greater than zero"
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid notification delivery", ErrValidation)
	}
	return nil
}

func scanDelivery(scanner interface{ Scan(...any) error }) (Delivery, error) {
	var delivery Delivery
	var projectID sql.NullString
	var domainEventID sql.NullString
	var idempotencyKey sql.NullString
	var policyName sql.NullString
	var destinationName sql.NullString
	var destinationService sql.NullString
	var subjectType sql.NullString
	var subjectID sql.NullString
	var payloadJSON string
	var nextAttemptAt sql.NullString
	var lastAttemptAt sql.NullString
	var deliveredAt sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&delivery.ID,
		&delivery.ScopeType,
		&projectID,
		&delivery.PolicyID,
		&delivery.DestinationID,
		&delivery.EventType,
		&domainEventID,
		&idempotencyKey,
		&policyName,
		&destinationName,
		&destinationService,
		&subjectType,
		&subjectID,
		&delivery.Message,
		&payloadJSON,
		&delivery.Status,
		&delivery.AttemptCount,
		&delivery.MaxAttempts,
		&nextAttemptAt,
		&lastAttemptAt,
		&deliveredAt,
		&delivery.LastError,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Delivery{}, err
	}
	payload, err := unmarshalMap(payloadJSON)
	if err != nil {
		return Delivery{}, err
	}
	created, err := parseTime(createdAt)
	if err != nil {
		return Delivery{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Delivery{}, err
	}
	delivery.ProjectID = nullString(projectID)
	delivery.DomainEventID = nullString(domainEventID)
	delivery.IdempotencyKey = nullString(idempotencyKey)
	delivery.PolicyName = nullString(policyName)
	delivery.DestinationName = nullString(destinationName)
	delivery.DestinationService = nullString(destinationService)
	delivery.SubjectType = nullString(subjectType)
	delivery.SubjectID = nullString(subjectID)
	delivery.Payload = payload
	delivery.NextAttemptAt = parseNullableTime(nextAttemptAt)
	delivery.LastAttemptAt = parseNullableTime(lastAttemptAt)
	delivery.DeliveredAt = parseNullableTime(deliveredAt)
	delivery.CreatedAt = created
	delivery.UpdatedAt = updated
	return delivery, nil
}

func deliveryScopeKey(scopeType string, projectID string) string {
	if scopeType == PolicyScopeProject {
		return projectID
	}
	return "global"
}

func validDeliveryStatus(status string) bool {
	switch status {
	case DeliveryStatusQueued, DeliveryStatusSending, DeliveryStatusDelivered, DeliveryStatusFailed, DeliveryStatusCanceled:
		return true
	default:
		return false
	}
}

func marshalMap(value map[string]any) (string, error) {
	data, err := json.Marshal(nonNilMap(value))
	if err != nil {
		return "", fmt.Errorf("marshal notification delivery payload: %w", err)
	}
	return string(data), nil
}

func unmarshalMap(value string) (map[string]any, error) {
	if value == "" {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal notification delivery payload: %w", err)
	}
	return nonNilMap(payload), nil
}
