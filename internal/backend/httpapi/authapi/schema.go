package authapi

import (
	"net/http"

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
	Body CreateTokenInputBody
}

type CreateTokenInputBody struct {
	Name string `json:"name,omitempty"`
}

type RevokeTokenInput struct {
	shared.AuthInput
	TokenID string `path:"token_id"`
}

type CreateUserInput struct {
	shared.AuthInput
	Body CreateUserInputBody
}

type CreateUserInputBody struct {
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
	Body   UpdateUserInputBody
}

type UpdateUserInputBody struct {
	Disabled *bool `json:"disabled,omitempty"`
}

type CreateGroupInput struct {
	shared.AuthInput
	Body CreateGroupInputBody
}

type CreateGroupInputBody struct {
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
	Body CreateRoleBindingInputBody
}

type CreateRoleBindingInputBody struct {
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

type ListTokensOutput = shared.ListOutput[auth.APIToken]
type CreateTokenOutput = shared.CreatedOutput[auth.CreatedAPIToken]
type ListUsersOutput = shared.ListOutput[auth.User]
type CreateUserOutput = shared.CreatedOutput[auth.CreatedUser]
type UserOutput struct {
	Body auth.User
}
type ListGroupsOutput = shared.ListOutput[auth.Group]
type CreateGroupOutput = shared.CreatedOutput[auth.Group]
type ListGroupMembersOutput = shared.ListOutput[auth.User]
type ListRolesOutput = shared.ListOutput[auth.Role]
type ListRoleBindingsOutput = shared.ListOutput[auth.RoleBinding]
type CreateRoleBindingOutput = shared.CreatedOutput[auth.RoleBinding]
