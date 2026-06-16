package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestSearchEndpointsAndSavedViews(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithTrackerService(tracker.NewService(db.SQL, authorizer)),
		WithSearchService(search.NewService(db.SQL, authorizer)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	project := createSearchTestProject(t, handler, session, csrf)
	ticket := createSearchTestTicket(t, handler, session, csrf, project.ID)

	searchReq := httptest.NewRequest(http.MethodPost, "/api/search", mustJSON(t, map[string]any{
		"text": "login",
		"sort": []map[string]string{{"field": "key", "direction": "asc"}},
	}))
	addSessionCSRF(searchReq, session, csrf)
	searchRec := httptest.NewRecorder()
	handler.ServeHTTP(searchRec, searchReq)
	if searchRec.Code != http.StatusOK || !strings.Contains(searchRec.Body.String(), ticket.Key) {
		t.Fatalf("unexpected search response %d: %s", searchRec.Code, searchRec.Body.String())
	}

	createViewReq := httptest.NewRequest(http.MethodPost, "/api/saved-views", mustJSON(t, map[string]any{
		"scope_type": search.SavedViewScopeUser,
		"project_id": project.ID,
		"name":       "My Login Tickets",
		"query": map[string]string{
			"text":   "login",
			"filter": `status != "done"`,
		},
		"columns": []string{"key", "title", "status"},
	}))
	addSessionCSRF(createViewReq, session, csrf)
	createView := httptest.NewRecorder()
	handler.ServeHTTP(createView, createViewReq)
	if createView.Code != http.StatusCreated {
		t.Fatalf("expected create saved view status 201, got %d: %s", createView.Code, createView.Body.String())
	}
	var view search.SavedView
	if err := json.Unmarshal(createView.Body.Bytes(), &view); err != nil {
		t.Fatalf("decode saved view: %v", err)
	}
	if view.ID == "" || view.Name != "My Login Tickets" {
		t.Fatalf("unexpected saved view: %#v", view)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/saved-views?project_id="+project.ID, nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), view.ID) {
		t.Fatalf("unexpected saved view list response %d: %s", list.Code, list.Body.String())
	}

	newName := "Updated Login Tickets"
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/saved-views/"+view.ID, mustJSON(t, map[string]any{
		"name": newName,
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), newName) {
		t.Fatalf("unexpected saved view update response %d: %s", update.Code, update.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/saved-views/"+view.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete saved view status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func createSearchTestProject(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie) tracker.Project {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects", mustJSON(t, map[string]any{
		"key":  "SEA",
		"name": "Search",
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create project status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var project tracker.Project
	if err := json.Unmarshal(rec.Body.Bytes(), &project); err != nil {
		t.Fatalf("decode project: %v", err)
	}
	return project
}

func createSearchTestTicket(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie, projectID string) tracker.Ticket {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/tickets", mustJSON(t, map[string]any{
		"title":       "Login search ticket",
		"description": "Find this login issue",
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create ticket status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var ticket tracker.Ticket
	if err := json.Unmarshal(rec.Body.Bytes(), &ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}
	return ticket
}
