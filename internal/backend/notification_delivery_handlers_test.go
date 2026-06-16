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
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

func TestNotificationDeliveryEndpoints(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	notificationService := notifications.NewService(db.SQL)
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notificationService),
	)

	destination, err := notificationService.CreateDestination(ctx, notifications.CreateDestinationInput{
		Name:        "ops",
		ScopeType:   notifications.DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}
	policy, err := notificationService.CreatePolicy(ctx, notifications.CreatePolicyInput{
		Name:           "ops",
		ScopeType:      notifications.PolicyScopeGlobal,
		EventTypes:     []string{"ticket_assigned"},
		DestinationIDs: []string{destination.ID},
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	delivery, err := notificationService.EnqueueDelivery(ctx, notifications.EnqueueDeliveryInput{
		IdempotencyKey: "event-1:" + policy.ID + ":" + destination.ID,
		PolicyID:       policy.ID,
		DestinationID:  destination.ID,
		EventType:      "ticket_assigned",
		SubjectType:    "ticket",
		SubjectID:      "ticket-1",
		Message:        "Assigned CORE-1",
		Payload:        map[string]any{"ticket_key": "CORE-1"},
	})
	if err != nil {
		t.Fatalf("enqueue delivery: %v", err)
	}

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	listReq := httptest.NewRequest(http.MethodGet, "/api/notification-deliveries?status=queued", nil)
	listReq.AddCookie(session)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, listReq)
	if list.Code != http.StatusOK {
		t.Fatalf("expected delivery list status 200, got %d: %s", list.Code, list.Body.String())
	}
	var listBody struct {
		Metadata struct {
			Count int `json:"count"`
		} `json:"metadata"`
		Status struct {
			Items []notificationDeliveryResourceBody `json:"items"`
		} `json:"status"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode delivery list: %v", err)
	}
	if listBody.Metadata.Count != 1 || len(listBody.Status.Items) != 1 || listBody.Status.Items[0].Metadata.ID != delivery.ID {
		t.Fatalf("unexpected delivery list: %#v", listBody)
	}
	if bytes.Contains(list.Body.Bytes(), []byte("logger://")) {
		t.Fatalf("delivery list leaked Shoutrrr URL: %s", list.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/notification-deliveries/"+delivery.ID, nil)
	getReq.AddCookie(session)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getReq)
	if get.Code != http.StatusOK {
		t.Fatalf("expected get delivery status 200, got %d: %s", get.Code, get.Body.String())
	}
	fetched := decodeNotificationDeliveryResource(t, get.Body.Bytes())
	if fetched.Metadata.PolicyName != "ops" || fetched.Metadata.DestinationService != "logger" || fetched.Spec.EventType != "ticket_assigned" || fetched.Status.State != "queued" {
		t.Fatalf("unexpected delivery resource: %#v", fetched)
	}

	retryMissingCSRFReq := httptest.NewRequest(http.MethodPost, "/api/notification-deliveries/"+delivery.ID+"/retry", nil)
	retryMissingCSRFReq.AddCookie(session)
	retryMissingCSRF := httptest.NewRecorder()
	handler.ServeHTTP(retryMissingCSRF, retryMissingCSRFReq)
	if retryMissingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF retry status 403, got %d: %s", retryMissingCSRF.Code, retryMissingCSRF.Body.String())
	}

	if _, err := db.SQL.ExecContext(ctx, `
		UPDATE notification_deliveries
		SET status = 'failed', last_error = 'temporary failure'
		WHERE id = ?
	`, delivery.ID); err != nil {
		t.Fatalf("mark delivery failed: %v", err)
	}
	retryReq := httptest.NewRequest(http.MethodPost, "/api/notification-deliveries/"+delivery.ID+"/retry", nil)
	addSessionCSRF(retryReq, session, csrf)
	retry := httptest.NewRecorder()
	handler.ServeHTTP(retry, retryReq)
	if retry.Code != http.StatusOK {
		t.Fatalf("expected retry status 200, got %d: %s", retry.Code, retry.Body.String())
	}
	retried := decodeNotificationDeliveryResource(t, retry.Body.Bytes())
	if retried.Status.State != "queued" || retried.Status.LastError != "" || retried.Status.NextAttemptAt == "" {
		t.Fatalf("unexpected retried delivery: %#v", retried)
	}
}

func TestNotificationDeliveryEndpointsRequireManagePermission(t *testing.T) {
	ctx := context.Background()
	db, _ := openBackendTestDB(t, ctx)
	authService := auth.NewService(db.SQL)
	user, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "viewer"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authz.NewSQLEvaluator(db.SQL)),
		WithNotificationService(notifications.NewService(db.SQL)),
	)
	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": user.Username,
		"password": user.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)

	req := httptest.NewRequest(http.MethodGet, "/api/notification-deliveries", nil)
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden delivery list, got %d: %s", rec.Code, rec.Body.String())
	}
}

type notificationDeliveryResourceBody struct {
	Metadata struct {
		ID                 string `json:"id"`
		IdempotencyKey     string `json:"idempotency_key"`
		ScopeType          string `json:"scope_type"`
		ProjectID          string `json:"project_id"`
		PolicyID           string `json:"policy_id"`
		PolicyName         string `json:"policy_name"`
		DestinationID      string `json:"destination_id"`
		DestinationName    string `json:"destination_name"`
		DestinationService string `json:"destination_service"`
	} `json:"metadata"`
	Spec struct {
		EventType   string         `json:"event_type"`
		SubjectType string         `json:"subject_type"`
		SubjectID   string         `json:"subject_id"`
		Message     string         `json:"message"`
		Payload     map[string]any `json:"payload"`
		MaxAttempts int            `json:"max_attempts"`
	} `json:"spec"`
	Status struct {
		State         string `json:"state"`
		AttemptCount  int    `json:"attempt_count"`
		NextAttemptAt string `json:"next_attempt_at"`
		LastError     string `json:"last_error"`
	} `json:"status"`
}

func decodeNotificationDeliveryResource(t *testing.T, data []byte) notificationDeliveryResourceBody {
	t.Helper()

	var body notificationDeliveryResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode notification delivery resource: %v", err)
	}
	return body
}
