package notifications

import (
	"context"
	"errors"
	"testing"
)

func TestDestinationCRUD(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	service := NewService(db.SQL)

	created, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        " Ops Alerts ",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}
	if created.ID == "" || created.Name != "ops alerts" || created.ScopeType != DestinationScopeGlobal || created.Service != "logger" || !created.URLSet || !created.Enabled {
		t.Fatalf("unexpected destination: %#v", created)
	}

	items, err := service.ListDestinations(ctx, ListDestinationsInput{ScopeType: DestinationScopeGlobal})
	if err != nil {
		t.Fatalf("list destinations: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID || !items[0].URLSet {
		t.Fatalf("unexpected destination list: %#v", items)
	}

	disabled := false
	name := "Ops Archive"
	updated, err := service.UpdateDestination(ctx, created.ID, UpdateDestinationInput{
		Name:    &name,
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("update destination: %v", err)
	}
	if updated.Name != "ops archive" || updated.Enabled || updated.Service != "logger" || !updated.URLSet {
		t.Fatalf("unexpected updated destination: %#v", updated)
	}

	if err := service.DeleteDestination(ctx, created.ID); err != nil {
		t.Fatalf("delete destination: %v", err)
	}
	if _, err := service.GetDestination(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestDestinationValidation(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	service := NewService(db.SQL)

	if _, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "bad",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "not a url",
		Enabled:     true,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid URL validation error, got %v", err)
	}

	if _, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "project",
		ScopeType:   DestinationScopeProject,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid project scope validation error, got %v", err)
	}

	if _, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "dup",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	}); err != nil {
		t.Fatalf("create first duplicate candidate: %v", err)
	}
	if _, err := service.CreateDestination(ctx, CreateDestinationInput{
		Name:        "Dup",
		ScopeType:   DestinationScopeGlobal,
		ShoutrrrURL: "logger://",
		Enabled:     true,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected duplicate validation error, got %v", err)
	}
}
