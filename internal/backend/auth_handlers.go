package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
)

const csrfCookieName = "rayboard_csrf"

type contextKey string

const (
	principalContextKey contextKey = "principal"
	userContextKey      contextKey = "user"
)

type authRoute struct {
	auth       *auth.Service
	authorizer authz.Evaluator
}

func registerAuthRoutes(mux *http.ServeMux, authService *auth.Service, authorizer authz.Evaluator) {
	route := authRoute{auth: authService, authorizer: authorizer}
	mux.HandleFunc("POST /api/login", route.login)
	mux.HandleFunc("POST /api/logout", route.logout)
	mux.HandleFunc("GET /api/me", route.requireAuth(route.me))
	mux.HandleFunc("GET /api/tokens", route.requireAuth(route.listTokens))
	mux.HandleFunc("POST /api/tokens", route.requireAuth(route.createToken))
	mux.HandleFunc("DELETE /api/tokens/{token_id}", route.requireAuth(route.revokeToken))
	mux.HandleFunc("GET /api/users", route.requireAuth(route.listUsers))
	mux.HandleFunc("POST /api/users", route.requireAuth(route.createUser))
	mux.HandleFunc("GET /api/users/{user_id}", route.requireAuth(route.getUser))
	mux.HandleFunc("PATCH /api/users/{user_id}", route.requireAuth(route.updateUser))
	mux.HandleFunc("DELETE /api/users/{user_id}", route.requireAuth(route.deleteUser))
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	User auth.User `json:"user"`
}

type meResponse struct {
	User      auth.User       `json:"user"`
	Principal authz.Principal `json:"principal"`
}

type createTokenRequest struct {
	Name string `json:"name"`
}

type createUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	Disabled    bool   `json:"disabled"`
}

type updateUserRequest struct {
	Disabled *bool `json:"disabled"`
}

func (route authRoute) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Username == "" || request.Password == "" {
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Username and password are required", map[string]string{
			"username": "Required",
			"password": "Required",
		})
		return
	}

	session, err := route.auth.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	setSessionCookie(w, r, session)
	csrf, err := randomURLToken()
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Could not create CSRF token", nil)
		return
	}
	setCSRFCookie(w, r, csrf, session.ExpiresAt)

	httpjson.Write(w, http.StatusOK, loginResponse{User: session.User})
}

func (route authRoute) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err == nil && cookie.Value != "" {
		principal, _, authErr := route.auth.AuthenticateSession(r.Context(), cookie.Value)
		if authErr == nil && principal.AuthKind == authz.AuthKindSession && !route.validCSRF(r) {
			httpjson.Error(w, http.StatusForbidden, "forbidden", "CSRF token is required", nil)
			return
		}
		if err := route.auth.Logout(r.Context(), cookie.Value); err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Could not log out", nil)
			return
		}
	}

	clearCookie(w, r, auth.SessionCookieName, true)
	clearCookie(w, r, csrfCookieName, false)
	w.WriteHeader(http.StatusNoContent)
}

func (route authRoute) me(w http.ResponseWriter, r *http.Request, principal authz.Principal, user auth.User) {
	httpjson.Write(w, http.StatusOK, meResponse{
		User:      user,
		Principal: principal,
	})
}

func (route authRoute) listTokens(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	tokens, err := route.auth.ListAPITokens(r.Context(), principal.UserID)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Could not list API tokens", nil)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": tokens})
}

