package auth

import (
	"context"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestBootstrapAdminCreatesAdminAndRoles(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)

	result, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap admin: %v", err)
	}

	if result.Username != "admin" || result.Password == "" || result.UserID == "" {
		t.Fatalf("unexpected result: %#v", result)
	}

	var hash string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT password_hash FROM users WHERE username = 'admin' AND is_disabled = 0
	`).Scan(&hash); err != nil {
		t.Fatalf("query admin: %v", err)
	}
	if !VerifyPassword(hash, result.Password) {
		t.Fatal("expected generated password to verify against stored hash")
	}

	var globalAdminPermissions int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM roles r
		JOIN role_permissions rp ON rp.role_id = r.id
		WHERE r.name = ? AND rp.permission = '*'
	`, string(authz.RoleGlobalAdmin)).Scan(&globalAdminPermissions); err != nil {
		t.Fatalf("query global admin permissions: %v", err)
	}
	if globalAdminPermissions != 1 {
		t.Fatalf("expected global admin wildcard permission, got %d", globalAdminPermissions)
	}
}

func TestBootstrapAdminResetsPasswordWithoutDuplicatingRecords(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)

	first, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}
	second, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	if first.Password == second.Password {
		t.Fatal("expected password to be regenerated")
	}

	var users int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&users); err != nil {
		t.Fatalf("count admin users: %v", err)
	}
	if users != 1 {
		t.Fatalf("expected one admin user, got %d", users)
	}

	var bindings int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM role_bindings WHERE id = 'binding_admin_global_admin'").Scan(&bindings); err != nil {
		t.Fatalf("count admin bindings: %v", err)
	}
	if bindings != 1 {
		t.Fatalf("expected one admin binding, got %d", bindings)
	}

	var hash string
	if err := db.SQL.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE username = 'admin'").Scan(&hash); err != nil {
		t.Fatalf("query admin hash: %v", err)
	}
	if VerifyPassword(hash, first.Password) {
		t.Fatal("expected first password to stop working after reset")
	}
	if !VerifyPassword(hash, second.Password) {
		t.Fatal("expected second password to verify")
	}
}

func openMigratedDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
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
