package webhooks

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestIncomingWebhookLifecycleAndTokenRotation(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithNow(func() time.Time { return now }),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        " GitHub ",
		Direction:   DirectionIncoming,
		Enabled:     true,
		ActorUserID: "actor",
		Engine: EngineSpec{
			Type:   EngineTypeLua,
			Script: `return { ok = true }`,
		},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}
	if created.ID == "" || created.Name != "github" || created.Token == "" || !created.TokenSet || created.TokenRotatedAt == nil {
		t.Fatalf("unexpected created webhook: %#v", created)
	}
	assertNoPlaintextWebhookToken(t, ctx, db, created.Token)

	authenticated, err := service.AuthenticateIncoming(ctx, created.ID, created.Token)
	if err != nil {
		t.Fatalf("authenticate incoming webhook: %v", err)
	}
	if authenticated.ID != created.ID || authenticated.ActorUserID != "actor" {
		t.Fatalf("unexpected authenticated webhook: %#v", authenticated)
	}

	items, err := service.List(ctx, principal, ListInput{ProjectID: "project-1", Direction: DirectionIncoming})
	if err != nil {
		t.Fatalf("list webhooks: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID || items[0].TokenSet != true {
		t.Fatalf("unexpected webhook list: %#v", items)
	}

	enabled := false
	name := "GitHub Archive"
	updated, err := service.Update(ctx, principal, created.ID, UpdateInput{
		Name:    &name,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("update webhook: %v", err)
	}
	if updated.Name != "github archive" || updated.Enabled {
		t.Fatalf("unexpected updated webhook: %#v", updated)
	}
	if _, err := service.AuthenticateIncoming(ctx, created.ID, created.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected disabled webhook auth to fail, got %v", err)
	}

	enabled = true
	if _, err := service.Update(ctx, principal, created.ID, UpdateInput{Enabled: &enabled}); err != nil {
		t.Fatalf("reenable webhook: %v", err)
	}
	rotated, err := service.RotateIncomingToken(ctx, principal, created.ID)
	if err != nil {
		t.Fatalf("rotate webhook token: %v", err)
	}
	if rotated.Token == "" || rotated.Token == created.Token {
		t.Fatalf("expected new rotated token, got %#v", rotated)
	}
	assertNoPlaintextWebhookToken(t, ctx, db, rotated.Token)
	if _, err := service.AuthenticateIncoming(ctx, created.ID, created.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected old token to fail, got %v", err)
	}
	if _, err := service.AuthenticateIncoming(ctx, created.ID, rotated.Token); err != nil {
		t.Fatalf("expected rotated token to authenticate: %v", err)
	}

	if err := service.Delete(ctx, principal, created.ID); err != nil {
		t.Fatalf("delete webhook: %v", err)
	}
	if _, err := service.Get(ctx, principal, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted webhook not found, got %v", err)
	}
}

func TestIncomingWebhookValidationAndRBAC(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "viewer", false)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
	)

	viewer := authz.Principal{UserID: "viewer", AuthKind: authz.AuthKindSession}
	if _, err := service.Create(ctx, viewer, CreateInput{
		ProjectID:   "project-1",
		Name:        "denied",
		Direction:   DirectionIncoming,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return {}`},
	}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden create, got %v", err)
	}

	admin := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	if _, err := service.Create(ctx, admin, CreateInput{
		ProjectID:   "project-1",
		Name:        "bad",
		Direction:   DirectionIncoming,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid lua engine validation, got %v", err)
	}
	if _, err := service.Create(ctx, admin, CreateInput{
		ProjectID:   "project-1",
		Name:        "outgoing",
		Direction:   DirectionOutgoing,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return {}`},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected outgoing validation, got %v", err)
	}
}

func openWebhookTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedWebhookProject(t *testing.T, ctx context.Context, db *store.DB, id string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES (?, ?, ?)
	`, id, "WEB", "Webhooks"); err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func seedWebhookUser(t *testing.T, ctx context.Context, db *store.DB, id string, disabled bool) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, is_disabled)
		VALUES (?, ?, ?, ?)
	`, id, id, id, disabled); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func assertNoPlaintextWebhookToken(t *testing.T, ctx context.Context, db *store.DB, token string) {
	t.Helper()

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM webhooks WHERE token_hash = ?", token).Scan(&count); err != nil {
		t.Fatalf("count plaintext webhook token: %v", err)
	}
	if count != 0 {
		t.Fatalf("webhook token stored in plaintext")
	}
}
