package notifications

import (
	"context"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/events"
)

func TestNotificationHookLuaSuppressesDelivery(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "actor")
	seedNotificationUser(t, ctx, db, "reporter")
	seedNotificationUser(t, ctx, db, "assignee")
	seedNotificationTicket(t, ctx, db, "ticket-1", "AUTO-1", "reporter", "assignee")

	eventStore := events.NewStore(db.SQL)
	service := NewService(db.SQL, WithEventStore(eventStore))
	destination := mustNotificationDestination(t, ctx, service, CreateDestinationInput{
		Name:        "global",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	mustNotificationPolicy(t, ctx, service, CreatePolicyInput{
		Name:           "comments",
		ScopeType:      PolicyScopeGlobal,
		EventTypes:     []string{"comment_added"},
		DestinationIDs: []string{destination.ID},
		Enabled:        true,
	})
	if _, err := service.CreateHook(ctx, CreateHookInput{
		Name:        "suppress comments",
		ScopeType:   PolicyScopeGlobal,
		ActorUserID: "actor",
		EventTypes:  []string{"comment_added"},
		Enabled:     true,
		Engine:      HookEngine{Type: HookEngineLua, Script: `return { suppress = true }`},
	}); err != nil {
		t.Fatalf("create notification hook: %v", err)
	}

	if err := eventStore.Append(ctx, nil, events.Event{
		Type:     "comment.created",
		ActorID:  "actor",
		ObjectID: "comment-1",
		Data:     map[string]any{"ticket_id": "ticket-1"},
	}); err != nil {
		t.Fatalf("append comment event: %v", err)
	}
	processed, err := service.ProcessPendingDomainEvents(ctx, 10)
	if err != nil {
		t.Fatalf("process domain events: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected one processed event, got %d", processed)
	}
	if got := countDeliveries(t, ctx, db, "comment_added"); got != 0 {
		t.Fatalf("expected hook to suppress deliveries, got %d", got)
	}
}

func TestNotificationHookLuaTransformsAndRoutesDelivery(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "actor")
	seedNotificationUser(t, ctx, db, "reporter")
	seedNotificationUser(t, ctx, db, "assignee")
	seedNotificationTicket(t, ctx, db, "ticket-1", "AUTO-1", "reporter", "assignee")

	eventStore := events.NewStore(db.SQL)
	service := NewService(db.SQL, WithEventStore(eventStore))
	globalDestination := mustNotificationDestination(t, ctx, service, CreateDestinationInput{
		Name:        "global",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	projectDestination := mustNotificationDestination(t, ctx, service, CreateDestinationInput{
		Name:        "project",
		ScopeType:   DestinationScopeProject,
		ProjectID:   "project-1",
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	blockedDestination := mustNotificationDestination(t, ctx, service, CreateDestinationInput{
		Name:        "blocked",
		ScopeType:   DestinationScopeProject,
		ProjectID:   "project-1",
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	_ = blockedDestination
	mustNotificationPolicy(t, ctx, service, CreatePolicyInput{
		Name:           "status",
		ScopeType:      PolicyScopeProject,
		ProjectID:      "project-1",
		EventTypes:     []string{"ticket_status_changed"},
		DestinationIDs: []string{globalDestination.ID, projectDestination.ID},
		Enabled:        true,
	})
	if _, err := service.CreateHook(ctx, CreateHookInput{
		Name:        "route status",
		ScopeType:   PolicyScopeProject,
		ProjectID:   "project-1",
		ActorUserID: "actor",
		EventTypes:  []string{"ticket_status_changed"},
		Enabled:     true,
		Engine: HookEngine{Type: HookEngineLua, Script: `
return {
  message = "Hooked " .. notification.plan.message,
  destination_ids = { notification.policy.destination_ids[2], "not-allowed" },
  payload = { hooked = true, ticket_id = notification.plan.payload.ticket_id }
}
`},
	}); err != nil {
		t.Fatalf("create notification hook: %v", err)
	}

	if err := eventStore.Append(ctx, nil, events.Event{
		Type:      "ticket.updated",
		ActorID:   "actor",
		ProjectID: "project-1",
		ObjectID:  "ticket-1",
		Data: map[string]any{
			"changes": map[string]any{
				"status": map[string]string{"old": "todo", "new": "done"},
			},
		},
	}); err != nil {
		t.Fatalf("append ticket event: %v", err)
	}
	processed, err := service.ProcessPendingDomainEvents(ctx, 10)
	if err != nil {
		t.Fatalf("process domain events: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected one processed event, got %d", processed)
	}
	deliveries, err := service.ListDeliveries(ctx, ListDeliveriesInput{
		ScopeType: PolicyScopeProject,
		ProjectID: "project-1",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected one routed delivery, got %#v", deliveries)
	}
	delivery := deliveries[0]
	if delivery.DestinationID != projectDestination.ID || delivery.Message != "Hooked AUTO-1 moved to done" {
		t.Fatalf("unexpected transformed delivery: %#v", delivery)
	}
	if delivery.Payload["hooked"] != true || delivery.Payload["ticket_id"] != "ticket-1" {
		t.Fatalf("unexpected transformed payload: %#v", delivery.Payload)
	}
}
