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
	if body := rec.Body.String(); !strings.Contains(body, "Rayboard") || !strings.Contains(body, "http://backend.test") {
		t.Fatalf("unexpected body: %s", body)
	}
}
