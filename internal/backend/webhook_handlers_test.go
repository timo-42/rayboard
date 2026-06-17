package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
)

func TestWebhookEndpointsLifecycle(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	webhookService := webhooks.NewService(db.SQL, authorizer, webhooks.WithRunStore(automation.NewRunStore(db.SQL)))
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithWebhookService(webhookService),
	)
	seedWebhookHandlerProject(t, ctx, db, "project-1")
	actor, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "webhook-actor"})
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/projects/project-1/webhooks", map[string]any{
		"spec": map[string]any{
			"name":          "github",
			"direction":     webhooks.DirectionIncoming,
			"enabled":       true,
			"actor_user_id": actor.ID,
			"engine": map[string]any{
				"type":   webhooks.EngineTypeLua,
				"script": "return { ok = true }",
			},
		},
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/projects/project-1/webhooks", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":          "github",
			"direction":     webhooks.DirectionIncoming,
			"enabled":       true,
			"actor_user_id": actor.ID,
			"engine": map[string]any{
				"type":   webhooks.EngineTypeLua,
				"script": `rayboard.log("received " .. request.payload.kind); return { ok = true, kind = request.payload.kind }`,
			},
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create webhook status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeWebhookResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Metadata.ProjectID != "project-1" || created.Spec.Name != "github" || created.Spec.Direction != "incoming" || !created.Status.TokenSet || created.Status.Token == "" {
		t.Fatalf("unexpected created webhook: %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/webhooks", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list webhook status 200, got %d: %s", list.Code, list.Body.String())
	}
	if bytes.Contains(list.Body.Bytes(), []byte(created.Status.Token)) {
		t.Fatalf("webhook list leaked token: %s", list.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/webhook-definitions/"+created.Metadata.ID, nil)
	getReq.AddCookie(session)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getReq)
	if get.Code != http.StatusOK {
		t.Fatalf("expected get webhook status 200, got %d: %s", get.Code, get.Body.String())
	}
	if bytes.Contains(get.Body.Bytes(), []byte(created.Status.Token)) {
		t.Fatalf("webhook get leaked token: %s", get.Body.String())
	}

	enabled := false
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/webhook-definitions/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"enabled": enabled,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update webhook status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeWebhookResource(t, update.Body.Bytes())
	if updated.Spec.Enabled {
		t.Fatalf("expected disabled webhook, got %#v", updated)
	}

	enabled = true
	reenableReq := httptest.NewRequest(http.MethodPatch, "/api/webhook-definitions/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"enabled": enabled,
		},
	}))
	addSessionCSRF(reenableReq, session, csrf)
	reenable := httptest.NewRecorder()
	handler.ServeHTTP(reenable, reenableReq)
	if reenable.Code != http.StatusOK {
		t.Fatalf("expected reenable webhook status 200, got %d: %s", reenable.Code, reenable.Body.String())
	}

	rotateReq := httptest.NewRequest(http.MethodPost, "/api/webhook-definitions/"+created.Metadata.ID+"/rotate-token", nil)
	addSessionCSRF(rotateReq, session, csrf)
	rotate := httptest.NewRecorder()
	handler.ServeHTTP(rotate, rotateReq)
	if rotate.Code != http.StatusOK {
		t.Fatalf("expected rotate webhook token status 200, got %d: %s", rotate.Code, rotate.Body.String())
	}
	rotated := decodeWebhookResource(t, rotate.Body.Bytes())
	if rotated.Status.Token == "" || rotated.Status.Token == created.Status.Token {
		t.Fatalf("expected rotated token, got %#v", rotated)
	}

	oldIncomingReq := httptest.NewRequest(http.MethodPost, "/api/webhooks/incoming/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{"payload": map[string]any{"ok": true}},
	}))
	oldIncomingReq.Header.Set("Authorization", "Bearer "+created.Status.Token)
	oldIncoming := httptest.NewRecorder()
	handler.ServeHTTP(oldIncoming, oldIncomingReq)
	if oldIncoming.Code != http.StatusNotFound {
		t.Fatalf("expected old webhook token status 404, got %d: %s", oldIncoming.Code, oldIncoming.Body.String())
	}

	incomingReq := httptest.NewRequest(http.MethodPost, "/api/webhooks/incoming/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{"payload": map[string]any{"kind": "demo"}},
	}))
	incomingReq.Header.Set("Authorization", "Bearer "+rotated.Status.Token)
	incoming := httptest.NewRecorder()
	handler.ServeHTTP(incoming, incomingReq)
	if incoming.Code != http.StatusOK {
		t.Fatalf("expected incoming webhook status 200, got %d: %s", incoming.Code, incoming.Body.String())
	}
	received := decodeIncomingRunResource(t, incoming.Body.Bytes())
	if received.Metadata.ID == "" || received.Spec.Payload["kind"] != "demo" || received.Status.State != automation.StatusSucceeded {
		t.Fatalf("unexpected incoming webhook response: %#v", received)
	}
	output, ok := received.Status.Output["output"].(map[string]any)
	if !ok || output["ok"] != true || output["kind"] != "demo" {
		t.Fatalf("unexpected incoming webhook output: %#v", received.Status.Output)
	}
	if bytes.Contains(incoming.Body.Bytes(), []byte(rotated.Status.Token)) {
		t.Fatalf("incoming webhook response leaked token: %s", incoming.Body.String())
	}

	runsReq := httptest.NewRequest(http.MethodGet, "/api/webhook-definitions/"+created.Metadata.ID+"/runs", nil)
	runsReq.AddCookie(session)
	runs := httptest.NewRecorder()
	handler.ServeHTTP(runs, runsReq)
	if runs.Code != http.StatusOK {
		t.Fatalf("expected webhook runs status 200, got %d: %s", runs.Code, runs.Body.String())
	}
	if !bytes.Contains(runs.Body.Bytes(), []byte(received.Metadata.ID)) {
		t.Fatalf("expected run list to include incoming run: %s", runs.Body.String())
	}
	if bytes.Contains(runs.Body.Bytes(), []byte(rotated.Status.Token)) {
		t.Fatalf("webhook run list leaked token: %s", runs.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/webhook-definitions/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete webhook status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func TestWebhookEndpointsRequirePermission(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithWebhookService(webhooks.NewService(db.SQL, authorizer, webhooks.WithRunStore(automation.NewRunStore(db.SQL)))),
	)
	seedWebhookHandlerProject(t, ctx, db, "project-1")
	viewer, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "viewer"})
	if err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": viewer.Username,
		"password": viewer.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/webhooks", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden webhook list, got %d: %s", rec.Code, rec.Body.String())
	}
}

type webhookResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Name        string `json:"name"`
		Direction   string `json:"direction"`
		Enabled     bool   `json:"enabled"`
		ActorUserID string `json:"actor_user_id"`
		Engine      struct {
			Type   string `json:"type"`
			Script string `json:"script"`
		} `json:"engine"`
	} `json:"spec"`
	Status struct {
		TokenSet bool   `json:"token_set"`
		Token    string `json:"token"`
	} `json:"status"`
}

func decodeWebhookResource(t *testing.T, data []byte) webhookResourceBody {
	t.Helper()

	var body webhookResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode webhook resource: %v", err)
	}
	return body
}

type incomingRunResourceBody struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Payload map[string]any `json:"payload"`
	} `json:"spec"`
	Status struct {
		State  string         `json:"state"`
		Output map[string]any `json:"output"`
	} `json:"status"`
}

func decodeIncomingRunResource(t *testing.T, data []byte) incomingRunResourceBody {
	t.Helper()

	var body incomingRunResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode incoming run resource: %v", err)
	}
	return body
}

func seedWebhookHandlerProject(t *testing.T, ctx context.Context, db *store.DB, id string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES (?, ?, ?)
	`, id, "WEB", "Webhooks"); err != nil {
		t.Fatalf("seed webhook project: %v", err)
	}
}
