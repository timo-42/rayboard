package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

func TestNotificationDestinationEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	auditStore := audit.NewStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuditStore(auditStore),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notifications.NewService(db.SQL)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/notification-destinations", map[string]any{
		"spec": map[string]any{
			"name":         "ops",
			"shoutrrr_url": "logger://",
			"enabled":      true,
		},
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/notification-destinations", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":         "Ops",
			"shoutrrr_url": "logger://",
			"enabled":      true,
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create destination status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeNotificationDestinationResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Metadata.ScopeType != "global" || created.Spec.Name != "ops" || created.Spec.Type != "logger" || !created.Spec.Enabled || !created.Status.URLSet {
		t.Fatalf("unexpected destination resource: %#v", created)
	}
	if bytes.Contains(create.Body.Bytes(), []byte("logger://")) {
		t.Fatalf("destination response leaked Shoutrrr URL: %s", create.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/notification-destinations", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK || bytes.Contains(list.Body.Bytes(), []byte("logger://")) {
		t.Fatalf("unexpected destination list response %d: %s", list.Code, list.Body.String())
	}

	newURL := "logger://"
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/notification-destinations/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"shoutrrr_url": &newURL,
			"enabled":      false,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update destination status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeNotificationDestinationResource(t, update.Body.Bytes())
	if updated.Spec.Enabled || !updated.Status.URLSet {
		t.Fatalf("unexpected updated destination: %#v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/notification-destinations/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete destination status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}

	entries, err := auditStore.List(ctx, 50)
	if err != nil {
		t.Fatalf("list audit entries: %v", err)
	}
	events := auditEvents(entries)
	for _, eventType := range []string{"notification.destination_created", "notification.destination_updated", "notification.destination_deleted"} {
		if events[eventType] == nil {
			t.Fatalf("expected audit event %s in %#v", eventType, entries)
		}
		if bytes.Contains(mustJSONBytes(t, events[eventType]), []byte("logger://")) {
			t.Fatalf("audit event leaked Shoutrrr URL: %#v", events[eventType])
		}
	}
	if events["notification.destination_updated"].Payload["url_rotated"] != true {
		t.Fatalf("expected URL rotation audit payload, got %#v", events["notification.destination_updated"].Payload)
	}
}

func TestNotificationDestinationEndpointsRequireManagePermission(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
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
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	req := httptest.NewRequest(http.MethodPost, "/api/notification-destinations", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":         "ops",
			"shoutrrr_url": "logger://",
		},
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden destination create, got %d: %s", rec.Code, rec.Body.String())
	}
}

type notificationDestinationResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ScopeType string `json:"scope_type"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Enabled     bool   `json:"enabled"`
		ShoutrrrURL string `json:"shoutrrr_url"`
	} `json:"spec"`
	Status struct {
		URLSet bool `json:"url_set"`
	} `json:"status"`
}

func decodeNotificationDestinationResource(t *testing.T, data []byte) notificationDestinationResourceBody {
	t.Helper()

	var body notificationDestinationResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification destination resource: %v", err)
	}
	if body.Spec.ShoutrrrURL != "" {
		t.Fatalf("destination resource leaked Shoutrrr URL field: %#v", body)
	}
	return body
}
