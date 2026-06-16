package store_test

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/migrations"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestOpenMigrateIdempotent(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	var migrations int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	expectedMigrations := countMigrationFiles(t)
	if migrations != expectedMigrations {
		t.Fatalf("expected %d applied migrations, got %d", expectedMigrations, migrations)
	}

	for _, table := range []string{
		"users",
		"sessions",
		"api_tokens",
		"groups",
		"group_memberships",
		"roles",
		"role_permissions",
		"role_bindings",
		"projects",
		"project_components",
		"project_versions",
		"custom_field_definitions",
		"custom_field_options",
		"ticket_custom_field_values",
		"tickets",
		"sprints",
		"ticket_comments",
		"ticket_activity",
		"ticket_attachments",
		"saved_views",
		"automation_runs",
		"cron_jobs",
		"notifications",
	} {
		t.Run(table, func(t *testing.T) {
			assertTableExists(t, ctx, db.SQL, table)
		})
	}
}

func countMigrationFiles(t *testing.T) int {
	t.Helper()

	names, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		t.Fatalf("list migration files: %v", err)
	}
	return len(names)
}

func TestForeignKeysEnforced(t *testing.T) {
	ctx := context.Background()
	db := openMigratedTestDB(t, ctx)

	_, err := db.SQL.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)
	`, "session-1", "missing-user", "hash-1", "2099-01-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected foreign key error, got nil")
	}
}

func TestWithTxCommitAndRollback(t *testing.T) {
	ctx := context.Background()
	db := openMigratedTestDB(t, ctx)

	if err := db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, username, display_name)
			VALUES (?, ?, ?)
		`, "user-commit", "commit-user", "Commit User")
		return err
	}); err != nil {
		t.Fatalf("commit transaction: %v", err)
	}

	assertUserCount(t, ctx, db.SQL, "user-commit", 1)

	rollbackErr := errors.New("rollback")
	err := db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, username, display_name)
			VALUES (?, ?, ?)
		`, "user-rollback", "rollback-user", "Rollback User")
		if err != nil {
			return err
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}

	assertUserCount(t, ctx, db.SQL, "user-rollback", 0)
}

func openMigratedTestDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	db := openTestDB(t, ctx)
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func openTestDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rayboard.sqlite")
	db, err := store.Open(ctx, path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
	})

	return db
}

func assertTableExists(t *testing.T, ctx context.Context, db *sql.DB, table string) {
	t.Helper()

	var exists int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table' AND name = ?
	`, table).Scan(&exists); err != nil {
		t.Fatalf("check table %s: %v", table, err)
	}
	if exists != 1 {
		t.Fatalf("expected table %s to exist", table)
	}
}

func assertUserCount(t *testing.T, ctx context.Context, db *sql.DB, userID string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&got); err != nil {
		t.Fatalf("count user %s: %v", userID, err)
	}
	if got != want {
		t.Fatalf("expected %d rows for user %s, got %d", want, userID, got)
	}
}
