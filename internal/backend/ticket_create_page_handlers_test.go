package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestTicketCreatePageEndpointsLifecycle(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer)
	createPageService := tracker.NewCreatePageService(db.SQL, trackerService, authorizer)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithTrackerService(trackerService),
		WithCreatePageService(createPageService),
	)
	project, err := trackerService.CreateProject(ctx, authz.Principal{UserID: bootstrap.UserID}, tracker.CreateProjectInput{
		Key:        "FORM",
		Name:       "Forms",
		LeadUserID: bootstrap.UserID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/projects/"+project.ID+"/ticket-create-pages", map[string]any{
		"spec": ticketCreatePageBody("Bug Intake", "bug-intake", bootstrap.UserID),
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/ticket-create-pages", mustJSON(t, map[string]any{
		"spec": ticketCreatePageBody("Bug Intake", "bug-intake", bootstrap.UserID),
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create page status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeTicketCreatePageResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Metadata.ProjectID != project.ID || created.Spec.Name != "Bug Intake" || created.Spec.Slug != "bug-intake" || !created.Spec.Enabled || created.Spec.OwnerUserID != bootstrap.UserID {
		t.Fatalf("unexpected created page: %#v", created)
	}
	if created.Spec.TargetType != "bug" || created.Spec.TargetStatus != "todo" || created.Spec.Defaults["priority"] != "High" {
		t.Fatalf("unexpected created page defaults: %#v", created.Spec)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/ticket-create-pages", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list page status 200, got %d: %s", list.Code, list.Body.String())
	}
	listed := decodeTicketCreatePageList(t, list.Body.Bytes())
	if listed.Metadata.Count != 1 || len(listed.Status.Items) != 1 || listed.Status.Items[0].Metadata.ID != created.Metadata.ID {
		t.Fatalf("unexpected create page list: %#v", listed)
	}

	schemaReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/ticket-create-pages/bug-intake/schema", nil)
	schemaReq.AddCookie(session)
	schemaRec := httptest.NewRecorder()
	handler.ServeHTTP(schemaRec, schemaReq)
	if schemaRec.Code != http.StatusOK {
		t.Fatalf("expected schema status 200, got %d: %s", schemaRec.Code, schemaRec.Body.String())
	}
	schema := decodeTicketCreatePageSchema(t, schemaRec.Body.Bytes())
	if schema.Metadata.PageID != created.Metadata.ID || schema.Metadata.Slug != "bug-intake" || !schema.Status.Enabled || len(schema.Spec.FieldLayout) != 1 {
		t.Fatalf("unexpected create page schema: %#v", schema)
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/ticket-create-pages/bug-intake/submit", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"ticket": map[string]any{
				"title":       "Broken login",
				"description": "Submitted through a custom page",
				"labels":      []string{"customer"},
			},
		},
	}))
	addSessionCSRF(submitReq, session, csrf)
	submit := httptest.NewRecorder()
	handler.ServeHTTP(submit, submitReq)
	if submit.Code != http.StatusCreated {
		t.Fatalf("expected submit status 201, got %d: %s", submit.Code, submit.Body.String())
	}
	ticket := decodeTicketCreatePageTicket(t, submit.Body.Bytes())
	if ticket.Metadata.ProjectID != project.ID || ticket.Spec.Title != "Broken login" || ticket.Spec.Type != "bug" || ticket.Spec.Priority != "high" || ticket.Spec.Status != "todo" || ticket.Status.ReporterID != bootstrap.UserID {
		t.Fatalf("unexpected submitted ticket: %#v", ticket)
	}
	if len(ticket.Spec.Labels) != 1 || ticket.Spec.Labels[0] != "customer" {
		t.Fatalf("expected submitted labels to override defaults, got %#v", ticket.Spec.Labels)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/ticket-create-pages/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":    "Bug Intake Disabled",
			"enabled": false,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeTicketCreatePageResource(t, update.Body.Bytes())
	if updated.Spec.Enabled || updated.Spec.Name != "Bug Intake Disabled" {
		t.Fatalf("unexpected updated page: %#v", updated)
	}

	disabledSchemaReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/ticket-create-pages/bug-intake/schema", nil)
	disabledSchemaReq.AddCookie(session)
	disabledSchema := httptest.NewRecorder()
	handler.ServeHTTP(disabledSchema, disabledSchemaReq)
	if disabledSchema.Code != http.StatusNotFound {
		t.Fatalf("expected disabled schema status 404, got %d: %s", disabledSchema.Code, disabledSchema.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/ticket-create-pages/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, session, csrf)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func TestTicketCreatePageEndpointsRequirePermission(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer)
	createPageService := tracker.NewCreatePageService(db.SQL, trackerService, authorizer)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithTrackerService(trackerService),
		WithCreatePageService(createPageService),
	)
	project, err := trackerService.CreateProject(ctx, authz.Principal{UserID: bootstrap.UserID}, tracker.CreateProjectInput{
		Key:        "PERM",
		Name:       "Permissions",
		LeadUserID: bootstrap.UserID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := createPageService.Create(ctx, authz.Principal{UserID: bootstrap.UserID}, tracker.CreateCreatePageInput{
		ProjectID: project.ID,
		Name:      "Task Intake",
		Slug:      "task-intake",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("create page: %v", err)
	}
	viewer, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "viewer"})
	if err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": viewer.Username,
		"password": viewer.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	listReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/ticket-create-pages", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden list, got %d: %s", list.Code, list.Body.String())
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/ticket-create-pages/task-intake/submit", mustJSON(t, map[string]any{
		"spec": map[string]any{"ticket": map[string]any{"title": "forbidden"}},
	}))
	addSessionCSRF(submitReq, session, csrf)
	submit := httptest.NewRecorder()
	handler.ServeHTTP(submit, submitReq)
	if submit.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden submit, got %d: %s", submit.Code, submit.Body.String())
	}
}

