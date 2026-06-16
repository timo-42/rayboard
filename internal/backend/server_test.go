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
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" || body["service"] != "backend" {
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
	assertRequestBodyFields(t, spec, "/api/login", http.MethodPost, []string{"username"}, []string{"password"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/tickets", http.MethodPost, []string{"title"}, []string{"labels"})
	assertRequestBodyFields(t, spec, "/api/cron-jobs", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "schedule"}, []string{"spec", "engine"})
	assertDiscriminatedEngineSchema(t, spec, requestBodySchema(t, spec, "/api/cron-jobs", http.MethodPost), []string{"spec", "engine"})
	assertResponseBodyFields(t, spec, "/api/cron-jobs/{job_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "next_run_at"})
	assertRequestBodyFields(t, spec, "/api/projects/{project_id}/sprints", http.MethodPost, []string{"spec"}, []string{"spec", "name"}, []string{"spec", "goal"}, []string{"spec", "start_date"}, []string{"spec", "end_date"})
	assertResponseBodyFields(t, spec, "/api/sprints/{sprint_id}", http.MethodGet, "200", []string{"metadata"}, []string{"spec"}, []string{"status"}, []string{"status", "state"})
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
		properties := mapField(t, current, "properties")
		next, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("expected OpenAPI schema field %s in %#v", strings.Join(fieldPath, "."), properties)
		}
		current = next
	}
	return current
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
