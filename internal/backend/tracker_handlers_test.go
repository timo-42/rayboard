package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
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
		"spec": map[string]any{
			"key":         "CORE",
			"name":        "Core Tracking",
			"description": "Project and ticket API",
		},
	}))
	addSessionCSRF(createProjectReq, session, csrf)
	createProject := httptest.NewRecorder()
	handler.ServeHTTP(createProject, createProjectReq)
	if createProject.Code != http.StatusCreated {
		t.Fatalf("expected create project status 201, got %d: %s", createProject.Code, createProject.Body.String())
	}
	project := decodeProjectResourceAsTracker(t, createProject.Body.Bytes())
	if project.ID == "" || project.Key != "CORE" {
		t.Fatalf("unexpected project: %#v", project)
	}

	listStatusesReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/statuses", nil)
	listStatusesReq.AddCookie(session)
	listStatuses := httptest.NewRecorder()
	handler.ServeHTTP(listStatuses, listStatusesReq)
	if listStatuses.Code != http.StatusOK {
		t.Fatalf("expected list statuses status 200, got %d: %s", listStatuses.Code, listStatuses.Body.String())
	}
	var statusBody struct {
		Items []tracker.ProjectStatus `json:"items"`
	}
	if err := json.Unmarshal(listStatuses.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode statuses: %v", err)
	}
	if len(statusBody.Items) != 3 || statusBody.Items[0].Slug != "todo" {
		t.Fatalf("unexpected statuses: %#v", statusBody.Items)
	}

	replaceStatusesReq := httptest.NewRequest(http.MethodPut, "/api/projects/"+project.ID+"/statuses", mustJSON(t, map[string]any{
		"statuses": []map[string]string{
			{"slug": "todo", "name": "Todo"},
			{"slug": "in_progress", "name": "In Progress"},
			{"slug": "review", "name": "Review"},
			{"slug": "done", "name": "Done"},
		},
	}))
	addSessionCSRF(replaceStatusesReq, session, csrf)
	replaceStatuses := httptest.NewRecorder()
	handler.ServeHTTP(replaceStatuses, replaceStatusesReq)
	if replaceStatuses.Code != http.StatusOK {
		t.Fatalf("expected replace statuses status 200, got %d: %s", replaceStatuses.Code, replaceStatuses.Body.String())
	}

	createBoardReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/boards", mustJSON(t, map[string]any{
		"name":         "Review Board",
		"description":  "Review workflow",
		"status_slugs": []string{"todo", "review", "done"},
	}))
	addSessionCSRF(createBoardReq, session, csrf)
	createBoard := httptest.NewRecorder()
	handler.ServeHTTP(createBoard, createBoardReq)
	if createBoard.Code != http.StatusCreated {
		t.Fatalf("expected create board status 201, got %d: %s", createBoard.Code, createBoard.Body.String())
	}
	var board tracker.Board
	if err := json.Unmarshal(createBoard.Body.Bytes(), &board); err != nil {
		t.Fatalf("decode board: %v", err)
	}
	if board.ID == "" || len(board.Columns) != 3 || board.Columns[1].StatusSlug != "review" {
		t.Fatalf("unexpected board: %#v", board)
	}

	createCustomFieldReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/custom-fields", mustJSON(t, map[string]any{
		"key":        "severity",
		"name":       "Severity",
		"field_type": "single_select",
		"required":   true,
		"options":    []string{"Low", "High"},
	}))
	addSessionCSRF(createCustomFieldReq, session, csrf)
	createCustomField := httptest.NewRecorder()
	handler.ServeHTTP(createCustomField, createCustomFieldReq)
	if createCustomField.Code != http.StatusCreated {
		t.Fatalf("expected create custom field status 201, got %d: %s", createCustomField.Code, createCustomField.Body.String())
	}
	var customField tracker.CustomFieldDefinition
	if err := json.Unmarshal(createCustomField.Body.Bytes(), &customField); err != nil {
		t.Fatalf("decode custom field: %v", err)
	}
	if customField.ID == "" || len(customField.Options) != 2 {
		t.Fatalf("unexpected custom field: %#v", customField)
	}

	listCustomFieldsReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/custom-fields", nil)
	listCustomFieldsReq.AddCookie(session)
	listCustomFields := httptest.NewRecorder()
	handler.ServeHTTP(listCustomFields, listCustomFieldsReq)
	if listCustomFields.Code != http.StatusOK {
		t.Fatalf("expected list custom fields status 200, got %d: %s", listCustomFields.Code, listCustomFields.Body.String())
	}

	listProjectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	listProjectsReq.AddCookie(session)
	listProjects := httptest.NewRecorder()
	handler.ServeHTTP(listProjects, listProjectsReq)
	if listProjects.Code != http.StatusOK {
		t.Fatalf("expected list projects status 200, got %d: %s", listProjects.Code, listProjects.Body.String())
	}

	createTicketReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/tickets", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"title":       "First API ticket",
			"description": "Created through HTTP",
			"priority":    "High",
			"type":        "Bug",
			"labels":      []string{"backend", "API", "backend"},
			"custom_fields": map[string]any{
				"severity": "High",
			},
		},
	}))
	addSessionCSRF(createTicketReq, session, csrf)
	createTicket := httptest.NewRecorder()
	handler.ServeHTTP(createTicket, createTicketReq)
	if createTicket.Code != http.StatusCreated {
		t.Fatalf("expected create ticket status 201, got %d: %s", createTicket.Code, createTicket.Body.String())
	}
	ticket := decodeTicketResourceAsTracker(t, createTicket.Body.Bytes())
	if ticket.ID == "" || ticket.Key != "CORE-1" || ticket.Status != "todo" {
		t.Fatalf("unexpected ticket: %#v", ticket)
	}
	if ticket.CustomFields["severity"] != "High" {
		t.Fatalf("unexpected ticket custom fields: %#v", ticket.CustomFields)
	}
	if !slices.Equal(ticket.Labels, []string{"api", "backend"}) {
		t.Fatalf("unexpected ticket labels: %#v", ticket.Labels)
	}

	getTicketReq := httptest.NewRequest(http.MethodGet, "/api/tickets/"+ticket.ID, nil)
	getTicketReq.AddCookie(session)
	getTicket := httptest.NewRecorder()
	handler.ServeHTTP(getTicket, getTicketReq)
	if getTicket.Code != http.StatusOK {
		t.Fatalf("expected get ticket status 200, got %d: %s", getTicket.Code, getTicket.Body.String())
	}
	fetchedTicket := decodeTicketResourceAsTracker(t, getTicket.Body.Bytes())
	if !slices.Equal(fetchedTicket.Labels, []string{"api", "backend"}) {
		t.Fatalf("unexpected fetched ticket labels: %#v", fetchedTicket.Labels)
	}

	boardTicketsReq := httptest.NewRequest(http.MethodGet, "/api/boards/"+board.ID+"/tickets", nil)
	boardTicketsReq.AddCookie(session)
	boardTickets := httptest.NewRecorder()
	handler.ServeHTTP(boardTickets, boardTicketsReq)
	if boardTickets.Code != http.StatusOK {
		t.Fatalf("expected board tickets status 200, got %d: %s", boardTickets.Code, boardTickets.Body.String())
	}
	var boardTicketsBody struct {
		Board   tracker.Board `json:"board"`
		Columns []struct {
			Column  tracker.BoardColumn  `json:"column"`
			Tickets []ticketResourceBody `json:"tickets"`
		} `json:"columns"`
	}
	if err := json.Unmarshal(boardTickets.Body.Bytes(), &boardTicketsBody); err != nil {
		t.Fatalf("decode board tickets: %v", err)
	}
	if len(boardTicketsBody.Columns) != 3 || len(boardTicketsBody.Columns[0].Tickets) != 1 {
		t.Fatalf("unexpected board tickets: %#v", boardTicketsBody)
	}
	if !slices.Equal(boardTicketsBody.Columns[0].Tickets[0].Spec.Labels, []string{"api", "backend"}) {
		t.Fatalf("unexpected board ticket labels: %#v", boardTicketsBody.Columns[0].Tickets[0].Spec.Labels)
	}

	createSecondReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/tickets", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"title":  "Second API ticket",
			"labels": []string{"docs"},
			"custom_fields": map[string]any{
				"severity": "Low",
			},
		},
	}))
	addSessionCSRF(createSecondReq, session, csrf)
	createSecond := httptest.NewRecorder()
	handler.ServeHTTP(createSecond, createSecondReq)
	if createSecond.Code != http.StatusCreated {
		t.Fatalf("expected create second ticket status 201, got %d: %s", createSecond.Code, createSecond.Body.String())
	}
	second := decodeTicketResourceAsTracker(t, createSecond.Body.Bytes())
	if !slices.Equal(second.Labels, []string{"docs"}) {
		t.Fatalf("unexpected second ticket labels: %#v", second.Labels)
	}

	listByLabelReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/tickets?label=Backend", nil)
	listByLabelReq.AddCookie(session)
	listByLabel := httptest.NewRecorder()
	handler.ServeHTTP(listByLabel, listByLabelReq)
	if listByLabel.Code != http.StatusOK {
		t.Fatalf("expected list by label status 200, got %d: %s", listByLabel.Code, listByLabel.Body.String())
	}
	var labelList struct {
		Items []ticketResourceBody `json:"items"`
	}
	if err := json.Unmarshal(listByLabel.Body.Bytes(), &labelList); err != nil {
		t.Fatalf("decode label ticket list: %v", err)
	}
	if len(labelList.Items) != 1 || labelList.Items[0].Metadata.ID != ticket.ID || !slices.Equal(labelList.Items[0].Spec.Labels, []string{"api", "backend"}) {
		t.Fatalf("unexpected label ticket list: %#v", labelList.Items)
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
		Items []ticketResourceBody `json:"items"`
	}
	if err := json.Unmarshal(listBacklog.Body.Bytes(), &backlog); err != nil {
		t.Fatalf("decode backlog: %v", err)
	}
	if len(backlog.Items) != 2 || backlog.Items[0].Metadata.ID != second.ID || backlog.Items[0].Spec.Rank != "000001" {
		t.Fatalf("unexpected backlog: %#v", backlog.Items)
	}
	if !slices.Equal(backlog.Items[0].Spec.Labels, []string{"docs"}) {
		t.Fatalf("unexpected backlog labels: %#v", backlog.Items[0].Spec.Labels)
	}

	createEpicReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/tickets", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"title":      "Roadmap epic",
			"type":       "Epic",
			"start_date": "2026-07-01",
			"due_date":   "2026-07-31",
			"labels":     []string{"roadmap"},
			"custom_fields": map[string]any{
				"severity": "High",
			},
		},
	}))
	addSessionCSRF(createEpicReq, session, csrf)
	createEpic := httptest.NewRecorder()
	handler.ServeHTTP(createEpic, createEpicReq)
	if createEpic.Code != http.StatusCreated {
		t.Fatalf("expected create epic status 201, got %d: %s", createEpic.Code, createEpic.Body.String())
	}
	epic := decodeTicketResourceAsTracker(t, createEpic.Body.Bytes())
	if epic.Type != "epic" || epic.StartDate != "2026-07-01" || epic.DueDate != "2026-07-31" {
		t.Fatalf("unexpected epic: %#v", epic)
	}

	roadmapChildReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/tickets", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"title":            "Roadmap child",
			"status":           "done",
			"parent_ticket_id": epic.ID,
			"custom_fields": map[string]any{
				"severity": "Low",
			},
		},
	}))
	addSessionCSRF(roadmapChildReq, session, csrf)
	roadmapChild := httptest.NewRecorder()
	handler.ServeHTTP(roadmapChild, roadmapChildReq)
	if roadmapChild.Code != http.StatusCreated {
		t.Fatalf("expected create roadmap child status 201, got %d: %s", roadmapChild.Code, roadmapChild.Body.String())
	}

	roadmapReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/roadmap", nil)
	roadmapReq.AddCookie(session)
	roadmap := httptest.NewRecorder()
	handler.ServeHTTP(roadmap, roadmapReq)
	if roadmap.Code != http.StatusOK {
		t.Fatalf("expected roadmap status 200, got %d: %s", roadmap.Code, roadmap.Body.String())
	}
	var roadmapBody struct {
		Items []struct {
			Epic     ticketResourceBody      `json:"epic"`
			Progress tracker.RoadmapProgress `json:"progress"`
		} `json:"items"`
	}
	if err := json.Unmarshal(roadmap.Body.Bytes(), &roadmapBody); err != nil {
		t.Fatalf("decode roadmap: %v", err)
	}
	if len(roadmapBody.Items) != 1 || roadmapBody.Items[0].Epic.Metadata.ID != epic.ID || roadmapBody.Items[0].Progress.Total != 1 || roadmapBody.Items[0].Progress.Done != 1 {
		t.Fatalf("unexpected roadmap body: %#v", roadmapBody.Items)
	}
	if !slices.Equal(roadmapBody.Items[0].Epic.Spec.Labels, []string{"roadmap"}) {
		t.Fatalf("unexpected roadmap labels: %#v", roadmapBody.Items[0].Epic.Spec.Labels)
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
		"spec": map[string]any{
			"component_id": component.ID,
			"version_id":   version.ID,
		},
	}))
	addSessionCSRF(componentVersionUpdateReq, session, csrf)
	componentVersionUpdate := httptest.NewRecorder()
	handler.ServeHTTP(componentVersionUpdate, componentVersionUpdateReq)
	if componentVersionUpdate.Code != http.StatusOK {
		t.Fatalf("expected component/version ticket update status 200, got %d: %s", componentVersionUpdate.Code, componentVersionUpdate.Body.String())
	}
	componentVersionTicket := decodeTicketResourceAsTracker(t, componentVersionUpdate.Body.Bytes())
	if componentVersionTicket.ComponentID != component.ID || componentVersionTicket.VersionID != version.ID {
		t.Fatalf("unexpected component/version ticket: %#v", componentVersionTicket)
	}

	createSprintReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/sprints", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":       "Sprint 1",
			"goal":       "Exercise sprint API",
			"start_date": "2026-06-16",
			"end_date":   "2026-06-30",
		},
	}))
	addSessionCSRF(createSprintReq, session, csrf)
	createSprint := httptest.NewRecorder()
	handler.ServeHTTP(createSprint, createSprintReq)
	if createSprint.Code != http.StatusCreated {
		t.Fatalf("expected create sprint status 201, got %d: %s", createSprint.Code, createSprint.Body.String())
	}
	var sprint sprintResourceBody
	if err := json.Unmarshal(createSprint.Body.Bytes(), &sprint); err != nil {
		t.Fatalf("decode sprint: %v", err)
	}
	if sprint.Metadata.ID == "" || sprint.Status.State != tracker.SprintStatePlanned {
		t.Fatalf("unexpected sprint: %#v", sprint)
	}

	assignSprintReq := httptest.NewRequest(http.MethodPut, "/api/tickets/"+ticket.ID+"/sprint", mustJSON(t, map[string]any{
		"sprint_id": sprint.Metadata.ID,
	}))
	addSessionCSRF(assignSprintReq, session, csrf)
	assignSprint := httptest.NewRecorder()
	handler.ServeHTTP(assignSprint, assignSprintReq)
	if assignSprint.Code != http.StatusOK {
		t.Fatalf("expected assign sprint status 200, got %d: %s", assignSprint.Code, assignSprint.Body.String())
	}
	sprintTicket := decodeTicketResourceAsTracker(t, assignSprint.Body.Bytes())
	if sprintTicket.SprintID != sprint.Metadata.ID {
		t.Fatalf("expected ticket sprint %s, got %#v", sprint.Metadata.ID, sprintTicket)
	}

	startSprintReq := httptest.NewRequest(http.MethodPost, "/api/sprints/"+sprint.Metadata.ID+"/start", nil)
	addSessionCSRF(startSprintReq, session, csrf)
	startSprint := httptest.NewRecorder()
	handler.ServeHTTP(startSprint, startSprintReq)
	if startSprint.Code != http.StatusOK {
		t.Fatalf("expected start sprint status 200, got %d: %s", startSprint.Code, startSprint.Body.String())
	}

	completeSprintReq := httptest.NewRequest(http.MethodPost, "/api/sprints/"+sprint.Metadata.ID+"/complete", nil)
	addSessionCSRF(completeSprintReq, session, csrf)
	completeSprint := httptest.NewRecorder()
	handler.ServeHTTP(completeSprint, completeSprintReq)
	if completeSprint.Code != http.StatusOK {
		t.Fatalf("expected complete sprint status 200, got %d: %s", completeSprint.Code, completeSprint.Body.String())
	}

	status := "In_Progress"
	updateTicketReq := httptest.NewRequest(http.MethodPatch, "/api/tickets/"+ticket.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"status": status,
		},
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
		"spec": map[string]any{
			"key":  "DENY",
			"name": "Denied",
		},
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

