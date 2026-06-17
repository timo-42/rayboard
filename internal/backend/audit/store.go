package audit

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

	"github.com/timo-42/rayboard/internal/backend/authz"
)

const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
)

type Store struct {
	db  *sql.DB
	now func() time.Time
}

type Option func(*Store)

type Entry struct {
	ID          string
	EventType   string
	ActorID     string
	AuthKind    authz.AuthKind
	SubjectType string
	SubjectID   string
	Outcome     string
	Payload     map[string]any
	OccurredAt  time.Time
}

type RecordInput struct {
	EventType   string
	ActorID     string
	AuthKind    authz.AuthKind
	SubjectType string
	SubjectID   string
	Outcome     string
	Payload     map[string]any
	OccurredAt  time.Time
}

type ListInput struct {
	Limit       int
	EventType   string
	ActorID     string
	SubjectType string
	SubjectID   string
	Outcome     string
}

func NewStore(db *sql.DB, options ...Option) *Store {
	store := &Store{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func WithNow(now func() time.Time) Option {
	return func(store *Store) {
		if now != nil {
			store.now = now
		}
	}
}

func (s *Store) Record(ctx context.Context, input RecordInput) (Entry, error) {
	if s == nil || s.db == nil {
		return Entry{}, errors.New("audit: database is required")
	}
	input.EventType = strings.TrimSpace(input.EventType)
	input.SubjectType = strings.TrimSpace(input.SubjectType)
	input.SubjectID = strings.TrimSpace(input.SubjectID)
	if input.EventType == "" {
		return Entry{}, errors.New("audit: event type is required")
	}
	if input.SubjectType == "" {
		return Entry{}, errors.New("audit: subject type is required")
	}
	if input.Outcome == "" {
		input.Outcome = OutcomeSuccess
	}
	if input.Outcome != OutcomeSuccess && input.Outcome != OutcomeFailure {
		return Entry{}, errors.New("audit: invalid outcome")
	}
	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = s.now().UTC()
	}
	payload := input.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Entry{}, fmt.Errorf("encode audit payload: %w", err)
	}
	id, err := newID()
	if err != nil {
		return Entry{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_log (
			id, event_type, actor_id, auth_kind, subject_type, subject_id,
			outcome, payload_json, occurred_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, input.EventType, nullableString(input.ActorID), nullableString(string(input.AuthKind)),
		input.SubjectType, nullableString(input.SubjectID), input.Outcome, string(encoded), formatTime(occurredAt)); err != nil {
		return Entry{}, fmt.Errorf("insert audit entry: %w", err)
	}
	return Entry{
		ID:          id,
		EventType:   input.EventType,
		ActorID:     input.ActorID,
		AuthKind:    input.AuthKind,
		SubjectType: input.SubjectType,
		SubjectID:   input.SubjectID,
		Outcome:     input.Outcome,
		Payload:     payload,
		OccurredAt:  occurredAt,
	}, nil
}

func (s *Store) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.ListEntries(ctx, ListInput{Limit: limit})
}

func (s *Store) ListEntries(ctx context.Context, input ListInput) ([]Entry, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("audit: database is required")
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	query := `
		SELECT id, event_type, actor_id, auth_kind, subject_type, subject_id,
			outcome, payload_json, occurred_at
		FROM audit_log
		WHERE (? = '' OR event_type = ?)
			AND (? = '' OR actor_id = ?)
			AND (? = '' OR subject_type = ?)
			AND (? = '' OR subject_id = ?)
			AND (? = '' OR outcome = ?)
		ORDER BY occurred_at DESC, created_at DESC, id DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query,
		strings.TrimSpace(input.EventType), strings.TrimSpace(input.EventType),
		strings.TrimSpace(input.ActorID), strings.TrimSpace(input.ActorID),
		strings.TrimSpace(input.SubjectType), strings.TrimSpace(input.SubjectType),
		strings.TrimSpace(input.SubjectID), strings.TrimSpace(input.SubjectID),
		strings.TrimSpace(input.Outcome), strings.TrimSpace(input.Outcome),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit entries: %w", err)
	}
	return entries, nil
}

func scanEntry(rows interface {
	Scan(dest ...any) error
}) (Entry, error) {
	var entry Entry
	var actor sql.NullString
	var authKind sql.NullString
	var subjectID sql.NullString
	var payloadJSON string
	var occurredAt string
	if err := rows.Scan(
		&entry.ID,
		&entry.EventType,
		&actor,
		&authKind,
		&entry.SubjectType,
		&subjectID,
		&entry.Outcome,
		&payloadJSON,
		&occurredAt,
	); err != nil {
		return Entry{}, fmt.Errorf("scan audit entry: %w", err)
	}
	entry.ActorID = actor.String
	entry.AuthKind = authz.AuthKind(authKind.String)
	entry.SubjectID = subjectID.String
	if payloadJSON == "" {
		payloadJSON = "{}"
	}
	if err := json.Unmarshal([]byte(payloadJSON), &entry.Payload); err != nil {
		return Entry{}, fmt.Errorf("decode audit payload: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, occurredAt)
	if err != nil {
		return Entry{}, fmt.Errorf("parse audit occurred_at: %w", err)
	}
	entry.OccurredAt = parsed
	return entry, nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func newID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate audit id: %w", err)
	}
	return "audit_" + strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(raw[:]), "=")), nil
}
