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
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestCommentEndpointsLifecycle(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithTrackerService(tracker.NewService(db.SQL, authorizer)),
		WithCommentService(comments.NewService(db.SQL, authorizer)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	project := createCommentTestProject(t, handler, session, csrf)
	ticket := createCommentTestTicket(t, handler, session, csrf, project.ID)

	createReq := httptest.NewRequest(http.MethodPost, "/api/tickets/"+ticket.ID+"/comments", mustJSON(t, map[string]any{
		"body": "Looks ready",
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create comment status 201, got %d: %s", create.Code, create.Body.String())
	}
	var comment comments.Comment
	if err := json.Unmarshal(create.Body.Bytes(), &comment); err != nil {
		t.Fatalf("decode comment: %v", err)
	}
	if comment.ID == "" || comment.Body != "Looks ready" {
		t.Fatalf("unexpected comment: %#v", comment)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/tickets/"+ticket.ID+"/comments", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "Looks ready") {
		t.Fatalf("unexpected list response %d: %s", list.Code, list.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/comments/"+comment.ID, nil)
	deleteReq.AddCookie(session)
	deleteReq.AddCookie(csrf)
	deleteReq.Header.Set("X-CSRF-Token", csrf.Value)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func createCommentTestProject(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie) tracker.Project {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects", mustJSON(t, map[string]any{
		"key":  "COM",
		"name": "Comments",
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

func createCommentTestTicket(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie, projectID string) tracker.Ticket {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/tickets", mustJSON(t, map[string]any{
		"title": "Comment target",
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
