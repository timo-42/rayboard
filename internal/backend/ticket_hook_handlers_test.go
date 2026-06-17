package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestTicketHookEndpointsLifecycle(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	hookService := tracker.NewHookService(db.SQL, authorizer)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithTicketHookService(hookService),
	)
	seedTicketHookHandlerProject(t, ctx, db, "project-1")

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/projects/project-1/ticket-hooks", map[string]any{
		"spec": ticketHookCreateBody("normalize-title", true, 10),
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/projects/project-1/ticket-hooks", mustJSON(t, map[string]any{
		"spec": ticketHookCreateBody("normalize-title", true, 10),
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create ticket hook status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeTicketHookResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Metadata.ProjectID != "project-1" || created.Spec.Name != "normalize-title" || created.Spec.Event != tracker.HookEventTicketCreate || created.Spec.Phase != tracker.HookPhaseBefore || created.Spec.Position != 10 || !created.Spec.Enabled {
		t.Fatalf("unexpected created ticket hook: %#v", created)
	}
	if created.Spec.Engine.Type != tracker.HookEngineLua || created.Spec.Engine.Script == "" {
		t.Fatalf("unexpected created ticket hook engine: %#v", created.Spec.Engine)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/ticket-hooks?event=ticket_create&phase=before", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list ticket hooks status 200, got %d: %s", list.Code, list.Body.String())
	}
	listBody := decodeTicketHookList(t, list.Body.Bytes())
	if listBody.Metadata.Count != 1 || len(listBody.Status.Items) != 1 || listBody.Status.Items[0].Metadata.ID != created.Metadata.ID {
		t.Fatalf("unexpected ticket hook list: %#v", listBody)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/ticket-hooks/"+created.Metadata.ID, nil)
	getReq.AddCookie(session)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getReq)
	if get.Code != http.StatusOK {
		t.Fatalf("expected get ticket hook status 200, got %d: %s", get.Code, get.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/ticket-hooks/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"enabled":  false,
			"position": 25,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update ticket hook status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeTicketHookResource(t, update.Body.Bytes())
	if updated.Spec.Enabled || updated.Spec.Position != 25 {
		t.Fatalf("unexpected updated ticket hook: %#v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/ticket-hooks/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete ticket hook status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/ticket-hooks/"+created.Metadata.ID, nil)
	notFoundReq.AddCookie(session)
	notFound := httptest.NewRecorder()
	handler.ServeHTTP(notFound, notFoundReq)
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("expected deleted ticket hook status 404, got %d: %s", notFound.Code, notFound.Body.String())
	}
}

func TestTicketHookEndpointsRequirePermission(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithTicketHookService(tracker.NewHookService(db.SQL, authorizer)),
	)
	seedTicketHookHandlerProject(t, ctx, db, "project-1")
	viewer, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "viewer"})
	if err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": viewer.Username,
		"password": viewer.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/ticket-hooks", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden ticket hook list, got %d: %s", rec.Code, rec.Body.String())
	}
}

func ticketHookCreateBody(name string, enabled bool, position int) map[string]any {
	return map[string]any{
		"name":     name,
		"event":    tracker.HookEventTicketCreate,
		"phase":    tracker.HookPhaseBefore,
		"enabled":  enabled,
		"position": position,
		"engine": map[string]any{
			"type":   tracker.HookEngineLua,
			"script": `return { ticket = ticket }`,
		},
	}
}

type ticketHookResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Name     string `json:"name"`
		Event    string `json:"event"`
		Phase    string `json:"phase"`
		Enabled  bool   `json:"enabled"`
		Position int    `json:"position"`
		Engine   struct {
			Type   string `json:"type"`
			Script string `json:"script"`
		} `json:"engine"`
	} `json:"spec"`
	Status struct {
		LastError string `json:"last_error"`
	} `json:"status"`
}

type ticketHookListBody struct {
	Metadata struct {
		Count int `json:"count"`
	} `json:"metadata"`
	Status struct {
		Items []ticketHookResourceBody `json:"items"`
	} `json:"status"`
}

func decodeTicketHookResource(t *testing.T, data []byte) ticketHookResourceBody {
	t.Helper()

	var body ticketHookResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode ticket hook resource: %v", err)
	}
	return body
}

func decodeTicketHookList(t *testing.T, data []byte) ticketHookListBody {
	t.Helper()

	var body ticketHookListBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode ticket hook list: %v", err)
	}
	return body
}

func seedTicketHookHandlerProject(t *testing.T, ctx context.Context, db *store.DB, id string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES (?, ?, ?)
	`, id, "HOOK", "Ticket Hooks"); err != nil {
		t.Fatalf("seed ticket hook project: %v", err)
	}
}
