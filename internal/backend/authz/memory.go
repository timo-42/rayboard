package authz

import (
	"errors"
	"sort"
	"sync"
)

var ErrForbidden = errors.New("authz: forbidden")

// InMemoryOption configures an in-memory evaluator at construction time.
type InMemoryOption func(*InMemoryEvaluator)

// InMemoryEvaluator evaluates RBAC state held in memory. It is intended for
// tests and early service integration; production storage can implement the
// Evaluator interface directly.
type InMemoryEvaluator struct {
	mu sync.RWMutex

	roles            map[RoleName]Role
	bindings         []Binding
	groupMemberships []GroupMembership
	disabledUsers    map[string]bool
}

// NewInMemoryEvaluator creates an evaluator seeded with built-in roles.
func NewInMemoryEvaluator(options ...InMemoryOption) *InMemoryEvaluator {
	evaluator := &InMemoryEvaluator{
		roles:         make(map[RoleName]Role),
		disabledUsers: make(map[string]bool),
	}
	for _, role := range BuiltInRoles() {
		evaluator.roles[role.Name] = role
	}
	for _, option := range options {
		option(evaluator)
	}
	return evaluator
}

// WithRoles adds or replaces roles in a new in-memory evaluator.
func WithRoles(roles ...Role) InMemoryOption {
	return func(evaluator *InMemoryEvaluator) {
		for _, role := range roles {
			evaluator.setRoleLocked(role)
		}
	}
}

// WithBindings adds bindings to a new in-memory evaluator.
func WithBindings(bindings ...Binding) InMemoryOption {
	return func(evaluator *InMemoryEvaluator) {
		evaluator.bindings = append(evaluator.bindings, cloneBindings(bindings)...)
	}
}

// WithGroupMemberships adds group memberships to a new in-memory evaluator.
func WithGroupMemberships(memberships ...GroupMembership) InMemoryOption {
	return func(evaluator *InMemoryEvaluator) {
		evaluator.groupMemberships = append(evaluator.groupMemberships, cloneMemberships(memberships)...)
	}
}

// WithDisabledUsers marks users disabled in a new in-memory evaluator.
func WithDisabledUsers(userIDs ...string) InMemoryOption {
	return func(evaluator *InMemoryEvaluator) {
		for _, userID := range userIDs {
			if userID != "" {
				evaluator.disabledUsers[userID] = true
			}
		}
	}
}

// SetRole adds or replaces a role.
func (e *InMemoryEvaluator) SetRole(role Role) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.setRoleLocked(role)
}

func (e *InMemoryEvaluator) setRoleLocked(role Role) {
	if role.Name == "" {
		return
	}
	e.roles[role.Name] = cloneRole(role)
}

// BindRole appends a role binding.
func (e *InMemoryEvaluator) BindRole(binding Binding) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bindings = append(e.bindings, binding)
}

// SetBindings replaces all role bindings.
func (e *InMemoryEvaluator) SetBindings(bindings []Binding) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bindings = cloneBindings(bindings)
}

// AddGroupMembership appends one group membership.
func (e *InMemoryEvaluator) AddGroupMembership(membership GroupMembership) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.groupMemberships = append(e.groupMemberships, membership)
}

// SetGroupMemberships replaces all group memberships.
func (e *InMemoryEvaluator) SetGroupMemberships(memberships []GroupMembership) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.groupMemberships = cloneMemberships(memberships)
}

// SetUserDisabled controls whether a user is denied by this evaluator.
func (e *InMemoryEvaluator) SetUserDisabled(userID string, disabled bool) {
	if e == nil || userID == "" {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if disabled {
		e.disabledUsers[userID] = true
		return
	}
	delete(e.disabledUsers, userID)
}

// Can reports whether a principal has permission in scope.
func (e *InMemoryEvaluator) Can(principal Principal, permission Permission, scope Scope) bool {
	if e == nil || principal.UserID == "" || principal.Disabled || !scope.valid() {
		return false
	}
	permission = NormalizePermission(permission)
	if permission == "" {
		return false
	}
	for _, granted := range e.EffectivePermissions(principal.UserID, scope) {
		if PermissionMatches(granted, permission) {
			return true
		}
	}
	return false
}

// Require returns ErrForbidden unless a principal has permission in scope.
func (e *InMemoryEvaluator) Require(principal Principal, permission Permission, scope Scope) error {
	if e.Can(principal, permission, scope) {
		return nil
	}
	return ErrForbidden
}

// EffectivePermissions returns the stored permissions granted to a user in
// scope, including direct user bindings and group bindings.
func (e *InMemoryEvaluator) EffectivePermissions(userID string, scope Scope) []Permission {
	if e == nil || userID == "" || !scope.valid() {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.disabledUsers[userID] {
		return nil
	}

	groups := make(map[string]struct{})
	for _, membership := range e.groupMemberships {
		if membership.UserID == userID && membership.GroupID != "" {
			groups[membership.GroupID] = struct{}{}
		}
	}

	permissions := make(map[Permission]struct{})
	for _, binding := range e.bindings {
		if !binding.Scope.appliesTo(scope) || !bindingMatchesUser(binding, userID, groups) {
			continue
		}
		role, ok := e.roles[binding.RoleName]
		if !ok {
			continue
		}
		for _, permission := range role.Permissions {
			permission = NormalizePermission(permission)
			if permission != "" {
				permissions[permission] = struct{}{}
			}
		}
	}

	result := make([]Permission, 0, len(permissions))
	for permission := range permissions {
		result = append(result, permission)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func bindingMatchesUser(binding Binding, userID string, groups map[string]struct{}) bool {
	switch binding.TargetKind {
	case BindingTargetUser:
		return binding.TargetID == userID
	case BindingTargetGroup:
		_, ok := groups[binding.TargetID]
		return ok
	default:
		return false
	}
}

func cloneBindings(bindings []Binding) []Binding {
	cloned := make([]Binding, len(bindings))
	copy(cloned, bindings)
	return cloned
}

func cloneMemberships(memberships []GroupMembership) []GroupMembership {
	cloned := make([]GroupMembership, len(memberships))
	copy(cloned, memberships)
	return cloned
}
