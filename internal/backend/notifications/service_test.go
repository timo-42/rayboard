package notifications

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestNotificationListAndReadState(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "user-1")
	seedNotificationUser(t, ctx, db, "user-2")

	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	service := NewService(db.SQL, WithNow(func() time.Time { return now }))
	first, err := service.Create(ctx, CreateInput{
		UserID:      "user-1",
		Type:        "ticket_assigned",
		SubjectType: "ticket",
		SubjectID:   "ticket-1",
		Body:        "Assigned to AUTO-1",
		Data:        map[string]any{"ticket_key": "AUTO-1"},
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	if _, err := service.Create(ctx, CreateInput{
		UserID: "user-2",
		Type:   "ticket_assigned",
		Body:   "Other user",
	}); err != nil {
		t.Fatalf("create other notification: %v", err)
	}

	principal := authz.Principal{UserID: "user-1", AuthKind: authz.AuthKindSession}
	items, err := service.List(ctx, principal, ListInput{})
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(items) != 1 || items[0].ID != first.ID || items[0].Data["ticket_key"] != "AUTO-1" {
		t.Fatalf("unexpected notifications: %#v", items)
	}

	read, err := service.SetRead(ctx, principal, first.ID, true)
	if err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if read.ReadAt == nil {
		t.Fatalf("expected read_at to be set: %#v", read)
	}
	unreadItems, err := service.List(ctx, principal, ListInput{UnreadOnly: true})
	if err != nil {
		t.Fatalf("list unread notifications: %v", err)
	}
	if len(unreadItems) != 0 {
		t.Fatalf("expected no unread notifications, got %#v", unreadItems)
	}

	unread, err := service.SetRead(ctx, principal, first.ID, false)
	if err != nil {
		t.Fatalf("mark unread: %v", err)
	}
	if unread.ReadAt != nil {
		t.Fatalf("expected read_at to clear: %#v", unread)
	}

	if err := service.MarkAllRead(ctx, principal); err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	items, err = service.List(ctx, principal, ListInput{UnreadOnly: true})
	if err != nil {
		t.Fatalf("list unread after all read: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no unread after mark all read, got %#v", items)
	}
}

func TestNotificationEventHandlers(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "actor")
	seedNotificationUser(t, ctx, db, "reporter")
	seedNotificationUser(t, ctx, db, "assignee")
	seedNotificationTicket(t, ctx, db, "ticket-1", "AUTO-1", "reporter", "assignee")

	service := NewService(db.SQL)
	bus := events.NewBus()
	service.RegisterEventHandlers(bus)

	errs := bus.Publish(ctx, events.Event{
		Type:     "comment.created",
		ActorID:  "actor",
		ObjectID: "comment-1",
		Data:     map[string]any{"ticket_id": "ticket-1"},
	})
	if len(errs) != 0 {
		t.Fatalf("comment event errors: %v", errs)
	}
	if got := countNotifications(t, ctx, db, "comment_added"); got != 2 {
		t.Fatalf("expected 2 comment notifications, got %d", got)
	}

	errs = bus.Publish(ctx, events.Event{
		Type:     "ticket.updated",
		ActorID:  "actor",
		ObjectID: "ticket-1",
		Data: map[string]any{
			"changes": map[string]any{
				"status":      map[string]string{"old": "todo", "new": "done"},
				"assignee_id": map[string]string{"old": "", "new": "assignee"},
			},
		},
	})
	if len(errs) != 0 {
		t.Fatalf("ticket event errors: %v", errs)
	}
	if got := countNotifications(t, ctx, db, "ticket_status_changed"); got != 2 {
		t.Fatalf("expected 2 status notifications, got %d", got)
	}
	if got := countNotifications(t, ctx, db, "ticket_assigned"); got != 1 {
		t.Fatalf("expected 1 assignment notification, got %d", got)
	}
}

