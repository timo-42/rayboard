package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/engines"
)

func TestEngineTestEndpoint(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	runStore := automation.NewRunStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithEngineService(engines.NewService(db.SQL, authorizer, runStore)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	sessionCookie := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrfCookie := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/engines/test", map[string]any{
		"spec": map[string]any{
			"engine": map[string]string{"type": "lua", "script": `return { ok = true }`},
		},
	}, []*http.Cookie{sessionCookie})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"surface": "ticket_hook_before",
			"context": map[string]any{"ticket_id": "ticket-1"},
			"input":   map[string]string{"title": "Preview"},
			"dry_run": false,
			"engine": map[string]string{
				"type":   "lua",
				"script": `rayboard.log("preview " .. input.title); return { ok = true, title = input.title, surface = context.surface, ticket_id = context.ticket_id, dry_run = context.dry_run }`,
			},
		},
	}))
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfCookie.Value)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected engine test status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Engine struct {
				Type   string `json:"type"`
				Script string `json:"script"`
			} `json:"engine"`
			Surface string         `json:"surface"`
			Context map[string]any `json:"context"`
			DryRun  bool           `json:"dry_run"`
		} `json:"spec"`
		Status struct {
			State  string         `json:"state"`
			Output map[string]any `json:"output"`
			Logs   []string       `json:"logs"`
			Engine map[string]any `json:"engine"`
		} `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode engine test response: %v", err)
	}
	if body.Metadata.ID == "" || body.Status.State != automation.StatusSucceeded {
		t.Fatalf("unexpected response metadata/status: %#v", body)
	}
	if body.Spec.Engine.Type != "lua" || body.Spec.Engine.Script != "" {
		t.Fatalf("expected engine source to be redacted, got %#v", body.Spec.Engine)
	}
	if body.Spec.Surface != "ticket_hook_before" || body.Spec.Context["ticket_id"] != "ticket-1" || !body.Spec.DryRun {
		t.Fatalf("expected normalized engine test spec, got %#v", body.Spec)
	}
	if body.Status.Output["ok"] != true || body.Status.Output["title"] != "Preview" || body.Status.Output["surface"] != "ticket_hook_before" || body.Status.Output["ticket_id"] != "ticket-1" || body.Status.Output["dry_run"] != true {
		t.Fatalf("unexpected engine output: %#v", body.Status.Output)
	}
	if len(body.Status.Logs) != 1 || body.Status.Logs[0] != "preview Preview" || body.Status.Engine["type"] != "lua" || body.Status.Engine["surface"] != "ticket_hook_before" || body.Status.Engine["dry_run"] != true {
		t.Fatalf("unexpected engine status metadata: %#v", body.Status)
	}
}
