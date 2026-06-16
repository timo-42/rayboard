package authz

import (
	"errors"
	"testing"
)

func TestDirectUserRoleBinding(t *testing.T) {
	evaluator := NewInMemoryEvaluator()
	principal := Principal{UserID: "user-1", AuthKind: AuthKindSession}

	evaluator.BindRole(UserBinding("user-1", RoleProjectViewer, ProjectScope("project-1")))

	if !evaluator.Can(principal, PermissionTicketsRead, ProjectScope("project-1")) {
		t.Fatal("expected direct project viewer to read tickets")
	}
	if evaluator.Can(principal, PermissionTicketsWrite, ProjectScope("project-1")) {
		t.Fatal("expected project viewer to be denied ticket writes")
	}
}

func TestGroupRoleBinding(t *testing.T) {
	evaluator := NewInMemoryEvaluator()
	principal := Principal{UserID: "user-1", AuthKind: AuthKindAPIToken}

	evaluator.AddGroupMembership(GroupMembership{UserID: "user-1", GroupID: "group-1"})
	evaluator.BindRole(GroupBinding("group-1", RoleProjectMember, ProjectScope("project-1")))

	if !evaluator.Can(principal, PermissionCommentsWrite, ProjectScope("project-1")) {
		t.Fatal("expected group project member to write comments")
	}
	if evaluator.Can(principal, PermissionUsersWrite, GlobalScope()) {
		t.Fatal("expected project group role to be denied global user writes")
	}
}

func TestScopeBehavior(t *testing.T) {
	evaluator := NewInMemoryEvaluator()

	projectPrincipal := Principal{UserID: "project-user", AuthKind: AuthKindSession}
	evaluator.BindRole(UserBinding("project-user", RoleProjectAdmin, ProjectScope("project-1")))

	if !evaluator.Can(projectPrincipal, PermissionTicketsWrite, ProjectScope("project-1")) {
		t.Fatal("expected project role to apply to its project")
	}
	if evaluator.Can(projectPrincipal, PermissionTicketsWrite, ProjectScope("project-2")) {
		t.Fatal("expected project role to be denied in another project")
	}
	if evaluator.Can(projectPrincipal, PermissionTicketsRead, GlobalScope()) {
		t.Fatal("expected project role to be denied at global scope")
	}

	globalPrincipal := Principal{UserID: "global-user", AuthKind: AuthKindSession}
	evaluator.BindRole(UserBinding("global-user", RoleGlobalAdmin, GlobalScope()))

	if !evaluator.Can(globalPrincipal, PermissionSettingsManage, GlobalScope()) {
		t.Fatal("expected global admin to manage global settings")
	}
	if !evaluator.Can(globalPrincipal, PermissionTicketsWrite, ProjectScope("project-2")) {
		t.Fatal("expected global binding to apply to project checks")
	}
}

func TestWildcardPermissions(t *testing.T) {
	evaluator := NewInMemoryEvaluator()
	principal := Principal{UserID: "user-1", AuthKind: AuthKindSession}

	evaluator.BindRole(UserBinding("user-1", RoleProjectAdmin, ProjectScope("project-1")))

	if !evaluator.Can(principal, PermissionTicketsRead, ProjectScope("project-1")) {
		t.Fatal("expected tickets:* to match tickets:read")
	}
	if !PermissionMatches(PermissionTicketsWildcard, Permission("tickets:delete")) {
		t.Fatal("expected tickets:* to match another ticket action")
	}
	if PermissionMatches(PermissionTicketsWildcard, PermissionUsersRead) {
		t.Fatal("expected tickets:* not to match users:read")
	}
}

func TestDisabledUsersAreDenied(t *testing.T) {
	evaluator := NewInMemoryEvaluator()
	principal := Principal{UserID: "user-1", AuthKind: AuthKindSession}

	evaluator.BindRole(UserBinding("user-1", RoleGlobalAdmin, GlobalScope()))

	if !evaluator.Can(principal, PermissionUsersWrite, GlobalScope()) {
		t.Fatal("expected enabled global admin to write users")
	}

	disabledPrincipal := principal
	disabledPrincipal.Disabled = true
	if evaluator.Can(disabledPrincipal, PermissionUsersWrite, GlobalScope()) {
		t.Fatal("expected disabled principal to be denied")
	}

	evaluator.SetUserDisabled("user-1", true)
	if evaluator.Can(principal, PermissionUsersWrite, GlobalScope()) {
		t.Fatal("expected evaluator-disabled user to be denied")
	}
	if permissions := evaluator.EffectivePermissions("user-1", GlobalScope()); len(permissions) != 0 {
		t.Fatalf("expected disabled user to have no effective permissions, got %#v", permissions)
	}
}

func TestDenyByDefault(t *testing.T) {
	evaluator := NewInMemoryEvaluator()
	principal := Principal{UserID: "user-1", AuthKind: AuthKindSession}

	if evaluator.Can(principal, PermissionUsersRead, GlobalScope()) {
		t.Fatal("expected unbound user to be denied")
	}
	if evaluator.Can(Principal{}, PermissionUsersRead, GlobalScope()) {
		t.Fatal("expected empty principal to be denied")
	}
	if evaluator.Can(principal, PermissionUsersRead, Scope{}) {
		t.Fatal("expected invalid scope to be denied")
	}
	if err := evaluator.Require(principal, PermissionUsersRead, GlobalScope()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if permissions := evaluator.EffectivePermissions("user-1", GlobalScope()); len(permissions) != 0 {
		t.Fatalf("expected no effective permissions by default, got %#v", permissions)
	}
}

func TestBuiltInRoleDefaults(t *testing.T) {
	wantRoles := []RoleName{
		RoleGlobalAdmin,
		RoleGlobalUserManager,
		RoleProjectOwner,
		RoleProjectAdmin,
		RoleProjectMember,
		RoleProjectViewer,
		RoleAutomationManager,
		RoleNotificationManager,
	}

	for _, roleName := range wantRoles {
		role, ok := BuiltInRole(roleName)
		if !ok {
			t.Fatalf("missing built-in role %s", roleName)
		}
		if !role.BuiltIn {
			t.Fatalf("expected %s to be marked built-in", roleName)
		}
		if len(role.Permissions) == 0 {
			t.Fatalf("expected %s to have default permissions", roleName)
		}
	}

	admin, ok := BuiltInRole(RoleGlobalAdmin)
	if !ok {
		t.Fatal("missing global admin role")
	}
	if len(admin.Permissions) != 1 || admin.Permissions[0] != PermissionAll {
		t.Fatalf("expected global admin to grant all permissions, got %#v", admin.Permissions)
	}
}
