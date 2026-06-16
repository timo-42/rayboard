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
	if body.OpenAPI == "" || body.Info.Title != "Rayboard API" {
		t.Fatalf("unexpected OpenAPI metadata: %#v", body)
	}
	for _, path := range []string{"/api/health", "/api/login", "/api/projects/{project_id}/tickets"} {
		if _, ok := body.Paths[path]; !ok {
			t.Fatalf("expected path %s in OpenAPI document", path)
		}
	}
	for _, scheme := range []string{"bearerToken", "sessionCookie", "csrfToken"} {
		if _, ok := body.Components.SecuritySchemes[scheme]; !ok {
			t.Fatalf("expected security scheme %s in OpenAPI document", scheme)
		}
	}
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
