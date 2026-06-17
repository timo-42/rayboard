package notifications

import (
	"context"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/automation"
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
	service := NewService(db.SQL, WithEventStore(eventStore), WithRunStore(automation.NewRunStore(db.SQL)))
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

func TestNotificationHookPreviewRecordsRun(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "actor")

	service := NewService(db.SQL, WithRunStore(automation.NewRunStore(db.SQL)))
	hook, err := service.CreateHook(ctx, CreateHookInput{
		Name:        "preview",
		ScopeType:   PolicyScopeGlobal,
		ActorUserID: "actor",
		EventTypes:  []string{"comment_added"},
		Enabled:     true,
		Engine: HookEngine{Type: HookEngineLua, Script: `
return {
  message = "Preview " .. notification.plan.message,
  payload = { seen = notification.plan.payload.ticket_id },
  destination_ids = { notification.policy.destination_ids[1] }
}
`},
	})
	if err != nil {
		t.Fatalf("create notification hook: %v", err)
	}

	result, err := service.PreviewHook(ctx, hook.ID, PreviewHookInput{
		EventType:      "comment_added",
		Message:        "comment",
		Payload:        map[string]any{"ticket_id": "ticket-1"},
		DestinationIDs: []string{"dest-1", "dest-2"},
	})
	if err != nil {
		t.Fatalf("preview notification hook: %v", err)
	}
	if result.Run.ID == "" || result.Run.Status != automation.StatusSucceeded {
		t.Fatalf("unexpected preview run: %#v", result.Run)
	}
	if result.Plan.Message != "Preview comment" || result.Plan.Payload["seen"] != "ticket-1" || len(result.Plan.DestinationIDs) != 1 || result.Plan.DestinationIDs[0] != "dest-1" {
		t.Fatalf("unexpected preview plan: %#v", result.Plan)
	}
	runs, err := service.ListHookRuns(ctx, hook.ID, 10, 0)
	if err != nil {
		t.Fatalf("list hook runs: %v", err)
	}
	if len(runs) != 1 || runs[0].ID != result.Run.ID || runs[0].TriggerType != "notification_hook_preview" {
		t.Fatalf("unexpected hook runs: %#v", runs)
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
	service := NewService(db.SQL, WithEventStore(eventStore), WithRunStore(automation.NewRunStore(db.SQL)))
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
	runs, err := service.ListHookRuns(ctx, "not-a-hook", 10, 0)
	if err == nil || len(runs) != 0 {
		t.Fatalf("expected missing hook run lookup to fail, got runs=%#v err=%v", runs, err)
	}
	hooks, err := service.ListHooks(ctx, ListHooksInput{ScopeType: PolicyScopeProject, ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("list hooks: %v", err)
	}
	if len(hooks) != 1 {
		t.Fatalf("expected one hook, got %#v", hooks)
	}
	runs, err = service.ListHookRuns(ctx, hooks[0].ID, 10, 0)
	if err != nil {
		t.Fatalf("list hook runs: %v", err)
	}
	if len(runs) != 1 || runs[0].TriggerType != "notification_hook" || runs[0].Status != automation.StatusSucceeded {
		t.Fatalf("unexpected recorded hook runs: %#v", runs)
	}
}
