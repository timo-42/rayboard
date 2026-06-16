package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestAuthEndpointsLoginMeAndLogout(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	handler := NewHandler(WithAuthService(auth.NewService(db.SQL)))

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d: %s", login.Code, login.Body.String())
	}
	sessionCookie := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrfCookie := responseCookie(t, login.Result(), csrfCookieName)
	if !sessionCookie.HttpOnly {
		t.Fatal("expected session cookie to be HttpOnly")
	}
	if csrfCookie.HttpOnly {
		t.Fatal("expected CSRF cookie to be readable by the browser")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.AddCookie(sessionCookie)
	me := httptest.NewRecorder()
	handler.ServeHTTP(me, meReq)
	if me.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d: %s", me.Code, me.Body.String())
	}

	logoutMissingCSRF := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	logoutMissingCSRF.AddCookie(sessionCookie)
	missingCSRF := httptest.NewRecorder()
	handler.ServeHTTP(missingCSRF, logoutMissingCSRF)
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	logout := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	logout.AddCookie(sessionCookie)
	logout.AddCookie(csrfCookie)
	logout.Header.Set("X-CSRF-Token", csrfCookie.Value)
	loggedOut := httptest.NewRecorder()
	handler.ServeHTTP(loggedOut, logout)
	if loggedOut.Code != http.StatusNoContent {
		t.Fatalf("expected logout status 204, got %d: %s", loggedOut.Code, loggedOut.Body.String())
	}
}

func TestAuthEndpointsAPITokensAndBearerAuth(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	handler := NewHandler(WithAuthService(auth.NewService(db.SQL)))

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	sessionCookie := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrfCookie := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/tokens", map[string]string{"name": "demo"}, []*http.Cookie{sessionCookie})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/tokens", mustJSON(t, map[string]string{"name": "demo"}))
	createReq.AddCookie(sessionCookie)
	createReq.AddCookie(csrfCookie)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create token status 201, got %d: %s", create.Code, create.Body.String())
	}

	var created auth.CreatedAPIToken
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created token: %v", err)
	}
	if created.ID == "" || created.Token == "" || created.Name != "demo" {
		t.Fatalf("unexpected created token: %#v", created)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+created.Token)
	me := httptest.NewRecorder()
	handler.ServeHTTP(me, meReq)
	if me.Code != http.StatusOK {
		t.Fatalf("expected bearer me status 200, got %d: %s", me.Code, me.Body.String())
	}

	revokeReq := httptest.NewRequest(http.MethodDelete, "/api/tokens/"+created.ID, nil)
	revokeReq.Header.Set("Authorization", "Bearer "+created.Token)
	revoke := httptest.NewRecorder()
	handler.ServeHTTP(revoke, revokeReq)
	if revoke.Code != http.StatusNoContent {
		t.Fatalf("expected revoke status 204, got %d: %s", revoke.Code, revoke.Body.String())
	}

	retry := httptest.NewRecorder()
	handler.ServeHTTP(retry, meReq)
	if retry.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked bearer status 401, got %d: %s", retry.Code, retry.Body.String())
	}
}

func openBackendTestDB(t *testing.T, ctx context.Context) (*store.DB, auth.BootstrapAdminResult) {
	t.Helper()

	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	bootstrap, err := auth.BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap admin: %v", err)
	}
	return db, bootstrap
}

func postJSON(t *testing.T, handler http.Handler, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, path, mustJSON(t, body))
	request.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func mustJSON(t *testing.T, body any) *bytes.Reader {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return bytes.NewReader(data)
}

func responseCookie(t *testing.T, response *http.Response, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("missing response cookie %q", name)
	return nil
}
