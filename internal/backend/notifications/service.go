package notifications

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
)

var (
	ErrNotFound   = errors.New("notifications: not found")
	ErrValidation = errors.New("notifications: validation failed")
	ErrDelivery   = errors.New("notifications: delivery failed")
)

type Notification struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Type        string         `json:"type"`
	SubjectType string         `json:"subject_type,omitempty"`
	SubjectID   string         `json:"subject_id,omitempty"`
	Body        string         `json:"body"`
	Data        map[string]any `json:"data"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type CreateInput struct {
	UserID      string
	Type        string
	SubjectType string
	SubjectID   string
	Body        string
	Data        map[string]any
}

type ListInput struct {
	UnreadOnly bool
	Limit      int
	Offset     int
}

type Service struct {
	db         *sql.DB
	now        func() time.Time
	eventStore *events.Store
}

type Option func(*Service)

func NewService(db *sql.DB, options ...Option) *Service {
	service := &Service{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithNow(now func() time.Time) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func WithEventStore(store *events.Store) Option {
	return func(service *Service) {
		service.eventStore = store
	}
}

func (s *Service) RegisterEventHandlers(bus *events.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe("comment.created", s.handleCommentCreated)
	bus.Subscribe("ticket.updated", s.handleTicketUpdated)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Notification, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Type = strings.TrimSpace(input.Type)
	input.SubjectType = strings.TrimSpace(input.SubjectType)
	input.SubjectID = strings.TrimSpace(input.SubjectID)
	input.Body = strings.TrimSpace(input.Body)
	fields := map[string]string{}
	if input.UserID == "" {
		fields["user_id"] = "Required"
	}
	if input.Type == "" {
		fields["type"] = "Required"
	}
	if input.Body == "" {
		fields["body"] = "Required"
	}
	if len(input.Body) > 2000 {
		fields["body"] = "Must be 2000 characters or fewer"
	}
	if len(fields) > 0 {
		return Notification{}, fmt.Errorf("%w: invalid notification", ErrValidation)
	}

	id, err := newID("notif")
	if err != nil {
		return Notification{}, err
	}
	now := s.now().UTC()
	dataJSON, err := marshalData(input.Data)
	if err != nil {
		return Notification{}, err
	}
	notification := Notification{
		ID:          id,
		UserID:      input.UserID,
		Type:        input.Type,
		SubjectType: input.SubjectType,
		SubjectID:   input.SubjectID,
		Body:        input.Body,
		Data:        nonNilMap(input.Data),
		CreatedAt:   now,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, subject_type, subject_id, body, data_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, notification.ID, notification.UserID, notification.Type, nullableString(notification.SubjectType), nullableString(notification.SubjectID), notification.Body, dataJSON, formatTime(notification.CreatedAt)); err != nil {
		return Notification{}, fmt.Errorf("insert notification: %w", err)
	}
	return notification, nil
}

func (s *Service) List(ctx context.Context, principal authz.Principal, input ListInput) ([]Notification, error) {
	if principal.UserID == "" || principal.Disabled {
		return nil, authz.ErrForbidden
	}
	limit, offset, err := normalizeList(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	where := []string{"user_id = ?"}
	args := []any{principal.UserID}
	if input.UnreadOnly {
		where = append(where, "read_at IS NULL")
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, type, subject_type, subject_id, body, data_json, read_at, created_at
		FROM notifications
		WHERE `+joinAnd(where)+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return notifications, nil
}

