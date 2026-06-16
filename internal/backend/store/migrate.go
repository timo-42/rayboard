package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/timo-42/rayboard/internal/backend/migrations"
)

var migrationFilePattern = regexp.MustCompile(`^([0-9]+)_.+\.sql$`)

const schemaMigrationsDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	checksum TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);`

type appliedMigration struct {
	name     string
	checksum string
}

type migration struct {
	version  int
	name     string
	sql      string
	checksum string
}

// Migrate applies all embedded SQL migrations in version order.
func Migrate(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("store: nil sql database")
	}

	if _, err := db.ExecContext(ctx, schemaMigrationsDDL); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := loadAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	migrationsToApply, err := loadMigrations()
	if err != nil {
		return err
	}

	for _, migration := range migrationsToApply {
		if current, ok := applied[migration.version]; ok {
			if current.name != migration.name {
				return fmt.Errorf("migration %06d was applied as %q, current file is %q", migration.version, current.name, migration.name)
			}
			if current.checksum != migration.checksum {
				return fmt.Errorf("migration %s checksum changed after being applied", migration.name)
			}
			continue
		}

		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}

	return nil
}

func loadAppliedMigrations(ctx context.Context, db *sql.DB) (map[int]appliedMigration, error) {
	rows, err := db.QueryContext(ctx, "SELECT version, name, checksum FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]appliedMigration)
	for rows.Next() {
		var version int
		var migration appliedMigration
		if err := rows.Scan(&version, &migration.name, &migration.checksum); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = migration
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return applied, nil
}

func loadMigrations() ([]migration, error) {
	names, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(names)

	loaded := make([]migration, 0, len(names))
	seenVersions := make(map[int]string, len(names))
	for _, name := range names {
		version, err := migrationVersion(name)
		if err != nil {
			return nil, err
		}
		if existing, ok := seenVersions[version]; ok {
			return nil, fmt.Errorf("duplicate migration version %06d in %q and %q", version, existing, name)
		}
		seenVersions[version] = name

		content, err := migrations.Files.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}

		sum := sha256.Sum256(content)
		loaded = append(loaded, migration{
			version:  version,
			name:     filepath.Base(name),
			sql:      strings.TrimSpace(string(content)),
			checksum: hex.EncodeToString(sum[:]),
		})
	}

	return loaded, nil
}

func migrationVersion(name string) (int, error) {
	base := filepath.Base(name)
	matches := migrationFilePattern.FindStringSubmatch(base)
	if matches == nil {
		return 0, fmt.Errorf("invalid migration filename %q", base)
	}

	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("parse migration version %q: %w", base, err)
	}
	if version <= 0 {
		return 0, fmt.Errorf("invalid migration version %d in %q", version, base)
	}

	return version, nil
}

func applyMigration(ctx context.Context, db *sql.DB, migration migration) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.name, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if migration.sql != "" {
		if _, err = tx.ExecContext(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.name, err)
		}
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, name, checksum)
		VALUES (?, ?, ?)
	`, migration.version, migration.name, migration.checksum); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.name, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.name, err)
	}

	return nil
}