func (route authRoute) listUsers(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if !route.requirePermission(w, principal, authz.PermissionUsersRead, authz.GlobalScope()) {
		return
	}
	users, err := route.auth.ListUsers(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": users})
}

func (route authRoute) createUser(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if !route.requirePermission(w, principal, authz.PermissionUsersWrite, authz.GlobalScope()) {
		return
	}

	var request createUserRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	user, err := route.auth.CreateUser(r.Context(), auth.CreateUserInput{
		Username:    request.Username,
		DisplayName: request.DisplayName,
		Password:    request.Password,
		Disabled:    request.Disabled,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, user)
}

func (route authRoute) getUser(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if !route.requirePermission(w, principal, authz.PermissionUsersRead, authz.GlobalScope()) {
		return
	}
	user, err := route.auth.GetUser(r.Context(), r.PathValue("user_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, user)
}

func (route authRoute) updateUser(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if !route.requirePermission(w, principal, authz.PermissionUsersWrite, authz.GlobalScope()) {
		return
	}

	var request updateUserRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Disabled == nil {
		writeServiceError(w, &auth.ValidationError{
			Message: "Invalid user update",
			Fields:  map[string]string{"disabled": "Required"},
		})
		return
	}
	user, err := route.auth.SetUserDisabled(r.Context(), r.PathValue("user_id"), *request.Disabled)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, user)
}

func (route authRoute) deleteUser(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if !route.requirePermission(w, principal, authz.PermissionUsersWrite, authz.GlobalScope()) {
		return
	}
	if err := route.auth.DeleteUser(r.Context(), r.PathValue("user_id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route authRoute) createToken(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if principal.AuthKind == authz.AuthKindSession && !route.validCSRF(r) {
		httpjson.Error(w, http.StatusForbidden, "forbidden", "CSRF token is required", nil)
		return
	}

	var request createTokenRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	token, err := route.auth.CreateAPIToken(r.Context(), principal.UserID, request.Name)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Could not create API token", nil)
		return
	}
	httpjson.Write(w, http.StatusCreated, token)
}

func (route authRoute) revokeToken(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if principal.AuthKind == authz.AuthKindSession && !route.validCSRF(r) {
		httpjson.Error(w, http.StatusForbidden, "forbidden", "CSRF token is required", nil)
		return
	}

	tokenID := r.PathValue("token_id")
	if tokenID == "" {
		httpjson.Error(w, http.StatusNotFound, "not_found", "API token was not found", nil)
		return
	}
	if err := route.auth.RevokeAPIToken(r.Context(), principal.UserID, tokenID); err != nil {
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Could not revoke API token", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type authedHandler func(http.ResponseWriter, *http.Request, authz.Principal, auth.User)

func (route authRoute) requireAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, user, err := route.authenticate(r)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		if principal.AuthKind == authz.AuthKindSession && mutatesState(r.Method) && !route.validCSRF(r) {
			httpjson.Error(w, http.StatusForbidden, "forbidden", "CSRF token is required", nil)
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey, principal)
		ctx = context.WithValue(ctx, userContextKey, user)
		next(w, r.WithContext(ctx), principal, user)
	}
}

func (route authRoute) authenticate(r *http.Request) (authz.Principal, auth.User, error) {
	if token := bearerToken(r.Header.Get("Authorization")); token != "" {
		return route.auth.AuthenticateBearer(r.Context(), token)
	}
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil || cookie.Value == "" {
		return authz.Principal{}, auth.User{}, auth.ErrUnauthenticated
	}
	return route.auth.AuthenticateSession(r.Context(), cookie.Value)
}

func (route authRoute) validCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	header := r.Header.Get("X-CSRF-Token")
	return header != "" && header == cookie.Value
}

func (route authRoute) requirePermission(w http.ResponseWriter, principal authz.Principal, permission authz.Permission, scope authz.Scope) bool {
	if route.authorizer == nil {
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
		return false
	}
	if err := route.authorizer.Require(principal, permission, scope); err != nil {
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
		return false
	}
	return true
}

func PrincipalFromContext(ctx context.Context) (authz.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(authz.Principal)
	return principal, ok
}

func UserFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(userContextKey).(auth.User)
	return user, ok
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Request body must be valid JSON", nil)
		return false
	}
	return true
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrUnauthenticated):
		httpjson.Error(w, http.StatusUnauthorized, "unauthenticated", "Authentication is required", nil)
	case errors.Is(err, auth.ErrDisabledUser):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "User is disabled", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Authentication failed", nil)
	}
}

func writeServiceError(w http.ResponseWriter, err error) {
	var validation *auth.ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, auth.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, auth.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, auth.ErrConflict):
		httpjson.Error(w, http.StatusConflict, "conflict", "Resource already exists", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
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

func mutatesState(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, session auth.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    session.Secret,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}

func setCSRFCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}

func clearCookie(w http.ResponseWriter, r *http.Request, name string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}

func randomURLToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
