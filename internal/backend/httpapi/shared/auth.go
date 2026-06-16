package shared

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
)

const CSRFCookieName = "rayboard_csrf"

type contextKey string

const (
	principalContextKey contextKey = "principal"
	userContextKey      contextKey = "user"
)

type AuthInput struct {
	Authorization string `header:"Authorization" doc:"Bearer API token."`
	CookieHeader  string `header:"Cookie" doc:"Raw Cookie header fallback."`
	SessionCookie string `cookie:"rayboard_session" doc:"Browser session cookie."`
	CSRFCookie    string `cookie:"rayboard_csrf" doc:"Browser CSRF cookie."`
	CSRFToken     string `header:"X-CSRF-Token" doc:"CSRF token for mutating cookie-authenticated requests."`
}

type Authenticator struct {
	Auth       *auth.Service
	Authorizer authz.Evaluator
}

func (a Authenticator) Authenticate(ctx context.Context, input AuthInput, mutating bool) (context.Context, authz.Principal, auth.User, error) {
	if a.Auth == nil {
		return ctx, authz.Principal{}, auth.User{}, huma.Error401Unauthorized("Authentication is required")
	}

	principal, user, err := a.authenticate(ctx, input)
	if err != nil {
		return ctx, authz.Principal{}, auth.User{}, AuthError(err)
	}
	if mutating && principal.AuthKind == authz.AuthKindSession && !validCSRF(input) {
		return ctx, authz.Principal{}, auth.User{}, huma.Error403Forbidden("CSRF token is required")
	}
	ctx = context.WithValue(ctx, principalContextKey, principal)
	ctx = context.WithValue(ctx, userContextKey, user)
	return ctx, principal, user, nil
}

func (a Authenticator) Require(principal authz.Principal, permission authz.Permission, scope authz.Scope) error {
	if a.Authorizer == nil {
		return huma.Error403Forbidden("Permission denied")
	}
	if err := a.Authorizer.Require(principal, permission, scope); err != nil {
		return huma.Error403Forbidden("Permission denied")
	}
	return nil
}

func PrincipalFromContext(ctx context.Context) (authz.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(authz.Principal)
	return principal, ok
}

func UserFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(userContextKey).(auth.User)
	return user, ok
}

func (a Authenticator) authenticate(ctx context.Context, input AuthInput) (authz.Principal, auth.User, error) {
	if token := bearerToken(input.Authorization); token != "" {
		return a.Auth.AuthenticateBearer(ctx, token)
	}
	sessionCookie := input.SessionCookieValue()
	if sessionCookie == "" {
		return authz.Principal{}, auth.User{}, auth.ErrUnauthenticated
	}
	return a.Auth.AuthenticateSession(ctx, sessionCookie)
}

func AuthError(err error) error {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrUnauthenticated):
		return huma.Error401Unauthorized("Authentication is required")
	case errors.Is(err, auth.ErrDisabledUser):
		return huma.Error403Forbidden("User is disabled")
	default:
		return huma.Error500InternalServerError("Authentication failed")
	}
}

func validCSRF(input AuthInput) bool {
	csrfCookie := input.CSRFCookieValue()
	return csrfCookie != "" && input.CSRFToken != "" && csrfCookie == input.CSRFToken
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	kind, value, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(kind, "Bearer") {
		return ""
	}
	return strings.TrimSpace(value)
}

func (input AuthInput) SessionCookieValue() string {
	if input.SessionCookie != "" {
		return input.SessionCookie
	}
	return cookieValue(input.CookieHeader, auth.SessionCookieName)
}

func (input AuthInput) CSRFCookieValue() string {
	if input.CSRFCookie != "" {
		return input.CSRFCookie
	}
	return cookieValue(input.CookieHeader, CSRFCookieName)
}

func cookieValue(header string, name string) string {
	if header == "" {
		return ""
	}
	request := http.Request{Header: http.Header{"Cookie": []string{header}}}
	cookie, err := request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
