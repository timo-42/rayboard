package authapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.PublicOperation(http.MethodPost, "/api/login", "Auth", "Log in with username and password"), provider.login)
	registerLogout(api, provider)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/me", "Auth", "Get current authenticated user"), provider.me)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tokens", "Auth", "List API tokens"), provider.listTokens)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/tokens", "Auth", "Create API token", http.StatusCreated), provider.createToken)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/tokens/{token_id}", "Auth", "Revoke API token", http.StatusNoContent), provider.revokeToken)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/users", "Users", "List users"), provider.listUsers)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/users", "Users", "Create user", http.StatusCreated), provider.createUser)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/users/{user_id}", "Users", "Get user"), provider.getUser)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/users/{user_id}", "Users", "Update user"), provider.updateUser)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/users/{user_id}", "Users", "Delete user", http.StatusNoContent), provider.deleteUser)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/groups", "RBAC", "List groups"), provider.listGroups)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/groups", "RBAC", "Create group", http.StatusCreated), provider.createGroup)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/groups/{group_id}/members", "RBAC", "List group members"), provider.listGroupMembers)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/groups/{group_id}/members/{user_id}", "RBAC", "Add group member"), provider.addGroupMember)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/groups/{group_id}/members/{user_id}", "RBAC", "Remove group member", http.StatusNoContent), provider.removeGroupMember)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/roles", "RBAC", "List roles"), provider.listRoles)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/role-bindings", "RBAC", "List role bindings"), provider.listRoleBindings)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/role-bindings", "RBAC", "Create role binding", http.StatusCreated), provider.createRoleBinding)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/role-bindings/{binding_id}", "RBAC", "Delete role binding", http.StatusNoContent), provider.deleteRoleBinding)
}

func (provider Provider) login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	session, err := provider.Auth.Login(ctx, input.Body.Username, input.Body.Password)
	if err != nil {
		return nil, shared.AuthError(err)
	}
	csrf, err := randomURLToken()
	if err != nil {
		return nil, huma.Error500InternalServerError("Could not create CSRF token")
	}
	return &LoginOutput{
		SetCookie: []http.Cookie{
			sessionCookie(session),
			csrfCookie(csrf, session.ExpiresAt),
		},
		Body: LoginOutputBody{User: session.User},
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
	}
	ctx.SetStatus(http.StatusNoContent)
}

func (provider Provider) me(ctx context.Context, input *MeInput) (*MeOutput, error) {
	ctx, principal, user, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	_ = ctx
	return &MeOutput{Body: MeOutputBody{User: user, Principal: principal}}, nil
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
	return &ListTokensOutput{Body: shared.ItemList[TokenResource]{Items: tokenResources(tokens)}}, nil
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
	return &ListUsersOutput{Body: shared.ItemList[UserResource]{Items: userResources(users)}}, nil
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
	return &ListGroupsOutput{Body: shared.ItemList[GroupResource]{Items: groupResources(groups)}}, nil
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
	return &ListGroupMembersOutput{Body: shared.ItemList[UserResource]{Items: userResources(users)}}, nil
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
	return &ListRolesOutput{Body: shared.ItemList[RoleResource]{Items: roleResources(roles)}}, nil
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
	return &ListRoleBindingsOutput{Body: shared.ItemList[RoleBindingResource]{Items: roleBindingResources(bindings)}}, nil
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
	return &CreateRoleBindingOutput{Body: roleBindingResource(binding)}, nil
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
	return &shared.EmptyOutput{}, nil
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
