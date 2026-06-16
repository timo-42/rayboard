package events_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestStoreAppendPersistsPendingDomainEvent(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	store := events.NewStore(db.SQL)

	occurredAt := time.Date(2026, 6, 16, 12, 30, 0, 0, time.UTC)
	if err := store.Append(ctx, nil, events.Event{
		Type:        "ticket.created",
		ActorID:     "user-1",
		ProjectID:   "project-1",
		ObjectID:    "ticket-1",
		RelatedType: "sprint",
		RelatedID:   "sprint-1",
		At:          occurredAt,
		Data: map[string]any{
			"key":    "CORE-1",
			"status": "todo",
		},
	}); err != nil {
		t.Fatalf("append domain event: %v", err)
	}

	row := readOnlyDomainEvent(t, ctx, db.SQL)
	if row.EventType != "ticket.created" || row.SubjectType != "ticket" || row.SubjectID != "ticket-1" {
		t.Fatalf("unexpected event subject: %#v", row)
	}
	if !row.ActorID.Valid || row.ActorID.String != "user-1" || !row.ProjectID.Valid || row.ProjectID.String != "project-1" {
		t.Fatalf("unexpected actor/project: %#v", row)
	}
	if !row.RelatedType.Valid || row.RelatedType.String != "sprint" || !row.RelatedID.Valid || row.RelatedID.String != "sprint-1" {
		t.Fatalf("unexpected related object: %#v", row)
	}
	if row.OccurredAt != "2026-06-16T12:30:00Z" || row.ProcessingStatus != "pending" || row.Attempts != 0 {
		t.Fatalf("unexpected processing metadata: %#v", row)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(row.PayloadJSON), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["key"] != "CORE-1" || payload["status"] != "todo" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestStoreAppendRollsBackWithTransaction(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	store := events.NewStore(db.SQL)

	tx, err := db.SQL.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := store.Append(ctx, tx, events.Event{
		Type:     "ticket.updated",
		ObjectID: "ticket-1",
		Data:     map[string]any{"status": "done"},
	}); err != nil {
		t.Fatalf("append domain event: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx: %v", err)
	}

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM domain_events").Scan(&count); err != nil {
		t.Fatalf("count domain events: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rollback to remove domain event, got %d rows", count)
	}
}

func TestStoreListsAndMarksPendingEvents(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	store := events.NewStore(db.SQL)

	if err := store.Append(ctx, nil, events.Event{Type: "ticket.updated", ObjectID: "ticket-1"}); err != nil {
		t.Fatalf("append ticket event: %v", err)
	}
	if err := store.Append(ctx, nil, events.Event{Type: "comment.created", ObjectID: "comment-1", Data: map[string]any{"ticket_id": "ticket-1"}}); err != nil {
		t.Fatalf("append comment event: %v", err)
	}

	pending, err := store.ListPending(ctx, 10, "comment.created")
	if err != nil {
		t.Fatalf("list pending events: %v", err)
	}
	if len(pending) != 1 || pending[0].Type != "comment.created" || pending[0].ObjectID != "comment-1" || pending[0].Data["ticket_id"] != "ticket-1" {
		t.Fatalf("unexpected pending events: %#v", pending)
	}

	if err := store.MarkProcessed(ctx, pending[0].ID); err != nil {
		t.Fatalf("mark processed: %v", err)
	}
	pending, err = store.ListPending(ctx, 10, "comment.created")
	if err != nil {
		t.Fatalf("list pending after processed: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected no pending comment events, got %#v", pending)
	}

	ticketEvents, err := store.ListPending(ctx, 10, "ticket.updated")
	if err != nil {
		t.Fatalf("list ticket events: %v", err)
	}
	if len(ticketEvents) != 1 {
		t.Fatalf("expected one ticket event, got %#v", ticketEvents)
	}
	wantErr := errors.New("handler failed")
	if err := store.MarkFailed(ctx, ticketEvents[0].ID, wantErr); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	var status string
	var attempts int
	var lastError string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT processing_status, attempts, COALESCE(last_error, '')
		FROM domain_events
		WHERE id = ?
	`, ticketEvents[0].ID).Scan(&status, &attempts, &lastError); err != nil {
		t.Fatalf("read failed event: %v", err)
	}
	if status != "failed" || attempts != 1 || lastError != wantErr.Error() {
		t.Fatalf("unexpected failed event state: status=%s attempts=%d error=%q", status, attempts, lastError)
	}
}

func TestStoreAppendRequiresSubject(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	store := events.NewStore(db.SQL)

	if err := store.Append(ctx, nil, events.Event{Type: "ticket.created"}); err == nil {
		t.Fatal("expected missing subject error")
	}
}

type domainEventRow struct {
	EventType        string
	ActorID          sql.NullString
	ProjectID        sql.NullString
	SubjectType      string
	SubjectID        string
	RelatedType      sql.NullString
	RelatedID        sql.NullString
	PayloadJSON      string
	OccurredAt       string
	ProcessingStatus string
	Attempts         int
}

func readOnlyDomainEvent(t *testing.T, ctx context.Context, db *sql.DB) domainEventRow {
	t.Helper()

	var row domainEventRow
	if err := db.QueryRowContext(ctx, `
		SELECT event_type, actor_id, project_id, subject_type, subject_id,
			related_type, related_id, payload_json, occurred_at, processing_status, attempts
		FROM domain_events
	`).Scan(&row.EventType, &row.ActorID, &row.ProjectID, &row.SubjectType, &row.SubjectID,
		&row.RelatedType, &row.RelatedID, &row.PayloadJSON, &row.OccurredAt,
		&row.ProcessingStatus, &row.Attempts); err != nil {
		t.Fatalf("read domain event: %v", err)
	}
	return row
}

func openMigratedDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "rayboard.sqlite"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
	})
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}
