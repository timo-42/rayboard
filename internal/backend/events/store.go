package events

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type SQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type Store struct {
	db  *sql.DB
	now func() time.Time
}

type StoredEvent struct {
	ID          string
	Type        string
	ActorID     string
	ProjectID   string
	ObjectID    string
	SubjectType string
	SubjectID   string
	RelatedType string
	RelatedID   string
	At          time.Time
	Data        map[string]any
}

type StoreOption func(*Store)

func NewStore(db *sql.DB, options ...StoreOption) *Store {
	store := &Store{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func WithNow(now func() time.Time) StoreOption {
	return func(store *Store) {
		if now != nil {
			store.now = now
		}
	}
}

func (s *Store) Append(ctx context.Context, exec SQLExecutor, event Event) error {
	if s == nil {
		return errors.New("events: nil store")
	}
	if exec == nil {
		if s.db == nil {
			return errors.New("events: database is required")
		}
		exec = s.db
	}
	event.Type = strings.TrimSpace(event.Type)
	if event.Type == "" {
		return errors.New("events: event type is required")
	}
	subjectType, subjectID := eventSubject(event)
	if subjectType == "" {
		return errors.New("events: subject type is required")
	}
	if subjectID == "" {
		return errors.New("events: subject id is required")
	}
	occurredAt := event.At
	if occurredAt.IsZero() {
		occurredAt = s.now().UTC()
	}
	payload := event.Data
	if payload == nil {
		payload = map[string]any{}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode domain event payload: %w", err)
	}
	id, err := newID("event")
	if err != nil {
		return err
	}
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO domain_events (
			id, event_type, actor_id, project_id, subject_type, subject_id,
			related_type, related_id, payload_json, occurred_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, event.Type, nullableString(event.ActorID), nullableString(event.ProjectID), subjectType, subjectID,
		nullableString(event.RelatedType), nullableString(event.RelatedID), string(encoded), formatTime(occurredAt)); err != nil {
		return fmt.Errorf("insert domain event: %w", err)
	}
	return nil
}

func (s *Store) ListPending(ctx context.Context, limit int, eventTypes ...string) ([]StoredEvent, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("events: database is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	where := []string{"processing_status = 'pending'", "(next_attempt_at IS NULL OR next_attempt_at <= ?)"}
	args := []any{formatTime(s.now().UTC())}
	if len(eventTypes) > 0 {
		placeholders := make([]string, 0, len(eventTypes))
		for _, eventType := range eventTypes {
			eventType = strings.TrimSpace(eventType)
			if eventType == "" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, eventType)
		}
		if len(placeholders) > 0 {
			where = append(where, "event_type IN ("+strings.Join(placeholders, ", ")+")")
		}
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, event_type, actor_id, project_id, subject_type, subject_id,
			related_type, related_id, payload_json, occurred_at
		FROM domain_events
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY occurred_at ASC, created_at ASC, id ASC
		LIMIT ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list pending domain events: %w", err)
	}
	defer rows.Close()

	var events []StoredEvent
	for rows.Next() {
		event, err := scanStoredEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending domain events: %w", err)
	}
	return events, nil
}

func (s *Store) MarkProcessed(ctx context.Context, eventID string) error {
	if s == nil || s.db == nil {
		return errors.New("events: database is required")
	}
	return s.mark(ctx, eventID, "processed", s.now().UTC(), nil)
}

func (s *Store) MarkFailed(ctx context.Context, eventID string, cause error) error {
	if s == nil || s.db == nil {
		return errors.New("events: database is required")
	}
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	return s.mark(ctx, eventID, "failed", time.Time{}, &message)
}

func (s *Store) mark(ctx context.Context, eventID string, status string, processedAt time.Time, lastError *string) error {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return errors.New("events: event id is required")
	}
	var processed any
	if !processedAt.IsZero() {
		processed = formatTime(processedAt)
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE domain_events
		SET processing_status = ?,
			attempts = attempts + 1,
			processed_at = ?,
			last_error = ?
		WHERE id = ?
	`, status, processed, nullableStringPtr(lastError), eventID)
	if err != nil {
		return fmt.Errorf("mark domain event %s: %w", status, err)
	}
	return nil
}

func eventSubject(event Event) (string, string) {
	subjectType := strings.TrimSpace(event.SubjectType)
	subjectID := strings.TrimSpace(event.SubjectID)
	if subjectID == "" {
		subjectID = strings.TrimSpace(event.ObjectID)
	}
	if subjectType == "" {
		subjectType = inferSubjectType(event.Type)
	}
	return subjectType, subjectID
}

func inferSubjectType(eventType string) string {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return ""
	}
	prefix, _, ok := strings.Cut(eventType, ".")
	if !ok {
		return eventType
	}
	return prefix
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate %s id: %w", prefix, err)
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:])
	return prefix + "_" + strings.ToLower(encoded), nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func scanStoredEvent(scanner interface{ Scan(...any) error }) (StoredEvent, error) {
	var event StoredEvent
	var actorID sql.NullString
	var projectID sql.NullString
	var relatedType sql.NullString
	var relatedID sql.NullString
	var payloadJSON string
	var occurredAt string
	if err := scanner.Scan(
		&event.ID,
		&event.Type,
		&actorID,
		&projectID,
		&event.SubjectType,
		&event.SubjectID,
		&relatedType,
		&relatedID,
		&payloadJSON,
		&occurredAt,
	); err != nil {
		return StoredEvent{}, fmt.Errorf("scan domain event: %w", err)
	}
	event.ActorID = nullSQLString(actorID)
	event.ProjectID = nullSQLString(projectID)
	event.RelatedType = nullSQLString(relatedType)
	event.RelatedID = nullSQLString(relatedID)
	event.ObjectID = event.SubjectID
	event.Data = map[string]any{}
	if payloadJSON != "" {
		if err := json.Unmarshal([]byte(payloadJSON), &event.Data); err != nil {
			return StoredEvent{}, fmt.Errorf("decode domain event payload: %w", err)
		}
	}
	parsed, err := time.Parse(time.RFC3339Nano, occurredAt)
	if err != nil {
		return StoredEvent{}, fmt.Errorf("parse domain event occurred_at: %w", err)
	}
	event.At = parsed
	return event, nil
}

func nullSQLString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableStringPtr(value *string) any {
	if value == nil {
		return nil
	}
	return nullableString(*value)
}
