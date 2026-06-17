package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestIncomingWebhookLifecycleAndTokenRotation(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithNow(func() time.Time { return now }),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        " GitHub ",
		Direction:   DirectionIncoming,
		Enabled:     true,
		ActorUserID: "actor",
		Engine: EngineSpec{
			Type:   EngineTypeLua,
			Script: `return { ok = true }`,
		},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}
	if created.ID == "" || created.Name != "github" || created.Token == "" || !created.TokenSet || created.TokenRotatedAt == nil {
		t.Fatalf("unexpected created webhook: %#v", created)
	}
	assertNoPlaintextWebhookToken(t, ctx, db, created.Token)

	authenticated, err := service.AuthenticateIncoming(ctx, created.ID, created.Token)
	if err != nil {
		t.Fatalf("authenticate incoming webhook: %v", err)
	}
	if authenticated.ID != created.ID || authenticated.ActorUserID != "actor" {
		t.Fatalf("unexpected authenticated webhook: %#v", authenticated)
	}

	items, err := service.List(ctx, principal, ListInput{ProjectID: "project-1", Direction: DirectionIncoming})
	if err != nil {
		t.Fatalf("list webhooks: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID || items[0].TokenSet != true {
		t.Fatalf("unexpected webhook list: %#v", items)
	}

	enabled := false
	name := "GitHub Archive"
	updated, err := service.Update(ctx, principal, created.ID, UpdateInput{
		Name:    &name,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("update webhook: %v", err)
	}
	if updated.Name != "github archive" || updated.Enabled {
		t.Fatalf("unexpected updated webhook: %#v", updated)
	}
	if _, err := service.AuthenticateIncoming(ctx, created.ID, created.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected disabled webhook auth to fail, got %v", err)
	}

	enabled = true
	if _, err := service.Update(ctx, principal, created.ID, UpdateInput{Enabled: &enabled}); err != nil {
		t.Fatalf("reenable webhook: %v", err)
	}
	rotated, err := service.RotateIncomingToken(ctx, principal, created.ID)
	if err != nil {
		t.Fatalf("rotate webhook token: %v", err)
	}
	if rotated.Token == "" || rotated.Token == created.Token {
		t.Fatalf("expected new rotated token, got %#v", rotated)
	}
	assertNoPlaintextWebhookToken(t, ctx, db, rotated.Token)
	if _, err := service.AuthenticateIncoming(ctx, created.ID, created.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected old token to fail, got %v", err)
	}
	if _, err := service.AuthenticateIncoming(ctx, created.ID, rotated.Token); err != nil {
		t.Fatalf("expected rotated token to authenticate: %v", err)
	}

	if err := service.Delete(ctx, principal, created.ID); err != nil {
		t.Fatalf("delete webhook: %v", err)
	}
	if _, err := service.Get(ctx, principal, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted webhook not found, got %v", err)
	}
}

func TestIncomingWebhookValidationAndRBAC(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "viewer", false)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
	)

	viewer := authz.Principal{UserID: "viewer", AuthKind: authz.AuthKindSession}
	if _, err := service.Create(ctx, viewer, CreateInput{
		ProjectID:   "project-1",
		Name:        "denied",
		Direction:   DirectionIncoming,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return {}`},
	}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden create, got %v", err)
	}

	admin := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	if _, err := service.Create(ctx, admin, CreateInput{
		ProjectID:   "project-1",
		Name:        "bad",
		Direction:   DirectionIncoming,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid lua engine validation, got %v", err)
	}
	if _, err := service.Create(ctx, admin, CreateInput{
		ProjectID:   "project-1",
		Name:        "outgoing missing events",
		Direction:   DirectionOutgoing,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return {}`},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected outgoing event type validation, got %v", err)
	}
}

func TestOutgoingWebhookDefinitionLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "Delivery Events",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine: EngineSpec{
			Type:   EngineTypeLua,
			Script: `return { method = "POST", path = "/events", body = event }`,
		},
	})
	if err != nil {
		t.Fatalf("create outgoing webhook: %v", err)
	}
	if created.ID == "" || created.Direction != DirectionOutgoing || created.Token != "" || created.TokenSet || created.TokenRotatedAt != nil || !equalStrings(created.EventTypes, []string{"ticket.updated"}) {
		t.Fatalf("unexpected outgoing webhook: %#v", created)
	}

	listed, err := service.List(ctx, principal, ListInput{ProjectID: "project-1", Direction: DirectionOutgoing})
	if err != nil {
		t.Fatalf("list outgoing webhooks: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID || listed[0].TokenSet {
		t.Fatalf("unexpected outgoing webhook list: %#v", listed)
	}

	if _, err := service.RotateIncomingToken(ctx, principal, created.ID); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected outgoing token rotation validation, got %v", err)
	}

	enabled := false
	updated, err := service.Update(ctx, principal, created.ID, UpdateInput{Enabled: &enabled})
	if err != nil {
		t.Fatalf("disable outgoing webhook: %v", err)
	}
	if updated.Enabled || updated.Direction != DirectionOutgoing || updated.TokenSet {
		t.Fatalf("unexpected disabled outgoing webhook: %#v", updated)
	}

	if err := service.Delete(ctx, principal, created.ID); err != nil {
		t.Fatalf("delete outgoing webhook: %v", err)
	}
	if _, err := service.Get(ctx, principal, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted outgoing webhook not found, got %v", err)
	}
}

func TestOutgoingWebhookDeliveryEnqueue(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	matching, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "ticket updates",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { method = "POST" }`},
	})
	if err != nil {
		t.Fatalf("create matching outgoing webhook: %v", err)
	}
	if _, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "comments",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"comment.created"},
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { method = "POST" }`},
	}); err != nil {
		t.Fatalf("create nonmatching outgoing webhook: %v", err)
	}
	disabled, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "disabled",
		Direction:   DirectionOutgoing,
		Enabled:     false,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { method = "POST" }`},
	})
	if err != nil {
		t.Fatalf("create disabled outgoing webhook: %v", err)
	}
	_ = disabled

	eventStore := events.NewStore(db.SQL)
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:      "ticket.updated",
		ActorID:   "actor",
		ProjectID: "project-1",
		ObjectID:  "ticket-1",
		Data: map[string]any{
			"changes": map[string]any{"status": map[string]string{"old": "todo", "new": "done"}},
		},
	}); err != nil {
		t.Fatalf("append ticket event: %v", err)
	}
	pending, err := eventStore.ListPending(ctx, 10, "ticket.updated")
	if err != nil {
		t.Fatalf("list pending events: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected one pending event, got %#v", pending)
	}

	enqueued, err := service.EnqueueOutgoingDeliveriesForEvent(ctx, pending[0])
	if err != nil {
		t.Fatalf("enqueue outgoing deliveries: %v", err)
	}
	if enqueued != 1 {
		t.Fatalf("expected one outgoing delivery, got %d", enqueued)
	}
	deliveries, err := service.ListOutgoingDeliveries(ctx, principal, matching.ID, 10, 0)
	if err != nil {
		t.Fatalf("list outgoing deliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected one delivery, got %#v", deliveries)
	}
	delivery := deliveries[0]
	if delivery.WebhookID != matching.ID || delivery.WebhookName != matching.Name || delivery.DomainEventID != pending[0].ID || delivery.Status != OutgoingDeliveryStatusQueued || delivery.NextAttemptAt == nil {
		t.Fatalf("unexpected delivery: %#v", delivery)
	}
	gotDelivery, err := service.GetOutgoingDelivery(ctx, principal, delivery.ID)
	if err != nil {
		t.Fatalf("get outgoing delivery: %v", err)
	}
	if gotDelivery.ID != delivery.ID || gotDelivery.WebhookID != matching.ID {
		t.Fatalf("unexpected outgoing delivery get result: %#v", gotDelivery)
	}
	viewer := authz.Principal{UserID: "viewer", AuthKind: authz.AuthKindSession}
	if _, err := service.GetOutgoingDelivery(ctx, viewer, delivery.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden outgoing delivery get, got %v", err)
	}
	eventPayload, ok := delivery.Payload["event"].(map[string]any)
	if !ok || eventPayload["type"] != "ticket.updated" || eventPayload["project_id"] != "project-1" {
		t.Fatalf("unexpected delivery event payload: %#v", delivery.Payload)
	}
	webhookPayload, ok := delivery.Payload["webhook"].(map[string]any)
	if !ok || webhookPayload["id"] != matching.ID || webhookPayload["name"] != matching.Name {
		t.Fatalf("unexpected delivery webhook payload: %#v", delivery.Payload)
	}

	enqueued, err = service.EnqueueOutgoingDeliveriesForEvent(ctx, pending[0])
	if err != nil {
		t.Fatalf("reenqueue outgoing deliveries: %v", err)
	}
	if enqueued != 0 {
		t.Fatalf("expected idempotent reenqueue, got %d", enqueued)
	}
	if got := countWebhookRows(t, ctx, db, "outgoing_webhook_deliveries"); got != 1 {
		t.Fatalf("expected one outgoing webhook delivery row, got %d", got)
	}
}

func TestProcessPendingOutgoingWebhookDeliveriesSendsLuaRequest(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	var gotMethod string
	var gotPath string
	var gotHeader string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.String()
		gotHeader = r.Header.Get("X-Rayboard")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode outgoing request body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithOutgoingBaseURL(server.URL),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	hook, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "sender",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine: EngineSpec{Type: EngineTypeLua, Script: `
