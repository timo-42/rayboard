package webhooks

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	lua "github.com/yuin/gopher-lua"
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

type ProcessDeliveriesInput struct {
	Limit int
}

type outgoingRequest struct {
	Method  string
	Path    string
	Query   map[string]string
	Headers map[string]string
	Body    any
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

func (s *Service) ProcessPendingDomainEvents(ctx context.Context, eventStore *events.Store, limit int) (int, error) {
	if s == nil || eventStore == nil {
		return 0, nil
	}
	pending, err := eventStore.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	var firstErr error
	for _, event := range pending {
		if _, err := s.EnqueueOutgoingDeliveriesForEvent(ctx, event); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		processed++
	}
	return processed, firstErr
}

func (s *Service) RetryOutgoingDelivery(ctx context.Context, principal authz.Principal, deliveryID string) (OutgoingDelivery, error) {
	delivery, err := s.GetOutgoingDelivery(ctx, principal, deliveryID)
	if err != nil {
		return OutgoingDelivery{}, err
	}
	if delivery.Status != OutgoingDeliveryStatusFailed && delivery.Status != OutgoingDeliveryStatusCanceled {
		return OutgoingDelivery{}, fmt.Errorf("%w: only failed or canceled deliveries can be retried", ErrValidation)
	}
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE outgoing_webhook_deliveries
		SET status = ?, next_attempt_at = ?, last_error = NULL, updated_at = ?
		WHERE id = ?
	`, OutgoingDeliveryStatusQueued, formatTime(now), formatTime(now), delivery.ID)
	if err != nil {
		return OutgoingDelivery{}, fmt.Errorf("retry outgoing webhook delivery: %w", err)
	}
	if err := requireRowsAffected(result, "outgoing webhook delivery retry"); err != nil {
		return OutgoingDelivery{}, err
	}
	return s.GetOutgoingDelivery(ctx, principal, delivery.ID)
}

func (s *Service) ProcessPendingDeliveries(ctx context.Context, input ProcessDeliveriesInput) (int, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	limit := input.Limit
	if limit == 0 {
		limit = 25
	}
	if limit < 0 || limit > 100 {
		return 0, fmt.Errorf("%w: delivery process limit must be between 1 and 100", ErrValidation)
	}
	if err := s.requeueStaleOutgoingDeliveries(ctx); err != nil {
		return 0, err
	}
	now := s.now().UTC()
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM outgoing_webhook_deliveries
		WHERE status = ? AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
		ORDER BY next_attempt_at ASC, created_at ASC, id ASC
		LIMIT ?
	`, OutgoingDeliveryStatusQueued, formatTime(now), limit)
	if err != nil {
		return 0, fmt.Errorf("list pending outgoing webhook deliveries: %w", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("scan pending outgoing webhook delivery: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return 0, fmt.Errorf("close pending outgoing webhook deliveries: %w", err)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate pending outgoing webhook deliveries: %w", err)
	}

	processed := 0
	var firstErr error
	for _, id := range ids {
		ok, err := s.claimOutgoingDelivery(ctx, id)
		if err != nil {
			return processed, err
		}
		if !ok {
			continue
		}
		if err := s.processOutgoingDelivery(ctx, id); err != nil && firstErr == nil {
			firstErr = err
		}
		processed++
	}
	return processed, firstErr
}

func (s *Service) claimOutgoingDelivery(ctx context.Context, deliveryID string) (bool, error) {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE outgoing_webhook_deliveries
		SET status = ?, updated_at = ?
		WHERE id = ? AND status = ? AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
	`, OutgoingDeliveryStatusSending, formatTime(now), deliveryID, OutgoingDeliveryStatusQueued, formatTime(now))
	if err != nil {
		return false, fmt.Errorf("claim outgoing webhook delivery: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check outgoing webhook delivery claim: %w", err)
	}
	return affected > 0, nil
}

func (s *Service) requeueStaleOutgoingDeliveries(ctx context.Context) error {
	now := s.now().UTC()
	staleBefore := now.Add(-15 * time.Minute)
	_, err := s.db.ExecContext(ctx, `
		UPDATE outgoing_webhook_deliveries
		SET status = ?, next_attempt_at = ?, last_error = ?, updated_at = ?
		WHERE status = ? AND updated_at <= ?
	`, OutgoingDeliveryStatusQueued, formatTime(now), "requeued after stale sending state", formatTime(now),
		OutgoingDeliveryStatusSending, formatTime(staleBefore))
	if err != nil {
		return fmt.Errorf("requeue stale outgoing webhook deliveries: %w", err)
	}
	return nil
}

func (s *Service) processOutgoingDelivery(ctx context.Context, deliveryID string) error {
	delivery, err := s.getOutgoingDelivery(ctx, deliveryID)
	if err != nil {
		return err
	}
	hook, err := s.get(ctx, delivery.WebhookID)
	if err != nil {
		failure := fmt.Errorf("webhook definition is not available")
		if markErr := s.markOutgoingDeliveryFailed(ctx, delivery, failure, true); markErr != nil {
			return markErr
		}
		return failure
	}
	if hook.Direction != DirectionOutgoing || !hook.Enabled {
		failure := fmt.Errorf("outgoing webhook is disabled")
		if markErr := s.markOutgoingDeliveryFailed(ctx, delivery, failure, true); markErr != nil {
			return markErr
		}
		return failure
	}
	if err := s.requireActiveActor(ctx, hook.ActorUserID); err != nil {
		if markErr := s.markOutgoingDeliveryFailed(ctx, delivery, err, true); markErr != nil {
			return markErr
		}
		return err
	}
	request, err := s.shapeOutgoingRequest(ctx, hook, delivery)
	if err != nil {
		if markErr := s.markOutgoingDeliveryFailed(ctx, delivery, err, true); markErr != nil {
			return markErr
		}
		_ = s.recordRunResult(ctx, hook.ID, "failed", err.Error())
		return err
	}
	if err := s.sendOutgoingRequest(ctx, request); err != nil {
		if markErr := s.markOutgoingDeliveryFailed(ctx, delivery, err, false); markErr != nil {
			return markErr
		}
		_ = s.recordRunResult(ctx, hook.ID, "failed", err.Error())
		return err
	}
	if err := s.markOutgoingDeliveryDelivered(ctx, delivery); err != nil {
		return err
	}
	return s.recordRunResult(ctx, hook.ID, "succeeded", "")
}

func (s *Service) shapeOutgoingRequest(ctx context.Context, hook Webhook, delivery OutgoingDelivery) (outgoingRequest, error) {
	switch hook.Engine.Type {
	case EngineTypeLua:
		return s.shapeOutgoingLua(ctx, hook, delivery)
	case EngineTypeAI:
		return s.shapeOutgoingAI(ctx, hook, delivery)
	default:
		return outgoingRequest{}, fmt.Errorf("%w: unsupported engine", ErrValidation)
	}
}

func (s *Service) shapeOutgoingLua(ctx context.Context, hook Webhook, delivery OutgoingDelivery) (outgoingRequest, error) {
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	sandbox := luasandbox.New(luasandbox.DefaultJSONLimits())
	defer sandbox.Close()
	sandbox.L.SetContext(runCtx)

	payload := nonNilMap(delivery.Payload)
	eventValue, err := sandbox.JSON.FromGo(nonNilMapFromAny(payload["event"]))
	if err != nil {
		return outgoingRequest{}, err
	}
	webhookValue, err := sandbox.JSON.FromGo(nonNilMapFromAny(payload["webhook"]))
	if err != nil {
		return outgoingRequest{}, err
	}
	deliveryValue, err := sandbox.JSON.FromGo(outgoingDeliveryContext(delivery))
	if err != nil {
		return outgoingRequest{}, err
	}
	sandbox.L.SetGlobal("event", eventValue)
	sandbox.L.SetGlobal("webhook", webhookValue)
	sandbox.L.SetGlobal("delivery", deliveryValue)

	fn, err := sandbox.L.LoadString(hook.Engine.Script)
	if err != nil {
		return outgoingRequest{}, err
	}
	top := sandbox.L.GetTop()
	sandbox.L.Push(fn)
	if err := sandbox.L.PCall(0, lua.MultRet, nil); err != nil {
		return outgoingRequest{}, err
	}
	if sandbox.L.GetTop() <= top {
		return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook script must return a request object", ErrValidation)
	}
	value, err := sandbox.JSON.ToGo(sandbox.L.Get(-1))
	if err != nil {
		return outgoingRequest{}, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook script must return an object", ErrValidation)
	}
	return outgoingRequestFromMap(object)
}

func (s *Service) shapeOutgoingAI(ctx context.Context, hook Webhook, delivery OutgoingDelivery) (outgoingRequest, error) {
	if s.openrouter == nil {
		return outgoingRequest{}, fmt.Errorf("%w: OpenRouter service is not configured", ErrValidation)
	}
	prompt, err := outgoingAIPrompt(hook, delivery)
	if err != nil {
		return outgoingRequest{}, err
	}
	result, err := s.openrouter.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: hook.Engine.ProviderID,
		Prompt:     prompt,
	})
	if err != nil {
		return outgoingRequest{}, err
	}
	request, err := outgoingRequestFromMap(result.Output)
	if err != nil {
		return outgoingRequest{}, err
	}
	return request, nil
}

func outgoingAIPrompt(hook Webhook, delivery OutgoingDelivery) (string, error) {
	payload := map[string]any{
		"context": map[string]any{
			"direction":   hook.Direction,
			"project_id":  hook.ProjectID,
			"webhook_id":  hook.ID,
			"delivery_id": delivery.ID,
			"user_id":     hook.ActorUserID,
		},
		"event":    nonNilMapFromAny(delivery.Payload["event"]),
		"webhook":  nonNilMapFromAny(delivery.Payload["webhook"]),
		"delivery": outgoingDeliveryContext(delivery),
		"instructions": []string{
			"Return only a JSON object describing one outbound HTTP request.",
			"Allowed fields are method, path, query, headers, and body.",
			"path must be a relative URL path beginning with /; do not return scheme, host, userinfo, or credentials.",
			"Allowed methods are POST, PUT, PATCH, DELETE, and GET.",
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode outgoing webhook AI input: %w", err)
	}
	return strings.TrimSpace(hook.Engine.Prompt) + "\n\nRayboard outgoing webhook input:\n" + string(data), nil
}

func outgoingRequestFromMap(input map[string]any) (outgoingRequest, error) {
	request := outgoingRequest{
		Method:  strings.ToUpper(strings.TrimSpace(stringValue(input, "method"))),
		Path:    strings.TrimSpace(stringValue(input, "path")),
		Query:   stringMapValue(input, "query"),
		Headers: stringMapValue(input, "headers"),
		Body:    input["body"],
	}
	if request.Method == "" {
		request.Method = http.MethodPost
	}
	switch request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodGet:
	default:
		return outgoingRequest{}, fmt.Errorf("%w: unsupported outgoing webhook method", ErrValidation)
	}
	if request.Path == "" {
		return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook path is required", ErrValidation)
	}
	parsed, err := url.Parse(request.Path)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.User != nil || !strings.HasPrefix(parsed.Path, "/") {
		return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook path must be a relative path", ErrValidation)
	}
	if parsed.RawQuery != "" {
		return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook query must use the query object", ErrValidation)
	}
	for key := range request.Headers {
		normalized := http.CanonicalHeaderKey(strings.TrimSpace(key))
		switch strings.ToLower(normalized) {
		case "host", "content-length":
			return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook header %q is not allowed", ErrValidation, key)
		}
		if strings.ContainsAny(key, "\r\n") || strings.ContainsAny(request.Headers[key], "\r\n") {
			return outgoingRequest{}, fmt.Errorf("%w: outgoing webhook headers must not contain newlines", ErrValidation)
		}
	}
	return request, nil
}

func (s *Service) sendOutgoingRequest(ctx context.Context, shaped outgoingRequest) error {
	base, err := url.Parse(strings.TrimSpace(s.outgoingBaseURL))
	if err != nil || base.Scheme == "" || base.Host == "" || base.User != nil {
		return fmt.Errorf("%w: outgoing webhook base URL is not configured", ErrValidation)
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return fmt.Errorf("%w: outgoing webhook base URL must use http or https", ErrValidation)
	}
	relative, err := url.Parse(shaped.Path)
	if err != nil {
		return fmt.Errorf("%w: outgoing webhook path is invalid", ErrValidation)
	}
	target := base.ResolveReference(relative)
	query := target.Query()
	for key, value := range shaped.Query {
		query.Set(key, value)
	}
	target.RawQuery = query.Encode()

	body, err := json.Marshal(shaped.Body)
	if err != nil {
		return fmt.Errorf("encode outgoing webhook body: %w", err)
	}
	if shaped.Body == nil || shaped.Method == http.MethodGet {
		body = nil
	}
	if len(body) > 1<<20 {
		return fmt.Errorf("%w: outgoing webhook body exceeds 1048576 bytes", ErrValidation)
	}
	request, err := http.NewRequestWithContext(ctx, shaped.Method, target.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create outgoing webhook request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range shaped.Headers {
		request.Header.Set(key, value)
	}
	resp, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send outgoing webhook request: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 256<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: outgoing webhook returned HTTP %d", ErrDelivery, resp.StatusCode)
	}
	return nil
}

func (s *Service) markOutgoingDeliveryDelivered(ctx context.Context, delivery OutgoingDelivery) error {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE outgoing_webhook_deliveries
		SET status = ?, attempt_count = attempt_count + 1, last_attempt_at = ?,
			delivered_at = ?, next_attempt_at = NULL, last_error = NULL, updated_at = ?
		WHERE id = ?
	`, OutgoingDeliveryStatusDelivered, formatTime(now), formatTime(now), formatTime(now), delivery.ID)
	if err != nil {
		return fmt.Errorf("mark outgoing webhook delivery delivered: %w", err)
	}
	return requireRowsAffected(result, "outgoing webhook delivery delivered")
}

