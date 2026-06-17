package attachments

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestAttachmentLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openAttachmentTestDB(t, ctx)
	seedAttachmentProject(t, ctx, db.SQL)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-1", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	service := NewService(db.SQL, evaluator, WithNow(func() time.Time {
		return time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	}))
	principal := authz.Principal{UserID: "user-1", ActorUserID: "user-1", AuthKind: authz.AuthKindSession}

	meta, err := service.Upload(ctx, principal, UploadInput{
		TicketID:    "ticket-1",
		FileName:    "spec.txt",
		ContentType: "text/plain",
		Data:        []byte("hello"),
	})
	if err != nil {
		t.Fatalf("upload attachment: %v", err)
	}
	if meta.ID == "" || meta.SizeBytes != 5 || meta.UploaderID != "user-1" {
		t.Fatalf("unexpected metadata: %#v", meta)
	}

	listed, err := service.List(ctx, principal, "ticket-1")
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != meta.ID {
		t.Fatalf("unexpected attachments: %#v", listed)
	}

	file, err := service.Download(ctx, principal, meta.ID)
	if err != nil {
		t.Fatalf("download attachment: %v", err)
	}
	if file.FileName != "spec.txt" || !bytes.Equal(file.Data, []byte("hello")) {
		t.Fatalf("unexpected file: %#v", file)
	}

	if err := service.Delete(ctx, principal, meta.ID); err != nil {
		t.Fatalf("delete attachment: %v", err)
	}
	if _, err := service.Download(ctx, principal, meta.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted attachment not found, got %v", err)
	}

	var activities int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM ticket_activity
		WHERE ticket_id = 'ticket-1' AND activity_type IN ('attachment.uploaded', 'attachment.deleted')
	`).Scan(&activities); err != nil {
		t.Fatalf("count attachment activity: %v", err)
	}
	if activities != 2 {
		t.Fatalf("expected two activity rows, got %d", activities)
	}
}

func TestAttachmentValidationAndAuthorization(t *testing.T) {
	ctx := context.Background()
	db := openAttachmentTestDB(t, ctx)
	seedAttachmentProject(t, ctx, db.SQL)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("viewer", authz.RoleProjectViewer, authz.ProjectScope("project-1")),
		authz.UserBinding("member", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	service := NewService(db.SQL, evaluator)

	viewer := authz.Principal{UserID: "viewer", ActorUserID: "viewer", AuthKind: authz.AuthKindSession}
	if _, err := service.Upload(ctx, viewer, UploadInput{
		TicketID: "ticket-1",
		FileName: "denied.txt",
		Data:     []byte("denied"),
	}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden upload, got %v", err)
	}

	member := authz.Principal{UserID: "member", ActorUserID: "member", AuthKind: authz.AuthKindSession}
	if _, err := service.Upload(ctx, member, UploadInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.Upload(ctx, member, UploadInput{
		TicketID: "ticket-1",
		FileName: "too-large.bin",
		Data:     make([]byte, MaxAttachmentSizeBytes+1),
	}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected too large error, got %v", err)
	}
}

func TestAttachmentPolicyProvider(t *testing.T) {
	ctx := context.Background()
	db := openAttachmentTestDB(t, ctx)
	seedAttachmentProject(t, ctx, db.SQL)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("member", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	service := NewService(db.SQL, evaluator, WithPolicyProvider(staticPolicy{
		policy: AttachmentPolicy{
			MaxSizeBytes:        4,
			AllowedContentTypes: []string{"text/plain"},
		},
	}))
	member := authz.Principal{UserID: "member", ActorUserID: "member", AuthKind: authz.AuthKindSession}

	if _, err := service.Upload(ctx, member, UploadInput{
		TicketID:    "ticket-1",
		FileName:    "bad.json",
		ContentType: "application/json",
		Data:        []byte("{}"),
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected content type validation, got %v", err)
	}
	if _, err := service.Upload(ctx, member, UploadInput{
		TicketID:    "ticket-1",
		FileName:    "large.txt",
		ContentType: "text/plain",
		Data:        []byte("large"),
	}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected policy too-large error, got %v", err)
	}
	if _, err := service.Upload(ctx, member, UploadInput{
		TicketID:    "ticket-1",
		FileName:    "ok.txt",
		ContentType: "text/plain; charset=utf-8",
		Data:        []byte("ok"),
	}); err != nil {
		t.Fatalf("expected allowed attachment upload, got %v", err)
	}
}

type staticPolicy struct {
	policy AttachmentPolicy
}

func (policy staticPolicy) AttachmentPolicy(context.Context) (AttachmentPolicy, error) {
	return policy.policy, nil
}

func openAttachmentTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedAttachmentProject(t *testing.T, ctx context.Context, db *sql.DB) {
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
		t.Fatalf("seed attachment project: %v", err)
	}
}
