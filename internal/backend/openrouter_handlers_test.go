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
	"github.com/timo-42/rayboard/internal/backend/openrouter"
)

func TestOpenRouterProviderEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	auditStore := audit.NewStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuditStore(auditStore),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithOpenRouterService(openrouter.NewService(db.SQL)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/openrouter-providers", map[string]any{
		"spec": map[string]any{
			"name":          "default",
			"default_model": "openai/gpt-4.1-mini",
			"api_key":       "sk-or-secret",
		},
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/openrouter-providers", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":                    "Default",
			"default_model":           "openai/gpt-4.1-mini",
			"api_key":                 "sk-or-secret",
			"allowed_models":          []string{"openai/gpt-4.1-mini", "anthropic/claude-sonnet-4"},
			"default_timeout_seconds": 45,
			"max_output_tokens":       4096,
			"enabled":                 true,
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create provider status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeOpenRouterProviderResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Spec.Name != "default" || created.Spec.DefaultModel != "openai/gpt-4.1-mini" || !created.Status.APIKeySet {
		t.Fatalf("unexpected provider resource: %#v", created)
	}
	if bytes.Contains(create.Body.Bytes(), []byte("sk-or-secret")) {
		t.Fatalf("provider response leaked API key: %s", create.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/openrouter-providers", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK || bytes.Contains(list.Body.Bytes(), []byte("sk-or-secret")) {
		t.Fatalf("unexpected provider list response %d: %s", list.Code, list.Body.String())
	}

	rotatedKey := "sk-or-rotated"
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/openrouter-providers/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"api_key": &rotatedKey,
			"enabled": false,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update provider status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeOpenRouterProviderResource(t, update.Body.Bytes())
	if updated.Spec.Enabled || !updated.Status.APIKeySet {
		t.Fatalf("unexpected updated provider: %#v", updated)
	}
	if bytes.Contains(update.Body.Bytes(), []byte(rotatedKey)) {
		t.Fatalf("provider update response leaked API key: %s", update.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/openrouter-providers/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete provider status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}

	entries, err := auditStore.List(ctx, 50)
	if err != nil {
		t.Fatalf("list audit entries: %v", err)
	}
	events := auditEvents(entries)
	for _, eventType := range []string{"openrouter.provider_created", "openrouter.provider_updated", "openrouter.provider_deleted"} {
		if events[eventType] == nil {
			t.Fatalf("expected audit event %s in %#v", eventType, entries)
		}
		if bytes.Contains(mustJSONBytes(t, events[eventType]), []byte(rotatedKey)) || bytes.Contains(mustJSONBytes(t, events[eventType]), []byte("sk-or-secret")) {
			t.Fatalf("audit event leaked API key: %#v", events[eventType])
		}
	}
	if events["openrouter.provider_updated"].Payload["api_key_rotated"] != true {
		t.Fatalf("expected key rotation audit payload, got %#v", events["openrouter.provider_updated"].Payload)
	}
}

func TestOpenRouterProviderEndpointsRequireAIManage(t *testing.T) {
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
		WithOpenRouterService(openrouter.NewService(db.SQL)),
	)
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	req := httptest.NewRequest(http.MethodPost, "/api/openrouter-providers", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":          "default",
			"default_model": "openai/gpt-4.1-mini",
			"api_key":       "sk-or-secret",
		},
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden provider create, got %d: %s", rec.Code, rec.Body.String())
	}
}

type openRouterProviderResourceBody struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Name                  string   `json:"name"`
		DefaultModel          string   `json:"default_model"`
		AllowedModels         []string `json:"allowed_models"`
		DefaultTimeoutSeconds int      `json:"default_timeout_seconds"`
		MaxOutputTokens       int      `json:"max_output_tokens"`
		Enabled               bool     `json:"enabled"`
		APIKey                string   `json:"api_key"`
	} `json:"spec"`
	Status struct {
		APIKeySet bool `json:"api_key_set"`
		Deleted   bool `json:"deleted"`
	} `json:"status"`
}

func decodeOpenRouterProviderResource(t *testing.T, data []byte) openRouterProviderResourceBody {
	t.Helper()

	var body openRouterProviderResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode OpenRouter provider resource: %v", err)
	}
	if body.Spec.APIKey != "" {
		t.Fatalf("provider resource leaked API key field: %#v", body)
	}
	return body
}
