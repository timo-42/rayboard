package authapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.PublicOperation(http.MethodPost, "/api/login", "Auth", "Log in with username and password"), provider.login)
	registerLogout(api, provider)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/me", "Auth", "Get current authenticated user"), provider.me)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/me/effective-permissions", "Auth", "Get current effective permissions"), provider.myEffectivePermissions)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tokens", "Auth", "List API tokens"), provider.listTokens)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/tokens", "Auth", "Create API token", http.StatusCreated), provider.createToken)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/tokens/{token_id}", "Auth", "Revoke API token", http.StatusNoContent), provider.revokeToken)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/users", "Users", "List users"), provider.listUsers)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/users", "Users", "Create user", http.StatusCreated), provider.createUser)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/users/{user_id}", "Users", "Get user"), provider.getUser)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/users/{user_id}/effective-permissions", "RBAC", "Get user effective permissions"), provider.userEffectivePermissions)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/users/{user_id}", "Users", "Update user"), provider.updateUser)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/users/{user_id}", "Users", "Delete user", http.StatusNoContent), provider.deleteUser)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/groups", "RBAC", "List groups"), provider.listGroups)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/groups", "RBAC", "Create group", http.StatusCreated), provider.createGroup)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/groups/{group_id}/members", "RBAC", "List group members"), provider.listGroupMembers)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/groups/{group_id}/members/{user_id}", "RBAC", "Add group member", http.StatusNoContent), provider.addGroupMember)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/groups/{group_id}/members/{user_id}", "RBAC", "Remove group member", http.StatusNoContent), provider.removeGroupMember)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/roles", "RBAC", "List roles"), provider.listRoles)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/roles/{role_name}", "RBAC", "Get role"), provider.getRole)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/role-bindings", "RBAC", "List role bindings"), provider.listRoleBindings)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/role-bindings", "RBAC", "Create role binding", http.StatusCreated), provider.createRoleBinding)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/role-bindings/{binding_id}", "RBAC", "Delete role binding", http.StatusNoContent), provider.deleteRoleBinding)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/members", "RBAC", "List project members"), provider.listProjectMembers)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/role-bindings", "RBAC", "List project role bindings"), provider.listProjectRoleBindings)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/role-bindings", "RBAC", "Create project role binding", http.StatusCreated), provider.createProjectRoleBinding)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/projects/{project_id}/role-bindings/{binding_id}", "RBAC", "Delete project role binding", http.StatusNoContent), provider.deleteProjectRoleBinding)
}

func (provider Provider) login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	session, err := provider.Auth.Login(ctx, input.Body.Spec.Username, input.Body.Spec.Password)
	if err != nil {
		_ = provider.recordAudit(ctx, audit.RecordInput{
			EventType:   "auth.login_failed",
			SubjectType: "user",
			SubjectID:   input.Body.Spec.Username,
			Outcome:     audit.OutcomeFailure,
			Payload: map[string]any{
				"username":    input.Body.Spec.Username,
				"auth_method": "password",
				"reason":      authFailureReason(err),
			},
		})
		return nil, shared.AuthError(err)
	}
	csrf, err := randomURLToken()
	if err != nil {
		return nil, huma.Error500InternalServerError("Could not create CSRF token")
	}
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "auth.session_created",
		ActorID:     session.User.ID,
		AuthKind:    authz.AuthKindSession,
		SubjectType: "session",
		SubjectID:   session.ID,
		Payload: map[string]any{
			"user_id":     session.User.ID,
			"username":    session.User.Username,
			"auth_method": "password",
			"expires_at":  session.ExpiresAt,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &LoginOutput{
		SetCookie: []http.Cookie{
			sessionCookie(session),
			csrfCookie(csrf, session.ExpiresAt),
		},
		Body: sessionResource(session.User, authz.Principal{
			UserID:   session.User.ID,
			AuthKind: authz.AuthKindSession,
			Disabled: session.User.Disabled,
		}),
	}, nil
}

func registerLogout(api huma.API, provider Provider) {
	op := shared.PublicOperationWithStatus(http.MethodPost, "/api/logout", "Auth", "Log out current session", http.StatusNoContent)
	api.OpenAPI().AddOperation(&op)
	api.Adapter().Handle(&op, api.Middlewares().Handler(op.Middlewares.Handler(func(ctx huma.Context) {
		provider.logout(api, ctx)
	})))
}

