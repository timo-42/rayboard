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

func TestNotificationPolicyEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	notificationService := notifications.NewService(db.SQL)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notificationService),
	)

	destination, err := notificationService.CreateDestination(ctx, notifications.CreateDestinationInput{
		Name:        "ops",
		ScopeType:   notifications.DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/notification-policies", map[string]any{
		"spec": map[string]any{
			"name":            "ops",
			"event_types":     []string{"ticket_assigned"},
			"destination_ids": []string{destination.ID},
			"enabled":         true,
		},
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/notification-policies", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":            "Ops",
			"event_types":     []string{"ticket_assigned", "ticket_status_changed"},
			"destination_ids": []string{destination.ID},
			"enabled":         true,
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create policy status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeNotificationPolicyResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Metadata.ScopeType != "global" || created.Spec.Name != "ops" || len(created.Spec.EventTypes) != 2 || len(created.Spec.DestinationIDs) != 1 || !created.Spec.Enabled {
		t.Fatalf("unexpected created policy: %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/notification-policies", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list policy status 200, got %d: %s", list.Code, list.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/notification-policies/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"enabled": false,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update policy status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeNotificationPolicyResource(t, update.Body.Bytes())
	if updated.Spec.Enabled {
		t.Fatalf("expected disabled policy, got %#v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/notification-policies/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete policy status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func TestNotificationPolicyEndpointsRequireManagePermission(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/api/notification-policies", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden policy list, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationHookEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	actor, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "hook-actor"})
	if err != nil {
		t.Fatalf("create hook actor: %v", err)
	}
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

	createReq := httptest.NewRequest(http.MethodPost, "/api/notification-hooks", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":          "Suppress noisy comments",
			"actor_user_id": actor.ID,
			"event_types":   []string{"comment_added"},
			"enabled":       true,
			"engine": map[string]any{
				"type":   "lua",
				"script": `return { suppress = true }`,
			},
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create hook status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeNotificationHookResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Metadata.ScopeType != notifications.PolicyScopeGlobal || created.Spec.Name != "suppress noisy comments" || created.Spec.ActorUserID != actor.ID || created.Spec.Engine.Type != "lua" {
		t.Fatalf("unexpected created hook: %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/notification-hooks", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list hook status 200, got %d: %s", list.Code, list.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/notification-hooks/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{"enabled": false},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update hook status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeNotificationHookResource(t, update.Body.Bytes())
	if updated.Spec.Enabled {
		t.Fatalf("expected disabled hook, got %#v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/notification-hooks/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete hook status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

type notificationPolicyResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ScopeType string `json:"scope_type"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Name           string   `json:"name"`
		EventTypes     []string `json:"event_types"`
		DestinationIDs []string `json:"destination_ids"`
		Enabled        bool     `json:"enabled"`
	} `json:"spec"`
	Status struct {
		Deleted bool `json:"deleted"`
	} `json:"status"`
}

func decodeNotificationPolicyResource(t *testing.T, data []byte) notificationPolicyResourceBody {
	t.Helper()

	var body notificationPolicyResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification policy resource: %v", err)
	}
	return body
}

type notificationHookResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ScopeType string `json:"scope_type"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Name        string   `json:"name"`
		ActorUserID string   `json:"actor_user_id"`
		EventTypes  []string `json:"event_types"`
		Enabled     bool     `json:"enabled"`
		Engine      struct {
			Type   string `json:"type"`
			Script string `json:"script"`
		} `json:"engine"`
	} `json:"spec"`
	Status struct {
		LastError string `json:"last_error"`
	} `json:"status"`
}

func decodeNotificationHookResource(t *testing.T, data []byte) notificationHookResourceBody {
	t.Helper()

	var body notificationHookResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification hook resource: %v", err)
	}
	return body
}
