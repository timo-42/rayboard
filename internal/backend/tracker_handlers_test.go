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

	createSecondReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/tickets", mustJSON(t, map[string]any{
		"title": "Second API ticket",
	}))
	addSessionCSRF(createSecondReq, session, csrf)
	createSecond := httptest.NewRecorder()
	handler.ServeHTTP(createSecond, createSecondReq)
	if createSecond.Code != http.StatusCreated {
		t.Fatalf("expected create second ticket status 201, got %d: %s", createSecond.Code, createSecond.Body.String())
	}
	var second tracker.Ticket
	if err := json.Unmarshal(createSecond.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode second ticket: %v", err)
	}

	reorderBacklogReq := httptest.NewRequest(http.MethodPatch, "/api/projects/"+project.ID+"/backlog", mustJSON(t, map[string]any{
		"ticket_ids": []string{second.ID, ticket.ID},
	}))
	addSessionCSRF(reorderBacklogReq, session, csrf)
	reorderBacklog := httptest.NewRecorder()
	handler.ServeHTTP(reorderBacklog, reorderBacklogReq)
	if reorderBacklog.Code != http.StatusOK {
		t.Fatalf("expected reorder backlog status 200, got %d: %s", reorderBacklog.Code, reorderBacklog.Body.String())
	}

	listBacklogReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/backlog", nil)
	listBacklogReq.AddCookie(session)
	listBacklog := httptest.NewRecorder()
	handler.ServeHTTP(listBacklog, listBacklogReq)
	if listBacklog.Code != http.StatusOK {
		t.Fatalf("expected list backlog status 200, got %d: %s", listBacklog.Code, listBacklog.Body.String())
	}
	var backlog struct {
		Items []tracker.Ticket `json:"items"`
	}
	if err := json.Unmarshal(listBacklog.Body.Bytes(), &backlog); err != nil {
		t.Fatalf("decode backlog: %v", err)
	}
	if len(backlog.Items) != 2 || backlog.Items[0].ID != second.ID || backlog.Items[0].Rank != "000001" {
		t.Fatalf("unexpected backlog: %#v", backlog.Items)
	}

	createComponentReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/components", mustJSON(t, map[string]any{
		"name":        "API",
		"description": "Backend API",
	}))
	addSessionCSRF(createComponentReq, session, csrf)
	createComponent := httptest.NewRecorder()
	handler.ServeHTTP(createComponent, createComponentReq)
	if createComponent.Code != http.StatusCreated {
		t.Fatalf("expected create component status 201, got %d: %s", createComponent.Code, createComponent.Body.String())
	}
	var component tracker.Component
	if err := json.Unmarshal(createComponent.Body.Bytes(), &component); err != nil {
		t.Fatalf("decode component: %v", err)
	}
	if component.ID == "" || component.Name != "API" {
		t.Fatalf("unexpected component: %#v", component)
	}

	createVersionReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/versions", mustJSON(t, map[string]any{
		"name":        "1.0",
		"description": "First release",
		"target_date": "2026-07-01",
	}))
	addSessionCSRF(createVersionReq, session, csrf)
	createVersion := httptest.NewRecorder()
	handler.ServeHTTP(createVersion, createVersionReq)
	if createVersion.Code != http.StatusCreated {
		t.Fatalf("expected create version status 201, got %d: %s", createVersion.Code, createVersion.Body.String())
	}
	var version tracker.Version
	if err := json.Unmarshal(createVersion.Body.Bytes(), &version); err != nil {
		t.Fatalf("decode version: %v", err)
	}
	if version.ID == "" || version.Status != tracker.VersionStatusPlanned {
		t.Fatalf("unexpected version: %#v", version)
	}

	componentVersionUpdateReq := httptest.NewRequest(http.MethodPatch, "/api/tickets/"+second.ID, mustJSON(t, map[string]any{
		"component_id": component.ID,
		"version_id":   version.ID,
	}))
	addSessionCSRF(componentVersionUpdateReq, session, csrf)
	componentVersionUpdate := httptest.NewRecorder()
	handler.ServeHTTP(componentVersionUpdate, componentVersionUpdateReq)
	if componentVersionUpdate.Code != http.StatusOK {
		t.Fatalf("expected component/version ticket update status 200, got %d: %s", componentVersionUpdate.Code, componentVersionUpdate.Body.String())
	}
	var componentVersionTicket tracker.Ticket
	if err := json.Unmarshal(componentVersionUpdate.Body.Bytes(), &componentVersionTicket); err != nil {
		t.Fatalf("decode component/version ticket: %v", err)
	}
	if componentVersionTicket.ComponentID != component.ID || componentVersionTicket.VersionID != version.ID {
		t.Fatalf("unexpected component/version ticket: %#v", componentVersionTicket)
	}

	createSprintReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/sprints", mustJSON(t, map[string]any{
		"name":       "Sprint 1",
		"goal":       "Exercise sprint API",
		"start_date": "2026-06-16",
		"end_date":   "2026-06-30",
	}))
	addSessionCSRF(createSprintReq, session, csrf)
	createSprint := httptest.NewRecorder()
	handler.ServeHTTP(createSprint, createSprintReq)
	if createSprint.Code != http.StatusCreated {
		t.Fatalf("expected create sprint status 201, got %d: %s", createSprint.Code, createSprint.Body.String())
	}
	var sprint tracker.Sprint
	if err := json.Unmarshal(createSprint.Body.Bytes(), &sprint); err != nil {
		t.Fatalf("decode sprint: %v", err)
	}
	if sprint.ID == "" || sprint.State != tracker.SprintStatePlanned {
		t.Fatalf("unexpected sprint: %#v", sprint)
	}

	assignSprintReq := httptest.NewRequest(http.MethodPut, "/api/tickets/"+ticket.ID+"/sprint", mustJSON(t, map[string]any{
		"sprint_id": sprint.ID,
	}))
	addSessionCSRF(assignSprintReq, session, csrf)
	assignSprint := httptest.NewRecorder()
	handler.ServeHTTP(assignSprint, assignSprintReq)
	if assignSprint.Code != http.StatusOK {
		t.Fatalf("expected assign sprint status 200, got %d: %s", assignSprint.Code, assignSprint.Body.String())
	}
	var sprintTicket tracker.Ticket
	if err := json.Unmarshal(assignSprint.Body.Bytes(), &sprintTicket); err != nil {
		t.Fatalf("decode sprint ticket: %v", err)
	}
	if sprintTicket.SprintID != sprint.ID {
		t.Fatalf("expected ticket sprint %s, got %#v", sprint.ID, sprintTicket)
	}

	startSprintReq := httptest.NewRequest(http.MethodPost, "/api/sprints/"+sprint.ID+"/start", nil)
	addSessionCSRF(startSprintReq, session, csrf)
	startSprint := httptest.NewRecorder()
	handler.ServeHTTP(startSprint, startSprintReq)
	if startSprint.Code != http.StatusOK {
		t.Fatalf("expected start sprint status 200, got %d: %s", startSprint.Code, startSprint.Body.String())
	}

	completeSprintReq := httptest.NewRequest(http.MethodPost, "/api/sprints/"+sprint.ID+"/complete", nil)
	addSessionCSRF(completeSprintReq, session, csrf)
	completeSprint := httptest.NewRecorder()
	handler.ServeHTTP(completeSprint, completeSprintReq)
	if completeSprint.Code != http.StatusOK {
		t.Fatalf("expected complete sprint status 200, got %d: %s", completeSprint.Code, completeSprint.Body.String())
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