func (provider Provider) logout(api huma.API, ctx huma.Context) {
	r, _ := humago.Unwrap(ctx)
	if r == nil {
		_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "Could not log out")
		return
	}
	sessionCookie, sessionErr := r.Cookie(auth.SessionCookieName)
	if sessionErr == nil && sessionCookie.Value != "" {
		principal, _, authErr := provider.Auth.AuthenticateSession(ctx.Context(), sessionCookie.Value)
		csrfCookie, csrfErr := r.Cookie(shared.CSRFCookieName)
		if authErr == nil && principal.AuthKind == authz.AuthKindSession && (csrfErr != nil || csrfCookie.Value == "" || csrfCookie.Value != r.Header.Get("X-CSRF-Token")) {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "CSRF token is required")
			return
		}
		if err := provider.Auth.Logout(ctx.Context(), sessionCookie.Value); err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "Could not log out")
			return
		}
		if authErr == nil {
			actorID, authKind := auditActor(principal)
			if err := provider.recordAudit(ctx.Context(), audit.RecordInput{
				EventType:   "auth.session_revoked",
				ActorID:     actorID,
				AuthKind:    authKind,
				SubjectType: "session",
				Payload: map[string]any{
					"reason": "logout",
				},
			}); err != nil {
				_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "Could not write audit log")
				return
			}
		}
	}
	ctx.SetStatus(http.StatusNoContent)
}

func (provider Provider) me(ctx context.Context, input *MeInput) (*MeOutput, error) {
	ctx, principal, user, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	_ = ctx
	return &MeOutput{Body: sessionResource(user, principal)}, nil
}

func (provider Provider) myEffectivePermissions(ctx context.Context, input *EffectivePermissionsInput) (*EffectivePermissionsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	scope, ok := effectivePermissionsScope(input.Scope, input.ProjectID)
	if !ok {
		return nil, huma.Error400BadRequest("Invalid permission scope")
	}
	if provider.Authenticator.Authorizer == nil {
		return nil, huma.Error500InternalServerError("Authorization is not configured")
	}
	permissions := provider.Authenticator.Authorizer.EffectivePermissions(principal.UserID, scope)
	return &EffectivePermissionsOutput{Body: effectivePermissionsResource(principal.UserID, scope, permissions)}, nil
}

func (provider Provider) listTokens(ctx context.Context, input *struct{ shared.AuthInput }) (*ListTokensOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	tokens, err := provider.Auth.ListAPITokens(ctx, principal.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Could not list API tokens")
	}
	return &ListTokensOutput{Body: shared.NewListResource[TokenResource](tokenResources(tokens))}, nil
}

