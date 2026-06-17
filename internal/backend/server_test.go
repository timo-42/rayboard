package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var body struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Service string `json:"service"`
		} `json:"spec"`
		Status struct {
			State string `json:"state"`
		} `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Metadata.ID != "backend" || body.Spec.Service != "backend" || body.Status.State != "ok" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestOpenAPIJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "openapi+json") {
		t.Fatalf("expected OpenAPI content type, got %q", contentType)
	}

	var body struct {
		OpenAPI string `json:"openapi"`
		Info    struct {
			Title string `json:"title"`
		} `json:"info"`
		Paths      map[string]any `json:"paths"`
		Components struct {
			SecuritySchemes map[string]any `json:"securitySchemes"`
		} `json:"components"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("decode raw OpenAPI response: %v", err)
	}
	if body.OpenAPI == "" || body.Info.Title != "Rayboard API" {
		t.Fatalf("unexpected OpenAPI metadata: %#v", body)
	}
	for _, path := range []string{"/api/health", "/api/login", "/api/projects/{project_id}/tickets", "/api/cron-jobs"} {
		if _, ok := body.Paths[path]; !ok {
			t.Fatalf("expected path %s in OpenAPI document", path)
		}
	}
	for _, scheme := range []string{"bearerToken", "sessionCookie", "csrfToken"} {
		if _, ok := body.Components.SecuritySchemes[scheme]; !ok {
			t.Fatalf("expected security scheme %s in OpenAPI document", scheme)
		}
	}
	assertRequestBodyFields(t, spec, "/api/login", http.MethodPost, []string{"spec"}, []string{"spec", "username"}, []string{"spec", "password"})
	assertResponseBodyFields(t, spec, "/api/login", http.MethodPost, "200", []string{"metadata"}, []string{"metadata", "user_id"}, []string{"spec"}, []string{"spec", "username"}, []string{"status"}, []string{"status", "auth_kind"})
	assertResponseBodyFields(t, spec, "/api/me/effective-permissions", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "user_id"}, []string{"spec"}, []string{"spec", "scope"}, []string{"status"}, []string{"status", "permissions"})
	assertRequestBodyFields(t, spec, "/api/tokens", http.MethodPost, []string{"spec"}, []string{"spec", "name"})
	assertResponseBodyFields(t, spec, "/api/tokens", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "name"}, []string{"status"}, []string{"status", "token"})
	assertRequestBodyFields(t, spec, "/api/users", http.MethodPost, []string{"spec"}, []string{"spec", "username"}, []string{"spec", "display_name"}, []string{"spec", "disabled"})
	assertResponseBodyFields(t, spec, "/api/users/{user_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "username"}, []string{"status"}, []string{"status", "disabled"})
	assertResponseBodyFields(t, spec, "/api/users/{user_id}/effective-permissions", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "user_id"}, []string{"spec"}, []string{"spec", "scope"}, []string{"status"}, []string{"status", "permissions"})
	assertRequestBodyFields(t, spec, "/api/groups", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "display_name"})
	assertResponseBodyFields(t, spec, "/api/groups", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "name"}, []string{"status"})
	assertRequestBodyFields(t, spec, "/api/role-bindings", http.MethodPost, []string{"spec"}, []string{"spec", "role_name"}, []string{"spec", "subject_type"}, []string{"spec", "subject_id"}, []string{"spec", "scope"})
	assertResponseBodyFields(t, spec, "/api/role-bindings", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "role_name"}, []string{"status"}, []string{"status", "role_id"})
	assertRequestBodyFields(t, spec, "/api/projects", http.MethodPost, []string{"spec"}, []string{"spec", "key"}, []string{"spec", "name"}, []string{"spec", "description"}, []string{"spec", "lead_user_id"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "key"}, []string{"status"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/tickets", http.MethodPost, []string{"spec"}, []string{"spec", "title"}, []string{"spec", "labels"})
	assertResponseBodyFields(t, spec, "/api/tickets/{ticket_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "title"}, []string{"status"}, []string{"status", "key"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/backlog", http.MethodPatch, []string{"spec"}, []string{"spec", "ticket_ids"})
	assertResponseBodyFields(t, spec, "/api/tickets/{ticket_id}/activity", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "count"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/statuses", http.MethodPut, []string{"spec"}, []string{"spec", "statuses"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/statuses", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "count"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/boards", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "status_slugs"})
	assertRequestBodyFields(t, spec, "/api/boards/{board_id}", http.MethodPatch, []string{"spec"}, []string{"spec", "name"})
	assertResponseBodyFields(t, spec, "/api/boards/{board_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "columns"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/components", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "owner_user_id"})
	assertResponseBodyFields(t, spec, "/api/components/{component_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/versions", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "target_date"})
	assertResponseBodyFields(t, spec, "/api/versions/{version_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "state"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/custom-fields", http.MethodPost, []string{"spec"}, []string{"spec", "key"}, []string{"spec", "field_type"})
	assertResponseBodyFields(t, spec, "/api/custom-fields/{field_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "options"})
	assertRequestBodyFields(t, spec, "/api/tickets/{ticket_id}/comments", http.MethodPost, []string{"spec"}, []string{"spec", "body"})
	assertResponseBodyFields(t, spec, "/api/tickets/{ticket_id}/comments", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertResponseBodyFields(t, spec, "/api/tickets/{ticket_id}/attachments", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertRequestBodyFields(t, spec, "/api/search", http.MethodPost, []string{"spec"}, []string{"spec", "text"}, []string{"spec", "sort"})
	assertResponseBodyFields(t, spec, "/api/search", http.MethodPost, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertRequestBodyFields(t, spec, "/api/saved-views", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "query"})
	assertResponseBodyFields(t, spec, "/api/saved-views/{view_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"})
	assertResponseBodyFields(t, spec, "/api/notifications", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"}, []string{"status", "items", "status", "read_at"})
	assertResponseBodyFields(t, spec, "/api/me/notification-preferences", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "scope_type"}, []string{"spec"}, []string{"spec", "in_app_enabled"}, []string{"spec", "external_enabled"}, []string{"status"}, []string{"status", "customized"})
	assertRequestBodyFields(t, spec, "/api/me/notification-preferences", http.MethodPatch, []string{"spec"}, []string{"spec", "in_app_enabled"}, []string{"spec", "external_enabled"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/notification-preferences", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"status"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/notification-preferences", http.MethodPatch, []string{"spec"}, []string{"spec", "comment_enabled"})
	assertRequestBodyFields(t, spec, "/api/notification-policies", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "event_types"}, []string{"spec", "destination_ids"})
	assertResponseBodyFields(t, spec, "/api/notification-policies/{policy_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "scope_type"}, []string{"spec"}, []string{"spec", "event_types"}, []string{"spec", "destination_ids"}, []string{"status"})
	assertRequestBodyFields(t, spec, "/api/notification-policies/{policy_id}", http.MethodPatch, []string{"spec"}, []string{"spec", "enabled"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/notification-policies", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "event_types"}, []string{"spec", "destination_ids"})
	assertResponseBodyFields(t, spec, "/api/notification-deliveries", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "count"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "metadata", "scope_type"}, []string{"status", "items", "metadata", "destination_id"}, []string{"status", "items", "spec"}, []string{"status", "items", "spec", "event_type"}, []string{"status", "items", "status"}, []string{"status", "items", "status", "state"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/notification-deliveries", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertResponseBodyFields(t, spec, "/api/notification-deliveries/{delivery_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "scope_type"}, []string{"metadata", "destination_id"}, []string{"spec"}, []string{"spec", "event_type"}, []string{"spec", "message"}, []string{"status"}, []string{"status", "state"}, []string{"status", "attempt_count"})
	assertResponseBodyFields(t, spec, "/api/notification-deliveries/{delivery_id}/retry", http.MethodPost, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "state"}, []string{"status", "next_attempt_at"})
	assertRequestBodyFields(t, spec, "/api/notification-destinations", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "shoutrrr_url"}, []string{"spec", "enabled"})
	assertResponseBodyFields(t, spec, "/api/notification-destinations/{destination_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"metadata", "scope_type"}, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "type"}, []string{"status"}, []string{"status", "url_set"})
	assertRequestBodyFields(t, spec, "/api/notification-destinations/{destination_id}", http.MethodPatch, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "shoutrrr_url"}, []string{"spec", "enabled"})
	assertRequestBodyFields(t, spec, "/api/notification-destinations/{destination_id}/test-send", http.MethodPost, []string{"spec"}, []string{"spec", "message"})
	assertResponseBodyFields(t, spec, "/api/notification-destinations/{destination_id}/test-send", http.MethodPost, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "last_delivery_status"}, []string{"status", "last_delivery_at"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/notification-destinations", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "shoutrrr_url"})
	assertRequestBodyFields(t, spec, "/api/openrouter-providers", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "default_model"}, []string{"spec", "api_key"})
	assertResponseBodyFields(t, spec, "/api/openrouter-providers/{provider_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"spec", "default_model"}, []string{"status"}, []string{"status", "api_key_set"})
	assertRequestBodyFields(t, spec, "/api/cron-jobs", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "schedule"}, []string{"spec", "engine"})
	assertDiscriminatedEngineSchema(t, spec, requestBodySchema(t, spec, "/api/cron-jobs", http.MethodPost), []string{"spec", "engine"})
	assertResponseBodyFields(t, spec, "/api/cron-jobs/{job_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "next_run_at"})
	assertResponseBodyFields(t, spec, "/api/cron-jobs/{job_id}/run", http.MethodPost, "202", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "state"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/webhooks", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "direction"}, []string{"spec", "actor_user_id"}, []string{"spec", "event_types"}, []string{"spec", "engine"})
	assertDiscriminatedEngineSchema(t, spec, requestBodySchema(t, spec, "/api/projects/{project_id}/webhooks", http.MethodPost), []string{"spec", "engine"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/webhooks", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "count"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"}, []string{"status", "items", "status", "token_set"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/webhooks", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"spec", "event_types"}, []string{"spec", "engine"}, []string{"status"}, []string{"status", "token_set"}, []string{"status", "token"})
	assertResponseBodyFields(t, spec, "/api/webhook-definitions/{webhook_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "event_types"}, []string{"status"}, []string{"status", "token_set"})
	assertRequestBodyFields(t, spec, "/api/webhook-definitions/{webhook_id}", http.MethodPatch, []string{"spec"}, []string{"spec", "enabled"}, []string{"spec", "event_types"})
	assertResponseBodyFields(t, spec, "/api/webhook-definitions/{webhook_id}/rotate-token", http.MethodPost, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "token"})
	assertResponseBodyFields(t, spec, "/api/webhook-definitions/{webhook_id}/runs", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "status"})
	assertRequestBodyFields(t, spec, "/api/webhooks/incoming/{webhook_id}", http.MethodPost, []string{"spec"}, []string{"spec", "payload"})
	assertResponseBodyFields(t, spec, "/api/webhooks/incoming/{webhook_id}", http.MethodPost, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "state"}, []string{"status", "output"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/ticket-hooks", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "event"}, []string{"spec", "phase"}, []string{"spec", "engine"})
	assertDiscriminatedEngineSchema(t, spec, requestBodySchema(t, spec, "/api/projects/{project_id}/ticket-hooks", http.MethodPost), []string{"spec", "engine"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/ticket-hooks", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "count"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "metadata", "project_id"}, []string{"status", "items", "spec"}, []string{"status", "items", "spec", "engine"}, []string{"status", "items", "status"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/ticket-hooks", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "id"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"spec", "event"}, []string{"spec", "phase"}, []string{"spec", "engine"}, []string{"status"}, []string{"status", "last_error"})
	assertResponseBodyFields(t, spec, "/api/ticket-hooks/{hook_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "engine"}, []string{"status"}, []string{"status", "last_error"})
	assertRequestBodyFields(t, spec, "/api/ticket-hooks/{hook_id}", http.MethodPatch, []string{"spec"}, []string{"spec", "enabled"}, []string{"spec", "position"}, []string{"spec", "engine"})
	assertRequestBodyFields(t, spec, "/api/ticket-hooks/{hook_id}/preview", http.MethodPost, []string{"spec"}, []string{"spec", "ticket"}, []string{"spec", "current"})
	assertResponseBodyFields(t, spec, "/api/ticket-hooks/{hook_id}/preview", http.MethodPost, "200", []string{"metadata"}, []string{"metadata", "hook_id"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"spec", "ticket"}, []string{"status"}, []string{"status", "output"}, []string{"status", "logs"}, []string{"status", "error"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/ticket-create-pages", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "slug"}, []string{"spec", "enabled"}, []string{"spec", "target_type"}, []string{"spec", "target_status"}, []string{"spec", "field_layout"}, []string{"spec", "defaults"}, []string{"spec", "owner_user_id"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/ticket-create-pages", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "count"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "metadata", "project_id"}, []string{"status", "items", "spec"}, []string{"status", "items", "spec", "field_layout"}, []string{"status", "items", "status"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/ticket-create-pages", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "id"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"spec", "slug"}, []string{"spec", "defaults"}, []string{"status"})
	assertResponseBodyFields(t, spec, "/api/ticket-create-pages/{page_id}", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "id"}, []string{"metadata", "owner_user_id"}, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "owner_user_id"}, []string{"status"}, []string{"status", "deleted_at"})
	assertRequestBodyFields(t, spec, "/api/ticket-create-pages/{page_id}", http.MethodPatch, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "slug"}, []string{"spec", "enabled"}, []string{"spec", "owner_user_id"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/ticket-create-pages/{slug}/schema", http.MethodGet, "200", []string{"metadata"}, []string{"metadata", "page_id"}, []string{"metadata", "project_id"}, []string{"metadata", "slug"}, []string{"spec"}, []string{"spec", "field_layout"}, []string{"spec", "defaults"}, []string{"status"}, []string{"status", "enabled"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/ticket-create-pages/{slug}/submit", http.MethodPost, []string{"spec"}, []string{"spec", "ticket"}, []string{"spec", "ticket", "title"}, []string{"spec", "ticket", "custom_fields"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/ticket-create-pages/{slug}/submit", http.MethodPost, "201", []string{"metadata"}, []string{"metadata", "id"}, []string{"metadata", "project_id"}, []string{"spec"}, []string{"spec", "title"}, []string{"status"}, []string{"status", "key"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/sprints", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "goal"}, []string{"spec", "start_date"}, []string{"spec", "end_date"})
	assertResponseBodyFields(t, spec, "/api/sprints/{sprint_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "state"})
	assertRequestBodyFields(t, spec, "/api/tickets/{ticket_id}/sprint", http.MethodPut, []string{"spec"}, []string{"spec", "sprint_id"})
	assertResponseBodyFields(t, spec, "/api/boards/{board_id}/tickets", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"spec", "board"}, []string{"status"}, []string{"status", "columns"}, []string{"status", "columns", "tickets"}, []string{"status", "columns", "tickets", "metadata"}, []string{"status", "columns", "tickets", "spec"}, []string{"status", "columns", "tickets", "status"})
	assertResponseBodyFields(t, spec, "/api/projects/{project_id}/roadmap", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "items"}, []string{"status", "items", "metadata"}, []string{"status", "items", "spec"}, []string{"status", "items", "spec", "epic"}, []string{"status", "items", "spec", "epic", "metadata"}, []string{"status", "items", "spec", "epic", "spec"}, []string{"status", "items", "spec", "epic", "status"}, []string{"status", "items", "status"}, []string{"status", "items", "status", "progress"})
	assertOpenAPIUsesResourceConvention(t, spec)
}

func TestAPIDocsAreServedLocally(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		if contentType != "text/html" {
			t.Fatalf("expected HTML content type, got %q", contentType)
		}
	}
	body := rec.Body.String()
	if !strings.Contains(body, "/api/openapi.json") {
		t.Fatalf("expected docs page to reference local OpenAPI JSON")
	}
	for _, localAsset := range []string{"/api/docs/swagger-ui.css", "/api/docs/swagger-ui-bundle.js", "/api/docs/swagger-ui-standalone-preset.js"} {
		if !strings.Contains(body, localAsset) {
			t.Fatalf("expected docs page to reference local Swagger UI asset %s", localAsset)
		}
	}
	for _, external := range []string{`src="https://`, `href="https://`, "unpkg.com"} {
		if strings.Contains(body, external) {
			t.Fatalf("docs page must not reference external asset %q", external)
		}
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/api/docs/swagger-ui-bundle.js", nil)
	assetRec := httptest.NewRecorder()

	NewHandler().ServeHTTP(assetRec, assetReq)

	if assetRec.Code != http.StatusOK {
		t.Fatalf("expected Swagger UI asset status 200, got %d: %s", assetRec.Code, assetRec.Body.String())
	}
	if !strings.Contains(assetRec.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("expected embedded Swagger UI bundle")
	}
}

func TestRedocDocsAreServedLocally(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/docs/redoc", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("expected HTML content type, got %q", contentType)
	}
	body := rec.Body.String()
	for _, expected := range []string{"Redoc.init", "/api/openapi.json", "Rayboard API"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected Redoc page to contain %q", expected)
		}
	}
	for _, external := range []string{`src="https://`, `href="https://`, "cdn.jsdelivr.net", "unpkg.com"} {
		if strings.Contains(body, external) {
			t.Fatalf("redoc page must not reference external asset %q", external)
		}
	}
}

func assertRequestBodyFields(t *testing.T, spec map[string]any, path string, method string, fieldPaths ...[]string) {
	t.Helper()

	schema := requestBodySchema(t, spec, path, method)
	for _, fieldPath := range fieldPaths {
		assertSchemaField(t, spec, schema, fieldPath)
	}
}

func assertResponseBodyFields(t *testing.T, spec map[string]any, path string, method string, status string, fieldPaths ...[]string) {
	t.Helper()

	schema := responseBodySchema(t, spec, path, method, status)
	for _, fieldPath := range fieldPaths {
		assertSchemaField(t, spec, schema, fieldPath)
	}
}

func requestBodySchema(t *testing.T, spec map[string]any, path string, method string) map[string]any {
	t.Helper()

	paths := mapField(t, spec, "paths")
	pathItem := mapField(t, paths, path)
	operation := mapField(t, pathItem, strings.ToLower(method))
	requestBody := mapField(t, operation, "requestBody")
	content := mapField(t, requestBody, "content")
	jsonMedia := mapField(t, content, "application/json")
	return resolveSchema(t, spec, mapField(t, jsonMedia, "schema"))
}

func responseBodySchema(t *testing.T, spec map[string]any, path string, method string, status string) map[string]any {
	t.Helper()

	paths := mapField(t, spec, "paths")
	pathItem := mapField(t, paths, path)
	operation := mapField(t, pathItem, strings.ToLower(method))
	responses := mapField(t, operation, "responses")
	response := mapField(t, responses, status)
	content := mapField(t, response, "content")
	jsonMedia := mapField(t, content, "application/json")
	return resolveSchema(t, spec, mapField(t, jsonMedia, "schema"))
}

func assertSchemaField(t *testing.T, spec map[string]any, schema map[string]any, fieldPath []string) {
	t.Helper()

	_ = schemaAtPath(t, spec, schema, fieldPath)
}

func assertDiscriminatedEngineSchema(t *testing.T, spec map[string]any, root map[string]any, fieldPath []string) {
	t.Helper()

	engine := resolveSchema(t, spec, schemaAtPath(t, spec, root, fieldPath))
	discriminator := mapField(t, engine, "discriminator")
	if discriminator["propertyName"] != "type" {
		t.Fatalf("expected engine discriminator propertyName type, got %#v", discriminator)
	}
	oneOf, ok := engine["oneOf"].([]any)
	if !ok || len(oneOf) != 2 {
		t.Fatalf("expected engine oneOf variants, got %#v", engine["oneOf"])
	}
	assertOneOfVariant(t, oneOf, "lua", []string{"type", "script"}, []string{"script"})
	assertOneOfVariant(t, oneOf, "ai", []string{"type", "prompt", "provider_id"}, []string{"prompt", "provider_id"})
}

func assertOpenAPIUsesResourceConvention(t *testing.T, spec map[string]any) {
	t.Helper()

	paths := mapField(t, spec, "paths")
	for path, pathValue := range paths {
		pathItem, ok := pathValue.(map[string]any)
		if !ok {
			continue
		}
		for method, operationValue := range pathItem {
			operation, ok := operationValue.(map[string]any)
			if !ok {
				continue
			}
			if requestBody, ok := operation["requestBody"].(map[string]any); ok {
				content, ok := requestBody["content"].(map[string]any)
				if ok {
					if media, ok := content["application/json"].(map[string]any); ok {
						schema := resolveSchema(t, spec, mapField(t, media, "schema"))
						assertSchemaField(t, spec, schema, []string{"spec"})
					}
				}
			}
			responses, ok := operation["responses"].(map[string]any)
			if !ok {
				continue
			}
			for status, responseValue := range responses {
				if !strings.HasPrefix(status, "2") {
					continue
				}
				response, ok := responseValue.(map[string]any)
				if !ok {
					continue
				}
				content, ok := response["content"].(map[string]any)
				if !ok {
					continue
				}
				media, ok := content["application/json"].(map[string]any)
				if !ok {
					continue
				}
				schema := resolveSchema(t, spec, mapField(t, media, "schema"))
				if !schemaHasField(t, spec, schema, []string{"metadata"}) || !schemaHasField(t, spec, schema, []string{"spec"}) || !schemaHasField(t, spec, schema, []string{"status"}) {
					t.Fatalf("expected %s %s response %s to use metadata/spec/status resource convention, got %#v", strings.ToUpper(method), path, status, schema)
				}
				if schemaHasField(t, spec, schema, []string{"status", "items"}) {
					assertSchemaField(t, spec, schema, []string{"status", "items", "metadata"})
					assertSchemaField(t, spec, schema, []string{"status", "items", "spec"})
					assertSchemaField(t, spec, schema, []string{"status", "items", "status"})
				}
			}
		}
	}
}

func schemaHasField(t *testing.T, spec map[string]any, schema map[string]any, fieldPath []string) bool {
	t.Helper()

	defer func() {
		_ = recover()
	}()
	return schemaAtPathOK(t, spec, schema, fieldPath)
}

func schemaAtPathOK(t *testing.T, spec map[string]any, schema map[string]any, fieldPath []string) bool {
	t.Helper()

	current := schema
	for _, field := range fieldPath {
		current = resolveSchema(t, spec, current)
		current = schemaArrayItem(t, spec, current)
		properties, ok := current["properties"].(map[string]any)
		if !ok {
			return false
		}
		next, ok := properties[field].(map[string]any)
		if !ok {
			return false
		}
		current = next
	}
	return true
}

func assertOneOfVariant(t *testing.T, variants []any, engineType string, required []string, properties []string) {
	t.Helper()

	for _, variant := range variants {
		schema, ok := variant.(map[string]any)
		if !ok {
			continue
		}
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			continue
		}
		typeSchema, ok := props["type"].(map[string]any)
		if !ok || !jsonArrayContains(typeSchema["enum"], engineType) {
			continue
		}
		for _, field := range required {
			if !jsonArrayContains(schema["required"], field) {
				t.Fatalf("expected %s engine required field %q in %#v", engineType, field, schema["required"])
			}
		}
		for _, field := range properties {
			if _, ok := props[field]; !ok {
				t.Fatalf("expected %s engine property %q in %#v", engineType, field, props)
			}
		}
		return
	}
	t.Fatalf("expected engine oneOf variant %q in %#v", engineType, variants)
}

func schemaAtPath(t *testing.T, spec map[string]any, schema map[string]any, fieldPath []string) map[string]any {
	t.Helper()

	current := schema
	for _, field := range fieldPath {
		current = resolveSchema(t, spec, current)
		current = schemaArrayItem(t, spec, current)
		properties := mapField(t, current, "properties")
		next, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("expected OpenAPI schema field %s in %#v", strings.Join(fieldPath, "."), properties)
		}
		current = next
	}
	return current
}

func schemaArrayItem(t *testing.T, spec map[string]any, schema map[string]any) map[string]any {
	t.Helper()

	typ := schema["type"]
	isArray := typ == "array" || jsonArrayContains(typ, "array")
	if !isArray {
		return schema
	}
	items, ok := schema["items"].(map[string]any)
	if !ok {
		t.Fatalf("expected array items schema in %#v", schema)
	}
	return resolveSchema(t, spec, items)
}

func resolveSchema(t *testing.T, spec map[string]any, schema map[string]any) map[string]any {
	t.Helper()

	for {
		ref, ok := schema["$ref"].(string)
		if !ok || ref == "" {
			return schema
		}
		const prefix = "#/components/schemas/"
		if !strings.HasPrefix(ref, prefix) {
			t.Fatalf("unsupported OpenAPI schema reference %q", ref)
		}
		schemas := mapField(t, mapField(t, spec, "components"), "schemas")
		resolved, ok := schemas[strings.TrimPrefix(ref, prefix)].(map[string]any)
		if !ok {
			t.Fatalf("missing OpenAPI schema reference %q", ref)
		}
		schema = resolved
	}
}

func mapField(t *testing.T, value map[string]any, field string) map[string]any {
	t.Helper()

	next, ok := value[field].(map[string]any)
	if !ok {
		t.Fatalf("expected object field %q in %#v", field, value)
	}
	return next
}

func jsonArrayContains(value any, expected string) bool {
	items, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}
