package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestSQLEvaluatorGlobalAndProjectBindings(t *testing.T) {
	ctx := context.Background()
	db := openAuthzTestDB(t, ctx)
	seedBuiltInRoles(t, ctx, db)
	seedUser(t, ctx, db, "user-admin", "admin", false)
	seedUser(t, ctx, db, "user-project", "project", false)

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO role_bindings (id, role_id, subject_type, subject_id, resource_type, resource_id)
		VALUES
			('binding-global-admin', 'role_global_admin', 'user', 'user-admin', NULL, NULL),
			('binding-project-viewer', 'role_project_viewer', 'user', 'user-project', 'project', 'project-1')
	`); err != nil {
		t.Fatalf("seed role bindings: %v", err)
	}

	evaluator := NewSQLEvaluator(db.SQL)
	admin := Principal{UserID: "user-admin", AuthKind: AuthKindSession}
	projectUser := Principal{UserID: "user-project", AuthKind: AuthKindAPIToken}

	if !evaluator.Can(admin, PermissionUsersWrite, GlobalScope()) {
		t.Fatal("expected global admin to write users")
	}
	if !evaluator.Can(admin, PermissionTicketsWrite, ProjectScope("project-2")) {
		t.Fatal("expected global admin to write tickets in any project")
	}
	if !evaluator.Can(projectUser, PermissionTicketsRead, ProjectScope("project-1")) {
		t.Fatal("expected project viewer to read its project")
	}
	if evaluator.Can(projectUser, PermissionTicketsRead, ProjectScope("project-2")) {
		t.Fatal("expected project viewer to be denied in another project")
	}
	if evaluator.Can(projectUser, PermissionUsersRead, GlobalScope()) {
		t.Fatal("expected project viewer to be denied global user reads")
	}
}

func TestSQLEvaluatorGroupBindingAndImmediateRevocation(t *testing.T) {
	ctx := context.Background()
	db := openAuthzTestDB(t, ctx)
	seedBuiltInRoles(t, ctx, db)
	seedUser(t, ctx, db, "user-1", "user", false)

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO groups (id, name, display_name) VALUES ('group-1', 'devs', 'Developers');
		INSERT INTO group_memberships (group_id, user_id) VALUES ('group-1', 'user-1');
		INSERT INTO role_bindings (id, role_id, subject_type, subject_id, resource_type, resource_id)
		VALUES ('binding-group-member', 'role_project_member', 'group', 'group-1', 'project', 'project-1')
	`); err != nil {
		t.Fatalf("seed group binding: %v", err)
	}

	evaluator := NewSQLEvaluator(db.SQL)
	principal := Principal{UserID: "user-1", AuthKind: AuthKindSession}

	if !evaluator.Can(principal, PermissionCommentsWrite, ProjectScope("project-1")) {
		t.Fatal("expected group member to write comments")
	}

	if _, err := db.SQL.ExecContext(ctx, "DELETE FROM group_memberships WHERE group_id = 'group-1' AND user_id = 'user-1'"); err != nil {
		t.Fatalf("delete group membership: %v", err)
	}
	if evaluator.Can(principal, PermissionCommentsWrite, ProjectScope("project-1")) {
		t.Fatal("expected deleted group membership to revoke permission immediately")
	}
}

func TestSQLEvaluatorDisabledUserDenied(t *testing.T) {
	ctx := context.Background()
	db := openAuthzTestDB(t, ctx)
	seedBuiltInRoles(t, ctx, db)
	seedUser(t, ctx, db, "user-1", "user", true)

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO role_bindings (id, role_id, subject_type, subject_id, resource_type, resource_id)
		VALUES ('binding-global-admin', 'role_global_admin', 'user', 'user-1', NULL, NULL)
	`); err != nil {
		t.Fatalf("seed role binding: %v", err)
	}

	evaluator := NewSQLEvaluator(db.SQL)
	principal := Principal{UserID: "user-1", AuthKind: AuthKindSession}
	if evaluator.Can(principal, PermissionUsersWrite, GlobalScope()) {
		t.Fatal("expected disabled user to be denied")
	}
	if err := evaluator.Require(principal, PermissionUsersWrite, GlobalScope()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func openAuthzTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedBuiltInRoles(t *testing.T, ctx context.Context, db *store.DB) {
	t.Helper()

	for _, role := range BuiltInRoles() {
		roleID := "role_" + string(role.Name)
		if _, err := db.SQL.ExecContext(ctx, `
			INSERT INTO roles (id, name, description)
			VALUES (?, ?, ?)
		`, roleID, string(role.Name), "test role"); err != nil {
			t.Fatalf("insert role %s: %v", role.Name, err)
		}
		for _, permission := range role.Permissions {
			if _, err := db.SQL.ExecContext(ctx, `
				INSERT INTO role_permissions (role_id, permission)
				VALUES (?, ?)
			`, roleID, string(permission)); err != nil {
				t.Fatalf("insert permission %s: %v", permission, err)
			}
		}
	}
}

func seedUser(t *testing.T, ctx context.Context, db *store.DB, id string, username string, disabled bool) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, is_disabled)
		VALUES (?, ?, ?, ?)
	`, id, username, username, disabled); err != nil {
		t.Fatalf("insert user %s: %v", username, err)
	}
}
