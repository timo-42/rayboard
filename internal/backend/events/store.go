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
