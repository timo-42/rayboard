package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestTrackerEndpointsProjectAndTicketFlow(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	handler := newTrackerTestHandler(db.SQL)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	createProjectReq := httptest.NewRequest(http.MethodPost, "/api/projects", mustJSON(t, map[string]any{
		"key":         "CORE",
		"name":        "Core Tracking",
		"description": "Project and ticket API",
	}))
	addSessionCSRF(createProjectReq, session, csrf)
	createProject := httptest.NewRecorder()
	handler.ServeHTTP(createProject, createProjectReq)
	if createProject.Code != http.StatusCreated {
		t.Fatalf("expected create project status 201, got %d: %s", createProject.Code, createProject.Body.String())
	}
	var project tracker.Project
	if err := json.Unmarshal(createProject.Body.Bytes(), &project); err != nil {
		t.Fatalf("decode project: %v", err)
	}
	if project.ID == "" || project.Key != "CORE" {
		t.Fatalf("unexpected project: %#v", project)
	}

	listProjectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	listProjectsReq.AddCookie(session)
	listProjects := httptest.NewRecorder()
	handler.ServeHTTP(listProjects, listProjectsReq)
	if listProjects.Code != http.StatusOK {
		t.Fatalf("expected list projects status 200, got %d: %s", listProjects.Code, listProjects.Body.String())
	}

	createTicketReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/tickets", mustJSON(t, map[string]any{
		"title":       "First API ticket",
		"description": "Created through HTTP",
		"priority":    "High",
		"type":        "Bug",
	}))
	addSessionCSRF(createTicketReq, session, csrf)
	createTicket := httptest.NewRecorder()
	handler.ServeHTTP(createTicket, createTicketReq)
	if createTicket.Code != http.StatusCreated {
		t.Fatalf("expected create ticket status 201, got %d: %s", createTicket.Code, createTicket.Body.String())
	}
	var ticket tracker.Ticket
	if err := json.Unmarshal(createTicket.Body.Bytes(), &ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}
	if ticket.ID == "" || ticket.Key != "CORE-1" || ticket.Status != "todo" {
		t.Fatalf("unexpected ticket: %#v", ticket)
	}

	status := "In_Progress"
	updateTicketReq := httptest.NewRequest(http.MethodPatch, "/api/tickets/"+ticket.ID, mustJSON(t, map[string]any{
		"status": status,
	}))
	addSessionCSRF(updateTicketReq, session, csrf)
	updateTicket := httptest.NewRecorder()
	handler.ServeHTTP(updateTicket, updateTicketReq)
	if updateTicket.Code != http.StatusOK {
		t.Fatalf("expected update ticket status 200, got %d: %s", updateTicket.Code, updateTicket.Body.String())
	}

	activityReq := httptest.NewRequest(http.MethodGet, "/api/tickets/"+ticket.ID+"/activity", nil)
	activityReq.AddCookie(session)
	activity := httptest.NewRecorder()
	handler.ServeHTTP(activity, activityReq)
	if activity.Code != http.StatusOK {
		t.Fatalf("expected activity status 200, got %d: %s", activity.Code, activity.Body.String())
	}
}

func TestTrackerEndpointsDenyUnprivilegedMutations(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	user, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "viewer"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	handler := newTrackerTestHandler(db.SQL)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	createProjectReq := httptest.NewRequest(http.MethodPost, "/api/projects", mustJSON(t, map[string]any{
		"key":  "DENY",
		"name": "Denied",
	}))
	addSessionCSRF(createProjectReq, session, csrf)
	createProject := httptest.NewRecorder()
	handler.ServeHTTP(createProject, createProjectReq)
	if createProject.Code != http.StatusForbidden {
		t.Fatalf("expected create project status 403, got %d: %s", createProject.Code, createProject.Body.String())
	}
}

func newTrackerTestHandler(db *sql.DB) http.Handler {
	authorizer := authz.NewSQLEvaluator(db)
	authService := auth.NewService(db)
	trackerService := tracker.NewService(db, authorizer)
	return NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithTrackerService(trackerService),
	)
}
