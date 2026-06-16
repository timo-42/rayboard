package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	_ "modernc.org/sqlite"
)

const sqliteDriverName = "sqlite"

// DB wraps the backend SQLite handle.
type DB struct {
	SQL *sql.DB
}

// Open opens a SQLite database and applies connection-level pragmas.
func Open(ctx context.Context, path string) (*DB, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("store: database path is required")
	}

	sqlDB, err := sql.Open(sqliteDriverName, sqliteDSN(path))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if err := verifyForeignKeys(ctx, sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	if !isInMemory(path) {
		if err := enableWAL(ctx, sqlDB); err != nil {
			_ = sqlDB.Close()
			return nil, err
		}
	}

	return &DB{SQL: sqlDB}, nil
}

// Close closes the underlying database handle.
func (db *DB) Close() error {
	if db == nil || db.SQL == nil {
		return nil
	}
	return db.SQL.Close()
}

// Migrate applies all embedded migrations to the database.
func (db *DB) Migrate(ctx context.Context) error {
	if db == nil || db.SQL == nil {
		return errors.New("store: nil database")
	}
	return Migrate(ctx, db.SQL)
}

func sqliteDSN(path string) string {
	const (
		foreignKeysPragma = "foreign_keys(ON)"
		busyTimeoutPragma = "busy_timeout(5000)"
	)

	if path == ":memory:" {
		return "file::memory:?mode=memory&cache=shared&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"
	}

	if strings.HasPrefix(path, "file:") {
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		return path + separator + "_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"
	}

	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Add("_pragma", foreignKeysPragma)
	q.Add("_pragma", busyTimeoutPragma)
	u.RawQuery = q.Encode()
	return u.String()
}

func verifyForeignKeys(ctx context.Context, db *sql.DB) error {
	var enabled int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil {
		return fmt.Errorf("read foreign_keys pragma: %w", err)
	}
	if enabled != 1 {
		return errors.New("store: sqlite foreign key enforcement is disabled")
	}
	return nil
}

func enableWAL(ctx context.Context, db *sql.DB) error {
	var mode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode = WAL").Scan(&mode); err != nil {
		return fmt.Errorf("enable sqlite WAL mode: %w", err)
	}
	if !strings.EqualFold(mode, "wal") {
		return fmt.Errorf("enable sqlite WAL mode: got journal mode %q", mode)
	}
	return nil
}

func isInMemory(path string) bool {
	if path == ":memory:" || strings.HasPrefix(path, "file::memory:") {
		return true
	}

	if strings.HasPrefix(path, "file:") {
		u, err := url.Parse(path)
		if err == nil && u.Query().Get("mode") == "memory" {
			return true
		}
	}

	return false
}
