package authapi

import (
	"net/http"
	"time"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type LoginInput struct {
	Body LoginInputBody
}

type LoginInputBody struct {
	Username string `json:"username,omitempty" doc:"Username."`
	Password string `json:"password,omitempty" doc:"Password."`
}

type LoginOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      LoginOutputBody
}

type LoginOutputBody struct {
	User auth.User `json:"user"`
}

type LogoutInput struct {
	shared.AuthInput
}

type MeInput struct {
	shared.AuthInput
}

type MeOutput struct {
	Body MeOutputBody
}

type MeOutputBody struct {
	User      auth.User       `json:"user"`
	Principal authz.Principal `json:"principal"`
}

type CreateTokenInput struct {
	shared.AuthInput
	Body shared.ResourceInput[CreateTokenSpec]
}

type CreateTokenSpec struct {
	Name string `json:"name,omitempty"`
}

type RevokeTokenInput struct {
	shared.AuthInput
	TokenID string `path:"token_id"`
}

type CreateUserInput struct {
	shared.AuthInput
	Body shared.ResourceInput[CreateUserSpec]
}

type CreateUserSpec struct {
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Password    string `json:"password,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

type UserIDInput struct {
	shared.AuthInput
	UserID string `path:"user_id"`
}

type UpdateUserInput struct {
	shared.AuthInput
	UserID string `path:"user_id"`
	Body   shared.ResourceInput[UpdateUserSpec]
}

type UpdateUserSpec struct {
	Disabled *bool `json:"disabled,omitempty"`
}

type CreateGroupInput struct {
	shared.AuthInput
	Body shared.ResourceInput[GroupSpec]
}

type GroupSpec struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type GroupIDInput struct {
	shared.AuthInput
	GroupID string `path:"group_id"`
}

type GroupMemberInput struct {
	shared.AuthInput
	GroupID string `path:"group_id"`
	UserID  string `path:"user_id"`
}

type CreateRoleBindingInput struct {
	shared.AuthInput
	Body shared.ResourceInput[RoleBindingSpec]
}

type RoleBindingSpec struct {
	RoleName    authz.RoleName          `json:"role_name,omitempty"`
	SubjectType authz.BindingTargetKind `json:"subject_type,omitempty"`
	SubjectID   string                  `json:"subject_id,omitempty"`
	Scope       string                  `json:"scope,omitempty"`
	ProjectID   string                  `json:"project_id,omitempty"`
}

type RoleBindingIDInput struct {
	shared.AuthInput
	BindingID string `path:"binding_id"`
}

type ResourceMetadata struct {
	ID string `json:"id"`
}

type TokenSpec struct {
	Name string `json:"name"`
}

type TokenStatus struct {
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type CreatedTokenStatus struct {
	TokenStatus
	Token string `json:"token"`
}

type TokenResource = shared.Resource[ResourceMetadata, TokenSpec, TokenStatus]
type CreatedTokenResource = shared.Resource[ResourceMetadata, TokenSpec, CreatedTokenStatus]

type UserSpec struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Disabled    bool   `json:"disabled"`
}

type UserStatus struct {
	Disabled bool `json:"disabled"`
}

type CreatedUserStatus struct {
	UserStatus
	Password string `json:"password"`
}

type UserResource = shared.Resource[ResourceMetadata, UserSpec, UserStatus]
type CreatedUserResource = shared.Resource[ResourceMetadata, UserSpec, CreatedUserStatus]

type GroupStatus struct {
}

type GroupResource = shared.Resource[ResourceMetadata, GroupSpec, GroupStatus]

type RoleSpec struct {
	Name        authz.RoleName     `json:"name"`
	Description string             `json:"description"`
	Permissions []authz.Permission `json:"permissions"`
}

type RoleStatus struct {
}

type RoleResource = shared.Resource[ResourceMetadata, RoleSpec, RoleStatus]

type RoleBindingStatus struct {
	RoleID string `json:"role_id"`
}

type RoleBindingResource = shared.Resource[ResourceMetadata, RoleBindingSpec, RoleBindingStatus]

type ListTokensOutput = shared.ListOutput[TokenResource]
type CreateTokenOutput = shared.CreatedOutput[CreatedTokenResource]
type ListUsersOutput = shared.ListOutput[UserResource]
type CreateUserOutput = shared.CreatedOutput[CreatedUserResource]
type UserOutput struct {
	Body UserResource
}
type ListGroupsOutput = shared.ListOutput[GroupResource]
type CreateGroupOutput = shared.CreatedOutput[GroupResource]
type ListGroupMembersOutput = shared.ListOutput[UserResource]
type ListRolesOutput = shared.ListOutput[RoleResource]
type ListRoleBindingsOutput = shared.ListOutput[RoleBindingResource]
type CreateRoleBindingOutput = shared.CreatedOutput[RoleBindingResource]

func tokenResource(token auth.APIToken) TokenResource {
	return TokenResource{
		Metadata: ResourceMetadata{ID: token.ID},
		Spec:     TokenSpec{Name: token.Name},
		Status: TokenStatus{
			CreatedAt:  token.CreatedAt,
			LastUsedAt: token.LastUsedAt,
			ExpiresAt:  token.ExpiresAt,
			RevokedAt:  token.RevokedAt,
		},
	}
}

func createdTokenResource(token auth.CreatedAPIToken) CreatedTokenResource {
	resource := tokenResource(token.APIToken)
	return CreatedTokenResource{
		Metadata: resource.Metadata,
		Spec:     resource.Spec,
		Status: CreatedTokenStatus{
			TokenStatus: resource.Status,
			Token:       token.Token,
		},
	}
}

func tokenResources(tokens []auth.APIToken) []TokenResource {
	resources := make([]TokenResource, 0, len(tokens))
	for _, token := range tokens {
		resources = append(resources, tokenResource(token))
	}
	return resources
}

func userResource(user auth.User) UserResource {
	return UserResource{
		Metadata: ResourceMetadata{ID: user.ID},
		Spec: UserSpec{
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Disabled:    user.Disabled,
		},
		Status: UserStatus{Disabled: user.Disabled},
	}
}

func createdUserResource(user auth.CreatedUser) CreatedUserResource {
	resource := userResource(user.User)
	return CreatedUserResource{
		Metadata: resource.Metadata,
		Spec:     resource.Spec,
		Status: CreatedUserStatus{
			UserStatus: resource.Status,
			Password:   user.Password,
		},
	}
}

func userResources(users []auth.User) []UserResource {
	resources := make([]UserResource, 0, len(users))
	for _, user := range users {
		resources = append(resources, userResource(user))
	}
	return resources
}

func groupResource(group auth.Group) GroupResource {
	return GroupResource{
		Metadata: ResourceMetadata{ID: group.ID},
		Spec: GroupSpec{
			Name:        group.Name,
			DisplayName: group.DisplayName,
		},
		Status: GroupStatus{},
	}
}

func groupResources(groups []auth.Group) []GroupResource {
	resources := make([]GroupResource, 0, len(groups))
	for _, group := range groups {
		resources = append(resources, groupResource(group))
	}
	return resources
}

func roleResource(role auth.Role) RoleResource {
	return RoleResource{
		Metadata: ResourceMetadata{ID: role.ID},
		Spec: RoleSpec{
			Name:        role.Name,
			Description: role.Description,
			Permissions: role.Permissions,
		},
		Status: RoleStatus{},
	}
}

func roleResources(roles []auth.Role) []RoleResource {
	resources := make([]RoleResource, 0, len(roles))
	for _, role := range roles {
		resources = append(resources, roleResource(role))
	}
	return resources
}

func roleBindingResource(binding auth.RoleBinding) RoleBindingResource {
	spec := RoleBindingSpec{
		RoleName:    binding.RoleName,
		SubjectType: binding.SubjectType,
		SubjectID:   binding.SubjectID,
		Scope:       binding.ResourceType,
	}
	if binding.ResourceType == string(authz.ScopeKindProject) {
		spec.ProjectID = binding.ResourceID
	}
	return RoleBindingResource{
		Metadata: ResourceMetadata{ID: binding.ID},
		Spec:     spec,
		Status:   RoleBindingStatus{RoleID: binding.RoleID},
	}
}

func roleBindingResources(bindings []auth.RoleBinding) []RoleBindingResource {
	resources := make([]RoleBindingResource, 0, len(bindings))
	for _, binding := range bindings {
		resources = append(resources, roleBindingResource(binding))
	}
	return resources
}