func ticketCreatePageBody(name string, slug string, ownerUserID string) map[string]any {
	return map[string]any{
		"name":          name,
		"slug":          slug,
		"description":   "External intake form",
		"enabled":       true,
		"target_type":   "bug",
		"target_status": "todo",
		"owner_user_id": ownerUserID,
		"field_layout": []map[string]any{
			{"key": "title", "required": true},
		},
		"defaults": map[string]any{
			"priority": "High",
			"labels":   []string{"intake"},
		},
	}
}

type ticketCreatePageResourceBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Name         string           `json:"name"`
		Slug         string           `json:"slug"`
		Description  string           `json:"description"`
		Enabled      bool             `json:"enabled"`
		TargetType   string           `json:"target_type"`
		TargetStatus string           `json:"target_status"`
		FieldLayout  []map[string]any `json:"field_layout"`
		Defaults     map[string]any   `json:"defaults"`
		OwnerUserID  string           `json:"owner_user_id"`
	} `json:"spec"`
}

type ticketCreatePageListBody struct {
	Metadata struct {
		Count int `json:"count"`
	} `json:"metadata"`
	Status struct {
		Items []ticketCreatePageResourceBody `json:"items"`
	} `json:"status"`
}

type ticketCreatePageSchemaBody struct {
	Metadata struct {
		PageID string `json:"page_id"`
		Slug   string `json:"slug"`
	} `json:"metadata"`
	Spec struct {
		FieldLayout []map[string]any `json:"field_layout"`
	} `json:"spec"`
	Status struct {
		Enabled bool `json:"enabled"`
	} `json:"status"`
}

type ticketCreatePageTicketBody struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Title       string   `json:"title"`
		Status      string   `json:"status"`
		Priority    string   `json:"priority"`
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Labels      []string `json:"labels"`
	} `json:"spec"`
	Status struct {
		ReporterID string `json:"reporter_id"`
	} `json:"status"`
}

func decodeTicketCreatePageResource(t *testing.T, data []byte) ticketCreatePageResourceBody {
	t.Helper()

	var body ticketCreatePageResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode ticket create page: %v", err)
	}
	return body
}

func decodeTicketCreatePageList(t *testing.T, data []byte) ticketCreatePageListBody {
	t.Helper()

	var body ticketCreatePageListBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode ticket create page list: %v", err)
	}
	return body
}

func decodeTicketCreatePageSchema(t *testing.T, data []byte) ticketCreatePageSchemaBody {
	t.Helper()

	var body ticketCreatePageSchemaBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode ticket create page schema: %v", err)
	}
	return body
}

func decodeTicketCreatePageTicket(t *testing.T, data []byte) ticketCreatePageTicketBody {
	t.Helper()

	var body ticketCreatePageTicketBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode submitted ticket: %v", err)
	}
	return body
}
