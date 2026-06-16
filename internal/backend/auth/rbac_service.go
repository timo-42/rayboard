package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type CreateGroupInput struct {
	Name        string
	DisplayName string
}

type Role struct {
	ID          string             `json:"id"`
	Name        authz.RoleName     `json:"name"`
	Description string             `json:"description"`
	Permissions []authz.Permission `json:"permissions"`
}

type RoleBinding struct {
	ID           string                  `json:"id"`
	RoleID       string                  `json:"role_id"`
	RoleName     authz.RoleName          `json:"role_name"`
	SubjectType  authz.BindingTargetKind `json:"subject_type"`
	SubjectID    string                  `json:"subject_id"`
	ResourceType string                  `json:"resource_type"`
	ResourceID   string                  `json:"resource_id,omitempty"`
}

type CreateRoleBindingInput struct {
	RoleName    authz.RoleName
	SubjectType authz.BindingTargetKind
	SubjectID   string
	Scope       authz.Scope
}

func (s *Service) CreateGroup(ctx context.Context, input CreateGroupInput) (Group, error) {
	input.Name = normalizeName(input.Name)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.DisplayName == "" {
		input.DisplayName = input.Name
	}

	fields := make(map[string]string)
	if input.Name == "" {
		fields["name"] = "Required"
	}
	if input.DisplayName == "" {
		fields["display_name"] = "Required"
	}
	if len(fields) > 0 {
		return Group{}, &ValidationError{Message: "Invalid group", Fields: fields}
	}

	var exists bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM groups WHERE name = ?)
	`, input.Name).Scan(&exists); err != nil {
		return Group{}, fmt.Errorf("check group name: %w", err)
	}
	if exists {
		return Group{}, fmt.Errorf("%w: group name already exists", ErrConflict)
	}

	group := Group{
		ID:          newID("group"),
		Name:        input.Name,
		DisplayName: input.DisplayName,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO groups (id, name, display_name, updated_at)
		VALUES (?, ?, ?, ?)
	`, group.ID, group.Name, group.DisplayName, formatTime(s.now())); err != nil {
		return Group{}, fmt.Errorf("insert group: %w", err)
	}
	return group, nil
}

func (s *Service) ListGroups(ctx context.Context) ([]Group, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, display_name
		FROM groups
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.Name, &group.DisplayName); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate groups: %w", err)
	}
	return groups, nil
}

