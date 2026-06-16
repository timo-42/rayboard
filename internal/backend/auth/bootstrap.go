package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

const (
	bootstrapAdminUsername = "admin"
	bootstrapAdminUserID   = "user_admin"
	globalScope            = "global"
)

type BootstrapAdminResult struct {
	UserID   string
	Username string
	Password string
}

func BootstrapAdmin(ctx context.Context, db *sql.DB) (BootstrapAdminResult, error) {
	password, err := randomPassword()
	if err != nil {
		return BootstrapAdminResult{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return BootstrapAdminResult{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	defer tx.Rollback()

	if err := seedBuiltInRoles(ctx, tx); err != nil {
		return BootstrapAdminResult{}, err
	}
	if err := upsertAdminUser(ctx, tx, hash); err != nil {
		return BootstrapAdminResult{}, err
	}

	adminID, err := adminUserID(ctx, tx)
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	if err := bindAdminRole(ctx, tx, adminID); err != nil {
		return BootstrapAdminResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return BootstrapAdminResult{}, err
	}

	return BootstrapAdminResult{
		UserID:   adminID,
		Username: bootstrapAdminUsername,
		Password: password,
	}, nil
}

func seedBuiltInRoles(ctx context.Context, tx *sql.Tx) error {
	for _, role := range authz.BuiltInRoles() {
		roleID := string(role.Name)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO roles (id, name, description)
			VALUES (?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				description = excluded.description,
				updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		`, roleID, roleID, "Built-in role"); err != nil {
			return fmt.Errorf("seed role %s: %w", role.Name, err)
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = ?", roleID); err != nil {
			return fmt.Errorf("reset role permissions %s: %w", role.Name, err)
		}
		for _, permission := range role.Permissions {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO role_permissions (role_id, permission)
				VALUES (?, ?)
			`, roleID, string(permission)); err != nil {
				return fmt.Errorf("seed role permission %s/%s: %w", role.Name, permission, err)
			}
		}
	}
	return nil
}

func upsertAdminUser(ctx context.Context, tx *sql.Tx, passwordHash string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, password_hash, is_disabled)
		VALUES (?, ?, ?, ?, 0)
		ON CONFLICT(username) DO UPDATE SET
			password_hash = excluded.password_hash,
			is_disabled = 0,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
	`, bootstrapAdminUserID, bootstrapAdminUsername, "Administrator", passwordHash)
	if err != nil {
		return fmt.Errorf("upsert admin user: %w", err)
	}
	return nil
}

func adminUserID(ctx context.Context, tx *sql.Tx) (string, error) {
	var id string
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM users WHERE username = ?
	`, bootstrapAdminUsername).Scan(&id); err != nil {
		return "", fmt.Errorf("query admin user: %w", err)
	}
	return id, nil
}

func bindAdminRole(ctx context.Context, tx *sql.Tx, adminID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO role_bindings (id, role_id, subject_type, subject_id, resource_type, resource_id)
		VALUES (?, ?, 'user', ?, ?, NULL)
		ON CONFLICT(id) DO NOTHING
	`, "binding_admin_global_admin", string(authz.RoleGlobalAdmin), adminID, globalScope)
	if err != nil {
		return fmt.Errorf("bind admin role: %w", err)
	}
	return nil
}

func randomPassword() (string, error) {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
