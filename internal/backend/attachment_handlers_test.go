package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestAttachmentEndpointsLifecycle(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithTrackerService(tracker.NewService(db.SQL, authorizer)),
		WithAttachmentService(attachments.NewService(db.SQL, authorizer)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	project := createAttachmentTestProject(t, handler, session, csrf)
	ticket := createAttachmentTestTicket(t, handler, session, csrf, project.ID)

	uploadReq := multipartUploadRequest(t, "/api/tickets/"+ticket.ID+"/attachments", "file", "notes.txt", "text/plain", []byte("hello attachment"))
	uploadReq.AddCookie(session)
	uploadReq.AddCookie(csrf)
	uploadReq.Header.Set("X-CSRF-Token", csrf.Value)
	upload := httptest.NewRecorder()
	handler.ServeHTTP(upload, uploadReq)
	if upload.Code != http.StatusCreated {
		t.Fatalf("expected upload status 201, got %d: %s", upload.Code, upload.Body.String())
	}
	var meta attachments.Metadata
	if err := json.Unmarshal(upload.Body.Bytes(), &meta); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if meta.ID == "" || meta.FileName != "notes.txt" || meta.SizeBytes != int64(len("hello attachment")) {
		t.Fatalf("unexpected metadata: %#v", meta)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/tickets/"+ticket.ID+"/attachments", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "notes.txt") {
		t.Fatalf("unexpected list response %d: %s", list.Code, list.Body.String())
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/attachments/"+meta.ID+"/download", nil)
	downloadReq.AddCookie(session)
	download := httptest.NewRecorder()
	handler.ServeHTTP(download, downloadReq)
	if download.Code != http.StatusOK || download.Body.String() != "hello attachment" {
		t.Fatalf("unexpected download response %d: %s", download.Code, download.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/attachments/"+meta.ID, nil)
	deleteReq.AddCookie(session)
	deleteReq.AddCookie(csrf)
	deleteReq.Header.Set("X-CSRF-Token", csrf.Value)
	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, deleteReq)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d: %s", deleted.Code, deleted.Body.String())
	}
}

func createAttachmentTestProject(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie) tracker.Project {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"key":  "ATT",
			"name": "Attachments",
		},
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create project status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	return decodeProjectResourceAsTracker(t, rec.Body.Bytes())
}

func createAttachmentTestTicket(t *testing.T, handler http.Handler, session *http.Cookie, csrf *http.Cookie, projectID string) tracker.Ticket {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/tickets", mustJSON(t, map[string]any{
		"title": "Attachment target",
	}))
	addSessionCSRF(req, session, csrf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create ticket status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var ticket tracker.Ticket
	if err := json.Unmarshal(rec.Body.Bytes(), &ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}
	return ticket
}

func multipartUploadRequest(t *testing.T, path string, field string, fileName string, contentType string, data []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if contentType != "" {
		req.Header.Set("X-Test-Content-Type", contentType)
	}
	return req
}
