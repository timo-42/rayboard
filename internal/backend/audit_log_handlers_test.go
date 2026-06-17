package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
)

func TestAuditLogEndpoint(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	auditStore := audit.NewStore(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuditStore(auditStore),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
	)

	if _, err := auditStore.Record(ctx, audit.RecordInput{
		EventType:   "settings.updated",
		ActorID:     bootstrap.UserID,
		AuthKind:    authz.AuthKindSession,
		SubjectType: "settings",
		SubjectID:   "global",
		Payload:     map[string]any{"changed_fields": []string{"system_health_note"}},
	}); err != nil {
		t.Fatalf("record audit entry: %v", err)
	}

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodGet, "/api/audit-log?event_type=settings.updated&limit=10", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected audit log status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := decodeAuditLogList(t, rec.Body.Bytes())
	if body.Metadata.Count != 1 {
		t.Fatalf("expected one filtered audit entry, got %#v", body)
	}
	item := body.Status.Items[0]
	if item.Metadata.ID == "" || item.Spec.EventType != "settings.updated" || item.Spec.SubjectType != "settings" {
		t.Fatalf("unexpected audit entry resource: %#v", item)
	}
	if item.Spec.Payload["changed_fields"] == nil || !item.Status.SecurityEvent {
		t.Fatalf("unexpected audit entry payload/status: %#v", item)
	}
}

func TestAuditLogEndpointRequiresGlobalSettingsManage(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	user, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "audit-viewer"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	handler := NewHandler(
		WithAuthService(authService),
		WithAuditStore(audit.NewStore(db.SQL)),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
	)
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodGet, "/api/audit-log", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden audit log get, got %d: %s", rec.Code, rec.Body.String())
	}
}

type auditLogListBody struct {
	Metadata struct {
		Count int `json:"count"`
	} `json:"metadata"`
	Status struct {
		Items []struct {
			Metadata struct {
				ID string `json:"id"`
			} `json:"metadata"`
			Spec struct {
				EventType   string         `json:"event_type"`
				ActorID     string         `json:"actor_user_id"`
				SubjectType string         `json:"subject_type"`
				SubjectID   string         `json:"subject_id"`
				Outcome     string         `json:"outcome"`
				Payload     map[string]any `json:"payload"`
			} `json:"spec"`
			Status struct {
				SecurityEvent bool `json:"security_event"`
			} `json:"status"`
		} `json:"items"`
	} `json:"status"`
}

func decodeAuditLogList(t *testing.T, data []byte) auditLogListBody {
	t.Helper()
	var body auditLogListBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode audit log list: %v", err)
	}
	return body
}