func (provider Provider) createToken(ctx context.Context, input *CreateTokenInput) (*CreateTokenOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	token, err := provider.Auth.CreateAPIToken(ctx, principal.UserID, input.Body.Spec.Name)
	if err != nil {
		return nil, huma.Error500InternalServerError("Could not create API token")
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "auth.api_token_created",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "api_token",
		SubjectID:   token.ID,
		Payload: map[string]any{
			"target_user_id": principal.UserID,
			"token_id":       token.ID,
			"token_name":     token.Name,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateTokenOutput{Body: createdTokenResource(token)}, nil
}

func (provider Provider) revokeToken(ctx context.Context, input *RevokeTokenInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Auth.RevokeAPIToken(ctx, principal.UserID, input.TokenID); err != nil {
		return nil, huma.Error500InternalServerError("Could not revoke API token")
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "auth.api_token_revoked",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "api_token",
		SubjectID:   input.TokenID,
		Payload: map[string]any{
			"target_user_id": principal.UserID,
			"token_id":       input.TokenID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listUsers(ctx context.Context, input *struct{ shared.AuthInput }) (*ListUsersOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionUsersRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	users, err := provider.Auth.ListUsers(ctx)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	return &ListUsersOutput{Body: shared.NewListResource[UserResource](userResources(users))}, nil
}

func (provider Provider) createUser(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionUsersWrite, authz.GlobalScope()); err != nil {
		return nil, err
	}
	user, err := provider.Auth.CreateUser(ctx, auth.CreateUserInput{
		Username:    input.Body.Spec.Username,
		DisplayName: input.Body.Spec.DisplayName,
		Password:    input.Body.Spec.Password,
		Disabled:    input.Body.Spec.Disabled,
	})
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "user.created",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "user",
		SubjectID:   user.ID,
		Payload: map[string]any{
			"username":           user.Username,
			"display_name":       user.DisplayName,
			"disabled":           user.Disabled,
			"password_generated": input.Body.Spec.Password == "",
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateUserOutput{Body: createdUserResource(user)}, nil
}

func (provider Provider) getUser(ctx context.Context, input *UserIDInput) (*UserOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionUsersRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	user, err := provider.Auth.GetUser(ctx, input.UserID)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	return &UserOutput{Body: userResource(user)}, nil
}

func (provider Provider) userEffectivePermissions(ctx context.Context, input *UserEffectivePermissionsInput) (*EffectivePermissionsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	scope, ok := effectivePermissionsScope(input.Scope, input.ProjectID)
	if !ok {
		return nil, huma.Error400BadRequest("Invalid permission scope")
	}
	if provider.Authenticator.Authorizer == nil {
		return nil, huma.Error500InternalServerError("Authorization is not configured")
	}
	if _, err := provider.Auth.GetUser(ctx, input.UserID); err != nil {
		return nil, shared.AuthServiceError(err)
	}
	permissions := provider.Authenticator.Authorizer.EffectivePermissions(input.UserID, scope)
	return &EffectivePermissionsOutput{Body: effectivePermissionsResource(input.UserID, scope, permissions)}, nil
}

func (provider Provider) updateUser(ctx context.Context, input *UpdateUserInput) (*UserOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionUsersWrite, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if input.Body.Spec.Disabled == nil {
		return nil, huma.Error400BadRequest("Invalid user update")
	}
	user, err := provider.Auth.SetUserDisabled(ctx, input.UserID, *input.Body.Spec.Disabled)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	eventType := "user.enabled"
	if user.Disabled {
		eventType = "user.disabled"
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   eventType,
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "user",
		SubjectID:   user.ID,
		Payload: map[string]any{
			"target_user_id":       user.ID,
			"disabled":             user.Disabled,
			"credentials_revoked":  user.Disabled,
			"target_user_username": user.Username,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &UserOutput{Body: userResource(user)}, nil
}

func (provider Provider) deleteUser(ctx context.Context, input *UserIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionUsersWrite, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if err := provider.Auth.DeleteUser(ctx, input.UserID); err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "user.deleted",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "user",
		SubjectID:   input.UserID,
		Payload: map[string]any{
			"target_user_id":      input.UserID,
			"deletion_mode":       "soft_delete",
			"credentials_revoked": true,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listGroups(ctx context.Context, input *struct{ shared.AuthInput }) (*ListGroupsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionGroupsRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	groups, err := provider.Auth.ListGroups(ctx)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	return &ListGroupsOutput{Body: shared.NewListResource[GroupResource](groupResources(groups))}, nil
}

func (provider Provider) createGroup(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionGroupsWrite, authz.GlobalScope()); err != nil {
		return nil, err
	}
	group, err := provider.Auth.CreateGroup(ctx, auth.CreateGroupInput{
		Name:        input.Body.Spec.Name,
		DisplayName: input.Body.Spec.DisplayName,
	})
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.group_created",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "group",
		SubjectID:   group.ID,
		Payload: map[string]any{
			"group_id":     group.ID,
			"name":         group.Name,
			"display_name": group.DisplayName,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateGroupOutput{Body: groupResource(group)}, nil
}

func (provider Provider) listGroupMembers(ctx context.Context, input *GroupIDInput) (*ListGroupMembersOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionGroupsRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	users, err := provider.Auth.ListGroupMembers(ctx, input.GroupID)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	return &ListGroupMembersOutput{Body: shared.NewListResource[UserResource](userResources(users))}, nil
}

func (provider Provider) addGroupMember(ctx context.Context, input *GroupMemberInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionGroupsWrite, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if err := provider.Auth.AddGroupMember(ctx, input.GroupID, input.UserID); err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.group_member_added",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "group_membership",
		SubjectID:   input.GroupID + ":" + input.UserID,
		Payload: map[string]any{
			"group_id":       input.GroupID,
			"target_user_id": input.UserID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) removeGroupMember(ctx context.Context, input *GroupMemberInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionGroupsWrite, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if err := provider.Auth.RemoveGroupMember(ctx, input.GroupID, input.UserID); err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.group_member_removed",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "group_membership",
		SubjectID:   input.GroupID + ":" + input.UserID,
		Payload: map[string]any{
			"group_id":       input.GroupID,
			"target_user_id": input.UserID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listRoles(ctx context.Context, input *struct{ shared.AuthInput }) (*ListRolesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	roles, err := provider.Auth.ListRoles(ctx)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	return &ListRolesOutput{Body: shared.NewListResource[RoleResource](roleResources(roles))}, nil
}

func (provider Provider) getRole(ctx context.Context, input *RoleNameInput) (*RoleOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	role, err := provider.roleByName(ctx, input.RoleName)
	if err != nil {
		return nil, err
	}
	return &RoleOutput{Body: roleResource(role)}, nil
}

func (provider Provider) listRoleBindings(ctx context.Context, input *struct{ shared.AuthInput }) (*ListRoleBindingsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesRead, authz.GlobalScope()); err != nil {
		return nil, err
	}
	bindings, err := provider.Auth.ListRoleBindings(ctx)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	return &ListRoleBindingsOutput{Body: shared.NewListResource[RoleBindingResource](roleBindingResources(bindings))}, nil
}

func (provider Provider) listProjectMembers(ctx context.Context, input *ProjectIDInput) (*ListGroupMembersOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesRead, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	bindings, err := provider.projectRoleBindings(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	usersByID := map[string]auth.User{}
	for _, binding := range bindings {
		switch binding.SubjectType {
		case authz.BindingTargetUser:
			user, err := provider.Auth.GetUser(ctx, binding.SubjectID)
			if err != nil {
				return nil, shared.AuthServiceError(err)
			}
			usersByID[user.ID] = user
		case authz.BindingTargetGroup:
			members, err := provider.Auth.ListGroupMembers(ctx, binding.SubjectID)
			if err != nil {
				return nil, shared.AuthServiceError(err)
			}
			for _, member := range members {
				usersByID[member.ID] = member
			}
		}
	}
	users := make([]auth.User, 0, len(usersByID))
	for _, user := range usersByID {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})
	return &ListGroupMembersOutput{Body: shared.NewListResource[UserResource](userResources(users))}, nil
}

func (provider Provider) listProjectRoleBindings(ctx context.Context, input *ProjectIDInput) (*ListRoleBindingsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesRead, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	bindings, err := provider.projectRoleBindings(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	return &ListRoleBindingsOutput{Body: shared.NewListResource[RoleBindingResource](roleBindingResources(bindings))}, nil
}

func (provider Provider) createRoleBinding(ctx context.Context, input *CreateRoleBindingInput) (*CreateRoleBindingOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesBind, authz.GlobalScope()); err != nil {
		return nil, err
	}
	binding, err := provider.Auth.CreateRoleBinding(ctx, auth.CreateRoleBindingInput{
		RoleName:    input.Body.Spec.RoleName,
		SubjectType: input.Body.Spec.SubjectType,
		SubjectID:   input.Body.Spec.SubjectID,
		Scope:       requestScope(input.Body.Spec),
	})
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.role_binding_created",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "role_binding",
		SubjectID:   binding.ID,
		Payload: map[string]any{
			"binding_id":    binding.ID,
			"role_id":       binding.RoleID,
			"role_name":     binding.RoleName,
			"subject_type":  binding.SubjectType,
			"subject_id":    binding.SubjectID,
			"resource_type": binding.ResourceType,
			"resource_id":   binding.ResourceID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateRoleBindingOutput{Body: roleBindingResource(binding)}, nil
}

func (provider Provider) createProjectRoleBinding(ctx context.Context, input *CreateProjectRoleBindingInput) (*CreateRoleBindingOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesBind, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	if !projectAssignableRole(input.Body.Spec.RoleName) {
		return nil, huma.Error400BadRequest("Role cannot be bound through the project role API")
	}
	binding, err := provider.Auth.CreateRoleBinding(ctx, auth.CreateRoleBindingInput{
		RoleName:    input.Body.Spec.RoleName,
		SubjectType: input.Body.Spec.SubjectType,
		SubjectID:   input.Body.Spec.SubjectID,
		Scope:       authz.ProjectScope(input.ProjectID),
	})
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.role_binding_created",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "role_binding",
		SubjectID:   binding.ID,
		Payload: map[string]any{
			"binding_id":    binding.ID,
			"role_id":       binding.RoleID,
			"role_name":     binding.RoleName,
			"subject_type":  binding.SubjectType,
			"subject_id":    binding.SubjectID,
			"resource_type": binding.ResourceType,
			"resource_id":   binding.ResourceID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateRoleBindingOutput{Body: roleBindingResource(binding)}, nil
}

func (provider Provider) deleteProjectRoleBinding(ctx context.Context, input *ProjectRoleBindingIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesBind, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	binding, err := provider.projectRoleBinding(ctx, input.ProjectID, input.BindingID)
	if err != nil {
		return nil, err
	}
	if err := provider.Auth.DeleteRoleBinding(ctx, input.BindingID); err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.role_binding_deleted",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "role_binding",
		SubjectID:   input.BindingID,
		Payload: map[string]any{
			"binding_id":    input.BindingID,
			"resource_type": binding.ResourceType,
			"resource_id":   binding.ResourceID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) deleteRoleBinding(ctx context.Context, input *RoleBindingIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionRolesBind, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if err := provider.Auth.DeleteRoleBinding(ctx, input.BindingID); err != nil {
		return nil, shared.AuthServiceError(err)
	}
	actorID, authKind := auditActor(principal)
	if err := provider.recordAudit(ctx, audit.RecordInput{
		EventType:   "rbac.role_binding_deleted",
		ActorID:     actorID,
		AuthKind:    authKind,
		SubjectType: "role_binding",
		SubjectID:   input.BindingID,
		Payload: map[string]any{
			"binding_id": input.BindingID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) roleByName(ctx context.Context, roleName authz.RoleName) (auth.Role, error) {
	roles, err := provider.Auth.ListRoles(ctx)
	if err != nil {
		return auth.Role{}, shared.AuthServiceError(err)
	}
	for _, role := range roles {
		if role.Name == roleName {
			return role, nil
		}
	}
	return auth.Role{}, huma.Error404NotFound("Role not found")
}

func (provider Provider) projectRoleBindings(ctx context.Context, projectID string) ([]auth.RoleBinding, error) {
	bindings, err := provider.Auth.ListRoleBindings(ctx)
	if err != nil {
		return nil, shared.AuthServiceError(err)
	}
	filtered := make([]auth.RoleBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.ResourceType == string(authz.ScopeKindProject) && binding.ResourceID == projectID {
			filtered = append(filtered, binding)
		}
	}
	return filtered, nil
}

func (provider Provider) projectRoleBinding(ctx context.Context, projectID string, bindingID string) (auth.RoleBinding, error) {
	bindings, err := provider.projectRoleBindings(ctx, projectID)
	if err != nil {
		return auth.RoleBinding{}, err
	}
	for _, binding := range bindings {
		if binding.ID == bindingID {
			return binding, nil
		}
	}
	return auth.RoleBinding{}, huma.Error404NotFound("Project role binding not found")
}

func projectAssignableRole(role authz.RoleName) bool {
	switch role {
	case authz.RoleProjectAdmin,
		authz.RoleProjectMember,
		authz.RoleProjectViewer,
		authz.RoleAutomationManager,
		authz.RoleNotificationManager:
		return true
	default:
		return false
	}
}

func requestScope(input RoleBindingSpec) authz.Scope {
	switch input.Scope {
	case "", string(authz.ScopeKindGlobal):
		return authz.GlobalScope()
	case string(authz.ScopeKindProject):
		return authz.ProjectScope(input.ProjectID)
	default:
		return authz.Scope{}
	}
}

func sessionCookie(session auth.Session) http.Cookie {
	return http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    session.Secret,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func csrfCookie(token string, expiresAt time.Time) http.Cookie {
	return http.Cookie{
		Name:     shared.CSRFCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
}

func randomURLToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func authFailureReason(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, auth.ErrDisabledUser):
		return "disabled_user"
	case errors.Is(err, auth.ErrInvalidCredentials):
		return "invalid_credentials"
	default:
		return "auth_error"
	}
}
