package authz

import (
	"sort"
	"strings"
)

const (
	RoleGlobalAdmin         RoleName = "global_admin"
	RoleGlobalUserManager   RoleName = "global_user_manager"
	RoleProjectOwner        RoleName = "project_owner"
	RoleProjectAdmin        RoleName = "project_admin"
	RoleProjectMember       RoleName = "project_member"
	RoleProjectViewer       RoleName = "project_viewer"
	RoleAutomationManager   RoleName = "automation_manager"
	RoleNotificationManager RoleName = "notification_manager"
)

const (
	PermissionAll Permission = "*"

	PermissionUsersRead   Permission = "users:read"
	PermissionUsersWrite  Permission = "users:write"
	PermissionGroupsRead  Permission = "groups:read"
	PermissionGroupsWrite Permission = "groups:write"
	PermissionRolesRead   Permission = "roles:read"
	PermissionRolesBind   Permission = "roles:bind"

	PermissionProjectsRead  Permission = "projects:read"
	PermissionProjectsWrite Permission = "projects:write"

	PermissionTicketsRead     Permission = "tickets:read"
	PermissionTicketsWrite    Permission = "tickets:write"
	PermissionTicketsWildcard Permission = "tickets:*"

	PermissionCommentsWrite    Permission = "comments:write"
	PermissionAttachmentsWrite Permission = "attachments:write"
	PermissionSprintsManage    Permission = "sprints:manage"
	PermissionBoardsManage     Permission = "boards:manage"
	PermissionFieldsManage     Permission = "fields:manage"
	PermissionViewsManage      Permission = "views:manage"

	PermissionNotificationsManage Permission = "notifications:manage"
	PermissionWebhooksManage      Permission = "webhooks:manage"
	PermissionAutomationsManage   Permission = "automations:manage"
	PermissionSettingsManage      Permission = "settings:manage"
	PermissionDemoReset           Permission = "demo:reset"
	PermissionAIManage            Permission = "ai:manage"
)

var defaultBuiltInRoles = map[RoleName]Role{
	RoleGlobalAdmin: {
		Name:        RoleGlobalAdmin,
		Permissions: []Permission{PermissionAll},
		BuiltIn:     true,
	},
	RoleGlobalUserManager: {
		Name: RoleGlobalUserManager,
		Permissions: []Permission{
			PermissionUsersRead,
			PermissionUsersWrite,
			PermissionGroupsRead,
			PermissionGroupsWrite,
			PermissionRolesRead,
			PermissionRolesBind,
		},
		BuiltIn: true,
	},
	RoleProjectOwner: {
		Name: RoleProjectOwner,
		Permissions: []Permission{
			PermissionProjectsRead,
			PermissionProjectsWrite,
			PermissionRolesRead,
			PermissionRolesBind,
			PermissionTicketsWildcard,
			PermissionCommentsWrite,
			PermissionAttachmentsWrite,
			PermissionSprintsManage,
			PermissionBoardsManage,
			PermissionFieldsManage,
			PermissionViewsManage,
			PermissionNotificationsManage,
			PermissionWebhooksManage,
			PermissionAutomationsManage,
			PermissionSettingsManage,
			PermissionAIManage,
		},
		BuiltIn: true,
	},
	RoleProjectAdmin: {
		Name: RoleProjectAdmin,
		Permissions: []Permission{
			PermissionProjectsRead,
			PermissionRolesRead,
			PermissionRolesBind,
			PermissionTicketsWildcard,
			PermissionCommentsWrite,
			PermissionAttachmentsWrite,
			PermissionSprintsManage,
			PermissionBoardsManage,
			PermissionFieldsManage,
			PermissionViewsManage,
			PermissionNotificationsManage,
			PermissionWebhooksManage,
			PermissionAutomationsManage,
		},
		BuiltIn: true,
	},
	RoleProjectMember: {
		Name: RoleProjectMember,
		Permissions: []Permission{
			PermissionProjectsRead,
			PermissionTicketsRead,
			PermissionTicketsWrite,
			PermissionCommentsWrite,
			PermissionAttachmentsWrite,
		},
		BuiltIn: true,
	},
	RoleProjectViewer: {
		Name: RoleProjectViewer,
		Permissions: []Permission{
			PermissionProjectsRead,
			PermissionTicketsRead,
		},
		BuiltIn: true,
	},
	RoleAutomationManager: {
		Name: RoleAutomationManager,
		Permissions: []Permission{
			PermissionProjectsRead,
			PermissionAutomationsManage,
			PermissionWebhooksManage,
		},
		BuiltIn: true,
	},
	RoleNotificationManager: {
		Name: RoleNotificationManager,
		Permissions: []Permission{
			PermissionProjectsRead,
			PermissionNotificationsManage,
		},
		BuiltIn: true,
	},
}

// BuiltInRoles returns the built-in v1 roles with their default permissions.
func BuiltInRoles() []Role {
	roles := make([]Role, 0, len(defaultBuiltInRoles))
	for _, role := range defaultBuiltInRoles {
		roles = append(roles, cloneRole(role))
	}
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})
	return roles
}

// BuiltInRole returns one built-in role by name.
func BuiltInRole(name RoleName) (Role, bool) {
	role, ok := defaultBuiltInRoles[name]
	if !ok {
		return Role{}, false
	}
	return cloneRole(role), true
}

// NormalizePermission makes permission checks consistent across stored grants
// and requested checks.
func NormalizePermission(permission Permission) Permission {
	return Permission(strings.ToLower(strings.TrimSpace(string(permission))))
}

// PermissionMatches reports whether a granted permission satisfies a requested
// permission. Wildcards only apply when they are stored in the granted side.
func PermissionMatches(granted Permission, requested Permission) bool {
	granted = NormalizePermission(granted)
	requested = NormalizePermission(requested)
	if granted == "" || requested == "" {
		return false
	}
	if granted == PermissionAll || granted == requested {
		return true
	}
	grantedText := string(granted)
	if !strings.HasSuffix(grantedText, ":*") {
		return false
	}
	prefix := strings.TrimSuffix(grantedText, "*")
	return strings.HasPrefix(string(requested), prefix)
}

func cloneRole(role Role) Role {
	permissions := make([]Permission, len(role.Permissions))
	copy(permissions, role.Permissions)
	role.Permissions = permissions
	return role
}