func TestProcessPendingDomainEvents(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "actor")
	seedNotificationUser(t, ctx, db, "reporter")
	seedNotificationUser(t, ctx, db, "assignee")
	seedNotificationTicket(t, ctx, db, "ticket-1", "AUTO-1", "reporter", "assignee")

	eventStore := events.NewStore(db.SQL)
	service := NewService(db.SQL, WithEventStore(eventStore))
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:     "comment.created",
		ActorID:  "actor",
		ObjectID: "comment-1",
		Data:     map[string]any{"ticket_id": "ticket-1"},
	}); err != nil {
		t.Fatalf("append comment event: %v", err)
	}
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:     "ticket.updated",
		ActorID:  "actor",
		ObjectID: "ticket-1",
		Data: map[string]any{
			"changes": map[string]any{
				"status":      map[string]string{"old": "todo", "new": "done"},
				"assignee_id": map[string]string{"old": "", "new": "assignee"},
			},
		},
	}); err != nil {
		t.Fatalf("append ticket event: %v", err)
	}

	processed, err := service.ProcessPendingDomainEvents(ctx, 10)
	if err != nil {
		t.Fatalf("process pending domain events: %v", err)
	}
	if processed != 2 {
		t.Fatalf("expected 2 processed events, got %d", processed)
	}
	if got := countNotifications(t, ctx, db, "comment_added"); got != 2 {
		t.Fatalf("expected 2 comment notifications, got %d", got)
	}
	if got := countNotifications(t, ctx, db, "ticket_status_changed"); got != 2 {
		t.Fatalf("expected 2 status notifications, got %d", got)
	}
	if got := countNotifications(t, ctx, db, "ticket_assigned"); got != 1 {
		t.Fatalf("expected 1 assignment notification, got %d", got)
	}
	assertDomainEventStatus(t, ctx, db, "comment.created", "processed", 1, "")
	assertDomainEventStatus(t, ctx, db, "ticket.updated", "processed", 1, "")

	processed, err = service.ProcessPendingDomainEvents(ctx, 10)
	if err != nil {
		t.Fatalf("process pending domain events again: %v", err)
	}
	if processed != 0 {
		t.Fatalf("expected no processed events on rerun, got %d", processed)
	}
	if got := totalNotifications(t, ctx, db); got != 5 {
		t.Fatalf("expected no duplicate notifications on rerun, got %d", got)
	}
}

func TestProcessPendingDomainEventsMarksFailures(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)

	eventStore := events.NewStore(db.SQL)
	service := NewService(db.SQL, WithEventStore(eventStore))
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:     "ticket.updated",
		ActorID:  "actor",
		ObjectID: "missing-ticket",
		Data: map[string]any{
			"changes": map[string]any{
				"status": map[string]string{"old": "todo", "new": "done"},
			},
		},
	}); err != nil {
		t.Fatalf("append missing ticket event: %v", err)
	}

	processed, err := service.ProcessPendingDomainEvents(ctx, 10)
	if err == nil {
		t.Fatal("expected processing error")
	}
	if processed != 0 {
		t.Fatalf("expected no processed events, got %d", processed)
	}
	assertDomainEventStatus(t, ctx, db, "ticket.updated", "failed", 1, "notifications: not found")
}

func openNotificationTestDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "rayboard.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return db
}

func seedNotificationUser(t *testing.T, ctx context.Context, db *store.DB, id string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name)
		VALUES (?, ?, ?)
	`, id, id, id); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

func seedNotificationProject(t *testing.T, ctx context.Context, db *store.DB, id string, key string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES (?, ?, ?)
	`, id, key, key); err != nil {
		t.Fatalf("seed project %s: %v", id, err)
	}
}

func seedNotificationTicket(t *testing.T, ctx context.Context, db *store.DB, id string, key string, reporterID string, assigneeID string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES ('project-1', 'AUTO', 'Automation')
	`); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO tickets (id, project_id, key, title, reporter_id, assignee_id)
		VALUES (?, 'project-1', ?, 'Ticket', ?, ?)
	`, id, key, reporterID, assigneeID); err != nil {
		t.Fatalf("seed ticket: %v", err)
	}
}

func countNotifications(t *testing.T, ctx context.Context, db *store.DB, notificationType string) int {
	t.Helper()

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM notifications WHERE type = ?", notificationType).Scan(&count); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	return count
}

func totalNotifications(t *testing.T, ctx context.Context, db *store.DB) int {
	t.Helper()

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM notifications").Scan(&count); err != nil {
		t.Fatalf("count all notifications: %v", err)
	}
	return count
}

func assertDomainEventStatus(t *testing.T, ctx context.Context, db *store.DB, eventType string, wantStatus string, wantAttempts int, wantError string) {
	t.Helper()

	var status string
	var attempts int
	var lastError string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT processing_status, attempts, COALESCE(last_error, '')
		FROM domain_events
		WHERE event_type = ?
	`, eventType).Scan(&status, &attempts, &lastError); err != nil {
		t.Fatalf("read domain event status for %s: %v", eventType, err)
	}
	if status != wantStatus || attempts != wantAttempts || (wantError != "" && lastError != wantError) {
		t.Fatalf("unexpected event status for %s: status=%s attempts=%d error=%q", eventType, status, attempts, lastError)
	}
}
