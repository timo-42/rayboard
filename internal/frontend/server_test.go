package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Rayboard") ||
		!strings.Contains(body, "http://backend.test") ||
		!strings.Contains(body, "/static/app.js") ||
		!strings.Contains(body, `href="/docs"`) ||
		!strings.Contains(body, `href="/api/docs"`) ||
		!strings.Contains(body, `href="/api/docs/redoc"`) ||
		!strings.Contains(body, `href="/1"`) ||
		!strings.Contains(body, `href="/5"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestDesignVariantRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/3", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `href="/3" aria-current="page"`) ||
		!strings.Contains(body, `href="/1"`) ||
		!strings.Contains(body, `href="/5"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestAPIProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Fatalf("unexpected proxied path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT proxy, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(backend.Close)

	req := httptest.NewRequest(http.MethodPut, "/api/health", nil)
	rec := httptest.NewRecorder()

	NewHandler(backend.URL).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("unexpected proxy body: %s", body)
	}
}

func TestDocsHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected HTML docs content type, got %q", contentType)
	}
	body := rec.Body.String()
	for _, expected := range []string{"Rayboard Docs", "Rayboard Documentation", "/docs/api", "/docs/auth-rbac"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected docs page to contain %q", expected)
		}
	}
}

func TestDocsNamedPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs/api", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "API Guide") || !strings.Contains(body, "/api/openapi.json") {
		t.Fatalf("unexpected docs body: %s", body)
	}
}