return {
  method = "POST",
  path = "/events",
  query = { source = "rayboard" },
  headers = { ["X-Rayboard"] = "webhook" },
  body = { event_type = event.type, webhook_id = webhook.id, attempt = delivery.attempt_count }
}
`},
	})
	if err != nil {
		t.Fatalf("create outgoing webhook: %v", err)
	}
	eventStore := events.NewStore(db.SQL)
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:      "ticket.updated",
		ActorID:   "actor",
		ProjectID: "project-1",
		ObjectID:  "ticket-1",
		Data:      map[string]any{"status": "done"},
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	pending, err := eventStore.ListPending(ctx, 10)
	if err != nil {
		t.Fatalf("list pending events: %v", err)
	}
	if _, err := service.EnqueueOutgoingDeliveriesForEvent(ctx, pending[0]); err != nil {
		t.Fatalf("enqueue outgoing delivery: %v", err)
	}
	staleSendingAt := time.Now().UTC().Add(-20 * time.Minute)
	if _, err := db.SQL.ExecContext(ctx, `
		UPDATE outgoing_webhook_deliveries
		SET status = ?, updated_at = ?
		WHERE webhook_id = ?
	`, OutgoingDeliveryStatusSending, formatTime(staleSendingAt), hook.ID); err != nil {
		t.Fatalf("mark delivery stale sending: %v", err)
	}

	processed, err := service.ProcessPendingDeliveries(ctx, ProcessDeliveriesInput{Limit: 10})
	if err != nil {
		t.Fatalf("process outgoing delivery: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected one processed delivery, got %d", processed)
	}
	deliveries, err := service.ListOutgoingDeliveries(ctx, principal, hook.ID, 10, 0)
	if err != nil {
		t.Fatalf("list outgoing deliveries: %v", err)
	}
	if len(deliveries) != 1 || deliveries[0].Status != OutgoingDeliveryStatusDelivered || deliveries[0].AttemptCount != 1 || deliveries[0].DeliveredAt == nil {
		t.Fatalf("unexpected delivered outgoing delivery: %#v", deliveries)
	}
	if gotMethod != http.MethodPost || gotPath != "/events?source=rayboard" || gotHeader != "webhook" {
		t.Fatalf("unexpected outgoing request: method=%s path=%s header=%s", gotMethod, gotPath, gotHeader)
	}
	if gotBody["event_type"] != "ticket.updated" || gotBody["webhook_id"] != hook.ID || gotBody["attempt"] != float64(0) {
		t.Fatalf("unexpected outgoing request body: %#v", gotBody)
	}
}

func TestProcessPendingOutgoingWebhookDeliveriesUsesAllowedBaseURLProvider(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithOutgoingBaseURLProvider(staticOutgoingBaseURLProvider{values: []string{server.URL + "/"}}),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	hook, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "settings-base",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { method = "POST", path = "/settings-base", body = event }`},
	})
	if err != nil {
		t.Fatalf("create outgoing webhook: %v", err)
	}
	deliveryID := enqueueTestOutgoingDelivery(t, ctx, db, service, hook.ID)

	processed, err := service.ProcessPendingDeliveries(ctx, ProcessDeliveriesInput{Limit: 10})
	if err != nil {
		t.Fatalf("process outgoing delivery: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected one processed delivery, got %d", processed)
	}
	delivery, err := service.GetOutgoingDelivery(ctx, principal, deliveryID)
	if err != nil {
		t.Fatalf("get delivery: %v", err)
	}
	if delivery.Status != OutgoingDeliveryStatusDelivered || gotPath != "/settings-base" {
		t.Fatalf("unexpected provider-backed delivery: delivery=%#v path=%s", delivery, gotPath)
	}
}

