package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/audit"
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

func TestAuthEndpointMalformedJSONUsesErrorEnvelope(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
	handler := NewHandler(WithAuthService(auth.NewService(db.SQL)))

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"spec":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected malformed login status 400/422, got %d: %s", rec.Code, rec.Body.String())
	}
	envelope := decodeAPIError(t, rec.Body.Bytes())
	if envelope.Error.Code != "validation_failed" || envelope.Error.Message == "" || len(envelope.Error.Fields) == 0 {
		t.Fatalf("unexpected malformed JSON error envelope: %#v", envelope)
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
	revokeReq.AddCookie(sessionCookie)
	revokeReq.AddCookie(csrfCookie)
	revoke := httptest.NewRecorder()
	handler.ServeHTTP(revoke, revokeReq)
	if revoke.Code != http.StatusNoContent {
		t.Fatalf("expected bearer revoke with session cookies and no CSRF status 204, got %d: %s", revoke.Code, revoke.Body.String())
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
	auditStore := audit.NewStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuditStore(auditStore),
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

	deniedEffectiveReq := httptest.NewRequest(http.MethodGet, "/api/users/"+createdUser.Metadata.ID+"/effective-permissions", nil)
	deniedEffectiveReq.AddCookie(userSession)
	deniedEffective := httptest.NewRecorder()
	handler.ServeHTTP(deniedEffective, deniedEffectiveReq)
	if deniedEffective.Code != http.StatusForbidden {
		t.Fatalf("expected user effective-permissions inspection status 403, got %d: %s", deniedEffective.Code, deniedEffective.Body.String())
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
	var binding struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(createBinding.Body.Bytes(), &binding); err != nil {
		t.Fatalf("decode binding: %v", err)
	}

	createProjectBindingReq := httptest.NewRequest(http.MethodPost, "/api/role-bindings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"role_name":    authz.RoleProjectMember,
			"subject_type": authz.BindingTargetGroup,
			"subject_id":   group.Metadata.ID,
			"scope":        authz.ScopeKindProject,
			"project_id":   "project-alpha",
		},
	}))
	addSessionCSRF(createProjectBindingReq, adminSession, adminCSRF)
	createProjectBinding := httptest.NewRecorder()
	handler.ServeHTTP(createProjectBinding, createProjectBindingReq)
	if createProjectBinding.Code != http.StatusCreated {
		t.Fatalf("expected create project binding status 201, got %d: %s", createProjectBinding.Code, createProjectBinding.Body.String())
	}

	getRoleReq := httptest.NewRequest(http.MethodGet, "/api/roles/"+string(authz.RoleProjectAdmin), nil)
	getRoleReq.AddCookie(adminSession)
	getRole := httptest.NewRecorder()
	handler.ServeHTTP(getRole, getRoleReq)
	if getRole.Code != http.StatusOK {
		t.Fatalf("expected get role status 200, got %d: %s", getRole.Code, getRole.Body.String())
	}
	var role struct {
		Spec struct {
			Name        authz.RoleName     `json:"name"`
			Permissions []authz.Permission `json:"permissions"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(getRole.Body.Bytes(), &role); err != nil {
		t.Fatalf("decode role: %v", err)
	}
	if role.Spec.Name != authz.RoleProjectAdmin || !slices.Contains(role.Spec.Permissions, authz.PermissionRolesBind) {
		t.Fatalf("unexpected role detail: %#v", role)
	}

	createProjectScopedBindingReq := httptest.NewRequest(http.MethodPost, "/api/projects/project-alpha/role-bindings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"role_name":    authz.RoleProjectViewer,
			"subject_type": authz.BindingTargetUser,
			"subject_id":   createdUser.Metadata.ID,
		},
	}))
	addSessionCSRF(createProjectScopedBindingReq, adminSession, adminCSRF)
	createProjectScopedBinding := httptest.NewRecorder()
	handler.ServeHTTP(createProjectScopedBinding, createProjectScopedBindingReq)
	if createProjectScopedBinding.Code != http.StatusCreated {
		t.Fatalf("expected create project-scoped binding status 201, got %d: %s", createProjectScopedBinding.Code, createProjectScopedBinding.Body.String())
	}

	createProjectAdminReq := httptest.NewRequest(http.MethodPost, "/api/users", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"username": "project-admin",
		},
	}))
	addSessionCSRF(createProjectAdminReq, adminSession, adminCSRF)
	createProjectAdmin := httptest.NewRecorder()
	handler.ServeHTTP(createProjectAdmin, createProjectAdminReq)
	if createProjectAdmin.Code != http.StatusCreated {
		t.Fatalf("expected create project admin status 201, got %d: %s", createProjectAdmin.Code, createProjectAdmin.Body.String())
	}
	var projectAdminUser struct {
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
	if err := json.Unmarshal(createProjectAdmin.Body.Bytes(), &projectAdminUser); err != nil {
		t.Fatalf("decode project admin user: %v", err)
	}
	createProjectAdminBindingReq := httptest.NewRequest(http.MethodPost, "/api/role-bindings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"role_name":    authz.RoleProjectAdmin,
			"subject_type": authz.BindingTargetUser,
			"subject_id":   projectAdminUser.Metadata.ID,
			"scope":        authz.ScopeKindProject,
			"project_id":   "project-alpha",
		},
	}))
	addSessionCSRF(createProjectAdminBindingReq, adminSession, adminCSRF)
	createProjectAdminBinding := httptest.NewRecorder()
	handler.ServeHTTP(createProjectAdminBinding, createProjectAdminBindingReq)
	if createProjectAdminBinding.Code != http.StatusCreated {
		t.Fatalf("expected project admin binding status 201, got %d: %s", createProjectAdminBinding.Code, createProjectAdminBinding.Body.String())
	}
	projectAdminLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": projectAdminUser.Spec.Username,
		"password": projectAdminUser.Status.Password,
	}, nil)
	projectAdminSession := responseCookie(t, projectAdminLogin.Result(), auth.SessionCookieName)
	projectAdminCSRF := responseCookie(t, projectAdminLogin.Result(), csrfCookieName)

	projectAdminCreateBindingReq := httptest.NewRequest(http.MethodPost, "/api/projects/project-alpha/role-bindings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"role_name":    authz.RoleProjectViewer,
			"subject_type": authz.BindingTargetUser,
			"subject_id":   createdUser.Metadata.ID,
		},
	}))
	addSessionCSRF(projectAdminCreateBindingReq, projectAdminSession, projectAdminCSRF)
	projectAdminCreateBinding := httptest.NewRecorder()
	handler.ServeHTTP(projectAdminCreateBinding, projectAdminCreateBindingReq)
	if projectAdminCreateBinding.Code != http.StatusCreated {
		t.Fatalf("expected project admin create project binding status 201, got %d: %s", projectAdminCreateBinding.Code, projectAdminCreateBinding.Body.String())
	}
	var projectAdminBinding struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(projectAdminCreateBinding.Body.Bytes(), &projectAdminBinding); err != nil {
		t.Fatalf("decode project admin binding: %v", err)
	}

	projectAdminOwnerReq := httptest.NewRequest(http.MethodPost, "/api/projects/project-alpha/role-bindings", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"role_name":    authz.RoleProjectOwner,
			"subject_type": authz.BindingTargetUser,
			"subject_id":   createdUser.Metadata.ID,
		},
	}))
	addSessionCSRF(projectAdminOwnerReq, projectAdminSession, projectAdminCSRF)
	projectAdminOwner := httptest.NewRecorder()
	handler.ServeHTTP(projectAdminOwner, projectAdminOwnerReq)
	if projectAdminOwner.Code != http.StatusBadRequest {
		t.Fatalf("expected project owner binding through project API status 400, got %d: %s", projectAdminOwner.Code, projectAdminOwner.Body.String())
	}

	projectAdminOtherProjectReq := httptest.NewRequest(http.MethodGet, "/api/projects/project-beta/role-bindings", nil)
	projectAdminOtherProjectReq.AddCookie(projectAdminSession)
	projectAdminOtherProject := httptest.NewRecorder()
	handler.ServeHTTP(projectAdminOtherProject, projectAdminOtherProjectReq)
	if projectAdminOtherProject.Code != http.StatusForbidden {
		t.Fatalf("expected project admin cross-project list status 403, got %d: %s", projectAdminOtherProject.Code, projectAdminOtherProject.Body.String())
	}

	projectAdminDeleteBindingReq := httptest.NewRequest(http.MethodDelete, "/api/projects/project-alpha/role-bindings/"+projectAdminBinding.Metadata.ID, nil)
	addSessionCSRF(projectAdminDeleteBindingReq, projectAdminSession, projectAdminCSRF)
	projectAdminDeleteBinding := httptest.NewRecorder()
	handler.ServeHTTP(projectAdminDeleteBinding, projectAdminDeleteBindingReq)
	if projectAdminDeleteBinding.Code != http.StatusNoContent {
		t.Fatalf("expected project admin delete project binding status 204, got %d: %s", projectAdminDeleteBinding.Code, projectAdminDeleteBinding.Body.String())
	}

	listProjectBindingsReq := httptest.NewRequest(http.MethodGet, "/api/projects/project-alpha/role-bindings", nil)
	listProjectBindingsReq.AddCookie(adminSession)
	listProjectBindings := httptest.NewRecorder()
	handler.ServeHTTP(listProjectBindings, listProjectBindingsReq)
	if listProjectBindings.Code != http.StatusOK {
		t.Fatalf("expected list project role bindings status 200, got %d: %s", listProjectBindings.Code, listProjectBindings.Body.String())
	}
	var projectBindings struct {
		Status struct {
			Items []struct {
				Spec struct {
					ProjectID string `json:"project_id"`
				} `json:"spec"`
			} `json:"items"`
		} `json:"status"`
	}
	if err := json.Unmarshal(listProjectBindings.Body.Bytes(), &projectBindings); err != nil {
		t.Fatalf("decode project bindings: %v", err)
	}
	if len(projectBindings.Status.Items) < 2 {
		t.Fatalf("expected project role bindings, got %#v", projectBindings)
	}
	for _, item := range projectBindings.Status.Items {
		if item.Spec.ProjectID != "project-alpha" {
			t.Fatalf("expected project-scoped binding, got %#v", item)
		}
	}

	listProjectMembersReq := httptest.NewRequest(http.MethodGet, "/api/projects/project-alpha/members", nil)
	listProjectMembersReq.AddCookie(adminSession)
	listProjectMembers := httptest.NewRecorder()
	handler.ServeHTTP(listProjectMembers, listProjectMembersReq)
	if listProjectMembers.Code != http.StatusOK {
		t.Fatalf("expected list project members status 200, got %d: %s", listProjectMembers.Code, listProjectMembers.Body.String())
	}
	var projectMembers struct {
		Status struct {
			Items []struct {
				Metadata struct {
					ID string `json:"id"`
				} `json:"metadata"`
			} `json:"items"`
		} `json:"status"`
	}
	if err := json.Unmarshal(listProjectMembers.Body.Bytes(), &projectMembers); err != nil {
		t.Fatalf("decode project members: %v", err)
	}
	memberIDs := map[string]bool{}
	for _, item := range projectMembers.Status.Items {
		memberIDs[item.Metadata.ID] = true
	}
	if !memberIDs[createdUser.Metadata.ID] || !memberIDs[projectAdminUser.Metadata.ID] {
		t.Fatalf("unexpected project members: %#v", projectMembers)
	}

	allowedReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	allowedReq.AddCookie(userSession)
	allowed := httptest.NewRecorder()
	handler.ServeHTTP(allowed, allowedReq)
	if allowed.Code != http.StatusOK {
		t.Fatalf("expected existing session to gain access status 200, got %d: %s", allowed.Code, allowed.Body.String())
	}

	selfEffectiveReq := httptest.NewRequest(http.MethodGet, "/api/me/effective-permissions", nil)
	selfEffectiveReq.AddCookie(userSession)
	selfEffective := httptest.NewRecorder()
	handler.ServeHTTP(selfEffective, selfEffectiveReq)
	if selfEffective.Code != http.StatusOK {
		t.Fatalf("expected self effective permissions status 200, got %d: %s", selfEffective.Code, selfEffective.Body.String())
	}
	selfPermissions := decodeEffectivePermissions(t, selfEffective.Body.Bytes())
	if selfPermissions.Metadata.UserID != createdUser.Metadata.ID || selfPermissions.Spec.Scope != authz.ScopeKindGlobal || !slices.Contains(selfPermissions.Status.Permissions, authz.PermissionRolesRead) {
		t.Fatalf("unexpected self effective permissions: %#v", selfPermissions)
	}

	projectEffectiveReq := httptest.NewRequest(http.MethodGet, "/api/me/effective-permissions?scope=project&project_id=project-alpha", nil)
	projectEffectiveReq.AddCookie(userSession)
	projectEffective := httptest.NewRecorder()
	handler.ServeHTTP(projectEffective, projectEffectiveReq)
	if projectEffective.Code != http.StatusOK {
		t.Fatalf("expected project effective permissions status 200, got %d: %s", projectEffective.Code, projectEffective.Body.String())
	}
	projectPermissions := decodeEffectivePermissions(t, projectEffective.Body.Bytes())
	if projectPermissions.Spec.ProjectID != "project-alpha" || !slices.Contains(projectPermissions.Status.Permissions, authz.PermissionTicketsWrite) {
		t.Fatalf("unexpected project effective permissions: %#v", projectPermissions)
	}

	adminEffectiveReq := httptest.NewRequest(http.MethodGet, "/api/users/"+createdUser.Metadata.ID+"/effective-permissions?scope=project&project_id=project-alpha", nil)
	adminEffectiveReq.AddCookie(adminSession)
	adminEffective := httptest.NewRecorder()
	handler.ServeHTTP(adminEffective, adminEffectiveReq)
	if adminEffective.Code != http.StatusOK {
		t.Fatalf("expected admin effective-permissions inspection status 200, got %d: %s", adminEffective.Code, adminEffective.Body.String())
	}
	adminPermissions := decodeEffectivePermissions(t, adminEffective.Body.Bytes())
	if adminPermissions.Metadata.UserID != createdUser.Metadata.ID || !slices.Contains(adminPermissions.Status.Permissions, authz.PermissionTicketsWrite) {
		t.Fatalf("unexpected admin inspected effective permissions: %#v", adminPermissions)
	}

	invalidScopeReq := httptest.NewRequest(http.MethodGet, "/api/me/effective-permissions?scope=project", nil)
	invalidScopeReq.AddCookie(userSession)
	invalidScope := httptest.NewRecorder()
	handler.ServeHTTP(invalidScope, invalidScopeReq)
	if invalidScope.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid scope status 400, got %d: %s", invalidScope.Code, invalidScope.Body.String())
	}

	invalidGlobalReq := httptest.NewRequest(http.MethodGet, "/api/me/effective-permissions?scope=global&project_id=project-alpha", nil)
	invalidGlobalReq.AddCookie(userSession)
	invalidGlobal := httptest.NewRecorder()
	handler.ServeHTTP(invalidGlobal, invalidGlobalReq)
	if invalidGlobal.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid global scope status 400, got %d: %s", invalidGlobal.Code, invalidGlobal.Body.String())
	}

	missingUserReq := httptest.NewRequest(http.MethodGet, "/api/users/user_missing/effective-permissions", nil)
	missingUserReq.AddCookie(adminSession)
	missingUser := httptest.NewRecorder()
	handler.ServeHTTP(missingUser, missingUserReq)
	if missingUser.Code != http.StatusNotFound {
		t.Fatalf("expected missing effective-permissions user status 404, got %d: %s", missingUser.Code, missingUser.Body.String())
	}

	deleteBindingReq := httptest.NewRequest(http.MethodDelete, "/api/role-bindings/"+binding.Metadata.ID, nil)
	addSessionCSRF(deleteBindingReq, adminSession, adminCSRF)
	deleteBinding := httptest.NewRecorder()
	handler.ServeHTTP(deleteBinding, deleteBindingReq)
	if deleteBinding.Code != http.StatusNoContent {
		t.Fatalf("expected delete binding status 204, got %d: %s", deleteBinding.Code, deleteBinding.Body.String())
	}

	removeMemberReq := httptest.NewRequest(http.MethodDelete, "/api/groups/"+group.Metadata.ID+"/members/"+createdUser.Metadata.ID, nil)
	addSessionCSRF(removeMemberReq, adminSession, adminCSRF)
	removeMember := httptest.NewRecorder()
	handler.ServeHTTP(removeMember, removeMemberReq)
	if removeMember.Code != http.StatusNoContent {
		t.Fatalf("expected remove member status 204, got %d: %s", removeMember.Code, removeMember.Body.String())
	}

	entries, err := auditStore.List(ctx, 50)
	if err != nil {
		t.Fatalf("list audit entries: %v", err)
	}
	events := auditEvents(entries)
	for _, eventType := range []string{"rbac.group_created", "rbac.group_member_added", "rbac.role_binding_created", "rbac.role_binding_deleted", "rbac.group_member_removed"} {
		if events[eventType] == nil {
			t.Fatalf("expected audit event %s in %#v", eventType, entries)
		}
	}
	if events["rbac.role_binding_created"].Payload["role_name"] != string(authz.RoleGlobalUserManager) {
		t.Fatalf("unexpected role binding payload: %#v", events["rbac.role_binding_created"].Payload)
	}
}

func TestAuthEndpointsWriteAuditLog(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	auditStore := audit.NewStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuditStore(auditStore),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
	)

	failedLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": "wrong",
	}, nil)
	if failedLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected failed login status 401, got %d: %s", failedLogin.Code, failedLogin.Body.String())
	}

	adminLogin := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	adminSession := responseCookie(t, adminLogin.Result(), auth.SessionCookieName)
	adminCSRF := responseCookie(t, adminLogin.Result(), csrfCookieName)

	createTokenReq := httptest.NewRequest(http.MethodPost, "/api/tokens", mustJSON(t, map[string]any{
		"spec": map[string]string{"name": "audit-token"},
	}))
	addSessionCSRF(createTokenReq, adminSession, adminCSRF)
	createToken := httptest.NewRecorder()
	handler.ServeHTTP(createToken, createTokenReq)
	if createToken.Code != http.StatusCreated {
		t.Fatalf("expected token create status 201, got %d: %s", createToken.Code, createToken.Body.String())
	}
	var token struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Status struct {
			Token string `json:"token"`
		} `json:"status"`
	}
	if err := json.Unmarshal(createToken.Body.Bytes(), &token); err != nil {
		t.Fatalf("decode token: %v", err)
	}

	createUserReq := httptest.NewRequest(http.MethodPost, "/api/users", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"username": "audit-user",
		},
	}))
	addSessionCSRF(createUserReq, adminSession, adminCSRF)
	createUser := httptest.NewRecorder()
	handler.ServeHTTP(createUser, createUserReq)
	if createUser.Code != http.StatusCreated {
		t.Fatalf("expected user create status 201, got %d: %s", createUser.Code, createUser.Body.String())
	}
	var createdUser struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(createUser.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("decode user: %v", err)
	}

	disableReq := httptest.NewRequest(http.MethodPatch, "/api/users/"+createdUser.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]bool{"disabled": true},
	}))
	addSessionCSRF(disableReq, adminSession, adminCSRF)
	disable := httptest.NewRecorder()
	handler.ServeHTTP(disable, disableReq)
	if disable.Code != http.StatusOK {
		t.Fatalf("expected disable user status 200, got %d: %s", disable.Code, disable.Body.String())
	}

	entries, err := auditStore.List(ctx, 20)
	if err != nil {
		t.Fatalf("list audit entries: %v", err)
	}
	events := auditEvents(entries)
	for _, eventType := range []string{"auth.login_failed", "auth.session_created", "auth.api_token_created", "user.created", "user.disabled"} {
		if events[eventType] == nil {
			t.Fatalf("expected audit event %s in %#v", eventType, entries)
		}
	}
	if events["auth.api_token_created"].Payload["token_name"] != "audit-token" {
		t.Fatalf("unexpected token audit payload: %#v", events["auth.api_token_created"].Payload)
	}
	if _, ok := events["auth.api_token_created"].Payload["token"]; ok {
		t.Fatalf("audit payload leaked API token: %#v", events["auth.api_token_created"].Payload)
	}
	if token.Status.Token == "" || bytes.Contains(mustJSONBytes(t, entries), []byte(token.Status.Token)) {
		t.Fatalf("audit entries leaked API token secret")
	}
	if events["user.disabled"].Payload["credentials_revoked"] != true {
		t.Fatalf("expected credential revocation payload, got %#v", events["user.disabled"].Payload)
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

	if path == "/api/login" {
		body = map[string]any{"spec": body}
	}
	request := httptest.NewRequest(http.MethodPost, path, mustJSON(t, body))
	request.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func auditEvents(entries []audit.Entry) map[string]*audit.Entry {
	events := make(map[string]*audit.Entry, len(entries))
	for index := range entries {
		entry := &entries[index]
		events[entry.EventType] = entry
	}
	return events
}

type effectivePermissionsBody struct {
	Metadata struct {
		UserID string `json:"user_id"`
	} `json:"metadata"`
	Spec struct {
		Scope     authz.ScopeKind `json:"scope"`
		ProjectID string          `json:"project_id"`
	} `json:"spec"`
	Status struct {
		Permissions []authz.Permission `json:"permissions"`
	} `json:"status"`
}

func decodeEffectivePermissions(t *testing.T, data []byte) effectivePermissionsBody {
	t.Helper()

	var body effectivePermissionsBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode effective permissions: %v", err)
	}
	return body
}

func mustJSONBytes(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
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