type sprintResourceBody struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Status struct {
		State string `json:"state"`
	} `json:"status"`
}

type projectResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		CreatedBy string `json:"created_by"`
	} `json:"metadata"`
	Spec struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		LeadUserID  string `json:"lead_user_id"`
	} `json:"spec"`
}

type ticketResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Title          string         `json:"title"`
		Description    string         `json:"description"`
		Status         string         `json:"status"`
		Priority       string         `json:"priority"`
		Type           string         `json:"type"`
		AssigneeID     string         `json:"assignee_id"`
		ParentTicketID string         `json:"parent_ticket_id"`
		SprintID       string         `json:"sprint_id"`
		ComponentID    string         `json:"component_id"`
		VersionID      string         `json:"version_id"`
		Rank           string         `json:"rank"`
		StartDate      string         `json:"start_date"`
		DueDate        string         `json:"due_date"`
		Labels         []string       `json:"labels"`
		CustomFields   map[string]any `json:"custom_fields"`
	} `json:"spec"`
	Status struct {
		Key        string `json:"key"`
		ReporterID string `json:"reporter_id"`
	} `json:"status"`
}

func decodeProjectResourceAsTracker(t *testing.T, data []byte) tracker.Project {
	t.Helper()

	var body projectResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode project resource: %v", err)
	}
	return tracker.Project{
		ID:          body.Metadata.ID,
		Key:         body.Spec.Key,
		Name:        body.Spec.Name,
		Description: body.Spec.Description,
		LeadUserID:  body.Spec.LeadUserID,
		CreatedBy:   body.Metadata.CreatedBy,
	}
}

func decodeTicketResourceAsTracker(t *testing.T, data []byte) tracker.Ticket {
	t.Helper()

	var body ticketResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode ticket resource: %v", err)
	}
	return tracker.Ticket{
		ID:             body.Metadata.ID,
		ProjectID:      body.Metadata.ProjectID,
		Key:            body.Status.Key,
		Title:          body.Spec.Title,
		Description:    body.Spec.Description,
		Status:         body.Spec.Status,
		Priority:       body.Spec.Priority,
		Type:           body.Spec.Type,
		ReporterID:     body.Status.ReporterID,
		AssigneeID:     body.Spec.AssigneeID,
		ParentTicketID: body.Spec.ParentTicketID,
		SprintID:       body.Spec.SprintID,
		ComponentID:    body.Spec.ComponentID,
		VersionID:      body.Spec.VersionID,
		Rank:           body.Spec.Rank,
		StartDate:      body.Spec.StartDate,
		DueDate:        body.Spec.DueDate,
		Labels:         body.Spec.Labels,
		CustomFields:   body.Spec.CustomFields,
	}
}
