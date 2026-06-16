package notifications

import (
	"context"
	"errors"
	"testing"
)

func TestDeliveryEnqueueListAndIdempotency(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	service := NewService(db.SQL)

	destination := mustNotificationDestination(t, ctx, service, CreateDestinationInput{
		Name:        "ops",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	policy := mustNotificationPolicy(t, ctx, service, CreatePolicyInput{
		Name:           "ops",
		ScopeType:      PolicyScopeGlobal,
		EventTypes:     []string{"ticket_assigned"},
		DestinationIDs: []string{destination.ID},
		Enabled:        true,
	})

	created, err := service.EnqueueDelivery(ctx, EnqueueDeliveryInput{
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
	if created.ID == "" || created.Status != DeliveryStatusQueued || created.PolicyName != "ops" || created.DestinationName != "ops" || created.DestinationService != "logger" || created.MaxAttempts != 3 {
		t.Fatalf("unexpected delivery: %#v", created)
	}

	duplicate, err := service.EnqueueDelivery(ctx, EnqueueDeliveryInput{
		IdempotencyKey: created.IdempotencyKey,
		PolicyID:       policy.ID,
		DestinationID:  destination.ID,
		EventType:      "ticket_assigned",
		Message:        "Duplicate",
	})
	if err != nil {
		t.Fatalf("idempotent enqueue delivery: %v", err)
	}
	if duplicate.ID != created.ID || duplicate.Message != created.Message {
		t.Fatalf("expected existing delivery for idempotency key, got %#v", duplicate)
	}

	items, err := service.ListDeliveries(ctx, ListDeliveriesInput{ScopeType: PolicyScopeGlobal, Status: DeliveryStatusQueued})
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("unexpected delivery list: %#v", items)
	}

	if _, err := service.ListDeliveries(ctx, ListDeliveriesInput{ScopeType: PolicyScopeGlobal, Status: "unknown"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid status validation error, got %v", err)
	}
}

func TestDeliveryRetry(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	service := NewService(db.SQL)
	destination := mustNotificationDestination(t, ctx, service, CreateDestinationInput{
		Name:        "ops",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	policy := mustNotificationPolicy(t, ctx, service, CreatePolicyInput{
		Name:           "ops",
		ScopeType:      PolicyScopeGlobal,
		EventTypes:     []string{"ticket_assigned"},
		DestinationIDs: []string{destination.ID},
		Enabled:        true,
	})
	delivery, err := service.EnqueueDelivery(ctx, EnqueueDeliveryInput{
		PolicyID:      policy.ID,
		DestinationID: destination.ID,
		EventType:     "ticket_assigned",
		Message:       "Assigned CORE-1",
	})
	if err != nil {
		t.Fatalf("enqueue delivery: %v", err)
	}
	if _, err := service.RetryDelivery(ctx, delivery.ID); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected queued retry validation error, got %v", err)
	}
	if _, err := db.SQL.ExecContext(ctx, `
		UPDATE notification_deliveries
		SET status = 'failed', last_error = 'temporary failure'
		WHERE id = ?
	`, delivery.ID); err != nil {
		t.Fatalf("mark delivery failed: %v", err)
	}
	retried, err := service.RetryDelivery(ctx, delivery.ID)
	if err != nil {
		t.Fatalf("retry delivery: %v", err)
	}
	if retried.Status != DeliveryStatusQueued || retried.LastError != "" || retried.NextAttemptAt == nil {
		t.Fatalf("unexpected retried delivery: %#v", retried)
	}
}

func mustNotificationDestination(t *testing.T, ctx context.Context, service *Service, input CreateDestinationInput) Destination {
	t.Helper()

	destination, err := service.CreateDestination(ctx, input)
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}
	return destination
}

func mustNotificationPolicy(t *testing.T, ctx context.Context, service *Service, input CreatePolicyInput) Policy {
	t.Helper()

	policy, err := service.CreatePolicy(ctx, input)
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	return policy
}
