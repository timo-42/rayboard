package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

func TestGroupsAndRoleBindingsGrantPermissions(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	bootstrap, err := BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	service := NewService(db.SQL)
	user, err := service.CreateUser(ctx, CreateUserInput{Username: "member"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	group, err := service.CreateGroup(ctx, CreateGroupInput{Name: "users", DisplayName: "Users"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := service.AddGroupMember(ctx, group.ID, user.ID); err != nil {
		t.Fatalf("add group member: %v", err)
	}

	members, err := service.ListGroupMembers(ctx, group.ID)
	if err != nil {
		t.Fatalf("list group members: %v", err)
	}
	if len(members) != 1 || members[0].ID != user.ID {
		t.Fatalf("unexpected group members: %#v", members)
	}

	binding, err := service.CreateRoleBinding(ctx, CreateRoleBindingInput{
		RoleName:    authz.RoleGlobalUserManager,
		SubjectType: authz.BindingTargetGroup,
		SubjectID:   group.ID,
		Scope:       authz.GlobalScope(),
	})
	if err != nil {
		t.Fatalf("create role binding: %v", err)
	}
	if binding.RoleName != authz.RoleGlobalUserManager || binding.SubjectID != group.ID {
		t.Fatalf("unexpected role binding: %#v", binding)
	}

	evaluator := authz.NewSQLEvaluator(db.SQL)
	if !evaluator.Can(authz.Principal{UserID: user.ID, AuthKind: authz.AuthKindSession}, authz.PermissionUsersRead, authz.GlobalScope()) {
		t.Fatal("expected group role binding to grant users:read")
	}
	if !evaluator.Can(authz.Principal{UserID: bootstrap.UserID, AuthKind: authz.AuthKindSession}, authz.PermissionUsersWrite, authz.GlobalScope()) {
		t.Fatal("expected bootstrap admin to keep users:write")
	}

	bindings, err := service.ListRoleBindings(ctx)
	if err != nil {
		t.Fatalf("list role bindings: %v", err)
	}
	if len(bindings) < 2 {
		t.Fatalf("expected bootstrap and group bindings, got %#v", bindings)
	}

	if err := service.DeleteRoleBinding(ctx, binding.ID); err != nil {
		t.Fatalf("delete role binding: %v", err)
	}
	if evaluator.Can(authz.Principal{UserID: user.ID, AuthKind: authz.AuthKindSession}, authz.PermissionUsersRead, authz.GlobalScope()) {
		t.Fatal("expected deleted role binding to revoke users:read")
	}

	if err := service.RemoveGroupMember(ctx, group.ID, user.ID); err != nil {
		t.Fatalf("remove group member: %v", err)
	}
	if _, err := service.ListGroupMembers(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing group not found, got %v", err)
	}
}

func TestGroupAndRoleBindingValidation(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	if _, err := BootstrapAdmin(ctx, db.SQL); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	service := NewService(db.SQL)

	if _, err := service.CreateGroup(ctx, CreateGroupInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected group validation error, got %v", err)
	}

	group, err := service.CreateGroup(ctx, CreateGroupInput{Name: "ops"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if _, err := service.CreateGroup(ctx, CreateGroupInput{Name: "ops"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected duplicate group conflict, got %v", err)
	}

	if err := service.AddGroupMember(ctx, group.ID, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing user not found, got %v", err)
	}
	if _, err := service.CreateRoleBinding(ctx, CreateRoleBindingInput{
		RoleName:    "missing",
		SubjectType: authz.BindingTargetGroup,
		SubjectID:   group.ID,
		Scope:       authz.GlobalScope(),
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing role not found, got %v", err)
	}
	if _, err := service.CreateRoleBinding(ctx, CreateRoleBindingInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected role binding validation error, got %v", err)
	}

	roles, err := service.ListRoles(ctx)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(roles) == 0 {
		t.Fatal("expected built-in roles")
	}
}
