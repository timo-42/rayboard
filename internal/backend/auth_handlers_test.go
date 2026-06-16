package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
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

	missingCSRF := postJSON(t, handler, "/api/tokens", map[string]any{"spec": map[string]string{"name": "demo"}}, []*http.Cookie{sessionCookie})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/tokens", mustJSON(t, map[string]any{"spec": map[string]string{"name": "demo"}}))
	createReq.AddCookie(sessionCookie)
	createReq.AddCookie(csrfCookie)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create token status 201, got %d: %s", create.Code, create.Body.String())
	}

	var created struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Name string `json:"name"`
		} `json:"spec"`
		Status struct {
			Token string `json:"token"`
		} `json:"status"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created token: %v", err)
	}
	if created.Metadata.ID == "" || created.Status.Token == "" || created.Spec.Name != "demo" {
		t.Fatalf("unexpected created token: %#v", created)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+created.Status.Token)
	me := httptest.NewRecorder()
	handler.ServeHTTP(me, meReq)
	if me.Code != http.StatusOK {
		t.Fatalf("expected bearer me status 200, got %d: %s", me.Code, me.Body.String())
	}

	revokeReq := httptest.NewRequest(http.MethodDelete, "/api/tokens/"+created.Metadata.ID, nil)
	revokeReq.Header.Set("Authorization", "Bearer "+created.Status.Token)
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

func TestUserAdminEndpointsRequireRBAC(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
	)

	adminLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	adminSession := responseCookie(t, adminLogin.Result(), auth.SessionCookieName)
	adminCSRF := responseCookie(t, adminLogin.Result(), csrfCookieName)

	createReq := httptest.NewRequest(http.MethodPost, "/api/users", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"username":     "demo-user",
			"display_name": "Demo User",
		},
	}))
	addSessionCSRF(createReq, adminSession, adminCSRF)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create user status 201, got %d: %s", create.Code, create.Body.String())
	}

	var created struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Username string `json:"username"`
		} `json:"spec"`
		Status struct {
			Password string `json:"password"`
		} `json:"status"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created user: %v", err)
	}
	if created.Metadata.ID == "" || created.Status.Password == "" || created.Spec.Username != "demo-user" {
		t.Fatalf("unexpected created user: %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	listReq.AddCookie(adminSession)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list users status 200, got %d: %s", list.Code, list.Body.String())
	}

	userLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": created.Spec.Username,
		"password": created.Status.Password,
	}, nil)
	userSession := responseCookie(t, userLogin.Result(), auth.SessionCookieName)
	deniedReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	deniedReq.AddCookie(userSession)
	denied := httptest.NewRecorder()
	handler.ServeHTTP(denied, deniedReq)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("expected list users without permission status 403, got %d: %s", denied.Code, denied.Body.String())
	}

	disableReq := httptest.NewRequest(http.MethodPatch, "/api/users/"+created.Metadata.ID, mustJSON(t, map[string]any{"spec": map[string]bool{"disabled": true}}))
	addSessionCSRF(disableReq, adminSession, adminCSRF)
	disable := httptest.NewRecorder()
	handler.ServeHTTP(disable, disableReq)
	if disable.Code != http.StatusOK {
		t.Fatalf("expected disable user status 200, got %d: %s", disable.Code, disable.Body.String())
	}

	disabledLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": created.Spec.Username,
		"password": created.Status.Password,
	}, nil)
	if disabledLogin.Code != http.StatusForbidden {
		t.Fatalf("expected disabled login status 403, got %d: %s", disabledLogin.Code, disabledLogin.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/users/"+created.Metadata.ID, nil)
	addSessionCSRF(deleteReq, adminSession, adminCSRF)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete user status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func TestGroupRoleBindingEndpointsAffectExistingSession(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
	)

	adminLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	adminSession := responseCookie(t, adminLogin.Result(), auth.SessionCookieName)
	adminCSRF := responseCookie(t, adminLogin.Result(), csrfCookieName)

	createUserReq := httptest.NewRequest(http.MethodPost, "/api/users", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"username": "delegate",
		},
	}))
	addSessionCSRF(createUserReq, adminSession, adminCSRF)
	createUser := httptest.NewRecorder()
	handler.ServeHTTP(createUser, createUserReq)
	if createUser.Code != http.StatusCreated {
		t.Fatalf("expected create user status 201, got %d: %s", createUser.Code, createUser.Body.String())
	}
	var createdUser struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Username string `json:"username"`
		} `json:"spec"`
		Status struct {
			Password string `json:"password"`
		} `json:"status"`
	}
	if err := json.Unmarshal(createUser.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("decode created user: %v", err)
	}

	userLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": createdUser.Spec.Username,
		"password": createdUser.Status.Password,
	}, nil)
	userSession := responseCookie(t, userLogin.Result(), auth.SessionCookieName)

	deniedReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	deniedReq.AddCookie(userSession)
	denied := httptest.NewRecorder()
	handler.ServeHTTP(denied, deniedReq)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("expected initial delegate status 403, got %d: %s", denied.Code, denied.Body.String())
	}

	createGroupReq := httptest.NewRequest(http.MethodPost, "/api/groups", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":         "delegates",
			"display_name": "Delegates",
		},
	}))
	addSessionCSRF(createGroupReq, adminSession, adminCSRF)
	createGroup := httptest.NewRecorder()
	handler.ServeHTTP(createGroup, createGroupReq)
	if createGroup.Code != http.StatusCreated {
		t.Fatalf("expected create group status 201, got %d: %s", createGroup.Code, createGroup.Body.String())
	}
	var group struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Name string `json:"name"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(createGroup.Body.Bytes(), &group); err != nil {
		t.Fatalf("decode group: %v", err)
	}

	addMemberReq := httptest.NewRequest(http.MethodPost, "/api/groups/"+group.Metadata.ID+"/members/"+createdUser.Metadata.ID, nil)
	addSessionCSRF(addMemberReq, adminSession, adminCSRF)
	addMember := httptest.NewRecorder()
	handler.ServeHTTP(addMember, addMemberReq)
	if addMember.Code != http.StatusNoContent {
		t.Fatalf("expected add member status 204, got %d: %s", addMember.Code, addMember.Body.String())
	}

	createBindingReq := httptest.NewRequest(http.MethodPost, "/api/role-bindings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"role_name":    authz.RoleGlobalUserManager,
			"subject_type": authz.BindingTargetGroup,
			"subject_id":   group.Metadata.ID,
			"scope":        authz.ScopeKindGlobal,
		},
	}))
	addSessionCSRF(createBindingReq, adminSession, adminCSRF)
	createBinding := httptest.NewRecorder()
	handler.ServeHTTP(createBinding, createBindingReq)
	if createBinding.Code != http.StatusCreated {
		t.Fatalf("expected create binding status 201, got %d: %s", createBinding.Code, createBinding.Body.String())
	}

	allowedReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	allowedReq.AddCookie(userSession)
	allowed := httptest.NewRecorder()
	handler.ServeHTTP(allowed, allowedReq)
	if allowed.Code != http.StatusOK {
		t.Fatalf("expected existing session to gain access status 200, got %d: %s", allowed.Code, allowed.Body.String())
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

func addSessionCSRF(request *http.Request, session *http.Cookie, csrf *http.Cookie) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrf.Value)
	request.AddCookie(session)
	request.AddCookie(csrf)
}
