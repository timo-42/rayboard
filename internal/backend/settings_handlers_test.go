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
	"github.com/timo-42/rayboard/internal/backend/settings"
)

func TestSettingsEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	auditStore := audit.NewStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuditStore(auditStore),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithSettingsService(settings.NewService(db.SQL)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	getReq := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getReq.AddCookie(session)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getReq)
	if get.Code != http.StatusOK {
		t.Fatalf("expected get settings status 200, got %d: %s", get.Code, get.Body.String())
	}
	got := decodeSettingsResource(t, get.Body.Bytes())
	if got.Metadata.ID != settings.GlobalSettingsKey || got.Spec.AttachmentMaxSizeBytes != 10<<20 {
		t.Fatalf("unexpected default settings: %#v", got)
	}

	missingCSRFReq := httptest.NewRequest(http.MethodPatch, "/api/settings", mustJSON(t, map[string]any{
		"spec": map[string]any{"attachment_max_size_bytes": 2048},
	}))
	missingCSRFReq.AddCookie(session)
	missingCSRFReq.Header.Set("Content-Type", "application/json")
	missingCSRF := httptest.NewRecorder()
	handler.ServeHTTP(missingCSRF, missingCSRFReq)
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/settings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"attachment_max_size_bytes":        2048,
			"attachment_allowed_content_types": []string{"text/plain"},
			"webhook_allowed_base_urls":        []string{"https://example.com/hooks/"},
			"demo_warning_enabled":             false,
			"system_health_note":               "green",
		},
	}))
	addSessionCSRF(patchReq, session, csrf)
	patch := httptest.NewRecorder()
	handler.ServeHTTP(patch, patchReq)
	if patch.Code != http.StatusOK {
		t.Fatalf("expected patch settings status 200, got %d: %s", patch.Code, patch.Body.String())
	}
	updated := decodeSettingsResource(t, patch.Body.Bytes())
	if updated.Spec.AttachmentMaxSizeBytes != 2048 || updated.Spec.DemoWarningEnabled || updated.Spec.SystemHealthNote != "green" {
		t.Fatalf("unexpected updated settings: %#v", updated)
	}
	if !updated.Status.AttachmentPolicyActive || !updated.Status.WebhookAllowlistActive {
		t.Fatalf("expected active settings status, got %#v", updated.Status)
	}

	entries, err := auditStore.List(ctx, 20)
	if err != nil {
		t.Fatalf("list audit entries: %v", err)
	}
	events := auditEvents(entries)
	if events["settings.updated"] == nil {
		t.Fatalf("expected settings.updated audit event in %#v", entries)
	}
	if bytes.Contains(mustJSONBytes(t, events["settings.updated"]), []byte("secret")) {
		t.Fatalf("settings audit leaked secret-like payload: %#v", events["settings.updated"])
	}
}

func TestSettingsEndpointsRequireGlobalSettingsManage(t *testing.T) {
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
		WithSettingsService(settings.NewService(db.SQL)),
	)
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden settings get, got %d: %s", rec.Code, rec.Body.String())
	}
}

type settingsResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		UpdatedBy string `json:"updated_by"`
	} `json:"metadata"`
	Spec struct {
		AttachmentMaxSizeBytes        int64    `json:"attachment_max_size_bytes"`
		AttachmentAllowedContentTypes []string `json:"attachment_allowed_content_types"`
		WebhookAllowedBaseURLs        []string `json:"webhook_allowed_base_urls"`
		DemoWarningEnabled            bool     `json:"demo_warning_enabled"`
		SystemHealthNote              string   `json:"system_health_note"`
	} `json:"spec"`
	Status struct {
		AttachmentPolicyActive bool `json:"attachment_policy_active"`
		WebhookAllowlistActive bool `json:"webhook_allowlist_active"`
	} `json:"status"`
}

func decodeSettingsResource(t *testing.T, data []byte) settingsResourceBody {
	t.Helper()
	var body settingsResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode settings resource: %v", err)
	}
	return body
}
