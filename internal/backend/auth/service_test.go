package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

func TestLoginCreatesSessionAndAuthenticates(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	bootstrap, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	service := NewService(db.SQL)
	session, err := service.Login(ctx, bootstrap.Username, bootstrap.Password)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if session.Secret == "" {
		t.Fatal("expected session secret")
	}

	principal, user, err := service.AuthenticateSession(ctx, session.Secret)
	if err != nil {
		t.Fatalf("authenticate session: %v", err)
	}
	if principal.UserID != bootstrap.UserID || principal.AuthKind != authz.AuthKindSession || user.Username != "admin" {
		t.Fatalf("unexpected auth result: %#v %#v", principal, user)
	}
}

func TestLoginRejectsBadPasswordAndDisabledUser(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	bootstrap, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	service := NewService(db.SQL)
	if _, err := service.Login(ctx, bootstrap.Username, "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	if _, err := db.SQL.ExecContext(ctx, "UPDATE users SET is_disabled = 1 WHERE id = ?", bootstrap.UserID); err != nil {
		t.Fatalf("disable admin: %v", err)
	}
	if _, err := service.Login(ctx, bootstrap.Username, bootstrap.Password); !errors.Is(err, ErrDisabledUser) {
		t.Fatalf("expected disabled user, got %v", err)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	bootstrap, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	service := NewService(db.SQL)
	session, err := service.Login(ctx, bootstrap.Username, bootstrap.Password)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if err := service.Logout(ctx, session.Secret); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, _, err := service.AuthenticateSession(ctx, session.Secret); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAPITokenLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	bootstrap, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	service := NewService(db.SQL, WithNow(func() time.Time {
		return time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	}))

	created, err := service.CreateAPIToken(ctx, bootstrap.UserID, "demo")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if created.Token == "" {
		t.Fatal("expected plaintext token")
	}

	tokens, err := service.ListAPITokens(ctx, bootstrap.UserID)
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(tokens) != 1 || tokens[0].ID != created.ID || tokens[0].Name != "demo" {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}

	principal, user, err := service.AuthenticateBearer(ctx, created.Token)
	if err != nil {
		t.Fatalf("authenticate bearer: %v", err)
	}
	if principal.UserID != bootstrap.UserID || principal.AuthKind != authz.AuthKindAPIToken || user.Username != "admin" {
		t.Fatalf("unexpected auth result: %#v %#v", principal, user)
	}

	if err := service.RevokeAPIToken(ctx, bootstrap.UserID, created.ID); err != nil {
		t.Fatalf("revoke token: %v", err)
	}
	if _, _, err := service.AuthenticateBearer(ctx, created.Token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}
