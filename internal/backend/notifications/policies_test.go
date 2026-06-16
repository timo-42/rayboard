package notifications

import (
	"context"
	"errors"
	"testing"
)

func TestPolicyCRUD(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	service := NewService(db.SQL)

	destination, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "ops",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}

	created, err := service.CreatePolicy(ctx, CreatePolicyInput{
		Name:           " Ticket Updates ",
		ScopeType:      PolicyScopeGlobal,
		EventTypes:     []string{"ticket_assigned", "ticket_status_changed", "ticket_assigned"},
		DestinationIDs: []string{destination.ID, destination.ID},
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if created.ID == "" || created.Name != "ticket updates" || len(created.EventTypes) != 2 || len(created.DestinationIDs) != 1 || !created.Enabled {
		t.Fatalf("unexpected created policy: %#v", created)
	}

	items, err := service.ListPolicies(ctx, ListPoliciesInput{ScopeType: PolicyScopeGlobal})
	if err != nil {
		t.Fatalf("list policies: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("unexpected policy list: %#v", items)
	}

	disabled := false
	name := "Ticket Archive"
	updated, err := service.UpdatePolicy(ctx, created.ID, UpdatePolicyInput{
		Name:    &name,
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("update policy: %v", err)
	}
	if updated.Name != "ticket archive" || updated.Enabled {
		t.Fatalf("unexpected updated policy: %#v", updated)
	}

	if err := service.DeletePolicy(ctx, created.ID); err != nil {
		t.Fatalf("delete policy: %v", err)
	}
	if _, err := service.GetPolicy(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted policy not found, got %v", err)
	}
}

func TestPolicyValidation(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationProject(t, ctx, db, "project-1", "CORE")
	seedNotificationProject(t, ctx, db, "project-2", "OPS")
	service := NewService(db.SQL)

	globalDestination, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "global",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create global destination: %v", err)
	}
	projectDestination, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "project",
		ScopeType:   DestinationScopeProject,
		ProjectID:   "project-1",
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create project destination: %v", err)
	}
	otherDestination, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "other",
		ScopeType:   DestinationScopeProject,
		ProjectID:   "project-2",
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create other destination: %v", err)
	}

	if _, err := service.CreatePolicy(ctx, CreatePolicyInput{
		Name:           "bad event",
		ScopeType:      PolicyScopeGlobal,
		EventTypes:     []string{"ticket.updated"},
		DestinationIDs: []string{globalDestination.ID},
		Enabled:        true,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected bad event validation error, got %v", err)
	}

	if _, err := service.CreatePolicy(ctx, CreatePolicyInput{
		Name:           "cross scope",
		ScopeType:      PolicyScopeGlobal,
		EventTypes:     []string{"ticket_assigned"},
		DestinationIDs: []string{projectDestination.ID},
		Enabled:        true,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected global policy destination validation error, got %v", err)
	}

	if _, err := service.CreatePolicy(ctx, CreatePolicyInput{
		Name:           "project policy",
		ScopeType:      PolicyScopeProject,
		ProjectID:      "project-1",
		EventTypes:     []string{"ticket_assigned"},
		DestinationIDs: []string{globalDestination.ID, projectDestination.ID},
		Enabled:        true,
	}); err != nil {
		t.Fatalf("expected same-project/global destinations to validate: %v", err)
	}

	if _, err := service.CreatePolicy(ctx, CreatePolicyInput{
		Name:           "wrong project",
		ScopeType:      PolicyScopeProject,
		ProjectID:      "project-1",
		EventTypes:     []string{"ticket_assigned"},
		DestinationIDs: []string{otherDestination.ID},
		Enabled:        true,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected cross-project destination validation error, got %v", err)
	}
}
