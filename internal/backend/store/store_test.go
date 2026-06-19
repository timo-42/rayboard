package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
		"domain_events",
		"projects",
		"project_statuses",
		"project_components",
		"project_versions",
		"project_labels",
		"custom_field_definitions",
		"custom_field_options",
		"ticket_custom_field_values",
		"tickets",
		"boards",
		"board_columns",
		"sprints",
		"ticket_labels",
		"ticket_comments",
		"ticket_activity",
		"ticket_attachments",
		"ticket_fts",
		"comment_fts",
		"attachment_fts",
		"saved_views",
		"automation_runs",
		"cron_jobs",
		"notifications",
		"notification_destinations",
		"notification_preferences",
		"notification_policies",
		"notification_deliveries",
		"ticket_hooks",
		"ticket_create_pages",
		"webhooks",
		"outgoing_webhook_deliveries",
		"audit_log",
		"openrouter_providers",
		"system_settings",
		"sprint_report_snapshots",
		"sprint_report_tickets",
		"version_report_snapshots",
		"version_report_tickets",
	} {
		t.Run(table, func(t *testing.T) {
			assertTableExists(t, ctx, db.SQL, table)
		})
	}
}

func TestSprintReportSnapshotMigrationBackfillsCompletedSprints(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)

	applyEmbeddedMigrationsThrough(t, ctx, db.SQL, 26)
	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES ('project_1', 'CORE', 'Core');
		INSERT INTO sprints (id, project_id, name, state, completed_at, created_at, updated_at)
		VALUES ('sprint_done', 'project_1', 'Done Sprint', 'completed', '2026-06-17T10:00:00Z', '2026-06-01T10:00:00Z', '2026-06-17T10:00:00Z');
		INSERT INTO tickets (id, project_id, key, title, status, sprint_id, created_at, updated_at)
		VALUES ('ticket_1', 'project_1', 'CORE-1', 'Migrated ticket', 'done', 'sprint_done', '2026-06-10T10:00:00Z', '2026-06-17T10:00:00Z');
	`); err != nil {
		t.Fatalf("seed pre-snapshot data: %v", err)
	}

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate snapshot backfill: %v", err)
	}

	var capturedAt string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT captured_at
		FROM sprint_report_snapshots
		WHERE sprint_id = 'sprint_done'
	`).Scan(&capturedAt); err != nil {
		t.Fatalf("get backfilled snapshot: %v", err)
	}
	if capturedAt != "2026-06-17T10:00:00Z" {
		t.Fatalf("expected completed_at capture, got %q", capturedAt)
	}

	var ticketID string
	var position int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT ticket_id, position
		FROM sprint_report_tickets
		WHERE sprint_id = 'sprint_done'
	`).Scan(&ticketID, &position); err != nil {
		t.Fatalf("get backfilled snapshot ticket: %v", err)
	}
	if ticketID != "ticket_1" || position != 0 {
		t.Fatalf("unexpected backfilled snapshot ticket %q at %d", ticketID, position)
	}
}

func TestVersionReportSnapshotMigrationBackfillsReleasedVersions(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)

	applyEmbeddedMigrationsThrough(t, ctx, db.SQL, 30)
	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES ('project_1', 'CORE', 'Core');
		INSERT INTO project_versions (id, project_id, name, status, release_date, created_at, updated_at)
		VALUES ('version_released', 'project_1', '2026.7', 'released', '2026-07-03', '2026-06-01T10:00:00Z', '2026-07-03T12:30:00Z');
		INSERT INTO tickets (id, project_id, key, title, status, version_id, created_at, updated_at)
		VALUES ('ticket_1', 'project_1', 'CORE-1', 'Migrated release ticket', 'done', 'version_released', '2026-06-10T10:00:00Z', '2026-07-03T12:30:00Z');
	`); err != nil {
		t.Fatalf("seed pre-version-snapshot data: %v", err)
	}

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate version snapshot backfill: %v", err)
	}

	var capturedAt string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT captured_at
		FROM version_report_snapshots
		WHERE version_id = 'version_released'
	`).Scan(&capturedAt); err != nil {
		t.Fatalf("get backfilled version snapshot: %v", err)
	}
	if capturedAt != "2026-07-03T00:00:00Z" {
		t.Fatalf("expected release_date capture timestamp, got %q", capturedAt)
	}

	var ticketID string
	var position int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT ticket_id, position
		FROM version_report_tickets
		WHERE version_id = 'version_released'
	`).Scan(&ticketID, &position); err != nil {
		t.Fatalf("get backfilled version snapshot ticket: %v", err)
	}
	if ticketID != "ticket_1" || position != 0 {
		t.Fatalf("unexpected backfilled version snapshot ticket %q at %d", ticketID, position)
	}
}

func TestProjectLabelMigrationBackfillsUsedLabels(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)

	applyEmbeddedMigrationsThrough(t, ctx, db.SQL, 31)
	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES ('project_1', 'CORE', 'Core');
		INSERT INTO tickets (id, project_id, key, title, status, created_at, updated_at)
		VALUES ('ticket_1', 'project_1', 'CORE-1', 'Migrated label ticket', 'todo', '2026-06-10T10:00:00Z', '2026-06-10T10:00:00Z');
		INSERT INTO ticket_labels (ticket_id, label, created_at)
		VALUES ('ticket_1', 'backend', '2026-06-11T10:00:00Z');
	`); err != nil {
		t.Fatalf("seed pre-project-label data: %v", err)
	}

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate project labels: %v", err)
	}

	var createdAt string
	var updatedAt string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT created_at, updated_at
		FROM project_labels
		WHERE project_id = 'project_1' AND label = 'backend'
	`).Scan(&createdAt, &updatedAt); err != nil {
		t.Fatalf("get backfilled project label: %v", err)
	}
	if createdAt != "2026-06-11T10:00:00Z" || updatedAt != "2026-06-11T10:00:00Z" {
		t.Fatalf("unexpected backfilled project label timestamps %q/%q", createdAt, updatedAt)
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

func applyEmbeddedMigrationsThrough(t *testing.T, ctx context.Context, db *sql.DB, maxVersion int) {
	t.Helper()

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		);
	`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}

	names, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	sort.Strings(names)
	for _, name := range names {
		versionText := filepath.Base(name)[:6]
		version, err := strconv.Atoi(versionText)
		if err != nil {
			t.Fatalf("parse migration version %q: %v", name, err)
		}
		if version > maxVersion {
			continue
		}
		content, err := migrations.Files.ReadFile(name)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		sum := sha256.Sum256(content)
		if _, err := db.ExecContext(ctx, strings.TrimSpace(string(content))); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO schema_migrations (version, name, checksum)
			VALUES (?, ?, ?)
		`, version, filepath.Base(name), hex.EncodeToString(sum[:])); err != nil {
			t.Fatalf("record migration %s: %v", name, err)
		}
	}
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
