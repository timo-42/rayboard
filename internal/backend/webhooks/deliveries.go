package webhooks

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
)

const (
	OutgoingDeliveryStatusQueued    = "queued"
	OutgoingDeliveryStatusSending   = "sending"
	OutgoingDeliveryStatusDelivered = "delivered"
	OutgoingDeliveryStatusFailed    = "failed"
	OutgoingDeliveryStatusCanceled  = "canceled"
)

type OutgoingDelivery struct {
	ID             string
	WebhookID      string
	WebhookName    string
	DomainEventID  string
	IdempotencyKey string
	ProjectID      string
	EventType      string
	SubjectType    string
	SubjectID      string
	Payload        map[string]any
	Status         string
	AttemptCount   int
	MaxAttempts    int
	NextAttemptAt  *time.Time
	LastAttemptAt  *time.Time
	DeliveredAt    *time.Time
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *Service) EnqueueOutgoingDeliveriesForEvent(ctx context.Context, event events.StoredEvent) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("webhooks: database is required")
	}
	projectID := strings.TrimSpace(event.ProjectID)
	if projectID == "" {
		return 0, nil
	}
	hooks, err := s.enabledOutgoingWebhooks(ctx, projectID, event.Type)
	if err != nil {
		return 0, err
	}
	enqueued := 0
	for _, hook := range hooks {
		created, err := s.enqueueOutgoingDelivery(ctx, hook, event)
		if err != nil {
			return enqueued, err
		}
		if created {
			enqueued++
		}
	}
	return enqueued, nil
}

func (s *Service) ListOutgoingDeliveries(ctx context.Context, principal authz.Principal, webhookID string, limit int, offset int) ([]OutgoingDelivery, error) {
	hook, err := s.Get(ctx, principal, webhookID)
	if err != nil {
		return nil, err
	}
	limit, offset, err = normalizeList(limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, webhook_id, COALESCE(webhook_name, ''), domain_event_id, idempotency_key, project_id, event_type,
			subject_type, subject_id, payload_json, status, attempt_count, max_attempts,
			next_attempt_at, last_attempt_at, delivered_at, COALESCE(last_error, ''),
			created_at, updated_at
		FROM outgoing_webhook_deliveries
		WHERE webhook_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, hook.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list outgoing webhook deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []OutgoingDelivery
	for rows.Next() {
		delivery, err := scanOutgoingDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, delivery)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outgoing webhook deliveries: %w", err)
	}
	return deliveries, nil
}

func (s *Service) GetOutgoingDelivery(ctx context.Context, principal authz.Principal, deliveryID string) (OutgoingDelivery, error) {
	deliveryID = strings.TrimSpace(deliveryID)
	if deliveryID == "" {
		return OutgoingDelivery{}, fmt.Errorf("%w: delivery id is required", ErrValidation)
	}
	delivery, err := s.getOutgoingDelivery(ctx, deliveryID)
	if err != nil {
		return OutgoingDelivery{}, err
	}
	if delivery.WebhookID == "" {
		return OutgoingDelivery{}, ErrNotFound
	}
	if _, err := s.Get(ctx, principal, delivery.WebhookID); err != nil {
		return OutgoingDelivery{}, err
	}
	return delivery, nil
}

