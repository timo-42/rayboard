package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

func TestNotificationInboxEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	notificationService := notifications.NewService(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notificationService),
	)

	first, err := notificationService.Create(ctx, notifications.CreateInput{
		UserID:      bootstrap.UserID,
		Type:        "comment_added",
		SubjectType: "ticket",
		SubjectID:   "ticket_1",
		Body:        "A comment was added",
		Data:        map[string]any{"ticket_key": "CORE-1"},
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	if _, err := notificationService.Create(ctx, notifications.CreateInput{
		UserID: bootstrap.UserID,
		Type:   "ticket_updated",
		Body:   "A ticket changed",
	}); err != nil {
		t.Fatalf("create notification: %v", err)
	}

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	listReq := httptest.NewRequest(http.MethodGet, "/api/notifications?unread=true&limit=20", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list notification status 200, got %d: %s", list.Code, list.Body.String())
	}
	listed := decodeNotificationList(t, list.Body.Bytes())
	if listed.Metadata.Count != 2 || listed.Status.Items[0].Spec.Body == "" {
		t.Fatalf("unexpected notification list: %#v", listed)
	}

	readReq := httptest.NewRequest(http.MethodPost, "/api/notifications/"+first.ID+"/read", nil)
	addSessionCSRF(readReq, session, csrf)
	read := httptest.NewRecorder()
	handler.ServeHTTP(read, readReq)
	if read.Code != http.StatusOK {
		t.Fatalf("expected mark read status 200, got %d: %s", read.Code, read.Body.String())
	}
	readResource := decodeNotificationResource(t, read.Body.Bytes())
	if readResource.Status.ReadAt == nil {
		t.Fatalf("expected read_at after mark read: %#v", readResource)
	}

	unreadReq := httptest.NewRequest(http.MethodPost, "/api/notifications/"+first.ID+"/unread", nil)
	addSessionCSRF(unreadReq, session, csrf)
	unread := httptest.NewRecorder()
	handler.ServeHTTP(unread, unreadReq)
	if unread.Code != http.StatusOK {
		t.Fatalf("expected mark unread status 200, got %d: %s", unread.Code, unread.Body.String())
	}
	unreadResource := decodeNotificationResource(t, unread.Body.Bytes())
	if unreadResource.Status.ReadAt != nil {
		t.Fatalf("expected nil read_at after mark unread: %#v", unreadResource)
	}

	missingCSRF := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
	missingCSRF.AddCookie(session)
	missingCSRFRec := httptest.NewRecorder()
	handler.ServeHTTP(missingCSRFRec, missingCSRF)
	if missingCSRFRec.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF read-all status 403, got %d: %s", missingCSRFRec.Code, missingCSRFRec.Body.String())
	}

	readAllReq := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
	addSessionCSRF(readAllReq, session, csrf)
	readAll := httptest.NewRecorder()
	handler.ServeHTTP(readAll, readAllReq)
	if readAll.Code != http.StatusNoContent {
		t.Fatalf("expected read-all status 204, got %d: %s", readAll.Code, readAll.Body.String())
	}
}

type notificationListBody struct {
	Metadata struct {
		Count int `json:"count"`
	} `json:"metadata"`
	Status struct {
		Items []notificationResourceBody `json:"items"`
	} `json:"status"`
}

type notificationResourceBody struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Type string `json:"type"`
		Body string `json:"body"`
	} `json:"spec"`
	Status struct {
		ReadAt *string `json:"read_at"`
	} `json:"status"`
}

func decodeNotificationList(t *testing.T, data []byte) notificationListBody {
	t.Helper()
	var body notificationListBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification list: %v", err)
	}
	return body
}

func decodeNotificationResource(t *testing.T, data []byte) notificationResourceBody {
	t.Helper()
	var body notificationResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification resource: %v", err)
	}
	return body
}
