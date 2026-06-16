package authz

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

type SQLEvaluator struct {
	db *sql.DB
}

func NewSQLEvaluator(db *sql.DB) *SQLEvaluator {
	return &SQLEvaluator{db: db}
}

func (e *SQLEvaluator) Can(principal Principal, permission Permission, scope Scope) bool {
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

func (e *SQLEvaluator) Require(principal Principal, permission Permission, scope Scope) error {
	if e.Can(principal, permission, scope) {
		return nil
	}
	return ErrForbidden
}

func (e *SQLEvaluator) EffectivePermissions(userID string, scope Scope) []Permission {
	permissions, err := e.effectivePermissions(context.Background(), userID, scope)
	if err != nil {
		return nil
	}
	return permissions
}

func (e *SQLEvaluator) effectivePermissions(ctx context.Context, userID string, scope Scope) ([]Permission, error) {
	if e == nil || userID == "" || !scope.valid() {
		return nil, nil
	}

	var disabled bool
	if err := e.db.QueryRowContext(ctx, `
		SELECT is_disabled
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&disabled); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query authorization user: %w", err)
	}
	if disabled {
		return nil, nil
	}

	rows, err := e.db.QueryContext(ctx, `
		SELECT DISTINCT rp.permission, rb.resource_type, rb.resource_id
		FROM role_bindings rb
		JOIN role_permissions rp ON rp.role_id = rb.role_id
		WHERE (rb.subject_type = 'user' AND rb.subject_id = ?)
		   OR (
		        rb.subject_type = 'group'
		        AND rb.subject_id IN (
		            SELECT group_id
		            FROM group_memberships
		            WHERE user_id = ?
		        )
		   )
	`, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("query effective permissions: %w", err)
	}
	defer rows.Close()

	set := make(map[Permission]struct{})
	for rows.Next() {
		var permission Permission
		var resourceType sql.NullString
		var resourceID sql.NullString
		if err := rows.Scan(&permission, &resourceType, &resourceID); err != nil {
			return nil, fmt.Errorf("scan effective permission: %w", err)
		}
		bindingScope := scopeFromResource(resourceType, resourceID)
		if !bindingScope.appliesTo(scope) {
			continue
		}
		permission = NormalizePermission(permission)
		if permission != "" {
			set[permission] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate effective permissions: %w", err)
	}

	result := make([]Permission, 0, len(set))
	for permission := range set {
		result = append(result, permission)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result, nil
}

func scopeFromResource(resourceType sql.NullString, resourceID sql.NullString) Scope {
	if !resourceType.Valid || resourceType.String == "" {
		return GlobalScope()
	}
	if resourceType.String == string(ScopeKindProject) && resourceID.Valid && resourceID.String != "" {
		return ProjectScope(resourceID.String)
	}
	return Scope{}
}