func TestProcessPendingOutgoingWebhookDeliveriesRejectsBaseOutsideAllowlist(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	hitCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hitCount++
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	allowed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer allowed.Close()

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithOutgoingBaseURL(server.URL),
		WithOutgoingBaseURLProvider(staticOutgoingBaseURLProvider{values: []string{allowed.URL}}),
		WithNow(func() time.Time { return now }),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	hook, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "denied-base",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { method = "POST", path = "/denied", body = event }`},
	})
	if err != nil {
		t.Fatalf("create outgoing webhook: %v", err)
	}
	deliveryID := enqueueTestOutgoingDelivery(t, ctx, db, service, hook.ID)

	processed, err := service.ProcessPendingDeliveries(ctx, ProcessDeliveriesInput{Limit: 10})
	if processed != 1 || !errors.Is(err, ErrValidation) {
		t.Fatalf("expected one validation failure, got processed=%d err=%v", processed, err)
	}
	if hitCount != 0 {
		t.Fatalf("disallowed outgoing base should not receive request, got %d hits", hitCount)
	}
	delivery, err := service.GetOutgoingDelivery(ctx, principal, deliveryID)
	if err != nil {
		t.Fatalf("get delivery: %v", err)
	}
	if delivery.Status != OutgoingDeliveryStatusQueued || delivery.AttemptCount != 1 || !strings.Contains(delivery.LastError, "not allowed") {
		t.Fatalf("unexpected denied-base delivery state: %#v", delivery)
	}
}

func TestProcessPendingOutgoingWebhookDeliveriesRetriesAndCanBeManuallyQueued(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithOutgoingBaseURL(server.URL),
		WithNow(func() time.Time { return now }),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	hook, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "sender",
		Direction:   DirectionOutgoing,
		Enabled:     true,
		ActorUserID: "actor",
		EventTypes:  []string{"ticket.updated"},
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { method = "POST", path = "/events", body = event }`},
	})
	if err != nil {
		t.Fatalf("create outgoing webhook: %v", err)
	}
	eventStore := events.NewStore(db.SQL, events.WithNow(func() time.Time { return now }))
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:      "ticket.updated",
		ActorID:   "actor",
		ProjectID: "project-1",
		ObjectID:  "ticket-1",
		Data:      map[string]any{"status": "done"},
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	pending, err := eventStore.ListPending(ctx, 10)
	if err != nil {
		t.Fatalf("list pending events: %v", err)
	}
	if _, err := service.EnqueueOutgoingDeliveriesForEvent(ctx, pending[0]); err != nil {
		t.Fatalf("enqueue outgoing delivery: %v", err)
	}
	deliveries, err := service.ListOutgoingDeliveries(ctx, principal, hook.ID, 10, 0)
	if err != nil {
		t.Fatalf("list outgoing deliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected one outgoing delivery, got %#v", deliveries)
	}
	deliveryID := deliveries[0].ID
	if _, err := db.SQL.ExecContext(ctx, `UPDATE outgoing_webhook_deliveries SET max_attempts = 2 WHERE id = ?`, deliveryID); err != nil {
		t.Fatalf("set max attempts: %v", err)
	}

	processed, err := service.ProcessPendingDeliveries(ctx, ProcessDeliveriesInput{Limit: 10})
	if processed != 1 || !errors.Is(err, ErrDelivery) {
		t.Fatalf("expected one retryable delivery failure, got processed=%d err=%v", processed, err)
	}
	failedOnce, err := service.GetOutgoingDelivery(ctx, principal, deliveryID)
	if err != nil {
		t.Fatalf("get failed-once delivery: %v", err)
	}
	if failedOnce.Status != OutgoingDeliveryStatusQueued || failedOnce.AttemptCount != 1 || failedOnce.NextAttemptAt == nil || failedOnce.LastAttemptAt == nil {
		t.Fatalf("unexpected first failure state: %#v", failedOnce)
	}
	if _, err := db.SQL.ExecContext(ctx, `UPDATE outgoing_webhook_deliveries SET next_attempt_at = ? WHERE id = ?`, formatTime(now), deliveryID); err != nil {
		t.Fatalf("reset next attempt: %v", err)
	}

	processed, err = service.ProcessPendingDeliveries(ctx, ProcessDeliveriesInput{Limit: 10})
	if processed != 1 || !errors.Is(err, ErrDelivery) {
		t.Fatalf("expected final delivery failure, got processed=%d err=%v", processed, err)
	}
	failed, err := service.GetOutgoingDelivery(ctx, principal, deliveryID)
	if err != nil {
		t.Fatalf("get failed delivery: %v", err)
	}
	if failed.Status != OutgoingDeliveryStatusFailed || failed.AttemptCount != 2 || failed.NextAttemptAt != nil || failed.LastError == "" {
		t.Fatalf("unexpected final failure state: %#v", failed)
	}
	retried, err := service.RetryOutgoingDelivery(ctx, principal, deliveryID)
	if err != nil {
		t.Fatalf("retry outgoing delivery: %v", err)
	}
	if retried.Status != OutgoingDeliveryStatusQueued || retried.NextAttemptAt == nil || retried.LastError != "" {
		t.Fatalf("unexpected retried delivery: %#v", retried)
	}
}