func (s *Service) AddGroupMember(ctx context.Context, groupID string, userID string) error {
	if err := s.ensureGroupExists(ctx, groupID); err != nil {
		return err
	}
	if _, err := s.GetUser(ctx, userID); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO group_memberships (group_id, user_id)
		VALUES (?, ?)
		ON CONFLICT(group_id, user_id) DO NOTHING
	`, groupID, userID); err != nil {
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, groupID string, userID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM group_memberships
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("remove group member: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check removed group member: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) ListGroupMembers(ctx context.Context, groupID string) ([]User, error) {
	if err := s.ensureGroupExists(ctx, groupID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.display_name, u.is_disabled
		FROM group_memberships gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ? AND u.deleted_at IS NULL
		ORDER BY u.username ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Disabled); err != nil {
			return nil, fmt.Errorf("scan group member: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group members: %w", err)
	}
	return users, nil
}

func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description, '')
		FROM roles
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		permissions, err := s.rolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = permissions
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return roles, nil
}

func (s *Service) CreateRoleBinding(ctx context.Context, input CreateRoleBindingInput) (RoleBinding, error) {
	fields := make(map[string]string)
	if input.RoleName == "" {
		fields["role_name"] = "Required"
	}
	if input.SubjectType != authz.BindingTargetUser && input.SubjectType != authz.BindingTargetGroup {
		fields["subject_type"] = "Must be user or group"
	}
	if strings.TrimSpace(input.SubjectID) == "" {
		fields["subject_id"] = "Required"
	}
	resourceType, resourceID, scopeOK := resourceFromScope(input.Scope)
	if !scopeOK {
		fields["scope"] = "Must be global or project"
	}
	if len(fields) > 0 {
		return RoleBinding{}, &ValidationError{Message: "Invalid role binding", Fields: fields}
	}

	roleID, err := s.roleID(ctx, input.RoleName)
	if err != nil {
		return RoleBinding{}, err
	}
	if err := s.ensureSubjectExists(ctx, input.SubjectType, input.SubjectID); err != nil {
		return RoleBinding{}, err
	}

	binding := RoleBinding{
		ID:           newID("binding"),
		RoleID:       roleID,
		RoleName:     input.RoleName,
		SubjectType:  input.SubjectType,
		SubjectID:    input.SubjectID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO role_bindings (id, role_id, subject_type, subject_id, resource_type, resource_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, binding.ID, binding.RoleID, string(binding.SubjectType), binding.SubjectID, binding.ResourceType, nullableString(binding.ResourceID)); err != nil {
		return RoleBinding{}, fmt.Errorf("insert role binding: %w", err)
	}
	return binding, nil
}

func (s *Service) ListRoleBindings(ctx context.Context) ([]RoleBinding, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT rb.id, rb.role_id, r.name, rb.subject_type, rb.subject_id, rb.resource_type, rb.resource_id
		FROM role_bindings rb
		JOIN roles r ON r.id = rb.role_id
		ORDER BY rb.created_at DESC, rb.id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list role bindings: %w", err)
	}
	defer rows.Close()

	var bindings []RoleBinding
	for rows.Next() {
		binding, err := scanRoleBinding(rows)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role bindings: %w", err)
	}
	return bindings, nil
}

func (s *Service) DeleteRoleBinding(ctx context.Context, bindingID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM role_bindings
		WHERE id = ?
	`, bindingID)
	if err != nil {
		return fmt.Errorf("delete role binding: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted role binding: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) rolePermissions(ctx context.Context, roleID string) ([]authz.Permission, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT permission
		FROM role_permissions
		WHERE role_id = ?
		ORDER BY permission ASC
	`, roleID)
	if err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []authz.Permission
	for rows.Next() {
		var permission authz.Permission
		if err := rows.Scan(&permission); err != nil {
			return nil, fmt.Errorf("scan role permission: %w", err)
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role permissions: %w", err)
	}
	return permissions, nil
}

func (s *Service) roleID(ctx context.Context, roleName authz.RoleName) (string, error) {
	var roleID string
	if err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM roles
		WHERE name = ?
	`, string(roleName)).Scan(&roleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("query role: %w", err)
	}
	return roleID, nil
}

func (s *Service) ensureSubjectExists(ctx context.Context, subjectType authz.BindingTargetKind, subjectID string) error {
	switch subjectType {
	case authz.BindingTargetUser:
		_, err := s.GetUser(ctx, subjectID)
		return err
	case authz.BindingTargetGroup:
		return s.ensureGroupExists(ctx, subjectID)
	default:
		return &ValidationError{
			Message: "Invalid subject",
			Fields:  map[string]string{"subject_type": "Must be user or group"},
		}
	}
}

func (s *Service) ensureGroupExists(ctx context.Context, groupID string) error {
	var exists bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)
	`, groupID).Scan(&exists); err != nil {
		return fmt.Errorf("check group: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func scanRoleBinding(rows interface {
	Scan(dest ...any) error
}) (RoleBinding, error) {
	var binding RoleBinding
	var resourceType sql.NullString
	var resourceID sql.NullString
	if err := rows.Scan(
		&binding.ID,
		&binding.RoleID,
		&binding.RoleName,
		&binding.SubjectType,
		&binding.SubjectID,
		&resourceType,
		&resourceID,
	); err != nil {
		return RoleBinding{}, fmt.Errorf("scan role binding: %w", err)
	}
	if resourceType.Valid {
		binding.ResourceType = resourceType.String
	}
	if resourceID.Valid {
		binding.ResourceID = resourceID.String
	}
	return binding, nil
}

func resourceFromScope(scope authz.Scope) (string, string, bool) {
	switch scope.Kind {
	case authz.ScopeKindGlobal:
		if scope.ProjectID != "" {
			return "", "", false
		}
		return string(authz.ScopeKindGlobal), "", true
	case authz.ScopeKindProject:
		if scope.ProjectID == "" {
			return "", "", false
		}
		return string(authz.ScopeKindProject), scope.ProjectID, true
	default:
		return "", "", false
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