func (s *Service) markOutgoingDeliveryFailed(ctx context.Context, delivery OutgoingDelivery, failure error, final bool) error {
	now := s.now().UTC()
	attemptCount := delivery.AttemptCount + 1
	status := OutgoingDeliveryStatusQueued
	var nextAttemptAt any = formatTime(nextOutgoingDeliveryAttemptAt(now, attemptCount))
	if final || attemptCount >= delivery.MaxAttempts {
		status = OutgoingDeliveryStatusFailed
		nextAttemptAt = nil
	}
	message := failure.Error()
	result, err := s.db.ExecContext(ctx, `
		UPDATE outgoing_webhook_deliveries
		SET status = ?, attempt_count = ?, last_attempt_at = ?, next_attempt_at = ?,
			last_error = ?, updated_at = ?
		WHERE id = ?
	`, status, attemptCount, formatTime(now), nextAttemptAt, message, formatTime(now), delivery.ID)
	if err != nil {
		return fmt.Errorf("mark outgoing webhook delivery failed: %w", err)
	}
	return requireRowsAffected(result, "outgoing webhook delivery failed")
}

func nextOutgoingDeliveryAttemptAt(now time.Time, attemptCount int) time.Time {
	switch {
	case attemptCount <= 1:
		return now.Add(1 * time.Minute)
	case attemptCount == 2:
		return now.Add(5 * time.Minute)
	default:
		return now.Add(15 * time.Minute)
	}
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

func outgoingDeliveryContext(delivery OutgoingDelivery) map[string]any {
	return map[string]any{
		"id":              delivery.ID,
		"webhook_id":      delivery.WebhookID,
		"domain_event_id": delivery.DomainEventID,
		"event_type":      delivery.EventType,
		"subject_type":    delivery.SubjectType,
		"subject_id":      delivery.SubjectID,
		"attempt_count":   delivery.AttemptCount,
		"max_attempts":    delivery.MaxAttempts,
	}
}

func nonNilMapFromAny(value any) map[string]any {
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return map[string]any{}
	}
	return object
}

func stringMapValue(input map[string]any, key string) map[string]string {
	value, ok := input[key]
	if !ok || value == nil {
		return nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(object))
	for itemKey, itemValue := range object {
		text, ok := itemValue.(string)
		if !ok {
			continue
		}
		result[itemKey] = text
	}
	return result
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
