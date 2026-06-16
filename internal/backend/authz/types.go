package authz

// AuthKind identifies how a principal was authenticated. Scheduler,
// automation, and webhook actors still authorize as their configured user.
type AuthKind string

const (
	AuthKindSession           AuthKind = "session"
	AuthKindAPIToken          AuthKind = "api_token"
	AuthKindCron              AuthKind = "cron"
	AuthKindWebhook           AuthKind = "webhook"
	AuthKindInternalAdminDemo AuthKind = "internal_admin_demo"
)

// Principal is the authenticated actor presented to the evaluator.
type Principal struct {
	UserID      string
	AuthKind    AuthKind
	ActorUserID string
	Disabled    bool
}

// Permission is a namespace/action authorization string such as users:read.
type Permission string

// RoleName is the stable name of a role.
type RoleName string

// Role maps a role name to its granted permissions.
type Role struct {
	Name        RoleName
	Permissions []Permission
	BuiltIn     bool
}

// ScopeKind identifies the authorization scope namespace.
type ScopeKind string

const (
	ScopeKindGlobal  ScopeKind = "global"
	ScopeKindProject ScopeKind = "project"
)

// Scope describes where a permission check is being performed.
type Scope struct {
	Kind      ScopeKind
	ProjectID string
}

// GlobalScope returns the global authorization scope.
func GlobalScope() Scope {
	return Scope{Kind: ScopeKindGlobal}
}

// ProjectScope returns an authorization scope for a project.
func ProjectScope(projectID string) Scope {
	return Scope{Kind: ScopeKindProject, ProjectID: projectID}
}

func (s Scope) valid() bool {
	switch s.Kind {
	case ScopeKindGlobal:
		return s.ProjectID == ""
	case ScopeKindProject:
		return s.ProjectID != ""
	default:
		return false
	}
}

func (s Scope) appliesTo(requested Scope) bool {
	if !s.valid() || !requested.valid() {
		return false
	}
	if s.Kind == ScopeKindGlobal {
		return true
	}
	return requested.Kind == ScopeKindProject && s.ProjectID == requested.ProjectID
}

// BindingTargetKind identifies whether a binding applies directly to a user or
// indirectly through a group.
type BindingTargetKind string

const (
	BindingTargetUser  BindingTargetKind = "user"
	BindingTargetGroup BindingTargetKind = "group"
)

// Binding grants a role to a user or group at a global or project scope.
type Binding struct {
	TargetKind BindingTargetKind
	TargetID   string
	RoleName   RoleName
	Scope      Scope
}

// UserBinding creates a role binding for a user.
func UserBinding(userID string, role RoleName, scope Scope) Binding {
	return Binding{
		TargetKind: BindingTargetUser,
		TargetID:   userID,
		RoleName:   role,
		Scope:      scope,
	}
}

// GroupBinding creates a role binding for a group.
func GroupBinding(groupID string, role RoleName, scope Scope) Binding {
	return Binding{
		TargetKind: BindingTargetGroup,
		TargetID:   groupID,
		RoleName:   role,
		Scope:      scope,
	}
}

// GroupMembership adds a user to a group for indirect role bindings.
type GroupMembership struct {
	UserID  string
	GroupID string
}

// Evaluator is the single authorization surface consumed by services.
type Evaluator interface {
	Can(principal Principal, permission Permission, scope Scope) bool
	Require(principal Principal, permission Permission, scope Scope) error
	EffectivePermissions(userID string, scope Scope) []Permission
}
