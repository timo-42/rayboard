package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

func TestNotificationPreferenceEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notifications.NewService(db.SQL)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	getReq := httptest.NewRequest(http.MethodGet, "/api/me/notification-preferences", nil)
	getReq.AddCookie(session)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getReq)
	if get.Code != http.StatusOK {
		t.Fatalf("expected preferences status 200, got %d: %s", get.Code, get.Body.String())
	}
	defaults := decodeNotificationPreferencesResource(t, get.Body.Bytes())
	if defaults.Metadata.ScopeType != "user" || defaults.Metadata.UserID != bootstrap.UserID || defaults.Status.Customized || !defaults.Spec.InAppEnabled || !defaults.Spec.ExternalEnabled {
		t.Fatalf("unexpected default preferences: %#v", defaults)
	}

	missingCSRFReq := httptest.NewRequest(http.MethodPatch, "/api/me/notification-preferences", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"external_enabled": false,
		},
	}))
	missingCSRFReq.Header.Set("Content-Type", "application/json")
	missingCSRFReq.AddCookie(session)
	missingCSRF := httptest.NewRecorder()
	handler.ServeHTTP(missingCSRF, missingCSRFReq)
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/me/notification-preferences", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"external_enabled":      false,
			"status_change_enabled": false,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update preferences status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeNotificationPreferencesResource(t, update.Body.Bytes())
	if updated.Metadata.ID == "" || !updated.Status.Customized || updated.Spec.ExternalEnabled || updated.Spec.StatusChangeEnabled || !updated.Spec.InAppEnabled {
		t.Fatalf("unexpected updated preferences: %#v", updated)
	}
}

func TestProjectNotificationPreferenceEndpointsRequireManagePermission(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES ('project-1', 'CORE', 'Core')
	`); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	authService := auth.NewService(db.SQL)
	user, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "viewer"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notifications.NewService(db.SQL)),
	)

	viewerLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	viewerSession := responseCookie(t, viewerLogin.Result(), auth.SessionCookieName)
	deniedReq := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/notification-preferences", nil)
	deniedReq.AddCookie(viewerSession)
	denied := httptest.NewRecorder()
	handler.ServeHTTP(denied, deniedReq)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("expected project preferences forbidden, got %d: %s", denied.Code, denied.Body.String())
	}

	adminLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	adminSession := responseCookie(t, adminLogin.Result(), auth.SessionCookieName)
	adminCSRF := responseCookie(t, adminLogin.Result(), csrfCookieName)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/projects/project-1/notification-preferences", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"comment_enabled": false,
		},
	}))
	addSessionCSRF(updateReq, adminSession, adminCSRF)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected project preference update status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeNotificationPreferencesResource(t, update.Body.Bytes())
	if updated.Metadata.ScopeType != "project" || updated.Metadata.ProjectID != "project-1" || updated.Spec.CommentEnabled || !updated.Status.Customized {
		t.Fatalf("unexpected project preferences: %#v", updated)
	}
}

type notificationPreferencesResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ScopeType string `json:"scope_type"`
		UserID    string `json:"user_id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		InAppEnabled        bool `json:"in_app_enabled"`
		ExternalEnabled     bool `json:"external_enabled"`
		StatusChangeEnabled bool `json:"status_change_enabled"`
		CommentEnabled      bool `json:"comment_enabled"`
	} `json:"spec"`
	Status struct {
		Customized bool `json:"customized"`
	} `json:"status"`
}

func decodeNotificationPreferencesResource(t *testing.T, data []byte) notificationPreferencesResourceBody {
	t.Helper()

	var body notificationPreferencesResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification preferences resource: %v", err)
	}
	return body
}
