package comments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestCommentLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openCommentTestDB(t, ctx)
	seedCommentProject(t, ctx, db.SQL)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-1", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	service := NewService(db.SQL, evaluator, WithNow(func() time.Time {
		return time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	}))
	principal := authz.Principal{UserID: "user-1", ActorUserID: "user-1", AuthKind: authz.AuthKindSession}

	comment, err := service.Create(ctx, principal, CreateInput{TicketID: "ticket-1", Body: "First comment"})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if comment.ID == "" || comment.AuthorID != "user-1" || comment.Body != "First comment" {
		t.Fatalf("unexpected comment: %#v", comment)
	}

	comments, err := service.List(ctx, principal, "ticket-1")
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 1 || comments[0].ID != comment.ID {
		t.Fatalf("unexpected comments: %#v", comments)
	}

	if err := service.Delete(ctx, principal, comment.ID); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	comments, err = service.List(ctx, principal, "ticket-1")
	if err != nil {
		t.Fatalf("list comments after delete: %v", err)
	}
	if len(comments) != 0 {
		t.Fatalf("expected no comments after delete, got %#v", comments)
	}

	var activities int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM ticket_activity
		WHERE ticket_id = 'ticket-1' AND activity_type IN ('comment.created', 'comment.deleted')
	`).Scan(&activities); err != nil {
		t.Fatalf("count comment activity: %v", err)
	}
	if activities != 2 {
		t.Fatalf("expected two activity rows, got %d", activities)
	}
}

func TestCommentCreateAppendsMentionEvents(t *testing.T) {
	ctx := context.Background()
	db := openCommentTestDB(t, ctx)
	seedCommentProject(t, ctx, db.SQL)
	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, is_disabled)
		VALUES
			('disabled', 'disabled', 'Disabled User', 1),
			('outsider', 'outsider', 'Outsider', 0)
	`); err != nil {
		t.Fatalf("seed mention users: %v", err)
	}

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-1", authz.RoleProjectMember, authz.ProjectScope("project-1")),
		authz.UserBinding("member", authz.RoleProjectMember, authz.ProjectScope("project-1")),
		authz.UserBinding("viewer", authz.RoleProjectViewer, authz.ProjectScope("project-1")),
	))
	service := NewService(db.SQL, evaluator, WithEventStore(events.NewStore(db.SQL)), WithNow(func() time.Time {
		return time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	}))
	principal := authz.Principal{UserID: "user-1", ActorUserID: "user-1", AuthKind: authz.AuthKindSession}

	comment, err := service.Create(ctx, principal, CreateInput{
		TicketID: "ticket-1",
		Body:     "Please check this @member and @viewer. Duplicate @Member, missing @ghost, disabled @disabled, inaccessible @outsider, self @user-1.",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	rows, err := db.SQL.QueryContext(ctx, `
		SELECT payload_json
		FROM domain_events
		WHERE event_type = 'comment.mentioned'
		ORDER BY payload_json ASC
	`)
	if err != nil {
		t.Fatalf("query mention events: %v", err)
	}
	defer rows.Close()
	var mentioned []string
	for rows.Next() {
		var payloadJSON string
		if err := rows.Scan(&payloadJSON); err != nil {
			t.Fatalf("scan mention payload: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
			t.Fatalf("decode mention payload: %v", err)
		}
		if payload["ticket_id"] != "ticket-1" || payload["comment_id"] != comment.ID {
			t.Fatalf("unexpected mention payload: %#v", payload)
		}
		mentioned = append(mentioned, payload["mentioned_username"].(string))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate mention events: %v", err)
	}
	if len(mentioned) != 2 || mentioned[0] != "member" || mentioned[1] != "viewer" {
		t.Fatalf("unexpected mentioned users: %#v", mentioned)
	}
}

func TestCommentValidationAndAuthorization(t *testing.T) {
	ctx := context.Background()
	db := openCommentTestDB(t, ctx)
	seedCommentProject(t, ctx, db.SQL)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("viewer", authz.RoleProjectViewer, authz.ProjectScope("project-1")),
		authz.UserBinding("member", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	service := NewService(db.SQL, evaluator)

	viewer := authz.Principal{UserID: "viewer", ActorUserID: "viewer", AuthKind: authz.AuthKindSession}
	if _, err := service.Create(ctx, viewer, CreateInput{TicketID: "ticket-1", Body: "Denied"}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden create, got %v", err)
	}

	member := authz.Principal{UserID: "member", ActorUserID: "member", AuthKind: authz.AuthKindSession}
	if _, err := service.Create(ctx, member, CreateInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.List(ctx, member, "missing-ticket"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing ticket not found, got %v", err)
	}
}

func openCommentTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedCommentProject(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name)
		VALUES
			('user-1', 'user-1', 'User One'),
			('viewer', 'viewer', 'Viewer'),
			('member', 'member', 'Member');
		INSERT INTO projects (id, key, name)
		VALUES ('project-1', 'CORE', 'Core');
		INSERT INTO tickets (id, project_id, key, title, status)
		VALUES ('ticket-1', 'project-1', 'CORE-1', 'Ticket One', 'todo');
	`); err != nil {
		t.Fatalf("seed comment project: %v", err)
	}
}
