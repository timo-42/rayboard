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
		!strings.Contains(body, `class="design-selector-home"`) ||
		!strings.Contains(body, "Design Selector") ||
		!strings.Contains(body, `href="/docs"`) ||
		!strings.Contains(body, `href="/api/docs"`) ||
		!strings.Contains(body, `href="/api/docs/redoc"`) ||
		!strings.Contains(body, "Engine Workbench") ||
		!strings.Contains(body, `id="engine-form"`) ||
		!strings.Contains(body, `id="notification-inbox"`) ||
		!strings.Contains(body, `id="search-panel"`) ||
		!strings.Contains(body, `id="ticket-columns"`) ||
		!strings.Contains(body, `href="/1"`) ||
		!strings.Contains(body, `href="/5"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestEmbeddedAppSupportsAttachments(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadAttachments",
		"normalizeAttachment",
		"/api/tickets/${ticketID}/attachments",
		"/api/attachments/${attachment.id}/download",
		"data-delete-attachment-id",
		"new FormData()",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-attachments",
		".attachment-item",
		".attachment-form",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsComments(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadComments",
		"normalizeComment",
		"/api/tickets/${ticketID}/comments",
		"/api/comments/${deleteComment.dataset.deleteCommentId}",
		"data-delete-comment-id",
		"data-comment-form",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-comments",
		".comment-item",
		".comment-form",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsNotifications(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadNotifications",
		"normalizeNotification",
		"/api/notifications${query}",
		"/api/notifications/read-all",
		"/api/notifications/${button.dataset.notificationId}/${action}",
		"data-notification-read-state",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".notification-inbox",
		".notification-item",
		".notification-item.is-unread",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsSearchSavedViews(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadSavedViews",
		"runSearch",
		"normalizeSavedView",
		"/api/search",
		"/api/saved-views",
		"data-apply-saved-view-id",
		"data-delete-saved-view-id",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".search-panel",
		".search-results",
		".saved-view-list",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestDesignVariantRoute(t *testing.T) {
	tests := []struct {
		path      string
		bodyClass string
		name      string
	}{
		{path: "/1", bodyClass: "design-operations", name: "Operations"},
		{path: "/2", bodyClass: "design-planning", name: "Planning"},
		{path: "/3", bodyClass: "design-automation", name: "Automation"},
		{path: "/4", bodyClass: "design-triage", name: "Triage"},
		{path: "/5", bodyClass: "design-executive", name: "Executive"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			NewHandler("http://backend.test").ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}
			body := rec.Body.String()
			for _, expected := range []string{
				`class="` + tt.bodyClass + `"`,
				tt.name,
				`href="` + tt.path + `" aria-current="page"`,
				`href="/1"`,
				`href="/5"`,
			} {
				if !strings.Contains(body, expected) {
					t.Fatalf("expected body to contain %q: %s", expected, body)
				}
			}
		})
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