func TestIncomingWebhookReceiveRunsLuaAndRecordsHistory(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithRunStore(automation.NewRunStore(db.SQL)),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "receiver",
		Direction:   DirectionIncoming,
		Enabled:     true,
		ActorUserID: "actor",
		Engine: EngineSpec{
			Type: EngineTypeLua,
			Script: `
				rayboard.log("event " .. request.payload.event)
				return { accepted = true, event = request.payload.event }
			`,
		},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	result, err := service.ReceiveIncoming(ctx, created.ID, created.Token, IncomingInput{
		Headers: map[string]string{"x-event": "push"},
		Query:   map[string]string{"dry_run": "true"},
		Payload: map[string]any{"event": "opened"},
	})
	if err != nil {
		t.Fatalf("receive incoming webhook: %v", err)
	}
	if result.Run.Status != automation.StatusSucceeded {
		t.Fatalf("expected succeeded run, got %#v", result.Run)
	}
	output, ok := result.Run.Output["output"].(map[string]any)
	if !ok || output["accepted"] != true || output["event"] != "opened" {
		t.Fatalf("unexpected run output: %#v", result.Run.Output)
	}
	logs, ok := result.Run.Output["logs"].([]any)
	if !ok || len(logs) != 1 || logs[0] != "event opened" {
		t.Fatalf("unexpected run logs: %#v", result.Run.Output)
	}
	if result.Webhook.LastError != "" {
		t.Fatalf("expected empty last error, got %#v", result.Webhook)
	}

	rejectScript := `return { reject = { message = "missing signature" } }`
	if _, err := service.Update(ctx, principal, created.ID, UpdateInput{Engine: &EngineSpec{Type: EngineTypeLua, Script: rejectScript}}); err != nil {
		t.Fatalf("update webhook reject script: %v", err)
	}
	rejected, err := service.ReceiveIncoming(ctx, created.ID, created.Token, IncomingInput{Payload: map[string]any{"event": "closed"}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for reject, got %v", err)
	}
	if rejected.Run.Status != automation.StatusFailed || rejected.Run.Error == "" {
		t.Fatalf("expected rejected failed run, got %#v", rejected.Run)
	}
	rejectOutput, ok := rejected.Run.Output["output"].(map[string]any)
	if !ok || rejectOutput["reject"] == nil {
		t.Fatalf("expected reject output, got %#v", rejected.Run.Output)
	}

	failingScript := `error("rejected")`
	if _, err := service.Update(ctx, principal, created.ID, UpdateInput{Engine: &EngineSpec{Type: EngineTypeLua, Script: failingScript}}); err != nil {
		t.Fatalf("update webhook script: %v", err)
	}
	failed, err := service.ReceiveIncoming(ctx, created.ID, created.Token, IncomingInput{Payload: map[string]any{"event": "closed"}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if failed.Run.Status != automation.StatusFailed || failed.Run.Error == "" {
		t.Fatalf("expected failed run, got %#v", failed.Run)
	}
	if failed.Webhook.LastError == "" {
		t.Fatalf("expected webhook last error, got %#v", failed.Webhook)
	}

	runs, err := service.ListRuns(ctx, principal, created.ID, 10, 0)
	if err != nil {
		t.Fatalf("list webhook runs: %v", err)
	}
	if len(runs) != 3 || runs[0].Status != automation.StatusFailed || runs[1].Status != automation.StatusFailed || runs[2].Status != automation.StatusSucceeded {
		t.Fatalf("unexpected webhook runs: %#v", runs)
	}
}

func TestIncomingWebhookLuaRayboardHelpersActAsActor(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	authorizer := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")),
		authz.UserBinding("actor", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	trackerService := tracker.NewService(db.SQL, authorizer)
	searchService := search.NewService(db.SQL, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	service := NewService(
		db.SQL,
		authorizer,
		WithRunStore(automation.NewRunStore(db.SQL)),
		WithTrackerService(trackerService),
		WithSearchService(searchService),
		WithCommentService(commentService),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "actions",
		Direction:   DirectionIncoming,
		Enabled:     true,
		ActorUserID: "actor",
		Engine: EngineSpec{
			Type: EngineTypeLua,
			Script: `
local ticket, err = rayboard.create_ticket({
  project_id = "project-1",
  title = request.payload.title,
  description = "Created from webhook Lua",
  start_date = "2026-07-01",
  due_date = "2026-07-15",
  labels = {"Webhook", "Lua"}
})
if err then error(err.message) end

local comment, comment_err = rayboard.comment({
  ticket_id = ticket.id,
  body = "Webhook helper comment"
})
if comment_err then error(comment_err.message) end

local updated, update_err = rayboard.update_ticket({
  ticket_id = ticket.id,
  priority = "High",
  labels = {"automation", "Webhook"}
})
if update_err then error(update_err.message) end

local fetched, get_err = rayboard.get_ticket({ ticket_id = updated.id })
if get_err then error(get_err.message) end

local results, search_err = rayboard.search({
  project_id = "project-1",
  filter = 'labels == "automation"',
  limit = 10
})
if search_err then error(search_err.message) end

rayboard.log(fetched.key .. ":" .. fetched.reporter_id .. ":" .. fetched.start_date .. ":" .. fetched.due_date .. ":" .. tostring(#results.items))
return { ticket_id = fetched.id, key = fetched.key, comments = 1 }
`,
		},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	result, err := service.ReceiveIncoming(ctx, created.ID, created.Token, IncomingInput{Payload: map[string]any{"title": "Lua-created ticket"}})
	if err != nil {
		t.Fatalf("receive incoming webhook: %v", err)
	}
	if result.Run.Status != automation.StatusSucceeded {
		t.Fatalf("unexpected run status: %#v", result.Run)
	}
	if countWebhookRows(t, ctx, db, "tickets") != 1 {
		t.Fatalf("expected one ticket")
	}
	if countWebhookRows(t, ctx, db, "ticket_comments") != 1 {
		t.Fatalf("expected one comment")
	}
	output, ok := result.Run.Output["output"].(map[string]any)
	if !ok || output["key"] != "WEB-1" || output["comments"] != float64(1) {
		t.Fatalf("unexpected helper output: %#v", result.Run.Output)
	}
	logs, ok := result.Run.Output["logs"].([]any)
	if !ok || len(logs) != 1 || logs[0] != "WEB-1:actor:2026-07-01:2026-07-15:1" {
		t.Fatalf("unexpected helper logs: %#v", result.Run.Output)
	}

	var reporterID string
	if err := db.SQL.QueryRowContext(ctx, "SELECT reporter_id FROM tickets LIMIT 1").Scan(&reporterID); err != nil {
		t.Fatalf("query reporter: %v", err)
	}
	if reporterID != "actor" {
		t.Fatalf("expected actor reporter, got %q", reporterID)
	}
}

func TestIncomingWebhookAIRayboardActionsActAsActor(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	var receivedAuth string
	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		var body struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode OpenRouter request: %v", err)
		}
		if len(body.Messages) > 1 {
			receivedPrompt = body.Messages[1].Content
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "gen_webhook",
			"choices": []map[string]any{{
				"message": map[string]any{
					"role":    "assistant",
					"content": `{"accepted":true,"actions":[{"type":"create_ticket","input":{"project_id":"project-1","title":"AI-created ticket","description":"Created from incoming AI webhook","labels":["ai","webhook"]}}]}`,
				},
			}},
			"usage": map[string]any{"prompt_tokens": 11, "completion_tokens": 7},
		})
	}))
	defer server.Close()

	authorizer := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")),
		authz.UserBinding("actor", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	trackerService := tracker.NewService(db.SQL, authorizer)
	searchService := search.NewService(db.SQL, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	openRouterService := openrouter.NewService(db.SQL, openrouter.WithBaseURL(server.URL))
	provider, err := openRouterService.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "Default",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-or-secret",
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("create OpenRouter provider: %v", err)
	}
	service := NewService(
		db.SQL,
		authorizer,
		WithRunStore(automation.NewRunStore(db.SQL)),
		WithTrackerService(trackerService),
		WithSearchService(searchService),
		WithCommentService(commentService),
		WithOpenRouterService(openRouterService),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "ai receiver",
		Direction:   DirectionIncoming,
		Enabled:     true,
		ActorUserID: "actor",
		Engine: EngineSpec{
			Type:       EngineTypeAI,
			Prompt:     "Create a ticket from the incoming payload.",
			ProviderID: provider.ID,
		},
	})
	if err != nil {
		t.Fatalf("create AI webhook: %v", err)
	}

	result, err := service.ReceiveIncoming(ctx, created.ID, created.Token, IncomingInput{Payload: map[string]any{"title": "from request"}})
	if err != nil {
		t.Fatalf("receive AI webhook: %v", err)
	}
	if result.Run.Status != automation.StatusSucceeded {
		t.Fatalf("unexpected AI run: %#v", result.Run)
	}
	if receivedAuth != "Bearer sk-or-secret" {
		t.Fatalf("unexpected OpenRouter auth header: %q", receivedAuth)
	}
	if !strings.Contains(receivedPrompt, "Create a ticket from the incoming payload.") || !strings.Contains(receivedPrompt, `"request"`) {
		t.Fatalf("unexpected AI prompt: %s", receivedPrompt)
	}
	if strings.Contains(receivedPrompt, "sk-or-secret") {
		t.Fatalf("AI prompt leaked OpenRouter secret: %s", receivedPrompt)
	}
	if countWebhookRows(t, ctx, db, "tickets") != 1 {
		t.Fatalf("expected one AI-created ticket")
	}
	output, ok := result.Run.Output["output"].(map[string]any)
	if !ok || output["accepted"] != true || output["action_results"] == nil || output["provider_id"] != provider.ID {
		t.Fatalf("unexpected AI output: %#v", result.Run.Output)
	}
	encoded, err := json.Marshal(result.Run)
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	if strings.Contains(string(encoded), "sk-or-secret") || strings.Contains(string(encoded), "Create a ticket from the incoming payload") {
		t.Fatalf("run history leaked secret or prompt: %s", string(encoded))
	}
}

func TestIncomingWebhookDisabledActorCannotExecute(t *testing.T) {
	ctx := context.Background()
	db := openWebhookTestDB(t, ctx)
	seedWebhookProject(t, ctx, db, "project-1")
	seedWebhookUser(t, ctx, db, "actor", false)
	seedWebhookUser(t, ctx, db, "admin", false)

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("admin", authz.RoleProjectOwner, authz.ProjectScope("project-1")))),
		WithRunStore(automation.NewRunStore(db.SQL)),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}
	created, err := service.Create(ctx, principal, CreateInput{
		ProjectID:   "project-1",
		Name:        "disabled-actor",
		Direction:   DirectionIncoming,
		Enabled:     true,
		ActorUserID: "actor",
		Engine:      EngineSpec{Type: EngineTypeLua, Script: `return { ok = true }`},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}
	if _, err := db.SQL.ExecContext(ctx, "UPDATE users SET is_disabled = 1 WHERE id = ?", "actor"); err != nil {
		t.Fatalf("disable actor: %v", err)
	}

	result, err := service.ReceiveIncoming(ctx, created.ID, created.Token, IncomingInput{Payload: map[string]any{"ok": true}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if result.Run.Status != automation.StatusFailed || result.Webhook.LastError == "" {
		t.Fatalf("expected failed run and last error, got %#v", result)
	}
}

type staticOutgoingBaseURLProvider struct {
	values []string
	err    error
}

func (provider staticOutgoingBaseURLProvider) OutgoingWebhookBaseURLs(context.Context) ([]string, error) {
	if provider.err != nil {
		return nil, provider.err
	}
	return append([]string(nil), provider.values...), nil
}

func enqueueTestOutgoingDelivery(t *testing.T, ctx context.Context, db *store.DB, service *Service, webhookID string) string {
	t.Helper()

	eventStore := events.NewStore(db.SQL)
	if err := eventStore.Append(ctx, nil, events.Event{
		Type:      "ticket.updated",
		ActorID:   "actor",
		ProjectID: "project-1",
		ObjectID:  "ticket-1",
		Data:      map[string]any{"status": "done"},
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	pending, err := eventStore.ListPending(ctx, 10)
	if err != nil {
		t.Fatalf("list pending events: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("expected pending event")
	}
	if _, err := service.EnqueueOutgoingDeliveriesForEvent(ctx, pending[0]); err != nil {
		t.Fatalf("enqueue outgoing delivery: %v", err)
	}
	var deliveryID string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT id
		FROM outgoing_webhook_deliveries
		WHERE webhook_id = ?
	`, webhookID).Scan(&deliveryID); err != nil {
		t.Fatalf("get enqueued outgoing delivery: %v", err)
	}
	return deliveryID
}

func openWebhookTestDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "rayboard.sqlite"))
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
	return db
}

func seedWebhookProject(t *testing.T, ctx context.Context, db *store.DB, id string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES (?, ?, ?)
	`, id, "WEB", "Webhooks"); err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func seedWebhookUser(t *testing.T, ctx context.Context, db *store.DB, id string, disabled bool) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, is_disabled)
		VALUES (?, ?, ?, ?)
	`, id, id, id, disabled); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func assertNoPlaintextWebhookToken(t *testing.T, ctx context.Context, db *store.DB, token string) {
	t.Helper()

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM webhooks WHERE token_hash = ?", token).Scan(&count); err != nil {
		t.Fatalf("count plaintext webhook token: %v", err)
	}
	if count != 0 {
		t.Fatalf("webhook token stored in plaintext")
	}
}

func countWebhookRows(t *testing.T, ctx context.Context, db *store.DB, table string) int {
	t.Helper()

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}
	return count
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