func (s *Service) enabledOutgoingWebhooks(ctx context.Context, projectID string, eventType string) ([]Webhook, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, direction, enabled, actor_user_id, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			event_types_json,
			token_hash, token_rotated_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhooks
		WHERE project_id = ? AND direction = ? AND enabled = 1 AND deleted_at IS NULL
		ORDER BY created_at ASC, id ASC
	`, projectID, DirectionOutgoing)
	if err != nil {
		return nil, fmt.Errorf("list enabled outgoing webhooks: %w", err)
	}
	defer rows.Close()

	var hooks []Webhook
	for rows.Next() {
		hook, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enabled outgoing webhooks: %w", err)
	}
	matched := hooks[:0]
	for _, hook := range hooks {
		if eventTypeMatches(hook.EventTypes, eventType) {
			matched = append(matched, hook)
		}
	}
	return matched, nil
}

func (s *Service) enqueueOutgoingDelivery(ctx context.Context, hook Webhook, event events.StoredEvent) (bool, error) {
	key := outgoingDeliveryIdempotencyKey(event.ID, hook.ID)
	existing, err := s.getOutgoingDeliveryByIdempotencyKey(ctx, key)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return false, err
	}
	if err == nil && existing.ID != "" {
		return false, nil
	}
	now := s.now().UTC()
	delivery := OutgoingDelivery{
		ID:             newID("outgoing_delivery"),
		WebhookID:      hook.ID,
		WebhookName:    hook.Name,
		DomainEventID:  event.ID,
		IdempotencyKey: key,
		ProjectID:      hook.ProjectID,
		EventType:      strings.TrimSpace(event.Type),
		SubjectType:    strings.TrimSpace(event.SubjectType),
		SubjectID:      strings.TrimSpace(event.SubjectID),
		Payload:        outgoingDeliveryPayload(event, hook),
		Status:         OutgoingDeliveryStatusQueued,
		MaxAttempts:    3,
		NextAttemptAt:  &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := validateOutgoingDelivery(delivery); err != nil {
		return false, err
	}
	payloadJSON, err := marshalOutgoingDeliveryPayload(delivery.Payload)
	if err != nil {
		return false, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO outgoing_webhook_deliveries (
			id, webhook_id, webhook_name, domain_event_id, idempotency_key, project_id, event_type,
			subject_type, subject_id, payload_json, status, attempt_count, max_attempts,
			next_attempt_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, delivery.ID, nullableString(delivery.WebhookID), delivery.WebhookName, nullableString(delivery.DomainEventID), nullableString(delivery.IdempotencyKey),
		delivery.ProjectID, delivery.EventType, nullableString(delivery.SubjectType), nullableString(delivery.SubjectID),
		payloadJSON, delivery.Status, delivery.AttemptCount, delivery.MaxAttempts, nullableTime(delivery.NextAttemptAt),
		formatTime(delivery.CreatedAt), formatTime(delivery.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return false, nil
		}
		return false, fmt.Errorf("insert outgoing webhook delivery: %w", err)
	}
	return true, nil
}

func (s *Service) getOutgoingDelivery(ctx context.Context, deliveryID string) (OutgoingDelivery, error) {
	delivery, err := scanOutgoingDelivery(s.db.QueryRowContext(ctx, `
		SELECT id, webhook_id, COALESCE(webhook_name, ''), domain_event_id, idempotency_key, project_id, event_type,
			subject_type, subject_id, payload_json, status, attempt_count, max_attempts,
			next_attempt_at, last_attempt_at, delivered_at, COALESCE(last_error, ''),
			created_at, updated_at
		FROM outgoing_webhook_deliveries
		WHERE id = ?
	`, deliveryID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OutgoingDelivery{}, ErrNotFound
		}
		return OutgoingDelivery{}, err
	}
	return delivery, nil
}

func (s *Service) getOutgoingDeliveryByIdempotencyKey(ctx context.Context, key string) (OutgoingDelivery, error) {
	delivery, err := scanOutgoingDelivery(s.db.QueryRowContext(ctx, `
		SELECT id, webhook_id, COALESCE(webhook_name, ''), domain_event_id, idempotency_key, project_id, event_type,
			subject_type, subject_id, payload_json, status, attempt_count, max_attempts,
			next_attempt_at, last_attempt_at, delivered_at, COALESCE(last_error, ''),
			created_at, updated_at
		FROM outgoing_webhook_deliveries
		WHERE idempotency_key = ?
	`, key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OutgoingDelivery{}, ErrNotFound
		}
		return OutgoingDelivery{}, err
	}
	return delivery, nil
}

func outgoingDeliveryPayload(event events.StoredEvent, hook Webhook) map[string]any {
	return map[string]any{
		"event": map[string]any{
			"id":           event.ID,
			"type":         event.Type,
			"actor_id":     event.ActorID,
			"project_id":   event.ProjectID,
			"subject_type": event.SubjectType,
			"subject_id":   event.SubjectID,
			"related_type": event.RelatedType,
			"related_id":   event.RelatedID,
			"occurred_at":  formatTime(event.At),
			"data":         nonNilMap(event.Data),
		},
		"webhook": map[string]any{
			"id":         hook.ID,
			"name":       hook.Name,
			"project_id": hook.ProjectID,
		},
	}
}

func eventTypeMatches(allowed []string, eventType string) bool {
	eventType = strings.TrimSpace(eventType)
	for _, item := range allowed {
		if item == eventType {
			return true
		}
	}
	return false
}

func outgoingDeliveryIdempotencyKey(domainEventID string, webhookID string) string {
	return "outgoing_webhook:" + strings.TrimSpace(domainEventID) + ":" + strings.TrimSpace(webhookID)
}

func validateOutgoingDelivery(delivery OutgoingDelivery) error {
	fields := map[string]string{}
	if delivery.WebhookID == "" {
		fields["webhook_id"] = "Required"
	}
	if delivery.WebhookName == "" {
		fields["webhook_name"] = "Required"
	}
	if delivery.ProjectID == "" {
		fields["project_id"] = "Required"
	}
	if delivery.EventType == "" {
		fields["event_type"] = "Required"
	}
	if delivery.Status == "" {
		fields["status"] = "Required"
	}
	if delivery.MaxAttempts <= 0 {
		fields["max_attempts"] = "Must be greater than zero"
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid outgoing webhook delivery", ErrValidation)
	}
	return nil
}

func scanOutgoingDelivery(scanner interface{ Scan(...any) error }) (OutgoingDelivery, error) {
	var delivery OutgoingDelivery
	var webhookID sql.NullString
	var domainEventID sql.NullString
	var idempotencyKey sql.NullString
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
		&webhookID,
		&delivery.WebhookName,
		&domainEventID,
		&idempotencyKey,
		&delivery.ProjectID,
		&delivery.EventType,
		&subjectType,
		&subjectID,
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
		return OutgoingDelivery{}, err
	}
	payload, err := unmarshalOutgoingDeliveryPayload(payloadJSON)
	if err != nil {
		return OutgoingDelivery{}, err
	}
	created, err := parseTime(createdAt)
	if err != nil {
		return OutgoingDelivery{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return OutgoingDelivery{}, err
	}
	delivery.WebhookID = nullString(webhookID)
	delivery.DomainEventID = nullString(domainEventID)
	delivery.IdempotencyKey = nullString(idempotencyKey)
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

func marshalOutgoingDeliveryPayload(value map[string]any) (string, error) {
	data, err := json.Marshal(nonNilMap(value))
	if err != nil {
		return "", fmt.Errorf("marshal outgoing webhook delivery payload: %w", err)
	}
	return string(data), nil
}

func unmarshalOutgoingDeliveryPayload(value string) (map[string]any, error) {
	if value == "" {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal outgoing webhook delivery payload: %w", err)
	}
	return nonNilMap(payload), nil
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableTime(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return formatTime(*value)
}
