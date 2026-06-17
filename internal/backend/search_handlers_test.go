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
		"spec": map[string]any{
			"text": "login",
			"sort": []map[string]string{{"field": "key", "direction": "asc"}},
		},
	}))
	addSessionCSRF(searchReq, session, csrf)
	searchRec := httptest.NewRecorder()
	handler.ServeHTTP(searchRec, searchReq)
	if searchRec.Code != http.StatusOK {
		t.Fatalf("unexpected search response %d: %s", searchRec.Code, searchRec.Body.String())
	}
	var searchBody struct {
		Status struct {
			Items []ticketResourceBody `json:"items"`
		} `json:"status"`
	}
	if err := json.Unmarshal(searchRec.Body.Bytes(), &searchBody); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(searchBody.Status.Items) != 1 || searchBody.Status.Items[0].Status.Key != ticket.Key {
		t.Fatalf("unexpected search items: %#v", searchBody.Status.Items)
	}

	createViewReq := httptest.NewRequest(http.MethodPost, "/api/saved-views", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"scope_type":   search.SavedViewScopeProject,
			"project_id":   project.ID,
			"name":         "My Login Tickets",
			"display_mode": "board",
			"group_by":     "status",
			"pinned":       true,
			"query": map[string]string{
				"text":   "login",
				"filter": `status != "done"`,
			},
			"columns": []string{"key", "title", "status"},
		},
	}))
	addSessionCSRF(createViewReq, session, csrf)
	createView := httptest.NewRecorder()
	handler.ServeHTTP(createView, createViewReq)
	if createView.Code != http.StatusCreated {
		t.Fatalf("expected create saved view status 201, got %d: %s", createView.Code, createView.Body.String())
	}
	var view savedViewResourceBody
	if err := json.Unmarshal(createView.Body.Bytes(), &view); err != nil {
		t.Fatalf("decode saved view: %v", err)
	}
	if view.Metadata.ID == "" || view.Spec.Name != "My Login Tickets" {
		t.Fatalf("unexpected saved view: %#v", view)
	}
	if view.Spec.DisplayMode != search.SavedViewDisplayBoard || view.Spec.GroupBy != "status" || !view.Spec.Pinned {
		t.Fatalf("unexpected saved view metadata: %#v", view)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/saved-views?project_id="+project.ID+"&pinned=true", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), view.Metadata.ID) {
		t.Fatalf("unexpected saved view list response %d: %s", list.Code, list.Body.String())
	}

	newName := "Updated Login Tickets"
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/saved-views/"+view.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name": newName,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), newName) {
		t.Fatalf("unexpected saved view update response %d: %s", update.Code, update.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/saved-views/"+view.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete saved view status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func TestSearchRequiresCSRFForSession(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithSearchService(search.NewService(db.SQL, authorizer)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodPost, "/api/search", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"text": "login",
		},
	}))
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF search status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

type savedViewResourceBody struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Name        string `json:"name"`
		DisplayMode string `json:"display_mode"`
		GroupBy     string `json:"group_by"`
		Pinned      bool   `json:"pinned"`
	} `json:"spec"`
}

func createSearchTestProject(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie) tracker.Project {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"key":  "SEA",
			"name": "Search",
		},
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create project status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	return decodeProjectResourceAsTracker(t, rec.Body.Bytes())
}

func createSearchTestTicket(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie, projectID string) tracker.Ticket {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/tickets", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"title":       "Login search ticket",
			"description": "Find this login issue",
		},
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create ticket status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	return decodeTicketResourceAsTracker(t, rec.Body.Bytes())
}