func (s *Service) SetRead(ctx context.Context, principal authz.Principal, notificationID string, read bool) (Notification, error) {
	if principal.UserID == "" || principal.Disabled {
		return Notification{}, authz.ErrForbidden
	}
	var readAt any
	if read {
		readAt = formatTime(s.now().UTC())
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = ?
		WHERE id = ? AND user_id = ?
	`, readAt, notificationID, principal.UserID)
	if err != nil {
		return Notification{}, fmt.Errorf("update notification read state: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Notification{}, fmt.Errorf("check notification update: %w", err)
	}
	if affected == 0 {
		return Notification{}, ErrNotFound
	}
	return s.getForUser(ctx, principal.UserID, notificationID)
}

func (s *Service) MarkAllRead(ctx context.Context, principal authz.Principal) error {
	if principal.UserID == "" || principal.Disabled {
		return authz.ErrForbidden
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, ?)
		WHERE user_id = ? AND read_at IS NULL
	`, formatTime(s.now().UTC()), principal.UserID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

func (s *Service) ProcessPendingDomainEvents(ctx context.Context, limit int) (int, error) {
	if s == nil || s.eventStore == nil {
		return 0, nil
	}
	pending, err := s.eventStore.ListPending(ctx, limit, "comment.created", "ticket.updated")
	if err != nil {
		return 0, err
	}

	processed := 0
	var firstErr error
	for _, stored := range pending {
		event := events.Event{
			Type:        stored.Type,
			ActorID:     stored.ActorID,
			ProjectID:   stored.ProjectID,
			ObjectID:    stored.ObjectID,
			SubjectType: stored.SubjectType,
			SubjectID:   stored.SubjectID,
			RelatedType: stored.RelatedType,
			RelatedID:   stored.RelatedID,
			At:          stored.At,
			Data:        stored.Data,
		}
		if err := s.handleDomainEvent(ctx, event); err != nil {
			if markErr := s.eventStore.MarkFailed(ctx, stored.ID, err); markErr != nil {
				return processed, markErr
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := s.eventStore.MarkProcessed(ctx, stored.ID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, firstErr
}

func (s *Service) handleDomainEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case "comment.created":
		return s.handleCommentCreated(ctx, event)
	case "ticket.updated":
		return s.handleTicketUpdated(ctx, event)
	default:
		return nil
	}
}

func (s *Service) handleCommentCreated(ctx context.Context, event events.Event) error {
	ticketID, _ := event.Data["ticket_id"].(string)
	if ticketID == "" {
		return nil
	}
	ticket, err := s.ticket(ctx, ticketID)
	if err != nil {
		return err
	}
	recipients := recipientSet(event.ActorID, ticket.ReporterID, ticket.AssigneeID)
	for userID := range recipients {
		if _, err := s.Create(ctx, CreateInput{
			UserID:      userID,
			Type:        "comment_added",
			SubjectType: "ticket",
			SubjectID:   ticket.ID,
			Body:        fmt.Sprintf("New comment on %s", ticket.Key),
			Data: map[string]any{
				"ticket_id":  ticket.ID,
				"ticket_key": ticket.Key,
				"comment_id": event.ObjectID,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) handleTicketUpdated(ctx context.Context, event events.Event) error {
	ticket, err := s.ticket(ctx, event.ObjectID)
	if err != nil {
		return err
	}
	if newAssignee := changeNew(event.Data, "assignee_id"); newAssignee != "" && newAssignee != event.ActorID {
		if _, err := s.Create(ctx, CreateInput{
			UserID:      newAssignee,
			Type:        "ticket_assigned",
			SubjectType: "ticket",
			SubjectID:   ticket.ID,
			Body:        fmt.Sprintf("Assigned to %s", ticket.Key),
			Data: map[string]any{
				"ticket_id":  ticket.ID,
				"ticket_key": ticket.Key,
			},
		}); err != nil {
			return err
		}
	}
	if status := changeNew(event.Data, "status"); status != "" {
		recipients := recipientSet(event.ActorID, ticket.ReporterID, ticket.AssigneeID)
		for userID := range recipients {
			if _, err := s.Create(ctx, CreateInput{
				UserID:      userID,
				Type:        "ticket_status_changed",
				SubjectType: "ticket",
				SubjectID:   ticket.ID,
				Body:        fmt.Sprintf("%s moved to %s", ticket.Key, status),
				Data: map[string]any{
					"ticket_id":  ticket.ID,
					"ticket_key": ticket.Key,
					"status":     status,
				},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

type eventTicket struct {
	ID         string
	Key        string
	ReporterID string
	AssigneeID string
}

func (s *Service) ticket(ctx context.Context, ticketID string) (eventTicket, error) {
	var ticket eventTicket
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, key, COALESCE(reporter_id, ''), COALESCE(assignee_id, '')
		FROM tickets
		WHERE id = ? AND deleted_at IS NULL
	`, ticketID).Scan(&ticket.ID, &ticket.Key, &ticket.ReporterID, &ticket.AssigneeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return eventTicket{}, ErrNotFound
		}
		return eventTicket{}, fmt.Errorf("get notification ticket: %w", err)
	}
	return ticket, nil
}

func (s *Service) getForUser(ctx context.Context, userID string, notificationID string) (Notification, error) {
	notification, err := scanNotification(s.db.QueryRowContext(ctx, `
		SELECT id, user_id, type, subject_type, subject_id, body, data_json, read_at, created_at
		FROM notifications
		WHERE id = ? AND user_id = ?
	`, notificationID, userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Notification{}, ErrNotFound
		}
		return Notification{}, fmt.Errorf("get notification: %w", err)
	}
	return notification, nil
}

func scanNotification(scanner interface{ Scan(...any) error }) (Notification, error) {
	var notification Notification
	var subjectType sql.NullString
	var subjectID sql.NullString
	var dataJSON string
	var readAt sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&subjectType,
		&subjectID,
		&notification.Body,
		&dataJSON,
		&readAt,
		&createdAt,
	); err != nil {
		return Notification{}, err
	}
	notification.SubjectType = nullString(subjectType)
	notification.SubjectID = nullString(subjectID)
	notification.Data = map[string]any{}
	if dataJSON != "" {
		if err := json.Unmarshal([]byte(dataJSON), &notification.Data); err != nil {
			return Notification{}, fmt.Errorf("decode notification data: %w", err)
		}
	}
	notification.ReadAt = parseNullableTime(readAt)
	created, err := parseTime(createdAt)
	if err != nil {
		return Notification{}, err
	}
	notification.CreatedAt = created
	return notification, nil
}

func recipientSet(actorID string, userIDs ...string) map[string]struct{} {
	recipients := map[string]struct{}{}
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" || userID == actorID {
			continue
		}
		recipients[userID] = struct{}{}
	}
	return recipients
}

func changeNew(data map[string]any, field string) string {
	changes, ok := data["changes"]
	if !ok {
		return ""
	}
	encoded, err := json.Marshal(changes)
	if err != nil {
		return ""
	}
	decoded := map[string]map[string]string{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return ""
	}
	return decoded[field]["new"]
}

func marshalData(data map[string]any) (string, error) {
	encoded, err := json.Marshal(nonNilMap(data))
	if err != nil {
		return "", fmt.Errorf("encode notification data: %w", err)
	}
	return string(encoded), nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func normalizeList(limit int, offset int) (int, int, error) {
	if limit < 0 || offset < 0 {
		return 0, 0, fmt.Errorf("%w: limit and offset must be non-negative", ErrValidation)
	}
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		return 0, 0, fmt.Errorf("%w: limit must be 200 or fewer", ErrValidation)
	}
	return limit, offset, nil
}

func joinAnd(parts []string) string {
	result := ""
	for index, part := range parts {
		if index > 0 {
			result += " AND "
		}
		result += part
	}
	return result
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate notification id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse notification time: %w", err)
	}
	return parsed, nil
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	parsed, err := parseTime(value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
