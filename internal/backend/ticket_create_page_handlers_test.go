package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestTicketCreatePageEndpointsLifecycle(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer)
	openRouterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "gen_create_page_handler",
			"choices": []map[string]any{{
				"message": map[string]any{
					"role":    "assistant",
					"content": `{"field_layout":[{"key":"title","required":true}],"defaults":{"priority":"High"}}`,
				},
			}},
		})
	}))
	defer openRouterServer.Close()
	openRouterService := openrouter.NewService(db.SQL, openrouter.WithBaseURL(openRouterServer.URL))
	createPageService := tracker.NewCreatePageService(db.SQL, trackerService, authorizer, tracker.WithCreatePageOpenRouterService(openRouterService))
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithTrackerService(trackerService),
		WithCreatePageService(createPageService),
		WithOpenRouterService(openRouterService),
	)
	project, err := trackerService.CreateProject(ctx, authz.Principal{UserID: bootstrap.UserID}, tracker.CreateProjectInput{
		Key:        "FORM",
		Name:       "Forms",
		LeadUserID: bootstrap.UserID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	provider, err := openRouterService.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "Form AI",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-form",
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("create OpenRouter provider: %v", err)
	}
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/projects/"+project.ID+"/ticket-create-pages", map[string]any{
		"spec": ticketCreatePageBody("Bug Intake", "bug-intake", bootstrap.UserID, provider.ID),
	}, []*http.Cookie{session})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/ticket-create-pages", mustJSON(t, map[string]any{
		"spec": ticketCreatePageBody("Bug Intake", "bug-intake", bootstrap.UserID, provider.ID),
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
	if created.Spec.FormAIPrompt != "Build a focused bug intake form." || created.Spec.FormAIProviderID != provider.ID {
		t.Fatalf("expected AI form fields in management response, got %#v", created.Spec)
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
	if listed.Status.Items[0].Spec.FormAIPrompt != created.Spec.FormAIPrompt || listed.Status.Items[0].Spec.FormAIProviderID != provider.ID {
		t.Fatalf("expected AI form fields in management list response, got %#v", listed.Status.Items[0].Spec)
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
	if schema.Spec.FormAIPrompt != "" || schema.Spec.FormAIProviderID != "" {
		t.Fatalf("schema response must redact AI form fields, got %#v", schema.Spec)
	}
	assertTicketCreatePageSchemaRedactsFormLogic(t, schemaRec.Body.Bytes())

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
			"name":                "Bug Intake Disabled",
			"enabled":             false,
			"form_ai_prompt":      "Updated prompt",
			"form_ai_provider_id": provider.ID,
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
	if updated.Spec.FormAIPrompt != "Updated prompt" || updated.Spec.FormAIProviderID != provider.ID {
		t.Fatalf("expected updated AI form fields, got %#v", updated.Spec)
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

func TestTicketCreatePageSchemaRunsLuaFormLogic(t *testing.T) {
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
		Key:        "DYN",
		Name:       "Dynamic Forms",
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

	createReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+project.ID+"/ticket-create-pages", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":            "Dynamic Intake",
			"slug":            "dynamic-intake",
			"enabled":         true,
			"owner_user_id":   bootstrap.UserID,
			"field_layout":    []map[string]any{{"key": "title", "type": "text"}},
			"defaults":        map[string]any{"priority": "Low"},
			"form_lua_script": `return { field_layout = { { key = "title", type = "text", required = true }, { key = "component", type = "single-select", options = { "api", "ui" } } }, defaults = { priority = "High" } }`,
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create page status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeTicketCreatePageResource(t, create.Body.Bytes())
	if created.Spec.FormLuaScript == "" {
		t.Fatalf("expected form Lua script in management response: %#v", created.Spec)
	}

	schemaReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+project.ID+"/ticket-create-pages/dynamic-intake/schema", nil)
	schemaReq.AddCookie(session)
	schemaRec := httptest.NewRecorder()
	handler.ServeHTTP(schemaRec, schemaReq)
	if schemaRec.Code != http.StatusOK {
		t.Fatalf("expected schema status 200, got %d: %s", schemaRec.Code, schemaRec.Body.String())
	}
	schema := decodeTicketCreatePageSchema(t, schemaRec.Body.Bytes())
	if len(schema.Spec.FieldLayout) != 2 || schema.Spec.FieldLayout[1]["key"] != "component" || schema.Spec.Defaults["priority"] != "High" {
		t.Fatalf("expected dynamic schema/defaults, got %#v", schema.Spec)
	}
	if schema.Spec.FormLuaScript != "" {
		t.Fatalf("schema response must redact form Lua script, got %#v", schema.Spec)
	}
	assertTicketCreatePageSchemaRedactsFormLogic(t, schemaRec.Body.Bytes())
}

func assertTicketCreatePageSchemaRedactsFormLogic(t *testing.T, data []byte) {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode schema response map: %v", err)
	}
	spec, ok := body["spec"].(map[string]any)
	if !ok {
		t.Fatalf("schema response missing spec object: %#v", body)
	}
	for _, field := range []string{"form_lua_script", "form_ai_prompt", "form_ai_provider_id"} {
		if _, ok := spec[field]; ok {
			t.Fatalf("schema response must omit %s, got %#v", field, spec)
		}
	}
}

func ticketCreatePageBody(name string, slug string, ownerUserID string, aiProviderID string) map[string]any {
	return map[string]any{
		"name":                name,
		"slug":                slug,
		"description":         "External intake form",
		"enabled":             true,
		"target_type":         "bug",
		"target_status":       "todo",
		"form_ai_prompt":      "Build a focused bug intake form.",
		"form_ai_provider_id": aiProviderID,
		"owner_user_id":       ownerUserID,
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
		Name             string           `json:"name"`
		Slug             string           `json:"slug"`
		Description      string           `json:"description"`
		Enabled          bool             `json:"enabled"`
		TargetType       string           `json:"target_type"`
		TargetStatus     string           `json:"target_status"`
		FieldLayout      []map[string]any `json:"field_layout"`
		Defaults         map[string]any   `json:"defaults"`
		FormLuaScript    string           `json:"form_lua_script"`
		FormAIPrompt     string           `json:"form_ai_prompt"`
		FormAIProviderID string           `json:"form_ai_provider_id"`
		OwnerUserID      string           `json:"owner_user_id"`
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
		FieldLayout      []map[string]any `json:"field_layout"`
		Defaults         map[string]any   `json:"defaults"`
		FormLuaScript    string           `json:"form_lua_script"`
		FormAIPrompt     string           `json:"form_ai_prompt"`
		FormAIProviderID string           `json:"form_ai_provider_id"`
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
